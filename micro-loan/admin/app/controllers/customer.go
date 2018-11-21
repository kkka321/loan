package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils/pagination"

	"micro-loan/common/dao"
	"micro-loan/common/i18n"
	"micro-loan/common/lib/redis/cache"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type CustomerController struct {
	BaseController
}

func (c *CustomerController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "customer"
}

func (c *CustomerController) List() {
	c.Data["Action"] = "list"

	var condCntr = map[string]interface{}{}
	mobile := c.GetString("mobile")
	if len(mobile) > 0 {
		condCntr["mobile"] = mobile
	}
	realname := c.GetString("realname")
	if len(realname) > 0 {
		condCntr["realname"] = realname
	}
	tagsInt, _ := tools.Str2Int(c.GetString("tags"))
	tags := types.CustomerTags(tagsInt)
	if tags > 0 {
		condCntr["tags"] = tags
	}
	idCheckStatus, _ := c.GetInt("id_check_status")
	if idCheckStatus > 0 {
		condCntr["idCheckStatus"] = idCheckStatus
	}
	mediaSource := c.GetString("media_source")
	if len(mediaSource) > 0 {
		condCntr["media_source"] = mediaSource
	}

	generalize := c.GetString("generalize")
	if len(generalize) > 0 {
		condCntr["generalize"] = generalize
	}
	campaign := c.GetString("campaign")
	if len(campaign) > 0 {
		condCntr["campaign"] = campaign
	}

	splitSep := " - "
	// s申请时间范围
	registerTimeRange := c.GetString("register_time_range")
	if len(registerTimeRange) > 16 {
		tr := strings.Split(registerTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["register_time_start"] = timeStart
				condCntr["register_time_end"] = timeEnd
			}
		}
	}

	// user admin id
	userAccountId, _ := c.GetInt64("user_account_id")
	if userAccountId > 0 {
		condCntr["user_account_id"] = userAccountId
	}

	identity, _ := c.GetInt64("identity")
	if identity > 0 {
		condCntr["identity"] = identity
	}

	sortfield := c.GetString("field")
	if len(sortfield) > 0 {
		condCntr["field"] = sortfield
	}

	sorttype := c.GetString("sort")
	if len(sorttype) > 0 {
		condCntr["sort"] = sorttype
	}

	showMore := c.GetString("show_more", "0")
	c.Data["showMore"] = showMore

	c.Data["registerTimeRange"] = registerTimeRange
	c.Data["identity"] = identity
	c.Data["mobile"] = mobile
	c.Data["realname"] = realname
	c.Data["tags"] = tags
	c.Data["idCheckStatus"] = idCheckStatus
	c.Data["userAccountId"] = userAccountId
	c.Data["mediaSource"] = mediaSource
	c.Data["campaign"] = campaign
	c.Data["generalize"] = generalize

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	//count, _ := service.CustomerCount(condCntr)
	list, count, _ := service.CustomerList(condCntr, page, pagesize)
	service.CustomerAddBalance(list)

	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list
	c.Data["CustomerTagsMap"] = types.CustomerTagsMap()

	c.Layout = "layout.html"
	c.TplName = "customer/list.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "customer/list_scripts.html"

}

