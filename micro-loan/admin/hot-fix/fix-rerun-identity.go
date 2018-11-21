package main

import (
	"fmt"
	"time"

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

type UserData struct {
	ID       int64  `orm:"column(id)"`
	Mobile   string `orm:"column(mobile)"`
	Realname string `orm:"column(realname)"`
	Identity string `orm:"column(identity)"`
}

func main() {
	// 设置进程 title
	procTitle := "fix-rerun-identity"
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

	accountBase := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(accountBase.Using())

	for {
		var userData []UserData
		sql := fmt.Sprintf(`SELECT id,realname,identity,mobile,third_id FROM %s WHERE realname<>"" and identity<>"" and mobile<>"" and third_id<>identity AND id>%d ORDER BY id ASC LIMIT 5`,
			accountBase.TableName(), lastID)

		num, err := o.Raw(sql).QueryRows(&userData)

		// logs.Debug("num", num)
		// os.Exit(0)

		if err != nil || num <= 0 {
			logs.Info("[%s] 没有更多待处理数据了...", procTitle)
			break
		}
		for _, userData := range userData {
			lastID = userData.ID
			channelType := tongdun.IDCheckChannelType
			channelCode := tongdun.IDCheckChannelCode
			name := userData.Realname
			identityCode := userData.Identity
			mobile := userData.Mobile
			code, _, err := tongdun.CreateTask(lastID, channelType, channelCode, name, identityCode, mobile)

			if err == nil && code == 0 {

				logs.Debug("[rerun-identity] 创建成功，睡眠2秒")
				time.Sleep(time.Duration(2) * time.Second)

				accountBase, _ := models.OneAccountBaseByPkId(lastID)
				logs.Debug("[rerun-identity] query start ")
				if len(accountBase.Realname) > 0 && len(accountBase.Identity) > 0 {
					tongdunModel, _ := models.GetOneAC(lastID, tongdun.ChannelCodeKTP)
					//如果有任务ID，并且该任务并未被处理 去主动查询同盾接口然后更新

					if tongdunModel.TaskID != "" &&
						tongdunModel.CheckCode == tongdun.IDCheckCodeCreate && //-1
						tongdunModel.IsMatch == tongdun.IsMatchCreateTask { //C
						//查询同盾接口

						logs.Debug("[rerun-identity] AccountID:", tongdunModel.AccountID, "TaskID:", tongdunModel.TaskID)
						idCheckData, err := tongdun.QueryTask(tongdunModel.AccountID, tongdunModel.TaskID)
						if err != nil {
							logs.Debug("[rerun-identity] 身份检查任务查询出现错误:", err)
						} else {

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
							tongdunModel.CreateTime, _ = tools.GetTimeParseWithFormat(idCheckData.Data.CreateTime, "2006-01-02 15:04:05")
							tongdunModel.Source = tongdun.SourceQueryTask

							models.UpdateTongdun(tongdunModel)

							//如果同盾检查未通过 ，则再调用一次Advance
							if idCheckData.Code == tongdun.IDCheckCodeYes { //Y
								//如果同盾识别成功，冗余身份信息到base
								accountBase.ThirdID = tongdunModel.OcrIdentity
								accountBase.ThirdName = tongdunModel.OcrRealName
								service.UpdateAccountBaseByThird(accountBase)

								logs.Debug("[rerun-identity]base冗余更新完毕")
							}
						}
					} else {

						service.IdentityVerify(lastID)
					}

				} else {
					logs.Warning("account_base has no complete realname and identity info", "account id:", lastID)
				}
			} else {
				fmt.Println("任务创建失败", err)
			}
		}
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[%s] politeness exit.", procTitle)
}
