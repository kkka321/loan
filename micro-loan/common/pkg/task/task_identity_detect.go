package task

import (
	"encoding/json"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/dao"
	"micro-loan/common/lib/gaws"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/advance"
	"micro-loan/common/thirdparty/faceid"
	"micro-loan/common/thirdparty/tongdun"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// ----------

type IdentityDetectTask struct {
}

func (c *IdentityDetectTask) Start() {
	logs.Info("Do work: TaskIdentityDetect")

	queueName := beego.AppConfig.String("account_identity_detect")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("identity_detect_lock")
	for {
		TaskHeartBeat(storageClient, lockKey)

		qValue, err := storageClient.Do("rpop", queueName)
		if err != nil || qValue == nil {
			logs.Info("queue", queueName, " is empty.")

			// 没有可供处理的数据,休眠半秒钟
			time.Sleep(500 * time.Millisecond)
			continue
		}

		accountId, _ := tools.Str2Int64(string(qValue.([]byte)))

		// 退出进程命令的特殊命令号
		if types.TaskExitCmd == accountId {
			logs.Info("[TaskIdentityDetect] receive exit cmd.")
			break
		}

		addCurrentData(tools.Int642Str(accountId), "accountId")
		DoTaskIdentityDetect(accountId)
		removeCurrentData(tools.Int642Str(accountId))
	}

	logs.Info("TaskIdentityDetect has done.")
}

func (c *IdentityDetectTask) Cancel() {

}

// TODO: 考虑重复操作,一期全部按新数据处理
func DoTaskIdentityDetect(accountId int64) (err error) {

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[DoTaskIdentityDetect] panic accountId:%d, err:%v", accountId, x)
			logs.Error(tools.FullStack())
		}
	}()

	profile, err := dao.CustomerProfile(accountId)
	if err != nil {
		logs.Error("Account does not exist. accountId:", accountId)
		return
	}

	// 日志记录,将结构体转为json
	logJSON, _ := tools.JsonEncode(*profile)

	// 1. 下载对应的图片到本地`/tmp/`下
	if profile.IdPhoto <= 0 || profile.HandHeldIdPhoto <= 0 {
		logs.Warning("User does not upload Identity Photo. AccountProfile:", logJSON)
		return
	}

	idPhotoResource, err1 := service.OneResource(profile.IdPhoto)
	handIdPhotoResource, err2 := service.OneResource(profile.HandHeldIdPhoto)
	if err1 != nil || err2 != nil {
		logs.Warning("The resource does not exist. AccountProfile:", logJSON)
		return
	}
	idPhotoTmp := gaws.BuildTmpFilename(profile.IdPhoto)
	handIdPhotoTmp := gaws.BuildTmpFilename(profile.HandHeldIdPhoto)
	_, err3 := gaws.AwsDownload(idPhotoResource.HashName, idPhotoTmp)
	_, err4 := gaws.AwsDownload(handIdPhotoResource.HashName, handIdPhotoTmp)
	// 删除可能下载到本地的临时文件
	defer tools.Remove(idPhotoTmp)
	defer tools.Remove(handIdPhotoTmp)
	if err3 != nil || err4 != nil {
		logs.Warning("Dowload resource from aws has wrong. idPhotoResource:", idPhotoResource.HashName, ", handIdPhotoResource:", handIdPhotoResource.HashName)

		return
	}

	// 2. 调用 FaceId,确认照片里面有人脸
	//// 2.1 处理 id_photo
	bson, httpCode, err := faceid.Detect(accountId, idPhotoTmp, faceid.MultiOrientedDetectionNo)
	if err != nil || 200 != httpCode {
		logs.Warning("idPhotoResource does not have faces data. AccountProfile:", logJSON, ", idPhotoResource:", idPhotoResource.HashName, "httpCode:", httpCode, "ERROR:", err)
		//tools.Remove(idPhotoTmp)
		//return
	}
	idPhotoApiRes := faceid.ResponseDetect{}
	idPhotoBridge := faceid.ResponseDetectFaces{}
	json.Unmarshal(bson, &idPhotoApiRes)
	if len(idPhotoApiRes.Faces) > 0 {
		idPhotoBridge = idPhotoApiRes.Faces[0]
		profile.SaveIdPhotoDetect(idPhotoBridge.Quality, idPhotoBridge.QualityThreshold)
	}
	//// 2.2 处理 hand_held_id_photo
	bson, httpCode, err = faceid.Detect(accountId, handIdPhotoTmp, faceid.MultiOrientedDetectionNo)
	if err != nil || 200 != httpCode {
		logs.Warning("handIdPhotoResource does not have faces data. AccountProfile:", logJSON, ", handIdPhotoResource:", handIdPhotoResource.HashName, "httpCode:", httpCode, "ERROR:", err)
		//tools.Remove(handIdPhotoTmp)
		//return
	}
	handPhotoApiRes := faceid.ResponseDetect{}
	handPhotoBridge := faceid.ResponseDetectFaces{}
	json.Unmarshal(bson, &handPhotoApiRes)
	if len(handPhotoApiRes.Faces) > 0 {
		handPhotoBridge = handPhotoApiRes.Faces[0]
		profile.SaveHandPhotoDetect(handPhotoBridge.Quality, handPhotoBridge.QualityThreshold)
	}
	if idPhotoBridge.Quality < idPhotoBridge.QualityThreshold || handPhotoBridge.Quality < handPhotoBridge.QualityThreshold {
		logs.Warning("Detect faces has wrong. AccountProfile:", *profile, ", idPhotoBridge:", idPhotoBridge, ", handPhotoBridge:", handPhotoBridge)
		//// 防止写满`/tmp`
		//tools.Remove(idPhotoTmp)
		//tools.Remove(handIdPhotoTmp)
		//return
	}

	// // 3. 如果有人脸,调用advance.ai,进行比对,并回写数据
	// //// 3.1 Face Comparison 人脸比对
	// param := map[string]interface{}{}
	// file := map[string]interface{}{
	// 	"firstImage":  idPhotoTmp,
	// 	"secondImage": handIdPhotoTmp,
	// }
	// _, faceCmpData, err := advance.Request(accountId, advance.ApiFaceComparison, param, file)
	// if advance.IsSuccess(faceCmpData.Code) {
	// 	data := faceCmpData.Data.(map[string]interface{})
	// 	if f, ok := data["similarity"]; ok {
	// 		profile.SaveFaceComparison(f.(float64))
	// 	}
	// }

	// 3. 如果account_base中有 identity和realname,进行身份认证, 先走同盾，如果同盾没有通过咋调用Advance,
	// 同盾数据更新在account_tongdun表，acvance写入到account_base中
	//// 3.1 Face Comparison 人脸比对
	accountBase, _ := models.OneAccountBaseByPkId(accountId)
	logs.Debug("[tongdun IdentityCheck] start ")
	if len(accountBase.Realname) > 0 && len(accountBase.Identity) > 0 {
		tongdunModel, _ := models.GetOneAC(accountId, tongdun.ChannelCodeKTP)
		//如果有任务ID，并且该任务并未被处理 去主动查询同盾接口然后更新

		if tongdunModel.TaskID != "" &&
			tongdunModel.CheckCode == tongdun.IDCheckCodeCreate && //-1
			tongdunModel.IsMatch == tongdun.IsMatchCreateTask { //C
			//查询同盾接口

			logs.Debug("[tongdun IdentityCheck] AccountID:", tongdunModel.AccountID, "TaskID:", tongdunModel.TaskID)
			idCheckData, err := tongdun.QueryTask(tongdunModel.AccountID, tongdunModel.TaskID)
			if err != nil {
				logs.Debug("[Tongdun IdentityCheck] 身份检查任务查询出现错误:", err)
			} else {
				logs.Debug("[tongdun IdentityCheck]  QueryTask idCheckData:%#v", idCheckData)
				tongdunModel.TaskID = idCheckData.TaskID
				tongdunModel.OcrRealName = idCheckData.Data.RealName
				tongdunModel.OcrIdentity = idCheckData.Data.IdentityCode
				tongdunModel.Mobile = idCheckData.Data.Mobile
				tongdunModel.CheckCode = idCheckData.Code
				tongdunModel.Message = idCheckData.Message
				tongdunModel.IsMatch = service.ParseIsMatch(idCheckData.Data.TaskData)
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
					// 用户激活事件触发
					event.Trigger(&evtypes.UserActiveEv{
						AccountID: accountId,
						Time:      tools.GetUnixMillis(),
					})
				}
			}
		}
	} else {
		logs.Warning("account_base has no complete realname and identity info", "account id:", accountId)
	}

	//身份检查如果已经通过，则不再调用advanceAI  KTP
	if !service.IdentityVerify(accountId) {
		//兼容老版本
		//Advance 身份检查
		idCheckData, err := advance.IdentiryCheck(accountId, accountBase.Realname, accountBase.Identity)
		logs.Debug("[advcance IdentityCheck] 接口已触发于兼容老版本 ")
		updateBaseByAdvanceIDcheckResult(accountId, idCheckData, err)
	}
	// 3.2 ID Holding Photo Check 手持识别=====身份检查V3版本 把手持照片比对结果放到API端同步处理
	fileHC := map[string]interface{}{
		"idHoldingImage": handIdPhotoTmp,
	}
	_, faceHoldData, err := advance.Request(accountId, advance.ApiIDCheck, map[string]interface{}{}, fileHC)
	if advance.IsSuccess(faceHoldData.Code) {
		profile.SaveHoldCheck(faceHoldData.Data.Similarity)
	}

	// 4. 删除临时文件 TODO: 是不是可以用defer简化?
	// tools.Remove(idPhotoTmp)
	// tools.Remove(handIdPhotoTmp)
	return
}

