package reduce

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/pkg/repayplan"
)

func RepayLowestMoney4ClearReduce(Order models.Order, repayPlan models.RepayPlan) (amount int64, err error) {
	amount, _ = repayplan.CaculateRepayTotalAmountWithPreReducedByRepayPlan(repayPlan)
	if amount <= 0 {
		err = fmt.Errorf("RepayLowestMoney amount is not valid")
	}
	return
}
