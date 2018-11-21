package nexmo

// Nexmo 短信接口
// 文档地址: https://developer.nexmo.com/messaging/sms/overview
// 包括两个接口:
// 1. api.send 发送短信接口
// 2. api.receipt 接收通知接口

import (
	"encoding/json"
	"micro-loan/common/lib/sms/api"
	"micro-loan/common/lib/sms/areacode"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"net/http"
	"net/url"
	"strconv"

	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"

	"micro-loan/common/types"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

var (
	apiKey, apiSecret, apiHost, from string
)

const (
	// APISingleSend 单条发送url片段
	singleSendURL = "/sms/json"
	codeSuccess   = "0"

	supportGroup = false
	supportTest  = false
)

func init() {
	apiKey = beego.AppConfig.String("nexmo_api_key")
	apiSecret = beego.AppConfig.String("nexmo_api_secret")
	apiHost = beego.AppConfig.String("nexmo_api_host")
	from = beego.AppConfig.String("nexmo_api_from")

}

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

// Send 发送短信
func (s *Sender) Send() (apiResponse api.Response, originalResp map[string]interface{}, err error) {
	apiURL := apiHost + singleSendURL

	req := make(url.Values)
	//req := make(map[string]interface{})
	req["api_key"] = []string{apiKey}
	req["api_secret"] = []string{apiSecret}
	// 手机号码，格式(区号+手机号码)，例如：8615800000000，其中86为中国的区号
	req["to"] = []string{areacode.PhoneWithServiceRegionCode(s.Mobile)}
	req["from"] = []string{from}
	req["text"] = []string{s.Msg}

	if err != nil {
		logs.Error("[Nexmo] json marshal error: ", err)
		return
	}

	reqHeader := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	// logs.Debug(apiURL)
	httpBody, httpCode, err := tools.SimpleHttpClient("POST", apiURL, reqHeader, req.Encode(), tools.DefaultHttpTimeout())

	monitor.IncrThirdpartyCount(models.ThirdpartyNexmo, httpCode)

	originalResp = make(map[string]interface{})
	originalResp["body"] = string(httpBody)
	originalResp["httpCode"] = httpCode
	originalResp["httpErr"] = err

	apiResponse = &APIResponse{}

	if err != nil {
		logs.Error("[Nexmo] has wrong, req: %s, httpCode: %d, httpBody: %s", req, httpCode, string(httpBody))
		// 此时 apiResponse 为nil , TODO 更多的内部处理
		return
	}

	responstType, fee := thirdparty.CalcFeeByApi(apiURL, req, originalResp)
	models.AddOneThirdpartyRecord(models.ThirdpartyNexmo, apiURL, s.RelatedID, req, originalResp, responstType, fee, httpCode)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: 0,
		OrderId:       s.RelatedID,
		ApiMd5:        tools.Md5(apiURL),
		Fee:           int64(fee),
		Result:        responstType,
	})

	err = json.Unmarshal(httpBody, apiResponse)
	return
}

// GetID 返回 ServiceID
func (s *Sender) GetID() types.SmsServiceID {
	return types.NexoID
}

// Delivery 返回 送达状态
func (s *Sender) Delivery(r *http.Request) (msgID string, deliveryStatus int, callbackContent interface{}, err error) {
	return
}

// APIResponse api返回结构
type APIResponse struct {
	MsgCount string        `json:"message-count"`
	Msgs     []ResponseMsg `json:"messages"`
}

// ResponseMsg 单条msg 发送结果
type ResponseMsg struct {
	Mobile           string `json:"to"`
	MsgID            string `json:"message-id"`
	Status           string `json:"status"`
	RemainingBalance string `json:"remaining-balance"`
	Network          string `json:"network"`
	MessagePrice     string `json:"message-price"`
}

// IsSuccess is a required func to check result
func (apiRes *APIResponse) IsSuccess() bool {
	c, _ := strconv.Atoi(apiRes.MsgCount)
	if c == 1 {
		if apiRes.Msgs[0].Status == codeSuccess {
			return true
		}
	}
	return false
}

// GetMsgID 返回第三方消息ID, 用于回调
func (apiRes *APIResponse) GetMsgID() string {
	c, _ := strconv.Atoi(apiRes.MsgCount)
	if c == 1 {
		return apiRes.Msgs[0].MsgID
	}
	return ""
}
