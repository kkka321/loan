package boomsms

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/sms/api"
	"micro-loan/common/lib/sms/areacode"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// Sender 短信发送器
type Sender struct {
	Msg         string
	Mobile      string
	GroupMobile []string
	RelatedID   int64
}

// APIResponse api返回结构
type APIResponse struct {
	Status           int    `json:"status"`
	MessageId        string `json:"message_id"`
	To               string `json:"to"`
	RemainingBalance string `json:"remaining_balance"`
	MessageCount     string `json:"message_count"`
	ClientRef        string `json:"client_ref"`
	ErrorText        string `json:"error-text"`
}

func (apiRes *APIResponse) IsSuccess() bool {
	return apiRes.Status == 0
}

// GetMsgID 返回第三方消息ID, 用于回调
func (apiRes *APIResponse) GetMsgID() string {
	return apiRes.MessageId
}

var (
	token, from, apiHost string
)

func init() {
	tokenFile := beego.AppConfig.String("boomsms_api_token")
	apiHost = beego.AppConfig.String("boomsms_api_host")
	from = beego.AppConfig.String("boomsms_api_from")

	f, err := os.Open(tokenFile)
	if err != nil {
		return
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	token, _ = rd.ReadString('\n')
}

// Send 发送信息
func (s *Sender) Send() (apiResponse api.Response, originalResp map[string]interface{}, err error) {
	// 必须初始化
	apiResponse = &APIResponse{}

	req := make(map[string]interface{})

	req["to"] = areacode.PhoneWithServiceRegionCode(s.Mobile)
	req["text"] = s.Msg
	req["from"] = from
	bytesData, err := json.Marshal(req)

	reqHeader := map[string]string{
		"Accept":        "application/json;charset=UTF-8",
		"Content-Type":  "application/json;charset=UTF-8",
		"Authorization": "Bearer " + token,
	}

	reqBody := string(bytesData)

	logs.Warn("[Send] url:%s, header:%s, body:%s", apiHost, reqHeader, reqBody)

	httpBody, httpCode, err := tools.SimpleHttpClient("POST", apiHost, reqHeader, reqBody, tools.DefaultHttpTimeout())

	monitor.IncrThirdpartyCount(models.ThirdpartyBoomsms, httpCode)

	if err != nil {
		logs.Error("[Send] send error, err:%s, header:%d, body:%s", err, reqHeader, reqBody)
		return
	}

	originalResp = make(map[string]interface{})
	originalResp["body"] = string(httpBody)
	originalResp["httpCode"] = httpCode
	originalResp["httpErr"] = err

	responstType, fee := thirdparty.CalcFeeByApi(apiHost, req, string(httpBody))
	models.AddOneThirdpartyRecord(models.ThirdpartyBoomsms, apiHost, s.RelatedID, req, string(httpBody), responstType, fee, httpCode)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: 0,
		OrderId:       s.RelatedID,
		ApiMd5:        tools.Md5(apiHost),
		Fee:           int64(fee),
		Result:        responstType,
	})
	err = json.Unmarshal(httpBody, &apiResponse)
	if err != nil {
		logs.Error("[Send] Unmarshal error, err:%s, httpBody:%s", err, string(httpBody))
		return
	}

	if !apiResponse.IsSuccess() {
		logs.Error("[Send] response status error, header:%s, body:%s, res:%s", reqHeader, reqBody, string(httpBody))
		return
	}

	return
}

// GetID 返回 ServiceID
func (s *Sender) GetID() types.SmsServiceID {
	return types.BoomSmsID
}

// Delivery 返回 送达状态
func (s *Sender) Delivery(r *http.Request) (msgID string, deliveryStatus int, callbackContent interface{}, err error) {
	return
}
