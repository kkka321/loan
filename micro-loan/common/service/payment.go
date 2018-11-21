package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/dao"
	"micro-loan/common/lib/payment"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/schema_task"
	"micro-loan/common/thirdparty"
	"micro-loan/common/thirdparty/bluepay"
	"micro-loan/common/thirdparty/doku"
	"micro-loan/common/thirdparty/xendit"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func CreatePaymentApi(companyId int, datas map[string]interface{}) (payment.PaymentInterface, error) {
	var pApi payment.PaymentInterface
	companyName := datas["company_name"].(string)
	var err error
	switch companyId {
	case types.DoKu:
		api := new(doku.DokuApi)
		api.CompanyId = companyId
		api.CompanyName = companyName
		api.HandleDisburseCallback = HandleDisburse
		api.HandleLoanFailCallback = UpdateOrderToLoanFail
		pApi = api
	case types.Xendit:
		api := new(xendit.XenditApi)
		api.CompanyId = companyId
		api.CompanyName = companyName
		pApi = api
	case types.Bluepay:
		api := new(bluepay.BluepayApi)
		api.CompanyId = companyId
		api.CompanyName = companyName
		pApi = api
	default:
		pApi = new(payment.PaymentApi)
		logs.Error("CreatePaymentApi unexcept companyid:", companyId)
		err = fmt.Errorf("unexcept companyid")
	}

	return pApi, err
}

func XenditCreateVirtualAccountCallback(router string, jsonData []byte) error {

	var accountId int64
	err := xendit.CreateVirtualAccountCallback(jsonData, &accountId)
	responstType, fee := thirdparty.CalcFeeByApi(router, string(jsonData), "")
	models.AddOneThirdpartyRecord(models.ThirdpartyXendit, router, accountId, string(jsonData), "", responstType, fee, 200)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: accountId,
		OrderId:       0,
		ApiMd5:        tools.Md5(router),
		Fee:           int64(fee),
		Result:        responstType,
	})
	return err
}

func XenditDisburseCallback(router string, jsonData []byte) error {
	var err error

	var accountId int64
	var bankCode string
	var status types.LoanStatus
	var tranData models.Mobi_E_Trans
	var isMatch bool
	var callBackOrderId int64
	var amount int64
	tranData, err = xendit.DisburseCallback(jsonData, &accountId, &bankCode, &status, &isMatch, &callBackOrderId, &amount)
	if err != nil {
		return err
	}

	// 如果是退款订单 去做退款处理
	if callBackOrderId > 0 && thirdparty.IsValiedId(callBackOrderId, int(types.RefundBiz)) {
		return XenditDisburseRefundCallback(router, jsonData, accountId, callBackOrderId, amount, bankCode)
	}

	//order, err := dao.AccountLastLoanOrder(accountId)
	order, err := models.GetOrder(callBackOrderId)
	logs.Info("[XenditDisburseCallback] callBackOrderId:%d", callBackOrderId)
	if err != nil {
		logs.Error("[XenditDisburseCallback] order nil err:%s, accountId:%d callBackOrderId:%d", err, accountId, callBackOrderId)
		return err
	}

	responstType, fee := 0, 0
	recordId, _ := models.AddOneThirdpartyRecord(models.ThirdpartyXendit, router, order.Id, string(jsonData), "", responstType, fee, 200)

	if order.CheckStatus != types.LoanStatusIsDoing &&
		order.CheckStatus != types.LoanStatusWait4Loan &&
		order.CheckStatus != types.LoanStatusLoanFail {
		err := fmt.Errorf("[XenditDisburseCallback] status error status:%d, orderid:%d", int(order.CheckStatus), order.Id)

		return err
	}

	if tranData.Status == "FAILED" {

		tranData.UpdateMobiEEtrans(&tranData)

		UpdateOrderToLoanFail(&order, err)

		return nil
	}

	if status == types.LoanStatusLoanFail {
		err := fmt.Errorf("[XenditDisburseCallback] status error, status:%d, orderid:%d", status, order.Id)

		return err
	}

	if !isMatch {
		err := fmt.Errorf("[XenditDisburseCallback] data not matched orderid:%d", order.Id)

		return err
	}

	tranData.UpdateMobiEEtrans(&tranData)

	err = HandleDisburse(types.Xendit, &order, bankCode, false)

	if err != nil {
		logs.Error("[XenditDisburseCallback] status error err:%s, orderid:%d", err, order.Id)

		return err
	}

	// 回调成功 更新记录
	responstType, fee = thirdparty.CalcFeeByApi(router, string(jsonData), "")
	record := models.ThirdpartyRecord{
		Id:           recordId,
		ResponseType: responstType,
		FeeForCall:   fee,
	}
	record.UpdateFee()
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: accountId,
		OrderId:       order.Id,
		ApiMd5:        tools.Md5(router),
		Fee:           int64(fee),
		Result:        responstType,
	})

	return err
}

