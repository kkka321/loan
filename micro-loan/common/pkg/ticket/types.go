package ticket

import (
	"micro-loan/common/models"
	"micro-loan/common/thirdparty/fantasy"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
)

// 工单风险评级
// 目前只有还款提醒使用
const (
	RiskLevelLow    = 10
	RiskLevelMedium = 20
	RiskLevelHigh   = 30
)

var riskLevelMap = map[int]string{
	RiskLevelLow:    "Low",
	RiskLevelMedium: "Medium",
	RiskLevelHigh:   "High",
}

// RiskLevelMap 返回风险评级与风险名称对应关系
func RiskLevelMap() map[int]string {
	return riskLevelMap
}

func calculateRiskLevel(orderID int64) int {
	orderData, _ := models.GetOrder(orderID)
	req := fantasy.NewSingleRequestByOrderPt(&orderData)
	if orderData.IsReloan == int(types.IsReloanYes) {
		s, _ := req.GetBScoreV1()
		if s >= 550 {
			return RiskLevelLow
		} else if s >= 460 && s <= 549 {
			return RiskLevelMedium
		} else {
			return RiskLevelHigh
		}
	}
	s, _ := req.GetAScoreV1()
	if s >= 618 {
		return RiskLevelLow
	} else if s >= 604 && s <= 617 {
		return RiskLevelMedium
	} else {
		return RiskLevelHigh
	}
}

func CalculateRiskScore(orderID int64) (s int) {
	orderData, _ := models.GetOrder(orderID)
	req := fantasy.NewSingleRequestByOrderPt(&orderData)
	if orderData.IsReloan == int(types.IsReloanYes) {
		s, _ = req.GetBScoreV1()
		logs.Info("[CalculateRiskScore] Get B score:", s, "isReloan:", orderData.IsReloan)
		return s
	}

	s, _ = req.GetAScoreV1()
	logs.Info("[CalculateRiskScore] Get A score:", s, "isReloan:", orderData.IsReloan)
	return s
}
