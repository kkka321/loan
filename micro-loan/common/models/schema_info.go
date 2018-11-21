package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

// SCHEMA_INFO_TABLENAME 表名
const SCHEMA_INFO_TABLENAME string = "schema_info"

// SCHEMA_INFO_TABLENAME 描述数据表结构与结构体的映射
type SchemaInfo struct {
	Id           int64 `orm:"pk;"`
	SchemaMode   types.SchemaMode
	SchemaStatus types.SchemaStatus
	SchemaTime   string
	StartDate    int64
	EndDate      int64
	FuncName     string
	TaskId       int64
	Utime        int64
	Ctime        int64
}

// TableName 返回当前模型对应的表名
func (r *SchemaInfo) TableName() string {
	return SCHEMA_INFO_TABLENAME
}

// Using 返回当前模型的数据库
func (r *SchemaInfo) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *SchemaInfo) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *SchemaInfo) Insert() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)
	r.Id = id

	return err
}

func (r *SchemaInfo) Update() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err := o.Update(r)

	return err
}

func GetSchemaInfo(id int64) (SchemaInfo, error) {
	obj := SchemaInfo{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	err := o.QueryTable(obj.TableName()).
		Filter("id", id).One(&obj)

	return obj, err
}

func LoadSchemaInfo() ([]SchemaInfo, error) {
	m := SchemaInfo{}
	o := orm.NewOrm()
	o.Using(m.UsingSlave())

	list := make([]SchemaInfo, 0)

	_, err := o.QueryTable(m.TableName()).
		Filter("schema_status", types.SchemaStatusOn).
		Exclude("schema_mode", types.SchemaModeBusiness).
		All(&list)

	return list, err
}
