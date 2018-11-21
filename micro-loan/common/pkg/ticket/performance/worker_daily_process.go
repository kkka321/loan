package performance

import (
	"encoding/json"
	"fmt"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
)

type ticketItemDailyProcess struct {
	ticketItem     types.TicketItemEnum
	day            int
	startTimestamp int64
	endTimestamp   int64
	baseModel      models.TicketWorkerDailyStats
	statsData      map[int64]models.TicketWorkerDailyStats
}

const startHour = 9
const endHour = 23

// data.RepayTotal
// data.DiffTargetRepay
// data.LoadLeftUnpaidPrincipal

type DailyWorkerProcessData struct {
	TicketItem              types.TicketItemEnum
	RepayTotal              int64
	DiffTargetRepay         int64
	LoadLeftUnpaidPrincipal int64
	RepayAmountRate         float64
	StatsHour               int
}

// DailyWorkerProcessStatsLastHour 执行当日上个小时的统计
func DailyWorkerProcessStatsLastHour() {
	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		return
	}
	hour := time.Now().In(loc).Hour()
	if hour-startHour < 0 {
		return
	}

	day := time.Now().In(loc).Format("20060102")

	DailyWorkerProcessStats(day, hour)
}

func getTodayTag() string {
	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		panic("ServiceTime zone config err:" + err.Error())
	}
	return time.Now().In(loc).Format("20060102")
}

// DailyWorkerProcessHistoryStatsByDay do history stats
func DailyWorkerProcessHistoryStatsByDay(day string) {
	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		return
	}
	today := time.Now().In(loc).Format("20060102")
	intDay, err := strconv.Atoi(day)
	if err != nil {
		logs.Error("[DailyWorkerProcessHistoryStatsByDay] stats day(%s) with a wrong format", day)
		return
	}
	intToday, _ := strconv.Atoi(today)

	var stopHour int
	if intDay > intToday {
		logs.Error("[DailyWorkerProcessHistoryStatsByDay] stats day(%s) cannot beyond today(%s)", day, today)
		return
	} else if intDay == intToday {
		stopHour = time.Now().In(loc).Hour()
	} else {
		stopHour = endHour
	}

	for statsHour := startHour; statsHour <= stopHour; statsHour++ {
		DailyWorkerProcessStats(day, statsHour)
	}
}

// DailyWorkerProcessStats 执行当日上个小时的统计
func DailyWorkerProcessStats(day string, hour int) {
	intDay, err := strconv.Atoi(day)
	if err != nil {
		logs.Error("[NewItemDailyStats] day(%s) is invalid:%v", day, err)
		return
	}

	needStatsTicketItem := []types.TicketItemEnum{
		types.TicketItemRM0,
		types.TicketItemRM1,
		types.TicketItemUrgeM11,
		types.TicketItemUrgeM12,
	}
	statsData := make(map[int64]DailyWorkerProcessData)
	for _, ticketItem := range needStatsTicketItem {
		itemStatsData := NewItemDailyProcessStatsAndRun(ticketItem, intDay, hour)
		for uid, sData := range itemStatsData {
			if val, ok := statsData[uid]; !ok || (val.TicketItem != ticketItem && val.LoadLeftUnpaidPrincipal < sData.LoadLeftUnpaidPrincipal) {
				d := DailyWorkerProcessData{ticketItem, sData.RepayTotal, sData.DiffTargetRepay, sData.LoadLeftUnpaidPrincipal, sData.RepayAmountRate, hour}
				statsData[uid] = d
			}
		}
	}

	saveStatsData(&statsData, day)
}

type DailyWorkerProcess struct {
}

func statsDataExpirteTime(day string) (expire int64) {
	loc, err := time.LoadLocation(tools.GetServiceTimezone())
	if err != nil {
		return
	}
	startTime, err := time.ParseInLocation("20060102", day, loc)
	if err != nil {
		return
	}
	expire = startTime.AddDate(0, 0, 3).Unix()
	return
}

func saveStatsData(statsData *map[int64]DailyWorkerProcessData, day string) {
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()

	lastestProcessHashName := getLastestDailyWorkerProcessDataHashName(day)
	expireTimestamp := statsDataExpirteTime(day)
	defer func() {
		if reply, _ := redis.Int64(redisCli.Do("TTL", lastestProcessHashName)); reply == -1 {
			redisCli.Do("EXPIREAT", lastestProcessHashName, expireTimestamp)
		}
	}()
	for uid, d := range *statsData {
		b, _ := json.Marshal(d)
		//lastestStatsData := getLastestDailyWorkerProcessData(uid, day)
		// TODO 释放注释
		//if lastestStatsData.StatsHour <= d.StatsHour {
		redisCli.Do("HSET", lastestProcessHashName, uid, b)
		//}
		hoursHashName := getHoursRepayAmountHashName(day, uid)
		redisCli.Do("HSET", hoursHashName, d.StatsHour, d.RepayTotal)
		if reply, _ := redis.Int64(redisCli.Do("TTL", hoursHashName)); reply == -1 {
			redisCli.Do("EXPIREAT", hoursHashName, expireTimestamp)
		}
	}
}

