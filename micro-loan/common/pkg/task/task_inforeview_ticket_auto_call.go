package task

import (
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/toolbox"
	"github.com/garyburd/redigo/redis"

	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/nxtele"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

var autoCallOnce bool
var autoCallName string
var AutoCallWg sync.WaitGroup

func init() {
	flag.BoolVar(&autoCallOnce, "auto-call-once", false, "run one task")
}

type InfoReviewAutoCallTask struct {
}

type infoReviewAutoCallTaskInfo struct {
	Time string
	Func func() error
}

var infoReviewAutoCallFunc map[string]infoReviewAutoCallTaskInfo = map[string]infoReviewAutoCallTaskInfo{
	//go run cli-task/task.go --name=info_review_auto_call --auto-call-once=true
	/*
		types.RiskCtlWaitAuqtoCallKeyPre: infoReviewAutoCallTaskInfo{"",
			func() error {
				return InfoReviewAutoCall()
			}},
	*/
}

func updateInfoReviewAutoCallFunc(times string) {
	if len(times) <= 0 || times == types.CallTimeBlankVal {
		return
	}

	IndonesiaTime := strings.Split(times, types.CallTimeDelimiter)
	for _, v := range IndonesiaTime {
		t := strings.Split(v, ":")
		if len(v) <= 0 || len(t) <= 0 {
			continue
		}

		hour := t[0]
		hourInt, _ := tools.Str2Int(hour)
		hourInt = (hourInt + 17) % 24 // 印尼小时转化为Unix对应的小时

		var minuteInt int
		if len(t) >= 2 {
			minute := t[1]
			minuteInt, _ = tools.Str2Int(minute)
		}

		timeStr := fmt.Sprintf("0 %d %d * * *", minuteInt, hourInt)
		logs.Info("[updateInfoReviewAutoCallFunc] timeStr:", timeStr)

		key := fmt.Sprintf("%s_%d_%d", types.InfoReviewAutoCallKeyPre, hourInt, minuteInt)
		var tmp infoReviewAutoCallTaskInfo
		tmp.Time = timeStr
		tmp.Func = func() error { return InfoReviewAutoCall() }

		infoReviewAutoCallFunc[key] = tmp

	}
}

func getCallNumConfig() (callNum int) {
	callNumStr := config.ValidItemString(types.InfoReviewAutoCallNumName)
	callNum, _ = tools.Str2Int(callNumStr)

	return
}

// 获取InfoReview工单自动外呼配置并升级全局变量
func getInfoReviewAutoCallConfig() (callNum int) {
	// 获取外呼时间配置
	callNum = getCallNumConfig()
	if callNum <= 0 {
		return
	}

	callTimes := config.ValidItemString(types.InfoReviewAutoCallTimeName)
	updateInfoReviewAutoCallFunc(callTimes)

	return
}

// 对 InfoReview 的工单，确认电话能否打通
func (c *InfoReviewAutoCallTask) Start() {

	if autoCallOnce {
		InfoReviewAutoCall()
		return

	} else {

		callNum := getInfoReviewAutoCallConfig()
		if callNum <= 0 || len(infoReviewAutoCallFunc) <= 0 {
			logs.Info("[InfoReviewAutoCallTask] politeness exit, not have timed task.")
			return
		}

		storageClient := storage.RedisStorageClient.Get()
		defer storageClient.Close()

		lockKey := beego.AppConfig.String("info_review_auto_call_task_lock")
		lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", tools.SECONDAHOUR/6, "NX")

		if err != nil || lock == nil {
			logs.Error("[InfoReviewAutoCallTask] process is working, so, I will exit.")
			// ***! // 很重要!
			close(done)
			return
		}

		for k, v := range infoReviewAutoCallFunc {
			if len(v.Time) > 0 {
				tk := toolbox.NewTask(k, v.Time, v.Func)
				toolbox.AddTask(k, tk)
			}
		}
		toolbox.StartTask()

		qName := beego.AppConfig.String("info_review_auto_call_task")

		for {
			if cancelled() {
				logs.Info("[InfoReviewAutoCallTask] receive exit cmd.")
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
				logs.Info("[InfoReviewAutoCallTask] receive exit cmd")
				close(done)
				// ***! // 很重要!
				continue
			}

			time.Sleep(time.Second)
		}

		TimerWg.Wait()

		// -1 正常退出时,释放锁
		storageClient.Do("DEL", lockKey)
		logs.Info("[InfoReviewAutoCallTask] politeness exit.")
	}

}

func (c *InfoReviewAutoCallTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("info_review_auto_call_task_lock")
	storageClient.Do("DEL", lockKey)
}

func getInfoReviewCallQueueName() (qName string) {
	qName = beego.AppConfig.String("info_review_auto_call_queue")
	return
}

func getInfoReviewCallSetName() (sName string) {
	sName = beego.AppConfig.String("info_review_auto_call_sets")
	return
}

// 获取订单的即将呼叫的次数缓存key，已呼叫次数保存到缓存
func getOrderCallNumKey(order int64) (key string) {
	key = fmt.Sprintf("%s:%d", types.InfoReviewAutoCallKeyPre, order)
	return
}

// 设置订单的将呼叫的次数
func setOrderCallNum(orderID int64) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	key := getOrderCallNumKey(orderID)
	val, err := storageClient.Do("EXISTS", key)
	if err == nil && 0 == val.(int64) {
		storageClient.Do("SET", key, 1, "EX", tools.SECONDADAY, "NX")

	} else {
		storageClient.Do("INCR", key)
	}
}

