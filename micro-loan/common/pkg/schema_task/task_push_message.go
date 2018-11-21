package schema_task

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gomodule/redigo/redis"

	"micro-loan/common/dao"
	"micro-loan/common/lib/gaws"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/google/push"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func pushCustomMsg(task *models.PushTask, param interface{}) (int, int) {
	var b []byte
	w := aws.NewWriteAtBuffer(b)
	gaws.AwsDownload2Stream(task.PushListPath, w)
	list := tools.ParseTargetList(string(w.Bytes()))

	chanList := make(chan string, 100)

	endSign := "-1"

	subFun := func(wg *sync.WaitGroup, no int) {
		defer wg.Done()

		total := 0
		succ := 0

		for {
			str, ok := <-chanList
			if str == endSign {
				close(chanList)
				break
			}

			if !ok {
				break
			}

			if task.PushWay == types.PushWayAccount {
				id, _ := tools.Str2Int64(str)
				if id == 0 {
					continue
				}

				t, s := push.SendFmsMessageV2(task.Id, id, task.Title, task.Body, task.MessageType, task.Mark, task.SkipTo, task.Version)
				total += t
				succ += s
			} else {
				uuid := tools.Md5(str)

				t, s := push.SendFmsMessageViaUuidV2(uuid, task.Title, task.Body, task.Version)
				total += t
				succ += s
			}
		}

		IncrPushCount(task.Id, total, succ)

		logs.Info("[runExportPush] push msg taskId:%d, no:%d, way:%d, total:%d, succ:%d", task.Id, no, task.PushWay, total, succ)
	}

	var wg sync.WaitGroup

	count := len(list)/10000 + 2
	for i := 0; i < count; i++ {
		wg.Add(1)
		go subFun(&wg, i)
	}

	for _, v := range list {
		chanList <- v
	}
	chanList <- endSign

	wg.Wait()

	return 0, 0
}

func pushBusinessMsg(task *models.PushTask, param interface{}) (int, int) {
	total := 0
	succ := 0

	accountId, ok := param.(int64)
	if ok {
		total, succ = push.SendFmsMessageV2(task.Id, accountId, task.Title, task.Body, task.MessageType, task.Mark, task.SkipTo, task.Version)
	} else {
		logs.Error("[pushBusinessMsg] unexcept param pushWay:%d, param:%v", task.PushWay, param)
	}

	return total, succ
}

func PushRegisterNoOrderAccount(task *models.PushTask) (int, int) {
	total := 0
	succ := 0

	now := tools.GetUnixMillis()
	nowStr := tools.MDateMHSDate(now)
	zero := tools.GetDateParseBackend(nowStr) * 1000

	maxId := int64(0)
	for {
		list, _ := dao.QueryRegisterNoOrderAccount(zero, now, maxId)
		if len(list) == 0 {
			break
		}

		for _, v := range list {
			t, s := push.SendFmsMessageV2(task.Id, v, task.Title, task.Body, task.MessageType, task.Mark, task.SkipTo, task.Version)
			total += t
			succ += s

			if v > maxId {
				maxId = v
			}
		}
	}

	return total, succ
}

func PushRegisterOrderNoKtpAccount(task *models.PushTask) (int, int) {
	total := 0
	succ := 0

	now := tools.GetUnixMillis()
	nowStr := tools.MDateMHSDate(now)
	zero := tools.GetDateParseBackend(nowStr) * 1000

	maxId := int64(0)
	for {
		list, _ := dao.QueryRegisterOrderNoKtp(zero, now, maxId)
		if len(list) == 0 {
			break
		}

		for _, v := range list {
			t, s := push.SendFmsMessageV2(task.Id, v, task.Title, task.Body, task.MessageType, task.Mark, task.SkipTo, task.Version)
			total += t
			succ += s

			if v > maxId {
				maxId = v
			}
		}
	}

	return total, succ
}

func PushNoRegister(task *models.PushTask) (int, int) {
	total := 0
	succ := 0

	list, _ := dao.QueryNoRegister()
	for _, v := range list {
		t, s := push.SendFmsMessageViaUuidV2(v, task.Title, task.Body, task.Version)
		total += t
		succ += s
	}

	return total, succ
}

