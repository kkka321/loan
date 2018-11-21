package controllers

import (
	//"github.com/astaxie/beego/logs"

	"micro-loan/common/service"
	"micro-loan/common/thirdparty/appsflyer"

	"github.com/astaxie/beego"
)

type AppsflyerController struct {
	beego.Controller
}

func (c *AppsflyerController) Install() {
	req := c.Ctx.Request
	originData, err := appsflyer.ParseOrigin(req)
	if err != nil {
		c.CustomAbort(400, "Invalid Request")
		return
	}
	service.CreateAccountOriginByAppsflyerPush(originData)
}
