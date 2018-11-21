package main

import (
	"micro-loan/common/lib/clogs"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	// 数据库初始化
	_ "micro-loan/api/app/routers"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"micro-loan/api/app/controllers"

	"fmt"
)

func init() {
	// 设置进程 title,以便运维平滑升级程序
	procTitle := fmt.Sprintf("%s:%s", beego.AppConfig.String("appname"), beego.AppConfig.String("httpport"))
	gspt.SetProcTitle(procTitle)
}

func main() {
	dir := beego.AppConfig.String("log_dir")
	port := beego.AppConfig.String("httpport")
	clogs.InitLog(dir, "api_"+port)

	logs.Info("start api.")

	// 注册失败控制器
	beego.ErrorController(&controllers.ErrorController{})

	beego.Run()
}
