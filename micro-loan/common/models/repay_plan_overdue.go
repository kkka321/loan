package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const REPAY_PLAN_OVERDUE_TABLENAME string = "repay_plan_overdue"

type RepayPlanOverdue struct {
	Id                  int64  `orm:"pk;"`
	OrderId             int64  `orm:"column(order_id)"`
	OverdueDate         string `orm:"column(overdue_date)"`
	Penalty             int64
	GracePeriodInterest int64
	Ctime               int64
	Utime               int64
}

// 当前模型对应的表名
func (r *RepayPlanOverdue) TableName() string {
	return REPAY_PLAN_OVERDUE_TABLENAME
}

// 当前模型的数据库
func (r *RepayPlanOverdue) Using() string {
	return types.OrmDataBaseApi
}

func (r *RepayPlanOverdue) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func AddRepayPlanOverdue(repayPlanOverdue *RepayPlanOverdue) (id int64, err error) {
	o := orm.NewOrm()
	obj := RepayPlanOverdue{}
	o.Using(obj.Using())
	id, err = o.Insert(repayPlanOverdue)
	if err != nil {
		logs.Error("model repay_plan_overdue AddRepayPlan failed.", err)
	}
	return
}

func UpdateRepayPlanOverdue(repayPlanOverdue *RepayPlanOverdue) (id int64, err error) {
	o := orm.NewOrm()
	obj := &RepayPlanOverdue{}
	o.Using(obj.Using())
	id, err = o.Update(repayPlanOverdue)
	if err != nil {
		logs.Error("model repay_plan_overdue UpdateRepayPlan failed.", err)
	}

	return
}

func GetRepayPlanOverdueByOrderId(orderId int64) (list []RepayPlanOverdue, err error) {
	r := RepayPlanOverdue{}
	o := orm.NewOrm()
	o.Using(r.UsingSlave())
	_, err = o.QueryTable(r.TableName()).Filter("order_id", orderId).OrderBy("-id").All(&list)
	if err != nil {
		logs.Error("[model.GetRepayPlanOverdueByOrderId] err:", err)
	}
	return
}
