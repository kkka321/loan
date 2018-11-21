package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/dao"
	"micro-loan/common/lib/device"
	"micro-loan/common/models"
	"micro-loan/common/pkg/coupon_event"
	"micro-loan/common/pkg/entrust/serveentrust"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/pkg/schema_task"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func repayNormalLoan(userAccountId int64, amount int64, bankCode string, paymentCode string, vaCompanyCode int, callbackStr string, dataOrder *models.Order, isRefund ...bool) (int64, error) {
	logs.Debug("[repayNormalLoan] begin userAccountId:%d, amount:%d, bankCode:%s, vaCompanyCode:%d, orderId:% isRefund:%v",
		userAccountId, amount, bankCode, vaCompanyCode, dataOrder.Id, isRefund)

	isRefundFlag := false
	if len(isRefund) > 0 && isRefund[0] == true {
		isRefundFlag = true
	}
	timetag := tools.GetUnixMillis()
	repayPlan, err := models.GetLastRepayPlanByOrderid(dataOrder.Id)
	preReduced, _ := dao.GetLastPrereducedByOrderid(repayPlan.OrderId)
	//如果有未生效的结清减免
	if (preReduced.GracePeriodInterestPrededuced > 0 || preReduced.PenaltyPrereduced > 0) && preReduced.ReduceStatus == types.ReduceStatusNotValid {
		//如果客户还款金额大于预减免应还总金额
		amountWithPrereduced, _ := repayplan.CaculateRepayTotalAmountWithPreReducedByRepayPlan(repayPlan)
		logs.Debug("[repayNormalLoan] 结清减免最低应还总额:", amountWithPrereduced, "当前还款额:", amount)
		if amount >= amountWithPrereduced {
			//更新预减免为生效状态
			preReduced.ReduceStatus = types.ReduceStatusValid
			preReduced.GraceInterestReduced = preReduced.GracePeriodInterestPrededuced
			preReduced.PenaltyReduced = preReduced.PenaltyPrereduced
			//更新还款计划
			olderRepayPlan := repayPlan
			repayPlan.GracePeriodInterestReduced = repayPlan.GracePeriodInterestReduced + preReduced.GracePeriodInterestPrededuced
			repayPlan.PenaltyReduced = repayPlan.PenaltyReduced + preReduced.PenaltyPrereduced
			models.UpdateRepayPlan(&repayPlan)
			models.OpLogWrite(0, repayPlan.Id, models.OpCodeRepayPlanUpdate, repayPlan.TableName(), olderRepayPlan, repayPlan)
		} else {
			//更新预减免为生效状态
			preReduced.ReduceStatus = types.ReduceStatusInvalid
			preReduced.InvalidReason = types.ClearReducedInvalidReasonNotClear
		}
		preReduced.ConfirmTime = timetag
		preReduced.Utime = timetag
		cols := []string{"reduce_status", "grace_interest_reduced", "penalty_reduced", "invalid_reason", "confirm_time", "utime"}
		models.OrmUpdate(&preReduced, cols)
	}

	if err != nil {
		//没有还款计划
		errStr := fmt.Sprintf("[repayNormalLoan] GetLastRepayPlanByOrderid does not get data, orderid:%d, err:%s", dataOrder.Id, err.Error())
		logs.Error(errStr)
		return 0, fmt.Errorf(errStr)
	}

	//用户出账
	eTrans := models.User_E_Trans{}
	eTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
	eTrans.OrderId = dataOrder.Id
	eTrans.UserAccountId = userAccountId
	eTrans.VaCompanyCode = vaCompanyCode
	eTrans.PayType = types.PayTypeMoneyOut
	if isRefundFlag {
		eTrans.PayType = types.PayTypeRefundOut
	}
	eTrans.Ctime = timetag
	eTrans.Utime = timetag
	needTotal, err := generateRepayTrans(dataOrder, &eTrans, &repayPlan, amount)
	if err != nil {
		str := fmt.Sprintf("[repayNormalLoan] generateRepayTrans return error orderId:%d, err:%v", dataOrder.Id, err)
		logs.Error(str)
		return 0, fmt.Errorf(str)
	}

	// 检查用户是否多还钱
	balance := int64(0)
	if amount >= needTotal {
		balance = amount - needTotal
		eTrans.Balance = balance
	}

	olderOrder := *dataOrder
	eAccount, _ := models.GetLastestActiveEAccountByVacompanyType(userAccountId, vaCompanyCode)

	payment := models.Payment{}
	payment.Id, _ = device.GenerateBizId(types.PaymentBiz)
	payment.Amount = amount
	payment.OrderId = dataOrder.Id
	payment.PayType = types.PayTypeMoneyIn
	payment.VaCompanyCode = vaCompanyCode
	payment.UserAccountId = tools.Int642Str(userAccountId)
	payment.UserBankCode = bankCode
	if paymentCode != "" {
		payment.VaCode = paymentCode
	} else {
		payment.VaCode = eAccount.BankCode + eAccount.EAccountNumber
	}
	payment.Ctime = timetag
	payment.Utime = timetag
	if !isRefundFlag {
		//退款到订单的操作没有余额的变动所以不需要记录payment
		payment.AddPayment(&payment)
	}
	//财务记录

	eInTrans := models.User_E_Trans{}
	eInTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
	eInTrans.OrderId = dataOrder.Id
	eInTrans.PaymentId = payment.Id
	eInTrans.UserAccountId = userAccountId
	eInTrans.VaCompanyCode = vaCompanyCode
	eInTrans.Total = amount
	eInTrans.PayType = types.PayTypeMoneyIn
	if isRefundFlag {
		eInTrans.PayType = types.PayTypeRefundIn
	}
	eInTrans.CallbackJson = callbackStr
	eInTrans.Ctime = timetag
	eInTrans.Utime = timetag
	eInTrans.AddEtrans(&eInTrans)
	//用户入账

	//用户出账完结
	eTrans.AddEtrans(&eTrans)

	//每次还款时间都更新下
	dataOrder.RepayTime = timetag

	if amount >= needTotal {
		UpdateOrderToAlreadyCleared(dataOrder, olderOrder)

		if olderOrder.IsReloan == int(types.IsReloanNo) && olderOrder.PreOrder == 0 &&
			(olderOrder.CheckStatus == types.LoanStatusWaitRepayment || olderOrder.CheckStatus == types.LoanStatusPartialRepayment) {
			param := coupon_event.InviteV3Param{}
			param.AccountId = dataOrder.UserAccountId
			param.TaskType = types.AccountTaskRepay
			HandleCouponEvent(coupon_event.TriggerInviteV3, param)
		}
		//改写订单状态，订单已还清
	} else {
		// 自动减免逻辑  此时还的钱已存在于 repayPlan内
		if ok, reduce := CanAutoReduce(dataOrder, &repayPlan, 0); ok {
			err = DoAutoReduce(dataOrder, &repayPlan, reduce)
			if err == nil {
				UpdateOrderToAlreadyCleared(dataOrder, olderOrder)

			} else {
				logs.Error("[repayNormalLoan] DoAutoReduce err:%v repayPlan:%#v", err, repayPlan)
			}
		} else {
			//更新订单状态
			dataOrder.Utime = timetag
			if dataOrder.CheckStatus != types.LoanStatusOverdue {
				dataOrder.CheckStatus = types.LoanStatusPartialRepayment
			}
			models.UpdateOrder(dataOrder)
			ticket.WatchPartialRepayment(dataOrder)
			models.OpLogWrite(0, dataOrder.Id, models.OpCodeOrderUpdate, dataOrder.TableName(), olderOrder, *dataOrder)
		}

	}

	models.UpdateRepayPlan(&repayPlan)
	//更新还款计划

	return balance, nil
}

