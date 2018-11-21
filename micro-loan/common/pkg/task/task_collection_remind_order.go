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
	"micro-loan/common/i18n"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type CollectionRemindOrderTask struct {
}

var CollectionRemindMsg = map[types.CollectionRemindDay]string{
	types.CollectionRemindDef:   i18n.GetMessageText(i18n.TextCollectionRemindDef),
	types.CollectionRemindTwo:   i18n.GetMessageText(i18n.TextCollectionRemindTwo),
	types.CollectionRemindFour:  i18n.GetMessageText(i18n.TextCollectionRemindFour),
	types.CollectionRemindEight: i18n.GetMessageText(i18n.TextCollectionRemindEight),
}

// 催收短信提醒
func (c *CollectionRemindOrderTask) Start() {
	logs.Info("[TaskHandleCollectionRemindOrder] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("collection_remind_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[TaskHandleCollectionRemindOrder] process is working, so, I will exit.")
		close(done)
		return
	}

	// 获取催收配置信息
	collectionRemindDayInt, collectionRemindMsg := getCollectionRemindConf()
	logs.Info("[TaskHandleCollectionRemindOrder] collectionRemindDay:", collectionRemindDayInt,
		", collectionRemindMsg:", collectionRemindMsg)
	if len(collectionRemindDayInt) <= 0 {
		logs.Error("[TaskHandleCollectionRemindOrder] get collection remind data failed.")
		close(done)
		return
	}

	for {
		if cancelled() {
			logs.Info("[TaskHandleCollectionRemindOrder] receive exit cmd.")
			break
		}

		TaskHeartBeat(storageClient, lockKey)

		setsName := beego.AppConfig.String("collection_remind_sets")
		todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))
		yesterdaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(-1)))

		num, _ := storageClient.Do("EXISTS", yesterdaySetName)
		if num != nil && num.(int64) == 1 {
			storageClient.Do("DEL", yesterdaySetName)
		}

		qVal, err := storageClient.Do("EXISTS", todaySetName)
		// 初始化去重集合
		if err == nil && 0 == qVal.(int64) {
			storageClient.Do("SADD", todaySetName, 1)
		}

		// 生产队列,小批量处理
		qName := beego.AppConfig.String("collection_remind_queue")
		qVal, err = storageClient.Do("LLEN", qName)
		if err == nil && qVal != nil && 0 == qVal.(int64) {
			logs.Info("[TaskHandleCollectionRemindOrder] %s 队列为空,开始按条件生成.", qName)

			var idsBox []string
			setsMem, err := redis.Values(storageClient.Do("SMEMBERS", todaySetName))
			if err != nil || setsMem == nil {
				logs.Error("[TaskHandleCollectionRemindOrder] 生产催收短信提醒订单队列无法从集合中取到元素,休眠1秒后将重试.")
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			for _, m := range setsMem {
				idsBox = append(idsBox, string(m.([]byte)))
			}
			// 理论上不会出现
			if len(idsBox) == 0 {
				logs.Error("[TaskHandleCollectionRemindOrder] 生产催收短信提醒订单队列出错了,集合中没有元素,不符合预期,程序将退出.")
				//! 很重要,确定程序正常退出
				close(done)
				break
			}

			// 获取订单列表
			orderList, _ := service.GetCollectionRemindOrderList(idsBox, collectionRemindDayInt)

			// 如果没有满足条件的数据,work goroutine 也不用启动了
			if len(orderList) == 0 {
				logs.Info("[TaskHandleCollectionRemindOrder] 生产催收短信提醒订单队列没有满足条件的数据,任务结束.")
				//这个为crontab用，只跑一次
				break
			}

			for _, orderId := range orderList {
				storageClient.Do("LPUSH", qName, orderId)
			}
		}

		// 消费队列
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go consumeCollectionRemindOrderQueue(&wg, i, collectionRemindMsg)
		}

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[TaskHandleCollectionRemindOrder] politeness exit.")
}

