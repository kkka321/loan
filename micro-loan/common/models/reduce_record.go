package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const REDUCE_RECORD_TABLENAME string = "reduce_record"

type ReduceRecord struct {
	Id              int64 `orm:"pk;"`
	OrderId         int64 `orm:"column(order_id)"`
	OpUid           int64 `orm:"column(op_uid)"`
	AmountReduced   int64 `orm:"column(amount_reduced)"`
	PenaltyReduced  int64 `orm:"column(penalty_reduced)"`
	InterestReduced int64 `orm:"column(interest_reduced)"`
	ReduceType      int
	OpReason        string `orm:"column(op_reason)"`
	Ctime           int64
	Utime           int64
}

// 当前模型对应的表名
func (r *ReduceRecord) TableName() string {
	return REDUCE_RECORD_TABLENAME
}

// 当前模型的数据库
func (r *ReduceRecord) Using() string {
	return types.OrmDataBaseApi
}

func (r *ReduceRecord) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func addReduceRecord(reduceRecord *ReduceRecord) (id int64, err error) {
	o := orm.NewOrm()
	obj := ReduceRecord{}
	o.Using(obj.Using())
	id, err = o.Insert(reduceRecord)
	if err != nil {
		logs.Error("model reduce_record AddReduceRecord failed.", err)
	}
	return
}

func updateReduceRecord(reduceRecord *ReduceRecord) (id int64, err error) {
	o := orm.NewOrm()
	obj := &ReduceRecord{}
	o.Using(obj.Using())
	id, err = o.Update(reduceRecord)
	if err != nil {
		logs.Error("model reduce_record UpdateReduceRecord failed.", err)
	}
	return
}

func InsertReduceRecord(orderId int64, reductionAmount int64, penaltyReduced int64, interestReduced int64, recordType int, opReason string, opUid int64) error {
	o := orm.NewOrm()
	obj := ReduceRecord{
		OrderId:         orderId,
		OpUid:           opUid,
		AmountReduced:   reductionAmount,
		PenaltyReduced:  penaltyReduced,
		InterestReduced: interestReduced,
		ReduceType:      recordType,
		OpReason:        opReason,
		Ctime:           tools.GetUnixMillis(),
		Utime:           tools.GetUnixMillis(),
	}
	o.Using(obj.Using())
	_, err := o.Insert(&obj)
	if err != nil {
		logs.Error("model reduce_record ReduceRecord failed.", err)
	}
	return err
}

func GetLastestReduceRecord(orderId int64) (one ReduceRecord, err error) {
	r := ReduceRecord{}
	o := orm.NewOrm()
	o.Using(r.Using())
	err = o.QueryTable(r.TableName()).Filter("order_id", orderId).OrderBy("-id").One(&one)
	return
}

func GetAllReduceRecord(orderId int64) (list []ReduceRecord, err error) {
	r := ReduceRecord{}
	o := orm.NewOrm()
	o.Using(r.UsingSlave())
	_, err = o.QueryTable(r.TableName()).Filter("order_id", orderId).OrderBy("-id").All(&list)
	if err != nil {
		logs.Error("[model.GetAllReduceRecord] err:", err)
	}
	return
}
