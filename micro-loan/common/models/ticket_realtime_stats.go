package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// TICKET_WORKER_HOURLY_STATS_TABLENAME 表名
const TICKET_WORKER_REALTIME_STATS_TABLENAME = "ticket_worker_realtime_stats"

// TicketWorkerRealtimeStats 工单,描述数据表结构与结构体的映射
type TicketWorkerRealtimeStats struct {
	Id                      int64                `orm:"pk"`
	TimeTag                 int64                // 毫秒
	TicketItemID            types.TicketItemEnum `orm:"column(ticket_item_id)"` // 决定默认优先级和分配给什么角色
	AdminUID                int64                `orm:"column(admin_uid)"`
	AssignNum               int64
	HandleNum               int64
	CompleteNum             int64
	LoadNum                 int64
	LoadLeftUnpaidPrincipal int64 // 负载未还本金， 包括今天已还的
	RepayPrincipal          int64
	RepayInterest           int64
	RepayTotal              int64
	RepayAmountRate         float64
	TargetRepayRate         float64
	Ranking                 int
	DiffTargetRepay         int64
	Ctime                   int64
	Utime                   int64
}

// TableName 返回当前模型对应的表名
func (r *TicketWorkerRealtimeStats) TableName() string {
	return TICKET_WORKER_REALTIME_STATS_TABLENAME
}

// Using 返回当前模型的数据库
func (r *TicketWorkerRealtimeStats) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *TicketWorkerRealtimeStats) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// Insert 插入新记录
func (r *TicketWorkerRealtimeStats) Insert() (int64, error) {
	r.Ctime = tools.GetUnixMillis()
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

// Update ..
func (r *TicketWorkerRealtimeStats) Update() (num int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	r.Utime = tools.GetUnixMillis()
	//columns = append(columns, "Utime")
	num, err = o.Update(r)

	return
}

// // GetTicketWorkerPerformanceCountByDateRange 根据日统计表获取特定工单类型月统计数据
// func GetTicketWorkerPerformanceCountByHourRange(ticketItem types.TicketItemEnum, startDay, endDay int) (handleStats, completeStats int64) {
// 	where := fmt.Sprintf("WHERE ticket_item_id=%d and hour>=%d and hour<=%d", ticketItem, startDay, endDay)
// 	sql := fmt.Sprintf("SELECT SUM(handle_num) as handle_stats,  SUM(complete_num) as complete_stats FROM `%s` %s", TICKET_WORKER_HOURLY_STATS_TABLENAME, where)
//
// 	obj := TicketWorkerRealtimeStats{}
// 	o := orm.NewOrm()
// 	o.Using(obj.UsingSlave())
// 	r := o.Raw(sql)
//
// 	container := struct {
// 		HandleStats   int64
// 		CompleteStats int64
// 	}{}
// 	r.QueryRow(&container)
// 	handleStats = container.HandleStats
// 	completeStats = container.CompleteStats
// 	return
// }