// 风险管理列表
func (c *CustomerController) Risk() {
	var condCntr = map[string]interface{}{}
	riskTypeStr := c.GetString("risk_type")
	riskType, _ := tools.Str2Int(riskTypeStr)
	if riskType > 0 {
		condCntr["risk_type"] = riskTypeStr
	}
	riskValue := c.GetString("risk_value")
	if len(riskValue) > 0 {
		condCntr["risk_value"] = riskValue
	}

	splitSep := " - "
	//
	cTimeRange := c.GetString("ctime_range")
	if len(cTimeRange) > 16 {
		tr := strings.Split(cTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["ctime_start"] = timeStart
				condCntr["ctime_end"] = timeEnd
			}
		}
	}
	c.Data["cTimeRange"] = cTimeRange

	reviewTimeRange := c.GetString("review_time_range")
	if len(reviewTimeRange) > 16 {
		tr := strings.Split(reviewTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["review_time_start"] = timeStart
				condCntr["review_time_end"] = timeEnd
			}
		}
	}
	c.Data["reviewTimeRange"] = reviewTimeRange

	status, err := c.GetInt("status")
	if err != nil || status == -1 {
		status = -1
	} else {
		condCntr["status"] = status
	}

	isDeleted, err := c.GetInt("is_deleted")
	if err != nil {
		isDeleted = 0
	}
	if isDeleted != -1 {
		condCntr["is_deleted"] = isDeleted
	}

	source, err := c.GetInt("source")
	if err != nil {
		source = -1
	}
	if source != -1 {
		condCntr["source"] = source
	}

	c.Data["risk_type"] = types.RiskTypeEnum(riskType)
	c.Data["RiskStatusMap"] = types.RiskStatusMap()
	c.Data["status"] = status
	c.Data["isDeleted"] = isDeleted
	c.Data["source"] = source

	c.Data["risk_value"] = riskValue

	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	count, list, _, _ := service.CustomerRiskList(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list
	c.Data["RiskTypeMap"] = types.RiskTypeMap()

	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "customer/risk_scripts.html"

	c.TplName = "customer/risk.html"
}

func (c *CustomerController) Detail() {
	action := "detail"
	gotoURL := "/customer/list"
	c.Data["Action"] = action

	id, _ := tools.Str2Int64(c.GetString("id"))
	c.Data["Id"] = id

	_, err := dao.CustomerOne(id)
	if err != nil {
		c.commonError(action, gotoURL, "客户不存在")
		return
	}

	c.isGrantedData(types.DataPrivilegeTypeCustomer, id)

	//公共
	c.Layout = "layout.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["CssPlugin"] = "plugin/css.html"
	c.LayoutSections["JsPlugin"] = "plugin/js.html"

	c.TplName = "customer/detail.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "customer/detail_scripts.html"
}

func (c *CustomerController) Follow() {
	action := "follow"
	gotoURL := "/customer/list"
	c.Data["Action"] = action

	id, _ := tools.Str2Int64(c.GetString("id"))
	baseInfo, err := dao.CustomerOne(id)
	if err != nil {
		c.commonError(action, gotoURL, "客户不存在")
		return
	}
	c.Data["BaseInfo"] = baseInfo

	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "customer/datepicker.html"

	c.TplName = "customer/follow.html"
}

func (c *CustomerController) FollowConfirm() {
	action := "follow/confirm"
	gotoURL := "/customer/list"
	c.Data["Action"] = action

	cid, _ := tools.Str2Int64(c.GetString("cid"))
	baseInfo, err := dao.CustomerOne(cid)
	if err != nil {
		c.commonError(action, gotoURL, "待跟进客户不存在")
		return
	}

	c.Data["OpMessage"] = "新增跟进成功."
	c.Data["BaseInfo"] = baseInfo

	var followTime int64
	followTimeStr := c.GetString("follow_time")
	expFollow := strings.Split(followTimeStr, ":")
	if len(expFollow) > 2 {
		followTimeStr = expFollow[0] + ":" + expFollow[1]
	}
	followTime = tools.GetTimeParse(followTimeStr) * 1000
	//logs.Debug("followTimeStr:", followTimeStr, "followTime:", followTime)
	if followTime <= 0 {
		followTime = tools.GetUnixMillis()
	}

	content := c.GetString("content")
	remark := c.GetString("remark")
	_, err = service.AddCustomerFollow(cid, c.AdminUid, followTime, content, remark)
	if err != nil {
		logs.Warning("service.AddCustomerFollow fail. opUid:", c.AdminUid, ", CustomerID:", cid)
	}

	c.Layout = "layout.html"
	c.TplName = "success.html"
}

