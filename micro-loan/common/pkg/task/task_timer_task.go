package task

import (
	"flag"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/toolbox"
	"github.com/gomodule/redigo/redis"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/repayremind"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/pkg/ticket/performance"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

var runOnce bool
var runName string
var runOnceParam string
var TimerWg sync.WaitGroup

func init() {
	flag.BoolVar(&runOnce, "run-once", false, "run one task")
	flag.StringVar(&runName, "run-name", "", "task name")
	flag.StringVar(&runOnceParam, "run-once-param", "", "custom param for run once, like stats day, not required")
}

type TimerTask struct {
}

type timerTaskInfo struct {
	Time string
	Func func() error
}

// six columns mean：
//       second：0-59
//       minute：0-59
//       hour：1-23
//       day：1-31
//       month：1-12
//       week：0-6（0 means Sunday）
var workName map[string]timerTaskInfo = map[string]timerTaskInfo{
	"thirdpartyFee":           timerTaskInfo{"0 10 17 * * *", ThirdpartyFeeOutFromCache},
	"businessDetailStatistic": timerTaskInfo{"0 30 17 * * *", BusinessDetailStatistic}, //印尼时间 0：30开始统计
	"roll": timerTaskInfo{"0 0 17 * * *", RunRollTask},
	"daily_ticket_final_assign":         timerTaskInfo{"0 0 2 * * *", handleDailyUnassignedTicket},
	"worker_ticket_daily_stats":         timerTaskInfo{"0 10 17 * * *", workerTicketDailyStats},
	"worker_ticket_daily_process_stats": timerTaskInfo{"0 2 * * * *", workerTicketDailyProcessStats},
	//go run cli-task/task.go --name=timer_task --run-once=true --run-name=worker_ticket_hourly_process_stats --run-once-param=2018101502
	"worker_ticket_hourly_stats": timerTaskInfo{"1 */5 * * * *", workerTicketHourlyStats},
	//"worker_ticket_realtime_stats": timerTaskInfo{"1 */5 * * * *", workerTicketRealtimeStats},
	"repay_remind_case_task": timerTaskInfo{"0 6 0 * * *", repayRemindCaseTask},
	//"ticket_month_stats_daily_update": timerTaskInfo{"0 20 8 * * *", ticketMonthStatsDailyUpdate},
	"coupon_expire":        timerTaskInfo{"0 0 18 * * *", RunCouponExpireTask},
	"account_coupon_reuse": timerTaskInfo{"0 1 17 * * *", RunAccountCouponReuseTask},
	//go run cli-task/task.go --name=timer_task --run-once=true --run-name=batch_apply_entrust
	"batch_apply_entrust": timerTaskInfo{"0 30 16 * * *", batchApplyEntrustTask}, //批量申请委外 印尼时间23：30
}

func (c *TimerTask) Start() {
	if runOnce {
		v, ok := workName[runName]
		if !ok {
			logs.Info("[TimerTask] run name wrong %s", runName)
			return
		}

		v.Func()
	} else {
		storageClient := storage.RedisStorageClient.Get()
		defer storageClient.Close()

		lockKey := beego.AppConfig.String("timer_task_lock")
		lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")

		if err != nil || lock == nil {
			logs.Error("[TimerTask] process is working, so, I will exit.")
			// ***! // 很重要!
			close(done)
			return
		}

		for k, v := range workName {
			tk := toolbox.NewTask(k, v.Time, v.Func)
			toolbox.AddTask(k, tk)
		}
		toolbox.StartTask()

		qName := beego.AppConfig.String("timer_task")

		for {
			if cancelled() {
				logs.Info("[TimerTask] receive exit cmd.")
				break
			}

			TaskHeartBeat(storageClient, lockKey)

			qValueByte, err := storageClient.Do("RPOP", qName)
			// 没有可供消费的数据,退出工作 goroutine
			if err != nil || qValueByte == nil {
				time.Sleep(time.Second)
				continue
			}

			id, _ := tools.Str2Int64(string(qValueByte.([]byte)))
			if id == types.TaskExitCmd {
				logs.Info("[TimerTask] receive exit cmd")
				close(done)
				// ***! // 很重要!
				continue
			}

			time.Sleep(time.Second)
		}

		TimerWg.Wait()

		storageClient.Do("DEL", lockKey)
		logs.Info("[TimerTask] politeness exit.")
	}
}

func (c *TimerTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("timer_task_lock")
	storageClient.Do("DEL", lockKey)
}

func handleDailyUnassignedTicket() error {
	TimerWg.Add(1)
	defer TimerWg.Done()

	ticketItems := ticket.DailyAvgAssignTicketItems()
	var wg sync.WaitGroup

	for _, ticketItem := range ticketItems {
		//copyTicketItem := ticketItem
		if ticketItem == types.TicketItemRM0 {
			// RM0 配额不做二次分单，弃掉不分
			continue
		}
		wg.Add(1)
		go ticket.DailyFinalAssign(&wg, ticketItem)
	}
	wg.Wait()
	return nil
}

func workerTicketDailyStats() error {
	TimerWg.Add(1)
	defer TimerWg.Done()

	// runOnceParam should be date like 20181003
	if runOnceParam != "" {
		performance.DailyWorkerPerformanceStatsTask(runOnceParam)
	} else {
		performance.LastDayDailyWorkerPerformanceStatsTask()
	}
	return nil
}

func workerTicketDailyProcessStats() error {
	TimerWg.Add(1)
	defer TimerWg.Done()

	// runOnceParam should be date like 20181003
	if runOnceParam != "" {
		performance.DailyWorkerProcessHistoryStatsByDay(runOnceParam)
	} else {
		performance.DailyWorkerProcessStatsLastHour()
	}
	return nil
}

func workerTicketHourlyStats() error {
	TimerWg.Add(1)
	defer TimerWg.Done()

	// runOnceParam should be date like 2018100312
	logs.Notice("[workerTicketHourlyStats] runOnceParam", runOnceParam)
	if runOnceParam != "" {
		performance.HourlyWorkerPerformanceStatsTask(runOnceParam)
	} else {
		performance.TodayHourlyWorkerPerformanceStatsTask()
	}
	return nil
}
func batchApplyEntrustTask() error {
	TimerWg.Add(1)
	defer TimerWg.Done()

	entrustDay, _ := config.ValidItemInt("outsource_day")
	overdueIDs := models.GetOverdueCaseIDs(entrustDay)
	logs.Debug("[overdueIDs]", overdueIDs, "num", len(overdueIDs))
	if len(overdueIDs) > 0 {
		ticketIDs := models.GetTicketIDByRelatedIDS(overdueIDs)
		logs.Debug("[ticketIDs]", ticketIDs)
		if len(ticketIDs) > 0 {
			success := ticket.BatchApplyEntrust(ticketIDs)
			logs.Notice("[batchApplyEntrustTask] success:", success, "total:", len(ticketIDs))
		}
	}

	return nil
}

func workerTicketRealtimeStats() error {
	TimerWg.Add(1)
	defer TimerWg.Done()

	// runOnceParam should be date like 2018100312
	logs.Notice("[workerTicketHourlyStats] runOnceParam", runOnceParam)
	if runOnceParam != "" {
		timeTag, err := strconv.ParseInt(runOnceParam, 10, 64)
		if err != nil {
			logs.Error(err)
			return err
		}
		performance.RealtimeWorkerPerformanceStatsTask(timeTag)
	} else {
		performance.TodayRealtimeWorkerPerformanceStatsTask()
	}
	return nil
}

func ticketMonthStatsDailyUpdate() error {
	TimerWg.Add(1)
	defer TimerWg.Done()

	performance.UpdateCurrentMonthStats()
	return nil
}

func ThirdpartyFeeOutFromCache() error {
	logs.Info("[TimerTask] ThirdpartyFeeOutFromCache run")
	thirdparty.MoveOutThirdpartyStatisticFeeFromCache()
	return nil
}

func BusinessDetailStatistic() error {
	logs.Info("[TimerTask] BusinessDetailStatistic run")

	// 统计昨天数据
	thirdparty.BusinessDetailStatistic(tools.GetUnixMillis())

	// 休息5分钟，留给从库 同步的时间
	time.Sleep(time.Second * 30)

	// 生成今天的临时记录
	thirdparty.BusinessDetailStatistic(tools.GetUnixMillis() + tools.MILLSSECONDADAY)
	return nil
}

func repayRemindCaseTask() error {
	TimerWg.Add(1)
	defer TimerWg.Done()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("repay_remind_case_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[OverdueMessageNotify] process is working, so, I will exit.")
		return nil
	}
	defer storageClient.Do("DEL", lockKey)

	setsName := beego.AppConfig.String("repay_remind_case_sets")
	todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))
	yesterdaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(-1)))

	num, _ := storageClient.Do("EXISTS", yesterdaySetName)
	if num != nil && num.(int64) == 1 {
		//如果存在就干掉
		storageClient.Do("DEL", yesterdaySetName)
	}

	qVal, err := storageClient.Do("EXISTS", todaySetName)

	if err == nil && 0 == qVal.(int64) {
		storageClient.Do("SADD", todaySetName, 1)
	}

	qName := beego.AppConfig.String("repay_remind_case_queue")
	qVal, err = storageClient.Do("LLEN", qName)
	if err == nil && qVal != nil && 0 == qVal.(int64) {
		logs.Info("[TaskHandleRepayRemindOrder] %s 队列为空,开始按条件生成.", qName)

		var idsBox []string
		setsMem, err := redis.Values(storageClient.Do("SMEMBERS", todaySetName))
		if err != nil || setsMem == nil {
			logs.Error("[TaskHandleRepayRemindOrder] 生产还款提醒订单队列无法从集合中取到元素,休眠1秒后将重试.")
			time.Sleep(1000 * time.Millisecond)
			return nil
		}
		for _, m := range setsMem {
			idsBox = append(idsBox, string(m.([]byte)))
		}
		// 理论上不会出现
		if len(idsBox) == 0 {
			logs.Error("[TaskHandleRepayRemindOrder] 生产还款提醒订单队列出错了,集合中没有元素,不符合预期,程序将退出.")
			//! 很重要,确定程序正常退出
			return nil
		}

		orderList, _ := service.GetRepayRemindCaseOrderList(idsBox)

		// 如果没有满足条件的数据,work goroutine 也不用启动了
		if len(orderList) == 0 {
			return nil
		}

		for _, orderID := range orderList {
			storageClient.Do("LPUSH", qName, orderID)
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		// 此处为预处理
		go consumeRepayRemindCaseOrderQueue(&wg, i)
	}
	wg.Wait()
	// 使用
	// TODO 防panic处理
	repayremind.FilterAndCreateCases()

	// calculateUnpaidAmount()

	ticket.CheckEarlyWorkerDailyAssignByItemForRM()
	return nil
}

