package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const REFUND_TABLENAME string = "refund"

type Refund struct {
	Id            int64 `orm:"pk;"`
	UserAccountId int64
	Amount        int64
	Fee           int64
	ReleatedOrder int64
	CheckStatus   int
	RefundType    int
	OpUid         int64
	CallTime      int64
	ResponseTime  int64
	Ctime         int64
	Utime         int64
}

func (r *Refund) TableName() string {
	return REFUND_TABLENAME
}

func (r *Refund) Using() string {
	return types.OrmDataBaseApi
}

func (r *Refund) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *Refund) Add() (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Insert(r)
	return
}
func (r *Refund) Update(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)
	return
}
