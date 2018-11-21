package evtypes

// anyPkg --import/call->trigger-new/import

import (
	"micro-loan/common/types"
)

//UserActiveEv 用户激活事件
type UserActiveEv struct {
	AccountID int64 `json:"aid"`
	Time      int64 `json:"time"`
}

// OrderApplyEv 借款订单提交申请事件
type OrderApplyEv struct {
	AccountID int64 `json:"aid"`
	OrderID   int64 `json:"order_id"`
	Time      int64 `json:"time"`
}

// OrderAuditEv 借款订单提交审核事件
type OrderAuditEv struct {
	AccountID int64 `json:"aid"`
	OrderID   int64 `json:"order_id"`
	Time      int64 `json:"time"`
}

// LoanSubmitEv 借款提交事件
type LoanSubmitEv struct {
	OrderID int64 `json:"i"`
	Time    int64 `json:"t"`
}

// LoanSuccessEv 放款成功事件
type LoanSuccessEv struct {
	OrderID int64 `json:"i"`
	Time    int64 `json:"t"`
}

// RepaySuccessEv 还款成功事件事件
type RepaySuccessEv struct {
	AccountID int64 `json:"aid"`
	OrderID   int64 `json:"order_id"`
	Time      int64 `json:"time"`
}

// OrderInvalidEv 订单失效事件
type OrderInvalidEv struct {
	AccountID int64 `json:"aid"`
	OrderID   int64 `json:"order_id"`
	Time      int64 `json:"time"`
}

// BlacklistEv 黑名单事件
type BlacklistEv struct {
	AccountID int64              `json:"aid"`
	RiskItem  types.RiskItemEnum `json:"risk_item"`
	RiskVal   string             `json:"risk_val"`
	Reason    types.RiskReason   `json:"reason"`
	RiskMark  string             `json:"risk_mark"`
}

// TicketCreateEv 创建工单
type TicketCreateEv struct {
	Item       types.TicketItemEnum   `json:"i"`
	CreateUID  int64                  `json:"c"`
	RelatedID  int64                  `json:"r"`
	OrderID    int64                  `json:"oid"`
	CustomerID int64                  `json:"cid"`
	Data       map[string]interface{} `json:"d"`
}

// CustomerStatisticEv 用户统计信息
type CustomerStatisticEv struct {
	UserAccountId int64  `json:"uid"`
	OrderId       int64  `json:"oid"`
	ApiMd5        string `json:"api_md5"`
	Fee           int64  `json:"fee"`
	Result        int    `json:"r"`
	MessageFlag   bool   `json:"m"` // 1 是短信  0 非短信消息
}

// WorkerDailyFirstOnlineEv 员工后台登录异步事件
type WorkerDailyFirstOnlineEv struct {
	AdminUID int64 `json:"id"`
	RoleID   int64 `json:"rid"`
}

//更新逾期付款码金额
type FixPaymentCodeEv struct {
	OrderID int64 `json:"order_id"`
}

// RegisterTrackEv 注册推送appsflyer事件
type RegisterTrackEv struct {
	AccountID           int64  `json:"uid"`
	StemFrom            string `json:"stem"`
	AppsflyerID         string `json:"afid"`
	GoogleAdvertisingID string `json:"gaid"`
	Time                int64  `json:"t"`
}
