package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// COUPON_TASK_TABLENAME 表名
const COUPON_TASK_TABLENAME string = "coupon_task"

// COUPON_TASK_TABLENAME 描述数据表结构与结构体的映射
type CouponTask struct {
	Id             int64 `orm:"pk;"`
	TaskName       string
	TaskStatus     types.SchemaStatus
	TaskDesc       string
	CouponId       int64
	CouponTarget   types.CouponTarget
	CouponListPath string
	Utime          int64
	Ctime          int64
}

// TableName 返回当前模型对应的表名
func (r *CouponTask) TableName() string {
	return COUPON_TASK_TABLENAME
}

// Using 返回当前模型的数据库
func (r *CouponTask) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *CouponTask) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *CouponTask) Insert() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)
	r.Id = id

	return err
}

func (r *CouponTask) Update() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err := o.Update(r)

	return err
}

func GetCouponTask(id int64) (CouponTask, error) {
	obj := CouponTask{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	err := o.QueryTable(obj.TableName()).
		Filter("id", id).One(&obj)

	return obj, err
}

func GetCouponTaskByTarget(target types.CouponTarget) ([]CouponTask, error) {
	m := CouponTask{}
	o := orm.NewOrm()
	o.Using(m.UsingSlave())

	list := make([]CouponTask, 0)

	_, err := o.QueryTable(m.TableName()).
		Filter("coupon_target", target).
		All(&list)

	return list, err
}