func (c *CustomerController) RiskReport() {
	// action := "risk_report"
	// //gotoURL := "/customer/list"
	// c.Data["Action"] = action

	// 风险上报不和客户强绑定
	//cid, _ := tools.Str2Int64(c.GetString("cid"))
	//baseInfo, err := service.CustomerOne(cid)
	//if err != nil {
	//	c.commonError(action, gotoURL, "待风险上报客户不存在")
	//	return
	//}
	//c.Data["BaseInfo"] = baseInfo
	cid, _ := c.GetInt64("cid")

	c.Data["cid"] = cid

	c.Data["TitlePrefix"] = "上报"
	c.Data["IsRelieve"] = false
	c.Data["op_action"] = "report"
	c.Data["RiskItemMap"] = types.RiskItemMap()
	c.Data["RiskTypeMap"] = types.RiskTypeMap()
	c.Data["RiskReportReasonMap"] = types.RiskReportReasonMap()

	c.Layout = "layout.html"
	c.TplName = "customer/risk_report.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "customer/risk_report_scripts.html"
}

func (c *CustomerController) RiskQueryVal() {
	//cid, err := c.GetInt64("cid")
	cid, err := c.GetInt64("cid")
	riskItem, err2 := c.GetInt("risk_item")
	if err != nil || err2 != nil || cid <= 0 || riskItem <= 0 {
		c.Data["json"] = map[string]string{"error": "Required Param Invaild"}
	} else {
		riskItemVal, err := service.GetCustomerRiskItemVal(cid, types.RiskItemEnum(riskItem))
		jsonData := make(map[string]interface{})
		if err != nil {
			jsonData["error"] = err.Error()
		}
		jsonData["data"] = riskItemVal
		c.Data["json"] = jsonData
	}

	c.ServeJSON()
}

func (c *CustomerController) RiskSave() {
	// 简单校验
	riskItem, _ := c.GetInt("risk_item")
	riskType, _ := c.GetInt("risk_type")
	reason, _ := c.GetInt("reason")
	riskValue := c.GetString("risk_value")
	orderIds := strings.TrimSpace(c.GetString("order_ids"))
	userAccountIds := strings.TrimSpace(c.GetString("user_account_ids"))
	remark := c.GetString("remark")
	// logs.Debug("riskItem:", riskItem, ", riskType:", riskType, ", reason:", reason, ", riskValue:", riskValue, ", remark:", remark)
	if riskItem <= 0 || riskType <= 0 || reason <= 0 || len(riskValue) <= 0 || len(userAccountIds) <= 0 {
		c.newCommonError("/customer/risk", "缺少必要参数")
		return
	}

	cid, _ := c.GetInt64("cid")
	//baseInfo, err := service.CustomerOne(cid)
	//_, err := service.CustomerOne(cid)
	//if err != nil {
	//	c.commonError(action, gotoURL, "待操作客户不存在")
	//	return
	//}

	_, err := service.AddCustomerRisk(cid, c.AdminUid, types.RiskItemEnum(riskItem), types.RiskTypeEnum(riskType), types.RiskReason(reason), riskValue, remark, types.RiskWaitReview, 0, orderIds, userAccountIds)
	if err != nil {
		c.newCommonError("/customer/risk", err.Error())
		return
	}

	c.Data["OpMessage"] = "保存风险数据成功."
	c.Layout = "layout.html"
	c.Data["Redirect"] = "/customer/risk"
	c.TplName = "success_redirect.html"
}

func (c *CustomerController) RiskReview() {
	id, _ := c.GetInt64("id")
	risk, err := dao.CustomerRiskOne(id)
	if err != nil || risk.IsDeleted != 0 {
		c.newCommonError("/customer/risk", "待解除风险数据有误")
		return
	}
	c.Data["Risk"] = risk
	userAccountIds := strings.Split(risk.UserAccountIds, ",")
	c.Data["UserAccountIds"] = userAccountIds

	c.Data["id"] = id
	c.Data["TitlePrefix"] = "上报"
	c.Data["RiskItemMap"] = types.RiskItemMap()
	c.Data["RiskTypeMap"] = types.RiskTypeMap()
	c.Data["RiskStatusMap"] = types.RiskStatusMap()

	c.Layout = "layout.html"
	c.TplName = "customer/risk_review.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "customer/risk_review_scripts.html"
}