// 获取催收配置信息，如果未配置按默认
func getCollectionRemindConf() (collectionRemindDayInt []types.CollectionRemindDay, collectionRemindMsg map[types.CollectionRemindDay]string) {

	collectionRemindDayConf := config.ValidItemString("collection_remind_day")
	logs.Info("[getCollectionRemindConf] collectionRemindDayConf: %s", collectionRemindDayConf)
	if len(collectionRemindDayConf) <= 0 {
		// 按照默认的逾期天数催收通知
		collectionRemindMsg = CollectionRemindMsg
		collectionRemindDayInt = []types.CollectionRemindDay{types.CollectionRemindTwo, types.CollectionRemindFour,
			types.CollectionRemindEight}
	} else {
		collectionRemindMsg = make(map[types.CollectionRemindDay]string)
		collectionRemindDayStr := strings.Split(collectionRemindDayConf, ",")
		for _, val := range collectionRemindDayStr {
			val = strings.TrimSpace(val)
			v, _ := strconv.Atoi(val)
			collectionRemindDayInt = append(collectionRemindDayInt, types.CollectionRemindDay(v))

			// 获取催收天数的对应参数配置
			msgConf := config.ValidItemString(fmt.Sprintf("%s%s", "collection_remind_", val))
			msgDef := CollectionRemindMsg[types.CollectionRemindDay(v)]
			if len(msgConf) <= 0 && len(msgDef) <= 0 {
				collectionRemindMsg[types.CollectionRemindDay(v)] = CollectionRemindMsg[types.CollectionRemindDef]
			} else if len(msgConf) <= 0 {
				collectionRemindMsg[types.CollectionRemindDay(v)] = msgDef
			} else {
				collectionRemindMsg[types.CollectionRemindDay(v)] = msgConf
			}
			logs.Info("[getCollectionRemindConf] collectionRemindMsg[%d]: %s", v, collectionRemindMsg[types.CollectionRemindDay(v)])
		}
	}

	return
}

func (c *CollectionRemindOrderTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("collection_remind_lock")
	storageClient.Do("DEL", lockKey)
}

// 消费催收短信提醒队列
func consumeCollectionRemindOrderQueue(wg *sync.WaitGroup, workerID int, collectionRemindMsg map[types.CollectionRemindDay]string) {
	defer wg.Done()
	logs.Info("It will do consumeCollectionRemindOrderQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	qName := beego.AppConfig.String("collection_remind_queue")
	for {
		if cancelled() {
			logs.Info("[consumeCollectionRemindOrderQueue] receive exit cmd, workID:", workerID)
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Info("[consumeCollectionRemindOrderQueue] no data for consume, I will exit after 500ms, workID:", workerID)
			time.Sleep(500 * time.Millisecond)
			break
		}

		orderID, _ := tools.Str2Int64(string(qValueByte.([]byte)))
		if orderID == types.TaskExitCmd {
			logs.Info("[consumeCollectionRemindOrderQueue] receive exit cmd, I will exit after jobs done. workID:",
				workerID, ", orderID:", orderID)
			close(done)
			break
		}

		addCurrentData(tools.Int642Str(orderID), "orderId")
		handleCollectionRemindOrder(orderID, workerID, collectionRemindMsg)
		removeCurrentData(tools.Int642Str(orderID))
	}
}

func handleCollectionRemindOrder(orderID int64, workerID int, collectionRemindMsg map[types.CollectionRemindDay]string) {
	logs.Info("[handleCollectionRemindOrder] orderID:", orderID, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleCollectionRemindOrder] painc orderId:%d, workId:%d, err:%v", orderID, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	setsName := beego.AppConfig.String("collection_remind_sets")
	todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))
	qVal, err := storageClient.Do("SADD", todaySetName, orderID)
	// 说明有错,或已经处理过,忽略本次操作
	if err != nil || 0 == qVal.(int64) {
		logs.Warning("[handleCollectionRemindOrder] 此订单已经处理过, 忽略之. orderID: %d, workerID: %d", orderID, workerID)
		return
	}

	order, _ := models.GetOrder(orderID)
	dao.CustomerOne(order.UserAccountId)

	overdueDay, repayMoney, bankCode, userEAccount, err := getCollectionRemindData(orderID)
	if err != nil {
		logs.Warning("[handleCollectionRemindOrder] 不满足催收条件. orderID: %d, workerID: %d", orderID, workerID)
		return
	}
	if repayMoney <= 0 {
		logs.Warning("[handleCollectionRemindOrder] 还款金额有误. orderID: %d, workerID: %d", orderID, workerID)
		return
	}

	if len(userEAccount.EAccountNumber) <= 0 {
		logs.Warning("[handleCollectionRemindOrder] 获取电子虚拟账户失败, 电子虚拟账户为空. orderID: %d, workerID: %d",
			orderID, workerID)
	}

	smsContent := collectionRemindMsg[types.CollectionRemindDay(overdueDay)]
	if len(smsContent) <= 0 {
		logs.Warning("[handleCollectionRemindOrder] 不满足催收条件, 逾期天数有误. orderID: %d, workerID: %d", orderID, workerID)
		return
	}

	smsContent = fmt.Sprintf(smsContent, repayMoney, bankCode, userEAccount.EAccountNumber)
	logs.Info("[handleCollectionRemindOrder] 催收短信内容:", smsContent)

	// 催收短信发送逻辑
	//sms.Send(types.ServiceCollectionRemind, accountBase.Mobile, smsContent, orderID)
}

