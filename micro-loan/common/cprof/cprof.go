package cprof

import (
	"net/http/pprof"

	"github.com/astaxie/beego"
)

type ProfController struct {
	beego.Controller
}

func (c *ProfController) Get() {
	tokenDetect := "hy0kle"
	token := c.GetString("token")
	if token != tokenDetect {
		c.Ctx.Output.Status = 403
		c.Ctx.Output.Header("Content-Type", "text/html")
		c.Ctx.Output.Body([]byte("Access denied"))
		return
	}

	switch c.Ctx.Input.Param(":pp") {
	default:
		pprof.Index(c.Ctx.ResponseWriter, c.Ctx.Request)
	case "":
		pprof.Index(c.Ctx.ResponseWriter, c.Ctx.Request)
	case "cmdline":
		pprof.Cmdline(c.Ctx.ResponseWriter, c.Ctx.Request)
	case "profile":
		pprof.Profile(c.Ctx.ResponseWriter, c.Ctx.Request)
	case "symbol":
		pprof.Symbol(c.Ctx.ResponseWriter, c.Ctx.Request)
	}

	c.Ctx.ResponseWriter.WriteHeader(200)
}
