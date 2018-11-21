/**
* 期望专门负责event事件处理的服务，降低event文件的臃肿
*
*
**/

package service

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// DoCustomerTags 为客户打标签
func DoCustomerTags(accountID, tags int64) (num int64, err error) {
	num, err = UpdateCustomer(accountID, tags)
	return
}

// DoReckonCustomerTags 计算用户tags并返回
func DoReckonCustomerTags(accountID int64) (tags int64) {
	tags = CustomerTags(accountID)
	return
}

// DoAddBlacklist 系统加入黑名单riskItem types.RiskItemEnum, riskType types.RiskTypeEnum,
func DoAddBlacklist(accountID int64, riskItem types.RiskItemEnum, reasion types.RiskReason, riskVal string, riskMark string) {

	//操作员ID
	AddCustomerRisk(
		accountID,
		0,
		riskItem,
		types.RiskBlacklist,
		reasion,
		riskVal,
		riskMark,
		types.RiskReviewPass,
		tools.GetUnixMillis(),
		"",
		"")
	return
}
