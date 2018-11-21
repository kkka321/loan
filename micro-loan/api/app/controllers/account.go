package controllers

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/cerror"
	"micro-loan/common/dao"
	"micro-loan/common/lib/gaws"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/accesstoken"
	"micro-loan/common/pkg/coupon_event"
	"micro-loan/common/pkg/google/push"
	"micro-loan/common/pkg/npwp"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/service"
	"micro-loan/common/strategy/limit"
	"micro-loan/common/thirdparty"
	"micro-loan/common/thirdparty/advance"
	"micro-loan/common/thirdparty/api253"
	"micro-loan/common/thirdparty/faceid"
	"micro-loan/common/thirdparty/tongdun"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type AccountController struct {
	ApiBaseController
}

func (c *AccountController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()

	// 统一将 ip 加到 RequestJSON 中
	c.RequestJSON["ip"] = c.Ctx.Input.IP()
	c.RequestJSON["related_id"] = int64(0)
}

func (c *AccountController) SaveClientInfo() {
	if !service.CheckClientInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 打开APP
	service.RecordClientInfoOpenApp(c.RequestJSON)

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) RequestLoginAuthCode() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckLoginAuthCodeRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	serviceType := types.ServiceRequestLogin
	authCodeType := types.AuthCodeTypeText
	// 过限制策略
	if limit.MobileStrategy(c.RequestJSON["mobile"].(string), serviceType, authCodeType) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LimitStrategyMobile, "")
		c.ServeJSON()
		return
	}

	// 写现场数据
	c.RequestJSON["service_type"] = serviceType
	service.RecordClientInfo(c.RequestJSON)

	// 调用短信服务
	if !service.SendSms(serviceType, authCodeType, c.RequestJSON["mobile"].(string), c.Ctx.Input.IP()) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.SMSServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// 发送短信验证码
func (c *AccountController) RequestLoginAuthCodeV2() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckLoginAuthCodeRequiredV2(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	mobile := c.RequestJSON["mobile"].(string)
	smsType, _ := tools.Str2Int(c.RequestJSON["sms_type"].(string))
	accountBase, err := models.OneAccountBaseByMobile(mobile)
	switch types.ServiceType(smsType) {
	case types.ServiceRegister:
		if err == nil && accountBase.Mobile == mobile {
			c.Data["json"] = cerror.BuildApiResponse(cerror.MobileHasRegistered, "")
			c.ServeJSON()
			return
		}
	case types.ServiceLogin:
		if err != nil && accountBase.Mobile != mobile {
			smsType = int(types.ServiceRegister)
		}
	case types.ServiceFindPassword:
		if err != nil && accountBase.Mobile != mobile {
			c.Data["json"] = cerror.BuildApiResponse(cerror.MobileNotRegistered, "")
			logs.Warning("[RequestLoginAuthCodeV2] Mobile not registered(get find password authcode), mobile:", mobile)
			c.ServeJSON()
			return
		} else if len(accountBase.Password) <= 0 {
			c.Data["json"] = cerror.BuildApiResponse(cerror.AccountPasswordUnset, "")
			c.ServeJSON()
			return
		}
	}

	serviceType := types.ServiceType(smsType)
	authCodeType := types.AuthCodeTypeText
	// 限制策略(一天6次，每次时间间隔至少60秒)
	smsHitStrategy := limit.MobileStrategyV2(mobile, serviceType, authCodeType)
	if smsHitStrategy > 0 {
		errcode := cerror.SMSRequestFrequencyTooHigh
		if smsHitStrategy == limit.SmsTimesTooMore {
			errcode = cerror.LimitStrategyMobile
		}
		c.Data["json"] = cerror.BuildApiResponse(errcode, "")
		c.ServeJSON()
		return
	}

	// 写现场数据
	c.RequestJSON["service_type"] = serviceType
	service.RecordClientInfo(c.RequestJSON)

	// 调用短信服务
	if !service.SendSms(serviceType, authCodeType, mobile, c.Ctx.Input.IP()) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.SMSServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// 发送语音验证码
func (c *AccountController) RequestVoiceAuthCode() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckVoiceAuthCodeRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	mobile := c.RequestJSON["mobile"].(string)
	smsType, _ := tools.Str2Int(c.RequestJSON["sms_type"].(string))
	accountBase, err := models.OneAccountBaseByMobile(mobile)
	switch types.ServiceType(smsType) {
	case types.ServiceRegister:
		if err == nil && accountBase.Mobile == mobile {
			c.Data["json"] = cerror.BuildApiResponse(cerror.MobileHasRegistered, "")
			c.ServeJSON()
			return
		}
	case types.ServiceLogin:
		if err != nil && accountBase.Mobile != mobile {
			smsType = int(types.ServiceRegister)
		}
	case types.ServiceFindPassword:
		if err != nil && accountBase.Mobile != mobile {
			c.Data["json"] = cerror.BuildApiResponse(cerror.MobileNotRegistered, "")
			logs.Warning("[RequestVoiceAuthCode] Mobile not registered(get find password authcode), mobile:", mobile)
			c.ServeJSON()
			return
		} else if len(accountBase.Password) <= 0 {
			c.Data["json"] = cerror.BuildApiResponse(cerror.AccountPasswordUnset, "")
			c.ServeJSON()
			return
		}
	}

	serviceType := types.ServiceType(smsType)
	authCodeType := types.AuthCodeTypeVoice
	// 限制策略(一天6次，每次时间间隔至少60秒)
	smsHitStrategy := limit.MobileStrategyV2(mobile, serviceType, authCodeType)
	if smsHitStrategy > 0 {
		errcode := cerror.SMSRequestFrequencyTooHigh
		if smsHitStrategy == limit.SmsTimesTooMore {
			errcode = cerror.LimitStrategyVoiceAuthCode
		}
		c.Data["json"] = cerror.BuildApiResponse(errcode, "")
		c.ServeJSON()
		return
	}

	// 写现场数据
	c.RequestJSON["service_type"] = serviceType
	service.RecordClientInfo(c.RequestJSON)

	// 调用语言验证码服务
	if !service.SendVoiceAuthCode(serviceType, authCodeType, mobile, c.Ctx.Input.IP()) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.VoiceAuthCodeServiceFail, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) Login() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckLoginRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 验证 auth_code 有效性
	ok := service.CheckSmsCode(c.RequestJSON["mobile"].(string), c.RequestJSON["auth_code"].(string))
	if !ok {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidAuthCode, "")
		c.ServeJSON()
		return
	}

	// 注册新用户或老用户登陆,并将 access_token 返回给客户端
	accountId, accessToken, isNew, err := service.RegisterOrLogin(c.RequestJSON)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	c.AccountID = accountId // 登陆成功时,将id设置为正确值

	// 写登陆或注册的现场数据
	serviceType := types.ServiceLogin
	if isNew {
		serviceType = types.ServiceRegister
	} else {
		c.updateAppsFlyerAndGoogleAdsID()
	}
	c.RequestJSON["service_type"] = serviceType
	c.RequestJSON["related_id"] = accountId
	service.RecordClientInfo(c.RequestJSON)
	service.UpdateClientInfoOpenAppIsRegister(c.RequestJSON)

	// 如果是新用户,创建profile
	service.InitAccountProfile(accountId)

	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"access_token":     accessToken,
		"is_repeat_loan":   dao.IsRepeatLoan(accountId),
		"account_profile":  service.BuildAccountProfile(accountId),
		"loan_lifetime":    service.GetLoanLifetime(c.AccountID),
		"current_step":     service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
		"product_suitable": service.ProductSuitablesForApp(c.AccountID),
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

	go service.SendFantasyGraphReq(accountId, "login")
}

// LoginTwo（首贷借贷流程变化）
func (c *AccountController) LoginTwo() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckLoginRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 验证 auth_code 有效性
	ok := service.CheckSmsCode(c.RequestJSON["mobile"].(string), c.RequestJSON["auth_code"].(string))
	if !ok {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidAuthCode, "")
		c.ServeJSON()
		return
	}

	// 注册新用户或老用户登陆,并将 access_token 返回给客户端
	accountId, accessToken, isNew, err := service.RegisterOrLogin(c.RequestJSON)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	c.AccountID = accountId // 登陆成功时,将id设置为正确值

	// 写登陆或注册的现场数据
	serviceType := types.ServiceLogin
	if isNew {
		serviceType = types.ServiceRegister
	} else {
		c.updateAppsFlyerAndGoogleAdsID()
	}
	c.RequestJSON["service_type"] = serviceType
	c.RequestJSON["related_id"] = accountId
	service.RecordClientInfo(c.RequestJSON)
	service.UpdateClientInfoOpenAppIsRegister(c.RequestJSON)

	// 如果是新用户,创建profile
	service.InitAccountProfile(accountId)

	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"access_token":     accessToken,
		"is_repeat_loan":   dao.IsRepeatLoan(accountId),
		"account_profile":  service.BuildAccountProfile(accountId),
		"loan_lifetime":    service.GetLoanLifetime(c.AccountID),
		"current_step":     phase,
		"progress":         progress,
		"product_suitable": service.ProductSuitablesForApp(c.AccountID),
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

	param := coupon_event.InviteV3Param{}
	param.AccountId = accountId
	param.TaskType = types.AccountTaskLogin
	service.HandleCouponEvent(coupon_event.TriggerInviteV3, param)

	go service.SendFantasyGraphReq(accountId, "login")
}

// 登录时更新'appsflyer的设备ID'以及'google广告ID'
func (c *AccountController) updateAppsFlyerAndGoogleAdsID() {
	res, _ := models.OneAccountBaseByMobile(c.RequestJSON["mobile"].(string))
	if len(res.Channel) > 0 && len(res.AppsflyerID) <= 0 {
		accBase := models.AccountBase{}
		accBase.Id = res.Id
		cond := make([]string, 0)

		appsflyerId, ok := c.RequestJSON["appsflyer_id"].(string)
		if ok {
			accBase.AppsflyerID = appsflyerId
			cond = append(cond, "appsflyer_id")
		}

		googleAdvertisingId, ok := c.RequestJSON["google_advertising_id"].(string)
		if ok {
			accBase.GoogleAdvertisingID = googleAdvertisingId
			cond = append(cond, "google_advertising_id")
		}

		stemFrom, ok := c.RequestJSON["cid"].(string)
		if ok {
			accBase.StemFrom = stemFrom
			cond = append(cond, "stem_from")
		}

		num, err := models.OrmUpdate(&accBase, cond)
		if err != nil || num < 0 {
			logs.Warning("[Login] update appsflyer_id, google_advertising_id, stem_from error", err)
		}
	}
}

func (c *AccountController) Register() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckRegisterRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 验证 auth_code 有效性
	ok := service.CheckSmsCodeV2(c.RequestJSON["mobile"].(string), c.RequestJSON["auth_code"].(string), types.ServiceRegister)
	if !ok {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidAuthCode, "")
		c.ServeJSON()
		return
	}

	// 注册新用户,并将 access_token 返回给客户端
	accountId, accessToken, isNew, err := service.RegisterOrLoginV2(c.RequestJSON)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceUnavailable, "")
		c.ServeJSON()
		return
	}
	// 如果手机号已经注册过，直接提示注册失败
	if !isNew {
		c.Data["json"] = cerror.BuildApiResponse(cerror.MobileHasRegistered, "")
		c.ServeJSON()
		return
	}

	c.AccountID = accountId // 登录成功时,将id设置为正确值

	// 写注册的现场数据
	serviceType := types.ServiceRegister

	c.RequestJSON["service_type"] = serviceType
	c.RequestJSON["related_id"] = accountId
	service.RecordClientInfo(c.RequestJSON)
	service.UpdateClientInfoOpenAppIsRegister(c.RequestJSON)

	// 如果是新用户,创建profile
	service.InitAccountProfile(accountId)

	versionCode := c.updateAppVersion()

	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"access_token":     accessToken,
		"is_repeat_loan":   dao.IsRepeatLoan(accountId),
		"account_profile":  service.BuildAccountProfile(accountId),
		"loan_lifetime":    service.GetLoanLifetime(c.AccountID),
		"current_step":     service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
		"product_suitable": service.ProductSuitablesForApp(c.AccountID),
		"is_exist_pwd":     true,
		"version_update":   versionCode,
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

	go service.SendFantasyGraphReq(accountId, "register")
}

