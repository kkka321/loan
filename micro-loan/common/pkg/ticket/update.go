package ticket

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
)

// ReopenByModel 根据已查询model打开工单
func ReopenByModel(ticketModel models.Ticket, nextCallTime string) bool {
	if !isTicketFinished(&ticketModel) {
		logs.Warning("[ReopenTicket] ticket is not finished ,so can't reopen: %d ", ticketModel.Id)
		return false
	}
	ticketModel.AssignUID = 0
	ticketModel.Status = types.TicketStatusCreated
	ticketModel.Utime = tools.GetUnixMillis()
	ticketModel.StartTime = 0
	ticketModel.AssignTime = 0
	ticketModel.CompleteTime = 0
	ticketModel.CloseTime = 0
	ticketModel.CloseReason = ""
	ticketModel.CustomerBestTime = nextCallTime
	cols := []string{"assign_uid", "status", "utime", "start_time", "assign_time", "complete_time", "close_time", "close_reason", "customer_best_time"}
	num, err := models.OrmUpdate(&ticketModel, cols)

	logs.Debug("[ReopenTicket] update ticket data, num:", num, "err:", err)
	if num > 0 && err == nil {
		//重新加入待分配队列
		enterWaitAssignQueue(ticketModel.Id, ticketModel.ItemID, "RPUSH")
	}

	return true
}

// ReopenPhoneVerifyOrInfoReviewByRelatedID 重新打开电核或者信息审查工单
func ReopenPhoneVerifyOrInfoReviewByRelatedID(relatedID int64, nextCallTime string) bool {
	ticketModel, err := models.GetTicketForPhoneVerifyOrInfoReivew(relatedID)
	if err != nil {
		logs.Error("[CompleteByRelatedID] query ticket, relatedID:%d, err: %s", relatedID, err.Error())
		return false
	}

	return ReopenByModel(ticketModel, nextCallTime)
}

// ReopenByRelatedID 根据相关case id ， 重新打开工单
func ReopenByRelatedID(relatedID int64, ticketItem types.TicketItemEnum, nextCallTime string) bool {
	ticketModel, err := models.GetTicketByItemAndRelatedID(ticketItem, relatedID)
	if err != nil || ticketModel.Id <= 0 {
		logs.Error("[ReopenTicket] query ticket :ticketItem:%d, relatedID:%d, err: %s",
			ticketItem, relatedID, err.Error())
		return false
	}

	return ReopenByModel(ticketModel, nextCallTime)
}

// UpdateByHandleCase 根据case处理记录, 更新ticket
func UpdateByHandleCase(caseID int64, itemID types.TicketItemEnum, handleTime, nextHandleTime int64, phoneObject int, communicationWay, isEmptyNumber int, caseLevel string) {
	tm, err := models.GetTicketByItemAndRelatedID(itemID, caseID)
	if err != nil {
		// 不存在的工单
		logs.Error("[ticket.UpdateByHandle] err:", err)
		return
	}
	cols := []string{"Utime", "LastHandleTime", "HandleNum"}
	tm.Utime = handleTime
	tm.LastHandleTime = handleTime
	tm.HandleNum++

	// hashKey := beego.AppConfig.String("hash:ticket_daily_handlenum")
	// storageClient := storage.RedisStorageClient.Get()
	// defer storageClient.Close()
	// qVal, _ := storageClient.Do("EXISTS", hashKey)
	// if 0 == qVal.(int64) {
	// 	storageClient.Do("HSET", hashKey, tm.OrderID, 1)
	// 	expireTimestamp, _ := tools.GetTodayTimestampByLocalTime("23:59:59")
	// 	storageClient.Do("EXPIREAT", hashKey, expireTimestamp)
	// } else {
	// 	hValue, _ := storageClient.Do("HGET", hashKey, tm.OrderID)
	// 	hadleNum, _ := tools.Str2Int(string(hValue.([]byte)))
	// 	hadleNum++
	// 	storageClient.Do("HSET", hashKey, tm.OrderID, hadleNum)
	// }

	if tm.StartTime == 0 {
		tm.StartTime = handleTime
		cols = append(cols, "StartTime")
	}
	if caseLevel != "" {
		tm.CaseLevel = caseLevel
		cols = append(cols, "CaseLevel")
	}
	if nextHandleTime > 0 {
		tm.NextHandleTime = nextHandleTime
		cols = append(cols, "NextHandleTime")
	}
	if phoneObject == types.PhoneObjectSelf {
		cols = append(cols, "CommunicationWay", "IsEmptyNumber")
		tm.CommunicationWay = communicationWay
		tm.IsEmptyNumber = isEmptyNumber
	}

	num, err := models.OrmUpdate(&tm, cols)
	if err != nil {
		logs.Error("[ticket.UpdateByHandleCase] update sql err:", err)
	}
	if num != 1 {
		logs.Error("[ticket.UpdateByHandleCase] update affected row err:", err)
	}
	return
}

