package bluepay

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	uuid "github.com/satori/go.uuid"

	"micro-loan/common/dao"
	"micro-loan/common/lib/device"
	"micro-loan/common/lib/payment"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type BluepayCreateVAResponse struct {
	Data        BluepayCreateVADataDetailResponse `json:"data"`
	Message     string                            `json:"message"`
	Status      int                               `json:"status"`
	VaFee       int64                             `json:"vaFee"`
	IsStatic    int                               `json:"isStatic"`
	OtcFee      int64                             `json:"otcFee"`
	PaymentCode string                            `json:"payment_code"`
}

type BluepayCreateVADataDetailResponse struct {
	Msisdn        string `json:"msisdn"`
	Paymentcode   string `json:"paymentCode"`
	TransactionId string `json:transactionId`
}

type BluepayCreateDisburseResponse struct {
	TransactionId  string `json:"transactionId"`
	TransferStatus string `json:"transferStatus"`
	Code           string `json:"code"`
}

type BluepayApi struct {
	payment.PaymentApi
}

type NpwpResp struct {
	Status       int    `json:"status"`
	Message      string `json:"message"`
	Npwp         string `json:"npwp"`
	CustomerName string `json:"customerName"`
}

var bluePayBankNameCodeMap = map[string]string{
	"Bank Rakyat Indonesia (BRI)":      "BRI",
	"Bank Mandiri":                     "MANDIRI",
	"Bank Negara Indonesia (BNI)":      "BNI",
	"Bank Danamon":                     "DANAMON",
	"Bank Permata":                     "PERMATA",
	"Bank Central Asia (BCA)":          "BCA",
	"Bank Maybank":                     "BII",
	"Bank Panin":                       "PANIN",
	"Bank CIMB Niaga":                  "CIMB",
	"Bank UOB Indonesia":               "UOB",
	"Bank Artha Graha International":   "ARTA GRAHA",
	"Bank BJB":                         "BANK BJB",
	"Bank Jatim":                       "BANK JATIM",
	"BPD Kalimantan Barat":             "BPD NUSA TENGGARA BARAT",
	"Bank Nusantara Parahyangan":       "BANK NUSANTARA PARAHYANGAN",
	"Bank Muamalat Indonesia":          "BANK MUAMALAT INDONESIA",
	"Sinarmas":                         "SINARMAS",
	"Bank Tabungan Negara (BTN)":       "BANK TABUNGAN NEGARA",
	"Bank Mega":                        "MEGA",
	"Bank Bukopin":                     "BUKOPIN",
	"Bank Hana":                        "BANK HANA",
	"Centratama Nasional Bank":         "BANK CENTRATAMA NASIONAL",
	"Bank Tabungan Pensiunan Nasional": "BANK TABUNGAN PENSIUNAN NASIONAL/BTPN",
}

func BluepayBankNameCodeMap() map[string]string {
	return bluePayBankNameCodeMap
}

func BluepayBankName2Code(name string) (code string, err error) {
	bankNameCodeMap := BluepayBankNameCodeMap()
	if v, ok := bankNameCodeMap[name]; ok {
		code = v
		return
	}
	err = fmt.Errorf("bank code undefined")
	return
}

func BankName2BluepaySupportCode(name string) (code string, err error) {
	conf := map[string]bool{
		"PERMATA": true,
		"BNI":     true,
	}

	code, err = BluepayBankName2Code(name)
	if err != nil {
		return
	}

	if !conf[code] {
		code = "BNI"
	}

	return
}

