package models

// `product`
import (
	//"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	//"micro-loan/common/tools"
	"micro-loan/common/types"
	//"fmt"
)

const PRODUCT_TABLENAME string = "product"

type Product struct {
	Id                 int64 `orm:"pk;"`
	Name               string
	Ver                int
	CustomerVisible    types.CustomerVisibleTypeEunm
	Status             int
	Period             types.ProductPeriodEunm
	PeriodLoan         int
	DayInterestRate    int64 `orm:"column(day_interest_rate)"`
	DayFeeRate         int64 `orm:"column(day_fee_rate)"`
	DayGraceRate       int64
	DayPenaltyRate     int64
	ChargeInterestType types.ProductChargeInterestTypeEnum `orm:"column(charge_interest_type)"`
	ChargeFeeType      int                                 `orm:"column(charge_fee_type)"`
	MinAmount          int64                               `orm:"column(min_amount)"`
	MaxAmount          int64                               `orm:"column(max_amount)"`
	MinPeriod          int                                 `orm:"column(min_period)"`
	MaxPeriod          int                                 `orm:"column(max_period)"`
	RepayRemind        int                                 `orm:"column(repay_remind)"`
	OverdueRemind      int                                 `orm:"column(overdue_remind)"`
	RepayOrder         string                              `orm:"column(repay_order)"`
	RepayType          types.ProductRepayTypeEunm
	CeilWay            types.ProductCeilWayEunm
	CeilWayUnit        types.ProductCeilWayUnitEunm
	GracePeriod        int
	ProductType        int
	PenaltyCalcExpr    string
	Remarks            string
	Ctime              int64
	Utime              int64
}

type ProductReturnApp struct {
	Id                 int64                               `json:"id"`
	DayInterestRate    int64                               `json:"day_interest_rate"`
	DayFeeRate         int64                               `json:"day_fee_rate"`
	ChargeInterestType types.ProductChargeInterestTypeEnum `json:"charge_interest_type"`
	ChargeFeeType      int                                 `json:"charge_fee_type"`
	MinAmount          int64                               `json:"min_amount"`
	MaxAmount          int64                               `json:"max_amount"`
	MinPeriod          int                                 `json:"min_period"`
	MaxPeriod          int                                 `json:"max_period"`
	CustomerVisible    int                                 `json:"customer_visible"`
	CeilWay            types.ProductCeilWayEunm            `json:"ceil_way"`
	CeilWayUnit        types.ProductCeilWayUnitEunm        `json:"ceil_way_unit"`
}

// 当前模型对应的表名
func (r *Product) TableName() string {
	return PRODUCT_TABLENAME
}

// 当前模型的数据库
func (r *Product) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *Product) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *Product) AddProduct() (int64, error) {

	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

func (r *Product) UpdateProduct(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}

func GetProduct(id int64) (Product, error) {

	var product Product

	o := orm.NewOrm()
	o.Using(product.Using())
	err := o.QueryTable(product.TableName()).Filter("Id", id).One(&product)

	return product, err
}

//func (r *Product) GetProductApp() (dst ProductReturnApp) {
//	dst.Id = r.Id
//	dst.DayInterestRate = r.DayInterestRate
//	dst.DayFeeRate = r.DayFeeRate
//	dst.ChargeInterestType = r.ChargeInterestType
//	dst.ChargeFeeType = r.ChargeFeeType
//	dst.MinAmount = r.MinAmount
//	dst.MaxAmount = r.MaxAmount
//	dst.MinPeriod = r.MinPeriod
//	dst.MaxPeriod = r.MaxPeriod
//	dst.CeilWay = r.CeilWay
//	dst.CeilWayUnit = r.CeilWayUnit
//	dst.CustomerVisible = int(r.CustomerVisible)
//	return
//}
