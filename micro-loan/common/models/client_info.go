package models

import (
	"micro-loan/common/types"

	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const CLIENT_INFO_TABLENAME string = "client_info"

type ClientInfo struct {
	Id             int64 `orm:"pk;"`
	Mobile         string
	ServiceType    types.ServiceType `orm:"column(service_type)"`
	RelatedId      int64             `orm:"column(related_id)"`
	IP             string            `orm:"column(ip)"`
	OS             string            `orm:"column(os)"`
	Imei           string            `orm:"column(imei)"`
	ImeiMd5        string            `orm:"column(imei_md5)"`
	UUID           string            `orm:"column(uuid)"`
	UUIDMd5        string            `orm:"column(uuid_md5)"`
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
	Ctime          int64
}

func (r *ClientInfo) TableName() string {
	return CLIENT_INFO_TABLENAME
}

func (r *ClientInfo) Using() string {
	return types.OrmDataBaseApi
}

func (r *ClientInfo) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func OneLastClientInfoByRelatedID(relatedID int64) (clientInfo ClientInfo, err error) {
	if relatedID <= 0 {
		err = fmt.Errorf("wrong related_id: %d", relatedID)

		return
	}

	o := orm.NewOrm()
	o.Using(clientInfo.Using())

	err = o.QueryTable(clientInfo.TableName()).Filter("related_id", relatedID).OrderBy("-id").Limit(1).One(&clientInfo)

	return
}

// 注册过的客户端信息
func LatestRegisteredClientInfoByUUIDMd5(uuidMd5 string) (clientInfo ClientInfo, err error) {
	if len(uuidMd5) <= 0 {
		err = fmt.Errorf("Parameter uuidMd5 is blank")
		return
	}

	o := orm.NewOrm()
	o.Using(clientInfo.Using())
	cond := orm.NewCondition()
	cond = cond.And("UUIDMd5", uuidMd5).And("RelatedId__gt", 0)

	err = o.QueryTable(clientInfo.TableName()).SetCond(cond).
		OrderBy("-id").Limit(1).One(&clientInfo)

	return
}

func OrderClientInfo(orderId int64) (clientInfo ClientInfo, err error) {
	if orderId <= 0 {
		logs.Warning("[OrderClientInfo] invalid orderId:", orderId)
		err = fmt.Errorf("invalid orderId: %d", orderId)
		return
	}

	o := orm.NewOrm()
	o.Using(clientInfo.UsingSlave())

	err = o.QueryTable(clientInfo.TableName()).Filter("related_id", orderId).Filter("service_type", types.ServiceCreateOrder).OrderBy("-id").Limit(1).One(&clientInfo)

	return
}
