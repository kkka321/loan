package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// THIRDPARTY_STATISTICS_FEE_TABLENAME 表名
const THIRDPARTY_STATISTICS_FEE_TABLENAME string = "thirdparty_statistic_fee"

// ThirdpartyStatisticFee 描述数据表结构与结构体的映射
type ThirdpartyStatisticFee struct {
	Id               int64 `orm:"pk;"`
	Name             string
	Api              string
	ApiMd5           string
	ChargeType       int
	Price            int
	TotalPrice       int64
	CallCount        int
	SuccessCallCount int
	HitCallCount     int
	StatisticDate    int64
	StatisticDateS   string
	Ctime            int64
}

// TableName 返回当前模型对应的表名
func (r *ThirdpartyStatisticFee) TableName() string {
	return THIRDPARTY_STATISTICS_FEE_TABLENAME
}

// Using 返回当前模型的数据库
func (r *ThirdpartyStatisticFee) Using() string {
	return types.OrmDataBaseApi
}

// 当前模型的数据库
func (r *ThirdpartyStatisticFee) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

// Add 添加
func (r *ThirdpartyStatisticFee) Add() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)
	r.Id = id

	return id, err
}

func (r *ThirdpartyStatisticFee) Update(col ...string) (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Update(r, col...)
	return id, err
}

func GetThirdpartyStatisticFeeByApiAndDate(api string, date int64) (one ThirdpartyStatisticFee, err error) {
	obj := ThirdpartyStatisticFee{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	err = o.QueryTable(obj.TableName()).Filter("api_md5", tools.Md5(api)).Filter("statistic_date", date).One(&one)
	return
}

func ListThirdpartyStatisticFee() (list []ThirdpartyStatisticFee, err error) {
	r := ThirdpartyStatisticFee{}
	o := orm.NewOrm()
	o.Using(r.Using())

	_, err = o.QueryTable(r.TableName()).OrderBy("-id").All(&list)

	return nil, err
}
