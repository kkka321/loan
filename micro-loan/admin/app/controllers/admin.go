package controllers

import (
	"micro-loan/common/models"
	"micro-loan/common/pkg/admin"
	"micro-loan/common/pkg/rbac"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"fmt"
	"micro-loan/common/dao"
	"micro-loan/common/i18n"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego/utils/pagination"
)

type AdminController struct {
	BaseController
}

func (c *AdminController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "admin"
}

func (c *AdminController) Password() {
	c.Layout = "layout.html"
	c.TplName = "account/fix_password.html"

}
func (c *AdminController) FixPassword() {
	response := map[string]interface{}{}
	password := strings.Trim(c.GetString("password"), "\"")
	newpassword := strings.Trim(c.GetString("newpassword"), "\"")
	uid := c.AdminUid
	admin, _ := models.OneAdminByUid(uid)

	md5password := tools.PasswordEncrypt(password, admin.RegisterTime)
	newpassd := tools.PasswordEncrypt(newpassword, admin.RegisterTime)
	if admin.Password == md5password {
		//密码匹配修改密码
		admin.Password = newpassd
		id, err := models.Update(admin)
		if err != nil {
			response["message"] = "false"
			response["code"] = 406
		} else {
			adp, _ := models.OneAdminByUid(id)
			if adp.Password == newpassd {
				response["message"] = "ok"
				response["code"] = 200
			}
		}

	} else {
		response["message"] = "old password is error"
		response["code"] = 405
	}
	c.Data["json"] = response
	c.ServeJSON()

}

func (c *AdminController) OpLog() {

	condCntr := map[string]interface{}{}
	opTable := c.GetString("op_table")
	if len(opTable) > 0 {
		condCntr["opTable"] = opTable
	}
	c.Data["opTable"] = opTable
	opCode, _ := c.GetInt("op_code")
	if opCode > 0 {
		condCntr["opCode"] = opCode
	}
	c.Data["opCode"] = models.OpCodeEnum(opCode)

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
	}
	c.Data["id"] = id

	relatedId, _ := c.GetInt64("related_id")
	if relatedId > 0 {
		condCntr["relatedId"] = relatedId
	}
	c.Data["relatedId"] = relatedId

	opUID, err := c.GetInt64("op_uid")
	if err == nil && opUID > -1 {
		condCntr["opUid"] = opUID
		c.Data["opUid"] = opUID
	} else {
		c.Data["opUid"] = -1
	}

	splitSep := " - "

	// 查询日志时间范围
	timeRange := c.GetString("time_range")
	if len(timeRange) > 16 {
		expApplyTime := strings.Split(timeRange, splitSep)
		if len(expApplyTime) == 2 {
			timeStart := tools.GetDateParseBackend(expApplyTime[0]) * 1000
			timeEnd := tools.GetDateParseBackend(expApplyTime[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["start_time"] = timeStart
				condCntr["end_time"] = timeEnd
			}
		}
	}
	c.Data["TimeRange"] = timeRange

	month, _ := c.GetInt64("month", 0)
	if month > 0 {
		condCntr["month"] = month
	}
	c.Data["month"] = month

	// 分页逻辑
	if c.GetString("search") == "1" {
		page, _ := tools.Str2Int(c.GetString("p"))
		pagesize := 15

		list, count, _ := service.OpLoggerList(condCntr, page, pagesize)
		paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

		c.Data["paginator"] = paginator
		c.Data["List"] = list
	}

	c.Data["OpTableList"] = service.OpLoggerTableMap
	c.Data["OpCodeList"] = models.OpCodeList
	c.Data["monthMap"] = service.GetMonthMap()

	c.Layout = "layout.html"
	c.TplName = "op_logger/list.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "op_logger/table_view.html"

}

func (c *AdminController) OpLogView() {
	id, err := c.GetInt64("id")
	ctime, err := c.GetInt64("ctime")

	if err == nil {
		m := models.OpLogger{}
		month := tools.GetMonth(ctime)
		monthMap := service.GetMonthMap()
		tableName := m.OriTableName()
		if _, ok := monthMap[month]; ok {
			tableName = m.TableNameByMonth(month)
		}

		data, _ := service.GetOpLogger(tableName, id)

		data.Original = service.ConvertInt64tString(data.Original)
		data.Edited = service.ConvertInt64tString(data.Edited)
		c.Data["json"] = data
	} else {
		c.Data["json"] = nil
	}
	c.ServeJSON()
	return
}

