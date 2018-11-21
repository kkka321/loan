package runner

import (
	"fmt"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/models"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/repayremind"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty"
	"micro-loan/common/thirdparty/appsflyer"
	"micro-loan/common/thirdparty/xendit"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// event run place
// will be call by event consumer goroutine

// calls any outside and events

// to place all event here
// it will call by any where
// and call any where

//userActiveEv 异步运行用户激活事件
func userActiveEv(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.UserActiveEv); ok {
		//验证用户
		accountBase, _ := models.OneAccountBaseByPkId(e.AccountID)
		if accountBase.Id <= 0 {
			err = fmt.Errorf("[Event UserActiveEv run] account base 不存在, UserAccountId: %d", e.AccountID)
			logs.Error(err)
			return
		}
		//DoCustomerTags 用户激活后为用户打标签 ，2：目标客户
		service.DoCustomerTags(e.AccountID, int64(types.CustomerTagsTarget))

		//do more ...
	} else {
		err = fmt.Errorf("[userActiveEv] did not get a *evtypes.UserActiveEv persistent param: %T", param)
	}
	return
}

// orderApplyEv 用户申请订单异步打标签事件, 疑似与 LoanSubmitEv 类似
// 是否可以把 appsflyer 推送事件放入同一个事件里处理?
// 分离和合并各有好处?
func orderApplyEv(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.OrderApplyEv); ok {
		//基础验证
		order, _ := models.GetOrder(e.OrderID)
		if order.Id <= 0 {
			err = fmt.Errorf("Event OrderApplyEv run 订单不存在, OrderID: %d", e.OrderID)
			logs.Error(err)
			return
		}
		accountBase, _ := models.OneAccountBaseByPkId(order.UserAccountId)
		if accountBase.Id <= 0 || accountBase.Id != e.AccountID {
			err = fmt.Errorf("[Event OrderApplyEv run] account base 不存在, UserAccountId: %d", order.UserAccountId)
			logs.Error(err)
			return
		}

		//如果用户还未达到忠实用户(tags<5)，则启动打标签
		if accountBase.Tags < types.CustomerTagsLoyal {
			tags := service.DoReckonCustomerTags(e.AccountID)
			//DoCustomerTags 用户激活后为用户打标签 ，2：目标客户
			service.DoCustomerTags(e.AccountID, tags)
		}
	} else {
		err = fmt.Errorf("[orderApplyEv] did not get a *evtypes.OrderApplyEv persistent param: %T", param)
	}

	return
}

// orderAuditEv 用户借款订单提交审核事件run方法
func orderAuditEv(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.OrderAuditEv); ok {

		//基础验证
		order, _ := models.GetOrder(e.OrderID)
		if order.Id <= 0 {
			err = fmt.Errorf("Event OrderAuditEv run 订单不存在, OrderID: %d", e.OrderID)
			logs.Error(err)
			return
		}
		accountBase, _ := models.OneAccountBaseByPkId(order.UserAccountId)
		if accountBase.Id <= 0 || accountBase.Id != e.AccountID {
			err = fmt.Errorf("[Event OrderAuditEv run] account base 不存在, UserAccountId: %d", order.UserAccountId)
			logs.Error(err)
			return
		}

		//如果用户还未达到忠实用户(tags<5)，则启动打标签
		if accountBase.Tags < types.CustomerTagsLoyal {
			//获取用户全部订单
			tags := service.DoReckonCustomerTags(e.AccountID)
			//DoCustomerTags 用户激活后为用户打标签 ，2：目标客户
			service.DoCustomerTags(e.AccountID, tags)
		}
	} else {
		err = fmt.Errorf("[orderAuditEv] did not get a *evtypes.OrderAuditEv persistent param: %T", param)
	}

	return
}

// repaySuccessEv 异步运行用户还款事件
func repaySuccessEv(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.RepaySuccessEv); ok {
		//基础验证
		order, _ := models.GetOrder(e.OrderID)
		if order.Id <= 0 {
			err = fmt.Errorf("Event RepaySuccessEv run 订单不存在, OrderID: %d", e.OrderID)
			logs.Error(err)
			return
		}
		accountBase, _ := models.OneAccountBaseByPkId(order.UserAccountId)
		if accountBase.Id <= 0 || accountBase.Id != e.AccountID {
			err = fmt.Errorf("[Event RepaySuccessEv run] account base 不存在, UserAccountId: %d", order.UserAccountId)
			logs.Error(err)
			return
		}

		// 在redis中记录用户是否时复贷（已结清订单的个数）
		service.UpdateIsRepeatUser(e.AccountID)

		//如果用户还未达到忠实用户(tags<5)，则启动打标签
		if accountBase.Tags < types.CustomerTagsLoyal {
			//获取用户全部订单
			tags := service.DoReckonCustomerTags(e.AccountID)
			//DoCustomerTags 用户激活后为用户打标签 ，2：目标客户
			service.DoCustomerTags(e.AccountID, tags)
		}

		// try 关闭RM case，如果存在的话
		// 逾期case在 handleOverdueCase中处理
		repayremind.TryCompleteCaseByCleared(e.OrderID)

	} else {
		err = fmt.Errorf("[repaySuccessEv] did not get a *evtypes.RepaySuccessEv persistent param: %T", param)
	}

	return
}

