package main

import (
	"fmt"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/tools"

	"encoding/json"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

func main() {

	productId := "1493"
	//checkUrl := "http://idtool.bluepay.asia//charge/express/npwpQuery"
	checkUrl := "http://120.76.101.146:21811/charge/express/npwpQuery"
	npwp := "821691227614000"
	//
	//uuidStr := uuid.Must(uuid.NewV4()).String()
	//
	//mobile := "081320785456"
	//name := "YAYAN ARYANA asdfa"
	//accountNo := "430901008901539"
	//bankCode := "BRI"
	//
	//paramStr := fmt.Sprintf("phoneNum=%s&customerName=%s&accountNo=%s&bankName=%s&transactionId=%s", tools.UrlEncode(mobile), tools.UrlEncode(name), tools.UrlEncode(accountNo), tools.UrlEncode(bankCode), tools.UrlEncode(uuidStr))
	//logs.Debug(paramStr)
	//
	//hash := bluepay.OpenSSLEncrypt(paramStr)
	keyStr := beego.AppConfig.String("bluepay_secret_key")
	//keyStr = "ddd"

	encrypt := tools.Md5(fmt.Sprintf("productId=%s&npwp=%s%s", productId, npwp, keyStr))
	checkUrl = fmt.Sprintf("%s?productId=%s&npwp=%s&encrypt=%s", checkUrl, productId, npwp, encrypt)
	logs.Debug(checkUrl)

	reqHeaders := map[string]string{}
	httpBody, httpCode, err := tools.SimpleHttpClient("GET", checkUrl, reqHeaders, "", tools.DefaultHttpTimeout())

	if err != nil {
		logs.Error(err)
		return
	}

	logs.Debug("httpBody:%v , httpCode:%d ", string(httpBody), httpCode)

	type Resp struct {
		Status       int    `json:"status"`
		Message      string `json:"message"`
		Npwp         string `json:"npwp"`
		CustomerName string `json:"customerName"`
	}

	resp := Resp{}
	err = json.Unmarshal(httpBody, &resp)

	if err != nil {
		err = fmt.Errorf("bluepay name validator response json unmarshal failed, err is %s", err.Error())
		logs.Error(err)
		return
	}

	if httpCode != 200 {
		err = fmt.Errorf("bluepay name validator httpCode is wrong [%d]", httpCode)
		logs.Error(err)
		return
	}

	logs.Warn("resp:%#v", resp)

}
