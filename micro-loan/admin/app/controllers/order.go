package controllers

import (
	"strings"

	"github.com/astaxie/beego/utils/pagination"

	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type OrderController struct {
	BaseController
}

func (c *OrderController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "order"
}

// 借款管理列表
func (c *OrderController) List() {
	c.Data["Action"] = "list"
	c.TplName = "order/list.html"
	var condCntr = map[string]interface{}{}

	id, idErr := c.GetInt64("id")
	if idErr == nil && id > 0 {
		condCntr["id"] = id
	}

	realname := c.GetString("realname")
	if len(realname) > 0 {
		condCntr["realname"] = realname
	}

	checkStatusMulti := c.GetStrings("check_status")
	if len(checkStatusMulti) > 0 {
		condCntr["check_status"] = checkStatusMulti
	}

	var hasRoll bool
	orderType := c.GetStrings("order_type")
	if len(orderType) > 0 {
		condCntr["order_type"] = orderType
		for _, typeKey := range orderType {
			if typeKey == "roll" {
				hasRoll = true
				break
			}
		}
	}
	c.Data["hasRoll"] = hasRoll

	// user account id
	userAccountId, _ := c.GetInt64("user_account_id")
	if userAccountId > 0 {
		condCntr["user_account_id"] = userAccountId
	}
	mobile := c.GetString("mobile")
	if len(mobile) > 0 {

		condCntr["mobile"] = mobile
	}
	splitSep := " - "
	// 申请时间范围
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
	// 创建时间范围
	ctime := c.GetString("ctime_range")
	if len(ctime) > 16 {
		tr := strings.Split(ctime, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["ctime_start_time"] = timeStart
				condCntr["ctime_end_time"] = timeEnd
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
	c.Data["ctime_range"] = ctime
	c.Data["order_type"] = types.OrderTypeMap()
	c.Data["orderTypeMultiBox"] = service.BuildJsVar("orderTypeMultiBox", orderType)
	c.Data["applyTimeRange"] = applyTimeRange
	c.Data["id"] = id
	c.Data["mobile"] = mobile
	c.Data["realname"] = realname
	c.Data["check_status"] = types.AllOrderStatusMap()
	c.Data["risk_ctl_status"] = types.RiskCtlMap()
	c.Data["statusSelectMultiBox"] = service.BuildJsVar("statusSelectMultiBox", checkStatusMulti)
	c.Data["LoanStatusMap"] = types.OrderStatusMap()
	c.Data["userAccountId"] = userAccountId

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := types.DefaultPagesize

	list, count := service.OrderListBackend(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, int64(count))

	c.Data["paginator"] = paginator
	c.Data["List"] = list
	c.Data["RiskTypeMap"] = types.RiskTypeMap()

	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "order/list_scripts.html"
	return
}

func (c *OrderController) BusinessHistory() {
	c.Data["Action"] = "business_history"
	orderId, _ := c.GetInt64("order_id")
	orderBusiness := service.GetLoanOrderBusiness(orderId)
	c.Data["orderBusiness"] = orderBusiness
	c.Layout = "layout.html"
	c.TplName = "order/order_business.html"
}

func (c *OrderController) RepayPlan() {
	c.Data["Action"] = "repay_plan"
	orderId, _ := c.GetInt64("order_id")
	repayPlan := service.GetBackendRepayPlan(orderId)
	c.Data["repayPlan"] = repayPlan
	c.Layout = "layout.html"
	c.TplName = "order/repay_plan.html"
}
