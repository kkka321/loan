package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils/pagination"

	"micro-loan/common/dao"
	"micro-loan/common/i18n"
	"micro-loan/common/models"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type RiskCtlController struct {
	BaseController
}

func (c *RiskCtlController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "riskctl"
}

func (c *RiskCtlController) List() {
	c.Data["Action"] = "list"

	var condCntr = map[string]interface{}{}

	realname := c.GetString("realname")
	if len(realname) > 0 {
		condCntr["realname"] = realname
	}
	c.Data["realname"] = realname

	riskCtlRegular := c.GetString("risk_ctl_regular")
	if len(riskCtlRegular) > 0 {
		condCntr["risk_ctl_regular"] = riskCtlRegular
	}
	c.Data["riskCtlRegular"] = riskCtlRegular

	// user admin id
	userAccountId, _ := c.GetInt64("user_account_id")
	if userAccountId > 0 {
		condCntr["user_account_id"] = userAccountId
	}
	c.Data["userAccountId"] = userAccountId

	id, _ := c.GetInt64("id")
	if id > 0 {
		condCntr["id"] = id
	}
	c.Data["id"] = id

	riskCtlStatusMulti := c.GetStrings("risk_ctl_status")
	if len(riskCtlStatusMulti) > 0 && !tools.InSlice("-1", riskCtlStatusMulti) {
		condCntr["risk_ctl_status"] = riskCtlStatusMulti
	}
	c.Data["statusSelectMultiBox"] = service.BuildJsVar("statusSelectMultiBox", riskCtlStatusMulti)

	platformMulti := c.GetStrings("platform_mark")
	if len(platformMulti) > 0 {
		condCntr["platform_mark"] = platformMulti
	}
	c.Data["platformMarkMultiBox"] = service.BuildJsVar("platformMarkMultiBox", platformMulti)

	//is_reloan
	isReloan, _ := c.GetInt64("is_reloan", -1)
	if isReloan > -1 {
		condCntr["is_reloan"] = isReloan
	}
	c.Data["is_reloan"] = isReloan

	splitSep := " - "

	// 一些查询时间
	applyTimeRange := c.GetString("apply_time_range")
	if len(applyTimeRange) > 16 {
		expApplyTime := strings.Split(applyTimeRange, splitSep)
		if len(expApplyTime) == 2 {
			applyTimeStart := tools.GetDateParseBackend(expApplyTime[0]) * 1000
			applyTimeEnd := tools.GetDateParseBackend(expApplyTime[1])*1000 + 3600*24*1000
			if applyTimeStart > 0 && applyTimeEnd > 0 {
				condCntr["apply_time_start"] = applyTimeStart
				condCntr["apply_time_end"] = applyTimeEnd
			}
		}
	}
	c.Data["apply_time_range"] = applyTimeRange

	checkTimeRange := c.GetString("check_time_range")
	if len(checkTimeRange) > 16 {
		expCheckTime := strings.Split(checkTimeRange, splitSep)
		if len(expCheckTime) == 2 {
			checkTimeStart := tools.GetDateParseBackend(expCheckTime[0]) * 1000
			checkTimeEnd := tools.GetDateParseBackend(expCheckTime[1])*1000 + 3600*24*1000
			if checkTimeStart > 0 && checkTimeEnd > 0 {
				condCntr["check_time_start"] = checkTimeStart
				condCntr["check_time_end"] = checkTimeEnd
			}
		}
	}
	c.Data["check_time_range"] = checkTimeRange

	//随机值查询
	randomValueStart, err := c.GetInt64("random_value_start")
	if err == nil && randomValueStart >= 0 {
		condCntr["random_value_start"] = randomValueStart
		c.Data["randomValueStart"] = randomValueStart
	}
	randomValueEnd, err := c.GetInt64("random_value_end")
	if err == nil && randomValueEnd >= 0 {
		condCntr["random_value_end"] = randomValueEnd
		c.Data["randomValueEnd"] = randomValueEnd
	}

	//修正值查询
	fixValueStart, err := c.GetInt64("fix_value_start")
	if err == nil && fixValueStart >= 0 {
		condCntr["fix_value_start"] = fixValueStart
		c.Data["fixValueStart"] = fixValueStart
	}

	fixValueEnd, err := c.GetInt64("fix_value_end")
	if err == nil && fixValueEnd >= 0 {
		condCntr["fix_value_end"] = fixValueEnd
		c.Data["fixValueEnd"] = fixValueEnd
	}

	sortfield := c.GetString("field")
	if len(sortfield) > 0 {
		condCntr["field"] = sortfield
	}

	sorttype := c.GetString("sort")
	if len(sorttype) > 0 {
		condCntr["sort"] = sorttype
	}

	c.Data["RiskCtlMap"] = types.RiskCtlMap()
	c.Data["IsReloanMap"] = types.IsReloanMap()
	c.Data["PlatformMarkMap"] = types.PlatformMarkMap()

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := service.Pagesize

	count, list, _, _ := service.RiskCtlList(condCntr, page, pagesize)

	logs.Debug("count:", count)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.Layout = "layout.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "riskctl/datepicker.html"

	c.TplName = "riskctl/list.html"
}

