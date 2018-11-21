package task

import (
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type OverdueOrderTask struct {
}

// 处理逾期订单 {{{
func (c *OverdueOrderTask) Start() {
	logs.Info("[TaskHandleOverdueOrder] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("overdue_order_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[TaskHandleOverdueOrder] process is working, so, I will exit.")
		// ***! // 很重要!
		close(done)
		return
	}

	for {
		if cancelled() {
			logs.Info("[TaskHandleOverdueOrder] receive exit cmd.")
			break
		}

		TaskHeartBeat(storageClient, lockKey)

		// 生产队列,小批量处理
		qName := beego.AppConfig.String("overdue_order")
		qVal, err := storageClient.Do("LLEN", qName)
		if err == nil && qVal != nil && 0 == qVal.(int64) {
			logs.Info("[TaskHandleOverdueOrder] %s 队列为空,开始按条件生成.", qName)

			timetag := tools.NaturalDay(0)
			orderIDs, _ := service.GetOverdueOrderIDList(timetag, 1000)

			// 如果没有满足条件的数据,work goroutine 也不用启动了
			if len(orderIDs) == 0 {
				logs.Info("[TaskHandleOverdueOrder] 生产逾期订单队列没有满足条件的数据,任务完成.")
				break
			}

			for _, orderID := range orderIDs {
				storageClient.Do("LPUSH", qName, orderID)
			}
		}

		// 消费队列
		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go consumeOverdueOrderQueue(&wg, i)
		}

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()
		// 每日一次, 执行完成之后, 触发对早来的员工分单
		logs.Debug("overdue trigger Early worker dailly assign")
	}
	ticket.CheckEarlyWorkerDailyAssignForUrge()

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[TaskHandleOverdueOrder] politeness exit.")
}

func (c *OverdueOrderTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("overdue_order_lock")
	storageClient.Do("DEL", lockKey)
}

// 消费逾期订单队列
func consumeOverdueOrderQueue(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	logs.Info("It will do consumeOverdueOrderQueue, workerID:", workerID)

	qName := beego.AppConfig.String("overdue_order")
	for {
		if cancelled() {
			logs.Info("[consumeOverdueOrderQueue] receive exit cmd, workID:", workerID)
			break
		}

		storageClient := storage.RedisStorageClient.Get()
		qValueByte, err := storageClient.Do("RPOP", qName)
		storageClient.Close()
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Info("[consumeOverdueOrderQueue] no data for consume, I will exit after 500ms, workID:", workerID)
			time.Sleep(500 * time.Millisecond)
			break
		}

		orderID, _ := tools.Str2Int64(string(qValueByte.([]byte)))
		if orderID == types.TaskExitCmd {
			logs.Info("[consumeOverdueOrderQueue] receive exit cmd, I will exit after jobs done. workID:", workerID, ", orderID:", orderID)
			// ***! // 很重要!
			close(done)
			break
		}

		// 真正开始工作了
		addCurrentData(tools.Int642Str(orderID), "orderId")
		handleOverdueOrder(orderID, workerID)
		removeCurrentData(tools.Int642Str(orderID))
	}
}

func handleOverdueOrder(orderID int64, workerID int) {
	logs.Info("[handleOverdueOrder] orderID:", orderID, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleOverdueOrder] panic orderId:%d, workId:%d, err:%v", orderID, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	// 将订单状态在合适的时机扭转为[逾期]
	orderData, _ := models.GetOrder(orderID)
	repayPlan, _ := models.GetLastRepayPlanByOrderid(orderID)
	product, _ := models.GetProduct(orderData.ProductId)
	//logs.Debug("[handleOverdueOrder] orderID: %d, CheckStatus: %d, RepayDate: %d, offset: %d, today: %d", orderID, orderData.CheckStatus, repayPlan.RepayDate, repayPlan.RepayDate  + 3600 * 1000 * 24, tools.NaturalDay(0))
	/**
	例子: 7号是应还日,8号是宽限期,9号记逾期
	如果宽限期内用户无操作,订单状态不会发生改变,即宽限期不会将订单状态标记为逾期.也就是说,9号才会将订单状态改为逾期
	宽限期内记宽限期利息,不记罚息;逾期后开始记罚息
	*/

	if orderData.CheckStatus != types.LoanStatusOverdue &&
		orderData.CheckStatus != types.LoanStatusAlreadyCleared &&
		orderData.CheckStatus != types.LoanStatusRollClear &&
		repayPlan.RepayDate > 0 &&
		repayPlan.RepayDate+3600*1000*24*int64(product.GracePeriod) < tools.NaturalDay(0) {

		origin := orderData

		orderData.CheckStatus = types.LoanStatusOverdue
		// 冗余字段标记订单逾期,因为其状态值可能会变为已结束
		orderData.IsOverdue = types.IsOverdueYes
		orderData.Utime = tools.GetUnixMillis()

		models.UpdateOrder(&orderData)

		// 写一条操作日志
		models.OpLogWrite(0, orderData.Id, models.OpCodeOrderUpdate, orderData.TableName(), origin, orderData)

		monitor.IncrOrderCount(orderData.CheckStatus)
	}

	//更新宽限期利息与罚息
	service.OverdueUpdatePenalty(orderID)

	// 处理逾期案件
	service.HandleOverdueCase(orderID)

	timetag := tools.NaturalDay(0)
	orderExt, err := models.GetOrderExt(orderID)
	if err != nil {
		orderExt = models.OrderExt{}
		orderExt.OrderId = orderID
		orderExt.OverdueRunTime = timetag
		orderExt.Ctime = timetag
		orderExt.Add()
	} else {
		orderExt.OverdueRunTime = timetag
		orderExt.Utime = timetag
		orderExt.Update()
	}
}
