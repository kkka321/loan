package main

import (
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
)

// 	充值日期 充值金额
// 11-May-18 3503679925
// 23-Apr-18 2766930825
// 13-Apr-18 6739425
// 12-Apr-18 612426945
// 21-Mar-18 16133775

var chargeMapXend = map[string]int64{
	//"2018-03-21": 16133775,
	//"2018-04-12": 612426945,
	//"2018-04-13": 6739425,
	//"2018-04-23": 2766930825,
	//"2018-05-11": 3503679925,
	//"2018-08-08": 7167301785,
	"2018-09-21": 8582926850,
	"2018-10-02": 14815600000,
}

var chargeMapDoku = map[string]int64{
	//"2018-08-01": 276000001,
	"2018-08-08": 1145044249,
	"2018-08-27": 7232427675,
	"2018-09-06": 11811926175,
	"2018-09-25": 14825000000,
	"2018-10-05": 14815600000,
}

//8.30	Xendit 6月份提取	106303431
//8.30	Xendit 7月份提取	158702307
//9.29	Xendit 8月份提取	453846484
//9.29	Xendit 9月份提取1	700000000
//9.29	Xendit 9月份提取2	121838702
//9.29	Xendit 9月份提取3	50000
//9.30	Xendit 9月份提取4	100000000

var withdrawMapXend = map[string]int64{
	"2018-08-30": 106303431 + 158702307,
	"2018-09-29": 453846484 + 700000000 + 121838702 + 50000,
	"2018-09-30": 100000000,
}

var withdrawMapDoku = map[string]int64{
	//"2018-10-05": 148156000,
}

var paymentNameXend = "xendit"
var paymentNameDoku = "doku"

func main() {
	// 设置进程 title
	// +1 分布式锁
	// -1 正常退出时,释放锁
	procTitle := "fix-business-detail"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := fmt.Sprintf("lock:%s", procTitle)
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}
	defer storageClient.Do("DEL", lockKey)

	startDate := "2018-03-21"
	startMill := tools.GetDateParseBackend(startDate) * 1000

	currentDateMill := (tools.GetUnixMillis() / tools.MILLSSECONDADAY) * tools.MILLSSECONDADAY
	logs.Info("currentDateMi:%d startMill:%d", currentDateMill, startMill)
	for startMill+tools.MILLSSECONDADAY < currentDateMill {
		// 更新充值记录
		d := tools.MDateMHSDate(startMill)
		logs.Info("handle date:%s", d)
		if v, ok := chargeMapXend[d]; ok {
			logs.Info("chareg %d date:%s", v, d)
			err = service.DoSaveRecharge(d, paymentNameXend, v)
			if err != nil {
				logs.Error("[BusinessDetailController].DoSaveRecharge err:%s", err)
			}
		}

		if v, ok := withdrawMapXend[d]; ok {
			logs.Info("withdraw xendit %d date:%s", v, d)
			err = service.DoSaveWithdraw(d, paymentNameXend, v)
			if err != nil {
				logs.Error("[BusinessDetailController].DoSaveWithdraw err:%s", err)
			}
		}

		if v, ok := chargeMapDoku[d]; ok {
			logs.Info("chareg %d date:%s", v, d)
			err = service.DoSaveRecharge(d, paymentNameDoku, v)
			if err != nil {
				logs.Error("[BusinessDetailController].DoSaveRecharge err:%s", err)
			}
		}

		if v, ok := withdrawMapDoku[d]; ok {
			logs.Info("withdraw Doku %d date:%s", v, d)
			err = service.DoSaveWithdraw(d, paymentNameXend, v)
			if err != nil {
				logs.Error("[BusinessDetailController].DoSaveWithdraw err:%s", err)
			}
		}
		logs.Info("startMill:", startMill)

		// 传入 2018-03-22 的时间戳 会统计 2018-03-21 的数据
		thirdparty.BusinessDetailStatistic(startMill + tools.MILLSSECONDADAY)
		startMill += tools.MILLSSECONDADAY
		// break
	}

	logs.Warn("statistic ok")
	logs.Info("[%s] politeness exit.", procTitle)
}
