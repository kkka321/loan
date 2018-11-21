package controllers

import (
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils/pagination"

	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type ThirdPartyController struct {
	BaseController
}

func (c *ThirdPartyController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "thirdparty"
}

func (c *ThirdPartyController) Add() {
	name := c.GetString("name")
	api := c.GetString("api")
	price, _ := c.GetInt("price")
	chargeType, _ := c.GetInt("charge_type")

	thirdparty := models.ThirdpartyInfo{
		Index:      0,
		Name:       name,
		Api:        api,
		ChargeType: chargeType,
		Price:      price,
		Ctime:      tools.GetUnixMillis(),
		Utime:      tools.GetUnixMillis(),
	}

	id, err := thirdparty.Add()
	if err != nil {
		logs.Error("[ThirdPartyController.Add] err:", err, " id:", id, " thirdparty:%#v", thirdparty)
	}

	c.Layout = "layout.html"
	c.TplName = "thirdparty/thirdparty.html"
	c.Redirect("/thirdparty/list", 302)
}

func (c *ThirdPartyController) Edit() {
	thirdparty := models.ThirdpartyInfo{}
	var err error

	id, _ := c.GetInt64("id")
	if id > 0 {
		thirdparty, err = models.GetThirdpartyInfoByPkId(id)
	}
	c.Data["isEdit"] = thirdparty.Id != 0
	if err == nil {
		c.Data["thirdparty"] = thirdparty
	} else {
		logs.Warning("[controller.ThirdPartyController.Edit] err", err)
		c.Data["thirdparty"] = thirdparty
	}

	c.Data["chargeTypeMap"] = types.ChargeTypeMap()

	c.Layout = "layout.html"
	c.TplName = "thirdparty/thirdparty.html"
	// c.LayoutSections = make(map[string]string)
	// c.LayoutSections["Scripts"] = "thirdparty/thirdparty.html"
}

func (c *ThirdPartyController) DoEdit() {
	id, _ := c.GetInt64("id")
	thirdparty, err := models.GetThirdpartyInfoByPkId(id)
	if err != nil {
		logs.Error("[controller.ThirdPartyController.DoEdit.GetThirdpartyInfoByPkId] id: ", id, " err:", err)
		return
	}
	org := thirdparty

	// api := c.GetString("api")
	price, _ := c.GetInt("price")
	chargeType, _ := c.GetInt("charge_type")
	remarks := c.GetString("remarks")

	// thirdparty.Api = api
	thirdparty.Price = price
	thirdparty.Remarks = remarks
	thirdparty.ChargeType = chargeType
	thirdparty.Utime = tools.GetUnixMillis()

	gotoURL := "/thirdparty/list"
	err = thirdparty.Upadte("price", "charge_type", "remarks", "utime")
	if err != nil {
		logs.Error("update thirdparty err:%v  thirdparty:%#v", err, thirdparty)
		c.commonError("", gotoURL, "更新数据库失败")
		return
	}

	//写日志
	models.OpLogWrite(c.AdminUid, thirdparty.Id, models.OpCodeProductEdit, thirdparty.TableName(), org, thirdparty)
	c.Redirect(gotoURL, 302)
}

func (c *ThirdPartyController) List() {
	c.Data["Action"] = "list"

	list, _ := models.ThirdpartyInfoList()
	c.Data["List"] = list
	c.Layout = "layout.html"
	c.TplName = "thirdparty/list.html"

}

func (c *ThirdPartyController) Reconciliation() {
	var condCntr = map[string]interface{}{}

	//1、name
	name := c.GetString("name")
	if name != "" {
		condCntr["name"] = name
	}
	c.Data["name"] = name

	apiUrl := c.GetString("api_url")
	if len(apiUrl) > 0 {
		apiUrl = tools.Strim(apiUrl)
		condCntr["api_url"] = apiUrl
	}
	c.Data["apiUrl"] = apiUrl

	//2、time range
	splitSep := " - "
	statisticTimeRange := c.GetString("statistic_time_range")
	if len(statisticTimeRange) > 16 {
		tr := strings.Split(statisticTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["statistic_start"] = timeStart
				condCntr["statistic_end"] = timeEnd
			}
		}
	}
	c.Data["statisticTimeRange"] = statisticTimeRange

	//3、charge type
	chargeType, _ := c.GetInt("charge_type")
	if chargeType > 0 {
		condCntr["charge_type"] = chargeType
	}
	c.Data["chargeType"] = chargeType

	//4、query
	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize
	list, total, totalCount, totalSuccessCount, totalChargeAmout, err := service.ListThirdpartyStatisticFee(condCntr, page, pageSize)
	if err != nil {
		logs.Error("[thirdparty.Reconciliation] err:", err)
	}
	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(total))
	c.Data["paginator"] = paginator

	//5、show
	c.Data["chargeTypeMap"] = types.ChargeTypeMap()
	c.Data["list"] = list
	c.Data["totalCount"] = totalCount
	c.Data["totalSuccessCount"] = totalSuccessCount
	c.Data["totalChargeAmout"] = totalChargeAmout

	c.Layout = "layout.html"
	c.TplName = "thirdparty/reconciliation.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "thirdparty/list_scripts.html"
}

// 客户ID；手机号;客户姓名；客户分类；当前支出； 第三方服务名称1支出；第三方服务名称2支出；……

func (c *ThirdPartyController) ReconciliationCustomer() {
	var condCntr = map[string]interface{}{}

	//userAccountId
	userAccountId, _ := c.GetInt64("user_account_id")
	if userAccountId != 0 {
		condCntr["user_account_id"] = userAccountId
	}
	c.Data["userAccountId"] = userAccountId

	//2、mobile
	mobile := c.GetString("mobile")
	if mobile != "" {
		condCntr["mobile"] = mobile
	}
	c.Data["mobile"] = mobile

	//3、tags
	// tagsInt, _ := tools.Str2Int(c.GetString("tags"))
	// tags := types.CustomerTags(tagsInt)
	// if tags > 0 {
	// 	condCntr["tags"] = tags
	// }
	// c.Data["tags"] = tags

	//4、mediaSource
	mediaSource := c.GetString("media_source")
	if len(mediaSource) > 0 {
		condCntr["media_source"] = mediaSource
	}
	c.Data["mediaSource"] = mediaSource

	//4、campaign
	campaign := c.GetString("campaign")
	if len(campaign) > 0 {
		condCntr["campaign"] = campaign
	}
	c.Data["campaign"] = campaign

	// //4、query
	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize
	list, total, err := service.ListThirdpartyStatisticCustomer(condCntr, page, pageSize)
	if err != nil {
		logs.Error("[thirdparty.Reconciliation] err:", err)
	}
	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(total))
	c.Data["paginator"] = paginator

	//5、show
	c.Data["CustomerTagsMap"] = types.CustomerTagsMap()
	c.Data["list"] = list

	c.Layout = "layout.html"
	c.TplName = "thirdparty/reconciliation_customer.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "thirdparty/list_scripts.html"
}

func (c *ThirdPartyController) ReconciliationCustomerDetail() {
	// var condCntr = map[string]interface{}{}

	//userAccountId

	id, err := c.GetInt64("id")
	if err == nil {
		c.Data["json"], err = service.ListThirdpartyStatisticCustomerDetail(id)
		if err != nil {
			logs.Warn("[thirdparty.ReconciliationCustomerDetail] query by id:", id, " err:", err)
		}
	} else {
		c.Data["json"] = nil
	}
	c.ServeJSON()
	return
}
