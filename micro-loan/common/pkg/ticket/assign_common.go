package ticket

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
)

// CanAssignUsers 获取指定Ticket可分配用户列表
func CanAssignUsers(id int64) (admins []models.Admin, num int64, err error) {
	ticket, err := models.GetTicket(id)
	if err != nil {
		logs.Error("[Data Miss][Ticket] id:", id)
		return
	}

	admins, num, err = canAssignUsersByTicketItem(ticket.ItemID)

	return
}

// ManualCanAssignUsers 获取指定Ticket可分配用户列表
func ManualCanAssignUsers(id, roleID, opUID int64, roleType types.RoleTypeEnum) (admins []models.Admin, num int64, err error) {
	ticket, err := models.GetTicket(id)
	if err != nil {
		logs.Error("[Data Miss][Ticket] id:", id)
		return
	}

	// get  user level

	// is filter  by roles (role id and his child role ids)

	// get the admins
	admins, num, err = canAssignUsersByTicketItem(ticket.ItemID)

	// ids := admin.GetLeaderManageUsers(roleID, opUID)
	//
	// idsMap := tools.SliceInt64ToMap(ids)
	// filterAdmins := make([]models.Admin{}, len(idsMap))
	// for _, m := range admins {
	// 	if _, ok := idsMap[m.Id]; ok {
	// 		filterAdmins = append(filterAdmins, m)
	// 	}
	// }

	return
}

func activeCanAssignUsersByTicketItem(ticketItem types.TicketItemEnum) (activeAdmins []models.Admin) {
	admins, num, err := canAssignUsersByTicketItem(ticketItem)

	if num == 0 {
		logs.Error("[canAssignUsersByTicketItem] return 0 admins for assign ticket item: %d", ticketItem, "with err:", err)
		return
	}

	if err != nil {
		logs.Error("[activeCanAssignUsersByTicketItem]", err)
	}
	for _, admin := range admins {
		if admin.WorkStatus == types.AdminWorkStatusNormal {
			if IsWorkerOnline(admin.Id) {
				activeAdmins = append(activeAdmins, admin)
			}
		}
	}
	if len(activeAdmins) == 0 && num != 0 {
		logs.Warn("[activeCanAssignUsersByTicketItem] ticket item(%d) have %d workers, but no one active", ticketItem, num)
	}
	return
}

func canAssignUsersByTicketItem(ticketItem types.TicketItemEnum) (admins []models.Admin, num int64, err error) {

	idstrings, err := canAssignRoles(ticketItem)
	if err != nil || len(idstrings) <= 0 {
		err = fmt.Errorf("[canAssignUsersByTicketItem] no roles or config err: %v", err)
		return
	}
	admins, num, err = models.GetUsersByRoleIDStringsFromDB(idstrings)

	return
}

// CanAssignUsersByTicketItem 根据 ticketItem 获取可分配后台用户
func CanAssignUsersByTicketItem(ticketItem types.TicketItemEnum) (admins []models.Admin, num int64, err error) {
	return canAssignUsersByTicketItem(ticketItem)
}

// 根据工单类型获取可分配角色列表
func canAssignRoles(ticketItem types.TicketItemEnum) (idStrings []string, err error) {
	//var roleConfigName string
	roleConfigName, exist := types.TicketAssignRoleConfigNameMap()[ticketItem]
	if !exist {
		err = fmt.Errorf("[canAssignRoles] No assign role configname for ticket item: %d", ticketItem)
		//logs.Error(err)
		return
	}

	idStrings, err = getAssignRolesByConfigName(roleConfigName)

	return
}

// CanAssignRoles 获取 Ticket Item 对应的角色列表
func CanAssignRoles(ticketItem types.TicketItemEnum) (idStrings []string, err error) {
	return canAssignRoles(ticketItem)
}
