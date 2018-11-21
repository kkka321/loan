package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const NPWP_MOBI_TABLENAME string = "npwp_mobi"

type NpwpMobi struct {
	Id           int64 `orm:"pk;"`
	NpwpNo       string
	Status       int
	CustomerName string
	Ctime        int64
	Utime        int64
}

func (r *NpwpMobi) TableName() string {
	return NPWP_MOBI_TABLENAME
}

func (r *NpwpMobi) Using() string {
	return types.OrmDataBaseApi
}

func (r *NpwpMobi) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *NpwpMobi) Add() (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Insert(r)
	return
}

func (r *NpwpMobi) Update(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)
	return
}

func OneNpwpMobi(npwpNo string) (one NpwpMobi, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).
		Filter("npwp_no", npwpNo).
		OrderBy("-id").
		One(&one)
	return
}
