package controllers

import (
	"micro-loan/common/cerror"
	"micro-loan/common/service"
	"micro-loan/common/tools"
)

type ProductController struct {
	ApiBaseController
}

func (c *ProductController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()

	// 统一将 ip 加到 RequestJSON 中
	c.RequestJSON["ip"] = c.Ctx.Input.IP()
	c.RequestJSON["related_id"] = int64(0)
}

func (c *ProductController) ProductInfoV1() {
	if !service.CheckClientInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 未登录状态调用
	data := map[string]interface{}{
		"server_time":      tools.GetUnixMillis(),
		"product_suitable": service.ProductSuitablesForApp(c.AccountID),
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}
