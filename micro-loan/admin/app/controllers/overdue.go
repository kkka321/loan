package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils/pagination"

	"micro-loan/common/dao"
	"micro-loan/common/i18n"
	"micro-loan/common/lib/device"
	"micro-loan/common/models"
	"micro-loan/common/pkg/admin"
	"micro-loan/common/pkg/entrust"
	"micro-loan/common/pkg/reduce"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/pkg/schema_task"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/xendit"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type OverdueController struct {
	BaseController
}

func (c *OverdueController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "overdue"
}

// 催收管理列表
func (c *OverdueController) List() {
	c.Data["Action"] = "list"
	c.TplName = "overdue/list.html"

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

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
		c.Data["id"] = id
	}

	orderIdStr := c.GetString("order_id", "")
	orderId, _ := tools.Str2Int64(orderIdStr)
	if orderId > 0 {
		condCntr["order_id"] = orderId
	}
	c.Data["order_id"] = orderIdStr

	accountIdStr := c.GetString("account_id", "")
	accountId, _ := tools.Str2Int64(accountIdStr)
	if accountId > 0 {
		condCntr["account_id"] = accountId
	}
	c.Data["account_id"] = accountIdStr

	filter, _ := tools.Str2Int(c.GetString("filter", "-1"))
	if filter >= 0 {
		condCntr["filter"] = filter
	}
	c.Data["filter"] = filter

	joinUrgeTimeRange := c.GetString("join_urge_time_range")
	c.Data["joinUrgeTimeRange"] = joinUrgeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(joinUrgeTimeRange); err == nil {
		condCntr["join_urge_time_start"], condCntr["join_urge_time_end"] = start, end
	}

	// 出催时间
	outUrgeTimeRange := c.GetString("out_urge_time_range")
	c.Data["outUrgeTimeRange"] = outUrgeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(outUrgeTimeRange); err == nil {
		condCntr["out_urge_time_start"], condCntr["out_urge_time_end"] = start, end
	}

	// 案件级别
	caseLevel, _ := tools.Str2Int(c.GetString("caselevel", "-1"))
	if caseLevel >= 0 {
		condCntr["case_level"] = types.GetOverdueLevelItemVal(caseLevel)
	}
	c.Data["caselevel"] = caseLevel
	c.Data["OverdueLevelMap"] = types.OverdueLevelItemMap()

	// 订单类型
	orderType, _ := tools.Str2Int(c.GetString("ordertype", "-1"))
	if orderType >= 0 {
		condCntr["order_type"] = types.GetUrgeOrderTypeVal(orderType)
	}
	c.Data["ordertype"] = orderType
	c.Data["UrgeOrderTypeMap"] = types.UrgeOrderTypeMap()

	// 逾期天数查询
	overdueDaysStart, err := c.GetInt64("overdue_days_start")
	if err == nil && overdueDaysStart >= 0 {
		condCntr["overdue_days_start"] = overdueDaysStart
		c.Data["overdueDaysStart"] = overdueDaysStart
	}

	overdueDaysEnd, err := c.GetInt64("overdue_days_end")
	if err == nil && overdueDaysEnd >= 0 {
		condCntr["overdue_days_end"] = overdueDaysEnd
		c.Data["overdueDaysEnd"] = overdueDaysEnd
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

	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize

	list, count, _ := service.OverdueListBackend(c.AdminUid, condCntr, page, pageSize)
	for idx, _ := range list {
		list[idx].TotalRepay = repayplan.CaculateRepayTotalAmount(
			list[idx].Amount, list[idx].AmountPayed, list[idx].AmountReduced,
			list[idx].GracePeriodInterest, list[idx].GracePeriodInterestPayed, list[idx].GracePeriodInterestReduced,
			list[idx].Penalty, list[idx].PenaltyPayed, list[idx].PenaltyReduced)
		list[idx].TotalRepayPayed = repayplan.CaculateTotalPayed(list[idx].AmountPayed, list[idx].GracePeriodInterestPayed, list[idx].PenaltyPayed)
	}
	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.Data["UrgeFilterMap"] = types.UrgeFilterMap()

	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "overdue/list_scripts.html"
	return
}

//CO2案件手动创建工单
func (c *OverdueController) CreateTicket() {
	idsString := c.GetString("id")
	idStrings := strings.Split(idsString, ",")

	jsonData := make(map[string]interface{})
	count := 0
	if len(idStrings) > 0 {
		for _, v := range idStrings {
			caseID, _ := tools.Str2Int64(v)
			logs.Debug("[CreateTicket] caseID:", caseID)
			oneCase, _ := models.OneOverueCaseByPkId(caseID)
			logs.Debug("[CreateTicket] oneCase:", oneCase)
			if oneCase.IsOut == types.IsUrgeOutNo {
				oneTicket, _ := models.GetTicketByItemAndRelatedID(types.MustGetTicketItemIDByCaseName(oneCase.CaseLevel), caseID)
				logs.Debug("[CreateTicket] oneTicket:", oneTicket)
				if oneTicket.Id == 0 {
					oneOrder, _ := models.GetOrder(oneCase.OrderId)
					ticket.CreateTicket(types.MustGetTicketItemIDByCaseName(oneCase.CaseLevel), oneCase.Id, c.AdminUid, oneCase.OrderId, oneOrder.UserAccountId, nil)
					count++
				}
			}
		}
	}
	jsonData["data"] = count
	c.Data["json"] = jsonData
	c.ServeJSON()
}

