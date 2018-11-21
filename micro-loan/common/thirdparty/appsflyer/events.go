package appsflyer

import (
	"encoding/json"
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"net/http"
	"strconv"
	"time"

	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

// see: https://support.appsflyer.com/hc/en-us/articles/207034486-Server-to-Server-In-App-Events-API-HTTP-API-
// "1、提交审核成功（对应appflyer的添加到购物车事件）；2放款成功（对应appflyer的启动结账事件）
//  接口文档：https://support.appsflyer.com/hc/en-us/articles/115005544169-Rich-In-App-Events-Android-and-iOS#add-to-cart
//  放款成功的可选参数：价格（取放款金额） 收入（取应还金额-放款金额）
//  appID :包名"

// const (
// 	// AppID
// 	AppID  = "com.loan.cash.credit.pinjam.uang.dana.rapiah"
// 	DevKey = "z4EY4D3yJVHeVzS93ba8vV"
// 	AppURL = "https://api2.appsflyer.com/inappevent/"
// )

const (
	appEventRouter = "inappevent/"
)

type appConfig struct {
	AppID  string
	DevKey string
	AppURL string
}

var (
	apiBaseURL string
)

// EventType Appsflyer Event 类型
// 详情: https://support.appsflyer.com/hc/en-us/articles/115005544169-Rich-In-App-Events-Android-and-iOS
type EventType string

// appsflyer 事件
const (
	// AddCartEv 添加购物车事件
	AddCartEv EventType = "af_add_cart"
	// InitCheckoutEv 初始化结账事件
	InitCheckoutEv EventType = "af_initiated_checkout"
	// 完成注册
	CompleteRegistration EventType = "af_complete_registration"
	// PurchaseEv 购买事件
	PurchaseEv EventType = "af_purchase"
)

// appsflyer config
const (
	appIDConfigPrefix  = "appsflyer_app_id_"
	devKeyConfigPrefix = "appsflyer_dev_key_"
)

// EventReq 描述请求body结构
type EventReq struct {
	AppsflyerID   string    `json:"appsflyer_id"`   // 必填
	AdvertisingID string    `json:"advertising_id"` // 必填
	EventName     EventType `json:"eventName"`      // 必填
	//	EventCurrency string    `json:"eventCurrency, omitempty"`
	EventVal    interface{} `json:"eventValue"`
	EventTime   string      `json:"eventTime,omitempty"`  // 可选, 如果不传此参数, 则事件时间为 请求收到时间
	IsEventsAPI bool        `json:"af_events_api,string"` // 必填
}

// AddCartEventVal 描述
// Recommended Attributes:  af_price, af_content_type, af_content_id, af_content, af_currency, af_quantity
type AddCartEventVal struct {
	Price int64 `json:"af_price"` // require
	//ContentType float64 `json:"af_content_type"`
	Currency string `json:"af_currency"`
}

// InitCheckoutEventVal 描述
// Recommended Attributes:  af_price, af_content_type, af_content_id, af_content, af_quantity, af_payment_info_available, af_currency
type InitCheckoutEventVal struct {
	Price int64 `json:"af_price"` // require
	//ContentType float64 `json:"af_content_type"`
	Currency string `json:"af_currency"`
}

// PurchaseEventVal 描述
// Recommended Attributes:  af_revenue, af_content_type, af_content_id, af_content, af_price, af_quantity, af_currency, af_order_id
type PurchaseEventVal struct {
	Revenue int64 `json:"af_revenue"` // require
	Price   int64 `json:"af_price"`   // require
	//ContentType float64 `json:"af_content_type"`
	Currency string `json:"af_currency"`
}

// CompleteRegistrationEventVal 描述
// Recommended Attributes:  af_registration_method
type CompleteRegistrationEventVal struct {
	RegMethod string `json:"af_registration_method"`
}

type response struct {
	HTTPCode int
	Body     []byte
}

func init() {

	//
	// Android example: https://api2.appsflyer.com/inappevent/com.appsflyer.myapp
	apiBaseURL = beego.AppConfig.String("appsflyer_api_base_url")
}

// eventName, eventValue, appsflyer_id (AppsFlyer Device id), idfa, advertising_id and af_events_api are all required parameters
// eventValue can be empty, e.g. "eventValue": “”, (no space is needed)
// IP address should be the IP of the mobile device.
// The customer_user_id is a user identifier parameter, mapped to the Customer User ID field in the Raw Reports. The Customer User ID SDK API attaches this value automatically to all SDK originated in-app events. Make sure to include this field will ALL your S2S events to enjoy the same functionality.
func (e *EventReq) check() error {
	if len(e.AppsflyerID) <= 0 {
		return fmt.Errorf("[Appsflyer check before send] AppsflyerID can not be empty,AppsflyerID: %s", e.AppsflyerID)
	}
	// if len(e.AdvertisingID) <= 0 {
	// 	return fmt.Errorf("[Appsflyer check before send] AdvertisingID 不能为空,AdvertisingID: %s", e.AdvertisingID)
	// }
	return nil
}

func getConfigByStemFrom(stemFrom string) {

}

// Send 发送事件给 appsflyer
// 无 relateID 则传 0
func (e *EventReq) Send(relateID int64, stemFrom string) (success bool, err error) {
	appID := beego.AppConfig.String(appIDConfigPrefix + stemFrom)
	devKey := beego.AppConfig.String(devKeyConfigPrefix + stemFrom)
	//   URL: https://api2.appsflyer.com/inappevent/{app_id}
	if len(appID) == 0 || len(devKey) == 0 {
		err = fmt.Errorf("[appsflyer_send_event] no config for stem: %s", stemFrom)
		logs.Error(err)
		return
	}

	success = false

	if err = e.check(); err != nil {
		return
	}
	apiURL := getEventAPIURL(appID)
	reqHeader := getAuthHeader(devKey)
	bytes, _ := json.Marshal(e)
	httpBody, httpCode, err := tools.SimpleHttpClient("POST", apiURL, reqHeader, string(bytes), tools.DefaultHttpTimeout())

	monitor.IncrThirdpartyCount(models.ThirdpartyAppsFlyer, httpCode)

	// 若进行单元测试, 需注释掉下面的 model 操作
	responstType, fee := thirdparty.CalcFeeByApi(apiBaseURL+appEventRouter, e, response{httpCode, httpBody})
	models.AddOneThirdpartyRecord(models.ThirdpartyAppsFlyer, apiURL, relateID, e, response{httpCode, httpBody}, responstType, fee, httpCode)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: 0,
		OrderId:       relateID,
		ApiMd5:        tools.Md5(apiURL),
		Fee:           int64(fee),
		Result:        responstType,
	})

	if err != nil {
		err = fmt.Errorf("[Appsflyer Send Event] has wrong, req: %v, httpCode: %d, httpBody: %s, httpError: %s", *e, httpCode, string(httpBody), err.Error())
		return
	}
	if httpCode == http.StatusOK {
		success = true
	} else {
		success = false
		err = fmt.Errorf("[Appsflyer Send Event] has wrong, req: %v, httpCode: %d, httpBody: %s", *e, httpCode, string(httpBody))
	}
	return
}

