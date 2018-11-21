package performance

import (
	"fmt"
	"math"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// UpdateCurrentMonthStats 更新当月统计
func UpdateCurrentMonthStats() {
	needStatsTicketItem := []types.TicketItemEnum{
		types.TicketItemUrgeM11,
		types.TicketItemUrgeM12,
		types.TicketItemUrgeM13,
		types.TicketItemRepayRemind,
	}
	for _, ticketItem := range needStatsTicketItem {
		doMonthlyPerformanceStatsDailyUpdate(ticketItem)
	}
}

// 提醒小组:    逾期率统计范围:  { 账单日 + 2天  in  当月}   ，逾期率 > N1的天数
// 逾期是指:超过还款提醒范围，　达到下一个case的订单
// 6月26日: 还款提醒逾期率: d2/user_num
// select user_num, d2 from risk_monitor.deadline_overdue_data_all WHERE identified_1='user' and identified_2="overdue_user" and deadline_dt = "2018-06-24"
//
//
// M1- 1:    逾期率统计范围:  { 账单日 + 8天  in  当月}   ，逾期率 > N2的天数
// 6月26日: M1- 1逾期率: d8/user_num
// select user_num, d8 from risk_monitor.deadline_overdue_data_all WHERE identified_1='user' and identified_2="overdue_user" and deadline_dt = "2018-06-18"

// 当月处理量  =  日处理量叠加 + sum(日处理数据) date >= 20180626 and date <= 20180725 and ticket_item_id=xx
//
// 当月回款数  =  当月每个小组各自所有员工中结清的订单数  sum(日处理数据) date >= 20180626 and date <= 20180725 and ticket_item_id=xx
//
// 当月回款率  =    当月回款数/当月所有工单数目 当月回款数 / date >= 20180626 and date <= 20180725 and ticket_item_id=xx

// 月统计开始时间:上月26-本月25
// 当月逾期率达标天数 ：
//
func doMonthlyPerformanceStatsDailyUpdate(ticketItem types.TicketItemEnum) {
	//
	//time.Now().Format("")
	month := getCurrentStatsMonth()

	startTimestamp, endTimestamp, err := getMonthStatsMillUnixtimestampRange(month)
	startDay, endDay := getMonthStatsDayRange(month)

	if err != nil {
		logs.Error("[doDailyPerformanceStatsTask] date parse err", err)
		return
	}

	md, err := models.GetMonthlyStatsByDateAndTicketItem(ticketItem, month)
	md.HandleNum, md.CompleteNum = models.GetTicketWorkerPerformanceCountByDateRange(ticketItem, startDay, endDay)
	md.CompleteRate = RoundFloat64(models.GetTicketItemCompleteRateInRange(ticketItem, startTimestamp, endTimestamp), 4)
	lastDay := getLastDay()
	day, _ := strconv.Atoi(lastDay)

	if err != nil {
		md.TicketItemID = ticketItem
		md.Date = month
		md.Ctime = tools.GetUnixMillis()
		md.Utime = tools.GetUnixMillis()
		md.OverdueRateAchieveDays = statsCurrentMonthOverdueRateAchieve(ticketItem, lastDay, startDay)
		id, _ := models.OrmInsert(&md)
		if id > 0 {
			//
		}
	} else {
		//
		if !IsLastUpdateOnYerstoday(md.Utime) {
			md.OverdueRateAchieveDays = statsCurrentMonthOverdueRateAchieve(ticketItem, lastDay, startDay)
		} else {
			if IsOverdueRateAchieve(ticketItem, day) {
				md.OverdueRateAchieveDays++
			}
		}
		md.Utime = tools.GetUnixMillis()
		num, _ := models.OrmAllUpdate(&md)
		if num == 1 {
			//
		}
	}

}

// RoundFloat64 简易四舍五入保留指定小数位数
func RoundFloat64(val float64, precision int) float64 {
	t := math.Pow10(precision)
	return float64(math.Round(val*t)) / t
}

// IsLastUpdateOnYerstoday 上一次更新是否在昨天
func IsLastUpdateOnYerstoday(milliUtime int64) bool {
	if milliUtime == 0 {
		return true
	}
	loc, _ := time.LoadLocation(tools.GetServiceTimezone())
	if time.Unix(milliUtime/1000, 0).In(loc).AddDate(0, 0, 1).Format("20060102") == time.Now().In(loc).Format("20060102") {
		return true
	}
	return false
}

func getOverdueRateStandard(ticketItem types.TicketItemEnum) float64 {
	configName := getCaseOverdueRateStandardConfigName(ticketItem)
	rate, _ := config.ValidItemFloat64(configName)
	return rate
}

func statsCurrentMonthOverdueRateAchieve(ticketItem types.TicketItemEnum, lastDay string, startDay int) (achieveDays int) {
	intLastDay, _ := strconv.Atoi(lastDay)

	loc, _ := time.LoadLocation(tools.GetServiceTimezone())
	st, _ := time.ParseInLocation("20060102", strconv.Itoa(startDay), loc)
	day := startDay
	for i := 1; day <= intLastDay; i++ {
		if IsOverdueRateAchieve(ticketItem, day) {
			achieveDays++
		}

		day, _ = strconv.Atoi(st.AddDate(0, 0, i).Format("20060102"))
	}
	return
}

// IsOverdueRateAchieve 指定天, 逾期率是否达到标准
func IsOverdueRateAchieve(ticketItem types.TicketItemEnum, statsDay int) bool {

	var column string
	var beforeDays int
	switch ticketItem {
	case types.TicketItemRepayRemind:
		column = "d2"
		beforeDays = 2
	case types.TicketItemUrgeM11:
		column = "d8"
		beforeDays = 8
	case types.TicketItemUrgeM12:
		column = "d16"
		beforeDays = 16
	case types.TicketItemUrgeM13:
		column = "d31"
		beforeDays = 31
	default:

		return false
	}

	t, _ := time.Parse("20060102", strconv.Itoa(statsDay))
	date := t.AddDate(0, 0, -beforeDays).Format("2006-01-02")

	where := fmt.Sprintf("WHERE deadline_dt='%s'", date)
	sql := fmt.Sprintf("SELECT %s as overdue_num, user_num FROM `deadline_overdue_data_all` %s  limit 1",
		column, where)

	o := orm.NewOrm()
	o.Using(types.OrmDataBaseRiskMonitorSlave)
	r := o.Raw(sql)

	container := struct {
		OverdueNum float64
		UserNum    float64
	}{}
	r.QueryRow(&container)
	if container.UserNum > 0 {
		rate := container.OverdueNum / container.UserNum
		if rate < getOverdueRateStandard(ticketItem) {
			return true
		}
	}

	return false
}

// 月统计起始日
const (
	MonthStartDay = 26
	MonthEndDay   = 25
)

// getCurrentStatsMonth 返回统计的月份, 起止时间戳, 或者起止day
// 月统计开始时间:上月26-本月25
func getCurrentStatsMonth() int {
	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		return 0
	}
	// day = time.Now().In(loc).Format("20060102")
	lt := time.Now().In(loc)
	year := lt.Year()
	lastCompleteDay := lt.Day() - 1
	var month int
	if lastCompleteDay >= MonthStartDay {
		month = int(lt.AddDate(0, 1, 0).Month())
	} else {
		month = int(lt.Month())
	}
	statsTag := year*100 + month

	return statsTag
}

// month format like : int 201806
// return startDay like 20180526, endDay like 20180625
func getMonthStatsDayRange(month int) (startDay, endDay int) {
	startDay = (month-1)*100 + MonthStartDay
	endDay = month*100 + MonthEndDay
	return
}

func getMonthStatsMillUnixtimestampRange(month int) (start, end int64, err error) {
	startDay, endDay := getMonthStatsDayRange(month)
	loc, _ := time.LoadLocation(tools.GetServiceTimezone())
	startT, startErr := time.ParseInLocation("20060102", strconv.Itoa(startDay), loc)
	if startErr != nil {
		err = fmt.Errorf("[getMonthStatsMillUnixtimestampRange] time parse err: %v", startErr)
		return
	}
	endT, endErr := time.ParseInLocation("20060102", strconv.Itoa(endDay), loc)
	if endErr != nil {
		err = fmt.Errorf("[getMonthStatsMillUnixtimestampRange] time parse err: %v", endErr)
		return
	}

	start = startT.UnixNano() / int64(time.Millisecond)
	end = endT.UnixNano() / int64(time.Millisecond)
	return
}
