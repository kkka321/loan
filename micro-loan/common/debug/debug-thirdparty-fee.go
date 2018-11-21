package main

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/cerror"
	"micro-loan/common/tools"
)

// +-------------------------------------------------------------------------------------------------------------------------------------+
// | api                                                                                                                                 |
// +-------------------------------------------------------------------------------------------------------------------------------------+
// | 1、/appsflyer/callback/install                                                                                                         |
// | 2、/bluepay/callback                                                                                                                   |
// | 3、/xendit/disburse_fund_callback/create                                                                                               |
// | 4、/xendit/fva_receive_payment_callback/create                                                                                         |
// | 5、/xendit/virtual_account_callback/create                                                                                             |
// | 6、http://intapi.253.com/send/json                                                                                                     |
// | 7、https://api-sgp.megvii.com/faceid/v1/detect                                                                                         |
// | 8、https://api-sgp.megvii.com/faceid/v2/verify                                                                                         |
// | 9、https://api.advance.ai/openapi/anti-fraud/v2/identity-check                                                                         |
// | 10、https://api.advance.ai/openapi/face-recognition/v2/check                                                                            |
// | 11、https://api.advance.ai/openapi/face-recognition/v2/id-check                                                                         |
// | 12、https://api.advance.ai/openapi/face-recognition/v2/ocr-check                                                                        |
// | 13、https://api.xendit.co/callback_virtual_accounts                                                                                     |
// | 14、https://api.xendit.co/disbursements                                                                                                 |
// | 15、https://api2.appsflyer.com/inappevent/com.loan.cash.credit.easy.kilat.cepat.pinjam.uang.dana.rupiah                                 |
// | 16、https://api2.appsflyer.com/inappevent/com.loan.cash.credit.pinjam.uang.dana.rapiah                                                  |
// | 17、https://credit.akulaku.com/api/v2/credit_query                                                                                      |
// | 18、https://rest.nexmo.com/sms/json                                                                                                     |
// | 19、https://talosapi.shujumohe.com/octopus/task.unify.acquire/v3?partner_code=mobi_hw_mohe&partner_key=9d8fb869e5a29d8eb806403b41406c43 |
// | 20、https://talosapi.shujumohe.com/octopus/task.unify.create/v3?partner_code=mobi_hw_mohe&partner_key=9d8fb869e5a29d8eb806403b41406c43  |
// | 21、https://talosapi.shujumohe.com/octopus/task.unify.query/v3?partner_code=mobi_hw_mohe&partner_key=9d8fb869e5a29d8eb806403b41406c43   |
// +-------------------------------------------------------------------------------------------------------------------------------------+

