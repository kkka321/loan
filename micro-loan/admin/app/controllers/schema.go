package controllers

import (
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"micro-loan/common/lib/gaws"

	"micro-loan/common/dao"

	"micro-loan/common/pkg/schema_task"

	"github.com/astaxie/beego/utils/pagination"
	"github.com/aws/aws-sdk-go/aws"
)

type SchemaController struct {
	BaseController
}

func (c *SchemaController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "schema"
}

func (c *SchemaController) PushList() {
	c.Layout = "layout.html"
	c.TplName = "schema/push_list.html"

	var condCntr = map[string]interface{}{}

	taskName := c.GetString("task_name")
	if taskName != "" {
		condCntr["task_name"] = taskName
	}
	c.Data["task_name"] = taskName

	messageType, _ := c.GetInt("message_type")
	if messageType > 0 {
		condCntr["message_type"] = messageType
	}
	c.Data["message_type"] = messageType

	schemaTime := c.GetString("schema_time")
	if schemaTime != "" {
		condCntr["schema_time"] = schemaTime
	}
	c.Data["schema_time"] = schemaTime

	pushWay, _ := c.GetInt("push_way")
	if pushWay > 0 {
		condCntr["push_way"] = pushWay
	}
	c.Data["push_way"] = pushWay

	schemaMode, _ := c.GetInt("schema_mode")
	if schemaMode > 0 {
		condCntr["schema_mode"] = schemaMode
	}
	c.Data["schema_mode"] = schemaMode

	schemaStatus, _ := c.GetInt("schema_status")
	if schemaStatus > 0 {
		condCntr["schema_status"] = schemaStatus
	}
	c.Data["schema_status"] = schemaStatus

	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize

	list, count, _ := service.QueryPushList(condCntr, page, pageSize)

	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))

	c.Data["paginator"] = paginator

	c.Data["List"] = list

	c.Data["MessageType"] = types.MessageTypeMap
	c.Data["PushWay"] = types.PushWayMap
	c.Data["SchemaMode"] = types.SchemaModeMap
	c.Data["SchemaStatus"] = types.SchemaStatusMap

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "schema/push_list_scripts.html"

	return
}

func (c *SchemaController) PushEdit() {
	c.Layout = "layout.html"
	c.TplName = "schema/push_edit.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "schema/push_edit_scripts.html"
	c.Data["MessageType"] = types.MessageTypeMap
	c.Data["PushWay"] = types.PushWayMap
	c.Data["SchemaMode"] = types.SchemaModeMap
	c.Data["PushTarget"] = types.PushTargetMap
	c.Data["MessageSkip"] = types.MessageSkipMap

	c.Data["message_type"] = types.MessageTypeReview
	c.Data["push_way"] = types.PushWayAccount
	c.Data["schema_mode"] = types.SchemaModeManual
	c.Data["push_target"] = types.PushTargetCustom
	c.Data["skip_to"] = types.MessageSkipToNo

	op, _ := c.GetInt("op")
	if op == 0 {
		op = 1
	}

	c.Data["op"] = op

	id, _ := c.GetInt64("id")
	if id == 0 {
		return
	}

	splitSep := " - "
	schema, err := models.GetSchemaInfo(id)
	if err != nil {
		return
	}

	task, err := models.GetPushTask(schema.TaskId)
	if err != nil {
		return
	}

	c.Data["id"] = id
	c.Data["task_name"] = task.TaskName
	c.Data["task_desc"] = task.TaskDesc
	c.Data["push_target"] = task.PushTarget
	c.Data["message_type"] = task.MessageType
	c.Data["title"] = task.Title
	c.Data["body"] = task.Body
	c.Data["skip_to"] = task.SkipTo
	c.Data["mark"] = task.Mark
	c.Data["schema_date"] = tools.MDateUTC(schema.StartDate) + splitSep + tools.MDateUTC(schema.EndDate)
	c.Data["schema_time"] = schema.SchemaTime
	c.Data["push_way"] = task.PushWay
	c.Data["schema_mode"] = schema.SchemaMode
	c.Data["version"] = task.Version
	if task.PushListPath != "" {
		var b []byte
		w := aws.NewWriteAtBuffer(b)
		gaws.AwsDownload2Stream(task.PushListPath, w)
		c.Data["target_list"] = string(w.Bytes())
	}

	return
}

