package ticket

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/pkg/privilege"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"
	"strings"
	"sync"

	"github.com/astaxie/beego/logs"
)

func getWorkTicketItemsByRoleID(roleID int64) (ticketItems []types.TicketItemEnum) {
	ticketAssignRoleConfigNameMap := types.TicketAssignRoleConfigNameMap()
	roleIDString := strconv.FormatInt(roleID, 10)
	for ticketItem, roleConfigName := range ticketAssignRoleConfigNameMap {
		idStrings, _ := getAssignRolesByConfigName(roleConfigName)
		for _, id := range idStrings {
			if id == roleIDString {
				ticketItems = append(ticketItems, ticketItem)
				break
			}
		}
	}
	return
}

func getAssignRolesByConfigName(roleConfigName string) (idStrings []string, err error) {
	roleConfig := config.ValidItemString(roleConfigName)
	if len(roleConfig) <= 0 {
		err = fmt.Errorf("[getAssignRolesByConfigName] No role config in system config for roleConfigName: %s", roleConfigName)
		return
	}
	idStrings = strings.Split(roleConfig, ",")
	return
}

// ManualBatchAssign 手动批量分单
func ManualBatchAssign(ids []int64, adminUID, opUID int64) (succCount int64) {
	succCounter := struct {
		m     sync.Mutex
		count int64
	}{}
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(ticketID, adminUID int64) {
			defer wg.Done()
			succ, _ := ManualAssign(ticketID, adminUID, opUID)
			if succ {
				succCounter.m.Lock()
				succCounter.count++
				succCounter.m.Unlock()
			}
		}(id, adminUID)
	}
	wg.Wait()
	succCount = succCounter.count
	return
}

// ManualAssign 使用 ticket id 手动分配工单
func ManualAssign(id, adminUID, opUID int64) (res bool, err error) {
	ticket, err := models.GetTicket(id)
	if err != nil {
		logs.Error("[ManualAssign]", err)
		return
	}

	if adminUID <= 0 {
		err = fmt.Errorf("[ManualAssign] no user to assign, ticket:%v", ticket)
		return
	}

	if adminUID == ticket.AssignUID {
		err = fmt.Errorf("[ManualAssign] same user to assign, or already assign to admin(%d), ticket:%v", adminUID, ticket)
		return
	}
	res = assignByTicketData(&ticket, adminUID, opUID)
	return
}

// AutoAssignByTicketID 分单
func AutoAssignByTicketID(id, adminUID int64) bool {
	ticket, err := models.GetTicket(id)
	if err != nil {
		logs.Error("[Data Miss][Ticket] id:", id)
		return false
	}
	if ticket.AssignUID != 0 {
		logs.Error("[Data Miss][Ticket] Ticket alreay assigned, dont't auto assigned again, id:", id)
		return false
	}
	if ticket.Status != types.TicketStatusCreated {
		logs.Error("[AutoAssignByTicketID] ticket not on wait assign status, ticket:", ticket)
		return false
	}
	if adminUID <= 0 {
		logs.Error("[AutoAssignByTicketID] no user to assign, ticket:", ticket)
		return false
	}
	return assignByTicketData(&ticket, adminUID, types.Robot)
}

// assignByTicketData 分配工单
func assignByTicketData(ticket *models.Ticket, adminUID, opUID int64) bool {

	if isTicketFinished(ticket) {
		logs.Warning("[assignByTicketData] finish ticket reassign : %d ,assign by: %d", ticket.Id, opUID)
		return false
	}
	oldTicket := *ticket

	oldAssignUID := ticket.AssignUID

	switch ticket.ItemID {
	case types.TicketItemPhoneVerify, types.TicketItemInfoReview:
		if ticket.OrderID > 0 {
			privilege.GrantOrder(ticket.OrderID, adminUID)
		}
		if ticket.CustomerID > 0 {
			privilege.GrantCustomer(ticket.CustomerID, adminUID)
		}
	case types.TicketItemUrgeM11, types.TicketItemUrgeM12, types.TicketItemUrgeM13, types.TicketItemUrgeM20, types.TicketItemUrgeM30:
		// 未避免, ticket新旧数据 orderID 和 CustomerID 未及时补入, 先从原始表case表中查取关键数据
		// 二次更则直接使用冗余数据 ticket.OrderID 和 ticket.CustomerID
		privilege.GrantOverdueCase(ticket.RelatedID, adminUID)

		if ticket.OrderID > 0 {
			privilege.GrantOrder(ticket.OrderID, adminUID)
		}
		if ticket.CustomerID > 0 {
			privilege.GrantCustomer(ticket.CustomerID, adminUID)
		}
	case types.TicketItemRepayRemind, types.TicketItemRMAdvance1, types.TicketItemRM0, types.TicketItemRM1:
		privilege.GrantRepayRemindCase(ticket.RelatedID, adminUID)

		if ticket.OrderID > 0 {
			privilege.GrantOrder(ticket.OrderID, adminUID)
		}
		if ticket.CustomerID > 0 {
			privilege.GrantCustomer(ticket.CustomerID, adminUID)
		}
	default:
		logs.Warning("[Ticket-Assign] No special privileges assign or handle for tikect item:", ticket.ItemID)
	}
	ticket.AssignTime = tools.GetUnixMillis()
	ticket.AssignUID = adminUID
	ticket.Status = types.TicketStatusAssigned
	ticket.Utime = ticket.AssignTime
	num, err := models.OrmAllUpdate(ticket)
	if err != nil || num != 1 {
		logs.Error("[Ticket-Assign] Error, update failed,ticket:", ticket, "affected rows:", num, ";err:", err)
		return false
	}
	models.OpLogWrite(opUID, ticket.Id, models.OpCodeTicketAssign, ticket.TableName(), oldTicket, *ticket)

	{
		// 特定策略下, 触发事件
		workerStrategy := GetItemWorkerStrategy(ticket.ItemID)
		if iws, ok := workerStrategy.(*IdleWorkerStrategy); ok {
			iws.WorkingTicketChange(ticket.AssignUID)
			if oldAssignUID != 0 {
				iws.WorkingTicketChange(oldAssignUID)
			}
		}
	}

	return true
}
