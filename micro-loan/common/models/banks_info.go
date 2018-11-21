package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const BANKS_INFO_TABLENAME string = "banks_info"

type BanksInfo struct {
	Id                 int64 `orm:"pk;"`
	FullName           string
	XenditBrevityName  string
	DokuFullName       string
	DokuBrevityName    string
	DokuBrevityId      string
	BluepayBrevityName string
	LoanCompanyCode    int
	RepayCompanyCode   int
	Ctime              int64
	Utime              int64
}

func (r *BanksInfo) TableName() string {
	return BANKS_INFO_TABLENAME
}

func (r *BanksInfo) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *BanksInfo) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *BanksInfo) Add() (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Insert(r)
	return id, err
}

func (r *BanksInfo) Update(col ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, col...)
	return id, err
}

func OneBankInfoByFullName(fullName string) (one BanksInfo, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).Filter("full_name", fullName).One(&one)
	return
}

func OneBankInfoByXenditBrevity(xenditBrevity string) (one BanksInfo, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).Filter("xendit_brevity_name", xenditBrevity).One(&one)
	return
}

func BankListByCompanyType(company int, loanRepayType int) (list []BanksInfo, err error) {
	one := BanksInfo{}
	o := orm.NewOrm()
	o.Using(one.Using())

	cond := orm.NewCondition()
	switch loanRepayType {
	case types.LoanRepayTypeLoan:
		{
			cond = cond.And("loan_company_code", company)
		}
	case types.LoanRepayTypeRepay:
		{
			cond = cond.And("repay_company_code", company)
		}
	}

	_, err = o.QueryTable(one.TableName()).
		SetCond(cond).
		All(&list)
	return
}
