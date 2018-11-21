package controllers

import (
	//"micro-loan/common/models"

	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/pkg/repayremind"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"
	"strings"

	//"github.com/astaxie/beego/orm"
	"fmt"

	"micro-loan/common/pkg/reduce"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils/pagination"
)

type RepayController struct {
	BaseController
}

func (c *RepayController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "repay"
}

// 还款管理列表
func (c *RepayController) List() {
	c.Data["Action"] = "list"
	c.Data["OpUid"] = c.AdminUid
	c.TplName = "repay/list.html"

	var condCntr = map[string]interface{}{}

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
	checkStatuss := c.GetStrings("check_status")
	if len(checkStatuss) > 0 {
		condCntr["check_status"] = checkStatuss
	}

	selectedTagMap := map[interface{}]interface{}{}
	for _, checkStatus := range checkStatuss {
		validTag, _ := strconv.Atoi(checkStatus)
		selectedTagMap[types.LoanStatus(validTag)] = nil
	}
	c.Data["checkStatus"] = selectedTagMap

	// checkStatus, _ := c.GetInt("check_status")
	//loanStatus := types.LoanStatus(1)
	// if checkStatus > 0 {
	// 	condCntr["check_status"] = checkStatus
	// }

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
	repayDateRange := c.GetString("repay_date_range")
	if len(repayDateRange) > 16 {
		tr := strings.Split(repayDateRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["repay_start_date"] = timeStart
				condCntr["repay_end_date"] = timeEnd
			}
		}
	}

	c.Data["repayDateRange"] = repayDateRange

	// 实际还款时间范围
	repayTimeRange := c.GetString("repay_time_range")
	if len(repayTimeRange) > 16 {
		tr := strings.Split(repayTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["repay_time_start"] = timeStart
				condCntr["repay_time_end"] = timeEnd
			}
		}
	}

	finishTimeRange := c.GetString("finish_time_range")
	c.Data["finishTimeRange"] = finishTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(finishTimeRange); err == nil {
		condCntr["finish_time_start"], condCntr["finish_time_end"] = start, end
	}

	leftAmount, _ := c.GetInt64("left_amount")
	if leftAmount > 0 {
		condCntr["left_amount"] = leftAmount
	}
	c.Data["leftAmount"] = leftAmount

	sortfield := c.GetString("field")
	if len(sortfield) > 0 {
		condCntr["field"] = sortfield
	}

	sorttype := c.GetString("sort")
	if len(sorttype) > 0 {
		condCntr["sort"] = sorttype
	}

	c.Data["repayTimeRange"] = repayTimeRange
	mobile, _ := c.GetInt64("mobile")
	if mobile > 0 {

		condCntr["mobile"] = mobile
	}
	c.Data["id"] = id
	c.Data["mobile"] = mobile
	c.Data["account_id"] = accountIDStr
	c.Data["realname"] = realname
	c.Data["check_status"] = types.RepayStatusMap()
	//c.Data["checkStatus"] = loanStatus
	// c.Data["apply_start_time"] = applyStartTimeStr
	c.Data["LoanStatusMap"] = types.RepayStatusMap()

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	list, count, totalRepay, totalRepayPayed, totalRepayReduce := service.RepayListBackend(condCntr, page, pagesize)
	for idx, _ := range list {
		list[idx].TotalRepayPayed = repayplan.CaculateTotalPayed(list[idx].AmountPayed, list[idx].GracePeriodInterestPayed, list[idx].PenaltyPayed)
		list[idx].ReduceTotal = list[idx].AmountReduced + list[idx].GracePeriodInterestReduced + list[idx].PenaltyReduced
		accounts := service.GetEaccountsDesc(list[idx].UserAccountId)
		list[idx].UserEAccounts = append(list[idx].UserEAccounts, accounts)

		order, _ := models.GetOrder(list[idx].Id)
		repayPlan, _ := models.GetLastRepayPlanByOrderid(list[idx].Id)
		repayBalanceAmount, _ := reduce.RepayLowestMoney4ClearReduce(order, repayPlan)
		list[idx].RepayBalanceAmount = repayBalanceAmount
		list[idx].TotalRepay = repayBalanceAmount
	}

	paginator := pagination.SetPaginator(c.Ctx, pagesize, int64(count))

	if len(condCntr) == 0 {
		totalRepay = 0
		totalRepayPayed = 0
		totalRepayReduce = 0
	}
	c.Data["paginator"] = paginator
	c.Data["List"] = list
	c.Data["totalRepay"] = totalRepay
	c.Data["totalRepayPayed"] = totalRepayPayed
	c.Data["totalRepayReduce"] = totalRepayReduce

	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "repay/list_scripts.html"
	return
}