func XenditDisburseRefundCallback(router string, jsonData []byte, accountId, callBackOrderId, amount int64, bankCode string) error {

	err := RefundDisburseCallback(accountId, callBackOrderId, amount, bankCode, jsonData)
	if err != nil {
		// 退款订单置为无效
		refund, errN := dao.GetRefund(callBackOrderId)
		if errN == nil && refund.CheckStatus == int(types.RefundStatusProcessing) {
			errI := SetRefundInvalid(&refund)
			// 冻结资金恢复
			if errI == nil {
				RestoreBalanceByRefund(&refund)
			} else {
				logs.Error("[XenditDisburseRefundCallback] SetRefundInvalid err:%v refund:%#v", err, refund)
			}
		}

		models.AddOneThirdpartyRecord(models.ThirdpartyXendit, router, callBackOrderId, string(jsonData), "", 0, 0, 200)
		return err
	}

	responstType, fee := thirdparty.CalcFeeByApi(router, string(jsonData), "")
	models.AddOneThirdpartyRecord(models.ThirdpartyXendit, router, callBackOrderId, string(jsonData), "", responstType, fee, 200)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: accountId,
		OrderId:       0,
		ApiMd5:        tools.Md5(router),
		Fee:           int64(fee),
		Result:        responstType,
	})

	accountBase, _ := dao.CustomerOne(accountId)
	//content := i18n.GetMessageText(i18n.TextSmsRefundDisburseSuccess)
	// 发短信新统一入口
	//sms.Send(types.ServiceDisburseSuccess, accountBase.Mobile, content, callBackOrderId)

	param := make(map[string]interface{})
	param["related_id"] = callBackOrderId
	schema_task.SendBusinessMsg(types.SmsTargetRefundDisburseSuccess, types.ServiceDisburseSuccess, accountBase.Mobile, param)

	return nil
}

func XenditPaymentCallback(router string, jsonData []byte) error {
	var err error

	var accountId int64
	var amount int64
	var bankCode string

	err, resp := xendit.ReceivePaymentCallback(jsonData, &accountId, &amount, &bankCode)
	if err != nil {
		return err
	}
	if isHandled(resp.PaymentId) {
		str := fmt.Sprintf("[XenditPaymentCallback] this payment has been handled.  jsonData:%s", string(jsonData))
		logs.Error(str)
		return errors.New(str)
	}

	callbackStr := string(jsonData)
	orderId, err := RepayLoan(accountId, amount, bankCode, "", types.Xendit, callbackStr)
	if err != nil {
		models.AddOneThirdpartyRecord(models.ThirdpartyXendit, router, orderId, callbackStr, "", 0, 0, 200)
		return err
	}

	// 保存处理记录
	if err == nil {
		pay := models.PaymentReceiveCallLog{
			UserAccountId: accountId,
			OrderId:       orderId,
			PaymentId:     resp.PaymentId,
			Ctime:         tools.GetUnixMillis(),
		}
		models.OrmInsert(&pay)
	}

	responstType, fee := thirdparty.CalcFeeByApi(router, callbackStr, "")
	models.AddOneThirdpartyRecord(models.ThirdpartyXendit, router, orderId, callbackStr, "", responstType, fee, 200)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: accountId,
		OrderId:       orderId,
		ApiMd5:        tools.Md5(router),
		Fee:           int64(fee),
		Result:        responstType,
	})

	return err
}

