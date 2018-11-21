package main

import (
	// 数据库初始化
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty/bluepay"
	"micro-loan/common/thirdparty/doku"
	"micro-loan/common/thirdparty/xendit"
	"micro-loan/common/tools"
)

func main() {
	procTitle := "fix-bank-info-new"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	// lock
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()
	lockKey := fmt.Sprintf("lock:%s", procTitle)
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}
	defer storageClient.Do("DEL", lockKey)

	// test
	//bank := models.BanksInfo{
	//	Id:                 1,
	//	FullName:           "2",
	//	XenditBrevityName:  "3",
	//	DokuFullName:       "4",
	//	DokuBrevityName:    "5",
	//	DokuBrevityId:      "6",
	//	BluepayBrevityName: "7",
	//	LoanCompanyCode:    8,
	//	RepayCompanyCode:   9,
	//	Ctime:              10,
	//	Utime:              11,
	//}
	//models.OrmInsert(&bank)

	// 开搞
	for key, value := range xendit.BankNameCodeMap() {

		// xendit
		bank := models.BanksInfo{}
		bank.FullName = key
		bank.XenditBrevityName = value

		// doku
		if doukuFullName, ok := doku.GetBankXenditDokuBankMap()[key]; ok {
			bank.DokuFullName = doukuFullName
			bank.DokuBrevityName, err = doku.BankName2Code(doukuFullName)
			if err != nil {
				logs.Error("[BankName2Code] doukuFullName:%s err:%v   key:%s", doukuFullName, err, key)
			}
			if id, ok := doku.GetBankXenditDokuBandIdMap()[doukuFullName]; ok {
				bank.DokuBrevityId = id
			} else {
				logs.Error("GetBankXenditDokuBandIdMap doukuFullName:%s err:%v key:%s", doukuFullName, err, key)
			}

		} else {
			logs.Info("bank:%s not in doku", key)
		}

		// bluepay
		if blueBrive, ok := bluepay.BluepayBankNameCodeMap()[key]; ok {
			bank.BluepayBrevityName = blueBrive
		} else {
			logs.Info("bank:%s not in blueku", key)
		}

		// insert
		tag := tools.GetUnixMillis()
		bank.Ctime = tag
		bank.Utime = tag
		_, err = models.OrmInsert(&bank)
		if err != nil {
			logs.Error("OrmInsert err:%v bank:%#v", err, bank)
		}
	}

	logs.Warning("[%s] end.", procTitle)
}