func (c *CustomerController) RiskReviewSave() {

	// 简单校验
	id, _ := c.GetInt64("id")
	status, _ := c.GetInt("status")
	remark := c.GetString("remark")
	//logs.Debug("riskItem:", riskItem, ", riskType:", riskType, ", reason:", reason, ", riskValue:", riskValue, ", remark:", remark)
	if id <= 0 || status <= 0 {
		c.newCommonError("/customer/risk", "缺少必要参数")
		return
	}

	service.ReviewCustomerRisk(id, types.RiskStatusEnum(status), remark, c.AdminUid)
	c.Data["id"] = id

	c.Data["OpMessage"] = "保存风险数据成功."
	c.Layout = "layout.html"
	c.TplName = "success_redirect.html"
}

func (c *CustomerController) RiskRelieve() {
	id, _ := c.GetInt64("id")
	risk, err := dao.CustomerRiskOne(id)
	if err != nil || risk.IsDeleted != 0 {
		c.newCommonError("/customer/risk", "待解除风险数据有误")
		return
	}
	c.Data["Risk"] = risk

	c.Data["id"] = id
	c.Data["RiskItemMap"] = types.RiskItemMap()
	c.Data["RiskTypeMap"] = types.RiskTypeMap()
	c.Data["RiskRelieveReason"] = types.RiskRelieveReasonMap()

	c.Layout = "layout.html"
	c.TplName = "customer/risk_relieve.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "customer/risk_relieve_scripts.html"
}

func (c *CustomerController) RiskRelieveSave() {

	// 简单校验
	id, _ := c.GetInt64("id")
	remark := c.GetString("remark")
	relieveReason, _ := c.GetInt("relieve_reason")
	//logs.Debug("riskItem:", riskItem, ", riskType:", riskType, ", reason:", reason, ", riskValue:", riskValue, ", remark:", remark)
	if id <= 0 || relieveReason <= 0 {
		c.newCommonError("/customer/risk", "缺少必要参数")
		return
	}

	service.RelieveCustomerRisk(id, types.RiskRelieveReason(relieveReason), remark, c.AdminUid)

	c.Data["OpMessage"] = "保存风险数据成功."
	c.Layout = "layout.html"
	c.TplName = "success_redirect.html"
}

//AjaxModify 修改某字段
func (c *CustomerController) AjaxModify() {

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

	if len(Obj.Value) == 0 {
		mapData["error"] = 1
		mapData["err_str"] = "Can't modify to empty"
		c.Data["json"] = &mapData
		c.ServeJSON()
		return
	}

	//验证 名字
	isValidName := tools.IsIndonesiaName(Obj.Value)
	if Obj.Field == "realname" && !isValidName {
		mapData["error"] = 1
		mapData["err_str"] = "Invalid Name"
		c.Data["json"] = &mapData
		c.ServeJSON()
		return
	}

	// 验证身份证号
	id, _ := tools.Str2Int64(Obj.ID)
	accountBase, _ := models.OneAccountBaseByPkId(id)

	origin := accountBase
	switch Obj.Field {
	case "realname":
		{
			if accountBase.Realname == Obj.Value {
				mapData["data"] = true
				goto RET
			}
			accountBase.Realname = Obj.Value
		}
	case "identity":
		{
			if accountBase.Identity == Obj.Value {
				mapData["data"] = true
				goto RET
			}

			// 身份证是否被使用
			_, err := models.OneAccountBaseByIdentity(Obj.Value)
			if err != orm.ErrNoRows {
				mapData["error"] = 1
				mapData["err_str"] = "Identity Already Been Used"
				goto RET
			}

			// 是否有在贷
			loanLife := service.GetLoanLifetime(id)
			if loanLife == types.LoanLifetimeInProgress {
				mapData["error"] = 1
				mapData["err_str"] = "There is Progressing Order"
				goto RET
			}
			accountBase.Identity = Obj.Value
		}
	}

	//修改
	{
		num, err := service.AjaxAccountBaseModify(id, Obj.Field, Obj.Value)

		//记日志，传数据
		if err == nil && num > 0 {
			mapData["data"] = true
			// 写操作日志
			models.OpLogWrite(c.AdminUid, accountBase.Id, models.OpCodeAccountBaseUpdate, accountBase.TableName(), origin, accountBase)
		} else {
			mapData["error"] = 1
			mapData["err_str"] = "Unknow err"
		}
	}

RET:
	c.Data["json"] = &mapData
	c.ServeJSON()
	return

}

