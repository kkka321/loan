package tools

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
)

const (
	SECONDAHOUR      int64 = 3600
	MILLSSECONDAHOUR       = SECONDAHOUR * 1000

	SECONDADAY      int64 = 3600 * 24
	MILLSSECONDADAY       = SECONDADAY * 1000
)

func GetDateFormat(timestamp int64, format string) string {
	if timestamp <= 0 {
		return ""
	}
	tm := time.Unix(timestamp, 0)
	return tm.Format(format)
}

func GetLocalDateFormat(timestamp int64, format string) string {
	tmp := timestamp / 1000

	if tmp <= 0 {
		return "-"
	}

	tm := time.Unix(tmp, 0)
	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)

	return tm.In(local).Format(format)
}

func GetDate(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	tm := time.Unix(timestamp, 0)
	return tm.Format("2006-01-02")
}

func GetDateMH(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	tm := time.Unix(timestamp, 0)
	return tm.Format("2006-01-02 15:04")
}

// 格式化毫秒时间
func MDateMH(timestamp int64) string {
	return GetDateMH(timestamp / 1000)
}

func GetDateMHS(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	tm := time.Unix(timestamp, 0)
	return tm.Format("2006-01-02 15:04:05")
}

// 毫秒,输出印尼时间
func MDateMHS(timestamp int64) string {
	tmp := timestamp / 1000

	if tmp <= 0 {
		return "-"
	}

	tm := time.Unix(tmp, 0)
	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)

	return tm.In(local).Format("2006-01-02 15:04:05")
}

func RFC3339TimeTransfer(datetime string) int64 {

	timeLayout := "2006-01-02T15:04:05Z" //转化所需模板
	loc, _ := time.LoadLocation("")      //获取时区

	tmp, _ := time.ParseInLocation(timeLayout, datetime, loc)
	timestamp := tmp.Unix() * 1000 //转化为时间戳 类型是int64

	return timestamp
}

// 毫秒,输出印尼时间
func MDate2MinuteFormat(timestamp int64) string {
	tmp := timestamp / 1000

	if tmp <= 0 {
		return "-"
	}

	tm := time.Unix(tmp, 0)
	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)

	return tm.In(local).Format("2006-01-02 15:04")
}

// 毫秒,输出印尼时间(时分秒)
func MDateMHSHMS(timestamp int64) string {
	tmp := timestamp / 1000

	if tmp <= 0 {
		return "-"
	}

	tm := time.Unix(tmp, 0)
	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)

	return tm.In(local).Format("15:04:05")
}

// 毫秒,输出印尼时间(月份日时分秒)
func MHSHMS(timestamp int64) string {
	tmp := timestamp / 1000

	if tmp <= 0 {
		return "-"
	}

	tm := time.Unix(tmp, 0)
	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)

	return tm.In(local).Format("01-02 15:04:05")
}

// 获取印尼当天的最后一秒
func GetIDNCurrDayLastSecond() int64 {
	return NaturalDay(0) + (3600*(24-7)-1)*1000
}

func MDateMHSDate(timestamp int64) string {
	tmp := timestamp / 1000

	if tmp <= 0 {
		return ""
	}

	tm := time.Unix(tmp, 0)
	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)

	return tm.In(local).Format("2006-01-02")
}

func MDateMHSLocalDate(timestamp int64) string {
	tmp := timestamp / 1000

	if tmp <= 0 {
		return "-"
	}

	tm := time.Unix(tmp, 0)
	local, _ := time.LoadLocation("Local")

	return tm.In(local).Format("20060102")
}

func MDateMHSLocalDateAllNum(timestamp int64) string {
	tmp := timestamp / 1000

	if tmp <= 0 {
		return "-"
	}

	tm := time.Unix(tmp, 0)
	local, _ := time.LoadLocation("Local")

	return tm.In(local).Format("20060102150405")
}

// GetTodayTimestampByLocalTime 获取今日某一个时间点的Unix 秒时间戳
// timeStr just hour 12, hour:min 12:40  , h:m:s 12:40:30
func GetTodayTimestampByLocalTime(timeStr string) (int64, error) {
	t := time.Now()
	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)
	dateStr := t.In(local).Format("2006-01-02") + " " + timeStr
	//
	count := strings.Count(timeStr, ":")
	var layout string
	switch count {
	case 1:
		layout = "2006-01-02 15:04"
	case 2:
		layout = "2006-01-02 15:04:05"
	case 0:
		layout = "2006-01-02 15"
	default:
		return 0, fmt.Errorf("[GetTodayTimestampByLocalTime] with wrong timeStr format, timeStr: %s", timeStr)
	}

	parse, _ := time.ParseInLocation(layout, dateStr, local)
	return parse.Unix(), nil
}

func LocalYearMonth(timestamp int64) string {
	tmp := timestamp / 1000

	if tmp <= 0 {
		return "-"
	}

	tm := time.Unix(tmp, 0)
	local, _ := time.LoadLocation("Local")

	return tm.In(local).Format("200601")
}

