package service

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/cerror"
	"micro-loan/common/dao"
	"micro-loan/common/i18n"
	"micro-loan/common/lib/device"
	"micro-loan/common/lib/gaws"
	"micro-loan/common/lib/redis/cache"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/accesstoken"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/strategy/limit"
	"micro-loan/common/thirdparty/advance"
	"micro-loan/common/thirdparty/api253"
	"micro-loan/common/thirdparty/credit_increase"
	"micro-loan/common/thirdparty/tongdun"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// api 层的特有方法 {
func ApiDataAddEAccountNumber(accountId int64, data map[string]interface{}) {
	eAccountDesc := ""

	_, eAccountDesc = DisplayVAInfoV2(accountId)

	data["e_account_number"] = eAccountDesc
}

func ApiDataAddCurrentLoanInfo(accountId int64, data map[string]interface{}) {
	orderData, err := dao.AccountLastLoanOrder(accountId)
	if err != nil {
		data["amount"] = 0
		data["remaining_days"] = 0
		return
	}

	repayPlan, err := models.GetLastRepayPlanByOrderid(orderData.Id)
	if err != nil {
		data["amount"] = 0
		data["remaining_days"] = 0
		logs.Notice("[ApiDataAddCurrentLoanInfo] models.GetLastRepayPlanByOrderid no data, orderID:", orderData.Id, ", err:", err)
		return
	}

	amount, _ := repayplan.CaculateRepayTotalAmountByRepayPlan(repayPlan)
	remainingDays := (repayPlan.RepayDate - tools.NaturalDay(0)) / (3600000 * 24)
	if remainingDays <= 0 || amount <= 0 {
		remainingDays = 0
	}
	data["remaining_days"] = remainingDays

	data["status"] = orderData.CheckStatus

	if orderData.CheckStatus == types.LoanStatusWaitRepayment || orderData.CheckStatus == types.LoanStatusPartialRepayment {
		data["amount"] = amount
	} else if orderData.CheckStatus == types.LoanStatusOverdue {
		clearReduced, err := repayplan.CaculatePenaltyClearReducedByOrderId(orderData.Id)
		if err == nil { // 结清减免, 并且未生效
			amount = amount - clearReduced
		}
		data["amount"] = amount
	} else if orderData.CheckStatus == types.LoanStatusRolling {
		data["amount"] = orderData.MinRepayAmount
	}

	return
}

func CheckClientInfoRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"os":           true,
		"imei":         true,
		"model":        true,
		"brand":        true,
		"app_version":  true,
		"longitude":    true,
		"latitude":     true,
		"city":         true,
		"time_zone":    true,
		"network":      true,
		"is_simulator": true,
		"platform":     true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckLoginAuthCodeRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"mobile": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckLoginAuthCodeRequiredV2(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"mobile":   true,
		"sms_type": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckVoiceAuthCodeRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"mobile":   true,
		"sms_type": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}
func CheckGetFloatingRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"etype": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckRegisterRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"mobile":    true,
		"auth_code": true,
		"password":  true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckLoginRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"mobile":    true,
		"auth_code": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckSmsLoginRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"mobile":    true,
		"auth_code": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckPwdLoginRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"mobile":   true,
		"password": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckSmsVerifyRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"mobile":    true,
		"auth_code": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckFindPwdRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"mobile":   true,
		"password": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckSetPwdRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"password": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckModifyPwdRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"old_pwd": true,
		"new_pwd": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckWebApiLoginRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"mobile":    true,
		"auth_code": true,
		"channel":   true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckWebApiLoginRequiredV2(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"mobile":    true,
		"auth_code": true,
		"channel":   true,
		"invite":    true,
		"op":        true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckRepeatLoanVerifyRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"auth_code": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckConfirmLoanVerifyRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"auth_code": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckRepeatLoanAllRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"offset": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckIdentityDetectRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"fs1_size": true,
		"fs2_size": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckIdentityDetectRequiredV2(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"fs1_size":  true,
		"fs2_size":  true,
		"is_manual": true,
		"realname":  true,
		"identity":  true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckPaymentVoucherRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"fs1_size": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckAccountVerifyRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"delta":    true,
		"fs1_size": true,
		"fs2_size": true,
		"fs3_size": true,
		"fs4_size": true,
		"fs5_size": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckUpdateBaseRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"gender":   true,
		"realname": true,
		"identity": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckUpdateBaseRequiredV2(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"realname": true,
		"identity": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckUpdateWorkInfoRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"job_type":        true,
		"monthly_income":  true,
		"company_name":    true,
		"company_city":    true,
		"company_address": true,
		"service_years":   true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckUpdateBankInfoRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"bank_name": true,
		"bank_no":   true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckUpdateWorkInfoRequiredV2(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"job_type":          true,
		"monthly_income":    true,
		"company_name":      true,
		"company_city":      true,
		"service_years":     true,
		"company_telephone": true,
		"salary_day":        true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckModifyRepayBankRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"bank_code": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckUpdateContactInfoRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"contact1":      true,
		"contact1_name": true,
		"relationship1": true,
		"contact2":      true,
		"contact2_name": true,
		"relationship2": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckUpdateOtherInfoRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"education":       true,
		"marital_status":  true,
		"children_number": true,
		"bank_name":       true,
		"bank_no":         true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckCreateOrderRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"loan":   true,
		"period": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckConfirmOrderRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"loan":   true,
		"period": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckOperatorAchieveCode(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"channel_type": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckOperatorVerifyCode(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"code":         true,
		"channel_type": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckOperatorPhoneVerifyRecall(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"reverify": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckNpwpVerify(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"npwp_no": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckTongdunInvokeRecord(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"channel_type": true,
		"channel_code": true,
		"return_code":  true,
		"mobile":       true,
		"task_id":      true,
	}
	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckConfirmOrderRequiredV2(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"loan":            true,
		"loan_new":        true,
		"contract_amount": true,
		"period":          true,
		"period_new":      true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

func CheckLoanQuotaRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"loan":   true,
		"period": true,
	}

	return checkRequiredParameter(parameter, requiredParameter)
}

// 用户复贷,查看之前手持证件是否存在
func IfOriginHandHeldIdPhoneExist(accountId int64) (yes bool) {

	accountProfile, _ := dao.GetAccountProfile(accountId)

	if accountProfile.HandHeldIdPhoto > 1 {
		yes = true
	}

	return
}

func RecordClientInfo(info map[string]interface{}) {
	client := models.ClientInfo{
		//Mobile:      info["mobile"].(string),
		ServiceType: info["service_type"].(types.ServiceType),
		RelatedId:   info["related_id"].(int64),
		IP:          info["ip"].(string),
		OS:          info["os"].(string),
		Imei:        info["imei"].(string),
		Model:       info["model"].(string),
		Brand:       info["brand"].(string),
		AppVersion:  info["app_version"].(string),
		Longitude:   info["longitude"].(string),
		Latitude:    info["latitude"].(string),
		City:        info["city"].(string),
		TimeZone:    info["time_zone"].(string),
		Network:     info["network"].(string),
		Platform:    info["platform"].(string),
		Ctime:       tools.GetUnixMillis(),
	}

	isSimulator, _ := tools.Str2Int(info["is_simulator"].(string))
	client.IsSimulator = isSimulator

	// 防止此参数没有传入时,引发内核恐慌...
	if m, ok := info["mobile"]; ok {
		if m != nil {
			client.Mobile = m.(string)
		}
	}
	if appCode, ok := info["app_version_code"]; ok {
		vCode, _ := tools.Str2Int(appCode.(string))
		client.AppVersionCode = vCode
	}
	if uiVersion, ok := info["ui_version"]; ok {
		client.UiVersion = uiVersion.(string)
	}
	if uiCid, ok := info["cid"]; ok {
		client.StemFrom = uiCid.(string)
	}
	if uuid, ok := info["uuid"]; ok {
		v := uuid.(string)
		if len(v) > 0 {
			client.UUID = v
			client.UUIDMd5 = tools.Md5(v)
		}
	}
	if len(client.Imei) > 0 {
		client.ImeiMd5 = tools.Md5(client.Imei)
	}

	o := orm.NewOrm()
	o.Using(client.Using())
	o.Insert(&client)
}

func RecordClientInfoOpenApp(info map[string]interface{}) {
	t := tools.GetUnixMillis()
	client := models.ClientInfoOpenApp{
		IP:         info["ip"].(string),
		OS:         info["os"].(string),
		Imei:       info["imei"].(string),
		Model:      info["model"].(string),
		Brand:      info["brand"].(string),
		AppVersion: info["app_version"].(string),
		Longitude:  info["longitude"].(string),
		Latitude:   info["latitude"].(string),
		City:       info["city"].(string),
		TimeZone:   info["time_zone"].(string),
		Network:    info["network"].(string),
		Platform:   info["platform"].(string),
		FcmToken:   info["fcm_token"].(string),
		Ctime:      t,
		Utime:      t,
	}

	isSimulator, _ := tools.Str2Int(info["is_simulator"].(string))
	client.IsSimulator = isSimulator

	if appCode, ok := info["app_version_code"]; ok {
		vCode, _ := tools.Str2Int(appCode.(string))
		client.AppVersionCode = vCode
	}
	if uiVersion, ok := info["ui_version"]; ok {
		client.UiVersion = uiVersion.(string)
	}
	if uiCid, ok := info["cid"]; ok {
		client.StemFrom = uiCid.(string)
	}
	if uuid, ok := info["uuid"]; ok {
		v := uuid.(string)
		if len(v) > 0 {
			client.UUID = v
			client.UUIDMd5 = tools.Md5(v)
		}
	}
	if len(client.Imei) > 0 {
		client.ImeiMd5 = tools.Md5(client.Imei)
	}

	// 在“打开app”的客户端信息中，判断uuid是否存在
	clientInfoOpenApp, err := models.GetClientInfoOpenAppByUUIDMd5(client.UUIDMd5)
	if err != nil {
		// 在“注册”过的客户端信息中，判断uuid是否存在
		_, err := models.LatestRegisteredClientInfoByUUIDMd5(client.UUIDMd5)
		if err != nil {
			client.Add()
		}
	} else {
		if clientInfoOpenApp.IsRegistered == types.UUIDUnRegistered {
			clientInfoOpenApp.FcmToken = client.FcmToken
			clientInfoOpenApp.Utime = tools.GetUnixMillis()
			clientInfoOpenApp.Updates("fcm_token", "utime")
		}
	}

	return
}

func UpdateClientInfoOpenAppIsRegister(info map[string]interface{}) {

	var uuidMd5 string
	if uuid, ok := info["uuid"]; ok {
		v := uuid.(string)
		if len(v) > 0 {
			uuidMd5 = tools.Md5(v)
		}
	}

	// 在“注册”过的客户端信息中，判断uuid是否存在
	_, err := models.LatestRegisteredClientInfoByUUIDMd5(uuidMd5)
	if err == nil {
		clientInfoOpenApp, err := models.GetClientInfoOpenAppByUUIDMd5(uuidMd5)
		if err == nil && clientInfoOpenApp.IsRegistered == types.UUIDUnRegistered {
			clientInfoOpenApp.IsRegistered = types.UUIDRegistered
			clientInfoOpenApp.Utime = tools.GetUnixMillis()
			clientInfoOpenApp.Updates("is_registered", "utime")
		}
	}

	return
}

func RegisterOrLogin(reqJSON map[string]interface{}) (accountId int64, accessToken string, isNew bool, err error) {

	accountBase, isNew, err := registerHandler(reqJSON)
	if err != nil {
		return
	}

	fcmToken := ""
	if fcmData, ok := reqJSON["fcm_token"]; ok {
		fcmToken = fcmData.(string)
	}
	accountId = accountBase.Id
	accessToken, err = accesstoken.GenTokenWithCache(accountId, reqJSON["platform"].(string), reqJSON["ip"].(string), fcmToken)

	return
}

func RegisterOrLoginV2(reqJSON map[string]interface{}) (accountId int64, accessToken string, isNew bool, err error) {
	accountBase, err := models.OneAccountBaseByMobile(reqJSON["mobile"].(string))
	if err == nil {
		// 手机号已被注册过
		isNew = false
		return
	}
	isNew = true

	// 注册新用户
	bizId, _ := device.GenerateBizId(types.AccountSystem)
	accountBase.Id = bizId
	accountBase.Mobile = reqJSON["mobile"].(string)
	accountBase.Gender = types.GenderSecrecy
	if v, ok := reqJSON["appsflyer_id"]; ok {
		accountBase.AppsflyerID = v.(string)
	}
	if v, ok := reqJSON["google_advertising_id"]; ok {
		accountBase.GoogleAdvertisingID = v.(string)
	}
	if v, ok := reqJSON["channel"]; ok {
		accountBase.Channel = v.(string)
	}
	t := tools.GetUnixMillis()
	accountBase.Status = types.StatusValid
	accountBase.RegisterTime = t
	accountBase.LastLoginTime = t
	accountBase.LatestSmsVerifyTime = t
	accountBase.Password = tools.PasswordEncrypt(reqJSON["password"].(string), accountBase.RegisterTime)
	accountBase.Tags = types.CustomerTagsPotential

	if uiCid, ok := reqJSON["cid"]; ok {
		accountBase.StemFrom = uiCid.(string)
	}

	o := orm.NewOrm()
	o.Using(accountBase.Using())
	_, err = o.Insert(&accountBase)
	if err != nil {
		bson, _ := json.Marshal(accountBase)
		logs.Error("create new account has wrong: ", string(bson), ", err:", err)
		return
	}

	fcmToken := ""
	if fcmData, ok := reqJSON["fcm_token"]; ok {
		fcmToken = fcmData.(string)
	}
	accountId = accountBase.Id
	accessToken, err = accesstoken.GenTokenWithCacheV2(accountId, reqJSON["platform"].(string), reqJSON["ip"].(string), fcmToken)

	return
}

// 手机号未注册过,直接注册
func registerHandler(reqJSON map[string]interface{}) (accountBase models.AccountBase, isNew bool, err error) {
	accountBase, err = models.OneAccountBaseByMobile(reqJSON["mobile"].(string))
	if err != nil { // 未注册过
		bizId, _ := device.GenerateBizId(types.AccountSystem)
		accountBase.Id = bizId
		accountBase.Mobile = reqJSON["mobile"].(string)
		accountBase.Gender = types.GenderSecrecy
		if v, ok := reqJSON["appsflyer_id"]; ok {
			accountBase.AppsflyerID = v.(string)
		}
		if v, ok := reqJSON["google_advertising_id"]; ok {
			accountBase.GoogleAdvertisingID = v.(string)
		}
		if v, ok := reqJSON["channel"]; ok {
			accountBase.Channel = v.(string)
		}

		t := tools.GetUnixMillis()
		accountBase.Status = types.StatusValid
		accountBase.RegisterTime = t
		accountBase.LastLoginTime = t
		accountBase.LatestSmsVerifyTime = t
		accountBase.Tags = types.CustomerTagsPotential

		if uiCid, ok := reqJSON["cid"]; ok {
			accountBase.StemFrom = uiCid.(string)
		}

		o := orm.NewOrm()
		o.Using(accountBase.Using())
		_, err = o.Insert(&accountBase)
		if err != nil {
			bson, _ := json.Marshal(accountBase)
			logs.Error("create new account has wrong: ", string(bson), ", err:", err)
			return
		}

		isNew = true
		// 触发异步追踪事件-注册
		event.Trigger(&evtypes.RegisterTrackEv{
			AccountID:           accountBase.Id,
			StemFrom:            accountBase.StemFrom,
			AppsflyerID:         accountBase.AppsflyerID,
			GoogleAdvertisingID: accountBase.GoogleAdvertisingID,
			Time:                accountBase.RegisterTime,
		})
	}

	return
}

func SmsCodeLogin(reqJSON map[string]interface{}) (accountId int64, accessToken string, isNew, isExistPwd bool, err error) {

	accountBase, isNew, err := registerHandler(reqJSON)
	if err != nil {
		return
	}

	if len(accountBase.Password) > 0 {
		isExistPwd = true
	}

	fcmToken := ""
	if fcmData, ok := reqJSON["fcm_token"]; ok {
		fcmToken = fcmData.(string)
	}
	accountId = accountBase.Id
	accessToken, err = accesstoken.GenTokenWithCacheV2(accountId, reqJSON["platform"].(string), reqJSON["ip"].(string), fcmToken)

	// 清除密码错误次数缓存
	limit.ClearPwdCache(reqJSON["mobile"].(string))

	return
}

func PasswordLogin(reqJSON map[string]interface{}) (accountId int64, accessToken string, isExist bool, errMsg string, err error) {
	mobile := reqJSON["mobile"].(string)
	accountBase, err := models.OneAccountBaseByMobile(mobile)
	if err != nil {
		// 手机号未注册过
		isExist = false
		return
	}
	// 手机号注册过
	isExist = true

	if len(accountBase.Password) <= 0 {
		errMsg = i18n.GetMessageText(i18n.MsgPasswordUnset)
		return
	}
	pwd := reqJSON["password"].(string)
	pwdMd5 := tools.PasswordEncrypt(pwd, accountBase.RegisterTime)
	if pwdMd5 != accountBase.Password {
		errMsg = limit.PasswordStrategy(mobile)
		return
	}

	fcmToken := ""
	if fcmData, ok := reqJSON["fcm_token"]; ok {
		fcmToken = fcmData.(string)
	}
	accountId = accountBase.Id
	accessToken, err = accesstoken.GenTokenWithCacheV2(accountId, reqJSON["platform"].(string), reqJSON["ip"].(string), fcmToken)

	// 清除密码错误次数缓存
	limit.ClearPwdCache(reqJSON["mobile"].(string))

	return
}

func FindPasswordHandler(reqJSON map[string]interface{}) (isExist bool, err error) {
	accountBase, err := models.OneAccountBaseByMobile(reqJSON["mobile"].(string))
	if err != nil {
		// 手机号未被注册过
		isExist = false
		return
	}
	isExist = true

	// 更新账号密码
	accountBase.Password = tools.PasswordEncrypt(reqJSON["password"].(string), accountBase.RegisterTime)

	o := orm.NewOrm()
	o.Using(accountBase.Using())
	_, err = o.Update(&accountBase, "password")
	if err != nil {
		bson, _ := json.Marshal(accountBase)
		logs.Error("Update account has wrong: ", string(bson), ", err:", err)
		return
	}

	// 清除密码错误次数缓存
	limit.ClearPwdCache(reqJSON["mobile"].(string))

	return
}

func SetPasswordHandler(reqJSON map[string]interface{}, accountId int64) (err error) {
	accountBase, err := models.OneAccountBaseByPkId(accountId)
	if err != nil {
		return
	}

	// 更新账号密码
	accountBase.Password = tools.PasswordEncrypt(reqJSON["password"].(string), accountBase.RegisterTime)

	o := orm.NewOrm()
	o.Using(accountBase.Using())
	_, err = o.Update(&accountBase, "password")
	if err != nil {
		bson, _ := json.Marshal(accountBase)
		logs.Error("Update account has wrong: ", string(bson), ", err:", err)
		return
	}

	return
}

func ModifyPasswordHandler(reqJSON map[string]interface{}, accountId int64) (isOldPwdWrong bool, err error) {
	accountBase, err := models.OneAccountBaseByPkId(accountId)
	if err != nil {
		return
	}

	oldPwd := tools.PasswordEncrypt(reqJSON["old_pwd"].(string), accountBase.RegisterTime)
	if oldPwd != accountBase.Password {
		isOldPwdWrong = true
		return
	}

	// 更新账号密码
	accountBase.Password = tools.PasswordEncrypt(reqJSON["new_pwd"].(string), accountBase.RegisterTime)

	o := orm.NewOrm()
	o.Using(accountBase.Using())
	_, err = o.Update(&accountBase, "password")
	if err != nil {
		bson, _ := json.Marshal(accountBase)
		logs.Error("Update account has wrong: ", string(bson), ", err:", err)
		return
	}

	return
}

func BuildAccountProfile(accountId int64) interface{} {
	accountProfile := models.AccountProfile{AccountId: accountId}

	o := orm.NewOrm()
	o.Using(accountProfile.Using())
	// 失败了就用默认值
	o.Read(&accountProfile)

	accountBase := models.AccountBase{Id: accountId}
	m := orm.NewOrm()
	m.Using(accountBase.Using())
	m.Read(&accountBase)

	// 兼容存量数据的逻辑
	if accountProfile.IdPhoto == 0 || accountProfile.HandHeldIdPhoto == 0 {
		accountBase.Realname = ""
		accountBase.Identity = ""
	}

	profile := map[string]interface{}{
		"base_info": map[string]interface{}{
			"account_id":             accountBase.Id,
			"gender":                 accountBase.Gender,
			"realname":               accountBase.Realname,
			"identity":               accountBase.Identity,
			"mobile":                 accountBase.Mobile,
			"id_photo_url":           BuildResourceUrl(accountProfile.IdPhoto),
			"hand_held_id_photo_url": BuildResourceUrl(accountProfile.HandHeldIdPhoto),
		},
		"work_info": map[string]interface{}{
			"job_type":          accountProfile.JobType,
			"monthly_income":    accountProfile.MonthlyIncome,
			"company_name":      accountProfile.CompanyName,
			"company_city":      accountProfile.CompanyCity,
			"company_address":   accountProfile.CompanyAddress,
			"service_years":     accountProfile.ServiceYears,
			"company_telephone": accountProfile.CompanyTelephone,
			"salary_day":        accountProfile.SalaryDay,
		},
		"contact_info": map[string]interface{}{
			"contact1":      accountProfile.Contact1,
			"contact1_name": accountProfile.Contact1Name,
			"relationship1": accountProfile.Relationship1,
			"contact2":      accountProfile.Contact2,
			"contact2_name": accountProfile.Contact2Name,
			"relationship2": accountProfile.Relationship2,
		},
		"other_info": map[string]interface{}{
			"education":        accountProfile.Education,
			"marital_status":   accountProfile.MaritalStatus,
			"children_number":  accountProfile.ChildrenNumber,
			"resident_city":    accountProfile.ResidentCity,
			"resident_address": accountProfile.ResidentAddress,
			"bank_name":        accountProfile.BankName,
			"bank_no":          accountProfile.BankNo,
		},
	}

	return profile
}

func simpleUpdateCheck(funcName string, accountId int64) (err error) {
	if accountId <= 0 {
		logs.Error("call:", funcName, "get wrong id input. accountId:", accountId)
		err = fmt.Errorf("wrong id input: accountId: %d", accountId)
		return
	}

	return
}

// 初始化profile,如果是新用户,则创建profile记录
func InitAccountProfile(accountId int64) (profile models.AccountProfile, err error) {
	profile.AccountId = accountId
	profile.ChildrenNumber = -1 // 硬编码,客户端无法区法默认0和用户选择0
	profile.Ctime = tools.GetUnixMillis()
	profile.Utime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(profile.Using())

	_, _, err = o.ReadOrCreate(&profile, "account_id")
	if err != nil {
		profileJSON, _ := tools.JsonEncode(profile)
		logs.Error("ReadOrCreate is fail. accountId:", accountId, ", profileJSON:", profileJSON)
		return
	}

	return
}

// 更新账户基本信息v2
func UpdateAccountBaseV2(accountId int64, realname, identity string, gender types.GenderEnum) (num int64, code cerror.ErrCode, err error) {
	err = simpleUpdateCheck("UpdateAccountBase", accountId)
	if err != nil {
		return
	}
	obj, _ := models.OneAccountBaseByPkId(accountId)
	origin := obj

	obj.Id = accountId
	obj.Realname = realname
	obj.Identity = identity
	obj.Gender = gender

	o := orm.NewOrm()
	o.Using(obj.Using())

	var list []models.AccountBase
	num, err = o.QueryTable(obj.TableName()).Filter("identity", identity).All(&list)

	if num > 1 {
		// 1. 已经有多条绑定记录
		code = cerror.IdentityBindRepeated
		listJSON, _ := tools.JsonEncode(list)
		err = fmt.Errorf("identity bind repeated, account: %d, identity: %s, list: %s", accountId, identity, listJSON)
		logs.Error("check has error:", err)
		return
	} else if num == 1 {
		// 其他人绑定过这个身份证号
		if accountId != list[0].Id {
			code = cerror.IdentityBindRepeated
			listJSON, _ := tools.JsonEncode(list)
			err = fmt.Errorf("identity already bind other account, account: %d, identity: %s, list: %s", accountId, identity, listJSON)
			logs.Error("bind error:", err)
			return
		}
	}

	num, err = o.Update(&obj, "realname", "identity", "gender")
	if err != nil || num > 1 {
		logs.Error("[UpdateAccountBaseV2] Update account base info has wrong. obj:", obj, ", err:", err, ", num:", num)
	} else {
		// 写操作日志
		models.OpLogWrite(accountId, accountId, models.OpCodeAccountBaseUpdate, obj.TableName(), origin, obj)
	}

	return
}

// 更新账户基本信息v3
func UpdateAccountBaseV3(accountId int64, realname, identity string, gender types.GenderEnum, isManual int) (num int64, code cerror.ErrCode, err error) {
	err = simpleUpdateCheck("UpdateAccountBase", accountId)
	if err != nil {
		return
	}
	obj, _ := models.OneAccountBaseByPkId(accountId)
	origin := obj

	obj.Id = accountId
	obj.Realname = realname
	obj.Identity = identity
	obj.Gender = gender

	o := orm.NewOrm()
	o.Using(obj.Using())

	var list []models.AccountBase
	num, err = o.QueryTable(obj.TableName()).Filter("identity", identity).All(&list)

	if num > 1 {
		// 1. 已经有多条绑定记录
		code = cerror.IdentityBindRepeated
		listJSON, _ := tools.JsonEncode(list)
		err = fmt.Errorf("identity bind repeated, account: %d, identity: %s, list: %s", accountId, identity, listJSON)
		logs.Error("check has error:", err)
		return
	} else if num == 1 {
		// 其他人绑定过这个身份证号
		if accountId != list[0].Id {
			code = cerror.IdentityBindRepeated
			listJSON, _ := tools.JsonEncode(list)
			err = fmt.Errorf("identity already bind other account, account: %d, identity: %s, list: %s", accountId, identity, listJSON)
			logs.Warn("bind error:", err)
			return
		}
	}

	num, err = o.Update(&obj, "realname", "identity", "gender")
	if err != nil || num > 1 {
		logs.Error("[UpdateAccountBaseV3] Update account base info has wrong. obj:", obj, ", err:", err, ", num:", num)
		return
	} else {
		// 写操作日志
		models.OpLogWrite(accountId, accountId, models.OpCodeAccountBaseUpdate, obj.TableName(), origin, obj)
	}

	ext, extErr := models.OneAccountBaseExtByPkId(accountId)
	if extErr == nil {
		ext.IsManualIdentity = isManual
		ext.Utime = tools.GetUnixMillis()

		models.OrmAllUpdate(&ext)
	} else {
		ext = models.AccountBaseExt{}
		ext.AccountId = accountId
		ext.IsManualIdentity = isManual
		ext.Utime = tools.GetUnixMillis()
		ext.Ctime = ext.Utime

		models.OrmInsert(&ext)
	}

	return
}

// 更新账户基本信息
func UpdateAccountBase(accountId int64, realname, identity string, gender types.GenderEnum) (num int64, err error) {
	err = simpleUpdateCheck("UpdateAccountBase", accountId)
	if err != nil {
		return
	}
	obj, _ := models.OneAccountBaseByPkId(accountId)
	origin := obj

	obj.Id = accountId
	obj.Realname = realname
	obj.Identity = identity
	obj.Gender = gender

	o := orm.NewOrm()
	o.Using(obj.Using())

	num, err = o.Update(&obj, "realname", "identity", "gender")
	if err != nil || num > 1 {
		logs.Error("[UpdateAccountBase] Update account base info has wrong. obj:", obj, ", err:", err, ", num:", num)
	} else {
		// 写操作日志
		models.OpLogWrite(accountId, accountId, models.OpCodeAccountBaseUpdate, obj.TableName(), origin, obj)
	}

	return
}

// UpdateAccountBaseByThird 使用第三方身份认证信息,更新account_base(实际是插入新字段)
func UpdateAccountBaseByThird(m models.AccountBase) (num int64, err error) {
	accountBase, _ := models.OneAccountBaseByPkId(m.Id)
	o := orm.NewOrm()
	o.Using(m.Using())

	num, err = o.Update(&m, "ThirdID", "ThirdName", "ThirdProvince", "ThirdCity", "ThirdDistrict", "ThirdVillage")
	if err != nil || num > 1 {
		logs.Error("[UpdateAccountBaseByThird] Update account base info has wrong. obj:", m, ",err:", err, ", num:", num)
	} else {
		//logs.Debug("Update account base info success. obj:", m, ", num:", num)
		// 写操作日志
		models.OpLogWrite(m.Id, m.Id, models.OpCodeAccountBaseUpdate, accountBase.TableName(), accountBase, m)
	}

	return
}

// 更新OCR识别数据
func UpdateAccountBaseOCR(accountId int64, realname, identity string) (num int64, err error) {

	obj, _ := models.OneAccountBaseByPkId(accountId)
	origin := obj

	obj.Id = accountId
	obj.OcrRealname = realname
	obj.OcrIdentity = identity

	o := orm.NewOrm()
	o.Using(obj.Using())

	num, err = o.Update(&obj, "ocr_realname", "ocr_identity")
	if err != nil || num > 1 {
		logs.Error("[UpdateAccountBaseOCR] Update account base info has wrong. obj:", obj, ",err:", err, ", num:", num)
	} else {
		// 写操作日志
		models.OpLogWrite(accountId, accountId, models.OpCodeAccountBaseUpdate, obj.TableName(), origin, obj)
	}

	return
}

// 清空身份证信息OCR识别数据
func ClearAccountBase(accountId int64) (num int64, err error) {

	obj, _ := models.OneAccountBaseByPkId(accountId)
	origin := obj

	obj.Id = accountId
	obj.Identity = ""
	obj.Realname = ""

	o := orm.NewOrm()
	o.Using(obj.Using())

	num, err = o.Update(&obj, "identity", "realname")
	if err != nil || num > 1 {
		logs.Error("[UpdateAccountBaseOCR] Update account base info has wrong. obj:", obj, ",err:", err, ", num:", num)
	} else {
		// 写操作日志
		models.OpLogWrite(accountId, accountId, models.OpCodeAccountBaseUpdate, obj.TableName(), origin, obj)
	}

	return
}

// 更新 profile 的身份证信息
func UpdateAccountProfileIdPhoto(accountId, idPhoto, handHeldIdPhoto int64) (num int64, err error) {
	if accountId <= 0 {
		logs.Error("UpdateAccountProfileIdPhoto -> wrong id input. accountId:", accountId, ", idPhoto:", idPhoto, ", handHeldIdPhoto:", handHeldIdPhoto)
		err = fmt.Errorf("wrong id input: accountId: %d, idPhoto: %d, handHeldIdPhoto: %d", accountId, idPhoto, handHeldIdPhoto)
		return
	}
	obj, _ := models.OneAccountProfileByAccountID(accountId)

	origin := obj

	obj.AccountId = accountId
	obj.Ctime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(obj.Using())

	_, _, err = o.ReadOrCreate(&obj, "account_id")
	if err != nil {
		logs.Error("ReadOrCreate is fail. accountId:", accountId)
		return
	}

	obj.IdPhoto = idPhoto
	obj.HandHeldIdPhoto = handHeldIdPhoto
	obj.Utime = tools.GetUnixMillis()

	num, err = o.Update(&obj, "id_photo", "hand_held_id_photo", "utime")
	if err != nil || num != 1 {
		logs.Error("Update account profile id photo has wrong. obj:", obj, ", err:", err)
	} else {
		// 写操作日志
		models.OpLogWrite(accountId, accountId, models.OpUserInfoUpdate, obj.TableName(), origin, obj)
	}

	return
}

// 更新profile的work info
func UpdateAccountWorkInfo(accountId int64, jobType, monthlyIncome, serviceYears int, companyName, companyCity, companyAddress string) (num int64, err error) {
	err = simpleUpdateCheck("UpdateAccountWorkInfo", accountId)
	if err != nil {
		return
	}
	obj, _ := models.OneAccountProfileByAccountID(accountId)
	origin := obj

	obj.AccountId = accountId
	obj.JobType = jobType
	obj.MonthlyIncome = monthlyIncome
	obj.CompanyName = companyName
	obj.CompanyCity = companyCity
	obj.CompanyAddress = companyAddress
	obj.ServiceYears = serviceYears
	obj.Utime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(obj.Using())

	num, err = o.Update(&obj, "job_type", "monthly_income", "company_name", "company_city", "company_address", "service_years", "utime")
	if err != nil || num != 1 {
		logs.Error("Update account profile work info has wrong. obj:", obj, ", err:", err)
	} else {
		// 写操作日志
		models.OpLogWrite(accountId, accountId, models.OpUserInfoUpdate, obj.TableName(), origin, obj)
	}

	return
}

// 更新profile的work info
func UpdateAccountWorkInfoV2(accountId int64, jobType, monthlyIncome, serviceYears int, companyName, companyCity, companyTelephone, salaryDay string) (num int64, err error) {
	err = simpleUpdateCheck("UpdateAccountWorkInfoV2", accountId)
	if err != nil {
		return
	}
	obj, _ := models.OneAccountProfileByAccountID(accountId)
	origin := obj

	obj.AccountId = accountId
	obj.JobType = jobType
	obj.MonthlyIncome = monthlyIncome
	obj.CompanyName = companyName
	obj.CompanyCity = companyCity
	obj.CompanyTelephone = companyTelephone
	obj.SalaryDay = salaryDay
	obj.ServiceYears = serviceYears
	obj.Utime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(obj.Using())

	num, err = o.Update(&obj, "job_type", "monthly_income", "company_name", "company_city", "company_telephone", "salary_day", "service_years", "utime")
	if err != nil || num != 1 {
		logs.Error("Update account profile work info has wrong. obj:", obj, ", err:", err)
	} else {
		// 写操作日志
		models.OpLogWrite(accountId, accountId, models.OpUserInfoUpdate, obj.TableName(), origin, obj)
	}

	return
}

// 更新 profile contact info
func UpdateAccountContactInfo(accountId int64, contact1Name, contact1, contact2Name, contact2 string, relationship1, relationship2 int) (num int64, err error) {
	err = simpleUpdateCheck("UpdateAccountContactInfo", accountId)
	if err != nil {
		return
	}
	accountProfile, _ := models.OneAccountProfileByAccountID(accountId)
	origin := accountProfile

	accountProfile.AccountId = accountId
	accountProfile.Contact1 = contact1
	accountProfile.Contact1Name = contact1Name
	accountProfile.Relationship1 = relationship1
	accountProfile.Contact2 = contact2
	accountProfile.Contact2Name = contact2Name
	accountProfile.Relationship2 = relationship2
	accountProfile.Utime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(accountProfile.Using())

	num, err = o.Update(&accountProfile, "contact1", "contact1_name", "relationship1", "contact2", "contact2_name", "relationship2", "utime")
	if err != nil || num != 1 {
		logs.Error("Update account profile contact info has wrong. obj:", accountProfile, ", err:", err)
	} else {
		// 写操作日志
		models.OpLogWrite(accountId, accountId, models.OpUserInfoUpdate, accountProfile.TableName(), origin, accountProfile)
	}

	return
}

// 更新 profile other info
func UpdateAccountOtherInfo(accountId int64, education, maritalStatus, childrenNumber int, bankName, bankNo string) (num int64, err error) {
	err = simpleUpdateCheck("UpdateAccountOtherInfo", accountId)
	if err != nil {
		return
	}

	accountProfile, _ := models.OneAccountProfileByAccountID(accountId)

	origin := accountProfile

	accountProfile.AccountId = accountId
	accountProfile.Education = education
	accountProfile.MaritalStatus = maritalStatus
	accountProfile.ChildrenNumber = childrenNumber
	accountProfile.BankName = bankName
	accountProfile.BankNo = bankNo
	accountProfile.Utime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(accountProfile.Using())

	num, err = o.Update(&accountProfile, "education", "marital_status", "children_number", "bank_name", "bank_no", "utime")
	if err != nil || num != 1 {
		logs.Error("Update account profile other info has wrong. obj:", accountProfile, ", err:", err)
	} else {
		// 写操作日志
		models.OpLogWrite(accountId, accountId, models.OpUserInfoUpdate, accountProfile.TableName(), origin, accountProfile)
	}

	return
}

// IdentityVerify 验证身份检查结果
// 优先检查同盾，同盾失败再检查advance ，都失败则失败
func IdentityVerify(accountID int64) (verify bool) {
	if accountID == 0 {
		verify = false
		return
	}
	accountBase, _ := models.OneAccountBaseByPkId(accountID)
	//同盾数据
	accountTongdun, _ := models.GetOneAC(accountID, tongdun.ChannelCodeKTP)
	//同盾命中，匹配 (增加数据不为空判断，老用户没有同盾数据导致该条件成立然后清空了thirdID和 third_name)
	if accountTongdun.AccountID == accountID &&
		accountTongdun.CheckCode == tongdun.IDCheckCodeYes { //Y
		verify = true
		//如果同盾识别成功，冗余身份信息到base
		accountBase.ThirdID = accountTongdun.OcrIdentity
		accountBase.ThirdName = accountTongdun.OcrRealName
		UpdateAccountBaseByThird(accountBase)
		return
	}
	//由于APP版本兼容问题，所以advance始终都得跑一次
	if accountBase.Identity == "" {
		verify = false
		logs.Debug("[service.IdentityVerify] verify不通过原因: base identity为空")
	} else {
		if accountBase.Identity != accountBase.ThirdID {
			verify = false
			logs.Debug("[service.IdentityVerify] verify不通过原因: identity不等于thirdID ", accountBase.Identity, accountBase.ThirdID)
		} else {
			verify = true
		}
	}
	return
}

//最近一次活体绑定订单ID
func LastLiveVerifyBindOrderID(orderID int64) {
	orderData, err := models.GetOrder(orderID)
	if err != nil {
		logs.Error("[LastLiveVerifyBindOrderID] get order happend err:", err, "orderID:", orderID)
	}
	liveVerify, err1 := dao.CustomerLiveVerify(orderData.UserAccountId)
	if err1 != nil {
		logs.Error("[LastLiveVerifyBindOrderID] get liveVerify happend err:", err, "accountID:", orderData.UserAccountId)
	}
	liveVerify.OrderID = orderID
	cols := []string{"order_id"}
	num, err := models.OrmUpdate(&liveVerify, cols)
	logs.Debug("[LastLiveVerifyBindOrderID] liveVerify.OrderID:", liveVerify.OrderID, "liveVerify.Id :", liveVerify.Id, "num:", num, "err:", err)
}

// 活体认证
func AccountLiveVerify(accountId int64, imageResourceIdMap map[string]int64, originRes []byte) (confidenceAvg float64, isAlive bool, err error) {
	err = simpleUpdateCheck("AccountLiveVerify", accountId)
	if err != nil {
		return
	}

	obj := models.LiveVerify{
		AccountId: accountId,
		Ctime:     tools.GetUnixMillis(),
	}

	o := orm.NewOrm()
	o.Using(obj.Using())

	// 图片id逻辑
	if imageRid, ok := imageResourceIdMap["image_best"]; ok {
		obj.ImageBest = imageRid
	}
	if imageRid, ok := imageResourceIdMap["image_env"]; ok {
		obj.ImageEnv = imageRid
	}
	if imageRid, ok := imageResourceIdMap["image_ref1"]; ok {
		obj.ImageRef1 = imageRid
	}
	if imageRid, ok := imageResourceIdMap["image_ref2"]; ok {
		obj.ImageRef2 = imageRid
	}
	if imageRid, ok := imageResourceIdMap["image_ref3"]; ok {
		obj.ImageRef3 = imageRid
	}

	// 识别情况
	resObj := map[string]interface{}{}
	err = json.Unmarshal(originRes, &resObj)
	if err != nil {
		logs.Error("response can NOT decode. accountId:", accountId, ", imageResourceIdMap:", imageResourceIdMap, ", originRes:", string(originRes), ", err:", err)
		return
	}

	var confidenceNum float64
	var confidenceCount float64
	var confidence float64
	var subObj = make(map[string]interface{})
	if jsonObj, ok := resObj["result_ref1"]; ok {
		subObj = jsonObj.(map[string]interface{})
		confidence = subObj["confidence"].(float64)
		obj.ConfidenceRef1 = confidence
		confidenceNum++
		confidenceCount += confidence
	}
	if jsonObj, ok := resObj["result_ref2"]; ok {
		subObj = jsonObj.(map[string]interface{})
		confidence = subObj["confidence"].(float64)
		obj.ConfidenceRef2 = confidence
		confidenceNum++
		confidenceCount += confidence
	}
	if jsonObj, ok := resObj["result_ref3"]; ok {
		subObj = jsonObj.(map[string]interface{})
		confidence = subObj["confidence"].(float64)
		obj.ConfidenceRef3 = confidence
		confidenceNum++
		confidenceCount += confidence
	}

	if confidenceNum > 0 {
		confidenceAvg = confidenceCount / confidenceNum
	}
	// 平均识别率,先拍死一个值,一期写死,后续改为后台可动态调整的配置荐. TODO
	if confidenceAvg > types.FaceidVerifyConfidence {
		isAlive = true
	}

	o.Insert(&obj)

	return
}

// ProfileCompletePhase ! 这个方法里面有不少魔术数字  针对首贷、复贷返回的码不一样，请调用者注意
func ProfileCompletePhase(accountId int64, appUIVersion string, appVersionCode int) (phase int) {
	accountBase, _ := models.OneAccountBaseByPkId(accountId)
	profile, _ := dao.CustomerProfile(accountId)
	reLoan := dao.IsRepeatLoan(accountBase.Id)
	if reLoan {
		return ProfileCompletePhaseReLoan(accountId, appUIVersion, appVersionCode)
	}

	// 0: 未提交任何资料,需要上传证件和补充`姓名,证件号,性别`
	if len(accountBase.Identity) < types.LimitIdentity || len(accountBase.Realname) < types.LimitName || accountBase.Gender == types.GenderSecrecy || profile.IdPhoto <= 0 || profile.HandHeldIdPhoto <= 0 {
		return
	}
	phase = types.AccountInfoCompletePhaseBase // 完成姓名,身份证,性别信息提交
	// TODO: 考虑最后一次认证的时间
	logs.Debug("[ProfileCompletePhase] appUIVersion:", appUIVersion, "appVersionCode", appVersionCode)

	liveVerify, err := dao.CustomerLiveVerify(accountId)
	if err != nil || LiveVerifyExpired(&liveVerify, appUIVersion, appVersionCode) || liveVerify.VerifyConfidence() < types.FaceidVerifyConfidence {
		return
	}

	phase = types.AccountInfoCompletePhaseLive // 完成活体识别

	// 校验是否需要进行移动数据抓取
	// uiversion 标识此app是印尼版本，其他版本不一定要做运营商抓取。
	// 旧版本app无uiversion同样跳过抓取流程
	// versionCode 大于特定的版本号才会走运营商抓取流程
	if types.IndonesiaAppUIVersion == appUIVersion && types.IndonesiaAppRipeVersionCode <= appVersionCode {
		flag := NeedCatchOptData(&accountBase, appUIVersion, appVersionCode)
		if flag {
			// 需要抓取  且抓取状态不为成功的时候 app跳转到抓取页面 此时直接返回   客户端收到完成活体验证后即进入数据抓取页
			return
		} else {
			phase = types.AccountInfoCompleteOptVerify // 不需要抓取的话直接返回   完成数据抓取
		}
	}

	if profile.JobType <= 0 || profile.MonthlyIncome <= 0 || profile.ServiceYears <= 0 {
		return
	}
	if types.IndonesiaAppRipeVersionSalaryDay <= appVersionCode {
		if len(profile.SalaryDay) < types.LimitSalaryDay {
			return
		}
	}

	phase = types.AccountInfoCompletePhaseWork // 完成基本工作信息提交
	if len(profile.Contact1) < types.LimitMobile || len(profile.Contact1Name) < types.LimitName || profile.Relationship1 <= 0 || len(profile.Contact2) < types.LimitMobile || len(profile.Contact2Name) < types.LimitName || profile.Relationship2 <= 0 {
		return
	}

	phase = types.AccountInfoCompletePhaseContact // 完成联系人信息提交
	if profile.Education <= 0 || profile.MaritalStatus <= 0 || len(profile.BankName) < types.LimitName || len(profile.BankNo) < types.LimitBankNo {
		return
	}

	// 完成其他信息提交.全部阶段都完成了
	phase = types.AccountInfoCompletePhaseDone

	// 校验是否需要显示授信页
	// uiversion 标识此app是印尼版本，其他版本不一定要做补充授信抓取。
	// 旧版本app无uiversion同样跳过抓取流程
	// versionCode 大于特定的版本号才会走授信抓取
	if types.IndonesiaAppUIVersion == appUIVersion && types.IndonesiaAppRipeVersionCodeAddition <= appVersionCode {
		flag := NeedCatchAdditonalAuthorize(&accountBase)
		if flag {
			// 需要抓取  且抓取状态不为成功的时候 app跳转到抓取页面 此时直接返回   客户端收到完成其他信息后即进入补充数据抓取页
			return
		} else {
			phase = types.AccountInfoCompleteAddition // 不需要抓取的话直接返回   完成数据抓取
		}
	}

	return
}

// ProfileCompletePhaseReLoan 校验复贷流程步骤
func ProfileCompletePhaseReLoan(accountId int64, appUIVersion string, appVersionCode int) (phase int) {
	// accountBase, _ := models.OneAccountBaseByPkId(accountId)
	// phase = types.AccountInfoCompletePhaseNoneReLoan

	// // 根据订单id判断 是否上传了手持证件照
	// isUploadHoldPhoto := dao.IsUploadHoldPhoto(accountId)
	// if !isUploadHoldPhoto {
	// 	return
	// }
	phase = types.AccountInfoCompletePhaseHoldReLoan

	liveVerify, err := dao.CustomerLiveVerify(accountId)
	// TODO: 考虑最后一次认证的时间
	if err != nil || LiveVerifyExpired(&liveVerify, appUIVersion, appVersionCode) || liveVerify.VerifyConfidence() < types.FaceidVerifyConfidence {
		return
	}

	accountBaseExt, _ := models.OneAccountBaseExtByPkId(accountId)
	repeatLoanQuota := GetRepeatLoanQuota()
	// ABTest 标识为B时，进入复贷提额页面
	if (accountBaseExt.PageAfterLiveFlag == types.ABTestDividerFlagB && repeatLoanQuota) || !repeatLoanQuota {
		// 只有新版本有 授信项     84版app 是由66版升级而来  所以 不认新码 跳过它
		if appVersionCode >= types.IndonesiaAppRipeVersionNewReloanStep &&
			appVersionCode != 84 &&
			appUIVersion == types.IndonesiaAppUIVersion {
			phase = types.AccountInfoCompletePhaseJumpToAuthoriation
			//判断是否完成全部授权项
			isDoneAuth, _, _, _ := CustomerAuthorize(accountId)
			if isDoneAuth == 0 {
				return
			}
		}
	}

	phase = types.AccountInfoCompletePhaseLiveReLoan

	return
}

func GetLoanFlowPhaseConfig(appVersionCode int) (isReadConfig bool, phaseConfig string) {
	// 获取借款流程配置参数(补充授信要放在最后一步)
	phaseConfig = types.DefaultLoanFlow // 默认的借款流程配置
	if appVersionCode != types.IndonesiaAppRipeVersionNewLoanFlow && appVersionCode != types.IndonesiaAppRipeVersionNewLoanFlowT {

		phaseStr := config.ValidItemString("loan_flow_phase")
		if len(phaseStr) > 0 {
			isReadConfig = true
			phaseConfig = phaseStr
		}
	}

	return
}

func GetLoanFlowProgress(accountBase models.AccountBase, profile *models.AccountProfile) (progress int) {

	progress = 1 // progress是下一步的进度
	if profile.JobType == types.WorkType1 || profile.JobType == types.WorkType2 {
		if profile.MonthlyIncome > 0 && profile.ServiceYears > 0 &&
			len(profile.SalaryDay) >= types.LimitSalaryDay {
			progress++
		}
	}
	if profile.JobType == types.WorkType3 || profile.JobType == types.WorkType4 {
		if profile.MonthlyIncome > 0 {
			progress++
		}
	}

	if len(profile.Contact1) >= types.LimitMobile && len(profile.Contact1Name) >= types.LimitName && profile.Relationship1 > 0 &&
		len(profile.Contact2) >= types.LimitMobile && len(profile.Contact2Name) >= types.LimitName && profile.Relationship2 > 0 {
		progress++
	}

	if profile.Education > 0 && profile.MaritalStatus > 0 && len(profile.BankName) >= types.LimitName && len(profile.BankNo) >= types.LimitBankNo {
		progress++
	}

	if len(accountBase.Identity) >= types.LimitIdentity && len(accountBase.Realname) >= types.LimitName && accountBase.Gender > types.GenderSecrecy &&
		profile.IdPhoto > 0 && profile.HandHeldIdPhoto > 0 {
		progress++
	}

	return
}

// ProfileCompletePhaseTwo 与 ProfileCompletePhase 的区别是：借款流程灵活配置
func ProfileCompletePhaseTwo(accountId int64, appUIVersion string, appVersionCode int) (progress, phase int) {
	accountBase, _ := models.OneAccountBaseByPkId(accountId)
	profile, _ := dao.CustomerProfile(accountId)
	reLoan := dao.IsRepeatLoan(accountBase.Id)
	if reLoan {
		return 0, ProfileCompletePhaseReLoan(accountId, appUIVersion, appVersionCode)
	}

	logs.Debug("[ProfileCompletePhaseTwo] appUIVersion:", appUIVersion, "appVersionCode", appVersionCode)

	progress = GetLoanFlowProgress(accountBase, profile)

	_, phaseStr := GetLoanFlowPhaseConfig(appVersionCode)
	phaseArr := strings.Split(phaseStr, ",")

	for i := 0; i < len(phaseArr); i++ {
		switch phaseArr[i] {
		case tools.Int2Str(types.AccountInfoPhaseWork):

			if profile.JobType <= 0 {
				phase = types.AccountInfoPhaseWork
				return
			} else {
				if profile.JobType == types.WorkType1 || profile.JobType == types.WorkType2 {
					if profile.MonthlyIncome <= 0 || profile.ServiceYears <= 0 ||
						len(profile.SalaryDay) < types.LimitSalaryDay {
						// 基本工作信息提交
						phase = types.AccountInfoPhaseWork
						return
					}
				}
				if profile.JobType == types.WorkType3 || profile.JobType == types.WorkType4 {
					if profile.MonthlyIncome <= 0 {
						// 基本工作信息提交
						phase = types.AccountInfoPhaseWork
						return
					}
				}
			}

		case tools.Int2Str(types.AccountInfoPhaseContact):
			if len(profile.Contact1) < types.LimitMobile || len(profile.Contact1Name) < types.LimitName || profile.Relationship1 <= 0 ||
				len(profile.Contact2) < types.LimitMobile || len(profile.Contact2Name) < types.LimitName || profile.Relationship2 <= 0 {
				// 联系人信息提交
				phase = types.AccountInfoPhaseContact
				return
			}

		case tools.Int2Str(types.AccountInfoPhaseOther):
			if profile.Education <= 0 || profile.MaritalStatus <= 0 || len(profile.BankName) < types.LimitName || len(profile.BankNo) < types.LimitBankNo {
				// 其他信息提交
				phase = types.AccountInfoPhaseOther
				return
			}

		case tools.Int2Str(types.AccountInfoPhaseBase):
			if len(accountBase.Identity) < types.LimitIdentity || len(accountBase.Realname) < types.LimitName || accountBase.Gender == types.GenderSecrecy ||
				profile.IdPhoto <= 0 || profile.HandHeldIdPhoto <= 0 {
				// 姓名,身份证,性别信息提交
				phase = types.AccountInfoPhaseBase
				return
			}

		case tools.Int2Str(types.AccountInfoPhaseLive):
			liveVerify, err := dao.CustomerLiveVerify(accountId)
			if err != nil || LiveVerifyExpired(&liveVerify, appUIVersion, appVersionCode) || liveVerify.VerifyConfidence() < types.FaceidVerifyConfidence {
				phase = types.AccountInfoPhaseLive // 活体识别提交
				return
			}

		case tools.Int2Str(types.AccountInfoAddition):
			// 校验是否需要显示授信页
			// uiversion 标识此app是印尼版本，其他版本不一定要做补充授信抓取。
			// 旧版本app无uiversion同样跳过抓取流程
			if types.IndonesiaAppUIVersion == appUIVersion {
				flag := NeedCatchAdditonalAuthorize(&accountBase)
				if flag {
					// 有未授权的 展示全都授信的话跳过页面
					isDoneAuth, _, _, _ := CustomerAuthorize(accountId)
					if isDoneAuth == 0 {
						// 需要抓取 且 抓取状态不为成功的时候, app跳转到抓取页面
						phase = types.AccountInfoAddition
						return
					}

				}
			}

		}
	}

	phase = types.AccountInfoComplete // 不需要抓取的话, 返回用户信息完成

	return
}

func CanUpdateBankInfo() bool {

	value, _ := config.ValidItemInt("allow_app_modify_bank")

	if value == types.ModifyBankAllow {
		return true
	} else if value == types.ModifyBankForbidden {
		return false
	} else {
		logs.Error("[CanUpdateBankInfo] ValidItemInt value:%v", value)
		return false
	}
}

func TryAddModifyBankTag(order models.Order) {
	accountErrorMap := map[string]bool{
		"INVALID_DESTINATION":      true, //xendit
		"Transfer Inquiry Decline": true, //doku
	}

	if order.CheckStatus != types.LoanStatusLoanFail {
		logs.Warn("[TryAddModifyBankTag] order is not loan faild. orderId:%d status:%d", order.Id, order.CheckStatus)
		return
	}

	reason := GetFailedDisburseOrderReason(order.Id)
	if !accountErrorMap[reason] {
		logs.Info("[TryAddModifyBankTag] reason:%s no in accountError map. order:%d", reason, order.Id)
		return
	}

	accountBaseExt, _ := models.OneAccountBaseExtByPkId(order.UserAccountId)
	if accountBaseExt.RecallTag == types.RecallTagModifyBank {
		logs.Info("[TryAddModifyBankTag] accountId:%d already have tag.", order.UserAccountId)
		return
	}
	err := ChangeCustomerRecall(order.UserAccountId, order.Id, types.RecallTagModifyBank, types.RemarkTagNone)
	if err != nil {
		logs.Error("[TryAddModifyBankTag] ChangeCustomerRecall ret err:%v order:%#v", err, order)
	}
}

func UpdateBankInfo(opUid int64, accountId int64, bankName, bankNo string) (num int64, err error) {

	originProfile, err := dao.GetAccountProfile(accountId)
	if err != nil {
		logs.Error("[UpdateBankInfo] CustomerProfile err:", err, "accountId is ", accountId)
		return
	}

	_, err = models.OneBankInfoByFullName(bankName)
	if err != nil {
		logs.Error("[UpdateBankInfo] OneBankInfoByFullName bankName no valid. err:%s accountId:%d bankName:%s", err, accountId, bankName)
		return
	}

	if originProfile.BankName == bankName &&
		originProfile.BankNo == bankNo {
		err = fmt.Errorf("[UpdateBankInfo] accountId:%d bank name and no are same with db.", accountId, bankName)
		logs.Warn(err)
		return
	}

	// 30分钟内只允许修改1次
	// +1 分布式锁
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()
	lockKey := beego.AppConfig.String("update_bank_lock")
	lockKey = fmt.Sprintf("%s:%d", lockKey, accountId)
	lock, err := cacheClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", 1800, "NX") // 1800秒分布式锁
	if err != nil || lock == nil {
		logs.Error("[UpdateBankInfo] modify interval limit. accountId:%d", accountId)
		err = fmt.Errorf("[UpdateBankInfo] lock err:%v", err)
		return 0, err
	}

	obj := originProfile

	obj.BankName = bankName
	obj.BankNo = bankNo
	obj.Utime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(obj.Using())

	num, err = o.Update(&obj, "bank_name", "bank_no", "utime")
	if err != nil || num > 1 {
		logs.Error("[UpdateBankInfo] Update account base info has wrong. obj:", obj, ", err:", err, ", num:", num)
	} else {
		models.OpLogWrite(opUid, accountId, models.OpUserInfoUpdate, obj.TableName(), originProfile, obj)
	}

	return
}

func VerifyTimes(id int64, channelType string) (times int) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	keyName := "hash:opt_verify_cache:" + channelType
	//HGET key field
	qValueByte, err := storageClient.Do("HGET", keyName, id)
	if err != nil || qValueByte == nil {
		times = 0
		return
	}
	times, _ = tools.Str2Int(string(qValueByte.([]byte)))
	return
}

func VerifyTimesInc(id int64, channelType string, inc int) {
	if inc == 0 {
		return
	}
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	keyName := "hash:opt_verify_cache:" + channelType
	//HGET key field
	qValueByte, err := storageClient.Do("HGET", keyName, id)
	if err != nil || qValueByte == nil {
		logs.Info("[VerifyTimesInc] no cache for accountId:%d key:%s", id, keyName)
		//insert
		storageClient.Do("HSET", keyName, id, int(math.Max(float64(0), float64(inc))))
		sec := tools.NaturalDay(1) / 1000
		secNow := tools.TimeNow()
		diff := sec - secNow
		logs.Debug("sec:%d secNow:%d diff:%d", sec, secNow, diff)
		storageClient.Do("EXPIRE", keyName, diff)
	} else {
		times, _ := tools.Str2Int(string(qValueByte.([]byte)))
		if times <= 0 && inc < 0 {
			return
		}
		// HINCRBY key field increment
		storageClient.Do("HINCRBY", keyName, id, inc)
	}
	return
}

func GetIdentifyCodeCountByMobile(mobile string) (count int) {
	_, _, count = tongdun.GetChannelByMobile(mobile)
	return
}

// AchieveCodeByAccountId 调用同盾接口抓取移动运营商数据，此方法先创建任务，然后调用登录验证接口
func AchieveCodeByAccountId(id int64, channelType string) (code cerror.ErrCode, err error) {
	//	1、获得account信息
	accountBase, err := models.OneAccountBaseByPkId(id)
	if nil != err {
		code = cerror.CodeUnknown
		logs.Error("[AchieveCodeByAccountId] get account base id catch error. id:%d. err:%s", id, err)
		return
	}

	//  2、创建任务
	channelCode := tongdun.GetChannelCodeByTypeAndMobile(channelType, accountBase.Mobile)
	if "" == channelCode {
		logs.Warn("[AchieveCodeByAccountId] get channelCode :%s err", channelCode)
		code = cerror.TongdunAcquireCodeFail
		return
	}

	name, identityCode := accountBase.Realname, accountBase.Identity
	code, idCheckData, err := tongdun.CreateTask(id, channelType, channelCode, name, identityCode, accountBase.Mobile)
	if code != cerror.CodeSuccess {
		logs.Error("[AchieveCodeByAccountId] CreateTask catch error. id:%d. accountBase:%#v", id, accountBase)
		err = fmt.Errorf("[AchieveCodeByAccountId] CreateTask catch error. id:%d. err:%s ", id, err)
		return
	}

	// 3、调用登录认证接口。发送验证码
	// tongdunModel, _ := models.GetOneByCondition("account_id", strconv.FormatInt(id, 10))
	code, err = tongdun.AcquireCode(id, idCheckData.TaskID, accountBase.Mobile)

	// 获取验证码成功 redis加一
	// if code == cerror.CodeSuccess {
	// 	VerifyTimesInc(id, 1)
	// }
	return
}

// VerifyCodeByAccountId 调用同盾接口抓取移动运营商数据，验证用户输入的验证码
func VerifyCodeByAccountId(id int64, channelType string, codeVerify string) (finishCraw bool, channelCode string, code cerror.ErrCode, err error) {
	//	1、获得account信息
	accountBase, err := models.OneAccountBaseByPkId(id)
	if nil != err {
		code = cerror.CodeUnknown
		logs.Error("[AchieveCodeByAccountId] get account base id catch error. id:%d. err:%s", id, err)
		return
	}

	//  2、创建任务
	channelCode = tongdun.GetChannelCodeByTypeAndMobile(channelType, accountBase.Mobile)
	if "" == channelCode {
		logs.Warn("[VerifyCodeByAccountId] get channelCode :%s err", channelCode)
		code = cerror.TongdunVerifyCodeFail
		return
	}

	// 3、调用登录认证接口。发送验证码
	tongdunModel, err := models.GetOneAC(id, channelCode)
	if err != nil {
		logs.Warn("[VerifyCodeByAccountId] query tongdun model :%s", err)
		code = cerror.TongdunVerifyCodeFail
		return
	}

	//防止未创建任务就直接输入验证码
	if tongdunModel.TaskData != "" &&
		tongdunModel.TaskData != "null" {
		logs.Warn("[VerifyCodeByAccountId] already finish craw data id:%d", tongdunModel.ID)
		code = cerror.TongdunVerifyCodeFail
		finishCraw = true
		return
	}

	code, err = tongdun.VerifyCode(id, codeVerify, tongdunModel.TaskID, accountBase.Mobile)
	return
}

// IsRandomMarkAccountByID 是否为测试用户(随机)
func IsRandomMarkAccountByID(id int64) (bool, error) {
	m, err := models.OneAccountBaseByPkId(id)
	if err != nil {
		logs.Error("Check Account whether Random mark account err:", err)
		return false, err
	}
	return IsRandomMarkAccountByAccountBase(&m), nil
}

// IsRandomMarkAccountByAccountBase 根据AccountBase struct 查看是否是测试用户
// 引用传递, 但方法中不会改变 struct 的任何属性
func IsRandomMarkAccountByAccountBase(m *models.AccountBase) bool {
	if m.RandomMark > 0 {
		return true
	}
	return false
}

// NeedCatchOptData 根据手机号和配置的状态 判断是否需要抓取用户运营商信息 （必走流程）
func NeedCatchOptData(accountBase *models.AccountBase, appUIVersion string, appVersionCode int) (flag bool) {
	status, _ := config.ValidItemInt(types.OperatorCatchFlag)
	flag = false

	// 配置为跳过抓取或者用户为复贷用户 都不需要抓取数据
	if types.OperatorCatchFlagValueSkip == status ||
		true == dao.IsRepeatLoan(accountBase.Id) {
		accountBase.OperatorVerifyStatus = types.OperatorVerifyStatusSkip
		accountBase.Update("operator_verify_status")
		return
	}

	_, code, _ := tongdun.GetChannelByMobile(accountBase.Mobile)

	if code == tongdun.ChannelCodeXI &&
		appUIVersion == types.IndonesiaAppUIVersion &&
		appVersionCode < types.IndonesiaAppRipeVersionXlCatchCode {
		return false
	}

	// 某一版本的app只支持4位验证码的Telkosmel运营商。后续版本三个都支持,另两家运营商是6位验证码。
	if code != tongdun.ChannelCodeTelkomsel &&
		appUIVersion == types.IndonesiaAppUIVersion &&
		appVersionCode < types.IndonesiaAppRipeVersionModifyCode {
		return
	}

	if types.OperatorCatchFlagValueCatch == status &&
		"" != code &&
		(types.OperatorVerifyStatusSuccess != accountBase.OperatorVerifyStatus ||
			((types.OperatorVerifyStatusSuccess == accountBase.OperatorVerifyStatus) &&
				(tools.GetUnixMillis()-accountBase.OperatorVerifyFinishTime) > 3600*24*1000)) { //距离上次验证成功大于1天则需要重新验证
		flag = true
	}
	return
}

func canCatchYysData(accountId int64) int {
	accountBase, _ := models.OneAccountBaseByPkId(accountId)
	_, code, _ := tongdun.GetChannelByMobile(accountBase.Mobile)

	catchFlag := tools.ThreeElementExpression(code != "", int(1), int(0)).(int)

	return catchFlag
}

// 侧边栏是否展示运营商抓取页面 补充授信页
func AuthorizationInfo(accountId int64) interface{} {
	one, _ := models.OneAccountBaseExtByPkId(accountId)
	accountBase, _ := models.OneAccountBaseByPkId(accountId)
	_, code, _ := tongdun.GetChannelByMobile(accountBase.Mobile)

	catchFlag := tools.ThreeElementExpression(code != "", int(1), int(0)).(int)

	npwpStat := one.NpwpStatus
	if npwpStat == types.AuthorizeStatusCrawleSuccess {
		npwpStat = types.AuthorizeStatusSuccess
	}

	author := map[string]interface{}{
		"is_catch_yys": catchFlag,
		"yys":          one.AuthorizeStatusYys,
		"go_jek":       one.AuthorizeStatusGoJek,
		"lazada":       one.AuthorizeStatusLazada,
		"tokopedia":    one.AuthorizeStatusTokopedia,
		"facebook":     one.AuthorizeStatusFacebook,
		"instagram":    one.AuthorizeStatusInstagram,
		"linkedin":     one.AuthorizeStatusLinkedin,
		"npwp_no":      one.NpwpNo,
		"npwp_status":  npwpStat,
	}

	return author
}

//CustomerAuthorize 用户授权信息，返回是否完成全部授、总临时提额、完成百分比、授权项列表[授权项:是否完成授权]==[1001:1 or 0]
//isDoneAuth: 0 quotaTotal: 0 increasePercent: 0 authList: map[1001:0 1002:0 1003:0]
type RetAuthInfo struct {
	Name   string `json:"auth_code"`
	Status int    `json:"auth_status"`
}

// 按照 RetAuthInfo.Name 从大到小排序
type RetAuthInfoSlice []RetAuthInfo

func (a RetAuthInfoSlice) Len() int { // 重写 Len() 方法
	return len(a)
}

func (a RetAuthInfoSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}

func (a RetAuthInfoSlice) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[j].Name > a[i].Name
}

func removeYysIfnoSupport(flag int, items []string) []string {
	if flag == 1 {
		return items
	}

	for k, v := range items {
		if v == credit.BackendCodeYys {
			items = append(items[:k], items[k+1:]...)
		}
	}
	return items
}

func CustomerAuthorize(accountID int64) (isDoneAuth int, quotaTotal, increasePercent int64, authList []RetAuthInfo) {
	accountBaseExt, _ := models.OneAccountBaseExtByPkId(accountID)
	broExt := reflect.ValueOf(&accountBaseExt).Elem()

	//根据首复贷获取配置的授权项
	items := credit.AuthorizeValidityCatchList(dao.IsRepeatLoan(accountID))
	flag := canCatchYysData(accountID)
	items = removeYysIfnoSupport(flag, items)

	countItems := len(items)
	var countStatus int
	stepPercent := 100

	if countItems > 0 {
		stepPercent = 100 / countItems
		for _, v := range items {
			authInfo, ok := credit.AuthorInfoByBackendCode(v)
			if !ok {
				logs.Warn("[CustomerAuthorize] config. unknow code:%s", v)
				continue
			}
			statusVal := broExt.FieldByName(authInfo.StatusColName).Int()
			quotaVal := broExt.FieldByName(authInfo.QuotaColName).Int()

			//0:未授权 1:授权成功 2:授权失败 3:已过期 4:抓取成功
			if statusVal == types.AuthorizeStatusCrawleSuccess {
				countStatus++
				quotaTotal += quotaVal
				increasePercent += int64(stepPercent)
			}
			auth := RetAuthInfo{
				Name:   v,
				Status: int(statusVal),
			}
			authList = append(authList, auth)
		}
	}
	sort.Sort(RetAuthInfoSlice(authList))

	if countItems == countStatus {
		isDoneAuth = 1
		increasePercent = 100
	}
	return
}

// LiveVerifyExpired check result is expired
func LiveVerifyExpired(obj *models.LiveVerify, appUIVersion string, appVersionCode int) (expired bool) {
	if obj == nil {
		return true
	}
	//兼容老版本，老版本不需要做 活体认证的有效期限制
	if types.IndonesiaAppRipeVersionLiveVerify > appVersionCode {
		return false
	}

	interval, _ := config.ValidItemInt(types.LiveVerifyInterval)
	logs.Debug("[service.account.LiveVerifyExpired] config interval:", interval,
		" obj.Ctime:", obj.Ctime,
		" tools.GetUnixMillis() - obj.Ctime:", tools.GetUnixMillis()-obj.Ctime)
	return (tools.GetUnixMillis() - obj.Ctime) > int64(interval*60*1000)
}

// NeedCatchAdditonalAuthorize 根据配置的标志位判断是否需要抓取 补充授信
func NeedCatchAdditonalAuthorize(accountBase *models.AccountBase) (flag bool) {
	status, _ := config.ValidItemInt(types.AdditionalCatchFlag)
	flag = false

	//配置为抓取 且不是复贷用户 才去补充授权
	if types.OperatorCatchFlagValueCatch == status &&
		false == dao.IsRepeatLoan(accountBase.Id) {
		flag = true
	}
	return
}

func AccountBaseByOrderId(id int64) (account models.AccountBase, err error) {
	order := models.Order{
		Id: id,
	}

	o := orm.NewOrm()
	o.Using(order.Using())
	err = o.Read(&order)
	if err != nil {
		return
	}

	account.Id = order.UserAccountId
	o.Using(account.Using())
	err = o.Read(&account)

	return
}

func IsVerifySms(accountId int64) (is_verify_sms bool) {
	accountBase, _ := models.OneAccountBaseByPkId(accountId)
	smsVerifyTime, _ := config.ValidItemInt64("sms_verify_time") // sms_verify_time 单位是分钟

	timeDiff := tools.GetUnixMillis() - accountBase.LatestSmsVerifyTime
	if timeDiff > smsVerifyTime*60*1000 {
		is_verify_sms = true
	}

	return
}

func HaveUnsetOrder(accountId int64) bool {
	orderData, err := dao.AccountLastLoanOrder(accountId)
	if err != nil {
		// 之前从来没有过有效订单,则创建新订单
		return false
	}

	// 有未完结订单,不能再创建新订单,哪怕是临时订单
	if orderData.CheckStatus != types.LoanStatusAlreadyCleared &&
		orderData.CheckStatus != types.LoanStatusInvalid {
		return true
	}
	return false
}

func SetAccountPlatform(id, platform int64) {
	accountBase, err := models.OneAccountBaseByPkId(id)
	if err != nil {
		return
	}

	if accountBase.IsPlatformMark(platform) {
		return
	}

	accountBase.SetPlatformMark(platform)
	accountBase.Update("platform_mark")
}

func ClrAccountPlatform(id, platform int64) {
	accountBase, err := models.OneAccountBaseByPkId(id)
	if err != nil {
		return
	}

	if !accountBase.IsPlatformMark(platform) {
		return
	}

	accountBase.ClrPlatformMark(platform)
	accountBase.Update("platform_mark")
}

func UpdateGojekMark(id int64, jsondata string) {
	gojekData := tongdun.GojekData{}
	err := json.Unmarshal([]byte(jsondata), &gojekData)
	if err != nil {
		return
	}

	if gojekData.AccountInfo.GojekPoin == "" {
		return
	}

	point, err := tools.Str2Int(gojekData.AccountInfo.GojekPoin)
	if err != nil {
		return
	}

	gopoint, _ := config.ValidItemInt("risk_gojek_gopoint")
	if point > gopoint {
		SetAccountPlatform(id, types.PlatformMark_Gojek)
	} else {
		ClrAccountPlatform(id, types.PlatformMark_Gojek)
	}
}

func GetCustomerTypeByAccountId(accountId int64) (customerType string) {

	accountBase, err := models.OneAccountBaseByPkId(accountId)
	if err != nil {
		return
	}
	// 获取印尼当天的零点时间戳（单位：秒）
	zeroTimeStamp, _ := tools.GetTodayTimestampByLocalTime("0")
	zeroTimeStampMS := zeroTimeStamp * 1000
	// 当天注册用户为新注册用户
	if zeroTimeStampMS <= accountBase.RegisterTime && accountBase.RegisterTime < zeroTimeStampMS+tools.MILLSSECONDADAY {
		customerType = types.CustomerTypeNewRegister
		return
	}

	if dao.IsRepeatLoan(accountId) {
		customerType = types.CustomerTypeRepeatLoan
		return
	}

	customerType = types.CustomerTypeFirstLoan

	return
}

func GetABTestDividerFlag(accountID int64) (flag string) {
	// 获取abtest对象配置
	abtestObject := config.ValidItemString("abtest_divider_object")
	if len(abtestObject) <= 0 {
		abtestObject = types.CustomerTypeAll
	}
	// 获取客户类型
	customerType := GetCustomerTypeByAccountId(accountID)

	if abtestObject == customerType || abtestObject == types.CustomerTypeAll {
		abtestPercentage, err := config.ValidItemInt("abtest_divider_percentage")
		if err != nil {
			abtestPercentage = 80 // 分流走A流程的默认是80%
		}

		// 分流器比例分配
		randomValue := tools.GenerateRandom(1, 101)
		if randomValue <= abtestPercentage {
			flag = types.ABTestDividerFlagA
		} else {
			flag = types.ABTestDividerFlagB
		}
	}

	return
}

// 升级活体认证后跳转页面的标记
func UpdatePageAfterLiveFlagInAccountBaseExt(accountID int64, pageAfterLive string) {
	accountBaseExt, _ := models.OneAccountBaseExtByPkId(accountID)

	t := tools.GetUnixMillis()
	accountBaseExt.PageAfterLiveFlag = pageAfterLive
	accountBaseExt.Utime = t
	if accountBaseExt.AccountId == 0 {
		accountBaseExt.AccountId = accountID
		accountBaseExt.Ctime = t
		accountBaseExt.InsertWithNoReturn()
	} else {
		cols := []string{"page_after_live_flag", "utime"}
		accountBaseExt.UpdateWithNoReturn(cols)
	}

	return
}

// 是否修改手机成功
func IsSuccessModifyMobileByIdentity(accountID int64) (isSuccess bool) {
	var configIdhandRecopySimilary, idhandRecopySimilary float64
	var configIdhandSimilary, idhandSimilarity float64
	var faceHoldData advance.ResponseData

	res_list, _ := models.GetLatestSecondResource(accountID)
	// 本次上传身份证
	idPhoto := res_list[1]
	idPhotoTmp := gaws.BuildTmpFilename(idPhoto.Id)
	gaws.AwsDownload(idPhoto.HashName, idPhotoTmp)
	// 方法执行完 删除tmp下的图片
	defer tools.Remove(idPhotoTmp)

	// 历史身份证
	// 获取当前账号
	newAccount, err := models.OneAccountBaseByPkId(accountID)
	if err != nil {
		return
	}
	// 根据当前账号的ocr身份证号，查找对应的之前的账号
	oldAccount, _ := models.OneAccountBaseByIdentity(newAccount.OcrIdentity)
	accountProfile, _ := dao.GetAccountProfile(oldAccount.Id)
	IdPhotoResource, _ := OneResource(accountProfile.IdPhoto)
	idPhotoHistoryTmp := gaws.BuildTmpFilename(accountProfile.IdPhoto)
	gaws.AwsDownload(IdPhotoResource.HashName, idPhotoHistoryTmp)
	// 方法执行完 删除tmp下的图片
	defer tools.Remove(idPhotoHistoryTmp)

	// 本次上传手持身份证
	handHeldIdPhoto := res_list[0]
	handHeldIdPhotoTmp := gaws.BuildTmpFilename(handHeldIdPhoto.Id)
	gaws.AwsDownload(handHeldIdPhoto.HashName, handHeldIdPhotoTmp)
	// 方法执行完 删除tmp下的图片
	defer tools.Remove(handHeldIdPhotoTmp)
	fileHC := map[string]interface{}{
		"idHoldingImage": handHeldIdPhotoTmp,
	}

	//// 本次上传身份证与历史上传身份证对比(身份证号相同)
	configIdSimilary, _ := config.ValidItemFloat64("unbind_idcards_similar")
	idSimilarity, _ := advance.FaceComparison(accountID, idPhotoTmp, idPhotoHistoryTmp)
	logs.Debug("[ IsSuccessModifyMobileByIdentity ] 上传身份证与历史上传身份证比对结果：", idSimilarity,
		" 上传身份证和历史身份证比对阈值：", configIdSimilary, " idPhotoTmp:", idPhotoTmp, " idPhotoHistoryTmp:", idPhotoHistoryTmp)
	if idSimilarity < configIdSimilary {
		isSuccess = false
		goto next
	}
	isSuccess = true

	//// 手持证件照翻拍对比
	configIdhandRecopySimilary, _ = config.ValidItemFloat64("unbind_idhand_recopy_similar")
	idhandRecopySimilary, _ = api253.FaceCheck(accountID, handHeldIdPhotoTmp)
	logs.Debug("[ IsSuccessModifyMobileByIdentity ] 手持证件照翻拍比对结果：", idhandRecopySimilary, " 手持证件照翻拍比对阈值：",
		configIdhandRecopySimilary, " handHeldIdPhotoTmp:", handHeldIdPhotoTmp)
	if idhandRecopySimilary < configIdhandRecopySimilary {
		isSuccess = false
		goto next
	}

	//// 手持比对结果（手持身份证中的大脸与小脸比对）
	_, faceHoldData, _ = advance.Request(accountID, advance.ApiIDCheck, map[string]interface{}{}, fileHC)
	logs.Debug("[ IsSuccessModifyMobileByIdentity ] 手持证件照比对结果（手持身份证中的头像与本人比对）：", faceHoldData,
		" code:", faceHoldData.Code)
	if !advance.IsSuccess(faceHoldData.Code) {
		isSuccess = false
		goto next
	}

	configIdhandSimilary, _ = config.ValidItemFloat64("unbind_idhand_idcard_similar")
	idhandSimilarity = faceHoldData.Data.Similarity
	logs.Debug("[ IsSuccessModifyMobileByIdentity ] 手持证件照比对结果（手持身份证中的头像与本人比对）：", idhandSimilarity,
		" 手持证件照比对阈值：", configIdhandSimilary, " handHeldIdPhotoTmp:", handHeldIdPhotoTmp)
	if idhandSimilarity < configIdhandSimilary {
		isSuccess = false
		goto next
	}

next:
	// 插入数据库（自助修改手机号过程中的阈值）
	t := tools.GetUnixMillis()
	threshold := models.AccountModifyMobileThreshold{
		AccountId:                accountID,
		IdPhoto:                  idPhoto.Id,
		IdPhotoThreshold:         fmt.Sprintf("%f", idSimilarity),
		HandPhoto:                handHeldIdPhoto.Id,
		HandPhotoRecopyThreshold: fmt.Sprintf("%f", idhandRecopySimilary),
		HandPhotoThreshold:       fmt.Sprintf("%f", idhandSimilarity),
		Ctime:                    t,
		Utime:                    t,
	}
	_, errs := threshold.Insert()
	if errs != nil {
		logs.Error("[ IsSuccessModifyMobileByIdentity ] insert accountModifyMobileThreshold failed, err is", errs)

		isSuccess = false
		return
	}

	// 更换手机号
	if isSuccess {
		err := modifyMobileHandler(newAccount, oldAccount)
		if err != nil {
			isSuccess = false
			return
		}
	}

	return
}

// 修改手机号处理（accountID是新的）
func modifyMobileHandler(newAccount, oldAccount models.AccountBase) (err error) {

	// 身份证已存在
	var modifyAccount, modifyNewAccount models.AccountBase
	// 更新身份证号对应的旧账号
	o := orm.NewOrm()
	o.Using(oldAccount.Using())

	o.Begin()
	// 更新当前账号手机号
	invalidMobile := fmt.Sprintf("%s%s", newAccount.Mobile, types.CustomerAccountInvalidSuffix)
	modifyNewAccount = newAccount
	modifyNewAccount.Mobile = invalidMobile
	_, err = o.Update(&modifyNewAccount)
	if err != nil {
		err = fmt.Errorf("[modifyMobileHandler] update accountbase(newAccount) err. account:%#v, mobile:%v, err:%v ", newAccount, invalidMobile, err)
		o.Rollback()
		return
	}

	// 更新旧账号手机号
	modifyAccount = oldAccount
	modifyAccount.Mobile = newAccount.Mobile
	_, err = o.Update(&modifyAccount)
	if err != nil {
		err = fmt.Errorf("[modifyMobileHandler] update accountbase(oldAccount) err. account:%#v, mobile:%v, err:%v ", oldAccount, newAccount.Mobile, err)
		o.Rollback()
		return
	}

	num, errs := models.GetAccountMobileModifyNum(oldAccount.Id)
	if errs != nil {
		o.Rollback()
		return
	}
	if num == 0 {
		err = AddAccountMobileHistory(oldAccount.Id, oldAccount.Mobile)
		if err != nil {
			logs.Error("[ modifyMobileHandler ] insert accountMobileHistory failed, err is", err, ", oldMobile:", oldAccount.Mobile)
			o.Rollback()
			return
		}
	}

	err = AddAccountMobileHistory(oldAccount.Id, newAccount.Mobile)
	if err != nil {
		logs.Error("[ modifyMobileHandler ] insert accountMobileHistory failed, err is", err, ", newMobile:", newAccount.Mobile)
		o.Rollback()
		return
	}

	err = o.Commit()
	if err != nil {
		o.Rollback()
		return
	}

	// 记录更新日志
	models.OpLogWrite(oldAccount.Id, oldAccount.Id, models.OpCodeAccountBaseUpdate, oldAccount.TableName(), oldAccount, modifyAccount)
	models.OpLogWrite(newAccount.Id, newAccount.Id, models.OpCodeAccountBaseUpdate, newAccount.TableName(), newAccount, modifyNewAccount)

	return
}

func AddAccountMobileHistory(accountID int64, mobile string) (err error) {
	t := tools.GetUnixMillis()
	history := models.AccountMobileHistory{
		AccountId: accountID,
		Mobile:    mobile,
		Ctime:     t,
		Utime:     t,
	}
	_, err = history.Insert()

	return
}

func UpdateAccountMobile(account models.AccountBase, mobile string) (err error) {

	oldAccount := account

	account.Mobile = mobile
	_, err = account.Update("mobile")
	if err != nil {
		logs.Error("[UpdateAccountMobile] update err:%v, accountId:%#v, mobile:%v ", err, account.Id, mobile)
		return err
	}
	//opUid int64, opCode OpCodeEnum, opTable string, original interface{}, edited interface{}
	models.OpLogWrite(account.Id, account.Id, models.OpCodeAccountBaseUpdate, account.TableName(), oldAccount, account)

	return nil
}
