package task

import (
	"fmt"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"

	"micro-loan/common/dao"
	"micro-loan/common/i18n"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type RepayRemindOrderTask struct {
}

// 还款订单提醒 {{{
func (c *RepayRemindOrderTask) Start() {
	logs.Info("[TaskHandleRepayRemindOrder] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("repay_remind_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[TaskHandleRepayRemindOrder] process is working, so, I will exit.")
		// ***! // 很重要!
		close(done)
		return
	}

	for {
		if cancelled() {
			logs.Info("[TaskHandleRepayRemindOrder] receive exit cmd.")
			break
		}

		TaskHeartBeat(storageClient, lockKey)

		setsName := beego.AppConfig.String("repay_remind_sets")
		todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))
		yesterdaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(-1)))

		num, _ := storageClient.Do("EXISTS", yesterdaySetName)
		if num != nil && num.(int64) == 1 {
			//如果存在就干掉
			storageClient.Do("DEL", yesterdaySetName)
		}

		qVal, err := storageClient.Do("EXISTS", todaySetName)

		// 初始化去重集合
		if err == nil && 0 == qVal.(int64) {
			storageClient.Do("SADD", todaySetName, 1)
			/*
				t := time.Now()
				now := t.Unix() * 1000
				tomorrow := tools.NaturalDay(1)
				diff := math.Ceil(float64(tomorrow-now) / 1000)
				storageClient.Do("EXPIRE", setsName, diff) // 有效期到今天晚上23:59:59
			*/
		}

		// 生产队列,小批量处理
		qName := beego.AppConfig.String("repay_remind_queue")
		qVal, err = storageClient.Do("LLEN", qName)
		if err == nil && qVal != nil && 0 == qVal.(int64) {
			logs.Info("[TaskHandleRepayRemindOrder] %s 队列为空,开始按条件生成.", qName)

			var idsBox []string
			setsMem, err := redis.Values(storageClient.Do("SMEMBERS", todaySetName))
			if err != nil || setsMem == nil {
				logs.Error("[TaskHandleRepayRemindOrder] 生产还款提醒订单队列无法从集合中取到元素,休眠1秒后将重试.")
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			for _, m := range setsMem {
				idsBox = append(idsBox, string(m.([]byte)))
			}
			// 理论上不会出现
			if len(idsBox) == 0 {
				logs.Error("[TaskHandleRepayRemindOrder] 生产还款提醒订单队列出错了,集合中没有元素,不符合预期,程序将退出.")
				//! 很重要,确定程序正常退出
				close(done)
				break
			}

			orderList, _ := service.GetRepayRemindOrderList(idsBox)

			// 如果没有满足条件的数据,work goroutine 也不用启动了
			if len(orderList) == 0 {
				logs.Info("[TaskHandleRepayRemindOrder] 生产还款提醒订单队列没有满足条件的数据,休眠1秒后将重试.")
				time.Sleep(1000 * time.Millisecond)
				continue
			}

			for _, orderId := range orderList {
				storageClient.Do("LPUSH", qName, orderId)
			}
		}

		// 消费队列
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go consumeRepayRemindOrderQueue(&wg, i)
		}

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()

		//这个为crontab用，只跑一次
		break
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[TaskHandleRepayRemindOrder] politeness exit.")
}

func (c *RepayRemindOrderTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("repay_remind_lock")
	storageClient.Do("DEL", lockKey)
}

// 消费还款提醒订单队列
func consumeRepayRemindOrderQueue(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	logs.Info("It will do consumeRepayRemindOrderQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	qName := beego.AppConfig.String("repay_remind_queue")
	for {
		if cancelled() {
			logs.Info("[consumeRepayRemindOrderQueue] receive exit cmd, workID:", workerID)
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Info("[consumeRepayRemindOrderQueue] no data for consume, I will exit after 500ms, workID:", workerID)
			time.Sleep(500 * time.Millisecond)
			break
		}

		orderID, _ := tools.Str2Int64(string(qValueByte.([]byte)))
		if orderID == types.TaskExitCmd {
			logs.Info("[consumeRepayRemindOrderQueue] receive exit cmd, I will exit after jobs done. workID:", workerID, ", orderID:", orderID)
			// ***! // 很重要!
			close(done)
			break
		}

		addCurrentData(tools.Int642Str(orderID), "orderId")
		handleRepayRemindOrder(orderID, workerID)
		removeCurrentData(tools.Int642Str(orderID))
	}
}

func handleRepayRemindOrder(orderID int64, workerID int) {
	logs.Info("[handleRepayRemindOrder] orderID:", orderID, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleRepayRemindOrder] panic orderId:%d, workId:%d, err:%v", orderID, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	setsName := beego.AppConfig.String("repay_remind_sets")
	todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))
	qVal, err := storageClient.Do("SADD", todaySetName, orderID)
	// 说明有错,或已经处理过,忽略本次操作
	if err != nil || 0 == qVal.(int64) {
		logs.Info("[handleRepayRemindOrder] 此订单已经处理过,忽略之. orderID: %d, workerID: %d", orderID, workerID)
		return
	}

	order, orderErr := models.GetOrder(orderID)
	repayPlan, rpErr := models.GetLastRepayPlanByOrderid(orderID)
	_, abErr := dao.CustomerOne(order.UserAccountId)
	if orderErr != nil || rpErr != nil || abErr != nil {
		logs.Error("[handleRepayRemindOrder] can not find should exist data:", orderID,
			"orderData:", orderErr, "repayPlan:", rpErr, "accountBase:", abErr)
		return
	}

	date := tools.MDateMHSDate(repayPlan.RepayDate)
	repayMoney, _ := repayplan.CaculateRepayTotalAmountByRepayPlan(repayPlan)
	logs.Debug("[handleRepayRemindOrder] Order ID: %d, caculate repayMoney: %d", orderID, repayMoney)
	if repayMoney > 0 {

		//获取VA账户信息
		userEAccount, eAccountErr := dao.GetActiveEaccountWithBankName(order.UserAccountId) // models.GetActiveEAccount(order.UserAccountId, 1)
		if eAccountErr != nil {
			logs.Error("[handleRepayRemindOrder] GetActiveEAccount happend err:", eAccountErr)
			return
		}
		//xenditCallback := models.GetXenditCallBack(userEAccount.CallbackJson)
		logs.Debug("[handleRepayRemindOrder] userEAccount:", userEAccount, " xendCallbackJSON:", userEAccount.CallbackJson)
		vaAccountNumber := userEAccount.BankCode + " " + userEAccount.EAccountNumber

		//repayMoney := (repayPlan.Amount - repayPlan.AmountPayed) + (repayPlan.GracePeriodInterest - repayPlan.GracePeriodInterestPayed)
		smsContent := fmt.Sprintf(i18n.GetMessageText(i18n.TextRepayRemind), date, repayMoney, vaAccountNumber)

		//smsContent = fmt.Sprintf(smsContent, date, repayMoney)

		// 新发送短信逻辑
		logs.Debug("[handleRepayRemindOrder] Order ID: %d, start send sms: %s", orderID, smsContent)
		//sms.Send(types.ServiceRepayRemind, accountBase.Mobile, smsContent, orderID)
		// 创建人工还款提醒case
		logs.Debug("[handleRepayRemindOrder] Order ID: %d, start judge whether create repayremind case or not", orderID)

		//
		//repayremind.Handle
	}
}

// }}}
