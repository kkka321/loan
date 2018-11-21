package controllers

import (
	"micro-loan/common/models"
	"micro-loan/common/pkg/rbac"
	"micro-loan/common/types"
)

// MenuController 所有menu相关的控制器入口
type MenuController struct {
	BaseController
}

// Prepare 进入Action前的逻辑
func (c *MenuController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["SuperAdminUID"] = types.SuperAdminUID

	if c.AdminUid != types.SuperAdminUID {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"

		c.Data["goto_url"] = "/index"
		c.Data["message"] = "没有权限操作"

		return
	}
}

// List 列表
func (c *MenuController) List() {
	// condCntr := map[string]interface{}{}
	// name := c.GetString("name")

	id, _ := c.GetInt64("id")
	//c.Data["status"] = status

	list, _ := rbac.ActiveMenuTreeForUpdate(id)

	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "rbac/menu.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "rbac/menu_scripts.html"
}

func (c *MenuController) Create() {
	response := map[string]interface{}{}
	modelData := &models.Menu{}

	modelData.Path = c.GetString("path")
	modelData.Class = c.GetString("class")

	if v, err := c.GetInt("status"); err == nil {
		modelData.Status = v
	} else {
		response["field"] = "status"
		response["error"] = "Status is required"
	}
	if v := c.GetString("name"); len(v) > 0 {
		modelData.Name = v
	} else {
		response["field"] = "name"
		response["error"] = "Name cannot be empty"
	}
	if v, err := c.GetInt("pid"); err == nil {
		modelData.Pid = int64(v)
	} else {
		response["field"] = "pid"
		response["error"] = "pid is required"
	}

	if v, err := c.GetInt("privilege_id"); err == nil {
		modelData.PrivilegeId = int64(v)
	} else {
		if modelData.Path != "" {
			response["field"] = "privilege_id"
			response["error"] = "PrivilegeId is required"
		}
	}

	if _, ok := response["error"]; ok {
		c.Data["json"] = response

		c.ServeJSON()
		return
	}
	id, err := rbac.AddOneMenu(modelData)
	if err != nil {
		response["error"] = err.Error()
	} else {
		response["id"] = id
	}

	c.Data["json"] = response

	c.ServeJSON()

	return
}

func (c *MenuController) UpdatePage() {

	if v, err := c.GetInt64("id"); err == nil {
		data, err := rbac.GetOneMenu(v)
		if err != nil {
			c.Data["error"] = err.Error()
		}
		c.Data["data"] = data
	} else {
		c.Data["error"] = "ID is required and must a number"
	}

	c.TplName = "rbac/menu_edit.html"

	return
}

func (c *MenuController) Update() {
	response := map[string]interface{}{}
	modelData := models.Menu{}
	cols := []string{}

	modelData.Id, _ = c.GetInt64("id")

	if v, err := c.GetInt("status"); err == nil {
		modelData.Status = v
		cols = append(cols, "Status")
	} else {
		response["field"] = "status"
		response["error"] = "Status is required"
	}
	if v := c.GetString("name"); len(v) > 0 {
		modelData.Name = v
		cols = append(cols, "Name")
	} else {
		response["field"] = "name"
		response["error"] = "Name cannot be empty"
	}
	// if v, err := c.GetInt64("pid"); err == nil {
	// 	modelData.Pid = v
	// 	cols = append(cols, "Pid")
	// } else {
	// 	response["field"] = "pid"
	// 	response["error"] = "pid is required"
	// }

	modelData.Path = c.GetString("path")
	cols = append(cols, "Path")
	modelData.Class = c.GetString("class")
	cols = append(cols, "Class")

	if v, err := c.GetInt("privilege_id"); err == nil {
		modelData.PrivilegeId = int64(v)
	} else {
		if modelData.Path != "" {
			response["field"] = "privilege_id"
			response["error"] = "PrivilegeId is required"
		}
	}

	if _, ok := response["error"]; ok {
		c.Data["json"] = response

		c.ServeJSON()
		return
	}
	id, err := rbac.UpdateOneMenu(&modelData, cols...)
	//id, err := modelData.Update(cols...)
	if id > 0 {
		response["status"] = "ok"
		response["info"] = "Update is successfully"
		response["id"] = modelData.Id
	} else {
		if err != nil {
			response["error"] = err.Error()
		} else {
			response["info"] = "No change"
			response["id"] = modelData.Id
		}

	}

	c.Data["json"] = response

	c.ServeJSON()

	return
}

func (c *MenuController) Delete() {
	response := map[string]interface{}{}
	modelData := models.Menu{}

	modelData.Id, _ = c.GetInt64("id")

	updateRows, err := rbac.DeleteOneMenu(&modelData)
	//id, err := modelData.Update(cols...)
	if updateRows > 0 {
		response["status"] = "ok"
		response["info"] = "删除成功"
		response["id"] = modelData.Id
	} else {
		if err != nil {
			response["error"] = err.Error()
		}
	}

	c.Data["json"] = response

	c.ServeJSON()

	return
}

func (c *MenuController) UpdateSort() {
	response := map[string]interface{}{}
	modelData := models.Menu{}

	modelData.Id, _ = c.GetInt64("id")
	operation := c.GetString("operation")
	affectedRows, err := rbac.UpdateMenuSort(&modelData, operation == "up")

	//id, err := modelData.Update(cols...)
	if affectedRows == 2 {
		response["status"] = "ok"
		response["info"] = "更新成功"
		response["id"] = modelData.Id
	} else {
		if err != nil {
			response["error"] = err.Error()
		} else {
			response["error"] = "未知错误"
		}
	}

	c.Data["json"] = response

	c.ServeJSON()

	return
}
