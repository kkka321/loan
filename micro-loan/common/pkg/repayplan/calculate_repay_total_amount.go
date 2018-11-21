package repayplan

import (
	"fmt"
	"math"
	"micro-loan/common/dao"
	"micro-loan/common/models"

	"github.com/astaxie/beego/logs"
)

// CaculateRepayTotalAmountByOrderID 根据 Order ID 计算应还总额
// 此方法适用于单个订单的计算, 内含数据库查询, 不适用于后台列表
// 原方法名 GetRepayAmount, 未发现被调用
func CaculateRepayTotalAmountByOrderID(orderID int64) (amount int64, err error) {
	repayPlan, err := models.GetLastRepayPlanByOrderid(orderID)
	if err != nil {
		logs.Error("There is no a repay plan for this order:", orderID)
		return
	}
	amount, err = CaculateRepayTotalAmountByRepayPlan(repayPlan)
	return
}

// CaculateRepayTotalAmountWithPreReducedByOrderID 根据 Order ID 计算结清减免应还总额（有条件减免）
// 此方法适用于单个订单的计算, 内含数据库查询, 不适用于后台列表
// 原方法名 GetRepayAmount, 未发现被调用
func CaculateRepayTotalAmountWithPreReducedByOrderID(orderID int64) (amount int64, err error) {
	repayPlan, err := models.GetLastRepayPlanByOrderid(orderID)
	if err != nil {
		logs.Error("There is no a repay plan for this order:", orderID)
		return
	}
	amount, err = CaculateRepayTotalAmountWithPreReducedByRepayPlan(repayPlan)
	return
}

// CaculateTotalGracePeriodAndPenaltyByOrderID 根据 Order ID 计算宽限期利息和罚息
// 此方法适用于单个订单的计算, 内含数据库查询, 不适用于后台列表
func CaculateTotalGracePeriodAndPenaltyByOrderID(orderID int64) (amount int64, err error) {
	repayPlan, err := models.GetLastRepayPlanByOrderid(orderID)
	if err != nil {
		logs.Error("There is no a repay plan for this order:", orderID)
		return
	}
	amount, err = CaculateTotalGracePeriodAndPenaltyByRepayPlan(repayPlan)
	return
}

// CaculateTotalPayedByOrderID 根据 Order ID 计算已还金额(已还本金,已还宽限期利息,已还罚息)
// 此方法适用于单个订单的计算, 内含数据库查询, 不适用于后台列表
func CaculateTotalPayedByOrderID(orderID int64) (amount int64, err error) {
	repayPlan, err := models.GetLastRepayPlanByOrderid(orderID)
	if err != nil {
		logs.Error("There is no a repay plan for this order:", orderID)
		return
	}
	amount, err = CaculateTotalPayedByRepayPlan(repayPlan)
	return
}

// CaculateTotalAmountByOrderID 根据 Order ID 计算账单总额(本金,宽限期利息,罚息)
// 此方法适用于单个订单的计算, 内含数据库查询, 不适用于后台列表
func CaculateTotalAmountByOrderID(orderID int64) (amount int64, err error) {
	repayPlan, err := models.GetLastRepayPlanByOrderid(orderID)
	if err != nil {
		logs.Error("There is no a repay plan for this order:", orderID)
		return
	}
	amount, err = CaculateTotalAmountByRepayPlan(repayPlan)
	return
}

// CaculateRepayTotalAmountByRepayPlan 根据 repayPlan model 计算应还总额
func CaculateRepayTotalAmountByRepayPlan(repayPlan models.RepayPlan) (amount int64, err error) {
	// 安全性检验, 防止空model
	if repayPlan.Amount <= 0 {
		err = fmt.Errorf("Invalid RepayPlan data, repayPlan.Amount must > 0 , repayPlan: %v", repayPlan)
		return
	}
	amount = CaculateRepayTotalAmount(repayPlan.Amount, repayPlan.AmountPayed, repayPlan.AmountReduced,
		repayPlan.GracePeriodInterest, repayPlan.GracePeriodInterestPayed, repayPlan.GracePeriodInterestReduced,
		repayPlan.Penalty, repayPlan.PenaltyPayed, repayPlan.PenaltyReduced)
	return
}

// CaculateAcutalRepayedTotalByRepayPlan 根据还款计划model指针，计算实还总额
func CaculateAcutalRepayedTotalByRepayPlan(repayPlan *models.RepayPlan) int64 {
	return repayPlan.AmountPayed + repayPlan.GracePeriodInterestPayed + repayPlan.InterestPayed + repayPlan.PenaltyPayed + repayPlan.ServiceFeePayed + repayPlan.PreInterestPayed
}

// CaculatePenaltyClearReducedByOrderId 计算结清减免金额
func CaculatePenaltyClearReducedByOrderId(orderId int64) (penaltyClearReduced int64, err error) {
	preReduced, err := dao.GetLastPrereducedByOrderid(orderId)
	if err == nil { // 结清减免, 并且未生效
		penaltyClearReduced = preReduced.GracePeriodInterestPrededuced + preReduced.PenaltyPrereduced
	}
	return
}

