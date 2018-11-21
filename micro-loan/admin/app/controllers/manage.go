package controllers

import (
	"encoding/json"
	"strings"

	"github.com/astaxie/beego/utils/pagination"

	"micro-loan/common/models"
	"micro-loan/common/pkg/admin"
	"micro-loan/common/pkg/rbac"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type ManageController struct {
	BaseController
}

// 此模块只能超级管理员调用
func (c *ManageController) Prepare() {
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

func (c *ManageController) AdminList() {
	c.Data["Action"] = "admin/list"

	c.Layout = "layout.html"
	c.TplName = "manage/admin_list.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "manage/admin_scripts.html"

	condCntr := map[string]interface{}{}
	list, _, _ := admin.List(condCntr, 1, 500)
	c.Data["List"] = list
}

func (c *ManageController) AdminCreate() {
	c.Data["Action"] = "admin/create"
	c.Data["IsEdit"] = false
	c.Data["RoleList"], _ = rbac.RoleList(map[string]interface{}{})
	c.Layout = "layout.html"
	c.TplName = "manage/admin_create.html"
}

func (c *ManageController) AdminSave() {
	action := "admin/create"

	email := c.GetString("email")
	mobile := c.GetString("mobile")
	nickname := c.GetString("nickname")
	password := c.GetString("password")
	// TODO 检查是否有添加此角色的权限
	// 角色是否在可添加角色列表中
	roleID, _ := c.GetInt64("role_id")

	if len(email) < 8 || len(mobile) < 8 || len(nickname) < 3 || len(password) < 8 {
		c.adminError(action, "缺少必要参数")
		return
	}

	registerTime := tools.GetUnixMillis()
	adminModel := models.Admin{
		Email:        email,
		Mobile:       mobile,
		Nickname:     nickname,
		RoleID:       roleID,
		CreateUid:    c.AdminUid,
		Password:     tools.PasswordEncrypt(password, registerTime),
		RegisterTime: registerTime,
		Status:       types.StatusValid,
	}
	_, err := admin.Add(&adminModel)
	if err != nil {
		c.adminError(action, "创建失败:"+err.Error())
		return
	}

	c.Redirect("/manage/admin/list", 302)
}

func (c *ManageController) AdminEdit() {
	c.Data["RoleList"], _ = rbac.RoleList(map[string]interface{}{"status": types.StatusValid})

	if v, err := c.GetInt64("id"); err == nil {
		data, err := models.OneAdminByUid(v)
		if err != nil {
			c.Data["error"] = err.Error()
		}
		c.Data["Data"] = data
	} else {
		c.Data["error"] = "非法请求,无用户ID参数"
	}

	c.TplName = "manage/admin_edit.html"

	return
}

func (c *ManageController) AdminUpdate() {
	response := map[string]interface{}{}
	modelData := models.Admin{}
	cols := []string{}

	modelData.Id, _ = c.GetInt64("id")
	data, _ := models.OneAdminByUid(modelData.Id)
	modelData = data
	if v := c.GetString("nickname"); len(v) > 0 {
		if data.Nickname != v {
			modelData.Nickname = v
			cols = append(cols, "Nickname")
		}
	} else {
		response["field"] = "nickname"
		response["error"] = "昵称必填"
	}

	if v := c.GetString("mobile"); len(v) > 8 {
		if data.Mobile != v {
			modelData.Mobile = v
			cols = append(cols, "Mobile")
		}
	} else {
		response["field"] = "mobile"
		response["error"] = "手机号最少9位"
	}

	if v := c.GetString("email"); len(v) > 5 {
		if data.Email != v {
			modelData.Email = v
			cols = append(cols, "Email")
		}
	} else {
		response["field"] = "email"
		response["error"] = "邮箱不合法"
	}

	if v := c.GetString("password"); len(v) > 0 {
		if len(v) > 8 {
			modelData.Password = tools.PasswordEncrypt(v, data.RegisterTime)
			if data.Password != modelData.Password {
				cols = append(cols, "Password")
			}
		} else {
			response["field"] = "password"
			response["error"] = "密码不合法, 最少9位"
		}
	}

	if v, err := c.GetInt64("role_id"); err == nil {
		if data.RoleID != v {
			modelData.RoleID = v
			cols = append(cols, "RoleID")
		}
	} else {
		response["field"] = "role_id"
		response["error"] = "角色必选"
	}

	if _, ok := response["error"]; ok {
		c.Data["json"] = response

		c.ServeJSON()
		return
	}

	nums, err := admin.Update(&modelData, &data, cols)
	if nums > 0 {
		response["status"] = "ok"
		response["info"] = "Update is successfully"
		response["id"] = modelData.Id
	} else {
		if err != nil {
			response["error"] = err.Error()
		} else {
			response["error"] = "No change , no need to update"
		}
	}
	c.Data["json"] = response

	c.ServeJSON()

	return
}

func (c *ManageController) AdminBlock() {
	action := "admin/block"
	id, _ := tools.Str2Int64(c.GetString("id"))
	_, err := admin.UpdateStatus(id, types.StatusInvalid)
	if err != nil {
		c.adminError(action, "操作失败:"+err.Error())
		return
	}

	c.Redirect("/manage/admin/list", 302)
}

func (c *ManageController) AdminUnblock() {
	action := "admin/unblock"
	id, _ := tools.Str2Int64(c.GetString("id"))
	_, err := admin.UpdateStatus(id, types.StatusValid)
	if err != nil {
		c.adminError(action, "操作失败:"+err.Error())
		return
	}

	c.Redirect("/manage/admin/list", 302)
}

func (c *ManageController) adminError(action, message string) {
	c.commonError(action, "/manage/admin/list", message)
}

func (c *ManageController) SmsVerifyCode() {
	c.Data["Action"] = "sms_verify_code"
	condCntr := map[string]interface{}{}
	mobile := c.GetString("mobile")
	if len(mobile) > 0 {
		condCntr["mobile"] = mobile
	}
	c.Data["mobile"] = mobile

	splitSep := " - "
	// s申请时间范围
	expires := c.GetString("expires")
	if len(expires) > 16 {
		//为了查询出验证码的过期时间
		exp, _, _ := service.SmsVerifyCodeList(map[string]interface{}{}, 1, 1)

		tr := strings.Split(expires, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["expires_start_time"] = (timeStart - int64(exp[0].Expires))
				condCntr["expires_end_time"] = (timeEnd - int64(exp[0].Expires))
			}
		}
	}
	c.Data["expires"] = expires
	ctime := c.GetString("ctime")
	if len(ctime) > 16 {
		tr := strings.Split(ctime, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["ctime_start_time"] = timeStart
				condCntr["ctime_end_time"] = timeEnd
			}
		}
	}
	c.Data["ctime"] = ctime
	utime := c.GetString("utime")
	if len(utime) > 16 {
		tr := strings.Split(utime, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["utime_start_time"] = timeStart
				condCntr["utime_end_time"] = timeEnd
			}
		}
	}
	c.Data["utime"] = utime
	Ip := c.GetString("Ip")
	if len(Ip) > 0 {
		condCntr["ip"] = Ip
	}
	c.Data["Ip"] = Ip
	status, err := c.GetInt("status")
	if err == nil && status >= 0 {
		condCntr["status"] = status
		c.Data["status"] = status
	} else {
		c.Data["status"] = -1
	}

	authCodeType, err := c.GetInt("authCodeType")
	if err == nil && authCodeType > 0 {
		condCntr["authcode_type"] = authCodeType
		c.Data["authCodeType"] = authCodeType
	} else {
		c.Data["authCodeType"] = -1
	}

	// 分页逻辑
	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	list, _, _ := service.SmsVerifyCodeList(condCntr, page, pagesize)
	count, _ := service.SmsVerifyCodeCount(condCntr)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["statusList"] = types.SmsVerifyCodeStatusMap
	c.Data["authCodeTypeList"] = types.AuthCodeTypeMap()
	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "sms_verify_code/list.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "sms_verify_code/list_scripts.html"
}

