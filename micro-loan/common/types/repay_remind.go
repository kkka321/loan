package types

import "sort"

// RepayRemindAdvanceDays 还款提醒提前天数
const RepayRemindAdvanceDays = 1

// 逾期案件等级
const (
	RMLevelAdvance1 = "RM-1" // 逾期2—7天，M1-1
	RMLevel0        = "RM0"  // 逾期8—15天，M1-2
	RMLevel1        = "RM1"  // 逾期16—30天，M1-3
)

// rmCaseCreateDaysMap 创建案子时间点与case的对应关系
var rmCaseCreateDaysMap = map[string]int{
	RMLevelAdvance1: RMLevelAdvance1CreateDay,
	RMLevel0:        RMLevel0CreateDay,
	RMLevel1:        RMLevel1CreateDay,
}

// RMCaseCreateDaysMap 返回受保护的 case 创建天map
func RMCaseCreateDaysMap() map[string]int {
	return rmCaseCreateDaysMap
}

// 还款提醒案件， 时间范围定义
// 只记录初始日期， 减少复杂性， 逾期案件为一个接着一个， 无日期跳跃
const (
	RMLevelAdvance1CreateDay = -1
	RMLevel0CreateDay        = 0
	RMLevel1CreateDay        = 1
)

var rmLevelTicketItemMap = map[string]TicketItemEnum{
	RMLevelAdvance1: TicketItemRMAdvance1,
	RMLevel0:        TicketItemRM0,
	RMLevel1:        TicketItemRM1,
}

// RMLevelTicketItemMap 读取逾期案件等级与ticket类型的映射关系表
func RMLevelTicketItemMap() map[string]TicketItemEnum {
	return rmLevelTicketItemMap
}

////// 用于对数据表risk_monitor.bill_repay_hour_deadline中还款提醒级别的转换
const (
	DLevelAdvance1 = "D-1" // 还款日期前一天
	DLevel0        = "D0"  // 还款当天
	DLevel1        = "D1"  // 还款日期后一天
)

var rmLevelRelatedMap = map[string]string{
	RMLevelAdvance1: DLevelAdvance1,
	RMLevel0:        DLevel0,
	RMLevel1:        DLevel1,
}

// RMLevelRelatedItemMap 用于对数据表risk_monitor.bill_repay_hour_deadline中还款提醒级别的转换
func RMLevelRelatedItemMap() map[string]string {
	return rmLevelRelatedMap
}

// GetRepayRemindCaseExpireTime 返回案件创建是在逾期的哪一天配置列表
func GetRepayRemindCaseExpireTime(ticketItem TicketItemEnum, repayDate int64) int64 {
	level := ticketItemMap[ticketItem]

	createDays := rmCaseCreateDaysMap[level]
	//
	var willUpSlice []int
	for _, d := range rmCaseCreateDaysMap {
		if d > createDays {
			willUpSlice = append(willUpSlice, d)
		}
	}
	if len(willUpSlice) == 0 {
		return repayDate + 2*24*3600*1000
	}
	sort.Ints(willUpSlice)
	return repayDate + int64(willUpSlice[0])*24*3600*1000
}
