package service

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"

	"micro-loan/common/lib/redis/cache"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/google/push"
	"micro-loan/common/pkg/schema_task"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type PushSchemaInfo struct {
	Id           int64
	SchemaMode   types.SchemaMode
	SchemaStatus types.SchemaStatus
	SchemaTime   string
	TaskId       int64
	Ctime        int64
	TaskName     string
	MessageType  int
	PushWay      int
	PushTarget   types.PushTarget
	Title        string
	StartDate    int64
	EndDate      int64
}

type CouponSchemaInfo struct {
	Id           int64
	SchemaMode   types.SchemaMode
	SchemaStatus types.SchemaStatus
	SchemaTime   string
	TaskId       int64
	Ctime        int64
	TaskName     string
	CouponId     int64
	CouponTarget types.CouponTarget
	StartDate    int64
	EndDate      int64
}

type SmsSchemaInfo struct {
	Id           int64
	SchemaMode   types.SchemaMode
	SchemaStatus types.SchemaStatus
	SchemaTime   string
	TaskId       int64
	Ctime        int64
	TaskName     string
	Sender       types.SmsServiceID
	SmsTarget    types.SmsTarget
	StartDate    int64
	EndDate      int64
}

var funcMap map[string]func(int64) error = map[string]func(int64) error{
	"push_message": schema_task.PushMessage,
	"coupon":       schema_task.DistributeCoupon,
	"send_message": schema_task.SmsMessage,
}

func PublishData(ch string, data interface{}) {
	logs.Info("[PublishData] ch:%s, data:%v", ch, data)

	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	d, _ := json.Marshal(data)
	_, err := cacheClient.Do("PUBLISH", ch, d)

	if err != nil {
		logs.Warn("[PublishData] ch:%s, data:%v, err:%v", ch, data, err)
	}
}

func RunSchemaManual(id int64, curTime int64) {
	logs.Info("[RunSchemaManual] start schema schemaId:%d, time:%d", id, curTime)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[RunSchemaManual] panic schemaId:%d, err:%v", id, x)
			logs.Error(tools.FullStack())

			record := models.SchemaRecord{}
			record.Ctime = tools.GetUnixMillis()
			record.SchemaId = id
			record.Result = fmt.Sprintf("%v", x)
			record.Insert()

			updateSchemaStatus(id, types.SchemaStatusError)
		}
	}()

	info, err := models.GetSchemaInfo(id)
	if err != nil {
		logs.Error("[RunSchemaManual] GetSchemaInfo error schemaId:%d, err:%v", id, err)
		return
	}

	v, ok := funcMap[info.FuncName]
	if !ok {
		logs.Error("[RunSchemaManual] unexcept handler func:%s", info.FuncName)
		return
	}

	oldstatus := info.SchemaStatus
	if oldstatus != types.SchemaStatusOff {
		oldstatus = types.SchemaStatusOn
	}
	updateSchemaStatus(id, types.SchemaStatusRunning)

	err = v(info.TaskId)
	if err != nil {
		logs.Error("[RunSchemaManual] task return error schemaId:%d, err:%v", id, err)
	}

	record := models.SchemaRecord{}
	record.Ctime = tools.GetUnixMillis()
	record.SchemaId = id
	if err != nil {
		record.Result = err.Error()
	}
	record.Insert()

	if err != nil {
		updateSchemaStatus(id, types.SchemaStatusError)
	} else {
		updateSchemaStatus(id, oldstatus)
	}
}

