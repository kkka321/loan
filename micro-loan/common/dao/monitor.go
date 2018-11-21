package dao

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/models"
)

func OrderStatisticsByDate(date string) (models.OrderStatistics, error) {
	var order models.OrderStatistics = models.OrderStatistics{}

	o := orm.NewOrm()
	o.Using(order.Using())

	err := o.QueryTable(order.TableName()).Filter("statistics_date", date).One(&order)

	return order, err
}

func ThirdpartyStatisticsByDate(date string) (list []*models.ThirdpartyStatistics, err error) {
	var s models.ThirdpartyStatistics = models.ThirdpartyStatistics{}

	o := orm.NewOrm()
	o.Using(s.Using())

	o.QueryTable(s.TableName()).Filter("statistics_date", date).All(&list)

	return
}