package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const REGISTER_REMIND_MESSAGE_TABLENAME string = "register_remind_message"

type RegisterRemindMessage struct {
	Id    int64  `orm:"pk;"`
	Date  string `orm:"column(date)"`
	Count int    `orm:"column(count)"`
	Ctime int64  `orm:"column(ctime)"`
	Utime int64  `orm:"column(utime)"`
}

func (r *RegisterRemindMessage) TableName() string {
	return REGISTER_REMIND_MESSAGE_TABLENAME
}

func (r *RegisterRemindMessage) Using() string {
	return types.OrmDataBaseApi
}

func (r *RegisterRemindMessage) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *RegisterRemindMessage) Add() (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Insert(r)

	return
}
