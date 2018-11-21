package models

// `product`
import (
	//"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	//"micro-loan/common/tools"
	"micro-loan/common/types"
	//"fmt"
	"github.com/astaxie/beego/logs"
)

const PAYMENT_VOUCHER_TABLENAME string = "payment_voucher"

type PaymentVoucher struct {
	Id         int64  `orm:"pk;"`
	OpUid      int64  `orm:"column(op_uid)"`
	AccountId  int64  `orm:"column(account_id)"`
	OrderId    int64  `orm:"column(order_id)"`
	ResourceId int64  `orm:"column(resource_id)"`
	ReimbMeans string `orm:"column(reimb_means)"`
	Status     int64  `orm:"column(status)"`
	Comment    string `orm:"column(comment)"`
	Ctime      int64  `orm:"column(ctime)"`
	Utime      int64  `orm:"column(utime)"`
}

// 当前模型对应的表名
func (r *PaymentVoucher) TableName() string {
	return PAYMENT_VOUCHER_TABLENAME
}

// 当前模型的数据库
func (r *PaymentVoucher) Using() string {
	return types.OrmDataBaseApi
}

func (r *PaymentVoucher) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *PaymentVoucher) AddPayment(payment *Payment) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Insert(payment)
	if err != nil {
		logs.Error("model payment insert failed. err:%v payment:%#v", err, payment)
	}
	return
}
func (r *PaymentVoucher) Updates(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}

func GetMultiPaymentByOrderId(accountId int64) (data []PaymentVoucher, err error) {
	o := orm.NewOrm()

	obj := PaymentVoucher{}

	o.Using(obj.Using())

	_, err = o.QueryTable(obj.TableName()).Filter("order_id", accountId).
		OrderBy("-ctime").
		All(&data)

	return
}
