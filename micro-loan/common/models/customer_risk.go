package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

type CustomerRisk struct {
	Id             int64              `orm:"pk"`
	CustomerId     int64              `orm:"column(customer_id)"`
	RiskItem       types.RiskItemEnum `orm:"column(risk_item)"`
	RiskType       types.RiskTypeEnum `orm:"column(risk_type)"`
	RiskValue      string             `orm:"column(risk_value)"`
	Reason         types.RiskReason
	RelieveReason  types.RiskRelieveReason `orm:"column(relieve_reason)"`
	ReportRemark   string                  `orm:"column(report_remark)"`
	ReviewRemark   string                  `orm:"column(review_remark)"`
	RelieveRemark  string                  `orm:"column(relieve_remark)"`
	OpUid          int64                   `orm:"column(op_uid)"`
	IsDeleted      int                     `orm:"column(is_deleted)"`
	Status         types.RiskStatusEnum
	Ctime          int64
	Utime          int64
	ReviewTime     int64 `orm:"column(review_time)"`
	RelieveTime    int64 `orm:"column(relieve_time)"`
	OrderIds       string
	UserAccountIds string
}

const CUSTOMER_RISK_TABLENAME string = "customer_risk"

func (r *CustomerRisk) TableName() string {
	return CUSTOMER_RISK_TABLENAME
}

func (r *CustomerRisk) Using() string {
	return types.OrmDataBaseAdmin
}

// 使用从库
func (r *CustomerRisk) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// IsBlacklistIdentity 是否命中内部身份证黑名单
func IsBlacklistIdentity(identity string) (yes bool, err error) {
	yes, err = isRiskItemValue(types.RiskItemIdentity, identity)
	return
}

// IsBlacklistMobile 是否命中内部电话黑名单
func IsBlacklistMobile(mobile string) (yes bool, err error) {
	yes, err = isRiskItemValue(types.RiskItemMobile, mobile)
	return
}

// IsBlacklistIP 是否命中内部IP黑名单
func IsBlacklistIP(mobile string) (yes bool, err error) {
	yes, err = isRiskItemValue(types.RiskItemIP, mobile)
	return
}

func IsBlacklistItem(riskType types.RiskItemEnum, value string) (yes bool, err error) {
	yes, err = isRiskItemValue(riskType, value)
	return
}

func isRiskItemValue(ri types.RiskItemEnum, v string) (yes bool, err error) {
	m := CustomerRisk{}
	o := orm.NewOrm()
	o.Using(m.UsingSlave())

	count, err := o.QueryTable(m.TableName()).
		Filter("risk_value", v).
		Filter("risk_item", ri).
		Filter("risk_type", types.RiskBlacklist).
		Filter("is_deleted", types.DeletedNo).
		Filter("status", types.RiskReviewPass).
		Count()
	if count > 0 {
		yes = true
	}

	return
}
