package controllers

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils/pagination"

	"micro-loan/common/i18n"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/xendit"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type LoanController struct {
	BaseController
}

func (c *LoanController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "loan"
}

// 放款管理列表
func (c *LoanController) List() {
	c.Data["Action"] = "list"
	c.TplName = "loan/list.html"

	var condCntr = map[string]interface{}{}
	var loanTotalNum int64
	id := c.GetString("id")
	if len(id) > 0 {
		condCntr["id"] = id
	}
	accountIDStr := c.GetString("account_id", "")
	accountID, _ := tools.Str2Int64(accountIDStr)
	if accountID > 0 {
		condCntr["account_id"] = accountID
	}

	realname := c.GetString("realname")
	if len(realname) > 0 {
		condCntr["realname"] = realname
	}
	bankname := c.GetString("bankname")
	bankname = strings.TrimSpace(bankname)
	if len(bankname) > 0 {
		condCntr["bankname"] = bankname
	}

	loanChannel, _ := c.GetInt("loan_channel")
	if loanChannel > 0 {
		condCntr["loan_channel"] = loanChannel
	}

	failedCode, _ := c.GetInt("failed_code")
	if failedCode > 0 {
		condCntr["failed_code"] = failedCode
	}

	checkStatusMulti := c.GetStrings("check_status")
	if len(checkStatusMulti) > 0 {
		condCntr["check_status"] = checkStatusMulti
	}

	// applyStartTimeStr := c.GetString("apply_start_time")
	// applyStartTime := tools.GetTimeParse(applyStartTimeStr)
	// if applyStartTime > 0 {
	// 	condCntr["apply_start_time"] = applyStartTime * 1000
	// }
	//
	// applyEndTimeStr := c.GetString("apply_end_time")
	// applyEndTime := tools.GetTimeParse(applyEndTimeStr)
	// if applyEndTime > 0 {
	// 	condCntr["apply_end_time"] = applyEndTime * 1000
	// }

	mobile := c.GetString("mobile")
	if len(mobile) > 0 {
		condCntr["mobile"] = mobile
	}

	splitSep := " - "
	// s申请时间范围
	applyTimeRange := c.GetString("apply_time_range")
	if len(applyTimeRange) > 16 {
		tr := strings.Split(applyTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["apply_start_time"] = timeStart
				condCntr["apply_end_time"] = timeEnd
			}
		}
	}
	c.Data["applyTimeRange"] = applyTimeRange

	// 还款时间范围
	loanTimeRange := c.GetString("loan_time_range")
	if len(loanTimeRange) > 16 {
		tr := strings.Split(loanTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["loan_start_time"] = timeStart
				condCntr["loan_end_time"] = timeEnd
			}
		}
	}
	c.Data["loanTimeRange"] = loanTimeRange

	// 结清时间范围
	finishTimeRange := c.GetString("finish_time_range")
	if len(finishTimeRange) > 16 {
		tr := strings.Split(finishTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["finish_start_time"] = timeStart
				condCntr["finish_end_time"] = timeEnd
			}
		}
	}

	sortfield := c.GetString("field")
	if len(sortfield) > 0 {
		condCntr["field"] = sortfield
	}

	sorttype := c.GetString("sort")
	if len(sorttype) > 0 {
		condCntr["sort"] = sorttype
	}

	c.Data["finishTimeRange"] = finishTimeRange

	c.Data["id"] = id
	c.Data["mobile"] = mobile
	c.Data["account_id"] = accountIDStr
	c.Data["realname"] = realname
	c.Data["check_status"] = types.LoanStatusMap()
	c.Data["statusSelectMultiBox"] = service.BuildJsVar("statusSelectMultiBox", checkStatusMulti)
	// c.Data["apply_start_time"] = applyStartTimeStr
	c.Data["LoanStatusMap"] = types.LoanStatusMap()
	c.Data["LoanChannel"] = loanChannel
	c.Data["FailedCode"] = failedCode
	c.Data["Bankname"] = bankname

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	list, count, loanTotalNum, loanTotalNumSuccess := service.LoanListBackend(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, int64(count))

	for k, v := range *list {
		dLog, err := models.GetLastestDisburseInvorkLogByPkOrderId(v.Id)
		if err == nil {
			(*list)[k].DisbursementId = dLog.DisbursementId
		}
		(*list)[k].LoanCompany = dLog.VaCompanyCode
	}

	// 放款总金额
	if len(condCntr) == 0 {
		loanTotalNum = 0
		loanTotalNumSuccess = 0
	}
	c.Data["loanTotal"] = loanTotalNum
	c.Data["loanTotalSuccess"] = loanTotalNumSuccess

	c.Data["paginator"] = paginator
	c.Data["List"] = list
	c.Data["RiskTypeMap"] = types.RiskTypeMap()
	c.Data["FundCodeNameMap"] = types.FundCodeNameMap()
	c.Data["FailureCodeMap"] = types.FailureCodeMap()

	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "loan/list_scripts.html"
	return
}

func (c *LoanController) EditBankInfo() {
	c.Data["Action"] = "edit_bank_info"
	accountId, _ := c.GetInt64("account_id")
	bankInfo := service.GetAccountBankInfo(accountId)
	c.Data["bankInfo"] = bankInfo
	c.Data["bankList"] = xendit.AllBankList()
	c.Layout = "layout.html"
	c.TplName = "loan/edit_bank_info.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "loan/edit_bank_info_script.html"
}

