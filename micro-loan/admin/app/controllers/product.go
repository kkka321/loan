package controllers

import (
	"fmt"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/device"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type ProductController struct {
	BaseController
}

func (c *ProductController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "product"
}

func (c *ProductController) Edit() {

	c.Data["Action"] = "edit"
	id, _ := c.GetInt64("id")

	product, err := models.GetProduct(id)

	c.Data["isEdit"] = product.Id != 0
	if err == nil {
		c.Data["product"] = product
	} else {
		product = models.Product{}
		c.Data["product"] = product
	}

	c.Data["StatusNever"] = types.ProductStatusNever
	c.Data["StatusInValid"] = types.ProductStatusInValid
	c.Data["StatusValid"] = types.ProductStatusValid

	c.Data["productTypeMap"] = types.GetProductTypeMap()
	c.Data["customerVisibleTypeMap"] = types.GetCustomerVisibleTypeMap()
	c.Data["productStatusMap"] = types.GetProductStatusMap()
	c.Data["productChargeInterestTypeMap"] = types.GetProductChargeInterestTypeMap()
	c.Data["productChargeFeeTypeMap"] = types.GetProductChargeFeeTypeMap()
	c.Data["productRepayTypeMap"] = types.GetProductRepayTypeMap()
	c.Data["productPeriodMap"] = types.GetProductPeriodMap()
	c.Data["productCeilWayMap"] = types.GetProductCeilWayMap()
	c.Data["productCeilWayUnitMap"] = types.GetProductCeilWayUnitMap()

	c.Layout = "layout.html"
	c.TplName = "product/product.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "product/product_scripts.html"
}

func (c *ProductController) List() {
	c.Data["Action"] = "list"

	condCntr := map[string]interface{}{}

	sortfield := c.GetString("field")
	if len(sortfield) > 0 {
		condCntr["field"] = sortfield
	}

	sorttype := c.GetString("sort")
	if len(sorttype) > 0 {
		condCntr["sort"] = sorttype
	}

	list, _, _ := service.ListProduct(condCntr, 1, 100)
	c.Data["List"] = list

	c.Data["StatusNever"] = types.ProductStatusNever
	c.Data["StatusInValid"] = types.ProductStatusInValid
	c.Data["StatusValid"] = types.ProductStatusValid

	c.Layout = "layout.html"
	c.TplName = "product/list.html"

}

func (c *ProductController) Add() {

	name := c.GetString("name")
	productType, _ := c.GetInt("product_type")
	ver, _ := c.GetInt("ver")
	status, _ := c.GetInt("status")
	dayInterestRate, _ := c.GetInt64("day_interest_rate")
	dayFeeRate, _ := c.GetInt64("day_fee_rate")
	dayGraceRate, _ := c.GetInt64("day_grace_rate")
	dayPenaltyRate, _ := c.GetInt64("day_penalty_rate")
	chargeInterestType, _ := c.GetInt("charge_interest_type")
	chargeFeeType, _ := c.GetInt("charge_fee_type")
	period, _ := c.GetInt("period")
	minAmount, _ := c.GetInt64("min_amount")
	maxAmount, _ := c.GetInt64("max_amount")
	repayRemind, _ := c.GetInt("repay_remind")
	overdueRemind, _ := c.GetInt("overdue_remind")
	repayOrder := c.GetString("repay_order")
	ceilWay, _ := c.GetInt("ceil_way")
	ceilWayUnit, _ := c.GetInt("ceil_way_unit")
	minPeriod, _ := c.GetInt("min_period")
	maxPeriod, _ := c.GetInt("max_period")
	repayType, _ := c.GetInt("repay_type")
	gracePeriod, _ := c.GetInt("grace_period")
	penaltyCalcExpr := c.GetString("penalty_calc_expr")
	customerVisible, _ := c.GetInt("customer_visible")
	remarks := c.GetString("remarks")

	ID, _ := device.GenerateBizId(types.FinancialProduct)
	product := models.Product{
		Id:                 ID,
		Name:               name,
		Ver:                ver,
		Status:             status,
		Period:             types.ProductPeriodEunm(period),
		DayInterestRate:    dayInterestRate,
		DayFeeRate:         dayFeeRate,
		DayGraceRate:       dayGraceRate,
		DayPenaltyRate:     dayPenaltyRate,
		ChargeInterestType: types.ProductChargeInterestTypeEnum(chargeInterestType),
		ChargeFeeType:      chargeFeeType,
		MinAmount:          minAmount,
		MaxAmount:          maxAmount,
		CeilWay:            types.ProductCeilWayEunm(ceilWay),
		CeilWayUnit:        types.ProductCeilWayUnitEunm(ceilWayUnit),
		MinPeriod:          minPeriod,
		MaxPeriod:          maxPeriod,
		RepayRemind:        repayRemind,
		OverdueRemind:      overdueRemind,
		RepayOrder:         repayOrder,
		RepayType:          types.ProductRepayTypeEunm(repayType),
		GracePeriod:        gracePeriod,
		ProductType:        productType,
		PenaltyCalcExpr:    penaltyCalcExpr,
		CustomerVisible:    types.CustomerVisibleTypeEunm(customerVisible),
		Remarks:            remarks,
		Ctime:              tools.GetUnixMillis(),
		Utime:              tools.GetUnixMillis(),
	}
	product.AddProduct()

	service.ProductOptRecordWrite(c.AdminUid, c.AdminNickname, types.ProductOptTypeCreate, &product, &product)
	c.Redirect("/product/list", 302)
}

