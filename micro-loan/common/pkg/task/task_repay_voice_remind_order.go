package task

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"

	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/nxtele"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type RepayVoiceOrderTask struct {
}

func getRepayVoiceRemindSetName(i int) (setsName string) {
	setsName = beego.AppConfig.String("repay_voice_remind_sets")
	setsName = fmt.Sprintf("%s:%s:%d", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)), i)

	return
}

// 对 RM-1 和 RM0 的订单，语音提醒
func (c *RepayVoiceOrderTask) Start() {
	logs.Info("[TaskHandleRepayVoiceOrder] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("repay_voice_remind_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[TaskHandleRepayVoiceOrder] process is working, so, I will exit.")
		close(done)
		return
	}

	var findNum int // 记录查询符合条件的订单的次数，当次数超过两次后，退出
	for {
		if cancelled() {
			logs.Info("[TaskHandleRepayVoiceOrder] receive exit cmd.")
			break
		}

		TaskHeartBeat(storageClient, lockKey)

		// 生产队列, 小批量处理
		var i int
		qName := beego.AppConfig.String("repay_voice_remind_queue")
		// 对于还款日期是（昨天/今天/明天）的分别存储到不同的queue，分别对应 i（-1/0/1）
		// 20181029的一个需求，只提醒还款日期是当天/明天的
		for i = 0; i < 2; i++ {

			setName := getRepayVoiceRemindSetName(i)
			sVal, err := storageClient.Do("EXISTS", setName)
			// 初始化去重集合
			if err == nil && 0 == sVal.(int64) {
				storageClient.Do("SADD", setName, 1)
				storageClient.Do("EXPIRE", setName, tools.SECONDAHOUR/3)
			}

			qRealName := fmt.Sprintf("%s_%d", qName, i)
			qVal, err := storageClient.Do("LLEN", qRealName)
			if err == nil && qVal != nil && 0 == qVal.(int64) {
				logs.Info("[TaskHandleRepayVoiceOrder] %s 队列为空,开始按条件生成.", qRealName)

				var idsBox []string
				setsMem, err := redis.Values(storageClient.Do("SMEMBERS", setName))
				if err != nil || setsMem == nil {
					logs.Error("[TaskHandleRepayVoiceOrder] 生产还款语音提醒订单队列无法从集合中取到元素,休眠1秒后将重试.")
					time.Sleep(1000 * time.Millisecond)
					continue
				}
				for _, m := range setsMem {
					idsBox = append(idsBox, string(m.([]byte)))
				}
				// 理论上不会出现
				if len(idsBox) == 0 {
					logs.Error("[TaskHandleRepayVoiceOrder] 生产还款语音提醒订单队列出错了,集合中没有元素,不符合预期,程序将退出.")
					close(done)
					break
				}

				orderList, _ := service.GetRepayVoiceRemindOrderList(idsBox, i)

				// 如果没有满足条件的数据, work goroutine 也不用启动了
				if len(orderList) == 0 {
					findNum++
					if findNum >= 2 {
						logs.Info("[TaskHandleRepayVoiceOrder] 生产还款语音提醒订单队列没有满足条件的数据,程序将退出.")
						break
					}
					logs.Info("[TaskHandleRepayVoiceOrder] 生产还款语音提醒订单队列没有满足条件的数据,休眠1秒后将重试.")
					time.Sleep(1000 * time.Millisecond)
					continue

				}

				for _, orderId := range orderList {
					storageClient.Do("LPUSH", qRealName, orderId)
				}
			}
		}

		// 消费队列
		var wg sync.WaitGroup
		// 对于还款日期是（昨天/今天/明天）的分别存储到不同的queue，分别对应 i（-1/0/1）
		// 20181029的一个需求，只提醒还款日期是当天/明天的
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go consumeRepayVoiceOrderQueue(&wg, i)
		}

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()

		//这个为crontab用，只跑一次
		break
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[TaskHandleRepayVoiceOrder] politeness exit.")
}

func (c *RepayVoiceOrderTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("repay_voice_remind_lock")
	storageClient.Do("DEL", lockKey)
}