func doOrderRoll(userAccountId int64, amount int64, bankCode string, paymentCode string, vaCompanyCode int, callbackStr string, dataOrder *models.Order, repayPlan *models.RepayPlan) (int64, error) {
	logs.Debug("[doOrderRoll] begin userAccountId:%d, amount:%d, bankCode:%s, vaCompanyCode:%d, orderId:%d",
		userAccountId, amount, bankCode, vaCompanyCode, dataOrder.Id)

	needTotal, err := repayplan.CaculateRepayTotalAmountByRepayPlan(*repayPlan)
	if err != nil {
		str := fmt.Sprintf("[repayRollLoan] CaculateRepayTotalAmountByRepayPlan return error orderId:%d, err:%v", dataOrder.Id, err)
		logs.Error(str)
		return 0, fmt.Errorf(str)
	}

	//clear
	// 自动减免逻辑，满足自动减免同样去结清订单
	ok, _ := CanAutoReduce(dataOrder, repayPlan, amount)
	if amount >= needTotal || ok {
		logs.Debug("[doOrderRoll] amount >= needTotal or auto reduce order clear, amount:%d, needTotal:%d, orderId:%d ok:%v",
			amount, needTotal, dataOrder.Id, ok)

		balance, err := repayNormalLoan(userAccountId, amount, bankCode, paymentCode, vaCompanyCode, callbackStr, dataOrder)
		if err != nil {
			return 0, err
		}

		rollOrder, err := models.GetRollOrder(dataOrder.Id)
		logs.Error("[doOrderRoll] GetRollOrder error, orderId:%d, err:%v", dataOrder.Id, err)
		if err != nil {
			return 0, nil
		}

		olderTollOrder := rollOrder

		rollOrder.CheckStatus = types.LoanStatusRollFail
		rollOrder.CheckTime = tools.GetUnixMillis()
		rollOrder.Utime = tools.GetUnixMillis()
		models.UpdateOrder(&rollOrder)

		monitor.IncrOrderCount(rollOrder.CheckStatus)

		models.OpLogWrite(0, rollOrder.Id, models.OpCodeOrderUpdate, rollOrder.TableName(), olderTollOrder, rollOrder)

		return balance, nil
	}

	rollOrder, err := models.GetRollOrder(dataOrder.Id)
	if err != nil {
		logs.Error("[doOrderRoll] GetRollOrder error, orderId:%d, err:%v", dataOrder.Id, err)
		return 0, err
	}

	rollProduct, err := models.GetProduct(rollOrder.ProductId)
	if err != nil {
		logs.Error("[doOrderRoll] GetProduct error, rollOrder:%d, productId:%d, err:%v", rollOrder.Id, rollOrder.ProductId, err)
		return 0, err
	}

	eAccount, _ := models.GetLastestActiveEAccountByVacompanyType(userAccountId, vaCompanyCode)
	olderOrder := dataOrder
	olderTollOrder := rollOrder

	//正常还款
	{
		payment := models.Payment{}
		payment.Id, _ = device.GenerateBizId(types.PaymentBiz)
		payment.Amount = amount
		payment.OrderId = dataOrder.Id
		payment.PayType = types.PayTypeMoneyIn
		payment.VaCompanyCode = vaCompanyCode
		if paymentCode != "" {
			payment.VaCode = paymentCode
		} else {
			payment.VaCode = eAccount.BankCode + eAccount.EAccountNumber
		}
		payment.UserAccountId = tools.Int642Str(userAccountId)
		payment.UserBankCode = bankCode
		payment.Ctime = tools.GetUnixMillis()
		payment.Utime = tools.GetUnixMillis()
		payment.AddPayment(&payment)

		eInTrans := models.User_E_Trans{}
		eInTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
		eInTrans.OrderId = dataOrder.Id
		eInTrans.PaymentId = payment.Id
		eInTrans.UserAccountId = userAccountId
		eInTrans.VaCompanyCode = vaCompanyCode
		eInTrans.Total = amount
		eInTrans.PayType = types.PayTypeMoneyIn
		eInTrans.CallbackJson = callbackStr
		eInTrans.Ctime = tools.GetUnixMillis()
		eInTrans.Utime = tools.GetUnixMillis()
		eInTrans.AddEtrans(&eInTrans)

		eOutTrans := models.User_E_Trans{}
		_, err = generateRepayTrans(dataOrder, &eOutTrans, repayPlan, amount)
		if err != nil {
			logs.Error("[doOrderRoll] generateRepayTrans error, orderId:%d, amount:%d, err:%v", dataOrder.Id, amount, err)
		}
		eOutTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
		eOutTrans.OrderId = dataOrder.Id
		eOutTrans.UserAccountId = userAccountId
		eOutTrans.VaCompanyCode = vaCompanyCode
		eOutTrans.PayType = types.PayTypeMoneyOut
		eOutTrans.CallbackJson = callbackStr
		eOutTrans.Ctime = tools.GetUnixMillis()
		eOutTrans.Utime = tools.GetUnixMillis()
		eOutTrans.AddEtrans(&eOutTrans)
	}

	balance := repayPlan.AmountPayed + repayPlan.AmountReduced
	logs.Warn("[doOrderRoll] rollOrder has balance, orderId:%d, rollOrderId:%d, balance:%d", dataOrder.Id, rollOrder.Id, balance)

	//余额结转
	{
		eTrans := models.User_E_Trans{}
		eTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
		eTrans.OrderId = dataOrder.Id
		eTrans.UserAccountId = userAccountId
		eTrans.VaCompanyCode = types.MobiFundTran
		eTrans.Total = 0
		eTrans.Amount = -1 * balance
		eTrans.Balance = balance
		eTrans.PayType = types.PayTypeTran
		eTrans.Ctime = tools.GetUnixMillis()
		eTrans.Utime = tools.GetUnixMillis()
		eTrans.AddEtrans(&eTrans)
	}

	//虚拟还款
	{
		vPayment := models.Payment{}
		vPayment.Id, _ = device.GenerateBizId(types.PaymentBiz)
		vPayment.Amount = repayPlan.Amount
		vPayment.OrderId = dataOrder.Id
		vPayment.VaCompanyCode = types.MobiFundVirtual
		vPayment.PayType = types.PayTypeRollIn
		vPayment.UserAccountId = tools.Int642Str(userAccountId)
		vPayment.Ctime = tools.GetUnixMillis()
		vPayment.Utime = tools.GetUnixMillis()
		vPayment.AddPayment(&vPayment)

		evInTrans := models.User_E_Trans{}
		evInTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
		evInTrans.OrderId = dataOrder.Id
		evInTrans.PaymentId = vPayment.Id
		evInTrans.UserAccountId = userAccountId
		evInTrans.Total = repayPlan.Amount
		evInTrans.VaCompanyCode = types.MobiFundVirtual
		evInTrans.PayType = types.PayTypeRollIn
		evInTrans.Ctime = tools.GetUnixMillis()
		evInTrans.Utime = tools.GetUnixMillis()
		evInTrans.AddEtrans(&evInTrans)

		repayPlan.AmountReduced = 0
		repayPlan.AmountPayed = 0
		evOutTrans := models.User_E_Trans{}
		_, err = generateRepayTrans(dataOrder, &evOutTrans, repayPlan, repayPlan.Amount)
		if err != nil {
			logs.Error("[doOrderRoll] generateRepayTrans error, orderId:%d, amount:%d, err:%v", dataOrder.Id, amount, err)
		}
		evOutTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
		evOutTrans.OrderId = dataOrder.Id
		evOutTrans.UserAccountId = userAccountId
		evOutTrans.VaCompanyCode = types.MobiFundVirtual
		evOutTrans.PayType = types.PayTypeRollOut
		evOutTrans.Ctime = tools.GetUnixMillis()
		evOutTrans.Utime = tools.GetUnixMillis()
		evOutTrans.AddEtrans(&evOutTrans)
	}

	//更新余额
	{
		IncreaseBalanceByRefund(dataOrder.UserAccountId, 0)
	}

	_, interest, serviceFee := repayplan.CalcRepayInfoV3(rollOrder.Amount, rollProduct, rollOrder.Period)

	dataOrder.RepayTime = tools.GetUnixMillis()
	dataOrder.Utime = tools.GetUnixMillis()
	dataOrder.CheckStatus = types.LoanStatusRollClear
	dataOrder.FinishTime = tools.GetUnixMillis()
	models.UpdateOrder(dataOrder)

	models.UpdateRepayPlan(repayPlan)
	//更新还款计划

	HandleOverdueCase(dataOrder.Id)

	monitor.IncrOrderCount(dataOrder.CheckStatus)

	models.OpLogWrite(0, dataOrder.Id, models.OpCodeOrderUpdate, dataOrder.TableName(), olderOrder, dataOrder)

	HandleDisburse(vaCompanyCode, &rollOrder, bankCode, true)

	balance = balance - interest - serviceFee

	if balance > 0 {
		rollOrder.CheckStatus = types.LoanStatusPartialRepayment
		rollOrder.RepayTime = tools.GetUnixMillis()
	} else {
		rollOrder.CheckStatus = types.LoanStatusWaitRepayment
	}

	rollOrder.IsTemporary = types.IsTemporaryNO
	rollOrder.CheckTime = tools.GetUnixMillis()
	rollOrder.LoanTime = tools.GetUnixMillis()
	rollOrder.PhoneVerifyTime = tools.GetUnixMillis()
	rollOrder.RiskCtlStatus = types.RiskCtlPhoneVerifyPass
	rollOrder.Utime = tools.GetUnixMillis()
	models.UpdateOrder(&rollOrder)

	monitor.IncrOrderCount(rollOrder.CheckStatus)

	models.OpLogWrite(0, rollOrder.Id, models.OpCodeOrderUpdate, rollOrder.TableName(), olderTollOrder, rollOrder)

	rollRepayPlan := models.RepayPlan{}
	rollRepayPlan, err = models.GetLastRepayPlanByOrderid(rollOrder.Id)
	//如果这里返回err，需要回滚之前的操作，暂不考虑
	if err != nil {
		return 0, nil
	}
	if balance > 0 {
		eTranInTrans := models.User_E_Trans{}
		eTranInTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
		eTranInTrans.OrderId = rollOrder.Id
		eTranInTrans.UserAccountId = userAccountId
		eTranInTrans.VaCompanyCode = types.MobiRefundToOrder
		eTranInTrans.Total = balance
		eTranInTrans.PayType = types.PayTypeRefundIn
		eTranInTrans.Ctime = tools.GetUnixMillis()
		eTranInTrans.Utime = tools.GetUnixMillis()
		eTranInTrans.AddEtrans(&eTranInTrans)

		eOutTrans := models.User_E_Trans{}
		eOutTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
		eOutTrans.OrderId = rollOrder.Id
		eOutTrans.UserAccountId = userAccountId
		eOutTrans.VaCompanyCode = types.MobiRefundToOrder
		eOutTrans.PayType = types.PayTypeRefundOut
		eOutTrans.Amount = balance
		eOutTrans.Ctime = tools.GetUnixMillis()
		eOutTrans.Utime = tools.GetUnixMillis()
		eOutTrans.AddEtrans(&eOutTrans)

		rollRepayPlan.AmountPayed += balance
		models.UpdateRepayPlan(&rollRepayPlan)
	}

	accountBase, _ := models.OneAccountBaseByPkId(rollOrder.UserAccountId)

	schema_task.PushBusinessMsg(types.PushTargetRollSuccess, dataOrder.UserAccountId)

	param := make(map[string]interface{})
	param["related_id"] = rollOrder.Id
	schema_task.SendBusinessMsg(types.SmsTargetRollSuccess, types.ServiceCreateOrder, accountBase.Mobile, param)

	/*
		_, eAccountErr := dao.GetActiveEaccountWithBankName(rollOrder.UserAccountId)
		if eAccountErr == nil {
			date := tools.GetLocalDateFormat(rollRepayPlan.RepayDate, "02/01")
			repayMoney, _ := repayplan.CaculateRepayTotalAmountByRepayPlan(rollRepayPlan)

			smsContent := fmt.Sprintf(i18n.GetMessageText(i18n.TextRollSuccess), repayMoney, date)
			sms.Send(types.ServiceCreateOrder, accountBase.Mobile, smsContent, rollOrder.Id)
		}
	*/

	return 0, nil
}

