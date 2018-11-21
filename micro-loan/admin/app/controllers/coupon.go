package controllers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego/utils/pagination"

	"micro-loan/common/dao"
	"micro-loan/common/i18n"
	"micro-loan/common/lib/device"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type CouponController struct {
	BaseController
}

func (c *CouponController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "coupon"
}

func (c *CouponController) Coupon() {
	c.Layout = "layout.html"
	c.TplName = "coupon/list.html"

	var condCntr = map[string]interface{}{}

	name := c.GetString("name")
	if len(name) > 0 {
		condCntr["name"] = name
	}
	c.Data["name"] = name

	status, err := c.GetInt("status")
	if err == nil && status >= 0 {
		condCntr["status"] = status
	} else if err != nil {
		status = -1
	}
	c.Data["status"] = status

	distributeStatus, err := c.GetInt("distribute_status")
	if err == nil && distributeStatus >= 0 {
		condCntr["distribute_status"] = distributeStatus
	} else if err != nil {
		distributeStatus = -1
	}
	c.Data["distribute_status"] = distributeStatus

	splitSep := " - "
	timeRange := c.GetString("time_range")
	if len(timeRange) > 16 {
		tr := strings.Split(timeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["start_time"] = timeStart
				condCntr["end_time"] = timeEnd
			}
		}
	}
	c.Data["timeRange"] = timeRange

	couponType, err := c.GetInt("coupon_type")
	if err == nil && couponType >= 0 {
		condCntr["coupon_type"] = couponType
	} else if err != nil {
		couponType = -1
	}
	c.Data["coupon_type"] = couponType

	distributeAlgo := c.GetString("distribute_algo")
	if len(distributeAlgo) > 0 {
		condCntr["distribute_algo"] = distributeAlgo
	}
	c.Data["distribute_algo"] = distributeAlgo

	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize

	list, count, _ := service.GetAllCoupon(condCntr, page, pageSize)
	c.Data["List"] = list
	c.Data["CouponMap"] = types.CouponMap
	c.Data["DistributeStatusMap"] = types.DistributeStatusMap
	c.Data["CouponTypeMap"] = types.CouponTypeMap
	c.Data["CouponName"] = service.GetHistoryCoupon()

	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))

	c.Data["paginator"] = paginator

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "coupon/list_scripts.html"

	return
}

func (c *CouponController) CouponList() {
	c.Layout = "layout.html"
	c.TplName = "coupon/account_list.html"

	var condCntr = map[string]interface{}{}

	coupon_type, _ := c.GetInt("coupon_type")
	if coupon_type > 0 {
		condCntr["coupon_type"] = coupon_type
	}
	c.Data["coupon_type"] = coupon_type

	coupon_status, _ := c.GetInt("coupon_status")
	if coupon_status > 0 {
		condCntr["coupon_status"] = coupon_status
	}
	c.Data["coupon_status"] = coupon_status

	name := c.GetString("name")
	if name != "" {
		condCntr["name"] = name
	}
	c.Data["name"] = name

	distr_algo := c.GetString("distr_algo")
	if distr_algo != "" {
		condCntr["distr_algo"] = distr_algo
	}
	c.Data["distr_algo"] = distr_algo

	account_id, _ := c.GetInt64("account_id")
	if account_id > 0 {
		condCntr["account_id"] = account_id
	}
	c.Data["account_id"] = account_id

	coupon_id, _ := c.GetInt64("coupon_id")
	if coupon_id > 0 {
		condCntr["coupon_id"] = coupon_id
	}
	c.Data["coupon_id"] = coupon_id

	distr_range := c.GetString("distr_range")
	c.Data["distrRange"] = distr_range
	if start, end, err := tools.PareseDateRangeToMillsecond(distr_range); err == nil {
		condCntr["distr_range_start"], condCntr["distr_range_end"] = start, end
	}

	used_range := c.GetString("used_range")
	c.Data["usedRange"] = used_range
	if start, end, err := tools.PareseDateRangeToMillsecond(used_range); err == nil {
		condCntr["used_range_start"], condCntr["used_range_end"] = start, end
	}

	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize

	list, count, _ := service.GetAllAccountCoupon(condCntr, page, pageSize)

	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.Data["CouponTypeMap"] = types.CouponTypeMap
	c.Data["CouponStatusMap"] = types.CouponStatusMap
	c.Data["CouponName"] = service.GetHistoryCoupon()

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "coupon/account_list_scripts.html"

	return
}

