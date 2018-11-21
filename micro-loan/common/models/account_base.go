package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const ACCOUNT_BASE_TABLENAME string = "account_base"

type AccountBase struct {
	Id                       int64 `orm:"pk;"`
	Mobile                   string
	Password                 string
	Nickname                 string
	Gender                   types.GenderEnum
	Realname                 string
	Identity                 string
	OcrRealname              string `orm:"column(ocr_realname)"`
	OcrIdentity              string `orm:"column(ocr_identity)"`
	AppsflyerID              string `orm:"column(appsflyer_id)"`
	GoogleAdvertisingID      string `orm:"column(google_advertising_id)"`
	ThirdID                  string `orm:"column(third_id)"`
	ThirdName                string `orm:"column(third_name)"`
	ThirdProvince            string `orm:"column(third_province)"`
	ThirdCity                string `orm:"column(third_city)"`
	ThirdDistrict            string `orm:"column(third_district)"`
	ThirdVillage             string `orm:"column(third_village)"`
	Status                   int
	LatestSmsVerifyTime      int64 `orm:"column(latest_sms_verify_time)"`
	OperatorVerifyStatus     int
	OperatorVerifyFinishTime int64
	RegisterTime             int64 `orm:"column(register_time)"`
	LastLoginTime            int64 `orm:"column(last_login_time)"`
	Tags                     types.CustomerTags
	RandomMark               int64
	IsDeleted                int
	StemFrom                 string
	Channel                  string
	PlatformMark             int64
}

func (r *AccountBase) TableName() string {
	return ACCOUNT_BASE_TABLENAME
}

func (r *AccountBase) Using() string {
	return types.OrmDataBaseApi
}

func (r *AccountBase) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func OneAccountBaseByMobile(mobile string) (AccountBase, error) {
	var obj = AccountBase{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	err := o.QueryTable(obj.TableName()).Filter("mobile", mobile).One(&obj)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneAccountBaseByMobile] sql error err:%v", err)
	}

	return obj, err
}

func OneAccountBaseByIdentity(identity string) (AccountBase, error) {
	var obj = AccountBase{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	err := o.QueryTable(obj.TableName()).Filter("identity", identity).OrderBy("-id").One(&obj)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneAccountBaseByMobile] sql error err:%v", err)
	}

	return obj, err
}

func OneAccountBaseByPkId(id int64) (AccountBase, error) {
	var obj = AccountBase{
		Id: id,
	}
	o := orm.NewOrm()
	o.Using(obj.Using())
	err := o.Read(&obj)

	return obj, err
}

func (r *AccountBase) Update(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}

func (r *AccountBase) IsPlatformMark(platform int64) bool {
	return (r.PlatformMark & platform) > 0
}

func (r *AccountBase) SetPlatformMark(platform int64) {
	r.PlatformMark = r.PlatformMark | platform
}

func (r *AccountBase) ClrPlatformMark(platform int64) {
	r.PlatformMark = r.PlatformMark & (^platform)
}

func (r *AccountBase) Delete() (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Delete(r)
	return
}
