package dao

import (
	"github.com/astaxie/beego/orm"
	"micro-loan/common/models"
)

func GetReduceById(reduceId int64) (one models.ReduceRecordNew, err error) {
	o := orm.NewOrm()
	o.Using(one.UsingSlave())

	err = o.QueryTable(one.TableName()).
		Filter("id", reduceId).
		One(&one)
	return
}