func repayRollLoan(userAccountId int64, amount int64, bankCode string, paymentCode string, vaCompanyCode int, callbackStr string, dataOrder *models.Order) (int64, error) {
	timetag := tools.GetUnixMillis()

	repayPlan, err := models.GetLastRepayPlanByOrderid(dataOrder.Id)
	if err != nil {
		//没有还款计划
		errStr := fmt.Sprintf("[repayRollLoan] GetLastRepayPlanByOrderid does not get data, orderid is %d, err is %s", dataOrder.Id, err.Error())
		logs.Error(errStr)
		return 0, fmt.Errorf(errStr)
	}

	preReduced, _ := dao.GetLastPrereducedByOrderid(repayPlan.OrderId)
	//如果有未生效的结清减免
	if (preReduced.GracePeriodInterestPrededuced > 0 || preReduced.PenaltyPrereduced > 0) && preReduced.ReduceStatus == types.ReduceStatusNotValid {
		//如果客户还款金额大于预减免应还总金额
		amountWithPrereduced, _ := repayplan.CaculateRepayTotalAmountWithPreReducedByRepayPlan(repayPlan)
		logs.Debug("[repayRollLoan] 展单=结清减免最低应还总额:", amountWithPrereduced, "当前还款额:", amount)
		if amount >= amountWithPrereduced {
			//更新预减免为生效状态
			preReduced.ReduceStatus = types.ReduceStatusValid
			preReduced.GraceInterestReduced = preReduced.GracePeriodInterestPrededuced
			preReduced.PenaltyReduced = preReduced.PenaltyPrereduced

			preReduced.ConfirmTime = timetag
			preReduced.Utime = timetag
			cols := []string{"reduce_status", "grace_interest_reduced", "penalty_reduced", "confirm_time", "utime"}
			models.OrmUpdate(&preReduced, cols)

			//更新还款计划
			olderRepayPlan := repayPlan
			repayPlan.GracePeriodInterestReduced = repayPlan.GracePeriodInterestReduced + preReduced.GracePeriodInterestPrededuced
			repayPlan.PenaltyReduced = repayPlan.PenaltyReduced + preReduced.PenaltyPrereduced
			models.UpdateRepayPlan(&repayPlan)
			models.OpLogWrite(0, repayPlan.Id, models.OpCodeRepayPlanUpdate, repayPlan.TableName(), olderRepayPlan, repayPlan)

			//展期订单置为失效
			rollOrder, _ := models.GetRollOrder(dataOrder.Id)
			olderTollOrder := rollOrder
			rollOrder.CheckStatus = types.LoanStatusRollFail
			rollOrder.CheckTime = timetag
			rollOrder.Utime = timetag
			models.UpdateOrder(&rollOrder)
			models.OpLogWrite(0, rollOrder.Id, models.OpCodeOrderUpdate, rollOrder.TableName(), olderTollOrder, rollOrder)

		} else {
			//预减免失效
			preReduced.ReduceStatus = types.ReduceStatusInvalid
			preReduced.InvalidReason = types.ClearReducedInvalidReasonNotClear
			preReduced.ConfirmTime = timetag
			preReduced.Utime = timetag
			cols := []string{"reduce_status", "invalid_reason", "confirm_time", "Utime"}
			models.OrmUpdate(&preReduced, cols)
		}
	}

	minRepayTotal := dataOrder.MinRepayAmount
	trans, err := dao.GetFrozenTrans(dataOrder.Id)
	tmpAmount := amount

	for _, v := range trans {
		tmpAmount += v.Total
	}

	if tmpAmount < minRepayTotal {
		eAccount, _ := models.GetLastestActiveEAccountByVacompanyType(userAccountId, vaCompanyCode)

		payment := models.Payment{}
		payment.Id, _ = device.GenerateBizId(types.PaymentBiz)
		payment.Amount = amount
		payment.OrderId = dataOrder.Id
		payment.PayType = types.PayTypeMoneyIn
		payment.VaCompanyCode = vaCompanyCode
		if paymentCode != "" {
			payment.VaCode = paymentCode
		} else {
			payment.VaCode = eAccount.BankCode + eAccount.EAccountNumber
		}
		payment.UserAccountId = tools.Int642Str(userAccountId)
		payment.UserBankCode = bankCode
		payment.Ctime = timetag
		payment.Utime = timetag
		payment.AddPayment(&payment)

		//用户入账
		eInTrans := models.User_E_Trans{}
		eInTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
		eInTrans.OrderId = dataOrder.Id
		eInTrans.PaymentId = payment.Id
		eInTrans.UserAccountId = userAccountId
		eInTrans.VaCompanyCode = vaCompanyCode
		eInTrans.Total = amount
		eInTrans.PayType = types.PayTypeMoneyIn
		eInTrans.IsFrozen = 1
		eInTrans.CallbackJson = callbackStr
		eInTrans.Ctime = timetag
		eInTrans.Utime = timetag
		eInTrans.AddEtrans(&eInTrans)

		logs.Warn("[repayRollLoan] tmpAmount < minRepayTotal orderId;%d, tmpAmount:%d, minRepayTotal:%d", dataOrder.Id, tmpAmount, minRepayTotal)

		return 0, nil
	}

	transList := make([]models.User_E_Trans, 0)
	for _, v := range trans {
		//用户出账
		eTrans := models.User_E_Trans{}
		eTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
		eTrans.OrderId = dataOrder.Id
		eTrans.UserAccountId = v.UserAccountId
		eTrans.VaCompanyCode = v.VaCompanyCode
		eTrans.PayType = types.PayTypeMoneyOut
		eTrans.Ctime = timetag
		eTrans.Utime = timetag
		generateRepayTrans(dataOrder, &eTrans, &repayPlan, v.Total)
		transList = append(transList, eTrans)
	}

	if len(transList) > 0 {
		//unfrozen

		for _, v := range trans {
			v.IsFrozen = 0
			v.Update()
		}

		for _, v := range transList {
			v.AddEtrans(&v)
		}

		models.UpdateRepayPlan(&repayPlan)
	}

	return doOrderRoll(userAccountId, amount, bankCode, paymentCode, vaCompanyCode, callbackStr, dataOrder, &repayPlan)
}