func isHandled(paymentId string) bool {
	recvLog := models.PaymentReceiveCallLog{}
	o := orm.NewOrm()
	o.Using(recvLog.Using())

	err := o.QueryTable(recvLog.TableName()).Filter("payment_id", paymentId).One(&recvLog)

	if err == orm.ErrNoRows {
		return false
	} else {
		return true
	}

}

func XenditMarketPaymentCallback(router string, jsonStr []byte) (err error) {
	strJson := string(jsonStr)
	var invoiceCallbackJson struct {
		Id                     string `json:"id"`
		UserId                 string `json:"user_id"`
		ExternalId             string `json:"external_id"`
		IsHigh                 bool   `json:"is_high"`
		MerchantName           string `json:"merchant_name"`
		Amount                 int64  `json:"amount"`
		FeesPaidAmount         int64  `json:"fees_paid_amount"`
		Status                 string `json:"status"`
		PayerMail              string `json:"payer_mail"`
		Description            string `json:"description"`
		AdjustedReceivedAmount int64  `json:"adjusted_received_amount"`
		PaymentMethod          string `json:"payment_method"`
		BankCode               string `json:"bank_code"`
		PaidAmount             int64  `json:"paid_amount"`
		Updated                string `json:"update"`
		Created                string `json:"created"`
	}

	err = json.Unmarshal(jsonStr, &invoiceCallbackJson)
	if err != nil {
		logs.Error("[XenditMarketReceivePaymentCreate callback Response] json.Unmarshal err:%s, json:%s", err.Error(), strJson)
		return
	}

	if isHandled(invoiceCallbackJson.Id) {
		str := fmt.Sprintf("[XenditMarketPaymentCallback] this payment has been handled.  jsonData:%s", string(jsonStr))
		logs.Error(str)
		return errors.New(str)
	}

	orderId, _ := tools.Str2Int64(invoiceCallbackJson.ExternalId)
	marketPayment, err := models.GetMarketPaymentByOrderId(orderId)
	if err != nil {
		logs.Error(err)
		return
	}

	marketPayment.Status = invoiceCallbackJson.Status
	marketPayment.CallbackJson = strJson
	marketPayment.PaidTime = tools.GetUnixMillis()
	marketPayment.Utime = tools.GetUnixMillis()
	models.UpdateMarketPayment(&marketPayment)

	orderId, err = RepayLoan(marketPayment.UserAccountId, invoiceCallbackJson.Amount, types.XenditMarketPaymentBankCode, marketPayment.PaymentCode, types.Xendit, strJson)
	if nil != err {
		models.AddOneThirdpartyRecord(models.ThirdpartyXendit, router, orderId, strJson, "", 0, 0, 200)
		return
	}

	// 保存处理记录
	if err == nil {
		pay := models.PaymentReceiveCallLog{
			UserAccountId: marketPayment.UserAccountId,
			OrderId:       orderId,
			PaymentId:     invoiceCallbackJson.Id,
			Ctime:         tools.GetUnixMillis(),
		}
		models.OrmInsert(&pay)
	}

	responstType, fee := thirdparty.CalcFeeByApi(router, strJson, "")
	models.AddOneThirdpartyRecord(models.ThirdpartyXendit, router, orderId, strJson, "", responstType, fee, 200)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: marketPayment.UserAccountId,
		OrderId:       orderId,
		ApiMd5:        tools.Md5(router),
		Fee:           int64(fee),
		Result:        responstType,
	})

	return

}

