package models

// `admin`
import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const ACCOUNT_BIGDATA_CONTACT_TABLENAME string = "account_bigdata_contact"

type AccountBigdataContact struct {
	Id          int64  `orm:"pk;"`
	AccountID   int64  `orm:"column(account_id)"`
	Mobile      string `orm:"column(mobile)"`
	ContactName string `orm:"column(contact_name)"`
	Ctime       int64  `orm:"column(ctime)"`
	Utime       int64  `orm:"column(utime)"`
	Itime       int64  `orm:"column(itime)"`
	S3key       string `orm:"column(s3key)"`
}

// 此处声明为指针方法,并不会修改传入的对象,只是为了省去拷贝对象的开消

// 当前模型对应的表名
func (r *AccountBigdataContact) TableName() string {
	return ACCOUNT_BIGDATA_CONTACT_TABLENAME
}

// 当前模型的数据库
func (r *AccountBigdataContact) Using() string {
	return types.OrmDataBaseApi
}

func (r *AccountBigdataContact) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func OneAccountBigdataContactByUid(id int64) (AccountBigdataContact, error) {
	obj := AccountBigdataContact{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	err := o.QueryTable(obj.TableName()).Filter("id", id).One(&obj)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneAccountBigdataContactByUid] sql error err:%v", err)
	}
	return obj, err
}

func OneAccountBigdataContactByIM(accountID int64, mobile string) (obj AccountBigdataContact, err error) {
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	err = o.QueryTable(obj.TableName()).Filter("account_id", accountID).Filter("mobile", mobile).One(&obj)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneAccountBigdataContactByUid] sql error err:%v", err)
	}
	return obj, err
}

func OneAccountBigdataContactByAccountID(accountID int64) (objs []AccountBigdataContact, num int64, err error) {
	obj := AccountBigdataContact{}

	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	num, err = o.QueryTable(obj.TableName()).Filter("account_id", accountID).All(&objs)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneAccountBigdataContactByAccountID] sql error err:%v", err)
	}
	return
}
