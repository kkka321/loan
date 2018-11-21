package thirdparty

import (
	"testing"

	"micro-loan/common/models"
	"micro-loan/common/types"
)

// {"body":"{\"code\": \"0\", \"error\":\"\", \"msgid\":\"18061910431000435722\"}","httpCode":200,"httpErr":null}
func TestCalcFeeFuncSms253(t *testing.T) {

	td := []struct {
		thirdPartyInfo models.ThirdpartyInfo
		response       map[string]interface{}
		fee            int
	}{
		{
			thirdPartyInfo: models.ThirdpartyInfo{
				Price:      100,
				ChargeType: types.ChargeForFree,
			},
			response: map[string]interface{}{
				"body":     "",
				"httpCode": types.HTTPCodeSuccess,
				"httpErr":  nil,
			},
			fee: 0,
		},
		{
			thirdPartyInfo: models.ThirdpartyInfo{
				Price:      100,
				ChargeType: types.ChargeForCall,
			},
			response: map[string]interface{}{
				"body":     "",
				"httpCode": types.HTTPCodeSuccess + 1,
				"httpErr":  nil,
			},
			fee: 100,
		},
		{
			thirdPartyInfo: models.ThirdpartyInfo{
				Price:      100,
				ChargeType: types.ChargeForCallSuccess,
			},
			response: map[string]interface{}{
				"httpCode": types.HTTPCodeSuccess,
				"httpErr":  nil,
			},
			fee: 100,
		},
		{
			thirdPartyInfo: models.ThirdpartyInfo{
				Price:      100,
				ChargeType: types.ChargeForHit,
			},
			response: map[string]interface{}{
				"body": struct {
					Code  string `json:"code"`
					Error string `json:"error"`
					MsgID string `json:"msgid"`
				}{
					Code: "0"},
				"httpCode": types.HTTPCodeSuccess,
				"httpErr":  nil,
			},
			fee: 100,
		},
	}

	for _, d := range td {
		sms := Sms253{}
		fee, _ := sms.CalcFeeFunc(nil, d.response, d.thirdPartyInfo)
		if fee != d.fee {
			t.Errorf("fee :%d except fee:%d", fee, d.fee)
		}
	}

}
func TestCalcFeeFuncAkulaku(t *testing.T) {

	td := []struct {
		thirdPartyInfo models.ThirdpartyInfo
		response       string
		fee            int
	}{
		{
			thirdPartyInfo: models.ThirdpartyInfo{
				Price:      100,
				ChargeType: types.ChargeForFree,
			},
			response: "{\"data\": {\"creditresult\": 0, \"risktype\": \"\", \"dataNo\": \"2018061941111\"}, \"success\": \"true\", \"sysTime\": 1529404109755}",
			fee:      0,
		},
		{
			thirdPartyInfo: models.ThirdpartyInfo{
				Price:      100,
				ChargeType: types.ChargeForCall,
			},
			response: "{\"data\": {\"creditresult\": 0, \"risktype\": \"\", \"dataNo\": \"2018061941111\"}, \"success\": \"true\", \"sysTime\": 1529404109755}",
			fee:      100,
		},
		{
			thirdPartyInfo: models.ThirdpartyInfo{
				Price:      100,
				ChargeType: types.ChargeForCallSuccess,
			},
			response: "{\"data\": {\"creditresult\": 0, \"risktype\": \"\", \"dataNo\": \"2018061941111\"}, \"success\": \"true\", \"sysTime\": 1529404109755}",
			fee:      100,
		},
		{
			thirdPartyInfo: models.ThirdpartyInfo{
				Price:      100,
				ChargeType: types.ChargeForHit,
			},
			response: "{\"data\": {\"creditresult\": 0, \"risktype\": \"\", \"dataNo\": \"2018061941111\"}, \"success\": \"true\", \"sysTime\": 1529404109755}",
			fee:      100,
		},
	}

	for _, d := range td {
		akulaku := Akulaku{}
		fee, _ := akulaku.CalcFeeFunc(nil, d.response, d.thirdPartyInfo)
		if fee != d.fee {
			t.Errorf("fee :%d except fee:%d d.thirdPartyInfo.type:%d", fee, d.fee, d.thirdPartyInfo.ChargeType)
		}
	}

}