// CaculateRepayTotalAmountWithPreReducedByRepayPlan 根据 repayPlan model 计算结清减免应还总额（有条件减免 ）
func CaculateRepayTotalAmountWithPreReducedByRepayPlan(repayPlan models.RepayPlan) (amount int64, err error) {
	// 安全性检验, 防止空model
	if repayPlan.Amount <= 0 {
		err = fmt.Errorf("Invalid RepayPlan data, repayPlan.Amount must > 0 , repayPlan: %v", repayPlan)
		return
	}
	preReduced, _ := dao.GetLastPrereducedByOrderid(repayPlan.OrderId)
	amount = CaculateRepayTotalAmountWithPreReduced(repayPlan.Amount, repayPlan.AmountPayed, repayPlan.AmountReduced,
		repayPlan.GracePeriodInterest, repayPlan.GracePeriodInterestPayed, repayPlan.GracePeriodInterestReduced, preReduced.GracePeriodInterestPrededuced,
		repayPlan.Penalty, repayPlan.PenaltyPayed, repayPlan.PenaltyReduced, preReduced.PenaltyPrereduced)
	return
}

// CaculateTotalGracePeriodAndPenaltyByRepayPlan 根据 repayPlan model 计算应还宽限期利息和罚息
func CaculateTotalGracePeriodAndPenaltyByRepayPlan(repayPlan models.RepayPlan) (amount int64, err error) {
	// 安全性检验, 防止空model
	if repayPlan.Amount <= 0 {
		err = fmt.Errorf("Invalid RepayPlan data, repayPlan.Amount must > 0 , repayPlan: %v", repayPlan)
		return
	}
	amount = CaculateTotalGracePeriodAndPenalty(repayPlan.GracePeriodInterest, repayPlan.GracePeriodInterestPayed, repayPlan.GracePeriodInterestReduced,
		repayPlan.Penalty, repayPlan.PenaltyPayed, repayPlan.PenaltyReduced)
	return
}

// CaculateTotalPayedByRepayPlan 根据 repayPlan model 计算已还金额(已还本金,已还宽限期利息,已还罚息)
func CaculateTotalPayedByRepayPlan(repayPlan models.RepayPlan) (amount int64, err error) {
	// 安全性检验, 防止空model
	if repayPlan.Amount <= 0 {
		err = fmt.Errorf("Invalid RepayPlan data, repayPlan.Amount must > 0 , repayPlan: %v", repayPlan)
		return
	}
	amount = CaculateTotalPayed(repayPlan.AmountPayed, repayPlan.GracePeriodInterestPayed, repayPlan.PenaltyPayed)
	return
}

// CaculateTotalReducedByRepayPlan 根据 repayPlan model 计算已减免金额(已减免本金,已减免宽限期利息,已减免罚息)
func CaculateTotalReducedByRepayPlan(repayPlan models.RepayPlan) (amount int64, err error) {
	// 安全性检验, 防止空model
	if repayPlan.Amount <= 0 {
		err = fmt.Errorf("Invalid RepayPlan data, repayPlan.Amount must > 0 , repayPlan: %v", repayPlan)
		return
	}
	amount = CaculateTotalPayed(repayPlan.AmountReduced, repayPlan.GracePeriodInterestReduced, repayPlan.PenaltyReduced)
	return
}

// CaculateTotalAmountByRepayPlan 根据 repayPlan model 计算账单总额(应还本金,应还宽限期利息,应还罚息)
func CaculateTotalAmountByRepayPlan(repayPlan models.RepayPlan) (amount int64, err error) {
	// 安全性检验, 防止空model
	if repayPlan.Amount <= 0 {
		err = fmt.Errorf("Invalid RepayPlan data, repayPlan.Amount must > 0 , repayPlan: %v", repayPlan)
		return
	}
	amount = CaculateOrderTotalAmount(repayPlan.Amount, repayPlan.GracePeriodInterest, repayPlan.Penalty)
	return
}

// CaculateRepayTotalAmount 计算应还总额
// 所有的计算应还总额, 都应该调用此底层方法, 确保计算公式统一性
// 计算公式: 应该总额 = (应该本金-已还本金-已减免本金) + (应还宽限期利息-已还宽限期利息-已减免宽限期利息) + (应还罚息-已还罚息- 已减免罚息)
func CaculateRepayTotalAmount(repayPlanAmount, repayPlanAmountPayed, repayPlanAmountReduced,
	gracePeriodInterest, gracePeriodInterestPayed, gracePereiodInterestReduced,
	penalty, penaltyPayed, penaltyReduced int64) int64 {
	return (repayPlanAmount - repayPlanAmountPayed - repayPlanAmountReduced) +
		(gracePeriodInterest - gracePeriodInterestPayed - gracePereiodInterestReduced) +
		(penalty - penaltyPayed - penaltyReduced)
}

