package product

import (
	"math"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"github.com/astaxie/beego/logs"
)

// TrialCalcRepayTypeOnce 一次性还款付息，这种方式最多为2行数据，如果当前时间大于 最终应还款日 则在最后展示一条当前日期列
func TrialCalcRepayTypeOnce(trialIn types.ProductTrialCalcIn, product models.Product) (trialResults []types.ProductTrialCalcResult, err error) {
	logs.Debug("[TrialCalcRepayTypeOnce] trialIn %#v product:%#v", trialIn, product)
	loanDate := tools.GetDateParse(trialIn.LoanDate) * 1000
	currentDate := tools.GetDateParse(trialIn.CurrentDate) * 1000
	repayDate := tools.GetDateParse(trialIn.RepayDate) * 1000
	repayDateShould := loanDate + int64(trialIn.Period)*int64(product.Period)*24*3600*1000
	graceInterestDate := repayDateShould + int64(product.GracePeriod)*int64(product.Period)*24*3600*1000
	repayOrder := strings.Split(product.RepayOrder, ";")
	interestTotal, feeTotal, loanTotal, amountTotal, _ := GetInterestAndFee(trialIn, product)

	logs.Debug("[TrialCalcRepayTypeOnce]  loanDate:%d currentDate:%d repayDate:%d repayDateShould:%d graceInterestDate:%d repayOrder:%s",
		loanDate, currentDate, repayDate, repayDateShould, graceInterestDate, repayOrder)
	logs.Debug("[TrialCalcRepayTypeOnce] interestTotal:%d, feeTotal:%d, loanTotal:%d, amountTotal::%d", interestTotal, feeTotal, loanTotal, amountTotal)
	status := types.ProductTrialCalcStatus{
		LoanDate:          loanDate,
		CurrentDate:       currentDate,
		RepayDate:         repayDate,
		RepayDateShould:   repayDateShould,
		GraceInterestDate: graceInterestDate,
		InterestTotal:     interestTotal,
		FeeTotal:          feeTotal,
		LoanTotal:         loanTotal,
		AmountTotal:       amountTotal,
		RepayOrder:        repayOrder,
	}
	//for 循环里逐条计算各个字段的值
	for i := 0; i < 2; i++ {
		result := calcTrialResultByNum(trialIn, product, i, status)
		trialResults = append(trialResults, result)
	}
	return
}

