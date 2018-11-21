package models

// `product`
import (
	//"github.com/astaxie/beego/logs"

	"github.com/astaxie/beego/orm"

	//"micro-loan/common/tools"

	"micro-loan/common/tools"
	"micro-loan/common/types"
	//"fmt"
)

const VOICE_REMIND_TABLENAME string = "voice_remind"

type VoiceRemind struct {
	Id       int64   `orm:"pk;"`
	Sid      int64   `orm:"column(sid)"`
	Mobile   string  `orm:"column(mobile)"`
	Duration int     `orm:"column(duration)"`
	Fee      float64 `orm:"column(fee)"`
	Status   int     `orm:"column(status)"`
	Ctime    int64
}

// 当前模型对应的表名
func (r *VoiceRemind) TableName() string {
	return VOICE_REMIND_TABLENAME
}

// 当前模型的数据库
func (r *VoiceRemind) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *VoiceRemind) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// Add 添加语音查询结果记录, 无需初始化Ctime, 内部已自动初始化
func (r *VoiceRemind) Add() (int64, error) {
	r.Ctime = tools.GetUnixMillis()
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

// 查询最近三次
func GetVoiceRemindByMobileAndStatus(mobile string, status types.VoiceType) (objs []VoiceRemind, err error) {
	o := orm.NewOrm()
	obj := VoiceRemind{}
	o.Using(obj.Using())

	_, err = o.QueryTable(obj.TableName()).Filter("mobile", mobile).Filter("status", status).OrderBy("-id").Limit(3).All(&objs)

	return
}

// 查询最近一次
func GetLatestVoiceRemindByMobileAndStatus(mobile string, status types.VoiceType) (VoiceRemind, error) {
	o := orm.NewOrm()
	obj := VoiceRemind{}
	o.Using(obj.Using())

	_, err := o.QueryTable(obj.TableName()).Filter("mobile", mobile).Filter("status", status).OrderBy("-id").Limit(1).All(&obj)

	return obj, err
}

// 查询最近一次的呼叫记录
func GetLatestVoiceRemindByMobile(mobile string) (VoiceRemind, error) {
	o := orm.NewOrm()
	obj := VoiceRemind{}
	o.Using(obj.Using())

	_, err := o.QueryTable(obj.TableName()).Filter("mobile", mobile).OrderBy("-id").Limit(1).All(&obj)

	return obj, err
}

// 查询所有呼叫记录
func GetAllVoiceRemindByMobile(mobile string) (objs []VoiceRemind, err error) {
	o := orm.NewOrm()
	obj := VoiceRemind{}
	o.Using(obj.Using())

	_, err = o.QueryTable(obj.TableName()).Filter("mobile", mobile).OrderBy("-id").All(&objs)

	return
}
