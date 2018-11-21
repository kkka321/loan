package task

import (
	"fmt"
	"sync"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/service"
	"micro-loan/common/tools"
)

func RunRollTask() error {
	logs.Info("[RunRollTask] start launch.")

	TimerWg.Add(1)
	defer TimerWg.Done()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("roll_order_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[RunRollTask] process is working, so, I will exit.")
		return nil
	}

	for {
		if cancelled() {
			logs.Info("[RunRollTask] receive exit cmd.")
			break
		}

		setsName := beego.AppConfig.String("roll_order_sets")
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
		}

		// 生产队列,小批量处理
		qName := beego.AppConfig.String("roll_order")
		qVal, err = storageClient.Do("LLEN", qName)
		if err == nil && qVal != nil && 0 == qVal.(int64) {
			logs.Info("[RunRollTask] %s 队列为空,开始按条件生成.", qName)

			var idsBox []string
			setsMem, err := redis.Values(storageClient.Do("SMEMBERS", todaySetName))
			if err != nil || setsMem == nil {
				logs.Error("[RunRollTask] 队列无法从集合中取到元素")
				break
			}
			for _, m := range setsMem {
				idsBox = append(idsBox, string(m.([]byte)))
			}
			// 理论上不会出现
			if len(idsBox) == 0 {
				logs.Error("[RunRollTask] 集合中没有元素,不符合预期,程序将退出.")
				break
			}

			orderList, _ := service.GetRollApplyOrderList(idsBox, 100)

			// 如果没有满足条件的数据,work goroutine 也不用启动了
			if len(orderList) == 0 {
				break
			}

			for _, order := range orderList {
				storageClient.Do("LPUSH", qName, order.Id)
			}

			// 消费队列
			var wg sync.WaitGroup
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go consumeRollOrderQueue(&wg, i)
			}

			// 主 goroutine,等待工作 goroutine 正常结束
			wg.Wait()

		} else {
			logs.Error("[RunRollTask] get roll_order wrong err:%v", err)
			break
		}
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[RunRollTask] politeness exit.")

	return nil
}

// 消费逾期订单队列
func consumeRollOrderQueue(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	logs.Info("It will do consumeRollOrderQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	qName := beego.AppConfig.String("roll_order")
	for {
		if cancelled() {
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Info("[consumeRollOrderQueue] no data for consume, I will exit, workID:", workerID)
			break
		}

		orderID, _ := tools.Str2Int64(string(qValueByte.([]byte)))

		// 真正开始工作了
		addCurrentData(tools.Int642Str(orderID), "orderId")
		handleRollApplyOrder(orderID, workerID)
		removeCurrentData(tools.Int642Str(orderID))
	}
}

func handleRollApplyOrder(orderID int64, workerID int) {
	logs.Info("[handleRollApplyOrder] orderID:", orderID, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleRollApplyOrder] panic orderId:%d, workId:%d, err:%v", orderID, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	setsName := beego.AppConfig.String("roll_order_sets")
	todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))
	qVal, err := storageClient.Do("SADD", todaySetName, orderID)

	// 说明有错,或已经处理过,忽略本次操作
	if err != nil || 0 == qVal.(int64) {
		logs.Info("[handleRollApplyOrder] 此订单已经处理过,忽略之. orderID: %d, workerID: %d", orderID, workerID)
		return
	}

	service.HandleRollOrder(orderID)
}