func (c *CouponController) CouponEdit() {
	c.Layout = "layout.html"
	c.TplName = "coupon/edit.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "coupon/edit_scripts.html"
	c.Data["CouponTypeMap"] = types.CouponTypeMap
	c.Data["coupon_type"] = types.CouponTypeRedPacket
	c.Data["status"] = types.CouponInvalid
	c.Data["CouponMap"] = types.CouponMap
	c.Data["CouponName"] = service.GetHistoryCoupon()

	op, _ := c.GetInt("op")
	if op == 0 {
		op = 1
	}

	c.Data["Op"] = op

	id, _ := c.GetInt64("id")
	if id == 0 {
		return
	}

	splitSep := " - "
	coupon, _ := dao.GetCouponById(id)

	c.Data["Id"] = id
	c.Data["name"] = coupon.Name
	c.Data["distr_algo"] = coupon.DistributeAlgo
	c.Data["distrRange"] = tools.MDateMHSDate(coupon.DistributeStart) + splitSep + tools.MDateMHSDate(coupon.DistributeEnd)
	c.Data["coupon_type"] = coupon.CouponType
	c.Data["status"] = coupon.IsAvailable
	c.Data["r_min_amount"] = coupon.ValidMin
	c.Data["r_amount"] = coupon.DiscountAmount
	c.Data["d_min_amount"] = coupon.ValidMin
	c.Data["d_max_amount"] = coupon.DiscountMax
	c.Data["d_rate"] = coupon.DiscountRate
	c.Data["i_day"] = coupon.DiscountDay
	c.Data["i_min_amount"] = coupon.ValidMin
	c.Data["i_max_amount"] = coupon.DiscountMax
	c.Data["l_amount"] = coupon.DiscountAmount
	c.Data["l_max_amount"] = coupon.DiscountMax
	c.Data["distr_as_start"] = coupon.DistributeAsStart
	c.Data["coupon_start"] = tools.MDateMHSDate(coupon.ValidStart)
	c.Data["coupon_days"] = coupon.ValidDays
	c.Data["coupon_end"] = tools.MDateMHSDate(coupon.ValidEnd)
	c.Data["count"] = coupon.DistributeSize
	c.Data["comment"] = coupon.Comment

	return
}

