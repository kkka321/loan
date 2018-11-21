package main

import (
	"encoding/json"
	"fmt"

	"micro-loan/common/cerror"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

func main() {

	// entrust.ProcessedNotify("111,12222,")
	// os.Exit(0)

	// orders, num := entrust.GetEntrustList("180316020021615014,22222,", 5)
	// logs.Debug("orders:", orders, "num:", num)
	// os.Exit(0)

	logs.Debug("debug api ...")

	// entrust.AssignUrgeOrder()
	// logs.Debug("[assign] num,orders,err", num, orders, err)

	// os.Exit(0)

	group := "jucegroup"

	// outsource/v1/case/sync/base_info

	// params := map[string]interface{}{
	// 	"noise":         tools.Int642Str(tools.GetUnixMillis()),
	// 	"request_time":  tools.Int642Str(tools.GetUnixMillis()),
	// 	"pname":         group,
	// 	"page_size":     "5",
	// 	"order_id_list": "",
	// }

	// params := map[string]interface{}{
	// 	"noise":         tools.Int642Str(tools.GetUnixMillis()),
	// 	"request_time":  tools.Int642Str(tools.GetUnixMillis()),
	// 	"pname":         group,
	// 	"page_size":     "5",
	// 	"order_id_list": "",
	// }

	// params := map[string]interface{}{
	// 	"noise":        tools.Int642Str(tools.GetUnixMillis()),
	// 	"request_time": tools.Int642Str(tools.GetUnixMillis()),
	// 	"pname":        group,
	// }

	// params := map[string]interface{}{
	// 	"noise":         tools.Int642Str(tools.GetUnixMillis()),
	// 	"request_time":  tools.Int642Str(tools.GetUnixMillis()),
	// 	"pname":         group,
	// 	"order_id_list": "180316020021615014",
	// 	"page_size":     "5",
	// }

	// params := map[string]interface{}{
	// 	"noise":         tools.Int642Str(tools.GetUnixMillis()),
	// 	"request_time":  tools.Int642Str(tools.GetUnixMillis()),
	// 	"pname":         group,
	// 	"page_size":     "5000",
	// 	"order_id_list": "180830020013009607",
	// }

	// params := map[string]interface{}{
	// 	"noise":        tools.Int642Str(tools.GetUnixMillis()),
	// 	"request_time": tools.Int642Str(tools.GetUnixMillis()),
	// 	"pname":        group,
	// 	"id_card_list": "3202345305940001,3578216403800002",
	// 	"page_size":    "5",
	// }

	// params := map[string]interface{}{
	// 	"noise":         tools.Int642Str(tools.GetUnixMillis()),
	// 	"request_time":  tools.Int642Str(tools.GetUnixMillis()),
	// 	"pname":         group,
	// 	"page_size":     "5000",
	// 	"order_id_list": "180830020013009607",
	// }

	params := map[string]interface{}{
		"noise":         tools.Int642Str(tools.GetUnixMillis()),
		"request_time":  tools.Int642Str(tools.GetUnixMillis()),
		"pname":         group,
		"order_id_list": "180910020004574995,180922020003278527,180906020003931823,180920020024086608,180912020013862288,180921020010773835,180925020017697186,180923020015819942,180925020015120596,180927020000606062,180910020002332589,180913020006892543",
	}

	secret := tools.GetEntrustSignatureSecret(group)
	logs.Debug("secret:", secret)

	signature := tools.Signature(params, secret)
	params["signature"] = signature
	//fmt.Printf("params: %v\n", params)

	reqJSON, _ := json.Marshal(params)
	fmt.Printf("reqJSON: %s\n", reqJSON)

	reqData := fmt.Sprintf("data=%s", reqJSON)
	fmt.Printf("reqData: %s\n", reqData)

	//dataDecrypt, err := tools.AesDecryptUrlCode(dataEncrypt, tools.AesCBCKey, tools.AesCBCIV)
	//fmt.Printf("dataDecrypt: %s, err: %v\n", dataDecrypt, err)

	// host := "http://127.0.0.1:8700/"
	// host := "http://microl-api-test.toolkits.mobi/"
	host := "https://api.rupiahcepatweb.com/"

	// testUrl := host +"outsource/v1/case/sync/base_info"
	// testUrl := host +"outsource/v1/case/sync/processed_callback"
	// testUrl := host + "outsource/v1/case/sync/repay_status"
	// testUrl := host +"outsource/v1/case/sync/repaylist"
	testUrl := host + "outsource/v1/case/sync/repay_status"
	// testUrl := host +"outsource/v1/case/contacts"
	// testUrl := host + "outsource/v1/case/rolltc"
	// testUrl := host + "outsource/v1/case/spayment_code"

	fmt.Printf("-----API: %s\n", testUrl)

	reqHeaders := map[string]string{
		"Connection":   "keep-alive",
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
	}

	httpBody, httpStatusCode, err := tools.SimpleHttpClient("POST", testUrl, reqHeaders, reqData, tools.DefaultHttpTimeout())
	fmt.Printf("httpBody: %s, httpStatusCode: %d, err: %v\n", httpBody, httpStatusCode, err)

	logs.Debug("httpBody:", string(httpBody))

	var apiData cerror.APIEntrustResponse
	err = json.Unmarshal(httpBody, &apiData)

	logs.Debug("apiData:", apiData)

	// if apiData.Code == cerror.CodeSuccess {
	// 	apiResData, _ := tools.AesDecryptUrlCode(apiData.Data.(string), tools.AesCBCKey, tools.AesCBCIV)
	// 	fmt.Printf("apiResData: %s\n", apiResData)
	// } else {
	// 	fmt.Printf("接口数据有误.\n")
	// }
}