// RepayLoanV2 发生还款事件 读取配置的product 信息按照还款顺序还款
func RepayLoan(userAccountId int64, amount int64, bankCode string, paymentCode string, vaCompanyCode int, callbackStr string) (int64, error) {
	dataOrder, err := dao.AccountLastLoanOrder(userAccountId)
	if err != nil {
		//没有订单
		errStr := fmt.Sprintf("RepayLoanV2 AccountLastLoanOrder does not have this recorde by userAccountId: %d, err: %s", userAccountId, err.Error())
		logs.Error(errStr)
		return 0, fmt.Errorf(errStr)
	}
	//加入委外还款队列
	serveentrust.EntrustRepayList(dataOrder.Id)
	if dataOrder.CheckStatus == types.LoanStatusWaitRepayment ||
		dataOrder.CheckStatus == types.LoanStatusPartialRepayment ||
		dataOrder.CheckStatus == types.LoanStatusOverdue {
		balance, err := repayNormalLoan(userAccountId, amount, bankCode, paymentCode, vaCompanyCode, callbackStr, &dataOrder)
		IncreaseBalanceByRefund(userAccountId, balance)
		logs.Debug("[RepayLoan] repayNormalLoan return:%v", err)

		return dataOrder.Id, err

	} else if dataOrder.CheckStatus == types.LoanStatusRolling {
		balance, err := repayRollLoan(userAccountId, amount, bankCode, paymentCode, vaCompanyCode, callbackStr, &dataOrder)
		logs.Debug("[RepayLoan] repayRollLoan return:%v", err)

		if err == nil {
			IncreaseBalanceByRefund(userAccountId, balance)
			return dataOrder.Id, err
		}

		//出错一律按逾期订单处理
		dataOrder.CheckStatus = types.LoanStatusOverdue
		balance, err = repayNormalLoan(userAccountId, amount, bankCode, paymentCode, vaCompanyCode, callbackStr, &dataOrder)
		IncreaseBalanceByRefund(userAccountId, balance)
		logs.Debug("[RepayLoan] repayNormalLoan return:%v", err)

		//展期订单置为失效
		rollOrder, rollErr := models.GetRollOrder(dataOrder.Id)
		if rollErr == nil {
			olderTollOrder := rollOrder
			rollOrder.CheckStatus = types.LoanStatusRollFail
			rollOrder.CheckTime = tools.GetUnixMillis()
			rollOrder.Utime = tools.GetUnixMillis()
			models.UpdateOrder(&rollOrder)
			models.OpLogWrite(0, olderTollOrder.Id, models.OpCodeOrderUpdate, rollOrder.TableName(), olderTollOrder, rollOrder)
		}

		return dataOrder.Id, err
	} else {
		// 这小子结清后多还钱了。直接记在最后一个订单上
		record(dataOrder, amount, bankCode, paymentCode, vaCompanyCode, callbackStr)
		err := IncreaseBalanceByRefund(userAccountId, amount)
		if err == nil {
			logs.Warn("[RepayLoan] 这小子结清后多还钱了。userAccountId：%d amount:%d dataOrder:%#v", userAccountId, amount, dataOrder)
			return dataOrder.Id, err
		}

		errStr := fmt.Sprintf("RepayLoan dataorder status is wrong, the order is %#v amount:%d", dataOrder, amount)
		logs.Error(errStr)
		err = fmt.Errorf(errStr)
	}
	return dataOrder.Id, err
}

