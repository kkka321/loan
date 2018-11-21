package controllers

import (
	"micro-loan/common/cerror"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type AdPositionController struct {
	ApiBaseController
}

func (c *AdPositionController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()
}

func (c *AdPositionController) GetAdPosition() {
	data, err := service.GetAdPositionDisplay(c.AccountID, types.AdPositionRejectPage)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.AdPositionGetCodeFail, "")
		c.ServeJSON()
		return
	}

	data["server_time"] = tools.GetUnixMillis()

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}