// GetInterestAndFee 计算利息和服务费
func GetInterestAndFee(trialIn types.ProductTrialCalcIn, product models.Product) (interest int64, fee int64, loan int64, amount int64, err error) {
	logs.Debug("[GetInterestAndFee]  trialIn %#v product:%#v", trialIn, product)

	amount = 0
	if 0 == trialIn.Loan {
		amount = trialIn.Amount
	}

	// 利息 = 周期 × 单位 × 日费率
	interest = int64(trialIn.Period) * int64(product.Period) * product.DayInterestRate
	// 利息 = 周期 × 单位 × 日费率
	fee = int64(trialIn.Period) * int64(product.Period) * product.DayFeeRate

	// 根据费率收取方式计算本金的值
	if 0 == amount {
		//用户填写了放款金额
		loan = trialIn.Loan
		if product.ChargeInterestType == types.ProductChargeInterestTypeHeadCut &&
			product.ChargeFeeType == types.ProductChargeFeeInterestBefore {

			// 费率和利息都提前砍头收
			amountTmp := float64(trialIn.Loan) / (float64(1) - float64(interest+fee)/float64(types.ProductFeeBase))
			amount = tools.CeilWay(float64(amountTmp), 1, product.CeilWay, product.CeilWayUnit)

			interest = int64(float64(interest*amount) / float64(types.ProductFeeBase)) // tools.CeilWay(float64(interest*amount), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)
			// fee = int64(float64(fee*amount) / float64(types.ProductFeeBase))           //tools.CeilWay(float64(fee*amount), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)

			// 保证 平衡 防止 利息+服务费+放款额不等于本金
			fee = amount - loan - interest

		} else if product.ChargeInterestType == types.ProductChargeInterestTypeHeadCut &&
			product.ChargeFeeType == types.ProductChargeFeeInterestAfter {

			// 费率置后  利息提前砍头收
			amountTmp := float64(trialIn.Loan) / (float64(1) - float64(interest)/float64(types.ProductFeeBase))
			amount = tools.CeilWay(float64(amountTmp), 1, product.CeilWay, product.CeilWayUnit)

			// interest = int64(float64(interest*amount) / float64(types.ProductFeeBase)) // tools.CeilWay(float64(interest*amount), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)
			fee = int64(float64(fee*amount) / float64(types.ProductFeeBase)) //tools.CeilWay(float64(fee*amount), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)

			// 保证 平衡 防止 利息+放款额不等于本金
			interest = amount - loan
		} else if product.ChargeInterestType == types.ProductChargeInterestTypeByStages &&
			product.ChargeFeeType == types.ProductChargeFeeInterestBefore {

			//  利息置后  费率提前砍头收
			amountTmp := float64(trialIn.Loan) / (float64(1) - float64(fee)/float64(types.ProductFeeBase))
			amount = tools.CeilWay(float64(amountTmp), 1, product.CeilWay, product.CeilWayUnit)

			interest = int64(float64(interest*amount) / float64(types.ProductFeeBase)) // tools.CeilWay(float64(interest*amount), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)
			// fee = int64(float64(fee*amount) / float64(types.ProductFeeBase))           //tools.CeilWay(float64(fee*amount), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)
			// 保证 平衡 防止 服务费+放款额不等于本金
			fee = amount - loan
		} else if product.ChargeInterestType == types.ProductChargeInterestTypeByStages &&
			product.ChargeFeeType == types.ProductChargeFeeInterestAfter {
			//  利息和费率都置后收
			amount = trialIn.Loan
			interest = int64(float64(interest*amount) / float64(types.ProductFeeBase)) // tools.CeilWay(float64(interest*amount), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)
			fee = int64(float64(fee*amount) / float64(types.ProductFeeBase))           //tools.CeilWay(float64(fee*amount), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)
		}
		return
	} else {
		interest = int64(float64(interest*amount) / float64(types.ProductFeeBase)) // tools.CeilWay(float64(interest*amount), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)
		fee = int64(float64(fee*amount) / float64(types.ProductFeeBase))           //tools.CeilWay(float64(fee*amount), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)

		if product.ChargeInterestType == types.ProductChargeInterestTypeHeadCut &&
			product.ChargeFeeType == types.ProductChargeFeeInterestBefore {

			// 费率和利息都提前砍头收
			loan = amount - interest - fee
		} else if product.ChargeInterestType == types.ProductChargeInterestTypeHeadCut &&
			product.ChargeFeeType == types.ProductChargeFeeInterestAfter {

			// 费率置后  利息提前砍头收
			loan = amount - interest
		} else if product.ChargeInterestType == types.ProductChargeInterestTypeByStages &&
			product.ChargeFeeType == types.ProductChargeFeeInterestBefore {

			//  利息置后  费率提前砍头收
			loan = amount - fee
		} else if product.ChargeInterestType == types.ProductChargeInterestTypeByStages &&
			product.ChargeFeeType == types.ProductChargeFeeInterestAfter {
			//  利息和费率都置后收
			loan = amount
		}
		return
	}
}

