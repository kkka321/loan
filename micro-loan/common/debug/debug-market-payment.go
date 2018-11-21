package main

import (
	//"encoding/json"

	"encoding/json"
	"fmt"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

func main() {

	/*
		accountId := int64(1234) // accountId
		externalId := int64(123) // orderId
		amount := int64(50000)

		invoiceUrl := beego.AppConfig.String("xendit_invoice")
		secretKey := beego.AppConfig.String("secret_key")

		paramStr := fmt.Sprintf("%s%d%s%s%d%s%d",
			"external_id=", externalId,
			"&payer_email=rupiahcepat@gmail.com",
			"&description=", accountId,
			"&amount=", amount,
		)

		auth := tools.BasicAuth(secretKey, "")
		reqHeaders := map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"Authorization": "Basic " + auth,
		}

		httpBody, httpCode, err := tools.SimpleHttpClient("POST", invoiceUrl, reqHeaders, paramStr, tools.DefaultHttpTimeout())
		logs.Debug(string(httpBody))
		var invoiceResp struct {
			Id                        string `json:"id"`
			ExternalId                string `json:"external_id"`
			UserId                    string `json:"user_id"`
			Status                    string `json:"status"`
			MerchantName              string `json:"merchant_name"`
			MerchantProfilePictureUrl string `json:"merchant_profile_picture_url"`
			Amount                    int64  `json:"amount"`
			PayerEmail                string `json:"payer_email"`
			Description               string `json:"description"`
			ExpiryDate                string `json:"expiry_date"`
			InvoiceUrl                string `json:"invoice_url"`
			AvailableBanks            []struct {
				BankCode          string `json:"bank_code"`
				CollectionType    string `json:"collection_type"`
				BankAccountNumber string `json:"bank_account_number"`
				TransferAmount    int64  `json:"transfer_amount"`
				BankBranch        string `json:"bank_branch"`
				AccountHolderName string `json:"account_holder_name"`
				IdentityAmount    int64  `json:"identity_amount"`
			} `json:"available_banks"`

			AvailableRetailOutlets []struct {
				RetailOutletName string `json:"retail_outlet_name"`
				PaymentCode      string `json:"payment_code"`
				TransferAmount   int64  `json:"transfer_amount"`
			} `json:"available_retail_outlets"`

			ShouldExcludeCreditCard bool   `json:"should_exclude_credit_card"`
			ShouldSendEmail         bool   `json:"should_send_email"`
			Created                 string `json:"created"`
			Updated                 string `json:"updated"`
		}

		err = json.Unmarshal(httpBody, &invoiceResp)
		if err != nil {
			logs.Error("[Xendit Invoice Create Response] json.Unmarshal err:%s, json:%s", err.Error(), string(httpBody))
		}

		marketPayment := &models.MarketPayment{}
		marketPayment.UserAccountId = accountId
		marketPayment.OrderId = externalId
		marketPayment.PaymentCode = invoiceResp.AvailableRetailOutlets[0].PaymentCode
		marketPayment.Status = invoiceResp.Status
		marketPayment.Amount = amount
		marketPayment.Response = string(httpBody)
		marketPayment.Ctime = tools.GetUnixMillis()
		marketPayment.Utime = tools.GetUnixMillis()
		models.AddMarketPayment(marketPayment)

		monitor.IncrThirdpartyCount(models.ThirdpartyXendit, httpCode)
	*/
	/*
		//TODO 放在回调做
		responstType, fee := thirdparty.CalcFeeByApi(invoiceUrl, paramStr, string(httpBody))
		models.AddOneThirdpartyRecord(models.ThirdpartyXendit, invoiceUrl, externalId, paramStr, string(httpBody), responstType, fee)
		event.Trigger(&evtypes.CustomerStatisticEv{
			UserAccountId: accountId,
			OrderId:       externalId,
			ApiMd5:        tools.Md5(invoiceUrl),
			Fee:           int64(fee),
			Result:        responstType,
		})
	*/

	//return body, err

	/*
		data, _ := models.GetMarketPaymentByOrderId(180821020000033584)
		var invoiceResp models.InvoiceResp
		json.Unmarshal([]byte(data.Response), &invoiceResp)
		logs.Debug(invoiceResp)
		data, _ = models.GetMarketPaymentByOrderId(180822020000014714)
		json.Unmarshal([]byte(data.Response), &invoiceResp)
		logs.Debug(invoiceResp)
	*/

	expire_url := beego.AppConfig.String("xendit_paymentcode_expire")
	secretKey := beego.AppConfig.String("secret_key")
	auth := tools.BasicAuth(secretKey, "")
	reqHeaders := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + auth,
	}

	data, _ := models.GetMarketPaymentsByOrderId(180918020000030029)
	for _, obj := range data {
		var invoiceResp models.InvoiceResp
		json.Unmarshal([]byte(obj.Response), &invoiceResp)
		if invoiceResp.Id != "" {
			expire_paymentcode_url := strings.Replace(expire_url, "{invoice_id}", invoiceResp.Id, -1)
			httpBody, _, err := tools.SimpleHttpClient("POST", expire_paymentcode_url, reqHeaders, "", tools.DefaultHttpTimeout())
			if err != nil {
				logs.Error("expire paymentcode error json:[%s], market_payment[%#v]", string(httpBody), obj)
				err = fmt.Errorf(string(httpBody))
				return
			}
			json.Unmarshal(httpBody, &invoiceResp)
			if invoiceResp.AvailableRetailOutlets[0].PaymentCode != "" {
				obj.Status = invoiceResp.Status
				models.UpdateMarketPayment(&obj)
			}
		}
	}
}
