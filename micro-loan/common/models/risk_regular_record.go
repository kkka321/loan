package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const RISK_REGULAR_RECORD_TABLENAME string = "risk_regular_record"

type RiskRegularRecord struct {
	Id         int64 `orm:"pk;"`
	OrderId    int64
	AccountId  int64
	HitRegular string
	Status     int
	Ctime      int64
}

func (r *RiskRegularRecord) TableName() string {
	return RISK_REGULAR_RECORD_TABLENAME
}

func (r *RiskRegularRecord) Using() string {
	return types.OrmDataBaseApi
}

func (r *RiskRegularRecord) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func AddOneRiskRegularRecord(one RiskRegularRecord) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	id, err = o.Insert(&one)

	return
}
