package tongdun

import (
	"encoding/json"
	"fmt"
	"micro-loan/common/cerror"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"strings"

	"github.com/astaxie/beego"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"
)

/**
* 同盾 ，数据魔盒API
* 文档地址： http://wiki.a.mobimagic.com/confluence/pages/viewpage.action?pageId=6750895
* 实现内容如下：
* 1. 创建任务
* 2. callback处理
* 3. 主动查询任务
* 2018-05-15
* wudahai
 */

// IdentityCheckCreateTask 通用结构体
type IdentityCheckCreateTask struct {
	Code    int64  `json:"code"`
	TaskID  string `json:"task_id"`
	Message string `json:"message"`
	Data    struct {
		ChannelCode  string      `json:"channel_code"`
		ChannelType  string      `json:"channel_type"`
		ChannelSrc   string      `json:"channel_src"`
		ChannelAttr  string      `json:"channel_attr"`
		CreateTime   string      `json:"created_time"`
		IdentityCode string      `json:"identity_code"`
		RealName     string      `json:"real_name"`
		Mobile       string      `json:"user_mobile"`
		TaskData     interface{} `json:"task_data"`
	} `json:"data"`
}

type TaskData struct {
	ReturnInfo struct {
		IsMatch string `json:"is_match"`
	} `json:"return_info"`
}

// IdentityCheckNotify 身份检查异步通知
type IdentityCheckNotify struct {
	NotifyEvent    string `json:"notify_event"`
	NotifyTime     string `json:"notify_time"`
	PassbackParams string `json:"passback_params"`
	NotifyData     IdentityCheckCreateTask
}

// PassbackParams 同盾提供给我们的透传参数，需要什么给什么，目前我只给账号ID
type PassbackParams struct {
	AccountID int64  `json:"account_id"`
	Mobile    string `json:"mobile"`
}

type TelkomselData struct {
	AccountIndo struct {
		ActiveUntil      string `json:"active_until"`
		TelkomselPoin    string `json:"telkomsel_poin"`
		RemainingCredits string `json:"remaining_credits"`
	} `json:"account_info"`
	PurchaseDate struct {
		FirstPurchasedate string `json:"first_purchasedate"`
	} `json:"vouchers_history"`
}

type GojekData struct {
	AccountInfo struct {
		GojekPoin string `json:"go_points"`
	} `json:"my_account"`
}

const (
	//同盾频道code
	ChannelCodeKTP       string = "107001"
	ChannelCodeTelkomsel string = "102001"
	ChannelCodeXI        string = "102002"
	ChannelCodeIndosat   string = "102003"
	ChannelCodeGoJek     string = "104001"
	ChannelCodeLazada    string = "101001"
	ChannelCodeTokopedia string = "101002"
	ChannelCodeFacebook  string = "103001"
	ChannelCodeInstagram string = "103002"
	ChannelCodeLinkedin  string = "903004"
	// IDCheckCodeCreate 检查code -1:创建任务 0:命中 190:未命中
	IDCheckCodeCreate int64 = -1
	IDCheckCodeYes    int64 = 0
	IDCheckCodeNo     int64 = 190
	// SourceCreateTask 数据来源 创建任务时 -1
	SourceCreateTask int64 = -1
	// IsMatchCreateTask 是否匹配创建任务时默认值
	IsMatchCreateTask string = "C"
	IsMatchNo         string = "N"
	IsMatchYes        string = "Y"
	IsMatchAcquire    string = "A" // 获取验证码
	IsMatchVerify     string = "V" // 验证验证码
	// SourceQueryTask 数据来源 手动请求 1
	SourceQueryTask int64 = 1
	// SourceNotify 数据来源 异步通知 2
	SourceNotify int64 = 2
	//SourceReRun 重跑数据修复的
	SourceReRun int64 = 3
	// IDCheckChannelType 同盾身份检查渠道类型
	IDCheckChannelType string = "KTP"
	// IDYYSChannelType 同盾运营商渠道类型
	IDYYSChannelType string = "YYS"
	// IDGoJekChannelType 同盾Go-jek渠道类型
	IDGoJekChannelType string = "TRIP"
	// IDSocialChannelType 同盾社交渠道类型
	IDSocialChannelType string = "SOCIAL"
	// IDDSChannelType 同盾电商渠道类型
	IDDSChannelType string = "DS"
	// IDCheckChannelCode 同盾身份检查渠道CODE
	IDCheckChannelCode string = "107001"
	// 认证完成 ,等待爬取
	IDInputOk = 173
	// CreateTaskURL 创建任务API call URL
	CreateTaskURL string = "https://talosapi.shujumohe.com/octopus/task.unify.create/v3?partner_code=%s&partner_key=%s"
	// QueryTaskURL 查询任务
	QueryTaskURL string = "https://talosapi.shujumohe.com/octopus/task.unify.query/v3?partner_code=%s&partner_key=%s"
	// AcquireCodeURL  获取验证码地址
	AcquireCodeURL string = "https://talosapi.shujumohe.com/octopus/task.unify.acquire/v3?partner_code=%s&partner_key=%s"
	// RetryURL 重试地址
	RetryURL string = "https://talosapi.shujumohe.com/octopus/task.unify.retry/v3?partner_code=%s&partner_key=%s"
)