func RunSchema(id int64, curTime int64) {
	defer func() {
		if x := recover(); x != nil {
			logs.Error("[RunSchema] panic schemaId:%d, err:%v", id, x)
			logs.Error(tools.FullStack())

			record := models.SchemaRecord{}
			record.Ctime = tools.GetUnixMillis()
			record.SchemaId = id
			record.Result = fmt.Sprintf("%v", x)
			record.Insert()

			updateSchemaStatus(id, types.SchemaStatusError)
		}
	}()

	redisLock := beego.AppConfig.String("schema_lock_set")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	nowStr := tools.GetDate(tools.NaturalDay(0) / 1000)
	lastStr := tools.GetDate(tools.NaturalDay(-1) / 1000)

	nowKey := redisLock + nowStr
	lastKey := redisLock + lastStr

	num, _ := redis.Int(storageClient.Do("EXISTS", lastKey))
	if num > 0 {
		storageClient.Do("DEL", lastKey)
	}

	redisField := fmt.Sprintf("%d_%d", id, curTime)
	num, err := redis.Int(storageClient.Do("HSETNX", nowKey, redisField, tools.GetUnixMillis()))
	if err == nil && num == 0 {
		logs.Error("[RunSchema] schema already done schemaId:%d, time:%d", id, curTime)
		return
	}

	info, err := models.GetSchemaInfo(id)
	if err != nil {
		logs.Error("[RunSchema] GetSchemaInfo error schemaId:%d, err:%v", id, err)
		return
	}

	if info.SchemaMode == types.SchemaModeManual {
		if info.StartDate != tools.NaturalDay(0) {
			logs.Debug("[RunSchema] update manual schema status schemaId:%d", id)

			info.SchemaStatus = types.SchemaStatusOff
			info.Utime = tools.GetUnixMillis()
			info.Update()

			PublishData(info.TableName(), info)

			return
		}
	}

	v, ok := funcMap[info.FuncName]
	if !ok {
		logs.Error("[RunSchema] unexcept handler func:%s", info.FuncName)
		return
	}

	updateSchemaStatus(id, types.SchemaStatusRunning)

	err = v(info.TaskId)
	if err != nil {
		logs.Error("[RunSchema] task return error schemaId:%d, err:%v", id, err)
	}

	record := models.SchemaRecord{}
	record.Ctime = tools.GetUnixMillis()
	record.SchemaId = id
	if err != nil {
		record.Result = err.Error()
	}
	record.Insert()

	if err != nil {
		updateSchemaStatus(id, types.SchemaStatusError)
	} else {
		updateSchemaStatus(id, types.SchemaStatusOn)
	}
}

func updateSchemaStatus(id int64, status types.SchemaStatus) {
	info, err := models.GetSchemaInfo(id)
	if err == nil {
		info.SchemaStatus = status
		info.Utime = tools.GetUnixMillis()
		info.Update()
	}

	if info.FuncName == "push_message" {
		task, err := models.GetPushTask(info.TaskId)
		if err == nil {
			task.TaskStatus = status
			task.Utime = tools.GetUnixMillis()
			task.Update()
		}
	} else if info.FuncName == "send_message" {
		task, err := models.GetSmsTask(info.TaskId)
		if err == nil {
			task.TaskStatus = status
			task.Utime = tools.GetUnixMillis()
			task.Update()
		}
	} else if info.FuncName == "coupon" {
		task, err := models.GetCouponTask(info.TaskId)
		if err == nil {
			task.TaskStatus = status
			task.Utime = tools.GetUnixMillis()
			task.Update()
		}
	}
}

func StopSchema(id int64) {
	info, err := models.GetSchemaInfo(id)
	if err != nil {
		return
	}

	task, err := models.GetPushTask(info.TaskId)
	if err != nil {
		return
	}

	if info.SchemaMode != types.SchemaModeManual {
		return
	}

	info.SchemaStatus = types.SchemaStatusOff
	info.Utime = tools.GetUnixMillis()
	info.Update()

	task.TaskStatus = types.SchemaStatusOff
	task.Utime = tools.GetUnixMillis()
	task.Update()
}