func getEventAPIURL(appID string) string {
	return apiBaseURL + appEventRouter + appID
}

func getAuthHeader(devKey string) map[string]string {
	return map[string]string{"authentication": devKey}
}

// TimeFormat 将毫秒时间戳转化为 AppsFlyer 要求的时间格式
// 下面是 appsflyer 文档的说明
// https://support.appsflyer.com/hc/en-us/articles/207034486-Server-to-Server-In-App-Events-API-HTTP-API-#timing-the-events
// You can use the optional eventTime parameter to specify the time of the event occurrence (in UTC timezone). If the parameter is not included in the message, AppsFlyer uses the timestamp from the HTTPS message received.
// eventTime format is: "yyyy-MM-dd HH:mm:ss.SSS" (e.g. "2014-05-15 12:17:00.000")
// In batch mode, for events to be recorded with their real time stamps, they must all be sent to AppsFlyer by 02:00 AM (UTC) of the following day.
func TimeFormat(mTime int64) string {
	if mTime <= 0 {
		return ""
	}

	tm := time.Unix(mTime/1000, 0)
	local, _ := time.LoadLocation("UTC")

	return tm.In(local).Format("2006-01-02 15:04:05") + "." + strconv.FormatInt((mTime-mTime/1000*1000), 10)
}
