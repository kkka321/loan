package service

import "micro-loan/common/tools"

const (
	Pagesize int = 15
)

var monthMap = map[int64]string{
	1: "历史月份",
	tools.GetDateParse("2018-07-01"): "201807",
	tools.GetDateParse("2018-08-01"): "201808",
	tools.GetDateParse("2018-09-01"): "201809",
	tools.GetDateParse("2018-10-01"): "201810",
	tools.GetDateParse("2018-11-01"): "201811",
	tools.GetDateParse("2018-12-01"): "201812",
	tools.GetDateParse("2019-01-01"): "201901",
	tools.GetDateParse("2019-02-01"): "201902",
	tools.GetDateParse("2019-03-01"): "201903",
	tools.GetDateParse("2019-04-01"): "201904",
	tools.GetDateParse("2019-05-01"): "201905",
	tools.GetDateParse("2019-06-01"): "201906",
	tools.GetDateParse("2019-07-01"): "201907",
	tools.GetDateParse("2019-08-01"): "201908",
	tools.GetDateParse("2019-09-01"): "201909",
	tools.GetDateParse("2019-10-01"): "201910",
	tools.GetDateParse("2019-11-01"): "201911",
	tools.GetDateParse("2019-12-01"): "201912",
}

func GetMonthMap() map[int64]string {
	return monthMap
}
