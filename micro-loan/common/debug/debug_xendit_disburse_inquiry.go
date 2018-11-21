package main

import (
	"encoding/json"
	"fmt"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/thirdparty/xendit"
	"micro-loan/common/tools"

	"github.com/astaxie/beego"
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

	//code := doku.DoKuVaBankCodeTransform("BMRIIDJA1")
	//logs.Debug(code)

	accountId := int64(180813018257657471)
	orderId := int64(180814028391632797)

	inquiryUrl := beego.AppConfig.String("xendit_disburse_inquiry")
	secretKey := "xnd_production_OoqIfL0j0batncU8L7IUGTfCbtGkptF8xSDi+Rxi+mDR/bCmDgN/jg==:"
	inquiryUrl = fmt.Sprintf("%s%d", inquiryUrl, accountId)

	logs.Debug(inquiryUrl)

	auth := tools.BasicAuth(secretKey, "")
	reqHeaders := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + auth,
	}

	var inquiryResp []xendit.XenditCorrectInquiryResp

	httpBody, httpCode, err := tools.SimpleHttpClient("GET", inquiryUrl, reqHeaders, "", tools.DefaultHttpTimeout())

	logs.Debug(string(httpBody))
	logs.Debug(httpCode)
	logs.Debug(err)
	err = json.Unmarshal(httpBody, &inquiryResp)

	if err != nil {
		err = fmt.Errorf("DisburseInquiry json.Unmarshal err, err is ", err.Error())
		//status = inquiryResp.
		logs.Error(err)
		return
	}

	if httpCode == 0 {
		//表示本次已经超时
		err = fmt.Errorf("DisburseInquiry timeout, httpCode is %d, err is %s,  httpBody is %s", httpCode, err.Error(), string(httpBody))
		logs.Error(err)
		return
	}

	if httpCode != 200 {
		//除了超时的其他错误
		err = fmt.Errorf("DisburseInquiry httpCode is wrong, httpCode is %d, err is %s,  httpBody is %s", httpCode, err.Error(), string(httpBody))
		logs.Error(err)
		return
	}

	/*

		resp := xendit.XenditCorrectInquiryResp{}

		for i := 0; i < len(inquiryResp); i++ {
			disbursementDescription, _ := tools.Str2Int64(inquiryResp[i].DisbursementDescription)
			if disbursementDescription == orderId && inquiryResp[i].Status == "COMPLETED" {
				//disbursementDescription是我方服务器在放款时候传入的orderId
				//查询的时候对方服务器会将此参数传回，方便我们查询
				//此处验证时，只要保证订单是一样的，并且其中一单是完成的，证明对方已经成功放款过
				resp = inquiryResp[i]
				break
			}
		}

		logs.Debug(resp)
	*/

	//accountId := int64(180813018257657471)
	//orderId := int64(180814028391632797)

	xendit.DisburseInquiry(accountId, orderId)

	/*
		//var content []xendit.XenditCorrectInquiryResp

		content := []xendit.XenditCorrectInquiryResp{
			xendit.XenditCorrectInquiryResp{
				Status: "asdfa",
				UserId: "asdfasdf",
			},
			xendit.XenditCorrectInquiryResp{
				Status: "asdfa",
				UserId: "asdfasdf",
			},
		}

		var test xendit.XenditEntireInquiryResp
		//test.Normal[0] = content
		test.ErrorCode = "asdf"
		test.Message = "asdfasdfasd"

		test.Normal = content

		//logs.Debug(test)

		jsonStr := `[{"k2":"v2"}]`

		//jsonStr := `{"k1":"v1"}`
		var jsonStruct []map[string]interface{}
		json.Unmarshal([]byte(jsonStr), &jsonStruct)

		if len(jsonStruct) != 0 {
			//如果请求响应正确返回数组
			logs.Debug(jsonStruct)
		} else {
			//否则返回字典。。。
			jsonErrStruct := map[string]interface{}{}
			json.Unmarshal([]byte(jsonStr), &jsonErrStruct)
			logs.Debug(jsonErrStruct)
		}
	*/
}
