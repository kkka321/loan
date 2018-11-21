package performance

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/pkg/rbac"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"sort"
	"strconv"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type userItemHourlyStats struct {
	Uid int64
	Num int64
}

// LastDayHourlyWorkerPerformanceStatsTask 执行最近一天的统计
func TodayHourlyWorkerPerformanceStatsTask() {
	// 获取昨天的日期

	hour := getHour()
	logs.Notice("hour:", hour)
	HourlyWorkerPerformanceStatsTask(hour)
}

// HourlyWorkerPerformanceStatsTask worker日绩效统计
// hour 格式 2018070112
func HourlyWorkerPerformanceStatsTask(hour string) {
	day := getToday()
	needStatsTicketItem := []types.TicketItemEnum{
		types.TicketItemUrgeM11,
		types.TicketItemUrgeM12,
		// types.TicketItemRMAdvance1,
		types.TicketItemRM0,
		types.TicketItemRM1,
		//types.TicketItemUrgeM13,
		//types.TicketItemRepayRemind,
	}
	for _, ticketItem := range needStatsTicketItem {
		NewItemHourlyStatsAndRun(day, hour, ticketItem)
	}

}

type ticketItemHourly struct {
	ticketItem            types.TicketItemEnum
	hour                  int
	dayStartTimeStamp     int64
	startTimestamp        int64
	endTimestamp          int64
	baseModel             models.TicketWorkerHourlyStats
	statsData             map[int64]models.TicketWorkerHourlyStats
	targetRepayAmountRate float64 // 目标回收率
}

// NewItemHourlyStatsAndRun 创建统计实体
func NewItemHourlyStatsAndRun(day, hour string, ticketItem types.TicketItemEnum) {
	intHour, err := strconv.Atoi(hour)
	if err != nil {
		logs.Error("[NewItemDailyStats] hour(%s) is invalid:%v", hour, err)
		return
	}

	stats := new(ticketItemHourly)
	stats.ticketItem = ticketItem
	stats.hour = intHour
	stats.targetRepayAmountRate = GetTargetRepayRateByTicketItem(stats.ticketItem, types.TicketMyProcess)
	// 实时统计数据，此数据会导致 最后
	stats.dayStartTimeStamp, _, _ = parseUnixMillTimestampRange(day)
	stats.endTimestamp, err = parseUnixMillTimestampHourRange(hour)
	stats.startTimestamp = stats.dayStartTimeStamp

	logs.Notice("daystarttimestamp:", stats.dayStartTimeStamp, "starttimestamp:", stats.startTimestamp, "endtimestamp:", stats.endTimestamp)
	if err != nil {
		logs.Error("[NewItemDailyStats] date parse err", err)
		return
	}
	stats.baseModel = models.TicketWorkerHourlyStats{
		Hour:         intHour,
		TicketItemID: ticketItem,
		Ctime:        tools.GetUnixMillis(),
	}

	stats.statsData = map[int64]models.TicketWorkerHourlyStats{}

	stats.Start()
}

func fixStartTimeForLoads(startTimestamp int64) int64 {
	// 修正偏移, 早上8点前关闭和完成的, 工作人员无法触及该工单, 固不算负载
	fixStartOffset := 8 * time.Hour
	return startTimestamp + int64(fixStartOffset)/int64(time.Millisecond)
}

type usersRepayTotalRankingHourlyStats []models.TicketWorkerHourlyStats

// 获取此 slice 的长度
func (s usersRepayTotalRankingHourlyStats) Len() int { return len(s) }

//
func (s usersRepayTotalRankingHourlyStats) Less(i, j int) bool {
	return s[i].RepayTotal > s[j].RepayTotal
}