// CloseByTicketModel 关闭ticket
func CloseByTicketModel(ticketPtr *models.Ticket, reason string) (bool, error) {
	return closeByTicketModel(ticketPtr, reason)
}

func closeByTicketModel(ticketPtr *models.Ticket, reason string) (bool, error) {
	if isTicketFinished(ticketPtr) {
		err := fmt.Errorf("[closeByTicketModel]ticket already finished, %v", ticketPtr)
		logs.Warn(err)
		return false, err
	}

	t := tools.GetUnixMillis()
	ticketPtr.CloseTime = t
	ticketPtr.Utime = t
	ticketPtr.Status = types.TicketStatusClosed
	ticketPtr.CloseReason = reason
	cols := []string{"Status", "Utime", "CloseTime", "CloseReason"}
	num, err := models.OrmUpdate(ticketPtr, cols)
	if err != nil || num <= 0 {
		logs.Error("[ticket.PartialComplete] Update failed with err:", err, ticketPtr)
		return false, err
	}

	afterFinish(ticketPtr)

	return true, nil
}

// CloseByRelatedID 根据相关ID和工单类型, 关闭工单
func CloseByRelatedID(relatedID int64, ticketItem types.TicketItemEnum, reason string) {
	ticketModel, err := models.GetTicketByItemAndRelatedID(ticketItem, relatedID)
	if err != nil || ticketModel.Id <= 0 {
		logs.Error("[CloseByRelatedID] query ticket :ticketItem:%d, relatedID:%d, err: %s",
			ticketItem, relatedID, err.Error())
		return
	}

	closeByTicketModel(&ticketModel, reason)
}

// CompleteByRelatedID 根据相关ID和工单类型, 自动完成工单
// 如果, 工单未开始, 则直接置为关闭状态, 关闭原因, types.TicketCloseReasonNoWork
func CompleteByRelatedID(relatedID int64, ticketItem types.TicketItemEnum) {
	ticketModel, err := models.GetTicketByItemAndRelatedID(ticketItem, relatedID)
	if err != nil {
		logs.Error("[CompleteByRelatedID] query ticket :ticketItem:%d, relatedID:%d, err: %s",
			ticketItem, relatedID, err.Error())
		return
	}

	completeByTicketModel(&ticketModel)
}

func afterFinish(ticketPtr *models.Ticket) {
	if ticketPtr.AssignUID > 0 {
		workerStrategy := GetItemWorkerStrategy(ticketPtr.ItemID)
		if iws, ok := workerStrategy.(*IdleWorkerStrategy); ok {
			iws.WorkingTicketChange(ticketPtr.AssignUID)
		}
	}

	if ticketPtr.AssignUID == 0 {
		// remove ticket from wait assign queue
		removeFromWaitAssignQueue(ticketPtr.Id, ticketPtr.ItemID)
	}
}

// CompletePhoneVerifyOrInfoReviewByRelatedID 根据相关ID和工单类型, 自动完成工单
// 如果, 工单未开始, 则直接置为关闭状态, 关闭原因, types.TicketCloseReasonNoWork
func CompletePhoneVerifyOrInfoReviewByRelatedID(relatedID int64) {
	ticketModel, err := models.GetTicketForPhoneVerifyOrInfoReivew(relatedID)
	if err != nil {
		logs.Error("[CompleteByRelatedID] query ticket, relatedID:%d, err: %s", relatedID, err.Error())
		return
	}

	completeByTicketModel(&ticketModel)
}

func completeByTicketModel(ticketPtr *models.Ticket) (bool, error) {
	if isTicketFinished(ticketPtr) {
		err := fmt.Errorf("[completeByTicketModel]ticket already finished, %v", ticketPtr)
		logs.Warn(err)
		return false, err
	}
	if ticketPtr.StartTime == 0 {
		// 如果工单未开始, 则直接关闭工单, 意味着工单未做, 所以不可置为已完成
		return closeByTicketModel(ticketPtr, types.TicketCloseReasonNoWork)
	}

	t := tools.GetUnixMillis()
	ticketPtr.CompleteTime = t
	ticketPtr.Utime = t
	ticketPtr.Status = types.TicketStatusCompleted
	cols := []string{"Status", "Utime", "CompleteTime"}
	num, err := models.OrmUpdate(ticketPtr, cols)
	if err != nil || num <= 0 {
		logs.Error("[ticket.completeByTicketModel] Update failed with err:", err, ticketPtr)
		return false, err
	}

	afterFinish(ticketPtr)
	return true, nil
}

