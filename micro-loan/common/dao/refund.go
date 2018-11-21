package dao

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/models"
)

func OneAccountBalanceByAccountId(id int64) (one models.AccountBalance, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).
		Filter("account_id", id).
		One(&one)
	return
}

func GetRefund(refundId int64) (one models.Refund, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).
		Filter("id", refundId).
		One(&one)
	return
}
