package models

import (
	"fmt"
	"micro-loan/common/types"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const REPAY_PLAN_TABLENAME string = "repay_plan"

type RepayPlan struct {
	Id                         int64 `orm:"pk;"`
	OrderId                    int64 `orm:"column(order_id)"`
	Amount                     int64
	AmountPayed                int64 `orm:"column(amount_payed)"`
	AmountReduced              int64 `orm:"column(amount_reduced)"`
	PreInterest                int64
	PreInterestPayed           int64
	GracePeriodInterest        int64
	GracePeriodInterestPayed   int64
	GracePeriodInterestReduced int64 `orm:"column(grace_period_interest_reduced)"`
	Interest                   int64
	InterestPayed              int64
	ServiceFee                 int64 `orm:"column(service_fee)"`
	ServiceFeePayed            int64 `orm:"column(service_fee_payed)"`
	Penalty                    int64
	PenaltyPayed               int64 `orm:"column(penalty_payed)"`
	PenaltyReduced             int64 `orm:"column(penalty_reduced)"`
	RepayDate                  int64 `orm:"column(repay_date)"`
	Ctime                      int64
	Utime                      int64
}

type RepayPlanHistory struct {
	Id         int
	Plan       RepayPlan
	PayOutTime int64 // 出帐时间 代表系统产生费用
	PayInTime  int64 // 入账时间 代表用户还钱或减免操作
	// RecordType int
}

// 当前模型对应的表名
func (r *RepayPlan) TableName() string {
	return REPAY_PLAN_TABLENAME
}

// 当前模型的数据库
func (r *RepayPlan) Using() string {
	return types.OrmDataBaseApi
}

func (r *RepayPlan) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func AddRepayPlan(repayPlan *RepayPlan) (id int64, err error) {
	o := orm.NewOrm()
	obj := RepayPlan{}
	o.Using(obj.Using())
	id, err = o.Insert(repayPlan)
	if err != nil {
		logs.Error("model repay_plan AddRepayPlan failed.", err)
	}
	return
}

func UpdateRepayPlan(repayPlan *RepayPlan) (id int64, err error) {
	o := orm.NewOrm()
	obj := &RepayPlan{}
	o.Using(obj.Using())
	id, err = o.Update(repayPlan)
	if err != nil {
		logs.Error("model repay_plan UpdateRepayPlan failed.", err)
	}

	return
}

func GetLastRepayPlanByOrderid(orderId int64) (RepayPlan, error) {
	o := orm.NewOrm()
	obj := RepayPlan{}
	o.Using(obj.Using())
	err := o.QueryTable(obj.TableName()).Filter("order_id", orderId).OrderBy("-id").Limit(1).One(&obj)
	return obj, err
}

// 添加
func (r *RepayPlan) AddRepayPlan(repayPlan *RepayPlan) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Insert(repayPlan)
	if err != nil {
		logs.Error("model repay_plan AddRepayPlan failed.", err)
	}
	return
}

// 改
func (r *RepayPlan) UpdateRepayPlan(repayPlan *RepayPlan) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Update(repayPlan)
	if err != nil {
		logs.Error("model repay_plan UpdateRepayPlan failed.", err)
	}

	return
}

// 查询
func (r *RepayPlan) GetLastRepayPlanByOrderid(orderId int64) (RepayPlan, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	err := o.QueryTable(r.TableName()).Filter("order_id", orderId).OrderBy("-id").Limit(1).One(r)
	return *r, err
}

// GetOrdersLeftUnpaidPrincipal 根据该订单ID列表获取，实时获取剩余未还本金
func GetOrdersLeftUnpaidPrincipal(orderIDStrings []string) (unpaidPrincipal int64, err error) {
	o := orm.NewOrm()
	obj := RepayPlan{}
	o.Using(obj.UsingSlave())

	sql := "SELECT sum(`amount`-`amount_payed`-`amount_reduced`) as unpaid_principal"
	sql += fmt.Sprintf(" FROM `%s` WHERE `order_id` in(%s)",
		obj.TableName(), strings.Join(orderIDStrings, ","))

	r := o.Raw(sql)
	err = r.QueryRow(&unpaidPrincipal)
	if err != nil {
		logs.Error("[GetOrdersLeftUnpaidPrincipal] query should be ok, but err:", err)
	}
	return
}

// GetOrdersLeftUnpaidInterest 根据该订单ID列表获取，实时获取剩余未还息费（未还宽限期+未还罚息）
func GetOrdersLeftUnpaidInterest(orderIDStrings []string) (unpaidInterest int64, err error) {
	o := orm.NewOrm()
	obj := RepayPlan{}
	o.Using(obj.UsingSlave())

	sql := "SELECT sum(`grace_period_interest`-`grace_period_interest_payed`-`grace_period_interest_reduced`+`penalty`-`penalty_payed`-`penalty_reduced`) as unpaid_Interest"
	sql += fmt.Sprintf(" FROM `%s` WHERE `order_id` in(%s)",
		obj.TableName(), strings.Join(orderIDStrings, ","))

	r := o.Raw(sql)
	err = r.QueryRow(&unpaidInterest)
	if err != nil {
		logs.Error("[GetOrdersLeftUnpaidInterest] query should be ok, but err:", err)
	}
	return
}
