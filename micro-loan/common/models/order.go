package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	//"micro-loan/common/tools"

	"micro-loan/common/types"
)

const ORDER_TABLENAME string = "orders"

type Order struct {
	Id                          int64  `orm:"pk;"`
	UserAccountId               int64  `orm:"column(user_account_id)"`
	EAccountNumber              string `orm:"column(e_account_number)"`
	ProductId                   int64  `orm:"column(product_id)"`
	ProductIdOrg                int64
	Amount                      int64
	Loan                        int64
	LoanOrg                     int64
	Period                      int
	PeriodOrg                   int
	CheckStatus                 types.LoanStatus `orm:"column(check_status)"`
	IsTemporary                 int
	IsOverdue                   int
	IsDeadDebt                  int
	IsReloan                    int
	IsUpHoldPhoto               int
	ApplyTime                   int64
	CheckTime                   int64
	RepayTime                   int64
	LoanTime                    int64
	FinishTime                  int64
	PenaltyUtime                int64
	PhoneVerifyTime             int64
	RejectReason                types.RejectReasonEnum
	RiskCtlStatus               types.RiskCtlEnum
	RiskCtlFinishTime           int64
	RiskCtlRegular              string
	RandomValue                 int
	FixedRandom                 int
	RandomMark                  int
	OpUid                       int64
	RollTimes                   int
	PreOrder                    int64
	MinRepayAmount              int64
	LivingbestReloanhandSimilar string
	AfterBlackSimilar           string
	Ctime                       int64
	Utime                       int64
	IsDeleted                   int
}

type OrderLoanBusiness struct {
	OpUid               int64
	ApplyTime           int64
	ApplyOperator       string
	CheckTime           int64
	CheckOperator       string
	RiskCtlFinishTime   int64
	PhoneVerifyTime     int64
	PhoneVerifyOperator string
	LoanTime            int64
	LoanStatus          string
	PayTime             int64
	PayOperator         string
	PayStatus           string
}

// 当前模型对应的表名
func (r *Order) TableName() string {
	return ORDER_TABLENAME
}

// 当前模型的数据库
func (r *Order) Using() string {
	return types.OrmDataBaseApi
}

// 当前模型的数据库
func (r *Order) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *Order) AddOrder(order *Order) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Insert(order)
	if err != nil {
		logs.Error("model order insert failed.", err)
	}
	return
}

func (r *Order) Delete() (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Delete(r)
	return
}

func GetOrder(orderId int64) (Order, error) {
	o := orm.NewOrm()
	order := Order{}
	o.Using(order.Using())
	err := o.QueryTable(order.TableName()).Filter("id", orderId).One(&order)

	return order, err
}

func GetRollOrder(orderId int64) (Order, error) {
	o := orm.NewOrm()
	order := Order{}
	o.Using(order.Using())
	err := o.QueryTable(order.TableName()).Filter("pre_order", orderId).OrderBy("-id").Limit(1).One(&order)

	return order, err
}

func UpdateOrder(order *Order) (num int64, err error) {
	o := orm.NewOrm()
	odr := Order{}
	o.Using(odr.Using())
	num, err = o.Update(order)
	if err != nil {
		logs.Error("model order update failed.", err)
	}

	return
}

func GetClearedOrderNumByAccountId(accountId int64) (num int64, err error) {
	orderM := Order{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	num, err = o.QueryTable(orderM.TableName()).Filter("user_account_id", accountId).
		Filter("check_status", types.LoanStatusAlreadyCleared).Count()

	return
}

func GetOrderByAccountId(accountId int64) (Order, error) {
	order := Order{}
	o := orm.NewOrm()
	o.Using(order.Using())

	err := o.QueryTable(order.TableName()).Filter("user_account_id", accountId).Limit(1).One(&order)

	return order, err
}

/**

这个方法一定慎用，特别是查询订单的时候
基本都用 service.AccountLastLoanOrder替代了！！！！
*/

func GetUserLastOrder(userAccountId int64) (*Order, error) {
	r := Order{}
	o := orm.NewOrm()
	o.Using(r.Using())
	order := &Order{}
	err := o.QueryTable(r.TableName()).Filter("user_account_id", userAccountId).OrderBy("-id").One(order)
	return order, err
}

//返回查询器
func (r *Order) GetQuerySeter() (query orm.QuerySeter) {
	o := orm.NewOrm()
	o.Using(r.Using())
	return o.QueryTable(r.TableName())

}

func (r *Order) UpdateOrder(order *Order) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Update(order)
	if err != nil {
		logs.Error("model order update failed.", err)
	}

	return
}

//! 这个方法有点副作用,请调用者注意
func (r *Order) Update(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}

// GetUserOrderNum 根据用户ID获取用户申请过的总订单数
func GetUserOrderNum(accountID int64) int64 {
	r := Order{}
	o := orm.NewOrm()
	o.Using(r.UsingSlave())
	num, err := o.QueryTable(r.TableName()).Filter("user_account_id", accountID).Count()
	if err != nil {
		logs.Error("[GetUserOrderNum] err:", err)
	}
	return num
}

// GetUserApplySuccOrderNum 根据用户ID获取用户申请成功的订单数
func GetUserApplySuccOrderNum(accountID int64) int64 {
	r := Order{}
	o := orm.NewOrm()
	o.Using(r.UsingSlave())
	num, err := o.QueryTable(r.TableName()).Filter("user_account_id", accountID).
		Filter("check_status__in", types.SuccLoanStatusSlice()).Count()
	if err != nil {
		logs.Error("[GetUserApplySuccOrderNum] err:", err)
	}
	return num
}
