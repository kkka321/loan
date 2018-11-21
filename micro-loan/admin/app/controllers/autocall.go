package controllers

import (
	"micro-loan/common/service"
)

type AutoCallController struct {
	BaseController
}

func (c *AutoCallController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "autocall"
}

func (c *AutoCallController) AutoCallRecord() {
	mobile := c.GetString("mobile")

	c.Data["mobile"] = mobile

	if len(mobile) > 0 {
		records := service.GetAllAutoCallResult(mobile)

		c.Data["autoCallRecord"] = records
	}

	c.Layout = "layout.html"
	c.TplName = "autocall/record.html"

	return
}
