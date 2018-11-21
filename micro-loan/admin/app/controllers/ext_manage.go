package controllers

import (
	"encoding/json"
	"micro-loan/common/i18n"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/voip"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils/pagination"
)

type ExtensionController struct {
	BaseController
}

func (c *ExtensionController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "extension"
}

//分机管理
func (c *ExtensionController) List() {
	c.Data["Action"] = "list"
	c.TplName = "extension_manage/ext_list.html"

	action := "extension/list"
	gotoURL := "extension/list"
	go func() {
		err := service.UpdateSipInfo()
		if err != nil {
			c.commonError(action, gotoURL, i18n.T(c.LangUse, types.UpdateSipNumberInfoFail))
			return
		}
	}()

	var condCntr = map[string]interface{}{}
	name := c.GetString("name")
	if len(name) > 0 {
		condCntr["name"] = name
	}

	//分机号码
	mobile := c.GetString("ext_number")
	if len(mobile) > 0 {
		condCntr["extnumber"] = mobile
	}

	//通话状态
	callStatus, _ := c.GetInt("call_status", -1)
	if callStatus > 0 {
		condCntr["call_status"] = callStatus
	}
	selectedCallStatusMap := map[interface{}]interface{}{}
	selectedCallStatusMap[callStatus] = nil
	c.Data["selectedCallStatusMap"] = selectedCallStatusMap

	//分机是否启用
	enableStatus, _ := c.GetInt("enable_status", -1)
	if enableStatus >= 0 {
		condCntr["enable_status"] = enableStatus
	}
	selectedEnableStatusMap := map[interface{}]interface{}{}
	selectedEnableStatusMap[enableStatus] = nil
	c.Data["selectedEnableStatusMap"] = selectedEnableStatusMap

	//分机分配状态
	assignStatus, _ := c.GetInt("assign_status", -1)
	if assignStatus >= 0 {
		condCntr["assign_status"] = assignStatus
	}
	selectedAssignStatusMap := map[interface{}]interface{}{}
	selectedAssignStatusMap[assignStatus] = nil
	c.Data["selectedAssignStatusMap"] = selectedAssignStatusMap

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	list, count, _ := service.ListExtManageBackend(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, int64(count))

	c.Data["List"] = list
	c.Data["tagCallStatusMap"] = voip.TagCallStatusMap()
	c.Data["tagExtIsUseMap"] = voip.TagExtIsUseMap()
	c.Data["tagExtStatusMap"] = voip.TagExtStatusMap()
	c.Data["paginator"] = paginator
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "extension_manage/ext_list_script.html"
	return
}

//更新voip信息
func (c *ExtensionController) UpdateExtInfo() {
	c.Data["Action"] = "list"

	action := "extension/list"
	gotoURL := "/extension/list"

	//更新分机信息
	err := service.UpdateSipInfo()
	if err != nil {
		c.commonError(action, gotoURL, i18n.T(c.LangUse, types.UpdateSipNumberInfoFail))
		return
	}

	c.TplName = "extension_manage/ext_list.html"
	c.Layout = "layout.html"
	c.List()
	return
}

func (c *ExtensionController) AssignPage() {
	c.TplName = "extension_manage/assign.html"
	//分机号码
	extNumber := c.GetString("extnumber")
	//分配－１，取消分配－０
	isAssign, _ := c.GetInt("is_assign", -1)

	admins, num, err := service.CanExtensionAssignUsers(extNumber)
	if err != nil || num <= 0 {
		c.Data["error"] = i18n.T(c.LangUse, types.UserCanAssignNotFound)
		return
	}
	c.Data["admins"] = admins
	c.Data["extnumber"] = extNumber
	c.Data["is_assign"] = isAssign

}

func (c *ExtensionController) Assign() {
	//分机号码
	extNumber := c.GetString("extnumber")
	//分配的人员id
	assignID, uidErr := c.GetInt64("assign_id")

	if uidErr != nil || assignID <= 0 {
		c.Data["json"] = map[string]interface{}{
			"error": i18n.T(c.LangUse, types.InvalidRequest),
		}
		c.ServeJSON()
		return
	}

	//分配－１，　取消分配－０
	isAssign, err := c.GetInt("is_assign", -1)
	if err != nil || isAssign < 0 {
		c.Data["json"] = map[string]interface{}{
			"error": i18n.T(c.LangUse, types.InvalidRequest),
		}

		c.ServeJSON()
		return
	}

	result, err := service.ManualAssignOperate(assignID, extNumber, isAssign)
	if !result {
		c.Data["json"] = map[string]interface{}{"error": i18n.T(c.LangUse, err.Error())}
	} else {
		c.Data["json"] = map[string]interface{}{"status": true}
	}
	c.ServeJSON()
	return
}

func (c *ExtensionController) CancelAssign() {
	action := "/extension/list"
	gotoURL := "/extension/list"
	c.Data["Action"] = "list"
	c.TplName = "extension_manage/ext_list.html"
	//分机号码
	extNumber := c.GetString("extnumber")

	isAssign, err := c.GetInt("is_assign", -1)
	if err != nil || isAssign < 0 {
		c.Data["json"] = map[string]interface{}{
			"error": i18n.T(c.LangUse, types.InvalidRequest),
		}

		c.ServeJSON()
		return
	}
	//分配的人员id
	assignId, err := c.GetInt64("assign_id")
	if err != nil || isAssign < 0 {
		c.Data["json"] = map[string]interface{}{
			"error": i18n.T(c.LangUse, types.InvalidRequest),
		}

		c.ServeJSON()
		return
	}

	result, err := service.ManualUnAssignOperate(assignId, extNumber, isAssign)
	if !result {
		c.commonError(action, gotoURL, i18n.T(c.LangUse, err.Error()))
		return
	} else {
		c.commonError(action, gotoURL, i18n.T(c.LangUse, types.UnAssignSuccess))
		return
	}
	c.ServeJSON()
	c.Layout = "layout.html"
	return
}