func (c *SchemaController) PushSave() {
	op, _ := c.GetInt("op")
	if op == 0 {
		op = 1
	}

	splitSep := " - "
	id, _ := c.GetInt64("id")
	taskName := c.GetString("task_name")
	taskDesc := c.GetString("task_desc")
	pushTrager, _ := c.GetInt("push_target")
	messageType, _ := c.GetInt("message_type")
	title := c.GetString("title")
	body := c.GetString("body")
	skipTo, _ := c.GetInt("skip_to")
	mark := c.GetString("mark")
	version := c.GetString("version")
	schemaDateStr := c.GetString("schema_date")
	var timeStart, timeEnd int64
	if len(schemaDateStr) > 16 {
		tr := strings.Split(schemaDateStr, splitSep)
		if len(tr) == 2 {
			timeStart = tools.GetDateParse(tr[0]) * 1000
			timeEnd = tools.GetDateParse(tr[1])*1000 + 3600*24*1000 - 1000
		}
	}
	schemaTime := c.GetString("schema_time")
	pushWay, _ := c.GetInt("push_way")
	schemaMode, _ := c.GetInt("schema_mode")
	targetList := c.GetString("target_list")
	s3Key := ""

	if targetList != "" {
		fileMd5 := tools.Md5(targetList)
		_, s3Key = tools.BuildHashName(fileMd5, "push")
	}

	//add
	if op == 1 {
		task := models.PushTask{}
		task.TaskName = taskName
		task.TaskStatus = types.SchemaStatusOn
		task.TaskDesc = taskDesc
		task.MessageType = messageType
		task.PushWay = pushWay
		task.PushTarget = types.PushTarget(pushTrager)
		task.PushListPath = s3Key
		task.Title = title
		task.Body = body
		task.Mark = mark
		task.SkipTo = skipTo
		task.Version = version
		task.Ctime = tools.GetUnixMillis()
		err := task.Insert()
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/push_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		schema := models.SchemaInfo{}
		schema.SchemaMode = types.SchemaMode(schemaMode)
		schema.SchemaStatus = types.SchemaStatusOn
		schema.SchemaTime = schemaTime
		if schema.SchemaMode == types.SchemaModeAuto {
			schema.StartDate = timeStart
			schema.EndDate = timeEnd
		} else if schema.SchemaMode == types.SchemaModeManual {
			schema.StartDate = tools.NaturalDay(0)
			schema.EndDate = schema.StartDate + 3600*24*1000 - 1000
		}
		schema.FuncName = "push_message"
		schema.TaskId = task.Id
		schema.Ctime = tools.GetUnixMillis()
		err = schema.Insert()
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/push_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		r := strings.NewReader(targetList)
		gaws.AwsUploadStream(s3Key, r)

		service.PublishData(schema.TableName(), schema)

		c.Data["OpMessage"] = "增加数据成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/schema/push_list.html"
		c.TplName = "success_redirect.html"
	} else if op == 2 {
		schema, err := models.GetSchemaInfo(id)
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/push_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		task, err := models.GetPushTask(schema.TaskId)
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/push_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		if task.PushListPath != s3Key {
			gaws.AwsDelete(task.PushListPath)

			r := strings.NewReader(targetList)
			gaws.AwsUploadStream(s3Key, r)
		}

		task.TaskName = taskName
		task.TaskDesc = taskDesc
		task.MessageType = messageType
		task.PushWay = pushWay
		task.PushTarget = types.PushTarget(pushTrager)
		task.Title = title
		task.Body = body
		task.Mark = mark
		task.Version = version
		task.SkipTo = skipTo
		task.PushListPath = s3Key
		task.Utime = tools.GetUnixMillis()
		err = task.Update()
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/push_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		schema.SchemaMode = types.SchemaMode(schemaMode)
		schema.SchemaTime = schemaTime
		if schema.SchemaMode == types.SchemaModeAuto {
			schema.StartDate = timeStart
			schema.EndDate = timeEnd
		}
		schema.Utime = tools.GetUnixMillis()
		err = schema.Update()
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/push_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		service.PublishData(schema.TableName(), schema)

		c.Data["OpMessage"] = "更新数据成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/schema/push_list.html"
		c.TplName = "success_redirect.html"
	}
}

