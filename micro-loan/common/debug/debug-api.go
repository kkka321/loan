package main

import (
	"encoding/json"
	"fmt"

	"micro-loan/common/cerror"
	_ "micro-loan/common/lib/clogs"
	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

func main() {
	logs.Debug("debug api ...")

	params := map[string]interface{}{
		"noise":        tools.Int642Str(tools.GetUnixMillis()),
		"request_time": "12345",
		// "access_token": "",
		"access_token":     "df8f91d58af9e6cd3b82688e93e6e7f0",
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
		//"mobile":           "15801598759",
		"mobile":           "8613588888888",
		"auth_code":        "4985",
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
		"npwp_no":          "6344567543525643",
		"loan":             "800000",
		"period":           "14",
		"loan_new":         "800000",
		"period_new":       "14",
		"quota":            "400",
		"contract_amount":  "8000000",
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

	host := "http://microl-api-test.toolkits.mobi"
	//host := "http://127.0.0.1:8700"

	//testUrl := host + "/api/v1/account/auth_list"
	//testUrl := host + "/api/v2/order/loan_quota"
	testUrl := host + "/api/loan_flow/v1/order/confirm"

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
