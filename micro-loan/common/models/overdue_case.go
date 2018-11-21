package models

import (
	"fmt"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const OVERDUE_CASE_TABLENAME string = "overdue_case"

type OverdueCase struct {
	Id           int64 `orm:"pk;"`
	OrderId      int64
	CaseLevel    string
	OverdueDays  int
	JoinUrgeTime int64
	AssignUid    int64
	UrgeUid      int64
	UrgeTime     int64
	Result       string
	IsOut        int
	OutReason    types.UrgeOutReasonEnum
	OutUrgeTime  int64
	Utime        int64
}

func (*OverdueCase) TableName() string {
	return OVERDUE_CASE_TABLENAME
}

func (*OverdueCase) Using() string {
	return types.OrmDataBaseApi
}
func (r *OverdueCase) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func OneOverueCaseByPkId(id int64) (oneCase OverdueCase, err error) {
	o := orm.NewOrm()
	o.Using(oneCase.Using())

	oneCase.Id = id
	err = o.Read(&oneCase)

	return
}

func OneOverdueCaseByUniqueKey(orderID int64, caseLevel string, isOut int) (oneCase OverdueCase, err error) {
	o := orm.NewOrm()
	o.Using(oneCase.Using())

	err = o.QueryTable(oneCase.TableName()).Filter("order_id", orderID).
		Filter("case_level", caseLevel).
		Filter("is_out", isOut).Limit(1).One(&oneCase)

	return
}

func OneOverdueCaseByOrderID(orderID int64) (oneCase OverdueCase, err error) {
	o := orm.NewOrm()
	o.Using(oneCase.Using())

	err = o.QueryTable(oneCase.TableName()).Filter("order_id", orderID).
		OrderBy("-id").
		Limit(1).
		One(&oneCase)

	return
}

// LatestValidOverdueCaseByOrderID 获取最新未出催的案件
func LatestValidOverdueCaseByOrderID(orderID int64) (oneCase OverdueCase, err error) {
	o := orm.NewOrm()
	o.Using(oneCase.Using())

	err = o.QueryTable(oneCase.TableName()).Filter("order_id", orderID).Filter("is_out", types.IsUrgeOutNo).
		OrderBy("-id").
		Limit(1).
		One(&oneCase)

	return
}

func UpdateOverdueCase(oneCase *OverdueCase) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(oneCase.Using())

	id, err = o.Update(oneCase)

	return
}

//根据逾期天数获取逾期案件ID集合
func GetOverdueCaseIDs(overdueDays int) (IDs []int64) {
	o := orm.NewOrm()
	oneCase := OverdueCase{}
	o.Using(oneCase.Using())
	sql := fmt.Sprintf("SELECT id FROM `%s` WHERE overdue_days =%d and is_out=0", oneCase.TableName(), overdueDays)
	r := o.Raw(sql)
	r.QueryRows(&IDs)
	return
}
