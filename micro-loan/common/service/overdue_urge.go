package service

import (
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/repayremind"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/thirdparty/voip"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func CalculateOverdueLevel(reapyDate int64) (level string, days int64, err error) {
	today := tools.NaturalDay(0)
	if reapyDate > today {
		err = fmt.Errorf("time has wrong, today: %d, reapyDate: %d", today, reapyDate)
		return
	}

	days = (today - reapyDate) / 86400000
	logs.Debug("days: %d", days)

	// 获取此应还日期，最新的case
	var lastStartDay int
	for l, startDay := range types.OverdueLevelCreateDaysMap() {
		// 大于等于， 则说明， 已经满足此case的最新条件
		if int(days) >= startDay {
			// 若此最新日期，大于最新case的初始日期， 则替换
			if lastStartDay < startDay {
				level = l
				lastStartDay = startDay
			}
		}
	}

	// 下限err
	if len(level) == 0 {
		err = fmt.Errorf("did not reach the rating time, days: %d", days)
	}

	// 目前没有上限err

	// if days >= 2 && days < 8 {
	// 	level = types.OverdueLevelM11
	// } else if days >= 8 && days < 16 {
	// 	level = types.OverdueLevelM12
	// } else if days >= 16 && days < 31 {
	// 	level = types.OverdueLevelM13
	// } else if days >= 31 && days < 61 {
	// 	level = types.OverdueLevelM2
	// } else if days >= 61 {
	// 	level = types.OverdueLevelM3
	// } else {
	// 	err = fmt.Errorf("did not reach the rating time, days: %d", days)
	// }

	return
}

func CalculateOverdueDays(order *models.Order) int64 {
	if order.CheckStatus == types.LoanStatusInvalid {
		return 0
	}

	if order.IsOverdue == types.IsOverdueNo {
		return 0
	}

	replan, err := models.GetLastRepayPlanByOrderid(order.Id)
	if err != nil {
		return 0
	}

	days := int64(0)
	if order.CheckStatus == types.LoanStatusRollClear || order.CheckStatus == types.LoanStatusAlreadyCleared {
		days = (order.FinishTime - replan.RepayDate) / 86400000
	} else {
		days = (tools.NaturalDay(0) - replan.RepayDate) / 86400000
	}

	if days < 0 {
		days = 0
	}

	return days
}

func CalculateOverdue(reapyDate int64) (days int64, err error) {
	today := tools.NaturalDay(0)
	if reapyDate > today {
		err = fmt.Errorf("time has wrong, today: %d, reapyDate: %d", today, reapyDate)
		return
	}

	days = (today - reapyDate) / 86400000
	return
}

// HandleOverdueCase 处理订单的逾期案件
// 必须走主库
// 本方法做了什么:
// 1. 检验订单是否存在
// 2. 计算逾期天数 overdueDays
// 3. 根据逾期天数 overdueDays 拉黑身份证和手机号码
// 4-1. 无逾期案件, 生成逾期案件, 生成ticket
// 4-2. 存在历史案件, 若结清或展期结清,则关闭案件
// 展期进行中则冻结， 否则发现被冻结，进行解冻
// 被哪儿调用了:
// 调用者 1. 日处理逾期task, 用于升级, 更新案件
// 调用者 2. 用户调用api 去创建 CreateRollOrder , 创建展期订单, 触发原订单更新状态之后, 处理逾期案件, 此时状态应该是 展期中....
// 调用者 3. 减免触发结清 ReducePenalty, 更新 逾期案件   结清状态
// 调用者 4. repayNormalLoan , 还款触发结清, 更新逾期案件
// 调用者 5. doOrderRoll 更新旧订单之后, 此时应该是 展期结清- 关闭case
// 调用者 6. HandleRollOrder 处理展期订单 - 冻结case
func HandleOverdueCase(orderID int64) (err error) {
	// orderID 合法性校验
	if orderID <= 0 {
		err = fmt.Errorf("no data need execute, orderID: %d", orderID)
		return
	}

	// 获取有效订单数据
	orderData, err := models.GetOrder(orderID)
	orderDataJSON, _ := tools.JsonEncode(orderData)
	if err != nil {
		logs.Error("[HandleOverdueCase] 订单数据有误, orderData: %d, err: %v", orderDataJSON, err)
		return
	}

	// 计算逾期天数和逾期case level
	repayPlan, err := models.GetLastRepayPlanByOrderid(orderID)
	repayPlanJSON, _ := tools.JsonEncode(repayPlan)
	if err != nil || repayPlan.RepayDate <= 0 {
		logs.Error("[HandleOverdueCase] 还款计划数据有误, orderID: %d, repayPlan: %s, err: %v", orderID, repayPlanJSON, err)
		return
	}
	caseLevel, overdueDays, err := CalculateOverdueLevel(repayPlan.RepayDate)
	//暂时是废话，-- 稳定之后可以移除 TODO rm
	if err != nil || overdueDays < 1 {
		logs.Warning("[HandleOverdueCase] 不满足入催条件, orderData:", orderDataJSON,
			", repayPlan", repayPlanJSON,
			", caseLevel:", caseLevel, ", overdueDays:", overdueDays,
			", err:", err)
		return
	}

	// 命中系统黑名单规则，逾期>= 设定天数，触发黑名单事件
	itemName := "overdue_blacklist_day"
	itemValue, err := config.ValidItemInt(itemName)
	//逾期>=30 并且 订单状态为 9（逾期）触发黑名单事件
	if overdueDays >= int64(itemValue) && orderData.CheckStatus == types.LoanStatusOverdue {
		accountBase, _ := models.OneAccountBaseByPkId(orderData.UserAccountId)
		event.Trigger(&evtypes.BlacklistEv{
			accountBase.Id,
			types.RiskItemMobile,
			accountBase.Mobile,
			types.RiskReasonHighRisk,
			"overdue>=" + tools.Int2Str(itemValue),
		})
		event.Trigger(&evtypes.BlacklistEv{
			accountBase.Id,
			types.RiskItemIdentity,
			accountBase.Identity,
			types.RiskReasonHighRisk,
			"overdue>=" + tools.Int2Str(itemValue),
		})
	}

	// 获取当前未出催逾期case, 对case 进行新建, 更新, 出催, 并同时更新 工单
	oneCase, err := dao.GetInOverdueCaseByOrderID(orderID)
	o := orm.NewOrm()
	o.Using(oneCase.Using())
	if err != nil {
		if orderData.CheckStatus == types.LoanStatusRolling {
			//skip
			return
		} else if orderData.CheckStatus != types.LoanStatusAlreadyCleared &&
			orderData.CheckStatus != types.LoanStatusRollClear {

			// 触发RM case过期
			repayremind.ExpireCaseByOrderID(orderID)

			// 说明没有对应的案件,生成之
			createNewCase(orderID, caseLevel, int(overdueDays), orderData.UserAccountId)
			return
		} else {
			logs.Info("[HandleOverdueCase] order clear and no case orderId:%d", orderData.Id)
			return
		}
	}

	// 存在历史案件
	if orderData.CheckStatus == types.LoanStatusAlreadyCleared ||
		orderData.CheckStatus == types.LoanStatusRollClear {
		/// 用户居然结清了
		oneCase.IsOut = types.IsUrgeOutYes
		if orderData.CheckStatus == types.LoanStatusRollClear {
			oneCase.OutReason = types.UrgeOutReasonRollCleared
		} else {
			oneCase.OutReason = types.UrgeOutReasonCleared
		}
		oneCase.OutUrgeTime = tools.GetUnixMillis()
		oneCase.Utime = oneCase.OutUrgeTime

		oneCaseJSON, _ := tools.JsonEncode(oneCase)
		_, err = o.Update(&oneCase, "is_out", "out_reason", "out_urge_time", "utime")
		if err != nil {
			logs.Error("[HandleOverdueCase] 案件结清更新失败, oneCase:", oneCaseJSON, ", err:", err)
			return
		}

		// 自动完成工单
		ticket.CompleteByRelatedID(oneCase.Id, types.OverdueLevelTicketItemMap()[oneCase.CaseLevel])

		logs.Informational("[HandleOverdueCase] 案件结清更新成功, oneCase:", oneCaseJSON)
		return
	} else if orderData.CheckStatus == types.LoanStatusRolling {
		if oneCase.IsOut == types.IsUrgeOutNo {
			oneCase.IsOut = types.IsUrgeOutFrozen
		}
		oneCase.Utime = tools.GetUnixMillis()

		oneCaseJSON, _ := tools.JsonEncode(oneCase)
		_, err = o.Update(&oneCase, "is_out", "out_reason", "out_urge_time", "utime")
		if err != nil {
			logs.Error("[HandleOverdueCase] 案件冻结更新失败, oneCase:", oneCaseJSON, ", err:", err)
			return
		}

		logs.Informational("[HandleOverdueCase] 案件冻结更新成功, oneCase:", oneCaseJSON)
		return
	}

	doCaseUpdate(oneCase, caseLevel, int(overdueDays), orderData.UserAccountId)
	return
}

func createNewCase(orderID int64, caseLevel string, overdueDays int, userAccountID int64) (id int64, err error) {
	// create new case
	newCase := models.OverdueCase{}
	// 说明没有对应的案件,生成之
	newCase.OrderId = orderID
	newCase.CaseLevel = caseLevel
	newCase.OverdueDays = int(overdueDays)
	//newCase.AssignUid = assignUid
	newCase.JoinUrgeTime = tools.GetUnixMillis()
	newCase.Utime = newCase.JoinUrgeTime
	id, err = models.OrmInsert(&newCase)
	if err != nil {
		logs.Error("[createNewCase] overdue case insert failed, oneCase:", newCase, ", err:", err)
		return
	}
	if id > 0 {
		// 非边缘订单，或者自催订单生成工单
		entrustDay, err := config.ValidItemInt("outsource_day")
		if err != nil {
			entrustDay = types.EntrustDay
			logs.Warning("[ApplyEntrustCondition] entrust day config losed:", entrustDay)
		}
		// 大于entrustDay 不自动生成工单， 需要手动生成
		if overdueDays <= entrustDay {
			ticket.CreateTicket(types.OverdueLevelTicketItemMap()[caseLevel], id, types.Robot, orderID, userAccountID, nil)
		}
	}
	return
}

func doCaseUpdate(oneCase models.OverdueCase, caseLevel string, overdueDays int, userAccountID int64) {

	//如果案件已委外不再升级案件
	orderExt, _ := models.GetOrderExt(oneCase.OrderId)
	// 如果案件评级没有变,则只更新逾期天数
	// 兼容案件生成天数变化，若已按旧配置生成更新一级 case，则不后退， 保持原case不变
	// 案件不后退, 无论配置如何改变，只升不降
	if oneCase.CaseLevel == caseLevel ||
		types.OverdueLevelCreateDaysMap()[oneCase.CaseLevel] >= types.OverdueLevelCreateDaysMap()[caseLevel] ||
		orderExt.IsEntrust == 1 {
		oneCase.OverdueDays = int(overdueDays)
		oneCase.Utime = tools.GetUnixMillis()
		if oneCase.IsOut == types.IsUrgeOutFrozen {
			oneCase.IsOut = types.IsUrgeOutNo
		}

		_, err := models.OrmUpdate(&oneCase, []string{"is_out", "overdue_days", "utime"})
		//_, err = o.Update(&oneCase, "is_out", "overdue_days", "utime")
		if err != nil {
			logs.Error("[doCaseUpdate] case update failed, oneCase:", oneCase, ", err:", err)
			return
		}
		return
	}
	// 案件调级

	// 出催
	oneCase.IsOut = types.IsUrgeOutYes
	oneCase.OutReason = types.UrgeOutReasonAdjust
	oneCase.OutUrgeTime = tools.GetUnixMillis()
	oneCase.Utime = oneCase.OutUrgeTime

	oneCaseJSON, _ := tools.JsonEncode(oneCase)
	_, err := models.OrmUpdate(&oneCase, []string{"is_out", "out_reason", "out_urge_time", "utime"})
	if err != nil {
		logs.Error("[HandleOverdueCase] 逾期案件调级出催出错了, oneCase:", oneCaseJSON, ", err:", err)
		// 理论上不会出错,需要往下走
	}
	logs.Informational("[HandleOverdueCase] 逾期案件调级出催, onCase:", oneCaseJSON)

	//案件升级，当前催收工单关闭，关闭原因：案件升级
	ticket.CloseByRelatedID(oneCase.Id, types.OverdueLevelTicketItemMap()[oneCase.CaseLevel], types.TicketCloseReasonCaseUp)

	doInvaildPreduced(oneCase.OrderId)

	createNewCase(oneCase.OrderId, caseLevel, overdueDays, userAccountID)
}

func doInvaildPreduced(orderID int64) {
	prereduced, err := dao.GetLastPrereducedByOrderid(orderID)
	if err != nil {
		logs.Warn("[doInvaildPreduced] no prereduced record", err)
		return
	}
	prereduced.ReduceStatus = types.ReduceStatusInvalid
	prereduced.InvalidReason = types.ClearReducedInvalidReasonCaseUp
	prereduced.ConfirmTime = tools.GetUnixMillis()
	prereduced.Utime = tools.GetUnixMillis()
	cols := []string{"reduce_status", "invalid_reason", "confirm_time", "Utime"}
	models.OrmUpdate(&prereduced, cols)
}

func GetOverdueCaseDetailList(overdueCaseId int64) (data []models.OverdueCaseDetail, err error) {
	data, err = models.GetMultiDatasByOverdueCaseId(overdueCaseId)
	return
}

func GetOverdueCaseDetailListByOrderId(orderId int64) (data []models.OverdueCaseDetail, err error) {
	data, err = models.GetMultiDatasByOrderId(orderId)

	return
}

type OverdueCaseDetails struct {
	Id                int64 `orm:"pk;"`
	OpUid             int64
	PhoneObject       int
	PhoneObjectMobile string
	PhoneTime         int64
	PhoneConnect      int
	PromiseRepayTime  int64
	OverdueReason     string
	OverdueReasonItem types.OverdueReasonItemEnum
	RepayInclination  int
	UnconnectReason   int
	Result            string

	AnswerTimestamp int64
	EndTimestamp    int64
	HangupDirection int
	HangupCause     int
	CallMethod      int
}

func GetOverdueCaseDetailListByOrderIds(orderId int64) (list []OverdueCaseDetails, err error) {

	obj := models.OverdueCaseDetail{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	// 初始化查询条件
	selectSql := fmt.Sprintf(`SELECT op_uid, phone_object, phone_object_mobile, phone_time, phone_connect, promise_repay_time, overdue_reason, overdue_reason_item, repay_inclination, unconnect_reason, result`)
	where := fmt.Sprintf(`where overdue_case_detail.order_id = %v`, orderId)
	sqlList := fmt.Sprintf(`%s FROM %s %s ORDER BY overdue_case_detail.id desc`, selectSql, obj.TableName(), where)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRows(&list)

	// 查询'通话记录'
	objSipCallRecord := models.SipCallRecord{}
	o.Using(objSipCallRecord.UsingSlave())

	selectSipCallRecord := `SELECT answer_timestamp, end_timestamp, hangup_direction, hangup_cause, call_method`
	for k, v := range list {
		if v.PhoneTime > 0 {
			var dataSipCallRecord OverdueCaseDetails
			whereSipCallRecord := fmt.Sprintf(`where start_timestamp = %d and call_method = 3`, v.PhoneTime)
			sql := fmt.Sprintf(`%s from %s %s`, selectSipCallRecord, objSipCallRecord.TableName(), whereSipCallRecord)
			r := o.Raw(sql)
			r.QueryRow(&dataSipCallRecord)

			if dataSipCallRecord.CallMethod == voip.VoipCallMethodSipCall {
				list[k].AnswerTimestamp = dataSipCallRecord.AnswerTimestamp
				list[k].EndTimestamp = dataSipCallRecord.EndTimestamp
				list[k].HangupCause = dataSipCallRecord.HangupCause
				list[k].HangupDirection = dataSipCallRecord.HangupDirection
				list[k].CallMethod = dataSipCallRecord.CallMethod
			} else {
				list[k].CallMethod = voip.VoipCallManual
			}
		} else {
			list[k].CallMethod = voip.VoipCallManual
		}
	}

	return
}
