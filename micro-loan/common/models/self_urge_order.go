package models

import "micro-loan/common/types"

const SELF_URGE_ORDER = "self_urge_order"

// SelfUrgeOrder 描述orm映射
type SelfUrgeOrder struct {
	Id         int64 `orm:"pk;"`
	OrderId    int64
	ExpireTime int64
	IsDeleted  int
	Ctime      int64
	Utime      int64
}

// TableName 返回对应表名
func (r *SelfUrgeOrder) TableName() string {
	return SELF_URGE_ORDER
}

// Using 主库
func (r *SelfUrgeOrder) Using() string {
	return types.OrmDataBaseApi
}

// UsingSlave 使用从库连接
func (r *SelfUrgeOrder) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}
