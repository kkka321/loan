package performance

import (
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/types"
)

// 处理量绩效点比率配置名map
var handlePointRateConfigNameMap = map[types.TicketItemEnum]string{
	types.TicketItemRepayRemind: "ticket_handle_point_rate_repay_remind",
	types.TicketItemUrgeM11:     "ticket_handle_point_rate_urge_m11",
	types.TicketItemUrgeM12:     "ticket_handle_point_rate_urge_m12",
	types.TicketItemUrgeM13:     "ticket_handle_point_rate_urge_m13",
}

// 完成工单绩效点比率配置名map
var completePointRateConfigNameMap = map[types.TicketItemEnum]string{
	types.TicketItemRepayRemind: "ticket_complete_point_rate_repay_remind",
	types.TicketItemUrgeM11:     "ticket_complete_point_rate_urge_m11",
	types.TicketItemUrgeM12:     "ticket_complete_point_rate_urge_m12",
	types.TicketItemUrgeM13:     "ticket_complete_point_rate_urge_m13",
}

// 案子逾期率标准配置名map
var caseOverdueRateStandardConfigNameMap = map[types.TicketItemEnum]string{
	types.TicketItemRepayRemind: "ticket_overdue_rate_standard_repay_remind",
	types.TicketItemUrgeM11:     "ticket_overdue_rate_standard_urge_m11",
	types.TicketItemUrgeM12:     "ticket_overdue_rate_standard_urge_m12",
	types.TicketItemUrgeM13:     "ticket_overdue_rate_standard_urge_m13",
}

// 回款比率配置名map
var repayRateConfigNameMap = map[types.TicketItemEnum]string{
	types.TicketItemRM0:     "ticket_target_repay_rate_rm0",
	types.TicketItemRM1:     "ticket_target_repay_rate_rm1",
	types.TicketItemUrgeM11: "ticket_target_repay_rate_urge_m11",
	types.TicketItemUrgeM12: "ticket_target_repay_rate_urge_m12",
}

func getHandlePointRateConfigName(ticketItem types.TicketItemEnum) string {
	return handlePointRateConfigNameMap[ticketItem]
}

func getCompletePointRateConfigName(ticketItem types.TicketItemEnum) string {
	return completePointRateConfigNameMap[ticketItem]
}

func getCaseOverdueRateStandardConfigName(ticketItem types.TicketItemEnum) string {
	return caseOverdueRateStandardConfigNameMap[ticketItem]
}

func getUrgeRepayRateStandardConfigName(ticketItem types.TicketItemEnum) string {
	return repayRateConfigNameMap[ticketItem]
}

// 小组回款比率配置名map
var groupRepayRateConfigNameMap = map[types.TicketItemEnum]string{
	types.TicketItemUrgeM11: "ticket_group_target_repay_rate_urge_m11",
	types.TicketItemUrgeM12: "ticket_group_target_repay_rate_urge_m12",
	types.TicketItemRM0:     "ticket_group_target_repay_rate_urge_rm0",
}

func getUrgeGroupTargetRepayRateConfigName(ticketItem types.TicketItemEnum) string {
	return groupRepayRateConfigNameMap[ticketItem]
}

// GetUrgeGroupTargetRepayRate 获取小组标准回款率
func GetUrgeGroupTargetRepayRate(ticketItem types.TicketItemEnum) float64 {
	configName := getUrgeGroupTargetRepayRateConfigName(ticketItem)
	rate, _ := config.ValidItemFloat64(configName)
	return rate
}

// Bonus 与回款金额比率
var salaryRepayAmountRateMap = map[types.TicketItemEnum]string{
	types.TicketItemUrgeM11: "ticket_M11_salary_repay_amount_rate",
	types.TicketItemUrgeM12: "ticket_M12_salary_repay_amount_rate",
	types.TicketItemRM0:     "ticket_RM0_salary_repay_amount_rate",
}

func getItemSalaryRepayRateConfigName(ticketItem types.TicketItemEnum) string {
	return salaryRepayAmountRateMap[ticketItem]
}
