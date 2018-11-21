package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const REDUCE_RECORD_NEW_TABLENAME string = "reduce_record_new"

type ReduceRecordNew struct {
	Id                            int64 `orm:"pk;"`
	OrderId                       int64 `orm:"column(order_id)"`
	UserAccountId                 int64
	ApplyUid                      int64
	ConfirmUid                    int64
	AmountReduced                 int64 `orm:"column(amount_reduced)"`
	PenaltyReduced                int64 `orm:"column(penalty_reduced)"`
	GraceInterestReduced          int64
	ReduceType                    int
	ReduceStatus                  int
	OpReason                      string `orm:"column(op_reason)"`
	ConfirmRemark                 string
	ApplyTime                     int64
	ConfirmTime                   int64
	CaseID                        int64   `orm:"column(case_id)"`
	DerateRatio                   float64 `orm:"column(derate_ratio)"`
	GracePeriodInterestPrededuced int64   `orm:"column(grace_period_interest_prereduced)"`
	PenaltyPrereduced             int64   `orm:"column(penalty_prereduced)"`
	InvalidReason                 string  `orm:"column(invalid_reason)"`
	Ctime                         int64
	Utime                         int64
}

type ReduceRecordListItem struct {
	ReduceRecordNew
	Name   string
	Mobile string
}

// 当前模型对应的表名
func (r *ReduceRecordNew) TableName() string {
	return REDUCE_RECORD_NEW_TABLENAME
}

// 当前模型的数据库
func (r *ReduceRecordNew) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *ReduceRecordNew) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

//func addReduceRecordNew(reduceRecord *ReduceRecordNew) (id int64, err error) {
//	o := orm.NewOrm()
//	obj := ReduceRecordNew{}
//	o.Using(obj.Using())
//	id, err = o.Insert(reduceRecord)
//	if err != nil {
//		logs.Error("model reduce_record AddReduceRecord failed.", err)
//	}
//	return
//}
//
//func updateReduceRecordNew(reduceRecord *ReduceRecordNew) (id int64, err error) {
//	o := orm.NewOrm()
//	obj := &ReduceRecordNew{}
//	o.Using(obj.Using())
//	id, err = o.Update(reduceRecord)
//	if err != nil {
//		logs.Error("model reduce_record UpdateReduceRecord failed.", err)
//	}
//	return
//}

//func InsertReduceRecordNew(orderId int64, reductionAmount int64, penaltyReduced int64, interestReduced int64, recordType int, opReason string, opUid int64) error {
//	o := orm.NewOrm()
//	obj := ReduceRecordNew{
//		OrderId:         orderId,
//		OpUid:           opUid,
//		AmountReduced:   reductionAmount,
//		PenaltyReduced:  penaltyReduced,
//		InterestReduced: interestReduced,
//		ReduceType:      recordType,
//		OpReason:        opReason,
//		Ctime:           tools.GetUnixMillis(),
//		Utime:           tools.GetUnixMillis(),
//	}
//	o.Using(obj.Using())
//	_, err := o.Insert(&obj)
//	if err != nil {
//		logs.Error("model reduce_record ReduceRecord failed.", err)
//	}
//	return err
//}

func GetLastestReduceRecordNew(orderId int64) (one ReduceRecordNew, err error) {
	r := ReduceRecordNew{}
	o := orm.NewOrm()
	o.Using(r.Using())
	err = o.QueryTable(r.TableName()).
		Filter("order_id", orderId).
		Filter("reduce_type", types.ReduceTypeManual).
		OrderBy("-id").One(&one)
	return
}

func GetAllReduceRecordNew(orderId int64) (list []ReduceRecordNew, err error) {
	r := ReduceRecordNew{}
	o := orm.NewOrm()
	o.Using(r.UsingSlave())
	_, err = o.QueryTable(r.TableName()).Filter("order_id", orderId).OrderBy("-id").All(&list)
	if err != nil {
		logs.Error("[model.GetAllReduceRecord] err:", err)
	}
	return
}
