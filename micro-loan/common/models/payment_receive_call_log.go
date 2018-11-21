package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const PAYMENT_RECEIVE_CALL_LOG_TABLENAME string = "payment_receive_call_log"

type PaymentReceiveCallLog struct {
	Id            int64 `orm:"pk;"`
	UserAccountId int64
	OrderId       int64
	PaymentId     string
	Ctime         int64
}

func (r *PaymentReceiveCallLog) TableName() string {
	return PAYMENT_RECEIVE_CALL_LOG_TABLENAME
}

func (r *PaymentReceiveCallLog) Using() string {
	return types.OrmDataBaseApi
}

func (r *PaymentReceiveCallLog) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *PaymentReceiveCallLog) Add() (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Insert(r)
	return
}
func (r *PaymentReceiveCallLog) Update(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)
	return
}
