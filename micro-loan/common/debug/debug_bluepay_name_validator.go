package main

import (
	"fmt"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/thirdparty/bluepay"
	"micro-loan/common/tools"

	"encoding/json"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	uuid "github.com/satori/go.uuid"
)

func main() {

	/*
		productId := beego.AppConfig.String("bluepay_product_id")
		checkUrl := beego.AppConfig.String("bluepay_name_validator")

		mobile := "081396559408"
		accountBase, err := models.OneAccountBaseByMobile(mobile)

		if err != nil {
			logs.Error("can not get account_base by mobile:", mobile)
			return
		}

		profile, err := dao.CustomerProfile(accountBase.Id)
		if err != nil {
			logs.Error("can not get account_profile by account_id:", accountBase.Id)
			return
		}

		name := accountBase.Realname
		accountNo := profile.BankNo
		bankCode, err := bluepay.BluepayBankName2Code(profile.BankName)

		logs.Debug(accountNo)
		logs.Debug(bankCode)

		logs.Debug(err)

		if err == nil {

			uuidStr := uuid.Must(uuid.NewV4()).String()

			paramStr := fmt.Sprintf("phoneNum=%s&customerName=%s&accountNo=%s&bankName=%s&transactionId=%s", tools.UrlEncode(mobile), tools.UrlEncode(name), tools.UrlEncode(accountNo), tools.UrlEncode(bankCode), tools.UrlEncode(uuidStr))
			logs.Debug(paramStr)

			hash := bluepay.OpenSSLEncrypt(paramStr)
			keyStr := beego.AppConfig.String("bluepay_secret_key")
			logs.Debug("[Disburse] hash:%s, keyStr:%s", hash, keyStr)

			md5val := tools.Md5(fmt.Sprintf("productId=%s&data=%s%s", productId, hash, keyStr))
			checkUrl = fmt.Sprintf("%s?productId=%s&data=%s&encrypt=%s", checkUrl, productId, hash, md5val)
			logs.Debug(checkUrl)

			reqHeaders := map[string]string{}
			httpBody, httpCode, err := tools.SimpleHttpClient("GET", checkUrl, reqHeaders, "", tools.DefaultHttpTimeout())

			logs.Debug(string(httpBody))
			logs.Debug(httpCode)
			logs.Debug(err)
		}
	*/

	/*
		bankNo, err := bluepay.NameValidator(180713010190406063)
		logs.Debug(bankNo)
		logs.Debug(err)
	*/

	productId := "1493"
	checkUrl := "http://idtool.bluepay.asia/charge/express/checkAccount"

	uuidStr := uuid.Must(uuid.NewV4()).String()

	mobile := "081320785456"
	name := "YAYAN ARYANA asdfa"
	accountNo := "430901008901539"
	bankCode := "BRI"

	paramStr := fmt.Sprintf("phoneNum=%s&customerName=%s&accountNo=%s&bankName=%s&transactionId=%s", tools.UrlEncode(mobile), tools.UrlEncode(name), tools.UrlEncode(accountNo), tools.UrlEncode(bankCode), tools.UrlEncode(uuidStr))
	logs.Debug(paramStr)

	hash := bluepay.OpenSSLEncrypt(paramStr)
	keyStr := beego.AppConfig.String("bluepay_secret_key")
	logs.Debug("[Disburse] hash:%s, keyStr:%s", hash, keyStr)

	md5val := tools.Md5(fmt.Sprintf("productId=%s&data=%s%s", productId, hash, keyStr))
	checkUrl = fmt.Sprintf("%s?productId=%s&data=%s&encrypt=%s", checkUrl, productId, hash, md5val)
	logs.Debug(checkUrl)

	reqHeaders := map[string]string{}
	httpBody, httpCode, err := tools.SimpleHttpClient("GET", checkUrl, reqHeaders, "", tools.DefaultHttpTimeout())

	if err != nil {
		logs.Error(err)
		return
	}

	var nameValidatorResp struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}

	err = json.Unmarshal(httpBody, &nameValidatorResp)

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

	logs.Debug(string(httpBody))

}
