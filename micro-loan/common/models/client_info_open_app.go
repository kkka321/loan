package models

import (
	"fmt"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const CLIENT_INFO_OPEN_APP_TABLENAME string = "client_info_open_app"

type ClientInfoOpenApp struct {
	Id             int64  `orm:"pk;"`
	IP             string `orm:"column(ip)"`
	OS             string `orm:"column(os)"`
	Imei           string `orm:"column(imei)"`
	ImeiMd5        string `orm:"column(imei_md5)"`
	UUID           string `orm:"column(uuid)"`
	UUIDMd5        string `orm:"column(uuid_md5)"`
	IsRegistered   int    `orm:"column(is_registered)"`
	Model          string
	Brand          string
	AppVersion     string `orm:"column(app_version)"`
	AppVersionCode int
	Longitude      string
	Latitude       string
	City           string
	TimeZone       string `orm:"column(time_zone)"`
	Network        string
	IsSimulator    int `orm:"column(is_simulator)"`
	Platform       string
	UiVersion      string `orm:"column(ui_version)"`
	StemFrom       string
	FcmToken       string
	Ctime          int64 `orm:"column(ctime)"`
	Utime          int64 `orm:"column(utime)"`
}

func (r *ClientInfoOpenApp) TableName() string {
	return CLIENT_INFO_OPEN_APP_TABLENAME
}

func (r *ClientInfoOpenApp) Using() string {
	return types.OrmDataBaseApi
}

func (r *ClientInfoOpenApp) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *ClientInfoOpenApp) Add() (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Insert(r)

	return
}

func (r *ClientInfoOpenApp) Updates(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}

func GetClientInfoOpenAppByUUIDMd5(uuidMd5 string) (clientInfo ClientInfoOpenApp, err error) {
	if len(uuidMd5) <= 0 {
		logs.Warning("[GetUnregisterClientInfoByUUIDMd5] Parameter uuidMd5 is blank")
		err = fmt.Errorf("Parameter uuidMd5 is blank")
		return
	}

	o := orm.NewOrm()
	o.Using(clientInfo.UsingSlave())

	err = o.QueryTable(clientInfo.TableName()).Filter("uuid_md5", uuidMd5).OrderBy("-id").Limit(1).One(&clientInfo)

	return
}

// 提醒注册列表
func GetNeedRemindRegisterUUID() (list []string, err error) {
	clientInfoOpenApp := ClientInfoOpenApp{}
	o := orm.NewOrm()
	o.Using(clientInfoOpenApp.UsingSlave())

	t := tools.GetUnixMillis()
	sql := fmt.Sprintf(`SELECT c.uuid_md5 FROM %s c 
		WHERE c.is_registered = %d AND c.fcm_token != '' AND (c.ctime > %d AND c.ctime <= %d)`,
		clientInfoOpenApp.TableName(),
		types.UUIDUnRegistered, t-24*3600*1000, t)
	_, err = o.Raw(sql).QueryRows(&list)

	return
}
