package models

import (
	"micro-loan/common/types"
)

// TICKET_ASSIGN_CONFIG_TABLENAME 表名
const TICKET_ASSIGN_CONFIG_TABLENAME string = "ticket_assign_config"

// TicketAssignConfig 工单,描述数据表结构与结构体的映射
type TicketAssignConfig struct {
	Id           int64                `orm:"pk"`
	TicketItemID types.TicketItemEnum `orm:"column(ticket_item_id)"` // 决定默认优先级和分配给什么角色
	AssignRoles  string               `orm:"column(assign_roles)"`
	Mode         int
	Status       int
	Ctime        int64
	Utime        int64
}

// TableName 返回当前模型对应的表名
func (r *TicketAssignConfig) TableName() string {
	return TICKET_ASSIGN_CONFIG_TABLENAME
}

// Using 返回当前模型的数据库
func (r *TicketAssignConfig) Using() string {
	return types.OrmDataBaseAdmin
}
func (r *TicketAssignConfig) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}
