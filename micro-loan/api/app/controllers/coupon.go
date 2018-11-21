package controllers

import (
	"micro-loan/common/service"

	"micro-loan/common/cerror"
	"micro-loan/common/tools"

	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
)

type CouponController struct {
	ApiBaseController
}

func (c *CouponController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()

	// 统一将 ip 加到 RequestJSON 中
	c.RequestJSON["ip"] = c.Ctx.Input.IP()
	c.RequestJSON["related_id"] = int64(0)
}

func (c *CouponController) List() {
	if !service.CheckCouponListRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	couponTypes := []types.CouponType{types.CouponTypeRedPacket, types.CouponTypeDiscount, types.CouponTypeInterest}

	offset, _ := tools.Str2Int(c.RequestJSON["offset"].(string))
	service.QueryAccountCoupon(c.AccountID, couponTypes, offset, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *CouponController) ListV2() {
	if !service.CheckCouponListRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	couponTypes := []types.CouponType{types.CouponTypeRedPacket, types.CouponTypeDiscount, types.CouponTypeInterest, types.CouponTypeLimit}

	offset, _ := tools.Str2Int(c.RequestJSON["offset"].(string))
	service.QueryAccountCoupon(c.AccountID, couponTypes, offset, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *CouponController) Active() {
	if !service.CheckCouponActiveRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	period, _ := tools.Str2Int(c.RequestJSON["period"].(string))
	loan, _ := tools.Str2Int64(c.RequestJSON["loan"].(string))
	amount, _ := tools.Str2Int64(c.RequestJSON["amount"].(string))

	product, err := service.ProductSuitablesByPeriod(c.AccountID, period, loan)
	if err != nil {
		logs.Error("[Active]ProductSuitablesByPeriod can not find product. accountId:%d, periodNew:%d, err:%v", c.AccountID, period, err)
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	couponTypes := []types.CouponType{types.CouponTypeRedPacket, types.CouponTypeDiscount, types.CouponTypeInterest}
	service.QueryAccountCouponActive(c.AccountID, couponTypes, loan, amount, period, &product, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *CouponController) ActiveV2() {
	if !service.CheckCouponActiveRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	period, _ := tools.Str2Int(c.RequestJSON["period"].(string))
	loan, _ := tools.Str2Int64(c.RequestJSON["loan"].(string))
	amount, _ := tools.Str2Int64(c.RequestJSON["amount"].(string))

	product, err := service.ProductSuitablesByPeriod(c.AccountID, period, loan)
	if err != nil {
		logs.Error("[ActiveV2]ProductSuitablesByPeriod can not find product. accountId:%d, periodNew:%d, err:%v loan:%d", c.AccountID, period, err, loan)
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	couponTypes := []types.CouponType{types.CouponTypeRedPacket, types.CouponTypeDiscount, types.CouponTypeInterest, types.CouponTypeLimit}
	service.QueryAccountCouponActive(c.AccountID, couponTypes, loan, amount, period, &product, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *CouponController) HasNew() {
	if !service.CheckCouponNewRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	hasNew := 0
	list, _ := service.QueryNewAccountCoupon(c.AccountID)
	if len(list) > 0 {
		hasNew = 1
	}

	data["has_new"] = hasNew

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *CouponController) MarkNew() {
	if !service.CheckMarkNewRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	go service.MarkNewAccountCoupon(c.AccountID)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}
