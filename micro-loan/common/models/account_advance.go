package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const ACCOUNT_ADVANCE_TABLENAME string = "account_advance"

// AccountAdvance 表结构
type AccountAdvance struct {
	Id        int64 `orm:"pk;"`
	AccountId int64
	OrderId   int64
	Response  string
	Type      int //1 blacklist  2 mulit platform
	Ctime     int64
}

func (r *AccountAdvance) TableName() string {
	return ACCOUNT_ADVANCE_TABLENAME
}

func (r *AccountAdvance) Using() string {
	return types.OrmDataBaseApi
}

func (r *AccountAdvance) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *AccountAdvance) Insert() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err := o.Insert(r)
	return err
}

func GetAdvanceBlacklist(accountId, orderId int64) (AccountAdvance, error) {
	m := AccountAdvance{}
	o := orm.NewOrm()
	o.Using(m.Using())

	err := o.QueryTable(m.TableName()).
		Filter("account_id", accountId).
		Filter("order_id", orderId).
		Filter("type", 1).
		OrderBy("-ctime").
		One(&m)

	return m, err
}