func (c *SchemaController) SchemaRun() {
	ids := c.GetStrings("ids[]")
	curTime := tools.GetUnixMillis()

	for _, id := range ids {
		i, _ := tools.Str2Int64(id)
		if i == 0 {
			continue
		}

		go service.RunSchemaManual(i, curTime)
	}

	c.ServeJSON()
}

func (c *SchemaController) PushActive() {
	id, _ := c.GetInt64("id")
	op, _ := c.GetInt("op")

	schema, err := models.GetSchemaInfo(id)

	if err != nil {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"

		c.Data["goto_url"] = "/schema/push_list.html"
		c.Data["message"] = "数据错误"
		return
	}

	task, err := models.GetPushTask(schema.TaskId)
	if err != nil {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"

		c.Data["goto_url"] = "/schema/push_list.html"
		c.Data["message"] = "数据错误"
		return
	}

	if schema.SchemaStatus == types.SchemaStatus(op) {
		c.Data["OpMessage"] = "操作成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/schema/push_list.html"
		c.TplName = "success_redirect.html"

		return
	}

	if schema.SchemaMode == types.SchemaModeManual {
		schema.StartDate = tools.NaturalDay(0)
		schema.EndDate = schema.StartDate + 3600*24*1000 - 1000
	}
	task.TaskStatus = types.SchemaStatus(op)
	task.Utime = tools.GetUnixMillis()
	task.Update()

	schema.SchemaStatus = types.SchemaStatus(op)
	schema.Utime = tools.GetUnixMillis()
	schema.Update()

	service.PublishData(schema.TableName(), schema)

	if err == nil {
		c.Data["OpMessage"] = "操作成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/schema/push_list.html"
		c.TplName = "success_redirect.html"
	} else {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"
		c.Data["goto_url"] = "/schema/push_list.html"
		c.Data["message"] = err.Error()
	}
}

func (c *SchemaController) PushDetail() {
	c.Layout = "layout.html"
	c.TplName = "schema/push_detail.html"

	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize

	task, _ := c.GetInt64("id")

	list, count, _ := service.QueryPushRecord(task, page, pageSize)

	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "schema/push_detail_scripts.html"

	return
}

func (c *SchemaController) CouponList() {
	c.Layout = "layout.html"
	c.TplName = "schema/coupon_list.html"

	var condCntr = map[string]interface{}{}

	taskName := c.GetString("task_name")
	if taskName != "" {
		condCntr["task_name"] = taskName
	}
	c.Data["task_name"] = taskName

	coupon_id, _ := c.GetInt64("coupon_id")
	if coupon_id > 0 {
		condCntr["coupon_id"] = coupon_id
	}
	c.Data["coupon_id"] = coupon_id

	schemaTime := c.GetString("schema_time")
	if schemaTime != "" {
		condCntr["schema_time"] = schemaTime
	}
	c.Data["schema_time"] = schemaTime

	schemaMode, _ := c.GetInt("schema_mode")
	if schemaMode > 0 {
		condCntr["schema_mode"] = schemaMode
	}
	c.Data["schema_mode"] = schemaMode

	schemaStatus, _ := c.GetInt("schema_status")
	if schemaStatus > 0 {
		condCntr["schema_status"] = schemaStatus
	}
	c.Data["schema_status"] = schemaStatus

	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize

	list, count, _ := service.QueryCouponList(condCntr, page, pageSize)

	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))

	c.Data["paginator"] = paginator

	c.Data["List"] = list

	c.Data["SchemaMode"] = types.SchemaModeMap
	c.Data["SchemaStatus"] = types.SchemaStatusMap

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "schema/coupon_list_scripts.html"

	return
}

