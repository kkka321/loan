package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// SALE_INVITE_TABLENAME 表名
const SALE_INVITE_TABLENAME string = "sale_invite"

// SALE_INVITE_TABLENAME 描述数据表结构与结构体的映射
type SaleInvite struct {
	Id           int64 `orm:"pk;"`
	AccountId    int64
	ShareUrl     string
	AnonymousUrl string
	InviterId    int64
	Ctime        int64
	Utime        int64
}

// TableName 返回当前模型对应的表名
func (r *SaleInvite) TableName() string {
	return SALE_INVITE_TABLENAME
}

// Using 返回当前模型的数据库
func (r *SaleInvite) Using() string {
	return types.OrmDataBaseApi
}

func (r *SaleInvite) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func GetSaleInviteById(id int64) (data SaleInvite, err error) {
	obj := SaleInvite{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	qs := o.QueryTable(obj.TableName())

	err = qs.Filter("id", id).One(&data)

	return
}

func GetSaleInvite(accountId int64) (data SaleInvite, err error) {
	obj := SaleInvite{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	qs := o.QueryTable(obj.TableName())

	err = qs.Filter("account_id", accountId).One(&data)

	return
}

func (r *SaleInvite) Insert() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err := o.Insert(r)
	return err
}

func (r *SaleInvite) Update() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err := o.Update(r)

	return err
}