func (c *CustomerController) Delete() {
	cid, _ := c.GetInt64("id")

	service.DeleteCustomer(cid)
}

func (c *CustomerController) SuperDelete() {

	// 只有超管才能做这种操作
	if c.AdminUid != 1 {
		return
	}

	cid, _ := c.GetInt64("id")
	service.SuperDeleteCustomer(c.AdminUid, cid)
}

func (c *CustomerController) ImportBlacklist() {
	c.Data["Action"] = "customer/import_blacklist"
	c.Data["risk_type_list"] = types.RiskItemMap()
	c.Layout = "layout.html"
	c.TplName = "customer/risk_import.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["JsPlugin"] = "plugin/js.html"
	c.LayoutSections["Scripts"] = "customer/risk_import_scripts.html"
}

func (c *CustomerController) BlacklistSave() {
	rawData := c.GetString("risk_data")
	if len(rawData) == 0 {
		c.ImportBlacklist()
		return
	}

	rawTypes := c.GetStrings("risk_type")
	if len(rawTypes) == 0 {
		c.ImportBlacklist()
		return
	}

	logs.Info("[ImportBlacklist] data:%s, type:%s", rawData, rawTypes)

	dataList := strings.Split(rawData, "\n")

	typeList := []int{}
	for _, v := range rawTypes {
		if d, e := tools.Str2Int(v); e == nil {
			typeList = append(typeList, d)
		}
	}

	if len(typeList) == 0 {
		return
	}

	var splitChar string = ""
	blackList := [][]string{}
	for _, v := range dataList {
		if splitChar == "" {
			v = strings.Trim(v, "\r")
			if strings.Contains(v, "\r") {
				splitChar = "\r"
			} else if strings.Contains(v, "_") {
				splitChar = "_"
			} else {
				splitChar = ","
			}
		}

		vec := strings.Split(v, splitChar)
		if len(vec) == 0 {
			continue
		}

		if _, e := tools.Str2Int64(vec[0]); e != nil {
			continue
		}

		blackList = append(blackList, vec)
	}

	service.ImportBlacklist(typeList, blackList, c.AdminUid)

	c.Layout = "layout.html"
	c.Data["Redirect"] = "/customer/import_blacklist"
	c.Data["OpMessage"] = "保存风险数据成功."
	c.TplName = "success_redirect.html"
}

//获取历史照片
func (c *CustomerController) PicShow() {

	action := "pic_show"
	gotoURL := "/customer/pic_show"

	c.Data["Action"] = action

	cid, err := tools.Str2Int64(c.GetString("cid"))
	if err != nil {
		c.commonError(action, gotoURL, "客户不存在")
		return
	}

	pic_list, _ := models.GetMultiPicShow(cid)

	c.Data["List"] = pic_list

	c.Layout = "layout.html"
	c.TplName = "customer/pic_show.html"

}

// 放款申请页
func (c *CustomerController) RefundApply() {
	action := "refund"
	gotoURL := "/customer/list"

	id, err := tools.Str2Int64(c.GetString("id"))

	if err != nil {
		c.commonError(action, gotoURL, "客户不存在")
		return
	}

	one, err := dao.OneAccountBalanceByAccountId(id)
	if err != nil {
		logs.Error("[RefundApply] OneRefundByAccountId err:%v id:%d", err, id)
		c.commonError(action, gotoURL, "获取客户余额信息失败")
		return
	}

	c.Data["One"] = one
	c.Data["Cid"] = id
	c.Data["Fee"], _ = config.ValidItemInt64("refund_fee_xendit")

	c.Layout = "layout.html"
	c.TplName = "customer/refund_apply.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "customer/refund_script.html"
}

