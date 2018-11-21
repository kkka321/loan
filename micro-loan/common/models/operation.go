package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// OPERATION_TABLENAME 表名
const OPERATION_TABLENAME string = "operation"

// Operation 描述数据表结构与结构体的映射
type Operation struct {
	Id    int64 `orm:"pk;"`
	Name  string
	Ctime int64
	Utime int64
}

// TableName 返回当前模型对应的表名
func (r *Operation) TableName() string {
	return OPERATION_TABLENAME
}

// Using 返回当前模型的数据库
func (r *Operation) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *Operation) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// Add
func (r *Operation) Add() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.Ctime = tools.GetUnixMillis()
	r.Utime = r.Ctime

	id, err := o.Insert(r)

	return id, err
}

// GetAll 返回所有 operation
func (r *Operation) GetAll() (data []Operation, err error) {
	o := orm.NewOrm()
	o.Using(r.UsingSlave())
	qs := o.QueryTable(r.TableName())

	_, err = qs.All(&data)
	return
}

// GetOneOperation 根据名称返回指定的 privilege
func GetOneOperation(name string) (Operation, error) {
	var p Operation
	o := orm.NewOrm()
	o.Using(p.UsingSlave())
	qs := o.QueryTable(p.TableName())

	err := qs.Filter("name", name).One(&p)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[GetOneOperation] sql error err:%v", err)
	}

	return p, err
}