// orderInvalidEv 异步运行订单失效事件
func orderInvalidEv(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.OrderInvalidEv); ok {
		//基础验证
		order, _ := models.GetOrder(e.OrderID)
		if order.Id <= 0 {
			err = fmt.Errorf("Event OrderInvalidEv run 订单不存在, OrderID: %d", e.OrderID)
			logs.Error(err)
			return
		}
		accountBase, _ := models.OneAccountBaseByPkId(order.UserAccountId)
		if accountBase.Id <= 0 || accountBase.Id != e.AccountID {
			err = fmt.Errorf("[Event OrderInvalidEv run] account base 不存在, UserAccountId: %d", order.UserAccountId)
			logs.Error(err)
			return
		}

		//如果用户还未达到忠实用户(tags<5)，则启动打标签
		if accountBase.Tags < types.CustomerTagsLoyal {
			//获取用户全部订单
			tags := service.DoReckonCustomerTags(e.AccountID)
			//DoCustomerTags 用户激活后为用户打标签 ，2：目标客户
			service.DoCustomerTags(e.AccountID, tags)
		}
	} else {
		err = fmt.Errorf("[orderInvalidEv] did not get a *evtypes.OrderInvalidEv persistent param: %T", param)
	}

	return
}

// blacklistEv 异步运行黑名单事件
func blacklistEv(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.BlacklistEv); ok {

		//基础验证
		accountBase, _ := models.OneAccountBaseByPkId(e.AccountID)
		if accountBase.Id <= 0 || accountBase.Id != e.AccountID {
			err = fmt.Errorf("[Event BlacklistEv run] account base 不存在, UserAccountId: %d", e.AccountID)
			logs.Error(err)
			return
		}

		if e.RiskItem > 0 && e.RiskVal != "" && e.Reason > 0 {
			service.DoAddBlacklist(e.AccountID, e.RiskItem, e.Reason, e.RiskVal, e.RiskMark)
		}
	} else {
		err = fmt.Errorf("[blacklistEv] did not get a *evtypes.BlacklistEv persistent param: %T", param)
	}

	return
}

// loanSubmitEv 运行申请贷款事件
func loanSubmitEv(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.LoanSubmitEv); ok {

		// 根据
		order, _ := models.GetOrder(e.OrderID)
		if order.Id <= 0 {
			err = fmt.Errorf("Event LoanSubmitEv  run order is nonexistence, OrderID: %d", e.OrderID)
			logs.Error(err)
			return
		}

		accountBase, _ := models.OneAccountBaseByPkId(order.UserAccountId)
		if accountBase.Id <= 0 {
			err = fmt.Errorf("[Event LoanSubmitEv run] account base is nonexistence, UserAccountId: %d", order.UserAccountId)
			logs.Error(err)
			return
		}

		eventReq := appsflyer.EventReq{
			AdvertisingID: accountBase.GoogleAdvertisingID,
			AppsflyerID:   accountBase.AppsflyerID,
			EventName:     appsflyer.AddCartEv,
			EventTime:     appsflyer.TimeFormat(e.Time),
			IsEventsAPI:   true,
			EventVal: appsflyer.AddCartEventVal{
				Price:    order.Loan,
				Currency: tools.GetServiceCurrency(),
			},
		}
		var errSend error
		success, errSend = eventReq.Send(e.OrderID, accountBase.StemFrom)
		if !success || errSend != nil {
			logs.Warn("[loanSubmitEv] appsflyer push event, orderID: %d, customerID: %d, error:%v", e.OrderID, accountBase.Id, errSend)
		}
	} else {
		err = fmt.Errorf("[loanSubmitEv] did not get a *evtypes.LoanSubmitEv persistent param: %T", param)
	}

	return
}

