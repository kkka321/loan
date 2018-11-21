package sms253

import (
	"encoding/json"
	"fmt"
	"micro-loan/common/lib/sms/api"
	"micro-loan/common/lib/sms/areacode"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"net/http"

	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

var (
	//http://intapi.253.com
	account, password, apiHost string
)

const (
	// APISingleSend 单条发送url片段
	singleSendURL = "/send/json"
	codeSuccess   = "0"

	supportGroup = false
	supportTest  = false
)

// Sender 短信发送器
type Sender struct {
	Msg         string
	Mobile      string
	GroupMobile []string
	RelatedID   int64
}

// IsSupportGroup 是否支持群发
func (s *Sender) IsSupportGroup() bool {
	return supportTest
}

// IsSupportTest 是否支持测试
func (s *Sender) IsSupportTest() bool {
	return supportGroup
}

// Send 发送信息
func (s *Sender) Send() (apiResponse api.Response, originalResp map[string]interface{}, err error) {
	// 必须初始化
	apiResponse = &APIResponse{}

	apiURL := apiHost + singleSendURL

	req := make(map[string]interface{})
	req["account"] = account
	req["password"] = password
	// 手机号码，格式(区号+手机号码)，例如：8615800000000，其中86为中国的区号
	req["mobile"] = areacode.PhoneWithServiceRegionCode(s.Mobile)
	req["msg"] = s.Msg
	bytesData, err := json.Marshal(req)
	if err != nil {
		logs.Error("[sms253] json marshal error: ", err)
		return
	}

	reqHeader := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}
	httpBody, httpCode, err := tools.SimpleHttpClient("POST", apiURL, reqHeader, string(bytesData), tools.DefaultHttpTimeout())

	monitor.IncrThirdpartyCount(models.ThirdpartySms253, httpCode)

	originalResp = make(map[string]interface{})
	originalResp["body"] = string(httpBody)
	originalResp["httpCode"] = httpCode
	originalResp["httpErr"] = err

	iMobile, _ := tools.Str2Int64(s.Mobile)
	realtedID := tools.ThreeElementExpression(s.RelatedID != 0, s.RelatedID, iMobile).(int64)
	responstType, fee := thirdparty.CalcFeeByApi(apiURL, req, originalResp)
	models.AddOneThirdpartyRecord(models.ThirdpartySms253, apiURL, realtedID, req, originalResp, responstType, fee, httpCode)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: 0,
		OrderId:       realtedID,
		ApiMd5:        tools.Md5(apiURL),
		Fee:           int64(fee),
		Result:        responstType,
		MessageFlag:   true,
	})

	if err != nil {
		logs.Error("[SendSms] has wrong, req: %s, httpCode: %d, httpBody: %s", req, httpCode, string(httpBody))
		// 此时 apiResponse 为nil , TODO 更多的内部处理
		return
	}

	err = json.Unmarshal(httpBody, apiResponse)
	return
}

// GetID 返回 ServiceID
func (s *Sender) GetID() types.SmsServiceID {
	return types.Sms253ID
}

// Delivery 返回 送达状态
// http://pushUrl?receiver=admin&pswd=12345&msgid=12345&reportTime=1012241002&mobile=13900210021&status=DELIVRD
func (s *Sender) Delivery(r *http.Request) (msgID string, deliveryStatus int, callbackContent interface{}, err error) {
	r.ParseForm()
	callbackContent = r.Form
	msgID, deliveryStatus, err = handleDeliveryParam(r)

	responstType, fee := thirdparty.CalcFeeByApi(r.URL.String(), callbackContent, "")
	models.AddOneThirdpartyRecord(models.ThirdpartySms253, r.URL.String(), 0, callbackContent, "", responstType, fee, 200)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: 0,
		OrderId:       0,
		ApiMd5:        tools.Md5(r.URL.String()),
		Fee:           int64(fee),
		Result:        responstType,
	})
	return
}

func handleDeliveryParam(r *http.Request) (msgID string, deliveryStatus int, err error) {
	deliveryStatus = types.DeliveryUnknown
	if len(r.Form["receiver"]) <= 0 {
		err = fmt.Errorf("[Sms253 deliver] Miss Required parameter %s, r.Form: %v", "receiver", r.Form)
		return
	}
	receiver := r.Form["receiver"][0]
	if receiver != beego.AppConfig.String("sms253_delivery_receiver") {
		err = fmt.Errorf("[Sms253 deliver] Error parameter %s, r.Form: %v", "receiver", r.Form)
		return
	}

	if len(r.Form["pswd"]) <= 0 {
		err = fmt.Errorf("[Sms253 deliver] Miss Required parameter %s, r.Form: %v", "pswd", r.Form)
		return
	}
	pw := r.Form["pswd"][0]
	if pw != beego.AppConfig.String("sms253_delivery_pw") {
		err = fmt.Errorf("[Sms253 deliver] Error parameter %s, r.Form: %v", "pswd", r.Form)
		return
	}

	if len(r.Form["msgid"]) <= 0 {
		err = fmt.Errorf("[Sms253 deliver] Miss Required parameter %s, r.Form: %v", "msgid", r.Form)
		return
	}
	msgID = r.Form["msgid"][0]
	if len(msgID) <= 0 {
		err = fmt.Errorf("[Sms253 deliver] Error parameter %s, r.Form: %v", "msgid", r.Form)
	}
	if len(r.Form["status"]) <= 0 {
		err = fmt.Errorf("[Sms253 deliver] Miss Required parameter %s, r.Form: %v", "status", r.Form)
		return
	}
	status := r.Form["status"][0]

	switch status {
	case "DELIVRD":
		deliveryStatus = types.DeliverySuccess
		return
	case "REJECTD", "MBBLACK", "SM11", "SM12":
		deliveryStatus = types.DeliveryFailed
		return
	case "UNKNOWN":
		deliveryStatus = types.DeliveryUnknown
		return
	default:
		deliveryStatus = types.DeliveryUnknown
		return
	}
}

func init() {
	account = beego.AppConfig.String("sms253_account")
	password = beego.AppConfig.String("sms253_password")
	apiHost = beego.AppConfig.String("sms253_api_host")
}

// APIResponse api返回结构
type APIResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
	MsgID string `json:"msgid"`
}

// IsSuccess is a required func to check result
func (apiRes *APIResponse) IsSuccess() bool {
	return apiRes.Code == codeSuccess
}

// GetMsgID 返回第三方消息ID, 用于回调
func (apiRes *APIResponse) GetMsgID() string {
	return apiRes.MsgID
}

// TODO 获取短信发送状态

// 先用 sms253 , 1天手机号, cache,
// 先取key ->　key 不存在用新的这家，　key 存在上次用的哪家，　上次用的哪家，　这次就不用哪家
