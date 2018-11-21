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

const PAYMENT_TABLENAME string = "payment"

type Payment struct {
	Id            int64 `orm:"pk;"`
	OrderId       int64 `orm:"column(order_id)"`
	Amount        int64
	PayType       int    `orm:"column(pay_type)"`
	VaCompanyCode int    `orm:"column(va_company_code)"`
	UserAccountId string `orm:"column(user_account_id)"`
	UserBankCode  string `orm:"column(user_bank_code)"`
	VaCode        string
	Ctime         int64
	Utime         int64
}

// 当前模型对应的表名
func (r *Payment) TableName() string {
	return PAYMENT_TABLENAME
}

// 当前模型的数据库
func (r *Payment) Using() string {
	return types.OrmDataBaseApi
}

func (r *Payment) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *Payment) AddPayment(payment *Payment) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Insert(payment)
	if err != nil {
		logs.Error("model payment insert failed. err:%v payment:%#v", err, payment)
	}
	return
}

func (r *Payment) GetDisburseOrder(orderId int64) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	payment := &Payment{}
	err = o.QueryTable(r.TableName()).Filter("order_id", orderId).Filter("pay_type", types.PayTypeMoneyOut).One(payment)
	id = payment.Id
	return
}

func GetPaymentByOrderIdPayType(orderId int64, payType int) (payment Payment, err error) {
	o := orm.NewOrm()
	o.Using(payment.Using())
	err = o.QueryTable(payment.TableName()).Filter("order_id", orderId).Filter("pay_type", payType).One(&payment)
	return
}

func GetPaymentById(id int64) (Payment, error) {
    r := Payment{}

    o := orm.NewOrm()
    o.Using(r.UsingSlave())
    err := o.QueryTable(r.TableName()).Filter("id", id).One(&r)

    return r, err
}