package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const SIPASSIGN_TABLENAME = "sip_assign_history"

type SipAssignHistory struct {
	Id           int64  `orm:"pk;"`                   //通话记录id
	ExtNumber    string `orm:"column(extnumber)"`     //分机号码
	AssignId     int64  `orm:"column(assign_id)"`     //分配人员id
	AssignTime   int64  `orm:"column(assign_time)"`   //分配时间
	UnAssignTime int64  `orm:"column(unassign_time)"` //未分配时间
	Utime        int64  `orm:"column(utime)"`         //更新时间
	Ctime        int64  `orm:"column(ctime)"`         //创建时间
}

// TableName 返回当前模型对应的表名
func (r *SipAssignHistory) TableName() string {
	return SIPASSIGN_TABLENAME
}

// Using 返回当前模型的数据库
func (r *SipAssignHistory) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *SipAssignHistory) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *SipAssignHistory) Insert() (int64, error) {
	r.Ctime = tools.GetUnixMillis()
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

func (r *SipAssignHistory) Update() (num int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	r.Utime = tools.GetUnixMillis()
	num, err = o.Update(r)

	return
}

func (r *SipAssignHistory) Updates(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}
func GetSipAssignHistory(extnumber string, assign_id int64) (SipAssignHistory, error) {
	o := orm.NewOrm()
	sipHistory := SipAssignHistory{}
	o.Using(sipHistory.Using())
	err := o.QueryTable(sipHistory.TableName()).Filter("extnumber", extnumber).Filter("assign_id", assign_id).OrderBy("-id").Limit(1).One(&sipHistory)

	return sipHistory, err
}
