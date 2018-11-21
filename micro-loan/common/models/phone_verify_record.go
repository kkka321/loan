package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const PHONE_VERIFY_RECORD_TABLENAME string = "phone_verify_record"

type PhoneVerifyRecord struct {
	Id                  int64 `orm:"pk;"`
	OrderId             int64
	Q1Id                int
	Q1Value             string
	Q1Status            int
	Q2Id                int
	Q2Value             string
	Q2Status            int
	Q3Id                int
	Q3Value             string
	Q3Status            int
	Q4Id                int
	Q4Value             string
	Q4Status            int
	Q5Id                int
	Q5Value             string
	Q5Status            int
	Q6Id                int
	Q6Value             string
	Q6Status            int
	AnswerPhoneStatus   int
	IdentityInfoStatus  int
	OwnerMobileStatus   int
	OwnerMobileWhatsapp int
	RedirectReject      int
	InvalidReason       int
	OpUid               int64
	Remark              string
	Result              int
	Ctime               int64
}

func (r *PhoneVerifyRecord) TableName() string {
	return PHONE_VERIFY_RECORD_TABLENAME
}

func (r *PhoneVerifyRecord) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *PhoneVerifyRecord) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func AddOnePhoneVerifyRecord(phoneVerify PhoneVerifyRecord) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(phoneVerify.Using())

	id, err = o.Insert(&phoneVerify)

	return
}

type RefuseRecord struct {
	OrderId int64
	Ctime   int64
}

// GetRefuseRecordByPhoneStatus 获取因电话异常被电核拒绝的记录
func GetRefuseRecordByPhoneStatus(beforeNday int64) (refuseRecords []RefuseRecord, err error) {
	record := PhoneVerifyRecord{}
	records := []PhoneVerifyRecord{}
	o := orm.NewOrm()
	o.Using(record.UsingSlave())
	_, err = o.QueryTable(record.TableName()).
		Filter("answer_phone_status", 2).
		Filter("ctime__gt", beforeNday).
		OrderBy("id").
		All(&records)

	if err == nil && len(records) > 0 {
		for _, v := range records {

			exist := false
			for _, vv := range refuseRecords {
				if v.OrderId == vv.OrderId {
					exist = true
				}
			}
			if exist == true {
				continue
			}
			refuseRecord := RefuseRecord{}
			refuseRecord.OrderId = v.OrderId
			refuseRecord.Ctime = v.Ctime
			refuseRecords = append(refuseRecords, refuseRecord)
		}
	}

	return
}
