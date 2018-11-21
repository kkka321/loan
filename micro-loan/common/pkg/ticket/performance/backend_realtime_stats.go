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

	"github.com/astaxie/beego/orm"
)

// DaySalaryChartData 统计图表所需数据模型
type DaySalaryChartData struct {
	Hour      string `json:"name"`
	StatsHour int    `json:"_"`
	Salary    int64  `json:"y"`
}

type DailySalaryChartDatas []DaySalaryChartData

func (p DailySalaryChartDatas) Len() int { return len(p) }

func (p DailySalaryChartDatas) Less(i, j int) bool {
	return p[i].StatsHour < p[j].StatsHour
}

func (p DailySalaryChartDatas) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func getTodayStartHour() (start int) {
	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		return
	}
	start, _ = strconv.Atoi(time.Now().In(loc).Format("20060102") + "00")
	return
}

// GetDailyLatestStats 获取当日实时统计的最新后台用户数据
func GetDailyLatestStats(adminUID int64) (models.TicketWorkerHourlyStats, error) {
	obj := models.TicketWorkerHourlyStats{}

	o := orm.NewOrm()
	o.Using(obj.Using())

	err := o.QueryTable(obj.TableName()).Filter("admin_uid", adminUID).
		Filter("ticket_item_id__in", types.TicketItemUrgeM11, types.TicketItemUrgeM12, types.TicketItemRM0).
		Filter("hour__gte", getTodayStartHour()).
		OrderBy("-hour", "-load_num").
		Limit(1).
		One(&obj)

	return obj, err
}

func GetCurrentRanking(ticketItem types.TicketItemEnum, hour int) (list []models.TicketWorkerHourlyStats) {
	obj := models.TicketWorkerHourlyStats{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	where := fmt.Sprintf("WHERE hour= %d and ticket_item_id=%d", hour, ticketItem)

	sqlList := fmt.Sprintf("SELECT * FROM `%s` %s ORDER BY `ranking` ASC", obj.TableName(), where)
	// 查询符合条件的所有条数
	r := o.Raw(sqlList)

	r.QueryRows(&list)

	return
}

func GetPersonalStatsList(adminUID int64, ticketItem types.TicketItemEnum) (chartData DailySalaryChartDatas) {
	obj := models.TicketWorkerHourlyStats{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	where := fmt.Sprintf("WHERE ticket_item_id=%d and admin_uid=%d", ticketItem, adminUID)

	sqlList := fmt.Sprintf("SELECT hour, repay_total FROM `%s` %s ORDER BY `hour` DESC LIMIT 8", obj.TableName(), where)
	// 查询符合条件的所有条数
	r := o.Raw(sqlList)
	var list []models.TicketWorkerHourlyStats

	r.QueryRows(&list)

	salaryRate, _ := config.ValidItemFloat64(getItemSalaryRepayRateConfigName(ticketItem))
	//var lastHour int
	//if len(list) != 0 {
	for _, hourStats := range list {
		h := hourStats.Hour - hourStats.Hour/100*100
		//d := DaySalaryChartData{TimeTag: strconv.Itoa(h), Salary: caculateSalary(salaryRate, hourStats.RepayTotal)}
		d := DaySalaryChartData{Hour: strconv.Itoa(h), StatsHour: hourStats.Hour, Salary: caculateSalary(salaryRate, hourStats.RepayTotal)}
		chartData = append(chartData, d)
	}
	// 	lastHour = list[0].Hour - list[0].Hour/100*100 + 1
	// } else {
	// 	lastHour = getTimeHour(latestStatsDataTimeTag)
	// }
	//

	sort.Sort(chartData)
	if len(chartData) > 0 {
		chartData[len(chartData)-1].Hour = ChartLatestTimeTag
	}

	return chartData
}

// func GetTeamStats(teamSuperRoleID int64) {
// 	//
// 	// 获取
// 	allLeaderRoleIDs := rbac.GetChildRoles()
// 	//
// }

func getTimeHour(unixMill int64) (hour int) {
	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		return
	}
	hour = time.Unix(unixMill/1000, 0).In(loc).Hour()
	return
}

func caculateSalary(salaryRate float64, repayTotalAmount int64) (salary int64) {
	return int64(float64(repayTotalAmount)*salaryRate) / 100
}

// type Group struct {
// 	Leader       int64
// 	Memebers     []int64
// 	MemeberRoles []int64
// 	LeaderRole   int64
// }
//
// func (g *Group) GetAllRoleIDs() {
//
// }
//
// func GetGroup(ticketItem types.TicketItemEnum) {
// 	// 获取工单类型对应的所有角色
// 	roleIDStrs, _ := ticket.CanAssignRoles(ticketItem)
// 	// 将角色拆分成组
// 	groupMap := make(map[int64]Group)
// 	for _, idStr := range roleIDStrs {
// 		id, _ := strconv.ParseInt(idStr, 10, 64)
// 		// 获取 id
// 		roleLevel := rbac.GetRoleLevel(id)
// 		if roleLevel == types.RoleLeader {
// 			groupMap[id] = Group{LeaderRole: id}
// 		} else if roleLevel == types.RoleEmployee {
//
// 		}
// 	}
//
// }
