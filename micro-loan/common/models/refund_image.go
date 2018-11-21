package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const REFUND_IMAGE_TABLENAME string = "refund_image"

type RefundImage struct {
	Id            int64 `orm:"pk;"`
	UserAccountId int64
	RefundId      int64
	Image0Id      int64
	Image1Id      int64
	Image2Id      int64
	Image3Id      int64
	Image4Id      int64
	Ctime         int64
}

func (r *RefundImage) TableName() string {
	return REFUND_IMAGE_TABLENAME
}

func (r *RefundImage) Using() string {
	return types.OrmDataBaseApi
}

func (r *RefundImage) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *RefundImage) Add() (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Insert(r)
	return
}

// func (r *RefundImage) Update(cols ...string) (id int64, err error) {
// 	o := orm.NewOrm()
// 	o.Using(r.Using())

// 	id, err = o.Update(r, cols...)
// 	return
// }
