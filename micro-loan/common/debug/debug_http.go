package main

import (
	//"encoding/json"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty/xendit"

	"github.com/astaxie/beego/logs"
)

func main() {

	/*
		secretKey := beego.AppConfig.String("secret_key")
		fixPaymentCode := "https://api.xendit.co/fixed_payment_code"
		fixPaymentCode = "https://api.xendit.co/fixed_payment_code/5bcd990e64c95e0f4d35d521"

		externalId := "demo_fixed_payment_code_124"
		retailOutletName := "ALFAMART"
		name := "name2"
		expectedAmount := 9999

		paramStr := fmt.Sprintf("%s%s%s%s%s%s%s%d",
			"external_id=", externalId,
			"&retail_outlet_name=", retailOutletName,
			"&name=", name,
			"&expected_amount=", expectedAmount)

		logs.Debug(paramStr)

		reqHeader := map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"Authorization": "Basic " + tools.BasicAuth(secretKey, ""),
		}
		logs.Debug(fixPaymentCode)
		//httpBody, httpCode, err := tools.SimpleHttpClient("POST", fixPaymentCode, reqHeader, paramStr, tools.DefaultHttpTimeout())
		httpBody, httpCode, err := tools.SimpleHttpClient("PATCH", fixPaymentCode, reqHeader, paramStr, tools.DefaultHttpTimeout())
		//httpBody, httpCode, err := tools.SimpleHttpClient("GET", fixPaymentCode, reqHeader, "", tools.DefaultHttpTimeout())
		logs.Debug(string(httpBody))
		logs.Debug(httpCode)
		if err != nil {
			logs.Error("[CreateVirtualAccount] SimpleHttpClient error url:%s, params:%s, err:%s", fixPaymentCode, paramStr, err.Error())
		}
	*/

	/*
		{"owner_id":"5a743292ea1830b877710ed2","external_id":"demo_fixed_payment_code_123","retail_outlet_name":"ALFAMART","prefix":"KF88","name":"name1","payment_code":"KF8869487","type":"USER","expected_amount":10000,"is_single_use":false,"expiration_date":"2049-10-21T17:00:00.000Z","id":"5bcd990e64c95e0f4d35d521"}
	*/

	/*
		obj := models.FixPaymentCode{}
		obj.Id = "test2"
		obj.UserAccountId = 123124
		obj.PaymentCode = "paymentcode2"
		obj.ResponseJson = "{'ab':'b'}"
		obj.ExpirationDate = 1
		models.AddFixPaymentCode(&obj)

		obj2, _ := models.OneFixPaymentCodeById("test1")
		obj2.ExpirationDate = 100
		fields := []string{"expiration_date"}
		models.UpdateFixPaymentCode(&obj2, fields)
	*/

	/*
		obj := models.FixPaymentCodeOrder{}
		obj.Id = 1
		obj.UserAccountId = 123124
		obj.PaymentCode = "paymentcode2"
		obj.OrderId = 1234
		obj.ExpectedAmount = 199999
		models.AddFixPaymentCodeOrder(&obj)
	*/

	/*
		data, _ := models.OneFixPaymentCodeOrderById(1)
		logs.Debug(data)
	*/

	order, _ := models.GetOrder(180917020000014039)
	//xendit.XenditAddMarketPayment(order, 793408)
	err, marketPayment, _ := xendit.MarketPaymentCodeGenerate(order.Id, false, 0)
	logs.Debug(marketPayment)
	logs.Debug(err)

}