// 获取催收短信需要的数据
func getCollectionRemindData(orderID int64) (overdueDay, repayMoney int64, bankCode string, userEAccount models.User_E_Account, err error) {

	orderData, err := models.GetOrder(orderID)
	orderDataJSON, _ := tools.JsonEncode(orderData)
	if err != nil {
		logs.Error("[getCollectionRemindData] 订单数据有误, orderData: %s, err: %v", orderDataJSON, err)
		return
	}

	repayPlan, err := models.GetLastRepayPlanByOrderid(orderID)
	repayPlanJSON, _ := tools.JsonEncode(repayPlan)
	if err != nil || repayPlan.RepayDate <= 0 {
		logs.Error("[getCollectionRemindData] 还款计划数据有误, orderID: %d, repayPlan: %s, err: %v", orderID, repayPlanJSON, err)
		return
	}

	overdueDay, err = service.CalculateOverdue(repayPlan.RepayDate)
	if err != nil {
		logs.Warning("[getCollectionRemindData] 不满足催收条件, orderData:", orderDataJSON,
			", repayPlan:", repayPlanJSON,
			", overdueDays:", overdueDay,
			", err:", err)
		return
	}

	repayMoney = (repayPlan.Amount - repayPlan.AmountPayed) + (repayPlan.GracePeriodInterest - repayPlan.GracePeriodInterestPayed) +
		(repayPlan.Penalty - repayPlan.PenaltyPayed)

	userEAccount, err = dao.GetActiveEaccountWithBankName(orderData.UserAccountId)
	//userEAccount, err = models.GetEAccount(orderData.UserAccountId, types.Xendit)
	if err != nil {
		logs.Error("[getCollectionRemindData] 获取电子虚拟账户失败, accountID: %d, err: %v",
			orderData.UserAccountId, err)
		return
	}
	/*
		xenditCallback := models.GetXenditCallBack(userEAccount.CallbackJson)
		logs.Debug("[getCollectionRemindData] xenditCallback struct :", xenditCallback, " xendCallbackJSON:", userEAccount.CallbackJson)
		bankCode = xenditCallback.BankCode
	*/
	bankCode = userEAccount.BankCode

	return
}
