package clogs

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"encoding/json"
)

func InitLog(dir, name string) {
	// 加入日志,先简单一点配置
	// 日志级别只在dev环境为Trace,其他环境均为Warning
	logs.EnableFuncCallDepth(true)
	var logsConfig = make(map[string]interface{})
	logname := dir + "/" + name + ".log"
	logsConfig["filename"] = logname
	logsConfig["rotate"] = false

	var runmode = beego.AppConfig.String("runmode")
	if "dev" != runmode {
		logsConfig["level"] = logs.LevelWarning
		logsConfig["separate"] = []string{
			"emergency", "alert", "critical",
			"error", "warning",
		}
	} else {
		logsConfig["separate"] = []string{
			"emergency", "alert", "critical",
			"error", "warning", "notice",
			"info", "debug",
		}
	}
	config, _ := json.Marshal(logsConfig)
	logs.SetLogger(logs.AdapterMultiFile, string(config))

	logs.Debug("runmode: ", runmode, ", config:", string(config))
}