func MDateMHSDateNumber(timestamp int64) string {
	tmp := timestamp / 1000

	if tmp <= 0 {
		return "-"
	}

	tm := time.Unix(tmp, 0)
	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)

	return tm.In(local).Format("20060102")
}

func DateMHSZ(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	tm := time.Unix(timestamp, 0)
	return tm.Format("2006-01-02")
}

// 毫秒
func MDateUTC(timestamp int64) string {
	return DateMHSZ(timestamp / 1000)
}

/*
*   从一种时间格式转为另一种 ，或者转为时间戳
*	@param timestr 即将处理的时间字符串
*	@param fromFormat 当前时间格式  Mon, 02 Jan 2006  MST
*	@param toFormat 目标时间格式   	2006-01-02 15:04:05
*	@param fromFormat 当前时间格式
*	@param unixtime 为真返回时间戳，否则正常转换时间格式
*	@return string []byte
 */
func TimeStrFormat(timestr, fromFormat, toFormat string, unixtime bool) interface{} {
	timeparse, _ := time.Parse(fromFormat, timestr)
	timestsmp := timeparse.Unix()
	if unixtime {
		return timestsmp
	}
	tm := time.Unix(timestsmp, 0)
	return tm.Format(toFormat)

}

// GetTimeParseWithFormat 指定格式的时间字符串转时间戳
func GetTimeParseWithFormat(times, format string) (int64, error) {
	if "" == times {
		return 0, fmt.Errorf("time not valid")
	}

	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)
	parse, err := time.ParseInLocation(format, times, local)
	return parse.Unix(), err
}

func GetTimeParse(times string) int64 {
	if "" == times {
		return 0
	}

	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)
	parse, _ := time.ParseInLocation("2006-01-02 15:04", times, local)
	return parse.Unix()
}

// GetDateParse 用于跑批, 或者需要以 UTC时区为基准的时间解析
func GetDateParse(dates string) int64 {
	if "" == dates {
		return 0
	}
	loc, _ := time.LoadLocation("Local")
	parse, _ := time.ParseInLocation("2006-01-02", dates, loc)
	return parse.Unix()
}

func GetDateParses(dates string) int64 {
	if "" == dates {
		return 0
	}
	loc, _ := time.LoadLocation("Local")
	parse, _ := time.ParseInLocation("2006-01-02 15:04:05", dates, loc)
	return parse.Unix()
}

// Str2TimeByLayout 使用layout将时间字符串转unix时间戳(毫秒)
func Str2TimeByLayout(layout, timeStr string) int64 {
	if "" == timeStr {
		return 0
	}

	loc, _ := time.LoadLocation("Local")
	parse, _ := time.ParseInLocation(layout, timeStr, loc)
	return parse.UnixNano() / 1000000
}

// GetDateParseBackend 所有后台使用
func GetDateParseBackend(dates string) int64 {
	if "" == dates {
		return 0
	}

	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)
	parse, _ := time.ParseInLocation("2006-01-02", dates, local)

	return parse.Unix()
}

//解析北京时间成印尼时间戳
func GetDateParseBackends(dates string) int64 {
	if "" == dates {
		return 0
	}

	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)
	parse, _ := time.ParseInLocation("2006-01-02 15:04:05", dates, local)

	return parse.Unix()
}

// GetDateParseBeijing, 解析北京时间成时间戳
func GetDateParseBeijing(dates string) int64 {
	if "" == dates {
		return 0
	}

	local, _ := time.LoadLocation("Asia/Shanghai")
	parse, _ := time.ParseInLocation("2006-01-02 15:04:05", dates, local)

	return parse.Unix()
}

// 毫秒,输出北京时间
func MDateMHSBeijing(timestamp int64) string {
	tmp := timestamp / 1000

	if tmp <= 0 {
		return "-"
	}

	tm := time.Unix(tmp, 0)
	local, _ := time.LoadLocation("Asia/Shanghai")

	return tm.In(local).Format("2006-01-02 15:04:05")
}

// PareseDateRangeToMillsecondWithSep 将时间范围字符串解析成毫秒时间戳
// start, end, err
func PareseDateRangeToMillsecondWithSep(dateRange string, splitSep string) (int64, int64, error) {
	if len(dateRange) == 0 {
		// 后台正常逻辑, 因此不记录log, 只是返回err, 便于处理
		return 0, 0, errors.New("Empty date range, just ignore it")
	}

	tr := strings.Split(dateRange, splitSep)
	if (len(tr)) != 2 {
		err := fmt.Errorf("[PareseDateRangeToMillsecondWithCustomSep][wrong date range format], (%s) cantnot split to 2 date by (%s)",
			dateRange, splitSep)
		logs.Error(err)
		return 0, 0, err
	}

	start := GetDateParseBackend(tr[0]) * 1000
	end := GetDateParseBackend(tr[1])*1000 + MILLSSECONDADAY
	if start <= 0 || end <= 0 {
		err := fmt.Errorf("[PareseDateRangeToMillsecondWithCustomSep][wrong date range format], (%s) cantnot split to 2 format date like 2006-01-02",
			dateRange)
		logs.Error(err)
		return 0, 0, err
	}

	return start, end, nil
}

