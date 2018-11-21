package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// TICKET_WORKER_DAILY_STATS_TABLENAME 表名
const TICKET_WORKER_DAILY_PROCESS_STATS_TABLENAME = "ticket_worker_daily_process_stats"

// TicketWorkerDailyProcessStats 工单,描述数据表结构与结构体的映射
type TicketWorkerDailyProcessStats struct {
	Id                      int64 `orm:"pk"`
	Date                    int
	TicketItemID            types.TicketItemEnum `orm:"column(ticket_item_id)"` // 决定默认优先级和分配给什么角色
	AdminUID                int64                `orm:"column(admin_uid)"`
	LoadLeftUnpaidPrincipal int64                // 负载未还本金， 包括今天已还的
	RepayTotal              int64
	H1Repay                 int64
	RepayAmountRate         float64
	DiffTargetRepay         int64
	RepayAmountStandardRate int64
	Ctime                   int64
	Utime                   int64
}

// TableName 返回当前模型对应的表名
func (r *TicketWorkerDailyProcessStats) TableName() string {
	return TICKET_WORKER_DAILY_STATS_TABLENAME
}

// Using 返回当前模型的数据库
func (r *TicketWorkerDailyProcessStats) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *TicketWorkerDailyProcessStats) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// Insert 插入新记录
func (r *TicketWorkerDailyProcessStats) Insert() (int64, error) {
	r.Ctime = tools.GetUnixMillis()
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

// Update ..
func (r *TicketWorkerDailyProcessStats) Update() (num int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	r.Utime = tools.GetUnixMillis()
	//columns = append(columns, "Utime")
	num, err = o.Update(r)

	return
}
