package task

import (
	"fmt"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
)

type RegisterRemindTask struct {
}

// 下载但未注册的用户，push注册消息
func (c *RegisterRemindTask) Start() {
	logs.Info("[TaskHandleRegisterRemind] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("register_remind_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[TaskHandleRegisterRemind] process is working, so, I will exit.")
		// ***! // 很重要!
		close(done)
		return
	}

	for {
		if cancelled() {
			logs.Info("[TaskHandleRegisterRemind] receive exit cmd.")
			break
		}

		TaskHeartBeat(storageClient, lockKey)

		setsName := beego.AppConfig.String("register_remind_sets")
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
			storageClient.Do("SADD", todaySetName, fmt.Sprintf("%d", 1))
		}

		// 生产队列,小批量处理
		qName := beego.AppConfig.String("register_remind_queue")
		qVal, err = storageClient.Do("LLEN", qName)
		if err == nil && qVal != nil && 0 == qVal.(int64) {
			logs.Info("[TaskHandleRegisterRemind] %s 队列为空,开始按条件生成.", qName)

			var uuidMd5Box []string
			setsMem, err := redis.Values(storageClient.Do("SMEMBERS", todaySetName))
			if err != nil || setsMem == nil {
				logs.Error("[TaskHandleRegisterRemind] 生产注册提醒队列无法从集合中取到元素,休眠1秒后将重试.")
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			for _, m := range setsMem {
				uuidMd5Box = append(uuidMd5Box, fmt.Sprintf("'%s'", string(m.([]byte))))
			}
			// 理论上不会出现
			if len(uuidMd5Box) == 0 {
				logs.Error("[TaskHandleRegisterRemind] 生产注册提醒队列出错了,集合中没有元素,不符合预期,程序将退出.")
				//! 很重要,确定程序正常退出
				close(done)
				break
			}

			uuidMd5List, _ := models.GetNeedRemindRegisterUUID()
			// 如果没有满足条件的数据,work goroutine 也不用启动了
			if len(uuidMd5List) == 0 {
				logs.Info("[TaskHandleRegisterRemind] 生产注册提醒队列没有满足条件的数据,休眠1秒后将重试.")
				time.Sleep(1000 * time.Millisecond)
				continue
			}

			// 记录发送注册提醒消息次数
			t := tools.GetUnixMillis()
			regRemindMsg := models.RegisterRemindMessage{}
			regRemindMsg.Date = tools.MDateMHSDate(t)
			regRemindMsg.Count = len(uuidMd5List)
			regRemindMsg.Ctime = t
			regRemindMsg.Utime = t
			regRemindMsg.Add()

			for _, uuidMd5 := range uuidMd5List {
				storageClient.Do("LPUSH", qName, uuidMd5)
			}
		}

		// 消费队列
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go consumeRegisterRemindQueue(&wg, i)
		}

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()

		//这个为crontab用，只跑一次
		break
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[TaskHandleRegisterRemind] politeness exit.")
}

func (c *RegisterRemindTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("register_remind_lock")
	storageClient.Do("DEL", lockKey)
}

// 消费注册提醒队列
func consumeRegisterRemindQueue(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	logs.Info("It will do consumeRegisterRemindQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	qName := beego.AppConfig.String("register_remind_queue")
	for {
		if cancelled() {
			logs.Info("[consumeRegisterRemindQueue] receive exit cmd, workID:", workerID)
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Info("[consumeRegisterRemindQueue] no data for consume, I will exit after 500ms, workID:", workerID)
			time.Sleep(500 * time.Millisecond)
			break
		}

		uuidMd5 := string(qValueByte.([]byte))

		addCurrentData(uuidMd5, "uuidMd5")
		handleRegisterRemind(uuidMd5, workerID)
		removeCurrentData(uuidMd5)
	}
}

func handleRegisterRemind(uuidMd5 string, workerID int) {
	logs.Info("[handleRegisterRemind] uuidMd5:", uuidMd5, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleRegisterRemind] panic uuidMd5:%s, workId:%d, err:%v", uuidMd5, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	setsName := beego.AppConfig.String("register_remind_sets")
	todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))
	qVal, err := storageClient.Do("SADD", todaySetName, uuidMd5)
	// 说明有错,或已经处理过,忽略本次操作
	if err != nil || 0 == qVal.(int64) {
		logs.Info("[handleRegisterRemind] 此uuidMd5已经处理过,忽略之. uuidMd5: %s, workerID: %d", uuidMd5, workerID)
		return
	}

	logs.Info("[handleRegisterRemind] Before SendRegisterRemindMessage. uuidMd5: %s, workerID: %d", uuidMd5, workerID)
	//push.SendRegisterRemindMessage(uuidMd5, i18n.GetMessageText(i18n.MsgRegisterRemindTitle), i18n.GetMessageText(i18n.MsgRegisterRemind))
}