func PushRepayMsg(task *models.PushTask) (int, int) {
	timetag := tools.NaturalDay(0)
	count := 2

	subFun := func(wg *sync.WaitGroup, ids []int64, no int) {
		defer wg.Done()

		total := 0
		succ := 0

		for _, v := range ids {
			order, err := models.GetOrder(v)
			if err != nil {
				continue
			}

			t, s := push.SendFmsMessageV2(task.Id, order.UserAccountId, task.Title, task.Body, task.MessageType, task.Mark, task.SkipTo, task.Version)
			total += t
			succ += s

			orderExt, err := models.GetOrderExt(v)
			if err != nil {
				orderExt = models.OrderExt{}
				orderExt.OrderId = v
				orderExt.RepayMsgRunTime = timetag
				orderExt.Ctime = timetag
				orderExt.Add()
			} else {
				orderExt.RepayMsgRunTime = timetag
				orderExt.Utime = timetag
				orderExt.Update()
			}
		}

		IncrPushCount(task.Id, total, succ)

		logs.Info("[PushRepayMsg] push msg taskId:%d, no:%d, way:%d, total:%d, succ:%d", task.Id, no, task.PushWay, total, succ)
	}

	for {
		orderList, _ := dao.GetRepayMessageOrderList(timetag, int64(100))
		if len(orderList) == 0 {
			break
		}

		var wg sync.WaitGroup
		preSize := len(orderList)/count + 1
		for i := 0; i < count; i++ {
			startIndex := i * preSize
			endIndex := startIndex + preSize

			if startIndex >= len(orderList) {
				break
			}

			if endIndex > len(orderList) {
				endIndex = len(orderList)
			}

			wg.Add(1)
			go subFun(&wg, orderList[startIndex:endIndex], i)
		}

		wg.Wait()
	}

	return 0, 0
}

func PushOverdueMsg(task *models.PushTask) (int, int) {
	count := 2

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	setsName := beego.AppConfig.String("overdue_message_sets")
	todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))
	yesterdaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(-1)))

	num, _ := storageClient.Do("EXISTS", yesterdaySetName)
	if num != nil && num.(int64) == 1 {
		storageClient.Do("DEL", yesterdaySetName)
	}

	qVal, err := storageClient.Do("EXISTS", todaySetName)
	if err == nil && 0 == qVal.(int64) {
		storageClient.Do("SADD", todaySetName, 1)
	}

	subFun := func(wg *sync.WaitGroup, ids []int64, no int) {
		subClient := storage.RedisStorageClient.Get()

		defer subClient.Close()
		defer wg.Done()

		total := 0
		succ := 0

		for _, v := range ids {
			qVal, err := subClient.Do("SADD", todaySetName, v)
			if err != nil || 0 == qVal.(int64) {
				logs.Info("[PushOverdueMsg] order repeat orderID:%d, workerID:%d, err:%v", v, no, err)
				continue
			}

			order, err := models.GetOrder(v)
			if err != nil {
				continue
			}

			t, s := push.SendFmsMessageV2(task.Id, order.UserAccountId, task.Title, task.Body, task.MessageType, task.Mark, task.SkipTo, task.Version)
			total += t
			succ += s
		}

		IncrPushCount(task.Id, total, succ)

		logs.Info("[PushOverdueMsg] push msg taskId:%d, no:%d, way:%d, total:%d, succ:%d", task.Id, no, task.PushWay, total, succ)
	}

	for {
		var idsBox []string
		setsMem, err := redis.Values(storageClient.Do("SMEMBERS", todaySetName))
		if err != nil || setsMem == nil {
			logs.Error("[PushOverdueMsg] SMEMBERS return error err:%v, setsMem:%v", err, setsMem)
			break
		}

		for _, m := range setsMem {
			idsBox = append(idsBox, string(m.([]byte)))
		}
		// 理论上不会出现
		if len(idsBox) == 0 {
			logs.Error("[idsBox] idsBox empty setsMem:%v", setsMem)
			break
		}

		orderList, _ := dao.GetOverdueMessageOrderList(idsBox)
		if len(orderList) == 0 {
			break
		}

		var wg sync.WaitGroup
		preSize := len(orderList)/count + 1
		for i := 0; i < count; i++ {
			startIndex := i * preSize
			endIndex := startIndex + preSize

			if startIndex >= len(orderList) {
				break
			}

			if endIndex > len(orderList) {
				endIndex = len(orderList)
			}

			wg.Add(1)
			go subFun(&wg, orderList[startIndex:endIndex], i)
		}

		wg.Wait()
	}

	return 0, 0
}

func StartPushBackup() {
	lockKey := beego.AppConfig.String("push_backup_lock")

	for {
		storageClient := storage.RedisStorageClient.Get()
		lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", 10*60, "NX")

		if err != nil || lock == nil {
			storageClient.Close()
			time.After(time.Hour)
			continue
		}

		backupHistoryPushData()

		storageClient.Do("DEL", lockKey)

		storageClient.Close()

		time.Sleep(time.Second)
	}
}

