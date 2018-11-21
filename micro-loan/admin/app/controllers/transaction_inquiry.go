package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils/pagination"

	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/doku"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

const (
	PAYMENT_CALLBACK_ID                          = "mobi_manual_id"
	PAYMENT_CALLBACK_PAYMENT_ID                  = "mobi_payment_id"
	PAYMENT_CALLBACK_CALLBACK_VIRTUAL_ACCOUNT_ID = "mobi_manual_callback_virtual_account_id"
	PAYMENT_CALLBACK_OWNER_ID                    = "mobi_manual_owner_id"
)

type TransactionInquiryController struct {
	BaseController
}

func (c *TransactionInquiryController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "transaction_inquiry"
}

func (c *TransactionInquiryController) DisburseInquiryForm() {
	//c.Data["Action"] = "disburse_inquiry_form"
	c.Layout = "layout.html"
	c.TplName = "transaction_inquiry/disburse_inquiry_form.html"
}

func (c *TransactionInquiryController) DisburseInquiryResult() {
	c.Data["Action"] = "disburse_inquiry_result"
	action := "/transaction_inquiry/disburse_inquiry_result"
	condCntr := map[string]interface{}{}

	accountId, _ := c.GetInt64("account_id")
	if accountId > 0 {
		condCntr["field"] = accountId
	} else {
		c.commonError(action, "/transaction_inquiry/disburse_inquiry_form", "缺少必要参数")
		return
	}

	inquiryUrl := beego.AppConfig.String("xendit_disburse_inquiry")
	secretKey := beego.AppConfig.String("secret_key")
	inquiryUrl = fmt.Sprintf("%s%d", inquiryUrl, accountId)

	auth := tools.BasicAuth(secretKey, "")
	reqHeaders := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + auth,
	}

	var inquiryResp []struct {
		UserId                  string `json:"user_id"`
		ExternalId              string `json:"external_id"`
		Amount                  int64  `json:"amount"`
		BankCode                string `json:"bank_code"`
		AccountHolderName       string `json:"account_holder_name"`
		DisbursementDescription string `json:"disbursement_description"`
		IsInstant               bool   `json:"is_instant"`
		Status                  string `json:"status"`
		Id                      string `json:"id"`
		ErrorCode               string `json:"error_code"`
		Message                 string `json:"message"`
	}

	httpBody, httpCode, err := tools.SimpleHttpClient("GET", inquiryUrl, reqHeaders, "", tools.DefaultHttpTimeout())
	logs.Debug(string(httpBody))

	pageSize := 15
	page, _ := tools.Str2Int(c.GetString("p"))
	if page < 1 {
		page = 1
	}

	if err != nil {
		c.Data["err"] = err.Error()
	} else {
		err = json.Unmarshal(httpBody, &inquiryResp)
		if err != nil {
			c.Data["err"] = err.Error()
		} else {
			if httpCode != 200 {
				c.Data["httpCode"] = httpCode
			} else {
				len := len(inquiryResp)
				s := (page - 1) * pageSize
				e := s + pageSize
				if e > len {
					e = len
				}
				list := inquiryResp[s:e]
				paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(len))
				c.Data["List"] = list
				c.Data["paginator"] = paginator
			}
		}
	}

	c.Layout = "layout.html"
	c.TplName = "transaction_inquiry/disburse_inquiry_result.html"
}

func (c *TransactionInquiryController) PaymentInquiryForm() {
	//c.Data["Action"] = "disburse_inquiry_form"

	action := "/transaction_inquiry/payment_inquiry_form"

	order_id, _ := c.GetInt64("order_id")
	c.Data["order"] = models.Order{}

	amount, _ := c.GetInt64("amount")
	c.Data["amount"] = amount

	if order_id > 0 {
		order, err := models.GetOrder(order_id)
		c.Data["order"] = order
		if err != nil {
			c.commonError(action, "/transaction_inquiry/payment_inquiry_form", "订单id不存在")
			return
		}
		eAccount, err := dao.GetActiveEaccountWithBankName(order.UserAccountId)
		if err != nil {
			c.commonError(action, "/transaction_inquiry/payment_inquiry_form", "user_e_account不存在")
			return
		}
		c.Data["eAccount"] = eAccount

		if amount <= 0 {
			c.commonError(action, "/transaction_inquiry/payment_inquiry_form", "金额输入有误")
			return
		}

		repayPlan, err := models.GetLastRepayPlanByOrderid(order_id)
		if err != nil {
			c.commonError(action, "/transaction_inquiry/payment_inquiry_form", "还款计划不存在")
			return
		}
		leftOver, err := repayplan.CaculateRepayTotalAmountWithPreReducedByRepayPlan(repayPlan)
		if err != nil {
			c.commonError(action, "/transaction_inquiry/payment_inquiry_form", "amount < 0")
			return
		}

		if amount > leftOver {
			msg := fmt.Sprintf("%s%d", "补单金额不能超过剩余还款金额, 剩余还款金额为:", leftOver)
			c.commonError(action, "/transaction_inquiry/payment_inquiry_form", msg)
			return
		}

	}

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "transaction_inquiry/list_scripts.html"
	c.Layout = "layout.html"
	c.TplName = "transaction_inquiry/payment_inquiry_form.html"

}

