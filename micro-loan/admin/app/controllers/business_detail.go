package controllers

import (
	"strings"

	"github.com/astaxie/beego/logs"

	"github.com/astaxie/beego/utils/pagination"

	"micro-loan/common/dao"
	"micro-loan/common/service"
	"micro-loan/common/tools"
)

type BusinessDetailController struct {
	BaseController
}

func (c *BusinessDetailController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "businessDetail"
}

func (c *BusinessDetailController) List() {

	var condCntr = map[string]interface{}{}
	mobile := c.GetString("mobile")
	if len(mobile) > 0 {
		condCntr["mobile"] = mobile
	}

	splitSep := " - "
	// s申请时间范围
	registerTimeRange := c.GetString("register_time_range")
	if len(registerTimeRange) > 16 {
		tr := strings.Split(registerTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + tools.MILLSSECONDADAY
			if timeStart > 0 && timeEnd > 0 {
				condCntr["register_time_start"] = timeStart
				condCntr["register_time_end"] = timeEnd
			}

			logs.Info("timeStart:%d timeEnd:%d", timeStart, timeEnd)
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
	c.Data["registerTimeRange"] = registerTimeRange
	c.Data["mobile"] = mobile

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	list, count, _ := service.BusinessDetailList(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "business_detail/list.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "business_detail/list_scripts.html"

}

func (c *BusinessDetailController) Recharge() {
	rechargeDate := c.GetString("recharge_date")

	list, err := dao.PaymentThirdpartyList()
	if err != nil {
		logs.Error("[BusinessDetailController].PaymentThirdpartyList err:%s", err)
	}
	c.Data["PaymentList"] = list
	c.Data["rechargeDate"] = rechargeDate

	c.Layout = "layout.html"
	c.TplName = "business_detail/recharge.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "business_detail/list_scripts.html"
}

func (c *BusinessDetailController) DoSaveRecharge() {

	action := "detail"
	gotoURL := "/business/detail/list"
	paymentName := c.GetString("payment_name")
	chargeAmount, _ := c.GetInt64("charge_amount")
	date := c.GetString("recharge_date")
	err := service.DoSaveRecharge(date, paymentName, chargeAmount)
	if err != nil {
		logs.Error("[BusinessDetailController].DoSaveRecharge err:%s", err)
		c.commonError(action, gotoURL, "失败")
		return
	}

	logs.Info("paymentName:%s chargeAmount:%d date:%s ", paymentName, chargeAmount, date)
	c.Redirect(gotoURL, 302)
}

func (c *BusinessDetailController) Withdraw() {
	withdrawDate := c.GetString("withdraw_date")

	list, err := dao.PaymentThirdpartyList()
	if err != nil {
		logs.Error("[BusinessDetailController].PaymentThirdpartyList err:%s", err)
	}
	c.Data["PaymentList"] = list
	c.Data["withdrawDate"] = withdrawDate

	c.Layout = "layout.html"
	c.TplName = "business_detail/withdraw.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "business_detail/list_scripts.html"
}

func (c *BusinessDetailController) DoWithdraw() {

	action := "detail"
	gotoURL := "/business/detail/list"
	paymentName := c.GetString("payment_name")
	chargeAmount, _ := c.GetInt64("withdraw_amount")
	date := c.GetString("withdraw_date")
	err := service.DoSaveWithdraw(date, paymentName, chargeAmount)
	if err != nil {
		logs.Error("[BusinessDetailController].DoWithdraw err:%s", err)
		c.commonError(action, gotoURL, "失败")
		return
	}

	logs.Info("paymentName:%s chargeAmount:%d date:%s ", paymentName, chargeAmount, date)
	c.Redirect(gotoURL, 302)
}

func (c *BusinessDetailController) ListDetail() {

	date := c.GetString("date")
	list, err := service.ListRecordByDate(date)
	if err != nil {
		logs.Error("[BusinessDetailController] ListDetail err:", err)

		action := "detail"
		gotoURL := "/business/detail/list"
		c.commonError(action, gotoURL, "日期错误")
		return
	}
	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "business_detail/detail.html"
	// c.LayoutSections = make(map[string]string)
	// c.LayoutSections["Scripts"] = "business_detail/list_scripts.html"
	return
}
