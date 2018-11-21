package controllers

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/pkg/rbac"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils/pagination"
)

type RBACController struct {
	BaseController
}

// 此模块只能超级管理员调用
func (c *RBACController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "manage"
	c.Data["SuperAdminUID"] = types.SuperAdminUID

	if c.AdminUid != types.SuperAdminUID {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"

		c.Data["goto_url"] = "/index"
		c.Data["message"] = "没有权限操作"

		return
	}
}

func (c *RBACController) OperationList() {
	condCntr := map[string]interface{}{}
	name := c.GetString("name")
	if len(name) > 0 {
		condCntr["name"] = name
	}
	c.Data["name"] = name

	//c.Data["status"] = status

	// 分页逻辑, 若 P 为
	page, _ := c.GetInt("p")
	pagesize := 20

	list, count, _ := rbac.ListOperation(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "rbac/operation.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "rbac/operation_scripts.html"
}

func (c *RBACController) OperationCreate() {
	response := map[string]interface{}{}
	modelData := &models.Operation{}

	if v := c.GetString("name"); len(v) > 0 {
		modelData.Name = v
	} else {
		response["field"] = "name"
		response["error"] = "Name cannot be empty"
	}
	if _, ok := response["error"]; ok {
		c.Data["json"] = response

		c.ServeJSON()
		return
	}
	id, err := rbac.AddOneOperation(modelData)
	if err != nil {
		response["error"] = err.Error()
	} else {
		response["id"] = id
	}

	c.Data["json"] = response

	c.ServeJSON()

	return

}

func (c *RBACController) OperationUpdatePage() {

	if v, err := c.GetInt64("id"); err == nil {
		data, err := rbac.GetOneOperation(v)
		if err != nil {
			c.Data["error"] = err.Error()
		}
		c.Data["data"] = data
	} else {
		c.Data["error"] = "ID is required and must a number"
	}

	c.TplName = "rbac/operation_edit.html"

	return
}

func (c *RBACController) OperationUpdate() {
	response := map[string]interface{}{}
	modelData := models.Operation{}

	modelData.Id, _ = c.GetInt64("id")
	cols := []string{}
	if v := c.GetString("name"); len(v) > 0 {
		modelData.Name = v
		cols = append(cols, "Name")

	} else {
		response["field"] = "name"
		response["error"] = "Name cannot be empty"
	}
	if _, ok := response["error"]; ok {
		c.Data["json"] = response

		c.ServeJSON()
		return
	}
	id, err := rbac.UpdateOneOperation(&modelData, cols...)
	//id, err := modelData.Update(cols...)
	if id > 0 {
		response["status"] = "ok"
		response["info"] = "Update is successfully"
	} else {
		if err != nil {
			response["error"] = err.Error()
		}
	}

	c.Data["json"] = response

	c.ServeJSON()

	return
}

func (c *RBACController) PrivilegeOperationManage() {

	//c.Data["status"] = status
	condCntr := map[string]interface{}{}
	condCntr["status"] = 1

	// 分页逻辑, 若 P 为

	list, _ := rbac.PrivilegeList(condCntr)

	c.Data["PrivilegeList"] = list

	c.Layout = "layout.html"
	c.TplName = "rbac/assign_operations.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "rbac/assign_operations_scripts.html"

}

func (c *RBACController) PrivilegeOperations() {
	privilegeID, _ := c.GetInt64("privilege_id[]")

	//c.Data["status"] = status

	// 分页逻辑, 若 P 为

	list, _ := rbac.AllOperationsForPrivilege(privilegeID)

	c.Data["json"] = list

	c.ServeJSON()

}

func (c *RBACController) PrivilegeAssignOperations() {

	//c.Data["status"] = status
	response := map[string]interface{}{}
	var privilegeID int64
	if v, err := c.GetInt64("privilege_id"); err == nil {
		privilegeID = v
	} else {
		response["error"] = "权限不合法, 刷新页面,重试."
		c.ServeJSON()
		return
	}

	assignOperations := c.GetStrings("assign_operations[]")
	succNum, existNum, err := rbac.AssignOperationsToPrivilege(assignOperations, privilegeID)
	if err != nil {
		response["error"] = err
	}
	response["info"] = fmt.Sprintf("准备分配操作个数 %d , 合法操作个数 %d, 分配成功个数 %d", len(assignOperations), existNum, succNum)

	response["list"], _ = rbac.AllOperationsForPrivilege(privilegeID)

	// 分页逻辑, 若 P 为
	c.Data["json"] = response

	c.ServeJSON()
}

func (c *RBACController) PrivilegeRevokeOperations() {

	//c.Data["status"] = status
	response := map[string]interface{}{}
	var privilegeID int64
	if v, err := c.GetInt64("privilege_id"); err == nil {
		privilegeID = v
	} else {
		response["error"] = "权限不合法, 刷新页面,重试."
		c.ServeJSON()
		return
	}

	assignOperations := c.GetStrings("assign_operations[]")
	succNum, err := rbac.RevokeOperationsFromPrivilege(assignOperations, privilegeID)
	if err != nil {
		response["error"] = err
	}
	response["info"] = fmt.Sprintf("准备移除操作个数 %d , 移除成功个数 %d", len(assignOperations), succNum)

	response["list"], _ = rbac.AllOperationsForPrivilege(privilegeID)

	//strings.Join(assignPrivileges, ",")
	//logs.Debug(assignPrivileges)
	// 分页逻辑, 若 P 为
	c.Data["json"] = response

	c.ServeJSON()

}

func (c *RBACController) RoleList() {
	condCntr := map[string]interface{}{}
	name := c.GetString("name")
	if len(name) > 0 {
		condCntr["name"] = name
	}
	c.Data["name"] = name

	id, _ := c.GetInt64("id")
	// 当 status 不存在,或者 status 为"" 时, err != nil
	if v, err := c.GetInt("status"); err == nil {
		condCntr["status"] = v
	}

	list, err := rbac.ActiveRoleTreeForUpdate(id)
	if err != nil {
		logs.Error(err)
	}
	//list, _ := rbac.RoleList(condCntr)

	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "rbac/role_tree.html"
	//c.TplName = "rbac/role.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "rbac/role_scripts.html"
}

func (c *RBACController) RoleCreate() {
	response := map[string]interface{}{}
	modelData := &models.Role{}

	if v, err := c.GetInt("status"); err == nil {
		modelData.Status = v
	} else {
		response["field"] = "status"
		response["error"] = "Status is required"
	}
	if v, err := c.GetInt64("pid"); err == nil {
		modelData.Pid = v
	} else {
		response["field"] = "pid"
		response["error"] = "Pid is required"
	}
	if v, err := c.GetInt("type"); err == nil {
		modelData.Type = types.RoleTypeEnum(v)
	} else {
		response["field"] = "type"
		response["error"] = "Type is required"
	}
	if v := c.GetString("name"); len(v) > 0 {
		modelData.Name = v
	} else {
		response["field"] = "name"
		response["error"] = "Name cannot be empty"
	}
	if _, ok := response["error"]; ok {
		c.Data["json"] = response

		c.ServeJSON()
		return
	}
	id, err := rbac.AddOneRole(modelData)
	if err != nil {
		response["error"] = err.Error()
	} else {
		response["id"] = id
	}

	c.Data["json"] = response

	c.ServeJSON()

	return
}

func (c *RBACController) RolePrivilegeManage() {

	//c.Data["status"] = status
	condCntr := map[string]interface{}{}
	condCntr["status"] = 1

	// 分页逻辑, 若 P 为

	list, _ := rbac.RoleList(condCntr)

	c.Data["RoleList"] = list

	c.Layout = "layout.html"
	c.TplName = "rbac/assign_privilege.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "rbac/assign_privilege_scripts.html"

}

func (c *RBACController) RolePrivileges() {
	roleID, _ := c.GetInt64("role_id[]")

	//c.Data["status"] = status

	// 分页逻辑, 若 P 为

	list, _ := rbac.AllPrivilegesForRole(roleID)

	c.Data["json"] = list

	c.ServeJSON()

}

func (c *RBACController) RoleAssignPrivileges() {

	//c.Data["status"] = status
	response := map[string]interface{}{}
	var roleID int64
	if v, err := c.GetInt64("role_id"); err == nil {
		roleID = v
	} else {
		response["error"] = "Role is invalid, reload the page , do it again."
		c.ServeJSON()
		return
	}

	assignPrivileges := c.GetStrings("assign_privileges[]")
	succNum, existNum, err := rbac.AssignPrivilegesToRole(assignPrivileges, roleID)
	if err != nil {
		response["error"] = err
	}
	response["info"] = fmt.Sprintf("Want to assign %d, exist privileges %d, assign successfully %d", len(assignPrivileges), existNum, succNum)

	response["list"], _ = rbac.AllPrivilegesForRole(roleID)

	//strings.Join(assignPrivileges, ",")
	//logs.Debug(assignPrivileges)
	// 分页逻辑, 若 P 为
	c.Data["json"] = response

	c.ServeJSON()
}

func (c *RBACController) RoleRevokePrivileges() {

	//c.Data["status"] = status
	response := map[string]interface{}{}
	var roleID int64
	if v, err := c.GetInt64("role_id"); err == nil {
		roleID = v
	} else {
		response["error"] = "权限不合法, 刷新页面,重试."
		c.ServeJSON()
		return
	}

	privileges := c.GetStrings("privileges[]")
	succNum, err := rbac.RevokePrivilegesFromRole(privileges, roleID)
	if err != nil {
		response["error"] = err
	}
	response["info"] = fmt.Sprintf("准备移除权限个数 %d , 移除成功个数 %d", len(privileges), succNum)

	response["list"], _ = rbac.AllPrivilegesForRole(roleID)

	//strings.Join(assignPrivileges, ",")
	//logs.Debug(assignPrivileges)
	// 分页逻辑, 若 P 为
	c.Data["json"] = response

	c.ServeJSON()

}

func (c *RBACController) RoleUpdate() {
	response := map[string]interface{}{}
	modelData := models.Role{}

	cols := []string{}
	if v, err := c.GetInt64("id"); err == nil {
		modelData.Id = v
	} else {
		response["field"] = "id"
		response["error"] = "ID 不合法, 刷新页面重试"
	}
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
	if _, ok := response["error"]; ok {
		c.Data["json"] = response

		c.ServeJSON()
		return
	}
	id, err := rbac.UpdateOneRole(&modelData, cols...)
	if id > 0 {
		response["status"] = "ok"
		response["info"] = "Update is successfully"
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

func (c *RBACController) RoleEditPage() {

	if v, err := c.GetInt("id"); err == nil {
		data, err := rbac.GetOneRole(int64(v))
		if err != nil {
			c.Data["error"] = err.Error()
		}
		c.Data["data"] = data
	} else {
		c.Data["error"] = "ID is required and must a number"
	}

	c.TplName = "rbac/role_edit.html"

	return
}

func (c *RBACController) PrivilegeList() {
	condCntr := map[string]interface{}{}
	name := c.GetString("name")
	if len(name) > 0 {
		condCntr["name"] = name
	}
	c.Data["name"] = name

	if v, err := c.GetInt64("group_id"); err == nil && v > 0 {
		condCntr["group_id"] = v
		c.Data["groupID"] = v
	} else {
		var groupID int64
		c.Data["groupID"] = groupID
	}
	//c.Data["status"] = status

	// 分页逻辑, 若 P 为
	page, _ := c.GetInt("p")
	pagesize := 20

	list, count, _ := rbac.ListPrivilege(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list
	c.Data["GroupList"], _ = rbac.PrivilegeGroupList()

	c.Layout = "layout.html"
	c.TplName = "rbac/privilege.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "rbac/privilege_scripts.html"
}

func (c *RBACController) PrivilegeCreate() {
	response := map[string]interface{}{}
	modelData := &models.Privilege{}

	if v := c.GetString("name"); len(v) > 0 {
		modelData.Name = v
	} else {
		response["field"] = "name"
		response["error"] = "Name cannot be empty"
	}
	if v, err := c.GetInt64("group_id"); err == nil {
		modelData.GroupID = v
	} else {
		response["field"] = "group_id"
		response["error"] = "选择一个权限组"
	}
	if _, ok := response["error"]; ok {
		c.Data["json"] = response

		c.ServeJSON()
		return
	}
	id, err := rbac.AddOnePrivilege(modelData)
	if err != nil {
		response["error"] = err.Error()
	} else {
		response["id"] = id
	}

	c.Data["json"] = response

	c.ServeJSON()

	return

}

func (c *RBACController) PrivilegeUpdatePage() {

	if v, err := c.GetInt64("id"); err == nil {
		data, err := rbac.GetOnePrivilege(v)
		if err != nil {
			c.Data["error"] = err.Error()
		}
		c.Data["data"] = data
	} else {
		c.Data["error"] = "ID is required and must a number"
	}

	c.TplName = "rbac/privilege_edit.html"

	return
}

func (c *RBACController) PrivilegeUpdate() {
	response := map[string]interface{}{}
	modelData := models.Privilege{}

	cols := []string{}
	if v, err := c.GetInt64("id"); err == nil {
		modelData.Id = v
	} else {
		response["field"] = "id"
		response["error"] = "ID 不合法, 刷新页面重试"
	}

	if v := c.GetString("name"); len(v) > 0 {
		modelData.Name = v
		cols = append(cols, "Name")

	} else {
		response["field"] = "name"
		response["error"] = "Name cannot be empty"
	}
	if _, ok := response["error"]; ok {
		c.Data["json"] = response
		c.ServeJSON()
		return
	}
	id, err := rbac.UpdateOnePrivilege(&modelData, cols...)
	if id > 0 {
		response["status"] = "ok"
		response["info"] = "Update is successfully"
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

func (c *RBACController) PrivilegeGroupList() {
	condCntr := map[string]interface{}{}
	name := c.GetString("name")
	if len(name) > 0 {
		condCntr["name"] = name
	}
	c.Data["name"] = name

	//c.Data["status"] = status

	list, _ := rbac.ListPrivilegeGroup(condCntr)

	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "rbac/privilege_group.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "rbac/privilege_group_scripts.html"
}

func (c *RBACController) PrivilegeGroupCreate() {
	response := map[string]interface{}{}
	modelData := &models.PrivilegeGroup{}

	if v := c.GetString("name"); len(v) > 0 {
		modelData.Name = v
	} else {
		response["field"] = "name"
		response["error"] = "Name cannot be empty"
	}

	if _, ok := response["error"]; ok {
		c.Data["json"] = response

		c.ServeJSON()
		return
	}
	id, err := rbac.AddOnePrivilegeGroup(modelData)
	if err != nil {
		response["error"] = err.Error()
	} else {
		response["id"] = id
	}

	c.Data["json"] = response

	c.ServeJSON()

	return

}

func (c *RBACController) PrivilegeGroupUpdatePage() {

	if v, err := c.GetInt64("id"); err == nil {
		data, err := rbac.GetOnePrivilegeGroup(v)
		if err != nil {
			c.Data["error"] = err.Error()
		}
		c.Data["data"] = data
	} else {
		c.Data["error"] = "ID is required and must a number"
	}

	c.TplName = "rbac/privilege_group_edit.html"

	return
}

func (c *RBACController) PrivilegeGroupUpdate() {
	response := map[string]interface{}{}
	modelData := models.PrivilegeGroup{}

	cols := []string{}
	if v, err := c.GetInt64("id"); err == nil {
		modelData.Id = v
	} else {
		response["field"] = "id"
		response["error"] = "ID 不合法, 刷新页面重试"
	}

	if v := c.GetString("name"); len(v) > 0 {
		modelData.Name = v
		cols = append(cols, "Name")

	} else {
		response["field"] = "name"
		response["error"] = "Name cannot be empty"
	}
	if _, ok := response["error"]; ok {
		c.Data["json"] = response
		c.ServeJSON()
		return
	}
	id, err := rbac.UpdateOnePrivilegeGroup(&modelData, cols...)
	if id > 0 {
		response["status"] = "ok"
		response["info"] = "Update is successfully"
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
