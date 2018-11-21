package performance

import (
	"fmt"
	"micro-loan/common/lib/redis/storage"
	"sort"
	"strconv"

	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"
)

// ProcessChartData 统计图表所需数据模型
type ProcessChartData struct {
	Hour      int     `json:"x"`
	RepayRate float64 `json:"y"`
}

type ProcessChartDatas []ProcessChartData

func (p ProcessChartDatas) Len() int { return len(p) }

func (p ProcessChartDatas) Less(i, j int) bool {
	return p[i].Hour < p[j].Hour
}

func (p ProcessChartDatas) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// GetTodayStats 获取指定admin的今日统计
func GetTodayStats(adminUID int64) (lastestStatsData DailyWorkerProcessData, processChartDatas ProcessChartDatas, err error) {
	today := getTodayTag()
	lastestStatsData = getLastestDailyWorkerProcessData(adminUID, today)
	if lastestStatsData.TicketItem <= 0 {
		err = fmt.Errorf("There is no any stats for admin:%d", adminUID)
		return
	}
	if lastestStatsData.LoadLeftUnpaidPrincipal <= 0 {
		return
	}
	processData := getHoursDailyWorkerProcessData(adminUID, today)

	for k, v := range processData {
		h, _ := strconv.Atoi(k)
		r := RoundFloat64(float64(100*v)/float64(lastestStatsData.LoadLeftUnpaidPrincipal), 2)
		processChartDatas = append(processChartDatas, ProcessChartData{h, r})
	}
	sort.Sort(processChartDatas)

	return
}

func getHoursDailyWorkerProcessData(uid int64, day string) (hoursProcessData map[string]int64) {
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()
	hashName := getHoursRepayAmountHashName(day, uid)
	hoursProcessData, errR := redis.Int64Map(redisCli.Do("HGETALL", hashName))
	if errR != nil {
		logs.Error("[getHoursDailyWorkerProcessData] redis err:", errR)
		return
	}
	logs.Debug("[getHoursDailyWorkerProcessData] get redis map data:%#v", hoursProcessData)

	return
}
