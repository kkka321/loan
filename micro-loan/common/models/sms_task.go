package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

// SMS_TASK_TABLENAME 表名
const SMS_TASK_TABLENAME string = "sms_task"

type SmsTask struct {
	Id          int64 `orm:"pk;"`
	TaskName    string
	TaskStatus  types.SchemaStatus
	TaskDesc    string
	SmsTarget   types.SmsTarget
	Sender      types.SmsServiceID
	SmsListPath string
	Body        string
	Utime       int64
	Ctime       int64
}

// TableName 返回当前模型对应的表名
func (r *SmsTask) TableName() string {
	return SMS_TASK_TABLENAME
}

// Using 返回当前模型的数据库
func (r *SmsTask) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *SmsTask) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *SmsTask) Insert() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)
	r.Id = id

	return err
}

func (r *SmsTask) Update() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err := o.Update(r)

	return err
}

func GetSmsTask(id int64) (SmsTask, error) {
	obj := SmsTask{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	err := o.QueryTable(obj.TableName()).
		Filter("id", id).One(&obj)

	return obj, err
}

func GetSmsTaskByTarget(target types.SmsTarget) ([]SmsTask, error) {
	m := SmsTask{}
	o := orm.NewOrm()
	o.Using(m.UsingSlave())

	list := make([]SmsTask, 0)

	_, err := o.QueryTable(m.TableName()).
		Filter("sms_target", target).
		All(&list)

	return list, err
}