func XenditFixPaymentCodeCallback(router string, jsonStr []byte) (err error) {
	strJson := string(jsonStr)
	var fixPaymentCodeCallbackJson struct {
		FixedPaymentCodePaymentId string `json:"fixed_payment_code_payment_id"`
		OwnerId                   string `json:"owner_id"`
		FixedPaymentCodeId        string `json:"fixed_payment_code_id"`
		PaymentId                 string `json:"payment_id"`
		ExternalId                string `json:"external_id"`
		PaymentCode               string `json:"payment_code"`
		Prefix                    string `json:"prefix"`
		RetailOutletName          string `json:"retail_outlet_name"`
		Amount                    int64  `json:"amount"`
		name                      string `json:"name"`
		transaction_timestamp     string `json:"transaction_timestamp"`
		Updated                   string `json:"update"`
		Created                   string `json:"created"`
	}

	err = json.Unmarshal(jsonStr, &fixPaymentCodeCallbackJson)
	if err != nil {
		logs.Error("[XenditFixPaymentCodeCallback callback] json.Unmarshal err:%s, json:%s", err.Error(), strJson)
		return
	}

	if isHandled(fixPaymentCodeCallbackJson.FixedPaymentCodePaymentId) {
		str := fmt.Sprintf("[XenditFixPaymentCodeCallback] this payment has been handled.  jsonData:%s", string(jsonStr))
		logs.Error(str)
		return errors.New(str)
	}

	userAccountId, _ := tools.Str2Int64(fixPaymentCodeCallbackJson.ExternalId)
	orderId, err := RepayLoan(userAccountId, fixPaymentCodeCallbackJson.Amount, types.XenditFixPaymentCode, fixPaymentCodeCallbackJson.PaymentCode, types.Xendit, strJson)
	if nil != err {
		models.AddOneThirdpartyRecord(models.ThirdpartyXendit, router, orderId, strJson, "", 0, 0, 200)
		return
	}

	// 保存处理记录
	if err == nil {
		pay := models.PaymentReceiveCallLog{
			UserAccountId: userAccountId,
			OrderId:       orderId,
			PaymentId:     fixPaymentCodeCallbackJson.FixedPaymentCodePaymentId,
			Ctime:         tools.GetUnixMillis(),
		}
		models.OrmInsert(&pay)
	}

	responstType, fee := thirdparty.CalcFeeByApi(router, strJson, "")
	models.AddOneThirdpartyRecord(models.ThirdpartyXendit, router, orderId, strJson, "", responstType, fee, 200)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: userAccountId,
		OrderId:       orderId,
		ApiMd5:        tools.Md5(router),
		Fee:           int64(fee),
		Result:        responstType,
	})

	return

}

func BluepayCreateVirtualAccountCallback(orderId int64, router string, rawQuery string) error {
	order, err := models.GetOrder(orderId)
	if err != nil {
		logs.Error("[BluepayCreateVirtualAccountCallback] orderid:%d, url:%s, err:%s", orderId, rawQuery, err.Error())
		return err
	}

	responstType, fee := thirdparty.CalcFeeByApi(router, rawQuery, "")
	models.AddOneThirdpartyRecord(models.ThirdpartyBluepay, router, order.UserAccountId, rawQuery, "", responstType, fee, 200)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: order.UserAccountId,
		OrderId:       orderId,
		ApiMd5:        tools.Md5(router),
		Fee:           int64(fee),
		Result:        responstType,
	})

	eAccount, err := models.GetEAccount(order.UserAccountId, types.Bluepay)
	if err != nil {
		logs.Error("[BluepayCreateVirtualAccountCallback] orderid:%d, accountid:%d, err:%s", orderId, order.UserAccountId, err.Error())
		return err
	}

	eAccount.Status = "ACTIVE"
	eAccount.CallbackJson = rawQuery
	eAccount.UpdateEAccount(&eAccount)

	return err
}

