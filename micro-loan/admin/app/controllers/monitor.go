package controllers

import (
	"micro-loan/common/service"
	"micro-loan/common/models"
)

type MonitorController struct {
	BaseController
}

func (c *MonitorController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "monitor"
}

func (c *MonitorController) Monitor() {
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "monitor/list_scripts.html"

	thirdparty, _ := c.GetInt("thirdparty")
	dataType, _ := c.GetInt("data_type")

	c.Data["data_type"] = dataType
	c.Data["thirdparty"] = thirdparty
	c.Data["ThirdpartyList"] = models.ThirdpartyNameMap
	c.Layout = "layout.html"
	c.TplName = "monitor/list_new.html"
}

func (c *MonitorController) List() {
	dataType, _ := c.GetInt("data_type")
	c.GetInt("chart_type")
	thirdparty, _ := c.GetInt("thirdparty")
	page := 0
	pagesize := 15

	condStr := map[string]interface{}{}
	condStr["thirdparty"] = thirdparty

	response := map[string]interface{}{}
	switch dataType {
	case 0:
		list := service.GetOrderTotalData(condStr, page, pagesize)
		response["response"] = list
	case 1:
		list := service.GetOrderStatisticsData(condStr, page, pagesize)
		response["response"] = list
	case 2:
		list := service.GetThirdpartyTotalData(condStr, page, pagesize)
		response["response"] = list
	case 3:
		list := service.GetThirdpartyStatisticsData(condStr, page, pagesize)
		response["response"] = list
	case 4:
		list := service.GetApiStatisticsData(condStr, page, pagesize)
		response["response"] = list
	}

	c.Data["data_type"] = dataType
	c.Data["thirdparty"] = thirdparty
	c.Data["ThirdpartyList"] = models.ThirdpartyNameMap
	c.Data["json"] = response
	c.ServeJSON()
}