func QueryPushList(condStr map[string]interface{}, page, pagesize int) (list []PushSchemaInfo, count int64, err error) {
	o := orm.NewOrm()
	s := models.SchemaInfo{}
	p := models.PushTask{}
	o.Using(s.UsingSlave())

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pagesize
	cond := "1=1"

	if f, ok := condStr["task_name"]; ok {
		cond = fmt.Sprintf("%s%s'%s'", cond, " AND p.task_name = ", f.(string))
	}
	if f, ok := condStr["message_type"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND p.message_type = ", f.(int))
	}
	if f, ok := condStr["schema_time"]; ok {
		cond = fmt.Sprintf("%s%s'%%%s%%'", cond, " AND s.schema_time LIKE ", f.(string))
	}
	if f, ok := condStr["push_way"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND p.push_way = ", f.(int))
	}
	if f, ok := condStr["schema_mode"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND s.schema_mode = ", f.(int))
	}
	if f, ok := condStr["schema_status"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND s.schema_status = ", f.(int))
	}

	sql := fmt.Sprintf(`SELECT s.id, s.schema_mode, s.schema_status, s.schema_time, s.task_id, s.ctime, s.start_date, s.end_date,
p.task_name, p.message_type, p.push_way, p.push_target, p.title
FROM %s p LEFT JOIN %s s ON s.task_id = p.id WHERE `,
		p.TableName(), s.TableName()) + cond

	orderBy := "ORDER BY s.task_id desc"

	limit := fmt.Sprintf("LIMIT %d, %d", offset, pagesize)

	sqlData := fmt.Sprintf("%s %s %s", sql, orderBy, limit)

	r := o.Raw(sqlData)
	_, err = r.QueryRows(&list)

	sqlCount := fmt.Sprintf(`SELECT count(s.id)
FROM %s p LEFT JOIN %s s ON s.task_id = p.id WHERE `,
		p.TableName(), s.TableName()) + cond

	r = o.Raw(sqlCount)
	r.QueryRow(&count)

	return
}

func QueryPushRecord(id int64, page, pagesize int) (list []models.PushTaskRecord, count int64, err error) {
	o := orm.NewOrm()
	c := models.PushTaskRecord{}
	o.Using(c.UsingSlave())

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pagesize
	if page == 1 {
		pagesize--
	} else {
		offset--
	}

	sql := fmt.Sprintf(`SELECT *
FROM %s WHERE task_id = %d`,
		c.TableName(), id)

	orderBy := "ORDER BY push_date desc"

	limit := fmt.Sprintf("LIMIT %d, %d", offset, pagesize)

	sqlData := fmt.Sprintf("%s %s %s", sql, orderBy, limit)

	r := o.Raw(sqlData)
	_, err = r.QueryRows(&list)

	r = o.Raw(sqlData)
	r.QueryRow(&count)

	if page == 1 {
		storageClient := storage.RedisStorageClient.Get()
		defer storageClient.Close()

		nowStr := tools.MDateMHSDate(tools.GetUnixMillis())
		key := fmt.Sprintf("push:%d:%s", id, nowStr)
		totalNum, _ := redis.Int(storageClient.Do("HGET", key, push.MessageKeyTotal))
		succNum, _ := redis.Int(storageClient.Do("HGET", key, push.MessageKeySucc))
		readNum, _ := redis.Int(storageClient.Do("HGET", key, push.MessageKeyRead))

		record := models.PushTaskRecord{}
		record.TaskId = id
		record.ReadNum = readNum
		record.TotalNum = totalNum
		record.SuccNum = succNum
		record.PushDate = tools.GetUnixMillis()
		if record.TotalNum > 0 {
			record.ReadRate = record.ReadNum * 100 / record.TotalNum
		}

		newList := make([]models.PushTaskRecord, 0)
		newList = append(newList, record)
		newList = append(newList, list...)
		list = newList

		count++
	}

	return
}

func QueryCouponList(condStr map[string]interface{}, page, pagesize int) (list []CouponSchemaInfo, count int64, err error) {
	o := orm.NewOrm()
	s := models.SchemaInfo{}
	c := models.CouponTask{}
	o.Using(s.UsingSlave())

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pagesize
	cond := "s.func_name = 'coupon' "

	if f, ok := condStr["task_name"]; ok {
		cond = fmt.Sprintf("%s%s'%s'", cond, " AND c.task_name = ", f.(string))
	}
	if f, ok := condStr["coupon_id"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND c.coupon_id = ", f.(int64))
	}
	if f, ok := condStr["schema_time"]; ok {
		cond = fmt.Sprintf("%s%s'%%%s%%'", cond, " AND s.schema_time LIKE ", f.(string))
	}
	if f, ok := condStr["schema_mode"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND s.schema_mode = ", f.(int))
	}
	if f, ok := condStr["schema_status"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND s.schema_status = ", f.(int))
	}

	sql := fmt.Sprintf(`SELECT s.id, s.schema_mode, s.schema_status, s.schema_time, s.task_id, s.ctime, s.start_date, s.end_date,
c.task_name, c.coupon_id, c.coupon_target
FROM %s c LEFT JOIN %s s ON s.task_id = c.id WHERE `,
		c.TableName(), s.TableName()) + cond

	orderBy := "ORDER BY s.task_id desc"

	limit := fmt.Sprintf("LIMIT %d, %d", offset, pagesize)

	sqlData := fmt.Sprintf("%s %s %s", sql, orderBy, limit)

	r := o.Raw(sqlData)
	_, err = r.QueryRows(&list)

	sqlCount := fmt.Sprintf(`SELECT count(s.id)
FROM %s c LEFT JOIN %s s ON s.task_id = c.id WHERE `,
		c.TableName(), s.TableName()) + cond

	r = o.Raw(sqlCount)
	r.QueryRow(&count)

	return
}