func (c *CustomerController) DoRefund() {
	action := "do_refund"
	gotoURL := "/customer/list"

	//1、获取参数
	refundAmount, _ := c.GetInt64("refund_amount")
	fee, _ := config.ValidItemInt64("refund_fee_xendit")
	refundType, _ := c.GetInt("refund_type")
	orederId, _ := c.GetInt64("order_id")
	otherAccount, _ := c.GetInt64("other_account")
	if refundAmount <= 0 || fee < 0 || refundType > 3 || refundType < 1 {
		c.commonError(action, gotoURL, i18n.T(c.LangUse, "参数错误"))
		return
	}
	accountId, err := tools.Str2Int64(c.GetString("Cid"))
	if err != nil {
		logs.Error("[DoRefund] accountId:%d err:%v", accountId, err)
		c.commonError(action, gotoURL, i18n.T(c.LangUse, "客户不存在"))
		return
	}

	// 只有admin可以退款到别人订单
	if refundType == types.RefundTypeToOrder {
		fee = 0
		// 只能退到自己帐号的订单下   admin 放开这个限制
		//order, _ := models.GetOrder(orederId)
		//if order.UserAccountId != accountId && c.AdminUid != 1 {
		//	logs.Error("[DoRefund] order_id:%d accountId:%d", orederId, accountId)
		//	c.commonError(action, gotoURL, i18n.T(c.LangUse, "订单号与用户不匹配,退款失败"))
		//	return
		//}

		_, err := models.GetOrder(orederId)
		if err != nil {
			logs.Error("[DoRefund] accountId:%d orederId :%d err:%v", accountId, orederId, err)
			c.commonError(action, gotoURL, i18n.T(c.LangUse, "目的orederId错误"))
			return
		}
	}

	//只有 admin 可以操作退到别人余额
	if refundType == types.RefundTypeToOtherAccount {
		fee = 0
		//if c.AdminUid != 1 {
		//	logs.Error("[DoRefund] order_id:%d accountId:%d", orederId, accountId)
		//	c.commonError(action, gotoURL, i18n.T(c.LangUse, "只有超管可以操作退到别人余额,退款失败"))
		//	return
		//}

		_, err := models.OneAccountBaseByPkId(otherAccount)
		if err != nil {
			logs.Error("[DoRefund] accountId:%d otherAccount :%d err:%v", accountId, otherAccount, err)
			c.commonError(action, gotoURL, i18n.T(c.LangUse, "目的账户ID错误"))
			return
		}
	}

	//2、获取客户信息
	err = service.CanRefund(accountId, refundAmount, fee)
	if err != nil {
		err = fmt.Errorf("[DoRefund] CanRefund err:%v accountId:%d", err, accountId)
		c.commonError(action, gotoURL, i18n.T(c.LangUse, "退款金额错误"))
		logs.Error(err)
		return
	}

	//3、检查是否有在贷订单
	if refundType == types.RefundTypeToBankCard {
		if !service.CanRefundToBankCard(accountId) {
			err = fmt.Errorf("[DoRefund] CanRefundToBankCard return false accountId:%d", accountId)
			c.commonError(action, gotoURL, i18n.T(c.LangUse, "客户有在贷订单,不允许退款到银行账户"))
			logs.Error(err)
			return
		}
	}

	//防止多协程同时退款  第二个退款可以稍后再试
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	keyPrefix := beego.AppConfig.String("refund_lock")
	key := fmt.Sprintf("%s%d", keyPrefix, accountId)
	logs.Debug("[DoRefund] lock key:%s", key)
	re, err := cacheClient.Do("SET", key, 1, "EX", 60*60, "NX") //锁住1个小时 防止两个退款同时进行
	if err != nil || re == nil {
		err = fmt.Errorf("[DoRefund] 已经有协程在处理！ accountId:%d orderid:%d amount:%d err:%v", accountId, orederId, refundAmount, err)
		c.commonError(action, gotoURL, i18n.T(c.LangUse, "退款进行中稍后重试"))
		logs.Error(err)
		return
	}
	defer cacheClient.Do("DEL", key)

	refund := models.Refund{
		UserAccountId: accountId,
		Amount:        refundAmount,
		OpUid:         c.AdminUid,
	}

	switch refundType {
	case types.RefundTypeToOrder:
		{

			refund.ReleatedOrder = orederId
			err = service.DoRefundToOrder(&refund)
		}
	case types.RefundTypeToBankCard:
		{
			refund.Fee = fee
			err = refundToBankCard(c, &refund)
		}
	case types.RefundTypeToOtherAccount:
		{
			refund.ReleatedOrder = otherAccount
			err = service.DoRefundToOtherAccount(&refund)
		}
	}

	if err != nil {
		logs.Error("[DoRefund]  err:%v accountId:%d refundAmount：%d refundType:%d refund:%#v", err, accountId, refundAmount, refundType, refund)
		c.commonError(action, gotoURL, i18n.T(c.LangUse, "退款失败"))
		return
	}

	c.Redirect(gotoURL, 302)
}

