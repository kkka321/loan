package controllers

import (
	"fmt"

	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type SysConfigController struct {
	BaseController
}

func (c *SysConfigController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "sysconfig"
}

func (c *SysConfigController) SystemConfigList() {
	c.Data["Action"] = "system_config/list"

	c.Data["StatusMap"] = types.StatusMap()
	c.Data["SystemConfigItemTypeMap"] = types.SystemConfigItemTypeMap()

	c.Layout = "layout.html"
	c.TplName = "sysconfig/system_config_list.html"

	condBox := map[string]interface{}{}

	isValid, _ := tools.Str2Int(c.GetString("status", fmt.Sprintf("%d", types.StatusValid)))
	if isValid >= 0 {
		condBox["status"] = isValid
	}
	c.Data["status"] = isValid
	itemName := c.GetString("item_name")
	if len(itemName) > 0 {
		condBox["item_name"] = itemName
	}
	c.Data["item_name"] = itemName

	list, _, _ := config.List(condBox)
	c.Data["List"] = list

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "sysconfig/system_config.js.html"
}

func (c *SysConfigController) SystemConfigSave() {
	response := map[string]interface{}{}

	itemName := c.GetString("item_name")
	itemTypeP := c.GetString("item_type")
	itemTypeInt, _ := tools.Str2Int(itemTypeP)
	itemType := types.SystemConfigItemType(itemTypeInt)
	itemValue := c.GetString("item_value")
	weight, _ := c.GetInt("weight")
	description := c.GetString("description")

	if len(itemName) <= 0 || len(itemValue) <= 0 || itemType <= 0 {
		response["code"] = 1
		response["message"] = "缺少必要参数"
	} else {
		id, err := config.Create(itemName, itemValue, itemType, weight, description, c.AdminUid)
		if err != nil {
			response["code"] = 2
			response["message"] = "保存新配置项失败"
		} else {
			response["code"] = 0
			response["message"] = "操作成功"
			response["data"] = map[string]interface{}{
				"lastInsertID": id,
			}
		}
	}

	c.Data["json"] = response
	c.ServeJSON()
}
