package performance

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"sort"
	"strconv"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type userItemStats struct {
	Uid int64
	Num int64
}

// LastDayDailyWorkerPerformanceStatsTask 执行最近一天的统计
func LastDayDailyWorkerPerformanceStatsTask() {
	// 获取昨天的日期
	day := getLastDay()
	DailyWorkerPerformanceStatsTask(day)
}

// DailyWorkerPerformanceStatsTask worker日绩效统计
// day 格式 20180701
func DailyWorkerPerformanceStatsTask(day string) {
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
		NewItemDailyStatsAndRun(day, ticketItem)
	}

}

type ticketItemDaily struct {
	ticketItem            types.TicketItemEnum
	day                   int
	startTimestamp        int64
	endTimestamp          int64
	baseModel             models.TicketWorkerDailyStats
	statsData             map[int64]models.TicketWorkerDailyStats
	targetRepayAmountRate float64 // 目标回收率
}

// NewItemDailyStatsAndRun 创建统计实体
func NewItemDailyStatsAndRun(day string, ticketItem types.TicketItemEnum) {
	intDay, err := strconv.Atoi(day)
	if err != nil {
		logs.Error("[NewItemDailyStats] day(%s) is invalid:%v", day, err)
		return
	}

	stats := new(ticketItemDaily)
	stats.ticketItem = ticketItem
	stats.targetRepayAmountRate = GetTargetRepayRateByTicketItem(ticketItem, types.TicketPerformanceManage)
	stats.day = intDay
	stats.startTimestamp, stats.endTimestamp, err = parseUnixMillTimestampRange(day)
	if err != nil {
		logs.Error("[NewItemDailyStats] date parse err", err)
		return
	}
	stats.baseModel = models.TicketWorkerDailyStats{
		Date:         intDay,
		TicketItemID: ticketItem,
		Ctime:        tools.GetUnixMillis(),
	}

	stats.statsData = map[int64]models.TicketWorkerDailyStats{}

	stats.Start()
}

func fixStartTimeForLoad(startTimestamp int64) int64 {
	// 修正偏移, 早上8点前关闭和完成的, 工作人员无法触及该工单, 固不算负载
	fixStartOffset := 9 * time.Hour
	return startTimestamp + int64(fixStartOffset)/int64(time.Millisecond)
}

type usersRepayTotalRankingDailyStats []models.TicketWorkerDailyStats

// 获取此 slice 的长度
func (s usersRepayTotalRankingDailyStats) Len() int { return len(s) }

//
func (s usersRepayTotalRankingDailyStats) Less(i, j int) bool {
	return s[i].RepayTotal > s[j].RepayTotal
}

