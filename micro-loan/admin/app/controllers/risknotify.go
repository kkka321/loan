package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/cerror"
	"micro-loan/common/service"
	"micro-loan/common/tools"
)

type RiskNotifyController struct {
	beego.Controller
}

// Notify 风控通知订单实时流
func (c *RiskNotifyController) Notify() {

	accessToken := c.GetString("access_token")
	reqTime, err1 := c.GetInt64("req_time")

	if err1 != nil {
		logs.Error("[ Notify ] c.GetInt64 happend error:", err1)
	}
	code := cerror.CodeSuccess
	//简单校验参数为空
	if accessToken == "" || reqTime == 0 {
		code = cerror.LostRequiredParameters
		c.Data["json"] = cerror.BuildAdminApiResponse(code, "")
		c.ServeJSON()
	}
	code, err := service.Notify(accessToken, reqTime)
	if err != nil {
		logs.Error("[ Notify ] risk notify happend error:", err)
	}
	c.Data["json"] = cerror.BuildAdminApiResponse(code, "")
	c.ServeJSON()
}

// QuotaConf 账号额度配置
func (c *RiskNotifyController) QuotaConf() {
	accountID, accountIDErr := c.GetInt64("account_id", 0)
	quota, quotaErr := c.GetInt64("quota", 0)
	quotaVisable, quotaVisableErr := c.GetInt64("quota_visable", 0)
	accountPeriod, periodErr := c.GetInt64("account_period", 0)
	isPhoneVerify, verifyErr := c.GetInt64("is_phone_verify")
	if accountIDErr != nil ||
		quotaErr != nil ||
		quotaVisableErr != nil ||
		periodErr != nil ||
		verifyErr != nil {
		logs.Error("[ Notify ] c.GetInt64 happend accountIDErr:", accountIDErr, "quotaVisableErr", quotaVisableErr, "quotaErr", quotaErr, "periodErr", periodErr, "verifyErr", verifyErr)
	}

	code := cerror.CodeSuccess
	//简单校验参数为空
	if accountID == 0 || quota == 0 || quotaVisable == 0 || accountPeriod == 0 {
		code = cerror.LostRequiredParameters
		c.Data["json"] = cerror.BuildAdminApiResponse(code, "")
		c.ServeJSON()
	}
	code, err := service.QuotaConf(accountID, quota, quotaVisable, accountPeriod, isPhoneVerify)
	if err != nil {
		logs.Error("[ Notify ] quota conf happend error:", err)
	}
	c.Data["json"] = cerror.BuildAdminApiResponse(code, "")
	c.ServeJSON()

}

// ThirdPartyQuery 第三方抓取接口
func (c *RiskNotifyController) ThirdPartyQuery() {

	mapData := make([]map[string]interface{}, 0)
	reqData := c.GetString("data")
	json.Unmarshal([]byte(reqData), &mapData)
	if len(mapData) > 500 && len(mapData) == 0 {
		code := cerror.RequestExceedsLimit
		c.Data["json"] = cerror.BuildAdminApiResponse(code, "")
		c.ServeJSON()
	}
	var accountID string
	var sourceFrom string
	var sourceCode string
	for k, v := range mapData {

		if _, ok := v["account_id"]; ok {
			accountID = v["account_id"].(string)
		}
		if _, ok := v["source_from"]; ok {
			sourceFrom = v["source_from"].(string)
		}
		if _, ok := v["source_code"]; ok {
			sourceCode = v["source_code"].(string)
		}
		if accountID == "" || sourceFrom == "" || sourceCode == "" {
			mapData[k]["result"] = ""
		}
		aID, _ := tools.Str2Int64(accountID)
		result, stime := service.QueryThirdParty(aID, sourceFrom, sourceCode)
		mapData[k]["result"] = result
		mapData[k]["service_time"] = stime

	}
	code := cerror.CodeSuccess
	c.Data["json"] = cerror.BuildAdminApiResponse(code, mapData)
	c.ServeJSON()

}

func (c *RiskNotifyController) RiskQuery() {
	mapData := make(map[string]interface{})
	section := c.GetString("type")
	accountId, _ := c.GetInt64("account_id")
	orderId, _ := c.GetInt64("order_id")
	value, err := service.QueryRiskValue(accountId, orderId, section)
	code := cerror.CodeSuccess
	if err == nil {
		mapData["status"] = 0
		mapData["msg"] = ""
	} else {
		mapData["status"] = 1
		mapData["msg"] = err.Error()
	}
	mapData["data"] = value
	c.Data["json"] = cerror.BuildAdminApiResponse(code, mapData)
	c.ServeJSON()
}
