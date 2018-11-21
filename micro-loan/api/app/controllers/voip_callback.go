package controllers

import (
	"micro-loan/common/service"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

type VoipCallbackController struct {
	beego.Controller
}

func (c *VoipCallbackController) SipBillMessageCB() {

	msg, err := service.SipBillMessageCallBack(c.Ctx.Input.RequestBody)
	if err != nil {
		c.Ctx.Output.Status = 200
		logs.Error("[SipBillMessageCB] err:", err)
	}

	c.Ctx.WriteString(msg)

	return

}
