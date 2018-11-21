package ticket

import (
	"encoding/json"
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/thirdparty/fantasy"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"sync"

	"github.com/astaxie/beego/logs"
)

// type assignFunc func(types.TicketItemEnum) int64

// DataUrge 描述 催收ticket 自定义数据data
type DataUrge struct {
	OrderID    int64
	CustomerID int64
}

// DataRepayRemindCase 描述 还款提醒 自定义数据data
type DataRepayRemindCase struct {
	OrderID    int64
	CustomerID int64
}

// CreateAfterRisk 进入等待电核状态后， 创建工单
func CreateAfterRisk(orderData models.Order) {
	itemID := types.TicketItemPhoneVerify
	if orderData.IsReloan == 0 {
		r := fantasy.NewSingleRequestByOrderPt(&orderData)
		score, _ := r.GetAScoreV1()
		infoReviewMinScore, _ := config.ValidItemInt("ticket_info_review_min_score")
		if infoReviewMinScore > 0 && score >= infoReviewMinScore {
			itemID = types.TicketItemInfoReview
		}
	} else {
		// 复贷直接进入info review
		itemID = types.TicketItemInfoReview
	}
	event.Trigger(&evtypes.TicketCreateEv{
		Item:       itemID,
		CreateUID:  types.Robot,
		RelatedID:  orderData.Id,
		OrderID:    orderData.Id,
		CustomerID: orderData.UserAccountId,
		Data:       nil})
}

// CreateTicket 创建工单
func CreateTicket(it types.TicketItemEnum, relatedID, createUID, orderID, customerID int64, data interface{}) (id int64, err error) {
	byteJSON, _ := json.Marshal(data)
	ticket := models.Ticket{
		ItemID:     it,
		RelatedID:  relatedID,
		CreateUID:  createUID,
		OrderID:    orderID,
		CustomerID: customerID,
		Data:       string(byteJSON),
		Priority:   types.TicketPirorityGeneral,
		Ctime:      tools.GetUnixMillis(),
		Status:     types.TicketStatusCreated,
	}
	jsonData, _ := json.Marshal(data)
	ticket.Data = string(jsonData)

	switch ticket.ItemID {
	case types.TicketItemPhoneVerify, types.TicketItemInfoReview:
		ticket.Link = fmt.Sprintf("/riskctl/phone_verify?order_id=%d", ticket.RelatedID)
	case types.TicketItemUrgeM11, types.TicketItemUrgeM12, types.TicketItemUrgeM13, types.TicketItemUrgeM20, types.TicketItemUrgeM30:
		rp, _ := models.GetLastRepayPlanByOrderid(orderID)
		ticket.ShouldRepayDate = rp.RepayDate
		ticket.ExpireTime = types.GetOverdueCaseExpireTime(ticket.ItemID, rp.RepayDate)
		ticket.Link = fmt.Sprintf("/overdue/urge?id=%d", ticket.RelatedID)
	case types.TicketItemRepayRemind:
		// 已废弃
		rp, _ := models.GetLastRepayPlanByOrderid(orderID)
		ticket.ExpireTime = rp.RepayDate + 2*24*3600*1000
		ticket.Link = fmt.Sprintf("/repay/remind_case/handle?id=%d", ticket.RelatedID)
	case types.TicketItemRMAdvance1, types.TicketItemRM0, types.TicketItemRM1:
		rp, _ := models.GetLastRepayPlanByOrderid(orderID)
		ticket.ShouldRepayDate = rp.RepayDate
		ticket.ExpireTime = types.GetRepayRemindCaseExpireTime(ticket.ItemID, rp.RepayDate)
		ticket.Link = fmt.Sprintf("/repay/remind_case/handle?id=%d", ticket.RelatedID)
		ticket.RiskLevel = calculateRiskLevel(orderID)
	default:
		logs.Warning("[Ticket-Create] No special handle for tikect item:", ticket.ItemID)
	}

	// 计算过期时间
	id, err = models.OrmInsert(&ticket)
	if id > 0 {
		enterWaitAssignQueue(id, ticket.ItemID, "LPUSH")
	}
	return
}

func isTicketFinished(ticket *models.Ticket) bool {
	switch ticket.Status {
	case types.TicketStatusClosed, types.TicketStatusCompleted:
		return true
	default:
		return false
	}
}

// IsTicketFinished 根据工单状态，判断工单是否完结
func IsTicketFinished(status types.TicketStatusEnum) bool {
	switch status {
	case types.TicketStatusClosed, types.TicketStatusCompleted:
		return true
	default:
		return false
	}
}

func BatchApplyEntrust(ids []int64) (succCount int64) {
	succCounter := struct {
		m     sync.Mutex
		count int64
	}{}
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(ticketID int64) {
			defer wg.Done()
			succ, _ := ApplyEntrust(ticketID)
			if succ {
				succCounter.m.Lock()
				succCounter.count++
				succCounter.m.Unlock()
			}
		}(id)
	}
	wg.Wait()
	succCount = succCounter.count
	return
}

func ApplyEntrust(ticketID int64) (res bool, err error) {
	ticket, err := models.GetTicket(ticketID)
	if err != nil {
		logs.Error("[ApplyEntrust] no ticket", err)
		return
	}
	oneCase, err1 := models.OneOverdueCaseByOrderID(ticket.OrderID)
	if err1 != nil {
		logs.Error("[ApplyEntrust] no overduecase", err1, "orderID:", ticket.OrderID)
		return
	}
	// 获取工单信息
	orderExt, _ := models.GetOrderExt(oneCase.OrderId)
	res = false
	//判断是否展示申请委外按钮
	if ApplyEntrustCondition(&ticket, &oneCase, &orderExt) {
		res, err = ApplyEntrustByTicketModel(&ticket)
	}
	return
}

// ApplyEntrustCondition 申请委外条件
func ApplyEntrustCondition(ticket *models.Ticket, overdueCase *models.OverdueCase, orderExt *models.OrderExt) (yes bool) {
	entrustDay, err := config.ValidItemInt("outsource_day")
	if err != nil {
		entrustDay = types.EntrustDay
		logs.Warning("[ApplyEntrustCondition] entrust day config losed:", entrustDay)
	}
	if (ticket.Status == types.TicketStatusAssigned ||
		ticket.Status == types.TicketStatusProccessing ||
		ticket.Status == types.TicketStatusPartialCompleted) &&
		overdueCase.OverdueDays >= entrustDay &&
		orderExt.IsEntrust == 0 {
		yes = true
	}
	return
}

// func HandleUpdate(ticket models.Ticket) {
// 	t := time.Now().Nanosecond()/time.Millisecond
//
// }
