package main

import (
	"flag"
	"fmt"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/models"
	"micro-loan/common/tools"

	"github.com/astaxie/beego/orm"
)

var isRun bool
var runNum int

func init() {
	flag.BoolVar(&isRun, "run", false, "whether Run")
	flag.IntVar(&runNum, "run-num", 0, "Run num")
	flag.Parse()
}

func main() {
	// 获取待修复数据
	var actualUpdate int64
	ticketM := models.Ticket{}
	limit := 500
	sql := fmt.Sprintf("select id,order_id from ticket where item_id>2 and ctime>%d and should_repay_date=0 order by id ASC limit %d",
		tools.NaturalDay(-13), limit,
	)

	o := orm.NewOrm()
	o.Using(ticketM.Using())
	var updateNum int
	for {
		var updateTickets []models.Ticket
		o.Raw(sql).QueryRows(&updateTickets)
		if len(updateTickets) == 0 {
			break
		}
		for _, tm := range updateTickets {
			if updateNum >= runNum {
				fmt.Printf("Update will stop by that runNum(%d) is run out,updateNum:%d \n", runNum, updateNum)
				fmt.Println("updated data:", actualUpdate)
				return
			}
			rp, _ := models.GetLastRepayPlanByOrderid(tm.OrderID)
			tm.ShouldRepayDate = rp.RepayDate
			if isRun {
				aff, _ := models.OrmUpdate(&tm, []string{"ShouldRepayDate"})
				actualUpdate += aff
			}
			updateNum++
		}
		if !isRun {
			break
		}
	}
	fmt.Println("updated data:", actualUpdate)

}
