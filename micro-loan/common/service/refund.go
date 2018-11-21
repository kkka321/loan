package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/dao"
	"micro-loan/common/lib/device"
	"micro-loan/common/models"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/thirdparty/xendit"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func DoRefundToOrder(refund *models.Refund) error {

	// 1、创建退款订单
	refund.Id, _ = device.GenerateBizId(types.RefundBiz)
	refund.RefundType = types.RefundTypeToOrder
	refund.Ctime = tools.GetUnixMillis()
	refund.Utime = refund.Ctime
	_, err := refund.Add()
	if err != nil {
		logs.Error("[DoRefundToOrder] add refund record err:%s refund:%#v", err, refund)
		return err
	}

	// 2、冻结客户的资金 并减小余额
	err = ReduceBalanceByRefund(refund)
	if err != orm.ErrNoRows && err != nil {
		logs.Error("[DoRefundToOrder] ReduceBalanceByRefund err:%s refund:%#v", err, refund)
		return err
	}

	// 3、更新退款订单
	refund.CheckStatus = int(types.RefundStatusProcessing)
	refund.Utime = tools.GetUnixMillis()
	_, err = refund.Update("check_status", "utime")
	if err != nil {
		logs.Error("[DoRefundToOrder] update refund record err:%s refund:%#v", err, refund)
		return err
	}

	// 4、还款到订单
	err = doRefundToOrder(refund)
	if err != nil {
		errI := SetRefundInvalid(refund)
		errN := RestoreBalanceByRefund(refund)

		logs.Error("[DoRefundToOrder] doRefundToOrder err:%s errI:%v errN:%v refund:%#v", err, errI, errN, refund)
		err = fmt.Errorf("[DoRefundToOrder] doRefundToOrder err:%s errI:%v errN:%v refund:%#v", err, errI, errN, refund)
		return err
	}

	return refundSuccess(refund, types.RefundTypeToOrder)
}

func DoRefundToOtherAccount(refund *models.Refund) error {
	// 1、创建退款订单
	refund.Id, _ = device.GenerateBizId(types.RefundBiz)
	refund.RefundType = types.RefundTypeToOtherAccount
	refund.Ctime = tools.GetUnixMillis()
	refund.Utime = refund.Ctime
	_, err := refund.Add()
	if err != nil {
		logs.Error("[DoRefundToOrder] add refund record err:%s refund:%#v", err, refund)
		return err
	}

	// 2、冻结客户的资金 并减小余额
	err = ReduceBalanceByRefund(refund)
	if err != orm.ErrNoRows && err != nil {
		logs.Error("[DoRefundToOrder] ReduceBalanceByRefund err:%s refund:%#v", err, refund)
		return err
	}

	// 3、更新退款订单
	refund.CheckStatus = int(types.RefundStatusProcessing)
	refund.Utime = tools.GetUnixMillis()
	_, err = refund.Update("check_status", "utime")
	if err != nil {
		logs.Error("[DoRefundToOrder] update refund record err:%s refund:%#v", err, refund)
		return err
	}

	// 4、还款到他人账户
	err = IncreaseBalanceByRefund(refund.ReleatedOrder, refund.Amount)
	if err != nil {
		errI := SetRefundInvalid(refund)
		errN := RestoreBalanceByRefund(refund)

		logs.Error("[DoRefundToOrder] doRefundToOrder err:%s errI:%v errN:%v refund:%#v", err, errI, errN, refund)
		err = fmt.Errorf("[DoRefundToOrder] doRefundToOrder err:%s errI:%v errN:%v refund:%#v", err, errI, errN, refund)
		return err
	}
	return refundSuccess(refund, types.RefundTypeToOtherAccount)
}

