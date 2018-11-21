package models

import (
	"micro-loan/common/types"

	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const FIX_PAYMENT_CODE_ORDER_TABLENAME string = "fix_payment_code_order"

type FixPaymentCodeOrder struct {
	Id             int64  `orm:"pk;"`
	PaymentCode    string `orm:"column(payment_code)"`
	UserAccountId  int64  `orm:"column(user_account_id)"`
	OrderId        int64  `orm:"column(order_id)"`
	ExpectedAmount int64  `orm:"column(expected_amount)"`
	Ctime          int64
	Utime          int64
}

func (r *FixPaymentCodeOrder) TableName() string {
	return FIX_PAYMENT_CODE_ORDER_TABLENAME
}

func (r *FixPaymentCodeOrder) Using() string {
	return types.OrmDataBaseApi
}

func (r *FixPaymentCodeOrder) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func AddFixPaymentCodeOrder(fixPaymentCodeOrder *FixPaymentCodeOrder) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(fixPaymentCodeOrder.Using())
	fixPaymentCodeOrder.Ctime = tools.GetUnixMillis()
	fixPaymentCodeOrder.Utime = tools.GetUnixMillis()
	id, err = o.Insert(fixPaymentCodeOrder)
	if err != nil {
		logs.Error("model fix_payment_code_order insert failed.", err)
	}
	return
}

func OneFixPaymentCodeOrderById(id int64) (FixPaymentCodeOrder, error) {
	var obj = FixPaymentCodeOrder{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	err := o.QueryTable(obj.TableName()).Filter("id", id).One(&obj)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneFixPaymentCodeOrderById] sql error err:%v", err)
	}

	return obj, err
}

func OneFixPaymentCodeOrderByOrderId(orderId int64) (FixPaymentCodeOrder, error) {
	var obj = FixPaymentCodeOrder{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	err := o.QueryTable(obj.TableName()).Filter("orderId", orderId).One(&obj)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneFixPaymentCodeOrderById] sql error err:%v", err)
	}

	return obj, err
}
