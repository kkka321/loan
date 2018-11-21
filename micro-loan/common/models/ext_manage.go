package models

import (
	"fmt"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const SIPINFO_TABLENAME = "sip_info"

type SipInfo struct {
	Id           int64  `orm:"pk;"`
	ExtNumber    string `orm:"column(extnumber)"`     //分机号码
	AssignId     int64  `orm:"column(assign_id)"`     //分配人员id
	CallStatus   int    `orm:"column(call_status)"`   //分机通话状态
	EnableStatus int    `orm:"column(enable_status)"` //分配是否启用
	AssignStatus int    `orm:"column(assign_status)"` //分机分配状态  0:未分配; 1:已分配
	Ctime        int64  `orm:"column(ctime)"`         //创建时间
	Utime        int64  `orm:"column(utime)"`         //更新时间
}

// TableName 返回当前模型对应的表名
func (r *SipInfo) TableName() string {
	return SIPINFO_TABLENAME
}

// Using 返回当前模型的数据库
func (r *SipInfo) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *SipInfo) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *SipInfo) Insert() (int64, error) {
	r.Ctime = tools.GetUnixMillis()
	r.Utime = tools.GetUnixMillis()
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

func (r *SipInfo) Update() (num int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	r.Utime = tools.GetUnixMillis()
	num, err = o.Update(r)

	return
}

func (r *SipInfo) Updates(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}

// GetAssignedSipInfo 根据主键,查询已经分配的分机
func GetAssignedSipInfo(extNumber string) (sipInfos []SipInfo, err error) {

	obj := SipInfo{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	sqlList := fmt.Sprintf("SELECT `assign_id` FROM `%s` WHERE `assign_id` > 0 AND `extnumber` <> %s",
		SIPINFO_TABLENAME, extNumber)
	r := o.Raw(sqlList)

	_, err = r.QueryRows(&sipInfos)

	return
}

// GetSipInfoByExtNumber 根据 extNumber,查询已分配的分机信息
func GetSipInfoByExtNumber(extNumber string) (SipInfo, error) {
	var sipInfo SipInfo

	o := orm.NewOrm()
	o.Using(sipInfo.Using())
	err := o.QueryTable(sipInfo.TableName()).
		Filter("extnumber", extNumber).
		One(&sipInfo)

	return sipInfo, err
}

func GetSipInfoByExtNumbers(extNumber string) (SipInfo, error) {
	var sipInfo SipInfo

	o := orm.NewOrm()
	o.Using(sipInfo.Using())
	err := o.QueryTable(sipInfo.TableName()).
		Filter("extnumber", extNumber).OrderBy("-extnumber").Limit(1).One(&sipInfo)

	return sipInfo, err
}

// GetSipInfoByAssignID 根据 assignId,查询已分配的分机信息
func GetSipInfoByAssignID(assignID int64) (SipInfo, error) {
	var sipInfo SipInfo

	o := orm.NewOrm()
	o.Using(sipInfo.Using())
	err := o.QueryTable(sipInfo.TableName()).
		Filter("assign_id", assignID).
		Filter("assign_status", types.ExtensionAssign).
		One(&sipInfo)

	return sipInfo, err
}
