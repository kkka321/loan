package api

import (
	"micro-loan/common/types"
	"net/http"
)

// Response 描述一个 Sms 借口返回值接口
type Response interface {
	IsSuccess() bool
	GetMsgID() string
}

// API 描述Sms 发送接口
type API interface {
	Send() (apiResponse Response, originalResp map[string]interface{}, err error)
	GetID() types.SmsServiceID
	Delivery(r *http.Request) (msgID string, status int, content interface{}, err error)
}

// Receiver 接收
type Receiver interface {
	Receiver()
}
