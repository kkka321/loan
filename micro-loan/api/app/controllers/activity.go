package controllers

import (
	"micro-loan/common/cerror"
	"micro-loan/common/service"
)

type ActivityController struct {
	ApiBaseController
}

func (c *ActivityController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()
}

//获取弹窗
func (c *ActivityController) GetPopoversor() {

	data, err := service.GetPopoversor()
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.PopGetCodeFail, "")
		c.ServeJSON()
		return
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

//获取浮窗
func (c *ActivityController) GetFloating() {
	if !service.CheckGetFloatingRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	etype := c.RequestJSON["etype"].(string)
	data, err := service.GetFloating(etype)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.FloatingGetCodeFail, "")
		c.ServeJSON()
		return
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}