// 消费还款提醒订单队列
func consumeRepayRemindCaseOrderQueue(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	logs.Info("It will do consumeRepayRemindOrderQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	qName := beego.AppConfig.String("repay_remind_case_queue")
	for {
		orderID, err := redis.Int64(storageClient.Do("RPOP", qName))
		if err != nil {
			logs.Warn("[consumeRepayRemindCaseOrderQueue] redis err:", err)
			break
		}
		if orderID == 0 {
			break
		}
		repayremind.PreHandle(orderID)
	}
}

/*
func calculateUnpaidAmount() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// 统计生成工单应还总金额
	// rm0 case的应还金额 unpaidCaseAmount
	// rm0的总的应还金额 unpaidPrincipal
	unpaidCaseAmount, unpaidAmount, _ := service.GetRepayRemindUnpaidAmount()
	unpaidCaseAmountKey := tools.GetUnpaidAmountKey("unpaid_case_amount", tools.GetUnixMillis())
	_, err := storageClient.Do("SET", unpaidCaseAmountKey, unpaidCaseAmount, "EX", tools.SECONDADAY*5, "NX")
	if err != nil {
		logs.Error("[repayRemindCaseTask] unpaid_case_amount set fail.")
	}
	unpaidAmountKey := tools.GetUnpaidAmountKey("unpaid_amount", tools.GetUnixMillis())
	_, err = storageClient.Do("SET", unpaidAmountKey, unpaidAmount, "EX", tools.SECONDADAY*5, "NX")
	if err != nil {
		logs.Error("[repayRemindCaseTask] unpaid_amount set fail.")
	}

	return
}
*/
