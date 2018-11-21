package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	//"os"
	"crypto/md5"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/cerror"
	"micro-loan/common/lib/device"
	"micro-loan/common/lib/gaws"
	"micro-loan/common/pkg/accesstoken"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type ApiBaseController struct {
	beego.Controller

	// AES解密后的数据
	DecryptData string
	// request json
	RequestJSON map[string]interface{}
	// 有效token对应的用户账户
	AccountID   int64
	UIVersion   string
	VersionCode int

	isTrace   bool
	beginTime int64
}

func (c *ApiBaseController) Prepare() {
	// 一期不对客户端数据做值类型校验,完全信任客户端 TODO
	// 经过的AES加解密和md5参数签名,理论是可以信任,除非客户端的人不靠谱 ^_*
	// 预处理

	////维护公告
	//c.Data["json"] = cerror.BuildApiResponse(cerror.ServiceIsDown, "")
	//c.ServeJSON()
	//c.Abort("")
	//return

	rv := tools.GenerateRandom(0, 100)
	rate, _ := beego.AppConfig.Int("monitor_api_trace_rate")
	if rv < rate {
		c.isTrace = true
		c.beginTime = tools.GetUnixMillis()
	}

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
		c.Abort("")
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
		c.Abort("")
		return
	}

	// json decode 通过
	c.RequestJSON = reqJSON
	logs.Debug("request json:", decryptData)

	// 必要参数检查,只检查存在,没有判值
	requiredParameter := map[string]bool{
		"noise":        true,
		"request_time": true,
		"access_token": true,
		"signature":    true,
	}
	var requiredCheck int = 0
	for k, _ := range reqJSON {
		if requiredParameter[k] {
			requiredCheck++
		}
	}
	if len(requiredParameter) != requiredCheck {
		logs.Warning("request json lost required parameter, json:", decryptData)
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		c.Abort("")
		return
	}

	// 参数签名检查
	originSignature := reqJSON["signature"]
	delete(reqJSON, "signature")
	checkSignature := tools.Signature(reqJSON, tools.GetSignatureSecret())
	logs.Debug("originSignature:", originSignature, ", checkSignature:", checkSignature)
	if originSignature != checkSignature {
		logs.Warning("signature check has wrong, json:", decryptData)
		c.Data["json"] = cerror.BuildApiResponse(cerror.SignatureVerifyFail, "")
		c.ServeJSON()
		c.Abort("")
		return
	}

	// 强制更新逻辑 VC = version_code
	if gpVCOrigin, ok := reqJSON["app_version_code"]; ok {
		gpVC, _ := tools.Str2Int(gpVCOrigin.(string))
		upVC, _ := config.ValidItemInt("app_force_upgrade_version_code")
		skipVersion := map[int]bool{
			24: true, // 强更 24 的包,有授权问题
		}
		// hard-code 12 是个死值,无法改动,因为app从12起才支持强制更新逻辑
		if skipVersion[gpVC] || (gpVC >= 12 && gpVC <= upVC) {
			jump_address := ""
			//判断下载渠道
			if channel, ok := reqJSON["is_google_service"]; ok {
				if channel.(string) == "1" { //google play
					res := config.ValidItemString("app_force_upgrade_gp_url")
					jump_address = res
				} else if channel.(string) == "0" { //非google play
					res := config.ValidItemString("app_force_upgrade_no_gp_url")
					jump_address = res
				}
			}

			data := map[string]interface{}{
				"server_time":     tools.GetUnixMillis(),
				"upgrade_message": config.ValidItemString("app_force_upgrade_message"),
				"jump_address":    jump_address,
			}
			c.Data["json"] = cerror.BuildApiResponse(cerror.AppForceUpgrade, data)
			c.ServeJSON()
			c.Abort("")
			return
		}
	}

	uri := c.Ctx.Request.RequestURI
	// 以下路由不需要持有 token
	notNeedTokenRoute := map[string]bool{
		"/api/v1/config/not_login":        true,
		"/api/v1/upload_client_info":      true,
		"/api/v1/request_login_auth_code": true,
		"/api/v2/request_login_auth_code": true,
		"/api/v1/request_voice_auth_code": true,
		"/api/v1/login":                   true,
		"/api/v1/register":                true,
		"/api/v1/login/sms":               true,
		"/api/v1/login/password":          true,
		"/api/v1/sms/verify":              true,
		"/api/v1/password/find":           true,

		"/api/loan_flow/v1/register":       true,
		"/api/loan_flow/v1/login":          true,
		"/api/loan_flow/v1/login/sms":      true,
		"/api/loan_flow/v1/login/password": true,
		"/t/:u":                           true,
		"/api/banner/v1/get":              true,
		"/api/activity/v1/get_popoversor": true,
		"/api/activity/v1/get_floating":   true,
		"/api/v1/log/boot":                true,
	}
	// 获得产品信息接口 {"/api/v1/product/info"} 允许未登录状态下调用， 当token 传空字符串时不去校验token
	if !notNeedTokenRoute[uri] &&
		!(uri == "/api/v1/product/info" && reqJSON["access_token"].(string) == "") {
		// 检查 token 有效性
		ok, accountId := accesstoken.IsValidAccessToken(types.PlatformAndroid, reqJSON["access_token"].(string))
		if !ok {
			logs.Warning("access_token is invalid, json:", decryptData)
			c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidAccessToken, "")
			c.ServeJSON()
			c.Abort("")
			return
		}

		c.AccountID = accountId
	}

	if v, ok := c.RequestJSON["ui_version"]; ok {
		c.UIVersion = v.(string)
	}

	if v, ok := c.RequestJSON["app_version_code"]; ok {
		versionCode, _ := tools.Str2Int(v.(string))
		c.VersionCode = versionCode
	}
}

/**
通用的资源上传方法,放在了 ApiBaseController,虽然不太好,但目前看没有更好的位置
# 入参:
  upFilename: post 的文件名
# 输出
*/

type UploadResourceResult struct {
	ResourceId  int64
	TmpFilename string
	Code        cerror.ErrCode
	Err         error
}

func (c *ApiBaseController) UploadResource(upFilename string, useMark types.ResourceUseMark) (resourceId int64, tmpFilename string, code cerror.ErrCode, err error) {
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
	//hashDir, hashName := tools.BuildHashName(fileMd5, extension)
	//localHashDir := tools.LocalHashDir(hashDir)
	//err = os.MkdirAll(localHashDir, 0755)

	_, err = gaws.AwsUpload(tmpFilename, hashName)
	if err != nil {
		logs.Error("Upload to aws fail. file:", upFilename, ", err:", err)
		code = cerror.UploadResourceFail
		return
	}
	// 写上传资源记录
	resourceId, _ = device.GenerateBizId(types.UploadResourceBiz)
	record := map[string]interface{}{
		"id":          resourceId,
		"op_uid":      c.AccountID,
		"content_md5": fileMd5,
		"hash_name":   hashName,
		"extension":   extension,
		"use_mark":    useMark,
		"mime":        mime,
	}
	service.AddOneUploadResource(record)

	return
}

func (c *ApiBaseController) Finish() {
	if c.isTrace {
		service.AddApiTraceData(c.beginTime, c.Ctx.Request.URL.String())
	}
}
