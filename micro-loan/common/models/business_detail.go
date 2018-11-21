package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const BUSSINESS_DETAIL_TABLENAME string = "business_detail"

type BusinessDetail struct {
	Id                    int64 `orm:"pk;"`
	PaymentName           string
	RechargeAmount        int64
	WithdrawAmount        int64
	PayOutAmount          int64
	PayOutForFee          int64
	PayInAmount           int64
	PayInForFee           int64
	AccountBalance        int64
	LendingBalance        int64
	FeeIncome             int64
	InterestIncome        int64
	GraceInterestIncome   int64
	PenaltyInterestIncome int64
	RecordDate            int64
	RecordDateS           string
	RecordType            int
	Ctime                 int64
	Utime                 int64
}

func (r *BusinessDetail) TableName() string {
	return BUSSINESS_DETAIL_TABLENAME
}

func (r *BusinessDetail) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *BusinessDetail) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *BusinessDetail) Add() (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Insert(r)
	return id, err
}

func (r *BusinessDetail) Update(col ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, col...)
	return id, err
}