func record(dataOrder models.Order, amount int64, bankCode string, paymentCode string, vaCompanyCode int, callbackStr string) error {
	eAccount, _ := models.GetLastestActiveEAccountByVacompanyType(dataOrder.UserAccountId, vaCompanyCode)

	//用户出账
	timetag := tools.GetUnixMillis()
	eTrans := models.User_E_Trans{}
	eTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
	eTrans.OrderId = dataOrder.Id
	eTrans.UserAccountId = dataOrder.UserAccountId
	eTrans.VaCompanyCode = vaCompanyCode
	eTrans.PayType = types.PayTypeMoneyOut
	eTrans.Ctime = timetag
	eTrans.Utime = timetag
	eTrans.Balance = amount
	eTrans.AddEtrans(&eTrans)

	payment := models.Payment{}
	payment.Id, _ = device.GenerateBizId(types.PaymentBiz)
	payment.Amount = amount
	payment.OrderId = dataOrder.Id
	payment.PayType = types.PayTypeMoneyIn
	payment.VaCompanyCode = vaCompanyCode
	if paymentCode != "" {
		payment.VaCode = paymentCode
	} else {
		payment.VaCode = eAccount.BankCode + eAccount.EAccountNumber
	}
	payment.UserAccountId = tools.Int642Str(dataOrder.UserAccountId)
	payment.UserBankCode = bankCode
	payment.Ctime = timetag
	payment.Utime = timetag
	payment.AddPayment(&payment)
	//财务记录

	eInTrans := models.User_E_Trans{}
	eInTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
	eInTrans.OrderId = dataOrder.Id
	eInTrans.PaymentId = payment.Id
	eInTrans.UserAccountId = dataOrder.UserAccountId
	eInTrans.VaCompanyCode = vaCompanyCode
	eInTrans.Total = amount
	eInTrans.PayType = types.PayTypeMoneyIn
	eInTrans.CallbackJson = callbackStr
	eInTrans.Ctime = timetag
	eInTrans.Utime = timetag
	eInTrans.AddEtrans(&eInTrans)
	//用户入账

	return nil
}