func (c *RiskCtlController) RegularAll() {
	orderID, _ := tools.Str2Int64(c.GetString("order_id"))
	list, num, _ := service.GetAllHitRegularByOrderID(orderID)
	var output = map[string]interface{}{}

	output["code"] = 0
	output["data"] = map[string]interface{}{
		"number": num,
		"list":   list,
	}

	c.Data["json"] = output
	c.ServeJSON()

	return
}

func (c *RiskCtlController) ShowVerifyResult() {
	orderID, _ := tools.Str2Int64(c.GetString("order_id"))
	list, _, remark, err := service.GetPhoneVerifyResultDetail(c.LangUse, orderID)

	var output = map[string]interface{}{}

	if err != nil {
		output["code"] = -1
		output["data"] = ""
	} else {
		output["code"] = 0
		output["data"] = map[string]interface{}{
			"list":   list,
			"remark": remark,
		}
	}

	c.Data["json"] = output
	c.ServeJSON()

	return
}

func (c *RiskCtlController) PhoneVerify() {
	action := "phone_verify"
	gotoURL := "/riskctl/list"
	// c.Data["Action"] = action
	c.Layout = "layout.html"

	orderId, _ := tools.Str2Int64(c.GetString("order_id"))
	// 动态数据权限检查
	c.isGrantedData(types.DataPrivilegeTypeOrder, orderId)

	ticket, err := models.GetTicketForPhoneVerifyOrInfoReivew(orderId)
	if err != nil {
		logs.Error("[PhoneVerify] GetTicketForPhoneVerifyOrInfoReivew query ticket, relatedID:%d, err: %v", orderId, err)
		c.commonError(action, gotoURL, fmt.Sprintf("该订单无工单信息:%d", orderId))
		return
	}
	c.Data["ItemID"] = ticket.ItemID

	// 状态有误,展示结果
	orderData, err := models.GetOrder(orderId)
	if err != nil || orderData.RiskCtlStatus != types.RiskCtlWaitPhoneVerify || orderData.IsTemporary == types.IsTemporaryYes {
		list, invalidReason, remark, errDetail := service.GetPhoneVerifyResultDetail(c.LangUse, orderId)
		if errDetail != nil {
			c.commonError(action, gotoURL, "订单状态有误")
			return
		}

		listRst := []service.PhoneVerifyResultDetailItem{}
		// info reviev 只展示自己的问题
		if ticket.ItemID == types.TicketItemInfoReview {
			for _, v := range list {
				if strings.HasPrefix(v.Qid, "18") {
					listRst = append(listRst, v)
				}
			}
		} else {
			listRst = list
		}

		c.Data["checkStatus"] = orderData.CheckStatus
		c.Data["invalidReason"] = invalidReason
		c.Data["list"] = listRst
		c.Data["remark"] = remark
		c.TplName = "riskctl/view_phone_verify.html"
		return
	}

	// 展示电核

	id := orderData.UserAccountId

	c.isGrantedData(types.DataPrivilegeTypeCustomer, id)
	baseInfo, _ := dao.CustomerOne(id)
	c.Data["BaseInfo"] = baseInfo

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"

	c.Data["OrderId"] = orderId

	questionHtml := c.GetString("phone_verify_question")
	if len(questionHtml) <= 0 {
		if ticket.ItemID == types.TicketItemInfoReview {
			questionHtml = service.BuildInfoReviewQuestionHtml(orderData.UserAccountId, orderId, c.LangUse)
		} else {
			questionHtml = service.BuildPhoneVerifyQuestionHtml(orderData.UserAccountId, orderId, c.LangUse)
		}
	}

	//logs.Warn("questionHtml is :%s", questionHtml)

	c.Data["QuestionHtml"] = questionHtml

	c.Data["ReloanFlag"] = 0
	if dao.IsRepeatLoan(orderData.UserAccountId) {
		c.Data["ReloanFlag"] = 1
	}

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "riskctl/phone_verify.js.html"
	c.TplName = "riskctl/phone_verify.html"
}