func (c *SchemaController) CouponEdit() {
	c.Layout = "layout.html"
	c.TplName = "schema/coupon_edit.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "schema/coupon_edit_scripts.html"

	c.Data["CouponSchemaMode"] = types.CouponSchemaMode
	c.Data["CouponTarget"] = types.CouponTargetMap

	c.Data["schema_mode"] = types.SchemaModeManual
	c.Data["coupon_target"] = types.CouponTargetCustom

	op, _ := c.GetInt("op")
	if op == 0 {
		op = 1
	}

	c.Data["op"] = op

	id, _ := c.GetInt64("id")
	if id == 0 {
		return
	}

	splitSep := " - "
	schema, err := models.GetSchemaInfo(id)
	if err != nil {
		return
	}

	task, err := models.GetCouponTask(schema.TaskId)
	if err != nil {
		return
	}

	c.Data["id"] = id
	c.Data["task_name"] = task.TaskName
	c.Data["task_desc"] = task.TaskDesc
	c.Data["coupon_target"] = task.CouponTarget
	c.Data["coupon_id"] = task.CouponId
	c.Data["schema_date"] = tools.MDateUTC(schema.StartDate) + splitSep + tools.MDateUTC(schema.EndDate)
	c.Data["schema_time"] = schema.SchemaTime
	c.Data["schema_mode"] = schema.SchemaMode
	if task.CouponListPath != "" {
		var b []byte
		w := aws.NewWriteAtBuffer(b)
		gaws.AwsDownload2Stream(task.CouponListPath, w)
		c.Data["target_list"] = string(w.Bytes())
	}

	return
}

func (c *SchemaController) CouponSave() {
	op, _ := c.GetInt("op")
	if op == 0 {
		op = 1
	}

	splitSep := " - "
	id, _ := c.GetInt64("id")
	taskName := c.GetString("task_name")
	taskDesc := c.GetString("task_desc")
	couponTrager, _ := c.GetInt("coupon_target")
	couponId, _ := c.GetInt64("coupon_id")
	schemaDateStr := c.GetString("schema_date")
	var timeStart, timeEnd int64
	if len(schemaDateStr) > 16 {
		tr := strings.Split(schemaDateStr, splitSep)
		if len(tr) == 2 {
			timeStart = tools.GetDateParse(tr[0]) * 1000
			timeEnd = tools.GetDateParse(tr[1])*1000 + 3600*24*1000 - 1000
		}
	}
	schemaTime := c.GetString("schema_time")
	schemaMode, _ := c.GetInt("schema_mode")
	targetList := c.GetString("target_list")
	s3Key := ""

	if targetList != "" {
		fileMd5 := tools.Md5(targetList)
		_, s3Key = tools.BuildHashName(fileMd5, "push")
	}

	_, err := dao.GetCouponById(couponId)
	if err != nil {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"

		c.Data["goto_url"] = "/schema/coupon_list.html"
		c.Data["message"] = "优惠券ID错误"

		return
	}

	//add
	if op == 1 {
		task := models.CouponTask{}
		task.TaskName = taskName
		task.TaskStatus = types.SchemaStatusOn
		task.TaskDesc = taskDesc
		task.CouponId = couponId
		task.CouponTarget = types.CouponTarget(couponTrager)
		task.CouponListPath = s3Key
		task.Ctime = tools.GetUnixMillis()
		err := task.Insert()
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/coupon_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		schema := models.SchemaInfo{}
		schema.SchemaMode = types.SchemaMode(schemaMode)
		schema.SchemaStatus = types.SchemaStatusOn
		schema.SchemaTime = schemaTime
		if schema.SchemaMode == types.SchemaModeAuto {
			schema.StartDate = timeStart
			schema.EndDate = timeEnd
		} else if schema.SchemaMode == types.SchemaModeManual {
			schema.StartDate = tools.NaturalDay(0)
			schema.EndDate = schema.StartDate + 3600*24*1000 - 1000
		}
		schema.FuncName = "coupon"
		schema.TaskId = task.Id
		schema.Ctime = tools.GetUnixMillis()
		err = schema.Insert()
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/coupon_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		r := strings.NewReader(targetList)
		gaws.AwsUploadStream(s3Key, r)

		service.PublishData(schema.TableName(), schema)

		c.Data["OpMessage"] = "增加数据成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/schema/coupon_list.html"
		c.TplName = "success_redirect.html"
	} else if op == 2 {
		schema, err := models.GetSchemaInfo(id)
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/coupon_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		task, err := models.GetCouponTask(schema.TaskId)
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/coupon_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		if task.CouponListPath != s3Key {
			gaws.AwsDelete(task.CouponListPath)

			r := strings.NewReader(targetList)
			gaws.AwsUploadStream(s3Key, r)
		}

		task.TaskName = taskName
		task.TaskDesc = taskDesc
		task.CouponId = couponId
		task.CouponTarget = types.CouponTarget(couponTrager)
		task.CouponListPath = s3Key
		task.Utime = tools.GetUnixMillis()
		err = task.Update()
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/coupon_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		schema.SchemaMode = types.SchemaMode(schemaMode)
		schema.SchemaTime = schemaTime
		if schema.SchemaMode == types.SchemaModeAuto {
			schema.StartDate = timeStart
			schema.EndDate = timeEnd
		}
		schema.Utime = tools.GetUnixMillis()
		err = schema.Update()
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/coupon_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		service.PublishData(schema.TableName(), schema)

		c.Data["OpMessage"] = "更新数据成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/schema/coupon_list.html"
		c.TplName = "success_redirect.html"
	}
}