func DoRefundToBankCard(refund *models.Refund, resIds []int64) error {

	// 1、创建退款订单
	refund.Id, _ = device.GenerateBizId(types.RefundBiz)
	refund.Ctime = tools.GetUnixMillis()
	refund.Utime = refund.Ctime
	refund.CallTime = refund.Ctime
	refund.RefundType = types.RefundTypeToBankCard
	_, err := refund.Add()
	if err != nil {
		logs.Error("[DoRefundToBankCard] add refund record err:%s refund:%#v", err, refund)
		return err
	}

	//2、保存还款凭证
	refundImage := models.RefundImage{
		UserAccountId: refund.UserAccountId,
		RefundId:      refund.Id,
		Image0Id:      resIds[0],
		Image1Id:      resIds[1],
		Image2Id:      resIds[2],
		Image3Id:      resIds[3],
		Image4Id:      resIds[4],
		Ctime:         tools.GetUnixMillis(),
	}
	_, err = refundImage.Add()
	if err != nil {
		logs.Error("[DoRefundToBankCard] add refundImage record err:%s refundImage:%#v", err, refundImage)
		SetRefundInvalid(refund)
		return err
	}

	// 3、冻结客户的资金 并减小余额
	err = ReduceBalanceByRefund(refund)
	if err != orm.ErrNoRows && err != nil {
		logs.Error("[DoRefundToBankCard] ReduceBalanceByRefund err:%s refund:%#v", err, refund)
		return err
	}

	// 4、更新退款订单
	refund.CheckStatus = int(types.RefundStatusProcessing)
	refund.CallTime = tools.GetUnixMillis()
	refund.Utime = tools.GetUnixMillis()
	_, err = refund.Update("check_status", "utime", "call_time")
	if err != nil {
		logs.Error("[DoRefundToBankCard] update refund record err:%s refund:%#v", err, refund)
		return err
	}

	// 5、调用第三方退款
	invokeId, err := CreateRefund(refund)
	invoke, _ := models.OneDisburseInvorkLogByPkId(invokeId)

	if invokeId == 0 || (invokeId != 0 && len(invoke.DisbursementId) > 0 && err != nil) {
		errI := SetRefundInvalid(refund)
		errN := RestoreBalanceByRefund(refund)

		logs.Error("[DoRefundToBankCard] doRefundToOrder err:%s errI:%v errN:%v refund:%#v", err, errI, errN, refund)
		err = fmt.Errorf("[DoRefundToBankCard] doRefundToOrder err:%s errI:%v errN:%v refund:%#v", err, errI, errN, refund)
		return err

	} else if invokeId != 0 && len(invoke.DisbursementId) == 0 && err != nil {
		// 第三方调用id 为空可能是超时
		logs.Warn("[handleWait4LoanOrder] 放款有可能因超时而失败,请检查. refundID:%d, err:%v invoke:%#v", refund.Id, err, invoke)
	} else if err != nil {
		logs.Error("[handleWait4LoanOrder] 未知错误,请检查. refundID:%d, err:%v invoke:%#v", refund.Id, err, invoke)
	}

	return nil
}

func refundSuccess(refund *models.Refund, refundType int) (err error) {
	//3、更新状态
	refund.CheckStatus = int(types.RefundStatusSuccess)
	if refundType == types.RefundTypeToBankCard {
		refund.ResponseTime = tools.GetUnixMillis()
	}
	refund.Utime = tools.GetUnixMillis()
	_, err = refund.Update("check_status", "utime", "response_time")
	if err != nil {
		logs.Error("[refundSuccess] update refund record err:%s refund:%#v", err, refund)
		return err
	}

	//4、冻结资金减小
	err = ReduceFrozenBalanceByRefund(refund)
	if err != nil && err != orm.ErrNoRows {
		err = fmt.Errorf("[refundSuccess] ReduceFrozenBalanceByRefund. err:%s refund:%#v", err, refund)
		logs.Error(err)
		return err
	}
	return nil
}