func main() {
	logs.Debug("debug thirdparty fee ...")

	params := map[string]interface{}{
		"noise":            tools.Int642Str(tools.GetUnixMillis()),
		"request_time":     "12345",
		"access_token":     "fd558bead1206353adff5d365f1b2504",
		"app_version":      "1.0.0.0",
		"app_version_code": "3",
		"platform":         "android",
		"network":          "wifi",
		"latitude":         "0.001",
		"is_simulator":     "0",
		"os":               "linux",
		"model":            "GX",
		"brand":            "google",
		"longitude":        "1.122",
		"imei":             "xxxwsssooll",
		"time_zone":        "GTM",
		"mobile":           "8615201694230",
		"auth_code":        "3641",
		"fs1_size":         "1024",
		"fs2_size":         "2048",
		"gender":           "0",
		"realname":         "MIRA AMALIA WULAN",
		"identity":         "3273205803840003",
		"job_type":         "2",
		"monthly_income":   "3",
		"company_name":     "Abs TTY",
		"company_city":     "x,y,z",
		"company_address":  "a-z",
		"service_years":    "5",
		"contact1":         "0876543219",
		"contact1_name":    "zhangsang",
		"relationship1":    "2",
		"contact2":         "12565678909",
		"contact2_name":    "wangmmz",
		"relationship2":    "6",
		"education":        "5",
		"marital_status":   "2",
		"children_number":  "3",
		"resident_city":    "h-y-c=cBBCX",
		"resident_address": "good good has wrong",
		"bank_name":        "ACCB",
		"bank_no":          "6344567543525643",
		"loan":             "20000",
		"period":           "14",
		"offset":           "0",
		"city":             "MALUKU",
		"ui_version":       "testuiversion",
		"tags":             "1,2,4,2048",
		"content":          "It's a feedback test content, from debug api, need more characters, just long enough for test",
		"code":             "8841",
		"channel_type":     "YYS",
	}

	signature := tools.Signature(params, tools.GetSignatureSecret())
	params["signature"] = signature
	//fmt.Printf("params: %v\n", params)

	reqJSON, _ := json.Marshal(params)
	fmt.Printf("reqJSON: %s\n", reqJSON)

	dataEncrypt, _ := tools.AesEncryptCBC(string(reqJSON), tools.AesCBCKey, tools.AesCBCIV)
	reqData := fmt.Sprintf("data=%s", dataEncrypt)
	fmt.Printf("reqData: %s\n", reqData)

	testUrl := testThirdParty__6()
	// testUrl := "http://127.0.0.1:8700/api/v1/request_login_auth_code"
	//testUrl := "http://127.0.0.1:8710/api/v1/request_login_auth_code"
	//testUrl := "http://127.0.0.1:8700/api/v1/login"
	//testUrl := "http://127.0.0.1:8700/api/v1/identity/detect"
	//testUrl := "http://127.0.0.1:8700/api/v1/account/info"
	//testUrl := "http://127.0.0.1:8700/api/v1/account/u/base"
	//testUrl := "http://127.0.0.1:8700/api/v1/account/u/work"
	//testUrl := "http://127.0.0.1:8700/api/v1/account/u/contact"
	//testUrl := "http://127.0.0.1:8700/api/v1/account/u/other"
	//testUrl := "http://127.0.0.1:8700/api/v1/order/repeat_auth_code"
	//testUrl := "http://127.0.0.1:8700/api/v1/order/repeat_verify"
	//testUrl := "http://127.0.0.1:8700/api/v1/order/current"
	//testUrl := "http://127.0.0.1:8700/api/v1/order/all"
	//testUrl := "http://127.0.0.1:8700/api/v1/order/confirm"
	//testUrl := "http://127.0.0.1:8700/api/v1/logout"
	//testUrl := "http://127.0.0.1:8700/api/v1/feedback/create"

	// testUrl := "http://127.0.0.1:8700/api/v1/account/operator_acquire_code"
	// testUrl := "http://127.0.0.1:8700/api/v1/account/operator_verify_code"
	// testUrl := "http://127.0.0.1:8700/api/v1/dot/dot1"
	// testUrl := "http://127.0.0.1:8700/api/v1/dot/dot2"

	// testUrl := "http://127.0.0.1:8700/api/v3/identity/detect"
	fmt.Printf("-----API: %s\n", testUrl)

	reqHeaders := map[string]string{
		"Connection":       "keep-alive",
		"Content-Type":     "application/x-www-form-urlencoded",
		"User-Agent":       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
		"X-Encrypt-Method": "AES",
	}

	httpBody, httpStatusCode, err := tools.SimpleHttpClient("POST", testUrl, reqHeaders, reqData, tools.DefaultHttpTimeout())
	fmt.Printf("httpBody: %s, httpStatusCode: %d, err: %v\n", httpBody, httpStatusCode, err)

	var apiData cerror.ApiResponse
	err = json.Unmarshal(httpBody, &apiData)
	if apiData.Code == cerror.CodeSuccess {
		apiResData, _ := tools.AesDecryptUrlCode(apiData.Data.(string), tools.AesCBCKey, tools.AesCBCIV)
		fmt.Printf("apiResData: %s\n", apiResData)
	} else {
		fmt.Printf("接口数据有误.\n")
	}
}

// +-------------------------------------------------------------------------------------------------------------------------------------+
// | api                                                                                                                                 |
// +-------------------------------------------------------------------------------------------------------------------------------------+
// | 1、/appsflyer/callback/install                                                                                                         |
// | 2、/bluepay/callback                                                                                                                   |
// | 3、/xendit/disburse_fund_callback/create                                                                                               |
// | 4、/xendit/fva_receive_payment_callback/create                                                                                         |
// | 5、/xendit/virtual_account_callback/create                                                                                             |
// | 6、http://intapi.253.com/send/json                                                                                                     |
// | 7、https://api-sgp.megvii.com/faceid/v1/detect                                                                                         |
// | 8、https://api-sgp.megvii.com/faceid/v2/verify                                                                                         |
// | 9、https://api.advance.ai/openapi/anti-fraud/v2/identity-check                                                                         |
// | 10、https://api.advance.ai/openapi/face-recognition/v2/check                                                                            |
// | 11、https://api.advance.ai/openapi/face-recognition/v2/id-check                                                                         |
// | 12、https://api.advance.ai/openapi/face-recognition/v2/ocr-check                                                                        |
// | 13、https://api.xendit.co/callback_virtual_accounts                                                                                     |
// | 14、https://api.xendit.co/disbursements                                                                                                 |
// | 15、https://api2.appsflyer.com/inappevent/com.loan.cash.credit.easy.kilat.cepat.pinjam.uang.dana.rupiah                                 |
// | 16、https://api2.appsflyer.com/inappevent/com.loan.cash.credit.pinjam.uang.dana.rapiah                                                  |
// | 17、https://credit.akulaku.com/api/v2/credit_query                                                                                      |
// | 18、https://rest.nexmo.com/sms/json                                                                                                     |
// | 19、https://talosapi.shujumohe.com/octopus/task.unify.acquire/v3?partner_code=mobi_hw_mohe&partner_key=9d8fb869e5a29d8eb806403b41406c43 |
// | 20、https://talosapi.shujumohe.com/octopus/task.unify.create/v3?partner_code=mobi_hw_mohe&partner_key=9d8fb869e5a29d8eb806403b41406c43  |
// | 21、https://talosapi.shujumohe.com/octopus/task.unify.query/v3?partner_code=mobi_hw_mohe&partner_key=9d8fb869e5a29d8eb806403b41406c43   |
// +-------------------------------------------------------------------------------------------------------------------------------------+

func testThirdParty__6() (url string) {
	// http://intapi.253.com/send/json
	return "http://127.0.0.1:8700/api/v1/request_login_auth_code"
}