func (c *ProductController) Clone() {
	idSrc, _ := c.GetInt64("id")
	productSrc, _ := models.GetProduct(idSrc)
	productSrc.Id = 0
	// ID, _ := device.GenerateBizId(types.FinancialProduct)
	// productSrc.Id = ID

	c.Data["Action"] = "edit"
	c.Data["isEdit"] = false
	c.Data["product"] = productSrc

	c.Data["StatusNever"] = types.ProductStatusNever
	c.Data["StatusInValid"] = types.ProductStatusInValid
	c.Data["StatusValid"] = types.ProductStatusValid

	c.Data["productTypeMap"] = types.GetProductTypeMap()
	c.Data["customerVisibleTypeMap"] = types.GetCustomerVisibleTypeMap()
	c.Data["productStatusMap"] = types.GetProductStatusMap()
	c.Data["productChargeInterestTypeMap"] = types.GetProductChargeInterestTypeMap()
	c.Data["productChargeFeeTypeMap"] = types.GetProductChargeFeeTypeMap()
	c.Data["productRepayTypeMap"] = types.GetProductRepayTypeMap()
	c.Data["productPeriodMap"] = types.GetProductPeriodMap()
	c.Data["productCeilWayMap"] = types.GetProductCeilWayMap()
	c.Data["productCeilWayUnitMap"] = types.GetProductCeilWayUnitMap()

	c.Layout = "layout.html"
	c.TplName = "product/product.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "product/product_scripts.html"
}

func (c *ProductController) DoEdit() {
	ID, _ := c.GetInt64("id")
	product, _ := models.GetProduct(ID)
	org := product
	customerVisible, _ := c.GetInt("customer_visible")
	remarks := c.GetString("remarks")
	product.Remarks = remarks
	product.CustomerVisible = types.CustomerVisibleTypeEunm(customerVisible)
	product.Utime = tools.GetUnixMillis()

	//  只有未定义状态，允许更改所有
	if product.Status == int(types.ProductStatusNever) {
		name := c.GetString("name")
		productType, _ := c.GetInt("product_type")
		dayInterestRate, _ := c.GetInt64("day_interest_rate")
		dayFeeRate, _ := c.GetInt64("day_fee_rate")
		dayGraceRate, _ := c.GetInt64("day_grace_rate")
		dayPenaltyRate, _ := c.GetInt64("day_penalty_rate")
		chargeInterestType, _ := c.GetInt("charge_interest_type")
		chargeFeeType, _ := c.GetInt("charge_fee_type")
		period, _ := c.GetInt("period")
		minAmount, _ := c.GetInt64("min_amount")
		maxAmount, _ := c.GetInt64("max_amount")
		repayOrder := c.GetString("repay_order")
		ceilWay, _ := c.GetInt("ceil_way")
		ceilWayUnit, _ := c.GetInt("ceil_way_unit")
		minPeriod, _ := c.GetInt("min_period")
		maxPeriod, _ := c.GetInt("max_period")
		repayType, _ := c.GetInt("repay_type")
		gracePeriod, _ := c.GetInt("grace_period")
		penaltyCalcExpr := c.GetString("penalty_calc_expr")
		customerVisible, _ := c.GetInt("customer_visible")
		remarks := c.GetString("remarks")

		product.Name = name
		product.Period = types.ProductPeriodEunm(period)
		product.DayInterestRate = dayInterestRate
		product.DayFeeRate = dayFeeRate
		product.DayGraceRate = dayGraceRate
		product.DayPenaltyRate = dayPenaltyRate
		product.ChargeInterestType = types.ProductChargeInterestTypeEnum(chargeInterestType)
		product.ChargeFeeType = chargeFeeType
		product.MinAmount = minAmount
		product.MaxAmount = maxAmount
		product.CeilWay = types.ProductCeilWayEunm(ceilWay)
		product.CeilWayUnit = types.ProductCeilWayUnitEunm(ceilWayUnit)
		product.MinPeriod = minPeriod
		product.MaxPeriod = maxPeriod
		product.RepayOrder = repayOrder
		product.RepayType = types.ProductRepayTypeEunm(repayType)
		product.GracePeriod = gracePeriod
		product.ProductType = productType
		product.PenaltyCalcExpr = penaltyCalcExpr
		product.CustomerVisible = types.CustomerVisibleTypeEunm(customerVisible)
		product.Remarks = remarks
		product.Utime = tools.GetUnixMillis()
	}

	_, err := product.UpdateProduct()
	if err != nil {
		logs.Warn("update product err:%v  product:%v", err, product)
	}

	//写日志
	models.OpLogWrite(c.AdminUid, product.Id, models.OpCodeProductEdit, product.TableName(), org, product)
	service.ProductOptRecordWrite(c.AdminUid, c.AdminNickname, types.ProductOptTypeModify, &org, &product)
	c.Redirect("/product/list", 302)
}

