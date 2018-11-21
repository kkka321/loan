package models

import (
	"fmt"

	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const OVERDUE_CASE_DETAIL_TABLENAME string = "overdue_case_detail"

type OverdueCaseDetail struct {
	Id                int64 `orm:"pk;"`
	OverdueCaseId     int64
	OrderId           int64
	PhoneConnect      int
	PromiseRepayTime  int64
	OverdueReason     string
	OverdueReasonItem types.OverdueReasonItemEnum
	RepayInclination  int
	UnconnectReason   int
	PhoneTime         int64
	OpUid             int64
	PhoneObject       int
	PhoneObjectMobile string
	Result            string
	Ctime             int64
	Utime             int64
}

func (*OverdueCaseDetail) TableName() string {
	return OVERDUE_CASE_DETAIL_TABLENAME
}

func (*OverdueCaseDetail) Using() string {
	return types.OrmDataBaseApi
}

func (r *OverdueCaseDetail) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func GetMultiDatasByOverdueCaseId(overdueCaseId int64) (data []OverdueCaseDetail, err error) {
	o := orm.NewOrm()

	obj := OverdueCaseDetail{}

	o.Using(obj.Using())

	_, err = o.QueryTable(obj.TableName()).Filter("overdue_case_id", overdueCaseId).
		OrderBy("-id").
		All(&data)

	return
}

func GetMultiDatasByOrderId(orderId int64) (data []OverdueCaseDetail, err error) {
	o := orm.NewOrm()

	obj := OverdueCaseDetail{}

	o.Using(obj.Using())

	_, err = o.QueryTable(obj.TableName()).Filter("order_id", orderId).
		OrderBy("-id").
		All(&data)

	return
}

func AddOverdueCaseDetail(oneCase *OverdueCaseDetail) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(oneCase.Using())

	id, err = o.Insert(oneCase)

	return
}

func UpdateOverdueCaseDetail(oneCase *OverdueCaseDetail) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(oneCase.Using())

	id, err = o.Update(oneCase)

	return
}

// GetUserUrgeHandleCount 获取指定时间范围，用户urge case 处理量
func GetUserUrgeHandleCount(ticketItem types.TicketItemEnum, startTimestamp, endTimestamp int64) (usersTicketCount []UserTicketCount) {
	caseLevel := types.GetOverdueLevelByTicketItem(ticketItem)

	where := fmt.Sprintf("WHERE s.case_level='%s' and p.ctime>=%d and p.ctime<%d", caseLevel, startTimestamp, endTimestamp)
	sql := fmt.Sprintf("SELECT p.op_uid as uid, COUNT(DISTINCT p.overdue_case_id) as num FROM `%s` p LEFT JOIN %s s ON p.overdue_case_id=s.id  %s GROUP BY p.op_uid",
		OVERDUE_CASE_DETAIL_TABLENAME, OVERDUE_CASE_TABLENAME, where)

	obj := OverdueCaseDetail{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	r := o.Raw(sql)
	r.QueryRows(&usersTicketCount)
	return
}
