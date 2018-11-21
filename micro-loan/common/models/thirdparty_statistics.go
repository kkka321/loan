package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// THIRDPARTY_STATISTICS_TABLENAME 表名
const THIRDPARTY_STATISTICS_TABLENAME string = "thirdparty_statistics"

// ThirdpartyStatistics 描述数据表结构与结构体的映射
type ThirdpartyStatistics struct {
	Id             int64 `orm:"pk;"`
	Thirdparty     int
	Success        int
	Fail           int
	StatisticsDate string
	Ctime          int64
}

// TableName 返回当前模型对应的表名
func (r *ThirdpartyStatistics) TableName() string {
	return THIRDPARTY_STATISTICS_TABLENAME
}

// Using 返回当前模型的数据库
func (r *ThirdpartyStatistics) Using() string {
	return types.OrmDataBaseApi
}

func (r *ThirdpartyStatistics) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

// Add 添加新的权限
func (r *ThirdpartyStatistics) Add() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.Ctime = tools.GetUnixMillis()

	id, err := o.Insert(r)
	r.Id = id

	return id, err
}

// Del 添加新的权限
func (r *ThirdpartyStatistics) Del() error {
	o := orm.NewOrm()
	o.Using(r.Using())

	_, err := o.Delete(r)

	return err
}