func refundToBankCard(c *CustomerController, refund *models.Refund) error {

	//1、文件数
	fileNum, _ := c.GetInt("file_num")
	logs.Debug("fileNum:", fileNum)
	if fileNum <= 0 {
		err := fmt.Errorf("[refundToBankCard] 获得凭证失败 fileNum:%d", fileNum)
		logs.Error(err)
		return err
	}

	//3、上传凭证图片
	resIds := make([]int64, 0)
	for i := 0; i < 5; i++ {
		if i < fileNum {
			fileName := fmt.Sprintf("file%d", i)
			idPic, idPicTmp, code, err := c.UploadResource(fileName, types.Use2Refund)
			logs.Debug("[refundToBankCard] idPhoto:%d, idPhotoTmp:%s, code:%d, err:%v fileName:%s", idPic, idPicTmp, code, err, fileName)
			defer tools.Remove(idPicTmp)
			if err != nil {
				err = fmt.Errorf("[refundToBankCard] 上传凭证失败  update pic err:%v", err)
				logs.Error(err)
				return err
			}
			resIds = append(resIds, idPic)
		} else {
			resIds = append(resIds, 0)
		}
	}
	logs.Debug("[refundToBankCard] resIds:%#v", resIds)

	//4、退款

	err := service.DoRefundToBankCard(refund, resIds)
	if err != nil {
		err = fmt.Errorf("[refundToBankCard] refundToBankCard .err:%v refund：%#v resIds:%#v", err, refund, resIds)
		logs.Error(err)
		return err
	}
	return nil
}

func (c *CustomerController) ModifyMobile() {

	action := "pic_show"
	gotoURL := "/customer/list"
	id, _ := tools.Str2Int64(c.GetString("id"))
	account, err := models.OneAccountBaseByPkId(id)
	if err != nil {
		c.commonError(action, gotoURL, "客户不存在")
		return
	}

	logs.Debug("id:%d", id)

	c.Data["Account"] = account
	c.TplName = "customer/modify_mobile.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "customer/list_scripts.html"
	//c.ServeJSON()

}

func (c *CustomerController) DoModifyMobile() {
	action := "pic_show"
	gotoURL := "/customer/list"

	id, _ := tools.Str2Int64(c.GetString("id"))
	mobileNew := c.GetString("mobile_new")
	mobileNew = tools.Strim(mobileNew)
	if len(mobileNew) <= 0 {
		c.commonError(action, gotoURL, "手机号错误")
		return
	}
	logs.Debug("id:%d newMobild:%s", id, mobileNew)

	err := service.ModifyMobile(c.AdminUid, id, mobileNew)
	if err != nil {
		logs.Error("[DoModifyMobile] ModifyMobile err:%s id:%d opId:%d newMobile:%s", err, id, c.AdminUid, mobileNew)
		c.commonError(action, gotoURL, "更新失败")
		return
	}

	c.Redirect("/customer/list", 302)
	return

}

