package task

import (
	"fmt"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type MonitorTask struct {
}

// TaskHandleEventPush 处理事件任务
func (c *MonitorTask) Start() {
	logs.Info("[MonitorTask] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("monitor_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")

	if err != nil || lock == nil {
		logs.Error("[MonitorTask] process is working, so, I will exit.")
		// ***! // 很重要!
		close(done)
		return
	}

	go backupHistoryMonitor()

	for {
		if cancelled() {
			logs.Info("[MonitorTask] receive exit cmd.")
			break
		}

		TaskHeartBeat(storageClient, lockKey)

		qName := beego.AppConfig.String("monitor")

		qValueByte, _ := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据,退出工作 goroutine
		if qValueByte != nil {
			orderID, _ := tools.Str2Int64(string(qValueByte.([]byte)))
			if orderID == types.TaskExitCmd {
				close(done)
				continue
			}
		}

		// 消费队列
		var wg sync.WaitGroup
		wg.Add(1)
		go monitorOrder(&wg)

		wg.Add(1)
		go monitorThirdparty(&wg)

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()

		time.Sleep(time.Second)
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)

	lockKey = beego.AppConfig.String("monitor_data_lock")
	storageClient.Do("DEL", lockKey)

	logs.Info("[MonitorTask] politeness exit.")
}

func (c *MonitorTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("monitor_lock")
	storageClient.Do("DEL", lockKey)

	lockKey = beego.AppConfig.String("monitor_data_lock")
	storageClient.Do("DEL", lockKey)
}

func backupOrderStatistics(date int64) {
	if !monitor.IsOrderKeyExist(date) {
		logs.Info("[backupOrderStatistics] no data exist key:%d", date)
		return
	}

	strDate := tools.GetDate(date / 1000)
	oldData, err := dao.OrderStatisticsByDate(strDate)
	if err == nil {
		logs.Info("[backupOrderStatistics] delete old date date:%s", strDate)
		oldData.Del()
	}

	var order models.OrderStatistics = models.OrderStatistics{}
	monitor.GetOrderStatistics(date, &order)
	order.Id = 0
	order.StatisticsDate = strDate
	order.Add()

	logs.Info("[backupOrderStatistics] save data date:%d", date)

	monitor.DelOrderKey(date)
}

func backupThirdpartyStatistics(date int64) {
	if !monitor.IsThirdpartyKeyExist(date) {
		logs.Info("[backupThirdpartyStatistics] no data exist key:%d", date)
		return
	}

	strDate := tools.GetDate(date / 1000)
	oldDatas, err := dao.ThirdpartyStatisticsByDate(strDate)
	if err == nil {
		logs.Info("[backupThirdpartyStatistics] delete old date date:%s", strDate)
		for _, v := range oldDatas {
			v.Del()
		}
	}

	list := monitor.GetThirdpartyStatistics(date)
	for _, v := range list {
		v.Id = 0
		v.StatisticsDate = strDate
		v.Add()
	}

	logs.Info("[backupThirdpartyStatistics] save data date:%d", date)

	monitor.DelThirdpartyKey(date)
}

func backupHistoryMonitor() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	for {
		lastDate := tools.NaturalDay(-1)
		lockKey := beego.AppConfig.String("monitor_data_lock")

		lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")

		if err != nil || lock == nil {
			logs.Warn("[backupHistoryMonitor] another thread run")
			time.Sleep(30 * time.Minute)
			continue
		}

		backupOrderStatistics(lastDate)

		backupThirdpartyStatistics(lastDate)

		storageClient.Do("DEL", lockKey)
		time.Sleep(30 * time.Minute)
	}
}

func monitorOrder(wg *sync.WaitGroup) {
	defer wg.Done()

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[monitorOrder] panic err:%v", x)
			logs.Error(tools.FullStack())
		}
	}()

	if cancelled() {
		return
	}

	freq, _ := beego.AppConfig.Int("monitor_order_freq")
	threshold, _ := beego.AppConfig.Int("monitor_order_threshold")
	dateKey := beego.AppConfig.String("monitor_order_date")

	var order models.OrderStatistics = models.OrderStatistics{}
	monitor.GetOrderStatistics(tools.NaturalDay(0), &order)

	if order.WaitReview == 0 {
		return
	}

	rv := (order.WaitManual + order.WaitLoan) * 100 / order.WaitReview
	logs.Info("[monitorOrder] pass:%d, total:%d", order.WaitManual+order.WaitLoan, order.WaitReview)
	if rv >= threshold {
		return
	}

	title := "Order Warning"
	body := fmt.Sprintf("pass rate %d < %d", rv, threshold)
	service.SendNotification(dateKey, freq, title, body)
}

func monitorThirdparty(wg *sync.WaitGroup) {
	defer wg.Done()

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[monitorOrder] panic err:%v", x)
			logs.Error(tools.FullStack())
		}
	}()

	if cancelled() {
		return
	}

	freq, _ := beego.AppConfig.Int("monitor_thirdparty_freq")
	threshold, _ := beego.AppConfig.Int("monitor_thirdparty_threshold")
	dateKey := beego.AppConfig.String("monitor_thirdparty_date")

	list := monitor.GetThirdpartyStatistics(tools.NaturalDay(0))

	for _, v := range list {
		total := v.Fail + v.Success

		if total == 0 {
			continue
		}

		rv := v.Success * 100 / total
		logs.Info("[monitorThirdparty] pass:%d, total:%d", v.Success, total)
		if rv >= threshold {
			continue
		}

		thirdname := ""
		if s, ok := models.ThirdpartyNameMap[v.Thirdparty]; ok {
			thirdname = s
		} else {
			thirdname = tools.Int2Str(v.Thirdparty)
		}

		key := dateKey + "_" + tools.Int2Str(v.Thirdparty)
		title := "Thirdparty Warning"
		body := fmt.Sprintf("thirdparty:%s pass rate %d < %d", thirdname, rv, threshold)
		service.SendNotification(key, freq, title, body)
	}
}
