package models

import (
	"github.com/astaxie/beego/orm"
	"micro-loan/common/types"
)

// PUSH_TASK_RECORD_TABLENAME 表名
const PUSH_TASK_RECORD_TABLENAME string = "push_task_record"

// PUSH_TASK_TABLENAME 描述数据表结构与结构体的映射
type PushTaskRecord struct {
	Id       int64 `orm:"pk;"`
	TaskId   int64
	ReadNum  int
	TotalNum int
	SuccNum  int
	PushDate int64
	ReadRate int
	Ctime    int64
}

// TableName 返回当前模型对应的表名
func (r *PushTaskRecord) TableName() string {
	return PUSH_TASK_RECORD_TABLENAME
}

// Using 返回当前模型的数据库
func (r *PushTaskRecord) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *PushTaskRecord) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *PushTaskRecord) Insert() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err := o.Insert(r)
	return err
}