func (c *CouponController) CouponEditSave() {
	op, _ := c.GetInt("op")
	if op == 0 {
		op = 1
	}

	splitSep := " - "
	distr_algo := c.GetString("distr_algo")
	name := c.GetString("name")
	commen := c.GetString("comment")

	//add
	if op == 1 {
		couponM := models.Coupon{}
		couponM.Id, _ = device.GenerateBizId(types.CouponBiz)
		couponM.Name = name
		couponM.DistributeAlgo = distr_algo

		distrRange := c.GetString("distr_range")
		if len(distrRange) > 16 {
			tr := strings.Split(distrRange, splitSep)
			if len(tr) == 2 {
				timeStart := tools.GetDateParseBackend(tr[0]) * 1000
				timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000 - 1000
				couponM.DistributeStart = timeStart
				couponM.DistributeEnd = timeEnd
			}
		}

		couponType, _ := c.GetInt("coupon_type")
		couponM.CouponType = types.CouponType(couponType)

		couponToday := c.GetString("coupon_today")
		if couponToday == "on" {
			couponM.DistributeAsStart = 1
		} else {
			validStart := c.GetString("coupon_start")
			if len(validStart) > 8 {
				timeStart := tools.GetDateParseBackend(validStart) * 1000
				couponM.ValidStart = timeStart
			}
		}
		couponDays, _ := c.GetInt("coupon_days")
		if couponDays > 0 {
			couponM.ValidDays = couponDays
		} else {
			validEnd := c.GetString("coupon_end")
			if len(validEnd) > 8 {
				timeEnd := tools.GetDateParseBackend(validEnd)*1000 + 3600*24*1000 - 1000
				couponM.ValidEnd = timeEnd
			}
		}
		switch couponM.CouponType {
		case types.CouponTypeRedPacket:
			{
				couponM.DiscountAmount, _ = c.GetInt64("r_amount")
				couponM.ValidMin, _ = c.GetInt64("r_min_amount")
				couponM.DiscountMax, _ = c.GetInt64("r_amount")
			}
		case types.CouponTypeDiscount:
			{
				couponM.DiscountRate, _ = c.GetInt64("d_rate")
				couponM.ValidMin, _ = c.GetInt64("d_min_amount")
				couponM.DiscountMax, _ = c.GetInt64("d_max_amount")
			}
		case types.CouponTypeInterest:
			{
				couponM.DiscountDay, _ = c.GetInt64("i_day")
				couponM.ValidMin, _ = c.GetInt64("i_min_amount")
				couponM.DiscountMax, _ = c.GetInt64("i_max_amount")
			}
		case types.CouponTypeLimit:
			{
				couponM.DiscountAmount, _ = c.GetInt64("l_amount")
				couponM.DiscountMax, _ = c.GetInt64("l_max_amount")
			}
		}

		count, _ := c.GetInt("count")
		couponM.DistributeSize = int64(count)

		couponM.Comment = commen
		couponM.IsAvailable = types.CouponAvailable
		err := service.AddCoupon(&couponM)
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/coupon"
			c.Data["message"] = "数据错误"

			return
		}

		c.Data["OpMessage"] = "增加数据成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/coupon"
		c.TplName = "success_redirect.html"
	} else if op == 2 {
		id, _ := c.GetInt64("id")

		couponM, err := dao.GetCouponById(id)
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/coupon"
			c.Data["message"] = "数据错误"

			return
		}

		count, _ := c.GetInt64("count")
		endDate := couponM.DistributeEnd
		distrRange := c.GetString("distr_range")
		if len(distrRange) > 16 {
			tr := strings.Split(distrRange, splitSep)
			if len(tr) == 2 {
				timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000 - 1000
				endDate = timeEnd
			}
		}

		couponM.Name = name
		couponM.Comment = commen

		status, _ := c.GetInt("status")
		err = service.ModifyCoupon(&couponM, status, count, endDate)
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"
			c.Data["goto_url"] = "/coupon"
			c.Data["message"] = err.Error()
		} else {
			c.Data["OpMessage"] = "更新数据成功."
			c.Layout = "layout.html"
			c.Data["Redirect"] = "/coupon"
			c.TplName = "success_redirect.html"
		}
	}
}

func (c *CouponController) CouponActive() {
	id, _ := c.GetInt64("id")

	if id == 0 {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"

		c.Data["goto_url"] = "/coupon"
		c.Data["message"] = "数据错误"
		return
	}

	err := service.ActiveCoupon(id)
	if err == nil {
		c.Data["OpMessage"] = "操作成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/coupon"
		c.TplName = "success_redirect.html"
	} else {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"
		c.Data["goto_url"] = "/coupon"
		c.Data["message"] = err.Error()
	}
}

