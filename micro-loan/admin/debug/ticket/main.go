package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/pkg/ticket/performance"
	"micro-loan/common/types"
)

func main() {
	// var orderID, userAccountID int64 = 180318020023511385, 180318010023497558
	// ticket.CreateTicket(types.TicketItemRepayRemind, 190, types.Robot, orderID, userAccountID, nil)

	// ticket.CreateTicket(types.TicketItemPhoneVerify, 180523020015578858, types.Robot,
	// 	ticket.DataRepayRemindCase{OrderID: orderID, CustomerID: userAccountID})
	//event.Trigger(&evtypes.TicketCreateEv{types.TicketItemPhoneVerify, types.Robot, 180511020007458399, nil})

	//ticket.WorkerIncompletedTicketsByTicketItem(types.TicketItemPhoneVerify)
	//result, err := ticket.IdleAssign(types.TicketItemPhoneVerify)
	// logs.Debug("Assign user:", result)
	// logs.Debug(err)
	//updatedOrder, _ := models.GetOrder(180723020209149649)
	//ticketModel, _ := models.GetTicket(9926)
	//ticket.CompleteByRelatedID(9925, 4)
	// ticket.CloseByRelatedID(9958, 3, "hha")
	// service.HandleOverdueCase(180702020169366173)
	//ticket.CheckEarlyWorkerDailyAssignByItem(types.TicketItemRepayRemind)
	// testStats()
	//testStatsMonth()
	//fmt.Println(tools.PasswordEncrypt("123456", 1516710209872))
	//performance.DailyWorkerPerformanceStatsTask("20180712")
	//ticket.ReopenTicket(180702020167688844, types.TicketItemPhoneVerify)
	// fmt.Println(models.GetTicketForPhoneVerifyOrInfoReivew(180702020167688844))
	// ticket.CompletePhoneVerifyOrInfoReviewByRelatedID(180702020167688844)

	ticket.CreateTicket(types.TicketItemUrgeM11, 20181113, types.Robot, 180724020209809565, 180724020209809565, nil)
	//orderData, _ := models.GetOrder(180724020214747671)
	//ticket.CreateAfterRisk(orderData)
	//ticket.ReopenPhoneVerifyOrInfoReviewByRelatedID(180702020167688844)
	//ticket.ReopenByRelatedID(180830020000009095, types.TicketItemPhoneVerify, "12:00-14:00")

}

func testStats() {
	//performance.DailyWorkerPerformanceStatsTask("20180701")
	//performance.LastDayDailyWorkerPerformanceStatsTask()
	defer performance.DailyWorkerPerformanceStatsTask("20180712")
	defer performance.DailyWorkerPerformanceStatsTask("20180711")
	defer performance.DailyWorkerPerformanceStatsTask("20180710")
	defer performance.DailyWorkerPerformanceStatsTask("20180709")
	defer performance.DailyWorkerPerformanceStatsTask("20180708")
	defer performance.DailyWorkerPerformanceStatsTask("20180707")
	defer performance.DailyWorkerPerformanceStatsTask("20180706")
	defer performance.DailyWorkerPerformanceStatsTask("20180705")
	defer performance.DailyWorkerPerformanceStatsTask("20180704")
	defer performance.DailyWorkerPerformanceStatsTask("20180703")
	defer performance.DailyWorkerPerformanceStatsTask("20180702")
	defer performance.DailyWorkerPerformanceStatsTask("20180630")
	defer performance.DailyWorkerPerformanceStatsTask("20180629")
	defer performance.DailyWorkerPerformanceStatsTask("20180628")
	defer performance.DailyWorkerPerformanceStatsTask("20180627")
	defer performance.DailyWorkerPerformanceStatsTask("20180626")
}

func testHourlyStats() {
	performance.TodayHourlyWorkerPerformanceStatsTask()
	// performance.HourlyWorkerPerformanceStatsTask("2018071222")
}

func testStatsMonth() {
	performance.UpdateCurrentMonthStats()

}