func (c *LoanController) DoEditBankInfo() {
	c.Data["Action"] = "do_edit_bank_info"
	accountId, _ := c.GetInt64("account_id")
	bankName := c.GetString("bank_name")
	bankNo := c.GetString("bank_no")
	_, err := service.UpdateBankInfo(c.AdminUid, accountId, bankName, bankNo)
	url := "/loan/list"
	if err != nil {
		logs.Error("[DoEditBankInfo] account_id:%d err:%v", accountId, err)
		c.commonError("", url, i18n.T(c.LangUse, "半小时内只能修改一次,修改失败"))
		return
	}
	// url := fmt.Sprintf("%s%d", "/loan/backend/edit_bank_info?account_id=", accountId)
	c.Redirect(url, 302)
}

func (c *LoanController) RepayPlan() {
	c.Data["Action"] = "repay_plan"
	orderId, _ := c.GetInt64("order_id")
	repayPlan := service.GetBackendRepayPlan(orderId)
	c.Data["repayPlan"] = repayPlan
	c.Layout = "layout.html"
	c.TplName = "order/repay_plan.html"
}

func (c *LoanController) DisbureAgain() {
	c.Data["Action"] = "disburse_again"
	orderId, _ := c.GetInt64("order_id")
	data := service.DisbureseAgainDetailBackend(orderId)
	currentCompany, supportCompany := service.LoanCompany(orderId, data.BankName)

	fmt.Printf("%#v\n", data)
	c.Data["data"] = data
	c.Data["order_id"] = orderId
	c.Data["CurrentCompany"] = currentCompany
	c.Data["SupportCompany"] = supportCompany
	c.Layout = "layout.html"
	c.TplName = "loan/disburse_again.html"
}

func (c *LoanController) DoDisbureAgain() {
	c.Data["Action"] = "do_disburse_again"
	orderId, _ := c.GetInt64("order_id")
	checkStatus, _ := c.GetInt64("check_status")
	loanCompanyCode, _ := c.GetInt("loan_comany_code")
	url := "/loan/list"
	action := "/loan/backend/disburse_again"
	if checkStatus != int64(types.LoanStatusWait4Loan) && checkStatus != int64(types.LoanStatusInvalid) {
		c.commonError(action, url, "订单状态有误")
		return
	}

	// 放开重新放款权限校验 2018年09月27日14:45:52
	//if !service.CanDisbureAgain(c.AdminUid, orderId, int(checkStatus)) {
	//	logs.Error("[DoDisbureAgain] admin:%d don't have authority to re loan this order:%d", c.AdminUid, orderId)
	//	c.commonError(action, url, "you don't have authority to re loan this order.")
	//	return
	//}

	err := service.DoDisbureseAgainBackend(c.AdminUid, orderId, types.LoanStatus(checkStatus))
	if err != nil {
		logs.Error("[ReDisburse] failed, err is", err)
		c.commonError(action, url, "ReDisburse failed")
		return
	}

	logs.Warn("loanCompanyCode:%d", loanCompanyCode)
	orderExt, _ := models.GetOrderExt(orderId)
	if orderExt.SpecialLoanCompany != loanCompanyCode &&
		loanCompanyCode > 0 {
		org := orderExt
		orderExt.SpecialLoanCompany = loanCompanyCode
		orderExt.Utime = tools.GetUnixMillis()
		if orderExt.OrderId == 0 {
			orderExt.OrderId = orderId
			orderExt.Ctime = orderExt.Utime
			models.OrmInsert(&orderExt)
		} else {
			cols := []string{"special_loan_company", "utime"}
			models.OrmUpdate(&orderExt, cols)
			models.OpLogWrite(c.AdminUid, orderExt.OrderId, models.OpCodeOrderUpdate, orderExt.TableName(), org, orderExt)
		}
	}

	c.Data["OpMessage"] = i18n.T(c.LangUse, "申请成功")
	c.Layout = "layout.html"
	c.TplName = "success_redirect.html"
}

func (c *LoanController) DoDisbureAgainMulti() {
	ids := c.GetStrings("ids[]")

	logs.Info(strings.Join(ids, ","))
	for _, id := range ids {

		logs.Info(id)
		idInt, _ := tools.Str2Int64(id)

		// 放开重新放款权限校验 2018年09月27日14:45:52
		//if !service.CanDisbureAgain(c.AdminUid, idInt, int(types.LoanStatusWait4Loan)) {
		//	logs.Error("[DoDisbureAgain] admin:%d don't have authority to re loan this order:%d", c.AdminUid, idInt)
		//	continue
		//}

		err := service.DoDisbureseAgainBackend(c.AdminUid, idInt, types.LoanStatusWait4Loan)
		if err != nil {
			logs.Error("[DoDisbureAgainMulti] DoDisbureseAgainBackend id:%d AdminUid:%d err:%v", idInt, c.AdminUid, err)
		}
	}
	c.ServeJSON()
	return
}

func (c *LoanController) DoRollBack() {
	strId := c.GetString("order_id")
	logs.Info(strId)
	orderId, _ := tools.Str2Int64(strId)

	if orderId > 0 {
		err := service.RollBackOrder(c.AdminUid, orderId)
		if err != nil {
			logs.Error("[DoRollBack] service.RollBackOrder id:%d AdminUid:%d err:%v", orderId, c.AdminUid, err)
		}
	} else {
		logs.Error("[DoRollBack] orderId==0 strId:%s AdminUid:%d ", strId, c.AdminUid)
	}

	c.ServeJSON()
	return
}
