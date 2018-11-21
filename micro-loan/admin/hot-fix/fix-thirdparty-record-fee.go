package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/erikdubbelboer/gspt"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
)

func main() {
	// 设置进程 title
	procTitle := "fix-thirdparty-record-fee"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := fmt.Sprintf("lock:%s", procTitle)
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}

	var lastID int64
	thirdpartyRecord := models.ThirdpartyRecord{}
	ormRecord := orm.NewOrm()
	ormRecord.Using(thirdpartyRecord.Using())

	thirdpartyInfo := models.ThirdpartyInfo{}
	ormInfo := orm.NewOrm()
	ormInfo.Using(thirdpartyInfo.Using())

	modifyNum := 0
	for {
		sql := fmt.Sprintf(`SELECT * FROM %s WHERE id > %d and response_type = 0  ORDER BY id ASC LIMIT 100`,
			thirdpartyRecord.TableName(), lastID)

		var records []models.ThirdpartyRecord
		num, err := ormRecord.Raw(sql).QueryRows(&records)
		if err != nil || num <= 0 {
			logs.Warn("读取记录 err:", err, " sql:", sql)
			logs.Info("[%s] 没有更多待处理数据了...", procTitle)
			break
		}

		for _, record := range records {

			if 0 == modifyNum%1000 {
				logs.Info("我需要休息一下防止搞崩数据库 modifyNum:%d", modifyNum)
				time.Sleep(time.Second)
			}

			lastID = record.Id
			logs.Info("处理该调用记录，id：", lastID)

			//查询此api对应的 thirdpartyInfo 信息
			// if record.Thirdparty != 9 {
			// 	continue
			// }

			//不修复当天的数据
			if record.ResponseType != 0 {
				byet, _ := json.Marshal(records)
				logs.Info("可能是当天数据不修复：%#v", string(byet))
				continue
			}
			responstType, fee := thirdparty.CalcFeeByApi(record.Api, record.Request, record.Response, record.Ctime)
			record.FeeForCall = fee
			record.ResponseType = responstType
			num, err := record.UpdateFee()
			if num < 0 || err != nil {
				jsonStr, _ := json.Marshal(record)
				logs.Error("更新调用费用记录出错。err:", err, " sql:", sql, " record:%#v", string(jsonStr))
				continue
			}
			modifyNum++
		}
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[%s] politeness exit.", procTitle)
}
