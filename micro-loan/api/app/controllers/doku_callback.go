package controllers

import (
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"encoding/xml"
	"micro-loan/common/thirdparty/doku"

	"micro-loan/common/lib/device"

	"micro-loan/common/dao"

	"micro-loan/common/models"

	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/thirdparty"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

const INIT_AMOUNT = "0"
const MAX_AMOUNT = "100000000"
const SUCCESSCODE = "0000"
const VA_FAIL_CODE = "9999"
const VA_INVALID_NUMBER = "3000"
const INDONESIA_CURRENCY = "360"
const SUCCESS_MSG = ""

type DoKuCallbackController struct {
	beego.Controller
}

func (c *DoKuCallbackController) VirtualAccountCreate() {
	reqStr := string(c.Ctx.Input.RequestBody)
	logs.Debug("[DoKuVirtualAccountCreate] reqStr:", reqStr)
	//MALLID=5870&CHAINMERCHANT=0&PAYMENTCHANNEL=36&PAYMENTCODE=8856060712341125&STATUSTYPE=I&WORDS=509a9e227426c4f2d0e9ef71ee6830c229479d8b&OCOID=null

	/*
		<?xml version="1.0"?>
		<INQUIRY_RESPONSE>
			<PAYMENTCODE>8975011200005642</PAYMENTCODE>
			<AMOUNT>100000.00<AMOUNT>
			<PURCHASEAMOUNT>100000.00</PURCHASEAMOUNT>
			<MINAMOUNT>10000.00<MINAMOUNT>
			<MAXAMOUNT>550000.00<MAXAMOUNT>
			<TRANSIDMERCHANT>1396430482839</TRANSIDMERCHANT>
			<WORDS>b5a22f37ad0693ebac1bf03a89a8faeae9e7f390</WORDS>
			<REQUESTDATETIME>20140402162122</REQUESTDATETIME>
			<CURRENCY>360</CURRENCY>
			<PURCHASECURRENCY>360</PURCHASECURRENCY>
			<SESSIONID>dxgcmvcbywhu3t5mwye7ngqhpf8i6edu</SESSIONID>
			<NAME>Nama Lengkap</NAME>
			<EMAIL>nama@xyx.com</EMAIL>
			<BASKET>ITEM 1,10000.00,2,20000.00;ITEM 2,20000.00,4,80000.00</BASKET>
			<ADDITIONALDATA>BORNEO TOUR AND TRAVEL</ADDITIONALDATA>
			<RESPONSECODE>0000</RESPONSECODE>
		</INQUIRY_RESPONSE>
	*/

	mallId := beego.AppConfig.String("doku_mallid")
	email := beego.AppConfig.String("doku_company_email")

	paymentCode := c.GetString("PAYMENTCODE")
	words := c.GetString("WORDS")
	code := SUCCESSCODE

	hash, err := doku.CheckVAWords(mallId, paymentCode, words)
	if err != nil {
		logs.Error("[DoKu VA words mismatched. ", words, "[hash]:", hash, "reqStr:", reqStr)
		code = VA_FAIL_CODE
		//return
	}

	userEAccount, err := doku.CheckDoKuVAExist(paymentCode)
	if err != nil {
		logs.Error("DoKu VA does not exist. reqStr is:", reqStr)
		code = VA_INVALID_NUMBER
	}

	accountBase, err := dao.CustomerOne(userEAccount.UserAccountId)

	if err != nil {
		logs.Error("DoKu accountBase does not exist. userAccountId is", userEAccount.UserAccountId, ", reqStr is: ", reqStr)
		code = VA_INVALID_NUMBER
	}

	//TODO xml是否需要这么多数据，是否可以去掉一些数据

	inquiryMainContent := doku.InquiryMainContent{}
	inquiryMainContent.PaymentCode = paymentCode
	inquiryMainContent.Amount = INIT_AMOUNT //0表示用户还款金额不是固定的
	inquiryMainContent.PurchaseAmount = INIT_AMOUNT
	inquiryMainContent.MinAmount = INIT_AMOUNT
	inquiryMainContent.MaxAmount = MAX_AMOUNT
	transIdMerchant, _ := device.GenerateBizId(types.DokuTransIdMerchant)
	inquiryMainContent.TransidMerchant = tools.Int642Str(transIdMerchant) //每次都要不一样
	inquiryMainContent.Words = hash
	inquiryMainContent.RequestDateTime = tools.MDateMHSLocalDateAllNum(tools.GetUnixMillis())
	inquiryMainContent.Currency = INDONESIA_CURRENCY
	inquiryMainContent.PurchaseCurrency = INDONESIA_CURRENCY
	inquiryMainContent.SessionId = tools.Int642Str(userEAccount.UserAccountId)
	inquiryMainContent.Name = accountBase.Realname
	inquiryMainContent.Email = email
	//inquiryMainContent.Basket = "ITEM 1,10000.00,2,20000.00;ITEM 2,20000.00,4,80000.00"
	//basket := fmt.Sprintf("%d,%d", order.Id, order.Amount)
	inquiryMainContent.Basket = "ITEM 1,10000.00,2,20000.00"
	//inquiryMainContent.AdditionalData = "BORNEO TOUR AND TRAVEL"
	inquiryMainContent.AdditionalData = ""
	inquiryMainContent.ResponseCode = code

	inquiryXMl := inquiryMainContent
	xmlBody, _ := xml.Marshal(inquiryXMl)
	c.Ctx.Output.Body([]byte(xml.Header + string(xmlBody)))
}

func (c *DoKuCallbackController) DisburseFundCreate() {
	//doku 没有付款回调
}

func (c *DoKuCallbackController) FVAReceivePaymentCreate() {

	reqStr := string(c.Ctx.Input.RequestBody)
	logs.Debug("[DoKu FVAReceivePaymentCreate] reqStr:", reqStr)

	amount := c.GetString("AMOUNT")
	paymentCode := c.GetString("PAYMENTCODE")
	words := c.GetString("WORDS")
	resultMsg := c.GetString("RESULTMSG")
	transIdMerchant := c.GetString("TRANSIDMERCHANT")
	bank := c.GetString("BANK")
	verifyStatus := c.GetString("VERIFYSTATUS")
	repayLoan, _ := tools.Str2Float64(amount)
	total := int64(repayLoan)

	hash, err := doku.CheckRepayVAWords(amount, transIdMerchant, resultMsg, verifyStatus, words)
	if err != nil {
		logs.Error(err)
		return
	}

	logs.Debug("[hash]: ", hash, "[post words]: ", words)

	router := "/doku/fva_receive_payment_callback/create"

	order, err := doku.CheckValidOrder(paymentCode)
	if err != nil {
		logs.Error(err)
		return
	}

	thirdPartyId, _ := models.AddOneThirdpartyRecord(models.ThirdpartyDoKu, router, order.Id, reqStr, "", 0, 0, 200)

	if hash == words {
		if resultMsg == "SUCCESS" {
			err := service.DoKuPaymentCallback(router, total, bank, paymentCode, reqStr)
			if err != nil {
				logs.Error("DoKuPaymentCallback err: ", err)
				c.Ctx.WriteString("STOP")
			} else {
				thirdPartyData, _ := models.GetThirpartyRecordById(thirdPartyId)
				responstType, fee := thirdparty.CalcFeeByApi(router, reqStr, "")
				event.Trigger(&evtypes.CustomerStatisticEv{
					UserAccountId: order.UserAccountId,
					OrderId:       order.Id,
					ApiMd5:        tools.Md5(router),
					Fee:           int64(fee),
					Result:        responstType,
				})
				thirdPartyData.ResponseType = responstType
				thirdPartyData.FeeForCall = fee
				thirdPartyData.UpdateFee()
				c.Ctx.WriteString("CONTINUE")
			}
		} else {
			logs.Error("DoKuPaymentCallback resultMsg is wrong, the reqStr is ", reqStr)
			c.Ctx.WriteString("STOP")
		}
	} else {
		logs.Error("DoKuPaymentCallback [hash]: ", hash, "[post words]: ", words)
		logs.Error("DoKuPaymentCallback hash mismatched,  the reqStr is ", reqStr)
		c.Ctx.WriteString("STOP")
	}

	return
}

func (c *DoKuCallbackController) IdentifyCreate() {
	strJson := string(c.Ctx.Input.RequestBody)

	logs.Error("[DoKuIdentifyCreate] json:", strJson)

}