// 还款管理列表
func (c *RepayController) VaSearch() {
	c.Data["Action"] = "list"
	c.TplName = "repay/va_search.html"

	var condCntr = map[string]interface{}{}

	// order id
	idS := c.GetString("id")
	id, _ := tools.Str2Int64(idS)
	if id > 0 {
		condCntr["id"] = id
	}

	accountIDStr := c.GetString("account_id", "")
	accountID, _ := tools.Str2Int64(accountIDStr)
	if accountID > 0 {
		condCntr["account_id"] = accountID
	}

	vaCode := c.GetString("va_code")
	if len(vaCode) > 0 {
		condCntr["va_code"] = vaCode
	}
	c.Data["vaCode"] = vaCode

	paymentCode := c.GetString("payment_code")
	if len(paymentCode) > 0 {
		condCntr["payment_code"] = paymentCode
	}
	c.Data["paymentCode"] = paymentCode

	repayType, _ := c.GetInt("repay_type")
	if repayType > 0 {
		condCntr["repay_type"] = repayType
	}
	c.Data["RepayType"] = repayType

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

	mobile := c.GetString("mobile")
	mobile = strings.Trim(mobile, " ")
	if len(mobile) > 0 {
		condCntr["mobile"] = mobile
	}

	c.Data["id"] = idS
	c.Data["mobile"] = mobile
	c.Data["account_id"] = accountIDStr
	c.Data["RepayTypeMap"] = types.RepayTypeMap()

	list := []service.RepayVaDisplay{}
	if len(condCntr) > 0 {
		list = service.RepayVaSearch(condCntr)
	}

	if len(list) > 0 {
		ab, _ := models.OneAccountBaseByPkId(list[0].UserAccountId)
		for k, _ := range list {
			list[k].RealName = ab.Realname
		}
	}

	c.Data["List"] = list
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "repay/list_scripts.html"
	return
}

func (c *RepayController) UserTrans() {
	c.Data["Action"] = "user_trans"
	accountId, _ := c.GetInt64("order_id")
	userTrans := service.GetBackendUserETrans(accountId)
	fmt.Printf("list is %#v", userTrans)
	c.Data["userTrans"] = userTrans
	c.Layout = "layout.html"
	c.TplName = "repay/user_e_trans.html"
}

func (c *RepayController) RepayPlan() {
	c.Data["Action"] = "repay_plan"
	orderId, _ := c.GetInt64("order_id")
	repayTime, _ := c.GetInt64("repay_time")
	repayPlan := service.GetBackendRepayPlan(orderId)
	c.Data["repayPlan"] = repayPlan
	c.Data["replayTime"] = repayTime
	c.Layout = "layout.html"
	c.TplName = "repay/repay_plan.html"
}
func (c *RepayController) RepayPlanRollBack() {
	orderId, _ := c.GetInt64("order_id")
	repayPlan := service.GetBackendRepayPlan(orderId)

	c.Data["repayPlan"] = repayPlan

	//0 是否是admin
	if c.AdminUid != 1 {
		logs.Error("[RepayPlanRollBack] only admin can do this. order_id:%d", orderId)
		return
	}

	//c.Data["OrderId"] = orderId
	c.Data["OrderId"] = orderId
	c.Data["AmountPayedTotal"] = service.GetRollBackDetail(orderId)

	c.TplName = "repay/repay_plan_rollback_apply.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "repay/list_scripts.html"
}

func (c *RepayController) DoRepayPlanRollBack() {
	orderId, _ := c.GetInt64("order_id")
	rollBackTotal, _ := c.GetInt64("roll_back_total")

	//0 是否是admin
	if c.AdminUid != 1 {
		logs.Error("[DoRepayPlanRollBack] only admin can do this. order_id:%d", orderId)
		return
	}
	err := service.DoRollBackRepayPlan(c.AdminUid, orderId, rollBackTotal)
	if err != nil {
		c.Data["json"] = 1
	} else {
		c.Data["json"] = 0
	}
	c.ServeJSON()
	return
}

