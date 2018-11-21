package types

import (
	"fmt"
	"sync"

	"github.com/astaxie/beego/logs"
)

// 工单关闭原因
const (
	TicketCloseReasonAbnormal string = "Abnormal"
	TicketCloseReasonNoWork   string = "No Need Handle"
	TicketCloseReasonCaseUp   string = "Case Up"
	TicketCloseReasonEntrust  string = "Already Entrust"
)

// 目标回收率类型
const (
	TicketMyProcess         int = 1 // 工作进度
	TicketPerformanceManage int = 2 // 人员绩效管理
)

// TicketPirorityEnum 工单优先级
type TicketPirorityEnum int

// Ticket 紧急程度优先级
const (
	// TicketPirorityEmergency 紧急事务
	TicketPirorityEmergency TicketPirorityEnum = 2
	// TicketPirorityUrgent 紧迫的
	TicketPirorityUrgent TicketPirorityEnum = 4
	// TicketPirorityGeneral 常规
	TicketPirorityGeneral TicketPirorityEnum = 8
)

var ticketPriorityMap = map[TicketPirorityEnum]string{
	TicketPirorityEmergency: "Emergency",
	TicketPirorityUrgent:    "Urgent",
	TicketPirorityGeneral:   "General",
}

// TicketItemEnum 工单优先级
type TicketItemEnum int

// Ticket 具体项
const (
	TicketItemPhoneVerify TicketItemEnum = 1
	TicketItemInfoReview  TicketItemEnum = 2
	TicketItemUrgeM11     TicketItemEnum = 3
	TicketItemUrgeM12     TicketItemEnum = 4
	TicketItemUrgeM13     TicketItemEnum = 5
	TicketItemUrgeM20     TicketItemEnum = 6
	TicketItemUrgeM30     TicketItemEnum = 7
	TicketItemRepayRemind TicketItemEnum = 8
	TicketItemRMAdvance1  TicketItemEnum = 9
	TicketItemRM0         TicketItemEnum = 10
	TicketItemRM1         TicketItemEnum = 11
)

var ticketItemMap = map[TicketItemEnum]string{
	TicketItemPhoneVerify: "Phone Verify",
	TicketItemInfoReview:  "Info Review",
	TicketItemUrgeM11:     OverdueLevelM11,
	TicketItemUrgeM12:     OverdueLevelM12,
	TicketItemUrgeM13:     OverdueLevelM13,
	TicketItemUrgeM20:     OverdueLevelM2,
	TicketItemUrgeM30:     OverdueLevelM3,
	TicketItemRepayRemind: "Repay Remind",
	TicketItemRMAdvance1:  RMLevelAdvance1,
	TicketItemRM0:         RMLevel0,
	TicketItemRM1:         RMLevel1,
}

var reverseTicketItemMap = map[string]TicketItemEnum{}
var onceInitReverseTicketmap = &sync.Once{}

// GetTicketItemIDByCaseName 根据Case名获取 Ticket Item ID
func GetTicketItemIDByCaseName(caseName string) (TicketItemEnum, error) {
	//
	onceInitReverseTicketmap.Do(func() {
		if len(reverseTicketItemMap) == 0 {
			for itemID, name := range ticketItemMap {
				reverseTicketItemMap[name] = itemID
			}
		}
	})

	//temp compatible for repay remind, TODO remove start
	if caseName == "" {
		return TicketItemRepayRemind, nil
	}
	// TODO remove end

	if v, ok := reverseTicketItemMap[caseName]; ok {
		return v, nil
	}

	return 0, fmt.Errorf("[GetTicketItemIDByCaseName]CaseName(%s) have no related item id", caseName)
}

// MustGetTicketItemIDByCaseName 根据Case名获取 Ticket Item ID
func MustGetTicketItemIDByCaseName(caseName string) TicketItemEnum {
	itemID, err := GetTicketItemIDByCaseName(caseName)
	if err != nil {
		logs.Error("[MustGetTicketItemIDByCaseName] should no happended:", err)
	}
	return itemID
}

// 角色类型与工单类型权限对应列表
// 系统管理员和风控拥有所有工单类型权限
var roleTicketItems = map[RoleTypeEnum][]TicketItemEnum{
	RoleTypePhoneVerify: {TicketItemPhoneVerify, TicketItemInfoReview},
	RoleTypeUrge: {
		TicketItemUrgeM11,
		TicketItemUrgeM12,
		TicketItemUrgeM13,
		TicketItemUrgeM20,
		TicketItemUrgeM30,
	},
	RoleTypeRepayReminder: {
		TicketItemRepayRemind,
		TicketItemRMAdvance1,
		TicketItemRM0,
		TicketItemRM1,
	},
}

