package main

import (
	"fmt"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty/tongdun"
	"micro-loan/common/tools"
	"os"

	"github.com/astaxie/beego/logs"
)

func main() {
	logs.Debug("debug api ...")

	CreateTask()
	// QueryTask()
}

func CreateTask() {
	accountID := int64(180209010000014081)
	channelType := tongdun.IDCheckChannelType
	channelCode := tongdun.IDCheckChannelCode
	name := "EGI INGRINALYS TIARA BAHRI"
	identityCode := "3213086509960001"
	mobile := "081221028250"

	// str := tools.TrimRealName("YOLANDA MEYSHA ZULFILIA  MAHMUDAH")
	// fmt.Println(str)
	// os.Exit(0)

	// tongdunMedel, _ := models.GetOneAC(180523010014564291, tongdun.ChannelCodeKTP)

	// fmt.Println(tongdunMedel)
	// fmt.Println("=====checkcode:", tongdunMedel.CheckCode)
	// fmt.Println("=======ismatch:", tongdunMedel.IsMatch)
	// os.Exit(0)

	code, _, err := tongdun.CreateTask(accountID, channelType, channelCode, name, identityCode, mobile)

	if err == nil && code == 0 {
		fmt.Println("任务创建成功")
	} else {
		fmt.Println("任务创建失败", err)
	}
	// //透传参数
	// passbackParams := tongdun.PassbackParams{}
	// passbackParams.AccountID = accountID
	// passbackParamsJSON, _ := tools.JSONMarshal(passbackParams)

	// params := map[string]interface{}{
	// 	"channel_type":    tongdun.IDCheckChannelType,
	// 	"channel_code":    tongdun.IDCheckChannelCode,
	// 	"real_name":       "EGI INGRINALYS TIARA BAHRI",
	// 	"identity_code":   "3213086509960001",
	// 	"user_mobile":     "081221028250",
	// 	"passback_params": passbackParamsJSON,
	// }

	// _, idCheckData, err := tongdun.Request(accountID, tongdun.CreateTask, params, map[string]interface{}{})

	// // tongdunTask := models.IdentityCheckTask{}
	// // if err := json.Unmarshal([]byte(taskJSON), &tongdunTask); err != nil {
	// // 	panic(err)
	// // }
	// if err != nil {
	// 	panic(err)
	// }

	// if idCheckData.Code == 0 {

	// 	tongdunModel := models.AccountTongdun{}
	// 	tongdunModel.TaskID = idCheckData.TaskID
	// 	tongdunModel.AccountID = accountID
	// 	tongdunModel.OcrRealName = idCheckData.Data.RealName
	// 	tongdunModel.OcrIdentity = idCheckData.Data.IdentityCode
	// 	tongdunModel.Mobile = idCheckData.Data.Mobile
	// 	tongdunModel.CheckCode = tongdun.IDCheckCodeCreate
	// 	tongdunModel.Message = idCheckData.Message
	// 	tongdunModel.IsMatch = tongdun.IsMatchCreateTask
	// 	tongdunModel.ChannelType = idCheckData.Data.ChannelType
	// 	tongdunModel.ChannelCode = idCheckData.Data.ChannelCode
	// 	tongdunModel.ChannelSrc = idCheckData.Data.ChannelSrc
	// 	tongdunModel.ChannelAttr = idCheckData.Data.ChannelAttr
	// 	tongdunModel.CreateTimeS = idCheckData.Data.CreateTime
	// 	tongdunModel.NotifyTimeS = ""
	// 	tongdunModel.CreateTime = tools.GetTimeParseWithFormat(idCheckData.Data.CreateTime, "2006-01-02 15:04:05")
	// 	tongdunModel.NotifyTime = 0
	// 	tongdunModel.Source = tongdun.SourceCreateTask

	// 	id, _ := models.InsertTongdun(tongdunModel)
	// 	fmt.Println("任务创建成功:====ID", id)

	// } else {
	// 	fmt.Println("任务创建失败")
	// }

	// fmt.Print("CODE---", idCheckData.Code)
	// os.Exit(0)

	// // accountTongdun := models.AccountTongdun{}

	// fmt.Print(idCheckData)
}

func QueryTask() {
	accountID := int64(180528010023403932)
	taskID := "TASKKTP107001201805281337531021360317"
	idCheckData, err := tongdun.QueryTask(accountID, taskID)

	if err != nil {
		panic(err)
	} else {

		notifyTime := "2018-05-19 11:11:11"

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
		tongdunModel.NotifyTimeS = notifyTime
		tongdunModel.CreateTime, _ = tools.GetTimeParseWithFormat(idCheckData.Data.CreateTime, "2006-01-02 15:04:05")
		tongdunModel.NotifyTime, _ = tools.GetTimeParseWithFormat(notifyTime, "2006-01-02 15:04:05")
		tongdunModel.Source = tongdun.SourceNotify

		fmt.Print(tongdunModel)
		os.Exit(0)

		models.UpdateTongdun(tongdunModel)
		fmt.Println("更新完毕")

	}

}