// PareseDateRangeToMillsecond 将时间范围字符串解析成毫秒时间戳
// 默认日期分隔符 " - "
// start, end, err
func PareseDateRangeToMillsecond(dateRange string) (start, end int64, err error) {
	splitSep := " - "
	start, end, err = PareseDateRangeToMillsecondWithSep(dateRange, splitSep)
	return
}

// PareseDateRangeToDayRange 将时间范围字符串解析成毫秒时间戳
// 默认日期分隔符 " - "
// start, end, err
func PareseDateRangeToDayRange(dateRange string) (start, end int, err error) {
	splitSep := " - "
	start, end, err = PareseDateRangeToDayRangeWithSep(dateRange, splitSep)
	return
}

// PareseDateRangeToDayRangeWithSep 将时间范围字符串解析成毫秒时间戳
// start, end, err
func PareseDateRangeToDayRangeWithSep(dateRange string, splitSep string) (int, int, error) {
	if len(dateRange) == 0 {
		// 后台正常逻辑, 因此不记录log, 只是返回err, 便于处理
		return 0, 0, errors.New("Empty date range, just ignore it")
	}

	tr := strings.Split(dateRange, splitSep)
	if (len(tr)) != 2 {
		err := fmt.Errorf("[PareseDateRangeToMillsecondWithCustomSep][wrong date range format], (%s) cantnot split to 2 date by (%s)",
			dateRange, splitSep)
		logs.Error(err)
		return 0, 0, err
	}

	start, _ := strconv.Atoi(strings.Replace(tr[0], "-", "", -1))
	end, _ := strconv.Atoi(strings.Replace(tr[1], "-", "", -1))

	if start <= 0 || end <= 0 {
		err := fmt.Errorf("[PareseDateRangeToMillsecondWithCustomSep][wrong date range format], (%s) cantnot split to 2 format date like 2006-01-02",
			dateRange)
		logs.Error(err)
		return 0, 0, err
	}

	return start, end, nil
}

// 取当前系统时间的毫秒
func GetUnixMillis() int64 {
	nanos := time.Now().UnixNano()
	millis := nanos / 1000000

	return millis
}

func TimeNow() int64 {
	return time.Now().Unix()
}

func NaturalDay(offset int64) (um int64) {
	t := time.Now()
	date := GetDate(t.Unix())
	baseUm := GetDateParse(date) * 1000
	offsetUm := MILLSSECONDADAY * offset

	um = baseUm + offsetUm

	return
}

func NaturalDayWithZone(timestamp int64) (um int64) {
	//t := time.Now()

	date := MDateMHSDate(timestamp)
	baseUm := GetDateParseBackend(date) * 1000

	um = baseUm

	return
}

/**
基于指定时间的偏移量
*/
func BaseDayOffset(baseDay int64, offset int64) (um int64) {
	date := GetDate(baseDay / 1000)
	baseUm := GetDateParse(date) * 1000
	offsetUm := MILLSSECONDADAY * offset
	um = baseUm + offsetUm
	return
}

func GetDateRange(begin, end int64) int64 {
	return (end - begin) / SECONDADAY
}

func GetDateRangeMillis(begin, end int64) int64 {
	return (end - begin) / MILLSSECONDADAY
}

// 获取印尼时间周几
func GetNowWeekDayConf() int {
	// 获取当前时周几
	timezone := GetServiceTimezone()
	local, _ := time.LoadLocation(timezone)
	t := time.Now()

	nowDay := t.In(local).Weekday()
	if 0 == nowDay {
		nowDay += 7
	}
	return int(nowDay)
}
func GetToday() (day string) {
	loc, err := time.LoadLocation(GetServiceTimezone())
	if err != nil {
		return
	}
	day = time.Now().AddDate(0, 0, 0).In(loc).Format("20060102")
	return
}

// 返回的单位是秒
func GetMonth(timetag int64) int64 {
	dateStr := GetDateFormat(timetag/1000, "2006-01-02")
	dateStr = dateStr[0:len(dateStr)-2] + "01"

	return GetDateParse(dateStr)
}

func GetDefaultDateRange(startDayOffset, endDayOffset int) string {
	loc, err := time.LoadLocation(GetServiceTimezone())
	if err != nil {
		return ""
	}

	startDay := time.Now().In(loc).AddDate(0, 0, startDayOffset).Format("2006-01-02")
	endDay := time.Now().In(loc).AddDate(0, 0, endDayOffset).Format("2006-01-02")
	return startDay + " - " + endDay
}