func backupHistoryPushData() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	setKey := beego.AppConfig.String("push_set") + tools.MDateMHSDate(tools.GetUnixMillis()-tools.MILLSSECONDADAY)

	num, _ := redis.Int(storageClient.Do("EXISTS", setKey))
	if num == 0 {
		return
	}

	keyList, _ := redis.Strings(storageClient.Do("SMEMBERS", setKey))

	count := 0
	for _, v := range keyList {
		list := strings.Split(v, ":")
		if len(list) < 3 {
			continue
		}

		id, _ := tools.Str2Int64(list[1])
		if id == 0 {
			continue
		}

		pushDate := tools.GetDateParseBackend(list[2]) * 1000

		totalNum, _ := redis.Int(storageClient.Do("HGET", v, push.MessageKeyTotal))
		succNum, _ := redis.Int(storageClient.Do("HGET", v, push.MessageKeySucc))
		readNum, _ := redis.Int(storageClient.Do("HGET", v, push.MessageKeyRead))

		record := models.PushTaskRecord{}
		record.TaskId = id
		record.ReadNum = readNum
		record.TotalNum = totalNum
		record.SuccNum = succNum
		record.PushDate = pushDate
		record.Ctime = tools.GetUnixMillis()
		if record.TotalNum > 0 {
			record.ReadRate = record.ReadNum * 100 / record.TotalNum
		}
		record.Insert()

		storageClient.Do("DEL", v)

		count++
	}

	logs.Info("[backupHistoryPushData] backup history data success key:%s, count:%d", setKey, count)

	storageClient.Do("DEL", setKey)
}

func runPushTask(task *models.PushTask, param interface{}) error {
	total := 0
	succ := 0

	if task.PushTarget == types.PushTargetCustom {
		total, succ = pushCustomMsg(task, param)
	} else if task.PushTarget == types.PushTargetRegisterNoOrder {
		total, succ = PushRegisterNoOrderAccount(task)
	} else if task.PushTarget == types.PushTargetRegisterOrderNoKtp {
		total, succ = PushRegisterOrderNoKtpAccount(task)
	} else if task.PushTarget == types.PushTargetNoRegister {
		total, succ = PushNoRegister(task)
	} else if task.PushTarget == types.PushTargetOverdue {
		total, succ = PushOverdueMsg(task)
	} else if task.PushTarget == types.PushTargetWaitRepayment {
		total, succ = PushRepayMsg(task)
	} else if task.PushTarget == types.PushTargetAllAccount {
		total, succ = PushAllAccount(task)
	} else {
		total, succ = pushBusinessMsg(task, param)
	}

	IncrPushCount(task.Id, total, succ)

	logs.Info("[runPushTask] push msg taskId:%d, way:%d, target:%d, param:%v, total:%d, succ:%d", task.Id, task.PushWay, task.PushTarget, param, total, succ)

	return nil
}

func IncrPushCount(id int64, total int, succ int) {
	nowStr := tools.MDateMHSDate(tools.GetUnixMillis())
	key := fmt.Sprintf("push:%d:%s", id, nowStr)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	exist, _ := redis.Int(storageClient.Do("HSETNX", key, push.MessageKeyTotal, 0))
	if exist > 0 {
		setKey := beego.AppConfig.String("push_set") + nowStr
		storageClient.Do("SADD", setKey, key)
	}

	storageClient.Do("HINCRBY", key, push.MessageKeyTotal, total)
	storageClient.Do("HINCRBY", key, push.MessageKeySucc, succ)
}

func PushMessage(id int64) error {
	taskInfo, err := models.GetPushTask(id)
	if err != nil {
		logs.Error("[PushMessage] GetPushTask return error taskId:%d, err:%v", id, err)
		return err
	}

	return runPushTask(&taskInfo, 0)
}

func PushBusinessMsg(target types.PushTarget, param interface{}) error {
	list, _ := models.GetPushTaskByTarget(target)
	for _, v := range list {
		runPushTask(&v, param)
	}

	return nil
}

func PushAllAccount(task *models.PushTask) (int, int) {
	count := 10

	chanList := make(chan int64, 1000)
	limit := int64(100)

	endSign := int64(-1)

	subFun := func(wg *sync.WaitGroup, no int) {
		defer wg.Done()

		total := 0
		succ := 0

		for {
			id, ok := <-chanList
			if id == endSign {
				close(chanList)
				break
			}

			if !ok {
				break
			}

			t, s := push.SendFmsMessageV2(task.Id, id, task.Title, task.Body, task.MessageType, task.Mark, task.SkipTo, task.Version)
			total += t
			succ += s
		}

		IncrPushCount(task.Id, total, succ)

		logs.Info("[runAllAccountPush] push msg taskId:%d, no:%d, way:%d, total:%d, succ:%d", task.Id, no, task.PushWay, total, succ)
	}

	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)
		go subFun(&wg, i)
	}

	for {
		maxId := int64(0)
		list, _ := dao.GetAllAccountList(maxId, limit)
		if len(list) == 0 {
			break
		}

		for _, v := range list {
			chanList <- v

			if v > maxId {
				maxId = v
			}
		}
	}

	chanList <- endSign

	wg.Wait()

	return 0, 0
}
