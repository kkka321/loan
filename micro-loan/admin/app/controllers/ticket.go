package controllers

import (
	"encoding/json"
	"fmt"
	"micro-loan/common/i18n"
	"micro-loan/common/models"
	"micro-loan/common/pkg/admin"
	"micro-loan/common/pkg/rbac"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/pkg/ticket/performance"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils/pagination"
)

// TicketController 所有menu相关的控制器入口
type TicketController struct {
	BaseController
}

// Prepare 进入Action前的逻辑
func (c *TicketController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["SuperAdminUID"] = types.SuperAdminUID
}

// List 列表
func (c *TicketController) ManageList() {
	var condCntr = map[string]interface{}{}

	// 获取指定
	c.Data["TicketStatusCreated"] = types.TicketStatusCreated
	c.Data["TicketStatusAssigned"] = types.TicketStatusAssigned
	c.Data["TicketStatusProccessing"] = types.TicketStatusProccessing
	c.Data["TicketStatusCompleted"] = types.TicketStatusCompleted
	c.Data["TicketStatusClosed"] = types.TicketStatusClosed
	c.Data["ticketStatusMap"] = types.TicketStatusMap()
	c.Data["RiskLevelMap"] = ticket.RiskLevelMap()
	c.Data["ticketItemMap"] = types.OwnTicketItemMap(c.RoleType)

	itemID, _ := c.GetInt("item_id")
	if itemID > 0 {
		condCntr["item_id"] = types.TicketItemEnum(itemID)
	} else {
		condCntr["item_id"] = c.Data["ticketItemMap"]
	}
	c.Data["itemID"] = types.TicketItemEnum(itemID)
	c.Layout = "layout.html"
	c.TplName = "ticket/list.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "ticket/list_scripts.html"

	isSubmit := c.GetString("submit")
	riskLevel, _ := c.GetInt("risk_level")
	c.Data["riskLevel"] = riskLevel
	if riskLevel > 0 {
		condCntr["risk_level"] = riskLevel
	}
	if isSubmit != "true" {
		return
	}

	//role, _ := rbac.GetOneRole(c.RoleID)

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
		c.Data["id"] = id
	}

	orderID, _ := c.GetInt64("order_id")
	if orderID > 0 {
		condCntr["order_id"] = orderID
		c.Data["orderID"] = orderID
	}

	statusList := c.GetStrings("status")
	if len(statusList) > 0 {
		condCntr["status"] = statusList
		selectedStatusMap := map[interface{}]interface{}{}
		for _, status := range statusList {
			validStatus, _ := strconv.Atoi(status)
			selectedStatusMap[types.TicketStatusEnum(validStatus)] = nil
		}
		c.Data["selectedStatusMap"] = selectedStatusMap
	}

	relatedID, _ := c.GetInt("related_id")
	if relatedID > 0 {
		condCntr["related_id"] = relatedID
		c.Data["relatedID"] = relatedID
	}

	ctimeRange := c.GetString("ctime_range")
	c.Data["ctimeRange"] = ctimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(ctimeRange); err == nil {
		condCntr["ctime_start"], condCntr["ctime_end"] = start, end
	}

	completeTimeRange := c.GetString("complete_time_range")
	c.Data["completeTimeRange"] = completeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(completeTimeRange); err == nil {
		condCntr["complete_time_start"], condCntr["complete_time_end"] = start, end
	}

	closeTimeRange := c.GetString("close_time_range")
	c.Data["closeTimeRange"] = closeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(closeTimeRange); err == nil {
		condCntr["close_time_start"], condCntr["close_time_end"] = start, end
	}
	assignUID, _ := c.GetInt64("assign_uid")
	if assignUID > 0 {
		condCntr["assign_uid"] = assignUID
		c.Data["assignUID"] = assignUID
	} else {
		opName := c.GetString("op_name")
		if len(opName) > 0 {
			c.Data["opName"] = opName
			admin, _ := models.OneAdminByNickName(opName)
			if admin.Id == 0 {
				//直接赋值-1000是当前条件不成立
				condCntr["assign_uid"] = -1000
			} else {
				condCntr["assign_uid"] = admin.Id
			}
		}
	}

	field := c.GetString("field", "id")
	sort := c.GetString("sort", "desc")

	page, _ := tools.Str2Int(c.GetString("p"))
	maxSize := 30
	pagesize, _ := c.GetInt("size", 15)
	if pagesize > maxSize {
		pagesize = maxSize
	}

	c.Data["size"] = pagesize

	list, count, _ := ticket.ListBackend(condCntr, c.RoleID, c.RolePid, page, pagesize, field, sort)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list
	// 获取指定
	c.Data["TicketStatusCreated"] = types.TicketStatusCreated
	c.Data["TicketStatusAssigned"] = types.TicketStatusAssigned
	c.Data["TicketStatusProccessing"] = types.TicketStatusProccessing
	c.Data["TicketStatusCompleted"] = types.TicketStatusCompleted
	c.Data["TicketStatusClosed"] = types.TicketStatusClosed
	c.Data["TicketStatusApplyEntrust"] = types.TicketStatusWaitingEntrust
	c.Data["ticketStatusMap"] = types.TicketStatusMap()
	c.Data["RiskLevelMap"] = ticket.RiskLevelMap()

}

