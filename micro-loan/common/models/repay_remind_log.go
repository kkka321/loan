package models

import (
	"fmt"

	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const REPAY_REMIND_CASE_LOG_TABLENAME string = "repay_remind_case_log"

type RepayRemindCaseLog struct {
	Id                int64 `orm:"pk;"`
	CaseId            int64
	OrderId           int64
	PhoneConnect      int
	PromiseRepayTime  int64
	UnrepayReason     string
	IsWillRepay       int
	UnconnectReason   int
	PhoneTime         int64
	OpUid             int64
	PhoneObject       int
	PhoneObjectMobile string
	UrgeType          int
	Result            string
	Ctime             int64
	Utime             int64
}

func (*RepayRemindCaseLog) TableName() string {
	return REPAY_REMIND_CASE_LOG_TABLENAME
}

func (*RepayRemindCaseLog) Using() string {
	return types.OrmDataBaseApi
}

func (r *RepayRemindCaseLog) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

//! 取最后一条还款提醒记录
func GetOneLastRepayRemindLogByOrderID(OrderID int64) (repayremind RepayRemindCaseLog, err error) {
	o := orm.NewOrm()
	o.Using(repayremind.UsingSlave())
	err = o.QueryTable(repayremind.TableName()).
		Filter("order_id", OrderID).
		OrderBy("-id").
		Limit(1).
		One(&repayremind)
	return
}

func GetRepayRemindCaseLogListByCaseID(caseID int64) (data []RepayRemindCaseLog, err error) {
	o := orm.NewOrm()

	obj := RepayRemindCaseLog{}

	o.Using(obj.Using())

	_, err = o.QueryTable(obj.TableName()).Filter("case_id", caseID).
		OrderBy("-id").
		All(&data)

	return
}

// GetUserRepayRemindHandleCount 获取指定时间范围，用户repay remind case 处理量
func GetUserRepayRemindHandleCount(ticketItem types.TicketItemEnum, startTimestamp, endTimestamp int64) (usersTicketCount []UserTicketCount) {
	caseLevel := types.TicketItemMap()[ticketItem]
	where := fmt.Sprintf("WHERE s.level='%s' and p.ctime>=%d and p.ctime<%d", caseLevel, startTimestamp, endTimestamp)
	sql := fmt.Sprintf("SELECT p.op_uid as uid, COUNT(DISTINCT p.case_id) as num FROM `%s` p LEFT JOIN %s s ON p.case_id=s.id  %s GROUP BY p.op_uid",
		REPAY_REMIND_CASE_LOG_TABLENAME, REPAY_REMIND_CASE_TABLENAME, where)

	// where := fmt.Sprintf("WHERE ctime>=%d and ctime<%d", startTimestamp, endTimestamp)
	// sql := fmt.Sprintf("SELECT op_uid as uid, COUNT(DISTINCT case_id) as num FROM `%s` %s GROUP BY op_uid",
	// 	REPAY_REMIND_CASE_LOG_TABLENAME, where)
	obj := RepayRemindCaseLog{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	r := o.Raw(sql)
	r.QueryRows(&usersTicketCount)
	return
}