// 退款成功 减少冻结资金，更新退款状态 记账
func RefundDisburseCallback(accountId, refundId, amount int64, bankCode string, jsonData []byte) error {
	resp := xendit.XenditDisburseFundCallBackData{}
	json.Unmarshal(jsonData, &resp)

	refund, err := dao.GetRefund(refundId)
	if err != nil {
		err = fmt.Errorf("[RefundDisburseCallback] GetRefund err:%s refundId:%d", err, refundId)
		logs.Error(err)
		return err
	}

	//1、基本检查
	if refund.Amount != amount || refund.CheckStatus != int(types.RefundStatusProcessing) || accountId != refund.UserAccountId {
		err = fmt.Errorf("[RefundDisburseCallback] amount or status not match. amount:%d refund:%#v jsonData:%s", amount, refund, string(jsonData))
		logs.Error(err)
		return err
	}

	if resp.Status != "COMPLETED" {
		err = fmt.Errorf("[RefundDisburseCallback] refund status err, json:%s", string(jsonData))
		logs.Error(err)
		return err
	}

	tranData, err := models.GetMobiEtrans(resp.Id)
	if err != nil {
		//目前发现这种超时的情况，如果回调显示COMPLETED,我们直接向表中插入数据，避免查询再次超时
		if resp.Status == "COMPLETED" {
			exteralId, _ := tools.Str2Int64(resp.ExternalId)
			mobiEtrans := &models.Mobi_E_Trans{}
			mobiEtrans.UserAcccountId = exteralId
			mobiEtrans.VaCompanyCode = types.Xendit
			mobiEtrans.Amount = resp.Amount
			//向上取整，百位取整
			mobiEtrans.PayType = types.PayTypeMoneyOut
			mobiEtrans.BankCode = resp.BankCode
			mobiEtrans.AccountHolderName = resp.AccountHolderName
			mobiEtrans.DisbursementDescription = resp.DisbursementDescription
			mobiEtrans.DisbursementId = resp.Id
			mobiEtrans.Status = resp.Status
			mobiEtrans.CallbackJson = string(jsonData)
			mobiEtrans.Utime = tools.GetUnixMillis()
			mobiEtrans.Ctime = tools.GetUnixMillis()
			_, err = mobiEtrans.AddMobiEtrans(mobiEtrans)

			*tranData = *mobiEtrans
		}
	}

	// 校验回调信息
	if tranData.DisbursementId != resp.Id ||
		tranData.AccountHolderName != resp.AccountHolderName ||
		tranData.BankCode != resp.BankCode {
		errStr := fmt.Sprintf("[RefundDisburseCallback] data not matched [%s],[%s] [%s],[%s] [%s],[%s]", tranData.DisbursementId, resp.Id, tranData.AccountHolderName, resp.AccountHolderName, tranData.BankCode, resp.BankCode)
		logs.Error(errStr)
		return errors.New(errStr)
	}
	tranData.Status = resp.Status
	tranData.CallbackJson = string(jsonData)
	tranData.Utime = tools.GetUnixMillis()
	tranData.UpdateMobiEEtrans(tranData)

	//2、记账 包括payment和user_e_trans
	timetag := tools.GetUnixMillis()
	payment := models.Payment{}
	payment.Id, _ = device.GenerateBizId(types.PaymentBiz)
	payment.Amount = amount
	payment.OrderId = refund.Id
	payment.PayType = types.PayTypeRefundOut
	payment.VaCompanyCode = types.Xendit
	payment.UserAccountId = tools.Int642Str(accountId)
	payment.UserBankCode = bankCode
	payment.Ctime = timetag
	payment.Utime = timetag
	payment.AddPayment(&payment)

	//user_e_trans 本金入账
	eTrans := models.User_E_Trans{}
	eTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
	eTrans.OrderId = refund.Id
    eTrans.PaymentId = payment.Id
	eTrans.UserAccountId = accountId
	eTrans.VaCompanyCode = types.Xendit
	eTrans.Total = amount
	eTrans.CallbackJson = string(jsonData)
	eTrans.PayType = types.PayTypeRefundIn
	eTrans.Ctime = timetag
	eTrans.Utime = timetag
	eTrans.AddEtrans(&eTrans)

	//user_e_trans服务费出帐                payment服务费入账
	if refund.Fee > 0 {
		//paymentFee := models.Payment{}
		//paymentFee.Id, _ = device.GenerateBizId(types.PaymentBiz)
		//paymentFee.Amount = refund.Fee
		//paymentFee.OrderId = refund.Id
		//paymentFee.PayType = types.PayTypeRefundIn
		//paymentFee.VaCompanyCode = types.Xendit
		//paymentFee.UserAccountId = tools.Int642Str(accountId)
		//paymentFee.UserBankCode = bankCode
		//paymentFee.Ctime = timetag
		//paymentFee.Utime = timetag
		//paymentFee.AddPayment(&paymentFee)

		eTransFee := models.User_E_Trans{}
		eTransFee.Id, _ = device.GenerateBizId(types.UserETransBiz)
		eTransFee.OrderId = refund.Id
		eTransFee.UserAccountId = accountId
		eTransFee.VaCompanyCode = types.Xendit
		eTransFee.Total = refund.Fee
		eTransFee.CallbackJson = string(jsonData)
		eTransFee.PayType = types.PayTypeRefundOut
		eTransFee.Ctime = timetag
		eTransFee.Utime = timetag
		eTransFee.AddEtrans(&eTransFee)
	}

	err = refundSuccess(&refund, types.RefundTypeToBankCard)
	if err != nil {
		logs.Error("[RefundDisburseCallback] refundSuccess err:%v", err)
	}

	// 退款实际成功了 返回nil
	return nil
}

