package ticket

import (
	"fmt"
	"micro-loan/common/i18n"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
)

// 纯后台逻辑放到此文件中
// 此文件方法不会被API逻辑或者task逻辑调用

type CollectionList struct {
	models.Ticket
	//common
	TicketID    int64  `json:"ticket_id"`    //工单ID
	AccountID   int64  `json:"account_id"`   //客户ID
	AccountName string `json:"account_ame"`  //客户姓名
	CoNumber    int    `json:"co_number"`    //催收次数
	LastCoTime  int64  `json:"last_co_time"` //上次催收时间
	LastCoLog   string `json:"last_co_log"`  //上次催收记录
	IsReloan    int    `json:"is_reloan"`    //客户类型-是否复贷
	CaseLevel   string `json:"case_level"`   //客户类型-案件评级

	OrderID int64 `json:"order_id"` //借款ID
	//PTP
	PromiseRepayTime int64 `json:"promise_repay_time"` //承若还款时间
	//OLD complete
	UrgeDay int `json:"urge_day"` //逾期天数
	//complete
	CompleteTime int64 `json:"complete_time"` //完成时间
}

type RmList struct {
	models.Ticket
	//common
	TicketID    int64  `json:"ticket_id"`    //工单ID
	AccountID   int64  `json:"account_id"`   //客户ID
	AccountName string `json:"account_ame"`  //客户姓名
	CoNumber    int    `json:"co_number"`    //催收次数
	LastCoTime  int64  `json:"last_co_time"` //上次催收时间
	LastCoLog   string `json:"last_co_log"`  //上次催收记录
	IsReloan    int    `json:"is_reloan"`    //客户类型-是否复贷

	PhoneConnect    int `json:"phone_connect"`    //拨打状态
	UnconnectReason int `json:"unconnect_reason"` //未接通原因

	OrderID int64 `json:"order_id"` //借款ID
	//complete
	CompleteTime int64 `json:"complete_time"` //完成时间
}
type PVList struct {
	models.Ticket

	IsReloan         int   `json:"is_reloan"`          //客户类型-是否复贷
	OrderID          int64 `json:"order_id"`           //借款ID
	CheckTime        int64 `json:"check_time"`         //审批时间
	PartCompleteTime int64 `json:"part_complete_time"` //部分完成时间
	CompleteTime     int64 `json:"complete_time"`      //完成时间

}

