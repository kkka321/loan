package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const ACCOUNT_BALANCE_TABLENAME string = "account_balance"

type AccountBalance struct {
	AccountId     int64 `orm:"pk;"`
	Balance       int64
	FrozenBalance int64
	Ctime         int64
	Utime         int64
}

func (r *AccountBalance) TableName() string {
	return ACCOUNT_BALANCE_TABLENAME
}

func (r *AccountBalance) Using() string {
	return types.OrmDataBaseApi
}

func (r *AccountBalance) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *AccountBalance) Add() (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Insert(r)
	return
}
func (r *AccountBalance) Update(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)
	return
}
