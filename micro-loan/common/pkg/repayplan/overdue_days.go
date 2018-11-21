package repayplan

import "micro-loan/common/tools"

func GetRepayDateByOverdueDays(overdueDays int64) int64 {
	return tools.NaturalDay(-overdueDays)

}