func (c *TransactionInquiryController) PaymentInquiryResult() {
	c.Data["Action"] = "payment_inquiry_result"
	resp := map[string]interface{}{}

	order_id, _ := c.GetInt64("order_id")
	if order_id <= 0 {
		resp["status"] = 40001
		resp["msg"] = "order_id is empty."
		c.Data["json"] = resp
		c.ServeJSON()
	}

	amount, _ := c.GetInt64("amount")
	if amount <= 0 {
		resp["status"] = 40002
		resp["msg"] = "the number of money is wrong."
		c.Data["json"] = resp
		c.ServeJSON()
	}

	order, err := models.GetOrder(order_id)
	c.Data["order"] = order
	if err != nil {
		resp["status"] = 40003
		msg := fmt.Sprintf("%s%d%s", "orderId is ", order_id, " but order does not exist.")
		resp["msg"] = msg
		c.Data["json"] = resp
		c.ServeJSON()
	}

	eAccount, err := dao.GetActiveEaccountWithBankName(order.UserAccountId)
	if err != nil {
		resp["status"] = 40004
		msg := fmt.Sprintf("%s%d%s", "orderId is ", order_id, " but userEaccount does not exist.")
		resp["msg"] = msg
		c.Data["json"] = resp
		c.ServeJSON()
	}

	originRepayPlan, err := models.GetLastRepayPlanByOrderid(order_id)
	if err != nil {
		resp["status"] = 40005
		msg := fmt.Sprintf("%s%d%s", "orderId is ", order_id, " but repayPlan does not exist.")
		resp["msg"] = msg
		c.Data["json"] = resp
		c.ServeJSON()
	}

	if order.CheckStatus != types.LoanStatusWaitRepayment &&
		order.CheckStatus != types.LoanStatusOverdue &&
		order.CheckStatus != types.LoanStatusPartialRepayment &&
		order.CheckStatus != types.LoanStatusRolling {

		statusDesc := service.GetLoanStatusDesc("zh-CN", order.CheckStatus)
		msg := fmt.Sprintf("%s%s%s", " 订单状态为:", statusDesc, ",无法补单.")

		resp["status"] = 40006
		msg = fmt.Sprintf("%s%d%s", "orderId is ", order_id, msg)
		resp["msg"] = msg
		c.Data["json"] = resp
		c.ServeJSON()
	}

	inquiryUrl := beego.AppConfig.String("xendit_payment_callback_url")
	secretKey := beego.AppConfig.String("secret_key")

	auth := tools.BasicAuth(secretKey, "")
	reqHeaders := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Basic " + auth,
	}

	var req struct {
		Id                       string `json:"id"`
		PaymentId                string `json:"payment_id"`
		CallbackVirtualAccountId string `json:"callback_virtual_account_id"`
		OwnerId                  string `json:"owner_id"`
		ExternalId               string `json:"external_id"`
		AccountNumber            string `json:"account_number"`
		BankCode                 string `json:"bank_code"`
		Amount                   int64  `json:"amount"`
		TransactionTimestamp     string `json:"transaction_timestamp"`
		MerchantCode             string `json:"merchant_code"`
		Updated                  string `json:"updated"`
		Created                  string `json:"created"`
	}

	req.Id = PAYMENT_CALLBACK_ID
	req.CallbackVirtualAccountId = PAYMENT_CALLBACK_CALLBACK_VIRTUAL_ACCOUNT_ID
	req.PaymentId = fmt.Sprintf("%s%d", PAYMENT_CALLBACK_PAYMENT_ID, order_id)
	req.OwnerId = PAYMENT_CALLBACK_OWNER_ID
	req.ExternalId = tools.Int642Str(order.UserAccountId)
	req.AccountNumber = eAccount.EAccountNumber
	req.BankCode = eAccount.BankCode
	req.Amount = amount
	now := tools.MDateMHS(tools.GetUnixMillis())
	req.TransactionTimestamp = now
	req.Updated = now
	req.Created = now

	reqB, err := json.Marshal(req)
	logs.Debug(string(reqB))
	//reqBody
	//inquiryUrl = "http://localhost:8700/xendit/fva_receive_payment_callback/create"
	//logs.Debug(string(inquiryUrl))
	_, httpCode, err := tools.SimpleHttpClient("POST", inquiryUrl, reqHeaders, string(reqB), tools.DefaultHttpTimeout())

	logs.Debug(httpCode)

	//c.Data["json"] = response
	if err != nil || httpCode != 200 {
		resp["status"] = 40007
		resp["msg"] = "http request failed"
		c.Data["json"] = resp
		c.ServeJSON()
	}

	repayPlan, _ := models.GetLastRepayPlanByOrderid(order_id)

	models.OpLogWrite(c.AdminUid, order.Id, models.OpCodeSupplementOrder, originRepayPlan.TableName(), originRepayPlan, repayPlan)

	resp["status"] = 0
	resp["msg"] = "succeeded."
	c.Data["json"] = resp

	c.ServeJSON()

}

