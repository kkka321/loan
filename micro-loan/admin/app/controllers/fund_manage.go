package controllers

import (
	"github.com/astaxie/beego/logs"

	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type FundController struct {
	BaseController
}

func (c *FundController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "fund"
}

func (c *FundController) LoanConfig() {
	c.Data["FundCodeNameMap"] = types.FundCodeNameMap()

	c.Layout = "layout.html"
	c.TplName = "fund_manage/loan_assign.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "fund_manage/loan_assign_scripts.html"
}

func (c *FundController) RepayConfig() {

	c.Data["FundCodeNameMap"] = types.FundCodeNameMap()

	c.Layout = "layout.html"
	c.TplName = "fund_manage/repay_assign.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "fund_manage/loan_assign_scripts.html"
}

func (c *FundController) BankQuery() {
	// return data
	resp := map[string]interface{}{}

	fundStr := c.GetString("fund_id")
	fundId, _ := tools.Str2Int(fundStr)

	loanRepay := c.GetString("loan_repay_type")
	loanrRepayType, _ := tools.Str2Int(loanRepay)
	logs.Info("[BankQuery] fundStr:%s loanRepay:%s", fundStr, loanRepay)

	assignList, unAssignList, allUnAssignedList, err := service.BankList(fundId, loanrRepayType)
	if err != nil {
		logs.Error("[BankQuery] BankList err:%v", err)
		c.ServeJSON()
		return
	}

	resp["AssignList"] = assignList
	resp["UnAssignList"] = unAssignList
	resp["AllUnAssignList"] = allUnAssignedList
	c.Data["json"] = resp

	c.ServeJSON()
	return
}

func (c *FundController) BankAssign() {
	resp := map[string]interface{}{}

	fundStr := c.GetString("fund_id")
	fundId, _ := tools.Str2Int(fundStr)

	loanRepay := c.GetString("loan_repay_type")
	loanrRepayType, _ := tools.Str2Int(loanRepay)

	assignOperations := c.GetStrings("assign_operations[]")

	logs.Info("[BankAssign] fundStr:%s loanRepay:%s assignOperations:%#v", fundStr, loanRepay, assignOperations)

	err := service.BankAssign(fundId, loanrRepayType, assignOperations)
	if err != nil {
		logs.Error("[BankAssign] BankAssign err:%v", err)

		resp["error"] = "BankAssign error"
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}

	assignList, unAssignList, allUnAssignedList, err := service.BankList(fundId, loanrRepayType)
	if err != nil {
		logs.Error("[BankAssign] BankList err:%v", err)
		c.ServeJSON()
		return
	}

	resp["AssignList"] = assignList
	resp["UnAssignList"] = unAssignList
	resp["AllUnAssignList"] = allUnAssignedList
	c.Data["json"] = resp

	c.ServeJSON()
	return
}

func (c *FundController) BankUnAssign() {
	resp := map[string]interface{}{}

	fundStr := c.GetString("fund_id")
	fundId, _ := tools.Str2Int(fundStr)

	loanRepay := c.GetString("loan_repay_type")
	loanrRepayType, _ := tools.Str2Int(loanRepay)

	assignOperations := c.GetStrings("assign_operations[]")

	logs.Info("[BankAssign] fundStr:%s loanRepay:%s assignOperations:%#v", fundStr, loanRepay, assignOperations)

	err := service.BankUnAssign(fundId, loanrRepayType, assignOperations)
	if err != nil {
		logs.Error("[BankAssign] BankAssign err:%v", err)

		resp["error"] = "BankAssign error"
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}

	assignList, unAssignList, allUnAssignedList, err := service.BankList(fundId, loanrRepayType)
	if err != nil {
		logs.Error("[BankAssign] BankList err:%v", err)
		c.ServeJSON()
		return
	}

	resp["AssignList"] = assignList
	resp["UnAssignList"] = unAssignList
	resp["AllUnAssignList"] = allUnAssignedList
	c.Data["json"] = resp

	c.ServeJSON()
	return
}
