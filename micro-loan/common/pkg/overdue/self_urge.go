package overdue

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// 对外接口中的 Edge 是指既定边界的单子，或者边界之外的
// 比如， 此处自催的单子并不包含， 初始生成委外订单， 并未把此类订单放入自催中
// 仅仅到达部分委外和部分自催的边界时， 把此边界订单放入自催表

const edgeTicketLevel = types.OverdueLevelM12

// CreateDailySelfUrgeOrders ...
func CreateDailySelfUrgeOrders() {
	data := RankingSelfUrgeOrders()
	insertSelfUrgeOrders(data)
}

// IsEdgeOrBeyond 是否边缘或者超出边缘的逾期订单
//  Edge定义见文件头
func IsEdgeOrBeyond(overdueDays int) bool {
	edgeMinDay := types.OverdueLevelCreateDaysMap()[edgeTicketLevel]
	if overdueDays >= edgeMinDay {
		return true
	}
	return false
}

// EdgeOrderIsSelfUrge 边缘order是否自己催收
// 边缘order， 界限之边， 处于自催和委外之间
func EdgeOrderIsSelfUrge(orderID int64) bool {
	obj := models.SelfUrgeOrder{}

	var num int64

	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE order_id = %d AND is_deleted=%d AND expire_time > %d",
		obj.TableName(), orderID, types.DeletedNo, tools.GetUnixMillis())
	r := o.Raw(sql)
	r.QueryRow(&num)
	if num > 0 {
		return true
	}
	return false
}

// EdgeMultiOrdersFilterSelfUrge 过滤掉准备自催的单子
// 边界是指
func EdgeMultiOrdersFilterSelfUrge(orderIDs []int64) map[int64]bool {
	obj := models.SelfUrgeOrder{}
	idString, err := tools.IntsSliceToWhereInString(orderIDs)
	checkMap := map[int64]bool{}

	if err != nil {
		return checkMap
	}

	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	sql := fmt.Sprintf("SELECT order_id FROM %s WHERE order_id in (%s) AND is_deleted=%d AND expire_time > %d",
		obj.TableName(), idString, types.DeletedNo, tools.GetUnixMillis())
	r := o.Raw(sql)
	var selfUrgeOrders []int64
	r.QueryRows(&selfUrgeOrders)

	for _, v := range selfUrgeOrders {
		checkMap[v] = true
	}
	filteredMap := map[int64]bool{}

	for _, i := range orderIDs {
		if _, ok := checkMap[i]; !ok {
			filteredMap[i] = true
		}
	}

	return filteredMap
}

// RankingSelfUrgeOrders ...
func RankingSelfUrgeOrders() (list []models.SelfUrgeOrder) {
	// SELECT oc.order_id, count(IF(promise_repay_time>0,1,null)) AS promise_repay_times FROM overdue_case oc
	// LEFT JOIN overdue_case_detail ocd ON oc.order_id=ocd.order_id
	// WHERE oc.overdue_days=13 AND oc.is_out=0 AND ocd.id>0
	// GROUP BY oc.order_id
	// ORDER BY promise_repay_times DESC
	// LIMIT 40
	nowOverdueDays := types.OverdueLevelCreateDaysMap()[edgeTicketLevel] - 1
	selfUrgeNum, _ := config.ValidItemInt64("overdue_edge_self_urge_num")

	obj := models.SelfUrgeOrder{}

	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	var sql = `SELECT oc.order_id, count(IF(promise_repay_time>0,1,null)) AS promise_repay_times FROM overdue_case oc
  LEFT JOIN overdue_case_detail ocd ON oc.order_id=ocd.order_id
  WHERE oc.overdue_days=%d AND oc.is_out=%d AND ocd.id>0
  GROUP BY oc.order_id
  ORDER BY promise_repay_times DESC
  LIMIT %d`
	sql = fmt.Sprintf(sql, nowOverdueDays, 0, selfUrgeNum)
	r := o.Raw(sql)
	r.QueryRows(&list)
	for i, d := range list {
		//
		repayPlan, _ := models.GetLastRepayPlanByOrderid(d.OrderId)
		d.ExpireTime = types.GetOverdueCaseExpireTime(types.MustGetTicketItemIDByCaseName(edgeTicketLevel), repayPlan.RepayDate)
		d.Ctime = tools.GetUnixMillis()
		d.Utime = d.Ctime
		list[i] = d
	}
	return
}

func insertSelfUrgeOrders(list []models.SelfUrgeOrder) {
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