func (c *CustomerController) DetailBaseInfo() {
	c.detailTab("base-info")
	return
}

func (c *CustomerController) DetailOtherInfo() {
	c.detailTab("other-info")
	return
}

func (c *CustomerController) DetailBigDataInfo() {
	c.detailTab("big-data-info")
	return
}

func (c *CustomerController) DetailCommunicationRecord() {
	c.detailTab("communication-record")
	return
}

func (c *CustomerController) DetailLoanHistory() {
	c.detailTab("loan-history")
	return
}

func (c *CustomerController) DetailCheckDuplicate() {
	c.detailTab("check-duplicate")
	return
}

func (c *CustomerController) detailTab(dataType string) {
	//公共
	c.TplName = "customer/detail_tab.html"
	id, _ := tools.Str2Int64(c.GetString("id"))
	orderID, _ := tools.Str2Int64(c.GetString("order_id"))

	c.Data["DataType"] = dataType

	c.isGrantedData(types.DataPrivilegeTypeCustomer, id)

	baseInfo, err := dao.CustomerOne(id)
	if err != nil {
		return
	}
	c.Data["BaseInfo"] = baseInfo

	var hasProfile = true
	profile, err := dao.CustomerProfile(id)
	if err != nil {
		hasProfile = false
	}
	c.Data["HasProfile"] = hasProfile
	c.Data["Profile"] = profile

	var queryID int64 = id
	if orderID > 0 {
		queryID = orderID
	}

	clientInfo, errCI := models.OneLastClientInfoByRelatedID(queryID)
	c.Data["ClientInfo"] = clientInfo

	if dataType == "base-info" {
		var hasLiveVerify = true
		liveVerify, err := dao.CustomerLiveVerify(id)
		if err != nil {
			hasLiveVerify = false
		}
		c.Data["HasLiveVerify"] = hasLiveVerify
		c.Data["LiveVerify"] = liveVerify
	} else if dataType == "other-info" {
		//
	} else if dataType == "big-data-info" {
		var hasBigData = true
		var bigData service.EsResponse

		if errCI != nil {
			hasBigData = false
		} else {
			if orderID > 0 {
				snapshot, errSS := models.OrderLastEsSnapshot(orderID)
				if errSS == nil {
					json.Unmarshal([]byte(snapshot.Data), &bigData)
				}
			} else {
				bigData, _, _, err = service.EsSearchById(tools.Md5(clientInfo.Imei))
			}
			//logs.Debug("bigData:", bigData)
			if bigData.Found != true {
				hasBigData = false
			}
		}
		c.Data["HasBigData"] = hasBigData
		c.Data["BigData"] = bigData
	} else if dataType == "communication-record" {
		var hasFollow = true
		followList, _, err := service.CustomerFollowList(id)
		if err != nil {
			hasFollow = false
		}
		c.Data["HasFollow"] = hasFollow
		c.Data["FollowList"] = followList
	} else if dataType == "loan-history" {
		//借款历史
		var condCntr = map[string]interface{}{}
		condCntr["user_account_id"] = id
		page, _ := tools.Str2Int(c.GetString("p"))
		pagesize := 500
		sortfield := c.GetString("field")
		if len(sortfield) > 0 {
			condCntr["field"] = sortfield
		}
		sorttype := c.GetString("sort")
		if len(sorttype) > 0 {
			condCntr["sort"] = sorttype

		}

		count, list, _, _ := service.RiskCtlList(condCntr, page, pagesize)
		paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

		c.Data["paginator"] = paginator
		c.Data["List"] = list
	} else if dataType == "check-duplicate" {
		dupOrderNo := service.GetDupOrderNo(&baseInfo, profile, &clientInfo)
		c.Data["DupOrderNo"] = dupOrderNo
	}
}