//分配历史
func (c *ExtensionController) ExtHistory() {
	c.Data["Action"] = "ext_history"
	c.TplName = "extension_manage/ext_assign_history.html"
	var condCntr = map[string]interface{}{}

	//分机号码
	extnumber := c.GetString("extnumber")
	if len(extnumber) > 0 {
		condCntr["extnumber"] = extnumber
	}

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	list, count, _ := service.ListAssignHistoryBackend(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, int64(count))

	c.Data["List"] = list
	c.Data["paginator"] = paginator
	c.Layout = "layout.html"
	return
}

//通话记录
func (c *ExtensionController) CallRecord() {
	c.Data["Action"] = "call_record"
	c.TplName = "extension_manage/ext_call_record.html"

	var condCntr = map[string]interface{}{}
	c.Data["ticketItemMap"] = types.OwnTicketItemMap(c.RoleType)
	//订单id
	orderId, _ := c.GetInt("order_id", -1)
	if orderId >= 0 {
		condCntr["order_id"] = orderId
	}

	//分配的人员名字
	name := c.GetString("name")
	if len(name) > 0 {
		condCntr["name"] = name
	}

	//分机号码
	extNumber := c.GetString("extnumber")
	if len(extNumber) > 0 {
		condCntr["ext_number"] = extNumber
	}

	//工单类型
	itemID, _ := c.GetInt("item_id")
	c.Data["itemID"] = types.TicketItemEnum(itemID)

	if itemID > 0 {
		condCntr["item_id"] = types.TicketItemEnum(itemID)
	}

	// 还款时间范围
	splitSep := " - "
	repayDateRange := c.GetString("assign_call_date_range")
	if len(repayDateRange) > 16 {
		tr := strings.Split(repayDateRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["callStartTime"] = timeStart
				condCntr["callEndTime"] = timeEnd
			}
		}
	}
	c.Data["assignDateRange"] = repayDateRange

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15
	list, count, _ := service.ListExtCallHistoryBackend(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, int64(count))
	c.Data["List"] = list
	c.Data["paginator"] = paginator
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "extension_manage/ext_call_record_script.html"
	c.Layout = "layout.html"
	return
}

// 定时获取第三方voip呼叫结果通知
func (c *ExtensionController) SipCallResult() {
	callRecordId, _ := tools.Str2Int64(c.GetString("call_record_id"))

	mapData := make(map[string]interface{})
	isDail := -1

	// 根据呼叫记录Id查询呼叫记录
	sipCallRecord, err := models.GetSipCallRecordById(callRecordId)
	if err != nil {
		logs.Info("[SipCallResult] GetSipCallRecordById failed, err:", err, "callRecordId:", callRecordId)

		mapData["is_dail"] = isDail
		c.Data["json"] = mapData
		c.ServeJSON()
		return
	}

	// 当通话时间大于0时，接通；当通话时间不大于0，并且结束时间大于0时，未接通
	if sipCallRecord.BillSec > 0 {
		isDail = 1
	} else if sipCallRecord.EndTimestamp > 0 {
		isDail = 0
	}
	mapData["is_dail"] = isDail
	mapData["start_time"] = tools.MDateMHS(sipCallRecord.StartTimestamp)

	c.Data["json"] = mapData
	c.ServeJSON()
}

// 获取第三方voip呼叫话单
func (c *ExtensionController) SipCallBill() {
	callRecordId := c.GetString("call_record_id")

	mapData := make(map[string]interface{})
	isDail := -1

	// 根据呼叫记录Id查询呼叫记录
	callListReq := voip.CallListRequest{
		StartTime: tools.MDateMHSBeijing(tools.GetUnixMillis() - 3600*12),
		EndTime:   tools.MDateMHSBeijing(tools.GetUnixMillis() + 3600*12),
		MemberID:  callRecordId,
	}
	sipCallRecord, err := voip.VoipCallList(callListReq)
	bills := sipCallRecord.Data.Result.Bills
	if err != nil || len(bills) == 0 {
		logs.Info("[SipCallBill] Voip call bill failed, err:", err, "callRecordId:", callRecordId)

		mapData["is_dail"] = isDail
		c.Data["json"] = mapData
		c.ServeJSON()
		return
	}

	// 当通话时间大于0时，接通；当通话时间不大于0，并且结束时间大于0时，未接通
	bill := bills[0]
	startTimeStamp := tools.GetDateParseBeijing(bill.StartTime) * 1000
	endTimeStamp := tools.GetDateParseBeijing(bill.EndTime) * 1000
	if bill.Billsec > 0 {
		isDail = 1
	} else if endTimeStamp > 0 {
		isDail = 0
	}
	mapData["is_dail"] = isDail
	mapData["start_time"] = tools.MDateMHS(startTimeStamp)

	c.Data["json"] = mapData
	c.ServeJSON()
}

func (c *ExtensionController) SipCall() {
	obj := service.ExtensionCallParams{}
	jsonStr := c.GetString("jsonStr")
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		panic(err)
	}

	mapData := make(map[string]interface{})

	isOk, callRecordId, msg := service.ExtensionCall(obj)
	mapData["isok"] = isOk
	mapData["msg"] = i18n.T(c.LangUse, msg)
	mapData["call_record_id"] = callRecordId

	c.Data["json"] = mapData
	c.ServeJSON()
}
