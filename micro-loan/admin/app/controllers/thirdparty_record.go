package controllers

import (
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils/pagination"

	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"
)

type ThirdPartyRecordController struct {
	BaseController
}

func (c *ThirdPartyRecordController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "thirdparty_record"
}

// List 列表
func (c *ThirdPartyRecordController) List() {
	var condCntr = map[string]interface{}{}
	relatedID := c.GetString("related_id")
	if len(relatedID) > 0 {
		condCntr["related_id"] = relatedID
	}
	c.Data["related_id"] = relatedID

	//ID检索
	idCheck := c.GetString("id_check")
	if len(idCheck) > 0 {
		condCntr["id_check"] = idCheck
	}
	c.Data["id_check"] = idCheck

	//api
	api := c.GetString("api")
	if len(api) > 0 {
		condCntr["api"] = api
	}
	c.Data["api"] = api

	//request
	request := c.GetString("request")
	if len(request) > 0 {
		condCntr["request"] = request
	}
	c.Data["request"] = request

	//response
	response := c.GetString("response")
	if len(response) > 0 {
		condCntr["response"] = response
	}
	c.Data["response"] = response

	//第三方类型
	thirdparty, _ := c.GetInt("thirdparty", -1)
	if thirdparty > 0 {
		condCntr["thirdparty"] = thirdparty
	}
	c.Data["thirdparty"] = thirdparty

	month, _ := c.GetInt64("month", 0)
	if month > 0 {
		condCntr["month"] = month
	}
	c.Data["month"] = month
	selectedThirdpartyMap := map[interface{}]interface{}{}
	// valid, _ := strconv.Atoi(thirdparty)
	selectedThirdpartyMap[thirdparty] = nil
	c.Data["selectedThirdpartyMap"] = selectedThirdpartyMap

	//每页条数
	pageNumber, _ := c.GetInt("page_number", 1)
	selectedTagPageMap := map[interface{}]interface{}{}
	selectedTagPageMap[pageNumber] = nil
	c.Data["selectedTagPageMap"] = selectedTagPageMap

	splitSep := " - "
	//
	cTimeRange := c.GetString("ctime_range")
	if len(cTimeRange) > 16 {
		tr := strings.Split(cTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["ctime_start"] = timeStart
				condCntr["ctime_end"] = timeEnd
			}
		}
	}
	c.Data["cTimeRange"] = cTimeRange
	//ctimeRange := c.GetString("ctime_range")
	// c.Data["ctimeRange"] = ctimeRange
	// if start, end, err := tools.PareseDateRangeToMillsecond(ctimeRange); err == nil {
	// 	condCntr["ctime_start"], condCntr["ctime_end"] = start, end
	// }

	c.Data["thirdpartyMap"] = models.ThirdpartyNameMap
	c.Data["monthMap"] = service.GetMonthMap()

	page, _ := c.GetInt("p")
	pagesize := 15

	list, count, _ := service.ThirdpartyListBackend(condCntr, page, pagesize)

	logs.Debug("list:", list)

	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list
	// 获取指定

	c.Layout = "layout.html"
	c.TplName = "thirdparty_record/list.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "thirdparty_record/list_scripts.html"
}

func (c *ThirdPartyRecordController) Detail() {
	m := models.ThirdpartyRecord{}

	id, _ := c.GetInt64("id")
	ctime, _ := c.GetInt64("ctime")
	month := tools.GetMonth(ctime)
	monthMap := service.GetMonthMap()
	tableName := m.OriTableName()
	if _, ok := monthMap[month]; ok {
		tableName = m.TableNameByMonth(month)
	}

	thirdparthInfo, _ := dao.GetThirdpartyOne(tableName, id)
	logs.Debug("[Detail] thirdpartyinfo:", thirdparthInfo)
	c.Data["data"] = thirdparthInfo
	c.Layout = "layout.html"
	c.TplName = "thirdparty_record/detail.html"
}