func (c *AdminController) List() {
	c.Layout = "layout.html"
	c.TplName = "admin/list.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "admin/scripts.html"

	condCntr := map[string]interface{}{}
	list, _, _ := admin.LowPrivilegeList(condCntr, 1, 500)
	c.Data["List"] = list
}

func (c *AdminController) Edit() {
	if v, err := c.GetInt64("id"); err == nil {
		data, err := models.OneAdminByUid(v)
		if err != nil {
			c.Data["error"] = err.Error()
		}

		c.Data["RoleList"], _ = rbac.LowPrivilegeRoleList(map[string]interface{}{"status": types.StatusValid})
		c.Data["Data"] = data
	} else {
		c.Data["error"] = "非法请求,无用户ID参数"
	}

	c.TplName = "admin/edit.html"

	return
}

func (c *AdminController) Update() {
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

func (c *AdminController) Block() {
	id, _ := tools.Str2Int64(c.GetString("id"))
	_, err := admin.UpdateStatus(id, types.StatusInvalid)
	if err != nil {
		c.newCommonError("/admin/list", "操作失败:"+err.Error())
		return
	}

	c.Redirect("/admin/list", 302)
}

func (c *AdminController) Unblock() {
	id, _ := tools.Str2Int64(c.GetString("id"))
	_, err := admin.UpdateStatus(id, types.StatusValid)
	if err != nil {
		c.newCommonError("/admin/list", "操作失败:"+err.Error())
		return
	}

	c.Redirect("/admin/list", 302)
}

func (c *AdminController) Create() {
	c.Data["Action"] = "admin/create"
	c.Data["IsEdit"] = false
	c.Data["RoleList"], _ = rbac.LowPrivilegeRoleList(map[string]interface{}{})
	c.Layout = "layout.html"
	c.TplName = "admin/create.html"
}

func (c *AdminController) Save() {

	email := c.GetString("email")
	mobile := c.GetString("mobile")
	nickname := c.GetString("nickname")
	password := c.GetString("password")
	// TODO 检查是否有添加此角色的权限
	// 角色是否在可添加角色列表中
	roleID, _ := c.GetInt64("role_id")

	if len(email) < 8 || len(mobile) < 8 || len(nickname) < 3 || len(password) < 8 {
		c.newCommonError("/admin/list", "无效参数")
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
		c.newCommonError("/admin/list", "操作失败:"+err.Error())
		return
	}

	c.Redirect("/admin/list", 302)
}

func (c *AdminController) Export() {
	str := c.Ctx.Input.Param(":id")

	accountList := make([]int64, 0)
	maxId := int64(0)

	nowDate := tools.MDateMHSDate(tools.GetUnixMillis())
	endTime := nowDate + " 11:30"

	startDate := tools.MDateMHSDate(tools.NaturalDay(-1))

	start := tools.GetDateParseBackend(startDate) * 1000
	end := tools.GetTimeParse(endTime) * 1000

	if str == "89" {
		//"注册未下单"
		for {
			list, _ := dao.QueryRegisterNoOrderAccount(start, end, maxId)
			if len(list) == 0 {
				break
			}

			for _, v := range list {
				if v > maxId {
					maxId = v
				}

				accountList = append(accountList, v)
			}
		}

	} else if str == "90" {
		//下单未填写资料
		for {
			list, _ := dao.QueryRegisterOrderNoKtp(start, end, maxId)
			if len(list) == 0 {
				break
			}

			for _, v := range list {
				accountList = append(accountList, v)

				if v > maxId {
					maxId = v
				}
			}
		}
	}

	mobiles := make([]string, 0)
	for _, id := range accountList {
		account, err := models.OneAccountBaseByPkId(id)
		if err != nil {
			continue
		}

		mobiles = append(mobiles, account.Mobile)
	}

	fileName := fmt.Sprintf("%s_%s.xlsx", startDate, str)
	lang := c.LangUse
	xlsx := excelize.NewFile()
	xlsx.SetCellValue("Sheet1", "A1", i18n.T(lang, "mobile"))

	for i, d := range mobiles {
		xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(i+2), d)
	}
	c.Ctx.Output.Header("Accept-Ranges", "bytes")
	c.Ctx.Output.Header("Content-Type", "application/octet-stream")
	c.Ctx.Output.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Ctx.Output.Header("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
	c.Ctx.Output.Header("Pragma", "no-cache")
	c.Ctx.Output.Header("Expires", "0")
	xlsx.Write(c.Ctx.ResponseWriter)
}