// updateBaseByAdvanceIDcheckResult  抽出来公共部分
func updateBaseByAdvanceIDcheckResult(accountID int64, idCheckData advance.ResponseData, err error) {
	if err != nil {
		logs.Error("[updateBaseByAdvanceIDcheckResult] error accountid:%d, err:%v", accountID, err)
		return
	}

	if advance.IsSuccess(idCheckData.Code) {
		logs.Debug("[updateBaseByAdvanceIDcheckResult] 接口获取数据: ", idCheckData.Data)
		accountBaseM := models.AccountBase{Id: accountID, ThirdID: idCheckData.Data.IDNumber}
		accountBaseM.ThirdName = idCheckData.Data.Name
		accountBaseM.ThirdProvince = idCheckData.Data.Province
		accountBaseM.ThirdCity = idCheckData.Data.City
		accountBaseM.ThirdDistrict = idCheckData.Data.District
		accountBaseM.ThirdVillage = idCheckData.Data.Village
		_, err := service.UpdateAccountBaseByThird(accountBaseM)
		if err != nil {
			logs.Error("[updateBaseByAdvanceIDcheckResult] account base Not success update accountId:%d, err:%v", accountID, err)
		}
		// 用户激活事件触发
		event.Trigger(&evtypes.UserActiveEv{
			AccountID: accountID,
			Time:      tools.GetUnixMillis(),
		})
	} else {
		logs.Warn("[updateBaseByAdvanceIDcheckResult] advance fail accountid:%d, data:%v", accountID, idCheckData)
	}

}