// CaculateRepayTotalAmountWithPreReduced 计算结清减免应还总额（有条件减免）
// 计算公式: 结清减免应还总额=(应还本金-已还本金-已减免本金)+(应还款期限利息-已还宽限期利息-已减免宽限期利息-预减免宽限期利息)+(应还罚息-已还罚息-已减免罚息-预减免罚息)
func CaculateRepayTotalAmountWithPreReduced(repayPlanAmount, repayPlanAmountPayed, repayPlanAmountReduced, gracePeriodInterest,
	gracePeriodInterestPayed, gracePereiodInterestReduced, gracePereiodInterestPreReduced,
	penalty, penaltyPayed, penaltyReduced, penaltyPreReduced int64) int64 {
	logs.Debug("[CaculateRepayTotalAmountWithPreReduced]结清减免应还总额=(应还本金-已还本金-已减免本金)+(应还款期限利息-已还宽限期利息-已减免宽限期利息-预减免宽限期利息)+(应还罚息-已还罚息-已减免罚息-预减免罚息)=(%d-%d-%d)+(%d-%d-%d-%d)+(%d-%d-%d-%d)",
		repayPlanAmount, repayPlanAmountPayed, repayPlanAmountReduced,
		gracePeriodInterest, gracePeriodInterestPayed, gracePereiodInterestReduced, gracePereiodInterestPreReduced,
		penalty, penaltyPayed, penaltyReduced, penaltyPreReduced)
	return (repayPlanAmount - repayPlanAmountPayed - repayPlanAmountReduced) +
		(gracePeriodInterest - gracePeriodInterestPayed - gracePereiodInterestReduced - gracePereiodInterestPreReduced) +
		(penalty - penaltyPayed - penaltyReduced - penaltyPreReduced)
}

//CaculateTotalGracePeriodAndPenalty 应还罚息和宽限期利息
// 计算公式： （ 应还罚息-已还罚息-已减免罚息 ）+（应还宽限期利息-已还宽限期利息-已减免宽限期利息）
func CaculateTotalGracePeriodAndPenalty(gracePeriodInterest, gracePeriodInterestPayed, gracePereiodInterestReduced,
	penalty, penaltyPayed, penaltyReduced int64) int64 {
	return (gracePeriodInterest - gracePeriodInterestPayed - gracePereiodInterestReduced) +
		(penalty - penaltyPayed - penaltyReduced)
}

//CaculateTotalPayed 计算已还金额(已还本金,已还宽限期利息,已还罚息)
// 计算公式： 已还本金 + 已还罚息 + 已还宽限期利息
func CaculateTotalPayed(amountPayed, gracePeriodInterestPayed, penaltyPayed int64) int64 {
	return (amountPayed + gracePeriodInterestPayed + penaltyPayed)
}

//CaculateTotalAmount 计算账单总额(应还本金,应还宽限期利息,应还罚息)
// 计算公式： 应还本金 + 应还罚息 + 应还宽限期利息
func CaculateOrderTotalAmount(amount, gracePeriodInterest, penalty int64) int64 {
	return (amount + gracePeriodInterest + penalty)
}

//CaculateTotalAmount 应还本金
// 计算公式： （ 应还本金-已还本金-已减免本金   ）
func CaculateTotalAmount(Amount, AmountPayed, AmountReduced int64) int64 {
	return (Amount - AmountPayed - AmountReduced)
}

//CaculateTotalGracePeriod 应还宽限期利息
// 计算公式： （ 应还罚息-已还罚息-已减免罚息 ）
func CaculateTotalGracePeriod(gracePeriodInterest, gracePeriodInterestPayed, gracePereiodInterestReduced int64) int64 {
	return (gracePeriodInterest - gracePeriodInterestPayed - gracePereiodInterestReduced)
}

//CaculateTotalPenalty 应还罚息
// 计算公式： （ 应还罚息-已还罚息-已减免罚息）
func CaculateTotalPenalty(penalty, penaltyPayed, penaltyReduced int64) int64 {
	return (penalty - penaltyPayed - penaltyReduced)
}

// CaculateCanReducedAmount 计算可减免金额
//计算公式 ：应还罚息和宽限期利息 * N N根据案件的等级定
func CaculateCanReducedAmount(gracePeriodAndPenaltyAmount int64, N float64) int64 {
	return int64(math.Floor(float64(gracePeriodAndPenaltyAmount) * N))
}

// CaculateGracPeriodAndPenaltyAmount 可减免金额与N逆算 应还罚息和宽限期利息
//计算公式 ：可减免金额 / N + 0.1 N根据案件的等级定,加0.1是因为在计算可减免金额是已经舍弃精度，逆算回去会更少，所以加浮点向上取整
func CaculateGracPeriodAndPenaltyAmount(canReduced int64, N float64) (val int64) {
	if N == 1 {
		val = int64(float64(canReduced) / N)
	} else {
		val = int64(math.Ceil(float64(canReduced)/N + 0.1))
	}
	return
}
