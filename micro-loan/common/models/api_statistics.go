package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// API_STATISTICS_TABLENAME 表名
const API_STATISTICS_TABLENAME string = "api_statistics"

// OrderStatistics 描述数据表结构与结构体的映射
type ApiStatistics struct {
	Id             int64 `orm:"pk;"`
	RequestUrl     string
	ConsumeTime    int64
	StatisticsDate string
	Ctime          int64
}

// TableName 返回当前模型对应的表名
func (r *ApiStatistics) TableName() string {
	return API_STATISTICS_TABLENAME
}

// Using 返回当前模型的数据库
func (r *ApiStatistics) Using() string {
	return types.OrmDataBaseApi
}

func (r *ApiStatistics) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

// Add 添加新的权限
func (r *ApiStatistics) Add() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.Ctime = tools.GetUnixMillis()

	id, err := o.Insert(r)
	r.Id = id

	return id, err
}
