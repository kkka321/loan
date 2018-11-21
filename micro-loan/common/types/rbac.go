package types

// RoleSuperPid 超级角色PID
const RoleSuperPid = 0

// 级别之间相差10 , 保留扩展空间
// 暂时未使用
const (
	// RoleSuper 超级用户类型
	RoleSuper = 1
	// RoleLeader 部门Owner, 每添加一个部门,应该自动加一个 Owner
	RoleLeader = 10
	// RoleEmploee 雇员
	RoleEmployee = 20
)

// roleLevelMap 角色类型列表
var roleLevelMap = map[int]string{
	RoleSuper:    "Super",
	RoleLeader:   "Leader",
	RoleEmployee: "Employee",
}

// RoleLevelMap 读取角色等级列表
func RoleLevelMap() map[int]string {
	return roleLevelMap
}

// RoleTypeEnum 描述角色类型
type RoleTypeEnum int

// 命名注意事项，　系统角色组　RoleType-System 放在前面，　方便统一调用
const (
	RoleTypeSystem          RoleTypeEnum = 1
	RoleTypePhoneVerify     RoleTypeEnum = 2
	RoleTypeRiskCtl         RoleTypeEnum = 3
	RoleTypeUrge            RoleTypeEnum = 4
	RoleTypeRepayReminder   RoleTypeEnum = 5
	RoleTypeCustomerService RoleTypeEnum = 6
)

var lowPrivilegeRoleTypeContainer = []RoleTypeEnum{
	RoleTypePhoneVerify,
	RoleTypeUrge,
	RoleTypeRepayReminder,
	RoleTypeCustomerService,
}

// LowPrivilegeRoleTypeContainer 返回低权限角色类型
func LowPrivilegeRoleTypeContainer() []RoleTypeEnum {
	return lowPrivilegeRoleTypeContainer
}

var roleTypeMap = map[RoleTypeEnum]string{
	RoleTypeSystem:          "System",
	RoleTypePhoneVerify:     "Phone Verify",
	RoleTypeRiskCtl:         "RiskCtl",
	RoleTypeUrge:            "Urge",
	RoleTypeRepayReminder:   "Repay Reminder",
	RoleTypeCustomerService: "Customer Service",
}

// RoleTypeMap 读取 roleTypeMap
func RoleTypeMap() map[RoleTypeEnum]string {
	return roleTypeMap
}

// SuperRolePid 超管角色父ID
const SuperRolePid = 0

// RBACBaseOpeartionList RBAC 不需要控制的基本操作集合
// 此处的集合为,基本操作, 不需要分配权限的操作集合,
// 目的: 减少后台操作,简化上线流程
var RBACBaseOpeartionList = []string{
	"IndexController@GET",
	"AdminController@Password",
	"AdminController@FixPassword",
	"TicketController@Me",
	"TicketController@UpdateStatus",
	"TicketController@UpdateMyOnlineStatus",
	"IndexController@Dashboard",
}
