package controllers

import (
	"fmt"
	"micro-loan/common/cerror"
	pt "micro-loan/common/lib/product"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/service"
	"micro-loan/common/strategy/limit"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	//"github.com/astaxie/beego/logs"
	"micro-loan/common/dao"

	"micro-loan/common/thirdparty/xendit"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

type LoanOrderController struct {
	ApiBaseController
}

func (c *LoanOrderController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()

	// 统一将 ip 加到 RequestJSON 中
	c.RequestJSON["ip"] = c.Ctx.Input.IP()
	c.RequestJSON["related_id"] = int64(0)
}

func (c *LoanOrderController) RepeatLoanAuthCode() {
	// 查看是否可以复贷
	if !dao.IsRepeatLoan(c.AccountID) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.MismatchRepeatLoan, "")
		c.ServeJSON()
		return
	}

	accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)

	serviceType := types.ServiceRepeatedLoan
	authCodeType := types.AuthCodeTypeText
	// 过限制策略
	if limit.MobileStrategy(accountBase.Mobile, serviceType, authCodeType) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LimitStrategyMobile, "")
		c.ServeJSON()
		return
	}

	// 调用短信服务
	if !service.SendSms(serviceType, authCodeType, accountBase.Mobile, c.Ctx.Input.IP()) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.SMSServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) RepeatLoanAuthCodeV2() {
	// 查看是否可以复贷
	if !dao.IsRepeatLoan(c.AccountID) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.MismatchRepeatLoan, "")
		c.ServeJSON()
		return
	}

	accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)

	serviceType := types.ServiceRepeatedLoan
	authCodeType := types.AuthCodeTypeText
	// 限制策略(一天6次，每次时间间隔至少60秒)
	smsHitStrategy := limit.MobileStrategyV2(accountBase.Mobile, serviceType, authCodeType)
	if smsHitStrategy > 0 {
		errcode := cerror.SMSRequestFrequencyTooHigh
		if smsHitStrategy == limit.SmsTimesTooMore {
			errcode = cerror.LimitStrategyMobile
		}
		c.Data["json"] = cerror.BuildApiResponse(errcode, "")
		c.ServeJSON()
		return
	}

	// 调用短信服务
	if !service.SendSms(serviceType, authCodeType, accountBase.Mobile, c.Ctx.Input.IP()) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.SMSServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) RepeatLoanVerify() {
	if !service.CheckRepeatLoanVerifyRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 查看是否可以复贷
	if !dao.IsRepeatLoan(c.AccountID) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.MismatchRepeatLoan, "")
		c.ServeJSON()
		return
	}

	accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
	// 验证 auth_code 有效性
	ok := service.CheckSmsCode(accountBase.Mobile, c.RequestJSON["auth_code"].(string))
	if !ok {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidAuthCode, "")
		c.ServeJSON()
		return
	}

	// 写一个redis hash 标志,后继创建订单的时候需要用到 TODO: 创建订单时检测此标记位
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()
	hashKey := beego.AppConfig.String("repeat_loan_verify")
	storageClient.Do("HSET", hashKey, c.AccountID, 1)

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) RepeatLoanVerifyV2() {
	if !service.CheckRepeatLoanVerifyRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 查看是否可以复贷
	if !dao.IsRepeatLoan(c.AccountID) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.MismatchRepeatLoan, "")
		c.ServeJSON()
		return
	}

	accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
	// 验证 auth_code 有效性
	ok := service.CheckSmsCodeV2(accountBase.Mobile, c.RequestJSON["auth_code"].(string), types.ServiceRepeatedLoan)
	if !ok {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidAuthCode, "")
		c.ServeJSON()
		return
	}

	// 写一个redis hash 标志,后继创建订单的时候需要用到 TODO: 创建订单时检测此标记位
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()
	hashKey := beego.AppConfig.String("repeat_loan_verify")
	storageClient.Do("HSET", hashKey, c.AccountID, 1)

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// 当前(最后)一条有效的订单,状态为非结清
func (c *LoanOrderController) Current() {
	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	order, err := dao.AccountLastLoanOrder(c.AccountID)
	// 客户最后一条有效订单,如果状态如果是[结清],或[无效],则不展示给用户
	if err != nil || order.CheckStatus == types.LoanStatusAlreadyCleared || order.CheckStatus == types.LoanStatusInvalid {
		service.BuildEmptyOrderData(data)
		c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
		c.ServeJSON()
		return
	}

	var orderList []models.Order
	orderList = append(orderList, order)
	service.BuildOrderData(data, orderList)
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// 当前(最后)一条有效的订单是申请过程中的,或者未逾期订单,或者逾期订单,或者两条有效的订单(一条是展期申请中，一条是等待展期)
func (c *LoanOrderController) CurrentV2() {
	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	var orderList []models.Order
	isExtension := false
	order, err := dao.AccountLastLoanOrder(c.AccountID)
	// 客户最后一条有效订单,如果状态如果是[结清],[无效],则不展示给用户
	if err != nil || order.CheckStatus == types.LoanStatusAlreadyCleared ||
		order.CheckStatus == types.LoanStatusInvalid {
		goto emptyOrder
	}

	// 客户的最后一条有效订单,如果状态是[展期失效],则当前订单就是该[展期失效]订单的父订单[原订单],状态为[逾期]
	if order.CheckStatus == types.LoanStatusRollFail {
		preOrder, err := models.GetOrder(order.PreOrder)
		if err != nil {
			goto emptyOrder
		}
		order = preOrder
	}

	if order.CheckStatus == types.LoanStatusOverdue {
		if service.IsOrderCanRoll(order) {
			isExtension = true
		}

		orderList = append(orderList, order)
	} else if order.CheckStatus == types.LoanStatusRolling {
		// 客户的最后一条有效订单,如果状态是[展期等待],则当前订单包括[展期订单]和[原订单]
		rollOrder, err := models.GetRollOrder(order.Id)
		if err != nil {
			goto emptyOrder
		}
		if rollOrder.CheckStatus == types.LoanStatusRollApply {
			orderList = append(orderList, rollOrder)
		}
		orderList = append(orderList, order)
	} else {
		orderList = append(orderList, order)
	}

	service.BuildOrderDataV2(data, orderList, isExtension)
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
	return

emptyOrder:
	service.BuildEmptyOrderData(data)
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
	return
}

func (c *LoanOrderController) All() {
	if !service.CheckRepeatLoanAllRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	offset, _ := tools.Str2Int64(c.RequestJSON["offset"].(string))
	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	orderList, num, err := dao.AccountHistoryLoanOrder(c.AccountID, offset)
	if err != nil || num <= 0 {
		service.BuildEmptyOrderData(data)
		// 修正偏移量,仿止从头循环
		if offset > 0 {
			data["offset"] = tools.Int642Str(offset)
		}
		c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
		c.ServeJSON()
		return
	}

	service.BuildOrderData(data, orderList)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) AllV2() {
	if !service.CheckRepeatLoanAllRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	offset, _ := tools.Str2Int64(c.RequestJSON["offset"].(string))
	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	orderList, num, err := dao.AccountHistoryLoanOrderV2(c.AccountID, offset)
	if err != nil || num <= 0 {
		service.BuildEmptyOrderData(data)
		// 修正偏移量,仿止从头循环
		if offset > 0 {
			data["offset"] = tools.Int642Str(offset)
		}
		c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
		c.ServeJSON()
		return
	}

	service.BuildOrderDataV2(data, orderList, false)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// 当前(最后)一条有效的订单是(等待还款,部分还款,逾期,展期(展期时显示的原订单信息))时,在首页显示订单
func (c *LoanOrderController) HomeOrder() {
	var data map[string]interface{}
	var order models.Order
	flag := true
	order, err := dao.AccountLastLoanOrder(c.AccountID)

	res, errs := service.GetUserLastPaymentVoucher(order.Id)
	if errs != nil {
		flag = false
	}
	if res == (models.PaymentVoucher{}) {
		flag = false
	}
	// 客户最后一条有效订单,如果状态不是[等待还款,部分还款,逾期,等待展期,展期失效],则不展示给用户
	if err != nil || (order.CheckStatus != types.LoanStatusWaitRepayment &&
		order.CheckStatus != types.LoanStatusPartialRepayment &&
		order.CheckStatus != types.LoanStatusOverdue &&
		order.CheckStatus != types.LoanStatusRolling &&
		order.CheckStatus != types.LoanStatusRollFail) {
		goto emptyOrder
	}

	// 客户的最后一条有效订单,如果状态是[展期失效],则当前订单就是该[展期失效]订单的父订单[原订单],状态为[逾期]
	if order.CheckStatus == types.LoanStatusRollFail {
		preOrder, err := models.GetOrder(order.PreOrder)
		if err != nil {
			goto emptyOrder
		}
		order = preOrder
	}

	data = service.BuildHomeOrderData(order)
	data["is_payment_voucher"] = flag
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
	return

emptyOrder:
	data = service.BuildEmptyHomeOrderData()
	data["is_payment_voucher"] = true
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
	return
}

func (c *LoanOrderController) ExtensionTrialCal() {
	if !service.CheckClientInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	order, err := dao.AccountLastOverdueLoanOrder(c.AccountID)
	if err != nil {
		logs.Error("[ExtensionTrialCal] Customer has no temporary order. accountId:", c.AccountID, ", err:", err)
		return
	}

	if order.CheckStatus != types.LoanStatusOverdue {
		logs.Error("[ExtensionTrialCal] Order can not roll or be rolling. accountId:", c.AccountID, ", orderId:", order.Id, ", err:", err)
		return
	}

	period, minRepay, _, err := service.CalcRollRepayAmount(order)
	if err != nil {
		logs.Error("[ExtensionTrialCal] Roll trial cal fail. accountId:", c.AccountID, ", err:", err)
		return
	}

	data := map[string]interface{}{
		"server_time":          tools.GetUnixMillis(),
		"id":                   order.Id,
		"min_repay":            minRepay,
		"latest_repay_time":    tools.GetIDNCurrDayLastSecond(), // 当前日期的最后一秒
		"extension_refund":     order.Amount,
		"extension_repay_time": tools.NaturalDay(int64(period)),
		"is_trial_cal":         service.IsTrialCalOrApply(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) ExtensionConfirm() {
	if !service.CheckClientInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	is_apply := service.IsTrialCalOrApply()
	if is_apply {
		err := service.CreateRollOrder(c.AccountID)
		if err != nil {
			c.Data["json"] = cerror.BuildApiResponse(cerror.CreateRollOrderFail, "")
			c.ServeJSON()
			return
		}
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
		"is_apply":    is_apply,
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) Confirm() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckConfirmOrderRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	loan, _ := tools.Str2Int64(c.RequestJSON["loan"].(string))
	period, _ := tools.Str2Int(c.RequestJSON["period"].(string))
	if loan <= 0 || period <= 0 {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidParameterValue, "")
		c.ServeJSON()
		return
	}

	orderId, err := service.ConfirmOrder(c.AccountID, loan, period)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.CreateOrderFail, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
		"order_id":    tools.Int642Str(orderId),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) ConfirmV2() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckConfirmOrderRequiredV2(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	loan, _ := tools.Str2Int64(c.RequestJSON["loan"].(string))
	loanNew, _ := tools.Str2Int64(c.RequestJSON["loan_new"].(string))
	contractAmount, _ := tools.Str2Int64(c.RequestJSON["contract_amount"].(string))
	period, _ := tools.Str2Int(c.RequestJSON["period"].(string))
	periodNew, _ := tools.Str2Int(c.RequestJSON["period_new"].(string))
	if loan <= 0 || loanNew <= 0 || contractAmount <= 0 || period <= 0 || periodNew <= 0 {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidParameterValue, "")
		c.ServeJSON()
		return
	}

	product, err := service.ProductSuitablesByPeriod(c.AccountID, periodNew, loanNew)
	if err != nil {
		logs.Error("[ConfirmV2] ProductSuitablesByPeriod can not find product. accountId:", c.AccountID, ", err:", err, " periodNew:", periodNew)
		c.Data["json"] = cerror.BuildApiResponse(cerror.ProductDoesNotExist, "")
		c.ServeJSON()
		return
	}

	loanOrderCond := service.LoanOrderCond{
		Loan:           loan,
		LoanNew:        loanNew,
		ContractAmount: contractAmount,
		Period:         period,
		PeriodNew:      periodNew,
	}
	orderId, err := service.ConfirmOrderV2(c.AccountID, product.Id, loanOrderCond)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.CreateOrderFail, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
		"order_id":    tools.Int642Str(orderId),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) ConfirmV3() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckConfirmOrderRequiredV2(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	loan, _ := tools.Str2Int64(c.RequestJSON["loan"].(string))
	loanNew, _ := tools.Str2Int64(c.RequestJSON["loan_new"].(string))
	contractAmount, _ := tools.Str2Int64(c.RequestJSON["contract_amount"].(string))
	period, _ := tools.Str2Int(c.RequestJSON["period"].(string))
	periodNew, _ := tools.Str2Int(c.RequestJSON["period_new"].(string))
	if loan <= 0 || loanNew <= 0 || contractAmount <= 0 || period <= 0 || periodNew <= 0 {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidParameterValue, "")
		c.ServeJSON()
		return
	}

	product, err := service.ProductSuitablesByPeriod(c.AccountID, periodNew, loanNew)
	if err != nil {
		logs.Error("[ConfirmV3] ProductSuitablesByPeriod can not find product. accountId:%d , err:%s periodNew:%d", c.AccountID, err, periodNew)
		c.Data["json"] = cerror.BuildApiResponse(cerror.ProductDoesNotExist, "")
		c.ServeJSON()
		return
	}

	loanOrderCond := service.LoanOrderCond{
		Loan:           loan,
		LoanNew:        loanNew,
		ContractAmount: contractAmount,
		Period:         period,
		PeriodNew:      periodNew,
	}

	phase, orderId, _, err := service.ConfirmOrderV3(c.AccountID, product.Id, loanOrderCond, 0)
	if err != nil {
		// 确认订单失败，新版本有可能是由于活体认证失效，需重新认证而失败。
		logs.Error("[ConfirmV3.ConfirmOrderV3] phase:", phase, " err:", err, " orderId", orderId)
		// c.Data["json"] = cerror.BuildApiResponse(cerror.CreateOrderFail, "")
		// c.ServeJSON()
		// return
	}

	data := map[string]interface{}{
		"server_time":  tools.GetUnixMillis(),
		"order_id":     tools.Int642Str(orderId),
		"current_step": phase,
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) ConfirmV4() {
	authCode := c.RequestJSON["auth_code"].(string)
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckConfirmOrderRequiredV2(c.RequestJSON) ||
		(service.IsVerifySms(c.AccountID) && len(authCode) <= 0) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 确认订单，校验验证码
	if service.IsVerifySms(c.AccountID) {
		accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
		ok := service.CheckSmsCodeV2(accountBase.Mobile, authCode, types.ServiceConfirmOrder)
		if !ok {
			c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidAuthCode, "")
			c.ServeJSON()
			return
		}
	}

	profile, _ := dao.CustomerProfile(c.AccountID)

	bankAccountNumber, _ := c.RequestJSON["bank_account_number"].(string)
	if bankAccountNumber != "" && bankAccountNumber != profile.BankNo {
		err := profile.ChangeBankNo(bankAccountNumber)
		if err != nil {
			errMsg := fmt.Sprintf("LoanOrderController ConfirmV4 profile.ChangeBankNo Update err")
			logs.Error(errMsg)
		}
	}

	loan, _ := tools.Str2Int64(c.RequestJSON["loan"].(string))
	loanNew, _ := tools.Str2Int64(c.RequestJSON["loan_new"].(string))
	contractAmount, _ := tools.Str2Int64(c.RequestJSON["contract_amount"].(string))
	period, _ := tools.Str2Int(c.RequestJSON["period"].(string))
	periodNew, _ := tools.Str2Int(c.RequestJSON["period_new"].(string))
	coupon := int64(0)
	if v, ok := c.RequestJSON["coupon"]; ok && v != nil {
		coupon, _ = tools.Str2Int64(v.(string))
	}
	if loan <= 0 || loanNew <= 0 || contractAmount <= 0 || period <= 0 || periodNew <= 0 {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidParameterValue, "")
		c.ServeJSON()
		return
	}

	product, err := service.ProductSuitablesByPeriod(c.AccountID, periodNew, loanNew)
	if err != nil {
		logs.Error("[ConfirmV4] ProductSuitablesByPeriod can not find product. accountId:%d periodNew:%d err:%v", c.AccountID, periodNew, err)
		c.Data["json"] = cerror.BuildApiResponse(cerror.ProductDoesNotExist, "")
		c.ServeJSON()
		return
	}

	loanOrderCond := service.LoanOrderCond{
		Loan:           loan,
		LoanNew:        loanNew,
		ContractAmount: contractAmount,
		Period:         period,
		PeriodNew:      periodNew,
	}

	phase, orderId, couponErr, err := service.ConfirmOrderV3(c.AccountID, product.Id, loanOrderCond, coupon)
	if err != nil {
		// 确认订单失败，新版本有可能是由于活体认证失效，需重新认证而失败。
		logs.Error("[ConfirmV3.ConfirmOrderV3] phase:", phase, " err:", err, " orderId", orderId)
		// c.Data["json"] = cerror.BuildApiResponse(cerror.CreateOrderFail, "")
		// c.ServeJSON()
		// return
	}
	couponMsg := ""
	if couponErr != nil {
		couponMsg = "Kupon anda telah kadaluarsa"
	}
	data := map[string]interface{}{
		"server_time":   tools.GetUnixMillis(),
		"order_id":      tools.Int642Str(orderId),
		"current_step":  phase,
		"coupon_result": couponMsg,
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// ConfirmTwo（首贷借贷流程变化）
func (c *LoanOrderController) ConfirmTwo() {
	authCode := c.RequestJSON["auth_code"].(string)
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckConfirmOrderRequiredV2(c.RequestJSON) ||
		(service.IsVerifySms(c.AccountID) && len(authCode) <= 0) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	// 确认订单，校验验证码
	if service.IsVerifySms(c.AccountID) {
		accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
		ok := service.CheckSmsCodeV2(accountBase.Mobile, authCode, types.ServiceConfirmOrder)
		if !ok {
			c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidAuthCode, "")
			c.ServeJSON()
			return
		}
	}

	profile, _ := dao.CustomerProfile(c.AccountID)

	bankAccountNumber, _ := c.RequestJSON["bank_account_number"].(string)
	if bankAccountNumber != "" && bankAccountNumber != profile.BankNo {
		err := profile.ChangeBankNo(bankAccountNumber)
		if err != nil {
			errMsg := fmt.Sprintf("LoanOrderController ConfirmTwo profile.ChangeBankNo Update err")
			logs.Error(errMsg)
		}
	}

	quota := int64(0)
	if v, ok := c.RequestJSON["quota"]; ok && v != nil {
		quota, _ = tools.Str2Int64(v.(string))
	}

	loan, _ := tools.Str2Int64(c.RequestJSON["loan"].(string))
	loanNew, _ := tools.Str2Int64(c.RequestJSON["loan_new"].(string))
	contractAmount, _ := tools.Str2Int64(c.RequestJSON["contract_amount"].(string))
	period, _ := tools.Str2Int(c.RequestJSON["period"].(string))
	periodNew, _ := tools.Str2Int(c.RequestJSON["period_new"].(string))
	if loan <= 0 || loanNew <= 0 || contractAmount <= 0 || period <= 0 || periodNew <= 0 {
		c.Data["json"] = cerror.BuildApiResponse(cerror.InvalidParameterValue, "")
		c.ServeJSON()
		return
	}

	coupon := int64(0)
	if v, ok := c.RequestJSON["coupon"]; ok && v != nil {
		coupon, _ = tools.Str2Int64(v.(string))
	}

	// 计算用户提升后的额度
	loanNew = loanNew + quota

	product, err := service.ProductSuitablesByPeriod(c.AccountID, periodNew, loanNew)
	if err != nil {
		logs.Error("[ConfirmTwo] ProductSuitablesByPeriod can not find product. accountId:%d periodNew:%d err:%v loanNew:%d", c.AccountID, periodNew, err, loanNew)
		c.Data["json"] = cerror.BuildApiResponse(cerror.ProductDoesNotExist, "")
		c.ServeJSON()
		return
	}

	loanOrderCond := service.LoanOrderCond{
		Loan:           loan,
		LoanNew:        loanNew,
		ContractAmount: contractAmount,
		Period:         period,
		PeriodNew:      periodNew,
	}

	phase, orderId, couponErr, err := service.ConfirmOrderTwo(c.AccountID, product.Id, loanOrderCond, coupon)
	if err != nil {
		// 确认订单失败，新版本有可能是由于活体认证失效，需重新认证而失败。
		logs.Error("[ConfirmTwo.ConfirmOrderTwo] phase:", phase, " err:", err, " orderId", orderId)
		// c.Data["json"] = cerror.BuildApiResponse(cerror.CreateOrderFail, "")
		// c.ServeJSON()
		// return
	} else {
		if quota > 0 {
			service.WriteOrdersQuota(orderId, quota)
		}
	}

	couponMsg := ""
	if couponErr != nil {
		couponMsg = "Kupon anda telah kadaluarsa"
	}
	data := map[string]interface{}{
		"server_time":   tools.GetUnixMillis(),
		"order_id":      tools.Int642Str(orderId),
		"current_step":  phase,
		"coupon_result": couponMsg,
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) ConfirmLoanAuthCode() {

	accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)

	serviceType := types.ServiceConfirmOrder
	authCodeType := types.AuthCodeTypeText
	// 限制策略(一天6次，每次时间间隔至少60秒)
	smsHitStrategy := limit.MobileStrategyV2(accountBase.Mobile, serviceType, authCodeType)
	if smsHitStrategy > 0 {
		errcode := cerror.SMSRequestFrequencyTooHigh
		if smsHitStrategy == limit.SmsTimesTooMore {
			errcode = cerror.LimitStrategyMobile
		}
		c.Data["json"] = cerror.BuildApiResponse(errcode, "")
		c.ServeJSON()
		return
	}

	// 调用短信服务
	if !service.SendSms(serviceType, authCodeType, accountBase.Mobile, c.Ctx.Input.IP()) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.SMSServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) ConfirmLoanVoiceAuthCode() {

	accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)

	serviceType := types.ServiceConfirmOrder
	authCodeType := types.AuthCodeTypeVoice
	// 限制策略(一天6次，每次时间间隔至少60秒)
	smsHitStrategy := limit.MobileStrategyV2(accountBase.Mobile, serviceType, authCodeType)
	if smsHitStrategy > 0 {
		errcode := cerror.SMSRequestFrequencyTooHigh
		if smsHitStrategy == limit.SmsTimesTooMore {
			errcode = cerror.LimitStrategyVoiceAuthCode
		}
		c.Data["json"] = cerror.BuildApiResponse(errcode, "")
		c.ServeJSON()
		return
	}

	// 调用语音验证码服务
	if !service.SendVoiceAuthCode(serviceType, authCodeType, accountBase.Mobile, c.Ctx.Input.IP()) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.SMSServiceUnavailable, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// 获取借款限额
func (c *LoanOrderController) LoanQuota() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckLoanQuotaRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	loanInt64, _ := tools.Str2Int64(c.RequestJSON["loan"].(string))
	periodInt, _ := tools.Str2Int(c.RequestJSON["period"].(string))
	tmpOrder, err := dao.AccountLastTmpLoanOrderByCond(c.AccountID, loanInt64, periodInt)
	if err != nil {
		logs.Warning("[LoanQuota] customer has no temporary order. accountId:", c.AccountID, ", err:", err)
		return
	}

	loan := tmpOrder.Loan
	period := tmpOrder.Period

	if !dao.IsRepeatLoan(c.AccountID) {
		// 首贷
		loanConf, _ := config.ValidItemInt64("first_loan_amount")
		loanConfWeekend, _ := config.ValidItemInt64("first_loan_amount_weekend")

		periodConf, _ := config.ValidItemInt("first_loan_period")
		periodConfWeekend, _ := config.ValidItemInt("first_loan_period_weekend")

		// 首贷如果是周末的话 使用周末的贷款配置
		if service.IsWeekend() {
			loanConf = loanConfWeekend
			periodConf = periodConfWeekend
		}

		if loan > loanConf {
			loan = loanConf
		}
		if period > periodConf {
			period = periodConf
		}

	} else {
		//获取风控对复贷账号的额度账期配置
		quotaConfModel, err := dao.GetLastAccountQuotaConf(c.AccountID)
		if err != nil {
			logs.Error("[LoanQuota] GetLastAccountQuotaConf happend err:", err)
		}
		//如果用户借款金额和期限
		if loan > quotaConfModel.Quota {
			loan = quotaConfModel.Quota
		}
		if period > int(quotaConfModel.AccountPeriod) {
			period = int(quotaConfModel.AccountPeriod)
		}
	}

	product, err := service.ProductSuitablesByPeriod(c.AccountID, period, loan)
	if err != nil {
		logs.Error("[LoanQuota] ProductSuitablesByPeriod can not find product. accountId:%d , err:%s period:%d", c.AccountID, err, period)
		c.Data["json"] = cerror.BuildApiResponse(cerror.ProductDoesNotExist, "")
		c.ServeJSON()
		return
	}

	trialIn := types.ProductTrialCalcIn{
		ID:           product.Id,
		Loan:         loan,
		Amount:       0,
		Period:       period,
		LoanDate:     "",
		CurrentDate:  "",
		RepayDate:    "",
		RepayedTotal: 0,
	}

	_, fee, _, contractAmount, _ := pt.GetInterestAndFee(trialIn, product)

	data := map[string]interface{}{
		"server_time":     tools.GetUnixMillis(),
		"loan":            loan,
		"contract_amount": contractAmount,
		"period":          period,
		"fee":             fee,
		"overdue_rate":    fmt.Sprintf("%g%%", float64(product.DayPenaltyRate)/float64(100)),
		"interest":        fmt.Sprintf("%g%%", float64(product.DayInterestRate*365)/float64(100)),
		"repay_time":      tools.MDateUTC(tools.NaturalDay(int64(period))),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// 获取借款限额
func (c *LoanOrderController) LoanQuotaV2() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckLoanQuotaRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	loanInt64, _ := tools.Str2Int64(c.RequestJSON["loan"].(string))
	periodInt, _ := tools.Str2Int(c.RequestJSON["period"].(string))
	tmpOrder, err := dao.AccountLastTmpLoanOrderByCond(c.AccountID, loanInt64, periodInt)
	if err != nil {
		logs.Warning("[LoanQuotaV2] customer has no temporary order. accountId:", c.AccountID, ", err:", err)
		return
	}

	isDoneAuth, quotaTotal, _, _ := service.CustomerAuthorize(c.AccountID)
	loan := tmpOrder.Loan
	period := tmpOrder.Period

	if !dao.IsRepeatLoan(c.AccountID) {
		//首贷不提额
		quotaTotal = 0

		// 首贷
		loanConf, _ := config.ValidItemInt64("first_loan_amount")
		loanConfWeekend, _ := config.ValidItemInt64("first_loan_amount_weekend")

		periodConf, _ := config.ValidItemInt("first_loan_period")
		periodConfWeekend, _ := config.ValidItemInt("first_loan_period_weekend")

		// 首贷如果是周末的话 使用周末的贷款配置
		if service.IsWeekend() {
			loanConf = loanConfWeekend
			periodConf = periodConfWeekend
		}

		tmpQuota := int64(0)
		if loan > loanConf {
			tmpQuota = loan - loanConf
			loan = loanConf
		}
		if quotaTotal > tmpQuota {
			quotaTotal = tmpQuota
		}

		if period > periodConf {
			period = periodConf
		}

	} else {
		//获取风控对复贷账号的额度账期配置
		quotaConfModel, err := dao.GetLastAccountQuotaConf(c.AccountID)
		if err != nil {
			logs.Error("[LoanQuotaV2] GetLastAccountQuotaConf happend err:", err)
		}
		//如果用户借款金额和期限
		tmpQuota := int64(0)
		if loan > quotaConfModel.Quota {
			tmpQuota = loan - quotaConfModel.Quota
			loan = quotaConfModel.Quota
		}

		//用户借的钱可能并没有达到 提额后的上线或还没达到提额前的值
		if quotaTotal > tmpQuota {
			quotaTotal = tmpQuota
		}

		if period > int(quotaConfModel.AccountPeriod) {
			period = int(quotaConfModel.AccountPeriod)
		}
	}

	//低版本不提额
	if c.VersionCode < types.IndonesiaAppRipeVersionNewReloanStep &&
		c.UIVersion == types.IndonesiaAppUIVersion {
		quotaTotal = 0
	}

	product, err := service.ProductSuitablesByPeriod(c.AccountID, period, loan+quotaTotal)
	if err != nil {
		logs.Error("[LoanQuotaV2] ProductSuitablesByPeriod can not find product. accountId:%d , err:%s period:%d loan:%d quotaTotal:%d", c.AccountID, err, period, loan, quotaTotal)
		c.Data["json"] = cerror.BuildApiResponse(cerror.ProductDoesNotExist, "")
		c.ServeJSON()
		return
	}

	trialIn := types.ProductTrialCalcIn{
		ID:           product.Id,
		Loan:         loan + quotaTotal,
		Amount:       0,
		Period:       period,
		LoanDate:     "",
		CurrentDate:  "",
		RepayDate:    "",
		RepayedTotal: 0,
	}

	_, fee, _, contractAmount, _ := pt.GetInterestAndFee(trialIn, product)

	//bankAccountNumber, _ := bluepay.NameValidator(c.AccountID)

	data := map[string]interface{}{
		"server_time":          tools.GetUnixMillis(),
		"loan":                 loan,
		"contract_amount":      contractAmount,
		"period":               period,
		"fee":                  fee,
		"overdue_rate":         fmt.Sprintf("%g%%", float64(product.DayPenaltyRate)/float64(100)),
		"interest":             fmt.Sprintf("%g%%", float64(product.DayInterestRate*365)/float64(100)),
		"repay_time":           tools.MDateUTC(tools.NaturalDay(int64(period))),
		"is_sms_verify":        service.IsVerifySms(c.AccountID), // 是否检查验证码的标示
		"bank_account_number":  "",
		"authorization_info":   service.AuthorizationInfo(c.AccountID),
		"is_done_auth":         isDoneAuth,
		"quota":                quotaTotal,
		"quota_after_increase": loan + quotaTotal,
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) CreateOrderV2() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckCreateOrderRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	if !service.HaveUnsetOrder(c.AccountID) {
		accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
		isHitMobile, _ := models.IsBlacklistMobile(accountBase.Mobile)
		if isHitMobile {
			logs.Warn("[AccountInfo] 手机号在内部黑名单内, Mobile: %s", accountBase.Mobile)
		}

		ip := c.RequestJSON["ip"].(string)
		isHitIP, _ := models.IsBlacklistIP(ip)
		if isHitIP {
			logs.Warn("[AccountInfo] IP在内部黑名单内, IP: %s", ip)
		}

		_, eAccountDesc := service.DisplayVAInfoV2(c.AccountID)
		if isHitMobile || isHitIP {
			data := map[string]interface{}{
				"server_time":      tools.GetUnixMillis(),
				"is_repeat_loan":   dao.IsRepeatLoan(c.AccountID),
				"account_profile":  service.BuildAccountProfile(c.AccountID),
				"loan_lifetime":    types.LoanHitBlackList,
				"current_step":     0,
				"e_account_number": eAccountDesc,
				"amount":           0,
				"remaining_days":   0,
			}

			c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
			c.ServeJSON()
			return
		}
	}

	loan, _ := tools.Str2Int64(c.RequestJSON["loan"].(string))
	period, _ := tools.Str2Int(c.RequestJSON["period"].(string))
	if loan > 0 && period > 0 {
		// 创建临时订单
		//// 1. 检查产品
		product, err := service.ProductSuitablesByPeriod(c.AccountID, period, loan)
		if err != nil {
			logs.Error("[CreateOrderV2] ProductSuitablesByPeriod can not find product. accountId:%d , err:%s period:%d", c.AccountID, err, period)
			c.Data["json"] = cerror.BuildApiResponse(cerror.ProductDoesNotExist, "")
			c.ServeJSON()
			return
		}
		//// 2. 创建借款订单
		_, orderId, err := service.CreateOrder(c.AccountID, product.Id, loan, period, types.IsTemporaryYes)
		if err == nil {
			// 写创建订单现场数据
			c.RequestJSON["service_type"] = types.ServiceCreateOrder
			c.RequestJSON["related_id"] = orderId
			c.RequestJSON["mobile"] = "" //! 注意,需要显示设置为空,否则有可能引起内核恐慌
			service.RecordClientInfo(c.RequestJSON)
		}
		//! 不能创建订单时,接口数据正常返回
	}

	data := map[string]interface{}{
		"server_time":     tools.GetUnixMillis(),
		"is_repeat_loan":  dao.IsRepeatLoan(c.AccountID),
		"account_profile": service.BuildAccountProfile(c.AccountID),
		"loan_lifetime":   service.GetLoanLifetime(c.AccountID),
		"current_step":    service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode),
		"menu_show":       service.MenuControlByOrderStatus(c.AccountID),
		"menu_show_v2":    service.MenuControlByOrderStatusV2(c.AccountID),
		//"_debug-AccountID": c.AccountID,
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// CreateOrderTwo（首贷借贷流程变化）
func (c *LoanOrderController) CreateOrderTwo() {
	if !service.CheckClientInfoRequired(c.RequestJSON) || !service.CheckCreateOrderRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	if !service.HaveUnsetOrder(c.AccountID) {
		accountBase, _ := models.OneAccountBaseByPkId(c.AccountID)
		isHitMobile, _ := models.IsBlacklistMobile(accountBase.Mobile)
		if isHitMobile {
			logs.Warn("[CreateOrderTwo] 手机号在内部黑名单内, Mobile: %s", accountBase.Mobile)
		}

		ip := c.RequestJSON["ip"].(string)
		isHitIP, _ := models.IsBlacklistIP(ip)
		if isHitIP {
			logs.Warn("[CreateOrderTwo] IP在内部黑名单内, IP: %s", ip)
		}

		_, eAccountDesc := service.DisplayVAInfoV2(c.AccountID)
		if isHitMobile || isHitIP {
			data := map[string]interface{}{
				"server_time":      tools.GetUnixMillis(),
				"is_repeat_loan":   dao.IsRepeatLoan(c.AccountID),
				"account_profile":  service.BuildAccountProfile(c.AccountID),
				"loan_lifetime":    types.LoanHitBlackList,
				"current_step":     0,
				"e_account_number": eAccountDesc,
				"amount":           0,
				"remaining_days":   0,
			}

			c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
			c.ServeJSON()
			return
		}
	}

	loan, _ := tools.Str2Int64(c.RequestJSON["loan"].(string))
	period, _ := tools.Str2Int(c.RequestJSON["period"].(string))
	if loan > 0 && period > 0 {
		// 创建临时订单
		//// 1. 检查产品
		product, err := service.ProductSuitablesByPeriod(c.AccountID, period, loan)
		if err != nil {
			logs.Error("[CreateOrderTwo] ProductSuitablesByPeriod can not find product. accountId:%d , err:%s period:%d loan:%d", c.AccountID, err, period, loan)
			c.Data["json"] = cerror.BuildApiResponse(cerror.ProductDoesNotExist, "")
			c.ServeJSON()
			return
		}
		//// 2. 创建借款订单
		_, orderId, err := service.CreateOrder(c.AccountID, product.Id, loan, period, types.IsTemporaryYes)
		if err == nil {
			// 写创建订单现场数据
			c.RequestJSON["service_type"] = types.ServiceCreateOrder
			c.RequestJSON["related_id"] = orderId
			c.RequestJSON["mobile"] = "" //! 注意,需要显示设置为空,否则有可能引起内核恐慌
			service.RecordClientInfo(c.RequestJSON)
		}
		//! 不能创建订单时,接口数据正常返回
	}

	progress, phase := service.ProfileCompletePhaseTwo(c.AccountID, c.UIVersion, c.VersionCode)
	data := map[string]interface{}{
		"server_time":     tools.GetUnixMillis(),
		"is_repeat_loan":  dao.IsRepeatLoan(c.AccountID),
		"account_profile": service.BuildAccountProfile(c.AccountID),
		"loan_lifetime":   service.GetLoanLifetime(c.AccountID),
		"current_step":    phase,
		"progress":        progress,
		"menu_show":       service.MenuControlByOrderStatus(c.AccountID),
		"menu_show_v2":    service.MenuControlByOrderStatusV2(c.AccountID),
		//"_debug-AccountID": c.AccountID,
	}

	service.ApiDataAddEAccountNumber(c.AccountID, data)
	service.ApiDataAddCurrentLoanInfo(c.AccountID, data)

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func xenditPaymentCodeResp(c *LoanOrderController, marketPayment models.FixPaymentCode) {

	logs.Debug("XenditPaymentCode step 2", marketPayment)
	data := map[string]interface{}{
		"payment_code": marketPayment.PaymentCode,
		"expire_time":  marketPayment.ExpirationDate,
		"amount":       marketPayment.ExpectedAmount,
		//"status":       marketPayment.Status,
	}
	logs.Debug("XenditPaymentCode step 3", marketPayment)
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *LoanOrderController) XenditPaymentCode() {

	//accountId := int64(180710010182740299)
	accountId := c.AccountID
	order, err := dao.AccountLastLoanOrder(accountId)

	if order.CheckStatus == types.LoanStatusRolling {
		//顺序提前
		c.Data["json"] = cerror.BuildApiResponse(cerror.RollOrderNotSupport, "")
		c.ServeJSON()
		return
	}

	if err != nil || (order.CheckStatus != types.LoanStatusWaitRepayment && order.CheckStatus != types.LoanStatusOverdue &&
		order.CheckStatus != types.LoanStatusPartialRepayment) {
		logs.Warning("[XenditPaymentCode] order does not exist, accountId is %d ", c.AccountID)
		c.Data["json"] = cerror.BuildApiResponse(cerror.OrderDoesNotExist, "")
		c.ServeJSON()
		return
	}

	marketPayment, err := models.OneFixPaymentCodeByUserAccountId(order.UserAccountId)
	if err != nil {
		//查不到，就去生成
		err, marketPayment, _ := xendit.MarketPaymentCodeGenerate(order.Id, 0)
		if err != nil {
			logs.Error("[XenditPaymentCode] marketPayment generation err [%#v], order_id is %d ", err, order.Id)
			c.Data["json"] = cerror.BuildApiResponse(cerror.PaymentGenerateErr, "")
			c.ServeJSON()
			return
		}
		logs.Debug("XenditPaymentCode step 1", marketPayment)
		xenditPaymentCodeResp(c, marketPayment)
		return
	}
	xenditPaymentCodeResp(c, marketPayment)
}

//

// 上传还款凭证
func (c *LoanOrderController) PaymentVoucher() {
	// 简单判断一下
	if !service.CheckPaymentVoucherRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	orderData, err := dao.AccountLastLoanOrder(c.AccountID)
	if err != nil {
		logs.Error("[PaymentVoucher] order nil err:%s, accountId:", err, c.AccountID)
		c.Data["json"] = cerror.BuildApiResponse(cerror.PaymentVoucherCode, "")
		c.ServeJSON()

	}

	// 1. 将文件流写入本地
	// 2. 将文件上传到s3
	resoureId, tmpFile, code, err := c.UploadResource("fs1", types.Use2PaymentVoucher)
	defer tools.Remove(tmpFile)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}
	//还款方式
	reimbMeans, ok := c.RequestJSON["reimb_means"].(string)
	if !ok {
		reimbMeans = ""
	}
	record := map[string]interface{}{
		"account_id":  c.AccountID,
		"resource_id": resoureId,
		"order_id":    orderData.Id,
		"reimb_means": reimbMeans,
	}

	service.AddOnePaymentVoucherResource(record)

	// 5. 返回结果给客户端
	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}
