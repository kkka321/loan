package models

// `admin`
import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const ACCOUNT_MOBILE_HISTORY_TABLENAME string = "account_mobile_history"

type AccountMobileHistory struct {
	Id        int64  `orm:"pk;"`                // 主键id
	AccountId int64  `orm:"column(account_id)"` // 账号ID
	Mobile    string `orm:"column(mobile)"`     // 手机号
	Ctime     int64  `orm:"column(ctime)"`      // 添加时间
	Utime     int64  `orm:"column(utime)"`      // 更新时间
}

// 当前模型对应的表名
func (r *AccountMobileHistory) TableName() string {
	return ACCOUNT_MOBILE_HISTORY_TABLENAME
}

// 当前模型的数据库
func (r *AccountMobileHistory) Using() string {
	return types.OrmDataBaseApi
}

func (r *AccountMobileHistory) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *AccountMobileHistory) Insert() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

func (r *AccountMobileHistory) Updates(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}

// GetAccountMobileModifyNum 根据用户ID获取用户修改过的手机号次数
func GetAccountMobileModifyNum(accountID int64) (num int64, err error) {
	r := AccountMobileHistory{}
	o := orm.NewOrm()
	o.Using(r.UsingSlave())

	num, err = o.QueryTable(r.TableName()).Filter("account_id", accountID).Count()
	if err != nil {
		logs.Error("[GetAccountMobileModifyNum] err:", err)
	}

	return
}