func QuerySmsList(condStr map[string]interface{}, page, pagesize int) (list []SmsSchemaInfo, count int64, err error) {
	o := orm.NewOrm()
	s := models.SchemaInfo{}
	t := models.SmsTask{}
	o.Using(s.UsingSlave())

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pagesize
	cond := "s.func_name = 'send_message' "

	if f, ok := condStr["task_name"]; ok {
		cond = fmt.Sprintf("%s%s'%s'", cond, " AND t.task_name = ", f.(string))
	}
	if f, ok := condStr["schema_time"]; ok {
		cond = fmt.Sprintf("%s%s'%%%s%%'", cond, " AND s.schema_time LIKE ", f.(string))
	}
	if f, ok := condStr["sender"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND t.sender = ", f.(int))
	}
	if f, ok := condStr["schema_mode"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND s.schema_mode = ", f.(int))
	}
	if f, ok := condStr["schema_status"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND s.schema_status = ", f.(int))
	}

	sql := fmt.Sprintf(`SELECT s.id, s.schema_mode, s.schema_status, s.schema_time, s.task_id, s.ctime, s.start_date, s.end_date, 
t.task_name, t.sender, t.sms_target
FROM %s t LEFT JOIN %s s ON s.task_id = t.id WHERE `,
		t.TableName(), s.TableName()) + cond

	orderBy := "ORDER BY s.task_id desc"

	limit := fmt.Sprintf("LIMIT %d, %d", offset, pagesize)

	sqlData := fmt.Sprintf("%s %s %s", sql, orderBy, limit)

	r := o.Raw(sqlData)
	_, err = r.QueryRows(&list)

	sqlCount := fmt.Sprintf(`SELECT count(s.id)
FROM %s t LEFT JOIN %s s ON s.task_id = t.id WHERE `,
		t.TableName(), s.TableName()) + cond

	r = o.Raw(sqlCount)
	r.QueryRow(&count)

	return

}

func QuerySmsRecord(id int64, page, pagesize int) (list []models.SmsTaskRecord, count int64, err error) {
	o := orm.NewOrm()
	c := models.SmsTaskRecord{}
	o.Using(c.UsingSlave())

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pagesize
	if page == 1 {
		pagesize--
	} else {
		offset--
	}

	sql := fmt.Sprintf(`SELECT *
FROM %s WHERE task_id = %d`,
		c.TableName(), id)

	orderBy := "ORDER BY send_date desc"

	limit := fmt.Sprintf("LIMIT %d, %d", offset, pagesize)

	sqlData := fmt.Sprintf("%s %s %s", sql, orderBy, limit)

	r := o.Raw(sqlData)
	_, err = r.QueryRows(&list)

	r = o.Raw(sqlData)
	r.QueryRow(&count)

	if page == 1 {
		storageClient := storage.RedisStorageClient.Get()
		defer storageClient.Close()

		nowStr := tools.MDateMHSDate(tools.GetUnixMillis())
		key := fmt.Sprintf("sms:%d:%s", id, nowStr)
		totalNum, _ := redis.Int(storageClient.Do("HGET", key, push.MessageKeyTotal))
		succNum, _ := redis.Int(storageClient.Do("HGET", key, push.MessageKeySucc))

		record := models.SmsTaskRecord{}
		record.TaskId = id
		record.TotalNum = totalNum
		record.SuccNum = succNum
		record.SendDate = tools.GetUnixMillis()

		newList := make([]models.SmsTaskRecord, 0)
		newList = append(newList, record)
		newList = append(newList, list...)
		list = newList

		count++
	}

	return
}