func (c *BluepayApi) CreateVirtualAccount(datas map[string]interface{}) (res []byte, err error) {
	//curl 'http://120.76.101.146:21921/indonesia/express/gather/mo?price=30000&productId=1483&payType=atm&transactionId=14615984398y&ui=none&promotionId=1000&bankType=permata'
	productId, _ := beego.AppConfig.Int64("bluepay_product_id")
	virtualAccountsUrl := beego.AppConfig.String("bluepay_create_va_url")

	bankName := datas["bank_name"].(string)
	bankCode, err := BankName2BluepaySupportCode(bankName)
	if err != nil {
		return []byte{}, err
	}

	bankType := strings.ToLower(bankCode)

	mobile := datas["mobile"].(string)
	headerStr := tools.SubString(mobile, 0, 2)
	if headerStr != "62" {
		mobile = fmt.Sprintf("%s%s", "62", mobile)
	}

	price := datas["amount"].(int64)
	orderId := datas["order_id"].(int64)
	externalId := datas["account_id"].(int64)

	virtualAccountsUrl = fmt.Sprintf("%s?msisdn=%s&price=%d&productId=%d&payType=atm&transactionId=%d&ui=none&promotionId=1000&bankType=%s", virtualAccountsUrl, mobile, price, productId, orderId, bankType)

	client := &http.Client{}
	req, err := http.NewRequest("GET", virtualAccountsUrl, nil)
	if err != nil {
		logs.Error("[CreateVirtualAccount] http.NewRequest url:%s, err:%s", virtualAccountsUrl, err.Error())
		return []byte{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.SetBasicAuth(secretKey, "")
	resp, err := client.Do(req)

	monitor.IncrThirdpartyCount(models.ThirdpartyBluepay, resp.StatusCode)

	if err != nil {
		logs.Error("[CreateVirtualAccount] client.Do url:%s, err:%s", virtualAccountsUrl, err.Error())
		return []byte{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("[CreateVirtualAccount] ioutil.ReadAll url:%s, err:%s", virtualAccountsUrl, err.Error())
		return []byte{}, err
	}

	responstType, fee := thirdparty.CalcFeeByApi(virtualAccountsUrl, "", string(body))
	models.AddOneThirdpartyRecord(models.ThirdpartyBluepay, virtualAccountsUrl, externalId, "", string(body), responstType, fee, resp.StatusCode)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: externalId,
		OrderId:       orderId,
		ApiMd5:        tools.Md5(virtualAccountsUrl),
		Fee:           int64(fee),
		Result:        responstType,
	})

	return body, err
}

func (c *BluepayApi) Disburse(datas map[string]interface{}) (res []byte, err error) {
	orderId := datas["order_id"].(int64)
	bankName := datas["bank_name"].(string)
	bankCode, err := BankName2BluepaySupportCode(bankName)
	if err != nil {
		return []byte{}, err
	}
	accountHolderName := datas["account_name"].(string)
	accountNumber := datas["account_num"].(string)
	amount := datas["amount"].(int64)

	paramStr := fmt.Sprintf("transactionId=%d&promotionId=1000&payeeCountry=%s&payeeBankName=%s&payeeName=%s&payeeAccount=%s&payeeMsisdn=%d&payeeType=%s&amount=%d&currency=%s",
		orderId, types.PayeeCountryIDId, bankCode, accountHolderName, accountNumber, types.PayeeMsisdnID, types.PayeeTypePersonal, amount, types.PayeeTypeIDCurrency)

	hash := OpenSSLEncrypt(paramStr)
	keyStr := beego.AppConfig.String("bluepay_secret_key")
	productId := beego.AppConfig.String("bluepay_product_id")
	logs.Debug("[Disburse] hash:%s, keyStr:%s", hash, keyStr)
	disburseUrl := beego.AppConfig.String("bluepay_disburse_url")

	md5val := tools.Md5(fmt.Sprintf("productId=%s&data=%s%s", productId, hash, keyStr))
	disburseUrl = fmt.Sprintf("%s?productId=%s&data=%s&encrypt=%s", disburseUrl, productId, hash, md5val)

	client := &http.Client{}
	req, err := http.NewRequest("GET", disburseUrl, nil)
	if err != nil {
		logs.Error("[Disburse] http.NewRequest url:%s, err:%s", disburseUrl, err.Error())
		return []byte{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)

	monitor.IncrThirdpartyCount(models.ThirdpartyBluepay, resp.StatusCode)

	if err != nil {
		logs.Error("[Disburse] client.Do url:%s, err:%s", disburseUrl, err.Error())
		return []byte{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logs.Error("[Disburse] ioutil.ReadAll url:%s, err:%s", disburseUrl, err.Error())
		return []byte{}, err
	}

	responstType, fee := thirdparty.CalcFeeByApi(disburseUrl, "", string(body))
	models.AddOneThirdpartyRecord(models.ThirdpartyBluepay, disburseUrl, orderId, "", string(body), responstType, fee, resp.StatusCode)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: 0,
		OrderId:       orderId,
		ApiMd5:        tools.Md5(disburseUrl),
		Fee:           int64(fee),
		Result:        responstType,
	})
	return body, err
}

func (c *BluepayApi) CheckVirtualAccount(datas map[string]interface{}) (res []byte, err error) {
	return []byte{}, err
}

func (c *BluepayApi) CreateVirtualAccountResponse(jsonData []byte, datas map[string]interface{}) error {
	var resp BluepayCreateVAResponse = BluepayCreateVAResponse{}
	err := json.Unmarshal(jsonData, &resp)
	if err != nil {
		logs.Error("[CreateVirtualAccountResponse] json.Unmarshal err:%v, json:%s", jsonData, err)
		return err
	}

	//if resJson.Status != 200 {
	if resp.Status != 201 {
		errStr := fmt.Sprintf("[CreateVirtualAccountResponse] response status is wrong retJson:%s", string(jsonData))
		err := fmt.Errorf(errStr)
		return err
	}

	userAccountId := datas["account_id"].(int64)
	_, err = models.GetEAccount(userAccountId, types.Bluepay)
	if err != nil {
		//不存在则创建
		eAccount := models.User_E_Account{}
		eAccount.Id, _ = device.GenerateBizId(types.UserEAccountBiz)
		eAccount.UserAccountId = userAccountId
		eAccount.EAccountNumber = resp.PaymentCode
		eAccount.VaCompanyCode = types.Bluepay
		eAccount.Status = "pending"
		eAccount.Ctime = tools.GetUnixMillis()
		eAccount.Utime = tools.GetUnixMillis()

		_, err = eAccount.AddEAccount(&eAccount)
	}

	return err
}

func (c *BluepayApi) DisburseResponse(jsonData []byte, datas map[string]interface{}) (err error) {
	var resp BluepayCreateDisburseResponse = BluepayCreateDisburseResponse{}
	err = json.Unmarshal(jsonData, &resp)
	if err != nil {
		logs.Error("[DisburseResponse] json.Unmarshal err:%s, json:%s", err, string(jsonData))
		return err
	}

	orderId := datas["order_id"].(int64)
	bankCode := datas["bank_code"].(string)
	accountHolderName := datas["account_name"].(string)

	transactionId, _ := tools.Str2Int64(resp.TransactionId)
	if transactionId != orderId {
		//response数据如果和请求的不一致，直接报警
		errStr := fmt.Sprintf("[DisburseResponse] response error orderId:%d, restJson:%s", orderId, string(jsonData))
		logs.Error(errStr)
		err = fmt.Errorf(errStr)
		return err
	}

	order, err := models.GetOrder(orderId)
	if err != nil {
		return err
	}

	o := models.Mobi_E_Trans{}

	orderIdStr := tools.Int642Str(orderId)
	o.UserAcccountId = order.UserAccountId
	o.VaCompanyCode = types.Bluepay
	o.Amount = order.Loan
	//向上取整，百位取整
	o.PayType = types.PayTypeMoneyOut
	o.BankCode = bankCode
	o.AccountHolderName = accountHolderName
	o.DisbursementDescription = orderIdStr
	o.DisbursementId = orderIdStr
	o.Status = resp.TransferStatus
	o.Utime = tools.GetUnixMillis()
	o.Ctime = tools.GetUnixMillis()
	_, err = o.AddMobiEtrans(&o)

	return err
}

func OpenSSLEncrypt(x string) string {
	keyStr := beego.AppConfig.String("bluepay_secret_key")
	ivStr := beego.AppConfig.String("bluepay_secret_iv")
	logs.Debug("ivStr is: ", ivStr)
	key := []byte(keyStr)
	iv := []byte(ivStr)
	var plaintextblock []byte
	// Turn struct into byte slice
	plaintext := x
	// Make sure the block size is a multiple of 16
	length := len(plaintext)

	extendBlock := 16 - (length % 16)
	plaintextblock = make([]byte, length+extendBlock)
	copy(plaintextblock[length:], bytes.Repeat([]byte{uint8(extendBlock)}, extendBlock))

	copy(plaintextblock, plaintext)
	cb, err := aes.NewCipher(key)
	if err != nil {
		log.Println("error NewCipher(): ", err)
	}

	ciphertext := make([]byte, len(plaintextblock))
	mode := cipher.NewCBCEncrypter(cb, iv)
	mode.CryptBlocks(ciphertext, plaintextblock)

	text := hex.EncodeToString(ciphertext)
	//二进制转换十六进制
	str := tools.UrlEncode(base64.StdEncoding.EncodeToString([]byte(text)))
	//urlencode

	return str
}

func NameValidator(accountId int64) (bankNo string, err error) {

	productId := beego.AppConfig.String("bluepay_product_id")
	checkUrl := beego.AppConfig.String("bluepay_name_validator")

	accountBase, err := models.OneAccountBaseByPkId(accountId)

	if err != nil {
		logs.Error("can not get account_base by accountId:", accountId)
		return
	}

	profile, err := dao.CustomerProfile(accountBase.Id)
	if err != nil {
		logs.Error("can not get account_profile by account_id:", accountBase.Id)
		return
	}

	name := accountBase.Realname
	bankNumber := profile.BankNo
	bankCode, err := BluepayBankName2Code(profile.BankName)

	if err == nil {
		//目前bluepay只支持放款银行列表的二要素检查
		//所以只有支持的银行，再调用此列表，不然没有意义

		uuidStr := uuid.Must(uuid.NewV4()).String()

		paramStr := fmt.Sprintf("phoneNum=%s&customerName=%s&accountNo=%s&bankName=%s&transactionId=%s", tools.UrlEncode(accountBase.Mobile), tools.UrlEncode(name), tools.UrlEncode(bankNumber), tools.UrlEncode(bankCode), tools.UrlEncode(uuidStr))
		logs.Debug(paramStr)

		hash := OpenSSLEncrypt(paramStr)
		keyStr := beego.AppConfig.String("bluepay_secret_key")
		logs.Debug("[NameValidator] hash:%s, keyStr:%s", hash, keyStr)

		md5val := tools.Md5(fmt.Sprintf("productId=%s&data=%s%s", productId, hash, keyStr))
		checkUrl = fmt.Sprintf("%s?productId=%s&data=%s&encrypt=%s", checkUrl, productId, hash, md5val)
		logs.Debug(checkUrl)

		reqHeaders := map[string]string{}
		httpBody, httpCode, err1 := tools.SimpleHttpClient("GET", checkUrl, reqHeaders, "", tools.DefaultHttpTimeout())

		//此处报错。。。err is shadowed during return
		//只能新申请一个err1变量了
		if err1 != nil {
			logs.Error(err1)
			err = err1
			return
		}

		var nameValidatorResp struct {
			Message string `json:"message"`
			Status  int    `json:"status"`
		}

		err1 = json.Unmarshal(httpBody, &nameValidatorResp)

		if err1 != nil {
			err1 = fmt.Errorf("bluepay name validator response json unmarshal failed, err is %s", err.Error())
			logs.Error(err1)
			err = err1
			return
		}

		logs.Debug(string(httpBody))

		if httpCode != 200 {
			err1 = fmt.Errorf("bluepay name validator httpCode is wrong [%d]", httpCode)
			logs.Error(err1)
			err = err1
			return
		}

		if nameValidatorResp.Status != 200 {
			//如果没匹配上，就返回银行账号，让客户去展示给用户，让用户可以修改
			bankNo = bankNumber
		}
	}

	return
}

func NpwpVerify(accountId int64, npwp string) (resp NpwpResp, err error) {

	productId := "1493"
	//checkUrl := "http://idtool.bluepay.asia//charge/express/npwpQuery"
	router := "http://120.76.101.146:21811/charge/express/npwpQuery"

	keyStr := beego.AppConfig.String("bluepay_secret_key")

	encrypt := tools.Md5(fmt.Sprintf("productId=%s&npwp=%s%s", productId, npwp, keyStr))
	reqParm := fmt.Sprintf("productId=%s&npwp=%s&encrypt=%s", productId, npwp, encrypt)

	checkUrl := router + "/" + reqParm
	logs.Debug(checkUrl)

	reqHeaders := map[string]string{}
	httpBody, httpCode, err := tools.SimpleHttpClient("GET", router, reqHeaders, "", tools.DefaultHttpTimeout())

	if err != nil {
		logs.Error(err)
		return
	}

	logs.Debug("httpBody:%v , httpCode:%d ", string(httpBody), httpCode)

	err = json.Unmarshal(httpBody, &resp)
	if err != nil {
		err = fmt.Errorf("[NpwpVerify] bluepay  response json unmarshal failed, err is %s httpBody:%s", err.Error(), string(httpBody))
		logs.Error(err)
		return
	}

	responstType, fee := thirdparty.CalcFeeByApi(checkUrl, reqParm, httpBody)
	models.AddOneThirdpartyRecord(models.ThirdpartyBluepay, checkUrl, accountId, reqParm, httpBody, responstType, fee, 200)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: accountId,
		OrderId:       0,
		ApiMd5:        tools.Md5(router),
		Fee:           int64(fee),
		Result:        responstType,
	})

	if httpCode != 200 {
		err = fmt.Errorf("[NpwpVerify] bluepay  response httpCode is wrong [%d]", httpCode)
		logs.Error(err)
		return
	}

	logs.Warn("resp:%#v", resp)

	return
}
