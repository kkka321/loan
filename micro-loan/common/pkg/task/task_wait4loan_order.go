package task

import (
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/pkg/schema_task"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type Wait4LoanOrderTask struct {
}

// 处理等待放款订单 {{{
func (c *Wait4LoanOrderTask) Start() {
	logs.Info("[Wait4LoanOrderTask] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁 {{{
	lockKey := beego.AppConfig.String("wait4loan_order_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[Wait4LoanOrderTask] process is working, so, I will exit.")
		// ***! // 很重要!
		close(done)
		return
	}

	qName := beego.AppConfig.String("wait4loan_order")
	for {
		if cancelled() {
			logs.Info("[Wait4LoanOrderTask] receive exit cmd.")
			break
		}

		TaskHeartBeat(storageClient, lockKey)

		// 1. 创建任务队列
		logs.Info("[Wait4LoanOrderTask] produceWait4LoanOrderQueue")
		qValueByte, err := storageClient.Do("LLEN", qName)
		logs.Debug("qValueByte:", qValueByte, ", err:", err)
		if err != nil {
			logs.Error("[Wait4LoanOrderTask] dependency service of redis does not work, err: %s", err.Error())
			break
		} else if qValueByte != nil && 0 == qValueByte.(int64) {
			//// 0. 非工作时间,不需要工作
			timetagStart := tools.NaturalDay(0) / 1000
			timetagEnd := timetagStart + int64(15.5*60*60)
			timetagNow := tools.TimeNow()
			if timetagNow > timetagEnd {
				logs.Info("[Wait4LoanOrderTask] skip unhandle time start:%d, end:%d, now:%d", timetagStart, timetagEnd, timetagNow)
				time.Sleep(15 * time.Second)
				continue
			}

			//// 队列是空,需要生成了
			//// 1. 取数据
			orderList, _ := service.OrderListByStatus(types.LoanStatusWait4Loan)
			if len(orderList) == 0 {
				logs.Info("[Wait4LoanOrderTask] no match data, sleep and retry.")
				time.Sleep(time.Second)
				continue
			}

			//// 2. 加队列
			for _, orderOne := range orderList {
				//// 写队列
				storageClient.Do("LPUSH", qName, orderOne.Id)
			}
		}

		// 2. 消费队列
		logs.Info("[Wait4LoanOrderTask] consume queue")
		var wg sync.WaitGroup
		// 可视情况加工作 goroutine 数,一期只开2个
		// 目前增加到4个协程
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go consumeWait4LoanOrderQueue(&wg, i)
		}

		// 3. 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()
	}

	// -1 正常退出时,释放锁 }}}
	storageClient.Do("DEL", lockKey)
	logs.Info("[Wait4LoanOrderTask] politeness exit.")
}

func (c *Wait4LoanOrderTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("wait4loan_order_lock")
	storageClient.Do("DEL", lockKey)
}

// 消费等待放款队列
func consumeWait4LoanOrderQueue(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	logs.Info("It will do consumeWait4LoanOrderQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	qName := beego.AppConfig.String("wait4loan_order")
	for {
		if cancelled() {
			logs.Info("[consumeWait4LoanOrderQueue] receive exit cmd, workID:", workerID)
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据
		if err != nil || qValueByte == nil {
			logs.Info("[consumeWait4LoanOrderQueue] no data for consume, exit work goroutine, workID: %d, err: %v", workerID, err)
			break
		}

		orderID, _ := tools.Str2Int64(string(qValueByte.([]byte)))
		if orderID == types.TaskExitCmd {
			logs.Info("[consumeWait4LoanOrderQueue] receive exit cmd, I will exit after jobs done. workID:", workerID, ", orderID:", orderID)
			// ***! // 很重要!
			close(done)
			break
		}

		// 真正开始工作了
		addCurrentData(tools.Int642Str(orderID), "orderId")
		handleWait4LoanOrder(orderID, workerID)
		removeCurrentData(tools.Int642Str(orderID))

		//绑定订单ID到最新一次活体
		service.LastLiveVerifyBindOrderID(orderID)
	}
}

func handleWait4LoanOrder(orderID int64, workerID int) {
	logs.Info("[handleWait4LoanOrder] orderID:", orderID, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleWait4LoanOrder] panic orderId:%d, workId:%d, err:%v", orderID, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	orderData, err := models.GetOrder(orderID)
	if err != nil || orderData.CheckStatus != types.LoanStatusWait4Loan {
		orderDataJSON, _ := tools.JsonEncode(orderData)
		logs.Error("[handleWait4LoanOrder] 订单状态不正确,请检查. orderDataJSON:", orderDataJSON, ", workerID", workerID, ", err:", err)
		return
	}

	// 调用放款模块
	invokeId, err := service.CreateDisburse(orderID)
	invoke, _ := models.OneDisburseInvorkLogByPkId(invokeId)

	if invokeId == 0 {
		logs.Warn("[handleWait4LoanOrder] 放款前检查失败. CreateDisburse field before call api. orderID:%d err:%v", orderID, err)
	} else if invokeId != 0 && invoke.DisbureStatus == types.DisbureStatusCallFailed && err != nil {
		// 未生成放款id 或 生成了id 以及第三方放款id 并返回了err就是真的放款失败了。
		schema_task.PushBusinessMsg(types.PushTargetLoanFail, orderData.UserAccountId)
		logs.Error("[handleWait4LoanOrder] 放款失败,请检查 disburse failed. orderID:%d, workerID:%d, err:%v invoke:%#v", orderID, workerID, err, invoke)

		// 将订单置为失败
		tag := tools.GetUnixMillis()
		orderData, _ = models.GetOrder(orderID)
		if orderData.CheckStatus == types.LoanStatusIsDoing {
			originOrder := orderData
			orderData.CheckStatus = types.LoanStatusLoanFail
			orderData.CheckTime = tag
			orderData.Utime = tag
			models.UpdateOrder(&orderData)
			monitor.IncrOrderCount(orderData.CheckStatus)
			// 写操作日志
			models.OpLogWrite(0, orderData.Id, models.OpCodeOrderUpdate, orderData.TableName(), originOrder, orderData)

			service.TryAddModifyBankTag(orderData)
		} else if orderData.CheckStatus == types.LoanStatusLoanFail {
			service.TryAddModifyBankTag(orderData)
		}

	} else if invokeId != 0 && invoke.DisbureStatus == types.DisbureStatusCallUnknow && err != nil {
		// 第三方调用id 为空可能是超时
		logs.Error("[handleWait4LoanOrder] 放款有可能因超时而失败,请检查. disburse failed, maybe due to delay. orderID:%d, workerID:%d, err:%v invoke:%#v", orderID, workerID, err, invoke)
	} else if err != nil {
		logs.Warn("[handleWait4LoanOrder] 未知错误,请检查. disburse unknown error, please check orderID:%d, workerID:%d, err:%v invoke:%#v", orderID, workerID, err, invoke)
	}

}

// }}}
