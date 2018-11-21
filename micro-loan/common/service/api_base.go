package service

import (
	"fmt"
	"strings"

	"micro-loan/common/models"
	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

// 通用的检查必要参数的方法,只检测参数存在,不关心参数值
func checkRequiredParameter(parameter map[string]interface{}, requiredParameter map[string]bool) bool {
	var requiredCheck int = 0
	var rpCopy = make(map[string]bool)
	for rp, v := range requiredParameter {
		rpCopy[rp] = v
	}

	for k, _ := range parameter {
		if requiredParameter[k] {
			requiredCheck++
			delete(rpCopy, k)
		}
	}

	if len(requiredParameter) != requiredCheck {
		var lostParam []string
		for l, _ := range rpCopy {
			lostParam = append(lostParam, l)
		}
		logs.Error("request lost required parameter, parameter:", parameter, fmt.Sprintf("lostParam: [%s]", strings.Join(lostParam, ", ")))
		return false
	}

	return true
}

func AddApiTraceData(beginTime int64, requestUrl string) {
	m := models.ApiStatistics{}
	m.ConsumeTime = tools.GetUnixMillis() - beginTime
	m.RequestUrl = requestUrl
	m.StatisticsDate = tools.GetDate(beginTime / 1000)

	m.Add()
}
