package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/service"
	"micro-loan/common/types"
)

func main() {
	//testRoleUser()
	//mobile := "8618518027928" //zhangao
	// mobile := "081382144181"  //曹俊鹏
	//mobile := "0821144358853" //刘开宏 已欠费
	// mobile := "081337898737" // 宾杰莹
	//mobile := "081246510493" // life  亚杰
	mobile := "082114370884" //
	//mobile := "08558330598" //
	// msg := "A test msg 2"

	service.SendSms(types.ServiceRequestLogin, types.AuthCodeTypeText, mobile, "127.0.0.1")
	// var relatedID int64 = 1111
	// sms.Send(mobile, msg, relatedID)
}
