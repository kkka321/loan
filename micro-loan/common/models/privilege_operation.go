package models

import "micro-loan/common/types"

// PRIVILEGE_OPERATION_TABLENAME 表名
const PRIVILEGE_OPERATION_TABLENAME string = "privilege_operation"

// PrivilegeOperation 描述数据表结构与结构体的映射
type PrivilegeOperation struct {
	Id          int64 `orm:"pk;"`
	PrivilegeID int64 `orm:"column(privilege_id);"`
	OperationID int64 `orm:"column(operation_id);"`
	Ctime       int64
}

// TableName 返回当前模型对应的表名
func (r *PrivilegeOperation) TableName() string {
	return PRIVILEGE_OPERATION_TABLENAME
}

// Using 返回当前模型的数据库
func (r *PrivilegeOperation) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *PrivilegeOperation) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}
