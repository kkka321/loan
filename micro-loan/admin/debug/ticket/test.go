package main

import (
	"fmt"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/pkg/ticket/performance"
	"micro-loan/common/types"
)

func main() {
	//repayremind.TryCompleteCaseByCleared(180318020023511385)
	performance.DailyWorkerProcessHistoryStatsByDay("20180820")

	fmt.Println(types.GetRepayRemindCaseExpireTime(types.TicketItemRMAdvance1, 0))
	fmt.Println(types.GetRepayRemindCaseExpireTime(types.TicketItemRM0, 0))
	fmt.Println(types.GetRepayRemindCaseExpireTime(types.TicketItemRM1, 0))
}
