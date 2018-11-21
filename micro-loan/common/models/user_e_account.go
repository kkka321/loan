package models

// `product`
import (
	//"github.com/astaxie/beego/logs"
	"encoding/json"

	"github.com/astaxie/beego/orm"

	//"micro-loan/common/tools"

	"micro-loan/common/types"
	//"fmt"
	"github.com/astaxie/beego/logs"
)

const E_ACCOUNT_TABLENAME string = "user_e_account"

type User_E_Account struct {
	Id             int64  `orm:"pk;"`
	UserAccountId  int64  `orm:"column(user_account_id)"`
	EAccountNumber string `orm:"column(e_account_number)"`
	BankCode       string `orm:"column(bank_code)"`
	RepayBankCode  string `orm:"column(repay_bank_code)"`
	VaCompanyCode  int    `orm:"column(va_company_code)"`
	IsClosed       int    `orm:"column(is_closed)"`
	Status         string
	CallbackJson   string
	Ctime          int64
	Utime          int64
}

type XenditCallBack struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	BankCode       string `json:"bank_code"`
	ExternalID     string `json:"external_id"`
	OwnerID        string `json:"owner_id"`
	MerchantCode   string `json:"merchant_code"`
	AccountNumber  string `json:"account_number"`
	IsSingleUse    bool   `json:"is_single_use"`
	Status         string `json:"status"`
	ExpirationDate string `json:"expiration_date"`
	IsClosed       bool   `json:"is_closed"`
	Updated        string `json:"updated"`
	Created        string `json:"created"`
}

// 当前模型对应的表名
func (r *User_E_Account) TableName() string {
	return E_ACCOUNT_TABLENAME
}

// 当前模型的数据库
func (r *User_E_Account) Using() string {
	return types.OrmDataBaseApi
}

func (r *User_E_Account) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *User_E_Account) AddEAccount(eAccount *User_E_Account) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Insert(eAccount)
	if err != nil {
		logs.Error("model user_e_account insert failed.", err)
	}
	return
}

// Xendit ballback转为结构体
func GetXenditCallBack(callbackJSON string) (xenditCallback XenditCallBack) {
	json.Unmarshal([]byte(callbackJSON), &xenditCallback)
	return
}

func GetEAccount(accountId int64, eType int) (User_E_Account, error) {

	eAccount := User_E_Account{}

	o := orm.NewOrm()
	o.Using(eAccount.Using())
	err := o.QueryTable(eAccount.TableName()).Filter("user_account_id", accountId).Filter("va_company_code", eType).One(&eAccount)

	return eAccount, err
}

//func GetActiveEAccount(accountId int64, eType int) (User_E_Account, error) {
//
//	eAccount := User_E_Account{}
//
//	o := orm.NewOrm()
//	o.Using(eAccount.Using())
//	err := o.QueryTable(eAccount.TableName()).Filter("user_account_id", accountId).Filter("va_company_code", eType).Filter("status", "ACTIVE").One(&eAccount)
//
//	return eAccount, err
//}

func GetLastestActiveEAccount(accountId int64) (User_E_Account, error) {
	eAccount := User_E_Account{}

	o := orm.NewOrm()
	o.Using(eAccount.Using())
	err := o.QueryTable(eAccount.TableName()).Filter("user_account_id", accountId).Filter("status", "ACTIVE").OrderBy("-id").One(&eAccount)

	return eAccount, err
}

func GetLastestActiveEAccountByVacompanyType(accountId int64, vacompanyType int) (User_E_Account, error) {
	eAccount := User_E_Account{}

	o := orm.NewOrm()
	o.Using(eAccount.Using())
	err := o.QueryTable(eAccount.TableName()).Filter("user_account_id", accountId).Filter("status", "ACTIVE").Filter("va_company_code", vacompanyType).OrderBy("-id").One(&eAccount)

	return eAccount, err
}

// 由"还款银行简码"，获取最新的va记录
func GetLastestActiveEAccountByRepayBank(accountId int64, repayBankCode string) (User_E_Account, error) {
	eAccount := User_E_Account{}

	o := orm.NewOrm()
	o.Using(eAccount.Using())
	err := o.QueryTable(eAccount.TableName()).Filter("user_account_id", accountId).Filter("status", "ACTIVE").
		Filter("repay_bank_code", repayBankCode).OrderBy("-id").One(&eAccount)

	return eAccount, err
}

// 由"还款银行简码"和"第三方支付公司"，获取最新的va记录
func GetLastestActiveEAccountByRepayBankAndVacompanyType(accountId int64, repayBankCode string, vacompanyType int) (User_E_Account, error) {
	eAccount := User_E_Account{}

	o := orm.NewOrm()
	o.Using(eAccount.Using())
	err := o.QueryTable(eAccount.TableName()).Filter("user_account_id", accountId).
		Filter("repay_bank_code", repayBankCode).Filter("va_company_code", vacompanyType).OrderBy("-id").One(&eAccount)

	return eAccount, err
}

// 由"放款银行简码"和"第三方支付公司"，获取最新的va记录
func GetLastestActiveEAccountByBankAndVacompanyType(accountId int64, BankCode string, vacompanyType int) (User_E_Account, error) {
	eAccount := User_E_Account{}

	o := orm.NewOrm()
	o.Using(eAccount.Using())
	err := o.QueryTable(eAccount.TableName()).Filter("user_account_id", accountId).
		Filter("bank_code", BankCode).Filter("va_company_code", vacompanyType).OrderBy("-id").One(&eAccount)

	return eAccount, err
}

func GetLastestActiveEAccountByBankCode(accountId int64, bankCode string) (User_E_Account, error) {
	eAccount := User_E_Account{}

	o := orm.NewOrm()
	o.Using(eAccount.Using())
	err := o.QueryTable(eAccount.TableName()).Filter("user_account_id", accountId).Filter("status", "ACTIVE").Filter("bank_code", bankCode).OrderBy("-id").One(&eAccount)

	return eAccount, err
}

func GetMultiEAccounts(accountId int64) ([]User_E_Account, error) {

	eAccount := User_E_Account{}

	var data []User_E_Account
	o := orm.NewOrm()
	o.Using(eAccount.Using())
	_, err := o.QueryTable(eAccount.TableName()).Filter("user_account_id", accountId).Filter("status", "ACTIVE").OrderBy("-id").All(&data)

	return data, err
}

func (r *User_E_Account) UpdateEAccount(eAccount *User_E_Account) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Update(eAccount)
	if err != nil {
		logs.Error("model user_e_account update failed.", err)
	}
	return
}

func GetEAccountByENumber(eAccountNumber string) (User_E_Account, error) {
	eAccount := User_E_Account{}
	o := orm.NewOrm()
	o.Using(eAccount.Using())
	err := o.QueryTable(eAccount.TableName()).Filter("e_account_number", eAccountNumber).Limit(1).One(&eAccount)
	return eAccount, err

}

func GetEAccountNumberByAccountId(accountId int64) ([]User_E_Account, error) {
	eAccount := User_E_Account{}
	var data []User_E_Account
	o := orm.NewOrm()
	o.Using(eAccount.Using())
	_, err := o.QueryTable(eAccount.TableName()).Filter("user_account_id", accountId).OrderBy("-id").All(&data)
	return data, err

}
