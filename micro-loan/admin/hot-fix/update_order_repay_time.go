package main

import (
	"fmt"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	//"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"micro-loan/common/models"
)

type OrderFixRepayTime struct {
	OrderId   int64
	Ctime     int64
	RepayTime int64
}

func main() {

	payment := models.Payment{}
	orders := models.Order{}

	var findData []OrderFixRepayTime
	sql := fmt.Sprintf(`select payment.order_id, payment.ctime, orders.repay_time from %s left join %s on payment.order_id = orders.id where (va_company_code = 1 or va_company_code = 2) and pay_type = 1 order by order_id desc`,
		payment.TableName(), orders.TableName())

	o := orm.NewOrm()
	o.Using(payment.Using())
	o.Raw(sql).QueryRows(&findData)

	for _, res := range findData {
		if res.RepayTime == 0 && res.Ctime > 0 {
			order, _ := models.GetOrder(res.OrderId)
			order.RepayTime = res.Ctime
			models.UpdateOrder(&order)
			fmt.Println("Update order id is:", order.Id, "update time is:", res.Ctime)
		}
	}

}
