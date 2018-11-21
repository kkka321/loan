package controllers

import (

	"micro-loan/common/cerror"
	"micro-loan/common/service"
	"micro-loan/common/strategy/limit"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type PopOrFloatController struct {
	ApiBaseController
}

func (c *PopOrFloatController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()

}

func (c *PopOrFloatController) PopoversOrFloatWindow() {
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