func getLastestDailyWorkerProcessData(uid int64, day string) (lastestStatsData DailyWorkerProcessData) {

	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()
	lastestProcessHashName := getLastestDailyWorkerProcessDataHashName(day)
	reply, errR := redis.Bytes(redisCli.Do("HGET", lastestProcessHashName, uid))
	if errR != nil {
		if errR != redis.ErrNil {
			logs.Error("[saveStatsData] redis err:")
		}
	} else {
		json.Unmarshal(reply, &lastestStatsData)
	}
	return
}

// NewItemDailyProcessStatsAndRun 创建统计实体
func NewItemDailyProcessStatsAndRun(ticketItem types.TicketItemEnum, intDay int, nowHour int) map[int64]models.TicketWorkerDailyStats {
	stats := new(ticketItemDaily)
	stats.ticketItem = ticketItem
	stats.day = intDay
	var err error
	stats.startTimestamp, _, err = parseUnixMillTimestampRange(strconv.Itoa(intDay))
	stats.endTimestamp = stats.startTimestamp + int64(nowHour*int(time.Hour/time.Millisecond))
	if err != nil {
		logs.Error("[NewItemDailyStats] date parse err", err)
		return nil
	}
	stats.baseModel = models.TicketWorkerDailyStats{
		Date:         intDay,
		TicketItemID: ticketItem,
		Ctime:        tools.GetUnixMillis(),
	}

	stats.statsData = map[int64]models.TicketWorkerDailyStats{}

	stats.doRepayAmountStats()

	return stats.statsData
}

func getLastestDailyWorkerProcessDataHashName(day string) string {
	return beego.AppConfig.String("latest_daily_worker_process_hash") + day
}

func getHoursRepayAmountHashName(day string, uid int64) string {
	return beego.AppConfig.String("admin_hours_repay_amount_hash") + day + ":" + strconv.FormatInt(uid, 10)
}

func (s *ticketItemDailyProcess) stats() {
	local, _ := time.LoadLocation(tools.GetServiceTimezone())
	h := time.Now().In(local).Hour() - startHour
	fmt.Println(h)

	// 获取 分案本金
	// 获取 已还总额

	fixStartTime := fixStartTimeForLoad(s.startTimestamp)

	userOrderContainer := []struct {
		Uid     int64
		OrderId int64
	}{}
	statusBox, err := tools.IntsSliceToWhereInString(types.TicketStatusSliceInDoing())
	if err != nil {
		logs.Error("[GetUserTicketLoadCount] occur err:", err)
		return
	}

	where := fmt.Sprintf("WHERE item_id=%d AND assign_time<%d AND (complete_time>=%d  OR  status in(%s))",
		s.ticketItem, s.endTimestamp, fixStartTime, statusBox)
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
		data.RepayTotal = repayInterest
		// 下属计算 LoadLeftUnpaidPrincipal 兼容任意时间的补算, 任何历史时间的待还本金都是准确的
		nowUnpaidPrincipal, _ := models.GetOrdersLeftUnpaidPrincipal(orders)
		alreadyPaidPrincipalSinceStatsDay, alreadyPaidInterest, _ := models.GetOrdersRepayPrincipalAndInterest(orders, s.startTimestamp, tools.GetUnixMillis())
		// 分案本金=工单分配的剩余应还本金之和
		data.LoadLeftUnpaidPrincipal = nowUnpaidPrincipal + alreadyPaidPrincipalSinceStatsDay

		// 回款总金额=已还本金+已还息费
		alreadyPaidTotal := alreadyPaidPrincipalSinceStatsDay + alreadyPaidInterest
		// 回收率
		data.RepayAmountRate = GetRepayAmountRate(s.ticketItem, orders, alreadyPaidTotal, data.LoadLeftUnpaidPrincipal, alreadyPaidInterest)

		//configName := getUrgeRepayRateStandardConfigName(s.ticketItem)
		//data.RepayAmountRate, _ = config.ValidItemFloat64(configName)
		//data.DiffTargetRepay = int64(float64(data.LoadLeftUnpaidPrincipal)*data.RepayAmountRate/100) - data.RepayTotal
		// 目标回收率
		targetRepayRate := GetTargetRepayRateByTicketItem(s.ticketItem, types.TicketMyProcess)
		// 差值金额
		data.DiffTargetRepay = int64(float64(data.LoadLeftUnpaidPrincipal)*targetRepayRate/100) - data.RepayTotal

		s.statsData[uid] = data
	}

	return

}

// 回款
// delete yesterday
// hash:LoadLeftUnpaidPrincipal
// hash:RepayTotal
// 计算 repayTotal / LoadLeftUnpaidPrincipal