func generateRepayTrans(order *models.Order, trans *models.User_E_Trans, repayPlan *models.RepayPlan, amount int64) (int64, error) {
	product, err := models.GetProduct(order.ProductId)
	if err != nil {
		//没有产品
		errStr := fmt.Sprintf("RepayLoanV2 GetProduct does not get data, productId is %d, err is %s", order.ProductId, err.Error())
		logs.Error(errStr)
		return 0, fmt.Errorf(errStr)
	}

	//应还服务费
	needFee := repayPlan.ServiceFee - repayPlan.ServiceFeePayed
	//应还罚息
	needPenalty := repayPlan.Penalty - repayPlan.PenaltyPayed - repayPlan.PenaltyReduced
	//应还宽限息
	needGracePeriodInterest := repayPlan.GracePeriodInterest - repayPlan.GracePeriodInterestPayed - repayPlan.GracePeriodInterestReduced
	//应还利息
	needInterest := repayPlan.Interest - repayPlan.InterestPayed
	//应还本金 考虑减免情况
	needAmount := repayPlan.Amount - repayPlan.AmountPayed - repayPlan.AmountReduced

	needTotal := needPenalty + needInterest + needAmount + needGracePeriodInterest + needFee

	logs.Debug("[RepayLoanV2] needFee%d needPenalty%d needGracePeriodInterest%d needInterest%d needAmount%d needTotal%d",
		needFee, needPenalty, needGracePeriodInterest, needInterest, needAmount, needTotal)

	if amount >= needTotal {
		//本次就可以结清

		trans.ServiceFee = needFee
		trans.GracePeriodInterest = needGracePeriodInterest
		trans.Interest = needInterest
		trans.Penalty = needPenalty
		trans.Amount = needAmount
		//用户出账相关费用更新

		repayPlan.ServiceFeePayed += needFee
		repayPlan.InterestPayed += needInterest
		repayPlan.PenaltyPayed += needPenalty
		repayPlan.GracePeriodInterestPayed += needGracePeriodInterest
		repayPlan.AmountPayed += needAmount
		//还款计划全部还清

		return needTotal, nil

	} else {
		if 0 < amount && amount < needTotal {
			//部分还款  考虑还款顺序
			repayOrderStr := strings.Split(product.RepayOrder, ";")
			logs.Debug("[RepayLoanV2] repayOrderStr %v ", repayOrderStr)

			for amount > 0 {
				logs.Debug("[RepayLoanV2] entry loop amount %d", amount)
				if len(repayOrderStr) == 0 {
					break
				}

				//根据还款顺序还款
				v := repayOrderStr[0]
				logs.Debug("[RepayLoanV2] before switch  v %s amount %d", v, amount)
				switch v {
				case types.ProductOrderAmount:
					{
						// 如果还款总额大于应还的本金 则还玩全部本金还有结余 否则 全部还给本金跳出for循环
						if amount > needAmount {
							repayPlan.AmountPayed += needAmount
							amount -= needAmount
							trans.Amount = needAmount
							repayOrderStr = repayOrderStr[1:]
						} else {
							repayPlan.AmountPayed += amount
							trans.Amount = amount
							amount = 0
						}
					}
				case types.ProductOrderInterest:
					{
						if amount > needInterest {
							repayPlan.InterestPayed += needInterest
							amount -= needInterest
							trans.Interest = needInterest
							repayOrderStr = repayOrderStr[1:]
						} else {
							repayPlan.InterestPayed += amount
							trans.Interest = amount
							amount = 0
						}
					}
				case types.ProductOrderFee:
					{
						if amount > needFee {
							repayPlan.ServiceFeePayed += needFee
							amount -= needFee
							trans.ServiceFee = needFee
							repayOrderStr = repayOrderStr[1:]
						} else {
							repayPlan.ServiceFeePayed += amount
							trans.ServiceFee = amount
							amount = 0
						}
					}
				case types.ProductOrderGraceInterest:
					{
						if amount > needGracePeriodInterest {
							repayPlan.GracePeriodInterestPayed += needGracePeriodInterest
							amount -= needGracePeriodInterest
							trans.GracePeriodInterest = needGracePeriodInterest
							repayOrderStr = repayOrderStr[1:]
						} else {
							repayPlan.GracePeriodInterestPayed += amount
							trans.GracePeriodInterest = amount
							amount = 0
						}
					}
				case types.ProductOrderPenalty:
					{
						if amount > needPenalty {
							repayPlan.PenaltyPayed += needPenalty
							amount -= needPenalty
							trans.Penalty = needPenalty
							repayOrderStr = repayOrderStr[1:]
						} else {
							repayPlan.PenaltyPayed += amount
							trans.Penalty = amount
							amount = 0
						}
					}
				case types.ProductOrderForfeitPenalty:
					{
						repayOrderStr = repayOrderStr[1:]
					}
				default:
					{
						logs.Warn("RepayLoanV2 repayOrder type undefine. v:%v  repayOrderStr:%v :", v, repayOrderStr)
						repayOrderStr = repayOrderStr[1:]
					}
				}
			}

			return needTotal, nil

			//改写订单状态，订单为部分还款
		} else {
			errStr := fmt.Sprintf("RepayLoanV2 invalid repay money,the orderid is %d ", order.Id)
			logs.Error(errStr)
			return 0, fmt.Errorf(errStr)
		}
	}
}

