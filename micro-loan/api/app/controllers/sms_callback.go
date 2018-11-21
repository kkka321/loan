package controllers

import (
	//"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/sms"

	"github.com/astaxie/beego"
)

type SmsCallbackController struct {
	beego.Controller
}

func (c *SmsCallbackController) Delivery() {
	req := c.Ctx.Request
	smsEncryptKey := c.Ctx.Input.Param(":smsEncryptKey")
	sms.HandleDelivery(smsEncryptKey, req)
}