func (c *RepayController) RepayPlanHistory() {
	c.Data["Action"] = "repay_plan_history"

	orderId, _ := c.GetInt64("order_id")
	repayPlanHistory := service.GetBackendRepayPlanHistory(orderId)

	c.Data["repayPlanHistory"] = repayPlanHistory
	c.Layout = "layout.html"
	c.TplName = "repay/repay_plan_history.html"
}

func (c *RepayController) RemindCaseList() {
	c.TplName = "repay/remind_case_list.html"

	// query {{{
	var condCntr = map[string]interface{}{}

	realname := c.GetString("realname")
	if len(realname) > 0 {
		condCntr["realname"] = realname
	}
	c.Data["realname"] = realname

	mobile := c.GetString("mobile")
	if len(mobile) > 0 {
		condCntr["mobile"] = mobile
	}
	c.Data["mobile"] = mobile

	level := c.GetString("level")
	if len(level) > 0 {
		condCntr["level"] = level
	}
	c.Data["level"] = level

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
		c.Data["id"] = id
	}
	orderID, _ := c.GetInt64("order_id")
	if orderID > 0 {
		condCntr["order_id"] = orderID
	}
	c.Data["orderID"] = orderID

	accountID, _ := c.GetInt64("account_id")
	if accountID > 0 {
		condCntr["account_id"] = accountID
	}
	c.Data["accountID"] = accountID

	// 提醒生成时间
	ctimeRange := c.GetString("ctime_range")
	c.Data["ctimeRange"] = ctimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(ctimeRange); err == nil {
		condCntr["ctime_start"], condCntr["ctime_end"] = start, end
	}
	// end }}}

	sortfield := c.GetString("field")
	if len(sortfield) > 0 {
		condCntr["field"] = sortfield
	}

	sorttype := c.GetString("sort")
	if len(sorttype) > 0 {
		condCntr["sort"] = sorttype
	}

	page, _ := c.GetInt("p")
	pageSize := service.Pagesize

	list, count, _ := repayremind.ListBackend(condCntr, page, pageSize)
	for i := range list {
		list[i].TotalRepay = repayplan.CaculateRepayTotalAmount(
			list[i].Amount, list[i].AmountPayed, list[i].AmountReduced,
			list[i].GracePeriodInterest, list[i].GracePeriodInterestPayed, list[i].GracePeriodInterestReduced,
			list[i].Penalty, list[i].PenaltyPayed, list[i].PenaltyReduced)
	}
	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.Data["RMCaseCreateDaysMap"] = types.RMCaseCreateDaysMap()

	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "repay/remind_case_scripts.html"
	return
}

func (c *RepayController) RemindCaseHandle() {
	id, _ := c.GetInt64("id", 0)
	oneCase, _ := models.OneRepayRemindCaseByPkID(id)

	c.isGrantedData(types.DataPrivilegeTypeRepayRemindCase, id)

	c.Data["OneCase"] = oneCase

	orderData, _ := models.GetOrder(oneCase.OrderId)
	c.Data["OrderData"] = orderData

	hasBigData, contactList := service.GetContactList(service.ContactListRepay, orderData.UserAccountId, oneCase.OrderId)
	c.Data["HasBigData"] = hasBigData
	c.Data["ContactList"] = contactList

	accountProfile, _ := models.OneAccountProfileByAccountID(orderData.UserAccountId)
	c.Data["AccountProfile"] = accountProfile

	repayPlan, _ := models.GetLastRepayPlanByOrderid(oneCase.OrderId)
	c.Data["leftRepayTotalAmount"], _ = repayplan.CaculateRepayTotalAmountByRepayPlan(repayPlan)
	c.Data["RepayPlan"] = repayPlan

	customer, _ := dao.CustomerOne(orderData.UserAccountId)

	records := service.GetAllAutoCallResult(customer.Mobile)
	c.Data["autoCallRecord"] = records
	result := service.GetLatestAutoCallResult(customer.Mobile)
	c.Data["AutoCallResult"] = result

	customer.Mobile = tools.MobileFormat(customer.Mobile)
	c.Data["Customer"] = customer
	c.Data["RepayInclinationMap"] = types.RepayInclinationMap()
	c.Data["UnconnectReasonMap"] = types.UnconnectReasonMap()
	c.Data["NotRepayReasonMap"] = types.NotRepayReasonMap()

	expireFlag := service.MarketPaymentCodeGenerateButton(oneCase.OrderId)
	c.Data["ExpireFlag"] = expireFlag
	marketPayment, _ := models.OneFixPaymentCodeByUserAccountId(orderData.UserAccountId)
	//logs.Debug(marketPayment)
	c.Data["ExpireDateTime"] = tools.MDateMHS(marketPayment.ExpirationDate)
	c.Data["PaymentCode"] = marketPayment.PaymentCode

	list, _ := service.GetRepayRemindCaseLogListByOrderId(oneCase.OrderId)
	c.Data["list"] = list

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "repay/remind_case_handle_scripts.html"
	c.LayoutSections["JsPlugin"] = "plugin/clipboard.html"
	c.LayoutSections["CssPlugin"] = "repay/repay.css.html"
	c.Layout = "layout.html"
	c.TplName = "repay/remind_case.html"
}