// Co2案件
func (c *OverdueController) Co2caseList() {
	c.Data["Action"] = "co2case_list"
	c.TplName = "overdue/co2case_list.html"

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

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
		c.Data["id"] = id
	}

	orderIdStr := c.GetString("order_id", "")
	orderId, _ := tools.Str2Int64(orderIdStr)
	if orderId > 0 {
		condCntr["order_id"] = orderId
	}
	c.Data["order_id"] = orderIdStr

	accountIdStr := c.GetString("account_id", "")
	accountId, _ := tools.Str2Int64(accountIdStr)
	if accountId > 0 {
		condCntr["account_id"] = accountId
	}
	c.Data["account_id"] = accountIdStr

	filter, _ := tools.Str2Int(c.GetString("filter", "-1"))
	if filter >= 0 {
		condCntr["filter"] = filter
	}
	c.Data["filter"] = filter

	joinUrgeTimeRange := c.GetString("join_urge_time_range")
	c.Data["joinUrgeTimeRange"] = joinUrgeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(joinUrgeTimeRange); err == nil {
		condCntr["join_urge_time_start"], condCntr["join_urge_time_end"] = start, end
	}

	// 出催时间
	outUrgeTimeRange := c.GetString("out_urge_time_range")
	c.Data["outUrgeTimeRange"] = outUrgeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(outUrgeTimeRange); err == nil {
		condCntr["out_urge_time_start"], condCntr["out_urge_time_end"] = start, end
	}

	// 案件级别
	// caseLevel, _ := tools.Str2Int(c.GetString("caselevel", "-1"))
	// if caseLevel >= 0 {
	// 	condCntr["case_level"] = types.GetOverdueLevelItemVal(caseLevel)
	// }
	// c.Data["caselevel"] = caseLevel
	// c.Data["OverdueLevelMap"] = types.OverdueLevelItemMap()

	// 订单类型
	orderType, _ := tools.Str2Int(c.GetString("ordertype", "-1"))
	if orderType >= 0 {
		condCntr["order_type"] = types.GetUrgeOrderTypeVal(orderType)
	}
	c.Data["ordertype"] = orderType
	c.Data["UrgeOrderTypeMap"] = types.UrgeOrderTypeMap()

	// 逾期天数查询
	overdueDaysStart, err := c.GetInt64("overdue_days_start")
	if err == nil && overdueDaysStart >= 0 {
		condCntr["overdue_days_start"] = overdueDaysStart
		c.Data["overdueDaysStart"] = overdueDaysStart
	}

	overdueDaysEnd, err := c.GetInt64("overdue_days_end")
	if err == nil && overdueDaysEnd >= 0 {
		condCntr["overdue_days_end"] = overdueDaysEnd
		c.Data["overdueDaysEnd"] = overdueDaysEnd
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

	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize

	list, count, _ := service.OverdueCO2caseListBackend(c.AdminUid, condCntr, page, pageSize)
	for idx, _ := range list {
		list[idx].TotalRepay = repayplan.CaculateRepayTotalAmount(
			list[idx].Amount, list[idx].AmountPayed, list[idx].AmountReduced,
			list[idx].GracePeriodInterest, list[idx].GracePeriodInterestPayed, list[idx].GracePeriodInterestReduced,
			list[idx].Penalty, list[idx].PenaltyPayed, list[idx].PenaltyReduced)
		list[idx].TotalRepayPayed = repayplan.CaculateTotalPayed(list[idx].AmountPayed, list[idx].GracePeriodInterestPayed, list[idx].PenaltyPayed)
	}
	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.Data["UrgeFilterMap"] = types.UrgeFilterMap()

	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "overdue/co2case_list_scripts.html"
	return
}

func (c *OverdueController) BatchAssignPage() {
	c.TplName = "overdue/batch_assign.html"
	c.Data["EntrustCompany"] = types.EntrustCompanyMap()
	// c.Data["EntrustEnumMap"] = types.EntrustEnumMap()
	c.Data["AgreeEnumMap"] = types.AgreeEnumMap()
	logs.Debug("[BatchAssignPage]", c.Data["EntrustCompany"])
}

func (c *OverdueController) BatchAssign() {
	idsString := c.GetString("ids")
	pname := c.GetString("pname")
	isAgree, _ := c.GetInt("is_agree")
	auditComment := c.GetString("audit_comment")
	remark := c.GetString("remark")
	idStrings := strings.Split(idsString, ",")
	// resultData:=make(map[string]interface{})

	logs.Debug("[BatchAssign] idStrings:%s,pname:%s,isAgree:%d,auditcommit:%s,remark:%s", idsString, pname, isAgree, auditComment, remark)
	valid := false
	if isAgree == 0 && auditComment != "" {
		valid = true
	}
	if isAgree == 1 && pname != "-1" && auditComment != "" {
		valid = true
	}
	if len(idStrings) == 0 || !valid {
		c.Data["json"] = map[string]interface{}{
			"error": "Invaild Request",
		}
		c.ServeJSON()
		return
	}
	var ids []int64
	for _, v := range idStrings {
		iv, err := strconv.ParseInt(v, 10, 64)
		if err == nil && iv > 0 {
			ids = append(ids, iv)
		}
	}
	logs.Debug("[BatchAssign] ids:", ids)
	result := entrust.ManualBatchAssign(ids, pname, auditComment, remark, isAgree)
	logs.Debug("[BatchAssign] result:", result)
	c.Data["json"] = map[string]interface{}{"result": fmt.Sprintf("Total order num: %d, actual assign num: %d", len(idStrings), result)}
	c.ServeJSON()
	return
}