func (s usersRepayTotalRankingDailyStats) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s *ticketItemDaily) Start() {
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
	var statsSlice usersRepayTotalRankingDailyStats
	for _, d := range s.statsData {
		statsSlice = append(statsSlice, d)
	}
	sort.Sort(statsSlice)
	for i := range statsSlice {
		statsSlice[i].Ranking = i + 1
	}

	statsNum := len(statsSlice)
	logs.Debug("[doDailyPerformanceStatsTask] will insert num:%d", statsNum)
	if statsNum == 0 {
		return
	}

	obj := models.TicketWorkerDailyStats{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	nums, err := o.InsertMulti(100, statsSlice)
	if err != nil {
		logs.Error("[doDailyPerformanceStatsTask] insert multi err", err)
	}
	if int(nums) != statsNum {
		logs.Error("[doDailyPerformanceStatsTask] want insert is %d, actual insert %d", statsNum, nums)
		// failed
	}
}

func (s *ticketItemDaily) doLoadStats() {
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

func (s *ticketItemDaily) doAssignStats() {
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

func (s *ticketItemDaily) doHandleStats() {
	var userHandleStats []models.UserTicketCount

	switch s.ticketItem {
	case types.TicketItemRepayRemind, types.TicketItemRM0, types.TicketItemRM1, types.TicketItemRMAdvance1:
		userHandleStats = models.GetUserRepayRemindHandleCount(s.ticketItem, s.startTimestamp, s.endTimestamp)
	case types.TicketItemUrgeM11, types.TicketItemUrgeM12, types.TicketItemUrgeM13:
		userHandleStats = models.GetUserUrgeHandleCount(s.ticketItem, s.startTimestamp, s.endTimestamp)
	default:
		//
		logs.Error("[doDailyPerformanceStatsTask] ticketItem(%d) have no handle stats on date(%d) , check it", s.ticketItem, s.day)
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

func (s *ticketItemDaily) doCompleteStats() {
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

func (s *ticketItemDaily) doRepayAmountStats() {

	userOrderContainer := []struct {
		Uid     int64
		OrderId int64
	}{}
	statusBox, err := tools.IntsSliceToWhereInString(types.TicketStatusSliceInDoing())
	if err != nil {
		logs.Error("[GetUserTicketLoadCount] occur err:", err)
		return
	}

	// where := fmt.Sprintf("WHERE assign_uid>0 AND item_id=%d AND assign_time<%d AND (complete_time>=%d  OR  status in(%s))",
	// 	s.ticketItem, s.endTimestamp, s.startTimestamp, statusBox)
	where := fmt.Sprintf(`WHERE assign_uid>0 AND item_id=%d AND assign_time<%d
		 AND ((complete_time>=%d) OR status in(%s) OR (close_time>=%d))`,
		s.ticketItem, s.endTimestamp, s.startTimestamp, statusBox,
		fixStartTimeForLoad(s.startTimestamp))
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
		data.RepayPrincipal = repayPrincipal
		data.RepayInterest = repayInterest
		data.RepayTotal = data.RepayPrincipal + data.RepayInterest
		// 下属计算 LoadLeftUnpaidPrincipal 兼容任意时间的补算, 任何历史时间的待还本金都是准确的
		nowUnpaidPrincipal, _ := models.GetOrdersLeftUnpaidPrincipal(orders)
		alreadyPaidPrincipalSinceStatsDay, alreadyPaidInterest, _ := models.GetOrdersRepayPrincipalAndInterest(orders, s.startTimestamp, tools.GetUnixMillis())
		// 分案本金=工单分配的剩余应还本金之和
		data.LoadLeftUnpaidPrincipal = nowUnpaidPrincipal + alreadyPaidPrincipalSinceStatsDay

		// 回款总金额=已还本金+已还息费
		alreadyPaidTotal := alreadyPaidPrincipalSinceStatsDay + alreadyPaidInterest
		// 回收率
		//data.RepayAmountRate = GetRepayAmountRate(s.ticketItem, orders, alreadyPaidTotal, data.LoadLeftUnpaidPrincipal, alreadyPaidInterest)
		data.RepayAmountRate = GetUrgeRepayAmountRate(alreadyPaidTotal, data.LoadLeftUnpaidPrincipal)
		// 目标回收率
		data.TargetRepayRate = s.targetRepayAmountRate
		// 差值金额
		data.DiffTargetRepay = int64(float64(data.LoadLeftUnpaidPrincipal)*data.TargetRepayRate/100) - data.RepayTotal

		s.statsData[uid] = data
	}

	return
}

// 获取回收率
func GetRepayAmountRate(ticketItem types.TicketItemEnum, orders []string, alreadPaidTotal, principal, alreadyPaidInterest int64) (repayAmountRate float64) {
	if ticketItem == types.TicketItemUrgeM11 || ticketItem == types.TicketItemUrgeM12 {
		// 10.11-10.17需求之前的co+rm回收率获取方法
		return GetUrgeRepayAmountRate(alreadPaidTotal, principal)
	} else if ticketItem == types.TicketItemRM0 || ticketItem == types.TicketItemRM1 {
		nowUnpaidInterest, _ := models.GetOrdersLeftUnpaidInterest(orders)
		// 分案交易额=分案本金+分案息费（分案息费=已还息费+未还息费）（10.11-10.17的需求）
		transaction := principal + nowUnpaidInterest + alreadyPaidInterest
		if transaction > 0 {
			repayAmountRate = RoundFloat64(float64(alreadPaidTotal*100)/float64(transaction), 2)
		}
	}

	return
}

// GetUrgeRepayAmountRate 获取催收回款率
func GetUrgeRepayAmountRate(repayTotal, loadLeftUnpaidPrincipal int64) float64 {
	if loadLeftUnpaidPrincipal == 0 {
		return 0
	}
	return RoundFloat64(float64(repayTotal*100)/float64(loadLeftUnpaidPrincipal), 2)
}

// GetTargetRepayRateByTicketItem 获取目的回收率
func GetTargetRepayRateByTicketItem(ticketItem types.TicketItemEnum, ticketTag int) (targetRepayRate float64) {
	// 获取标准回收率
	repayAmountStandardRate := GetStandardRepayRateByTicketItem(ticketItem)
	targetRepayRate = repayAmountStandardRate

	/*
		if ticketItem == types.TicketItemRM0 {
			originalRate := GetRM0OriginalRepayRate(ticketItem, ticketTag)
			if originalRate > 0 {
				targetRepayRate = ((originalRate-repayAmountStandardRate)/originalRate)*100 + 5
			}
		}
	*/

	//if ticketItem == types.TicketItemRM0 || ticketItem == types.TicketItemRM1 {
	if ticketItem == types.TicketItemRM1 {
		originalRate := GetRMOriginalRepayRate(ticketItem, ticketTag)
		if originalRate > 0 {
			targetRepayRate = ((originalRate-repayAmountStandardRate)/originalRate)*100 + 5
		}
	}

	return
}

// GetStandardRepayRateByTicketItem 获取标准回款率
func GetStandardRepayRateByTicketItem(ticketItem types.TicketItemEnum) float64 {
	configName := getUrgeRepayRateStandardConfigName(ticketItem)
	repayAmountStandardRate, _ := config.ValidItemFloat64(configName)
	return repayAmountStandardRate
}

/*
// GetRM0OriginalRepayRate 获取初始回收率，初始值=生成工单的未还款金额/应还总金额
func GetRM0OriginalRepayRate(ticketItem types.TicketItemEnum, ticketTag int) (originalRate float64) {

	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	var unpaidCaseAmount, unpaidAmount int64
	// 生成工单的未还款金额
	var unpaidCaseAmountKey string
	if ticketTag == types.TicketMyProcess { // 工作进度查询依据当天
		unpaidCaseAmountKey = tools.GetUnpaidAmountKey("unpaid_case_amount", tools.GetUnixMillis())
	} else if ticketTag == types.TicketPerformanceManage { // 人员绩效管理查询依据前一天
		unpaidCaseAmountKey = tools.GetUnpaidAmountKey("unpaid_case_amount", tools.GetUnixMillis()-tools.MILLSSECONDADAY)
	}
	cValue, _ := cacheClient.Do("GET", unpaidCaseAmountKey)
	if cValue == nil {
		return
	} else {
		value := string(cValue.([]byte))
		unpaidCaseAmount, _ = tools.Str2Int64(value)
	}

	// 应还总金额
	var unpaidAmountKey string
	if ticketTag == types.TicketMyProcess { // 工作进度查询依据当天
		unpaidAmountKey = tools.GetUnpaidAmountKey("unpaid_amount", tools.GetUnixMillis())
	} else if ticketTag == types.TicketPerformanceManage { // 人员绩效管理查询依据前一天
		unpaidAmountKey = tools.GetUnpaidAmountKey("unpaid_amount", tools.GetUnixMillis()-tools.MILLSSECONDADAY)
	}
	cValue, _ = cacheClient.Do("GET", unpaidAmountKey)
	if cValue == nil {
		return
	} else {
		value := string(cValue.([]byte))
		unpaidAmount, _ = tools.Str2Int64(value)
	}

	if unpaidAmount > 0 {
		originalRate = float64(unpaidCaseAmount/unpaidAmount) * 100
	}

	return
}
*/

// GetRMOriginalRepayRate 获取初始回收率，初始值=H1时未还款金额/应还金额
func GetRMOriginalRepayRate(ticketItem types.TicketItemEnum, ticketTag int) (originalRate float64) {

	var date string
	if ticketTag == types.TicketMyProcess { // 工作进度查询依据当天
		date = tools.MDateMHSDate(tools.NaturalDay(0))
		if ticketItem == types.TicketItemRM1 {
			date = tools.MDateMHSDate(tools.NaturalDay(-1))
		}
	} else if ticketTag == types.TicketPerformanceManage { // 人员绩效管理查询依据前一天
		date = tools.MDateMHSDate(tools.NaturalDay(-1))
		if ticketItem == types.TicketItemRM1 {
			date = tools.MDateMHSDate(tools.NaturalDay(-2))
		}
	}

	ticketItemStr := types.TicketItemMap()[ticketItem]
	bs := types.RMLevelRelatedItemMap()[ticketItemStr]

	where := fmt.Sprintf("WHERE dt='%s' AND bs='%s'", date, bs)
	sql := fmt.Sprintf("SELECT amount_total, hour_1 as one_unpay FROM `bill_repay_hour_deadline` %s ", where)

	o := orm.NewOrm()
	o.Using(types.OrmDataBaseRiskMonitor)
	r := o.Raw(sql)

	var amountTotal float64
	var oneUnpay float64
	err := r.QueryRow(&amountTotal, &oneUnpay)
	if err != nil {
		//logs.Warn("[GetRMOriginalRepayRate] amountTotal:", amountTotal, ", oneUnpay:", oneUnpay, ", sql:", sql, ", err:", err)
	}

	if amountTotal > 0 {
		originalRate = (oneUnpay / amountTotal) * 100
	}

	return
}

func getCompletePointRate(ticketItem types.TicketItemEnum) float64 {
	configName := getCompletePointRateConfigName(ticketItem)
	rate, _ := config.ValidItemFloat64(configName)
	return rate
}

func getHandlePointRate(ticketItem types.TicketItemEnum) float64 {
	configName := getHandlePointRateConfigName(ticketItem)
	rate, _ := config.ValidItemFloat64(configName)
	return rate
}

func parseUnixMillTimestampRange(day string) (start, end int64, err error) {
	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		return
	}
	startTime, err := time.ParseInLocation("20060102", day, loc)
	if err != nil {
		return
	}
	start = startTime.UnixNano() / int64(time.Millisecond)
	end = startTime.AddDate(0, 0, 1).UnixNano() / int64(time.Millisecond)
	return
}

// getLastDay day
func getLastDay() (day string) {
	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		return
	}
	day = time.Now().AddDate(0, 0, -1).In(loc).Format("20060102")
	return
}
