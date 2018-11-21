package main

import (
	// 数据库初始化
	"fmt"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/models"
)

func main() {
	var orderlist = []models.Order{}
	new(models.Order).GetQuerySeter().All(&orderlist)
	if len(orderlist) > 0 {
		for k, v := range orderlist {
			fmt.Print(k)
			if v.CheckStatus == 8 || v.CheckStatus == 9 || v.CheckStatus == 7 || v.CheckStatus == 11 {
				if v.LoanTime == 0 {
					replan, _ := new(models.RepayPlan).GetLastRepayPlanByOrderid(v.Id)
					if replan.Id != 0 {
						v.LoanTime = replan.Ctime
						//插入数据
						id, _ := new(models.Order).UpdateOrder(&v)
						fmt.Print(id)
					}
				}
			}
			//判断已经结清应还日期为空，添加最后结清日期。
			if v.CheckStatus == 8 {
				if v.RepayTime == 0 {
					v.RepayTime = v.FinishTime
					//插入数据
					id, _ := new(models.Order).UpdateOrder(&v)
					fmt.Print(id, "aaa")
				}
			}

		}

	}

}