// EntrustApprovalList 委外审批列表
func (c *OverdueController) EntrustApprovalList() {
	c.Data["Action"] = "entrust_approval_list"
	c.TplName = "overdue/entrust_aproval_list.html"

	var condCntr = map[string]interface{}{}

	orderIdStr := c.GetString("order_id", "")
	orderId, _ := tools.Str2Int64(orderIdStr)
	if orderId > 0 {
		condCntr["order_id"] = orderId
	}
	c.Data["order_id"] = orderIdStr

	accountIdStr := c.GetString("account_id", "")
	accountId, _ := tools.Str2Int64(accountIdStr)
	if accountId > 0 {
		condCntr["account_id"] = accountId
	}
	c.Data["account_id"] = accountIdStr

	filter, _ := tools.Str2Int(c.GetString("filter", "-1"))
	if filter >= 0 {
		condCntr["filter"] = filter
	}
	c.Data["filter"] = filter

	isAgree, _ := tools.Str2Int(c.GetString("is_agree", "-1"))
	if isAgree >= 0 {
		condCntr["isAgree"] = isAgree
	}
	c.Data["isAgree"] = isAgree

	isEntrust, _ := tools.Str2Int(c.GetString("is_entrust", "-1"))
	if isEntrust >= 0 {
		condCntr["isEntrust"] = isEntrust
	}
	c.Data["isEntrust"] = isEntrust

	pname := c.GetString("pname", "-1")
	if pname != "-1" {

		condCntr["pname"] = pname
	}
	c.Data["pname"] = pname

	entrustApplyRange := c.GetString("entrust_apply_range")
	// if entrustApplyRange == "" {
	// 	dateStr := tools.GetDate((tools.GetUnixMillis() - tools.MILLSSECONDADAY) / 1000)
	// 	entrustApplyRange = dateStr + " - " + dateStr
	// }
	c.Data["entrustApplyRange"] = entrustApplyRange
	if start, end, err := tools.PareseDateRangeToMillsecond(entrustApplyRange); err == nil {
		condCntr["entrust_apply_start"], condCntr["entrust_apply_end"] = start, end
	}

	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := 60 //service.Pagesize

	list, count, _ := service.EntrustApprovalListBackend(c.AdminUid, condCntr, page, pageSize)
	for idx, _ := range list {
		list[idx].TotalRepay = repayplan.CaculateRepayTotalAmount(
			list[idx].Amount, list[idx].AmountPayed, list[idx].AmountReduced,
			list[idx].GracePeriodInterest, list[idx].GracePeriodInterestPayed, list[idx].GracePeriodInterestReduced,
			list[idx].Penalty, list[idx].PenaltyPayed, list[idx].PenaltyReduced)
		list[idx].TotalRepayPayed = repayplan.CaculateTotalPayed(list[idx].AmountPayed, list[idx].GracePeriodInterestPayed, list[idx].PenaltyPayed)
	}
	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.Data["EntrustCompanyMap"] = types.EntrustCompanyMap()
	c.Data["EntrustEnumMap"] = types.EntrustEnumMap()
	c.Data["AgreeEnumMap"] = types.AgreeEnumMap()

	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "overdue/entrust_aproval_list_script.html"
	return

}

