package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/tools"
	"micro-loan/common/types"
)

const ORDERS_EXT_TABLENAME string = "orders_ext"

type OrderExt struct {
	OrderId            int64 `orm:"pk;"`
	OverdueRunTime     int64
	RepayMsgRunTime    int64
	IsEntrust          int64
	EntrustPname       string
	EntrustTime        int64
	SpecialLoanCompany int
	PhyInvalidTag      int
	QuotaIncreased     int64
	Ctime              int64
	Utime              int64
}

// TableName 返回当前模型对应的表名
func (r *OrderExt) TableName() string {
	return ORDERS_EXT_TABLENAME
}

// Using 返回当前模型的数据库
func (r *OrderExt) Using() string {
	return types.OrmDataBaseApi
}

func (r *OrderExt) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *OrderExt) Add() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.Ctime = tools.GetUnixMillis()

	id, err := o.Insert(r)

	return id, err
}

func (r *OrderExt) Update() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	return o.Update(r)
}

func GetOrderExt(orderId int64) (OrderExt, error) {
	o := orm.NewOrm()
	orderExt := OrderExt{}
	o.Using(orderExt.Using())
	err := o.QueryTable(orderExt.TableName()).Filter("order_id", orderId).One(&orderExt)

	return orderExt, err
}
