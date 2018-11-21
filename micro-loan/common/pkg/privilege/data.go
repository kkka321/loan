package privilege

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// GrantOrder ..
func GrantOrder(orderID, grantTo int64) {
	grantData(types.DataPrivilegeTypeOrder, orderID, grantTo)
}

// GrantOverdueCase ...
func GrantOverdueCase(dataID, grantTo int64) {
	grantData(types.DataPrivilegeTypeOverdueCase, dataID, grantTo)
}

// GrantCustomer ..
func GrantCustomer(accountID, grantTo int64) {
	grantData(types.DataPrivilegeTypeCustomer, accountID, grantTo)
}

// GrantRepayRemindCase ..
func GrantRepayRemindCase(dataID, grantTo int64) {
	grantData(types.DataPrivilegeTypeRepayRemindCase, dataID, grantTo)
}

func grantData(dataPrivilegeType types.DataPrivilegeTypeEnum, dataID, grantTo int64) {
	dp := models.DataPrivilege{
		Type:      dataPrivilegeType,
		DataID:    dataID,
		GrantType: types.DataGrantUser,
		GrantTo:   grantTo,
		Ctime:     tools.GetUnixMillis(),
		Status:    1,
	}
	models.OrmInsert(&dp)
}

// IsGrantedDataPrivilege 是否已授权该数据权限
func IsGrantedDataPrivilege(dataType types.DataPrivilegeTypeEnum, dataID int64, grantTo []string) bool {
	m := models.DataPrivilege{}
	o := orm.NewOrm()
	o.Using(m.UsingSlave())

	// 初始化查询条件
	wheres := []string{}
	wheres = append(wheres, fmt.Sprintf("type=%d", dataType))
	wheres = append(wheres, fmt.Sprintf("data_id=%d", dataID))
	wheres = append(wheres, fmt.Sprintf("grant_type=%d", types.DataGrantUser))
	wheres = append(wheres, fmt.Sprintf("grant_to in(%s)", strings.Join(grantTo, ",")))
	wheres = append(wheres, fmt.Sprintf("is_deleted=%d", types.DeletedNo))
	wheres = append(wheres, fmt.Sprintf("status=%d", 1))

	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` WHERE %s", m.TableName(), strings.Join(wheres, " AND "))

	var count int64
	r := o.Raw(sqlCount)
	r.QueryRow(&count)
	logs.Debug("[IsGrantedDataPrivilege] count:", count)
	if count > 0 {
		return true
	}

	return false
}

// IsGrantedData 根据用户ID, 角色信息, 判断是否已拥有单条数据权限
func IsGrantedData(dataType types.DataPrivilegeTypeEnum, dataID int64,
	uid int64, roleID int64, rolePid int64) bool {

	if rolePid == types.SuperRolePid {
		return true
	}
	var sharePrivilegeUids []string
	models.GetUserIDsByRolePidFromDB(roleID, &sharePrivilegeUids)
	sharePrivilegeUids = append(sharePrivilegeUids, strconv.FormatInt(uid, 10))
	return IsGrantedDataPrivilege(dataType, dataID, sharePrivilegeUids)
}
