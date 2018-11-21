package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// PRIVILEGE_TABLENAME 表名
const PRIVILEGE_TABLENAME string = "privilege"

// Privilege 描述数据表结构与结构体的映射
type Privilege struct {
	Id      int64 `orm:"pk;"`
	Name    string
	GroupID int64 `orm:"column(group_id);"`
	Ctime   int64
	Utime   int64
}

// TableName 返回当前模型对应的表名
func (r *Privilege) TableName() string {
	return PRIVILEGE_TABLENAME
}

// Using 返回当前模型的数据库
func (r *Privilege) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *Privilege) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// Add 添加新的权限
func (r *Privilege) Add() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.Ctime = tools.GetUnixMillis()
	r.Utime = r.Ctime

	id, err := o.Insert(r)

	return id, err
}

// GetAll 返回所有 privielge
func (r *Privilege) GetAll() (data []Privilege, err error) {
	o := orm.NewOrm()
	o.Using(r.UsingSlave())
	qs := o.QueryTable(r.TableName())

	_, err = qs.All(&data)
	return
}

// GetOnePrivilege 根据名称返回指定的 privilege
func GetOnePrivilege(name string) (Privilege, error) {
	var p Privilege
	o := orm.NewOrm()
	o.Using(p.UsingSlave())
	qs := o.QueryTable(p.TableName())

	err := qs.Filter("name", name).One(&p)

	return p, err
}
