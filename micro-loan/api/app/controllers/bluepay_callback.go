package controllers

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/service"
	"micro-loan/common/tools"
)

type BluePayCallbackController struct {
	beego.Controller
}

func (c *BluePayCallbackController) CallBack() {
	var err error

	pos := strings.Index(c.Ctx.Request.URL.RawQuery, "encrypt")
	encrypt := c.GetString("encrypt")
	paramStr := tools.SubString(c.Ctx.Request.URL.RawQuery, 0, pos-1)
	interfacetype := c.GetString("interfacetype")
	status := c.GetString("status")
	orderId, _ := c.GetInt64("t_id")
	bankCode := c.GetString("operator")

	secretKey := beego.AppConfig.String("bluepay_secret_key")
	md5Params := tools.Md5(fmt.Sprintf("%s%s", paramStr, secretKey))

	if md5Params != encrypt {
		c.Ctx.Output.Status = 401
		logs.Error("[BluePayCallbackController] md5 failed url:", c.Ctx.Request.URL.RawQuery)
		return
	}

	if interfacetype == "bank" {
		if status == "201" {
			err = service.BluepayCreateVirtualAccountCallback(orderId, "/bluepay/callback", c.Ctx.Request.URL.RawQuery)
		} else if status == "200" {
			eAccountNumber := c.GetString("paytype")
			amount, _ := c.GetInt64("price")
			err = service.BluepayPaymentCallback(eAccountNumber, amount, "/bluepay/callback", c.Ctx.Request.URL.RawQuery)
		} else {
			errStr := fmt.Sprintf("[BluePayCallbackController] status error interfacetype:%s, status:%s", interfacetype, status)
			err = fmt.Errorf(errStr)
		}
	} else if interfacetype == "cashout" {
		if status == "200" {
			err = service.BluepayDisburseCallback(orderId, bankCode, status, "/bluepay/callback", c.Ctx.Request.URL.RawQuery)
		} else {
			errStr := fmt.Sprintf("[BluePayCallbackController] status error interfacetype:%s, status:%s", interfacetype, status)
			err = fmt.Errorf(errStr)
		}
	} else {
		errStr := fmt.Sprintf("[BluePayCallbackController] interfacetype error interfacetype:%s, status:%s", interfacetype, status)
		err = fmt.Errorf(errStr)
	}

	if err != nil {
		c.Ctx.Output.Status = 401
		logs.Error("[BluePayCallbackController] err:%s, url:", err, c.Ctx.Request.URL.RawQuery)
	}
}
