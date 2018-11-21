package task

import (
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type AutoReduceOrderTask struct {
}

// 处理逾期订单 {{{
func (c *AutoReduceOrderTask) Start() {
	logs.Info("[AutoReduceOrderTask] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("auto_reduce_order_lock")
	qName := beego.AppConfig.String("auto_reduce")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", 24*60*60, "NX")
	if err != nil || lock == nil {
		logs.Error("[AutoReduceOrderTask] process is working, so, I will exit.")
		// ***! // 很重要!
		close(done)
		return
	}
	defer storageClient.Do("DEL", lockKey)

	lastedOrderId := int64(0)
	for {
		if cancelled() {
			logs.Info("[AutoReduceOrderTask] receive exit cmd.")
			break
		}

		TaskHeartBeat(storageClient, lockKey)

		// 生产队列,小批量处理
		qVal, err := storageClient.Do("LLEN", qName)
		if err == nil && qVal != nil && 0 == qVal.(int64) {
			logs.Info("[AutoReduceOrderTask] %s 队列为空,开始按条件生成.", qName)

			// 队列是空,需要生成了
			// 1. 取数据
			orderList, _ := service.OrderList4ReviewAutoReduce(lastedOrderId)

			logs.Info("len(orderList):%d lastedOrderId:%d", len(orderList), lastedOrderId)
			// 如果没有满足条件的数据,work goroutine 也不用启动了
			if len(orderList) == 0 {
				time.Sleep(500 * time.Millisecond)
				logs.Info("[AuotReduceOrders] 生产待审核自动减免订单队列没有满足条件的数据,退出.")
				break
			}
			lastedOrderId = orderList[len(orderList)-1].Id

			// 2. 加队列
			for _, order := range orderList {
				storageClient.Do("LPUSH", qName, order.Id)
			}
		}

		if err != nil {
			logs.Error("[AutoReduceOrderTask] err :%v qVal:%v", err, qVal)
		}
		// 消费队列
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			logs.Error("lets go:%d", i)
			go consumeNeedAutoReduceOrderQueue(&wg, i, qName)
		}

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()
	}

	logs.Info("[AuotReduceOrders] politeness exit.")
}

func (c *AutoReduceOrderTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("overdue_order_lock")
	storageClient.Do("DEL", lockKey)
}

func consumeNeedAutoReduceOrderQueue(wg *sync.WaitGroup, workerID int, qName string) {
	defer wg.Done()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	for {
		if cancelled() {
			logs.Info("[consumeNeedAutoReduceOrderQueue] receive exit cmd, workID:", workerID)
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Warn("[consumeNeedAutoReduceOrderQueue] no data for consume, I will exit after 500ms, workID :%d qName:%s  err:%v qValueByte:%v", workerID, qName, err, qValueByte)
			time.Sleep(500 * time.Millisecond)
			break
		}

		queueValToCmd, _ := tools.Str2Int64(string(qValueByte.([]byte)))
		if queueValToCmd == types.TaskExitCmd {
			logs.Info("[consumeNeedAutoReduceOrderQueue] receive exit cmd, I will exit after jobs done. workID:", workerID, ", queueVal:", queueValToCmd)
			// ***! // 很重要!
			close(done)
			break
		}
		// 真正开始工作了
		str := string(qValueByte.([]byte))
		orderId, _ := tools.Str2Int64(str)
		addCurrentData(str, "data")
		handleAutoReduceOrder(workerID, orderId)
		removeCurrentData(str)

	}
}

func handleAutoReduceOrder(workerID int, orderId int64) {
	logs.Info("[handleAutoReduceOrder] workerID:%d orderId:%d", workerID, orderId)
	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleAutoReduceOrder] panic orderId:%d, workId:%d, err:%v", orderId, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	service.CheckAndDoAutoReduce(orderId)
}
