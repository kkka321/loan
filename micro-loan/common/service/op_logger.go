package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/models"
	"micro-loan/common/tools"
)

// OpLoggerTableMap 定义了 op_logger 中包含的所有日志表名的列表
// 新增请直接附加
var OpLoggerTableMap = map[string]string{
	"orders":               "orders",
	"repay_plan":           "repay_plan",
	"customer_risk":        "customer_risk",
	"account_base":         "account_base",
	"account_profile":      "account_profile",
	"overdue_case":         "overdue_case",
	"product":              "product",
	"repay_remind_case":    "repay_remind_case",
	"worker_online_status": "worker_online_status",
	"ticket":               "ticket",
}

// OpLoggerListStru 描述后台 op_logger 日志列表的结构
type OpLoggerListStru struct {
	Id        int64
	RelatedId int64
	OpUid     int64
	OpCode    models.OpCodeEnum
	OpTable   string
	Ctime     int64
}

func opLoggerListCond(condCntr map[string]interface{}) (cond *orm.Condition) {
	cond = orm.NewCondition()

	// 生成查询条件
	if value, ok := condCntr["start_time"]; ok {
		cond = cond.And("ctime__gte", value)
	}
	if value, ok := condCntr["end_time"]; ok {
		cond = cond.And("ctime__lte", value)
	}
	if value, ok := condCntr["opTable"]; ok {
		cond = cond.And("op_table", value)
	}

	return
}

// OpLoggerList 返回符合查询条件的所有记录
// 注意： 当前后台列表，仅需显示部分字段，可优化 models.OpLogger
func OpLoggerList(condCntr map[string]interface{}, page, pagesize int) (list []OpLoggerListStru, total int64, err error) {
	obj := models.OpLogger{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	var cond string
	// start_time 为空, 则查最近三个月的数据, ctime为索引,避免全表扫描
	if f, ok := condCntr["start_time"]; ok {
		cond = fmt.Sprintf("`ctime` >= %d", f)
	} else {
		// 取三个月之前的毫秒时间戳
		before3Month := time.Now().AddDate(0, -3, 0).UnixNano() / 1000000
		cond = fmt.Sprintf("`ctime` >= %d", before3Month)
	}
	if f, ok := condCntr["end_time"]; ok {
		cond += fmt.Sprintf(" AND `ctime` <= %d", f.(int64))
	}
	if f, ok := condCntr["opTable"]; ok {
		cond += fmt.Sprintf(" AND `op_table` = '%s'", tools.Escape(f.(string)))
	}
	if f, ok := condCntr["opCode"]; ok {
		cond += fmt.Sprintf(" AND `op_code` = %d", f)
	}
	if f, ok := condCntr["id"]; ok {
		cond += fmt.Sprintf(" AND `id` = %d", f)
	}

	if f, ok := condCntr["relatedId"]; ok {
		cond += fmt.Sprintf(" AND `related_id` = %d", f)
	}

	if f, ok := condCntr["opUid"]; ok {
		cond += fmt.Sprintf(" AND `op_uid` = %d", f)
	}

	tableName := obj.TableName()
	var sqlList string
	if v, ok := condCntr["month"]; ok {
		if v.(int64) == 1 {
			tableName = obj.OriTableName()
			sqlList = fmt.Sprintf("SELECT `id`,`op_uid`,`op_code`,`op_table`,`ctime` FROM `%s` WHERE %s ORDER BY `id` desc LIMIT %d,%d", tableName, cond, offset, pagesize)
		} else {
			tableName = obj.TableNameByMonth(v.(int64))
			sqlList = fmt.Sprintf("SELECT `id`,`related_id`, `op_uid`,`op_code`,`op_table`,`ctime` FROM `%s` WHERE %s ORDER BY `id` desc LIMIT %d,%d", tableName, cond, offset, pagesize)
		}
	} else {
		sqlList = fmt.Sprintf("SELECT `id`,`related_id`, `op_uid`,`op_code`,`op_table`,`ctime` FROM `%s` WHERE %s ORDER BY `id` desc LIMIT %d,%d", tableName, cond, offset, pagesize)
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` WHERE %s", tableName, cond)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

// GetOpLogger 根据 ID 获取单条 OpLogger 详情
func GetOpLogger(tableName string, id int64) (data models.OpLogger, err error) {
	o := orm.NewOrm()
	o.Using(data.UsingSlave())

	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE id = %d", tableName, id)

	r := o.Raw(sql)
	r.QueryRow(&data)

	return
}

func ConvertInt64tString(in string) string {
	var orgInt64Map = make(map[string]int64)
	json.Unmarshal([]byte(in), &orgInt64Map)
	logs.Info("orgInt64Map:%#v", orgInt64Map)

	var orgStrMap = make(map[string]interface{})
	json.Unmarshal([]byte(in), &orgStrMap)
	logs.Info("orgStrMap:%#v", orgStrMap)

	if len(orgInt64Map) == 0 {
		return in
	}

	for k, v := range orgInt64Map {
		if v > 0 {
			orgStrMap[k] = tools.Int642Str(v)
		}
	}

	out, _ := json.Marshal(orgStrMap)
	return string(out)
}
