package models

import "micro-loan/common/types"

const TICKET_TAG_TABLENAME = "ticket_tag"

// TicketTag 描述映射表
type TicketTag struct {
	Id       int64
	TicketID int64 `orm:"column(ticket_id);"`
	TagID    int64 `orm:"column(tag_id);"`
	Ctime    int64
}

// TableName 当前模型对应的表名
func (r *TicketTag) TableName() string {
	return TICKET_TAG_TABLENAME
}

// Using 当前模型的主库
func (r *TicketTag) Using() string {
	return types.OrmDataBaseAdmin
}

// UsingSlave 返回从库
func (r *TicketTag) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}