func (c *OverdueController) Urge() {
	action := "urge"
	c.Data["Action"] = action
	backRoute := "/overdue/list"
	id, _ := c.GetInt64("id", 0)
	//account_id, _ := c.GetInt64("account_id", 0)
	oneCase, _ := models.OneOverueCaseByPkId(id)

	if oneCase.OrderId > 0 {
		c.isGrantedData(types.DataPrivilegeTypeOverdueCase, id)

		c.Data["OneCase"] = oneCase

		isExist := admin.IsExistPrereduced(oneCase.Id, oneCase.OrderId)
		c.Data["isExistAddminPrereduced"] = isExist

		orderData, _ := models.GetOrder(oneCase.OrderId)
		c.Data["OrderData"] = orderData

		hasBigData, contactList := service.GetContactList(service.ContactListUrge, orderData.UserAccountId, oneCase.OrderId)

		c.Data["HasBigData"] = hasBigData
		c.Data["ContactList"] = contactList

		repayPlanModel, _ := models.GetLastRepayPlanByOrderid(oneCase.OrderId)
		c.Data["RepayPlan"] = repayPlanModel
		c.Data["acutalRepayedTotal"] = repayplan.CaculateAcutalRepayedTotalByRepayPlan(&repayPlanModel)
		amount, _ := repayplan.CaculateRepayTotalAmountByRepayPlan(repayPlanModel)
		c.Data["LeftOver"] = amount //repayPlan.Amount + repayPlan.GracePeriodInterest + repayPlan.Penalty - repayPlan.AmountPayed - repayPlan.GracePeriodInterestPayed - repayPlan.PenaltyPayed

		//应还罚息和宽限期利息、可减免金额、结算最低应还款额如果有申请，取申请数据计算，否则现算
		preReducedObj, _ := dao.GetLastPrereducedByOrderid(oneCase.OrderId)
		if preReducedObj.Id > 0 {
			canReducedAmount := preReducedObj.GracePeriodInterestPrededuced + preReducedObj.PenaltyPrereduced
			penaltyAndGracePeriodInterest := repayplan.CaculateGracPeriodAndPenaltyAmount(canReducedAmount, preReducedObj.DerateRatio)
			repayClearLowstAmount := amount - canReducedAmount
			c.Data["PenaltyAndGracePeriodInterest"] = penaltyAndGracePeriodInterest
			c.Data["CanReducedAmount"] = canReducedAmount
			c.Data["RepayClearLowstAmount"] = repayClearLowstAmount
		} else {
			//应还罚息和宽限期利息
			penaltyAndGracePeriodInterest, _ := repayplan.CaculateTotalGracePeriodAndPenaltyByRepayPlan(repayPlanModel)
			c.Data["PenaltyAndGracePeriodInterest"] = penaltyAndGracePeriodInterest
			//可减免金额
			caseLevel := oneCase.CaseLevel
			derateRatio, _ := config.ValidItemFloat64("derate_ratio_" + caseLevel)
			canReducedAmount := repayplan.CaculateCanReducedAmount(penaltyAndGracePeriodInterest, derateRatio)
			logs.Debug("[Urge] 可减免金额， caseLevel:", caseLevel, "derateRatio:", derateRatio, "canReducedAmount:", canReducedAmount)

			c.Data["CanReducedAmount"] = canReducedAmount
			//结清最低应还款额
			repayClearLowstAmount := amount - canReducedAmount
			c.Data["RepayClearLowstAmount"] = repayClearLowstAmount
		}

		customer, _ := dao.CustomerOne(orderData.UserAccountId)

		result := service.GetLatestAutoCallResult(customer.Mobile)
		c.Data["AutoCallResult"] = result
		c.Data["Mobile"] = customer.Mobile
		records := service.GetAllAutoCallResult(customer.Mobile)
		c.Data["autoCallRecord"] = records

		customer.Mobile = tools.MobileFormat(customer.Mobile)
		c.Data["Customer"] = customer
		accountProfile, _ := dao.CustomerProfile(orderData.UserAccountId)
		accountProfile.Contact1 = tools.MobileFormat(accountProfile.Contact1)
		accountProfile.Contact2 = tools.MobileFormat(accountProfile.Contact2)
		accountProfile.CompanyTelephone = tools.MobileFormat(accountProfile.CompanyTelephone)

		c.Data["AccountProfile"] = accountProfile

		eAccount, _ := dao.GetActiveEaccountWithBankName(orderData.UserAccountId)
		c.Data["EAccount"] = eAccount

		marketPayment, _ := models.OneFixPaymentCodeByUserAccountId(orderData.UserAccountId)
		//marketPayment, _ := models.GetMarketPaymentByOrderId(oneCase.OrderId)

		//expireFlag := service.MarketPaymentCodeGenerateButton(oneCase.OrderId)

		c.Data["PaymentCode"] = marketPayment.PaymentCode
		//c.Data["PaymentStatus"] = marketPayment.Status
		c.Data["ExpireDateTime"] = tools.MDateMHS(marketPayment.ExpirationDate)
		c.Data["ExpireFlag"] = false

		flag := false
		order, err := dao.AccountLastOverdueLoanOrder(orderData.UserAccountId)
		if err == nil {
			flag = service.IsOrderCanRoll(order)
		}
		c.Data["IsDefer"] = flag
		c.Data["AccountId"] = orderData.UserAccountId

		repayBalanceAmount, _ := reduce.RepayLowestMoney4ClearReduce(order, repayPlanModel)
		c.Data["RepayBalanceAmount"] = repayBalanceAmount

		// 获取工单信息
		ticketData, _ := models.GetTicketByItemAndRelatedID(types.OverdueLevelTicketItemMap()[oneCase.CaseLevel], oneCase.Id)
		if ticketData.CaseLevel == "" {
			ticketData.CaseLevel = "A"
		}
		orderExt, _ := models.GetOrderExt(oneCase.OrderId)
		c.Data["showEntrust"] = false
		//判断是否展示申请委外按钮
		if ticket.ApplyEntrustCondition(&ticketData, &oneCase, &orderExt) {
			c.Data["showEntrust"] = true
		}

		c.Data["ticketData"] = ticketData

		c.Data["RepayInclinationMap"] = types.RepayInclinationMap()
		//c.Data["UnconnectReasonMap"] = types.UnconnectReasonMap()
		c.Data["OverdueReasonItemMap"] = types.OverdueReasonItemMap()
		c.Data["UrgeResultMap"] = types.UrgeResultMap()
		c.Data["CommnicationWayMap"] = types.CommnicationWayMap()
		c.LayoutSections = make(map[string]string)
		c.LayoutSections["Scripts"] = "overdue/urge_scripts.html"
		c.LayoutSections["JsPlugin"] = "plugin/clipboard.html"
		c.LayoutSections["CssPlugin"] = "overdue/overdue.css.html"

		// c.isGrantedData(types.DataPrivilegeTypeOrder, orderId)
		urgeDetailList, _ := service.GetOverdueCaseDetailListByOrderIds(oneCase.OrderId)
		c.Data["urgeDetailList"] = urgeDetailList
		repayPlan := service.GetBackendRepayPlan(oneCase.OrderId)
		c.Data["repayPlan"] = repayPlan

		c.Layout = "layout.html"
		c.TplName = "overdue/urge.html"
	} else {
		c.Redirect(backRoute, 302)
	}

}

func (c *OverdueController) urgeCallFail(mapData map[string]interface{}) {
	c.Data["json"] = mapData
	c.ServeJSON()
}