func (c *SchemaController) CouponActive() {
	id, _ := c.GetInt64("id")
	op, _ := c.GetInt("op")

	schema, err := models.GetSchemaInfo(id)

	if err != nil {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"

		c.Data["goto_url"] = "/schema/coupon_list.html"
		c.Data["message"] = "数据错误"
		return
	}

	task, err := models.GetCouponTask(schema.TaskId)
	if err != nil {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"

		c.Data["goto_url"] = "/schema/coupon_list.html"
		c.Data["message"] = "数据错误"
		return
	}

	if schema.SchemaStatus == types.SchemaStatus(op) {
		c.Data["OpMessage"] = "操作成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/schema/coupon_list.html"
		c.TplName = "success_redirect.html"

		return
	}

	if schema.SchemaMode == types.SchemaModeManual {
		schema.StartDate = tools.NaturalDay(0)
		schema.EndDate = schema.StartDate + 3600*24*1000 - 1000
	}
	task.TaskStatus = types.SchemaStatus(op)
	task.Utime = tools.GetUnixMillis()
	task.Update()

	schema.SchemaStatus = types.SchemaStatus(op)
	schema.Utime = tools.GetUnixMillis()
	schema.Update()

	service.PublishData(schema.TableName(), schema)

	if err == nil {
		c.Data["OpMessage"] = "操作成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/schema/coupon_list.html"
		c.TplName = "success_redirect.html"
	} else {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"
		c.Data["goto_url"] = "/schema/coupon_list.html"
		c.Data["message"] = err.Error()
	}
}

func (c *SchemaController) SmsList() {
	c.Layout = "layout.html"
	c.TplName = "schema/sms_list.html"

	var condCntr = map[string]interface{}{}

	taskName := c.GetString("task_name")
	if taskName != "" {
		condCntr["task_name"] = taskName
	}
	c.Data["task_name"] = taskName

	schemaTime := c.GetString("schema_time")
	if schemaTime != "" {
		condCntr["schema_time"] = schemaTime
	}
	c.Data["schema_time"] = schemaTime

	sender, _ := c.GetInt("sender")
	if sender > 0 {
		condCntr["sender"] = sender
	}
	c.Data["sender"] = sender

	schemaMode, _ := c.GetInt("schema_mode")
	if schemaMode > 0 {
		condCntr["schema_mode"] = schemaMode
	}
	c.Data["schema_mode"] = schemaMode

	schemaStatus, _ := c.GetInt("schema_status")
	if schemaStatus > 0 {
		condCntr["schema_status"] = schemaStatus
	}
	c.Data["schema_status"] = schemaStatus

	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize

	list, count, _ := service.QuerySmsList(condCntr, page, pageSize)

	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))

	c.Data["paginator"] = paginator

	c.Data["List"] = list

	c.Data["SmsSender"] = types.SmsServiceIdMap
	c.Data["SchemaMode"] = types.SchemaModeMap
	c.Data["SchemaStatus"] = types.SchemaStatusMap
	c.Data["SmsTarget"] = types.SmsTargetMap

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "schema/sms_list_scripts.html"

	return
}

