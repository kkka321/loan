package cmtelecom

import (
	"encoding/json"
	"micro-loan/common/lib/sms/api"
	"micro-loan/common/lib/sms/areacode"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"net/http"
	"strings"

	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"

	"micro-loan/common/types"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

//
// Document address：
// https://docs.cmtelecom.com/bulk-sms/v1.0
// HTTP address：
// https://sgw01.cm.nl/gateway.ashx (if support TLS cryptographic protocol)
// http://sgw01.cm.nl/gateway.ashx
//
// ProductToken
// 687FDCA9-BCD7-4375-BA39-86AB245B18EA

// https://gw.cmtelecom.com/v1.0/message

const (
	// APISingleSend 单条发送url片段
	apiURL      = "https://gw.cmtelecom.com/v1.0/message"
	codeSuccess = "0"

	supportGroup = false
	supportTest  = false
)

var companyName string

// Sender 短信发送器
type Sender struct {
	Msg         string
	Mobile      string
	GroupMobile []string
	RelatedID   int64
}

var producttoken string

func init() {
	producttoken = beego.AppConfig.String("cmtelecom_producttoken")
}

type request struct {
	Messages struct {
		Authentication struct {
			Producttoken string `json:"producttoken"`
		} `json:"authentication"`
		Msg []ReqMsgBody `json:"msg"`
	} `json:"messages"`
}

// ReqMsgBody 请求信息Body
type ReqMsgBody struct {
	Body struct {
		Content string `json:"content"`
	} `json:"body"`
	From string     `json:"from"`
	To   []ToNumber `json:"to"`
}

// ToNumber 子结构体
type ToNumber struct {
	Number string `json:"number"`
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

	var req request
	req.Messages.Authentication.Producttoken = producttoken
	var reqMsgBody ReqMsgBody
	reqMsgBody.Body.Content = s.Msg
	reqMsgBody.To = []ToNumber{{"00" + areacode.PhoneWithServiceRegionCode(s.Mobile)}}
	if strings.HasPrefix(s.Mobile, "86") {
		reqMsgBody.From = "111"
	} else {
		reqMsgBody.From = beego.AppConfig.String("cmtelecom_company_name")
	}
	req.Messages.Msg = []ReqMsgBody{reqMsgBody}

	// 手机号码，格式(区号+手机号码)，例如：8615800000000，其中86为中国的区号
	bytesData, err := json.Marshal(req)
	if err != nil {
		logs.Error("[cmtelecom] json marshal error: ", err)
		return
	}

	reqHeader := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}
	logs.Warn(req)
	httpBody, httpCode, err := tools.SimpleHttpClient("POST", apiURL, reqHeader, string(bytesData), tools.DefaultHttpTimeout())

	logs.Warn(string(httpBody))
	monitor.IncrThirdpartyCount(models.ThirdpartySmsCmtelecom, httpCode)

	originalResp = make(map[string]interface{})
	originalResp["body"] = string(httpBody)
	originalResp["httpCode"] = httpCode
	originalResp["httpErr"] = err

	iMobile, _ := tools.Str2Int64(s.Mobile)
	realtedID := tools.ThreeElementExpression(s.RelatedID != 0, s.RelatedID, iMobile).(int64)
	responstType, fee := thirdparty.CalcFeeByApi(apiURL, req, originalResp)
	models.AddOneThirdpartyRecord(models.ThirdpartySmsCmtelecom, apiURL, realtedID, req, originalResp, responstType, fee, httpCode)
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
	return types.CmtelcomSmsID
}

// Delivery 返回 送达状态
// http://pushUrl?receiver=admin&pswd=12345&msgid=12345&reportTime=1012241002&mobile=13900210021&status=DELIVRD
func (s *Sender) Delivery(r *http.Request) (msgID string, deliveryStatus int, callbackContent interface{}, err error) {
	// TODO
	return
}

// APIResponse api返回结构
type APIResponse struct {
	Details   string `json:"details"`
	ErrorCode int64  `json:"errorCode"`
	Messages  []struct {
		MessageDetails   string `json:"messageDetails"`
		MessageErrorCode int64  `json:"messageErrorCode"`
		Parts            int64  `json:"parts"`
		Reference        string `json:"reference"`
		Status           string `json:"status"`
		To               string `json:"to"`
	} `json:"messages"`
}

// IsSuccess is a required func to check result
func (apiRes *APIResponse) IsSuccess() bool {
	if len(apiRes.Messages) > 0 {
		return apiRes.Messages[0].Status == "Accepted"
	}
	return false
}

// GetMsgID 返回第三方消息ID, 用于回调
func (apiRes *APIResponse) GetMsgID() string {
	if len(apiRes.Messages) > 0 {
		return apiRes.Messages[0].Reference
	}
	return ""
}
