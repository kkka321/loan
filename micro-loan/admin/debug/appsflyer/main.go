package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/thirdparty/appsflyer"
	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

func main() {
	eventReq := appsflyer.EventReq{
		AdvertisingID: "11112212312",
		AppsflyerID:   "12312312312",
		EventName:     appsflyer.PurchaseEv,
		EventTime:     appsflyer.TimeFormat(tools.GetUnixMillis()),
		IsEventsAPI:   true,
		EventVal: appsflyer.PurchaseEventVal{
			Revenue:  1,
			Price:    2,
			Currency: tools.GetServiceCurrency(),
		},
	}
	success, errSend := eventReq.Send(123213122131, "")
	if !success || errSend != nil {
		//ordderDataJSON, _ := tools.JsonEncode(order)
		logs.Warn("[loanSuccessEv] appsflyer push event, orderID: %d, customerID: %d, error:%v", 123213122131, 1111, errSend)
	}
}
