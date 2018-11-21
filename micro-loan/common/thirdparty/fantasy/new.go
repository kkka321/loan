package fantasy

import (
	"fmt"
	"micro-loan/common/dao"
	"micro-loan/common/models"

	"github.com/astaxie/beego/logs"
)

// 先New Request
// 然后 GetAScoreV1  GetBScoreV1 ...

// NewSingleRequestByOrderID 返回请求 struct
func NewSingleRequestByOrderID(orderID int64) (r RiskRequestInfo, err error) {
	orderData, errQ := models.GetOrder(orderID)
	if errQ != nil {
		err = fmt.Errorf("[NewSingleRequestByOrderID] can not find order ,query err: %v, orderID:%d", errQ, orderID)
		return
	}

	return NewSingleRequestByOrderPt(&orderData), nil
}

// NewSingleRequestByOrderPt 返回请求 struct
func NewSingleRequestByOrderPt(orderDataPtr *models.Order) (r RiskRequestInfo) {
	accountBase, _ := models.OneAccountBaseByPkId(orderDataPtr.UserAccountId)
	accountProfile, _ := dao.CustomerProfile(orderDataPtr.UserAccountId)
	clientInfo, _ := models.OrderClientInfo(orderDataPtr.Id)

	return NewSingleRequest(orderDataPtr, &accountBase, accountProfile, &clientInfo)
}

// NewSingleRequest 返回请求 struct
func NewSingleRequest(oo *models.Order, uu *models.AccountBase, ap *models.AccountProfile, aci *models.ClientInfo) (r RiskRequestInfo) {
	detail := RiskRequestDetail{}
	FillFantasyRiskDetail(&detail, oo, uu, ap, aci)
	r.Data = append(r.Data, detail)
	return
}

// GetBScoreV1 根据 初始化后的RiskRequestInfo 获取A卡得分
// err log 已记录
func (r *RiskRequestInfo) GetBScoreV1() (score int, err error) {
	r.Model = "bscore"
	r.Version = "v1"
	// TODO  修改传值为传指针
	// 此处为保持原接口方法 GetFantasyRisk 不变， 故先使用传值
	_, _, riskResponse, err := GetFantasyRisk(*r)
	if err != nil {
		logs.Error("[fantasy.GetBScoreV1], err:", err, "request info:", r)
		return
	}
	if len(riskResponse.Data) > 0 {
		score = riskResponse.Data[0].Score
		return
	}
	err = fmt.Errorf("[fantasy.GetBScoreV1] err: response data is empty:%v, request info: %v", riskResponse.Data, r)
	logs.Error(err)
	return
}

// GetAScoreV1 根据 初始化后的RiskRequestInfo 获取A卡得分
// err log 已记录
func (r *RiskRequestInfo) GetAScoreV1() (score int, err error) {
	r.Model = "ascore"
	r.Version = "v1"
	_, _, riskResponse, err := GetFantasyRisk(*r)
	if err != nil {
		logs.Error("[fantasy.GetAScoreV1], err:", err, "request info:", r)
		return
	}
	if len(riskResponse.Data) > 0 {
		score = riskResponse.Data[0].Score
		return
	}
	err = fmt.Errorf("[fantasy.GetAScoreV1] err: response data is empty:%v, request info: %v", riskResponse.Data, r)
	logs.Error(err)
	return
}

// GetAScoreV2 根据 初始化后的RiskRequestInfo 获取A卡得分
// err log 已记录
func (r *RiskRequestInfo) GetAScoreV2() (score int, err error) {
	r.Model = "ascore"
	r.Version = "v2"
	_, _, riskResponse, err := GetFantasyRisk(*r)
	if err != nil {
		logs.Error("[fantasy.GetAScoreV2], err:", err, "request info:", r)
		return
	}
	if len(riskResponse.Data) > 0 {
		score = riskResponse.Data[0].Score
		return
	}
	err = fmt.Errorf("[fantasy.GetAScoreV2] err: response data is empty:%v, request info: %v", riskResponse.Data, r)
	logs.Error(err)
	return
}
