package controllers

import (
	"micro-loan/common/cerror"
	"micro-loan/common/tools"

	"micro-loan/common/lib/gaws"

	"github.com/astaxie/beego/logs"
)

//"github.com/astaxie/beego/logs"

type LogController struct {
	ApiBaseController
}

func (c *LogController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()

	// 统一将 ip 加到 RequestJSON 中
	c.RequestJSON["ip"] = c.Ctx.Input.IP()
	c.RequestJSON["related_id"] = int64(0)
}

func (c *LogController) Boot() {
	clientData := c.RequestJSON["data"]
	go func(data interface{}) {
		logs.Debug(data)
		if clientData, ok := data.(string); ok {
			location, err := gaws.BootLogUpload(clientData)
			logs.Debug("[BootLog] location:", location)
			if err != nil {
				logs.Error("[BootLog] Boot log upload failed:", err)
			}
		} else {
			logs.Error("[BootLog] client data is invalid:", data)
		}
	}(clientData)

	retData := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, retData)
	c.ServeJSON()
}