func CanRefund(accountId int64, refundAmount, fee int64) error {
	//1、获取客户信息
	accountBalance, err := dao.OneAccountBalanceByAccountId(accountId)
	if err != nil {
		err = fmt.Errorf("[CanRefund] OneAccountBalanceByAccountId accountId:%d err:%v", accountId, err)
		logs.Error(err)
		return err
	}

	//2、校验余额是否正确
	if refundAmount+fee > accountBalance.Balance {
		err = fmt.Errorf("[CanRefund] amount err. refundAmount:%d  fee:%d accountBalance:%#v", refundAmount, fee, accountBalance)
		logs.Error(err)
		return err
	}
	return nil
}

func CanRefundToBankCard(accountId int64) bool {
	order, err := dao.AccountLastLoanOrder(accountId)
	if err != nil {
		// 没有有效订单,状态为初始
		return true
	}

	if order.CheckStatus == types.LoanStatusWaitRepayment ||
		order.CheckStatus == types.LoanStatusOverdue ||
		order.CheckStatus == types.LoanStatusPartialRepayment ||
		order.CheckStatus == types.LoanStatusRolling ||
		order.CheckStatus == types.LoanStatusRollApply {
		return false
	}

	return true
}

// 退款申请时增加冻结金额
func ReduceBalanceByRefund(refund *models.Refund) error {
	obj := models.AccountBalance{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	sql := "update %s set balance = balance -%d , frozen_balance = frozen_balance + %d, utime=%d"
	sql = fmt.Sprintf(sql, obj.TableName(), refund.Amount+refund.Fee, refund.Amount+refund.Fee, tools.GetUnixMillis())

	where := " where account_id = %d"
	where = fmt.Sprintf(where, refund.UserAccountId)

	sql = fmt.Sprintf("%s %s", sql, where)

	err := o.Raw(sql).QueryRow()
	return err
}

// 退款成功回调时减少冻结金额
func ReduceFrozenBalanceByRefund(refund *models.Refund) error {
	obj := models.AccountBalance{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	sql := "update %s set frozen_balance = frozen_balance - %d, utime=%d"
	sql = fmt.Sprintf(sql, obj.TableName(), refund.Amount+refund.Fee, tools.GetUnixMillis())

	where := " where account_id = %d"
	where = fmt.Sprintf(where, refund.UserAccountId)

	sql = fmt.Sprintf("%s %s", sql, where)

	err := o.Raw(sql).QueryRow()
	return err
}

// 退款失败时减少冻结金额 增加余额
func RestoreBalanceByRefund(refund *models.Refund) error {
	obj := models.AccountBalance{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	sql := "update %s set  balance = balance +%d , frozen_balance = frozen_balance - %d, utime=%d"
	sql = fmt.Sprintf(sql, obj.TableName(), refund.Amount+refund.Fee, refund.Amount+refund.Fee, tools.GetUnixMillis())

	where := " where account_id = %d"
	where = fmt.Sprintf(where, refund.UserAccountId)

	sql = fmt.Sprintf("%s %s", sql, where)

	err := o.Raw(sql).QueryRow()
	return err
}

// 多还钱时 增加余额
func IncreaseBalanceByRefund(accountId int64, amount int64) error {
	if 0 == amount {
		return nil
	}
	obj := models.AccountBalance{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	// 查询是否有用户记录
	err := o.QueryTable(obj.TableName()).
		Filter("account_id", accountId).
		One(&obj)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[IncreaseBalanceByRefund] query account:%d amount:%d err:%v", accountId, amount, err)
		return err
	}

	// 有记录
	sql := ""
	if obj.AccountId != 0 {
		sql = "update %s set  balance = balance + %d, utime=%d"
		sql = fmt.Sprintf(sql, obj.TableName(), amount, tools.GetUnixMillis())
		where := " where account_id = %d"
		where = fmt.Sprintf(where, accountId)

		sql = fmt.Sprintf("%s %s", sql, where)

	} else {
		sql = "insert %s (account_id, balance, frozen_balance, ctime, utime ) value (%d,%d,%d,%d,%d)"
		sql = fmt.Sprintf(sql, obj.TableName(),
			accountId, amount, 0, tools.GetUnixMillis(), tools.GetUnixMillis())
	}
	err = o.Raw(sql).QueryRow()
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[IncreaseBalanceByRefund] insert account:%d amount:%d err:%v sql:%s", accountId, amount, err, sql)
		return err
	}
	return nil
}

func SetRefundInvalid(refund *models.Refund) error {
	refund.CheckStatus = int(types.RefundStatusFailed)
	refund.Utime = tools.GetUnixMillis()

	_, err := refund.Update("check_status", "utime")
	if err != nil {
		logs.Error("[SetRefundInvalid] update refund record err:%s refund:%#v", err, refund)
		err = fmt.Errorf("[SetRefundInvalid] update refund record err:%s refund:%#v", err, refund)
		return err
	}
	return nil
}

func doRefundToOrder(refund *models.Refund) (err error) {

	if refund == nil || refund.ReleatedOrder <= 0 {
		err = fmt.Errorf("[doRefundToOrder] refund err. ")
		return err
	}

	order, err := models.GetOrder(refund.ReleatedOrder)
	if err != nil {
		logs.Error("[doRefundToOrder] GetOrder err:%v", err)
		return err
	}

	if order.CheckStatus != types.LoanStatusWaitRepayment &&
		order.CheckStatus != types.LoanStatusOverdue &&
		order.CheckStatus != types.LoanStatusPartialRepayment &&
		order.CheckStatus != types.LoanStatusRolling {

		errStr := fmt.Sprintf("[doRefundToOrder] order status wrong.  order:%#v ", order)

		logs.Error(errStr)
		return errors.New(errStr)
	}

	repayPlan, err := models.GetLastRepayPlanByOrderid(refund.ReleatedOrder)
	if err != nil {
		logs.Error("[doRefundToOrder] GetRepayPlan err:%v", err)
		return err
	}

	amount, err := repayplan.CaculateRepayTotalAmountWithPreReducedByRepayPlan(repayPlan)
	if err != nil || amount < refund.Amount || 0 == amount {
		//退款金额太大了 暂不允许
		err = fmt.Errorf("[doRefundToOrder] amount err:%v  amount:%d repayPlan:%#v", err, amount, repayPlan)
		logs.Error(err)
		return err
	}

	if order.CheckStatus == types.LoanStatusRolling {
		_, err := repayRollLoan(refund.UserAccountId, refund.Amount, "", "", types.MobiRefundToOrder, "", &order)
		logs.Error("[doRefundToOrder] repayRollLoan return:%v", err)
		if err == nil {
			return err
		}

		//出错一律按逾期订单处理
		order.CheckStatus = types.LoanStatusOverdue
		_, err = repayNormalLoan(refund.UserAccountId, refund.Amount, "", "", types.MobiRefundToOrder, "", &order)
		logs.Error("[doRefundToOrder] repayNormalLoan return:%v", err)

		//展期订单置为失效
		rollOrder, rollErr := models.GetRollOrder(order.Id)
		if rollErr == nil {
			olderTollOrder := rollOrder
			rollOrder.CheckStatus = types.LoanStatusRollFail
			rollOrder.CheckTime = tools.GetUnixMillis()
			rollOrder.Utime = tools.GetUnixMillis()
			models.UpdateOrder(&rollOrder)
			models.OpLogWrite(0, olderTollOrder.Id, models.OpCodeOrderUpdate, rollOrder.TableName(), olderTollOrder, rollOrder)
		}

		return err
	} else {
		// 还钱
		_, err = repayNormalLoan(refund.UserAccountId, refund.Amount, "", "", types.MobiRefundToOrder, "", &order, true)
		if err != nil {
			err = fmt.Errorf("[doRefundToOrder] repayNormalLoan  err:%v refund:%#v", err, refund)
			logs.Error(err)
			return err
		}
		return nil
	}

}
