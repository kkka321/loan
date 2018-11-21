package models

import (

	//"github.com/astaxie/beego/logs"

	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const RISK_NOTIFY_TABLENAME string = "risk_notify"

// risk_notify 表结构
type RiskNotify struct {
	ID          int64  `orm:"pk;column(id)"`
	AccessToken string `orm:"column(access_token)"`
	ReqTime     int64  `orm:"column(req_time)"`
	AccountID   int64  `orm:"column(account_id)"`
	Ctime       int64  `orm:"column(ctime)"`
}

func (r *RiskNotify) TableName() string {
	return RISK_NOTIFY_TABLENAME
}

func (r *RiskNotify) Using() string {
	return types.OrmDataBaseApi
}

func (r *RiskNotify) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

// InsertRiskNotify 入库
func InsertRiskNotify(riskNotify RiskNotify) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(riskNotify.Using())
	id, err = o.Insert(&riskNotify)
	return
}

// UpdateRiskNotify 更新数据库
func UpdateRiskNotify(riskNotify RiskNotify) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(riskNotify.Using())
	id, err = o.Update(&riskNotify)
	return
}

// GetRiskNotifyCount 根据orderID
func GetRiskNotifyCount(orderID int64) (count int64, err error) {
	var atIns = RiskNotify{}
	o := orm.NewOrm()
	o.Using(atIns.Using())
	count, err = o.QueryTable(atIns.TableName()).Filter("order_id", orderID).Count()
	return
}