func (c *TicketController) Me() {
	var condCntr = map[string]interface{}{}
	condCntr["assign_uid"] = c.AdminUid
	c.Data["ticketItemMap"] = types.OwnTicketItemMap(c.RoleType)

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
	}
	c.Data["id"] = id

	orderID, _ := c.GetInt64("order_id")
	if orderID > 0 {
		condCntr["order_id"] = orderID
	}
	c.Data["orderID"] = orderID

	itemID, _ := c.GetInt("item_id")
	c.Data["itemID"] = types.TicketItemEnum(itemID)
	if itemID > 0 {
		condCntr["item_id"] = types.TicketItemEnum(itemID)
	} else {
		condCntr["item_id"] = c.Data["ticketItemMap"]
	}

	riskLevel, _ := c.GetInt("risk_level")
	c.Data["riskLevel"] = riskLevel
	if riskLevel > 0 {
		condCntr["risk_level"] = riskLevel
	}

	statusList := c.GetStrings("status")
	if len(statusList) > 0 {
		condCntr["status"] = statusList
		selectedStatusMap := map[interface{}]interface{}{}
		for _, status := range statusList {
			validStatus, _ := strconv.Atoi(status)
			selectedStatusMap[types.TicketStatusEnum(validStatus)] = nil
		}
		c.Data["selectedStatusMap"] = selectedStatusMap
	} else {
		if c.GetString("is_search") != "1" {
			condCntr["status"] = []string{strconv.Itoa(int(types.TicketStatusAssigned)), strconv.Itoa(int(types.TicketStatusProccessing))}
			c.Data["selectedStatusMap"] = map[interface{}]interface{}{
				types.TicketStatusAssigned:         nil,
				types.TicketStatusProccessing:      nil,
				types.TicketStatusPartialCompleted: nil,
			}
		}
	}

	relatedID, _ := c.GetInt("related_id")
	c.Data["relatedID"] = relatedID
	if relatedID > 0 {
		condCntr["related_id"] = relatedID
	}

	ctimeRange := c.GetString("ctime_range")
	c.Data["ctimeRange"] = ctimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(ctimeRange); err == nil {
		condCntr["ctime_start"], condCntr["ctime_end"] = start, end
	}

	completeTimeRange := c.GetString("complete_time_range")
	c.Data["completeTimeRange"] = completeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(completeTimeRange); err == nil {
		condCntr["complete_time_start"], condCntr["complete_time_end"] = start, end
	}

	closeTimeRange := c.GetString("close_time_range")
	c.Data["closeTimeRange"] = closeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(closeTimeRange); err == nil {
		condCntr["close_time_start"], condCntr["close_time_end"] = start, end
	}

	field := c.GetString("field", "last_handle_time")
	sort := c.GetString("sort", "ASC")

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	list, count, _ := ticket.ListBackend(condCntr, c.RoleID, c.RolePid, page, pagesize, field, sort)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list
	c.Data["IsWorkerAcceptTicket"] = ticket.IsWorkerOnline(c.AdminUid)
	// 获取指定
	c.Data["TicketStatusCreated"] = types.TicketStatusCreated
	c.Data["TicketStatusAssigned"] = types.TicketStatusAssigned
	c.Data["TicketStatusProccessing"] = types.TicketStatusProccessing
	c.Data["TicketStatusCompleted"] = types.TicketStatusCompleted
	c.Data["TicketStatusClosed"] = types.TicketStatusClosed
	c.Data["ticketStatusMap"] = types.TicketStatusMap()
	c.Data["RiskLevelMap"] = ticket.RiskLevelMap()
	//
	todayStartstamp, _ := tools.GetTodayTimestampByLocalTime("00:00")
	c.Data["TodayStartTime"] = todayStartstamp * 1000
	c.Layout = "layout.html"
	c.TplName = "ticket/me.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "ticket/list_scripts.html"
}

func (c *TicketController) Collection() {
	var condCntr = map[string]interface{}{}
	condCntr["assign_uid"] = c.AdminUid

	action := c.GetString("action", "new")
	c.Data["action"] = action

	c.Data["ticketItemMap"] = types.OwnTicketItemMap(c.RoleType)

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
	}
	c.Data["id"] = id

	overdueDays, _ := c.GetInt64("overdue_days")
	if overdueDays > 0 {
		condCntr["overdue_days"] = overdueDays
		c.Data["overdueDays"] = overdueDays
	}

	orderID, _ := c.GetInt64("order_id")
	if orderID > 0 {
		condCntr["order_id"] = orderID
		c.Data["orderID"] = orderID
	}

	assignUID, _ := c.GetInt64("assign_uid")
	if assignUID > 0 {
		condCntr["assign_uid"] = assignUID
		c.Data["assignUID"] = assignUID
	} else {
		opName := c.GetString("op_name")
		if len(opName) > 0 {
			c.Data["opName"] = opName
			admin, _ := models.OneAdminByNickName(opName)
			if admin.Id == 0 {
				//直接赋值-1000是当前条件不成立
				condCntr["assign_uid"] = -1000
			} else {
				condCntr["assign_uid"] = admin.Id
			}
		}
	}

	mobile := c.GetString("mobile")
	if len(mobile) > 0 {
		c.Data["mobile"] = mobile
		accountBase, _ := models.OneAccountBaseByMobile(mobile)
		if accountBase.Id > 0 {
			condCntr["customer_id"] = accountBase.Id
		} else {
			condCntr["assign_uid"] = -1000
		}
	}

	// itemID, _ := c.GetInt("item_id")
	// c.Data["itemID"] = types.TicketItemEnum(itemID)
	// if itemID > 0 {
	// 	condCntr["item_id"] = types.TicketItemEnum(itemID)
	// } else {
	// 	condCntr["item_id"] = c.Data["ticketItemMap"]
	// }
	condCntr["item_id"] = map[types.TicketItemEnum]string{
		types.TicketItemUrgeM11: "",
		types.TicketItemUrgeM12: "",
	}

	riskLevel, _ := c.GetInt("risk_level")
	c.Data["riskLevel"] = riskLevel
	if riskLevel > 0 {
		condCntr["risk_level"] = riskLevel
	}

	statusList := c.GetStrings("status")
	if len(statusList) > 0 {
		condCntr["status"] = statusList
		selectedStatusMap := map[interface{}]interface{}{}
		for _, status := range statusList {
			validStatus, _ := strconv.Atoi(status)
			selectedStatusMap[types.TicketStatusEnum(validStatus)] = nil
		}
		c.Data["selectedStatusMap"] = selectedStatusMap
	} else {
		if c.GetString("is_search") != "1" {
			condCntr["status"] = []string{strconv.Itoa(int(types.TicketStatusAssigned)), strconv.Itoa(int(types.TicketStatusProccessing))}
			c.Data["selectedStatusMap"] = map[interface{}]interface{}{
				types.TicketStatusAssigned:         nil,
				types.TicketStatusProccessing:      nil,
				types.TicketStatusPartialCompleted: nil,
			}
		}
	}

	relatedID, _ := c.GetInt("related_id")
	c.Data["relatedID"] = relatedID
	if relatedID > 0 {
		condCntr["related_id"] = relatedID
	}

	ctimeRange := c.GetString("ctime_range")
	c.Data["ctimeRange"] = ctimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(ctimeRange); err == nil {
		condCntr["ctime_start"], condCntr["ctime_end"] = start, end
	}

	completeTimeRange := c.GetString("complete_time_range")
	c.Data["completeTimeRange"] = completeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(completeTimeRange); err == nil {
		condCntr["complete_time_start"], condCntr["complete_time_end"] = start, end
	}

	lostUrgeTimeRange := c.GetString("last_urge_time_range")
	c.Data["lostUrgeTimeRange"] = lostUrgeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(lostUrgeTimeRange); err == nil {
		condCntr["last_urge_time_start"], condCntr["last_urge_time_end"] = start, end
	}

	field := c.GetString("field", "last_handle_time")
	sort := c.GetString("sort", "ASC")
	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 30

	list, count, _ := ticket.CollectionListBackend(condCntr, c.RoleID, c.RolePid, page, pagesize, field, sort, action)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list
	c.Data["IsWorkerAcceptTicket"] = ticket.IsWorkerOnline(c.AdminUid)
	// 获取指定
	c.Data["TicketStatusCreated"] = types.TicketStatusCreated
	c.Data["TicketStatusAssigned"] = types.TicketStatusAssigned
	c.Data["TicketStatusProccessing"] = types.TicketStatusProccessing
	c.Data["TicketStatusCompleted"] = types.TicketStatusCompleted
	c.Data["TicketStatusClosed"] = types.TicketStatusClosed
	c.Data["ticketStatusMap"] = types.TicketStatusMap()
	c.Data["RiskLevelMap"] = ticket.RiskLevelMap()
	//
	todayStartstamp, _ := tools.GetTodayTimestampByLocalTime("00:00")
	c.Data["TodayStartTime"] = todayStartstamp * 1000
	c.Layout = "layout.html"
	c.TplName = "ticket/collection.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "ticket/list_scripts.html"
	c.LayoutSections["CssPlugin"] = "ticket/collection.css.html"
}