func getCaseByOrder(updatedOrder *models.Order) (ticket models.Ticket, err error) {
	// 后期有 order所属索引, 可简化此处操作, TODO
	// 若已逾期则先查询有无逾期案件, 若有案件
	if updatedOrder.IsOverdue == types.IsOverdueYes {
		// overdue case ticket
		c, cErr := models.LatestValidOverdueCaseByOrderID(updatedOrder.Id)
		if cErr == nil {
			ticket, err = models.GetTicketByItemAndRelatedID(types.OverdueLevelTicketItemMap()[c.CaseLevel], c.Id)
			if err == nil {
				return
			}
		}
	}
	// 逾期1天 repay_remind_case ticket
	rc, rcErr := models.OneRepayRemindCaseByOrderID(updatedOrder.Id, types.StatusValid)
	if rcErr != nil {
		err = rcErr
		return
	}

	ticket, err = models.GetTicketByItemAndRelatedID(types.MustGetTicketItemIDByCaseName(rc.Level), rc.Id)
	return
}

// WatchPartialRepayment 监听部分还款事件
func WatchPartialRepayment(updatedOrder *models.Order) {
	ticket, err := getCaseByOrder(updatedOrder)
	if err != nil {
		logs.Error("[WatchPartialRepayment] query ticket by order: %v, err: %s", updatedOrder, err.Error())
		return
	}
	partialComplete(&ticket)

}

// PartialCompleteByRelatedID 根据相关ID和工单类型, 自动部分完成工单
func PartialCompleteByRelatedID(relatedID int64, ticketItem types.TicketItemEnum) {
	ticketModel, err := models.GetTicketByItemAndRelatedID(ticketItem, relatedID)
	if err != nil {
		logs.Error("[PartialCompleteByRelatedID] query ticket :ticketItem:%d, relatedID:%d, err: %s",
			ticketItem, relatedID, err.Error())
		return
	}

	partialComplete(&ticketModel)
}

func partialComplete(ticketPtr *models.Ticket) (bool, error) {
	// 工单未开始， 不可置为部分完成状态， 否则会影响工单自动分配
	if ticketPtr.StartTime == 0 {
		return false, fmt.Errorf("[partialComplete]ticket is not started, %v", ticketPtr)
	}
	if isTicketFinished(ticketPtr) {
		return false, fmt.Errorf("[partialComplete]ticket already finished, %v", ticketPtr)
	}

	oldStatus := ticketPtr.Status

	t := tools.GetUnixMillis()
	ticketPtr.PartialCompleteTime = t
	ticketPtr.Utime = t
	ticketPtr.Status = types.TicketStatusPartialCompleted
	num, err := models.OrmUpdate(ticketPtr, []string{"Status", "Utime", "PartialCompleteTime"})
	if err != nil || num <= 0 {
		logs.Error("[ticket.PartialComplete] Update failed with err:", err, ticketPtr)
		return false, err
	}
	// 触发worker持有工作中工单 sorted set 变化
	if ticketPtr.AssignUID > 0 && oldStatus != types.TicketStatusPartialCompleted {
		workerStrategy := GetItemWorkerStrategy(ticketPtr.ItemID)
		if iws, ok := workerStrategy.(*IdleWorkerStrategy); ok {
			iws.WorkingTicketChange(ticketPtr.AssignUID)
		}
	}
	return true, nil
}

// StartByTicketModel 关闭ticket
func StartByTicketModel(ticketPtr *models.Ticket) (bool, error) {
	return startByTicketModel(ticketPtr)
}

func startByTicketModel(ticketPtr *models.Ticket) (bool, error) {
	if isTicketFinished(ticketPtr) {
		return false, fmt.Errorf("[partialComplete]ticket already finished, %v", ticketPtr)
	}

	t := tools.GetUnixMillis()
	ticketPtr.StartTime = t
	ticketPtr.Utime = t
	ticketPtr.Status = types.TicketStatusProccessing
	num, err := models.OrmUpdate(ticketPtr, []string{"Status", "Utime", "StartTime"})
	if err != nil || num <= 0 {
		logs.Error("[ticket.startByTicketModel] Update failed with err:", err, ticketPtr)
		return false, err
	}

	return true, nil
}

func ApplyEntrustByTicketModel(ticketPtr *models.Ticket) (bool, error) {
	if isTicketFinished(ticketPtr) || ticketPtr.Status == types.TicketStatusCreated {
		return false, fmt.Errorf("[ApplyEntrustByTicketModel]ticket status not match %v", ticketPtr.Status)
	}
	t := tools.GetUnixMillis()
	ticketPtr.Utime = t
	ticketPtr.ApplyEntrustTime = t
	ticketPtr.Status = types.TicketStatusWaitingEntrust
	num, err := models.OrmUpdate(ticketPtr, []string{"Status", "ApplyEntrustTime", "Utime"})
	if err != nil || num <= 0 {
		logs.Error("[ticket.ApplyEntrustByTicketModel] Update failed with err:", err, ticketPtr)
		return false, err
	}
	return true, nil
}