func (c *SchemaController) SmsEdit() {
	c.Layout = "layout.html"
	c.TplName = "schema/sms_edit.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "schema/sms_edit_scripts.html"
	c.Data["SmsSender"] = types.SmsServiceIdMap
	c.Data["SchemaMode"] = types.SchemaModeMap
	c.Data["SmsTarget"] = types.SmsTargetMap

	c.Data["sender"] = types.Sms253ID
	c.Data["schema_mode"] = types.SchemaModeManual
	c.Data["sms_target"] = types.SmsTargetCustom

	c.Data["TplVar"] = schema_task.SmsVarComment

	op, _ := c.GetInt("op")
	if op == 0 {
		op = 1
	}

	c.Data["op"] = op

	id, _ := c.GetInt64("id")
	if id == 0 {
		return
	}

	splitSep := " - "
	schema, err := models.GetSchemaInfo(id)
	if err != nil {
		return
	}

	task, err := models.GetSmsTask(schema.TaskId)
	if err != nil {
		return
	}

	c.Data["id"] = id
	c.Data["task_name"] = task.TaskName
	c.Data["task_desc"] = task.TaskDesc
	c.Data["sms_target"] = task.SmsTarget
	c.Data["body"] = task.Body
	c.Data["schema_date"] = tools.MDateUTC(schema.StartDate) + splitSep + tools.MDateUTC(schema.EndDate)
	c.Data["schema_time"] = schema.SchemaTime
	c.Data["sender"] = task.Sender
	c.Data["schema_mode"] = schema.SchemaMode
	if task.SmsListPath != "" {
		var b []byte
		w := aws.NewWriteAtBuffer(b)
		gaws.AwsDownload2Stream(task.SmsListPath, w)
		c.Data["target_list"] = string(w.Bytes())
	}

	return
}

func (c *SchemaController) SmsSave() {
	op, _ := c.GetInt("op")
	if op == 0 {
		op = 1
	}

	splitSep := " - "
	id, _ := c.GetInt64("id")
	taskName := c.GetString("task_name")
	taskDesc := c.GetString("task_desc")
	smsTrager, _ := c.GetInt("sms_target")
	sender, _ := c.GetInt("sender")
	body := c.GetString("body")
	schemaDateStr := c.GetString("schema_date")
	var timeStart, timeEnd int64
	if len(schemaDateStr) > 16 {
		tr := strings.Split(schemaDateStr, splitSep)
		if len(tr) == 2 {
			timeStart = tools.GetDateParse(tr[0]) * 1000
			timeEnd = tools.GetDateParse(tr[1])*1000 + 3600*24*1000 - 1000
		}
	}
	schemaTime := c.GetString("schema_time")
	schemaMode, _ := c.GetInt("schema_mode")
	targetList := c.GetString("target_list")
	s3Key := ""

	if targetList != "" {
		fileMd5 := tools.Md5(targetList)
		_, s3Key = tools.BuildHashName(fileMd5, "push")
	}

	//add
	if op == 1 {
		task := models.SmsTask{}
		task.TaskName = taskName
		task.TaskStatus = types.SchemaStatusOn
		task.TaskDesc = taskDesc
		task.Sender = types.SmsServiceID(sender)
		task.SmsTarget = types.SmsTarget(smsTrager)
		task.SmsListPath = s3Key
		task.Body = body
		task.Ctime = tools.GetUnixMillis()
		err := task.Insert()
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/sms_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		schema := models.SchemaInfo{}
		schema.SchemaMode = types.SchemaMode(schemaMode)
		schema.SchemaStatus = types.SchemaStatusOn
		schema.SchemaTime = schemaTime
		if schema.SchemaMode == types.SchemaModeAuto {
			schema.StartDate = timeStart
			schema.EndDate = timeEnd
		} else if schema.SchemaMode == types.SchemaModeManual {
			schema.StartDate = tools.NaturalDay(0)
			schema.EndDate = schema.StartDate + 3600*24*1000 - 1000
		}
		schema.FuncName = "send_message"
		schema.TaskId = task.Id
		schema.Ctime = tools.GetUnixMillis()
		err = schema.Insert()
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/sms_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		r := strings.NewReader(targetList)
		gaws.AwsUploadStream(s3Key, r)

		service.PublishData(schema.TableName(), schema)

		c.Data["OpMessage"] = "增加数据成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/schema/sms_list.html"
		c.TplName = "success_redirect.html"
	} else if op == 2 {
		schema, err := models.GetSchemaInfo(id)
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/sms_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		task, err := models.GetSmsTask(schema.TaskId)
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/sms_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		if task.SmsListPath != s3Key {
			gaws.AwsDelete(task.SmsListPath)

			r := strings.NewReader(targetList)
			gaws.AwsUploadStream(s3Key, r)
		}

		task.TaskName = taskName
		task.TaskDesc = taskDesc
		task.Sender = types.SmsServiceID(sender)
		task.SmsTarget = types.SmsTarget(smsTrager)
		task.Body = body
		task.SmsListPath = s3Key
		task.Utime = tools.GetUnixMillis()
		err = task.Update()
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/sms_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		schema.SchemaMode = types.SchemaMode(schemaMode)
		schema.SchemaTime = schemaTime
		if schema.SchemaMode == types.SchemaModeAuto {
			schema.StartDate = timeStart
			schema.EndDate = timeEnd
		}
		schema.Utime = tools.GetUnixMillis()
		err = schema.Update()
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "error.tpl"

			c.Data["goto_url"] = "/schema/sms_list.html"
			c.Data["message"] = "数据错误"

			return
		}

		service.PublishData(schema.TableName(), schema)

		c.Data["OpMessage"] = "更新数据成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/schema/sms_list.html"
		c.TplName = "success_redirect.html"
	}
}

