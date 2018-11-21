package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// ROLE_TABLENAME 表名
const ROLE_TABLENAME string = "role"

// Role 描述数据表结构与结构体的映射
type Role struct {
	Id     int64 `orm:"pk;"`
	Name   string
	Type   types.RoleTypeEnum
	Pid    int64
	Status int
	Ctime  int64
	Utime  int64
}

// TableName 返回当前模型对应的表名
func (r *Role) TableName() string {
	return ROLE_TABLENAME
}

// Using 返回当前模型的数据库
func (r *Role) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *Role) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// GetOneRole 获取指定ID的角色信息
func GetOneRole(id int64) (data Role, err error) {
	obj := Role{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	qs := o.QueryTable(obj.TableName())

	err = qs.Filter("id", id).One(&data)

	return
}
