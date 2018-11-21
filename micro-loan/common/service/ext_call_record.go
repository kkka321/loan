package service

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"github.com/astaxie/beego/orm"
)

type SipCallRecords struct {
	Id              int64 `orm:"pk;"`              //通话记录id
	OrderId         int64 `orm:"column(order_id)"` //借款订单id
	ItemId          int   `orm:"column(item_id)"`  //工单类型id
	ItemIdS         string
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
	NickName        string `orm:"column(nickname)"`
}

func ListExtCallHistoryBackend(condCntr map[string]interface{}, page int, pagesize int) (lists []SipCallRecords, total int64, err error) {
	obj := models.SipCallRecord{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pagesize

	// 初始化查询条件
	where := whereExtCallHistoryBackend(condCntr)
	sqlCount := fmt.Sprintf(`SELECT COUNT(sip_call_record.id) FROM %s %s`, obj.TableName(), where)
	sqlList := fmt.Sprintf(`SELECT sip_call_record.id, order_id,item_id,call_id,assign_id,extnumber, disnumber,
		destnumber,call_direction,start_timestamp,answer_timestamp,end_timestamp,is_dial,billsec,duration,
		hangup_direction,hangup_cause,audio_record_name,call_method,ctime,
		nickname FROM %s %s ORDER BY sip_call_record.id desc LIMIT %d,%d`, obj.TableName(), where, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	list := []SipCallRecords{}
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	for _, v := range list {
		tp := SipCallRecords{}
		tp.AnswerTimestamp = v.AnswerTimestamp
		tp.BillSec = v.BillSec
		tp.CallDirection = v.CallDirection
		tp.CallId = v.CallId
		tp.Ctime = v.Ctime
		tp.DestNumber = tools.MobileDesensitization(v.DestNumber)
		tp.DisNumber = v.DisNumber
		tp.Duration = v.Duration
		tp.EndTimestamp = v.EndTimestamp
		tp.ExtNumber = v.ExtNumber
		tp.HangupCause = v.HangupCause
		tp.HangupDirection = v.HangupDirection
		tp.Id = v.Id
		tp.IsDial = v.IsDial
		tp.NickName = v.NickName
		tp.OrderId = v.OrderId
		tp.AudioRecordName = v.AudioRecordName
		tp.StartTimestamp = v.StartTimestamp
		itemIds, ok := types.TicketItemMap()[types.TicketItemEnum(v.ItemId)]
		if !ok {
			itemIds = "-"
		}
		tp.ItemIdS = itemIds
		tp.CallMethod = v.CallMethod

		lists = append(lists, tp)
	}

	return
}

func whereExtCallHistoryBackend(condCntr map[string]interface{}) string {
	// 初始化查询条件
	cond := []string{}

	//借款id
	if v, ok := condCntr["order_id"]; ok {
		cond = append(cond, fmt.Sprintf("order_id=%v", v))
	}

	//分机号
	if v, ok := condCntr["ext_number"]; ok {
		cond = append(cond, fmt.Sprintf("extnumber=%v", v))
	}
	//工单类型
	if v, ok := condCntr["item_id"]; ok {
		cond = append(cond, fmt.Sprintf("item_id=%v", v))
	}

	//时间筛选
	if v, ok := condCntr["callStartTime"]; ok {
		callEndTime := condCntr["callEndTime"]
		cond = append(cond, fmt.Sprintf("start_timestamp > %v AND start_timestamp < %v", v.(int64), callEndTime))
	}
	//用户姓名
	if v, ok := condCntr["name"]; ok {
		if len(cond) > 0 {
			return fmt.Sprintf("left join microloan_admin.admin on assign_id = microloan_admin.admin.id where microloan_admin.admin.nickname = '%v' AND ", v) +
				strings.Join(cond, " AND ")
		} else {
			return fmt.Sprintf("left join microloan_admin.admin on assign_id = microloan_admin.admin.id where microloan_admin.admin.nickname = '%v' ", v)
		}

	}

	if len(cond) > 0 {
		return "left join microloan_admin.admin on assign_id = microloan_admin.admin.id WHERE " + strings.Join(cond, " AND ")
	} else {
		return "left join microloan_admin.admin on assign_id = microloan_admin.admin.id"
	}

}
