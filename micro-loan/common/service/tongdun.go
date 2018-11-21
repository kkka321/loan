package service

import (
	"encoding/json"
	"reflect"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/pkg/schema_task"
	"micro-loan/common/thirdparty/credit_increase"
	"micro-loan/common/thirdparty/tongdun"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func HandleTongdunNormalCallback(idCheckData *tongdun.IdentityCheckCreateTask, accountID int64, notifyTime string) {
	channelCode := idCheckData.Data.ChannelCode
	//accountID := passbackParamsData.AccountID
	tongdunModel, _ := models.GetOneAC(accountID, channelCode)

	//数据基本判断
	if idCheckData.TaskID != tongdunModel.TaskID {
		logs.Error("[HandleTongdunNormalCallback] 异步数据与数据库TaskID不一致", "数据库中:", tongdunModel.TaskID, "异步数通知:", idCheckData.TaskID)
	}

	//如果有任务ID，并且该任务并未被处理 去主动查询同盾接口然后更新
	if idCheckData.TaskID == tongdunModel.TaskID &&
		((tongdunModel.CheckCode != tongdun.IDCheckCodeYes && tongdunModel.CheckCode != tongdun.IDCheckCodeNo && tongdunModel.ChannelType == tongdun.IDCheckChannelType) ||
			tongdunModel.ChannelType != tongdun.IDCheckChannelType) {

		tongdunModel.TaskID = idCheckData.TaskID
		tongdunModel.OcrRealName = idCheckData.Data.RealName
		tongdunModel.OcrIdentity = idCheckData.Data.IdentityCode
		tongdunModel.Mobile = idCheckData.Data.Mobile
		tongdunModel.CheckCode = idCheckData.Code
		tongdunModel.Message = idCheckData.Message
		tongdunModel.IsMatch = ParseIsMatch(idCheckData.Data.TaskData)
		tongdunModel.ChannelType = idCheckData.Data.ChannelType
		tongdunModel.ChannelCode = idCheckData.Data.ChannelCode
		tongdunModel.ChannelSrc = idCheckData.Data.ChannelSrc
		tongdunModel.ChannelAttr = idCheckData.Data.ChannelAttr
		tongdunModel.CreateTimeS = idCheckData.Data.CreateTime
		tongdunModel.NotifyTimeS = notifyTime
		tongdunModel.CreateTime, _ = tools.GetTimeParseWithFormat(idCheckData.Data.CreateTime, "2006-01-02 15:04:05")
		tongdunModel.NotifyTime, _ = tools.GetTimeParseWithFormat(notifyTime, "2006-01-02 15:04:05")
		tongdunModel.Source = tongdun.SourceNotify
		jsonData, e := json.Marshal(idCheckData.Data.TaskData)
		if e == nil {
			tongdunModel.TaskData = string(jsonData)
		} else {
			logs.Warn("[HandleTongdunNormalCallback] callback Marshal taskData :%s data:%#v", e, idCheckData.Data.TaskData)
		}
		models.UpdateTongdun(tongdunModel)

		if tongdunModel.ChannelType == tongdun.IDGoJekChannelType && tongdunModel.ChannelCode == tongdun.ChannelCodeGoJek {
			UpdateGojekMark(accountID, tongdunModel.TaskData)
		}
		go IncreaseCreditByAuthoriation(tongdunModel, tools.GetUnixMillis())
	} else {
		logs.Warn("[TongdunCallback] 数据已被更新", "DB_CODE:", tongdunModel.CheckCode, "DB_MATCH:", tongdunModel.IsMatch, "更新源:", tongdunModel.Source)
	}
}

func HandleTongdunSocialCallback(idCheckData *tongdun.IdentityCheckCreateTask, accountID int64, notifyTime string) {
	tongdunModel, _ := models.GetOneByCondition("task_id", idCheckData.TaskID)
	tongdunModel.AccountID = accountID
	tongdunModel.TaskID = idCheckData.TaskID
	tongdunModel.OcrRealName = idCheckData.Data.RealName
	tongdunModel.OcrIdentity = idCheckData.Data.IdentityCode
	tongdunModel.Mobile = idCheckData.Data.Mobile
	tongdunModel.CheckCode = idCheckData.Code
	tongdunModel.Message = idCheckData.Message
	tongdunModel.ChannelType = idCheckData.Data.ChannelType
	tongdunModel.ChannelCode = idCheckData.Data.ChannelCode
	tongdunModel.ChannelSrc = idCheckData.Data.ChannelSrc
	tongdunModel.ChannelAttr = idCheckData.Data.ChannelAttr
	tongdunModel.CreateTimeS = idCheckData.Data.CreateTime
	tongdunModel.NotifyTimeS = notifyTime
	tongdunModel.CreateTime, _ = tools.GetTimeParseWithFormat(idCheckData.Data.CreateTime, "2006-01-02 15:04:05")
	tongdunModel.NotifyTime, _ = tools.GetTimeParseWithFormat(notifyTime, "2006-01-02 15:04:05")
	tongdunModel.Source = tongdun.SourceNotify
	jsonData, e := json.Marshal(idCheckData.Data.TaskData)
	if e == nil {
		tongdunModel.TaskData = string(jsonData)
	} else {
		logs.Warn("[HandleTongdunNormalCallback] callback Marshal taskData :%s data:%#v", e, idCheckData.Data.TaskData)
	}

	_, err := dao.InsertOrUpdateTongdunManual(tongdunModel)
	if err != nil {
		logs.Error("[HandleTongdunNormalCallback] InsertOrUpdateTongdunManual err:%v tongdunModel:%#v", err, tongdunModel)
	}

	if tongdunModel.CheckCode == tongdun.IDInputOk {
		// 认证完成 ,等待爬取
		SaveAuthorizeResult(accountID, idCheckData.Data.ChannelCode, types.AuthorizeStatusSuccess)
	}

	if tongdunModel.TaskData != "" &&
		tongdunModel.TaskData != "null" &&
		tongdunModel.CheckCode == 0 {
		go IncreaseCreditByAuthoriation(tongdunModel, tools.GetUnixMillis())
	}
}

// app透传的参数是 手机号
func RepirePassParams(params *tongdun.PassbackParams) {
	if params.AccountID > 0 {
		return
	}

	if len(params.Mobile) == 0 {
		logs.Warn("[RepirePassParams] both zero. params:%#v", params)
		return
	}

	account, _ := models.OneAccountBaseByMobile(params.Mobile)
	params.AccountID = account.Id
}

func ParseIsMatch(data interface{}) (match string) {
	if data == nil {
		logs.Info("[parseIsMatch] data nil")
		return ""
	}
	strData, _ := json.Marshal(data)
	taskData := tongdun.TaskData{}
	json.Unmarshal(strData, &taskData)

	return taskData.ReturnInfo.IsMatch
}

// 修改逻辑后 此函数
func SaveAuthorizeResult(accountId int64, channelCode string, status int) {
	accountExt, _ := models.OneAccountBaseExtByPkId(accountId)

	timestamp := tools.GetUnixMillis()
	switch channelCode {
	case tongdun.ChannelCodeTelkomsel, tongdun.ChannelCodeXI, tongdun.ChannelCodeIndosat:
		{
			if accountExt.AuthorizeStatusYys == status {
				return
			}
			accountExt.AuthorizeFinishTimeYys = timestamp
			accountExt.AuthorizeStatusYys = status
		}
	case tongdun.ChannelCodeGoJek:
		{
			if accountExt.AuthorizeStatusGoJek == status {
				return
			}
			accountExt.AuthorizeFinishTimeGoJek = timestamp
			accountExt.AuthorizeStatusGoJek = status
		}

	case tongdun.ChannelCodeLazada:
		{
			if accountExt.AuthorizeStatusLazada == status {
				return
			}
			accountExt.AuthorizeFinishTimeLazada = timestamp
			accountExt.AuthorizeStatusLazada = status
		}

	case tongdun.ChannelCodeTokopedia:
		{
			if accountExt.AuthorizeStatusTokopedia == status {
				return
			}
			accountExt.AuthorizeFinishTimeTokopedia = timestamp
			accountExt.AuthorizeStatusTokopedia = status
		}

	case tongdun.ChannelCodeFacebook:
		{
			if accountExt.AuthorizeStatusFacebook == status {
				return
			}
			accountExt.AuthorizeFinishTimeFacebook = timestamp
			accountExt.AuthorizeStatusFacebook = status
		}

	case tongdun.ChannelCodeInstagram:
		{
			if accountExt.AuthorizeStatusInstagram == status {
				return
			}
			accountExt.AuthorizeFinishTimeInstagram = timestamp
			accountExt.AuthorizeStatusInstagram = status
		}
	case tongdun.ChannelCodeLinkedin:
		{
			if accountExt.AuthorizeStatusLinkedin == status {
				return
			}
			accountExt.AuthorizeFinishTimeLinkedin = timestamp
			accountExt.AuthorizeStatusLinkedin = status
		}
	default:
		{
			logs.Error("[SaveAuthorizeResult] unknow channel_code:%s accountExt:%#v  accountId:%d status:%d", channelCode, accountExt, accountId, status)
			return
		}
	}
	accountExt.Utime = timestamp

	if accountExt.AccountId == 0 {
		accountExt.AccountId = accountId
		accountExt.Ctime = timestamp
		_, err := models.OrmInsert(&accountExt)
		if err != nil {
			logs.Error("[SaveAuthorizeResult] OrmInsert err:%v accountExt:%#v channel_code:%s accountId:%d status:%d", err, accountExt, channelCode, accountId, status)
		}
	} else {
		_, err := models.OrmAllUpdate(&accountExt)
		if err != nil {
			logs.Error("[SaveAuthorizeResult] OrmAllUpdate err:%v accountExt:%#v channel_code:%s accountId:%d status:%d", err, accountExt, channelCode, accountId, status)
		}
	}

	// 记录日志
	models.InsertLogAccountBaseExt(accountExt)
	return
}

// 上一次查询最大值
func AccountList4ReviewAuthoriation(lastedAccountId int64) (list []models.AccountBaseExt, err error) {
	order := models.AccountBaseExt{}

	o := orm.NewOrm()
	o.Using(order.UsingSlave())

	cond := orm.NewCondition()
	cond = cond.And("account_id__gt", lastedAccountId)

	_, err = o.QueryTable(order.TableName()).
		SetCond(cond).
		OrderBy("account_id").
		Limit(100).
		All(&list)
	return
}

/*
stName,  状态字段名
ctName,  爬取成功时间字段名
qrName   提升额度字段名
*/
func checkAuthorOneExpired(aExt models.AccountBaseExt, stName, ctName, qrName string, period int) (retExt models.AccountBaseExt, expired bool) {
	st := getValueByColName(aExt, stName)
	ct := getValueByColName(aExt, ctName)
	qr := getValueByColName(aExt, qrName)

	if !st.IsValid() || !ct.IsValid() || !qr.IsValid() {
		logs.Error("[checkAuthorOneExpired] reflect err. aExt:%#v colName:%s", aExt)
		return aExt, false
	}

	status := st.Int()
	cwTime := ct.Int()
	if int(status) != types.AuthorizeStatusCrawleSuccess {
		logs.Info("[checkAuthorOneExpired] stName:%s no need to  change.", stName)
		return aExt, false
	}

	now := tools.GetUnixMillis()
	diff := (now - cwTime) / tools.MILLSSECONDADAY
	if diff < int64(period) {
		return aExt, false
	}

	// 已过期
	aExt = changeValueByColName(aExt, stName, types.AuthorizeStatusExpired)
	aExt = changeValueByColName(aExt, qrName, int64(0))
	return aExt, true
}

func CheckAuthoriationStatus(accountId int64) {
	aExt, err := models.OneAccountBaseExtByPkId(accountId)
	if err != nil {
		logs.Error("[CheckAuthoriationStatus] OneAccountBaseExtByPkId id:%d err:%v", accountId, err)
		return
	}

	isReloan := dao.IsRepeatLoan(accountId)
	hasExpird := false
	expired := false

	// 校验各个授信状态
	for _, v := range credit.BackendCodeMap() {
		period, _ := credit.AuthorizeValidityPeriod(v.BackendCode, isReloan)
		logs.Info("[CheckAuthoriationStatus] %s check:%s period:%d", v.IndonesiaName, v.BackendCode, period)

		aExt, expired = checkAuthorOneExpired(aExt,
			v.StatusColName,
			v.CrawTimeColName,
			v.QuotaColName,
			period)

		if expired {
			hasExpird = true
		}
	}

	if hasExpird {
		models.InsertLogAccountBaseExt(aExt)
		models.OrmAllUpdate(&aExt)
	}
}

// 增加同盾授信提升额度
func IncreaseCreditByAuthoriation(accountTongdun models.AccountTongdun, callBackTime int64) {
	if accountTongdun.ChannelType == tongdun.IDCheckChannelType &&
		accountTongdun.ChannelCode == tongdun.ChannelCodeKTP {
		return
	}

	aExt, err := models.OneAccountBaseExtByPkId(accountTongdun.AccountID)
	if err != nil {
		logs.Error("[IncreaseCreditByAuthoriation]  OneAccountBaseExtByPkId err:%v aExt:%#v accountTongdun:%#v", err, aExt, accountTongdun)
		return
	}

	authInfo, ok := credit.AuthorInfoByTongdunChannelCode(accountTongdun.ChannelCode)
	if !ok {
		logs.Error("[changeAuthorStatusToCrawleSuccess] channelcode:%s accountID:%d", accountTongdun.ChannelCode, accountTongdun.AccountID)
		return
	}

	//1 状态更新为抓取成功
	aExt = changeAuthorStatusToCrawleSuccess(aExt, authInfo, callBackTime)

	reloan := dao.IsRepeatLoan(accountTongdun.AccountID)
	if !reloan {
		//首贷不提额，只需要把状态改为抓取成功
		return
	}

	//2 调用接口去提额
	r := credit.NewRequestIncreaseCreditByTongdunModel(accountTongdun)
	ret := credit.GetIncreaseCredit(r)
	if ret.Code == 0 {
		//接口出问题 直接使用默认配置
		ret = credit.DefaultRet(aExt.AccountId)
	}

	if ret.Code != 200 {
		logs.Error("[IncreaseCreditByAuthoriation] do not increase. aId:%d req:%#v ret:%#v", aExt.AccountId, r, ret)
		return
	}

	increaseByRespons(ret, authInfo.RespondColName, aExt, authInfo.QuotaColName, authInfo.BackendCode, authInfo.IndonesiaName)
	return
}

func IncreaseCreditByAuthoriation4Npwp(aExt models.AccountBaseExt, callBackTime int64) {

	logs.Info("[IncreaseCreditByAuthoriation4Npwp] aExt:%#v", aExt)

	authInfo, _ := credit.AuthorInfoByBackendCode(credit.BackendCodeNpwp)

	//1 状态更新为抓取成功
	aExt = changeAuthorStatusToCrawleSuccess(aExt, authInfo, callBackTime)

	reloan := dao.IsRepeatLoan(aExt.AccountId)
	if !reloan {
		//首贷不提额，只需要把状态改为抓取成功
		return
	}

	//2 调用接口去提额
	r := credit.NewRequestIncreaseCreditByTNpwp(aExt)
	ret := credit.GetIncreaseCredit(r)
	if ret.Code == 0 {
		//接口出问题 直接使用默认配置
		ret = credit.DefaultRet(aExt.AccountId)
	}

	if ret.Code != 200 {
		logs.Error("[IncreaseCreditByAuthoriation4Npwp] do not increase. aId:%d req:%#v ret:%#v", aExt.AccountId, r, ret)
		return
	}

	increaseByRespons(ret, authInfo.RespondColName, aExt, authInfo.QuotaColName, authInfo.BackendCode, authInfo.IndonesiaName)
	return
}

// 将状态变为爬取完成
func changeAuthorStatusToCrawleSuccess(aExt models.AccountBaseExt, auth credit.AuthorInfo, successTime int64) models.AccountBaseExt {
	status := getValueByColName(aExt, auth.StatusColName).Int()
	if status == types.AuthorizeStatusCrawleSuccess {
		logs.Warn("[changeAuthorStatusToCrawleSuccess] no need to change. backendCode:%s accoundId:%d", auth.BackendCode, aExt.AccountId)
		return aExt
	}
	aExt = changeAuthorStatusToCrawleSuccessByFiledName(aExt, auth.StatusColName, auth.CrawTimeColName, successTime)
	return aExt
}

/*
stName,  状态字段名
ctName,  爬取成功时间字段名
*/
func changeAuthorStatusToCrawleSuccessByFiledName(aExt models.AccountBaseExt, stName, ctName string, successTime int64) models.AccountBaseExt {
	aExt = changeValueByColName(aExt, ctName, successTime)
	aExt = changeValueByColName(aExt, stName, types.AuthorizeStatusCrawleSuccess)

	models.OrmAllUpdate(&aExt)
	models.InsertLogAccountBaseExt(aExt)
	return aExt
}

/*
retQuotaName 返回额度的字段名
aQuotaName   aExt 额度的字段名
*/
func increaseByRespons(ret credit.RespondIncreaseCredit, retQuota string, aExt models.AccountBaseExt, aQuotaName string, backendCode string, sendName string) {

	accountId := aExt.AccountId
	//根据后台配置决定是否提额
	raiseQuota := int64(0)
	reloan := dao.IsRepeatLoan(accountId)
	_, isCatch := credit.AuthorizeValidityPeriod(backendCode, reloan)

	if isCatch == 1 {
		raiseQuotaNew := getValueByColNameV2(&ret.Data, retQuota).Int()
		raiseQuotaOld := getValueByColName(aExt, aQuotaName).Int()

		if raiseQuotaNew != raiseQuotaOld &&
			raiseQuotaNew != 0 {
			raiseQuota = raiseQuotaNew
			aExt = changeValueByColName(aExt, aQuotaName, raiseQuota)
		}
	} else {
		logs.Warn("[IncreaseByTongdunRespons] backendCode:%s no increase .", backendCode)
		return
	}

	if raiseQuota > 0 {
		models.OrmAllUpdate(&aExt)
		models.InsertLogAccountBaseExt(aExt)

		if reloan {
			schema_task.PushBusinessMsg(types.PushTargetReCreditIncrease, accountId)
		} else {
			schema_task.PushBusinessMsg(types.PushTargetCreditIncrease, accountId)
		}
	}
}

// 首贷用户结清订单时, 将他所有的 抓取成功状态置为过期
func ExpireAllAuthorStatus(accountId int64) {
	aExt, err := models.OneAccountBaseExtByPkId(accountId)
	if err != nil {
		logs.Error("[ExpireAllAuthorStatus]  OneAccountBaseExtByPkId err:%v accountId:%v", err, accountId)
		return
	}

	hasExpired := false
	expired := false
	for _, aInfo := range credit.BackendCodeMap() {

		aExt, expired = checkAuthorOneExpired(aExt,
			aInfo.StatusColName,
			aInfo.CrawTimeColName,
			aInfo.QuotaColName,
			0)

		if expired {
			hasExpired = true
		}
	}

	if hasExpired {
		models.InsertLogAccountBaseExt(aExt)
		models.OrmAllUpdate(&aExt)
	}

	return
}

// 通过反射更改结构体的值
func changeValueByColName(aExt models.AccountBaseExt, colName string, dstValue interface{}) models.AccountBaseExt {
	v := reflect.ValueOf(&aExt)
	v = v.Elem()

	col := v.FieldByName(colName)
	if !col.IsValid() {
		logs.Error("[changeValueByColName] reflect err. aExt:%#v colName:%s", aExt, colName)
		return aExt
	}

	col.Set(reflect.ValueOf(dstValue))
	return aExt
}

func getValueByColName(aExt models.AccountBaseExt, colName string) reflect.Value {
	v := reflect.ValueOf(&aExt)
	v = v.Elem()

	col := v.FieldByName(colName)
	return col
}

func getValueByColNameV2(struc interface{}, colName string) reflect.Value {
	v := reflect.ValueOf(struc)
	v = v.Elem()

	col := v.FieldByName(colName)
	return col
}
