package controllers

import (
	"encoding/json"
	//"os"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/cerror"
	"micro-loan/common/tools"
)

// APIBaseEntrustController 勤为接口基类
type APIBaseEntrustController struct {
	beego.Controller
	// request json
	RequestJSON map[string]interface{}
}

func (c *APIBaseEntrustController) Prepare() {
	// 一期不对客户端数据做值类型校验,完全信任客户端 TODO
	// 经过的AES加解密和md5参数签名,理论是可以信任,除非客户端的人不靠谱 ^_*
	// 预处理

	////维护公告
	//c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceIsDown, "")
	//c.ServeJSON()
	//c.Abort("")
	//return
	data := c.GetString("data")
	if len(data) < 16 {
		logs.Warning("post data is empty.")
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		c.Abort("")
		return
	}

	// 为了联调,先打出来
	logs.Debug(">>> origData:", data)

	var reqJSON map[string]interface{}
	err := json.Unmarshal([]byte(data), &reqJSON)
	if err != nil {
		logs.Warning("cat NOT json decode request data:", data)
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidRequestData, "")
		c.ServeJSON()
		c.Abort("")
		return
	}

	logs.Debug("reqJSON:", reqJSON)

	// json decode 通过
	c.RequestJSON = reqJSON

	// 必要参数检查,只检查存在,没有判值
	requiredParameter := map[string]bool{
		"pname":        true,
		"noise":        true,
		"request_time": true,
		"signature":    true,
	}
	var requiredCheck int = 0
	for k, _ := range reqJSON {
		if requiredParameter[k] {
			requiredCheck++
		}
	}
	if len(requiredParameter) != requiredCheck {
		logs.Warning("request json lost required parameter, json:", data)
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		c.Abort("")
		return
	}

	// 参数签名检查
	originSignature := reqJSON["signature"]
	delete(reqJSON, "signature")
	pname := reqJSON["pname"].(string)
	checkSignature := tools.Signature(reqJSON, tools.GetEntrustSignatureSecret(pname))
	logs.Debug("originSignature:", originSignature, ", checkSignature:", checkSignature)
	if originSignature != checkSignature {
		logs.Warning("signature check has wrong, json:", data)
		c.Data["json"] = cerror.BuildApiResponse(cerror.SignatureVerifyFail, "")
		c.ServeJSON()
		c.Abort("")
		return
	}
}
