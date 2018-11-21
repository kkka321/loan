package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// PUSH_TASK_TABLENAME 表名
const PUSH_TASK_TABLENAME string = "push_task"

// PUSH_TASK_TABLENAME 描述数据表结构与结构体的映射
type PushTask struct {
	Id           int64 `orm:"pk;"`
	TaskName     string
	TaskStatus   types.SchemaStatus
	TaskDesc     string
	MessageType  int
	PushWay      int
	PushTarget   types.PushTarget
	PushListPath string
	Title        string
	Body         string
	Mark         string
	SkipTo       int
	Version      string
	Utime        int64
	Ctime        int64
}

// TableName 返回当前模型对应的表名
func (r *PushTask) TableName() string {
	return PUSH_TASK_TABLENAME
}

// Using 返回当前模型的数据库
func (r *PushTask) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *PushTask) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *PushTask) Insert() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)
	r.Id = id

	return err
}

func (r *PushTask) Update() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err := o.Update(r)

	return err
}

func GetPushTask(id int64) (PushTask, error) {
	obj := PushTask{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	err := o.QueryTable(obj.TableName()).
		Filter("id", id).One(&obj)

	return obj, err
}

func GetPushTaskByTarget(target types.PushTarget) ([]PushTask, error) {
	m := PushTask{}
	o := orm.NewOrm()
	o.Using(m.UsingSlave())

	list := make([]PushTask, 0)

	_, err := o.QueryTable(m.TableName()).
		Filter("push_target", target).
		All(&list)

	return list, err
}
