package main

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/erikdubbelboer/gspt"

	// 数据库初始化
	"micro-loan/common/dao"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/service"
	credit "micro-loan/common/thirdparty/credit_increase"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

//var period = int64(60)
var MaxInterval = tools.MILLSSECONDADAY * 60

func get100Ids(lastedId int64) (aExts []models.AccountBaseExt) {
	accountExt := models.AccountBaseExt{}
	o := orm.NewOrm()
	o.Using(accountExt.UsingSlave())

	o.QueryTable(accountExt.TableName()).
		Filter("account_id__gt", lastedId).
		OrderBy("account_id").
		Limit(100).
		All(&aExts)
	return

}

func getValueByColName(aExt models.AccountBaseExt, colName string) reflect.Value {
	v := reflect.ValueOf(&aExt)
	v = v.Elem()

	col := v.FieldByName(colName)
	return col
}

func fixOneAccount(i int, group *sync.WaitGroup, aExt models.AccountBaseExt) {
	defer group.Done()

	logs.Info("[%d]  %d", i, aExt.AccountId)

	// 不是复贷直接跳过
	reloan := dao.IsRepeatLoan(aExt.AccountId)
	if !reloan {
		return
	}

	now := tools.GetUnixMillis()
	for code, aInfo := range credit.BackendCodeMap() {
		aExt, _ = models.OneAccountBaseExtByPkId(aExt.AccountId)
		status := getValueByColName(aExt, aInfo.StatusColName).Int()
		if status == types.AuthorizeStatusCrawleSuccess {
			continue
		}

		period, _ := credit.AuthorizeValidityPeriod(aInfo.BackendCode, reloan)

		if code != credit.BackendCodeNpwp {
			for _, v := range aInfo.TongdunChannelCodes {
				// 取最新一条记录
				at, _ := models.GetLatestSuccessACByChannelCode(aExt.AccountId, v)

				if at.AccountID == aExt.AccountId &&
					(now-at.CreateTime*1000) < int64(period)*tools.MILLSSECONDADAY {
					service.IncreaseCreditByAuthoriation(at, at.NotifyTime*1000)
				}
			}

		} else {

			if aExt.NpwpStatus == types.AuthorizeStatusSuccess &&
				(now-aExt.NpwpTime) < int64(period)*tools.MILLSSECONDADAY {
				service.IncreaseCreditByAuthoriation4Npwp(aExt, aExt.NpwpTime)
			}
		}
	}
}

func main() {
	// 设置进程 title
	// +1 分布式锁
	// -1 正常退出时,释放锁
	procTitle := "fix_increase_credit"
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

	lastedId := int64(0)
	wg := sync.WaitGroup{}
	for {
		ids := get100Ids(lastedId)
		if len(ids) == 0 {
			logs.Warn("no data to handel")
			break
		}

		// 3个携程同步 速度取决于最慢的那个
		for i := 0; i < len(ids); {
			if i < len(ids) {
				vi := ids[i]
				wg.Add(1)
				go fixOneAccount(0, &wg, vi)
				i++
				lastedId = vi.AccountId
			}

			if i < len(ids) {
				vi := ids[i]
				wg.Add(1)
				go fixOneAccount(1, &wg, vi)
				i++
				lastedId = vi.AccountId
			}

			if i < len(ids) {
				vi := ids[i]
				wg.Add(1)
				go fixOneAccount(1, &wg, vi)
				i++
				lastedId = vi.AccountId
			}
			wg.Wait()
		}
	}

	logs.Notice("ok")

}

//func init() {
//	period, _ = config.ValidItemInt64("additional_authorize_validity_period_reloan")
//}
