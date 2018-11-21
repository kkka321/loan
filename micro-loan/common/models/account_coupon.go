package models

import (
	"micro-loan/common/types"
)

// ACCOUNT_COUPON_TABLENAME 表名
const ACCOUNT_COUPON_TABLENAME string = "account_coupon"

// AccountCoupon 描述数据表结构与结构体的映射
type AccountCoupon struct {
	Id            int64 `orm:"pk;"`
	CouponId      int64
	UserAccountId int64
	OrderId       int64
	Status        types.CouponStatus
	Amount        int64
	UsedTime      int64
	EffectiveDate int64
	ExpireDate    int64
	ValidStart    int64
	ValidEnd      int64
	IsNew         int
	Utime         int64
	Ctime         int64
}

// TableName 返回当前模型对应的表名
func (r *AccountCoupon) TableName() string {
	return ACCOUNT_COUPON_TABLENAME
}

// Using 返回当前模型的数据库
func (r *AccountCoupon) Using() string {
	return types.OrmDataBaseApi
}

func (r *AccountCoupon) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}
