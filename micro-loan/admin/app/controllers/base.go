package controllers

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/cerror"
	"micro-loan/common/i18n"
	"micro-loan/common/lib/device"
	"micro-loan/common/lib/gaws"
	"micro-loan/common/models"
	"micro-loan/common/pkg/privilege"
	"micro-loan/common/pkg/rbac"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type BaseController struct {
	beego.Controller
	IsLogin          bool
	AdminUid         int64
	AdminNickname    string
	RoleID           int64
	RoleType         types.RoleTypeEnum
	RolePid          int64
	IsLeaderOrBeyond bool
	LangUse          string
}

func (this *BaseController) Prepare() {
	this.Data["ServiceRegion"] = tools.GetServiceRegion()

	isLogin := this.GetSession(types.SessAdminIsLogin)
	var adminUId int64
	if isLogin == nil {
		this.IsLogin = false
	} else {
		this.IsLogin = true

		adminUId, _ = this.GetSession(types.SessAdminUid).(int64)
		this.Data["LoginAdminUid"] = adminUId
		nickname := this.GetSession(types.SessAdminNickname).(string)
		this.Data["LoginNickname"] = nickname
		this.AdminUid = adminUId
		this.AdminNickname = nickname
		roleID := this.GetSession(types.SessAdminRoleID).(int64)
		this.RoleID = roleID
		// 兼容旧session方案, 下次上线可移除
		var rolePid int64
		var roleType types.RoleTypeEnum
		if this.GetSession(types.SessAdminRolePid) == nil {
			role, _ := rbac.GetOneRole(this.RoleID)
			rolePid = role.Pid
			roleType = role.Type
			this.SetSession(types.SessAdminRolePid, rolePid)
			this.SetSession(types.SessAdminRoleType, int(roleType))
		} else {
			rolePid = this.GetSession(types.SessAdminRolePid).(int64)
			roleType = types.RoleTypeEnum(this.GetSession(types.SessAdminRoleType).(int))
		}
		// 二次上线用以下替换
		//rolePid := this.GetSession(types.SessAdminRolePid).(int64)
		//roleType = types.RoleTypeEnum(this.GetSession(types.SessAdminRoleType).(int))

		this.RolePid = rolePid
		this.RoleType = roleType
	}

	this.Data["IsSuperAdmin"] = types.SuperAdminUID == adminUId

	this.IsLeaderOrBeyond = rbac.IsLeaderRoleAndBeyond(this.RoleID, this.RolePid)
	this.Data["IsLeaderOrBeyond"] = this.IsLeaderOrBeyond
	// 获取授权菜单和权限校验
	this.authAndGetMenu(types.SuperAdminUID == adminUId)

	this.Data["IsLogin"] = this.IsLogin
	this.Data["WebsiteTitle"] = "Micro-Loan Admin"
	this.Data["AdminStaticVersion"] = types.AdminStaticVersion

	// 多语言支持 {{{
	this.Data["LangSupportConf"] = i18n.LangSupportConf()

	hasCookie := false
	// 1. Check URL arguments.
	lang := this.Input().Get("lang")

	// 2. Get language information from cookies.
	if len(lang) == 0 {
		lang = this.Ctx.GetCookie("lang")
		hasCookie = true
	}

	// Check again in case someone modify by purpose.
	if !i18n.IsExist(lang) {
		lang = ""
		hasCookie = false
	}

	// 3. Get language information from 'Accept-Language'.
	if len(lang) == 0 {
		al := this.Ctx.Request.Header.Get("Accept-Language")
		if len(al) > 4 {
			al = al[:5] // Only compare first 5 letters.
			if i18n.IsExist(al) {
				lang = al
			}
		}
	}

	// 4. Default language is English.
	if len(lang) == 0 {
		lang = i18n.LangEnUS
	}

	this.LangUse = lang

	// Save language information in cookies.
	if !hasCookie {
		this.Ctx.SetCookie("lang", this.LangUse, 1<<31-1, "/")
	}
	this.Data["LangUse"] = this.LangUse
	// }}}
}

func (c *BaseController) commonError(action, gotoURL, message string) {
	c.Data["Action"] = action
	c.Data["goto_url"] = gotoURL
	c.Data["message"] = message

	c.Layout = "layout.html"
	c.TplName = "error.tpl"
}

func (c *BaseController) newCommonError(gotoURL, message string) {
	c.Data["goto_url"] = gotoURL
	c.Data["message"] = message

	c.Layout = "layout.html"
	c.TplName = "error.tpl"
}

