package task

import (
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type CustomerRecallTask struct {
}

// 高质量被拒用户召回 {{{
func (c *CustomerRecallTask) Start() {
	logs.Info("[CustomerRecallTask] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("customer_recall_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[CustomerRecallTask] process is working, so, I will exit.")
		// ***! // 很重要!
		close(done)
		return
	}
	lastAccountID := int64(0)

	//打标签
	timetag := tools.NaturalDay(0)
	//给客户打电核拒绝召回标签
	service.CustomerRecallTag(timetag)

	for {
		if cancelled() {
			logs.Info("[CustomerRecallTask] receive exit cmd.")
			break
		}

		// 生产队列,小批量处理
		qName := beego.AppConfig.String("customer_recall_queue")
		qVal, err := storageClient.Do("LLEN", qName)
		if err == nil && qVal != nil && 0 == qVal.(int64) {
			logs.Info("[CustomerRecallTask] %s 队列为空,开始按条件生成.", qName)

			//获取评分不足可召回用户列表
			accountList, _ := models.GetNeedRecallCustomer(lastAccountID)
			// 如果没有满足条件的数据,work goroutine 也不用启动了
			if len(accountList) == 0 {
				logs.Info("[CustomerRecallTask] 没有可召回用户,任务完成.")
				break
			}

			for _, account := range accountList {
				lastAccountID = account.AccountId
				storageClient.Do("LPUSH", qName, account.AccountId)
			}
		}

		// 消费队列
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go consumeRecallQueue(&wg, i)
		}

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()

	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[CustomerRecallTask] politeness exit.")
}

func (c *CustomerRecallTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("customer_recall_lock")
	storageClient.Do("DEL", lockKey)
}

// 消费逾期订单队列
func consumeRecallQueue(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	logs.Info("It will do consumeRecallQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	qName := beego.AppConfig.String("customer_recall_queue")
	for {
		if cancelled() {
			logs.Info("[consumeRecallQueue] receive exit cmd, workID:", workerID)
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Info("[consumeRecallQueue] no data for consume, I will exit after 500ms, workID:", workerID)
			time.Sleep(500 * time.Millisecond)
			break
		}

		accountID, _ := tools.Str2Int64(string(qValueByte.([]byte)))
		if accountID == types.TaskExitCmd {
			logs.Info("[consumeRecallQueue] receive exit cmd, I will exit after jobs done. workID:", workerID, ", orderID:", accountID)
			// ***! // 很重要!
			close(done)
			break
		}

		// 真正开始工作了
		handleRecallScore(accountID, workerID)
	}
}

func handleRecallScore(accountID int64, workerID int) {
	logs.Info("[handleRecallScore] accountID:", accountID, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleRecallScore] panic orderId:%d, workId:%d, err:%v", accountID, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	service.HandleRecallCancleScore(accountID)
}
