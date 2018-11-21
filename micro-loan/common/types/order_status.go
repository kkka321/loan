package types

// 借款状态,只为单一值,不存在组合值
type LoanStatus int

const (
	// 1、已提交申请：客户提交借款申请后生成借款订单，状态为“已提交申请”状态，此状态30日内有效，30日内仍未提交借款审核的，置为“失效”
	LoanStatusSubmit LoanStatus = 1
	// 2、已提交审核：首贷客户完成客户资料的填写/复贷客户完成短信认证和活体认证，提交借款审核后，借款订单状态更新为“已提交审核”
	LoanStatus4Review LoanStatus = 2
	// 3、审核拒绝：借款订单在 反欺诈拒绝/人工审核拒绝/评分模型拒绝时，状态更新为“审核拒绝”
	LoanStatusReject LoanStatus = 3
	// 4、等待人工审核：借款订单在评分模型不足需要人工审核时，状态更新为“等待人工审核”
	LoanStatusWaitManual LoanStatus = 4
	// 5、等待放款：借款订单 反欺诈直批/人工审核通过/评分模型通过时，状态更新为“等待放款”
	LoanStatusWait4Loan LoanStatus = 5
	// 6、放款失败：借款订单在放款失败后，状态更新为“放款失败”；如后续选择重新放款，则状态更新为“等待放款”
	LoanStatusLoanFail LoanStatus = 6
	// 7、等待还款：借款订单在放款成功后，状态更新为“等待还款”
	LoanStatusWaitRepayment LoanStatus = 7
	// 8、已结清：借款订单在客户将应还本金、利息、服务费、罚息、滞纳金全部还完后，状态更新为“已结清”
	LoanStatusAlreadyCleared LoanStatus = 8
	// 9、逾期：借款订单在超过应还日期+宽限期仍未结清，状态更新为“逾期”
	LoanStatusOverdue LoanStatus = 9
	// 10、失效：用户提交借款申请后30日内仍未提交借款审核的，置为“失效”状态；用户存在其他提交审核的订单后，用户的其余订单自动置为失效
	LoanStatusInvalid LoanStatus = 10
	// 11.部分还款
	LoanStatusPartialRepayment LoanStatus = 11
	// 12. 放款中,程序关心的状态
	LoanStatusIsDoing LoanStatus = 12
	// 13. 等待第三方黑名单验证
	LoanStatusThirdBlacklistIsDoing LoanStatus = 13
	// 14. 等待展期(原订单的状态)
	LoanStatusRolling LoanStatus = 14
	// 15. 展期结清(原订单的状态)
	LoanStatusRollClear LoanStatus = 15
	// 16. 展期申请中(展期订单的状态)
	LoanStatusRollApply LoanStatus = 16
	// 17. 展期失效(展期订单的状态)
	LoanStatusRollFail LoanStatus = 17
	// 18. 等待自动外呼(infoReview工单在审核通过，A卡分在某区间后的订单状态)
	LoanStatusWaitAutoCall LoanStatus = 18
	// 19. 等待人脸比对（第三方黑名单通过之后，如果人脸比对开关打开时进入该状态）
	LoanStatusWaitPhotoCompare LoanStatus = 19
)

// 成功放款状态集
var succLoanStatusSlice = []LoanStatus{
	LoanStatusWait4Loan,
	LoanStatusLoanFail,
	LoanStatusWaitRepayment,
	LoanStatusAlreadyCleared,
	LoanStatusOverdue,
	LoanStatusPartialRepayment,
	LoanStatusIsDoing,
	LoanStatusRolling,
	LoanStatusRollClear,
}

// SuccLoanStatusSlice 返回 成功放款状态集
func SuccLoanStatusSlice() []LoanStatus {
	return succLoanStatusSlice
}

// ProcessingLoanStatus 贷款正在进行中的状态集合
var ProcessingLoanStatus = []LoanStatus{
	//	LoanStatusSubmit, //"已提交申请",
	LoanStatus4Review,       //"已提交审核",
	LoanStatusReject,        //"审核拒绝",
	LoanStatusWaitManual,    //"等待人工审核",
	LoanStatusWait4Loan,     //"等待放款",
	LoanStatusLoanFail,      //"放款失败",
	LoanStatusWaitRepayment, //"等待还款",
	//	LoanStatusAlreadyCleared:   "已结清",
	LoanStatusOverdue, //"逾期",
	//	LoanStatusInvalid:          "失效",
	LoanStatusPartialRepayment,      //"部分还款",
	LoanStatusIsDoing,               //"正在放款中",
	LoanStatusThirdBlacklistIsDoing, //"等待第三方黑名单验证",
	LoanStatusRollClear,             //"展期结清",
	LoanStatusRollApply,             //"展期申请中",
	LoanStatusRollFail,              //"展期失效",
	LoanStatusRolling,               //"等待展期",
	LoanStatusWaitAutoCall,          // "等待自动外呼"
	LoanStatusWaitPhotoCompare,      // "等待人脸比对（第三方黑名单通过之后，如果人脸比对开关打开时进入该状态）"
}

