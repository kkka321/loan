package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const SIPCALL_TABLENAME = "sip_call_record"

type SipCallRecord struct {
	Id              int64  `orm:"pk;"`                       //通话记录id
	OrderId         int64  `orm:"column(order_id)"`          //借款订单id
	ItemId          int64  `orm:"column(item_id)"`           //工单类型id
	CallId          int64  `orm:"column(call_id)"`           //话单id
	AssignId        int64  `orm:"column(assign_id)"`         //分配人员id
	ExtNumber       string `orm:"column(extnumber)"`         //分机号码
	DisNumber       string `orm:"column(disnumber)"`         //主叫号码
	DestNumber      string `orm:"column(destnumber)"`        //被叫号码
	CallDirection   int    `orm:"column(call_direction)"`    //呼叫方向  0:呼出; 1:呼入
	CallMethod      int    `orm:"column(call_method)"`       //呼叫方法 1：分机互拨，2：分机直拔，3：API呼叫，4：双呼
	StartTime       string `orm:"column(start_time)"`        //拨打时间, 目前是北京时间
	AnswerTime      string `orm:"column(answer_time)"`       //应答时间, 目前是北京时间
	EndTime         string `orm:"column(end_time)"`          //结束时间, 目前是北京时间
	StartTimestamp  int64  `orm:"column(start_timestamp)"`   //拨打时间戳
	AnswerTimestamp int64  `orm:"column(answer_timestamp)"`  //拨打时间戳
	EndTimestamp    int64  `orm:"column(end_timestamp)"`     //结束时间戳
	IsDial          int    `orm:"column(is_dial)"`           //是否拨通  0:未拨通; 1:已拨通
	BillSec         int64  `orm:"column(billsec)"`           //通话时长
	Duration        int64  `orm:"column(duration)"`          //接通前等待时长
	HangupDirection int    `orm:"column(hangup_direction)"`  //挂机方向
	HangupCause     int    `orm:"column(hangup_cause)"`      //挂机原因
	AudioRecordName string `orm:"column(audio_record_name)"` //录音文件名
	Ctime           int64  `orm:"column(ctime)"`             //创建时间
	Utime           int64  `orm:"column(utime)"`             //更新时间
}

// TableName 返回当前模型对应的表名
func (r *SipCallRecord) TableName() string {
	return SIPCALL_TABLENAME
}

// Using 返回当前模型的数据库
func (r *SipCallRecord) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *SipCallRecord) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *SipCallRecord) Insert() (int64, error) {
	timestamp := tools.GetUnixMillis()
	r.Ctime = timestamp
	r.Utime = timestamp
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

func (r *SipCallRecord) Update() (num int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	r.Utime = tools.GetUnixMillis()
	num, err = o.Update(r)

	return
}

func (r *SipCallRecord) Updates(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}

// GetSipCallRecordById 根据 id,查询通话记录
func GetSipCallRecordById(id int64) (callRecord SipCallRecord, err error) {

	o := orm.NewOrm()
	o.Using(callRecord.Using())
	err = o.QueryTable(callRecord.TableName()).
		Filter("id", id).
		One(&callRecord)

	return
}
