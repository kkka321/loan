package task

import (
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type InvalidOrderTask struct {
}

// 处理无效订单 {{{
func (c *InvalidOrderTask) Start() {
	logs.Info("[TaskHandleInvalidOrder] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("invalid_order_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[produceInvalidOrderQueue] process is working, so, I will exit.")
		// ***! // 很重要!
		close(done)
		return
	}

	qName := beego.AppConfig.String("invalid_order")
	for {
		if cancelled() {
			logs.Info("[TaskHandleInvalidOrder] receive exit cmd.")
			break
		}

		TaskHeartBeat(storageClient, lockKey)

		// 1. 生产队列
		logs.Info("[TaskHandleInvalidOrder] produceWait4LoanOrderQueue")
		qValueByte, err := storageClient.Do("LLEN", qName)
		logs.Debug("qValueByte:", qValueByte, ", err:", err)
		if err == nil && qValueByte != nil && 0 == qValueByte.(int64) {
			// 队列是空,需要生成了
			// 1. 取数据
			orderList, _ := service.InvalidOrderList()
			if len(orderList) == 0 {
				logs.Info("[TaskHandleInvalidOrder] produceWait4LoanOrderQueue 没有满足条件的数据,可以退出了.")
				time.Sleep(500 * time.Millisecond)
				break
			}

			// 2. 加队列
			for _, orderOne := range orderList {
				// 写队列
				storageClient.Do("LPUSH", qName, orderOne.Id)
			}
		}

		// 2. 消费队列
		logs.Info("[TaskHandleInvalidOrder] consume queue")
		var wg sync.WaitGroup
		// 可视情况加工作 goroutine 数,一期只开2个
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go consumeInvalidOrderQueue(&wg, i)
		}

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[TaskHandleInvalidOrder] politeness exit.")
}

func (c *InvalidOrderTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("invalid_order_lock")
	storageClient.Do("DEL", lockKey)
}

// 消费无效订单队列
func consumeInvalidOrderQueue(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	logs.Info("It will do consumeInvalidOrderQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	qName := beego.AppConfig.String("invalid_order")
	for {
		if cancelled() {
			logs.Info("[consumeInvalidOrderQueue] receive exit cmd, workID:", workerID)
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据
		if err != nil || qValueByte == nil {
			logs.Info("[consumeInvalidOrderQueue] no data for consume, I will exit work goroutine, workID:", workerID)
			break
		}

		orderID, _ := tools.Str2Int64(string(qValueByte.([]byte)))
		if orderID == types.TaskExitCmd {
			logs.Info("[consumeInvalidOrderQueue] receive exit cmd, I will exit after jobs done. workID:", workerID, ", orderID:", orderID)
			// ***! // 很重要!
			close(done)
			break
		}

		// 真正开始工作了
		addCurrentData(tools.Int642Str(orderID), "orderId")
		handleInvalidOrder(orderID, workerID)
		removeCurrentData(tools.Int642Str(orderID))
	}
}

func handleInvalidOrder(orderID int64, workerID int) {
	logs.Info("[handleInvalidOrder] orderID:", orderID, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleInvalidOrder] panic orderId:%d, workId:%d, err:%v", orderID, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	orderData, err := models.GetOrder(orderID)
	orderDataJSON, _ := tools.JsonEncode(orderData)
	if err != nil || orderData.CheckStatus != types.LoanStatusSubmit {
		logs.Error("[handleInvalidOrder] 订单状态不正确,请检查. orderDataJSON:", orderDataJSON, ", workerID", workerID, ", err:", err)

		return
	}

	originOrder := orderData

	orderData.CheckStatus = types.LoanStatusInvalid
	orderData.CheckTime = tools.GetUnixMillis()
	models.UpdateOrder(&orderData)

	// 订单失效事件触发
	event.Trigger(&evtypes.OrderInvalidEv{
		OrderID:   orderID,
		AccountID: orderData.UserAccountId,
		Time:      tools.GetUnixMillis(),
	})

	// 写操作日志
	models.OpLogWrite(0, orderData.Id, models.OpCodeOrderUpdate, orderData.TableName(), originOrder, orderData)
}

// }}}