func (c *CouponController) CouponListExport() {
	var condCntr = map[string]interface{}{}

	coupon_type, _ := c.GetInt("coupon_type")
	if coupon_type > 0 {
		condCntr["coupon_type"] = coupon_type
	}
	c.Data["coupon_type"] = coupon_type

	coupon_status, _ := c.GetInt("coupon_status")
	if coupon_status > 0 {
		condCntr["coupon_status"] = coupon_status
	}
	c.Data["coupon_status"] = coupon_status

	name := c.GetString("name")
	if name != "" {
		condCntr["name"] = name
	}
	c.Data["name"] = name

	distr_algo := c.GetString("distr_algo")
	if distr_algo != "" {
		condCntr["distr_algo"] = name
	}
	c.Data["distr_algo"] = distr_algo

	account_id, _ := c.GetInt64("account_id")
	if account_id > 0 {
		condCntr["account_id"] = account_id
	}
	c.Data["account_id"] = account_id

	distr_range := c.GetString("distr_range")
	c.Data["distrRange"] = distr_range
	if start, end, err := tools.PareseDateRangeToMillsecond(distr_range); err == nil {
		condCntr["distr_range_start"], condCntr["distr_range_end"] = start, end
	}

	used_range := c.GetString("used_range")
	c.Data["usedRange"] = used_range
	if start, end, err := tools.PareseDateRangeToMillsecond(used_range); err == nil {
		condCntr["used_range_start"], condCntr["used_range_end"] = start, end
	}

	page := 1
	pageSize := service.Pagesize * 100
	list, _, _ := service.GetAllAccountCoupon(condCntr, page, pageSize)

	fileName := fmt.Sprintf("account_coupon_%d.xlsx", tools.GetUnixMillis())
	lang := c.LangUse
	xlsx := excelize.NewFile()
	xlsx.SetCellValue("Sheet1", "A1", i18n.T(lang, "ID"))
	xlsx.SetCellValue("Sheet1", "B1", i18n.T(lang, "客户ID"))
	xlsx.SetCellValue("Sheet1", "C1", i18n.T(lang, "订单ID"))
	xlsx.SetCellValue("Sheet1", "D1", i18n.T(lang, "活动名称"))
	xlsx.SetCellValue("Sheet1", "E1", i18n.T(lang, "券种类"))
	xlsx.SetCellValue("Sheet1", "F1", i18n.T(lang, "活动生效金额"))
	xlsx.SetCellValue("Sheet1", "G1", i18n.T(lang, "派发时间"))
	xlsx.SetCellValue("Sheet1", "H1", i18n.T(lang, "使用时间"))
	xlsx.SetCellValue("Sheet1", "I1", i18n.T(lang, "过期时间"))
	xlsx.SetCellValue("Sheet1", "J1", i18n.T(lang, "状态"))

	for i, d := range list {
		xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(i+2), d.Id)
		xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(i+2), d.UserAccountId)
		xlsx.SetCellValue("Sheet1", "C"+strconv.Itoa(i+2), d.OrderId)
		xlsx.SetCellValue("Sheet1", "D"+strconv.Itoa(i+2), d.Name)
		xlsx.SetCellValue("Sheet1", "E"+strconv.Itoa(i+2), service.CouponTypeDisplay(lang, d.CouponType))
		xlsx.SetCellValue("Sheet1", "F"+strconv.Itoa(i+2), d.Amount)
		xlsx.SetCellValue("Sheet1", "G"+strconv.Itoa(i+2), tools.MDateMHS(d.Ctime))
		xlsx.SetCellValue("Sheet1", "H"+strconv.Itoa(i+2), tools.MDateMHS(d.UsedTime))
		xlsx.SetCellValue("Sheet1", "I"+strconv.Itoa(i+2), tools.MDateMHS(d.ExpireDate))
		xlsx.SetCellValue("Sheet1", "J"+strconv.Itoa(i+2), service.CouponStatusDisplay(lang, d.Status))
	}
	c.Ctx.Output.Header("Accept-Ranges", "bytes")
	c.Ctx.Output.Header("Content-Type", "application/octet-stream")
	c.Ctx.Output.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Ctx.Output.Header("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
	c.Ctx.Output.Header("Pragma", "no-cache")
	c.Ctx.Output.Header("Expires", "0")
	xlsx.Write(c.Ctx.ResponseWriter)
}

func (c *CouponController) CouponDetail() {
	c.Layout = "layout.html"
	c.TplName = "coupon/coupon_detail.html"

	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize

	id, _ := c.GetInt64("id")

	list, count, _ := service.QueryCouponRecord(id, page, pageSize)

	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "coupon/coupon_detail_scripts.html"

	return
}
