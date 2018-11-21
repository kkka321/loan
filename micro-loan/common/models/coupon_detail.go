package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// COUPON_DETAIL_TABLENAME 表名
const COUPON_DETAIL_TABLENAME string = "coupon_detail"

// COUPON_DETAIL_TABLENAME 描述数据表结构与结构体的映射
type CouponDetail struct {
	Id         int64 `orm:"pk;"`
	CouponId   int64
	TotalNum   int
	UsedNum    int
	SuccNum    int
	CouponDate int64
	UsedRate   int
	SuccRate   int
	Ctime      int64
}

// TableName 返回当前模型对应的表名
func (r *CouponDetail) TableName() string {
	return COUPON_DETAIL_TABLENAME
}

// Using 返回当前模型的数据库
func (r *CouponDetail) Using() string {
	return types.OrmDataBaseApi
}

func (r *CouponDetail) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *CouponDetail) Insert() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err := o.Insert(r)
	return err
}
