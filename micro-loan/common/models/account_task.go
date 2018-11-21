package models

import "micro-loan/common/types"

// ACCOUNT_TASK_TABLENAME 表名
const ACCOUNT_TASK_TABLENAME string = "account_task"

// AccountCoupon 描述数据表结构与结构体的映射
type AccountTask struct {
	Id         int64 `orm:"pk;"`
	AccountId  int64
	InviterId  int64
	CouponId   int64
	TaskType   types.AccountTask
	TaskStatus types.AccountTaskStatus
	DoneTime   int64
	Utime      int64
	Ctime      int64
}

// TableName 返回当前模型对应的表名
func (r *AccountTask) TableName() string {
	return ACCOUNT_TASK_TABLENAME
}

// Using 返回当前模型的数据库
func (r *AccountTask) Using() string {
	return types.OrmDataBaseApi
}

func (r *AccountTask) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}