func calcTrialResultByNum(trialIn types.ProductTrialCalcIn, product models.Product, index int, status types.ProductTrialCalcStatus) (result types.ProductTrialCalcResult) {
	result.ID = product.Id
	result.Name = product.Name
	result.NumberOfPeriods = index
	result.RepayDateShould = tools.ThreeElementExpression(index == 0, status.LoanDate, status.RepayDateShould).(int64)
	result.Loan = tools.ThreeElementExpression(index == 0, status.LoanTotal, int64(0)).(int64)

	result.RepayAmountShould = tools.ThreeElementExpression(index == 0, int64(0), status.AmountTotal).(int64)
	result.RepayInterestShould = tools.ThreeElementExpression((index == 0 && types.ProductChargeInterestTypeHeadCut == product.ChargeInterestType) || (index == 1 && types.ProductChargeInterestTypeByStages == product.ChargeInterestType),
		status.InterestTotal, int64(0)).(int64)
	result.RepayFeeShould = tools.ThreeElementExpression((index == 0 && types.ProductChargeFeeInterestBefore == product.ChargeFeeType) || (index == 1 && types.ProductChargeFeeInterestAfter == product.ChargeFeeType),
		status.FeeTotal, int64(0)).(int64)
	result.RepayedDate = getResultRepayedDateByNum(trialIn, product, index, status)

	//先计算还款金额才能 计算应还罚息 和 宽限期利息
	rv := doRepay(trialIn, product, index, status, &result)
	if !rv {
		result.RepayGraceInterestShould = getResultGraceInterestShouldByNum(trialIn, product, index, status)
		result.RepayPenaltyShould = getResultPenaltyShouldByNum(trialIn, product, index, status)
		result.RepayedAmount = 0
		result.RepayedInterest = tools.ThreeElementExpression(index == 0, result.RepayInterestShould, int64(0)).(int64)
		result.RepayedFee = tools.ThreeElementExpression(index == 0, result.RepayFeeShould, int64(0)).(int64)
		result.RepayedGraceInterest = 0
		result.RepayedPenalty = 0
	}

	result.ForfeitPenalty = 0
	result.RepayedForfeitPenalty = 0
	result.RepayTotalShould = result.RepayAmountShould + result.RepayInterestShould + result.RepayFeeShould +
		result.RepayGraceInterestShould + result.RepayPenaltyShould + result.ForfeitPenalty

	result.RepayedTotal = result.RepayedAmount + result.RepayedInterest + result.RepayedGraceInterest +
		result.RepayedFee + result.RepayedPenalty + result.RepayedForfeitPenalty

	result.OverdueDays = getResultOverdueDaysByNum(trialIn, product, index, status, &result)
	result.RepayStatus = getResultRepayStatusByNum(trialIn, product, index, status, &result)
	return
}

func getResultRepayedDateByNum(trialIn types.ProductTrialCalcIn, product models.Product, index int, status types.ProductTrialCalcStatus) (repayedDate int64) {
	if 0 == index {
		if product.ChargeInterestType == types.ProductChargeInterestTypeByStages &&
			product.ChargeFeeType == types.ProductChargeFeeInterestAfter {
			repayedDate = 0
		} else {
			repayedDate = status.LoanDate
		}
	} else {
		if trialIn.RepayedTotal > 0 {
			repayedDate = status.RepayDate
		} else {
			repayedDate = 0
		}
	}
	return
}

func getResultGraceInterestShouldByNum(trialIn types.ProductTrialCalcIn, product models.Product, index int, status types.ProductTrialCalcStatus) (graceInterest int64) {
	if 0 == index {
		graceInterest = 0
	} else {
		if status.RepayDateShould >= status.CurrentDate {
			graceInterest = 0
		} else {
			graceInterestDays := (status.CurrentDate - status.RepayDateShould) / (24 * 3600 * 1000)
			if status.GraceInterestDate <= status.CurrentDate {
				graceInterestDays = int64(product.GracePeriod)
			}
			graceInterest = int64(graceInterestDays) * int64(float64(status.AmountTotal*product.DayGraceRate)/float64(types.ProductFeeBase))
			// graceInterest = tools.CeilWay(float64(graceInterest), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)
			// int64(math.Ceil(float64(graceInterest)/float64(types.ProductFeeBase)/float64(product.CeilWay))) * int64(product.CeilWay)
		}
	}
	return
}