func (s usersRepayTotalRankingHourlyStats) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s *ticketItemHourly) Start() {
	s.doLoadStats()
	s.doAssignStats()
	s.doHandleStats()
	s.doCompleteStats()

	switch s.ticketItem {
	case types.TicketItemUrgeM11, types.TicketItemUrgeM12, types.TicketItemRM0, types.TicketItemRM1:
		// need config repayRateConfigNameMap
		s.doRepayAmountStats()
	}

	// TODO
	// LoadLeftRepayPrincipal int64
	// RepayPrincipal
	// RepayInterest
	// RepayTotal
	// RepayAmountRatio
	// Ranking
	// DiffTargetRepay

	// statsSlice := assemblyStatsData(s.day, s.ticketItem, userAssignStats, userCompleteStats, userHandleStats, userLoadStats)
	// do ranking
	var statsSlice usersRepayTotalRankingHourlyStats
	for _, d := range s.statsData {
		statsSlice = append(statsSlice, d)
	}
	sort.Sort(statsSlice)
	for i := range statsSlice {
		statsSlice[i].Ranking = i + 1
		statsSlice[i].LeaderRoleID = GetGroupLeaderRoleID(statsSlice[i].AdminUID)
	}

	statsNum := len(statsSlice)
	logs.Debug("[doHourlyPerformanceStatsTask] will insert num:%d", statsNum)
	if statsNum == 0 {
		return
	}

	obj := models.TicketWorkerHourlyStats{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	{
		// delete old current hour data first
		sql := fmt.Sprintf("DELETE FROM %s WHERE ticket_item_id=%d AND hour=%d", obj.TableName(), s.ticketItem, s.hour)
		r := o.Raw(sql)
		r.Exec()
	}

	nums, err := o.InsertMulti(100, statsSlice)
	if err != nil {
		logs.Error("[doHourlyPerformanceStatsTask] insert multi err", err)
	}
	if int(nums) != statsNum {
		logs.Error("[doHourlyPerformanceStatsTask] want exe is %d, actual exe %d", statsNum, nums)
		// failed
	}
}

func (s *ticketItemHourly) doLoadStats() {
	userLoadStats := models.GetUserTicketLoadCount(s.ticketItem, fixStartTimeForLoads(s.startTimestamp), s.startTimestamp, s.endTimestamp)
	// 组合分配数
	for _, ua := range userLoadStats {
		data := s.baseModel
		if val, ok := s.statsData[ua.Uid]; ok {
			data = val
		}
		data.LoadNum = ua.Num
		data.AdminUID = ua.Uid
		s.statsData[ua.Uid] = data
	}
}

func (s *ticketItemHourly) doAssignStats() {
	userAssignStats := models.GetUserTicketAssignCount(s.ticketItem, s.startTimestamp, s.endTimestamp)
	// 组合分配数
	for _, ua := range userAssignStats {
		data := s.baseModel
		if val, ok := s.statsData[ua.Uid]; ok {
			data = val
		}
		data.AssignNum = ua.Num
		data.AdminUID = ua.Uid
		s.statsData[ua.Uid] = data
	}
}

func (s *ticketItemHourly) doHandleStats() {
	var userHandleStats []models.UserTicketCount

	switch s.ticketItem {
	case types.TicketItemRepayRemind, types.TicketItemRM0, types.TicketItemRM1, types.TicketItemRMAdvance1:
		userHandleStats = models.GetUserRepayRemindHandleCount(s.ticketItem, s.startTimestamp, s.endTimestamp)
	case types.TicketItemUrgeM11, types.TicketItemUrgeM12, types.TicketItemUrgeM13:
		userHandleStats = models.GetUserUrgeHandleCount(s.ticketItem, s.startTimestamp, s.endTimestamp)
	default:
		//
		logs.Error("[doHourlyPerformanceStatsTask] ticketItem(%d) have no handle stats on date(%d) , check it", s.ticketItem, s.hour)
		return
	}
	// handlePointRate := getHandlePointRate(s.ticketItem)

	// 组合处理数和处理绩效点
	for _, ua := range userHandleStats {
		data := s.baseModel
		if val, ok := s.statsData[ua.Uid]; ok {
			data = val
		}
		data.AdminUID = ua.Uid
		data.HandleNum = ua.Num
		// data.HandlePoint = float64(ua.Num) * handlePointRate
		s.statsData[ua.Uid] = data
	}
}

func (s *ticketItemHourly) doCompleteStats() {
	userCompleteStats := models.GetUserTicketCompleteCount(s.ticketItem, s.startTimestamp, s.endTimestamp)
	// completePointRate := getCompletePointRate(s.ticketItem)

	// 组合完成数
	for _, ua := range userCompleteStats {
		data := s.baseModel
		if val, ok := s.statsData[ua.Uid]; ok {
			data = val
		}
		data.AdminUID = ua.Uid
		data.CompleteNum = ua.Num
		// data.CompletePoint = float64(ua.Num) * completePointRate
		s.statsData[ua.Uid] = data
	}
}

func (s *ticketItemHourly) doRepayAmountStats() {

	userOrderContainer := []struct {
		Uid     int64
		OrderId int64
	}{}
	statusBox, err := tools.IntsSliceToWhereInString(types.TicketStatusSliceInDoing())
	if err != nil {
		logs.Error("[GetUserTicketLoadCount] occur err:", err)
		return
	}

	where := fmt.Sprintf("WHERE assign_uid>0 AND item_id=%d AND assign_time<%d AND (complete_time>=%d  OR  status in(%s))",
		s.ticketItem, s.endTimestamp, s.startTimestamp, statusBox)
	sql := fmt.Sprintf("SELECT assign_uid as uid, order_id FROM `%s` %s ORDER BY uid DESC", models.TICKET_TABLENAME, where)

	obj := models.Ticket{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	r := o.Raw(sql)
	r.QueryRows(&userOrderContainer)

	//
	if len(userOrderContainer) == 0 {
		//
		return
	}

	userOrders := map[int64][]string{}

	for _, userOrder := range userOrderContainer {
		if _, ok := userOrders[userOrder.Uid]; ok {
			userOrders[userOrder.Uid] = append(userOrders[userOrder.Uid], strconv.FormatInt(userOrder.OrderId, 10))
		} else {
			userOrders[userOrder.Uid] = []string{strconv.FormatInt(userOrder.OrderId, 10)}
		}
	}

	for uid, orders := range userOrders {
		if len(orders) > 1000 {
			logs.Warn("[getUserLoadOrders] user own too many orders(%d), try to optimize this func:", len(orders))
		}
		if len(orders) == 0 {
			continue
		}

		repayPrincipal, repayInterest, _ := models.GetOrdersRepayPrincipalAndInterest(orders, s.startTimestamp, s.endTimestamp)
		data := s.baseModel
		if val, ok := s.statsData[uid]; ok {
			data = val
		}
		//回款本金
		data.RepayPrincipal = repayPrincipal
		//回款利息包括罚息和宽限期利息
		data.RepayInterest = repayInterest
		data.RepayTotal = data.RepayPrincipal + data.RepayInterest
		// 下属计算 LoadLeftUnpaidPrincipal 兼容任意时间的补算, 任何历史时间的待还本金都是准确的
		nowUnpaidPrincipal, _ := models.GetOrdersLeftUnpaidPrincipal(orders)
		alreadyPaidPrincipalSinceStatsDay, alreadyPaidInterest, _ := models.GetOrdersRepayPrincipalAndInterest(orders, s.dayStartTimeStamp, tools.GetUnixMillis())
		//分案本金=工单分配的剩余应还本金之和
		data.LoadLeftUnpaidPrincipal = nowUnpaidPrincipal + alreadyPaidPrincipalSinceStatsDay
		alreadyPaidTotal := alreadyPaidPrincipalSinceStatsDay + alreadyPaidInterest
		//回收率
		//data.RepayAmountRate = GetRepayAmountRate(s.ticketItem, orders, alreadPaidTotal, data.LoadLeftUnpaidPrincipal, alreadyPaidInterest)
		data.RepayAmountRate = GetUrgeRepayAmountRate(alreadyPaidTotal, data.LoadLeftUnpaidPrincipal)
		//目标回收率
		data.TargetRepayRate = s.targetRepayAmountRate //GetStandardRepayRateByTicketItem(s.ticketItem)
		//回款总额与目标回款的差额
		data.DiffTargetRepay = int64(float64(data.LoadLeftUnpaidPrincipal)*data.TargetRepayRate/100) - data.RepayTotal

		s.statsData[uid] = data
	}

	return
}

// // GetStandardRepayRateByTicketItem 获取标准回款率
// func GetStandardRepayRateByTicketItem(ticketItem types.TicketItemEnum) float64 {
// 	configName := getUrgeRepayRateStandardConfigName(ticketItem)
// 	repayAmountStandardRate, _ := config.ValidItemFloat64(configName)
// 	return repayAmountStandardRate
// }

// func getCompletePointRate(ticketItem types.TicketItemEnum) float64 {
// 	configName := getCompletePointRateConfigName(ticketItem)
// 	rate, _ := config.ValidItemFloat64(configName)
// 	return rate
// }

// func getHandlePointRate(ticketItem types.TicketItemEnum) float64 {
// 	configName := getHandlePointRateConfigName(ticketItem)
// 	rate, _ := config.ValidItemFloat64(configName)
// 	return rate
// }

func parseUnixMillTimestampHourRange(hour string) (end int64, err error) {
	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		return
	}
	statsTime, err := time.ParseInLocation("2006010215", hour, loc)

	logs.Notice("[parseUnixMillTimestampHourRange] startTime:", statsTime)

	if err != nil {
		return
	}
	end = statsTime.UnixNano() / int64(time.Millisecond)
	return
}

// getToday day
func getToday() (day string) {
	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		return
	}
	day = time.Now().AddDate(0, 0, 0).In(loc).Format("20060102")
	return
}

func getHour() (hour string) {
	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		return
	}
	hour = time.Now().Add(time.Hour - realtimeStatsInterval).In(loc).Format("2006010215")
	return
}

// GetGroupLeaderRoleID 获取群组 Group Role ID
func GetGroupLeaderRoleID(uid int64) int64 {
	admin, err := models.OneAdminByUid(uid)
	if err != nil {
		logs.Error("[GetGroupLeaderRoleID] user cannot be find in database:", err)
		return 0
	}
	roleLevel := rbac.GetRoleLevel(admin.RoleID)
	if roleLevel == types.RoleLeader {
		return admin.RoleID
	}
	if roleLevel == types.RoleEmployee {
		role, _ := models.GetOneRole(admin.RoleID)
		return role.Pid
	}
	logs.Error("[GetGroupLeaderRoleID] unexpected super role occur in stats:", admin.Id, admin.RoleID)
	return 0
}