func (c *SchemaController) SmsActive() {
	id, _ := c.GetInt64("id")
	op, _ := c.GetInt("op")

	schema, err := models.GetSchemaInfo(id)

	if err != nil {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"

		c.Data["goto_url"] = "/schema/sms_list.html"
		c.Data["message"] = "数据错误"
		return
	}

	task, err := models.GetSmsTask(schema.TaskId)
	if err != nil {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"

		c.Data["goto_url"] = "/schema/sms_list.html"
		c.Data["message"] = "数据错误"
		return
	}

	if schema.SchemaStatus == types.SchemaStatus(op) {
		c.Data["OpMessage"] = "操作成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/schema/sms_list.html"
		c.TplName = "success_redirect.html"

		return
	}

	if schema.SchemaMode == types.SchemaModeManual {
		schema.StartDate = tools.NaturalDay(0)
		schema.EndDate = schema.StartDate + 3600*24*1000 - 1000
	}
	task.TaskStatus = types.SchemaStatus(op)
	task.Utime = tools.GetUnixMillis()
	task.Update()

	schema.SchemaStatus = types.SchemaStatus(op)
	schema.Utime = tools.GetUnixMillis()
	schema.Update()

	service.PublishData(schema.TableName(), schema)

	if err == nil {
		c.Data["OpMessage"] = "操作成功."
		c.Layout = "layout.html"
		c.Data["Redirect"] = "/schema/sms_list.html"
		c.TplName = "success_redirect.html"
	} else {
		c.Layout = "layout.html"
		c.TplName = "error.tpl"
		c.Data["goto_url"] = "/schema/sms_list.html"
		c.Data["message"] = err.Error()
	}
}

func (c *SchemaController) SmsDetail() {
	c.Layout = "layout.html"
	c.TplName = "schema/sms_detail.html"

	page, _ := tools.Str2Int(c.GetString("p"))
	pageSize := service.Pagesize

	task, _ := c.GetInt64("id")

	list, count, _ := service.QuerySmsRecord(task, page, pageSize)

	paginator := pagination.SetPaginator(c.Ctx, pageSize, int64(count))

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "schema/sms_detail_scripts.html"

	return
}
