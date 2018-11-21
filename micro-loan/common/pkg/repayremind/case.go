package repayremind

import (
	"fmt"
	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/thirdparty/fantasy"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
)

// CreateCase 创建人工提醒事件
func CreateCase(orderID int64, level string, userAccountID int64) (id int64, err error) {
	c := models.RepayRemindCase{}
	c.OrderId = orderID
	c.Level = level
	c.UserAccountId = userAccountID
	c.Ctime = tools.GetUnixMillis()
	c.Status = StatusValid
	id, err = models.OrmInsert(&c)
	if err != nil {
		logs.Error("[repayremind.CreateCase] err:", err)
	}
	if id > 0 {
		ticket.CreateTicket(types.MustGetTicketItemIDByCaseName(level), id, types.Robot, orderID, userAccountID, nil)
	}
	return
}

func expireCaseByModel(oneCase *models.RepayRemindCase) {
	oneCase.Status = types.StatusInvalid
	oneCase.InvalidReason = models.RepayRemindInvalidReasonExpired
	oneCase.InvalidTime = tools.GetUnixMillis()
	oneCase.Utime = oneCase.InvalidTime

	_, err := models.OrmUpdate(oneCase, []string{"Status", "InvalidReason", "InvalidTime", "Utime"})
	if err != nil {
		logs.Error("[expireCaseByModel] sql update err, oneCase:", oneCase, ", err:", err)
		// 理论上不会出错,需要往下走
	}

	//案件升级，当前催收工单关闭，关闭原因：案件升级
	ticket.CloseByRelatedID(oneCase.Id, types.MustGetTicketItemIDByCaseName(oneCase.Level), types.TicketCloseReasonCaseUp)
}

// ExpireCaseByOrderID 触发RM case过期
func ExpireCaseByOrderID(orderID int64) {
	oneCase, err := models.OneVaildRepayRemindCaseByOrderID(orderID)
	if err != nil {
		logs.Error("[repayremind.ExpireCaseByOrderID] valid case should be find,but not, err:", err)
		return
	}
	expireCaseByModel(&oneCase)
}

// DailyHandleCase 处理订单的逾期案件
func DailyHandleCase(orderID int64) (err error) {
	orderData, err := models.GetOrder(orderID)
	if err != nil {
		logs.Error("[DailyHandleCase] cannot find order ,query err:", err)
		return
	}

	if orderData.CheckStatus == types.LoanStatusAlreadyCleared ||
		orderData.CheckStatus == types.LoanStatusRollClear {
		logs.Warn("[DailyHandleCase] order already cleared, query err:", err)
		return
	}

	// 计算逾期天数和逾期case level
	repayPlan, err := models.GetLastRepayPlanByOrderid(orderData.Id)
	if err != nil || repayPlan.RepayDate <= 0 {
		logs.Error("[doCaseUpdate] 还款计划数据有误, orderID: %d, repayPlan: %s, err: %v", orderData.Id, repayPlan, err)
		return
	}
	caseLevel, _, err := calculateLevel(repayPlan.RepayDate)
	if caseLevel == types.RMLevelAdvance1 && checkScoreJump(&orderData) {
		return
	}

	if err != nil {
		logs.Error("[DailyHandleCase] should not happened in CalculateLevel, err:", err)
		return
	}
	// 获取当前未出催逾期case, 对case 进行新建, 更新, 出催, 并同时更新 工单
	oneCase, errQuery := models.OneVaildRepayRemindCaseByOrderID(orderData.Id)

	if errQuery != nil {
		// 说明没有对应的案件,生成之
		CreateCase(orderData.Id, caseLevel, orderData.UserAccountId)
		return
	}

	if oneCase.Level != caseLevel && types.RMCaseCreateDaysMap()[oneCase.Level] < types.RMCaseCreateDaysMap()[caseLevel] {
		// 失效旧案件
		expireCaseByModel(&oneCase)

		CreateCase(oneCase.OrderId, caseLevel, orderData.UserAccountId)
	}
	return
}

func checkScoreJump(orderDataPtr *models.Order) bool {
	riskReq := fantasy.RiskRequestInfo{}
	accountBase, _ := models.OneAccountBaseByPkId(orderDataPtr.UserAccountId)
	accountProfile, _ := dao.CustomerProfile(orderDataPtr.UserAccountId)
	clientInfo, _ := models.OrderClientInfo(orderDataPtr.Id)
	fantasy.FillFantasyRiskRequest(&riskReq, orderDataPtr, &accountBase, accountProfile, &clientInfo)

	if orderDataPtr.IsReloan == int(types.IsReloanYes) {
		// B score
		riskReq.Model = "bscore"
		riskReq.Version = "v1"
		_, _, riskB, _ := fantasy.GetFantasyRisk(riskReq)
		jumpScoreB, _ := config.ValidItemInt("repay_remind_case_jump_min_bscore")
		logs.Debug("[checkScoreJump] Score B:", jumpScoreB)
		if len(riskB.Data) > 0 && riskB.Data[0].Score >= jumpScoreB {
			return true
		}
		return false
	}
	// A v1 score
	riskReq.Model = "ascore"
	riskReq.Version = "v1"
	_, _, riskA, _ := fantasy.GetFantasyRisk(riskReq)
	jumpScoreA, _ := config.ValidItemInt("repay_remind_case_jump_min_ascore")
	logs.Debug("[checkScoreJump] Score A:", jumpScoreA)
	if len(riskA.Data) > 0 && riskA.Data[0].Score >= jumpScoreA {
		return true
	}
	return false
}

func calculateLevel(reapyDate int64) (level string, days int64, err error) {
	today := tools.NaturalDay(0)

	days = (today - reapyDate) / 86400000
	logs.Debug("days: %d", days)

	// 获取此应还日期，最新的case
	var lastStartDay = -9999
	for l, startDay := range types.RMCaseCreateDaysMap() {
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

	return
}

// TryCompleteCaseByCleared 尝试完成case,如果存在有效case的情况下
func TryCompleteCaseByCleared(orderID int64) {
	oneCase, err := models.OneVaildRepayRemindCaseByOrderID(orderID)
	if err != nil {
		logs.Debug("[TryCompleteCaseByCleared] cannot find valid rm case: %v, orderID: %d", err, orderID)
		return
	}
	oneCase.Status = types.StatusInvalid
	oneCase.InvalidReason = models.RepayRemindInvalidReasonCleared
	oneCase.InvalidTime = tools.GetUnixMillis()
	oneCase.Utime = oneCase.InvalidTime
	models.OrmUpdate(&oneCase, []string{"Status", "InvalidReason", "InvalidTime", "Utime"})

	itemID := types.MustGetTicketItemIDByCaseName(oneCase.Level)
	ticket.CompleteByRelatedID(oneCase.Id, itemID)
}
