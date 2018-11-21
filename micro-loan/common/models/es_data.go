package models

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

// ROLE_TABLENAME 表名
const ES_DATA_TABLENAME string = "es_data"

// Role 描述数据表结构与结构体的映射
type EsData struct {
	Id        int64  `orm:"pk;"`
	OrderId   int64  `orm:"column(order_id)"`
	AccountId int64  `orm:"column(account_id)"`
	EsIndex   string `orm:"column(es_index)"`
	Data      string
	Ctime     int64
	Utime     int64
}

// TableName 返回当前模型对应的表名
func (r *EsData) TableName() string {
	return ES_DATA_TABLENAME
}

// Using 返回当前模型的数据库
func (r *EsData) Using() string {
	return types.OrmDataBaseApi
}

func (r *EsData) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func OrderLastEsSnapshot(orderID int64) (esData EsData, err error) {
	if orderID <= 0 {
		err = fmt.Errorf("wrong orderID: %d", orderID)
		return
	}

	o := orm.NewOrm()
	o.Using(esData.UsingSlave())

	err = o.QueryTable(esData.TableName()).
		Filter("order_id", orderID).
		Filter("es_index", beego.AppConfig.String("es_index")).
		OrderBy("-id").Limit(1).One(&esData)

	return
}
