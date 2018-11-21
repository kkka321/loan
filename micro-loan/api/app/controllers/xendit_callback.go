package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/service"
)

type XenditCallbackController struct {
	beego.Controller
}

func (c *XenditCallbackController) VirtualAccountCreate() {
	err := service.XenditCreateVirtualAccountCallback("/xendit/virtual_account_callback/create", c.Ctx.Input.RequestBody)
	if err != nil {
		c.Ctx.Output.Status = 401
		logs.Error("[XenditVirtualAccountController] err:", err)
	}

	return
}

func (c *XenditCallbackController) DisburseFundCreate() {
	err := service.XenditDisburseCallback("/xendit/disburse_fund_callback/create", c.Ctx.Input.RequestBody)
	if err != nil {
		c.Ctx.Output.Status = 401
		logs.Error("[XenditDisburseFundController] err:", err)
	}

	return
}

func (c *XenditCallbackController) FVAReceivePaymentCreate() {
	logs.Debug(string(c.Ctx.Input.RequestBody))
	err := service.XenditPaymentCallback("/xendit/fva_receive_payment_callback/create", c.Ctx.Input.RequestBody)
	if err != nil {
		c.Ctx.Output.Status = 401
		logs.Error("[XenditFVAReceivePaymentController] err:", err)
	}

	return
}

func (c *XenditCallbackController) MarketReceivePaymentCreate() {
	err := service.XenditMarketPaymentCallback("/xendit/market_receive_payment_callback/create", c.Ctx.Input.RequestBody)
	if err != nil {
		c.Ctx.Output.Status = 401
		logs.Error("[XenditMarketReceivePaymentCreate] err:", err)
	}

	return
}

func (c *XenditCallbackController) FixPaymentcodeCreate() {
	err := service.XenditFixPaymentCodeCallback("/xendit/fix_payment_code_callback/create", c.Ctx.Input.RequestBody)
	if err != nil {
		c.Ctx.Output.Status = 401
		logs.Error("[XenditFixPaymentcodeCreate] err:", err)
	}

	return
}
