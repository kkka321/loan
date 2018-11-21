package entrust

import (
	"fmt"
	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/pkg/overdue"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"

	"github.com/astaxie/beego/orm"
)

// ！！！！目前自动分配这里不需要实现了， 暂时保留，以防以后需要自动分

// GetFilterEntrustOrders 获取可被委外的订单集合
func GetFilterEntrustOrders() (num int64, orders []int64, err error) {

	overdueCase := models.OverdueCase{}
	orderExt := models.OrderExt{}

	o := orm.NewOrm()
	o.Using(overdueCase.UsingSlave())
	sql := fmt.Sprintf(`SELECT oc.order_id FROM %s oc
	LEFT JOIN %s oe on oc.order_id=oe.order_id
	WHERE oc.overdue_days>=%d
	AND oc.is_out = %d
	AND oe.is_entrust = %d`,
		overdueCase.TableName(),
		orderExt.TableName(),
		13,
		types.IsOverdueNo,
		0,
	)
	num, err = o.Raw(sql).QueryRows(&orders)

	return
}

// AssignUrgeOrder 平均分配逾期订单到委外头上
func AssignUrgeOrder() {

	num, orders, err := GetFilterEntrustOrders()
	logs.Debug("[assign] num,orders,err", num, orders, err)
	if err == nil && num > 0 {

		canEntustOrderIDSMapData := overdue.EdgeMultiOrdersFilterSelfUrge(orders)
		canEntrustNumber := len(canEntustOrderIDSMapData)
		logs.Debug("[AssignUrgeOrder]canEntustOrderIDSMapData:", canEntustOrderIDSMapData)
		if canEntrustNumber > 0 {

			N, _ := config.ValidItemInt("entrust_company_number")
			avg := canEntrustNumber / N

			logs.Debug("[AssignUrgeOrder]canEntrustNumber:%d, N:%d, avg:%d", canEntrustNumber, N, avg)

		}

	}

	return
}

func ManualBatchAssign(orderids []int64, pname, auditComment, remark string, isAgree int) (seccusscount int) {

	if len(orderids) > 0 {
		for _, orderID := range orderids {

			orderExt, _ := models.GetOrderExt(orderID)

			//如果重复委外，则跳过
			if orderExt.IsEntrust == 1 && isAgree == 1 {
				continue
			}
			//记录审核数据
			record := models.EntrustApprovalRecord{
				OrderId:      orderID,
				IsAgree:      isAgree,
				AuditComment: auditComment,
				Pname:        pname,
				Remark:       remark,
				Ctime:        tools.GetUnixMillis(),
			}
			id, err := models.OrmInsert(&record)
			if id >= 0 && err == nil {
				unixtime := tools.GetUnixMillis()
				//如果审核通过，关闭工单，原因已委外, 写入待委外数据
				if isAgree == 1 {
					oneCase, _ := dao.GetInOverdueCaseByOrderID(orderID)
					item := types.MustGetTicketItemIDByCaseName(oneCase.CaseLevel)
					ticket.CloseByRelatedID(oneCase.Id, item, types.TicketCloseReasonEntrust)
					//标记待委外
					orderExt.IsEntrust = 0
					orderExt.Utime = unixtime
					orderExt.EntrustPname = pname
					orderExt.Update()
				} else {
					//如果审核拒绝 ，工单状态修改为进行中
					oneCase, _ := dao.GetInOverdueCaseByOrderID(orderID)
					item := types.MustGetTicketItemIDByCaseName(oneCase.CaseLevel)
					ticketModel, _ := models.GetTicketByItemAndRelatedID(item, oneCase.Id)
					ticketModel.Status = types.TicketStatusProccessing
					ticketModel.Utime = tools.GetUnixMillis()
					cols := []string{"status", "utime"}
					models.OrmUpdate(&ticketModel, cols)
				}
				seccusscount++
			}
		}
	}
	return
}