func BluepayDisburseCallback(orderId int64, bankCode string, status string, router string, url string) error {
	order, err := models.GetOrder(orderId)
	if err != nil {
		logs.Error("[BluepayDisburseCallback] err:%s, url:", err, url)
		return err
	}

	responstType, fee := thirdparty.CalcFeeByApi(router, url, "")
	models.AddOneThirdpartyRecord(models.ThirdpartyBluepay, router, order.Id, url, "", responstType, fee, 200)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: order.UserAccountId,
		OrderId:       orderId,
		ApiMd5:        tools.Md5(router),
		Fee:           int64(fee),
		Result:        responstType,
	})

	if order.CheckStatus != types.LoanStatusIsDoing && order.CheckStatus != types.LoanStatusWait4Loan {
		err := fmt.Errorf("[BluepayDisburseCallback] status error status:%d, orderid:%d", order.CheckStatus, order.Id)
		UpdateOrderToLoanFail(&order, err)

		return err
	}

	orderIdStr := tools.Int642Str(orderId)
	mobileEtrans, err := models.GetMobiEtrans(orderIdStr)
	if err != nil {
		logs.Error("[BluepayDisburseCallback] err:%s, url:%s, orderid:%s", err, url, orderIdStr)
		UpdateOrderToLoanFail(&order, err)
		return err
	}

	mobileEtrans.Status = status
	mobileEtrans.CallbackJson = url
	mobileEtrans.UpdateMobiEEtrans(mobileEtrans)

	err = HandleDisburse(types.Bluepay, &order, bankCode, false)

	if err != nil {
		logs.Error("[BluepayDisburseCallback] err:%s, url:%s:", err, url)

		UpdateOrderToLoanFail(&order, err)

		return err
	}

	return nil
}

func BluepayPaymentCallback(eAccountNumber string, amount int64, router string, rawQuery string) error {
	//bluepay还款时,paytype为eAccountNumber
	userEAccount, err := models.GetEAccountByENumber(eAccountNumber)
	if err != nil {
		logs.Error("[BluepayPaymentCallback] err:%s, url:%s", err, rawQuery)
		return err
	}

	accountProfile, err := dao.CustomerProfile(userEAccount.UserAccountId)
	if err != nil {
		logs.Error("[BluepayPaymentCallback] accountProfile does not exsit accountid:%d, url:%s", userEAccount.UserAccountId, rawQuery)
		return err
	}

	order, err := dao.AccountLastLoanOrder(userEAccount.UserAccountId)
	if err != nil {
		logs.Error("[BluepayPaymentCallback] order does not exsit accountid:%s, url:%s", userEAccount.UserAccountId, rawQuery)
		return err
	}

	responstType, fee := thirdparty.CalcFeeByApi(router, rawQuery, "")
	models.AddOneThirdpartyRecord(models.ThirdpartyBluepay, router, order.Id, rawQuery, "", responstType, fee, 200)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: order.UserAccountId,
		OrderId:       order.Id,
		ApiMd5:        tools.Md5(router),
		Fee:           int64(fee),
		Result:        responstType,
	})

	bankCode, _ := bluepay.BluepayBankName2Code(accountProfile.BankName)
	_, err = RepayLoan(order.UserAccountId, amount, bankCode, "", types.Bluepay, rawQuery)
	if err != nil {
		logs.Error("[BluepayPaymentCallback] err:%s, url:%s", err, rawQuery)
		return err
	}

	return err
}

func DoKuPaymentCallback(router string, repayLoan int64, bank string, paymentCode string, reqStr string) (err error) {

	order, err := doku.CheckValidOrder(paymentCode)

	if err != nil {
		return err
	}

	_, err = RepayLoan(order.UserAccountId, repayLoan, bank, "", types.DoKu, reqStr)

	return err
}