func getResultPenaltyShouldByNum(trialIn types.ProductTrialCalcIn, product models.Product, index int, status types.ProductTrialCalcStatus) (penalty int64) {
	if 0 == index {
		penalty = 0
	} else {
		// 当前日期 小于宽限期 无罚息
		if status.GraceInterestDate >= status.CurrentDate {
			penalty = 0
		} else {
			penaltyDays := (status.CurrentDate - status.GraceInterestDate) / (24 * 3600 * 1000)
			// 还款日总的逾期天数，包含宽限期和逾期天数
			penaltyDaysTotal := (status.CurrentDate - status.RepayDateShould) / (24 * 3600 * 1000)
			if penaltyDaysTotal > 90 {
				penaltyDays = int64(90 - product.GracePeriod)
			}
			penalty = int64(penaltyDays) * int64(float64(status.AmountTotal*product.DayPenaltyRate)/float64(types.ProductFeeBase))
			// penalty = tools.CeilWay(float64(penalty), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)
			// int64(math.Ceil(float64(penalty)/float64(types.ProductFeeBase)/float64(product.CeilWay))) * int64(product.CeilWay)
		}
	}
	return
}

// getResultOverdueDaysByNum(trialIn, product, index, status)
func getResultOverdueDaysByNum(trialIn types.ProductTrialCalcIn, product models.Product, index int, status types.ProductTrialCalcStatus, result *types.ProductTrialCalcResult) (overdueDays int) {
	if 0 == index {
		overdueDays = 0
	} else {
		//当前日期在应还日里 无逾期
		if status.RepayDateShould >= status.CurrentDate {
			overdueDays = 0
		} else {
			//已还清 无逾期
			if result.RepayTotalShould <= result.RepayedTotal {
				overdueDays = 0
			} else {
				// if status.CurrentDate <= status.GraceInterestDate {
				// 	overdueDays = 0
				// } else {
				overdueDays = int((status.CurrentDate - status.RepayDateShould) / (24 * 3600 * 1000))
				// }
			}
		}
	}
	return
}

// getResultRepayStatusByNum(trialIn, product, index, status)
func getResultRepayStatusByNum(trialIn types.ProductTrialCalcIn, product models.Product, index int, status types.ProductTrialCalcStatus, result *types.ProductTrialCalcResult) (repayStatus int) {
	if 0 == index || result.RepayTotalShould == 0 {
		repayStatus = 0
	} else {
		// 已结清的直接返回
		if isClear(result) {
			repayStatus = int(types.LoanStatusAlreadyCleared)
			return
		}

		//当前日期在应还日里 状态可选为 等待还款、部分还款、结清
		if status.RepayDateShould >= status.CurrentDate {
			// 未发生还款 状态 等待还款
			if trialIn.RepayedTotal == 0 {
				repayStatus = int(types.LoanStatusWaitRepayment)
			} else {
				repayStatus = int(types.LoanStatusPartialRepayment)
			}
		} else {
			//超过还款日  状态可为逾期 --结清
			repayStatus = int(types.LoanStatusOverdue)
		}
	}
	return
}

// isDone 判断当期是否结清
func isClear(result *types.ProductTrialCalcResult) (flag bool) {
	return (result.RepayTotalShould <= result.RepayedTotal)
}

