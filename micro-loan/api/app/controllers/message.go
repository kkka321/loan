package controllers

import (
	"micro-loan/common/cerror"
	"micro-loan/common/pkg/google/push"
	"micro-loan/common/tools"
)

type MessageController struct {
	ApiBaseController
}

func (c *MessageController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()

	// 统一将 ip 加到 RequestJSON 中
	c.RequestJSON["ip"] = c.Ctx.Input.IP()
	c.RequestJSON["related_id"] = int64(0)
}

func (c *MessageController) New() {
	if !push.CheckMessageNewRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	list, num, err := push.AccountNewMessage(c.AccountID)
	if err != nil || num <= 0 {
		push.BuildEmptyMessageData(data, false)
		// 修正偏移量,仿止从头循环
	} else {
		push.BuildMessageData(data, list, false)
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *MessageController) All() {
	if !push.CheckMessageAllRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	msgType, _ := tools.Str2Int(c.RequestJSON["type"].(string))
	offset, _ := tools.Str2Int64(c.RequestJSON["offset"].(string))

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	list, num, err := push.AccountAllMessage(c.AccountID, offset, msgType)
	if err != nil || num <= 0 {
		push.BuildEmptyMessageData(data, true)
		// 修正偏移量,仿止从头循环
		if offset > 0 {
			data["offset"] = tools.Int642Str(offset)
		}
	} else {
		push.BuildMessageData(data, list, true)
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *MessageController) Confirm() {
	if !push.CheckMessageConfirmRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	ids := c.RequestJSON["id"].(string)
	push.AccountConfirmMessage(ids)

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}