func (c *ManageController) OrderChange() {
	gotoURL := "/order/list"

	orderID, err := c.GetInt64("order_id")
	if err != nil {
		c.commonError("", gotoURL, "订单号有误, err:"+err.Error())
		return
	}

	orderData, err := models.GetOrder(orderID)
	if err != nil {
		c.commonError("", gotoURL, "订单数据有误, err:"+err.Error())
		return
	}

	orderDataBson, _ := json.MarshalIndent(orderData, "", "    ")
	c.Data["OrderDataJson"] = string(orderDataBson)
	c.Data["OrderData"] = orderData
	c.Data["OrderStatusMap"] = types.AllOrderStatusMap()

	c.Layout = "layout.html"
	c.TplName = "manage/order_change.html"
}

func (c *ManageController) OrderChangeSave() {
	gotoURL := "/order/list"

	orderID, err := c.GetInt64("order_id")
	if err != nil {
		c.commonError("", gotoURL, "订单号有误, err:"+err.Error())
		return
	}

	orderData, err := models.GetOrder(orderID)
	if err != nil {
		c.commonError("", gotoURL, "订单数据有误, err:"+err.Error())
		return
	}

	orderStatusMap := types.AllOrderStatusMap()
	newStatus, errC := c.GetInt("check_status")
	checkStatus := types.LoanStatus(newStatus)
	if errC != nil || orderStatusMap[checkStatus] == "" {
		c.commonError("", gotoURL, "提交数据有误")
		return
	}

	if checkStatus != orderData.CheckStatus {
		originData := orderData
		orderData.CheckStatus = checkStatus
		orderData.Utime = tools.GetUnixMillis()
		orderData.Update("check_status", "utime")

		models.OpLogWrite(c.AdminUid, orderData.Id, models.OpCodeOrderUpdate, orderData.TableName(), originData, orderData)
	}

	c.Redirect(gotoURL, 302)
}
