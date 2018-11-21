package task

import (
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type TicketRealtimeAssignTask struct {
	wg sync.WaitGroup
}

// Start 处理事件任务
func (c *TicketRealtimeAssignTask) Start() {
	logs.Info("[TicketRealtimeAssignTask] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("ticket_realtime_assign_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")

	if err != nil || lock == nil {
		logs.Error("[TicketRealtimeAssignTask] process is working, so, I will exit.")
		// ***! // 很重要!
		close(done)
		return
	}

	// 消费队列
	//var wg sync.WaitGroup
	for _, ticketItem := range ticket.RealtimeAssignTicketItems() {
		c.wg.Add(1)
		go autoAssign(&c.wg, cancelled, ticketItem, int(ticketItem))
	}
	//
	c.watch()

	// 主 goroutine,等待工作 goroutine 正常结束
	c.wg.Wait()

	// -1 正常退出时,释放锁
	c.Cancel()
}

func (c *TicketRealtimeAssignTask) watch() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		storageClient := storage.RedisStorageClient.Get()
		defer storageClient.Close()
		qtaskCommandKey := beego.AppConfig.String("ticket_realtime_assign_task_command")
		for {
			if cancelled() {
				logs.Info("[TicketRealtimeAssignTask] watch goroutine receive exit cmd.")
				break
			}

			lockKey := beego.AppConfig.String("ticket_realtime_assign_lock")
			TaskHeartBeat(storageClient, lockKey)

			// 获取任务命令
			command, _ := redis.Int64(storageClient.Do("RPOP", qtaskCommandKey))

			if command == types.TaskExitCmd {
				logs.Info("[TicketRealtimeAssignTask] catch receive exit cmd, will broadcase exit signal")
				close(done)
				// ***! // 很重要!
				return
			}

			time.Sleep(time.Second)

		}
	}()
}

func (c *TicketRealtimeAssignTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("ticket_realtime_assign_lock")
	storageClient.Do("DEL", lockKey)
}

func autoAssign(wg *sync.WaitGroup, isCanceled func() bool, ticketItem types.TicketItemEnum, workerID int) {
	defer wg.Done()
	logs.Debug("[autoAssign]start auto assign ticket item:", ticketItem)
	// 实时工单队列池
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()

	key := ticket.GetAssignQueueKey(ticketItem)
	s := ticket.GetItemWorkerStrategy(ticketItem)
	logs.Debug("[autoAssign]assign queue:", key)

	for {

		// check before [start] and before[sleep]
		if isCanceled() {
			logs.Info("[TicketRealtimeAssignTask] working task receive exit cmd.")
			return
		}
		// 实时分配的先决条件
		// 1. 有人力 2.有待分配ticket
		// 先判断人力, 还是先用 llen 判断ticket,
		// 目前, 工单数较少, 所以先判断待分配ticket
		waitCount, _ := redis.Int(redisCli.Do("LLEN", key))
		logs.Debug("[autoAssign]wait count:", waitCount)
		var adminUID, ticketID int64
		if waitCount > 0 {
			// 有待分配ticket
			adminUID, _ = s.OneWorker()
			if adminUID > 0 {
				// 有人力
				var popErr error

				ticketID, popErr = redis.Int64(redisCli.Do("RPOP", key))
				waitCount--
				if ticketID > 0 {
					ticket.AutoAssignByTicketID(ticketID, adminUID)
				} else {
					logs.Error("[AutoAssign] Get Ticket from queue err:", popErr, ";ticketID:", ticketID)
				}
			}
		}
		// single sub task is completed, check isCanceled,
		// best position to check,
		if isCanceled() {
			logs.Info("[TicketRealtimeAssignTask] working task receive exit cmd.")
			return
		}

		if waitCount < 1 || adminUID == 0 {
			// 无可用人力或者无更多待处理订单, 则休眠 1s, 等待新工单产生或者人力释放
			time.Sleep(1000 * time.Millisecond)
		} else {
			// 人力充沛, 有未分配订单, 快速处理
			// 后期可缩减间隔以应对, 批量自动分配
			time.Sleep(10 * time.Millisecond)
		}
	}

}
