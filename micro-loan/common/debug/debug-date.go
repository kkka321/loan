package main

import (
	"fmt"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/tools"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	//logs.Debug("debug api ...")

	//testDate()

	//date()
	// QueryTask()

	//datetime := "2018-07-17T03:42:53.479Z"
	//
	////datetime.Unix
	//
	//udbtime, err := time.Parse.EST(RFC3339, dbtime)
	////time.Parse.

	datetime := "2018-07-17T03:42:53.479Z" //待转化为时间戳的字符串

	//日期转化为时间戳
	timeLayout := "2006-01-02T15:04:05Z" //转化所需模板
	loc, _ := time.LoadLocation("")      //获取时区

	tmp, _ := time.ParseInLocation(timeLayout, datetime, loc)
	timestamp := tmp.Unix() //转化为时间戳 类型是int64
	fmt.Println(timestamp)

	fmt.Println(time.Now().Unix())

	//logs.Debug(tools.MDateMHS(1531710844005))
	//时间戳转化为日期
	datetime = time.Unix(timestamp, 0).Format(timeLayout)
	fmt.Println(datetime)

}

func date() {

	// Mon, 02 Jan 2006 15:04:05 MST
	// May 25th, 2018 14:49:06
	// timestr := strings.Replace(td.Text(), "th", "", -1)
	// timeStr := TimeStrFormat(timestr, "Jan 02, 2006 15:04:05", "2006-01-02 15:04:05", false)
	// timestamp := util.TimeStrFormat(timestr, "Jan 02, 2006 15:04:05", "2006-01-02 15:04:05", true)
	// fmt.Println(timeStr)   //2018-05-25 22:49:06
	// fmt.Println(timestamp) //1527259746

	// fmt.Print(time.shortDayNames)
	// strToTime("May 25th, 2018 14:49:06")
	strToTime("Tue, 24 Aug 2018 13:01:35 MST")

}

func strToTime(dateStr string) (timestamp int64) {

	fmt.Println("dateStr: ", dateStr)
	var longDayNames = []string{
		"Sunday",
		"Monday",
		"Tuesday",
		"Wednesday",
		"Thursday",
		"Friday",
		"Saturday",
	}
	var shortDayNames = []string{
		"Sun",
		"Mon",
		"Tue",
		"Wed",
		"Thu",
		"Fri",
		"Sat",
	}
	var shortMonthNames = []string{
		"---",
		"Jan",
		"Feb",
		"Mar",
		"Apr",
		"May",
		"Jun",
		"Jul",
		"Aug",
		"Sep",
		"Oct",
		"Nov",
		"Dec",
	}
	var longMonthNames = []string{
		"---",
		"January",
		"February",
		"March",
		"April",
		"May",
		"June",
		"July",
		"August",
		"September",
		"October",
		"November",
		"December",
	}

	reg := regexp.MustCompile(`\d{1,2}:\d{2}:\d{2}`)
	dateStr = reg.ReplaceAllString(dateStr, "15:04:05")
	reg = regexp.MustCompile(`\d{4}`)
	dateStr = reg.ReplaceAllString(dateStr, "2006")

	dateSlice := strings.Split(dateStr, " ")

	for k, v := range dateSlice {

		dayShort := false
		monthShort := false

		if tools.InSlice(v, shortDayNames) {
			fmt.Println("shortDayNames====2")
			dateSlice[k] = "Mon"
			dayShort = true
		}
		if dayShort == false && tools.InSlice(v, longDayNames) {
			fmt.Println("longDayNames===1")
			dateSlice[k] = "Monday"
		}

		if tools.InSlice(v, shortMonthNames) {
			fmt.Println("shortMonthNames======3")
			dateSlice[k] = "Jan"
			monthShort = true
		}
		if monthShort == false && tools.InSlice(v, longMonthNames) {
			fmt.Println("longMonthNames======4")
			dateSlice[k] = "January"
		}

	}

	fmt.Print(dateSlice)
	os.Exit(0)

	return
}

/*
*   从一种时间格式转为另一种 ，或者转为时间戳
*	@param timestr 即将处理的时间字符串
*	@param fromFormat 当前时间格式  Mon, 02 Jan 2006 15:04:05 MST
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

func testDate() {
	//t := tools.NaturalDay(0) + 3600*(24+7)*1000
	t := tools.NaturalDay(0) + 3600*(24)*1000
	fmt.Println("t = ", t)
}
