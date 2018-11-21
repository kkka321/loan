package models

import "micro-loan/common/types"

// COUPON_TABLENAME 表名
const COUPON_TABLENAME string = "coupon"

// COUPON_TABLENAME 描述数据表结构与结构体的映射
type Coupon struct {
	Id                int64 `orm:"pk;"`
	Name              string
	DistributeAlgo    string
	DistributeStart   int64
	DistributeEnd     int64
	CouponType        types.CouponType
	DiscountRate      int64
	DiscountDay       int64
	DiscountAmount    int64
	ValidStart        int64
	DistributeAsStart int
	ValidEnd          int64
	ValidDays         int
	ValidMin          int64
	DiscountMax       int64
	IsAvailable       int
	DistributeSize    int64
	UsedNum           int64 `orm:"-"`
	UsedAmount        int64 `orm:"-"`
	DistributeAll     int64 `orm:"-"`
	Comment           string
	Ctime             int64
	Utime             int64
}

// TableName 返回当前模型对应的表名
func (r *Coupon) TableName() string {
	return COUPON_TABLENAME
}

// Using 返回当前模型的数据库
func (r *Coupon) Using() string {
	return types.OrmDataBaseApi
}

func (r *Coupon) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}
