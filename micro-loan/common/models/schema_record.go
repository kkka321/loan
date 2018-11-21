package models

import (
	"github.com/astaxie/beego/orm"
	"micro-loan/common/types"
)

// SCHEMA_RECORD_TABLENAME 表名
const SCHEMA_RECORD_TABLENAME string = "schema_record"

// SCHEMA_RECORD_TABLENAME 描述数据表结构与结构体的映射
type SchemaRecord struct {
	Id       int64 `orm:"pk;"`
	SchemaId int64
	Result   string
	Ctime    int64
}

// TableName 返回当前模型对应的表名
func (r *SchemaRecord) TableName() string {
	return SCHEMA_RECORD_TABLENAME
}

// Using 返回当前模型的数据库
func (r *SchemaRecord) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *SchemaRecord) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *SchemaRecord) Insert() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err := o.Insert(r)
	return err
}
