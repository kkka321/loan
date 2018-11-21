package appsflyer

import (
	"micro-loan/common/tools"
	// // 数据库初始化
	// _ "micro-loan/common/lib/clogs"
	// _ "micro-loan/common/lib/db/mysql"
	"testing"
)

func init() {
	//   URL: https://api2.appsflyer.com/inappevent/{app_id}
	//
	// Android example: https://api2.appsflyer.com/inappevent/com.appsflyer.myapp
	apiBaseURL = "https://api2.appsflyer.com/"
}

func TestSendAddCartEvent(t *testing.T) {

	time := tools.GetUnixMillis()

	eventReq := EventReq{
		AppsflyerID:   "1523342424891-3057811623465736545",
		AdvertisingID: "3a764abb-d8bf-4231-8cf2-89c1193dd60e",
		EventName:     AddCartEv,
		EventTime:     TimeFormat(time),
		IsEventsAPI:   true,
		EventVal: AddCartEventVal{
			Price:    50,
			Currency: "IDR",
		},
	}
	success, err := eventReq.Send(1123123, "104488")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if !success {
		t.Fail()
	}
}

func TestSendInitCheckoutEvent(t *testing.T) {

	time := tools.GetUnixMillis()

	eventReq := EventReq{
		AppsflyerID:   "1523342424891-3057811623465736545",
		AdvertisingID: "3a764abb-d8bf-4231-8cf2-89c1193dd60e",
		EventName:     InitCheckoutEv,
		EventTime:     TimeFormat(time),
		IsEventsAPI:   true,
		EventVal: InitCheckoutEventVal{
			Price:    50,
			Currency: "IDR",
		},
	}
	success, err := eventReq.Send(1123123, "104488")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if !success {
		t.Fail()
	}
}

func TestSendPurchaseEvent(t *testing.T) {

	time := tools.GetUnixMillis()

	eventReq := EventReq{
		AppsflyerID:   "1523342424891-3057811623465736545",
		AdvertisingID: "3a764abb-d8bf-4231-8cf2-89c1193dd60e",
		EventName:     PurchaseEv,
		EventTime:     TimeFormat(time),
		IsEventsAPI:   true,
		EventVal: PurchaseEventVal{
			Revenue:  50,
			Price:    5000,
			Currency: "IDR",
		},
	}
	success, err := eventReq.Send(1123123, "104488")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if !success {
		t.Fail()
	}
}

func TestSendPurchaseEventInvaild(t *testing.T) {

	time := tools.GetUnixMillis()

	eventReq := EventReq{
		AppsflyerID:   "",
		AdvertisingID: "3a764abb-d8bf-4231-8cf2-89c1193dd60e",
		EventName:     PurchaseEv,
		EventTime:     TimeFormat(time),
		IsEventsAPI:   true,
		EventVal: PurchaseEventVal{
			Revenue:  50,
			Price:    5000,
			Currency: "IDR",
		},
	}
	success, err := eventReq.Send(1123123, "104488")
	if err == nil {
		t.Log("应该因为参数检查而失败")
		t.Fail()
	}
	if success {
		t.Fail()
	}
}
