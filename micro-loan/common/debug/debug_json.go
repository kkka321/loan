package main

import (
	"encoding/json"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"github.com/astaxie/beego/logs"
)

func main() {

	/*
		orderId := int64(180716020192591708)
		accountId := int64(180627010158727345)
		//accountIdStr := tools.Int642Str(accountId)
		bankName := "Bank Permata"
		accountHolderName := "YOLANDA MEYSHA ZULFILIA MAHMUDAH"
		accountNumber := "145478588858"
		//desc := datas["desc"].(string)
		amount := int64(90000)

		err := service.ThirdPartyDisburse(orderId, accountId, bankName, accountHolderName, accountNumber, amount, 3)
		logs.Debug(err)
	*/

	str := `{"status":0,"message":"Inquiry succeed","inquiry":{"idToken":"I0777315610184804","fund":{"origin":{"amount":600000.000000,"currency":"IDR"},"destination":{"amount":600000.00,"currency":"IDR"},"fees":{"total":5500.000000,"currency":"IDR","components":[{"description":null,"amount":5500.000000}],"additionalFee":0.000000}},"senderCountry":{"code":"ID","name":"Indonesia","currency":{"code":"IDR"}},"senderCurrency":{"code":"IDR"},"beneficiaryCountry":{"code":"ID","name":"Indonesia","currency":{"code":"IDR"}},"beneficiaryCurrency":{"code":"IDR"},"channel":{"code":"07","name":"Bank Deposit"},"forexReference":{"id":14722,"forex":{"origin":{"code":"IDR"},"destination":{"code":"IDR"}},"rate":1.0,"createdTime":1532433467529},"beneficiaryAccount":{"id":null,"bank":{"id":"013","code":"BBBAIDJA","name":"Bank Permata","city":null,"countryCode":"ID","groupBank":null,"province":null,"dcBankId":null},"number":"1224826571","name":"DYAH RORO PURBOLARAS","city":"Jakarta","address":null,"inputMode":null}}}`

	var dokuDisburseInquiryResponse struct {
		Status  interface{} `json:"status"`
		Message string      `json:"message"`
		Inquiry struct {
			IdToken string `json:"idToken"`
		}
	}

	json.Unmarshal([]byte(str), &dokuDisburseInquiryResponse)

	logs.Debug("%#v", dokuDisburseInquiryResponse)

	if _, ok := dokuDisburseInquiryResponse.Status.(float64); ok {
		logs.Debug(ok)
		logs.Debug("i am here")
	}

}
