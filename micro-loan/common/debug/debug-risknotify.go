package main

import (
	"encoding/json"
	"fmt"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	_ "micro-loan/common/lib/redis/storage"
	"micro-loan/common/service"

	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

func main() {
	logs.Debug("debug api ...")

	service.InsertDefaultQuotaConf(11101010101)

	// token, err := models.GenerateAccountToken(180211010000040290, "android", "127.0.0.1")
	// token, err := service.GenTokenWithCache(180211010000040290, "android", "127.0.0.1")
	// if err != nil {
	// 	logs.Error("error:", err)
	// }
	// logs.Debug("Token:", token)

	// params := map[string]interface{}{
	// 	"access_token": token,
	// 	"req_time":     tools.TimeNow(),
	// }

	params := make([]map[string]interface{}, 0)
	param := map[string]interface{}{
		"account_id":  "180523010015856588",
		"source_from": "tongdun",
		"source_code": "107001",
	}
	params = append(params, param)
	param = map[string]interface{}{
		"account_id":  "180328010027698208",
		"source_from": "tongdun",
		"source_code": "102001",
	}
	params = append(params, param)

	reqJSON, _ := json.Marshal(params)
	fmt.Printf("reqJSON: %s\n", reqJSON)

	requestPostBody := fmt.Sprintf("data=%s", reqJSON)
	fmt.Printf("requestPostBody: %s\n", requestPostBody)

	requestHeaders := map[string]string{
		"Connection":   "keep-alive",
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
	}

	//根据请求contentType ，构造不同的包体， x-www-form-urlencoded 与 json 两种常用方式
	// requestPostBody, requestHeaders := tools.MakeReqHeadAndBody("json", params)

	// logs.Debug("requestPostBody:", requestPostBody)
	// logs.Debug("requestHeaders:", requestHeaders)

	// testUrl := "http://127.0.0.1:8600/risknotify/notify"
	// testUrl := "http://127.0.0.1:8600/risknotify/quota_conf"
	testUrl := "http://127.0.0.1:8600/risknotify/thirdparty_query"

	// testUrl := "http://127.0.0.1:8700/api/v3/identity/detect"
	fmt.Printf("-----API: %s\n", testUrl)

	// httpBody, httpStatusCode, err := tools.SimpleHttpClient("POST", testUrl, reqHeaders, reqData, tools.DefaultHttpTimeout())
	httpBody, httpStatusCode, err := tools.SimpleHttpClient("POST", testUrl, requestHeaders, requestPostBody, tools.DefaultHttpTimeout())
	fmt.Printf("httpBody: %s, httpStatusCode: %d, err: %v\n", httpBody, httpStatusCode, err)

}
