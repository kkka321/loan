package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const CUSTOMER_RECALL_TAG_LOG_TABLENAME string = "customer_recall_tag_change_log"

type CustomerRecallTagChangeLog struct {
	Id              int64
	AccountId       int64
	OrderId         int64
	OrgionRecallTag int
	EditRecallTag   int
	Remark          int
	Ctime           int64
	Utime           int64
}

func (r *CustomerRecallTagChangeLog) TableName() string {
	return CUSTOMER_RECALL_TAG_LOG_TABLENAME
}

func (r *CustomerRecallTagChangeLog) Using() string {
	return types.OrmDataBaseApi
}
func (r *CustomerRecallTagChangeLog) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

// OneRecallPhoneVerifyTagLogByAOID 获取首次标记为电核拒绝需要召回的记录
func OneRecallPhoneVerifyTagLogByAOID(accountId, orderId int64) (one CustomerRecallTagChangeLog, err error) {
	o := orm.NewOrm()
	o.Using(one.UsingSlave())

	err = o.QueryTable(one.TableName()).
		Filter("account_id", accountId).
		Filter("order_id", orderId).
		OrderBy("-id").
		One(&one)
	return
}

//
//func OneAccountBaseExtByPkId(accountId int64) (one AccountBaseExt, err error) {
//	o := orm.NewOrm()
//	o.Using(one.Using())
//
//	err = o.QueryTable(one.TableName()).
//		Filter("account_id", accountId).
//		Filter("operate_type", opType).
//		Filter("edit_recall_tag", recallTag).
//		One(&one)
//	if err != nil && err != orm.ErrNoRows {
//		logs.Error("[OneCustomerRecallByIdAndType] err:%v accountId:%d opType:%d", err, accountId, opType)
//	}
//	return
//}
