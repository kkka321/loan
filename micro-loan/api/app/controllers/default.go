package controllers

import (
	"github.com/astaxie/beego"

	"micro-loan/common/cerror"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type MainController struct {
	beego.Controller
}

// 以下的两个路由,不走统一的加解密,参数签名检查

func (c *MainController) Get() {
	res := cerror.ApiResponse{
		Code: cerror.CodeSuccess,
		Data: "What are you doing?",
	}

	c.Data["json"] = res
	c.ServeJSON()
}

func (c *MainController) Ping() {
	data := map[string]interface{}{
		"message":     "pong",
		"server_time": tools.GetUnixMillis(),
		"version":     types.AppVersion,
		"head_hash":   tools.GitRevParseHead(),
		//"router":      c.Ctx.Request.RequestURI,
	}
	res := cerror.ApiResponse{
		Code: cerror.CodeSuccess,
		Data: data,
	}

	c.Data["json"] = res
	c.ServeJSON()
}
