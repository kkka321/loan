package sms

import (
	"micro-loan/common/lib/sms/api"
	"micro-loan/common/thirdparty/nexmo"
	"micro-loan/common/thirdparty/sms253"
	"reflect"
	"testing"
)

func Test_generateSmsCallbackKey(t *testing.T) {
	td := []struct {
		in  api.SmsServiceID
		out string
	}{
		{
			api.Sms253ID, "73d7a156bf120441f35dfd51125ec971",
		},
		{
			api.NexoID, "057bb802e0aa16b7ce19844a2372e8b9",
		},
	}
	for _, d := range td {
		if d.out != generateSmsCallbackKey(d.in) {
			t.Errorf("[SMS delivery] Encrypt key[original:%d, encrypt:%s] are not equal to the test, may be encypt func or service id be changed, but it shouldn't, callback will failed",
				d.in, d.out)
		}
	}
}

func TestGetHandlerByEncryptKey(t *testing.T) {
	td := []struct {
		in  string
		out api.API
		err error
	}{
		{
			"73d7a156bf120441f35dfd51125ec971", &sms253.Sender{}, nil,
		},
		{
			"057bb802e0aa16b7ce19844a2372e8b9", &nexmo.Sender{}, nil,
		},
		{
			"057bb802e0aa16b7ce19844a237", nil, nil,
		},
	}
	for _, d := range td {
		if o, _ := getHandlerByEncryptKey(d.in); reflect.TypeOf(o) != reflect.TypeOf(d.out) {
			t.Errorf("[SMS delivery] key [%s]对应的预设sender类型[%v] 与实际返回类型[%v]不符 ",
				d.in, reflect.TypeOf(d.out), reflect.TypeOf(o))
		}
	}
}
