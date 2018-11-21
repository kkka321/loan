package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// ORDER_STATISTICS_TABLENAME 表名
const ORDER_STATISTICS_TABLENAME string = "order_statistics"

// OrderStatistics 描述数据表结构与结构体的映射
type OrderStatistics struct {
	Id               int64 `orm:"pk;"`
	Submit           int
	WaitReview       int
	Reject           int
	WaitManual       int
	WaitLoan         int
	LoanFail         int
	WaitRepayment    int
	Cleared          int
	Overdue          int
	Invalid          int
	PartialRepayment int
	Loaning          int
	WaitAutoCall     int
	StatisticsDate   string
	Ctime            int64
}

// TableName 返回当前模型对应的表名
func (r *OrderStatistics) TableName() string {
	return ORDER_STATISTICS_TABLENAME
}

// Using 返回当前模型的数据库
func (r *OrderStatistics) Using() string {
	return types.OrmDataBaseApi
}

func (r *OrderStatistics) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

// Add 添加新的权限
func (r *OrderStatistics) Add() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.Ctime = tools.GetUnixMillis()

	id, err := o.Insert(r)
	r.Id = id

	return id, err
}

// Del 添加新的权限
func (r *OrderStatistics) Del() error {
	o := orm.NewOrm()
	o.Using(r.Using())

	_, err := o.Delete(r)

	return err
}