func (c *TicketController) RmTicket() {
	var condCntr = map[string]interface{}{}
	condCntr["assign_uid"] = c.AdminUid

	action := c.GetString("action", "new")

	c.Data["action"] = action

	c.Data["ticketItemMap"] = types.OwnTicketItemMap(c.RoleType)

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
	}
	c.Data["id"] = id

	orderID, _ := c.GetInt64("order_id")
	if orderID > 0 {
		condCntr["order_id"] = orderID
		c.Data["orderID"] = orderID
	}

	assignUID, _ := c.GetInt64("assign_uid")
	if assignUID > 0 {
		condCntr["assign_uid"] = assignUID
		c.Data["assignUID"] = assignUID
	} else {
		opName := c.GetString("op_name")
		if len(opName) > 0 {
			c.Data["opName"] = opName
			admin, _ := models.OneAdminByNickName(opName)
			if admin.Id == 0 {
				//直接赋值-1000是当前条件不成立
				condCntr["assign_uid"] = -1000
			} else {
				condCntr["assign_uid"] = admin.Id
			}
		}
	}

	mobile := c.GetString("mobile")
	if len(mobile) > 0 {
		c.Data["mobile"] = mobile
		accountBase, _ := models.OneAccountBaseByMobile(mobile)
		if accountBase.Id > 0 {
			condCntr["customer_id"] = accountBase.Id
		} else {
			condCntr["assign_uid"] = -1000
		}
	}

	// itemID, _ := c.GetInt("item_id")
	// c.Data["itemID"] = types.TicketItemEnum(itemID)
	// if itemID > 0 {
	// 	condCntr["item_id"] = types.TicketItemEnum(itemID)
	// } else {
	// 	condCntr["item_id"] = c.Data["ticketItemMap"]
	// }
	condCntr["item_id"] = map[types.TicketItemEnum]string{
		types.TicketItemRM0: "",
	}

	riskLevel, _ := c.GetInt("risk_level")
	c.Data["riskLevel"] = riskLevel
	if riskLevel > 0 {
		condCntr["risk_level"] = riskLevel
	}

	statusList := c.GetStrings("status")
	if len(statusList) > 0 {
		condCntr["status"] = statusList
		selectedStatusMap := map[interface{}]interface{}{}
		for _, status := range statusList {
			validStatus, _ := strconv.Atoi(status)
			selectedStatusMap[types.TicketStatusEnum(validStatus)] = nil
		}
		c.Data["selectedStatusMap"] = selectedStatusMap
	} else {
		if c.GetString("is_search") != "1" {
			condCntr["status"] = []string{strconv.Itoa(int(types.TicketStatusAssigned)), strconv.Itoa(int(types.TicketStatusProccessing))}
			c.Data["selectedStatusMap"] = map[interface{}]interface{}{
				types.TicketStatusAssigned:         nil,
				types.TicketStatusProccessing:      nil,
				types.TicketStatusPartialCompleted: nil,
			}
		}
	}

	relatedID, _ := c.GetInt("related_id")
	c.Data["relatedID"] = relatedID
	if relatedID > 0 {
		condCntr["related_id"] = relatedID
	}

	ctimeRange := c.GetString("ctime_range")
	c.Data["ctimeRange"] = ctimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(ctimeRange); err == nil {
		condCntr["ctime_start"], condCntr["ctime_end"] = start, end
	}

	completeTimeRange := c.GetString("complete_time_range")
	c.Data["completeTimeRange"] = completeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(completeTimeRange); err == nil {
		condCntr["complete_time_start"], condCntr["complete_time_end"] = start, end
	}

	lostUrgeTimeRange := c.GetString("last_urge_time_range")
	c.Data["lostUrgeTimeRange"] = lostUrgeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(lostUrgeTimeRange); err == nil {
		condCntr["last_urge_time_start"], condCntr["last_urge_time_end"] = start, end
	}

	field := c.GetString("field", "last_handle_time")
	sort := c.GetString("sort", "ASC")
	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 30

	list, count, _ := ticket.RmListBackend(condCntr, c.RoleID, c.RolePid, page, pagesize, field, sort, action)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list
	c.Data["IsWorkerAcceptTicket"] = ticket.IsWorkerOnline(c.AdminUid)
	// 获取指定
	c.Data["TicketStatusCreated"] = types.TicketStatusCreated
	c.Data["TicketStatusAssigned"] = types.TicketStatusAssigned
	c.Data["TicketStatusProccessing"] = types.TicketStatusProccessing
	c.Data["TicketStatusCompleted"] = types.TicketStatusCompleted
	c.Data["TicketStatusClosed"] = types.TicketStatusClosed
	c.Data["ticketStatusMap"] = types.TicketStatusMap()
	c.Data["RiskLevelMap"] = ticket.RiskLevelMap()
	//
	todayStartstamp, _ := tools.GetTodayTimestampByLocalTime("00:00")
	c.Data["TodayStartTime"] = todayStartstamp * 1000
	c.Layout = "layout.html"
	c.TplName = "ticket/rmticket.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "ticket/list_scripts.html"
	c.LayoutSections["CssPlugin"] = "ticket/collection.css.html"
}

