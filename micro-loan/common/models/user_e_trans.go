package models

// `product`
import (
	//"github.com/astaxie/beego/logs"
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
	//"fmt"
	"micro-loan/common/lib/device"
	"micro-loan/common/tools"
)

const E_TRANS_TABLENAME string = "user_e_trans"

type User_E_Trans struct {
	Id                  int64 `orm:"pk;"`
	UserAccountId       int64 `orm:"column(user_account_id)"`
	VaCompanyCode       int   `orm:"column(va_company_code)"`
	OrderId             int64 `orm:"column(order_id)"`
	PaymentId           int64
	Total               int64
	Amount              int64
	Interest            int64
	PreInterest         int64
	GracePeriodInterest int64
	ServiceFee          int64 `orm:"column(service_fee)"`
	Penalty             int64
	Balance             int64
	PayType             int
	CallbackJson        string
	IsFrozen            int
	Ctime               int64
	Utime               int64
}

// 当前模型对应的表名
func (r *User_E_Trans) TableName() string {
	return E_TRANS_TABLENAME
}

// 当前模型的数据库
func (r *User_E_Trans) Using() string {
	return types.OrmDataBaseApi
}
func (r *User_E_Trans) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *User_E_Trans) AddEtrans(eTrans *User_E_Trans) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Insert(eTrans)
	if err != nil {
		logs.Error("model user_e_trans insert failed.:%v eTrans:%#v", err, eTrans)
	}
	return
}

func (r *User_E_Trans) Update() (err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err = o.Update(r)
	if err != nil {
		logs.Error("[Update]model user_e_trans insert failed err:%v eTrans:%#v", err, r)
	}
	return
}

func GetETransByOrderId(orderId int64) []User_E_Trans {
	o := orm.NewOrm()
	eTrans := User_E_Trans{}
	o.Using(eTrans.UsingSlave())

	var data []User_E_Trans
	_, _ = o.QueryTable(eTrans.TableName()).
		Filter("order_id", orderId).
		All(&data)

	return data
}

func GetAllETransByCompany(orderId int64, company int) []User_E_Trans {
	o := orm.NewOrm()
	eTrans := User_E_Trans{}
	o.Using(eTrans.UsingSlave())

	var data []User_E_Trans
	_, _ = o.QueryTable(eTrans.TableName()).
		Filter("order_id", orderId).
		Filter("va_company_code", company).
		OrderBy("-id").
		All(&data)

	return data
}

func GetOutETransByOrderId(orderId int64) []User_E_Trans {
	o := orm.NewOrm()
	eTrans := User_E_Trans{}
	o.Using(eTrans.UsingSlave())

	var data []User_E_Trans
	_, _ = o.QueryTable(eTrans.TableName()).
		Filter("order_id", orderId).
		Exclude("pay_type", types.PayTypeMoneyIn).
		Exclude("pay_type", types.PayTypeRefundIn).
		All(&data)

	return data
}

// GetLastInPayETransByOrderID 获取最后一次入账
func GetLastInPayETransByOrderID(orderID int64) (etransModel User_E_Trans) {
	o := orm.NewOrm()
	eTrans := User_E_Trans{}
	o.Using(eTrans.UsingSlave())
	o.QueryTable(eTrans.TableName()).
		Filter("order_id", orderID).
		Filter("pay_type", types.PayTypeMoneyIn).
		Filter("va_company_code__lt", 1000).
		OrderBy("-id").
		One(&etransModel)
	return
}

// GetLastOutPayETransByOrderID 获取最后一次出账
func GetLastOutPayETransByOrderID(orderID int64) (etransModel User_E_Trans) {
	o := orm.NewOrm()
	eTrans := User_E_Trans{}
	o.Using(eTrans.UsingSlave())
	o.QueryTable(eTrans.TableName()).
		Filter("order_id", orderID).
		Filter("pay_type", types.PayTypeMoneyOut).
		OrderBy("-id").
		One(&etransModel)
	return
}

func AddReductionPenalty(orderId int64, userAccountId int64, reductionAmount int64, reduction_penalty int64, reduction_grace_period_interest int64) error {
	o := orm.NewOrm()
	eTrans := User_E_Trans{}
	o.Using(eTrans.Using())

	eTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
	eTrans.UserAccountId = userAccountId
	eTrans.OrderId = orderId
	eTrans.Total = reduction_penalty + reduction_grace_period_interest + reductionAmount
	eTrans.VaCompanyCode = types.MobiReductionPenalty
	eTrans.PayType = types.PayTypeMoneyIn
	eTrans.Ctime = tools.GetUnixMillis()
	eTrans.Utime = tools.GetUnixMillis()
	_, err := eTrans.AddEtrans(&eTrans)
	if err != nil {
		return err
	}

	//入账

	eTrans = User_E_Trans{}
	eTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
	eTrans.UserAccountId = userAccountId
	eTrans.OrderId = orderId
	eTrans.Amount = eTrans.Amount + reductionAmount
	eTrans.Penalty = eTrans.Penalty + reduction_penalty
	eTrans.GracePeriodInterest = eTrans.GracePeriodInterest + reduction_grace_period_interest
	eTrans.VaCompanyCode = types.MobiReductionPenalty
	eTrans.PayType = types.PayTypeMoneyOut
	eTrans.Ctime = tools.GetUnixMillis()
	eTrans.Utime = tools.GetUnixMillis()
	_, err = eTrans.AddEtrans(&eTrans)

	//出账
	return err
}

// GetOrdersRepayPrincipalAndInterest 根据该订单ID列表获取，这些订单在指定时间范围内的回款本金和回款息费
func GetOrdersRepayPrincipalAndInterest(orderIDStrings []string, startTime, endTime int64) (repayPrincipal, repayInterest int64, err error) {
	o := orm.NewOrm()
	eTrans := User_E_Trans{}
	o.Using(eTrans.UsingSlave())

	sql := "SELECT sum(`amount`), sum(`grace_period_interest`)+sum(`penalty`) as total_interest"
	sql += fmt.Sprintf(" FROM `%s` WHERE `order_id` in(%s) AND pay_type=%d AND ctime>=%d AND ctime<%d",
		eTrans.TableName(), strings.Join(orderIDStrings, ","), types.PayTypeMoneyOut, startTime, endTime)
	r := o.Raw(sql)
	err = r.QueryRow(&repayPrincipal, &repayInterest)
	if err != nil {
		logs.Error("[GetOrdersRepayPrincipalAndInterest] query should be ok, but err:", err)
	}
	return
}

func GetEtranByOrderIdPayTypeVaCompanyCode(orderID int64, payType int, vaCompanyCode int) (etransModel User_E_Trans, err error) {
	o := orm.NewOrm()
	eTrans := User_E_Trans{}
	o.Using(eTrans.UsingSlave())
	err = o.QueryTable(eTrans.TableName()).
		Filter("order_id", orderID).
		Filter("pay_type", payType).
		Filter("va_company_code", vaCompanyCode).
		One(&etransModel)
	return
}
