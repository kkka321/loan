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

type AuthoriationStatusCheck struct {
}

//  检查授信过期信息 {{{
func (c *AuthoriationStatusCheck) Start() {
	logs.Info("[AuthoriationStatusCheck] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("authoriation_status_check_lock")
	qName := beego.AppConfig.String("authoriation_status_check_queue")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", 24*60*60, "NX")
	if err != nil || lock == nil {
		logs.Error("[AuthoriationStatusCheck] process is working, so, I will exit.")
		// ***! // 很重要!
		close(done)
		return
	}
	defer storageClient.Do("DEL", lockKey)

	lastedAccountId := int64(0)
	for {
		if cancelled() {
			logs.Info("[AuthoriationStatusCheck] receive exit cmd.")
			break
		}

		TaskHeartBeat(storageClient, lockKey)

		// 生产队列,小批量处理
		qVal, err := storageClient.Do("LLEN", qName)
		if err == nil && qVal != nil && 0 == qVal.(int64) {
			logs.Info("[AuthoriationStatusCheck] %s 队列为空,开始按条件生成.", qName)

			// 队列是空,需要生成了
			// 1. 取数据
			accountList, _ := service.AccountList4ReviewAuthoriation(lastedAccountId)

			logs.Info("len(accountList):%d lastedAccountId:%d", len(accountList), lastedAccountId)
			// 如果没有满足条件的数据,work goroutine 也不用启动了
			if len(accountList) == 0 {
				time.Sleep(500 * time.Millisecond)
				logs.Info("[AuthoriationStatusCheck] 生产待审核自动减免订单队列没有满足条件的数据,退出.")
				break
			}
			lastedAccountId = accountList[len(accountList)-1].AccountId

			// 2. 加队列
			for _, a := range accountList {
				storageClient.Do("LPUSH", qName, a.AccountId)
			}
		}

		if err != nil {
			logs.Error("[AuthoriationStatusCheck] err :%v qVal:%v", err, qVal)
		}
		// 消费队列
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go consumeAuthoriationStatusCheckQueue(&wg, i, qName)
		}

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()
	}

	logs.Info("[AuthoriationStatusCheck] politeness exit.")
}

func (c *AuthoriationStatusCheck) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("authoriation_status_check_lock")
	storageClient.Do("DEL", lockKey)
}

func consumeAuthoriationStatusCheckQueue(wg *sync.WaitGroup, workerID int, qName string) {
	defer wg.Done()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	for {
		if cancelled() {
			logs.Info("[consumeAuthoriationStatusCheckQueue] receive exit cmd, workID:", workerID)
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Warn("[consumeAuthoriationStatusCheckQueue] no data for consume, I will exit after 500ms, workID :%d qName:%s  err:%v qValueByte:%v", workerID, qName, err, qValueByte)
			time.Sleep(500 * time.Millisecond)
			break
		}

		queueValToCmd, _ := tools.Str2Int64(string(qValueByte.([]byte)))
		if queueValToCmd == types.TaskExitCmd {
			logs.Info("[consumeAuthoriationStatusCheckQueue] receive exit cmd, I will exit after jobs done. workID:", workerID, ", queueVal:", queueValToCmd)
			// ***! // 很重要!
			close(done)
			break
		}
		// 真正开始工作了
		str := string(qValueByte.([]byte))
		accountId, _ := tools.Str2Int64(str)
		addCurrentData(str, "data")
		handleAuthoriationStatusCheck(workerID, accountId)
		removeCurrentData(str)

	}
}

func handleAuthoriationStatusCheck(workerID int, accountId int64) {
	logs.Info("[handleAuthoriationStatusCheck] workerID:%d accountId:%d", workerID, accountId)
	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleAuthoriationStatusCheck] panic accountId:%d, workId:%d, err:%v", accountId, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	service.CheckAuthoriationStatus(accountId)
}
