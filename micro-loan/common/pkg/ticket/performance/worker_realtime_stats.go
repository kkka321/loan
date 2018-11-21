package performance

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"sort"
	"strconv"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type userItemRealtimeStats struct {
	Uid int64
	Num int64
}

const realtimeStatsInterval time.Duration = 5 * time.Minute

// TodayRealtimeWorkerPerformanceStatsTask 执行最近一天的统计
func TodayRealtimeWorkerPerformanceStatsTask() {
	// 获取昨天的日期

	timeTag := getRealtimeTimeTag()
	logs.Notice("timeTag:", timeTag)
	RealtimeWorkerPerformanceStatsTask(timeTag)
}

func getRealtimeTimeTag() (timeTag int64) {
	frequency := int64(realtimeStatsInterval)

	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		return
	}
	t := time.Now().In(loc)
	//m := t.Minute()

	return (t.UnixNano() / frequency) * frequency / int64(time.Millisecond)
}

// RealtimeWorkerPerformanceStatsTask worker日绩效统计
// timeTag 格式 2018070112
func RealtimeWorkerPerformanceStatsTask(timeTag int64) {
	day := getToday()
	needStatsTicketItem := []types.TicketItemEnum{
		types.TicketItemUrgeM11,
		//types.TicketItemUrgeM12,
		// types.TicketItemRMAdvance1,
		//types.TicketItemRM0,
		//types.TicketItemRM1,
		//types.TicketItemUrgeM13,
	}
	for _, ticketItem := range needStatsTicketItem {
		NewItemRealtimeStatsAndRun(day, timeTag, ticketItem)
	}

}

type ticketItemRealtime struct {
	ticketItem        types.TicketItemEnum
	timeTag           int64
	dayStartTimeStamp int64
	startTimestamp    int64
	endTimestamp      int64
	baseModel         models.TicketWorkerRealtimeStats
	statsData         map[int64]models.TicketWorkerRealtimeStats
}

// NewItemRealtimeStatsAndRun 创建统计实体
func NewItemRealtimeStatsAndRun(day string, timeTag int64, ticketItem types.TicketItemEnum) {

	stats := new(ticketItemRealtime)
	stats.ticketItem = ticketItem
	stats.timeTag = timeTag
	stats.dayStartTimeStamp, _, _ = parseUnixMillTimestampRange(day)
	stats.endTimestamp = timeTag
	stats.startTimestamp = stats.dayStartTimeStamp

	logs.Notice("daystarttimestamp:", stats.dayStartTimeStamp, "starttimestamp:", stats.startTimestamp, "endtimestamp:", stats.endTimestamp)

	stats.baseModel = models.TicketWorkerRealtimeStats{
		TimeTag:      stats.timeTag,
		TicketItemID: ticketItem,
		Ctime:        tools.GetUnixMillis(),
	}

	stats.statsData = map[int64]models.TicketWorkerRealtimeStats{}

	stats.Start()
}

type usersRepayTotalRankingRealtimeStats []models.TicketWorkerRealtimeStats

// 获取此 slice 的长度
func (s usersRepayTotalRankingRealtimeStats) Len() int { return len(s) }

//
func (s usersRepayTotalRankingRealtimeStats) Less(i, j int) bool {
	return s[i].RepayTotal > s[j].RepayTotal
}

func (s usersRepayTotalRankingRealtimeStats) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s *ticketItemRealtime) Start() {
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
	var statsSlice usersRepayTotalRankingRealtimeStats
	for _, d := range s.statsData {
		statsSlice = append(statsSlice, d)
	}
	sort.Sort(statsSlice)
	for i := range statsSlice {
		statsSlice[i].Ranking = i + 1
	}

	statsNum := len(statsSlice)
	logs.Debug("[doHourlyPerformanceStatsTask] will insert num:%d", statsNum)
	if statsNum == 0 {
		return
	}

	obj := models.TicketWorkerRealtimeStats{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	var opNum int
	for _, oneStats := range statsSlice {
		_, err := o.InsertOrUpdate(&oneStats, "ranking")
		if err != nil {
			logs.Error(err)
			continue
		}
		opNum++
	}
	// nums, err := o.InsertMulti(100, statsSlice)
	// if err != nil {
	// 	logs.Error("[doHourlyPerformanceStatsTask] insert multi err", err)
	// }
	if opNum != statsNum {
		logs.Error("[doHourlyPerformanceStatsTask] want exe is %d, actual exe %d", statsNum, opNum)
		// failed
	}
	// nums, err := o.InsertMulti(100, statsSlice)
	// if err != nil {
	// 	logs.Error("[doHourlyPerformanceStatsTask] insert multi err", err)
	// }
	// if int(nums) != statsNum {
	// 	logs.Error("[doHourlyPerformanceStatsTask] want insert is %d, actual insert %d", statsNum, nums)
	// 	// failed
	// }
}

func (s *ticketItemRealtime) doLoadStats() {
	userLoadStats := models.GetUserTicketLoadCount(s.ticketItem, fixStartTimeForLoad(s.startTimestamp), s.startTimestamp, s.endTimestamp)
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

func (s *ticketItemRealtime) doAssignStats() {
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

func (s *ticketItemRealtime) doHandleStats() {
	var userHandleStats []models.UserTicketCount

	switch s.ticketItem {
	case types.TicketItemRepayRemind, types.TicketItemRM0, types.TicketItemRM1, types.TicketItemRMAdvance1:
		userHandleStats = models.GetUserRepayRemindHandleCount(s.ticketItem, s.startTimestamp, s.endTimestamp)
	case types.TicketItemUrgeM11, types.TicketItemUrgeM12, types.TicketItemUrgeM13:
		userHandleStats = models.GetUserUrgeHandleCount(s.ticketItem, s.startTimestamp, s.endTimestamp)
	default:
		//
		logs.Error("[doHourlyPerformanceStatsTask] ticketItem(%d) have no handle stats on date(%d) , check it", s.ticketItem, s.timeTag)
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

func (s *ticketItemRealtime) doCompleteStats() {
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

func (s *ticketItemRealtime) doRepayAmountStats() {

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
		alreadPaidTotal := alreadyPaidPrincipalSinceStatsDay + alreadyPaidInterest
		//回收率
		data.RepayAmountRate = GetRepayAmountRate(s.ticketItem, orders, alreadPaidTotal, data.LoadLeftUnpaidPrincipal, alreadyPaidInterest)
		//目标回收率
		data.TargetRepayRate = GetTargetRepayRateByTicketItem(s.ticketItem, types.TicketMyProcess) //GetStandardRepayRateByTicketItem(s.ticketItem)
		//回款总额与目标回款的差额
		data.DiffTargetRepay = int64(float64(data.LoadLeftUnpaidPrincipal)*data.TargetRepayRate/100) - data.RepayTotal

		s.statsData[uid] = data
	}

	return
}
