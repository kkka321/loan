package models

import (
	"micro-loan/common/types"
)

// LimitDataPrivilege 有限的数据权限, 此处主要存储不随时间线性暴涨的数据权限
// 所有 limitDataPrivilege 动态权限均可以缓存
// 该动态权限可以分配给角色, 也可以分配给对应角色的后台管理人员

// LIMIT_DATA_PRIVILEGE_TABLENAME 表名
const LIMIT_DATA_PRIVILEGE_TABLENAME string = "limit_data_privilege"

// LimitDataPrivilege 动态资源,描述数据表结构与结构体的映射
type LimitDataPrivilege struct {
	Id         int64 `orm:"pk"`
	Type       types.LimitDataPrivilegeTypeEnum
	DataID     int64
	GrantType  types.DataGrantTypeEnum
	GrantTo    int64
	Status     int
	Ctime      int64
	RevokeTime int64
	Utime      int64
}

// TableName 返回当前模型对应的表名
func (r *LimitDataPrivilege) TableName() string {
	return LIMIT_DATA_PRIVILEGE_TABLENAME
}

// Using 返回当前模型的数据库
func (r *LimitDataPrivilege) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *LimitDataPrivilege) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}
