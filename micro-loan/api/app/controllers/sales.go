package controllers

import (
	"micro-loan/common/cerror"
	"micro-loan/common/service/sales"
	"micro-loan/common/tools"
	"strings"

	"github.com/astaxie/beego/logs"
)

type SalesController struct {
	ApiBaseController
}

func (c *SalesController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()

	// 统一将 ip 加到 RequestJSON 中
	c.RequestJSON["ip"] = c.Ctx.Input.IP()
	c.RequestJSON["related_id"] = int64(0)
}

func (c *SalesController) InviteInfo() {
	if !sales.CheckInviteInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	clientTag := 0
	if v, ok := c.RequestJSON["tag"]; ok && v != nil {
		clientTag, _ = tools.Str2Int(v.(string))
	}

	sales.QueryAccountInviteInfo(c.AccountID, clientTag, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *SalesController) Invite() {
	if !sales.CheckInviteRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	mobilesStr, ok := c.RequestJSON["mobile_list"].(string)
	if !ok {
		logs.Debug("[Invite] get mobiles error data:%v", c.RequestJSON)
	}

	logs.Debug("[Invite] get mobiles str:%s", mobilesStr)
	mobiles := strings.Split(mobilesStr, ",")

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	clientTag := 0
	if v, ok := c.RequestJSON["tag"]; ok && v != nil {
		clientTag, _ = tools.Str2Int(v.(string))
	}

	result := sales.SendInviteMessage(c.AccountID, clientTag, mobiles)
	data["result"] = result

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *SalesController) InviteList() {
	if !sales.CheckInviteListRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	clientTag := 0
	if v, ok := c.RequestJSON["tag"]; ok && v != nil {
		clientTag, _ = tools.Str2Int(v.(string))
	}

	sales.QueryAccountInviteList(c.AccountID, clientTag, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}
