package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"encoding/json"
	//"os"

	"micro-loan/common/cerror"
	"micro-loan/common/tools"
)

type DotBaseController struct {
	beego.Controller

	// AES解密后的数据
	DecryptData string
	// request json
	RequestJSON map[string]interface{}
	// 有效token对应的用户账户
	AccountID   int64
	UIVersion   string
	VersionCode int
}

const (
	Dot1 string = "http://d1.toolkits.mobi/dot_common.php?ver=%s&uiver=%s&parae=%s"
	Dot2 string = "https://d2.toolkits.mobi/dot2?pkg=%s&ver=%s&uiver=%s"
)

func (c *DotBaseController) Prepare() {
	// 一期不对客户端数据做值类型校验,完全信任客户端 TODO
	// 经过的AES加解密和md5参数签名,理论是可以信任,除非客户端的人不靠谱 ^_*
	// 预处理

	//维护公告
	//c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceIsDown, "")
	//c.ServeJSON()
	//return

	data := c.GetString("data")
	if len(data) < 16 {
		logs.Warning("post data is empty.")
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 为了联调,先打出来
	logs.Debug(">>> origData:", data)

	//logs.Debug("ReqHeader:", c.Ctx.Request.Header)
	//// 需要配置文件配合 copyrequestbody = true
	//origByte := c.Ctx.Input.RequestBody
	//logs.Debug(">>> origByte:", string(origByte))
	//postData := c.Ctx.Request.PostForm.Get("data")
	//logs.Debug("postData:", postData)

	// beego 框架会自行urldecode,但客户端传过来又不会,为什么???
	// 客户端找到问题了,是urlencode的问题
	//decryptData, err := tools.AesDecryptUrlCode(data, tools.AesCBCKey, tools.AesCBCIV)
	decryptData, err := tools.AesDecryptCBC(data, tools.AesCBCKey, tools.AesCBCIV)
	if err != nil {
		logs.Warning("post data can NOT decrypt, data:", data, ", err:", err)
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// AES解密通过
	c.DecryptData = decryptData
	logs.Debug("decryptData:", decryptData)

	var reqJSON map[string]interface{}
	err = json.Unmarshal([]byte(decryptData), &reqJSON)
	if err != nil {
		logs.Warning("cat NOT json decode request data:", decryptData)
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidRequestData, "")
		c.ServeJSON()
		return
	}

	// json decode 通过
	c.RequestJSON = reqJSON
	logs.Debug("request json:", decryptData)

}