func IsTrialCalOrApply() (is_trial_cal bool) {
	currTime, _ := time.Parse("15:04:05", tools.MDateMHSHMS(tools.GetUnixMillis()))
	startTime, _ := time.Parse("15:04:05", "08:00:00")
	endTime, _ := time.Parse("15:04:05", "23:59:59")
	timeQuantum := config.ValidItemString("overdue_roll_time_quantum")
	if len(timeQuantum) > 0 {
		times := strings.Split(timeQuantum, "-")
		if len(times) >= 2 {
			startTime, _ = time.Parse("15:04:05", times[0])
			endTime, _ = time.Parse("15:04:05", times[1])
		}
	}

	if currTime.After(startTime) && endTime.After(currTime) {
		is_trial_cal = true
	}

	return
}

func IsOrderExtension(order models.Order) (isExtension bool) {
	if order.CheckStatus == types.LoanStatusOverdue && IsOrderCanRoll(order) && IsTrialCalOrApply() {
		isExtension = true
	}
	return
}

// 客户的最后一条有效订单,如果状态是[逾期],逾期天数>=n,展期次数<times,剩余应还本金/全部应还本金>N,开启展期功能
func IsOrderCanRoll(order models.Order) bool {
	repayPlan, err := models.GetLastRepayPlanByOrderid(order.Id)
	if err != nil {
		//没有还款计划
		logs.Error("[IsOrderCanRoll] GetLastRepayPlanByOrderid error, orderid:%d, err:%v", order.Id, err)
		return false
	}

	overdueDay, err := CalculateOverdue(repayPlan.RepayDate)
	if err != nil {
		//没有还款计划
		logs.Error("[IsOrderCanRoll] CalculateOverdue day error, orderid:%d, err:%v", order.Id, err)
		return false
	}

	minDay, _ := config.ValidItemInt("overdue_roll_min_day")
	if overdueDay < int64(minDay) {
		logs.Info("[IsOrderCanRoll] overdueDay < minDay orderid:%d, overdueDay:%d, minDay:%d", order.Id, overdueDay, minDay)
		return false
	}

	minPercent, _ := config.ValidItemFloat64("overdue_roll_min_percent")
	if float64(repayPlan.Amount-repayPlan.AmountPayed-repayPlan.AmountReduced)/float64(repayPlan.Amount) < minPercent {
		logs.Info("[IsOrderCanRoll] (amount-amountPayed)/amount < minAmount orderid:%d, amountPayed:%d, amount:%d, minPercent:%v",
			order.Id, repayPlan.AmountPayed, repayPlan.Amount, minPercent)
		return false
	}

	maxTimes, _ := config.ValidItemInt("overdue_roll_max_times")
	if order.RollTimes >= maxTimes {
		logs.Info("[IsOrderCanRoll] roll times >= maxTimes orderid:%d, times:%d, maxTimes:%d", order.Id, order.RollTimes, maxTimes)
		return false
	}

	return true
}

func CalcRollRepayAmount(order models.Order) (int, int64, int64, error) {

	repayPlan, err := models.GetLastRepayPlanByOrderid(order.Id)
	if err != nil {
		//没有还款计划
		logs.Error("[CalcRollRepayAmount] GetLastRepayPlanByOrderid error, orderid:%d, err:%v", order.Id, err)
		return 0, 0, 0, err
	}

	p, err := ProductRollSuitables()
	if err != nil {
		logs.Error("[CalcRollRepayAmount] ProductSuitablesByPeriod error, orderid:%d, err:%v", order.Id, err)
		return 0, 0, 0, err
	}

	period := p.MinPeriod
	total, interest, fee := repayplan.CalcRepayInfoV3(order.Amount, p, period)

	//应还服务费
	needFee := repayPlan.ServiceFee - repayPlan.ServiceFeePayed
	//应还罚息
	needPenalty := repayPlan.Penalty - repayPlan.PenaltyPayed - repayPlan.PenaltyReduced
	//应还宽限息
	needGracePeriodInterest := repayPlan.GracePeriodInterest - repayPlan.GracePeriodInterestPayed - repayPlan.GracePeriodInterestReduced
	//应还利息
	needInterest := repayPlan.Interest - repayPlan.InterestPayed

	min := needInterest + needPenalty + needGracePeriodInterest + fee + needFee + interest

	return period, min, total, nil
}

func HandleRollOrder(orderId int64) error {
	logs.Debug("[HandleRollOrder] begin orderid:%d", orderId)
	timetag := tools.GetUnixMillis()

	order, err := models.GetOrder(orderId)
	if err != nil {
		logs.Error("[HandleRollOrder] GetOrder error orderid:%d, err:%v", orderId, err)
		return err
	}

	if order.CheckStatus != types.LoanStatusRolling {
		logs.Error("[HandleRollOrder] order status error orderid:%d, status:%v", orderId, order.CheckStatus)
		return nil
	}

	repayPlan, err := models.GetLastRepayPlanByOrderid(order.Id)
	if err != nil {
		//没有还款计划
		errStr := fmt.Sprintf("[HandleRollOrder] GetLastRepayPlanByOrderid does not get data, orderid is %d, err is %s", order.Id, err.Error())
		logs.Error(errStr)
		return fmt.Errorf(errStr)
	}

	olderOrder := order

	trans, err := dao.GetFrozenTrans(orderId)
	amount := int64(0)
	transList := make([]models.User_E_Trans, 0)
	for _, v := range trans {
		amount += v.Total

		//用户出账
		eTrans := models.User_E_Trans{}
		eTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
		eTrans.OrderId = order.Id
		eTrans.UserAccountId = v.UserAccountId
		eTrans.VaCompanyCode = v.VaCompanyCode
		eTrans.PayType = types.PayTypeMoneyOut
		eTrans.Ctime = timetag
		eTrans.Utime = timetag
		generateRepayTrans(&order, &eTrans, &repayPlan, v.Total)
		transList = append(transList, eTrans)
	}

	minRepayTotal := order.MinRepayAmount

	if err != nil || amount < minRepayTotal {
		//frozen
		logs.Error("[HandleRollOrder] amount < minRepayTotal orderid:%d, amount:%d, minRepayTotal:%d, err:%v ", orderId, amount, minRepayTotal, err)
		for _, v := range trans {
			v.IsFrozen = 0
			v.Update()
		}

		for _, v := range transList {
			v.AddEtrans(&v)
		}

		models.UpdateRepayPlan(&repayPlan)
		//更新还款计划

		if amount > 0 {
			order.RepayTime = timetag
		}

		order.Utime = timetag
		order.CheckStatus = types.LoanStatusOverdue

		models.UpdateOrder(&order)

		HandleOverdueCase(order.Id)

		monitor.IncrOrderCount(order.CheckStatus)

		//更新订单状态
		models.OpLogWrite(0, order.Id, models.OpCodeOrderUpdate, order.TableName(), olderOrder, order)

		rollOrder, err := models.GetRollOrder(order.Id)
		if err != nil {
			logs.Error("[HandleRollOrder] GetRollOrder error orderId:%d, err:%v ", order.Id, err)
			return nil
		}

		olderTollOrder := rollOrder

		rollOrder.CheckStatus = types.LoanStatusRollFail
		rollOrder.CheckTime = timetag
		rollOrder.Utime = timetag
		models.UpdateOrder(&rollOrder)

		monitor.IncrOrderCount(rollOrder.CheckStatus)

		models.OpLogWrite(0, rollOrder.Id, models.OpCodeOrderUpdate, rollOrder.TableName(), olderTollOrder, rollOrder)

		return nil
	}

	//理论上不可能
	str := fmt.Sprintf("[HandleRollOrder] order amout >= minRepayTotal orderId:%d, amout:%d, minRepayTotal:%d", orderId, amount, minRepayTotal)
	logs.Error(str)
	return fmt.Errorf(str)
}

