package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	//"micro-loan/common/tools"
	"micro-loan/common/types"
	//"fmt"
)

const RE_LOAN_IMAGE string = "re_loan_image"

type ReLoanImage struct {
	Id            int64 `orm:"pk;"`
	UserAccountId int64 `orm:"column(user_account_id)"`
	ReLoanPhoto   int64
	Ctime         int64
}

// 当前模型对应的表名
func (r *ReLoanImage) TableName() string {
	return RE_LOAN_IMAGE
}

// 当前模型的数据库
func (r *ReLoanImage) Using() string {
	return types.OrmDataBaseApi
}

func (r *ReLoanImage) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func AddReLoanImage(reLoanImage ReLoanImage) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(reLoanImage.Using())
	id, err = o.Insert(&reLoanImage)
	if err != nil {
		logs.Error("model order insert failed.", err)
	}
	return
}
