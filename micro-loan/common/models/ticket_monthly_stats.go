package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// TICKET_ITEM_MONTHLY_STATS_TABLENAME 表名
const TICKET_ITEM_MONTHLY_STATS_TABLENAME = "ticket_item_monthly_stats"

// TicketItemMonthlyStats 工单,描述数据表结构与结构体的映射
type TicketItemMonthlyStats struct {
	Id                     int64 `orm:"pk"`
	Date                   int
	TicketItemID           types.TicketItemEnum `orm:"column(ticket_item_id)"` // 决定默认优先级和分配给什么角色
	HandleNum              int64
	CompleteNum            int64
	CompleteRate           float64
	OverdueRateAchieveDays int
	Ctime                  int64
	Utime                  int64
}

// TableName 返回当前模型对应的表名
func (r *TicketItemMonthlyStats) TableName() string {
	return TICKET_ITEM_MONTHLY_STATS_TABLENAME
}

// Using 返回当前模型的数据库
func (r *TicketItemMonthlyStats) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *TicketItemMonthlyStats) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// GetMonthlyStatsByDateAndTicketItem 根据日期和itemID 获取,获取月统计数据
func GetMonthlyStatsByDateAndTicketItem(ticketItem types.TicketItemEnum, date int) (TicketItemMonthlyStats, error) {
	o := orm.NewOrm()

	obj := TicketItemMonthlyStats{}

	o.Using(obj.UsingSlave())

	err := o.QueryTable(obj.TableName()).Filter("ticket_item_id", ticketItem).
		Filter("date", date).One(&obj)

	return obj, err
}
