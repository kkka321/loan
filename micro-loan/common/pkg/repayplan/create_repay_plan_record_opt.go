package repayplan

import (
	"micro-loan/common/lib/device"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	libProduct "micro-loan/common/lib/product"

	"github.com/astaxie/beego/logs"
)

func CreateRepayPlan(total, interest, serviceFee int64, dataOrder *models.Order, dataProduct *models.Product) models.RepayPlan {
	//添加payment记录，方便后续财务统计
	repayPlan := models.RepayPlan{}
	repayPlan.Id, _ = device.GenerateBizId(types.RepayPlanBiz)
	repayPlan.OrderId = dataOrder.Id
	repayPlan.Amount = total
	repayPlan.PreInterest = interest
	repayPlan.ServiceFee = serviceFee
	repayDate := tools.NaturalDay(int64(dataOrder.Period) * int64(dataProduct.Period))
	logs.Info("caculate repayDate is :", repayDate, "int64(dataOrder.Period) is :", int64(dataOrder.Period), " int64(dataProduct.Period) is:", int64(dataProduct.Period))
	repayPlan.RepayDate = repayDate
	//还款日期 如果还款期限是7 ，而3.11为放款日期， 那么正常还款日期是3.18，宽限期是3.19，大于3.19的日期就是逾期
	repayPlan.Ctime = tools.GetUnixMillis()
	repayPlan.Utime = tools.GetUnixMillis()
	//初始化还款计划
	if dataProduct.ChargeFeeType == types.ProductChargeFeeInterestBefore || dataProduct.ChargeInterestType == types.ProductChargeInterestTypeHeadCut {

		if dataProduct.ChargeInterestType == types.ProductChargeInterestTypeHeadCut {
			repayPlan.PreInterestPayed = interest
		}
		if dataProduct.ChargeFeeType == types.ProductChargeFeeInterestBefore {
			repayPlan.ServiceFeePayed = serviceFee
		}
	}

	models.AddRepayPlan(&repayPlan)
	//添加还款计划

	return repayPlan
}

// calcRepayInfoV2 使用配置的产品信息计算本金 利息 服务费
func CalcRepayInfoV2(loan int64, product models.Product, period int) (int64, int64, int64) {
	var total, interest, serviceFee int64

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

	interest, serviceFee, _, total, _ = libProduct.GetInterestAndFee(trialIn, product)
	logs.Debug("[calcRepayInfoV2]  interest %d  serviceFee %d  total %d", interest, serviceFee, total)
	return total, interest, serviceFee
}

// calcRepayInfoV3 使用配置的产品信息计算本金 利息 服务费
func CalcRepayInfoV3(amount int64, product models.Product, period int) (int64, int64, int64) {
	var total, interest, serviceFee int64

	trialIn := types.ProductTrialCalcIn{
		ID:           product.Id,
		Loan:         0,
		Amount:       amount,
		Period:       period,
		LoanDate:     "",
		CurrentDate:  "",
		RepayDate:    "",
		RepayedTotal: 0,
	}
	interest, serviceFee, _, total, _ = libProduct.GetInterestAndFee(trialIn, product)
	logs.Debug("[calcRepayInfoV3]  interest %d  serviceFee %d  total %d", interest, serviceFee, total)
	return total, interest, serviceFee
}
