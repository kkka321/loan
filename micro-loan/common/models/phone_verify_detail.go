package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const PHONE_VERIFY_CALL_DETAIL_TABLENAME string = "phone_verify_call_detail"

type PhoneVerifyCallDetail struct {
	Id           int64 `orm:"pk;"`
	OpUid        int64
	OrderId      int64
	PhoneConnect int
	PhoneTime    int64
	Result       int
	Remark       string
	Ctime        int64
	Utime        int64
}

func (*PhoneVerifyCallDetail) TableName() string {
	return PHONE_VERIFY_CALL_DETAIL_TABLENAME
}

func (*PhoneVerifyCallDetail) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *PhoneVerifyCallDetail) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func GetMultiPhoneVerifyCallDetailsByOrderId(orderId int64) (data []PhoneVerifyCallDetail, err error) {
	o := orm.NewOrm()

	obj := OverdueCaseDetail{}

	o.Using(obj.Using())

	_, err = o.QueryTable(obj.TableName()).Filter("order_id", orderId).
		OrderBy("-id").
		All(&data)

	return
}

func AddPhoneVerifyCallDetail(data *PhoneVerifyCallDetail) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(data.Using())

	id, err = o.Insert(data)

	return
}

func UpdatePhoneVerifyCallDetail(data *PhoneVerifyCallDetail) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(data.Using())

	id, err = o.Update(data)

	return
}
