package models

import (
	"micro-loan/common/types"

	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const FIX_PAYMENT_CODE_TABLENAME string = "fix_payment_code"

type FixPaymentCode struct {
	Id             string `orm:"pk;"`
	UserAccountId  int64  `orm:"column(user_account_id)"`
	OrderId        int64  `orm:"column(order_id)"`
	PaymentCode    string `orm:"column(payment_code)"`
	ExpectedAmount int64  `orm:"column(expected_amount)"`
	ExpirationDate int64  `orm:"column(expiration_date)"`
	ResponseJson   string `orm:"column(response_json)"`
	Ctime          int64
	Utime          int64
}

func (r *FixPaymentCode) TableName() string {
	return FIX_PAYMENT_CODE_TABLENAME
}

func (r *FixPaymentCode) Using() string {
	return types.OrmDataBaseApi
}

func (r *FixPaymentCode) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func AddFixPaymentCode(fixPaymentCode *FixPaymentCode) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(fixPaymentCode.Using())
	fixPaymentCode.Ctime = tools.GetUnixMillis()
	fixPaymentCode.Utime = tools.GetUnixMillis()
	id, err = o.Insert(fixPaymentCode)
	if err != nil {
		logs.Error("model fix_payment_code insert failed.", err)
	}
	return
}

func OneFixPaymentCodeById(id string) (FixPaymentCode, error) {
	var obj = FixPaymentCode{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	err := o.QueryTable(obj.TableName()).Filter("id", id).One(&obj)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneFixPaymentCodeById] sql error err:%v", err)
	}

	return obj, err
}

func OneFixPaymentCodeByPaymentCode(paymentCode string) (FixPaymentCode, error) {
	var obj = FixPaymentCode{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	err := o.QueryTable(obj.TableName()).Filter("payment_code", paymentCode).One(&obj)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneFixPaymentCodeById] sql error err:%v", err)
	}

	return obj, err
}

func OneFixPaymentCodeByUserAccountId(userAccountId int64) (FixPaymentCode, error) {
	var obj = FixPaymentCode{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	err := o.QueryTable(obj.TableName()).Filter("user_account_id", userAccountId).One(&obj)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneFixPaymentCodeByUserAccountId] sql error err:%v", err)
	}

	return obj, err
}

func UpdateFixPaymentCode(obj *FixPaymentCode, cols []string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(obj.Using())
	id, err = o.Update(obj, cols...)
	if err != nil {
		logs.Error("[UpdateFixPaymentCode] sql error err:%v", err)
	}

	return
}