//PvAndInfoReview 电核和InfoReview工单
func (c *TicketController) PvAndInfoReview() {
	var condCntr = map[string]interface{}{}
	condCntr["assign_uid"] = c.AdminUid

	action := c.GetString("action", "new")

	c.Data["action"] = action

	c.Data["ticketItemMap"] = types.OwnTicketItemMap(c.RoleType)

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
		c.Data["id"] = id
	}

	orderID, _ := c.GetInt64("order_id")
	if orderID > 0 {
		condCntr["order_id"] = orderID
		c.Data["orderID"] = orderID
	}

	assignUID, _ := c.GetInt64("assign_uid")
	if assignUID > 0 {
		condCntr["assign_uid"] = assignUID
		c.Data["assignUID"] = assignUID
	} else {
		opName := c.GetString("op_name")
		if len(opName) > 0 {
			c.Data["opName"] = opName
			admin, _ := models.OneAdminByNickName(opName)
			if admin.Id == 0 {
				//直接赋值-1000是当前条件不成立
				condCntr["assign_uid"] = -1000
			} else {
				condCntr["assign_uid"] = admin.Id
			}
		}
	}

	mobile := c.GetString("mobile")
	if len(mobile) > 0 {
		c.Data["mobile"] = mobile
		accountBase, _ := models.OneAccountBaseByMobile(mobile)
		if accountBase.Id > 0 {
			condCntr["customer_id"] = accountBase.Id
		} else {
			condCntr["assign_uid"] = -1000
		}
	}

	// itemID, _ := c.GetInt("item_id")
	// c.Data["itemID"] = types.TicketItemEnum(itemID)
	// if itemID > 0 {
	// 	condCntr["item_id"] = types.TicketItemEnum(itemID)
	// } else {
	// 	condCntr["item_id"] = c.Data["ticketItemMap"]
	// }
	condCntr["item_id"] = map[types.TicketItemEnum]string{
		types.TicketItemPhoneVerify: "",
		types.TicketItemInfoReview:  "",
	}

	riskLevel, _ := c.GetInt("risk_level")
	c.Data["riskLevel"] = riskLevel
	if riskLevel > 0 {
		condCntr["risk_level"] = riskLevel
	}

	statusList := c.GetStrings("status")
	if len(statusList) > 0 {
		condCntr["status"] = statusList
		selectedStatusMap := map[interface{}]interface{}{}
		for _, status := range statusList {
			validStatus, _ := strconv.Atoi(status)
			selectedStatusMap[types.TicketStatusEnum(validStatus)] = nil
		}
		c.Data["selectedStatusMap"] = selectedStatusMap
	} else {
		if c.GetString("is_search") != "1" {
			condCntr["status"] = []string{strconv.Itoa(int(types.TicketStatusAssigned)), strconv.Itoa(int(types.TicketStatusProccessing))}
			c.Data["selectedStatusMap"] = map[interface{}]interface{}{
				types.TicketStatusAssigned:         nil,
				types.TicketStatusProccessing:      nil,
				types.TicketStatusPartialCompleted: nil,
			}
		}
	}

	relatedID, _ := c.GetInt("related_id")
	c.Data["relatedID"] = relatedID
	if relatedID > 0 {
		condCntr["related_id"] = relatedID
	}

	ctimeRange := c.GetString("ctime_range")
	c.Data["ctimeRange"] = ctimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(ctimeRange); err == nil {
		condCntr["ctime_start"], condCntr["ctime_end"] = start, end
	}

	completeTimeRange := c.GetString("complete_time_range")
	c.Data["completeTimeRange"] = completeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(completeTimeRange); err == nil {
		condCntr["complete_time_start"], condCntr["complete_time_end"] = start, end
	}

	lostUrgeTimeRange := c.GetString("last_urge_time_range")
	c.Data["lostUrgeTimeRange"] = lostUrgeTimeRange
	if start, end, err := tools.PareseDateRangeToMillsecond(lostUrgeTimeRange); err == nil {
		condCntr["last_urge_time_start"], condCntr["last_urge_time_end"] = start, end
	}

	field := c.GetString("field", "id")
	sort := c.GetString("sort", "DESC")
	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 30

	list, count, _ := ticket.PVAndInfoReviewListBackend(condCntr, c.RoleID, c.RolePid, page, pagesize, field, sort, action)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list
	c.Data["IsWorkerAcceptTicket"] = ticket.IsWorkerOnline(c.AdminUid)
	// 获取指定
	c.Data["TicketStatusCreated"] = types.TicketStatusCreated
	c.Data["TicketStatusAssigned"] = types.TicketStatusAssigned
	c.Data["TicketStatusProccessing"] = types.TicketStatusProccessing
	c.Data["TicketStatusCompleted"] = types.TicketStatusCompleted
	c.Data["TicketStatusClosed"] = types.TicketStatusClosed
	c.Data["ticketStatusMap"] = types.TicketStatusMap()
	c.Data["RiskLevelMap"] = ticket.RiskLevelMap()
	//
	todayStartstamp, _ := tools.GetTodayTimestampByLocalTime("00:00")
	c.Data["TodayStartTime"] = todayStartstamp * 1000
	c.Layout = "layout.html"
	c.TplName = "ticket/pv_inforeview.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "ticket/list_scripts.html"
	c.LayoutSections["CssPlugin"] = "ticket/collection.css.html"
}

