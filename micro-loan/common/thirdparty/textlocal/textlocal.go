// @see: https://api.textlocal.in/docs/sendsms

package textlocal

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/sms/api"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"net/http"
)

const (
	SenderDefault = "TXTLCL"
	ApiKey        = "3uZi5imGSP8-5kWZE8tWXcYaTuiPKMh3rtuawAe43n"
	ApiUrl        = "https://api.textlocal.in/send/"
	//ApiUrl = "http://localhost/post.php"
)

const (
	TestYes = true
	TestNo  = false

	supportGroup = false
	supportTest  = true
)

const resSuccess = "success"

type apiErrors struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type apiMessages struct {
	Id        string `json:"id"`
	Recipient int64  `json:"recipient"`
}

type ApiResponse struct {
	Errors   []apiErrors   `json:"errors"`
	Status   string        `json:"status"`
	Messages []apiMessages `json:"messages"`
}

// 实现 api.Response 接口

// IsSuccess is a required func to check result
func (apiRes *ApiResponse) IsSuccess() bool {
	return apiRes.Status == resSuccess
}

func (apiRes *ApiResponse) GetMsgID() string {
	if len(apiRes.Messages) > 0 {
		return apiRes.Messages[0].Id
	} else {
		return ""
	}
}

func sendSms(numbers []string, message string, relatedId int64, sender string, test bool) (apiResponse *ApiResponse, httpBody []byte, httpCode int, err error) {
	if len(numbers) <= 0 {
		err = fmt.Errorf("please set mobile number")
		return
	}
	if len(message) < 4 {
		err = fmt.Errorf("please send message")
		return
	}

	messageRaw := tools.RawUrlEncode(message)

	reqBody := map[string]string{
		"apikey":  tools.UrlEncode(ApiKey),
		"numbers": tools.UrlEncode(strings.Join(numbers, ",")),
		"sender":  sender,
		"message": messageRaw,
		"test":    fmt.Sprintf("%v", test),
	}

	var reqBox []string
	for k, v := range reqBody {
		reqBox = append(reqBox, fmt.Sprintf("%s=%s", k, v))
	}

	req := strings.Join(reqBox, "&")
	logs.Debug("req: %s", req)
	reqHeader := map[string]string{
		"User-Agent":   "curl/7.54.0",
		"Content-Type": "application/x-www-form-urlencoded",
	}
	httpBody, httpCode, err = tools.SimpleHttpClient("POST", ApiUrl, reqHeader, req, tools.DefaultHttpTimeout())

	monitor.IncrThirdpartyCount(models.ThirdpartyTextLocal, httpCode)

	logs.Debug("httpBody: %s", string(httpBody))
	var httpBodyMap map[string]interface{}
	json.Unmarshal(httpBody, &httpBodyMap)

	responstType, fee := thirdparty.CalcFeeByApi(ApiUrl, reqBody, httpBodyMap)
	models.AddOneThirdpartyRecord(models.ThirdpartyTextLocal, ApiUrl, relatedId, reqBody, httpBodyMap, responstType, fee, httpCode)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: 0,
		OrderId:       relatedId,
		ApiMd5:        tools.Md5(ApiUrl),
		Fee:           int64(fee),
		Result:        responstType,
	})

	if err != nil {
		logs.Error("[SendSms] has wrong, req: %s, httpCode: %d, httpBody: %s", req, httpCode, string(httpBody))
		return
	}

	apiRes := ApiResponse{}
	err = json.Unmarshal(httpBody, &apiRes)

	apiResponse = &apiRes

	return
}

// 实现发送相关接口

// Sender 短信发送器
type Sender struct {
	Msg         string
	Mobile      string
	GroupMobile []string
	RelatedID   int64
	IsTest      bool
}

// IsSupportGroup 是否支持群发
func (s *Sender) IsSupportGroup() bool {
	return supportGroup
}

// IsSupportTest 是否支持测试
func (s *Sender) IsSupportTest() bool {
	return supportTest
}

// 设置为测试模式,并不真的发送短信,但接口会返回真实值
func (s *Sender) SetTestFlag(flag bool) {
	s.IsTest = flag
}

// Send 发送信息
func (s *Sender) Send() (apiResponse api.Response, originalResp map[string]interface{}, err error) {
	var numbers []string
	if len(s.Mobile) > 0 {
		numbers = append(numbers, s.Mobile)
	}
	for _, mobile := range s.GroupMobile {
		numbers = append(numbers, mobile)
	}

	apiResponse, httpBody, httpCode, err := sendSms(numbers, s.Msg, s.RelatedID, SenderDefault, s.IsTest)

	originalResp = make(map[string]interface{})
	originalResp["body"] = string(httpBody)
	originalResp["httpCode"] = httpCode
	originalResp["httpErr"] = err

	return
}

// GetID 返回 ServiceID
func (s *Sender) GetID() types.SmsServiceID {
	return types.TextlocalID
}

// Delivery 返回 送达状态
func (s *Sender) Delivery(r *http.Request) (msgID string, deliveryStatus int, callbackContent interface{}, err error) {
	return
}
