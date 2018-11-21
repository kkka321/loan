package repayremind

import (
	"fmt"
	"math/rand"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/thirdparty/fantasy"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"sync"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
)

// 评分中风险最大值
const (
	AScoreMiddleRiskMaxScore = 624
	BScoreMiddleRiskMaxScore = 569
)

// PreHandle 生成 case前的预处理
func PreHandle(orderID int64) {
	orderData, err := models.GetOrder(orderID)
	if err != nil {
		logs.Error("[case.PreHandle] query order error")
		return
	}

	riskReq := fantasy.NewSingleRequestByOrderPt(&orderData)
	if orderData.IsReloan == int(types.IsReloanYes) {
		score, err := riskReq.GetBScoreV1()
		if err != nil {
			logs.Error("[case.PreHandle] query fantasy score error")
		}
		if score <= BScoreMiddleRiskMaxScore {
			pushToRiskList(orderID, orderData.IsReloan)
		}
	} else {
		score, err := riskReq.GetAScoreV2()
		if err != nil {
			logs.Error("[case.PreHandle] query fantasy score error")
		}
		if score <= AScoreMiddleRiskMaxScore {
			pushToRiskList(orderID, orderData.IsReloan)
		}
	}
}

// GetRepayRemindCaseOrderList 还款提醒
func GetRepayRemindCaseOrderList() (reloanNum, firstLoanNum int, err error) {

	orderM := models.Order{}
	repayPlan := models.RepayPlan{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	today := tools.NaturalDay(0)
	sql := fmt.Sprintf(`SELECT count(IF(o.is_reloan=%d, 1, null)) as reloan_num,count(IF(o.is_reloan=%d, 1, null)) as first_loan_num
FROM %s o
LEFT JOIN %s r ON r.order_id = o.id
WHERE o.check_status IN(%d, %d) AND (r.repay_date = %d)`,
		types.IsReloanYes, types.IsReloanNo,
		orderM.TableName(),
		repayPlan.TableName(), types.LoanStatusWaitRepayment, types.LoanStatusPartialRepayment, today)
	err = o.Raw(sql).QueryRow(&reloanNum, &firstLoanNum)

	return
}

func getQuotaDetail(totalQuota int) (reloanQuota, firstLoanQuota int) {
	reloanNum, firstLoanNum, _ := GetRepayRemindCaseOrderList()
	logs.Warn("reloanNum:", reloanNum, "firstLoanNum:", firstLoanNum)
	todayTotal := reloanNum + firstLoanNum
	if todayTotal == 0 {
		return
	}
	reloanQuota = int(float64(reloanNum) / float64(todayTotal) * float64(totalQuota))
	firstLoanQuota = totalQuota - reloanQuota
	return
}

// FilterAndCreateCases 开始处理已经排序好的订单
func FilterAndCreateCases() {
	// 获取今日生成复贷和首贷还款提醒配额
	singleUserWorkLoad, _ := config.ValidItemInt("ticket_rm0_user_workload")
	_, num, _ := ticket.CanAssignUsersByTicketItem(types.TicketItemRM0)
	totalQuota := singleUserWorkLoad * int(num)
	reloanQuota, firstLoanQuota := getQuotaDetail(totalQuota)

	logs.Debug("CaseHandle: totalQuota:%d, reloanQuota: %d, firstLoanQuota: %d", totalQuota, reloanQuota, firstLoanQuota)

	// 此处之所以要一次性全取出来， 因为分单时是按创建顺序分单， 如果按照过滤逻辑生成工单
	// 则会造成， 用户获取风险度和首付贷工单不均衡， 特全部取出，进行打乱生成
	reloanOrderIDs, reloanLeft := getRandomOrders(types.IsReloanYes, reloanQuota)
	firstLoanOrderIDs, firstLoanLeft := getRandomOrders(types.IsReloanNo, firstLoanQuota)
	logs.Debug("Get orders: reloanActualGet: %d, reloanLeft:%d; firstLoanActualGet: %d,firstLoanLeft: %d",
		len(reloanOrderIDs), reloanLeft, len(firstLoanOrderIDs), firstLoanLeft)
	// 补入
	if reloanLeft > 0 && firstLoanLeft == 0 {
		substituteOrderIDs, _ := getRandomOrders(types.IsReloanNo, reloanLeft)
		firstLoanOrderIDs = append(firstLoanOrderIDs, substituteOrderIDs...)
	} else if reloanLeft == 0 && firstLoanLeft > 0 {
		substituteOrderIDs, _ := getRandomOrders(types.IsReloanYes, firstLoanLeft)
		reloanOrderIDs = append(reloanOrderIDs, substituteOrderIDs...)
	}

	allWillCaseOrders := append(firstLoanOrderIDs, reloanOrderIDs...)
	logs.Debug("max quota: %d, actual get orders: %d, reloan:%d, firstLoan:%d ", totalQuota, len(allWillCaseOrders), len(reloanOrderIDs), len(firstLoanOrderIDs))
	rand.Shuffle(len(allWillCaseOrders), func(i int, j int) {
		allWillCaseOrders[i], allWillCaseOrders[j] = allWillCaseOrders[j], allWillCaseOrders[i]
	})
	for _, orderID := range allWillCaseOrders {
		DailyHandleCase(orderID)
	}
}

// PreHandleTest 测试
func PreHandleTest(orderID int64) {
	isReloan := rand.Intn(2)
	score := rand.Intn(800)
	orderID = int64(isReloan*10000000000+10000000000+score) + rand.Int63n(10)*10000000 + rand.Int63n(10)*1000000
	pushToRiskList(orderID, isReloan)
}

var onceReloanExpire sync.Once
var onceFirstLoanExpire sync.Once

func pushToRiskList(orderID int64, isReloanTag int) {
	key := getScoreListName(isReloanTag)

	redisConn := storage.RedisStorageClient.Get()
	defer redisConn.Close()

	redisConn.Do("SADD", key, orderID)

	// do the expire set
	todayEndTimestamp, _ := tools.GetTodayTimestampByLocalTime("23:59:59")
	if isReloanTag == int(types.IsReloanYes) {
		onceReloanExpire.Do(func() {
			logs.Debug("[pushToRiskList] set expire: %s", key)
			redisConn.Do("EXPIREAT", key, todayEndTimestamp)
		})
	} else {
		onceFirstLoanExpire.Do(func() {
			logs.Debug("[pushToRiskList] set expire: %s", key)
			redisConn.Do("EXPIREAT", key, todayEndTimestamp)
		})
	}
}

func getScoreListName(isReloanTag int) string {
	keyPrefix := beego.AppConfig.String("rm_case_score_sets_prefix")
	if len(keyPrefix) == 0 {
		logs.Error("[case.pushToRiskList] redis config miss rm_case_score_sorted_set_prefix")
		panic("required config miss")
	}
	return fmt.Sprintf(keyPrefix, isReloanTag, tools.GetToday())
}

// no pop for
func getOneRandomRiskOrderFromSets(isReloanTag int) (orderID int64, isEmpty bool, err error) {
	redisConn := storage.RedisStorageClient.Get()
	defer redisConn.Close()
	key := getScoreListName(isReloanTag)
	//ZRANGE sorted_test 0 0 WITHSCORES
	reply, redisErr := redis.Int64(redisConn.Do("SPOP", key))
	if redisErr != nil && redisErr != redis.ErrNil {
		err = fmt.Errorf("[getOneHighestRiskOrderFromList] redis err: %v", redisErr)
		return
	}
	if redisErr == redis.ErrNil {
		// redis
		isEmpty = true
		return
	}
	orderID = reply
	//score, _ := strconv.Atoi(reply[1])
	//logs.Debug("[getOneHighestRiskOrderFromList] get from sorted set reloan tag: %d,  rank:%d,orderID: %d, score: %d", isReloanTag, index, orderID, score)
	return
}

func getRandomOrders(isReloanTag types.IsReloanEnum, quota int) (orderIDs []int64, left int) {
	i := 0
	for ; i < quota; i++ {
		orderID, isEmpty, _ := getOneRandomRiskOrderFromSets(int(isReloanTag))
		if isEmpty {
			break
		}
		orderIDs = append(orderIDs, orderID)
	}
	left = quota - i
	return
}