func (c *TicketController) AssignPage() {
	c.TplName = "ticket/assign.html"

	id, err := c.GetInt64("id")
	if err != nil || id <= 0 {
		c.Data["error"] = i18n.T(c.LangUse, "Request is invalid")
		return
	}
	admins, num, err := ticket.CanAssignUsers(id)
	var filterAdmins []models.Admin

	roleLevel := rbac.GetRoleLevel(c.RoleID)
	if roleLevel == types.RoleSuper {
		filterAdmins = admins
	} else if roleLevel == types.RoleLeader {
		ids := admin.GetLeaderManageUsers(c.RoleID, c.AdminUid)
		idsMap := tools.SliceInt64ToMap(ids)
		for _, m := range admins {
			if _, ok := idsMap[m.Id]; ok {
				filterAdmins = append(filterAdmins, m)
			}
		}
	}
	c.Data["admins"] = filterAdmins

	if err != nil || num <= 0 {
		c.Data["error"] = i18n.T(c.LangUse, "No invalid admin users to assign")
		return
	}
	c.Data["id"] = id
}

func (c *TicketController) Assign() {
	id, idErr := c.GetInt64("id")
	assignUID, uidErr := c.GetInt64("assign_uid")
	if idErr != nil || uidErr != nil || id <= 0 || assignUID <= 0 {
		c.Data["json"] = map[string]interface{}{
			"error": "Invaild Request",
		}
		c.ServeJSON()
		return
	}

	result, err := ticket.ManualAssign(id, assignUID, c.AdminUid)

	if !result {
		c.Data["json"] = map[string]interface{}{"error": err.Error()}
	} else {
		c.Data["json"] = map[string]interface{}{"status": true}
	}
	c.ServeJSON()
	return
}

func (c *TicketController) BatchAssignPage() {
	c.TplName = "ticket/batch_assign.html"

	itemID, err := c.GetInt64("item_id")
	if err != nil || itemID <= 0 {
		c.Data["error"] = i18n.T(c.LangUse, "Request is invalid")
		return
	}
	admins, num, err := ticket.CanAssignUsersByTicketItem(types.TicketItemEnum(itemID))
	if err != nil || num <= 0 {
		c.Data["error"] = i18n.T(c.LangUse, "No invalid admin users to assign")
		return
	}

	var filterAdmins []models.Admin

	roleLevel := rbac.GetRoleLevel(c.RoleID)
	if roleLevel == types.RoleSuper {
		filterAdmins = admins
	} else if roleLevel == types.RoleLeader {
		ids := admin.GetLeaderManageUsers(c.RoleID, c.AdminUid)
		idsMap := tools.SliceInt64ToMap(ids)
		for _, m := range admins {
			if _, ok := idsMap[m.Id]; ok {
				filterAdmins = append(filterAdmins, m)
			}
		}
	}
	c.Data["admins"] = filterAdmins
	c.Data["itemID"] = itemID
	c.Data["ticketName"] = types.TicketItemMap()[types.TicketItemEnum(itemID)]
}

func (c *TicketController) BatchAssign() {
	idsString := c.GetString("ids")
	idStrings := strings.Split(idsString, ",")
	assignUID, uidErr := c.GetInt64("assign_uid")
	if len(idStrings) == 0 || assignUID <= 0 || uidErr != nil {
		c.Data["json"] = map[string]interface{}{
			"error": "Invaild Request",
		}
		c.ServeJSON()
		return
	}

	var ids []int64
	for _, v := range idStrings {
		iv, err := strconv.ParseInt(v, 10, 64)
		if err == nil && iv > 0 {
			ids = append(ids, iv)
		}
	}

	result := ticket.ManualBatchAssign(ids, assignUID, c.AdminUid)
	c.Data["json"] = map[string]interface{}{"result": fmt.Sprintf("Want assign ticket num: %d, actual assign num: %d", len(idStrings), result)}
	c.ServeJSON()
	return
}

func (c *TicketController) BatchApplyEntrust() {
	idsString := c.GetString("ids")
	idStrings := strings.Split(idsString, ",")
	if len(idStrings) == 0 {
		c.Data["json"] = map[string]interface{}{
			"error": "Invaild Request",
		}
		c.ServeJSON()
		return
	}

	var ids []int64
	for _, v := range idStrings {
		iv, err := strconv.ParseInt(v, 10, 64)
		if err == nil && iv > 0 {
			ids = append(ids, iv)
		}
	}
	result := ticket.BatchApplyEntrust(ids)
	c.Data["json"] = map[string]interface{}{"result": fmt.Sprintf("Batch apply total num: %d, success num: %d", len(idStrings), result)}
	c.ServeJSON()
	return
}

func (c *TicketController) UpdateStatus() {
	id, err := c.GetInt64("id")
	action := c.GetString("action")
	if err != nil || id <= 0 || len(action) <= 0 {
		c.Data["json"] = map[string]interface{}{
			"error": "Invaild Request",
		}
		c.ServeJSON()
		return
	}

	ticketModel, err := models.GetTicket(id)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"error": "Invaild Request",
		}
		c.ServeJSON()
		return
	}
	var result bool
	switch action {
	case "start":
		result, err = ticket.StartByTicketModel(&ticketModel)
		break
	case "close":
		result, err = ticket.CloseByTicketModel(&ticketModel, types.TicketCloseReasonAbnormal)
		break
	default:
		err = fmt.Errorf("Invaild action:%s", action)
	}
	if result != true {
		if err != nil {
			c.Data["json"] = map[string]interface{}{"error": err}
		} else {
			c.Data["json"] = map[string]interface{}{"error": "Unknown error"}
		}
	} else {
		c.Data["json"] = map[string]interface{}{"status": true}
	}
	c.ServeJSON()
	return
}