// ApplyEntrust 申请委外
func (c *OverdueController) ApplyEntrust() {
	mapData := make(map[string]interface{})
	ticketID, _ := c.GetInt64("ticket_id", 0)
	applay, err := ticket.ApplyEntrust(ticketID)
	if applay && err == nil {
		//修改为待委外审批
		mapData["status"] = 1
		mapData["msg"] = "Apply successful"
	} else {
		mapData["status"] = 0
		mapData["msg"] = err
	}
	c.Data["json"] = &mapData
	c.ServeJSON()
}

// PreReduced 结清减免
func (c *OverdueController) PreReduced() {

	mapData := make(map[string]interface{})
	adminID := c.AdminUid
	caseID, _ := c.GetInt64("case_id", 0)
	orderID, _ := c.GetInt64("order_id", 0)

	oneCase, _ := models.OneOverueCaseByPkId(caseID)
	orderData, _ := models.GetOrder(orderID)

	if orderData.CheckStatus == types.LoanStatusInvalid ||
		orderData.CheckStatus == types.LoanStatusRolling ||
		orderData.CheckStatus == types.LoanStatusRollClear ||
		orderData.CheckStatus == types.LoanStatusRollApply ||
		orderData.CheckStatus == types.LoanStatusRollFail {
		mapData["msg"] = i18n.T(c.LangUse, "借款状态") + ":" +
			i18n.T(c.LangUse, types.AllOrderStatusMap()[orderData.CheckStatus]) +
			i18n.T(c.LangUse, ", 该案件不允许申请结清减免")
		c.Data["json"] = &mapData
		c.ServeJSON()
		return
	}

	isExist := admin.IsExistPrereduced(caseID, orderID)

	if isExist {
		mapData["msg"] = i18n.T(c.LangUse, "该案件已申请结清减免")
		c.Data["json"] = &mapData
		c.ServeJSON()
		return
	}
	quotaConf := admin.GetReducedQuotaConf(c.AdminUid)
	quotaToday := admin.GetReducedQuotaToday(c.AdminUid)

	logs.Debug("[PreReduced] AdminID:", adminID, "quotaToday:", quotaToday, "quotaConf:", quotaConf)
	//如果今天额度用尽返回提示消息
	if quotaToday >= int64(quotaConf) {
		///今日可减免的客户数已达上限,如有需要,请与主管申请
		mapData["msg"] = i18n.T(c.LangUse, "今日可减免的客户数已达上限,如有需要,请与主管申请")
	} else {

		//保存预减免宽限期利息和罚息
		repayPlan, _ := models.GetLastRepayPlanByOrderid(orderID)
		caseLevel := oneCase.CaseLevel //service.CalculateOverdueLevel(repayPlan.RepayDate)
		derateRatio, _ := config.ValidItemFloat64("derate_ratio_" + caseLevel)

		logs.Debug("[PreReduced]结清减免 caseLevel:", caseLevel, "derateRatio:", derateRatio)
		//计算减免宽限期利息
		PrereducedGracePeriod := repayplan.CaculateTotalGracePeriod(repayPlan.GracePeriodInterest, repayPlan.GracePeriodInterestPayed, repayPlan.GracePeriodInterestReduced)
		canReducedGracePeriodAmount := repayplan.CaculateCanReducedAmount(PrereducedGracePeriod, derateRatio)
		//计算减免罚息
		PrereducedPenalty := repayplan.CaculateTotalPenalty(repayPlan.Penalty, repayPlan.PenaltyPayed, repayPlan.PenaltyReduced)
		canReducedPenalty := repayplan.CaculateCanReducedAmount(PrereducedPenalty, derateRatio)

		idP, _ := device.GenerateBizId(types.ReduceRecordBiz)
		//结清减免记录插入
		tag := tools.GetUnixMillis()
		adminReducedQuota := models.ReduceRecordNew{
			Id:                            idP,
			ApplyUid:                      c.AdminUid,
			ConfirmUid:                    c.AdminUid,
			UserAccountId:                 orderData.UserAccountId,
			CaseID:                        caseID,
			OrderId:                       orderID,
			DerateRatio:                   derateRatio,
			ReduceType:                    types.ReduceTypePrereduced,
			ApplyTime:                     tag,
			GraceInterestReduced:          canReducedGracePeriodAmount,
			PenaltyReduced:                canReducedPenalty,
			GracePeriodInterestPrededuced: canReducedGracePeriodAmount,
			PenaltyPrereduced:             canReducedPenalty,
			ReduceStatus:                  types.ReduceStatusNotValid,
			Ctime:                         tag,
			Utime:                         tag,
		}
		id, _ := models.OrmInsert(&adminReducedQuota)

		if id > 0 {
			amount, _ := repayplan.CaculateRepayTotalAmountByRepayPlan(repayPlan)
			mapData["repayClearLowstAmount"] = amount - canReducedPenalty

			mapData["msg"] = i18n.T(c.LangUse, "结清减免申请成功")
			mapData["status"] = 1
		} else {
			mapData["msg"] = i18n.T(c.LangUse, "Try again")
			mapData["status"] = 0
		}

	}

	c.Data["json"] = &mapData
	c.ServeJSON()

}

