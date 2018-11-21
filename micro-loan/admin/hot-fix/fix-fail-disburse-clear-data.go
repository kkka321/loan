package main

import (
	// 数据库初始化

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/models"

	"micro-loan/common/tools"

	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

func clearData(orderId int64) {
	order, err := models.GetOrder(orderId)
	if err != nil {
		logs.Error("fix-fail-disburse-clear-data orderId is not valid", orderId)
		return
	}
	repayPlan, _ := models.GetLastRepayPlanByOrderid(orderId)

	o := orm.NewOrm()
	o.Using(repayPlan.Using())
	if repayPlan.Id > 0 {
		o.Delete(&repayPlan)
		//删除还款计划
	}

	payment, _ := models.GetPaymentByOrderIdPayType(orderId, 2)
	if payment.Id > 0 {
		o.Delete(&payment)
		//删除payment放款记录
	}

	userEtrans1, _ := models.GetEtranByOrderIdPayTypeVaCompanyCode(orderId, 1, 1001)
	userEtrans2, _ := models.GetEtranByOrderIdPayTypeVaCompanyCode(orderId, 2, 1001)

	if userEtrans1.Id > 0 {
		o.Delete(&userEtrans1)
	}

	if userEtrans2.Id > 0 {
		o.Delete(&userEtrans2)
	}
	//删除砍头息进账出账

	overdueList, _ := models.GetRepayPlanOverdueByOrderId(orderId)
	for _, v := range overdueList {
		if v.Id > 0 {
			o.Delete(&v)
		}
		//删除所有的逾期罚息记录
	}

	desc := tools.Int642Str(orderId)
	mobiEtrans, _ := models.GetMobiEtransByAccountIdDescription(order.UserAccountId, desc)
	if mobiEtrans.Id > 0 {
		o.Delete(&mobiEtrans)
		//删除mobi放款记录
	}

	order.CheckStatus = types.LoanStatusLoanFail
	order.UpdateOrder(&order)
	logs.Debug("orderId has been cleared data.", orderId)

}

func main() {
	//var fixOrders = []int64{180904020007644528, 180905020004429110}

	// 2018年09月10日15:31:35  修复订单
	//var fixOrders = []int64{180906020000985024}

	// 2018年09月11日16:01:15  刘东强 修复订单
	var fixOrders = []int64{180906020000615125,
		180905020010172831,
		180905020013263496,
		180905020010247184}

	//var fixOrders = []int64{180807021463513794}
	for _, v := range fixOrders {
		clearData(v)
	}

}