func (c *RepayController) RemindCaseUpdate() {

	id, _ := c.GetInt64("id")
	result := c.GetString("result")
	promiseRepayTime := tools.GetTimeParse(c.GetString("promise_repay_time")) * 1000

	c.isGrantedData(types.DataPrivilegeTypeOverdueCase, id)

	oneCase, _ := models.OneRepayRemindCaseByPkID(id)

	origin := oneCase

	oneCase.OpUid = c.AdminUid
	oneCase.Result = result
	oneCase.PromiseRepayTime = promiseRepayTime
	oneCase.Utime = tools.GetUnixMillis()

	// 更新案件

	models.OrmAllUpdate(&oneCase)

	phoneConnect, _ := c.GetInt("phone_connected")
	isWillRepay, _ := c.GetInt("is_will_repay")
	unconnectReason, _ := c.GetInt("phone_unconnect_reason")
	phoneObject, _ := c.GetInt("phone_objects")
	phoneObjectMobile := c.GetString("remind-call-input")
	urgetype, _ := c.GetInt("urge_type")
	phoneTime := c.GetString("phone_time")
	phoneTimeInt64, _ := tools.GetTimeParseWithFormat(phoneTime, "2006-01-02 15:04:05")

	caseLog := models.RepayRemindCaseLog{}
	caseLog.CaseId = id
	caseLog.UrgeType = urgetype
	caseLog.OrderId = oneCase.OrderId
	caseLog.UnrepayReason = c.GetString("unrepay_reason")
	caseLog.PhoneConnect = phoneConnect
	caseLog.PromiseRepayTime = promiseRepayTime
	caseLog.Result = tools.TrimRealName(result)
	caseLog.IsWillRepay = isWillRepay
	caseLog.UnconnectReason = unconnectReason
	caseLog.PhoneObject = phoneObject
	caseLog.PhoneObjectMobile = phoneObjectMobile
	caseLog.PhoneTime = phoneTimeInt64 * 1000
	caseLog.OpUid = c.AdminUid

	caseLog.Ctime = tools.GetUnixMillis()
	caseLog.Utime = tools.GetUnixMillis()
	logID, _ := models.OrmInsert(&caseLog)

	if logID > 0 {
		isEmptyNumber := 0
		if caseLog.UnconnectReason == types.UnconnectReasonEmptyNumber {
			isEmptyNumber = 1
		}
		ticket.UpdateByHandleCase(id, types.MustGetTicketItemIDByCaseName(oneCase.Level),
			caseLog.Ctime, caseLog.PromiseRepayTime, phoneObject, 0, isEmptyNumber, "")
	}

	// 写操作日志
	models.OpLogWrite(c.AdminUid, oneCase.Id, models.OpRepayRemindCaseUpdate, oneCase.TableName(), origin, oneCase)

	c.Data["OpMessage"] = "Urge result save success."
	//c.Data["Redirect"] = "/overdue/list"
	c.Layout = "layout.html"
	c.TplName = "success_redirect.html"
}

func (c *RepayController) RemindCaseView() {

}

func (c *RepayController) RemindCaseLog() {
	orderID, _ := c.GetInt64("order_id")

	c.isGrantedData(types.DataPrivilegeTypeOrder, orderID)

	list, _ := service.GetRepayRemindCaseLogListByOrderId(orderID)

	c.Data["list"] = list
	c.Layout = "layout.html"
	c.TplName = "repay/remind_case_log.html"

}