func (c *OverdueController) UrgeSave() {
	c.Data["Action"] = "urge/save"

	result := c.GetString("result")

	idStr := c.GetString("id", "0")
	id, _ := tools.Str2Int64(idStr)

	c.isGrantedData(types.DataPrivilegeTypeOverdueCase, id)

	oneCase, _ := models.OneOverueCaseByPkId(id)

	origin := oneCase

	oneCase.UrgeUid = c.AdminUid
	oneCase.UrgeTime = tools.GetUnixMillis()
	oneCase.Result = result
	oneCase.Utime = oneCase.UrgeTime

	// 更新案件
	models.UpdateOverdueCase(&oneCase)

	orderId, _ := c.GetInt64("order_id", 0)
	phoneConnect, _ := c.GetInt("phone_connected")
	promiseRepayTime := c.GetString("promise_repay_time")
	repayInclination, _ := c.GetInt("repay_inclination")
	unconnectReason, _ := c.GetInt("phone_unconnect_reason")
	phoneObject, _ := c.GetInt("phone_objects")
	phoneObjectMobile := c.GetString("urge-call-input")

	overdueReasonItem, _ := c.GetInt("overdue_reason_item")
	communicationWay, _ := c.GetInt("communication_way")
	caseLevel := c.GetString("case_level")

	phoneTime := c.GetString("phone_time")
	phoneTimeInt64, _ := tools.GetTimeParseWithFormat(phoneTime, "2006-01-02 15:04:05")

	overdueCaseDetail := models.OverdueCaseDetail{}
	overdueCaseDetail.OverdueCaseId = id
	overdueCaseDetail.OrderId = orderId
	overdueCaseDetail.OverdueReasonItem = types.OverdueReasonItemEnum(overdueReasonItem)
	overdueCaseDetail.PhoneConnect = phoneConnect
	overdueCaseDetail.PromiseRepayTime = tools.GetTimeParse(promiseRepayTime) * 1000
	overdueCaseDetail.Result = result
	overdueCaseDetail.RepayInclination = repayInclination
	overdueCaseDetail.UnconnectReason = unconnectReason
	overdueCaseDetail.PhoneObject = phoneObject
	overdueCaseDetail.PhoneObjectMobile = phoneObjectMobile
	overdueCaseDetail.PhoneTime = phoneTimeInt64 * 1000
	overdueCaseDetail.OpUid = c.AdminUid
	timeTag := tools.GetUnixMillis()
	overdueCaseDetail.Ctime = timeTag
	overdueCaseDetail.Utime = timeTag
	detailID, _ := models.AddOverdueCaseDetail(&overdueCaseDetail)
	if detailID > 0 {
		isEmptyNumber := 0
		if overdueCaseDetail.UnconnectReason == types.UnconnectReasonEmptyNumber {
			isEmptyNumber = 1
		}
		ticket.UpdateByHandleCase(id, types.OverdueLevelTicketItemMap()[oneCase.CaseLevel],
			timeTag, overdueCaseDetail.PromiseRepayTime, phoneObject, communicationWay, isEmptyNumber, caseLevel)
	}

	// 写操作日志
	models.OpLogWrite(c.AdminUid, oneCase.Id, models.OpOverdueCaseUpdate, oneCase.TableName(), origin, oneCase)

	c.Data["OpMessage"] = "Urge result save success."
	c.Data["Redirect"] = fmt.Sprintf("/overdue/urge?id=%d", id)
	//c.Data["Redirect"] = "/overdue/list"
	c.Layout = "layout.html"
	c.TplName = "success_redirect.html"
}

func (c *OverdueController) UrgeDetail() {
	orderId, _ := c.GetInt64("order_id")

	c.isGrantedData(types.DataPrivilegeTypeOrder, orderId)

	list, _ := service.GetOverdueCaseDetailListByOrderIds(orderId)

	c.Data["list"] = list
	c.Layout = "layout.html"
	c.TplName = "overdue/urge_detail.html"
}

func (c *OverdueController) ReductionList() {
	caseId, _ := c.GetInt64("case_id")
	orderId, _ := c.GetInt64("order_id")
	accountId, _ := c.GetInt64("account_id")
	reduceType, _ := c.GetInt("reduce_type")
	reduceStatus, _ := c.GetInt("reduce_status")

	var condCntr = map[string]interface{}{}
	if orderId > 0 {
		condCntr["order_id"] = orderId
	}
	c.Data["orderId"] = orderId

	if caseId > 0 {
		condCntr["case_id"] = caseId
	}
	c.Data["caseId"] = caseId

	if accountId > 0 {
		condCntr["account_id"] = accountId
	}
	c.Data["accountId"] = accountId

	if reduceType > 0 {
		condCntr["reduce_type"] = reduceType
	}
	c.Data["ReduceType"] = reduceType

	if reduceStatus > 0 {
		condCntr["reduce_status"] = reduceStatus
	}
	c.Data["ReduceStatus"] = reduceStatus

	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize

	list, count, _ := service.ReductionListBackend(condCntr, page, pageSize)
	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))
	c.Data["paginator"] = paginator

	// 查询姓名和手机号
	for k, one := range list {
		account, _ := models.OneAccountBaseByPkId(one.UserAccountId)
		list[k].Name = account.Realname
		list[k].Mobile = account.Mobile
	}

	c.Data["ReduceTypeMap"] = types.ReduceTypeMap
	c.Data["ReduceStatusMap"] = types.ReduceStatusMap
	c.Data["list"] = list
	c.Layout = "layout.html"
	c.TplName = "overdue/reduction_list.html"
}

