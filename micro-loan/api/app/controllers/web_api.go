/**
h5 导流,目前的需求,只有登陆/注册,做好防刷
因为h5页面拿不到用户其他有效信息,无法时行大数据风控,所以,只是导流
*/

package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/cerror"
	"micro-loan/common/models"
	"micro-loan/common/pkg/accesstoken"
	"micro-loan/common/pkg/coupon_event"
	"micro-loan/common/pkg/schema_task"
	"micro-loan/common/service"
	"micro-loan/common/strategy/limit"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type WebApiController struct {
	beego.Controller

	// request json
	RequestJSON map[string]interface{}
	// 有效token对应的用户账户
	AccountID int64

	isTrace   bool
	beginTime int64
}

func (c *WebApiController) BuildWebApiResponse(code cerror.ErrCode, data interface{}) cerror.ApiResponse {
	r := cerror.ApiResponse{
		Code: code,
		Data: data,
	}
	logs.Debug(">>> ResponseCode:", code, ", >>> ResponseJSON:", r)

	return r
}

func (c *WebApiController) Prepare() {
	////维护公告
	//c.Data["json"] = c.BuildWebApiResponse(cerror.ServiceIsDown, "")
	//c.ServeJSON()
	//return

	rv := tools.GenerateRandom(0, 100)
	rate, _ := beego.AppConfig.Int("monitor_api_trace_rate")
	if rv < rate {
		c.isTrace = true
		c.beginTime = tools.GetUnixMillis()
	}

	data := c.GetString("data")
	if len(data) < 16 {
		logs.Warning("post data is empty.")
		c.Data["json"] = c.BuildWebApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 为了联调,先打出来
	logs.Debug(">>> origData:", data)

	var reqJSON map[string]interface{}
	err := json.Unmarshal([]byte(data), &reqJSON)
	if err != nil {
		logs.Warning("cat NOT json decode request data:", data)
		c.Data["json"] = c.BuildWebApiResponse(cerror.InvalidRequestData, "")
		c.ServeJSON()
		return
	}

	// json decode 通过
	c.RequestJSON = reqJSON

	// 必要参数检查,只检查存在,没有判值
	requiredParameter := map[string]bool{
		"noise":        true,
		"request_time": true,
		"access_token": true,
	}
	var requiredCheck int = 0
	for k, _ := range reqJSON {
		if requiredParameter[k] {
			requiredCheck++
		}
	}
	if len(requiredParameter) != requiredCheck {
		logs.Warning("request json lost required parameter, json:", data)
		c.Data["json"] = c.BuildWebApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	uri := c.Ctx.Request.RequestURI
	// 以下路由不需要持有 token
	notNeedTokenRoute := map[string]bool{
		"/webapi/v1/request_login_auth_code": true,
		"/webapi/v1/login":                   true,
		"/webapi/v2/request_login_auth_code": true,
		"/webapi/v2/login":                   true,
	}
	if !notNeedTokenRoute[uri] {
		// 检查 token 有效性
		ok, accountId := accesstoken.IsValidAccessToken(types.PlatformH5, reqJSON["access_token"].(string))
		if !ok {
			logs.Notice("access_token is invalid, json:", data)
			c.Data["json"] = c.BuildWebApiResponse(cerror.InvalidAccessToken, "")
			c.ServeJSON()
			return
		}

		c.AccountID = accountId
	}

	c.RequestJSON["ip"] = c.Ctx.Input.IP()
}

func (c *WebApiController) Finish() {
	if c.isTrace {
		service.AddApiTraceData(c.beginTime, c.Ctx.Request.URL.String())
	}
}

func (c *WebApiController) RequestLoginAuthCode() {
	if !service.CheckLoginAuthCodeRequired(c.RequestJSON) {
		c.Data["json"] = c.BuildWebApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 手机号码校验
	ok, _ := tools.IsValidIndonesiaMobile(c.RequestJSON["mobile"].(string))
	if !ok {
		c.Data["json"] = c.BuildWebApiResponse(cerror.InvalidMobile, "")
		c.ServeJSON()
		return
	}

	// 60秒内重复请求限制 TODO

	serviceType := types.ServiceRequestLogin
	authCodeType := types.AuthCodeTypeText
	// 过限制策略
	if limit.MobileStrategy(c.RequestJSON["mobile"].(string), serviceType, authCodeType) {
		c.Data["json"] = c.BuildWebApiResponse(cerror.LimitStrategyMobile, "")
		c.ServeJSON()
		return
	}

	// 调用短信服务
	if !service.SendSms(serviceType, authCodeType, c.RequestJSON["mobile"].(string), c.Ctx.Input.IP()) {
		c.Data["json"] = c.BuildWebApiResponse(cerror.SMSServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	c.Data["json"] = c.BuildWebApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *WebApiController) IsLogin() {
	data := map[string]interface{}{
		"is_login":    1,
		"server_time": tools.GetUnixMillis(),
	}

	c.Data["json"] = c.BuildWebApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *WebApiController) Login() {
	if !service.CheckWebApiLoginRequired(c.RequestJSON) {
		c.Data["json"] = c.BuildWebApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 手机号码校验
	ok, _ := tools.IsValidIndonesiaMobile(c.RequestJSON["mobile"].(string))
	if !ok {
		c.Data["json"] = c.BuildWebApiResponse(cerror.InvalidMobile, "")
		c.ServeJSON()
		return
	}

	// 验证 auth_code 有效性
	ok = service.CheckSmsCode(c.RequestJSON["mobile"].(string), c.RequestJSON["auth_code"].(string))
	if !ok {
		c.Data["json"] = c.BuildWebApiResponse(cerror.InvalidAuthCode, "")
		c.ServeJSON()
		return
	}

	// 注册新用户或老用户登陆,并将 access_token 返回给客户端
	c.RequestJSON["platform"] = types.PlatformH5
	accountId, accessToken, _, err := service.RegisterOrLogin(c.RequestJSON)
	if err != nil {
		c.Data["json"] = c.BuildWebApiResponse(cerror.ServiceUnavailable, "")
		c.ServeJSON()
		return
	}
	//注册成功发送短信
	//service.SendMessage(types.ServiceRegisterOrLogin, c.RequestJSON["mobile"].(string))
	param := make(map[string]interface{})
	schema_task.SendBusinessMsg(types.SmsTargetH5Register, types.ServiceRegisterOrLogin, c.RequestJSON["mobile"].(string), param)

	c.AccountID = accountId // 登陆成功时,将id设置为正确值

	// 如果是新用户,创建profile
	service.InitAccountProfile(accountId)

	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"access_token": accessToken,
	}

	c.Data["json"] = c.BuildWebApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *WebApiController) RequestLoginAuthCodeV2() {
	if !service.CheckLoginAuthCodeRequired(c.RequestJSON) {
		c.Data["json"] = c.BuildWebApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 手机号码校验
	ok, _ := tools.IsValidIndonesiaMobile(c.RequestJSON["mobile"].(string))
	if !ok {
		c.Data["json"] = c.BuildWebApiResponse(cerror.InvalidMobile, "")
		c.ServeJSON()
		return
	}

	_, err := models.OneAccountBaseByMobile(c.RequestJSON["mobile"].(string))
	if err == nil {
		c.Data["json"] = c.BuildWebApiResponse(cerror.MobileHasRegistered, "")
		c.ServeJSON()
		return
	}

	// 60秒内重复请求限制 TODO

	serviceType := types.ServiceRequestLogin
	authCodeType := types.AuthCodeTypeText
	// 过限制策略
	if limit.MobileStrategy(c.RequestJSON["mobile"].(string), serviceType, authCodeType) {
		c.Data["json"] = c.BuildWebApiResponse(cerror.LimitStrategyMobile, "")
		c.ServeJSON()
		return
	}

	// 调用短信服务
	if !service.SendSms(serviceType, authCodeType, c.RequestJSON["mobile"].(string), c.Ctx.Input.IP()) {
		c.Data["json"] = c.BuildWebApiResponse(cerror.SMSServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	c.Data["json"] = c.BuildWebApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *WebApiController) LoginV2() {
	if !service.CheckWebApiLoginRequiredV2(c.RequestJSON) {
		c.Data["json"] = c.BuildWebApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 手机号码校验
	ok, _ := tools.IsValidIndonesiaMobile(c.RequestJSON["mobile"].(string))
	if !ok {
		c.Data["json"] = c.BuildWebApiResponse(cerror.InvalidMobile, "")
		c.ServeJSON()
		return
	}

	_, err := models.OneAccountBaseByMobile(c.RequestJSON["mobile"].(string))
	if err == nil {
		c.Data["json"] = c.BuildWebApiResponse(cerror.MobileHasRegistered, "")
		c.ServeJSON()
		return
	}

	// 验证 auth_code 有效性
	ok = service.CheckSmsCode(c.RequestJSON["mobile"].(string), c.RequestJSON["auth_code"].(string))
	if !ok {
		c.Data["json"] = c.BuildWebApiResponse(cerror.InvalidAuthCode, "")
		c.ServeJSON()
		return
	}

	// 注册新用户或老用户登陆,并将 access_token 返回给客户端
	c.RequestJSON["platform"] = types.PlatformH5
	accountId, accessToken, isNew, err := service.RegisterOrLogin(c.RequestJSON)
	if err != nil {
		c.Data["json"] = c.BuildWebApiResponse(cerror.ServiceUnavailable, "")
		c.ServeJSON()
		return
	}
	//注册成功发送短信
	//service.SendMessage(types.ServiceRegisterOrLogin, c.RequestJSON["mobile"].(string))
	param := make(map[string]interface{})
	schema_task.SendBusinessMsg(types.SmsTargetH5Register, types.ServiceRegisterOrLogin, c.RequestJSON["mobile"].(string), param)

	c.AccountID = accountId // 登陆成功时,将id设置为正确值

	// 如果是新用户,创建profile
	service.InitAccountProfile(accountId)

	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"access_token": accessToken,
	}

	if isNew {
		handleNewAccount(accountId, c.RequestJSON)
	}

	c.Data["json"] = c.BuildWebApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func handleNewAccount(newAccountId int64, data map[string]interface{}) {
	logs.Debug("[handleNewAccount] begin newAccount:%d, data:%v", newAccountId, data)

	inviteStr, ok1 := data["invite"].(string)
	opStr, ok2 := data["op"].(string)
	if !ok1 || !ok2 {
		logs.Warn("[handleNewAccount] get data error, ok1:%v, data:%v", ok1, data)
		return
	}

	inviteId, err1 := tools.Str2Int64(inviteStr)
	op, err2 := tools.Str2Int(opStr)
	if err1 != nil || err2 != nil {
		logs.Warn("[handleNewAccount] format data error, err1:%v, data:%v", err1, data)
		return
	}

	{
		param := coupon_event.InviteEventParam{}
		param.NewAccountId = newAccountId
		param.InviteId = inviteId
		param.InviteType = op
		//service.HandleCouponEvent(coupon_event.TriggerWebRegister, param)
	}

	{
		param := coupon_event.InviteV3Param{}
		param.AccountId = newAccountId
		param.InviteId = inviteId
		param.TaskType = types.AccountTaskRegister
		service.HandleCouponEvent(coupon_event.TriggerInviteV3, param)
	}
}
