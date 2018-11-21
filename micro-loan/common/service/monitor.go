package service

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/lib/mail"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/lib/sms"
	"micro-loan/common/models"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type MonitorDetail struct {
	Num     int
	Date    string
	Timetag int64
}

type MonitorData struct {
	iName int
	Name  string
	List  []MonitorDetail
}

type HistogramData struct {
	Name string
	Num  int
}

func fillMonitorData(sortData map[int]MonitorData, dataType int, num int, date string) {
	if _, ok := sortData[dataType]; !ok {
		d := MonitorData{}
		d.iName = dataType
		d.List = make([]MonitorDetail, 0)
		sortData[dataType] = d
	}

	timetag := tools.GetDateParse(date) * 1000

	v := sortData[dataType]
	v.List = append(v.List, MonitorDetail{num, date, timetag})
	sortData[dataType] = v
}

func GetOrderTotalData(condStr map[string]interface{}, page, pagesize int) (list []MonitorData) {
	obj := models.OrderStatistics{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	qb, _ := orm.NewQueryBuilder(tools.DBDriver())

	where := "1 = 1"

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	qb.Select("*").
		From(obj.TableName()).
		Where(where)

	// 导出 SQL 语句
	sql := qb.String()

	orderBy := "ORDER BY statistics_date"

	sql = fmt.Sprintf("%s %s LIMIT %d, %d", sql, orderBy, offset, pagesize)

	totalList := make([]models.OrderStatistics, 0)
	o.Raw(sql).QueryRows(&totalList)

	var order models.OrderStatistics = models.OrderStatistics{}
	monitor.GetOrderStatistics(tools.NaturalDay(0), &order)
	totalList = append(totalList, order)

	sortData := make(map[int]MonitorData)
	for _, v := range totalList {
		fillMonitorData(sortData, int(types.LoanStatusSubmit), v.Submit, v.StatisticsDate)
		fillMonitorData(sortData, int(types.LoanStatus4Review), v.WaitReview, v.StatisticsDate)
		fillMonitorData(sortData, int(types.LoanStatusReject), v.Reject, v.StatisticsDate)
		fillMonitorData(sortData, int(types.LoanStatusWaitManual), v.WaitManual, v.StatisticsDate)
		fillMonitorData(sortData, int(types.LoanStatusWait4Loan), v.WaitLoan, v.StatisticsDate)
		fillMonitorData(sortData, int(types.LoanStatusLoanFail), v.LoanFail, v.StatisticsDate)
		fillMonitorData(sortData, int(types.LoanStatusWaitRepayment), v.WaitRepayment, v.StatisticsDate)
		fillMonitorData(sortData, int(types.LoanStatusOverdue), v.Overdue, v.StatisticsDate)
		fillMonitorData(sortData, int(types.LoanStatusAlreadyCleared), v.Cleared, v.StatisticsDate)
		fillMonitorData(sortData, int(types.LoanStatusInvalid), v.Invalid, v.StatisticsDate)
		fillMonitorData(sortData, int(types.LoanStatusPartialRepayment), v.PartialRepayment, v.StatisticsDate)
		fillMonitorData(sortData, int(types.LoanStatusIsDoing), v.Loaning, v.StatisticsDate)
		fillMonitorData(sortData, int(types.LoanStatusWaitAutoCall), v.WaitAutoCall, v.StatisticsDate)
	}

	nameMap := types.LoanStatusMap()
	for _, v := range sortData {
		if name, ok := nameMap[types.LoanStatus(v.iName)]; ok {
			v.Name = name
			list = append(list, v)
		}
	}

	return
}

func GetOrderStatisticsData(condStr map[string]interface{}, page, pagesize int) (list []MonitorData) {
	obj := models.OrderStatistics{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	qb, _ := orm.NewQueryBuilder(tools.DBDriver())

	where := "1 = 1"

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	qb.Select("*").
		From(obj.TableName()).
		Where(where)

	// 导出 SQL 语句
	sql := qb.String()

	orderBy := "ORDER BY statistics_date"

	sql = fmt.Sprintf("%s %s LIMIT %d, %d", sql, orderBy, offset, pagesize)

	totalList := make([]models.OrderStatistics, 0)
	o.Raw(sql).QueryRows(&totalList)

	var order models.OrderStatistics = models.OrderStatistics{}
	monitor.GetOrderStatistics(tools.NaturalDay(0), &order)
	totalList = append(totalList, order)

	sortData := make(map[int]MonitorData)
	for _, v := range totalList {
		{
			total := v.WaitLoan + v.Reject
			num := 0
			if total > 0 {
				num = v.WaitLoan * 100 / total
			}
			fillMonitorData(sortData, 1, num, v.StatisticsDate)
		}
		{
			total := v.WaitRepayment + v.LoanFail
			num := 0
			if total > 0 {
				num = v.WaitLoan * 100 / total
			}
			fillMonitorData(sortData, 2, num, v.StatisticsDate)
		}
	}

	nameMap := map[int]string{
		1: "审核通过",
		2: "放款成功",
	}
	for _, v := range sortData {
		if name, ok := nameMap[v.iName]; ok {
			v.Name = name
			list = append(list, v)
		}
	}

	return
}

func GetThirdpartyTotalData(condStr map[string]interface{}, page, pagesize int) (list []MonitorData) {
	obj := models.ThirdpartyStatistics{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	qb, _ := orm.NewQueryBuilder(tools.DBDriver())

	where := "1 = 1"

	thirdparty := 0
	if sValue, ok := condStr["thirdparty"]; ok {
		thirdparty = sValue.(int)
		where = fmt.Sprintf("%s%s%d", where, " AND thirdparty = ", thirdparty)
	}

	qb.Select("*").
		From(obj.TableName()).
		Where(where)

	// 导出 SQL 语句
	sql := qb.String()

	orderBy := "ORDER BY statistics_date"

	sql = fmt.Sprintf("%s %s", sql, orderBy)

	totalList := make([]models.ThirdpartyStatistics, 0)
	o.Raw(sql).QueryRows(&totalList)

	realDatas := monitor.GetThirdpartyStatistics(tools.NaturalDay(0))
	for _, v := range realDatas {
		if v.Thirdparty == thirdparty {
			totalList = append(totalList, v)
		}
	}

	sortData := make(map[int]MonitorData)
	for _, v := range totalList {
		fillMonitorData(sortData, v.Thirdparty, v.Success+v.Fail, v.StatisticsDate)
	}

	nameMap := models.ThirdpartyNameMap
	for _, v := range sortData {
		if name, ok := nameMap[v.iName]; ok {
			v.Name = name
			list = append(list, v)
		}
	}

	return
}

func GetThirdpartyStatisticsData(condStr map[string]interface{}, page, pagesize int) (list []MonitorData) {
	obj := models.ThirdpartyStatistics{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	qb, _ := orm.NewQueryBuilder(tools.DBDriver())

	where := "1 = 1"

	thirdparty := 0
	if sValue, ok := condStr["thirdparty"]; ok {
		thirdparty = sValue.(int)
		where = fmt.Sprintf("%s%s%d", where, " AND thirdparty = ", thirdparty)
	}

	qb.Select("*").
		From(obj.TableName()).
		Where(where)

	// 导出 SQL 语句
	sql := qb.String()

	orderBy := "ORDER BY statistics_date"

	sql = fmt.Sprintf("%s %s", sql, orderBy)

	totalList := make([]models.ThirdpartyStatistics, 0)
	o.Raw(sql).QueryRows(&totalList)

	realDatas := monitor.GetThirdpartyStatistics(tools.NaturalDay(0))
	for _, v := range realDatas {
		if v.Thirdparty == thirdparty {
			totalList = append(totalList, v)
		}
	}

	sortData := make(map[int]MonitorData)
	for _, v := range totalList {
		total := v.Success + v.Fail
		num := 0
		if total > 0 {
			num = v.Success * 100 / total
		}
		fillMonitorData(sortData, 1, num, v.StatisticsDate)
	}

	nameMap := map[int]string{
		1: "通过",
	}
	for _, v := range sortData {
		if name, ok := nameMap[v.iName]; ok {
			v.Name = name
			list = append(list, v)
		}
	}

	return
}

func GetApiStatisticsData(cond map[string]interface{}, page, pagesize int) (list []HistogramData) {
	obj := models.ApiStatistics{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	qb, _ := orm.NewQueryBuilder(tools.DBDriver())

	where := "1 = 1"

	qb.Select("*").
		From(obj.TableName()).
		Where(where)

	// 导出 SQL 语句
	sql := qb.String()

	orderBy := "ORDER BY statistics_date"

	sql = fmt.Sprintf("%s %s", sql, orderBy)

	totalList := make([]models.ApiStatistics, 0)
	o.Raw(sql).QueryRows(&totalList)

	type ApiCounter struct {
		count int
		num   int
	}

	numMap := make(map[string]ApiCounter)
	for _, l := range totalList {
		if v, ok := numMap[l.RequestUrl]; ok {
			v.count++
			v.num += int(l.ConsumeTime)
			numMap[l.RequestUrl] = v
		} else {
			v := ApiCounter{}
			v.count = 1
			v.count = int(l.ConsumeTime)
			numMap[l.RequestUrl] = v
		}
	}

	for k, v := range numMap {
		d := HistogramData{}
		d.Name = k
		if v.count == 0 {
			d.Num = 0
		} else {
			d.Num = v.num * 100 / v.count
		}
		list = append(list, d)
	}

	return
}

func SendNotification(dateKey string, freq int, title, body string) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	runmode := beego.AppConfig.String("runmode")
	if runmode == "dev" {
		return
	}

	now := tools.TimeNow()

	hValue, err := storageClient.Do("GET", dateKey)
	if err != nil || hValue == nil {
		doSendNotification(title, body)
		storageClient.Do("SET", dateKey, now)
		return
	}

	iValue, err := tools.Str2Int64(string(hValue.([]byte)))
	if err != nil {
		doSendNotification(title, body)
		storageClient.Do("SET", dateKey, now)
		return
	}

	if (now - iValue) < int64(freq)*60 {
		logs.Info("[sendMail] skip send now:%d, last:%d, freq:%d", now, iValue, freq)
		return
	}

	doSendNotification(title, body)

	storageClient.Do("SET", dateKey, now)
	return
}

func doSendNotification(title, body string) {
	runmode := beego.AppConfig.String("runmode")
	addr := beego.AppConfig.String("monitor_server")
	sender := beego.AppConfig.String("monitor_sender")
	rcpterStr := beego.AppConfig.String("monitor_rcpter")
	phoneStr := beego.AppConfig.String("monitor_sms")

	datetime := tools.GetDateMHS(tools.TimeNow())
	body = body + "\ndatetime : " + datetime + "\nrunmode : " + runmode

	addrList := strings.Split(rcpterStr, ",")
	for _, v := range addrList {
		rcpter := strings.Trim(v, " ")
		if rcpter == "" {
			continue
		}

		mail.SendMail(title, body, addr, sender, rcpter)
	}

	phoneList := strings.Split(phoneStr, ",")
	for _, v := range phoneList {
		phone := strings.Trim(v, " ")
		if phone == "" {
			continue
		}

		status, err := sms.SendByKey(types.Sms253ID, types.ServiceMonitor, phone, body, int64(911))
		if !status {
			logs.Error("[doSendNotification] send sms error err:%v, body:%s", err, body)
		}
	}
}
