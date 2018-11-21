package privilege

// 暂时未用到

import (
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// GrantLimitDataPrivilege ..
func GrantLimitDataPrivilege(dataID, grantTo int64, grantType types.DataGrantTypeEnum) {
	dp := models.LimitDataPrivilege{
		Type:      types.LimitDataPrivilegeTypeTicketItem,
		DataID:    dataID,
		GrantType: grantType,
		GrantTo:   grantTo,
		Ctime:     tools.GetUnixMillis(),
	}
	models.OrmInsert(&dp)
}