// 获取订单已呼叫的次数
func getOrderCallNum(orderID int64) (num int) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	key := getOrderCallNumKey(orderID)
	val, err := storageClient.Do("GET", key)
	if err != nil || val == nil {
		return
	}

	num, err = tools.Str2Int(string(val.([]byte)))

	return
}

// 更新订单状态
func updateOrderStatus(orderID int64, isPass bool) {
	orderData, _ := models.GetOrder(orderID)

	if isPass {
		orderData.CheckStatus = types.LoanStatusWait4Loan
		orderData.RiskCtlStatus = types.RiskCtlAutoCallPass
	} else {
		orderData.CheckStatus = types.LoanStatusReject
		orderData.RiskCtlStatus = types.RiskCtlAutoCallReject
		orderData.RejectReason = types.RejectReasonLackCredit
	}

	orderData.CheckTime = tools.GetUnixMillis()
	orderData.Utime = orderData.CheckTime
	//orderData.RiskCtlFinishTime = orderData.CheckTime
	models.UpdateOrder(&orderData)
}

// 对 InfoReview 的工单，确认电话能否打通
func InfoReviewAutoCall() error {
	logs.Info("[InfoReviewAutoCall] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("info_review_auto_call_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", tools.SECONDAHOUR/6, "NX")
	if err != nil || lock == nil {
		logs.Error("[InfoReviewAutoCall] process is working, so, I will exit.")
		close(done)
		return nil
	}

	var findNum int // 记录查询符合条件的订单的次数，当次数超过两次后，退出
	for {
		if cancelled() {
			logs.Info("[InfoReviewAutoCall] receive exit cmd.")
			break
		}

		setName := getInfoReviewCallSetName()
		qVal, err := storageClient.Do("EXISTS", setName)
		// 初始化去重集合
		if err == nil && 0 == qVal.(int64) {
			storageClient.Do("SADD", setName, 1)
			storageClient.Do("EXPIRE", setName, tools.SECONDAHOUR/6)
		}

		// 生产队列, 小批量处理
		qRealName := getInfoReviewCallQueueName()
		qVal, err = storageClient.Do("LLEN", qRealName)
		if err == nil && qVal != nil && 0 == qVal.(int64) {
			logs.Info("[InfoReviewAutoCall] %s 队列为空,开始按条件生成.", qRealName)

			var idsBox []string
			setsMem, err := redis.Values(storageClient.Do("SMEMBERS", setName))
			if err != nil || setsMem == nil {
				logs.Error("[InfoReviewAutoCall] 生产InfoReview工单自动外呼队列无法从集合中取到元素,休眠1秒后将重试.")
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			for _, m := range setsMem {
				idsBox = append(idsBox, string(m.([]byte)))
			}
			// 理论上不会出现
			if len(idsBox) == 0 {
				logs.Error("[InfoReviewAutoCall] 生产InfoReview工单自动外呼队列出错了,集合中没有元素,不符合预期,程序将退出.")
				close(done)
				break
			}

			orderList, _ := service.GetWaitAutoCallOrderList(idsBox)

			// 如果没有满足条件的数据, work goroutine 也不用启动了
			if len(orderList) == 0 {
				findNum++
				if findNum >= 2 {
					logs.Info("[InfoReviewAutoCall] 生产InfoReview工单自动外呼队列没有满足条件的数据,程序将退出.")
					break
				}
				logs.Info("[InfoReviewAutoCall] 生产InfoReview工单自动外呼队列没有满足条件的数据,休眠1秒后将重试.")
				time.Sleep(1000 * time.Millisecond)
				continue
			}

			for _, orderID := range orderList {

				setOrderCallNum(orderID)

				// 即将外呼的次数
				num := getOrderCallNum(orderID)
				callNumConfig := getCallNumConfig()
				// 即将外呼次数超过配置次数时，修改订单风控状态（自动外呼拒绝）和订单状态（审核拒绝）
				if callNumConfig < num {
					logs.Info("[InfoReviewAutoCall] 该订单的外呼次数到达配置值, orderID:", orderID, ", callNumConfig:", callNumConfig, ", 即将外呼的次数num:", num)

					updateOrderStatus(orderID, false)
					continue
				}

				storageClient.Do("LPUSH", qRealName, orderID)
			}
		}

		// 消费队列
		var wg sync.WaitGroup
		wg.Add(1)
		consumeInfoReviewAutoCallQueue(&wg, int(types.VoiceTypeInfoReview))

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()

		//这个为crontab用，只跑一次
		break
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[InfoReviewAutoCall] politeness exit.")

	return nil
}

// 消费InfoReview工单自动外呼队列
func consumeInfoReviewAutoCallQueue(wg *sync.WaitGroup, workerID int) {

	defer wg.Done()
	logs.Info("It will do consumeInfoReviewAutoCallQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	var orderIDArr []int64
	var endFlag int
	var orderID int64

	qRealName := getInfoReviewCallQueueName()

	postMobileNum, _ := config.ValidItemInt("nxtele_post_mobile_num")
	if postMobileNum <= 0 {
		postMobileNum = 5000
	}
	logs.Info("[consumeInfoReviewAutoCallQueue] postMobileNum:", postMobileNum, "workID:", workerID)

	if cancelled() {
		logs.Info("[consumeInfoReviewAutoCallQueue] receive exit cmd, workID:", workerID)
		return
	}

	for {
		qValueByte, err := storageClient.Do("RPOP", qRealName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Info("[consumeInfoReviewAutoCallQueue] no data for consume, I will exit after 500ms, workID:", workerID)
			time.Sleep(500 * time.Millisecond)
			endFlag = 1
			goto next
		}

		orderID, _ = tools.Str2Int64(string(qValueByte.([]byte)))
		if orderID == types.TaskExitCmd {
			logs.Info("[consumeInfoReviewAutoCallQueue] receive exit cmd, I will exit after jobs done. workID:", workerID,
				", orderID:", orderID)
			close(done)
			break
		}
		orderIDArr = append(orderIDArr, orderID)

	next:
		if len(orderIDArr) == postMobileNum || endFlag == 1 {
			logs.Info("[consumeInfoReviewAutoCallQueue] endFlag:", endFlag, "orderIDArr:", orderIDArr, "workID:", workerID)
			if len(orderIDArr) > 0 {
				orderIDArrTmp := orderIDArr
				addCurrentData(tools.Int642Str(orderIDArrTmp[0]), "orderId")
				handleInfoReviewAutoCall(orderIDArrTmp, workerID)
				removeCurrentData(tools.Int642Str(orderIDArrTmp[0]))

				for 0 < len(orderIDArr) {
					orderIDArr = append(orderIDArr[:0], orderIDArr[1:]...)
				}
			} else {
				break
			}
		}
	}
}

func handleInfoReviewAutoCall(orderIDArr []int64, workerID int) {
	logs.Info("[handleInfoReviewAutoCall] orderIDArr:", orderIDArr, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleInfoReviewAutoCall] panic orderId:%d, workId:%d, err:%v", orderIDArr, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	setName := getInfoReviewCallSetName()

	var accountBase models.AccountBase
	var mobileArr []string
	for _, orderID := range orderIDArr {

		qVal, err := storageClient.Do("SADD", setName, orderID)
		// 说明有错,或已经处理过,忽略本次操作
		if err != nil || 0 == qVal.(int64) {
			logs.Info("[handleInfoReviewAutoCall] 此订单已经处理过,忽略之. orderID: %d, workerID: %d", orderIDArr, workerID)
			continue
		}

		order, _ := models.GetOrder(orderID)
		accountBase, _ = dao.CustomerOne(order.UserAccountId)
		if len(accountBase.Mobile) > 0 {
			mobileArr = append(mobileArr, accountBase.Mobile)
		}
	}

	mobile := strings.Join(mobileArr, ",")
	// 发送InfoReview工单的自动呼叫逻辑
	if len(mobile) > 0 {
		// 自动外呼
		nxteleResp, err := nxtele.Send(types.VoiceTypeInfoReview, mobile)
		if err != nil || nxteleResp == nil {
			logs.Error("[handleInfoReviewAutoCall] Send voice call request or parse response occur error, orderId:%d, workId:%d, err:%v",
				orderIDArr, workerID, err)
			return
		}
		isSuccess := nxteleResp.IsSuccess()
		sid := nxteleResp.GetSID()
		logs.Info("[handleInfoReviewAutoCall] Send voice call response, isSuccess:%d, SID:%d, orderId:%d, workId:%d",
			isSuccess, sid, orderIDArr, workerID)

		if isSuccess == 1 && sid > 0 {
			// 保存呼叫结果
			SaveSidStatus(sid, workerID)
		}
	}

	// 查询自动外呼结果
	queryInfoReviewAutoCall(orderIDArr)

	return
}

// 查询自动外呼结果
func queryInfoReviewAutoCall(orderIDArr []int64) {
	logs.Info("[queryInfoReviewAutoCall] queryInfoReviewAutoCall handle")

	for _, orderID := range orderIDArr {

		order, _ := models.GetOrder(orderID)
		accountBase, _ := dao.CustomerOne(order.UserAccountId)

		if len(accountBase.Mobile) > 0 {
			mobile := nxtele.MobileFormat(accountBase.Mobile)
			// 检查最近三次并且两个小时内的自动外呼记录
			voiceReminds, _ := models.GetVoiceRemindByMobileAndStatus(mobile, types.VoiceTypeInfoReview)
			for _, v := range voiceReminds {
				diff := tools.GetUnixMillis() - v.Ctime
				if v.Duration > 0 && diff < tools.MILLSSECONDADAY {
					updateOrderStatus(orderID, true)
					break
				}

				num := getOrderCallNum(orderID)
				callNumConfig := getCallNumConfig()
				// 已外呼次数超过配置次数时，修改订单风控状态（自动外呼拒绝）和订单状态（审核拒绝）
				if callNumConfig <= num {
					logs.Info("[queryInfoReviewAutoCall] 该订单的外呼次数到达配置值, orderID:", orderID, ", callNumConfig:", callNumConfig, ", 已外呼的次数num:", num)

					updateOrderStatus(orderID, false)
					break
				}

			}
		}
	}

	return
}
