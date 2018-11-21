package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const ENTRUST_APPROVAL_RECORD string = "entrust_approval_record"

type EntrustApprovalRecord struct {
	Id           int64
	OrderId      int64
	IsAgree      int
	AuditComment string
	Pname        string
	Remark       string
	Ctime        int64
}

func (r *EntrustApprovalRecord) TableName() string {
	return ENTRUST_APPROVAL_RECORD
}

func (r *EntrustApprovalRecord) Using() string {
	return types.OrmDataBaseAdmin
}
func (r *EntrustApprovalRecord) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func OneEntrustApprovalRecordById(Id int64) (one EntrustApprovalRecord, err error) {
	o := orm.NewOrm()
	o.Using(one.UsingSlave())
	err = o.QueryTable(one.TableName()).
		Filter("id", Id).
		OrderBy("-id").
		One(&one)
	return
}
func OneEntrustApprovalRecordByOrderId(OrderId int64) (one EntrustApprovalRecord, err error) {
	o := orm.NewOrm()
	o.Using(one.UsingSlave())
	err = o.QueryTable(one.TableName()).
		Filter("order_id", OrderId).
		OrderBy("-id").
		One(&one)
	return
}
