package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const REPAY_REMIND_CASE_TABLENAME string = "repay_remind_case"

const (
	RepayRemindStatusValid   int = 1
	RepayRemindStatusInValid int = 2
)

// RM 案件失效原因
const (
	RepayRemindInvalidReasonExpired = 1
	RepayRemindInvalidReasonCleared = 2
)

type RepayRemindCase struct {
	Id            int64 `orm:"pk;"`
	OrderId       int64
	Level         string
	UserAccountId int64
	//AssignUid        int64
	OpUid            int64
	PromiseRepayTime int64
	Result           string
	Status           int
	InvalidReason    int
	Ctime            int64
	InvalidTime      int64
	Utime            int64
}

func (*RepayRemindCase) TableName() string {
	return REPAY_REMIND_CASE_TABLENAME
}

func (*RepayRemindCase) Using() string {
	return types.OrmDataBaseApi
}

func (r *RepayRemindCase) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func OneRepayRemindCaseByPkID(id int64) (oneCase RepayRemindCase, err error) {
	o := orm.NewOrm()
	o.Using(oneCase.Using())

	oneCase.Id = id
	err = o.Read(&oneCase)

	return
}

// OneRepayRemindCaseByOrderID 根据OrderID获取一个指定状态的RM case
func OneRepayRemindCaseByOrderID(orderID int64, status int) (oneCase RepayRemindCase, err error) {
	o := orm.NewOrm()
	o.Using(oneCase.Using())
	err = o.QueryTable(oneCase.TableName()).Filter("order_id", orderID).
		Filter("status", status).
		OrderBy("-id").
		Limit(1).
		One(&oneCase)

	return
}

// OneVaildRepayRemindCaseByOrderID 根据OrderID获取一个有效的RM case
func OneVaildRepayRemindCaseByOrderID(orderID int64) (oneCase RepayRemindCase, err error) {
	return OneRepayRemindCaseByOrderID(orderID, types.StatusValid)
}
