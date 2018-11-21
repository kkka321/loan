package ticket

import (
	"micro-loan/common/types"
)

// 分配模式对应分配策略
const (
	RealtimeAssign = "IdleWorkerStrategy"
	DailyAvg       = "DayAvgAlternateWorkerStrategy"
)

func getItemWorkerStrategy(ticketItem types.TicketItemEnum) WorkerAssignStrategy {
	strategyConfig := ticketItemAssignMode[ticketItem]
	switch strategyConfig {
	case RealtimeAssign:
		return &IdleWorkerStrategy{ticketItem}
	case DailyAvg:
		return &DayAvgAlternateWorkerStrategy{TicketItem: ticketItem}
	default:
		//logs.Error("[getItemWorkerStrategy] no worker strategy for ticket item:", ticketItem)
		return &EmptyWorkerStrategy{TicketItem: ticketItem}
	}
}

// GetItemWorkerStrategy 获取 ticketItem 配置的人力策略
func GetItemWorkerStrategy(ticketItem types.TicketItemEnum) WorkerAssignStrategy {
	return getItemWorkerStrategy(ticketItem)
}

var ticketItemAssignMode = map[types.TicketItemEnum]string{
	types.TicketItemPhoneVerify: RealtimeAssign,
	types.TicketItemInfoReview:  RealtimeAssign,
	types.TicketItemUrgeM11:     DailyAvg,
	types.TicketItemUrgeM12:     DailyAvg,
	// types.TicketItemUrgeM13:     DailyAvg,
	// types.TicketItemUrgeM20:     DailyAvg,
	// types.TicketItemUrgeM30:     DailyAvg,
	// types.TicketItemRepayRemind: DailyAvg,
	//types.TicketItemRMAdvance1: DailyAvg,
	types.TicketItemRM0: DailyAvg,
	//types.TicketItemRM1: DailyAvg,
}

// RealtimeAssignTicketItems 返回实时分配的ticket item slice
func RealtimeAssignTicketItems() []types.TicketItemEnum {
	var tis []types.TicketItemEnum
	for ti, mode := range ticketItemAssignMode {
		if mode == RealtimeAssign {
			tis = append(tis, ti)
		}
	}
	return tis
}

// DailyAvgAssignTicketItems 返回日均分单的 ticket item slice
func DailyAvgAssignTicketItems() []types.TicketItemEnum {
	var tis []types.TicketItemEnum
	for ti, mode := range ticketItemAssignMode {
		if mode == DailyAvg {
			tis = append(tis, ti)
		}
	}
	return tis
}

func isDailyAvg(ticketItem types.TicketItemEnum) bool {
	if mode := ticketItemAssignMode[ticketItem]; mode == DailyAvg {
		return true
	}
	return false
}