type PhoneVerifyCallRecordParams struct {
	OrderId        string
	PhoneTime      string
	PhoneConnected string
	Remark         string
	Result         string
}

func (c *RiskCtlController) PhoneVerifySave() {
	action := "/riskctl/phone_verify/save"
	gotoURL := "/riskctl/list"
	c.Data["Action"] = action

	orderId, _ := tools.Str2Int64(c.GetString("order_id"))

	orderData, err := models.GetOrder(orderId)
	if err != nil || orderData.RiskCtlStatus != types.RiskCtlWaitPhoneVerify || orderData.IsTemporary == types.IsTemporaryYes {
		c.commonError(action, gotoURL, "订单状态有误")
		return
	}
	ticket, err := models.GetTicketForPhoneVerifyOrInfoReivew(orderId)
	if err != nil {
		logs.Error("[PhoneVerify] GetTicketForPhoneVerifyOrInfoReivew query ticket, relatedID:%d, err: %v", orderId, err)
		c.commonError(action, gotoURL, "无此工单")
		return
	}

	if ticket.ItemID == types.TicketItemPhoneVerify {
		// 保存电核通话记录
		obj := PhoneVerifyCallRecordParams{
			OrderId:        c.GetString("order_id"),
			PhoneTime:      c.GetString("phone_time"),
			PhoneConnected: c.GetString("phone_connected"),
			Remark:         c.GetString("remark"),
			Result:         c.GetString("result"),
		}

		err = c.savePhoneVerifyCallRecord(obj)
		if err != nil {
			c.commonError(action, gotoURL, i18n.T(c.LangUse, "保存电核记录失败"))
			return
		}
	}

	// 保存电核记录
	qids := c.GetString("qids")
	qidsBox := strings.Split(qids, ",")
	//! 魔术数字...
	if len(qidsBox) != 6 {
		c.commonError(action, gotoURL, "提交数据有问题,请确认.")
		return
	}

	var qValueBox = make([]string, 6)
	var qStatusBox = make([]int, 6)
	var qidsIntBox = make([]int, 6)
	var qid2StatusMap = map[int]int{}
	for index, qid := range qidsBox {
		fieldName := fmt.Sprintf("qid_value_%s", qid)
		qValueBox[index] = c.GetString(fieldName)

		statusName := fmt.Sprintf("qid_status_%s", qid)
		qStatus, _ := c.GetInt(statusName, 0)
		qStatusBox[index] = qStatus

		qidInt, _ := tools.Str2Int(qid)
		qidsIntBox[index] = qidInt

		qid2StatusMap[qidInt] = qStatus
	}

	//logs.Warn("qValueBox:%#v", qValueBox)
	//logs.Warn("qStatusBox:%#v", qStatusBox)
	//logs.Warn("qidsIntBox:%#v", qidsIntBox)

	redirectReject, _ := tools.Str2Int(c.GetString("redirect_reject"))

	answerPhoneStatus, _ := tools.Str2Int(c.GetString("answer_phone_status"))
	identityInfoStatus, _ := tools.Str2Int(c.GetString("identity_info_status"))
	ownerMobileStatus, _ := tools.Str2Int(c.GetString("owner_mobile_status"))
	//ownerMobileWhatsapp, _ := tools.Str2Int(c.GetString("owner_mobile_whatsapp"))
	ownerMobileWhatsapp := service.QuestionStatusNormal
	// 解决info review电核页面， 设专有页面后， answerPhoneStatus 永远为 0 ，
	// 则命中二级随机数时， 会直接拒绝 answerPhoneStatus 不为正常接通状态的单子
	if ticket.ItemID == types.TicketItemInfoReview {
		answerPhoneStatus = service.QuestionStatusNormal
		identityInfoStatus = service.QuestionStatusNormal
		ownerMobileStatus = service.QuestionStatusNormal
		ownerMobileWhatsapp = service.QuestionStatusNormal
	}
	result, _ := tools.Str2Int(c.GetString("result"))
	invalidReason, _ := tools.Str2Int(c.GetString("invalid_reason"))

	logs.Info("invalidReason:%d", invalidReason)

	if dao.IsRepeatLoan(orderData.UserAccountId) {
		//如果是复贷用户的话，直接算“您的手机号码是？”通过
		ownerMobileStatus = 1
		ownerMobileWhatsapp = 1
	}

	phoneVerify := models.PhoneVerifyRecord{
		OrderId: orderId,

		Q1Id:     qidsIntBox[0],
		Q1Value:  qValueBox[0],
		Q1Status: qStatusBox[0],

		Q2Id:     qidsIntBox[1],
		Q2Value:  qValueBox[1],
		Q2Status: qStatusBox[1],

		Q3Id:     qidsIntBox[2],
		Q3Value:  qValueBox[2],
		Q3Status: qStatusBox[2],

		Q4Id:     qidsIntBox[3],
		Q4Value:  qValueBox[3],
		Q4Status: qStatusBox[3],

		Q5Id:     qidsIntBox[4],
		Q5Value:  qValueBox[4],
		Q5Status: qStatusBox[4],

		Q6Id:     qidsIntBox[5],
		Q6Value:  qValueBox[5],
		Q6Status: qStatusBox[5],

		AnswerPhoneStatus:   answerPhoneStatus,
		IdentityInfoStatus:  identityInfoStatus,
		OwnerMobileStatus:   ownerMobileStatus,
		OwnerMobileWhatsapp: ownerMobileWhatsapp,
		RedirectReject:      redirectReject,
		InvalidReason:       invalidReason,
		OpUid:               c.AdminUid,
		Remark:              c.GetString("remark"),
		Result:              result,
		Ctime:               tools.GetUnixMillis(),
	}

	err = service.PhoneVerifySave(phoneVerify, orderData, ticket, redirectReject, qid2StatusMap, c.AdminUid)

	if err != nil {
		orderDataJSON, _ := tools.JsonEncode(orderData)
		logs.Error("Update order has wrong, orderData:", orderDataJSON, ", err:", err)
		c.commonError(action, gotoURL, i18n.T(c.LangUse, "更新订单出错了,请检查订单数据."))
		return
	}

	c.Data["OpMessage"] = i18n.T(c.LangUse, "保存电话审核结果成功.")
	c.Layout = "layout.html"
	c.TplName = "success.html"
}

