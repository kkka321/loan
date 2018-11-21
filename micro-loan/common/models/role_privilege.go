package models

import "micro-loan/common/types"

// ROLE_PRIVILEGE_TABLENAME 表名
const ROLE_PRIVILEGE_TABLENAME string = "role_privilege"

// RolePrivilege 描述数据表结构与结构体的映射
type RolePrivilege struct {
	Id          int64 `orm:"pk;"`
	RoleID      int64 `orm:"column(role_id);"`
	PrivilegeID int64 `orm:"column(privilege_id);"`
	Ctime       int64
}

// TableName 返回当前模型对应的表名
func (r *RolePrivilege) TableName() string {
	return ROLE_PRIVILEGE_TABLENAME
}

// Using 返回当前模型的数据库
func (r *RolePrivilege) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *RolePrivilege) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}
