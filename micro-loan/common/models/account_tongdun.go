package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const ACCOUNT_TONGDUN_TABLENAME string = "account_tongdun"

// AccountTongdun 表结构
type AccountTongdun struct {
	ID          int64  `orm:"pk;column(id)"`
	AccountID   int64  `orm:"column(account_id)"`
	TaskID      string `orm:"column(task_id)"`
	OcrRealName string `orm:"column(ocr_realname)"`
	OcrIdentity string `orm:"column(ocr_identity)"`
	Mobile      string `orm:"column(mobile)"`
	CheckCode   int64  `orm:"column(check_code)"`
	Message     string `orm:"column(message)"`
	IsMatch     string `orm:"column(is_match)"`
	ChannelType string `orm:"column(channel_type)"`
	ChannelCode string `orm:"column(channel_code)"`
	ChannelSrc  string `orm:"column(channel_src)"`
	ChannelAttr string `orm:"column(channel_attr)"`
	CreateTimeS string `orm:"column(create_time_s)"`
	CreateTime  int64  `orm:"column(create_time)"`
	NotifyTimeS string `orm:"column(notify_time_s)"`
	NotifyTime  int64  `orm:"column(notify_time)"`
	Source      int64  `orm:"column(source)"`
	TaskData    string
}

func (r *AccountTongdun) TableName() string {
	return ACCOUNT_TONGDUN_TABLENAME
}

func (r *AccountTongdun) Using() string {
	return types.OrmDataBaseApi
}

func (r *AccountTongdun) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

// InsertTongdun 入库
func InsertTongdun(tongdun AccountTongdun) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(tongdun.Using())
	id, err = o.Insert(&tongdun)
	return
}

// UpdateTongdun 更新数据库
func UpdateTongdun(tongdun AccountTongdun, cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(tongdun.Using())
	id, err = o.Update(&tongdun, cols...)
	return
}

// GetOneByCondition 根据任意字段获取一条记录
func GetOneByCondition(filed, val string) (AccountTongdun, error) {
	var atIns = AccountTongdun{}
	o := orm.NewOrm()
	o.Using(atIns.Using())
	err := o.QueryTable(atIns.TableName()).Filter(filed, val).One(&atIns)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[GetOneByCondition] sql error err:%v", err)
	}
	return atIns, err
}

// GetOneAC 根据accountID,channelCode
func GetOneAC(accountID int64, channelCode string) (AccountTongdun, error) {
	var atIns = AccountTongdun{}
	o := orm.NewOrm()
	o.Using(atIns.Using())
	err := o.QueryTable(atIns.TableName()).
		Filter("account_id", accountID).
		Filter("channel_code", channelCode).
		OrderBy("-id").
		One(&atIns)
	return atIns, err
}

// GetLatestSuccessACByChannelCode 根据accountID,channelCode
func GetLatestSuccessACByChannelCode(accountID int64, channelCode string) (AccountTongdun, error) {
	var atIns = AccountTongdun{}
	o := orm.NewOrm()
	o.Using(atIns.Using())
	err := o.QueryTable(atIns.TableName()).
		Filter("account_id", accountID).
		Filter("channel_code", channelCode).
		Exclude("task_data", "").
		Exclude("task_data", "null").
		OrderBy("-id").
		One(&atIns)
	return atIns, err
}