// CollectionListBackend 返回
func CollectionListBackend(condCntr map[string]interface{}, roleID int64, rolePid int64, page int, pagesize int, sortField, sort, action string) (collectionlist []CollectionList, total int64, err error) {
	allowSortKeys := map[string]bool{
		"id":               true,
		"handle_num":       true,
		"last_handle_time": true,
		"next_handle_time": true,
		"case_level":       true,
	}
	var orderBy string
	if _, ok := allowSortKeys[sortField]; ok && len(sort) > 0 {
		orderBy = sortField + " " + sort
	} else {
		orderBy = "`id` desc"
	}

	obj := models.Ticket{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = types.DefaultPagesize
	}
	offset := (page - 1) * pagesize

	// 初始化查询条件
	where := whereCollectionBackend(condCntr, roleID, rolePid, action)

	var list []models.Ticket
	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` %s", models.TICKET_TABLENAME, where)

	sqlList := fmt.Sprintf("SELECT * FROM `%s` %s ORDER BY %s  LIMIT %d,%d", models.TICKET_TABLENAME, where, orderBy, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	if len(list) > 0 {
		collection := CollectionList{}
		for _, ticket := range list {
			overdueCase, _ := models.OneOverueCaseByPkId(ticket.RelatedID)
			order, err := models.GetOrder(overdueCase.OrderId)
			if err != nil {
				continue
			}
			accountBase, _ := models.OneAccountBaseByPkId(order.UserAccountId)
			collection.Id = ticket.Id
			collection.Status = ticket.Status
			collection.Link = ticket.Link
			collection.TicketID = ticket.Id               //int64  `json:"ticket_id"`        //工单ID
			collection.AccountID = accountBase.Id         //int64  `json:"account_id"`       //客户ID
			collection.AccountName = accountBase.Realname //string `json:"account_ame"`      //客户姓名
			collection.CoNumber = ticket.HandleNum        //int    `json:"co_number"`        //催收次数
			collection.LastCoTime = overdueCase.UrgeTime  //int64  `json:"last_co_time"`     //上次催收时间
			collection.LastCoLog = overdueCase.Result     //string `json:"last_co_log"`      //上次催收记录
			collection.IsReloan = order.IsReloan
			collection.CaseLevel = ticket.CaseLevel
			collection.IsEmptyNumber = ticket.IsEmptyNumber
			collection.OrderID = order.Id //int64  `json:"order_id"`         //借款ID
			collection.LastHandleTime = ticket.LastHandleTime
			//PTP
			collection.PromiseRepayTime = ticket.NextHandleTime //int64 `json:"promise_repay_time"` //承若还款时间
			//OLD complete
			collection.UrgeDay = overdueCase.OverdueDays //int `json:"urge_day"` //逾期天数
			//complete
			collection.CompleteTime = ticket.CompleteTime //int64 `json:"complete_time"` //完成时间

			collectionlist = append(collectionlist, collection)
		}

	}

	return
}

// RmListBackend 返回
func RmListBackend(condCntr map[string]interface{}, roleID int64, rolePid int64, page int, pagesize int, sortField, sort, action string) (rmList []RmList, total int64, err error) {
	allowSortKeys := map[string]bool{
		"id":               true,
		"handle_num":       true,
		"last_handle_time": true,
		"next_handle_time": true,
	}
	var orderBy string
	if _, ok := allowSortKeys[sortField]; ok && len(sort) > 0 {
		orderBy = sortField + " " + sort
	} else {
		orderBy = "`id` desc"
	}

	obj := models.Ticket{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = types.DefaultPagesize
	}
	offset := (page - 1) * pagesize

	// 初始化查询条件
	where := whereCollectionBackend(condCntr, roleID, rolePid, action)

	var list []models.Ticket
	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` %s", models.TICKET_TABLENAME, where)

	sqlList := fmt.Sprintf("SELECT * FROM `%s` %s ORDER BY %s  LIMIT %d,%d", models.TICKET_TABLENAME, where, orderBy, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	if len(list) > 0 {
		rm := RmList{}
		for _, ticket := range list {

			order, err := models.GetOrder(ticket.OrderID)
			if err != nil {
				continue
			}
			repayremindLog, _ := models.GetOneLastRepayRemindLogByOrderID(order.Id)
			accountBase, _ := models.OneAccountBaseByPkId(order.UserAccountId)
			rm.Id = ticket.Id
			rm.Status = ticket.Status
			rm.Link = ticket.Link
			rm.TicketID = ticket.Id               //int64  `json:"ticket_id"`        //工单ID
			rm.AccountID = accountBase.Id         //int64  `json:"account_id"`       //客户ID
			rm.AccountName = accountBase.Realname //string `json:"account_ame"`      //客户姓名
			rm.CoNumber = ticket.HandleNum        //int    `json:"co_number"`        //催收次数
			// rm.LastCoTime = overdueCase.UrgeTime  //int64  `json:"last_co_time"`     //上次催收时间
			rm.LastCoLog = repayremindLog.Result //string `json:"last_co_log"`      //上次催收记录
			rm.IsReloan = order.IsReloan
			rm.CaseLevel = ticket.CaseLevel
			rm.IsEmptyNumber = ticket.IsEmptyNumber
			rm.NextHandleTime = ticket.NextHandleTime
			rm.LastHandleTime = ticket.LastHandleTime
			rm.OrderID = order.Id //int64  `json:"order_id"`         //借款ID
			rm.UnconnectReason = repayremindLog.UnconnectReason
			rm.PhoneConnect = repayremindLog.PhoneConnect
			//complete
			rm.CompleteTime = ticket.CompleteTime //int64 `json:"complete_time"` //完成时间

			rmList = append(rmList, rm)
		}

	}

	return
}

// RmListBackend 返回
func PVAndInfoReviewListBackend(condCntr map[string]interface{}, roleID int64, rolePid int64, page int, pagesize int, sortField, sort, action string) (pvList []PVList, total int64, err error) {
	allowSortKeys := map[string]bool{
		"id":               true,
		"handle_num":       true,
		"last_handle_time": true,
		"next_handle_time": true,
	}
	var orderBy string
	if _, ok := allowSortKeys[sortField]; ok && len(sort) > 0 {
		orderBy = sortField + " " + sort
	} else {
		orderBy = "`id` desc"
	}

	obj := models.Ticket{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = types.DefaultPagesize
	}
	offset := (page - 1) * pagesize

	// 初始化查询条件
	where := wherePVBackend(condCntr, roleID, rolePid, action)

	var list []models.Ticket
	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` %s", models.TICKET_TABLENAME, where)

	sqlList := fmt.Sprintf("SELECT * FROM `%s` %s ORDER BY %s  LIMIT %d,%d", models.TICKET_TABLENAME, where, orderBy, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	if len(list) > 0 {
		pv := PVList{}
		for _, ticket := range list {

			order, err := models.GetOrder(ticket.OrderID)
			if err != nil {
				continue
			}
			pv.Id = ticket.Id
			pv.ItemID = ticket.ItemID
			pv.OrderID = order.Id
			pv.IsReloan = order.IsReloan
			pv.CheckTime = order.CheckTime
			pv.LastHandleTime = ticket.LastHandleTime
			pv.NextHandleTime = ticket.NextHandleTime
			pv.HandleNum = ticket.HandleNum
			pv.Ctime = ticket.Ctime
			pv.AssignTime = ticket.AssignTime
			pv.Link = ticket.Link
			pv.CompleteTime = ticket.CompleteTime
			pv.PartialCompleteTime = ticket.PartialCompleteTime
			pvList = append(pvList, pv)
		}

	}

	return
}

//针对电核 info review
func wherePVBackend(condCntr map[string]interface{}, roleID, rolePid int64, action string) string {
	// 初始化查询条件
	cond := []string{}
	//co
	if action == "new" {
		cond = append(cond, fmt.Sprintf("status in(%d,%d)", types.TicketStatusAssigned, types.TicketStatusProccessing))
	}
	if action == "part" {
		cond = append(cond, fmt.Sprintf("status in(%d)", types.TicketStatusPartialCompleted))
	}

	if action == "complete" {
		cond = append(cond, fmt.Sprintf("status in(%d)", types.TicketStatusCompleted))
	}
	if v, ok := condCntr["id"]; ok {
		cond = append(cond, fmt.Sprintf("id=%d", v))
	}
	if v, ok := condCntr["order_id"]; ok {
		cond = append(cond, fmt.Sprintf("order_id=%d", v))
	}
	if v, ok := condCntr["item_id"]; ok {
		if itemID, ok := v.(types.TicketItemEnum); ok {
			cond = append(cond, fmt.Sprintf("item_id=%d", itemID))
		} else if itemIDMap, ok := v.(map[types.TicketItemEnum]string); ok {
			// for manager
			var itemIDs []string
			for k := range itemIDMap {
				itemIDs = append(itemIDs, strconv.Itoa(int(k)))
			}
			cond = append(cond, fmt.Sprintf("item_id in(%s)", strings.Join(itemIDs, ",")))

		}
	}
	if v, ok := condCntr["assign_uid"]; ok {
		// it means already
		cond = append(cond, fmt.Sprintf("assign_uid=%d", v))
	} else {
		// for manager page
		if rolePid != types.RoleSuperPid {
			var uids []string
			subUserNums, _ := models.GetUserIDsByRolePidFromDB(roleID, &uids)
			if subUserNums == 0 {
				uids = append(uids, "")
			}
			cond = append(cond, fmt.Sprintf("assign_uid in(%s)", strings.Join(uids, ",")))
		}
	}
	if v, ok := condCntr["customer_id"]; ok {
		cond = append(cond, fmt.Sprintf("customer_id=%d", v))
	}
	if v, ok := condCntr["complete_time_start"]; ok {
		cond = append(cond, fmt.Sprintf("complete_time>=%d", v))
	}
	if v, ok := condCntr["complete_time_end"]; ok {
		cond = append(cond, fmt.Sprintf("complete_time<%d", v))
	}
	if v, ok := condCntr["last_urge_time_start"]; ok {
		cond = append(cond, fmt.Sprintf("last_handle_time>=%d", v))
	}
	if v, ok := condCntr["last_urge_time_end"]; ok {
		cond = append(cond, fmt.Sprintf("last_handle_time<%d", v))
	}

	if v, ok := condCntr["op_uid"]; ok {
		cond = append(cond, fmt.Sprintf("assign_uid=%d", v))
	}
	if len(cond) > 0 {
		return "WHERE " + strings.Join(cond, " AND ")
	}
	return ""
}

func whereCollectionBackend(condCntr map[string]interface{}, roleID, rolePid int64, action string) string {
	// 初始化查询条件
	cond := []string{}
	//co
	if action == "new" {
		todayStr := time.Now().AddDate(0, 0, 0).Format("2006-01-02")
		todayBegin := todayStr + " 00:00:00"
		todayEnd := todayStr + " 23:59:59"
		startDate := tools.GetDateParseBackends(todayBegin) * 1000
		endDate := tools.GetDateParseBackends(todayEnd) * 1000
		cond = append(cond, fmt.Sprintf("assign_time>=%d and assign_time<=%d", startDate, endDate))
		cond = append(cond, fmt.Sprintf("status in(%d,%d,%d)", types.TicketStatusAssigned, types.TicketStatusProccessing, types.TicketStatusPartialCompleted))
	}
	if action == "ptp" {

		cond = append(cond, fmt.Sprintf("status in(%d,%d,%d)", types.TicketStatusAssigned, types.TicketStatusProccessing, types.TicketStatusPartialCompleted))
		cond = append(cond, fmt.Sprintf("next_handle_time > %d", 0))
	}
	if action == "old" {
		todayStr := time.Now().AddDate(0, 0, 0).Format("2006-01-02")
		todayBegin := todayStr + " 00:00:00"
		startDate := tools.GetDateParseBackends(todayBegin) * 1000
		cond = append(cond, fmt.Sprintf("assign_time<%d", startDate))
		cond = append(cond, fmt.Sprintf("status in(%d,%d,%d)", types.TicketStatusAssigned, types.TicketStatusProccessing, types.TicketStatusPartialCompleted))
		cond = append(cond, fmt.Sprintf("next_handle_time = %d", 0))
	}
	if action == "complete" {
		cond = append(cond, fmt.Sprintf("status in(%d)", types.TicketStatusCompleted))
	}
	if v, ok := condCntr["id"]; ok {
		cond = append(cond, fmt.Sprintf("id=%d", v))
	}
	if v, ok := condCntr["overdue_days"]; ok {
		if int64v, ok := v.(int64); ok {
			cond = append(cond, fmt.Sprintf("should_repay_date=%d", tools.NaturalDay(-int64v)))
		}
	}
	if v, ok := condCntr["order_id"]; ok {
		cond = append(cond, fmt.Sprintf("order_id=%d", v))
	}
	if v, ok := condCntr["item_id"]; ok {
		if itemID, ok := v.(types.TicketItemEnum); ok {
			cond = append(cond, fmt.Sprintf("item_id=%d", itemID))
		} else if itemIDMap, ok := v.(map[types.TicketItemEnum]string); ok {
			// for manager
			var itemIDs []string
			for k := range itemIDMap {
				itemIDs = append(itemIDs, strconv.Itoa(int(k)))
			}
			cond = append(cond, fmt.Sprintf("item_id in(%s)", strings.Join(itemIDs, ",")))

		}
	}
	if v, ok := condCntr["assign_uid"]; ok {
		// it means already
		cond = append(cond, fmt.Sprintf("assign_uid=%d", v))
	} else {
		// for manager page
		if rolePid != types.RoleSuperPid {
			var uids []string
			subUserNums, _ := models.GetUserIDsByRolePidFromDB(roleID, &uids)
			if subUserNums == 0 {
				uids = append(uids, "")
			}
			cond = append(cond, fmt.Sprintf("assign_uid in(%s)", strings.Join(uids, ",")))
		}
	}
	if v, ok := condCntr["customer_id"]; ok {
		cond = append(cond, fmt.Sprintf("customer_id=%d", v))
	}
	if v, ok := condCntr["complete_time_start"]; ok {
		cond = append(cond, fmt.Sprintf("complete_time>=%d", v))
	}
	if v, ok := condCntr["complete_time_end"]; ok {
		cond = append(cond, fmt.Sprintf("complete_time<%d", v))
	}
	if v, ok := condCntr["last_urge_time_start"]; ok {
		cond = append(cond, fmt.Sprintf("last_handle_time>=%d", v))
	}
	if v, ok := condCntr["last_urge_time_end"]; ok {
		cond = append(cond, fmt.Sprintf("last_handle_time<%d", v))
	}

	if v, ok := condCntr["op_uid"]; ok {
		cond = append(cond, fmt.Sprintf("assign_uid=%d", v))
	}
	if len(cond) > 0 {
		return "WHERE " + strings.Join(cond, " AND ")
	}
	return ""
}

// ListBackend 返回
func ListBackend(condCntr map[string]interface{}, roleID int64, rolePid int64, page int, pagesize int, sortField, sort string) (list []models.Ticket, total int64, err error) {
	allowSortKeys := map[string]bool{
		"id":               true,
		"last_handle_time": true,
		"next_handle_time": true,
	}
	var orderBy string
	if _, ok := allowSortKeys[sortField]; ok && len(sort) > 0 {
		orderBy = sortField + " " + sort
	} else {
		orderBy = "`id` desc"
	}

	obj := models.Ticket{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = types.DefaultPagesize
	}
	offset := (page - 1) * pagesize

	// 初始化查询条件
	where := whereBackend(condCntr, roleID, rolePid)

	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` %s", models.TICKET_TABLENAME, where)
	sqlList := fmt.Sprintf("SELECT * FROM `%s` %s ORDER BY %s LIMIT %d,%d", models.TICKET_TABLENAME, where, orderBy, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

func whereBackend(condCntr map[string]interface{}, roleID, rolePid int64) string {
	// 初始化查询条件
	cond := []string{}
	if v, ok := condCntr["id"]; ok {
		cond = append(cond, fmt.Sprintf("id=%d", v))
	}
	if v, ok := condCntr["order_id"]; ok {
		cond = append(cond, fmt.Sprintf("order_id=%d", v))
	}
	if v, ok := condCntr["status"]; ok {
		cond = append(cond, fmt.Sprintf("status in(%s)", strings.Join(v.([]string), ",")))
	}
	if v, ok := condCntr["risk_level"]; ok {
		cond = append(cond, fmt.Sprintf("risk_level=%d", v))
	}
	if v, ok := condCntr["item_id"]; ok {
		if itemID, ok := v.(types.TicketItemEnum); ok {
			cond = append(cond, fmt.Sprintf("item_id=%d", itemID))
		} else if itemIDMap, ok := v.(map[types.TicketItemEnum]string); ok {
			// for manager
			var itemIDs []string
			for k := range itemIDMap {
				itemIDs = append(itemIDs, strconv.Itoa(int(k)))
			}
			cond = append(cond, fmt.Sprintf("item_id in(%s)", strings.Join(itemIDs, ",")))

		}
	}
	if v, ok := condCntr["assign_uid"]; ok {
		// it means already
		cond = append(cond, fmt.Sprintf("assign_uid=%d", v))
	} else {
		// for manager page
		if rolePid != types.RoleSuperPid {
			var uids []string
			subUserNums, _ := models.GetUserIDsByRolePidFromDB(roleID, &uids)
			if subUserNums == 0 {
				uids = append(uids, "")
			}
			cond = append(cond, fmt.Sprintf("assign_uid in(%s)", strings.Join(uids, ",")))
		}
	}

	if v, ok := condCntr["related_id"]; ok {
		cond = append(cond, fmt.Sprintf("related_id=%d", v))
	}

	if v, ok := condCntr["ctime_start"]; ok {
		cond = append(cond, fmt.Sprintf("ctime>=%d", v))
	}
	if v, ok := condCntr["ctime_end"]; ok {
		cond = append(cond, fmt.Sprintf("ctime<%d", v))
	}
	if v, ok := condCntr["complete_time_start"]; ok {
		cond = append(cond, fmt.Sprintf("complete_time>=%d", v))
	}
	if v, ok := condCntr["complete_time_end"]; ok {
		cond = append(cond, fmt.Sprintf("complete_time<%d", v))
	}
	if v, ok := condCntr["close_time_start"]; ok {
		cond = append(cond, fmt.Sprintf("close_time>=%d", v))
	}
	if v, ok := condCntr["close_time_end"]; ok {
		cond = append(cond, fmt.Sprintf("close_time<%d", v))
	}
	if v, ok := condCntr["op_uid"]; ok {
		cond = append(cond, fmt.Sprintf("assign_uid=%d", v))
	}
	if len(cond) > 0 {
		return "WHERE " + strings.Join(cond, " AND ")
	}
	return ""
}

// DisplayWorkerCanAcceptTicket 后台展示, 工作人员是否可接收工单
func DisplayWorkerCanAcceptTicket(lang string, adminUID int64) string {
	if IsWorkerOnline(adminUID) {
		return i18n.T(lang, "是")
	}
	return i18n.T(lang, "否")
}

// GetRiskLevelDisplay 返回风险评级定义
func GetRiskLevelDisplay(lang string, riskLevel int) string {
	if val, ok := riskLevelMap[riskLevel]; ok {
		return i18n.T(lang, val)
	}
	return i18n.T(lang, "未定义")
}