func (c *TicketController) WorkerManage() {
	// var condCntr = map[string]interface{}{}
	c.Layout = "layout.html"
	c.TplName = "ticket/worker_manage.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "ticket/worker_manage_scripts.html"

	var condCntr = map[string]interface{}{}

	opName := strings.TrimSpace(c.GetString("op_name"))
	if len(opName) > 0 {
		c.Data["opName"] = opName
		admin, _ := models.OneAdminByNickName(opName)
		if admin.Id == 0 {
			//直接赋值-1000是当前条件不成立
			condCntr["op_uid"] = -1000
		} else {
			condCntr["op_uid"] = admin.Id
		}
	}

	status, _ := c.GetInt("status", -1)

	c.Data["status"] = status
	if status != -1 {
		condCntr["status"] = status
	}

	var list []admin.Worker
	var count int64

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15
	if c.AdminUid == types.SuperAdminUID || c.RoleType == types.RoleTypeSystem {
		list, count, _ = admin.GetUsersByType(0, condCntr, page, pagesize)
	} else {
		if c.RolePid == types.RoleSuperPid {
			list, count, _ = admin.GetUsersByType(c.RoleType, condCntr, page, pagesize)
		} else if rbac.GetRoleLevel(c.RoleID) == types.RoleLeader {
			condCntr["leader_role_id"] = c.RoleID
			condCntr["leader_user_id"] = c.AdminUid
			list, count, _ = admin.GetUsersByType(c.RoleType, condCntr, page, pagesize)
		}
	}

	for i := range list {
		list[i].OnlineStatus = ticket.IsWorkerOnline(list[i].Id)
	}
	c.Data["List"] = list

	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)
	c.Data["paginator"] = paginator

}

//AjaxModifyReducedQuota 修改某字段
func (c *TicketController) AjaxModifyReducedQuota() {

	mapData := make(map[string]interface{})
	mapData["data"] = false

	var Obj struct {
		ID    string
		Field string
		Value string
	}
	jsonStr := c.GetString("jsonStr")
	if err := json.Unmarshal([]byte(jsonStr), &Obj); err != nil {
		panic(err)
	}
	//验证
	isValidName := tools.IsNumber(Obj.Value)

	if isValidName {
		id, _ := tools.Str2Int64(Obj.ID)
		val, _ := tools.Str2Int(Obj.Value)
		adminModel, _ := models.OneAdminByUid(id)
		origin := adminModel
		if adminModel.Id > 0 {
			adminModel.ReducedQuota = val
			models.Update(adminModel)
		}
		// 写操作日志
		models.OpLogWrite(c.AdminUid, adminModel.Id, models.OpCodeAccountBaseUpdate, adminModel.TableName(), origin, adminModel)
		mapData["data"] = true
	} else {
		//配合不合法
		mapData["error"] = 1
	}

	c.Data["json"] = &mapData
	c.ServeJSON()

}

func (c *TicketController) UpdateWorkerStatus() {
	id, err := c.GetInt64("id")
	action := c.GetString("action")

	if err != nil || id <= 0 || len(action) <= 0 {
		c.Data["json"] = map[string]interface{}{
			"error": "Invaild Request",
		}
		c.ServeJSON()
		return
	}
	oldData, _ := models.OneAdminByUid(id)
	modelData := models.Admin{}
	modelData.Id, _ = c.GetInt64("id")
	modelData.RoleID = oldData.RoleID

	var result bool
	switch action {
	case "start":
		modelData.WorkStatus = types.AdminWorkStatusNormal
		admin.Update(&modelData, &oldData, []string{"WorkStatus"})

		if todayStartTime, _ := tools.GetTodayTimestampByLocalTime("00:00"); oldData.LastLoginTime > todayStartTime {
			ticket.ManualWorkerOnline(id, c.AdminUid)
		}
		result = true
		break
	case "stop":
		modelData.WorkStatus = types.AdminWorkStatusStop
		admin.Update(&modelData, &oldData, []string{"WorkStatus"})
		if ticket.IsWorkerOnline(id) {
			ticket.ManualWorkerOffline(id, c.AdminUid)
		}
		result = true
		break
	default:
		err = fmt.Errorf("Invaild action:%s", action)
	}
	if result != true {
		if err != nil {
			c.Data["json"] = map[string]interface{}{"error": err}
		} else {
			c.Data["json"] = map[string]interface{}{"error": "Unknown error"}
		}
	} else {
		c.Data["json"] = map[string]interface{}{"status": true}
	}
	c.ServeJSON()
	return
}

func (c *TicketController) UpdateWorkerOnlineStatus() {
	id, err := c.GetInt64("id")
	action := c.GetString("action")

	if err != nil || id <= 0 || len(action) <= 0 {
		c.Data["json"] = map[string]interface{}{
			"error": "Invaild Request",
		}
		c.ServeJSON()
		return
	}
	_, err = models.OneAdminByUid(id)
	if err != nil {
		logs.Error("[UpdateWorkerOnlineStatus]", err)
		c.Data["json"] = map[string]interface{}{
			"error": "Invaild Request",
		}
		c.ServeJSON()
		return
	}

	switch action {
	case "online":
		ticket.ManualWorkerOnline(id, c.AdminUid)
	case "offline":
		ticket.ManualWorkerOffline(id, c.AdminUid)

	default:
		err = fmt.Errorf("Invaild action:%s", action)
	}
	if err != nil {
		c.Data["json"] = map[string]interface{}{"error": err}
	} else {
		c.Data["json"] = map[string]interface{}{"status": true}
	}

	c.ServeJSON()
	return
}