func (c *BaseController) authAndGetMenu(isSuperAdmin bool) {
	// 获取权限列表
	//var privielgeMap = map[int64]string{1: true, 2: true, 3: true}
	cName, aName := c.GetControllerAndAction()
	pName := cName + "@" + aName

	operation, _ := models.GetOneOperation(pName)
	if operation.Id == 0 {
		operation.Name = pName
		operation.Id, _ = operation.Add()
	}

	if isSuperAdmin {
		c.Data["MenuList"], _ = rbac.SuperMenuTree(operation.Id)
		// 辅助菜单权限绑定
		if c.IsAjax() && c.GetString("token") == "menu_bind" {
			response := map[string]interface{}{}
			response["id"] = operation.Id
			response["name"] = operation.Name
			c.Data["json"] = response
			c.ServeJSON()
		}
	} else {

		pidMap, _ := rbac.GetRoleOperationIDMap(c.RoleID)
		if _, ok := pidMap[operation.Id]; !ok {
			c.Abort("403")
			//c.CustomAbort(403, "You have no access privileges, please apply it first")
		} else {
			// ajax 情况下, 不需要左边栏菜单
			if !c.IsAjax() {
				c.Data["MenuList"], _ = rbac.AuthMenuTree(operation.Id, pidMap)
			}
		}
	}

}

// 动态数据是否授权
// 如果角色为 指定type下超管用户, 则直接开放
// 否则, 该角色或者其直接下级角色必须包含该动态数据权限
func (c *BaseController) isGrantedData(dataType types.DataPrivilegeTypeEnum, dataID int64) {
	if !privilege.IsGrantedData(dataType, dataID, c.AdminUid, c.RoleID, c.RolePid) {
		//c.CustomAbort(403, "You have no this data privileges, please apply it first")
	}
}

func (c *BaseController) Finish() {
	strJson := string(c.Ctx.Input.RequestBody)
	fmt.Println(strJson)

	str := string(c.Ctx.Request.URL.String())
	fmt.Println(str)
}

func (c *BaseController) UploadResource(upFilename string, useMark types.ResourceUseMark) (resourceId int64, tmpFilename string, code cerror.ErrCode, err error) {
	code = cerror.CodeSuccess
	f, h, err := c.GetFile(upFilename)
	if err != nil {
		code = cerror.PermissionDenied
		logs.Error("Permission denied, can't upload file. for:", upFilename, "err:", err)
		return
	}
	defer f.Close()

	md5h := md5.New()
	io.Copy(md5h, f)
	ok_md5 := md5h.Sum([]byte(""))
	fileMd5 := fmt.Sprintf("%x", ok_md5)
	ext := tools.GetFileExt(h.Filename)
	tmpFilename = fmt.Sprintf("/tmp/%s.%s", fileMd5, ext)

	// 将上传文件保存到系统临时目录下
	c.SaveToFile(upFilename, tmpFilename)

	extension, mime, err := tools.DetectFileType(tmpFilename)
	if err != nil {
		code = cerror.FileTypeUnsupported
		logs.Error("Unrecognized file type: ", upFilename, ", err:", err)
		return
	}
	if !strings.Contains(mime, "image") {
		logs.Error("File type is not supported. file:", upFilename, ", mime:", mime, ", extension:", extension)
		code = cerror.FileTypeUnsupported
		err = fmt.Errorf("File type is not supported. file: %s", upFilename)
		return
	}

	_, hashName := tools.BuildHashName(fileMd5, extension)
	logs.Debug("hashName:", hashName)
	hashDir, hashName := tools.BuildHashName(fileMd5, extension)
	localHashDir := tools.LocalHashDir(hashDir)
	err = os.MkdirAll(localHashDir, 0755)

	if useMark == types.Use2Advertisement || useMark == types.Use2Banner ||
		useMark == types.Use2Pop || useMark == types.Use2Float || useMark == types.Use2AdPosition {
		_, err = gaws.AdsUpload(tmpFilename, hashName)
		if err != nil {
			logs.Error("Upload to ad fail. file:", upFilename, ", err:", err)
			code = cerror.UploadResourceFail
			return
		}
	} else {
		_, err = gaws.AwsUpload(tmpFilename, hashName)
		if err != nil {
			logs.Error("Upload to aws fail. file:", upFilename, ", err:", err)
			code = cerror.UploadResourceFail
			return
		}
	}

	// 写上传资源记录
	resourceId, _ = device.GenerateBizId(types.UploadResourceBiz)
	record := map[string]interface{}{
		"id":          resourceId,
		"op_uid":      c.AdminUid,
		"content_md5": fileMd5,
		"hash_name":   hashName,
		"extension":   extension,
		"use_mark":    useMark,
		"mime":        mime,
	}
	service.AddOneUploadResource(record)

	return
}
