package models

import (
	//"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	//"micro-loan/common/tools"
	"micro-loan/common/types"
	//"fmt"
	"github.com/astaxie/beego/logs"
)

const MOBI_E_TRANS_TABLENAME string = "mobi_e_trans"

type Mobi_E_Trans struct {
	Id                      int64 `orm:"pk;"`
	UserAcccountId          int64 `orm:"column(user_account_id)"`
	VaCompanyCode           int   `orm:"column(va_company_code)"`
	Amount                  int64
	PayType                 int
	BankCode                string `orm:"column(bank_code)"`
	AccountHolderName       string `orm:"column(account_holder_name)"`
	DisbursementDescription string `orm:"column(disbursement_description)"`
	DisbursementId          string `orm:"column(disbursement_id)"`
	Status                  string `orm:"column(status)"`
	CallbackJson            string `orm:"column(callback_json)"`
	Ctime                   int64
	Utime                   int64
}

// 当前模型对应的表名
func (r *Mobi_E_Trans) TableName() string {
	return MOBI_E_TRANS_TABLENAME
}

// 当前模型的数据库
func (r *Mobi_E_Trans) Using() string {
	return types.OrmDataBaseApi
}

func (r *Mobi_E_Trans) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *Mobi_E_Trans) AddMobiEtrans(eTrans *Mobi_E_Trans) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Insert(eTrans)
	if err != nil {
		logs.Error("model mobi_e_trans insert failed.", err)
	}
	return
}

func GetMobiEtrans(disbursementId string) (mobi_e_trans *Mobi_E_Trans, err error) {
	o := orm.NewOrm()
	r := Mobi_E_Trans{}
	o.Using(r.Using())

	mobi_e_trans = &Mobi_E_Trans{}
	//var mobi_e_trans Mobi_E_Trans

	err = o.QueryTable(r.TableName()).Filter("disbursement_id", disbursementId).One(mobi_e_trans)

	return mobi_e_trans, err
}

func (r *Mobi_E_Trans) UpdateMobiEEtrans(oTrans *Mobi_E_Trans) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(oTrans)
	if err != nil {
		logs.Error("model Mobi_E_Trans update failed.", err)
	}

	return
}

func GetLastMobiEtrans(disbursementDescription string) (mobi_e_trans *Mobi_E_Trans, err error) {
	o := orm.NewOrm()
	r := Mobi_E_Trans{}
	o.Using(r.Using())

	mobi_e_trans = &Mobi_E_Trans{}

	err = o.QueryTable(r.TableName()).Filter("disbursement_description", disbursementDescription).OrderBy("-id").One(mobi_e_trans)

	return mobi_e_trans, err
}

func GetMobiEtransByAccountIdDescription(accountId int64, disbursementDescription string) (mobi_e_trans Mobi_E_Trans, err error) {
	o := orm.NewOrm()
	o.Using(mobi_e_trans.Using())
	err = o.QueryTable(mobi_e_trans.TableName()).Filter("user_account_id", accountId).Filter("disbursement_description", disbursementDescription).OrderBy("-id").One(&mobi_e_trans)

	return mobi_e_trans, err
}
