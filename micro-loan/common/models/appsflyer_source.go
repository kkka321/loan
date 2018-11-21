package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// APPSFLYER_SOURCE_TABLENAME 定义table名
const APPSFLYER_SOURCE_TABLENAME = "appsflyer_source"

// AppsflyerSource 描述 ORM
type AppsflyerSource struct {
	Id                  int64  `orm:"pk;"`
	AppsflyerID         string `orm:"column(appsflyer_id)"`
	MediaSource         string `orm:"column(media_source)"`
	Campaign            string `orm:"column(campaign)"`
	GoogleAdvertisingID string `orm:"column(google_advertising_id)"`
	AppVersion          string
	City                string
	DeviceModel         string
	Status              int
	InstallTime         int64
	Ctime               int64
	Utime               int64
}

func (r *AppsflyerSource) TableName() string {
	return APPSFLYER_SOURCE_TABLENAME
}

func (r *AppsflyerSource) Using() string {
	return types.OrmDataBaseApi
}

func (r *AppsflyerSource) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *AppsflyerSource) Add() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	if r.Ctime == 0 {
		r.Ctime = tools.GetUnixMillis()
	}
	r.Status = types.StatusValid
	id, err := o.Insert(r)
	return id, err
}

// OneAppsflyerSourceByAppsflyerID 获取单条 appsflyersoruce
func OneAppsflyerSourceByAppsflyerID(appsflyerID string) (AppsflyerSource, error) {
	obj := AppsflyerSource{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	err := o.QueryTable(obj.TableName()).Filter("appsflyer_id", appsflyerID).One(&obj)

	return obj, err
}
