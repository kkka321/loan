package models

import (
	"github.com/astaxie/beego/logs"

	//"github.com/astaxie/beego/logs"

	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const ACCOUNT_QUOTA_CONF_TABLENAME string = "account_quota_conf"

// AccountQuotaConf 表结构
type AccountQuotaConf struct {
	ID            int64 `orm:"pk;column(id)"`
	AccountID     int64 `orm:"column(account_id)"`
	Quota         int64
	QuotaVisable  int64
	AccountPeriod int64
	IsPhoneVerify int64
	Status        int64
	IsDefault     int64
	Ctime         int64
	Utime         int64
}

func (r *AccountQuotaConf) TableName() string {
	return ACCOUNT_QUOTA_CONF_TABLENAME
}

func (r *AccountQuotaConf) Using() string {
	return types.OrmDataBaseApi
}

func (r *AccountQuotaConf) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func OneAccountQuotaConfByAccountID(accountID int64) (AccountQuotaConf, error) {
	var obj = AccountQuotaConf{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	err := o.QueryTable(obj.TableName()).Filter("account_id", accountID).Filter("status", 1).One(&obj)

	return obj, err
}

// InsertDefaultConf 写入默认配置
func InsertDefaultConf(accountID int64) {
	accountQuotaConfInsert := AccountQuotaConf{
		AccountID:     accountID,
		Quota:         600000,
		QuotaVisable:  3000000,
		AccountPeriod: 14,
		IsPhoneVerify: 1,
		Status:        1,
		IsDefault:     1,
		Ctime:         tools.GetUnixMillis(),
		Utime:         tools.GetUnixMillis(),
	}
	num, err := OrmInsert(&accountQuotaConfInsert)
	if num == 0 && err != nil {
		logs.Error("[InsertDefaultConf] account_quota_conf default val insert happend err:", err)
	}
}
