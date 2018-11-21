package main

import (
	"fmt"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/tongdun"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/erikdubbelboer/gspt"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
)

type TongdunData struct {
	ID        int64  `orm:"column(id)"`
	TaskID    string `orm:"column(task_id)"`
	AccountID int64  `orm:"column(account_id)"`
}

func main() {
	// 设置进程 title
	procTitle := "fix-rerun-tongdun"
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

	accountTongdun := models.AccountTongdun{}
	o := orm.NewOrm()
	o.Using(accountTongdun.Using())

	for {
		var tongdunData []TongdunData
		sql := fmt.Sprintf(`SELECT id,task_id,account_id FROM %s WHERE check_code !=0 and channel_type="KTP" AND id>%d ORDER BY id ASC LIMIT 5`,
			accountTongdun.TableName(), lastID)
		num, err := o.Raw(sql).QueryRows(&tongdunData)
		if err != nil || num <= 0 {
			logs.Info("[%s] 没有更多待处理数据了...", procTitle)
			break
		}

		for _, tongdunData := range tongdunData {

			lastID = tongdunData.ID

			idCheckData, err := tongdun.QueryTask(tongdunData.AccountID, tongdunData.TaskID)

			if err != nil {

				logs.Error("[fix-rerun-tongdun] err:", err)
			} else {
				timestamp, _ := tools.GetTimeParseWithFormat(idCheckData.Data.CreateTime, "2006-01-02 15:04:05")
				tongdunModel, _ := models.GetOneByCondition("task_id", idCheckData.TaskID)
				tongdunModel.TaskID = idCheckData.TaskID
				tongdunModel.OcrRealName = idCheckData.Data.RealName
				tongdunModel.OcrIdentity = idCheckData.Data.IdentityCode
				tongdunModel.Mobile = idCheckData.Data.Mobile
				tongdunModel.CheckCode = idCheckData.Code
				tongdunModel.Message = idCheckData.Message
				tongdunModel.IsMatch = idCheckData.Data.TaskData.ReturnInfo.IsMatch
				tongdunModel.ChannelType = idCheckData.Data.ChannelType
				tongdunModel.ChannelCode = idCheckData.Data.ChannelCode
				tongdunModel.ChannelSrc = idCheckData.Data.ChannelSrc
				tongdunModel.ChannelAttr = idCheckData.Data.ChannelAttr
				tongdunModel.CreateTimeS = idCheckData.Data.CreateTime
				tongdunModel.NotifyTimeS = ""
				tongdunModel.CreateTime = timestamp
				tongdunModel.NotifyTime = 0
				tongdunModel.Source = tongdun.SourceReRun

				models.UpdateTongdun(tongdunModel)

				logs.Debug("[fix-rerun-tongdun]更新完毕")

				accountBase, _ := models.OneAccountBaseByPkId(tongdunData.AccountID)
				//同盾命中，匹配
				if idCheckData.Code == tongdun.IDCheckCodeYes { //Y

					//如果同盾识别成功，冗余身份信息到base
					accountBase.ThirdID = idCheckData.Data.IdentityCode
					accountBase.ThirdName = idCheckData.Data.RealName

					logs.Debug("[fix-rerun-tongdun] accountBase ：", accountBase)

					service.UpdateAccountBaseByThird(accountBase)
					logs.Debug("[fix-rerun-tongdun]冗余字段更新完毕")
				} else {

					logs.Error("[fix-rerun-tongdun]没有更新，checkcode：", idCheckData.Code, "is_match:", idCheckData.Data.TaskData.ReturnInfo.IsMatch)
				}

			}

		}
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[%s] politeness exit.", procTitle)
}