// TrialCalc 产品试算接口
func (c *ProductController) TrialCalc() {

	id, _ := c.GetInt64("id")
	loan, _ := c.GetInt64("loan")
	amount, _ := c.GetInt64("amount")
	period, _ := c.GetInt("period")
	loanDate := c.GetString("loan_date")
	currentDate := c.GetString("current_date")
	repayDate := c.GetString("repay_date")
	repayedTotal, _ := c.GetInt64("repayed_total")

	c.Data["id"] = id
	c.Data["loan"] = loan
	c.Data["amount"] = amount
	c.Data["period"] = period
	c.Data["loanDate"] = loanDate
	c.Data["currentDate"] = currentDate
	c.Data["repayDate"] = repayDate
	c.Data["repayedTotal"] = repayedTotal

	if 0 != id {
		gotoURL := "trial_calc"
		action := "trial_calc"
		product, err := models.GetProduct(id)
		if nil != err {
			c.commonError(action, gotoURL, "获取产品信息异常,请检查产品地是否正确")
			return
		}

		trialIn := types.ProductTrialCalcIn{
			ID:           id,
			Loan:         loan,
			Amount:       amount,
			Period:       period,
			LoanDate:     loanDate,
			CurrentDate:  currentDate,
			RepayDate:    repayDate,
			RepayedTotal: repayedTotal,
		}
		result, err := service.ProductTrialCalc(trialIn, product)
		c.Data["product"] = product
		c.Data["result"] = result
	}
	c.Layout = "layout.html"
	c.TplName = "product/trial_calc.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "product/product_scripts.html"
}

func (c *ProductController) Up() {
	id, _ := c.GetInt64("id")
	c.Data["id"] = id
	if 0 != id {

		gotoURL := "list"
		action := "list"
		conflictId, err := service.IsProductCanActice(id)
		if err != nil {
			errMsg := fmt.Sprintf("借款期限或类型冲突。已存在上线状态产品:%d", conflictId)
			c.commonError(action, gotoURL, errMsg)
			return
		}

		product, err := models.GetProduct(id)
		origin := product
		if nil != err {
			c.commonError(action, gotoURL, "获取产品信息异常,请检查产品地是否正确")
			return
		}
		product.Status = int(types.ProductStatusValid)
		product.Utime = tools.GetUnixMillis()
		product.UpdateProduct("status", "utime")

		//写日志
		models.OpLogWrite(c.AdminUid, product.Id, models.OpCodeProductPublish, product.TableName(), origin, product)
		service.ProductOptRecordWrite(c.AdminUid, c.AdminNickname, types.ProductOptTypeUp, &product, &product)
	}

	c.Redirect("/product/list", 302)
}

func (c *ProductController) Down() {
	id, _ := c.GetInt64("id")
	c.Data["id"] = id
	if 0 != id {
		gotoURL := "list"
		action := "list"
		product, err := models.GetProduct(id)
		origin := product
		if nil != err {
			c.commonError(action, gotoURL, "获取产品信息异常,请检查产品地是否正确")
			return
		}
		product.Status = int(types.ProductStatusInValid)
		product.Utime = tools.GetUnixMillis()
		product.UpdateProduct("status", "utime")

		//写日志
		models.OpLogWrite(c.AdminUid, product.Id, models.OpCodeProductSetOff, product.TableName(), origin, product)
		service.ProductOptRecordWrite(c.AdminUid, c.AdminNickname, types.ProductOptTypeDown, &origin, &product)
	}

	c.Redirect("/product/list", 302)
}

// OptRecord 通过id返回操作记录
func (c *ProductController) OptRecordView() {
	id, err := c.GetInt64("id")
	if err == nil {
		c.Data["json"], err = models.GetProductOptRecordByPkId(id)
		if err != nil {
			logs.Warn("[product.OptRecordView] query by id:", id, " err:", err)
		}
	} else {
		c.Data["json"] = nil
	}
	c.ServeJSON()
	return
}

// OptRecord 展示产品的操作流水
func (c *ProductController) OptRecord() {
	id, _ := c.GetInt64("id")
	if id > 0 {
		productOptRecords, err := service.ListProductOptRecord(id)
		if err != nil {
			logs.Warn("[product.OptRecord] service.ListProductOptRecord id", id, " err:", err)
			return
		}
		c.Data["productOptRecords"] = productOptRecords
		c.Layout = "layout.html"
		c.TplName = "product/product_opt_reduce.html"
		c.LayoutSections = make(map[string]string)
		c.LayoutSections["Scripts"] = "product/product_scripts.html"
	} else {
		logs.Warn("[product.OptRecord] service.ListProductOptRecord id == 0")
		c.Redirect("/product/list", 302)
	}

}
