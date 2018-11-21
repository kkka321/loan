package ticket

import (
	"context"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/types"
	"strconv"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"
)

// RealtimeAssign 实时分配
type RealtimeAssignBak struct {
}

// AutoAssign 自动分配
// 每一种工单要启动一个实时分配 goroutine, 因为每种工单的人力池不同
// 可能会因人力不足出现, 任务排队情况, 不能因为其他类别工单阻塞而阻塞
func (a *RealtimeAssignBak) AutoAssign(ctx context.Context, ticketItem types.TicketItemEnum, s WorkerAssignStrategy) {
	// 实时工单队列池
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()

	key := getAssignQueueKey(ticketItem)

	for {
		// 实时分配的先决条件
		// 1. 有人力 2.有待分配ticket
		// 先判断人力, 还是先用 llen 判断ticket,
		// 目前, 工单数较少, 所以先判断待分配ticket
		waitCount, _ := redis.Int(redisCli.Do("LLEN", key))
		var adminUID int64
		if waitCount > 0 {
			// 有待分配ticket
			adminUID, _ = s.OneWorker()
			if adminUID > 0 {
				// 有人力
				ticketID, popErr := redis.Int64(redisCli.Do("RPOP", key))
				waitCount--
				if ticketID > 0 {
					AutoAssignByTicketID(ticketID, adminUID)
				} else {
					logs.Error("[AutoAssign] Get Ticket from queue err:", popErr, ";ticketID:", ticketID)
				}
			}
		}

		// 处理关闭信号
		// 平滑关闭
		select {
		case <-ctx.Done():
			return
		default:
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

func enterWaitAssignQueue(id int64, ticketItem types.TicketItemEnum, push string) bool {
	key := getAssignQueueKey(ticketItem)
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()
	if id <= 0 {
		logs.Error("[enterWaitAssignQueue] unexpected ticket id:", id)
		return false
	}
	redisCli.Do(push, key, id)
	return true
}

func removeFromWaitAssignQueue(id int64, ticketItem types.TicketItemEnum) bool {
	key := getAssignQueueKey(ticketItem)
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()
	if id <= 0 {
		logs.Error("[removeFromWaitAssignQueue] unexpected ticket id:", id)
		return false
	}
	_, err := redisCli.Do("LREM", key, 0, id)
	if err != nil {
		logs.Error("[removeFromWaitAssignQueue] remove operation redis err:", err)
		return false
	}
	return true
}

func getAssignQueueKey(ticketItem types.TicketItemEnum) string {
	prefix := beego.AppConfig.String("ticket_wait_assign_queue_prefix")
	if len(prefix) == 0 {
		logs.Error("[getAssignQueueKey] ticket_wait_assign_queue_prefix is not configured")
	}
	return prefix + strconv.Itoa(int(ticketItem))
}

// GetAssignQueueKey 根据ticket item 获取待分配ticket队列
func GetAssignQueueKey(ticketItem types.TicketItemEnum) string {
	return getAssignQueueKey(ticketItem)
}

// AssignAfterDayFirstOnline 首次登录触发自动分单
func AssignAfterDayFirstOnline(adminUID int64, roleID int64) {
	// get all can be assigned ticketitems
	ticketItems := getWorkTicketItemsByRoleID(roleID)
	if len(ticketItems) == 0 {
		return
	}

	for _, ticketItem := range ticketItems {
		// 判断ticket是否需要, 日均分配
		doDailyAvgAssign(ticketItem, adminUID)
	}
}

func doDailyAvgAssign(ticketItem types.TicketItemEnum, adminUID int64) {
	// 判断ticket是否需要, 日均分配
	if !isDailyAvg(ticketItem) {
		return
	}

	// get day ticket total by ctime
	todayWillAssignTotalNum := models.GetTodayTotalWillAssignNumByItem(ticketItem)

	// 若没有待处理工单, 则直接退出
	if todayWillAssignTotalNum == 0 {
		logs.Warn("[doDailyAvgAssign] no ticket for ticketItem:", ticketItem)
		return
	}

	key := getAssignQueueKey(ticketItem)

	var workAdminNum int64
	// 获取非休假状态的员工数目
	{
		admins, _, _ := canAssignUsersByTicketItem(ticketItem)
		for _, admin := range admins {
			if admin.WorkStatus == types.AdminWorkStatusNormal {
				workAdminNum++
			}
		}
	}

	var shouldAssign int64
	if todayWillAssignTotalNum > workAdminNum {
		shouldAssign = todayWillAssignTotalNum / workAdminNum
	} else {
		// 若工单数小于人数, 则确保此人可以获得分配工单的资格
		shouldAssign = 1
	}
	// 查询已分配 , 保证幂等
	todayAlreadyAssignedHim := models.GetTodayTotalAlreadyAssignedNumByItemAssignUID(ticketItem, adminUID)
	shouldAssign = shouldAssign - todayAlreadyAssignedHim

	// 分配给 adminUID
	logs.Debug("FinalshouldAssign:", shouldAssign, "toadyWillAssignTotalNum:", todayWillAssignTotalNum,
		"todayAlreadyAssignedHim", todayAlreadyAssignedHim, "workAdminNum:", workAdminNum)
	// 获取当天, 指定类型, shouldAssign个工单

	if shouldAssign == 0 {
		logs.Warn("[doDailyAvgAssign] admin user(%d) ticket item(%d) assign shouldn't 0", adminUID, ticketItem, "FinalshouldAssign:", shouldAssign, "toadyWillAssignTotalNum:", todayWillAssignTotalNum,
			"todayAlreadyAssignedHim", todayAlreadyAssignedHim, "workAdminNum:", workAdminNum)
		return
	}

	// 有待分配ticket
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()

	waitNum, _ := redis.Int64(redisCli.Do("LLEN", key))

	if waitNum == 0 && shouldAssign != 0 {
		logs.Warn("[doDailyAvgAssign] ticket item(%d) should not be 0 , when daily assign for admin:%d", ticketItem, adminUID)
		return
	}

	var actualWantNum int64
	if waitNum > shouldAssign {
		actualWantNum = shouldAssign
	} else {
		actualWantNum = waitNum
	}

	var actualAssignSucc int64
	for actualAssignSucc < actualWantNum {
		// 此处,防止无限循环, 假定每次弹出都是成功的, 且数据不存在异常
		// 队列长度总大于等于shouldAssign
		ticketID, popErr := redis.Int64(redisCli.Do("RPOP", key))
		if popErr != nil {
			logs.Warn("[doDailyAvgAssign] Get Ticket from queue err:", popErr, ";ticketID:", ticketID)
			// 报错则说明， 队列不存在， 或者其他错误
			break
		}
		if ticketID > 0 {
			assignResult := AutoAssignByTicketID(ticketID, adminUID)
			if assignResult {
				actualAssignSucc++
			}
		}
	}

	if actualWantNum != actualAssignSucc {
		logs.Warn("[doDailyAvgAssign] daily assign for admin(%d), item id: %d, actualWantNum:%d, actualAssignSucc:%d", adminUID, ticketItem, actualWantNum, actualAssignSucc)
	}
}

// DailyFinalAssign 单日组中分配
func DailyFinalAssign(wg *sync.WaitGroup, ticketItem types.TicketItemEnum) {
	defer wg.Done()

	logs.Debug("[DailyFinalAssign] start, ticket item:", ticketItem)
	key := getAssignQueueKey(ticketItem)
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()

	waitCount, _ := redis.Int(redisCli.Do("LLEN", key))
	logs.Debug("waitCount", waitCount)

	var preparedWorker int64
	if waitCount > 0 {
		workerStrategy := getItemWorkerStrategy(ticketItem)
		for waitCount > 0 {
			waitCount--
			// 此处,防止无限循环, 假定每次弹出都是成功的, 且数据不存在异常
			// 队列长度总大于等于 waitCount
			if preparedWorker == 0 {
				var workerErr error
				preparedWorker, workerErr = workerStrategy.OneWorker()
				if workerErr != nil || preparedWorker == 0 {
					logs.Error("[DailyFinalAssign] worker err, ticket item %d, err: %v, must be manual run in time, it is emergency",
						ticketItem, workerErr)
					return
				}
			}

			ticketID, popErr := redis.Int64(redisCli.Do("RPOP", key))
			if ticketID > 0 {
				assignResult := AutoAssignByTicketID(ticketID, preparedWorker)
				if assignResult {
					// 分配成功则, 置0, 获取下一个人力
					preparedWorker = 0
				}
			} else {
				logs.Error("[AutoAssign] Get Ticket from queue err:", popErr, ";ticketID:", ticketID)
			}
		}
	}
}

// CheckEarlyWorkerDailyAssignByItemForRM 针对早于日工单生成的员工, 做一个check, 和补分逻辑
func CheckEarlyWorkerDailyAssignByItemForRM() {
	// get all earlier active worker before daily tickets create
	ticketItems := []types.TicketItemEnum{
		types.TicketItemRMAdvance1,
		types.TicketItemRM0,
		types.TicketItemRM1,
	}
	for _, ticketItem := range ticketItems {
		makeUpDailyAssignForEarlyBirdsByTicketItem(ticketItem)
	}
}

// CheckEarlyWorkerDailyAssignForUrge 针对早于日工单生成的员工, 做一个check, 和补分逻辑
func CheckEarlyWorkerDailyAssignForUrge() {
	// get all earlier active worker before daily tickets create
	urgeTicketItems := []types.TicketItemEnum{
		types.TicketItemUrgeM11,
		types.TicketItemUrgeM12,
		types.TicketItemUrgeM13,
		types.TicketItemUrgeM20,
		types.TicketItemUrgeM30,
	}
	for _, ticketItem := range urgeTicketItems {
		makeUpDailyAssignForEarlyBirdsByTicketItem(ticketItem)
	}
}

func makeUpDailyAssignForEarlyBirdsByTicketItem(ticketItem types.TicketItemEnum) {
	logs.Debug("[makeUpDailyAssignForEarlyBirdsByTicketItem] started check, whether have early birds:")
	admins := activeCanAssignUsersByTicketItem(ticketItem)
	logs.Debug("[makeUpDailyAssignForEarlyBirdsByTicketItem] early birds:", admins)
	for _, admin := range admins {
		logs.Debug("[makeUpDailyAssignForEarlyBirdsByTicketItem] doDailyAvgAssign for early bird:", admin.Id, admin.Nickname)
		doDailyAvgAssign(ticketItem, admin.Id)
	}
}