func (c *TicketController) UpdateMyOnlineStatus() {
	id := c.AdminUid
	action := c.GetString("action")

	_, err := models.OneAdminByUid(id)
	if err != nil {
		logs.Error("[UpdateWorkerOnlineStatus]", err)
		c.Data["json"] = map[string]interface{}{
			"error": "Invaild Request",
		}
		c.ServeJSON()
		return
	}

	switch action {
	case "online":
		ticket.ManualWorkerOnline(id, c.AdminUid)
	case "offline":
		ticket.ManualWorkerOffline(id, c.AdminUid)
	default:
		err = fmt.Errorf("Invaild action:%s", action)
	}
	if err != nil {
		c.Data["json"] = map[string]interface{}{"error": err}
	} else {
		c.Data["json"] = map[string]interface{}{"status": true}
	}

	c.ServeJSON()
	return
}

func (c *TicketController) PerformanceMe() {

	var condCntr = map[string]interface{}{}

	//role, _ := rbac.GetOneRole(c.RoleID)
	c.Data["ticketItemMap"] = types.OwnTicketItemMap(c.RoleType)
	condCntr["admin_uid"] = c.AdminUid

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
	}
	c.Data["id"] = id

	itemID, _ := c.GetInt("item_id")
	c.Data["itemID"] = types.TicketItemEnum(itemID)
	if itemID > 0 {
		condCntr["ticket_item_id"] = types.TicketItemEnum(itemID)
	} else {
		condCntr["ticket_item_id"] = c.Data["ticketItemMap"]
	}

	dateRange := c.GetString("date_range")
	c.Data["dateRange"] = dateRange
	if start, end, err := tools.PareseDateRangeToDayRange(dateRange); err == nil {
		condCntr["date_start"], condCntr["date_end"] = start, end
	}

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	list, count, totalStats, _ := performance.WorkerStatsListBackend(condCntr, c.RoleID, c.AdminUid, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["totalStats"] = totalStats
	c.Data["List"] = list
	// 获取指定

	c.Layout = "layout.html"
	c.TplName = "ticket/performance_me.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "ticket/performance_list_scripts.html"
}

func (c *TicketController) PerformanceManagement() {
	var condCntr = map[string]interface{}{}

	//role, _ := rbac.GetOneRole(c.RoleID)
	c.Data["ticketItemMap"] = types.OwnTicketItemMap(c.RoleType)

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
	}
	c.Data["id"] = id

	itemID, _ := c.GetInt("item_id")
	c.Data["itemID"] = types.TicketItemEnum(itemID)
	if itemID > 0 {
		condCntr["ticket_item_id"] = types.TicketItemEnum(itemID)
	} else {
		condCntr["ticket_item_id"] = c.Data["ticketItemMap"]
	}

	opName := strings.TrimSpace(c.GetString("op_name"))
	if len(opName) > 0 {
		c.Data["opName"] = opName
		admin, _ := models.OneAdminByNickName(opName)
		if admin.Id == 0 {
			//直接赋值-1000是当前条件不成立
			condCntr["admin_uid"] = -1000
		} else {
			condCntr["admin_uid"] = admin.Id
		}

	}

	dateRange := c.GetString("date_range", tools.GetDefaultDateRange(-1, -1))
	c.Data["dateRange"] = dateRange
	if start, end, err := tools.PareseDateRangeToDayRange(dateRange); err == nil {
		condCntr["date_start"], condCntr["date_end"] = start, end
	}

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	list, count, totalStats, _ := performance.WorkerStatsListBackend(condCntr, c.RoleID, c.RolePid, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["totalStats"] = totalStats
	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "ticket/performance_list.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "ticket/performance_list_scripts.html"
}

func (c *TicketController) DailyPerformanceExport() {
	var condCntr = map[string]interface{}{}

	//role, _ := rbac.GetOneRole(c.RoleID)
	c.Data["ticketItemMap"] = types.OwnTicketItemMap(c.RoleType)

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
	}
	c.Data["id"] = id

	itemID, _ := c.GetInt("item_id")
	c.Data["itemID"] = types.TicketItemEnum(itemID)
	if itemID > 0 {
		condCntr["ticket_item_id"] = types.TicketItemEnum(itemID)
	} else {
		condCntr["ticket_item_id"] = c.Data["ticketItemMap"]
	}

	opName := strings.TrimSpace(c.GetString("op_name"))
	if len(opName) > 0 {
		c.Data["opName"] = opName
		admin, _ := models.OneAdminByNickName(opName)
		if admin.Id == 0 {
			//直接赋值-1000是当前条件不成立
			condCntr["admin_uid"] = -1000
		} else {
			condCntr["admin_uid"] = admin.Id
		}
	}

	dateRange := c.GetString("date_range", tools.GetDefaultDateRange(-1, -1))
	c.Data["dateRange"] = dateRange
	if start, end, err := tools.PareseDateRangeToDayRange(dateRange); err == nil {
		condCntr["date_start"], condCntr["date_end"] = start, end
	}

	list, _, _, _ := performance.WorkerStatsListExportBackend(condCntr, c.RoleID, c.AdminUid)

	fileName := fmt.Sprintf("stats_%d.xlsx", tools.GetUnixMillis())
	lang := c.LangUse
	xlsx := excelize.NewFile()
	xlsx.SetCellValue("Sheet1", "A1", i18n.T(lang, "日期"))
	xlsx.SetCellValue("Sheet1", "B1", i18n.T(lang, "工单分类"))
	xlsx.SetCellValue("Sheet1", "C1", i18n.T(lang, "排名"))
	xlsx.SetCellValue("Sheet1", "D1", i18n.T(lang, "分配给"))
	xlsx.SetCellValue("Sheet1", "E1", i18n.T(lang, "分案本金"))
	xlsx.SetCellValue("Sheet1", "F1", i18n.T(lang, "回款本金"))
	xlsx.SetCellValue("Sheet1", "G1", i18n.T(lang, "回款息费"))
	xlsx.SetCellValue("Sheet1", "H1", i18n.T(lang, "回款总金额"))
	xlsx.SetCellValue("Sheet1", "I1", i18n.T(lang, "回收率"))
	xlsx.SetCellValue("Sheet1", "J1", i18n.T(lang, "目标回收率"))
	xlsx.SetCellValue("Sheet1", "K1", i18n.T(lang, "差值金额"))
	xlsx.SetCellValue("Sheet1", "L1", i18n.T(lang, "新分配数"))
	xlsx.SetCellValue("Sheet1", "M1", i18n.T(lang, "处理数"))
	xlsx.SetCellValue("Sheet1", "N1", i18n.T(lang, "完成数"))
	xlsx.SetCellValue("Sheet1", "O1", i18n.T(lang, "负载数"))

	for i, d := range list {
		xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(i+2), d.Date)
		xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(i+2), service.GetTicketItemDisplay(lang, d.TicketItemID))
		xlsx.SetCellValue("Sheet1", "C"+strconv.Itoa(i+2), d.Ranking)
		xlsx.SetCellValue("Sheet1", "D"+strconv.Itoa(i+2), admin.OperatorName(d.AdminUID))
		xlsx.SetCellValue("Sheet1", "E"+strconv.Itoa(i+2), d.LoadLeftUnpaidPrincipal)
		xlsx.SetCellValue("Sheet1", "F"+strconv.Itoa(i+2), d.RepayPrincipal)
		xlsx.SetCellValue("Sheet1", "G"+strconv.Itoa(i+2), d.RepayInterest)
		xlsx.SetCellValue("Sheet1", "H"+strconv.Itoa(i+2), d.RepayTotal)
		xlsx.SetCellValue("Sheet1", "I"+strconv.Itoa(i+2), d.RepayAmountRate)
		xlsx.SetCellValue("Sheet1", "J"+strconv.Itoa(i+2), d.TargetRepayRate)
		xlsx.SetCellValue("Sheet1", "K"+strconv.Itoa(i+2), d.DiffTargetRepay)
		xlsx.SetCellValue("Sheet1", "L"+strconv.Itoa(i+2), d.AssignNum)
		xlsx.SetCellValue("Sheet1", "M"+strconv.Itoa(i+2), d.HandleNum)
		xlsx.SetCellValue("Sheet1", "N"+strconv.Itoa(i+2), d.CompleteNum)
		xlsx.SetCellValue("Sheet1", "O"+strconv.Itoa(i+2), d.LoadNum)
	}
	c.Ctx.Output.Header("Accept-Ranges", "bytes")
	c.Ctx.Output.Header("Content-Type", "application/octet-stream")
	c.Ctx.Output.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Ctx.Output.Header("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
	c.Ctx.Output.Header("Pragma", "no-cache")
	c.Ctx.Output.Header("Expires", "0")
	xlsx.Write(c.Ctx.ResponseWriter)
}