// RegisterTwo（首贷借贷流程变化）
func (c *AccountController) RegisterTwo() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckRegisterRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 验证 auth_code 有效性
	ok := service.CheckSmsCodeV2(c.RequestJSON["mobile"].(string), c.RequestJSON["auth_code"].(string), types.ServiceRegister)
	if !ok {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidAuthCode, "")
		c.ServeJSON()
		return
	}

	// 注册新用户,并将 access_token 返回给客户端
	accountId, accessToken, isNew, err := service.RegisterOrLoginV2(c.RequestJSON)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceUnavailable, "")
		c.ServeJSON()
		return
	}
	// 如果手机号已经注册过，直接提示注册失败
	if !isNew {
		c.Data["json"] = cerror.BuildApiResponse(cerror.MobileHasRegistered, "")
		c.ServeJSON()
		return
	}

	c.AccountID = accountId // 登录成功时,将id设置为正确值

	// 写注册的现场数据
	serviceType := types.ServiceRegister

	c.RequestJSON["service_type"] = serviceType
	c.RequestJSON["related_id"] = accountId
	service.RecordClientInfo(c.RequestJSON)
	service.UpdateClientInfoOpenAppIsRegister(c.RequestJSON)

	// 如果是新用户,创建profile
	service.InitAccountProfile(accountId)

	versionCode := c.updateAppVersion()

	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"access_token":     accessToken,
		"is_repeat_loan":   dao.IsRepeatLoan(accountId),
		"account_profile":  service.BuildAccountProfile(accountId),
		"loan_lifetime":    service.GetLoanLifetime(c.AccountID),
		"current_step":     phase,
		"progress":         progress,
		"product_suitable": service.ProductSuitablesForApp(c.AccountID),
		"is_exist_pwd":     true,
		"version_update":   versionCode,
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

	go service.SendFantasyGraphReq(accountId, "register")
}

// 手机号注册过,则直接登录; 手机号未注册过,则生成账号并登录
func (c *AccountController) SmsLogin() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckSmsLoginRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 判断账号是否被锁定
	if limit.IsAccountLocked(c.RequestJSON["mobile"].(string)) {
		c.Data["json"] = cerror.BuildApiResponseV2(cerror.AccountLocked, "", "")
		c.ServeJSON()
		return
	}

	serviceType := types.ServiceLogin
	_, err := models.OneAccountBaseByMobile(c.RequestJSON["mobile"].(string))
	if err != nil {
		serviceType = types.ServiceRegister
	}

	// 验证 auth_code 有效性
	ok := service.CheckSmsCodeV2(c.RequestJSON["mobile"].(string), c.RequestJSON["auth_code"].(string), serviceType)
	if !ok {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidAuthCode, "")
		c.ServeJSON()
		return
	}

	// 手机短信验证码登录,并将 access_token 返回给客户端
	accountId, accessToken, isNew, isExistPwd, err := service.SmsCodeLogin(c.RequestJSON)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	c.AccountID = accountId // 登录成功时,将id设置为正确值

	if isNew {
		// 如果是新用户,创建profile
		service.InitAccountProfile(accountId)
	} else {
		c.updateAppsFlyerAndGoogleAdsID()
	}

	// 写注册或登录的现场数据
	c.RequestJSON["service_type"] = serviceType
	c.RequestJSON["related_id"] = accountId
	service.RecordClientInfo(c.RequestJSON)
	service.UpdateClientInfoOpenAppIsRegister(c.RequestJSON)

	versionCode := c.updateAppVersion()

	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"access_token":     accessToken,
		"is_repeat_loan":   dao.IsRepeatLoan(accountId),
		"account_profile":  service.BuildAccountProfile(accountId),
		"loan_lifetime":    service.GetLoanLifetime(c.AccountID),
		"current_step":     service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
		"product_suitable": service.ProductSuitablesForApp(c.AccountID),
		"is_exist_pwd":     isExistPwd,
		"is_register":      isNew,
		"version_update":   versionCode,
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

	if isNew {
		go service.SendFantasyGraphReq(accountId, "register")
	} else {
		go service.SendFantasyGraphReq(accountId, "login")
	}
}

// SmsLoginTwo 手机号注册过,则直接登录; 手机号未注册过,则生成账号并登录（首贷借贷流程变化）
func (c *AccountController) SmsLoginTwo() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckSmsLoginRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 判断账号是否被锁定
	if limit.IsAccountLocked(c.RequestJSON["mobile"].(string)) {
		c.Data["json"] = cerror.BuildApiResponseV2(cerror.AccountLocked, "", "")
		c.ServeJSON()
		return
	}

	serviceType := types.ServiceLogin
	_, err := models.OneAccountBaseByMobile(c.RequestJSON["mobile"].(string))
	if err != nil {
		serviceType = types.ServiceRegister
	}

	// 验证 auth_code 有效性
	ok := service.CheckSmsCodeV2(c.RequestJSON["mobile"].(string), c.RequestJSON["auth_code"].(string), serviceType)
	if !ok {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidAuthCode, "")
		c.ServeJSON()
		return
	}

	// 手机短信验证码登录,并将 access_token 返回给客户端
	accountId, accessToken, isNew, isExistPwd, err := service.SmsCodeLogin(c.RequestJSON)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	c.AccountID = accountId // 登录成功时,将id设置为正确值

	if isNew {
		// 如果是新用户,创建profile
		service.InitAccountProfile(accountId)
	} else {
		c.updateAppsFlyerAndGoogleAdsID()
	}

	// 写注册或登录的现场数据
	c.RequestJSON["service_type"] = serviceType
	c.RequestJSON["related_id"] = accountId
	service.RecordClientInfo(c.RequestJSON)
	service.UpdateClientInfoOpenAppIsRegister(c.RequestJSON)

	versionCode := c.updateAppVersion()

	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"access_token":     accessToken,
		"is_repeat_loan":   dao.IsRepeatLoan(accountId),
		"account_profile":  service.BuildAccountProfile(accountId),
		"loan_lifetime":    service.GetLoanLifetime(c.AccountID),
		"current_step":     phase,
		"progress":         progress,
		"product_suitable": service.ProductSuitablesForApp(c.AccountID),
		"is_exist_pwd":     isExistPwd,
		"is_register":      isNew,
		"version_update":   versionCode,
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

	param := coupon_event.InviteV3Param{}
	param.AccountId = accountId
	param.TaskType = types.AccountTaskLogin
	service.HandleCouponEvent(coupon_event.TriggerInviteV3, param)

	if isNew {
		go service.SendFantasyGraphReq(accountId, "register")
	} else {
		go service.SendFantasyGraphReq(accountId, "login")
	}
}

func (c *AccountController) PwdLogin() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckPwdLoginRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponseV2(cerror.LostRequiredParameters, "", "")
		c.ServeJSON()
		return
	}

	// 判断账号是否被锁定
	if limit.IsAccountLocked(c.RequestJSON["mobile"].(string)) {
		c.Data["json"] = cerror.BuildApiResponseV2(cerror.AccountLocked, "", "")
		c.ServeJSON()
		return
	}

	// 手机密码登录,并将 access_token 返回给客户端
	accountId, accessToken, isExist, message, _ := service.PasswordLogin(c.RequestJSON)
	if !isExist {
		c.Data["json"] = cerror.BuildApiResponseV2(cerror.MobileNotRegistered, "", "")
		logs.Warning("[PwdLogin] Mobile not exist(pwd login verify), mobile:", c.RequestJSON["mobile"].(string))
		c.ServeJSON()
		return
	}
	if len(message) > 0 {
		c.Data["json"] = cerror.BuildApiResponseV2(cerror.InvalidPassword, message, "")
		c.ServeJSON()
		return
	}

	c.AccountID = accountId // 登陆成功时,将id设置为正确值

	// 写登录的现场数据
	serviceType := types.ServiceLogin

	c.updateAppsFlyerAndGoogleAdsID()

	c.RequestJSON["service_type"] = serviceType
	c.RequestJSON["related_id"] = accountId
	service.RecordClientInfo(c.RequestJSON)
	service.UpdateClientInfoOpenAppIsRegister(c.RequestJSON)

	versionCode := c.updateAppVersion()

	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"access_token":     accessToken,
		"is_repeat_loan":   dao.IsRepeatLoan(accountId),
		"account_profile":  service.BuildAccountProfile(accountId),
		"loan_lifetime":    service.GetLoanLifetime(c.AccountID),
		"current_step":     service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
		"product_suitable": service.ProductSuitablesForApp(c.AccountID),
		"is_exist_pwd":     true,
		"version_update":   versionCode,
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponseV2(cerror.CodeSuccess, "", data)
	c.ServeJSON()

	go service.SendFantasyGraphReq(accountId, "login")
}

// PwdLoginTwo（首贷借贷流程变化）
func (c *AccountController) PwdLoginTwo() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckPwdLoginRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponseV2(cerror.LostRequiredParameters, "", "")
		c.ServeJSON()
		return
	}

	// 判断账号是否被锁定
	if limit.IsAccountLocked(c.RequestJSON["mobile"].(string)) {
		c.Data["json"] = cerror.BuildApiResponseV2(cerror.AccountLocked, "", "")
		c.ServeJSON()
		return
	}

	// 手机密码登录,并将 access_token 返回给客户端
	accountId, accessToken, isExist, message, _ := service.PasswordLogin(c.RequestJSON)
	if !isExist {
		c.Data["json"] = cerror.BuildApiResponseV2(cerror.MobileNotRegistered, "", "")
		logs.Warning("[PwdLoginTwo] Mobile not exist(pwd login verify), mobile:", c.RequestJSON["mobile"].(string))
		c.ServeJSON()
		return
	}
	if len(message) > 0 {
		c.Data["json"] = cerror.BuildApiResponseV2(cerror.InvalidPassword, message, "")
		c.ServeJSON()
		return
	}

	c.AccountID = accountId // 登陆成功时,将id设置为正确值

	// 写登录的现场数据
	serviceType := types.ServiceLogin

	c.updateAppsFlyerAndGoogleAdsID()

	c.RequestJSON["service_type"] = serviceType
	c.RequestJSON["related_id"] = accountId
	service.RecordClientInfo(c.RequestJSON)
	service.UpdateClientInfoOpenAppIsRegister(c.RequestJSON)

	versionCode := c.updateAppVersion()

	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"access_token":     accessToken,
		"is_repeat_loan":   dao.IsRepeatLoan(accountId),
		"account_profile":  service.BuildAccountProfile(accountId),
		"loan_lifetime":    service.GetLoanLifetime(c.AccountID),
		"current_step":     phase,
		"progress":         progress,
		"product_suitable": service.ProductSuitablesForApp(c.AccountID),
		"is_exist_pwd":     true,
		"version_update":   versionCode,
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponseV2(cerror.CodeSuccess, "", data)
	c.ServeJSON()

	param := coupon_event.InviteV3Param{}
	param.AccountId = accountId
	param.TaskType = types.AccountTaskLogin
	service.HandleCouponEvent(coupon_event.TriggerInviteV3, param)

	go service.SendFantasyGraphReq(accountId, "login")
}

//检测是否需要升级
func (c *AccountController) updateAppVersion() (versionCode string) {
	if gpVCOrigin, ok := c.RequestJSON["app_version_code"]; ok {
		intAppVersionCode, _ := tools.Str2Int(gpVCOrigin.(string))

		strVersionCode := config.ValidItemString("app_version_code")
		tmpAppVersionCode, err := tools.Str2Int(strVersionCode)
		if err != nil {
			tmpAppVersionCode = 0
		}

		if tmpAppVersionCode > intAppVersionCode {
			versionCode = config.ValidItemString("app_version")
		}
	}

	return
}

