package nxtele

import (
	"encoding/json"
	"errors"
	"fmt"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

var (
	username, nxteleHost, ccode, showMobile                         string
	hash, yesterday, today, tomorrow, infoReview, overdue, voiceUrl string
	nxcloudHost, appkey, secretkey                                  string
)

const (
	// 发送呼叫API
	nxteleCallApi = "/Api/Index/index"
	// 查询呼叫状态API
	nxteleCallStatusApi = "/Api/Index/result"
	// 语音验证码API
	nxcloudVoiceAuthCodeApi = "/api/voiceSms/versend"
)

// 发送呼叫API返回结构
type NxteleCallResponse struct {
	Result int    `json:"result"`
	Msg    string `json:"msg"`
	SID    int64  `json:"sid"` // nxtele系统中的订单号
}

func init() {
	username = beego.AppConfig.String("nxtele_user")
	nxteleHost = beego.AppConfig.String("nxtele_api_host")
	ccode = beego.AppConfig.String("nxtele_country_code")

	nxcloudHost = beego.AppConfig.String("nxcloud_voice_authcode_host")
}

// nxtele 系统中, 成功: result=1; 失败: result=0
func (nxteleResp *NxteleCallResponse) IsSuccess() int {
	return nxteleResp.Result
}

// GetSID 返回第三方消息中的订单号SID
func (nxteleResp *NxteleCallResponse) GetSID() int64 {
	return nxteleResp.SID
}

func getSystemConfig() {

	showMobile = config.ValidItemString("nxtele_show_number")
	if len(showMobile) <= 0 {
		showMobile = beego.AppConfig.String("nxtele_show_number")
	}

	hash = config.ValidItemString("nxtele_hash")
	yesterday = config.ValidItemString("nxtele_yesterday")              // 昨天到期，逾期一天
	today = config.ValidItemString("nxtele_today")                      // 今天到期
	tomorrow = config.ValidItemString("nxtele_tomorrow")                // 明天到期
	overdue = config.ValidItemString("nxtele_overdue_call_record_file") // 逾期案件的自动呼叫录音文件

	infoReview = config.ValidItemString("inforeview_call_record_file") // inforeview 自动外呼语音文件地址

	appkey = config.ValidItemString("nxcloud_appkey")       // appkey
	secretkey = config.ValidItemString("nxcloud_secretkey") // secretkey
}

// 手机号格式化
func MobileFormat(mobile string) string {
	if strings.HasPrefix(mobile, "08") {
		mobile = strings.Replace(mobile, "08", "628", 1)
	}
	mobile = strings.Replace(mobile, ",08", ",628", -1)

	return mobile
}

// 发送语音提醒
func Send(voiceType types.VoiceType, mobile string) (nxteleResp *NxteleCallResponse, err error) {
	getSystemConfig()

	nxteleResp = &NxteleCallResponse{}
	nxteleCallUrl := nxteleHost + nxteleCallApi

	switch voiceType {
	case types.VoiceTypeYesterday:
		voiceUrl = yesterday
	case types.VoiceTypeToday:
		voiceUrl = today
	case types.VoiceTypeTomorrow:
		voiceUrl = tomorrow
	case types.VoiceTypeInfoReview:
		voiceUrl = infoReview
	case types.VoiceTypeOverdue:
		voiceUrl = overdue
	}

	if !mustAllParamsComplete() {
		logs.Error("[nxtele] Send auto voice call request failed, request paramter incomplete!!")
		err = errors.New("[nxtele] Request paramter incomplete!!")
		return
	}

	mobile = MobileFormat(mobile)

	reqBody := map[string]string{
		"user":       username,
		"hash":       hash,
		"showmobile": showMobile,
		"mobile":     mobile,
		"ccode":      ccode,
		"url":        voiceUrl,
	}

	var reqBox []string
	for k, v := range reqBody {
		reqBox = append(reqBox, fmt.Sprintf("%s=%s", k, v))
	}

	reqParma := strings.Join(reqBox, "&")
	logs.Info("[nxtele] Auto voice call request url:", nxteleCallUrl, ", reqParma:", reqParma)
	reqHeaders := map[string]string{
		"User-Agent":   "curl/7.54.0",
		"Content-Type": "application/x-www-form-urlencoded",
	}

	body, _, err := tools.SimpleHttpClient("POST", nxteleCallUrl, reqHeaders, reqParma, tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[nxtele] Auto voice call request has wrong. nxteleCallUrl:", nxteleCallUrl, ", err:", err)
		return
	}

	err = json.Unmarshal(body, &nxteleResp)
	if err != nil {
		logs.Error("[nxtele] Send auto voice call request, parse body failed")
		return
	}

	return
}

func mustAllParamsComplete() (yes bool) {
	if len(username) > 0 && len(hash) > 0 && len(showMobile) > 0 && len(ccode) > 0 && len(voiceUrl) > 0 {
		yes = true
	}
	return
}

// 查询呼叫状态数据
type NxteleCallStatus struct {
	Phone    string `json:"phone"`
	Duration string `json:"duration"` // 通话时长，0表示未接通
	Fee      string `json:"fee"`      // 所需费用
}

// 查询呼叫状态API返回结构
type NxteleCallStatusResp struct {
	Result int    `json:"result"`
	Msg    string `json:"msg"`
	SID    string `json:"sid"` // nxtele系统中的订单号
	Data   []NxteleCallStatus
}

func mustAccountComplete() (yes bool) {
	if len(username) > 0 && len(hash) > 0 {
		yes = true
	}
	return
}

// 查询呼叫状态
func GetSidStatus(sid int64) (nxteleStatusResp *NxteleCallStatusResp, err error) {

	nxteleStatusResp = &NxteleCallStatusResp{}

	if sid <= 0 || !mustAccountComplete() {
		logs.Error("[nxtele] Get voice call status request failed, request paramter error!!")
		return
	}

	nxteleCallStatusUrl := nxteleHost + nxteleCallStatusApi

	nxteleCallStatusUrl = fmt.Sprintf("%s?user=%s&hash=%s&sid=%d",
		nxteleCallStatusUrl, username, hash, sid)
	logs.Info("[nxtele] Get voice call status request url:", nxteleCallStatusUrl, ", sid:", sid)
	reqHeaders := map[string]string{}

	body, _, err := tools.SimpleHttpClient("GET", nxteleCallStatusUrl, reqHeaders, "", tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[nxtele] Get voice call status request has wrong. nxteleCallUrl:", nxteleCallStatusUrl, ", err:", err)
		return
	}

	bodyStr := string(body)
	if strings.Contains(bodyStr, "result") {
		err = json.Unmarshal(body, &nxteleStatusResp)
		if err != nil {
			logs.Error("[nxtele] Get voice call status request, parse body failed")
			return
		}
	} else {
		var nxteleCallStatus []NxteleCallStatus
		err = json.Unmarshal(body, &nxteleCallStatus)
		if err != nil {
			logs.Error("[nxtele] Get voice call status request, parse body failed")
			return
		}

		nxteleStatusResp.Data = nxteleCallStatus
		nxteleStatusResp.SID = strconv.FormatInt(sid, 10)
		nxteleStatusResp.Result = types.VoiceCallSuccess
		nxteleStatusResp.Msg = "Success"
	}

	return
}

// 语音验证码
// 发送语音验证码的返回
type VoiceAuthCodeResponse struct {
	Result    string `json:"result"`
	MessageId string `json:"messageid"`
	Code      string `json:"code"`
}

// 服务区域语言码
func GetLanguageCode() (lang string) {
	serviceRegion := tools.GetServiceRegion()

	switch serviceRegion {
	case tools.ServiceRegionIndonesia:
		lang = "id"
	}

	return
}

func SendVoice(relatedId int64, phoneNumber, content string) (VoiceAuthCodeResp *VoiceAuthCodeResponse, err error) {
	getSystemConfig()

	VoiceAuthCodeResp = &VoiceAuthCodeResponse{}
	voiceAuthCodeCallUrl := nxcloudHost + nxcloudVoiceAuthCodeApi

	reqBody := map[string]string{
		"appkey":       appkey,
		"secretkey":    secretkey,
		"phone":        phoneNumber,
		"country_code": ccode,
		"show_phone":   showMobile,
		"content":      content,
		"lang":         GetLanguageCode(),
	}

	var reqBox []string
	for k, v := range reqBody {
		reqBox = append(reqBox, fmt.Sprintf("%s=%s", k, v))
	}

	reqParma := strings.Join(reqBox, "&")
	logs.Info("[SendVoice] Voice authcode call request url:", voiceAuthCodeCallUrl, ", reqParma:", reqParma)
	reqHeaders := map[string]string{
		"User-Agent":   "curl/7.54.0",
		"Content-Type": "application/x-www-form-urlencoded",
	}

	body, _, err := tools.SimpleHttpClient("POST", voiceAuthCodeCallUrl, reqHeaders, reqParma, tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[SendVoice] Voice authcode call request has wrong. voiceAuthCodeCallUrl:", voiceAuthCodeCallUrl, ", err:", err)
		return
	}

	err = json.Unmarshal(body, &VoiceAuthCodeResp)
	if err != nil {
		logs.Error("[SendVoice] Send voice authcode call request, parse body failed")
		return
	}

	return
}

// 第三方要求验证码数字用"-"连接,例如:1-5-7-2
func FormatAuthCode(code string) (codeFormat string) {
	length := len(code)
	var codeTmp []string
	for i := 0; i < length; i++ {
		codeTmp = append(codeTmp, string(code[i:i+1]))
	}

	codeFormat = strings.Join(codeTmp, "-")

	return
}
