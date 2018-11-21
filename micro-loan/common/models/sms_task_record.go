package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// SMS_TASK_RECORD_TABLENAME 表名
const SMS_TASK_RECORD_TABLENAME string = "sms_task_record"

type SmsTaskRecord struct {
	Id       int64 `orm:"pk;"`
	TaskId   int64
	TotalNum int
	SuccNum  int
	SendDate int64
	Ctime    int64
}

// TableName 返回当前模型对应的表名
func (r *SmsTaskRecord) TableName() string {
	return SMS_TASK_RECORD_TABLENAME
}

// Using 返回当前模型的数据库
func (r *SmsTaskRecord) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *SmsTaskRecord) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *SmsTaskRecord) Insert() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err := o.Insert(r)
	return err
}
