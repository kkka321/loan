package models

import (
	"fmt"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// TICKET_TABLENAME 表名
const TICKET_TABLENAME string = "ticket"

// Ticket 工单,描述数据表结构与结构体的映射
type Ticket struct {
	Id                  int64 `orm:"pk"`
	Title               string
	Description         string
	ItemID              types.TicketItemEnum `orm:"column(item_id)"` // 决定默认优先级和分配给什么角色
	RelatedID           int64                `orm:"column(related_id)"`
	OrderID             int64                `orm:"column(order_id)"`
	CustomerID          int64                `orm:"column(customer_id)"`
	CreateUID           int64                `orm:"column(create_uid)"`
	AssignUID           int64                `orm:"column(assign_uid)"`
	Link                string
	Data                string
	Priority            types.TicketPirorityEnum
	CommunicationWay    int // 默认未知, 1是, 2, 否; 0 未知
	IsEmptyNumber       int // 默认未知, 1 是, 2, 否; 0 未知
	RiskLevel           int
	CaseLevel           string
	Status              types.TicketStatusEnum
	Ctime               int64
	ExpireTime          int64
	AssignTime          int64
	StartTime           int64
	HandleNum           int
	LastHandleTime      int64
	NextHandleTime      int64
	CustomerBestTime    string
	PartialCompleteTime int64
	CompleteTime        int64
	CloseTime           int64
	ApplyEntrustTime    int64
	CloseReason         string
	Utime               int64
	ShouldRepayDate     int64
}

// TableName 返回当前模型对应的表名
func (r *Ticket) TableName() string {
	return TICKET_TABLENAME
}

// Using 返回当前模型的数据库
func (r *Ticket) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *Ticket) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// Insert 插入新记录
func (r *Ticket) Insert() (int64, error) {
	r.Ctime = tools.GetUnixMillis()
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

// Update ..
func (r *Ticket) Update() (num int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	r.Utime = tools.GetUnixMillis()
	//columns = append(columns, "Utime")
	num, err = o.Update(r)

	return
}

// GetTicket 根据主键,查询工单
// 读主库
func GetTicket(id int64) (Ticket, error) {
	var ticket Ticket

	o := orm.NewOrm()
	o.Using(ticket.Using())
	err := o.QueryTable(ticket.TableName()).Filter("Id", id).One(&ticket)

	return ticket, err
}

// GetTicketByItemAndRelatedID 根据 itemID与relatedID,查询工单
func GetTicketByItemAndRelatedID(itemID types.TicketItemEnum, relatedID int64) (Ticket, error) {
	var ticket Ticket

	o := orm.NewOrm()
	o.Using(ticket.Using())
	err := o.QueryTable(ticket.TableName()).
		Filter("RelatedID", relatedID).
		Filter("ItemID", itemID).
		One(&ticket)

	return ticket, err
}

// GetTicketForPhoneVerifyOrInfoReivew 返回电核或者InfoReview产生的工单
func GetTicketForPhoneVerifyOrInfoReivew(relatedID int64) (ticket Ticket, err error) {
	o := orm.NewOrm()
	o.Using(ticket.Using())
	sql := fmt.Sprintf("select * from %s WHERE related_id=%d and item_id IN(%d,%d) order by id desc limit 1",
		ticket.TableName(), relatedID, types.TicketItemInfoReview, types.TicketItemPhoneVerify)
	r := o.Raw(sql)
	err = r.QueryRow(&ticket)
	if err != nil {
		logs.Error("[GetTicketForPhoneVerifyOrInfoReivew] err:", err)
		return
	}
	if ticket.Id == 0 {
		err = fmt.Errorf("[GetTicketForPhoneVerifyOrInfoReivew] cannot find ticket by related ID:%d", relatedID)
		logs.Error(err)
	}
	return
}

// GetTicketByTicketTypeAndRelatedID 根据 ticketType("电核(phone-verify)","催收(urge)","还款提醒(repay-remind)")与relatedID,查询工单
func GetTicketByTicketTypeAndRelatedID(ticketType string, relatedID int64) (Ticket, error) {

	cond := orm.NewCondition()
	cond = cond.And("RelatedID", relatedID)
	if ticketType == "phone-verify" {
		cond = cond.And("ItemID", types.TicketItemPhoneVerify)
	} else if ticketType == "urge" {
		cond = cond.And("ItemID__in", types.TicketItemUrgeM11, types.TicketItemUrgeM12, types.TicketItemUrgeM13, types.TicketItemUrgeM20, types.TicketItemUrgeM30)
	} else if ticketType == "repay-remind" {
		cond = cond.And("ItemID__in", types.TicketItemRepayRemind, types.TicketItemRMAdvance1, types.TicketItemRM0, types.TicketItemRM1)
	}

	var ticket Ticket

	o := orm.NewOrm()
	o.Using(ticket.UsingSlave())
	err := o.QueryTable(ticket.TableName()).
		SetCond(cond).
		One(&ticket)

	return ticket, err
}

func GetTicketIDByRelatedIDS(relatedIDS []int64) (ticketIDs []int64) {
	ids := tools.ArrayToString(relatedIDS, ",")
	o := orm.NewOrm()
	ticket := Ticket{}
	o.Using(ticket.Using())
	sql := fmt.Sprintf("SELECT id FROM `%s` WHERE related_id in(%s) and status in(%d,%d,%d)", ticket.TableName(), ids, types.TicketStatusAssigned, types.TicketStatusProccessing, types.TicketStatusPartialCompleted)
	r := o.Raw(sql)
	r.QueryRows(&ticketIDs)
	return
}

// GetWorkerIncompletedTicketNumByItem 获取工作者某一工单类型下,未完成工单数
// 由于此处方法触发于， ticket被更新之后， 故使用 主库连接， 以防止从库的延迟带来未知的bug
// 注：读主库
func GetWorkerIncompletedTicketNumByItem(adminUID int64, itemID types.TicketItemEnum) (
	partialNum, workingNumWithOutPartial int64, err error) {
	var obj Ticket

	o := orm.NewOrm()
	o.Using(obj.Using())

	statusWhereIn, _ := tools.IntsSliceToWhereInString(types.TicketStatusSliceInDoing())

	sqlCount := fmt.Sprintf("SELECT COUNT(IF(status=%d, id,null)) as partial_num, count(id)  FROM %s WHERE item_id=%d AND status IN(%s) AND assign_uid=%d ",
		types.TicketStatusPartialCompleted, obj.TableName(), itemID, statusWhereIn, adminUID)
	r := o.Raw(sqlCount)
	var total int64
	err = r.QueryRow(&partialNum, &total)
	workingNumWithOutPartial = total - partialNum
	if err != nil {
		logs.Error("[GetWorkerIncompletedTicketNumByItem] sql count err:", err)
	}
	return
}

// GetTodayTotalWillAssignNumByItem 获取某一工单类型下, 今天一共生成多少工单
// 此方法适用于, 日工单类型
func GetTodayTotalWillAssignNumByItem(itemID types.TicketItemEnum) (num int64) {
	var obj Ticket

	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	ctimeStart := tools.NaturalDay(0)
	ctimeEnd := tools.NaturalDay(1)
	// 要确保当天的任何时间查询数据一致: 今天所有将分配的单子 = 所有未分配 + 所有分配时间在今天的单子
	sqlCount := fmt.Sprintf("SELECT COUNT(id) FROM %s WHERE item_id=%d AND ((assign_time>=%d AND assign_time<%d) OR status=%d)",
		obj.TableName(), itemID, ctimeStart, ctimeEnd, types.TicketStatusCreated)
	r := o.Raw(sqlCount)
	err := r.QueryRow(&num)
	if err != nil {
		logs.Error("[GetTodayTotalWillAssignNumByItem] sql count err:", err)
	}
	return
}

// GetTodayTotalAlreadyAssignedNumByItemAssignUID 获取今天已分配给指定worker 指定工单类型,工单数
// 此方法适用于日上线均等分配, 用于查询今天已分配工单数, 保证,日初分配工单操作幂等
func GetTodayTotalAlreadyAssignedNumByItemAssignUID(itemID types.TicketItemEnum, adminUID int64) (num int64) {
	var obj Ticket

	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	ctimeStart := tools.NaturalDay(0)
	ctimeEnd := tools.NaturalDay(1)
	sqlCount := fmt.Sprintf("SELECT COUNT(id) FROM %s WHERE item_id=%d AND assign_time>=%d AND assign_time<%d AND assign_uid=%d",
		obj.TableName(), itemID, ctimeStart, ctimeEnd, adminUID)
	r := o.Raw(sqlCount)
	err := r.QueryRow(&num)
	if err != nil {
		logs.Error("[GetTodayTotalAlreadyAssignedNumByItemAssignUID] sql count err:", err)
	}
	return
}

// GetTicketItemCompleteRateInRange 获取指定工单类型完成率, 时间戳范围, 左闭右开
func GetTicketItemCompleteRateInRange(ticketItem types.TicketItemEnum, startTimestamp, endTimestamp int64) (rate float64) {
	where := fmt.Sprintf("WHERE item_id=%d and ctime>=%d and ctime<%d", ticketItem, startTimestamp, endTimestamp)
	sql := fmt.Sprintf("SELECT count(*) as total,  count(IF(complete_time>0, 1, null)) as complete FROM `%s` %s", TICKET_TABLENAME, where)

	obj := Ticket{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	r := o.Raw(sql)

	container := struct {
		Total    int64
		Complete int64
	}{}
	r.QueryRow(&container)
	if container.Total == 0 {
		return
	}
	rate = float64(container.Complete) / float64(container.Total)
	return
}

// UserTicketCount 描述用户工单量计数
// 被多个方法使用
type UserTicketCount struct {
	Uid int64
	Num int64
}

// GetUserTicketAssignCount 获取指定时间范围，用户被分配了多少工单
func GetUserTicketAssignCount(ticketItem types.TicketItemEnum, startTimestamp, endTimestamp int64) (usersTicketCount []UserTicketCount) {
	where := fmt.Sprintf("WHERE assign_uid>0 AND item_id=%d and assign_time>=%d and assign_time<%d", ticketItem, startTimestamp, endTimestamp)
	sql := fmt.Sprintf("SELECT assign_uid as uid, COUNT(*) as num FROM `%s` %s GROUP BY assign_uid", TICKET_TABLENAME, where)

	obj := Ticket{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	r := o.Raw(sql)
	r.QueryRows(&usersTicketCount)
	return
}

// GetUserTicketCompleteCount 获取指定时间范围，用户完成了多少工单
func GetUserTicketCompleteCount(ticketItem types.TicketItemEnum, startTimestamp, endTimestamp int64) (usersTicketCount []UserTicketCount) {
	where := fmt.Sprintf("WHERE assign_uid>0 AND item_id=%d and complete_time>=%d and complete_time<%d", ticketItem, startTimestamp, endTimestamp)
	sql := fmt.Sprintf("SELECT assign_uid as uid, COUNT(*) as num FROM `%s` %s GROUP BY assign_uid", TICKET_TABLENAME, where)

	obj := Ticket{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	r := o.Raw(sql)
	r.QueryRows(&usersTicketCount)
	return
}

// GetUserTicketLoadCount 获取指定时间范围，用户负载了多少工单
// 注意此处有时间有偏移
func GetUserTicketLoadCount(ticketItem types.TicketItemEnum, fixStartTimestamp, startTimestamp, endTimestamp int64) (usersTicketCount []UserTicketCount) {
	statusBox, err := tools.IntsSliceToWhereInString(types.TicketStatusSliceInDoing())
	if err != nil {
		logs.Error("[GetUserTicketLoadCount] occur err:", err)
		return
	}

	where := fmt.Sprintf(`WHERE assign_uid>0 AND item_id=%d AND assign_time<%d
		 AND ((complete_time>=%d ) OR status in(%s) OR (close_time>=%d ))`,
		ticketItem, endTimestamp, startTimestamp, statusBox, fixStartTimestamp)
	sql := fmt.Sprintf("SELECT assign_uid as uid, COUNT(*) as num FROM `%s` %s GROUP BY assign_uid", TICKET_TABLENAME, where)

	obj := Ticket{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	r := o.Raw(sql)
	r.QueryRows(&usersTicketCount)
	return
}
