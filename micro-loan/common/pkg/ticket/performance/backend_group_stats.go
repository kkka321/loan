package performance

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/types"
	"sort"
	"strconv"

	"github.com/astaxie/beego/orm"
)

type baseStats struct {
	RepayTotal              int64
	LoadLeftUnpaidPrincipal int64
	DiffTargetRepay         int64
	RepayAmountRate         float64
}

// ChartLatestTimeTag char 最新统计时间字段 label
const ChartLatestTimeTag = "Now"

// GroupStats 小组统计
type GroupStats struct {
	Hour                    int
	Ranking                 int
	LeaderRoleID            int64 `orm:"column(leader_role_id)"`
	RepayTotal              int64
	LoadLeftUnpaidPrincipal int64
	DiffTargetRepay         int64
	RepayAmountRate         float64
}

// GroupChart 小组图表
type GroupChart struct {
	Hour            int     `json:"_"`
	ChartHour       string  `json:"name"`
	RepayAmountRate float64 `json:"y"`
}

// GroupCharts 小组图表
type GroupCharts []GroupChart

func (s GroupCharts) Len() int { return len(s) }

//
func (s GroupCharts) Less(i, j int) bool {
	return s[i].Hour < s[j].Hour
}

func (s GroupCharts) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GroupTotalStatsList 多个team统计列表
type GroupTotalStatsList []GroupStats

// 获取此 slice 的长度
func (s GroupTotalStatsList) Len() int { return len(s) }

//
func (s GroupTotalStatsList) Less(i, j int) bool {
	return s[i].RepayTotal > s[j].RepayTotal
}

func (s GroupTotalStatsList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetGroupDailyLatestStats 获取当日实时统计的最新后台小组数据
func GetGroupDailyLatestStats(leaderRoleID int64, latestHour int, ticketItem types.TicketItemEnum) (groupTotal GroupStats) {
	obj := models.TicketWorkerHourlyStats{}

	o := orm.NewOrm()
	o.Using(obj.Using())
	field := "sum(repay_total) as repay_total, sum(load_left_unpaid_principal) as load_left_unpaid_principal, sum(diff_target_repay) as diff_target_repay"
	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE hour=%d and ticket_item_id=%d and leader_role_id=%d`,
		field, obj.TableName(), latestHour, ticketItem, leaderRoleID)

	r := o.Raw(sql)
	r.QueryRow(&groupTotal)
	groupTotal.RepayAmountRate = GetUrgeRepayAmountRate(groupTotal.RepayTotal, groupTotal.LoadLeftUnpaidPrincipal)
	groupTotal.LeaderRoleID = leaderRoleID
	return
}

// GetGroupListStats 获取各个小组的最新统计
func GetGroupListStats(leaderRoleID int64, latestHour int, ticketItem types.TicketItemEnum) (list GroupTotalStatsList) {
	obj := models.TicketWorkerHourlyStats{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	field := "leader_role_id,sum(repay_total) as repay_total, sum(load_left_unpaid_principal) as load_left_unpaid_principal"
	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE hour=%d and ticket_item_id=%d  GROUP BY leader_role_id`,
		field, obj.TableName(), latestHour, ticketItem)

	r := o.Raw(sql)
	r.QueryRows(&list)
	sort.Sort(list)

	for i := range list {
		list[i].Ranking = i + 1
		list[i].RepayAmountRate = GetUrgeRepayAmountRate(list[i].RepayTotal, list[i].LoadLeftUnpaidPrincipal)
	}

	return
}

// GetSingleGroupChart 获取指定小组的统计chart
func GetSingleGroupChart(leaderRoleID int64, latestHour int, ticketItem types.TicketItemEnum) (chartData GroupCharts) {
	obj := models.TicketWorkerHourlyStats{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	field := "hour, sum(repay_total) as repay_total, sum(load_left_unpaid_principal) as load_left_unpaid_principal, sum(diff_target_repay) as diff_target_repay"
	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE  ticket_item_id=%d and leader_role_id=%d group by hour order by hour DESC LIMIT 8`,
		field, obj.TableName(), ticketItem, leaderRoleID)

	r := o.Raw(sql)
	var list GroupTotalStatsList
	r.QueryRows(&list)
	for _, v := range list {
		cv := GroupChart{v.Hour, strconv.Itoa(v.Hour - v.Hour/100*100), GetUrgeRepayAmountRate(v.RepayTotal, v.LoadLeftUnpaidPrincipal)}
		chartData = append(chartData, cv)
	}
	sort.Sort(chartData)
	if len(chartData) > 0 {
		chartData[len(chartData)-1].ChartHour = ChartLatestTimeTag
	}

	return
}
