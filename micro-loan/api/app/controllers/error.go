package controllers

import (
	"github.com/astaxie/beego"

	"micro-loan/common/cerror"
)

type ErrorController struct {
	beego.Controller
}

func (c *ErrorController) Error404() {
	c.Data["json"] = cerror.BuildApiResponse(cerror.ApiNotFound, "api not found, please check out.")
	c.ServeJSON()
}

func (c *ErrorController) Error501() {
	c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceUnavailable, "server error")
	c.ServeJSON()
}

func (c *ErrorController) Error500() {
	c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceUnavailable, "Back-end service is not available")
	c.ServeJSON()
}

func (c *ErrorController) ErrorDb() {
	c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceUnavailable, "database is now down")
	c.ServeJSON()
}
