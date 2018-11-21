package okdollar

import (
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/payment"
	"micro-loan/common/tools"
)

type OkdollarApi struct {
	payment.PaymentApi
}

type PaymentRequest struct {
	EncryptedText string `json:"EncryptedText "`
}

type PaymentRequestDetail struct {
	Destination    string `json:"Destination"`				//Merchant OK$ Number
	Amount  	   int64  `json:"Amount"`
	Source         string `json:"Source"`					//Customer OK$/Mobile Number
	ApiKey         string `json:"ApiKey"`

	MerchantName   string `json:"MerchantName"`				//Merchant Name


	RefNumber      string `json:"RefNumber"`
}

type PaymentCallBackData struct {
	ResponseCode            string `json:"ResponseCode"`
	Destination             string `json:"Destination"`
	Source                  string `json:"Source"`
	Amount               	string `json:"Amount"`
	TransactionId           string `json:"TransactionId"`
	TransactionTime 		string `json:"TransactionTime"`
	AgentName             	string `json:"AgentName"`
	Kickvalue               string `json:"Kickvalue"`
	Loyaltypoints           string `json:"Loyaltypoints"`
	Description     		string `json:"Description"`
	MerRefNo                string `json:"MerRefNo"`
	CustomerNumber          string `json:"CustomerNumber"`
}

func (c *OkdollarApi) CreateVirtualAccount(datas map[string]interface{}) (res []byte, err error) {
	return []byte{}, err
}

func (c *OkdollarApi) CheckVirtualAccount(datas map[string]interface{}) (res []byte, err error) {
	return []byte{}, err
}

func (c *OkdollarApi) Disburse(datas map[string]interface{}) (res []byte, err error) {
	return []byte{}, err
}

func (c *OkdollarApi) CreateVirtualAccountResponse(jsonData []byte, datas map[string]interface{}) (err error) {
	return nil
}

func (c *OkdollarApi) DisburseResponse(jsonData []byte, datas map[string]interface{}) (err error) {
	return nil
}

func ReceivePayment(datas map[string]interface{}) error {
	apiHost := beego.AppConfig.String("okdollar_host")

	account := beego.AppConfig.String("okdollar_account")
	accountName := beego.AppConfig.String("okdollar_account_name")
	apiKey := beego.AppConfig.String("okdollar_api_key")
	secretKey := beego.AppConfig.String("okdollar_secret_key")
	orderId := datas["order_id"].(int64)
	orderAccount := datas["account_num"].(string)

	detailReq := PaymentRequestDetail{}
	detailReq.Amount = datas["amount"].(int64)
	detailReq.Destination = account
	detailReq.ApiKey = apiKey
	detailReq.MerchantName = accountName
	detailReq.RefNumber = tools.Int642Str(orderId)
	detailReq.Source = orderAccount

	iv := tools.Int642Str(orderId)

	detailRawJson, _ := json.Marshal(detailReq)
	//`{"Destination":"00959790648563","Amount":10,"Source":"00959976830268","ApiKey":"26E6D18B4497","MerchantName":"CGM","RefNumber":"CGMEComAx20161331924438"}`
	//detailRawJson := []byte(`{"Destination":"00959790648563","Amount":10,"Source":"00959976830268","ApiKey":"26E6D18B4497","MerchantName":"CGM","RefNumber":"CGMEComAx20161331924438"}`)
	fmt.Println(string(detailRawJson))
	detailData, _ := tools.Encrypter(detailRawJson, []byte(secretKey), []byte(iv), tools.AesPKCS7)
	fmt.Println(string(detailData))
	detailStr := tools.Base64Encode(detailData) + "," + iv + "," + account

	var clusterinfo = url.Values{}
	clusterinfo.Set("requestToJson", detailStr)

	paramStr := clusterinfo.Encode()

	client := &http.Client{}

	req, err := http.NewRequest("POST", apiHost, strings.NewReader(paramStr))
	if err != nil {
		fmt.Println(err)
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if resp.StatusCode != 200 {
		//return err
	}

	fmt.Println(string(body), resp.StatusCode, err)

	return nil
}

//{"requestToJson", "cAl970TBvkEDORELBzwNXT0b++9UpOW7nOS6cwBdf09B9CyzMmFblZud4oAkaso4fNbP/bi4lUPN4PjAdfcZ6SBmvEHtcD97kaol6cNXLPTywUwdP7kA5TVt7ArcIT8cTVGApBQRzymP5Cy
// KEC1O0y+khLi92oW4KUI1Ougfi8eJ/HO3wKPnA+T8kNh4QkXcio6mPBtjnS4MeqKqoGFfYg==,1234567890123456,00959790648563"}
func ReceivePaymentCallback(rawData []byte) {
	secretKey := beego.AppConfig.String("okdollar_secret_key")

	rawStr := string(rawData)

	logs.Error("[okdollar] [ReceivePaymentCallback]" + rawStr)

	vecStr := strings.Split(rawStr, ",")
	if len(vecStr) < 3 {
		return
	}

	rawJsonData, err := tools.Base64Decode(vecStr[0])
	logs.Error("[okdollar] [ReceivePaymentCallback]" + string(rawJsonData))
	if err != nil {
		logs.Error("[okdollar] [ReceivePaymentCallback] %v", err)
		return
	}

	strOrder := vecStr[1]
	account := vecStr[2]

	detailData, err := tools.Decrypter(rawJsonData, []byte(secretKey), []byte(strOrder), tools.AesPKCS7)
	if err != nil {
		logs.Error("[okdollar] [ReceivePaymentCallback] %v", err)
		return
	}

	logs.Error("[okdollar] [ReceivePaymentCallback]" + string(detailData))

	resp := PaymentCallBackData{}
	err = json.Unmarshal(detailData, &resp)
	if err != nil {
		logs.Error("[okdollar] [ReceivePaymentCallback] %v", err)
		return
	}

	logs.Error("[okdollar] [ReceivePaymentCallback] %s %s %v", strOrder, account, resp)

	return
}