// TicketItemMap return Map
func TicketItemMap() map[TicketItemEnum]string {
	return ticketItemMap
}

// OwnTicketItemMap 获取该角色拥有的 ticket item 类型
// 系统管理员和风控拥有所有工单类型权限
func OwnTicketItemMap(roleType RoleTypeEnum) map[TicketItemEnum]string {
	if roleType == RoleTypeSystem || roleType == RoleTypeRiskCtl {
		return TicketItemMap()
	}
	ownTicketItems := make(map[TicketItemEnum]string)
	if l, ok := roleTicketItems[roleType]; ok {
		for _, i := range l {
			if v, ok := ticketItemMap[i]; ok {
				ownTicketItems[i] = v
			}
		}
	}
	return ownTicketItems
}

// TicketStatusEnum 工单状态-枚举类型定义
type TicketStatusEnum int

// Ticket 状态
const (
	TicketStatusCreated          TicketStatusEnum = 0
	TicketStatusAssigned         TicketStatusEnum = 1
	TicketStatusProccessing      TicketStatusEnum = 3
	TicketStatusCompleted        TicketStatusEnum = 4
	TicketStatusClosed           TicketStatusEnum = 5
	TicketStatusPartialCompleted TicketStatusEnum = 6
	TicketStatusWaitingEntrust   TicketStatusEnum = 7
)

var ticketStatusSliceInDoing = []TicketStatusEnum{
	TicketStatusAssigned,
	TicketStatusProccessing,
	TicketStatusPartialCompleted,
}

// TicketStatusSliceInDoing 返回正在工作中的状态集合
// 目前主要用于统计
func TicketStatusSliceInDoing() []TicketStatusEnum {
	return ticketStatusSliceInDoing
}

var ticketStatusMap = map[TicketStatusEnum]string{
	TicketStatusCreated:          "已创建",
	TicketStatusAssigned:         "已分配",
	TicketStatusProccessing:      "进行中",
	TicketStatusCompleted:        "已完成",
	TicketStatusClosed:           "已关闭",
	TicketStatusPartialCompleted: "部分完成",
	TicketStatusWaitingEntrust:   "等待委外审批",
}

// TicketStatusMap return Map
func TicketStatusMap() map[TicketStatusEnum]string {
	return ticketStatusMap
}

var ticketAssignRoleConfigNameMap = map[TicketItemEnum]string{
	TicketItemPhoneVerify: "ticket_assign_role_phone_verify",
	TicketItemInfoReview:  "ticket_assign_role_info_review",
	TicketItemUrgeM11:     "ticket_assign_role_urge_m11",
	TicketItemUrgeM12:     "ticket_assign_role_urge_m12",
	TicketItemUrgeM13:     "ticket_assign_role_urge_m13",
	TicketItemUrgeM20:     "ticket_assign_role_urge_m20",
	TicketItemUrgeM30:     "ticket_assign_role_urge_m30",
	TicketItemRepayRemind: "ticket_assign_role_repay_remind",
	TicketItemRMAdvance1:  "ticket_assign_role_rm-1",
	TicketItemRM0:         "ticket_assign_role_rm0",
	TicketItemRM1:         "ticket_assign_role_rm1",
}

// TicketAssignRoleConfigNameMap 返回
func TicketAssignRoleConfigNameMap() map[TicketItemEnum]string {
	return ticketAssignRoleConfigNameMap
}

// TicketQueueNameVarTemp 工单分配池队列名称中包含变量的模板类型

// 工单分配池队列名称中包含变量的模板,
// 模板便于做动态匹配指定角色或者指定ticket item的分配池, 便于操作队列
const (
	TicketQueueNameItemVar   string = "TI%d_"
	TicketQueueNameRoleIDVar string = "R%s_"
)

// TicketAssignPollQueueNameTemplate queueNamePrefix + ItemID + RoleIDs
const TicketAssignPollQueueNameTemplate = "%s%s%s"

// 后台员工工作状态
const (
	AdminWorkStatusNormal = 1
	AdminWorkStatusStop   = 2
)

var communicationWayMap = map[int]string{
	1: "Whatsapp",
	2: "电话",
}

// CommnicationWayMap 返回交流方式map
func CommnicationWayMap() map[int]string {
	return communicationWayMap
}
