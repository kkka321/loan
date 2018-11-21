package types

// DataPrivilegeTypeEnum 描述角色类型
type DataPrivilegeTypeEnum int

// 动态权限类型
const (
	DataPrivilegeTypeOrder           DataPrivilegeTypeEnum = 1
	DataPrivilegeTypeCustomer        DataPrivilegeTypeEnum = 2
	DataPrivilegeTypeOverdueCase     DataPrivilegeTypeEnum = 3
	DataPrivilegeTypeRepayRemindCase DataPrivilegeTypeEnum = 4
)

// DataGrantTypeEnum 资源授权类型
type DataGrantTypeEnum int

const (
	// DataGrantUser 资源授权给用户
	DataGrantUser DataGrantTypeEnum = 1
	// DataGrantRole 资源授权给角色
	DataGrantRole DataGrantTypeEnum = 2
)

// LimitDataPrivilegeTypeEnum 描述角色类型
type LimitDataPrivilegeTypeEnum int

const (
	// LimitDataPrivilegeTypeTicketItem ticketItem 权限
	LimitDataPrivilegeTypeTicketItem LimitDataPrivilegeTypeEnum = 1
)