// loanSuccessEv 异步运行放款成功事件
func loanSuccessEv(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.LoanSuccessEv); ok {

		success = false

		order, _ := models.GetOrder(e.OrderID)
		if order.Id <= 0 {
			err = fmt.Errorf("Event LoanSuccessEv run 订单不存在, OrderID: %d", e.OrderID)
			logs.Error(err)
			return
		}

		accountBase, _ := models.OneAccountBaseByPkId(order.UserAccountId)
		if accountBase.Id <= 0 {
			err = fmt.Errorf("[Event LoanSuccessEv run] account base 不存在, UserAccountId: %d", order.UserAccountId)
			logs.Error(err)
			return
		}

		eventReq := appsflyer.EventReq{
			AdvertisingID: accountBase.GoogleAdvertisingID,
			AppsflyerID:   accountBase.AppsflyerID,
			EventName:     appsflyer.PurchaseEv,
			EventTime:     appsflyer.TimeFormat(e.Time),
			IsEventsAPI:   true,
			EventVal: appsflyer.PurchaseEventVal{
				Revenue:  order.Amount - order.Loan,
				Price:    order.Loan,
				Currency: tools.GetServiceCurrency(),
			},
		}
		var errSend error
		success, errSend = eventReq.Send(e.OrderID, accountBase.StemFrom)
		if !success || errSend != nil {
			//ordderDataJSON, _ := tools.JsonEncode(order)
			logs.Warn("[loanSuccessEv] appsflyer push event, orderID: %d, customerID: %d, error:%v", e.OrderID, accountBase.Id, errSend)
		}

		success, err = service.MarkCustomerIfHitRandom(order)
	} else {
		err = fmt.Errorf("[loanSuccessEv] did not get a *evtypes.LoanSuccessEv persistent param: %T", param)
	}
	return
}

// ticketCreateEv 运行事件,并创建工单
func ticketCreateEv(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.TicketCreateEv); ok {
		var id int64
		id, err = ticket.CreateTicket(e.Item, e.RelatedID, e.CreateUID, e.OrderID, e.CustomerID, nil)
		if id <= 0 {
			success = false
			return
		}
		success = true
	} else {
		err = fmt.Errorf("[ticketCreateEv] did not get a *evtypes.TicketCreateEv persistent param: %T", param)
	}
	return
}

// customerStatisticEv  用户统计信息 发生了api调用，更新用户的统计信息
func customerStatisticEv(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.CustomerStatisticEv); ok {
		return thirdparty.CustomerStatistic(e)
	} else {
		err = fmt.Errorf("[customerStatisticEv] did not get a *evtypes.customerStatisticEv persistent param: %T", param)
	}
	return
}

// workerDailyFirstOnlineEv 员工后台异步事件
func workerDailyFirstOnlineEv(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.WorkerDailyFirstOnlineEv); ok {
		ticket.AssignAfterDayFirstOnline(e.AdminUID, e.RoleID)
		success = true
	} else {
		err = fmt.Errorf("[workerDailyFirstOnlineEv] did not get a *evtypes.WorkerDailyFirstOnlineEv persistent param: %T", param)
	}
	return
}

// updateFixPaymentCodeAmount 更新固定付款码金额
func updateFixPaymentCodeAmount(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.FixPaymentCodeEv); ok {
		xendit.MarketPaymentCodeGenerate(e.OrderID, 0)
	} else {
		err = fmt.Errorf("[updateFixPaymentCodeAmount] did not get a *evtypes.FixPaymentCodeEv persistent param: %T", param)
	}
	return
}

// registerTrackEv 运行appsflyer 注册打点事件
func registerTrackEv(param interface{}) (success bool, err error) {
	if e, ok := param.(*evtypes.RegisterTrackEv); ok {
		eventReq := appsflyer.EventReq{
			AdvertisingID: e.GoogleAdvertisingID,
			AppsflyerID:   e.AppsflyerID,
			EventName:     appsflyer.CompleteRegistration,
			EventTime:     appsflyer.TimeFormat(e.Time),
			IsEventsAPI:   true,
			EventVal: appsflyer.CompleteRegistrationEventVal{
				RegMethod: "mobile",
			},
		}
		var errSend error
		success, errSend = eventReq.Send(e.AccountID, e.StemFrom)
		if !success || errSend != nil {
			logs.Warn("[registerTrackEv] appsflyer push event, AccountID: %d, error:%v", e.AccountID, errSend)
		}
	} else {
		err = fmt.Errorf("[registerTrackEv] did not get a *evtypes.RegisterTrackEv persistent param: %T", param)
	}

	return
}
