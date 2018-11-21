package task

import (
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/event/runner"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type EventPushTask struct {
}

// TaskHandleEventPush 处理事件任务
func (c *EventPushTask) Start() {
	logs.Info("[EventPush] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("event_queue_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")

	if err != nil || lock == nil {
		logs.Error("[EventPush] process is working, so, I will exit.")
		// ***! // 很重要!
		close(done)
		return
	}

	go func() {
		for {
			storageClientHeart := storage.RedisStorageClient.Get()
			TaskHeartBeat(storageClientHeart, lockKey)
			storageClientHeart.Close()
			time.Sleep(5 * time.Minute)
		}
	}()

	for {
		if cancelled() {
			logs.Info("[EventPush] receive exit cmd.")
			break
		}

		var wg sync.WaitGroup

		// 消费队列
		// 统一任务
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go consumeEventQueue(&wg, i)
		}

		// 拆分任务
		queues := evtypes.SeparateQueueMap()
		for sepEv := range queues {
			//maxGoroutine := runner.QueueConsumerGoroutineConfig(q, 4)
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go consumeSeperateEventQueue(&wg, i, sepEv)
			}
		}

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)

	logs.Info("[EventPush] politeness exit.")
}

func (c *EventPushTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("event_queue_lock")
	storageClient.Do("DEL", lockKey)
}

// 消费逾期订单队列
func consumeEventQueue(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	logs.Info("It will do consumeOverdueOrderQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	qName := beego.AppConfig.String("event_queue")
	for {
		if cancelled() {
			logs.Info("[EventPush] receive exit cmd, workID:", workerID)
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Info("[EventPush] no data for consume, I will exit after 500ms, workID:", workerID)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		queueValToCmd, _ := tools.Str2Int64(string(qValueByte.([]byte)))
		if queueValToCmd == types.TaskExitCmd {
			logs.Info("[EventPush] receive exit cmd, I will exit after jobs done. workID:", workerID, ", queueVal:", queueValToCmd)
			// ***! // 很重要!
			close(done)
			break
		}
		// 真正开始工作了
		str := string(qValueByte.([]byte))
		addCurrentData(str, "data")
		//EventParser.Run(qValueByte.([]byte))
		runner.ParseAndRun(qValueByte.([]byte))
		removeCurrentData(str)

	}
}

// 消费逾期订单队列
func consumeSeperateEventQueue(wg *sync.WaitGroup, workerID int, eventName string) {
	defer wg.Done()
	logs.Info("It will do consumeOverdueOrderQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	qName := evtypes.GetQueueName(eventName)

	//qName := beego.AppConfig.String("event_queue")
	for {
		if cancelled() {
			logs.Info("[EventPush] receive exit cmd, workID:", workerID)
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Info("[EventPush] no data for consume, I will exit after 500ms, workID:", workerID)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		queueValToCmd, _ := tools.Str2Int64(string(qValueByte.([]byte)))
		if queueValToCmd == types.TaskExitCmd {
			logs.Info("[EventPush] receive exit cmd, I will exit after jobs done. workID:", workerID, ", queueVal:", queueValToCmd)
			// ***! // 很重要!
			close(done)
			break
		}
		// 真正开始工作了
		str := string(qValueByte.([]byte))
		addCurrentData(str, "data")
		//EventParser.Run(qValueByte.([]byte))
		runner.ParseAndRunSeperate(qValueByte.([]byte), eventName)
		removeCurrentData(str)
	}
}