// 消费还款/逾期语音订单队列
func consumeRepayVoiceOrderQueue(wg *sync.WaitGroup, workerID int) {

	defer wg.Done()
	logs.Info("It will do consumeRepayVoiceOrderQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	var orderIDArr []int64
	var endFlag int
	var orderID int64

	qName := beego.AppConfig.String("repay_voice_remind_queue")
	qRealName := fmt.Sprintf("%s_%d", qName, workerID)

	postMobileNum, _ := config.ValidItemInt("nxtele_post_mobile_num")
	if postMobileNum <= 0 {
		postMobileNum = 5000
	}
	logs.Info("[consumeRepayVoiceOrderQueue] postMobileNum:", postMobileNum, "workID:", workerID)

	if cancelled() {
		logs.Info("[consumeRepayVoiceOrderQueue] receive exit cmd, workID:", workerID)
		return
	}

	for {
		qValueByte, err := storageClient.Do("RPOP", qRealName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Info("[consumeRepayVoiceOrderQueue] no data for consume, I will exit after 500ms, workID:", workerID)
			time.Sleep(500 * time.Millisecond)
			endFlag = 1
			goto next
		}

		orderID, _ = tools.Str2Int64(string(qValueByte.([]byte)))
		if orderID == types.TaskExitCmd {
			logs.Info("[consumeRepayVoiceOrderQueue] receive exit cmd, I will exit after jobs done. workID:", workerID, ", orderID:", orderID)
			close(done)
			break
		}
		orderIDArr = append(orderIDArr, orderID)

	next:
		if len(orderIDArr) == postMobileNum || endFlag == 1 {
			logs.Info("[consumeRepayVoiceOrderQueue] endFlag:", endFlag, "orderIDArr:", orderIDArr, "workID:", workerID)
			if len(orderIDArr) > 0 {
				orderIDArrTmp := orderIDArr
				addCurrentData(tools.Int642Str(orderIDArrTmp[0]), "orderId")
				handleRepayVoiceOrder(orderIDArrTmp, int(workerID))
				removeCurrentData(tools.Int642Str(orderIDArrTmp[0]))

				for 0 < len(orderIDArr) {
					orderIDArr = append(orderIDArr[:0], orderIDArr[1:]...)
				}
			} else {
				break
			}
		}
	}
}

func handleRepayVoiceOrder(orderIDArr []int64, workerID int) {
	logs.Info("[handleRepayVoiceOrder] orderIDArr:", orderIDArr, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleRepayVoiceOrder] panic orderId:%d, workId:%d, err:%v", orderIDArr, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	setName := getRepayVoiceRemindSetName(workerID)

	var repayPlan models.RepayPlan
	var accountBase models.AccountBase
	var mobileArr []string
	for _, orderID := range orderIDArr {

		sVal, err := storageClient.Do("SADD", setName, orderID)
		// 说明有错,或已经处理过,忽略本次操作
		if err != nil || 0 == sVal.(int64) {
			logs.Info("[handleRepayVoiceOrder] 此订单已经处理过,忽略之. orderID: %d, workerID: %d", orderIDArr, workerID)
			continue
		}

		order, _ := models.GetOrder(orderID)
		accountBase, _ = dao.CustomerOne(order.UserAccountId)
		repayPlan, _ = models.GetLastRepayPlanByOrderid(orderID)
		if len(accountBase.Mobile) > 0 {
			mobileArr = append(mobileArr, accountBase.Mobile)
		}
	}

	mobile := strings.Join(mobileArr, ",")
	voiceType := GetVoiceType(repayPlan)
	// 新发送语音提醒逻辑
	if len(mobile) > 0 {
		nxteleResp, err := nxtele.Send(voiceType, mobile)
		if err != nil || nxteleResp == nil {
			logs.Error("[handleRepayVoiceOrder] Send voice call request or parse response occur error, orderId:%d, workId:%d, err:%v",
				orderIDArr, workerID, err)
			return
		}
		isSuccess := nxteleResp.IsSuccess()
		sid := nxteleResp.GetSID()
		logs.Info("[handleRepayVoiceOrder] Send voice call response, isSuccess:%d, SID:%d, orderId:%d, workId:%d",
			isSuccess, sid, orderIDArr, workerID)

		if isSuccess == 1 && sid > 0 {
			SaveSidStatus(sid, workerID)
		}
	}
}

func GetVoiceType(repayPlan models.RepayPlan) (voiceType types.VoiceType) {

	yesterday := tools.NaturalDay(-1) // 还款日期是前一天，逾期一天提醒
	today := tools.NaturalDay(0)      // 还款日期是当天，当天提醒
	tomorrow := tools.NaturalDay(1)   // 还款日期是后一天，提前一天提醒
	switch repayPlan.RepayDate {
	case yesterday:
		voiceType = types.VoiceTypeYesterday
	case today:
		voiceType = types.VoiceTypeToday
	case tomorrow:
		voiceType = types.VoiceTypeTomorrow
	}

	return
}

// 保存呼叫结果
func SaveSidStatus(sid int64, workerID int) {

	var status *nxtele.NxteleCallStatusResp
	var err error

	for {
		// 查询呼叫结果
		status, err = nxtele.GetSidStatus(sid)
		if err != nil || status == nil {
			logs.Error("[SaveSidStatus] Get sid status failed, sid:", sid, "workerID:", workerID)
			return
		}

		if status.Result == types.VoiceCallSuccess {
			logs.Info("[SaveSidStatus] Get sid status success, sid:", sid, "workerID:", workerID)
			break
		}

		time.Sleep(60 * time.Second)
	}

	logs.Info("[SaveSidStatus] Get sid status, data:", status.Data)
	for k, v := range status.Data {
		logs.Info("[SaveSidStatus] data[%d]: mobile-%s, duration-%s, fee-%s, workerID-%d", k, v.Phone, v.Duration, v.Fee, workerID)

		d, _ := strconv.Atoi(v.Duration)
		f, _ := strconv.ParseFloat(v.Fee, 32)
		voiceRemind := models.VoiceRemind{
			Sid:      sid,
			Mobile:   v.Phone,
			Duration: d,
			Status:   workerID,
			Fee:      f,
		}
		voiceRemind.Add()
	}

	logs.Info("[SaveSidStatus] Add voice remind end")
}