func DoAutoReduce(order *models.Order, repayPlan *models.RepayPlan, autoReducedAmount int64) error {

	eTrans := models.User_E_Trans{}
	newPlan := *repayPlan
	oldPlan := *repayPlan
	_, err := generateRepayTrans(order, &eTrans, &newPlan, autoReducedAmount)
	if err != nil {
		logs.Error("[DoAutoReduce] generateRepayTrans err:%v", err)
		return err
	}

	if eTrans.Amount > 0 {
		repayPlan.AmountReduced += eTrans.Amount
	}

	if eTrans.GracePeriodInterest > 0 {
		repayPlan.GracePeriodInterestReduced += eTrans.GracePeriodInterest
	}

	if eTrans.Penalty > 0 {
		repayPlan.PenaltyReduced += eTrans.Penalty
	}

	logs.Info("[DoAutoReduce] autoReducedAmount:%d repayPlan:%#v eTrans:%#v", autoReducedAmount, repayPlan, eTrans)

	// 讲道理此时应该处于结清状态,因为已做好了减免   amount应该为0
	amount, err := repayplan.CaculateRepayTotalAmountByRepayPlan(*repayPlan)
	if amount > 0 {
		logs.Warn("[DoAutoReduce] 自动减免失败。还款计划不符合预期。repayPlan:%#v", repayPlan)
		*repayPlan = oldPlan
		return err
	}

	caseOver, _ := models.OneOverdueCaseByOrderID(order.Id)
	// 写减免记录
	tag := tools.GetUnixMillis()
	id, _ := device.GenerateBizId(types.ReduceRecordBiz)
	reduce := models.ReduceRecordNew{
		Id:                   id,
		OrderId:              order.Id,
		UserAccountId:        order.UserAccountId,
		ApplyUid:             0,
		ConfirmUid:           0,
		AmountReduced:        eTrans.Amount,
		PenaltyReduced:       eTrans.Penalty,
		GraceInterestReduced: eTrans.GracePeriodInterest,
		ReduceType:           types.ReduceTypeAuto,
		ReduceStatus:         types.ReduceStatusValid,
		OpReason:             "系统减免",
		CaseID:               caseOver.Id,
		ApplyTime:            tag,
		ConfirmTime:          tag,
		Ctime:                tag,
		Utime:                tag,
	}

	_, err = models.OrmInsert(&reduce)
	if err != nil {
		logs.Error("[DoAutoReduce] OrmInsert err:%v reduce:%#v", err, reduce)
	}
	return err
}

func CanAutoReduce(order *models.Order, repayPlan *models.RepayPlan, repayAmount int64) (flag bool, autoReducedAmount int64) {

	//1、计算剩余金额
	amountRemained, _ := repayplan.CaculateRepayTotalAmountByRepayPlan(*repayPlan)

	// 剩余未还 - 本次还款 = 实际剩余
	amountRemained -= repayAmount

	logs.Info("[CanAutoReduce] amountRemained:%d repayAmount:%d repayPlan:%#v order:%#v", amountRemained, repayAmount, repayPlan, order)
	// 已结清啦不需要自动减免了
	if 0 >= amountRemained {
		return false, 0
	}

	//2、根据配置 是否满足自动减免
	if types.IsOverdueNo == order.IsOverdue {
		// 未逾期状态使用M值判断
		m, _ := config.ValidItemInt64("auto_reduce_nocase_m")
		logs.Debug("[CanAutoReduce] auto_reduce_nocase_m:%v amountRemained:%d", m, amountRemained)
		if amountRemained <= m {
			logs.Info("[CanAutoReduce] can auto reduce . amountRemained:%d ", amountRemained)
			return true, amountRemained
		}
	} else if types.IsOverdueYes == order.IsOverdue {
		// 逾期状态使用N值判断
		n, _ := config.ValidItemFloat64("auto_reduce_case_n")

		logs.Debug("[CanAutoReduce] auto_reduce_case_n:%v result:%v", n, float64(amountRemained)/float64(repayPlan.Amount))
		if float64(amountRemained)/float64(repayPlan.Amount) <= n {
			logs.Info("[CanAutoReduce] can auto reduce . amountRemained:%d", amountRemained)
			return true, amountRemained
		}
	} else {
		logs.Error("[CanAutoReduce] unknow err.  order.IsOverdue:%v", order.IsOverdue)
	}
	return false, 0
}

func GetFailedDisburseOrderReason(OrderId int64) (desc string) {
	desc = "其他原因"
	data, _ := models.GetLastestDisburseInvorkLogByPkOrderId(OrderId)
	if data.FailureCode != "" {
		desc = data.FailureCode
	}
	return
}