// 找回密码时，验证短信验证码/语音验证码
func (c *AccountController) SmsVerify() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckSmsVerifyRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 验证 auth_code 有效性
	ok := service.CheckSmsCodeV2(c.RequestJSON["mobile"].(string), c.RequestJSON["auth_code"].(string), types.ServiceFindPassword)
	if !ok {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidAuthCode, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{}
	_, err := models.OneAccountBaseByMobile(c.RequestJSON["mobile"].(string))
	if err != nil {
		// 手机号未被注册过
		data["is_exist"] = false

		c.Data["json"] = cerror.BuildApiResponse(cerror.MobileNotRegistered, data)
		logs.Warning("[SmsVerify] Mobile not exist(find password send auth code verify), mobile:", c.RequestJSON["mobile"].(string))
		c.ServeJSON()
		return
	}

	data["server_time"] = tools.GetUnixMillis()
	data["is_exist"] = true
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) FindPassword() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckFindPwdRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	// 找回密码处理，更新账号密码
	isExist, err := service.FindPasswordHandler(c.RequestJSON)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceUnavailable, "")
		c.ServeJSON()
		return
	}
	if !isExist {
		c.Data["json"] = cerror.BuildApiResponse(cerror.MobileNotRegistered, data)
		logs.Warning("[FindPassword] Mobile not exist(find password), mobile:", c.RequestJSON["mobile"].(string))
		c.ServeJSON()
		return
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) SetPassword() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckSetPwdRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	// 设置密码处理，更新账号密码
	err := service.SetPasswordHandler(c.RequestJSON, c.AccountID)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) ModifyPassword() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckModifyPwdRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	// 设置密码处理，更新账号密码
	isOldPwdWrong, err := service.ModifyPasswordHandler(c.RequestJSON, c.AccountID)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceUnavailable, "")
		c.ServeJSON()
		return
	}
	if isOldPwdWrong {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidOldPassword, "")
		c.ServeJSON()
		return
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// AuthReport 更新client_info
func (c *AccountController) AuthReport() {

	if !service.CheckClientInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}
	accessToken := c.RequestJSON["access_token"].(string)
	var fcmToken string
	if v, ok := c.RequestJSON["fcm_token"]; ok {
		fcmToken = v.(string)
	} else {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 授权上报更新clientinfo
	c.RequestJSON["service_type"] = types.ServiceAuthReport
	c.RequestJSON["related_id"] = c.AccountID
	c.RequestJSON["mobile"] = "" //! 注意,需要显示设置为空,否则有可能引起内核恐慌
	service.RecordClientInfo(c.RequestJSON)

	//更新fcm_token
	if fcmToken != "" {
		accesstoken.UpdateFcmToken(accessToken, fcmToken)
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

}

func (c *AccountController) Logout() {
	if !service.CheckClientInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 写注销登陆态的现场数据
	c.RequestJSON["service_type"] = types.ServiceLogout
	c.RequestJSON["related_id"] = c.AccountID
	c.RequestJSON["mobile"] = "" //! 注意,需要显示设置为空,否则有可能引起内核恐慌
	service.RecordClientInfo(c.RequestJSON)

	accesstoken.CleanTokenCache(types.PlatformAndroid, c.RequestJSON["access_token"].(string))

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// 此接口有副作用,为了减少接口数,有可能会生成临时订单
func (c *AccountController) AccountInfo() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckCreateOrderRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	if !service.HaveUnsetOrder(c.AccountID) {
		accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
		isHitMobile, _ := models.IsBlacklistMobile(accountBase.Mobile)
		if isHitMobile {
			logs.Warn("[AccountInfo] 手机号在内部黑名单内, Mobile: %s", accountBase.Mobile)
		}

		ip := c.RequestJSON["ip"].(string)
		isHitIP, _ := models.IsBlacklistIP(ip)
		if isHitIP {
			logs.Warn("[AccountInfo] IP在内部黑名单内, IP: %s", ip)
		}

		_, eAccountDesc := service.DisplayVAInfoV2(c.AccountID)
		if isHitMobile || isHitIP {
			data := map[string]interface{}{
				"server_time":      tools.GetUnixMillis(),
				"is_repeat_loan":   dao.IsRepeatLoan(c.AccountID),
				"account_profile":  service.BuildAccountProfile(c.AccountID),
				"loan_lifetime":    types.LoanHitBlackList,
				"current_step":     0,
				"e_account_number": eAccountDesc,
				"amount":           0,
				"remaining_days":   0,
			}

			c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
			c.ServeJSON()
			return
		}
	}

	loan, _ := tools.Str2Int64(c.RequestJSON["loan"].(string))
	period, _ := tools.Str2Int(c.RequestJSON["period"].(string))
	if loan > 0 && period > 0 {
		// 创建临时订单
		//// 1. 检查产品
		product, err := service.ProductSuitablesByPeriod(c.AccountID, period, loan)
		if err != nil {
			logs.Error("[AccountInfo] ProductSuitablesByPeriod can not find product. accountId:", c.AccountID, ", err:", err)
			c.Data["json"] = cerror.BuildApiResponse(cerror.ProductDoesNotExist, "")
			c.ServeJSON()
			return
		}
		//// 2. 创建借款订单
		_, orderId, err := service.CreateOrder(c.AccountID, product.Id, loan, period, types.IsTemporaryYes)
		if err == nil {
			// 写创建订单现场数据
			c.RequestJSON["service_type"] = types.ServiceCreateOrder
			c.RequestJSON["related_id"] = orderId
			c.RequestJSON["mobile"] = "" //! 注意,需要显示设置为空,否则有可能引起内核恐慌
			service.RecordClientInfo(c.RequestJSON)
		}
		//! 不能创建订单时,接口数据正常返回
	}

	data := map[string]interface{}{
		"server_time":     tools.GetUnixMillis(),
		"is_repeat_loan":  dao.IsRepeatLoan(c.AccountID),
		"account_profile": service.BuildAccountProfile(c.AccountID),
		"loan_lifetime":   service.GetLoanLifetime(c.AccountID),
		"current_step":    service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
		"menu_show":       service.MenuControlByOrderStatus(c.AccountID),
		"menu_show_v2":    service.MenuControlByOrderStatusV2(c.AccountID),
		//"_debug-AccountID": c.AccountID,
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) AccountInfoV2() {
	if !service.CheckClientInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	order, _ := dao.AccountLastLoanOrder(c.AccountID)
	isExtension := service.IsOrderExtension(order)

	accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
	var isExistPwd bool
	if len(accountBase.Password) > 0 {
		isExistPwd = true
	}

	if !service.HaveUnsetOrder(c.AccountID) {

		isHitMobile, _ := models.IsBlacklistMobile(accountBase.Mobile)
		if isHitMobile {
			logs.Warn("[AccountInfoV2] 手机号在内部黑名单内, Mobile: %s", accountBase.Mobile)
		}

		ip := c.RequestJSON["ip"].(string)
		isHitIP, _ := models.IsBlacklistIP(ip)
		if isHitIP {
			logs.Warn("[AccountInfoV2] IP在内部黑名单内, IP: %s", ip)
		}

		_, eAccountDesc := service.DisplayVAInfoV2(c.AccountID)
		if isHitMobile || isHitIP {
			data := map[string]interface{}{
				"server_time":        tools.GetUnixMillis(),
				"is_repeat_loan":     dao.IsRepeatLoan(c.AccountID),
				"account_profile":    service.BuildAccountProfile(c.AccountID),
				"loan_lifetime":      types.LoanHitBlackList,
				"current_step":       0,
				"e_account_number":   eAccountDesc,
				"amount":             0,
				"remaining_days":     0,
				"authorization_info": service.AuthorizationInfo(c.AccountID),
				"menu_show":          service.MenuControlByOrderStatus(c.AccountID),
				"menu_show_v2":       service.MenuControlByOrderStatusV2(c.AccountID),
				"is_exist_pwd":       isExistPwd,
				"home_order_type":    0,
				"new_msg":            0,
				"is_extension":       false,
			}

			c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
			c.ServeJSON()
			return
		}
	}

	num, _ := push.AccountNewMessageSize(c.AccountID)
	data := map[string]interface{}{
		"server_time":        tools.GetUnixMillis(),
		"is_repeat_loan":     dao.IsRepeatLoan(c.AccountID),
		"account_profile":    service.BuildAccountProfile(c.AccountID),
		"loan_lifetime":      service.GetLoanLifetime(c.AccountID),
		"current_step":       service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
		"menu_show":          service.MenuControlByOrderStatus(c.AccountID),
		"menu_show_v2":       service.MenuControlByOrderStatusV2(c.AccountID),
		"authorization_info": service.AuthorizationInfo(c.AccountID),
		"is_exist_pwd":       isExistPwd,
		"home_order_type":    service.GetHomeOrderType(c.AccountID),
		"new_msg":            num,
		"is_extension":       isExtension,
		//"_debug-AccountID": c.AccountID,
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// AccountInfoTwo（首贷借贷流程变化）
func (c *AccountController) AccountInfoTwo() {
	if !service.CheckClientInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	order, _ := dao.AccountLastLoanOrder(c.AccountID)
	isExtension := service.IsOrderExtension(order)

	accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
	var isExistPwd bool
	if len(accountBase.Password) > 0 {
		isExistPwd = true
	}
	if !service.HaveUnsetOrder(c.AccountID) {

		isHitMobile, _ := models.IsBlacklistMobile(accountBase.Mobile)
		if isHitMobile {
			logs.Warn("[AccountInfoTwo] 手机号在内部黑名单内, Mobile: %s", accountBase.Mobile)
		}

		ip := c.RequestJSON["ip"].(string)
		isHitIP, _ := models.IsBlacklistIP(ip)
		if isHitIP {
			logs.Warn("[AccountInfoTwo] IP在内部黑名单内, IP: %s", ip)
		}

		_, eAccountDesc := service.DisplayVAInfoV2(c.AccountID)
		if isHitMobile || isHitIP {
			data := map[string]interface{}{
				"server_time":        tools.GetUnixMillis(),
				"is_repeat_loan":     dao.IsRepeatLoan(c.AccountID),
				"account_profile":    service.BuildAccountProfile(c.AccountID),
				"loan_lifetime":      types.LoanHitBlackList,
				"current_step":       0,
				"e_account_number":   eAccountDesc,
				"amount":             0,
				"remaining_days":     0,
				"authorization_info": service.AuthorizationInfo(c.AccountID),
				"menu_show":          service.MenuControlByOrderStatus(c.AccountID),
				"menu_show_v2":       service.MenuControlByOrderStatusV2(c.AccountID),
				"is_exist_pwd":       isExistPwd,
				"home_order_type":    0,
				"new_msg":            0,
				"is_extension":       false,
			}

			c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
			c.ServeJSON()
			return
		}
	}

	num, _ := push.AccountNewMessageSize(c.AccountID)
	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":        tools.GetUnixMillis(),
		"is_repeat_loan":     dao.IsRepeatLoan(c.AccountID),
		"account_profile":    service.BuildAccountProfile(c.AccountID),
		"loan_lifetime":      service.GetLoanLifetime(c.AccountID),
		"current_step":       phase,
		"progress":           progress,
		"menu_show":          service.MenuControlByOrderStatus(c.AccountID),
		"menu_show_v2":       service.MenuControlByOrderStatusV2(c.AccountID),
		"authorization_info": service.AuthorizationInfo(c.AccountID),
		"is_exist_pwd":       isExistPwd,
		"home_order_type":    service.GetHomeOrderType(c.AccountID),
		"new_msg":            num,
		"is_extension":       isExtension,
		//"_debug-AccountID": c.AccountID,
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// 身份证OCR识别
func (c *AccountController) IdentityDetect() {
	// 简单判断一下
	if !service.CheckIdentityDetectRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 1. 将文件流写入本地
	// 2. 将文件上传到s3
	idPhoto, idPhotoTmp, code, err := c.UploadResource("fs1", types.Use2IdentityDetect)
	defer tools.Remove(idPhotoTmp)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}
	handHeldIdPhoto, handHeldIdPhotoTmp, code, err := c.UploadResource("fs2", types.Use2IdentityDetect)
	defer tools.Remove(handHeldIdPhotoTmp)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}

	var realname string
	var identity string

	// 3. 调用 advance TODO: 图片尺寸过大压缩,理论应该先去faceid识别一下,但那样有可能响应速度会奇慢
	param := map[string]interface{}{}
	file := map[string]interface{}{
		"ocrImage": idPhotoTmp,
	}
	_, resData, err := advance.Request(c.AccountID, advance.ApiOCR, param, file)
	if err == nil && advance.IsSuccess(resData.Code) {
		ocrRealname := resData.Data.Name
		identity = resData.Data.IDNumber
		// 处理名字中的特殊字符
		realname = tools.TrimRealName(ocrRealname)

		// 身份证号是固定位数,如果识别位数不够,则认为无效,需要重新提交
		if len(realname) > 0 && len(identity) == types.LimitIdentity {
			// 更新基本信息
			service.UpdateAccountBase(c.AccountID, realname, identity, types.GenderSecrecy)

			// 更新OCR识别数据
			service.UpdateAccountBaseOCR(c.AccountID, ocrRealname, identity)

			// 更新用户profile 身份证信息
			service.UpdateAccountProfileIdPhoto(c.AccountID, idPhoto, handHeldIdPhoto)

			// 写待处理队列
			storageClient := storage.RedisStorageClient.Get()
			defer storageClient.Close()
			queueName := beego.AppConfig.String("account_identity_detect")
			storageClient.Do("lpush", queueName, c.AccountID)
		}
	}

	// 5. 返回结果给客户端
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"realname":     realname,
		"identity":     identity,
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// 身份证OCR识别v2
func (c *AccountController) IdentityDetectV2() {
	// 简单判断一下
	if !service.CheckIdentityDetectRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 1. 将文件流写入本地
	// 2. 将文件上传到s3
	idPhoto, idPhotoTmp, code, err := c.UploadResource("fs1", types.Use2IdentityDetect)
	defer tools.Remove(idPhotoTmp)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}
	handHeldIdPhoto, handHeldIdPhotoTmp, code, err := c.UploadResource("fs2", types.Use2IdentityDetect)
	defer tools.Remove(handHeldIdPhotoTmp)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}

	var realname string
	var identity string

	// 3. 调用 advance TODO: 图片尺寸过大压缩,理论应该先去faceid识别一下,但那样有可能响应速度会奇慢
	param := map[string]interface{}{}
	file := map[string]interface{}{
		"ocrImage": idPhotoTmp,
	}
	_, resData, err := advance.Request(c.AccountID, advance.ApiOCR, param, file)
	if err == nil && advance.IsSuccess(resData.Code) {
		ocrRealname := resData.Data.Name
		identity = resData.Data.IDNumber
		// 处理名字中的特殊字符
		realname = tools.TrimRealName(ocrRealname)

		// 身份证号是固定位数,如果识别位数不够,则认为无效,需要重新提交
		if len(realname) > 0 && len(identity) == types.LimitIdentity {
			// 更新基本信息
			_, code, err = service.UpdateAccountBaseV2(c.AccountID, realname, identity, types.GenderSecrecy)
			if err != nil {
				realname = ""
				identity = ""
			} else {
				// 更新OCR识别数据
				service.UpdateAccountBaseOCR(c.AccountID, ocrRealname, identity)

				// 更新用户profile 身份证信息
				service.UpdateAccountProfileIdPhoto(c.AccountID, idPhoto, handHeldIdPhoto)

				// 写待处理队列
				storageClient := storage.RedisStorageClient.Get()
				defer storageClient.Close()
				queueName := beego.AppConfig.String("account_identity_detect")
				storageClient.Do("lpush", queueName, c.AccountID)
			}
		}
	}

	if code != cerror.CodeSuccess {
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}

	// 5. 返回结果给客户端
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"realname":     realname,
		"identity":     identity,
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// IdentityDetectV3 身份证OCR识别v3 增加同盾身份检查
// 由于同盾异步原因，这里只负责创建查询任务
// 与V2相比，只增加了同盾创建查询任务逻辑

// OCR识别-> 手持比对 -> 创建同盾任务
func (c *AccountController) IdentityDetectV3() {
	// 简单判断一下
	if !service.CheckIdentityDetectRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 1. 将文件流写入本地
	// 2. 将文件上传到s3
	idPhoto, idPhotoTmp, code, err := c.UploadResource("fs1", types.Use2IdentityDetect)
	defer tools.Remove(idPhotoTmp)
	if err != nil {
		logs.Debug("[service.IdentityDetectV3] fs1 身份证 error:", err)
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}
	handHeldIdPhoto, handHeldIdPhotoTmp, code, err := c.UploadResource("fs2", types.Use2IdentityDetect)
	defer tools.Remove(handHeldIdPhotoTmp)
	if err != nil {
		logs.Debug("[service.IdentityDetectV3] fs2 手持身份证 error:", err)
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}

	var realname string
	var identity string

	// 3. 调用 advance TODO: 图片尺寸过大压缩,理论应该先去faceid识别一下,但那样有可能响应速度会奇慢
	param := map[string]interface{}{}
	file := map[string]interface{}{
		"ocrImage": idPhotoTmp,
	}
	_, resData, err := advance.Request(c.AccountID, advance.ApiOCR, param, file)

	for {

		if err == nil && advance.IsSuccess(resData.Code) {

			baseAccount, _ := models.OneAccountBaseByPkId(c.AccountID)
			ocrRealname := resData.Data.Name
			identity = resData.Data.IDNumber
			// 处理名字中的特殊字符
			realname = tools.TrimRealName(ocrRealname)
			// 身份证号是固定位数,如果识别位数不够,则认为无效,需要重新提交
			if len(realname) > 0 && len(identity) == types.LimitIdentity {
				// 更新OCR识别数据
				service.UpdateAccountBaseOCR(c.AccountID, ocrRealname, identity)
				// 3.2 ID Holding Photo Check 手持识别
				code, err = advance.IDHoldingPhotoCheck(c.AccountID, handHeldIdPhotoTmp, idPhotoTmp)
				if code != cerror.CodeSuccess {
					break
				}
				if err != nil {
					realname = ""
					identity = ""
				} else {
					// 更新基本信息
					_, code, err = service.UpdateAccountBaseV3(c.AccountID, realname, identity, types.GenderSecrecy, 0)
					if code != cerror.CodeSuccess {
						break
					}
					// 更新用户profile 身份证信息
					service.UpdateAccountProfileIdPhoto(c.AccountID, idPhoto, handHeldIdPhoto)

					//创建同盾身份检查查询任务
					_, _, err = tongdun.CreateTask(c.AccountID, tongdun.IDCheckChannelType, tongdun.IDCheckChannelCode, realname, identity, baseAccount.Mobile)
					if err != nil {
						logs.Error("[IdentityDetectV3] tongdun createTask error：", err)
					}
					// 写待处理队列
					storageClient := storage.RedisStorageClient.Get()
					defer storageClient.Close()
					queueName := beego.AppConfig.String("account_identity_detect")
					storageClient.Do("lpush", queueName, c.AccountID)
				}

			} else {
				//OCR识别错误
				code = cerror.OcrIdentifyError
				break
			}
		} else {
			//OCR识别错误
			code = cerror.OcrIdentifyError
			break
		}

		break
	}

	if code == cerror.IdentityBindRepeated {
		// 身份证号重复时，返回身份证号对应的手机号
		account, err := models.OneAccountBaseByIdentity(identity)
		if err == nil {
			mobile := tools.MobileDesensitization(account.Mobile)
			c.Data["json"] = cerror.BuildApiResponseV2(code, mobile, "")
			c.ServeJSON()
		}
		return
	} else if code != cerror.CodeSuccess {
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}

	// 5. 返回结果给客户端
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"realname":     realname,
		"identity":     identity,
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}

	logs.Debug("[service.IdentityDetectV3] 给客户端返回内容：", data)
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// IdentityDetectTwo（首贷借贷流程变化）
func (c *AccountController) IdentityDetectTwo() {
	// 简单判断一下
	if !service.CheckIdentityDetectRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 1. 将文件流写入本地
	// 2. 将文件上传到s3
	idPhoto, idPhotoTmp, code, err := c.UploadResource("fs1", types.Use2IdentityDetect)
	defer tools.Remove(idPhotoTmp)
	if err != nil {
		logs.Debug("[service.IdentityDetectTwo] fs1 身份证 error:", err)
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}
	handHeldIdPhoto, handHeldIdPhotoTmp, code, err := c.UploadResource("fs2", types.Use2IdentityDetect)
	defer tools.Remove(handHeldIdPhotoTmp)
	if err != nil {
		logs.Debug("[service.IdentityDetectTwo] fs2 手持身份证 error:", err)
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}

	var realname string
	var identity string

	// 3. 调用 advance TODO: 图片尺寸过大压缩,理论应该先去faceid识别一下,但那样有可能响应速度会奇慢
	param := map[string]interface{}{}
	file := map[string]interface{}{
		"ocrImage": idPhotoTmp,
	}
	_, resData, err := advance.Request(c.AccountID, advance.ApiOCR, param, file)

	for {

		if err == nil && advance.IsSuccess(resData.Code) {

			baseAccount, _ := models.OneAccountBaseByPkId(c.AccountID)
			ocrRealname := resData.Data.Name
			identity = resData.Data.IDNumber
			// 处理名字中的特殊字符
			realname = tools.TrimRealName(ocrRealname)
			// 身份证号是固定位数,如果识别位数不够,则认为无效,需要重新提交
			if len(realname) > 0 && len(identity) == types.LimitIdentity {
				// 更新OCR识别数据
				service.UpdateAccountBaseOCR(c.AccountID, ocrRealname, identity)
				// 3.2 ID Holding Photo Check 手持识别
				code, err = advance.IDHoldingPhotoCheck(c.AccountID, handHeldIdPhotoTmp, idPhotoTmp)
				if code != cerror.CodeSuccess {
					break
				}
				if err != nil {
					realname = ""
					identity = ""
				} else {
					// 更新基本信息
					_, code, err = service.UpdateAccountBaseV3(c.AccountID, realname, identity, types.GenderSecrecy, 0)
					if code != cerror.CodeSuccess {
						break
					}
					// 更新用户profile 身份证信息
					service.UpdateAccountProfileIdPhoto(c.AccountID, idPhoto, handHeldIdPhoto)

					//创建同盾身份检查查询任务
					_, _, err = tongdun.CreateTask(c.AccountID, tongdun.IDCheckChannelType, tongdun.IDCheckChannelCode, realname, identity, baseAccount.Mobile)
					if err != nil {
						logs.Error("[IdentityDetectTwo] tongdun createTask error：", err)
					}
					// 写待处理队列
					storageClient := storage.RedisStorageClient.Get()
					defer storageClient.Close()
					queueName := beego.AppConfig.String("account_identity_detect")
					storageClient.Do("lpush", queueName, c.AccountID)
				}

			} else {
				//OCR识别错误
				code = cerror.OcrIdentifyError
				break
			}
		} else {
			//OCR识别错误
			code = cerror.OcrIdentifyError
			break
		}

		break
	}

	if code == cerror.IdentityBindRepeated {
		// 身份证号重复时，返回身份证号对应的手机号
		account, err := models.OneAccountBaseByIdentity(identity)
		if err == nil {
			mobile := tools.MobileDesensitization(account.Mobile)
			c.Data["json"] = cerror.BuildApiResponseV2(code, mobile, "")
			c.ServeJSON()
		}
		return
	} else if code != cerror.CodeSuccess {
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}

	// 5. 返回结果给客户端
	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"realname":     realname,
		"identity":     identity,
		"current_step": phase,
		"progress":     progress,
	}

	logs.Debug("[service.IdentityDetectTwo] 给客户端返回内容：", data)
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) IdentityDetectTwoV2() {
	if !service.CheckIdentityDetectRequiredV2(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 1. 将文件流写入本地
	// 2. 将文件上传到s3
	idPhoto, idPhotoTmp, code, err := c.UploadResource("fs1", types.Use2IdentityDetect)
	defer tools.Remove(idPhotoTmp)
	if err != nil {
		logs.Debug("[service.IdentityDetectTwo] fs1 身份证 error:", err)
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}

	handHeldIdPhoto, handHeldIdPhotoTmp, code, err := c.UploadResource("fs2", types.Use2IdentityDetect)
	defer tools.Remove(handHeldIdPhotoTmp)
	if err != nil {
		logs.Debug("[service.IdentityDetectTwo] fs2 手持身份证 error:", err)
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}

	baseAccount, _ := models.OneAccountBaseByPkId(c.AccountID)

	var realname string
	var ocrRealname string
	var identity string
	var ocrIdentity string
	succ := cerror.CodeSuccess

	isManual, _ := tools.Str2Int(c.RequestJSON["is_manual"].(string))
	if isManual == 0 {
		param := map[string]interface{}{}
		file := map[string]interface{}{
			"ocrImage": idPhotoTmp,
		}
		_, resData, err := advance.Request(c.AccountID, advance.ApiOCR, param, file)
		if err == nil && advance.IsSuccess(resData.Code) {
			ocrRealname = resData.Data.Name
			ocrIdentity = resData.Data.IDNumber
			realname = tools.TrimRealName(ocrRealname)
			identity = resData.Data.IDNumber
		} else {
			succ = cerror.OcrServiceError
		}
	} else {
		realname = c.RequestJSON["realname"].(string)
		identity = c.RequestJSON["identity"].(string)
	}

	if succ == cerror.CodeSuccess {
		if len(realname) > 0 && len(identity) == types.LimitIdentity {
			service.UpdateAccountBaseOCR(c.AccountID, ocrRealname, ocrIdentity)

			succ, err = advance.IDHoldingPhotoCheck(c.AccountID, handHeldIdPhotoTmp, idPhotoTmp)
		} else {
			succ = cerror.OcrIdentifyError
		}
	}

	if succ == cerror.CodeSuccess {
		if err != nil {
			realname = ""
			identity = ""
		} else {
			_, succ, err = service.UpdateAccountBaseV3(c.AccountID, realname, identity, types.GenderSecrecy, isManual)
		}
	}

	if succ == cerror.CodeSuccess {
		service.UpdateAccountProfileIdPhoto(c.AccountID, idPhoto, handHeldIdPhoto)

		//创建同盾身份检查查询任务
		_, _, err = tongdun.CreateTask(c.AccountID, tongdun.IDCheckChannelType, tongdun.IDCheckChannelCode, realname, identity, baseAccount.Mobile)
		if err != nil {
			logs.Error("[IdentityDetectTwo] tongdun createTask error：", err)
		}
		// 写待处理队列
		storageClient := storage.RedisStorageClient.Get()
		defer storageClient.Close()
		queueName := beego.AppConfig.String("account_identity_detect")
		storageClient.Do("lpush", queueName, c.AccountID)
	}

	if succ == cerror.IdentityBindRepeated {
		// 身份证号重复时，返回身份证号对应的手机号
		account, err := models.OneAccountBaseByIdentity(identity)
		if err == nil {
			mobile := tools.MobileDesensitization(account.Mobile)
			c.Data["json"] = cerror.BuildApiResponseV2(succ, mobile, "")
			c.ServeJSON()
		}
		return
	} else if succ != cerror.CodeSuccess {
		c.Data["json"] = cerror.BuildApiResponse(succ, "")
		c.ServeJSON()
		return
	}

	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"realname":     realname,
		"identity":     identity,
		"current_step": phase,
		"progress":     progress,
	}

	logs.Debug("[service.IdentityDetectTwo] 给客户端返回内容：", data)
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// IdentityVerifyV3 验证身份检查结果
func (c *AccountController) IdentityVerifyV3() {

	code := cerror.CodeSuccess
	// 1. 身份检查，同盾，advance身份检测
	verify := service.IdentityVerify(c.AccountID)
	if !verify {
		code = cerror.IdentityVerifyNotPass

		//清空身份证信息
		service.ClearAccountBase(c.AccountID)

		// 更新用户profile 身份证信息
		service.UpdateAccountProfileIdPhoto(c.AccountID, 0, 0)
	}
	if code != cerror.CodeSuccess {
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}
	// 2. 返回结果给客户端
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// IdentityVerifyTwo 验证身份检查结果（首贷借贷流程变化）
func (c *AccountController) IdentityVerifyTwo() {

	code := cerror.CodeSuccess
	// 1. 身份检查，同盾，advance身份检测
	verify := service.IdentityVerify(c.AccountID)
	if !verify {
		code = cerror.IdentityVerifyNotPass

		//清空身份证信息
		service.ClearAccountBase(c.AccountID)

		// 更新用户profile 身份证信息
		service.UpdateAccountProfileIdPhoto(c.AccountID, 0, 0)
	}
	if code != cerror.CodeSuccess {
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}
	// 2. 返回结果给客户端
	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": phase,
		"progress":     progress,
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// IdentityModify 修改身份证对应手机号
func (c *AccountController) IdentityModify() {

	isSuccess := service.IsSuccessModifyMobileByIdentity(c.AccountID)

	// 返回客户端
	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
		"is_success":  isSuccess,
	}

	logs.Debug("[ IdentityModify ] 给客户端返回内容：", data)
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) AccountVerify() {
	// 简单判断一下
	if !service.CheckAccountVerifyRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 1. 上传活体照片到aws
	uploadConf := map[string]UploadResourceResult{}
	files := map[string]string{}
	imageResourceIdMap := map[string]int64{}
	fileKeyMap := map[string]string{
		"fs1": "image_best",
		"fs2": "image_env",
		"fs3": "image_ref1",
		"fs4": "image_ref2",
		"fs5": "image_ref3",
	}
	for i := 1; i <= 5; i++ {
		filename := fmt.Sprintf("fs%d", i)

		urr := UploadResourceResult{}
		urr.ResourceId, urr.TmpFilename, urr.Code, urr.Err = c.UploadResource(filename, types.Use2FaceidVerify)
		defer tools.Remove(urr.TmpFilename)
		if urr.Err != nil {
			c.Data["json"] = cerror.BuildApiResponse(urr.Code, "")
			c.ServeJSON()
			return
		}

		uploadConf[filename] = urr
		files[fileKeyMap[filename]] = urr.TmpFilename
		imageResourceIdMap[fileKeyMap[filename]] = urr.ResourceId
	}
	//fmt.Printf("uploadConf: %#v\n", uploadConf)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"confidence":   float64(0.0000),
		"is_alive":     false,
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}

	// 2. 调用faceid 检测
	originRes, httCode, err := faceid.Verify(c.AccountID, faceid.ComparisonTypeDefault, faceid.FaceImageTypeDefault, files, c.RequestJSON["delta"].(string))
	if err != nil || 200 != httCode {
		c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
		c.ServeJSON()
		return
	}

	// 3. 根据调用结果更新数据
	var confidenceAvg float64 // faceid 识别的平均值
	var isAlive bool
	confidenceAvg, isAlive, err = service.AccountLiveVerify(c.AccountID, imageResourceIdMap, originRes)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
		c.ServeJSON()
		return
	}

	data["confidence"] = confidenceAvg
	data["is_alive"] = isAlive
	data["current_step"] = service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode)

	// 4. 删除临时文件
	for _, k := range uploadConf {
		tools.Remove(k.TmpFilename)
	}

	//! 注: 此处不再改变订单状态,有专门的接口做这件事.2018.03.08
	//// 5. 尝试更改订单状态
	//service.TryConvertTemporaryOrder2Normal(c.AccountID)

	// 6. 返回结果给客户端
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

//AccountVerifyV2 活体检测去掉调用advance活体检测接口
func (c *AccountController) AccountVerifyV2() {
	// 简单判断一下
	if !service.CheckAccountVerifyRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}
	// 1. 上传活体照片到aws
	uploadConf := map[string]UploadResourceResult{}
	files := map[string]string{}
	imageResourceIdMap := map[string]int64{}
	fileKeyMap := map[string]string{
		"fs1": "image_best",
		"fs2": "image_env",
		"fs3": "image_ref1",
		"fs4": "image_ref2",
		"fs5": "image_ref3",
	}
	for i := 1; i <= 5; i++ {
		filename := fmt.Sprintf("fs%d", i)

		urr := UploadResourceResult{}
		urr.ResourceId, urr.TmpFilename, urr.Code, urr.Err = c.UploadResource(filename, types.Use2FaceidVerify)
		defer tools.Remove(urr.TmpFilename)
		if urr.Err != nil {
			c.Data["json"] = cerror.BuildApiResponse(urr.Code, "")
			c.ServeJSON()
			return
		}

		uploadConf[filename] = urr
		files[fileKeyMap[filename]] = urr.TmpFilename
		imageResourceIdMap[fileKeyMap[filename]] = urr.ResourceId
	}

	// 2. 虚构数据, 为了兼容老活体接口，客户端步骤判断需要检查活体数据，所以虚拟一条分值都为99.999的假数据
	JSONData := "{\"face_genuineness\":{\"face_replaced\":0,\"mask_confidence\":0,\"mask_threshold\":0.5,\"screen_replay_confidence\":0,\"screen_replay_threshold\":0.5,\"synthetic_face_confidence\":0,\"synthetic_face_threshold\":0.5},\"request_id\":\"1534817315,e5023067-4e04-4e1a-b921-38264923b893\",\"result_ref1\":{\"confidence\":99.999,\"thresholds\":{\"1e-3\":99.999,\"1e-4\":99.999,\"1e-5\":99.999,\"1e-6\":99.999}},\"result_ref2\":{\"confidence\":99.999,\"thresholds\":{\"1e-3\":99.999,\"1e-4\":99.999,\"1e-5\":99.999,\"1e-6\":99.999}},\"result_ref3\":{\"confidence\":99.999,\"thresholds\":{\"1e-3\":99.999,\"1e-4\":99.999,\"1e-5\":99.999,\"1e-6\":99.999}},\"time_used\":916}"

	// 3. 根据调用结果更新数据
	var confidenceAvg float64 // faceid 识别的平均值
	var isAlive bool
	confidenceAvg, isAlive, _ = service.AccountLiveVerify(c.AccountID, imageResourceIdMap, []byte(JSONData))

	logs.Debug("[AccountVerifyV2] confidenceAvg:", confidenceAvg, "isAlive:", isAlive)

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
		// "confidence":   confidenceAvg,
		"is_alive":     isAlive,
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	// 4. 返回结果给客户端
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

//AccountVerifyTwo 活体检测去掉调用advance活体检测接口（首贷借贷流程变化）
func (c *AccountController) AccountVerifyTwo() {
	// 简单判断一下
	if !service.CheckAccountVerifyRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}
	// 1. 上传活体照片到aws
	uploadConf := map[string]UploadResourceResult{}
	files := map[string]string{}
	imageResourceIdMap := map[string]int64{}
	fileKeyMap := map[string]string{
		"fs1": "image_best",
		"fs2": "image_env",
		"fs3": "image_ref1",
		"fs4": "image_ref2",
		"fs5": "image_ref3",
	}
	for i := 1; i <= 5; i++ {
		filename := fmt.Sprintf("fs%d", i)

		urr := UploadResourceResult{}
		urr.ResourceId, urr.TmpFilename, urr.Code, urr.Err = c.UploadResource(filename, types.Use2FaceidVerify)
		defer tools.Remove(urr.TmpFilename)
		if urr.Err != nil {
			c.Data["json"] = cerror.BuildApiResponse(urr.Code, "")
			c.ServeJSON()
			return
		}

		uploadConf[filename] = urr
		files[fileKeyMap[filename]] = urr.TmpFilename
		imageResourceIdMap[fileKeyMap[filename]] = urr.ResourceId
	}

	// 2. 虚构数据, 为了兼容老活体接口，客户端步骤判断需要检查活体数据，所以虚拟一条分值都为99.999的假数据
	JSONData := "{\"face_genuineness\":{\"face_replaced\":0,\"mask_confidence\":0,\"mask_threshold\":0.5,\"screen_replay_confidence\":0,\"screen_replay_threshold\":0.5,\"synthetic_face_confidence\":0,\"synthetic_face_threshold\":0.5},\"request_id\":\"1534817315,e5023067-4e04-4e1a-b921-38264923b893\",\"result_ref1\":{\"confidence\":99.999,\"thresholds\":{\"1e-3\":99.999,\"1e-4\":99.999,\"1e-5\":99.999,\"1e-6\":99.999}},\"result_ref2\":{\"confidence\":99.999,\"thresholds\":{\"1e-3\":99.999,\"1e-4\":99.999,\"1e-5\":99.999,\"1e-6\":99.999}},\"result_ref3\":{\"confidence\":99.999,\"thresholds\":{\"1e-3\":99.999,\"1e-4\":99.999,\"1e-5\":99.999,\"1e-6\":99.999}},\"time_used\":916}"

	// 3. 根据调用结果更新数据
	var confidenceAvg float64 // faceid 识别的平均值
	var isAlive bool
	confidenceAvg, isAlive, _ = service.AccountLiveVerify(c.AccountID, imageResourceIdMap, []byte(JSONData))

	logs.Debug("[AccountVerifyTwo] confidenceAvg:", confidenceAvg, "isAlive:", isAlive)

	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
		// "confidence":   confidenceAvg,
		"is_alive":     isAlive,
		"current_step": phase,
		"progress":     progress,
	}
	// 4. 返回结果给客户端
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

//AccountVerifyTwo 活体检测去掉调用advance活体检测接口（首贷借贷流程变化）
func (c *AccountController) AccountVerifyTwoV2() {

	// 1. 上传活体照片到aws
	uploadConf := map[string]UploadResourceResult{}
	files := map[string]string{}
	imageResourceIdMap := map[string]int64{}
	fileKeyMap := map[string]string{
		"fs1": "image_best",
		"fs2": "image_env",
		"fs3": "image_ref1",
		"fs4": "image_ref2",
		"fs5": "image_ref3",
	}
	uploadFile := make([]string, 0)
	var isAlive bool
	for i := 1; i <= 5; i++ {
		filename := fmt.Sprintf("fs%d", i)
		_, _, err := c.GetFile(filename)
		if err == nil {
			uploadFile = append(uploadFile, filename)
		}
	}
	logs.Notice("[AccountVerifyTwoV2]uploadFile:", uploadFile)

	for _, filename := range uploadFile {
		urr := UploadResourceResult{}
		urr.ResourceId, urr.TmpFilename, urr.Code, urr.Err = c.UploadResource(filename, types.Use2FaceidVerify)
		defer tools.Remove(urr.TmpFilename)
		if urr.Err != nil {
			c.Data["json"] = cerror.BuildApiResponse(urr.Code, "")
			c.ServeJSON()
			return
		}
		uploadConf[filename] = urr
		files[fileKeyMap[filename]] = urr.TmpFilename
		imageResourceIdMap[fileKeyMap[filename]] = urr.ResourceId
	}

	logs.Debug("[AccountVerifyTwoV2] aws alredy upload files:", files)

	uploadCount := len(uploadFile)
	//上传5张图片走原逻辑
	if uploadCount == 5 {

		// 2. 虚构数据, 为了兼容老活体接口，客户端步骤判断需要检查活体数据，所以虚拟一条分值都为99.999的假数据
		JSONData := "{\"face_genuineness\":{\"face_replaced\":0,\"mask_confidence\":0,\"mask_threshold\":0.5,\"screen_replay_confidence\":0,\"screen_replay_threshold\":0.5,\"synthetic_face_confidence\":0,\"synthetic_face_threshold\":0.5},\"request_id\":\"1534817315,e5023067-4e04-4e1a-b921-38264923b893\",\"result_ref1\":{\"confidence\":99.999,\"thresholds\":{\"1e-3\":99.999,\"1e-4\":99.999,\"1e-5\":99.999,\"1e-6\":99.999}},\"result_ref2\":{\"confidence\":99.999,\"thresholds\":{\"1e-3\":99.999,\"1e-4\":99.999,\"1e-5\":99.999,\"1e-6\":99.999}},\"result_ref3\":{\"confidence\":99.999,\"thresholds\":{\"1e-3\":99.999,\"1e-4\":99.999,\"1e-5\":99.999,\"1e-6\":99.999}},\"time_used\":916}"
		// 3. 根据调用结果更新数据
		var confidenceAvg float64 // faceid 识别的平均值
		confidenceAvg, isAlive, _ = service.AccountLiveVerify(c.AccountID, imageResourceIdMap, []byte(JSONData))
		logs.Debug("[AccountVerifyTwoV2] confidenceAvg:", confidenceAvg, "isAlive:", isAlive)
	}
	//走创蓝
	if uploadCount >= 1 && uploadCount < 5 {

		date := tools.MDateUTC(tools.GetUnixMillis())
		beginDate := date + " 00:00:00"
		endDate := date + " 23:59:59"
		beginTimeStamp, _ := tools.GetTimeParseWithFormat(beginDate, "2006-01-02 15:04:05")
		endTimeStamp, _ := tools.GetTimeParseWithFormat(endDate, "2006-01-02 15:04:05")
		thirdparty.MoveOutThirdpartyStatisticFeeFromCache()
		thirdpartyStatisticFee, _ := dao.GetThirdparthStatisticFeeByMd5("eaf4fb63969ff2cf682622b15b2176c3", beginTimeStamp*1000, endTimeStamp*1000)

		logs.Debug("[AccountVerifyTwoV2] successCallCount:", thirdpartyStatisticFee.CallCount)
		clcount, _ := config.ValidItemInt("api_cl_call_count")
		if thirdpartyStatisticFee.CallCount < (clcount - 1) {
			score, err := api253.FaceCheck(c.AccountID, files[fileKeyMap["fs2"]])
			logs.Notice("[AccountVerifyTwoV2] chuanglan score:", score, "resourceid:", imageResourceIdMap[fileKeyMap["fs2"]])
			if err == nil && score > 87 {
				isAlive = true
				//. 虚构数据, 为了兼容老活体接口，客户端步骤判断需要检查活体数据，所以虚拟一条分值都为100.000的假数据
				JSONData := "{\"face_genuineness\":{\"face_replaced\":0,\"mask_confidence\":0,\"mask_threshold\":0.5,\"screen_replay_confidence\":0,\"screen_replay_threshold\":0.5,\"synthetic_face_confidence\":0,\"synthetic_face_threshold\":0.5},\"request_id\":\"1534817315,e5023067-4e04-4e1a-b921-38264923b893\",\"result_ref1\":{\"confidence\":100.000,\"thresholds\":{\"1e-3\":100.000,\"1e-4\":100.000,\"1e-5\":100.000,\"1e-6\":100.000}},\"result_ref2\":{\"confidence\":100.000,\"thresholds\":{\"1e-3\":100.000,\"1e-4\":100.000,\"1e-5\":100.000,\"1e-6\":100.000}},\"result_ref3\":{\"confidence\":100.000,\"thresholds\":{\"1e-3\":100.000,\"1e-4\":100.000,\"1e-5\":100.000,\"1e-6\":100.000}},\"time_used\":916}"
				service.AccountLiveVerify(c.AccountID, imageResourceIdMap, []byte(JSONData))
			}
		}
	}
	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"is_alive":     isAlive,
		"current_step": phase,
		"progress":     progress,
	}
	// 4. 返回结果给客户端
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

}

