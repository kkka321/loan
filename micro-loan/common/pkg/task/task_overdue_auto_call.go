package task

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"

	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/nxtele"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type OverdueAutoCallTask struct {
}

// 对 逾期 的订单，语音提醒
func (c *OverdueAutoCallTask) Start() {
	logs.Info("[TaskHandleOverdueAutoCall] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("overdue_auto_call_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[TaskHandleOverdueAutoCall] process is working, so, I will exit.")
		close(done)
		return
	}

	var findNum int // 记录查询符合条件的订单的次数，当次数超过两次后，退出
	for {
		if cancelled() {
			logs.Info("[TaskHandleOverdueAutoCall] receive exit cmd.")
			break
		}

		TaskHeartBeat(storageClient, lockKey)

		setsName := beego.AppConfig.String("overdue_auto_call_sets")
		todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))

		qVal, err := storageClient.Do("EXISTS", todaySetName)
		// 初始化去重集合
		if err == nil && 0 == qVal.(int64) {
			storageClient.Do("SADD", todaySetName, 1)
			storageClient.Do("EXPIRE", todaySetName, tools.SECONDAHOUR/3)
		}

		// 生产队列, 小批量处理
		qName := beego.AppConfig.String("overdue_auto_call_queue")
		// 对于 逾期的订单 存储到同一的queue
		qVal, err = storageClient.Do("LLEN", qName)
		if err == nil && qVal != nil && 0 == qVal.(int64) {
			logs.Info("[TaskHandleOverdueAutoCall] %s 队列为空,开始按条件生成.", qName)

			var idsBox []string
			setsMem, err := redis.Values(storageClient.Do("SMEMBERS", todaySetName))
			if err != nil || setsMem == nil {
				logs.Error("[TaskHandleOverdueAutoCall] 生产逾期订单队列无法从集合中取到元素,休眠1秒后将重试.")
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			for _, m := range setsMem {
				idsBox = append(idsBox, string(m.([]byte)))
			}
			// 理论上不会出现
			if len(idsBox) == 0 {
				logs.Error("[TaskHandleOverdueAutoCall] 生产逾期订单队列出错了,集合中没有元素,不符合预期,程序将退出.")
				close(done)
				break
			}

			days := config.ValidItemString("nxtele_overdue_call_days")
			orderList, _ := service.GetOverdueOrderListByDays(idsBox, days)

			// 如果没有满足条件的数据, work goroutine 也不用启动了
			if len(orderList) == 0 {
				findNum++
				if findNum >= 2 {
					logs.Info("[TaskHandleOverdueAutoCall] 生产逾期订单队列没有满足条件的数据,程序将退出.")
					break
				}
				logs.Info("[TaskHandleOverdueAutoCall] 生产逾期订单队列没有满足条件的数据,休眠1秒后将重试.")
				time.Sleep(1000 * time.Millisecond)
				continue

			}

			for _, orderId := range orderList {
				storageClient.Do("LPUSH", qName, orderId)
			}
		}

		// 消费队列
		var wg sync.WaitGroup

		wg.Add(1)
		go consumeOverdueAutoCallQueue(&wg, int(types.VoiceTypeOverdue))

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()

		//这个为crontab用，只跑一次
		break
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[TaskHandleOverdueAutoCall] politeness exit.")
}

func (c *OverdueAutoCallTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("overdue_auto_call_lock")
	storageClient.Do("DEL", lockKey)
}

// 消费还款/逾期语音订单队列
func consumeOverdueAutoCallQueue(wg *sync.WaitGroup, workerID int) {

	defer wg.Done()
	logs.Info("It will do consumeOverdueOrderQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	var orderIDArr []int64
	var endFlag int
	var orderID int64

	qName := beego.AppConfig.String("overdue_auto_call_queue")

	postMobileNum, _ := config.ValidItemInt("nxtele_post_mobile_num")
	if postMobileNum <= 0 {
		postMobileNum = 5000
	}
	logs.Info("[consumeOverdueOrderQueue] postMobileNum:", postMobileNum, "workID:", workerID)

	if cancelled() {
		logs.Info("[consumeOverdueOrderQueue] receive exit cmd, workID:", workerID)
		return
	}

	for {
		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Info("[consumeOverdueOrderQueue] no data for consume, I will exit after 500ms, workID:", workerID)
			time.Sleep(500 * time.Millisecond)
			endFlag = 1
			goto next
		}

		orderID, _ = tools.Str2Int64(string(qValueByte.([]byte)))
		if orderID == types.TaskExitCmd {
			logs.Info("[consumeOverdueOrderQueue] receive exit cmd, I will exit after jobs done. workID:", workerID, ", orderID:", orderID)
			close(done)
			break
		}
		orderIDArr = append(orderIDArr, orderID)

	next:
		if len(orderIDArr) == postMobileNum || endFlag == 1 {
			logs.Info("[consumeOverdueOrderQueue] endFlag:", endFlag, "orderIDArr:", orderIDArr, "workID:", workerID)
			if len(orderIDArr) > 0 {
				orderIDArrTmp := orderIDArr
				addCurrentData(tools.Int642Str(orderIDArrTmp[0]), "orderId")
				handleOverdueAutoCall(orderIDArrTmp, int(workerID))
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

func handleOverdueAutoCall(orderIDArr []int64, workerID int) {
	logs.Info("[handleOverdueAutoCall] orderIDArr:", orderIDArr, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleOverdueAutoCall] panic orderId:%d, workId:%d, err:%v", orderIDArr, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	setsName := beego.AppConfig.String("overdue_auto_call_sets")
	todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))

	var accountBase models.AccountBase
	var mobileArr []string
	for _, orderID := range orderIDArr {

		qVal, err := storageClient.Do("SADD", todaySetName, orderID)
		// 说明有错,或已经处理过,忽略本次操作
		if err != nil || 0 == qVal.(int64) {
			logs.Info("[handleOverdueAutoCall] 此订单已经处理过,忽略之. orderID: %d, workerID: %d", orderIDArr, workerID)
			continue
		}

		order, _ := models.GetOrder(orderID)
		accountBase, _ = dao.CustomerOne(order.UserAccountId)
		if len(accountBase.Mobile) > 0 {
			mobileArr = append(mobileArr, accountBase.Mobile)
		}
	}

	mobile := strings.Join(mobileArr, ",")
	// overdue发送自动外呼
	if len(mobile) > 0 {
		nxteleResp, err := nxtele.Send(types.VoiceTypeOverdue, mobile)
		if err != nil || nxteleResp == nil {
			logs.Error("[handleOverdueAutoCall] Send voice call request or parse response occur error, orderId:%d, workId:%d, err:%v",
				orderIDArr, workerID, err)
			return
		}
		isSuccess := nxteleResp.IsSuccess()
		sid := nxteleResp.GetSID()
		logs.Info("[handleOverdueAutoCall] Send voice call response, isSuccess:%d, SID:%d, orderId:%d, workId:%d",
			isSuccess, sid, orderIDArr, workerID)

		if isSuccess == 1 && sid > 0 {
			SaveSidStatus(sid, workerID)
		}
	}
}