func (c *OverdueController) ReductionConfirm() {
	id, _ := c.GetInt64("id")

	logs.Info("[ReductionConfirm] id:%d", id)

	c.Data["Id"] = id
	c.Data["OptionMap"] = types.ReduceConfirmOptionMap
	c.Layout = "layout.html"
	c.TplName = "overdue/reduction_confirm.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "overdue/reduction_confirm_scripts.html"
}

func (c *OverdueController) ReductionConfirmSave() {
	id, _ := c.GetInt64("id")
	confirmOption, _ := c.GetInt("confirm_option")
	remark := c.GetString("remark")
	logs.Info("[ReductionConfirmSave] id:%d confirmOption:%d remark:%s", id, confirmOption, remark)

	if id <= 0 || remark == "" || confirmOption == 0 {
		action := "/overdue/backend/do_reduction"
		gotoURL := "/overdue/backend/reduction/list"

		logs.Error("[ReductionConfirmSave] param err. id:%d confirmOption:%d remark:%s", id, confirmOption, remark)
		c.commonError(action, gotoURL, "参数错误")
		return
	}

	err := service.ReductionConfirmSave(id, c.AdminUid, confirmOption, remark)
	if err != nil {
		action := "/overdue/backend/do_reduction"
		gotoURL := "/overdue/backend/reduction/list"

		logs.Error("[ReductionConfirmSave] param err:%v. id:%d confirmOption:%d remark:%s", err, id, confirmOption, remark)
		c.commonError(action, gotoURL, "减免失败")
		return
	}

	c.Data["OpMessage"] = "reduce confirm success."
	c.Layout = "layout.html"
	c.TplName = "success_redirect.html"
}

func (c *OverdueController) Reduction() {
	c.Data["Action"] = "reduction"
	orderId, _ := c.GetInt64("order_id")
	repayPlan := service.GetBackendRepayPlan(orderId)
	c.Data["repay_plan"] = repayPlan

	order, _ := models.GetOrder(orderId)
	c.Data["order"] = order

	reduceRecord, _ := models.GetLastestReduceRecordNew(orderId)
	c.Data["reduceRecord"] = reduceRecord

	couldReduceAmount := repayPlan.Amount - repayPlan.AmountPayed - repayPlan.AmountReduced
	couldReduceGraceInterest := repayPlan.GracePeriodInterest - repayPlan.GracePeriodInterestPayed - repayPlan.GracePeriodInterestReduced
	couldReducePenalty := repayPlan.Penalty - repayPlan.PenaltyPayed - repayPlan.PenaltyReduced
	c.Data["could_reduce_amount"] = couldReduceAmount
	c.Data["could_reduce_grace_interest"] = couldReduceGraceInterest
	c.Data["could_reduce_penalty"] = couldReducePenalty

	c.Data["order_id"] = orderId
	c.Layout = "layout.html"
	c.TplName = "overdue/reduction.html"
}

func (c *OverdueController) DoReduction() {
	c.Data["Action"] = "do_reduction"

	orderId, _ := c.GetInt64("order_id")
	reduction_amount, _ := c.GetInt64("reduction_amount")
	reduction_penalty, _ := c.GetInt64("reduction_penalty")
	reduction_interest, _ := c.GetInt64("reduction_interest")
	reduction_reason := c.GetString("reason")
	action := "/overdue/backend/do_reduction"
	gotoURL := fmt.Sprintf("/overdue/backend/reduction?order_id=%d", orderId)

	if reduction_penalty <= 0 && reduction_interest <= 0 && reduction_amount <= 0 {
		c.commonError(action, gotoURL, "减免数据异常, 输入的参数全为0 或存在非数值输入")
		return
	}

	err := service.ReducePenaltyApply(orderId, c.AdminUid, reduction_amount, reduction_penalty, reduction_interest, reduction_reason)
	if err != nil {
		c.commonError(action, gotoURL, err.Error())
		return
	}

	c.Data["OpMessage"] = "reduce apply success."
	//c.Data["Redirect"] = "/overdue/list"
	c.Layout = "layout.html"
	c.TplName = "success_redirect.html"
}

func (c *OverdueController) DeferShow() {

	// gotoURL := "overdue/defer_show"
	//
	// date := tools.MDateUTC(tools.GetUnixMillis())
	// endDate := date + " 23:59:59"
	// endTimeStamp, _ := tools.GetTimeParseWithFormat(endDate, "2006-01-02 15:04:05")
	//
	// //
	// var startTime time.Time
	// var endTime time.Time
	// currTime, _ := time.Parse("15:04:05", tools.MDateMHSHMS(tools.GetUnixMillis()))
	// timeQuantum := config.ValidItemString("overdue_roll_time_quantum")
	// if len(timeQuantum) > 0 {
	// 	times := strings.Split(timeQuantum, "-")
	// 	if len(times) >= 2 {
	// 		startTime, _ = time.Parse("15:04:05", times[0])
	// 		endTime, _ = time.Parse("15:04:05", times[1])
	// 	}
	// }
	//
	// if !currTime.After(startTime) || !endTime.After(currTime) {
	// 	c.newCommonError("/overdue/list", "不在展期申请时间范围内")
	// 	return
	// }
	//
	// accountId, err := tools.Str2Int64(c.GetString("account_id"))
	//
	// if err != nil {
	// 	c.newCommonError(gotoURL, "客户不存在")
	// 	return
	// }
	//
	// order, err := dao.AccountLastOverdueLoanOrder(accountId)
	// if err != nil {
	// 	logs.Error("[DeferShow] Customer has no temporary order. accountId:", accountId, ", err:", err)
	// 	return
	// }
	//
	// if order.CheckStatus != types.LoanStatusOverdue {
	// 	logs.Error("[DeferShow] Order can not roll or be rolling. accountId:", accountId, ", orderId:", order.Id, ", err:", err)
	// 	return
	// }
	//
	// period, minRepay, _, err := service.CalcRollRepayAmount(order)
	// if err != nil {
	// 	logs.Error("[DeferShow] Roll trial cal fail. accountId:", accountId, ", err:", err)
	// 	return
	// }
	//
	accountID, err := c.GetInt64("account_id")
	if err != nil {
		c.newCommonError("/", "Invalid account_id")
		return
	}

	// c.Data["min_repay"] = minRepay
	// c.Data["latest_repay_time"] = endTimeStamp * 1000
	// c.Data["extension_refund"] = order.Amount
	// c.Data["extension_repay_time"] = tools.NaturalDay(int64(period))
	c.Data["min_repay"], c.Data["latest_repay_time"], c.Data["extension_refund"], c.Data["extension_repay_time"], err = c.getRollTrialData(accountID)
	if err != nil {
		c.newCommonError("/", "Invalid Request")
		return
	}

	c.Layout = "layout.html"
	c.TplName = "overdue/defer_show.html"

}