func prepare(apiName string, param map[string]interface{}, file map[string]interface{}) (string, string, map[string]string, error) {

	partnerCode := beego.AppConfig.String("partner_code")
	partnerKey := beego.AppConfig.String("partner_key")

	if partnerCode == "" || partnerKey == "" {
		logs.Error("[ tongdun ] 同盾接口必要参数缺失！！请检查配置文件")
	} else {
		logs.Debug("[ tongdun ] partnerCode, 注意区分正式与测试 :", partnerCode)
	}

	requestURL := fmt.Sprintf(apiName, partnerCode, partnerKey)

	var requestPostBody string
	var requestHeaders = make(map[string]string)

	//根据请求contentType ，构造不同的包体， x-www-form-urlencoded 与 json 两种常用方式
	requestPostBody, requestHeaders = tools.MakeReqHeadAndBody("x-www-form-urlencoded", param)

	return requestURL, requestPostBody, requestHeaders, nil
}

// Request 发送请求
func Request(relatedID int64, apiName string, param map[string]interface{}, file map[string]interface{}) ([]byte, IdentityCheckCreateTask, error) {
	var original []byte
	resData := IdentityCheckCreateTask{}
	reqURL, reqBody, reqHeaders, err := prepare(apiName, param, file)
	logs.Debug("reqURL:", reqURL, ", reqBody:", reqBody, ", reqHeaders:", reqHeaders)
	// logs.Debug("reqURL:", reqURL, ", reqHeaders:", reqHeaders)

	if err != nil {
		return original, resData, err
	}

	httpBody, httpCode, err := tools.SimpleHttpClient("POST", reqURL, reqHeaders, reqBody, tools.DefaultHttpTimeout())

	monitor.IncrThirdpartyCount(models.ThirdpartyTongdun, httpCode)

	if err != nil {
		return original, resData, err
	}
	requestMap := map[string]interface{}{
		"query_string": param,
	}
	resMap := map[string]interface{}{}
	json.Unmarshal(httpBody, &resMap)
	//记录调用
	responstType, fee := thirdparty.CalcFeeByApi(reqURL, requestMap, resMap)
	models.AddOneThirdpartyRecord(models.ThirdpartyTongdun, reqURL, relatedID, requestMap, resMap, responstType, fee, httpCode)

	api := strings.Split(reqURL, "?")
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: relatedID,
		OrderId:       0,
		ApiMd5:        tools.Md5(api[0]),
		Fee:           int64(fee),
		Result:        responstType,
	})

	err = json.Unmarshal(httpBody, &resData)
	if err != nil {
		logs.Warning("API data has wrong:", httpBody)
		return original, resData, err
	}
	return httpBody, resData, nil
}

func IsSuccess(code string) (s bool) {
	s = false
	if "SUCCESS" == code || "FAILURE" == code {
		s = true
	}
	return
}