func doRepay(trialIn types.ProductTrialCalcIn, product models.Product, index int, status types.ProductTrialCalcStatus, result *types.ProductTrialCalcResult) (rv bool) {
	if 0 == index || 0 == trialIn.RepayedTotal {
		return false
	}
	logs.Debug("[doRepay] index:%d product %v  trialIn:%v", index, product, trialIn)

	//1、统一计算出还款日当天的罚息 宽限息 和滞纳金 防止不同的还款顺序导致未赋值
	// 宽限息    还款日在应还日里 不需要宽限期利息
	if status.RepayDate > status.RepayDateShould {
		// 还款日距应还日的时间
		graceInterestDays := (status.RepayDate - status.RepayDateShould) / (24 * 3600 * 1000)
		//如果还款日 晚于宽限期日  直接用宽限期
		if status.RepayDate > status.GraceInterestDate {
			graceInterestDays = int64(product.GracePeriod)
		}
		// 还款日当前的宽限息
		currentGraceShould := int64(graceInterestDays) * int64(float64((result.RepayAmountShould-result.RepayedAmount)*product.DayGraceRate)/float64(types.ProductFeeBase))
		result.RepayGraceInterestShould = currentGraceShould
		// currentGraceShould = tools.CeilWay(float64(currentGraceShould), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)
		logs.Debug("[doRepay] graceInterestDays %d  currentGraceShould %d  result.RepayAmountShould %d result.RepayedAmount %d",
			graceInterestDays, currentGraceShould, result.RepayAmountShould, result.RepayedAmount)
	}

	// 罚息 (1)还款日在宽限期里 不需要罚息; (2)大于90天的不再计算,罚息值
	if status.RepayDate > status.GraceInterestDate {
		// 还款日距应宽限期的时间
		penaltyDays := (status.RepayDate - status.GraceInterestDate) / (24 * 3600 * 1000)

		// 还款日总的逾期天数，包含宽限期和逾期天数
		penaltyDaysTotal := (status.RepayDate - status.RepayDateShould) / (24 * 3600 * 1000)
		if penaltyDaysTotal > 90 {
			penaltyDays = int64(90 - product.GracePeriod)
		}
		// 还款日当前的罚息
		currentPenalty := int64(penaltyDays) * int64(float64((result.RepayAmountShould-result.RepayedAmount)*product.DayPenaltyRate)/float64(types.ProductFeeBase))
		// currentPenalty = tools.CeilWay(float64(currentPenalty), types.ProductFeeBase, product.CeilWay, product.CeilWayUnit)
		logs.Debug("[doRepay] penaltyDays %d  currentPenalty %d  result.RepayAmountShould %d result.RepayedAmount %d",
			penaltyDays, currentPenalty, result.RepayAmountShould, result.RepayedAmount)
		result.RepayPenaltyShould = currentPenalty
	}

	// 2、 完成部分还款
	repayedTotal := trialIn.RepayedTotal
	for repayedTotal > 0 {
		if len(status.RepayOrder) == 0 {
			logs.Debug("[doRepay] len(status.RepayOrder) == 0")
			break
		}
		//根据还款顺序还款
		v := status.RepayOrder[0]
		logs.Debug("[doRepay] before switch repayedTotal %d v:%s", repayedTotal, v)
		switch v {
		case types.ProductOrderAmount:
			{
				//应还本金小于已还款 结清本金
				if result.RepayAmountShould < repayedTotal {
					result.RepayedAmount = result.RepayAmountShould
					repayedTotal -= result.RepayAmountShould
					status.RepayOrder = status.RepayOrder[1:]
				} else {
					result.RepayedAmount = repayedTotal
					repayedTotal = 0
				}
			}
		case types.ProductOrderInterest:
			{
				//应还利息小于已还款 结清利息
				if result.RepayInterestShould < repayedTotal {
					result.RepayedInterest = result.RepayInterestShould
					repayedTotal -= result.RepayInterestShould
					status.RepayOrder = status.RepayOrder[1:]
				} else {
					result.RepayedInterest = repayedTotal
					repayedTotal = 0
				}
			}
		case types.ProductOrderFee:
			{
				//应还费率小于已还款 结清费率
				if result.RepayFeeShould < repayedTotal {
					result.RepayedFee = result.RepayFeeShould
					repayedTotal -= result.RepayFeeShould
					status.RepayOrder = status.RepayOrder[1:]
				} else {
					result.RepayedFee = repayedTotal
					repayedTotal = 0
				}
			}
		case types.ProductOrderGraceInterest:
			{
				//已还金额 大于当前应还宽限期利息  结清宽限息
				if repayedTotal > result.RepayGraceInterestShould {
					result.RepayedGraceInterest = result.RepayGraceInterestShould
					status.RepayOrder = status.RepayOrder[1:]
					repayedTotal -= result.RepayGraceInterestShould
				} else {
					result.RepayedGraceInterest = repayedTotal
					repayedTotal = 0
				}
			}
		case types.ProductOrderPenalty:
			{
				//已还金额 大于当前应还罚息  结清罚息
				if repayedTotal > result.RepayPenaltyShould {
					result.RepayedPenalty = result.RepayPenaltyShould
					status.RepayOrder = status.RepayOrder[1:]
					repayedTotal -= result.RepayPenaltyShould
				} else {
					result.RepayedPenalty = repayedTotal
					repayedTotal = 0
				}

			}
		case types.ProductOrderForfeitPenalty:
			{
				status.RepayOrder = status.RepayOrder[1:]
			}
		default:
			{
				status.RepayOrder = status.RepayOrder[1:]
			}
		}

		logs.Debug("[doRepay] after switch repayedTotal %d", repayedTotal)
	}

	// 3、还款日过后 可能未结清账单 会生成额外的 宽限息或罚息
	// 当当前日期大于还款日期且账单未结清时 修正 应还宽限期和罚息的值
	if status.CurrentDate > status.RepayDate && status.CurrentDate > status.RepayDateShould {
		// 只有当本金未还玩的时候才会继续生成利息
		if result.RepayedAmount != result.RepayAmountShould {
			if status.CurrentDate < status.GraceInterestDate {
				// 当前日期小于宽限期 只有多余的宽限期无罚息
				graceInterestDays := (status.CurrentDate - int64(math.Max(float64(status.RepayDate), float64(status.RepayDateShould)))) / (24 * 3600 * 1000)
				modifyGraceShould := int64(graceInterestDays) * int64(float64((result.RepayAmountShould-result.RepayedAmount)*product.DayGraceRate)/float64(types.ProductFeeBase))
				result.RepayGraceInterestShould += modifyGraceShould
				logs.Debug("[doRepay] graceInterestDays %d  modifyGraceShould %d  result.RepayAmountShould %d result.RepayedAmount %d result.RepayGraceInterestShould %d",
					graceInterestDays, modifyGraceShould, result.RepayAmountShould, result.RepayedAmount, result.RepayGraceInterestShould)
			} else {
				// 当前日期大于宽限期 还款日小于宽限期有多余的宽限期
				if status.RepayDate < status.GraceInterestDate {
					graceInterestDays := (status.GraceInterestDate - int64(math.Max(float64(status.RepayDate), float64(status.RepayDateShould)))) / (24 * 3600 * 1000)
					modifyGraceShould := int64(graceInterestDays) * int64(float64((result.RepayAmountShould-result.RepayedAmount)*product.DayGraceRate)/float64(types.ProductFeeBase))
					result.RepayGraceInterestShould += modifyGraceShould

					logs.Debug("[doRepay] graceInterestDays %d  modifyGraceShould %d  result.RepayAmountShould %d result.RepayedAmount %d result.RepayGraceInterestShould %d",
						graceInterestDays, modifyGraceShould, result.RepayAmountShould, result.RepayedAmount, result.RepayGraceInterestShould)
				}

				// 当前总的逾期天数，包含宽限期和逾期天数
				penaltyDays := int64(0)
				penaltyDaysTotal := (status.CurrentDate - status.RepayDateShould) / (24 * 3600 * 1000)
				if penaltyDaysTotal <= 90 {
					penaltyDays = (status.CurrentDate - int64(math.Max(float64(status.RepayDate), float64(status.GraceInterestDate)))) / (24 * 3600 * 1000) // (status.CurrentDate - status.GraceInterestDate) / (24 * 3600 * 1000)
				} else {
					penaltyDays = (status.RepayDateShould-int64(math.Max(float64(status.RepayDate), float64(status.GraceInterestDate))))/(24*3600*1000) + 90
					// 还款期大于 逾期90天时没有多余的罚息
					penaltyDays = int64(math.Max(0, float64(penaltyDays)))
				}
				modifyPenalty := int64(penaltyDays) * int64(float64((result.RepayAmountShould-result.RepayedAmount)*product.DayPenaltyRate)/float64(types.ProductFeeBase))
				result.RepayPenaltyShould += modifyPenalty

				logs.Debug("[doRepay] penaltyDays %d  modifyPenalty %d  result.RepayAmountShould %d result.RepayedAmount %d result.RepayGraceInterestShould %d",
					penaltyDays, modifyPenalty, result.RepayAmountShould, result.RepayedAmount, result.RepayPenaltyShould)
			}
		}
	}
	return true
}