func (c *TransactionInquiryController) GenerateFixDoku() {
	dokuRemitResp := doku.DokuRemitResp{}

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "transaction_inquiry/list_scripts.html"
	c.Layout = "layout.html"
	c.TplName = "transaction_inquiry/generate_data_doku.html"

	// 1、获得参数
	order_id, _ := c.GetInt64("order_id")
	record_id, _ := c.GetInt("thirdparty_record_id")
	inquiryId := c.GetString("inquiry_id")
	transactionId := c.GetString("transaction_id")
	tableNameCode, _ := c.GetInt("table_name")
	order, err := models.GetOrder(order_id)
	tableName := types.TableNameMap()[tableNameCode]
	defer func() {
		c.Data["OrderId"] = order_id
		c.Data["InquiryId"] = inquiryId
		c.Data["TransactionId"] = transactionId
		c.Data["TableName"] = tableNameCode
		c.Data["ThirdpartyRecordId"] = record_id
		c.Data["TableNameMap"] = types.TableNameMap()
	}()

	//2\ 校验参数
	if order_id <= 0 ||
		order.Id == 0 ||
		len(inquiryId) == 0 ||
		len(transactionId) == 0 ||
		tableNameCode == 0 ||
		record_id == 0 ||
		len(tableName) == 0 {

		logs.Error("[GenerateFixDoku] order_id:%d  err:%v len(inquiryId):%d len(transactionId):%d",
			order_id, err, len(inquiryId), len(transactionId))
		c.Data["Resp"] = "数据错误"
		return
	}

	dokuRemitResp.Status = 0
	dokuRemitResp.Message = "Remit accepted"
	dokuRemitResp.Remit.TransactionId = transactionId
	dokuRemitResp.Remit.PaymentData.MallId = "2"
	dokuRemitResp.Remit.PaymentData.AccountNumber = "0000000899"
	dokuRemitResp.Remit.PaymentData.AccountName = "DOKU"
	dokuRemitResp.Remit.PaymentData.ChannelCode = "07"
	dokuRemitResp.Remit.PaymentData.InquiryId = inquiryId
	dokuRemitResp.Remit.PaymentData.Currency = "IDR"
	dokuRemitResp.Remit.PaymentData.Amount = tools.Int642Str(order.Loan)
	dokuRemitResp.Remit.PaymentData.TrxCode = "mobi_fix"
	dokuRemitResp.Remit.PaymentData.ResponseCode = "00"
	dokuRemitResp.Remit.PaymentData.ResponseMsg = "Transfer Approve"

	err = service.FixThirdParty(dokuRemitResp, order_id, tableName, record_id, c.AdminUid)

	ss, _ := json.Marshal(dokuRemitResp)
	c.Data["Resp"] = string(ss)
	if err != nil {
		logs.Error("[GenerateFixDoku] FixThirdParty err:%v", err)
		c.Data["Resp"] = err.Error()
	}
	return
}
