package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// DATA_PRIVILEGE_TABLENAME 表名
const DATA_PRIVILEGE_TABLENAME string = "data_privilege"

// DataPrivilege 动态资源,描述数据表结构与结构体的映射
type DataPrivilege struct {
	Id         int64 `orm:"pk"`
	Type       types.DataPrivilegeTypeEnum
	DataID     int64 `orm:"column(data_id)"`
	GrantType  types.DataGrantTypeEnum
	GrantTo    int64
	Status     int
	IsDeleted  int
	Ctime      int64
	RevokeTime int64
	Utime      int64
}

// TableName 返回当前模型对应的表名
func (r *DataPrivilege) TableName() string {
	return DATA_PRIVILEGE_TABLENAME
}

// Using 返回当前模型的数据库
func (r *DataPrivilege) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *DataPrivilege) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// Insert 插入新记录
func (r *DataPrivilege) Insert() (int64, error) {
	r.Ctime = tools.GetUnixMillis()
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

// select * from data_privilege where grant_to in
