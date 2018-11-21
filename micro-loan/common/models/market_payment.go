package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const MARKET_PAYMENT_TABLENAME string = "market_payment"

type MarketPayment struct {
	Id            int64  `orm:"pk;"`
	UserAccountId int64  `orm:"column(user_account_id)"`
	OrderId       int64  `orm:"column(order_id)"`
	PaymentCode   string `orm:"column(payment_code)"`
	Amount        int64
	Status        string
	ExpiryDate    int64 `orm:"column(expiry_date)"`
	PaidTime      int64 `orm:"column(paid_time)"`
	Response      string
	CallbackJson  string `orm:"column(callback_json)"`
	Ctime         int64
	Utime         int64
}

type InvoiceResp struct {
	Id                        string `json:"id"`
	ExternalId                string `json:"external_id"`
	UserId                    string `json:"user_id"`
	Status                    string `json:"status"`
	MerchantName              string `json:"merchant_name"`
	MerchantProfilePictureUrl string `json:"merchant_profile_picture_url"`
	Amount                    int64  `json:"amount"`
	PayerEmail                string `json:"payer_email"`
	Description               string `json:"description"`
	ExpiryDate                string `json:"expiry_date"`
	InvoiceUrl                string `json:"invoice_url"`
	AvailableBanks            []struct {
		BankCode          string `json:"bank_code"`
		CollectionType    string `json:"collection_type"`
		BankAccountNumber string `json:"bank_account_number"`
		TransferAmount    int64  `json:"transfer_amount"`
		BankBranch        string `json:"bank_branch"`
		AccountHolderName string `json:"account_holder_name"`
		IdentityAmount    int64  `json:"identity_amount"`
	} `json:"available_banks"`

	AvailableRetailOutlets []struct {
		RetailOutletName string `json:"retail_outlet_name"`
		PaymentCode      string `json:"payment_code"`
		TransferAmount   int64  `json:"transfer_amount"`
	} `json:"available_retail_outlets"`

	ShouldExcludeCreditCard bool   `json:"should_exclude_credit_card"`
	ShouldSendEmail         bool   `json:"should_send_email"`
	Created                 string `json:"created"`
	Updated                 string `json:"updated"`
	ErrorCode               string `json:"error_code"`
	Message                 string `json:"message"`
}

// 当前模型对应的表名
func (r *MarketPayment) TableName() string {
	return MARKET_PAYMENT_TABLENAME
}

// 当前模型的数据库
func (r *MarketPayment) Using() string {
	return types.OrmDataBaseApi
}

// 当前模型的数据库
func (r *MarketPayment) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func AddMarketPayment(marketPayment *MarketPayment) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(marketPayment.Using())
	id, err = o.Insert(marketPayment)
	if err != nil {
		logs.Error("model marketPayment insert failed.", err)
	}
	return
}

func GetMarketPaymentByOrderId(orderId int64) (MarketPayment, error) {
	o := orm.NewOrm()
	marketPayment := MarketPayment{}
	o.Using(marketPayment.Using())
	err := o.QueryTable(marketPayment.TableName()).Filter("order_id", orderId).Filter("status", "PENDING").OrderBy("-id").Limit(1).One(&marketPayment)

	return marketPayment, err
}

func GetMarketPaymentsByOrderId(orderId int64) (marketPayments []MarketPayment, err error) {
	o := orm.NewOrm()
	marketPayment := MarketPayment{}
	o.Using(marketPayment.Using())
	_, err = o.QueryTable(marketPayment.TableName()).Filter("order_id", orderId).Filter("status", "PENDING").OrderBy("-id").All(&marketPayments)
	return
}

func GetMarketPaymentByPaymentCode(paymentCode string) (MarketPayment, error) {
	o := orm.NewOrm()
	marketPayment := MarketPayment{}
	o.Using(marketPayment.Using())
	err := o.QueryTable(marketPayment.TableName()).Filter("payment_code", paymentCode).OrderBy("-id").Limit(1).One(&marketPayment)

	return marketPayment, err
}

func UpdateMarketPayment(marketPayment *MarketPayment) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(marketPayment.Using())
	id, err = o.Update(marketPayment)
	if err != nil {
		logs.Error("model marketPayment update failed.", err)
	}

	return
}

func UpdateMarketPaymentStatus(marketPayment *MarketPayment) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(marketPayment.Using())
	id, err = o.Update(marketPayment, "status")
	if err != nil {
		logs.Error("model marketPayment update failed.", err)
	}

	return
}
