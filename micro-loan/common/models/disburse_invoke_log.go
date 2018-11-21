package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const DISBURSE_INVOKE_LOG_TABLENAME string = "disburse_invoke_log"

type DisburseInvokeLog struct {
	Id             int64 `orm:"pk;column(id)"`
	OrderId        int64
	UserAccountId  int64
	VaCompanyCode  int
	BankName       string
	BankNo         string
	DisbursementId string
	DisbureStatus  int
	FailureCode    string
	HttpCode       int
	Ctime          int64
	Utime          int64
}

func (r *DisburseInvokeLog) TableName() string {
	return DISBURSE_INVOKE_LOG_TABLENAME
}

func (r *DisburseInvokeLog) Using() string {
	return types.OrmDataBaseApi
}
func (r *DisburseInvokeLog) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func OneDisburseInvorkLogByPkId(id int64) (one DisburseInvokeLog, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).
		Filter("id", id).
		One(&one)
	return
}

func GetLastestDisburseInvorkLogByPkOrderId(OrderId int64) (one DisburseInvokeLog, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).
		Filter("order_id", OrderId).
		OrderBy("-id").
		One(&one)
	return
}