// CreateTask 创建任务
func CreateTask(accountID int64, channelType, channelCode, name, identityCode, mobile string) (code cerror.ErrCode, idCheckData IdentityCheckCreateTask, err error) {
	//透传参数
	passbackParams := PassbackParams{}
	passbackParams.AccountID = accountID
	passbackParamsJSON, _ := tools.JSONMarshal(passbackParams)
	params := map[string]interface{}{
		"channel_type":    channelType,
		"channel_code":    channelCode,
		"real_name":       name,
		"identity_code":   identityCode,
		"user_mobile":     mobile,
		"passback_params": passbackParamsJSON,
	}
	_, idCheckData, err = Request(accountID, CreateTaskURL, params, map[string]interface{}{})

	if err != nil || idCheckData.Code > 0 {
		logs.Debug("[tongdun.CreateTask] 同盾任务创建失败， ERROR ：", err, " CODE:", idCheckData.Code)
	}

	if idCheckData.Code == 0 {
		tongdunModel := models.AccountTongdun{}
		tongdunModel.TaskID = idCheckData.TaskID
		tongdunModel.AccountID = accountID
		tongdunModel.OcrRealName = idCheckData.Data.RealName
		tongdunModel.OcrIdentity = idCheckData.Data.IdentityCode
		tongdunModel.Mobile = idCheckData.Data.Mobile
		tongdunModel.CheckCode = IDCheckCodeCreate
		tongdunModel.Message = idCheckData.Message
		tongdunModel.IsMatch = IsMatchCreateTask
		tongdunModel.ChannelType = idCheckData.Data.ChannelType
		tongdunModel.ChannelCode = idCheckData.Data.ChannelCode
		tongdunModel.ChannelSrc = idCheckData.Data.ChannelSrc
		tongdunModel.ChannelAttr = idCheckData.Data.ChannelAttr
		tongdunModel.CreateTimeS = idCheckData.Data.CreateTime
		tongdunModel.NotifyTimeS = ""
		tongdunModel.CreateTime, _ = tools.GetTimeParseWithFormat(idCheckData.Data.CreateTime, "2006-01-02 15:04:05")
		tongdunModel.NotifyTime = 0
		tongdunModel.Source = SourceCreateTask
		models.InsertTongdun(tongdunModel)

		code = cerror.CodeSuccess
	} else {

		tongdunModel := models.AccountTongdun{}
		tongdunModel.TaskID = "errortask"
		tongdunModel.AccountID = accountID
		tongdunModel.OcrRealName = name
		tongdunModel.OcrIdentity = identityCode
		tongdunModel.Mobile = mobile
		tongdunModel.CheckCode = idCheckData.Code
		tongdunModel.Message = idCheckData.Message
		tongdunModel.IsMatch = ""
		tongdunModel.ChannelType = channelType
		tongdunModel.ChannelCode = channelCode
		tongdunModel.ChannelSrc = idCheckData.Data.ChannelSrc
		tongdunModel.ChannelAttr = idCheckData.Data.ChannelAttr
		tongdunModel.CreateTimeS = tools.GetDateMHS(tools.TimeNow())
		tongdunModel.NotifyTimeS = ""
		tongdunModel.CreateTime = tools.TimeNow()
		tongdunModel.NotifyTime = 0
		tongdunModel.Source = SourceCreateTask
		models.InsertTongdun(tongdunModel)

		code = cerror.TongdunIDCheckCreateTaskFail
		// 增加短信通知，邮件通知的最佳位置

		//....

	}
	return
}

// QueryTask 查询任务
func QueryTask(accountID int64, taskID string) (data IdentityCheckCreateTask, err error) {
	params := map[string]interface{}{
		"task_id": taskID,
	}
	_, data, err = Request(accountID, QueryTaskURL, params, map[string]interface{}{})
	return
}

func GetTongdunTaskData(accountID int64, channelCode string, channelType string) (models.AccountTongdun, error) {
	var atIns = models.AccountTongdun{}
	o := orm.NewOrm()
	o.Using(atIns.Using())
	err := o.QueryTable(atIns.TableName()).
		Filter("account_id", accountID).
		Filter("channel_code", channelCode).
		Filter("channel_type", channelType).
		Exclude("task_data", "").
		OrderBy("-id").
		One(&atIns)
	return atIns, err
}

func GetPurchaseDate(accountID int64) (models.AccountTongdun, TelkomselData, error) {
	data := TelkomselData{}
	tdData, err := GetTongdunTaskData(accountID, ChannelCodeTelkomsel, IDYYSChannelType)
	if err != nil {
		return tdData, data, err
	}

	err = json.Unmarshal([]byte(tdData.TaskData), &data)
	if err != nil {
		return tdData, data, err
	}

	return tdData, data, nil
}

func GetGojekData(accountID int64) (models.AccountTongdun, GojekData, error) {
	data := GojekData{}
	tdData, err := GetTongdunTaskData(accountID, ChannelCodeGoJek, IDGoJekChannelType)
	if err != nil {
		return tdData, data, err
	}

	err = json.Unmarshal([]byte(tdData.TaskData), &data)
	if err != nil {
		return tdData, data, err
	}

	return tdData, data, nil
}
