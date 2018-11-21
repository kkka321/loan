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

type OrderFixLoanTime struct {
	Id        int64
	Ctime     int64
	LoanTime  int64
	RepayDate int64
}

func main() {

	repayPlan := models.RepayPlan{}
	orders := models.Order{}

	var findData []OrderFixLoanTime
	sql := fmt.Sprintf(`select orders.id, orders.loan_time,repay_plan.ctime,repay_plan.repay_date from %s left join %s on repay_plan.order_id = orders.id where repay_plan.order_id = orders.id and orders.loan_time = 0`,
		repayPlan.TableName(), orders.TableName())

	o := orm.NewOrm()
	o.Using(repayPlan.Using())
	o.Raw(sql).QueryRows(&findData)

	for _, res := range findData {
		if res.LoanTime == 0 && res.RepayDate > 0 {
			obj := models.Order{
				Id:       res.Id,
				LoanTime: res.Ctime,
			}
			o.Update(&obj, "loan_time")
			fmt.Println("Update order id is:", res.Id, "update time is:", res.Ctime)
		}
	}

}
