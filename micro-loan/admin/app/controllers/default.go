package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "github.com/astaxie/beego/cache/redis"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/utils/captcha"
	//"github.com/astaxie/beego/logs"

	"micro-loan/common/models"
	"micro-loan/common/pkg/rbac"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type MainController struct {
	BaseController
}

var cpt *captcha.Captcha

func init() {
	cacheConf := fmt.Sprintf(`{"conn": "%s:%s"}`,
		beego.AppConfig.String("cache_redis_host"), beego.AppConfig.String("cache_redis_port"))
	store, err := cache.NewCache("redis", cacheConf)
	if err != nil {
		fmt.Printf("session cache init fail, err: %#v\n", err)
		os.Exit(11)
	}

	cpt = captcha.NewWithFilter("/captcha/", store) //一定要写在构造函数里面，要不然第一次打开页面有可能是X
}

func (c *MainController) Prepare() {
	c.Data["ServiceRegion"] = tools.GetServiceRegion()

	c.Data["Controller"] = "index"
}

func (c *MainController) Get() {
	//c.Data["Action"] = "get"
	//c.Data["Website"] = "beego.me"
	//c.Data["Email"] = "astaxie@gmail.com"
	//
	//c.Layout = "layout.html"
	//c.TplName = "debug.tpl"

	c.Redirect("/index", 302)
	return
}

func (c *MainController) Login() {
	_, ok := c.GetSession(types.SessAdminIsLogin).(bool)
	if ok {
		c.Redirect("/index", 302)
		return
	}

	// referer := c.Ctx.Request.Referer()
	// if strings.Contains(referer, "login") {
	// 	referer = ""
	// }
	// c.SetSession("referer", referer)
	c.TplName = "login.tpl"
}

func (c *MainController) HealthChecker() {
	res := map[string]interface{}{
		"code":        0,
		"message":     "ok",
		"version":     types.AdminVersion,
		"head_hash":   tools.GitRevParseHead(),
		"server_time": time.Now().Unix(),
	}

	c.Data["json"] = res
	c.ServeJSON()

	return
}

func (c *MainController) Ping() {
	//logs.Debug("Http Method:", c.Ctx.Request.Method)
	if c.Ctx.Request.Method == "POST" {
		res := map[string]interface{}{
			"code":        0,
			"message":     "ping",
			"version":     types.AdminVersion,
			"head_hash":   tools.GitRevParseHead(),
			"server_time": time.Now().Unix(),
		}

		c.Data["json"] = res
		c.ServeJSON()

		return
	}

	c.Data["Action"] = "ping"
	c.Data["ServerTime"] = fmt.Sprintf("%v", time.Now())
	c.Data["AdminVersion"] = types.AdminVersion
	c.Data["HeadHash"] = tools.GitRevParseHead()
	c.Layout = "layout.html"
	c.TplName = "ping.tpl"
}

// 内部系统,简单验证
func (c *MainController) LoginConfirm() {
	c.Data["Action"] = "login"
	c.Layout = "layout.html"
	c.TplName = "error.tpl"
	c.Data["goto_url"] = "/login"

	email := c.GetString("email")
	password := c.GetString("password")

	if !cpt.VerifyReq(c.Ctx.Request) {
		c.Data["message"] = "captcha has wrong"
	} else if email == "" || password == "" {
		c.Data["message"] = "用户名或密码为空"
	} else if !models.CheckLoginIsValid(email, password) {
		c.Data["message"] = "用户名或密码有误"
	} else {
		admin, _ := models.OneAdminByEmail(email)
		if admin.Status != 1 {
			c.Data["message"] = "此用户已被封禁"
			return
		}

		ticket.WorkerLogin(admin.Id, admin.RoleID, admin.LastLoginTime)
		c.SetSession(types.SessAdminIsLogin, true)
		c.SetSession(types.SessAdminUid, admin.Id)
		c.SetSession(types.SessAdminNickname, admin.Nickname)
		c.SetSession(types.SessAdminRoleID, admin.RoleID)
		role, _ := rbac.GetOneRole(admin.RoleID)
		c.SetSession(types.SessAdminRoleType, int(role.Type))
		c.SetSession(types.SessAdminRolePid, role.Pid)

		// 一期不做登陆历史. TODO
		ip := c.Ctx.Input.IP()
		models.AddLoginLog(admin.Id, ip)
		models.UpdateLastLoginTime(admin.Id)

		redirect := "/index"
		if val, ok := c.GetSession("going_url").(string); ok {
			if len(val) > 0 {
				redirect = val
				c.DelSession("going_url")
			}
		}
		c.Redirect(redirect, 302)
	}
}

func (c *MainController) Logout() {
	c.DelSession(types.SessAdminIsLogin)
	c.DelSession(types.SessAdminUid)
	c.DelSession(types.SessAdminNickname)
	c.DestroySession()

	c.Redirect("/login", 302)
}

func (c *MainController) Crypto() {
	c.Data["Action"] = "crypto"

	var jsonStr string
	var err error
	var isShow = false
	text := c.GetString("text")
	if len(text) > 16 {
		jsonStr, err = tools.AesDecryptUrlCode(text, tools.AesCBCKey, tools.AesCBCIV)
		if err != nil {
			jsonStr = fmt.Sprintf("解析失败. err:%v", err)
		} else {
			isShow = true
		}
		jsonObj := make(map[string]interface{})
		_ = json.Unmarshal([]byte(jsonStr), &jsonObj)
		bson, _ := json.MarshalIndent(jsonObj, "", "  ")
		jsonStr = string(bson)
	}

	c.Data["isShow"] = isShow
	c.Data["jsonStr"] = jsonStr

	c.Layout = "layout.html"
	c.TplName = "crypto.html"
}