func (c *RiskCtlController) savePhoneVerifyCallRecord(obj PhoneVerifyCallRecordParams) (err error) {

	orderId, _ := tools.Str2Int64(obj.OrderId)
	phoneConnect, _ := tools.Str2Int(obj.PhoneConnected)
	remark := obj.Remark
	result, _ := tools.Str2Int(obj.Result)

	phoneTime := obj.PhoneTime
	phoneTimeInt64, _ := tools.GetTimeParseWithFormat(phoneTime, "2006-01-02 15:04:05")

	phoneVerifyCallDetail := models.PhoneVerifyCallDetail{}
	phoneVerifyCallDetail.OrderId = orderId
	phoneVerifyCallDetail.PhoneConnect = phoneConnect
	phoneVerifyCallDetail.PhoneTime = phoneTimeInt64 * 1000
	phoneVerifyCallDetail.Result = result
	phoneVerifyCallDetail.Remark = remark
	phoneVerifyCallDetail.OpUid = c.AdminUid
	timeTag := tools.GetUnixMillis()
	phoneVerifyCallDetail.Ctime = timeTag
	phoneVerifyCallDetail.Utime = timeTag

	_, err = models.AddPhoneVerifyCallDetail(&phoneVerifyCallDetail)
	if err != nil {
		phoneVerifyDataJSON, _ := tools.JsonEncode(phoneVerifyCallDetail)
		logs.Error("Insert phone verify call detail has wrong, phoneVerifyData:", phoneVerifyDataJSON, ", err:", err)
		return
	}

	// 写操作日志
	models.OpLogWrite(c.AdminUid, orderId, models.OpPhoneVerifyCaseUpdate, phoneVerifyCallDetail.TableName(), "", phoneVerifyCallDetail)
	return
}

