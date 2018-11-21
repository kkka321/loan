package controllers

import (
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"github.com/astaxie/beego/utils/pagination"
)

type SmsController struct {
	BaseController
}

func (c *SmsController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "sms"
}

func (c *SmsController) SmsStatusList() {
	c.Data["Action"] = "sms_status_list"

	condCntr := map[string]interface{}{}
	sms_service, err := c.GetInt("sms_service")
	if err == nil && sms_service >= 0 {
		condCntr["sms_service"] = sms_service
	}
	c.Data["sms_service"] = types.SmsServiceID(sms_service)

	sms_type, err := c.GetInt("sms_type")
	if err == nil && sms_type >= 0 {
		condCntr["sms_type"] = sms_type
	}
	c.Data["sms_type"] = types.ServiceType(sms_type)

	sms_status, err := c.GetInt("sms_status")
	if err == nil && sms_status >= 0 {
		condCntr["sms_status"] = sms_status
	}
	c.Data["sms_status"] = sms_status

	related_id, err := c.GetInt64("relatedId")
	if related_id > 0 {
		condCntr["related_id"] = related_id
	}
	c.Data["relatedId"] = related_id

	splitSep := " - "
	sendTimeRange := c.GetString("send_time_range")
	if len(sendTimeRange) > 16 {
		tr := strings.Split(sendTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["send_start_time"] = timeStart
				condCntr["send_end_time"] = timeEnd
			}
		}
	}
	c.Data["sendTimeRange"] = sendTimeRange

	sortfield := c.GetString("field")
	if len(sortfield) > 0 {
		condCntr["field"] = sortfield
	}

	sorttype := c.GetString("sort")
	if len(sorttype) > 0 {
		condCntr["sort"] = sorttype
	}

	// 分页逻辑
	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := service.Pagesize

	count, list, _, _ := service.SmsStatusList(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["serviceList"] = types.SmsServiceIdMap
	c.Data["statusList"] = types.DeliveryStatusMap
	c.Data["smsTypeList"] = types.ServiceTypeEnumMap()
	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "sms_status/list.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "sms_status/list_scripts.html"
}
