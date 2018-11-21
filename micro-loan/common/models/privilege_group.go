package models

import "micro-loan/common/types"

// PRIVILEGE_GROUP_TABLENAME 表名
const PRIVILEGE_GROUP_TABLENAME string = "privilege_group"

// PrivilegeGroup 描述数据表结构与结构体的映射
type PrivilegeGroup struct {
	Id    int64 `orm:"pk;"`
	Name  string
	Ctime int64
	Utime int64
}

// TableName 返回当前模型对应的表名
func (r *PrivilegeGroup) TableName() string {
	return PRIVILEGE_GROUP_TABLENAME
}

// Using 返回当前模型的数据库
func (r *PrivilegeGroup) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *PrivilegeGroup) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}
