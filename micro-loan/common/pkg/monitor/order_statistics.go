package monitor

import (
	"math"

	"micro-loan/common/types"
	"micro-loan/common/models"
	"micro-loan/common/tools"
)

func GetOrderStatistics(date int64, statistics *models.OrderStatistics) {
	statistics.Id = math.MaxInt64

	key := getMonitorKey(orderHashPrefix, date)

	statistics.Submit = getCountFromCache(key, types.LoanStatusSubmit)
	statistics.WaitReview = getCountFromCache(key, types.LoanStatus4Review)
	statistics.Reject = getCountFromCache(key, types.LoanStatusReject)
	statistics.WaitManual = getCountFromCache(key, types.LoanStatusWaitManual)
	statistics.WaitLoan = getCountFromCache(key, types.LoanStatusWait4Loan)
	statistics.LoanFail = getCountFromCache(key, types.LoanStatusLoanFail)
	statistics.WaitRepayment = getCountFromCache(key, types.LoanStatusWaitRepayment)
	statistics.Cleared = getCountFromCache(key, types.LoanStatusAlreadyCleared)
	statistics.Overdue = getCountFromCache(key, types.LoanStatusOverdue)
	statistics.Invalid = getCountFromCache(key, types.LoanStatusInvalid)
	statistics.PartialRepayment = getCountFromCache(key, types.LoanStatusPartialRepayment)
	statistics.Loaning = getCountFromCache(key, types.LoanStatusIsDoing)
	statistics.StatisticsDate = tools.GetDate(date / 1000)
}

func GetOrderMonitorKey(date int64) string {
	return getMonitorKey(orderHashPrefix, date)
}

func IsOrderKeyExist(date int64) bool {
	return IsKeyExist(orderHashPrefix, date)
}

func DelOrderKey(date int64) {
	DelKey(orderHashPrefix, date)
}

func IncrOrderCount(field types.LoanStatus)  {
	date := tools.NaturalDay( 0)
	incrCount(orderHashPrefix, date, field)
}