//AccountVerifyCL 创蓝活体检测，客户端SDK动作失败才调用此接口
func (c *AccountController) AccountVerifyCL() {

	date := tools.MDateUTC(tools.GetUnixMillis())
	beginDate := date + " 00:00:00"
	endDate := date + " 23:59:59"
	beginTimeStamp, _ := tools.GetTimeParseWithFormat(beginDate, "2006-01-02 15:04:05")
	endTimeStamp, _ := tools.GetTimeParseWithFormat(endDate, "2006-01-02 15:04:05")
	thirdparty.MoveOutThirdpartyStatisticFeeFromCache()
	thirdpartyStatisticFee, _ := dao.GetThirdparthStatisticFeeByMd5("eaf4fb63969ff2cf682622b15b2176c3", beginTimeStamp*1000, endTimeStamp*1000)

	logs.Debug("[AccountVerifyCL] successCallCount:", thirdpartyStatisticFee.CallCount)
	isAlive := false
	isReloan := dao.IsRepeatLoan(c.AccountID)
	clcount, _ := config.ValidItemInt("api_cl_call_count")
	if (thirdpartyStatisticFee.CallCount < (clcount - 1)) && !isReloan {
		accountProfile, _ := dao.GetAccountProfile(c.AccountID)
		handIdPhotoResource, _ := service.OneResource(accountProfile.HandHeldIdPhoto)
		idPhotoTmp := gaws.BuildTmpFilename(accountProfile.HandHeldIdPhoto)
		gaws.AwsDownload(handIdPhotoResource.HashName, idPhotoTmp)
		//方法执行完 删除tmp下的图片
		defer tools.Remove(idPhotoTmp)
		score, err := api253.FaceCheck(accountProfile.AccountId, idPhotoTmp)
		if err == nil && score > 87 {
			isAlive = true

			imageResourceIdMap := map[string]int64{}
			fileKeyMap := map[string]string{
				"fs1": "image_best",
				"fs2": "image_env",
				"fs3": "image_ref1",
				"fs4": "image_ref2",
				"fs5": "image_ref3",
			}
			for i := 1; i <= 5; i++ {
				filename := fmt.Sprintf("fs%d", i)
				imageResourceIdMap[fileKeyMap[filename]] = accountProfile.HandHeldIdPhoto
			}

			//. 虚构数据, 为了兼容老活体接口，客户端步骤判断需要检查活体数据，所以虚拟一条分值都为100.000的假数据
			JSONData := "{\"face_genuineness\":{\"face_replaced\":0,\"mask_confidence\":0,\"mask_threshold\":0.5,\"screen_replay_confidence\":0,\"screen_replay_threshold\":0.5,\"synthetic_face_confidence\":0,\"synthetic_face_threshold\":0.5},\"request_id\":\"1534817315,e5023067-4e04-4e1a-b921-38264923b893\",\"result_ref1\":{\"confidence\":100.000,\"thresholds\":{\"1e-3\":100.000,\"1e-4\":100.000,\"1e-5\":100.000,\"1e-6\":100.000}},\"result_ref2\":{\"confidence\":100.000,\"thresholds\":{\"1e-3\":100.000,\"1e-4\":100.000,\"1e-5\":100.000,\"1e-6\":100.000}},\"result_ref3\":{\"confidence\":100.000,\"thresholds\":{\"1e-3\":100.000,\"1e-4\":100.000,\"1e-5\":100.000,\"1e-6\":100.000}},\"time_used\":916}"
			service.AccountLiveVerify(c.AccountID, imageResourceIdMap, []byte(JSONData))
		}
	}
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"is_alive":     isAlive,
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	// 2. 返回结果给客户端
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

//AccountVerifyCLTwo 创蓝活体检测，客户端SDK动作失败才调用此接口（首贷借贷流程变化）
func (c *AccountController) AccountVerifyCLTwo() {

	date := tools.MDateUTC(tools.GetUnixMillis())
	beginDate := date + " 00:00:00"
	endDate := date + " 23:59:59"
	beginTimeStamp, _ := tools.GetTimeParseWithFormat(beginDate, "2006-01-02 15:04:05")
	endTimeStamp, _ := tools.GetTimeParseWithFormat(endDate, "2006-01-02 15:04:05")
	thirdparty.MoveOutThirdpartyStatisticFeeFromCache()
	thirdpartyStatisticFee, _ := dao.GetThirdparthStatisticFeeByMd5("eaf4fb63969ff2cf682622b15b2176c3", beginTimeStamp*1000, endTimeStamp*1000)

	logs.Debug("[AccountVerifyCLTwo] successCallCount:", thirdpartyStatisticFee.CallCount)
	isAlive := false
	isReloan := dao.IsRepeatLoan(c.AccountID)
	clcount, _ := config.ValidItemInt("api_cl_call_count")
	if (thirdpartyStatisticFee.CallCount < (clcount - 1)) && !isReloan {
		accountProfile, _ := dao.GetAccountProfile(c.AccountID)
		handIdPhotoResource, _ := service.OneResource(accountProfile.HandHeldIdPhoto)
		idPhotoTmp := gaws.BuildTmpFilename(accountProfile.HandHeldIdPhoto)
		gaws.AwsDownload(handIdPhotoResource.HashName, idPhotoTmp)
		//方法执行完 删除tmp下的图片
		defer tools.Remove(idPhotoTmp)
		score, err := api253.FaceCheck(accountProfile.AccountId, idPhotoTmp)
		if err == nil && score > 87 {
			isAlive = true
			imageResourceIdMap := map[string]int64{}
			fileKeyMap := map[string]string{
				"fs1": "image_best",
				"fs2": "image_env",
				"fs3": "image_ref1",
				"fs4": "image_ref2",
				"fs5": "image_ref3",
			}
			for i := 1; i <= 5; i++ {
				filename := fmt.Sprintf("fs%d", i)
				imageResourceIdMap[fileKeyMap[filename]] = accountProfile.HandHeldIdPhoto
			}

			//. 虚构数据, 为了兼容老活体接口，客户端步骤判断需要检查活体数据，所以虚拟一条分值都为100.000的假数据
			JSONData := "{\"face_genuineness\":{\"face_replaced\":0,\"mask_confidence\":0,\"mask_threshold\":0.5,\"screen_replay_confidence\":0,\"screen_replay_threshold\":0.5,\"synthetic_face_confidence\":0,\"synthetic_face_threshold\":0.5},\"request_id\":\"1534817315,e5023067-4e04-4e1a-b921-38264923b893\",\"result_ref1\":{\"confidence\":100.000,\"thresholds\":{\"1e-3\":100.000,\"1e-4\":100.000,\"1e-5\":100.000,\"1e-6\":100.000}},\"result_ref2\":{\"confidence\":100.000,\"thresholds\":{\"1e-3\":100.000,\"1e-4\":100.000,\"1e-5\":100.000,\"1e-6\":100.000}},\"result_ref3\":{\"confidence\":100.000,\"thresholds\":{\"1e-3\":100.000,\"1e-4\":100.000,\"1e-5\":100.000,\"1e-6\":100.000}},\"time_used\":916}"
			service.AccountLiveVerify(c.AccountID, imageResourceIdMap, []byte(JSONData))
		}

	}

	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"is_alive":     isAlive,
		"current_step": phase,
		"progress":     progress,
	}
	// 2. 返回结果给客户端
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) UpdateBase() {
	if !service.CheckUpdateBaseRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	gender, _ := tools.Str2Int(c.RequestJSON["gender"].(string))
	// 更新基本信息
	service.UpdateAccountBase(c.AccountID, c.RequestJSON["realname"].(string), c.RequestJSON["identity"].(string), types.GenderEnum(gender))

	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func getGenderFromIdentity(identity string) (gender types.GenderEnum) {
	// 印尼身份证号码的第7、8位>=40 为女，否则为男
	if len(identity) >= 8 {
		genderStr := identity[6:8]
		genderInt, _ := tools.Str2Int(genderStr)
		if genderInt >= 40 {
			gender = types.GenderFemale
		} else {
			gender = types.GenderMale
		}
	} else {
		logs.Warn("[getGenderFromIdentity] identity length < 8, identity:", identity)
	}

	return
}

func (c *AccountController) UpdateBaseV2() {
	if !service.CheckUpdateBaseRequiredV2(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	identity := c.RequestJSON["identity"].(string)
	gender := getGenderFromIdentity(identity)
	// 更新基本信息
	service.UpdateAccountBase(c.AccountID, c.RequestJSON["realname"].(string), identity, gender)

	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// UpdateBaseTwo（首贷借贷流程变化）
func (c *AccountController) UpdateBaseTwo() {
	if !service.CheckUpdateBaseRequiredV2(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	identity := c.RequestJSON["identity"].(string)
	gender := getGenderFromIdentity(identity)
	// 更新基本信息
	service.UpdateAccountBase(c.AccountID, c.RequestJSON["realname"].(string), identity, gender)

	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": phase,
		"progress":     progress,
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) UpdateWorkInfo() {
	if !service.CheckUpdateWorkInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	jobType, _ := tools.Str2Int(c.RequestJSON["job_type"].(string))
	monthlyIncome, _ := tools.Str2Int(c.RequestJSON["monthly_income"].(string))
	serviceYears, _ := tools.Str2Int(c.RequestJSON["service_years"].(string))
	service.UpdateAccountWorkInfo(c.AccountID, jobType, monthlyIncome, serviceYears, c.RequestJSON["company_name"].(string), c.RequestJSON["company_city"].(string), c.RequestJSON["company_address"].(string))

	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) UpdateWorkInfoV2() {
	if !service.CheckUpdateWorkInfoRequiredV2(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	jobType, _ := tools.Str2Int(c.RequestJSON["job_type"].(string))
	monthlyIncome, _ := tools.Str2Int(c.RequestJSON["monthly_income"].(string))
	serviceYears, _ := tools.Str2Int(c.RequestJSON["service_years"].(string))
	companyTelephone := c.RequestJSON["company_telephone"].(string)
	salaryDay := c.RequestJSON["salary_day"].(string)
	service.UpdateAccountWorkInfoV2(c.AccountID, jobType, monthlyIncome, serviceYears, c.RequestJSON["company_name"].(string), c.RequestJSON["company_city"].(string), companyTelephone, salaryDay)

	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// UpdateWorkInfoTwo（首贷借贷流程变化）
func (c *AccountController) UpdateWorkInfoTwo() {
	if !service.CheckUpdateWorkInfoRequiredV2(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	jobType, _ := tools.Str2Int(c.RequestJSON["job_type"].(string))
	monthlyIncome, _ := tools.Str2Int(c.RequestJSON["monthly_income"].(string))
	serviceYears, _ := tools.Str2Int(c.RequestJSON["service_years"].(string))
	companyTelephone := c.RequestJSON["company_telephone"].(string)
	salaryDay := c.RequestJSON["salary_day"].(string)
	service.UpdateAccountWorkInfoV2(c.AccountID, jobType, monthlyIncome, serviceYears, c.RequestJSON["company_name"].(string),
		c.RequestJSON["company_city"].(string), companyTelephone, salaryDay)

	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": phase,
		"progress":     progress,
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) UpdateBankInfo() {
	if !service.CheckUpdateBankInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	if !service.CanUpdateBankInfo() {
		logs.Warn("[UpdateBankInfo] accountId:%d CanUpdateBankInfo return false", c.AccountID)
		c.Data["json"] = cerror.BuildApiResponse(cerror.ModifyBankFail, "")
		c.ServeJSON()
		return
	}

	// 获取参数
	bankName := c.RequestJSON["bank_name"].(string)
	bankNo := c.RequestJSON["bank_no"].(string)
	if len(bankName) == 0 || len(bankNo) == 0 {
		logs.Error("[UpdateBankInfo] please check bankName:%s bankNo:%s accountId:%d", bankName, bankNo, c.AccountID)
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	order, _ := dao.AccountLastLoanOrder(c.AccountID)
	accoutBaseExt, _ := models.OneAccountBaseExtByPkId(c.AccountID)
	if accoutBaseExt.RecallTag != types.RecallTagModifyBank ||
		order.CheckStatus != types.LoanStatusLoanFail {
		logs.Warn("[UpdateBankInfo] accountId:%d account tag no match. accoutBaseExt:%#v order:%#v", c.AccountID, accoutBaseExt, order)
		c.Data["json"] = cerror.BuildApiResponse(cerror.ModifyBankFail, "")
		c.ServeJSON()
		return
	}

	// 更新信息
	_, err := service.UpdateBankInfo(c.AccountID, c.AccountID, bankName, bankNo)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.ModifyBankFail, "")
		c.ServeJSON()
		return
	}

	// 成功后修改订单状态 以及取消标签
	service.ChangeCustomerRecall(c.AccountID, order.Id, types.RecallTagNone, types.RemarkTagNone)
	service.DoDisbureseAgainBackendV2(c.AccountID, order, types.LoanStatusWait4Loan)

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) UpdateContactInfo() {
	if !service.CheckUpdateContactInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	relationship1, _ := tools.Str2Int(c.RequestJSON["relationship1"].(string))
	relationship2, _ := tools.Str2Int(c.RequestJSON["relationship2"].(string))
	service.UpdateAccountContactInfo(c.AccountID,
		c.RequestJSON["contact1_name"].(string), c.RequestJSON["contact1"].(string),
		c.RequestJSON["contact2_name"].(string), c.RequestJSON["contact2"].(string),
		relationship1, relationship2)

	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// UpdateContactInfoTwo（首贷借贷流程变化）
func (c *AccountController) UpdateContactInfoTwo() {
	if !service.CheckUpdateContactInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	relationship1, _ := tools.Str2Int(c.RequestJSON["relationship1"].(string))
	relationship2, _ := tools.Str2Int(c.RequestJSON["relationship2"].(string))
	service.UpdateAccountContactInfo(c.AccountID,
		c.RequestJSON["contact1_name"].(string), c.RequestJSON["contact1"].(string),
		c.RequestJSON["contact2_name"].(string), c.RequestJSON["contact2"].(string),
		relationship1, relationship2)

	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": phase,
		"progress":     progress,
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) UpdateOtherInfo() {
	if !service.CheckUpdateOtherInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	education, _ := tools.Str2Int(c.RequestJSON["education"].(string))
	maritalStatus, _ := tools.Str2Int(c.RequestJSON["marital_status"].(string))
	childrenNumber, _ := tools.Str2Int(c.RequestJSON["children_number"].(string))
	service.UpdateAccountOtherInfo(c.AccountID, education, maritalStatus, childrenNumber,
		c.RequestJSON["bank_name"].(string),
		c.RequestJSON["bank_no"].(string))

	//! 注: 此处不再改变订单状态,有专门的接口做这件事.2018.03.08
	// 显示的调用,更改订单状态
	//service.TryConvertTemporaryOrder2Normal(c.AccountID)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// UpdateOtherInfoTwo（首贷借贷流程变化）
func (c *AccountController) UpdateOtherInfoTwo() {
	if !service.CheckUpdateOtherInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	education, _ := tools.Str2Int(c.RequestJSON["education"].(string))
	maritalStatus, _ := tools.Str2Int(c.RequestJSON["marital_status"].(string))
	childrenNumber, _ := tools.Str2Int(c.RequestJSON["children_number"].(string))
	service.UpdateAccountOtherInfo(c.AccountID, education, maritalStatus, childrenNumber,
		c.RequestJSON["bank_name"].(string),
		c.RequestJSON["bank_no"].(string))

	//! 注: 此处不再改变订单状态,有专门的接口做这件事.2018.03.08
	// 显示的调用,更改订单状态
	//service.TryConvertTemporaryOrder2Normal(c.AccountID)
	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": phase,
		"progress":     progress,
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// OperatorAcquireCode 获取验证码
func (c *AccountController) OperatorAcquireCode() {
	if !service.CheckOperatorAchieveCode(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	channelType := c.RequestJSON["channel_type"].(string)
	times := service.VerifyTimes(c.AccountID, channelType)
	service.VerifyTimesInc(c.AccountID, channelType, 1)
	accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
	if times >= types.OperatorAcquireCodeMax {
		accountBase.OperatorVerifyStatus = types.OperatorVerifyStatusFailed
		accountBase.Update("operator_verify_status")
		//获取验证码6次失败 直接跳到下一步
		logs.Warn("[AchieveCodeByAccountId]  已经获取验证码%s次失败 直接跳到下一步  channelType:%s", types.OperatorAcquireCodeMax, channelType)
		data := map[string]interface{}{
			"server_time":   tools.GetUnixMillis(),
			"acquire_times": times,
		}
		c.Data["json"] = cerror.BuildApiResponse(cerror.LimitStrategyMobile, data)
		c.ServeJSON()
		return
	}

	//异步调用发送验证码接口
	go func(accountID int64) {
		code, err := service.AchieveCodeByAccountId(accountID, channelType)
		if cerror.CodeSuccess != code {
			logs.Warn("[OperatorAcquireCode] AchieveCodeByAccountId catch err%s", err)
		}
	}(c.AccountID)

	data := map[string]interface{}{
		"server_time":         tools.GetUnixMillis(),
		"acquire_times":       times,
		"identify_code_count": service.GetIdentifyCodeCountByMobile(accountBase.Mobile),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// OperatorVerifyCode 提交验证码给同盾
func (c *AccountController) OperatorVerifyCode() {
	if !service.CheckOperatorVerifyCode(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	channelType := c.RequestJSON["channel_type"].(string)
	codeVerify := c.RequestJSON["code"].(string)
	_, _, code, err := service.VerifyCodeByAccountId(c.AccountID, channelType, codeVerify)
	if cerror.CodeSuccess != code {
		logs.Warn("[OperatorVerifyCode] OperatorVerifyCode catch err%s", err)
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}
	// 更新标志位状态
	accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
	accountBase.OperatorVerifyStatus = types.OperatorVerifyStatusSuccess
	accountBase.OperatorVerifyFinishTime = tools.GetUnixMillis()
	accountBase.Update("operator_verify_status", "operator_verify_finish_time")

	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// OperatorVerifyCode 提交验证码给同盾
func (c *AccountController) OperatorVerifyCodeV2() {
	if !service.CheckOperatorVerifyCode(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	channelType := c.RequestJSON["channel_type"].(string)
	codeVerify := c.RequestJSON["code"].(string)
	_, channelCode, code, err := service.VerifyCodeByAccountId(c.AccountID, channelType, codeVerify)
	if cerror.CodeSuccess != code {
		logs.Warn("[OperatorVerifyCode] OperatorVerifyCode catch err%s", err)
		service.SaveAuthorizeResult(c.AccountID, channelCode, types.AuthorizeStatusFailed)
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}

	service.SaveAuthorizeResult(c.AccountID, channelCode, types.AuthorizeStatusSuccess)
	if channelType == tongdun.IDYYSChannelType {
		// 更新标志位状态
		accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
		timestamp := tools.GetUnixMillis()
		accountBase.LatestSmsVerifyTime = timestamp
		accountBase.OperatorVerifyStatus = types.OperatorVerifyStatusSuccess
		accountBase.OperatorVerifyFinishTime = timestamp
		accountBase.Update("latest_sms_verify_time", "operator_verify_status", "operator_verify_finish_time")
	}

	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// OperatorVerifyCodeTwo 提交验证码给同盾（首贷借贷流程变化）
func (c *AccountController) OperatorVerifyCodeTwo() {
	if !service.CheckOperatorVerifyCode(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	channelType := c.RequestJSON["channel_type"].(string)
	codeVerify := c.RequestJSON["code"].(string)
	finishCraw, channelCode, code, err := service.VerifyCodeByAccountId(c.AccountID, channelType, codeVerify)
	if cerror.CodeSuccess != code {
		if !finishCraw {
			logs.Warn("[OperatorVerifyCodeTwo] OperatorVerifyCode catch err%s", err)
			service.SaveAuthorizeResult(c.AccountID, channelCode, types.AuthorizeStatusFailed)
		}
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}

	service.SaveAuthorizeResult(c.AccountID, channelCode, types.AuthorizeStatusSuccess)
	if channelType == tongdun.IDYYSChannelType {
		// 更新标志位状态
		accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
		timestamp := tools.GetUnixMillis()
		accountBase.LatestSmsVerifyTime = timestamp
		accountBase.OperatorVerifyStatus = types.OperatorVerifyStatusSuccess
		accountBase.OperatorVerifyFinishTime = timestamp
		accountBase.Update("latest_sms_verify_time", "operator_verify_status", "operator_verify_finish_time")
	}

	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"current_step": phase,
		"progress":     progress,
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// TongdongInvokeRecord 客户端调用同盾服务记录
func (c *AccountController) TongdunInvokeRecord() {
	if !service.CheckTongdunInvokeRecord(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}
	request := make(map[string]interface{})
	response := make(map[string]interface{})

	urlInvoke := "/app/tongdun/invoke/"

	channelType := c.RequestJSON["channel_type"].(string)
	channelCode := c.RequestJSON["channel_code"].(string)
	returnCode := c.RequestJSON["return_code"].(string)
	mobile := c.RequestJSON["mobile"].(string)
	taskId := c.RequestJSON["task_id"].(string)

	request["channel_type"] = channelType
	request["channelCode"] = channelCode
	request["returnCode"] = returnCode
	request["mobile"] = mobile
	response["taskId"] = taskId

	models.AddOneThirdpartyRecord(models.ThirdpartyTongdun, urlInvoke+channelType, c.AccountID, request, response, 0, 0, 200)

	// task_id 不为空 说明创建成功
	if len(taskId) > 0 {
		service.SaveAuthorizeResult(c.AccountID, channelCode, types.AuthorizeStatusSuccess)

		tongdunModel, _ := models.GetOneByCondition("task_id", taskId)

		if tongdunModel.ID == 0 {
			code, _ := tools.Str2Int64(returnCode)
			tongdunModel.AccountID = c.AccountID
			tongdunModel.TaskID = taskId
			tongdunModel.Mobile = mobile
			tongdunModel.CheckCode = code
			tongdunModel.Message = ""
			tongdunModel.ChannelType = channelType
			tongdunModel.ChannelCode = channelCode
			tongdunModel.CreateTime = tools.GetUnixMillis() / 1000
			tongdunModel.CreateTimeS = tools.GetLocalDateFormat(tongdunModel.CreateTime*1000, "2006-01-02 15:04:05")

			// 为了应对 和api同时更新此数据记录
			_, err := dao.InsertOrUpdateTongdunManual(tongdunModel)
			if err != nil {
				logs.Error("[TongdunInvokeRecord] InsertOrUpdateTongdun err:%v tongdunModel:%#v", err, tongdunModel)
			}
		}
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// 再次审核
func (c *AccountController) RiskReCheck() {
	if !service.CheckClientInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 2.最后一条订单变为等待审核
	err := service.RiskReCheck(c.AccountID)
	if err != nil {
		logs.Error("[RiskReCheck] RiskReCheck err:%v AccountID:%d", err, c.AccountID)
		c.Data["json"] = cerror.BuildApiResponse(cerror.RiskReCheckError, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

}

func (c *AccountController) PhoneVeiryRefuseRecallHint() {
	timeOption := []string{
		"jam berapa saja",
		"7:00 - 9:00",
		"9:00 - 11:00",
		"11:00 - 14:00",
		"14:00 - 17:00",
		"17:00 - 20:00",
	}
	data := map[string]interface{}{
		"hint":        "Saya bersedia dihubungi customer service untuk diverifikasi pada jam - jam berikut ini",
		"time_option": timeOption,
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

}

// 电核拒绝召回
func (c *AccountController) PhoneVeiryRefuseRecall() {
	if !service.CheckOperatorPhoneVerifyRecall(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}
	reverifyStr := c.RequestJSON["reverify"].(string)
	var callTime string
	if v, ok := c.RequestJSON["call_time"]; ok {
		callTime = v.(string)
	}

	reverify, _ := tools.Str2Int(reverifyStr)

	logs.Debug("[PhoneVeiryRefuseRecall] reverify:%d, callTime:%s", reverify, callTime)
	err := service.PhoneVrifyRefuseRecall(c.AccountID, reverify, callTime)
	if err != nil {
		logs.Error("[PhoneVeiryRefuseRecall] PhoneVeiryRefuseRecall err:%v AccountID:%d", err, c.AccountID)
		c.Data["json"] = cerror.BuildApiResponse(cerror.RiskReCheckError, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

}

// 税号认证
func (c *AccountController) NpwpVerify() {
	if !service.CheckNpwpVerify(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	npwpName := ""
	npwpStatus := 0
	npwpNo := c.RequestJSON["npwp_no"].(string)
	npwpNo = strings.Trim(npwpNo, " ")

	one, _ := models.OneNpwpMobi(npwpNo)
	if one.Id != 0 {
		npwpName = one.CustomerName
		npwpStatus = one.Status
	} else {
		resp, err := npwp.NpwpVerify(c.AccountID, npwpNo)
		if err != nil {
			logs.Error("[NpwpVerify] npwp.NpwpVerify err:%v AccountID:%d npwpNo:%s resp:%#v", err, c.AccountID, npwpNo, resp)
		}

		npwpName = resp.CustomerName
		npwpStatus = resp.Status
	}

	data := map[string]interface{}{
		"verify_result": npwp.GiveReurn(c.AccountID, npwpName, npwpNo, npwpStatus),
		"server_time":   tools.GetUnixMillis(),
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

}

func (c *AccountController) AccountVAInfo() {

	bankCode, eAccountDesc := service.DisplayVAInfoV2(c.AccountID)

	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"bank_code":        bankCode,
		"e_account_number": eAccountDesc,
		"bank_code_list":   service.DisplayBankCode(bankCode),
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) AccountVAInfoV2() {

	bankCode, eAccountDesc := service.DisplayVAInfoV2(c.AccountID)

	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"bank_code":        bankCode,
		"e_account_number": eAccountDesc,
		"bank_code_list":   service.DisplayBankCodeV2(bankCode),
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) ModifyRepayBank() {
	if !service.CheckModifyRepayBankRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	bankCode := c.RequestJSON["bank_code"].(string)
	eAccountNumber, err := service.ModifyRepayBankAndVA(c.AccountID, bankCode)
	if err != nil {
		logs.Warn("[ModifyRepayBank] ModifyRepayBank err:%v AccountID:%d", err, c.AccountID)
		c.Data["json"] = cerror.BuildApiResponse(cerror.ModifyRepayBankAndGetVAFail, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"bank_code":        bankCode,
		"e_account_number": fmt.Sprintf("%s %s", bankCode, eAccountNumber),
		"bank_code_list":   service.DisplayBankCode(bankCode),
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) ModifyRepayBankV2() {
	if !service.CheckModifyRepayBankRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	bankCode := c.RequestJSON["bank_code"].(string)
	eAccountNumber, err := service.ModifyRepayBankAndVA(c.AccountID, bankCode)
	if err != nil {
		logs.Warn("[ModifyRepayBank] ModifyRepayBank err:%v AccountID:%d", err, c.AccountID)
		c.Data["json"] = cerror.BuildApiResponse(cerror.ModifyRepayBankAndGetVAFail, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"bank_code":        bankCode,
		"e_account_number": fmt.Sprintf("%s %s", bankCode, eAccountNumber),
		"bank_code_list":   service.DisplayBankCodeV2(bankCode),
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) ConfigNotLogin() {

	data := map[string]interface{}{}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) ConfigLogin() {

	var pageAfterLive string
	repeatLoanQuota := service.GetRepeatLoanQuota()
	if repeatLoanQuota {
		pageAfterLive = service.GetABTestDividerFlag(c.AccountID)
		service.UpdatePageAfterLiveFlagInAccountBaseExt(c.AccountID, pageAfterLive)
	}

	data := map[string]interface{}{
		"server_time":          tools.GetUnixMillis(),
		"loan_flow_flag":       "",
		"page_after_live_flag": pageAfterLive,
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) AccountVas() {
	vas, err := service.GetAccountVas(c.AccountID)
	if err != nil {
		logs.Warn("[AccountVas] GetAccountVas err:%v AccountID:%d", err, c.AccountID)
		c.Data["json"] = cerror.BuildApiResponse(cerror.AccountGetVAFail, "")

		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
		"va_list":     vas,
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *AccountController) AuthList() {
	isDoneAuth, quotaTotal, persent, authList := service.CustomerAuthorize(c.AccountID)
	logs.Notice("isDoneAuth:", isDoneAuth, "quotaTotal:", quotaTotal, "persent:", persent, "authList:", authList)
	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"is_done_auth": isDoneAuth,
		"quota":        quotaTotal,
		"persent":      persent,
		"auth_list":    authList,
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}
