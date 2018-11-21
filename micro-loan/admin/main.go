package main

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	_ "micro-loan/admin/app/helper"
	_ "micro-loan/admin/app/routers"
	_ "micro-loan/common/lib/db/mysql"

	_ "github.com/astaxie/beego/session/redis"

	"micro-loan/common/lib/clogs"
	"micro-loan/common/types"
)

func init() {
	// 设置进程 title,以便运维平滑升级程序
	procTitle := fmt.Sprintf("%s:%s", beego.AppConfig.String("appname"), beego.AppConfig.String("httpport"))
	gspt.SetProcTitle(procTitle)
}

var FilterUser = func(ctx *context.Context) {
	notNeedLoginRoute := map[string]bool{
		"/health_checker":              true,
		"/login":                       true,
		"/login_confirm":               true,
		"/ping":                        true,
		"/debug":                       true,
		"/risknotify/notify":           true,
		"/risknotify/quota_conf":       true,
		"/risknotify/thirdparty_query": true,
		"/risknotify/risk_query":       true,
	}
	_, ok := ctx.Input.Session(types.SessAdminIsLogin).(bool)
	if !ok &&
		!notNeedLoginRoute[ctx.Request.RequestURI] &&
		!strings.Contains(ctx.Request.RequestURI, "/debug/pprof") {
		ctx.Input.CruSession.Set("going_url", ctx.Request.RequestURI)
		ctx.Redirect(302, "/login")
		// Usually put return after redirect.
		return
	}
}

func main() {
	dir := beego.AppConfig.String("log_dir")
	port := beego.AppConfig.String("httpport")
	clogs.InitLog(dir, "admin_"+port)

	logs.Info("start admin.")

	// 统一登陆态检查
	beego.InsertFilter("/*", beego.BeforeRouter, FilterUser)

	beego.Run()
}