func (c *TicketController) PerformanceManagementHour() {
	var condCntr = map[string]interface{}{}

	//role, _ := rbac.GetOneRole(c.RoleID)
	c.Data["ticketItemMap"] = types.OwnTicketItemMap(c.RoleType)

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
	}
	c.Data["id"] = id

	itemID, _ := c.GetInt("item_id")
	c.Data["itemID"] = types.TicketItemEnum(itemID)
	if itemID > 0 {
		condCntr["ticket_item_id"] = types.TicketItemEnum(itemID)
	} else {
		condCntr["ticket_item_id"] = c.Data["ticketItemMap"]
	}

	currentHour := c.GetString("hour", "-1")
	c.Data["currentHour"] = currentHour
	if currentHour != "-1" {
		today := tools.GetToday()
		currentHour := today + currentHour
		condCntr["currentHour"], _ = tools.Str2Int(currentHour)
	}

	opName := strings.TrimSpace(c.GetString("op_name"))
	if len(opName) > 0 {
		c.Data["opName"] = opName
		admin, _ := models.OneAdminByNickName(opName)
		if admin.Id == 0 {
			//直接赋值-1000是当前条件不成立
			condCntr["admin_uid"] = -1000
		} else {
			condCntr["admin_uid"] = admin.Id
		}

	}

	// dateRange := c.GetString("date_range")
	// c.Data["dateRange"] = dateRange
	// if start, end, err := tools.PareseDateRangeToDayRange(dateRange); err == nil {
	// 	condCntr["date_start"], condCntr["date_end"] = start, end
	// }

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	list, count, totalStats, _ := performance.WorkerHourStatsListBackend(condCntr, c.RoleID, c.RolePid, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["hourList"] = []string{"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}
	c.Data["paginator"] = paginator
	c.Data["totalStats"] = totalStats
	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "ticket/performance_hour_list.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "ticket/performance_hour_list_scripts.html"
}

func (c *TicketController) ItemPerformanceMonthStats() {
	var condCntr = map[string]interface{}{}

	//role, _ := rbac.GetOneRole(c.RoleID)
	c.Data["ticketItemMap"] = types.OwnTicketItemMap(c.RoleType)

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
	}
	c.Data["id"] = id

	itemID, _ := c.GetInt("item_id")
	c.Data["itemID"] = types.TicketItemEnum(itemID)
	if itemID > 0 {
		condCntr["ticket_item_id"] = types.TicketItemEnum(itemID)
	} else {
		condCntr["ticket_item_id"] = c.Data["ticketItemMap"]
	}

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	list, count, _ := performance.ItemStatsListBackend(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "ticket/item_month_performance_list.html"

}

func (c *TicketController) MyProcess() {
	var uid int64
	testID, _ := c.GetInt64("test_id")
	if testID > 0 {
		uid = testID
	} else {
		uid = c.AdminUid
	}

	lastestStatsData, processChartDatas, err := performance.GetTodayStats(uid)
	var standardRepayRate float64
	if err == nil {
		standardRepayRate = performance.GetTargetRepayRateByTicketItem(lastestStatsData.TicketItem, types.TicketMyProcess)
	}
	jsonProcessChartDatas, _ := json.Marshal(processChartDatas)

	c.Data["lastestStatsData"] = lastestStatsData
	c.Data["standardRepayRate"] = standardRepayRate
	c.Data["jsonProcessChartDatas"] = string(jsonProcessChartDatas)

	c.Layout = "layout.html"
	c.TplName = "ticket/process.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "ticket/process_scripts.html"
}