func (c *OverdueController) GetRollTrialData() {
	res := make(map[string]interface{})

	accountID, err := c.GetInt64("account_id")
	if err != nil {
		res["error"] = "Invalid account_id"
	} else {
		res["minRepay"], res["lastestRepayTime"], res["rollRepayAmount"], res["rollRepayTime"], err = c.getRollTrialData(accountID)
		if err != nil {
			res["error"] = "Invalid Request"
		}
	}

	c.Data["json"] = res
	c.ServeJSON()
	return
}

func (c *OverdueController) getRollTrialData(accountID int64) (minRepay, lastestRepayTime, rollRepayAmount, rollRepayTime int64, err error) {
	date := tools.MDateUTC(tools.GetUnixMillis())
	endDate := date + " 23:59:59"
	endTimeStamp, _ := tools.GetTimeParseWithFormat(endDate, "2006-01-02 15:04:05")

	//
	var startTime, endTime time.Time
	currTime, _ := time.Parse("15:04:05", tools.MDateMHSHMS(tools.GetUnixMillis()))
	timeQuantum := config.ValidItemString("overdue_roll_time_quantum")
	if len(timeQuantum) > 0 {
		times := strings.Split(timeQuantum, "-")
		if len(times) >= 2 {
			startTime, _ = time.Parse("15:04:05", times[0])
			endTime, _ = time.Parse("15:04:05", times[1])
		}
	}

	if !currTime.After(startTime) || !endTime.After(currTime) {
		err = errors.New("Not in apply time range")
		return
	}

	if err != nil {
		return
	}

	order, err := dao.AccountLastOverdueLoanOrder(accountID)
	if err != nil {
		err = fmt.Errorf("[DeferShow] Customer has no temporary order. accountId: %d, err: %v", accountID, err)
		return
	}

	if order.CheckStatus != types.LoanStatusOverdue {
		err = fmt.Errorf("[DeferShow] Order can not roll or be rolling. accountId: %d, orderId: %d,err: %v", accountID, order.Id, err)
		return
	}
	var period int
	period, minRepay, _, err = service.CalcRollRepayAmount(order)
	if err != nil {
		err = fmt.Errorf("[DeferShow] Roll trial cal fail. accountId: %d, orderId: %d,err: %v", accountID, order.Id, err)
		return
	}

	lastestRepayTime = endTimeStamp * 1000
	rollRepayAmount = order.Amount
	rollRepayTime = tools.NaturalDay(int64(period))
	return

}

func (c *OverdueController) MarketPaymentCodeGenerate() {
	response := map[string]interface{}{}
	response["error"] = ""

	orderId, _ := c.GetInt64("order_id")
	balance, _ := c.GetInt64("balance")

	logs.Debug("chester_balance", balance)
	logs.Debug("orderId", orderId)

	err, marketPayment, amount := xendit.MarketPaymentCodeGenerate(orderId, balance)

	if err != nil {
		response["error"] = err.Error()
	} else {
		account, _ := dao.CustomerOne(marketPayment.UserAccountId)

		param := make(map[string]interface{})
		param["related_id"] = orderId
		schema_task.SendBusinessMsg(types.SmsTargetPaymentCode, types.ServiceMarketPaymentCode, account.Mobile, param)

		response["amount"] = amount
		response["paymentCode"] = marketPayment.PaymentCode
		response["expireDate"] = marketPayment.ExpirationDate
	}
	c.Data["json"] = response

	c.ServeJSON()

	return
}

func (c *OverdueController) RollApply() {
	accountID, err := c.GetInt64("account_id")
	if err != nil {
		c.newCommonError("/", "Invalid account_id")
		return
	}

	is_apply := service.IsTrialCalOrApply()
	if !is_apply {
		c.newCommonError("/", i18n.T(c.LangUse, "该时间段不允许展期"))
		return
	}

	err = service.CreateRollOrder(accountID)
	if err != nil {
		c.newCommonError("/", err.Error())
		return
	}

	c.Data["OpMessage"] = i18n.T(c.LangUse, "展期申请成功")
	c.Layout = "layout.html"
	c.TplName = "success.html"
}
