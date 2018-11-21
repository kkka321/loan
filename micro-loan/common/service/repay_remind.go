package service

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty/voip"

	"github.com/astaxie/beego/orm"
)

type RepayRemindCaseLogs struct {
	Id                int64 `orm:"pk;"`
	PhoneConnect      int
	PromiseRepayTime  int64
	UnrepayReason     string
	IsWillRepay       int
	UnconnectReason   int
	PhoneTime         int64
	OpUid             int64
	PhoneObject       int
	PhoneObjectMobile string
	Result            string
	UrgeType          int

	AnswerTimestamp int64
	EndTimestamp    int64
	HangupDirection int
	HangupCause     int

	CallMethod int
}

func GetRepayRemindCaseLogListByOrderId(orderId int64) (data []RepayRemindCaseLogs, err error) {
	// 查询'还款提醒'
	o := orm.NewOrm()
	obj := models.RepayRemindCaseLog{}
	o.Using(obj.UsingSlave())

	// 初始化查询条件
	selectSql := `SELECT op_uid, phone_object, phone_object_mobile, phone_time, phone_connect, promise_repay_time, unrepay_reason, is_will_repay, unconnect_reason, urge_type,result`
	where := fmt.Sprintf(`where repay_remind_case_log.order_id = %v `, orderId)
	sqlList := fmt.Sprintf(`%s FROM %s %s ORDER BY repay_remind_case_log.id desc`, selectSql, obj.TableName(), where)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRows(&data)

	// 查询'通话记录'
	objSipCallRecord := models.SipCallRecord{}
	o.Using(objSipCallRecord.UsingSlave())

	selectSipCallRecord := `SELECT answer_timestamp, end_timestamp, hangup_direction, hangup_cause, call_method`
	for k, v := range data {
		if v.PhoneTime > 0 {
			var dataSipCallRecord RepayRemindCaseLogs
			whereSipCallRecord := fmt.Sprintf(`where start_timestamp = %d and call_method = 3`, v.PhoneTime)
			sql := fmt.Sprintf(`%s from %s %s`, selectSipCallRecord, objSipCallRecord.TableName(), whereSipCallRecord)
			r := o.Raw(sql)
			r.QueryRow(&dataSipCallRecord)

			if dataSipCallRecord.CallMethod == voip.VoipCallMethodSipCall {
				data[k].AnswerTimestamp = dataSipCallRecord.AnswerTimestamp
				data[k].EndTimestamp = dataSipCallRecord.EndTimestamp
				data[k].HangupCause = dataSipCallRecord.HangupCause
				data[k].HangupDirection = dataSipCallRecord.HangupDirection
				data[k].CallMethod = dataSipCallRecord.CallMethod
			} else {
				data[k].CallMethod = voip.VoipCallManual
			}
		} else {
			data[k].CallMethod = voip.VoipCallManual
		}
	}

	return
}
