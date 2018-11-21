package main

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/service"
	"micro-loan/common/tools"

	"micro-loan/common/models"
	"micro-loan/common/types"
)

func main() {
	createDailySelfUrgeOrdersTemp()

	overdueCase := models.OverdueCase{}
	o := orm.NewOrm()
	o.Using(overdueCase.UsingSlave())
	sql := fmt.Sprintf("select * from overdue_case where overdue_days=12 and is_out=0 and join_urge_time<1534608000000")

	var overdueCases []models.OverdueCase
	o.Raw(sql).QueryRows(&overdueCases)
	for _, c := range overdueCases {
		service.HandleOverdueCase(c.OrderId)
	}

}

func createDailySelfUrgeOrdersTemp() {
	data := tempRankingSelfUrgeOrders()
	tempInsertSelfUrgeOrders(data)
}

// RankingSelfUrgeOrders ...
func tempRankingSelfUrgeOrders() (list []models.SelfUrgeOrder) {
	// SELECT oc.order_id, count(IF(promise_repay_time>0,1,null)) AS promise_repay_times FROM overdue_case oc
	// LEFT JOIN overdue_case_detail ocd ON oc.order_id=ocd.order_id
	// WHERE oc.overdue_days=13 AND oc.is_out=0 AND ocd.id>0
	// GROUP BY oc.order_id
	// ORDER BY promise_repay_times DESC
	// LIMIT 40
	nowOverdueDays := types.OverdueLevelCreateDaysMap()[types.OverdueLevelM12] - 1
	selfUrgeNum, _ := config.ValidItemInt64("overdue_edge_self_urge_num")

	obj := models.SelfUrgeOrder{}

	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	var sql = `SELECT oc.order_id, count(IF(promise_repay_time>0,1,null)) AS promise_repay_times FROM overdue_case oc
  LEFT JOIN overdue_case_detail ocd ON oc.order_id=ocd.order_id
  WHERE oc.overdue_days=%d AND oc.is_out=%d AND ocd.id>0 AND oc.join_urge_time<1534608000000
  GROUP BY oc.order_id
  ORDER BY promise_repay_times DESC
  LIMIT %d`
	sql = fmt.Sprintf(sql, nowOverdueDays, 0, selfUrgeNum)
	r := o.Raw(sql)
	r.QueryRows(&list)
	for i, d := range list {
		//
		repayPlan, _ := models.GetLastRepayPlanByOrderid(d.OrderId)
		d.ExpireTime = types.GetOverdueCaseExpireTime(types.MustGetTicketItemIDByCaseName(types.OverdueLevelM12), repayPlan.RepayDate)
		d.Ctime = tools.GetUnixMillis()
		d.Utime = d.Ctime
		list[i] = d
	}
	return
}

func tempInsertSelfUrgeOrders(list []models.SelfUrgeOrder) {
	// TODO check unique or is inserted
	// 否则会因为某条数据的唯一性问题， 导致整体插入失败
	obj := models.SelfUrgeOrder{}

	o := orm.NewOrm()
	o.Using(obj.Using())
	_, err := o.InsertMulti(100, list)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			for _, v := range list {
				o.Insert(&v)
			}
		}
		logs.Error("[insertSelfUrgeOrders] insert multi err", err)
	}
	return
}
