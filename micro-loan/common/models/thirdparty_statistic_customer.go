package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

// THIRDPARTY_STATISTICS_FEE_TABLENAME 表名
const THIRDPARTY_STATISTICS_CUSTOMER_TABLENAME string = "thirdparty_statistic_customer"

// 	 `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
//   `user_account_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '客户id',
//   `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '产品名称',
//   `mobile` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '电话',
//   `media_source` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用来来源,Organic 为自然增长,为空:未识别,其他为广告如:facebook, google, standard',
//   `campaign` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '渠道中广告详细分类',
//   `api` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '接口地址',
//   `api_md5` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '接口地址的md5值，api字段太长无法作为索引',
//   `api_fee` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '调用费用',
//   `call_count` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '调用次数',
//   `success_call_count` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '调用成功次数',
//   `hit_call_count` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '调用命中次数',
//   `cutomer_total_cost` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户总成本',
//   `record_type` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '1:单个api的统计数据  2：用户总的统计数据',
//   `ctime` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '建表时间',
//   `utime` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',

type ThirdpartyStatisticCustomer struct {
	Id               int64 `orm:"pk;"`
	UserAccountId    int64
	OrderId          int64
	Mobile           string
	Api              string
	ApiMd5           string
	ApiFee           int64
	CallCount        int
	SuccessCallCount int
	HitCallCount     int
	CutomerTotalCost int64
	RecordType       int
	Ctime            int64
	Utime            int64
}

type ThirdpartyStatisticCustomerInfo struct {
	Customer    ThirdpartyStatisticCustomer
	Realname    string
	MediaSource string
	Campaign    string
	Tags        int
}

// TableName 返回当前模型对应的表名
func (r *ThirdpartyStatisticCustomer) TableName() string {
	return THIRDPARTY_STATISTICS_CUSTOMER_TABLENAME
}

// Using 返回当前模型的数据库
func (r *ThirdpartyStatisticCustomer) Using() string {
	return types.OrmDataBaseApi
}

// 当前模型的数据库
func (r *ThirdpartyStatisticCustomer) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

// Add 添加
func (r *ThirdpartyStatisticCustomer) Add() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)
	r.Id = id

	return id, err
}

func (r *ThirdpartyStatisticCustomer) Update(col ...string) (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Update(r, col...)
	return id, err
}

func GetThirdpartyStatisticCustomerByApiMd5AndUId(apiMd5 string, userAccountId int64, recordType int) (one ThirdpartyStatisticCustomer, err error) {

	obj := ThirdpartyStatisticCustomer{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	err = o.QueryTable(obj.TableName()).
		Filter("api_md5", apiMd5).
		Filter("user_account_id", userAccountId).
		Filter("record_type", recordType).
		One(&one)
	return
}
