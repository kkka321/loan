package models

import (
	"micro-loan/common/types"
)

type CustomerFollow struct {
	Id         int64 `orm:"pk"`
	CustomerId int64 `orm:"column(customer_id)"`
	FollowTime int64 `orm:"column(follow_time)"`
	OpUid      int64 `orm:"column(op_uid)"`
	Content    string
	Remark     string
	Ctime      int64
}

const CUSTOMER_FOLLOW_TABLENAME string = "customer_follow"

func (r *CustomerFollow) TableName() string {
	return CUSTOMER_FOLLOW_TABLENAME
}

func (r *CustomerFollow) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *CustomerFollow) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}