var allOrderStatusMap = map[LoanStatus]string{
	LoanStatusSubmit:                "已提交申请",
	LoanStatus4Review:               "已提交审核",
	LoanStatusReject:                "审核拒绝",
	LoanStatusWaitManual:            "等待人工审核",
	LoanStatusWait4Loan:             "等待放款",
	LoanStatusLoanFail:              "放款失败",
	LoanStatusWaitRepayment:         "等待还款",
	LoanStatusAlreadyCleared:        "已结清",
	LoanStatusOverdue:               "逾期",
	LoanStatusInvalid:               "失效",
	LoanStatusPartialRepayment:      "部分还款",
	LoanStatusIsDoing:               "正在放款中",
	LoanStatusThirdBlacklistIsDoing: "等待第三方黑名单验证",
	LoanStatusRollClear:             "展期结清",
	LoanStatusRollApply:             "展期申请中",
	LoanStatusRollFail:              "展期失效",
	LoanStatusRolling:               "等待展期",
	LoanStatusWaitAutoCall:          "等待自动外呼",
	LoanStatusWaitPhotoCompare:      "等待人脸比对",
}

func AllOrderStatusMap() map[LoanStatus]string {
	return allOrderStatusMap
}

// orderStatusMap 借款状态
var orderStatusMap = map[LoanStatus]string{
	LoanStatusSubmit: "已提交申请",
	//LoanStatus4Review:          "已提交审核",
	LoanStatusReject: "审核拒绝",
	//LoanStatusWaitManual:       "等待人工审核",
	LoanStatusWait4Loan: "等待放款",
	//LoanStatusLoanFail:         "放款失败",
	LoanStatusWaitRepayment:  "等待还款",
	LoanStatusAlreadyCleared: "已结清",
	LoanStatusOverdue:        "逾期",
	//LoanStatusInvalid:          "失效",
	//LoanStatusPartialRepayment: "部分还款",
	//LoanStatusIsDoing:          "正在放款中",
}

// OrderStatusMap 借款状态
func OrderStatusMap() map[LoanStatus]string {
	return orderStatusMap
}

// orderTypeMap 订单类型定义
// 订单本身没有类型之说,是多个字段组合而来
var orderTypeMap = map[string]string{
	"normal":    "普通订单",
	"temporary": "临时订单",
	"first":     "首贷",
	"repeat":    "复贷",
	"roll":      "展单",
	"overdue":   "历史逾期",
	"dead_debt": "坏帐",
}

func OrderTypeMap() map[string]string {
	return orderTypeMap
}

var loanStatusMap = map[LoanStatus]string{
	LoanStatusWait4Loan:        "等待放款",
	LoanStatusLoanFail:         "放款失败",
	LoanStatusIsDoing:          "正在放款中",
	LoanStatusWaitRepayment:    "等待还款",
	LoanStatusPartialRepayment: "部分还款",
	LoanStatusOverdue:          "逾期",
}

// LoanStatusMap 订单状态
func LoanStatusMap() map[LoanStatus]string {
	return loanStatusMap
}

var repayStatusMap = map[LoanStatus]string{
	LoanStatusWaitRepayment:    "等待还款",
	LoanStatusPartialRepayment: "部分还款",
	LoanStatusAlreadyCleared:   "已结清",
	LoanStatusOverdue:          "逾期",
	LoanStatusRolling:          "等待展期",
	LoanStatusRollClear:        "展期结清",
}

// RepayStatusMap 还款状态
func RepayStatusMap() map[LoanStatus]string {
	return repayStatusMap
}

const (
	DisbureStatusCallSuccess     = 1
	DisbureStatusCallFailed      = 2
	DisbureStatusCallUnknow      = 3
	DisbureStatusCallBackSuccess = 4
	DisbureStatusCallBackFailed  = 5
)

type HomeOrderType int

const (
	HomeOrderTypeLoaning                   HomeOrderType = 1 // 等待还款, 部分还款, 逾期, 展期
	HomeOrderTypePhoneVerify               HomeOrderType = 2 // 等待人工电核
	HomeOrderTypeCustomerRecallScore       HomeOrderType = 3 // 评分模型需召回的客户首页
	HomeOrderTypeCustomerRecallPhoneVerify HomeOrderType = 4 // 电核环节需召回的客户首页
	HomeOrderTypeModifyBank                HomeOrderType = 5 // 因账号信息错误导致放款失败,需修改信息
)

const (
	RepayTypeVa          = 1
	RepayTypePaymentCode = 2
)

var repayTypeMap = map[int]string{
	RepayTypeVa:          "VA",
	RepayTypePaymentCode: "超市付款码",
}

// RepayStatusMap 还款状态
func RepayTypeMap() map[int]string {
	return repayTypeMap
}
