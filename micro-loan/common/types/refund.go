package types

// 退款状态,只为单一值,不存在组合值
type RefundStatus int

const (
	//0：无效值 1：退款中 2：退款失败 3：退款成功',
	// 1、：退款中
	RefundStatusProcessing RefundStatus = 1
	// 2、：退款失败
	RefundStatusFailed RefundStatus = 2
	// 3、：退款成功
	RefundStatusSuccess RefundStatus = 3
)

const (
	RefundTypeToOrder        = 1
	RefundTypeToBankCard     = 2
	RefundTypeToOtherAccount = 3
)