func (c *RiskCtlController) PhoneVerifyCallRecord() {
	action := "/riskctl/phone_verify/call_record"
	gotoURL := "/ticket/me"

	obj := PhoneVerifyCallRecordParams{}
	jsonStr := c.GetString("jsonStr")
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		panic(err)
	}

	err := c.savePhoneVerifyCallRecord(obj)
	if err != nil {
		c.commonError(action, gotoURL, i18n.T(c.LangUse, "保存电核通话记录失败"))
		return
	}
	// 更新指定工单为部分完成状态
	orderID, _ := strconv.ParseInt(obj.OrderId, 10, 64)
	handleTimestamp, _ := tools.GetTimeParseWithFormat(obj.PhoneTime, "2006-01-02 15:04:05")
	handleTime := handleTimestamp * 1000
	nextHandleTime := handleTime + 3600*1000
	ticket.UpdateByHandleCase(orderID, types.TicketItemPhoneVerify, handleTime, nextHandleTime, types.PhoneObjectSelf, 0, 0, "")
	ticket.PartialCompleteByRelatedID(orderID, types.TicketItemPhoneVerify)

	mapData := make(map[string]interface{})
	mapData["msg"] = i18n.T(c.LangUse, "保存电核通话记录成功")

	c.Data["json"] = mapData
	c.ServeJSON()
}

func (c *RiskCtlController) PhoneVerifyCallDetail() {
	orderId, _ := c.GetInt64("order_id")

	c.isGrantedData(types.DataPrivilegeTypeOrder, orderId)

	list, _ := service.GetPhoneVerifyCallDetailListByOrderIds(orderId)

	c.Data["list"] = list
	c.Layout = "layout.html"
	c.TplName = "riskctl/phone_verify_call_detail.html"
}

func (c *RiskCtlController) CheckBlacklist() {
	orderId, _ := c.GetInt64("id")

	service.RecheckThirdBlacklist(c.AdminUid, orderId)

	response := map[string]interface{}{}
	response["status"] = "ok"
	c.Data["json"] = response

	c.ServeJSON()
}

func (c *RiskCtlController) Follow() {
	orderId, _ := c.GetInt64("id")

	clientInfo, _ := service.OrderClientInfo(orderId)
	c.Data["data"] = clientInfo

	account, err := service.AccountBaseByOrderId(orderId)
	if err != nil {
		c.Data["StemFrom"] = clientInfo.StemFrom
	} else {
		c.Data["StemFrom"] = account.StemFrom
	}

	c.Layout = "layout.html"
	c.TplName = "riskctl/follow.html"
}
