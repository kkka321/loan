package main

import (
	"encoding/json"
	"fmt"

	_ "micro-loan/common/lib/clogs"
	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

func main() {
	logs.Debug("debug api ...")

	params := map[string]interface{}{
		"noise":        tools.Int642Str(tools.GetUnixMillis()),
		"request_time": "12345",
		"access_token": "104634ac523307a808360c3118fdbcc4",
		"app_version":  "1.0.0.0",
		"platform":     "android",
		"network":      "wifi",
		"latitude":     "0.001",
		"is_simulator": "0",
		"os":           "linux",
		"model":        "GX",
		"brand":        "google",
		"longitude":    "1.122",
		"imei":         "xxxwsssooll",
		"time_zone":    "GTM",
		"mobile":       "13811223456",
		"auth_code":    "3641",
		"fs1_size":     "1024",
		"fs2_size":     "2048",
		"fs3_size":     "2048",
		"fs4_size":     "2048",
		"fs5_size":     "2048",
		"delta":        "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",
	}
	signature := tools.Signature(params, tools.GetSignatureSecret())
	params["signature"] = signature
	fmt.Printf("params: %v\n", params)

	reqJSON, _ := json.Marshal(params)
	fmt.Printf("reqJSON: %s\n", reqJSON)
	dataEncrypt, _ := tools.AesEncryptCBC(string(reqJSON), tools.AesCBCKey, tools.AesCBCIV)
	dataEncryptUrlDecode, _ := tools.UrlDecode(dataEncrypt)
	reqData := map[string]string{
		"data": fmt.Sprintf("%s", dataEncryptUrlDecode),
	}
	fmt.Printf("reqData: %s\n", reqData)

	dataDecrypt, err := tools.AesDecryptUrlCode(dataEncrypt, tools.AesCBCKey, tools.AesCBCIV)
	fmt.Printf("dataDecrypt: %s, err: %v\n", dataDecrypt, err)

	// testUrl := "http://127.0.0.1:8700/api/v1/identity/detect"
	// testUrl := "http://127.0.0.1:8700/api/v1/account/verify"
	testUrl := "http://127.0.0.1:8700/api/loan_flow/v2/account/verify"
	fmt.Printf("-----API: %s\n", testUrl)

	reqHeaders := map[string]string{
		"Connection":       "keep-alive",
		"User-Agent":       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
		"X-Encrypt-Method": "AES",
	}

	// 身份识别
	// files := map[string]string{
	// 	"fs1": "/opt/data/Indonesia/1.jpeg",
	// 	"fs2": "/opt/data/Indonesia/11.jpeg",
	// }

	// 活体识别
	// files := map[string]string{
	// 	"fs1": "/opt/data/faceid/image_best.jpg",
	// 	"fs2": "/opt/data/faceid/image_env.jpg",
	// 	"fs3": "/opt/data/faceid/image_action1.jpg",
	// 	"fs4": "/opt/data/faceid/image_action2.jpg",
	// 	"fs5": "/opt/data/faceid/image_action3.jpg",
	// }
	// 活体识别
	files := map[string]string{
		"fs1": "/Users/mac/Documents/livingbig.jpeg",
		"fs2": "/Users/mac/Documents/livingbest.jpeg",
		"fs3": "/Users/mac/Documents/live_action1.jpeg",
		"fs4": "/Users/mac/Documents/live_action2.jpeg",
		"fs5": "/Users/mac/Documents/live_action3.jpeg",
	}

	resByte, httpCode, err := tools.MultipartClient(testUrl, reqData, reqHeaders, files, tools.DefaultHttpTimeout())
	fmt.Printf("resByte: %s, httpCode: %d, err: %v\n", resByte, httpCode, err)
}
