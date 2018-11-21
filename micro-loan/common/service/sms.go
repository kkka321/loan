package service

import (
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/i18n"
	"micro-loan/common/lib/device"
	"micro-loan/common/models"
	"micro-loan/common/pkg/schema_task"
	"micro-loan/common/thirdparty/nxtele"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

const (
	SMS_SEND_SUCC = 1
	SMS_SEND_FAIL = 0
)

const SMS_DISBURSE_SUCCESS = "Pelanggan yang terhormat, pinjaman yang Anda ajukan pada %s, dana telah dikirim ke rekening bank Anda, silakan periksa kembali."
const SEND_MESSAGE_H5_REGISTER = "RUPIAH CEPAT：Pendaftaran Anda telah berhasil, pinjaman besar sedang tunggu Anda! http://bit.ly/2LMWtkO"

//短信发送状态

type Message struct {
	Status string `json:"status"`
}

type SmsRet struct {
	Messages []Message `json:"messages"`
}

var smsFieldMap = map[string]string{
	"Id":        "id",
	"Ctime":     "ctime",
	"RelatedID": "related_id",
}

// SendSms 实际是发送验证码, 不是发送短信接口
func SendSms(serviceType types.ServiceType, authCodeType types.AuthCodeType, phoneNumber string, ip string) bool {
	//发送短信

	code := tools.GenerateMobileCaptcha(types.SMS_CODE_LEN)
	//content := formatSmsContent(serviceType, code)

	smsVerifyCodeID, _ := device.GenerateBizId(types.SmsVerifyCodeBiz)

	//status, _ := sms.Send(serviceType, phoneNumber, content, smsVerifyCodeID)
	param := make(map[string]interface{})
	param["auth_code"] = code
	param["related_id"] = smsVerifyCodeID

	status := false
	succ := 0

	switch serviceType {
	case types.ServiceRequestLogin:
		_, succ = schema_task.SendBusinessMsg(types.SmsTargetRequestLogin, serviceType, phoneNumber, param)
	case types.ServiceRepeatedLoan:
		_, succ = schema_task.SendBusinessMsg(types.SmsTargetRepeatedLoan, serviceType, phoneNumber, param)
	default:
		_, succ = schema_task.SendBusinessMsg(types.SmsTargetAuthCode, serviceType, phoneNumber, param)
	}

	if succ > 0 {
		status = true
	}

	o := orm.NewOrm()
	smsVerifyCode := new(models.SmsVerifyCode)
	smsVerifyCode.Id = smsVerifyCodeID
	smsVerifyCode.Mobile = phoneNumber
	smsVerifyCode.Code = code
	smsVerifyCode.AuthCodeType = authCodeType
	smsVerifyCode.Expires = types.SMS_EXPIRES
	smsVerifyCode.ServiceType = serviceType

	//移除 status, 此处仅为验证码状态, 不是短信发送状态
	if status {
		smsVerifyCode.Status = types.VerifyCodeUnchecked
	} else {
		smsVerifyCode.Status = types.VerifyCodeSendFailed
	}
	smsVerifyCode.Ip = ip
	smsVerifyCode.Ctime = tools.GetUnixMillis()
	smsVerifyCode.Utime = tools.GetUnixMillis()

	o.Using(smsVerifyCode.Using())
	// err = o.Begin()
	_, err := o.Insert(smsVerifyCode)

	if err != nil {
		return false
	}
	return status
}

func formatSmsContent(serviceType types.ServiceType, smsCode string) (messge string) {
	serviceRegin := tools.GetServiceRegion()

	switch serviceRegin {
	case tools.ServiceRegionIndonesia:
		messge = formatSmsContent4Indonesia(serviceType, smsCode)
	case tools.ServiceRegionIndia:
		messge = formatSmsContent4India(serviceType, smsCode)
	default:
		logs.Error("Unsupported service regin: ", serviceRegin)
	}

	return
}

func formatSmsContent4India(serviceType types.ServiceType, smsCode string) string {
	var msg string

	switch serviceType {
	case types.ServiceRequestLogin:
		msg = fmt.Sprintf(i18n.GetMessageText(i18n.TextLoginSmsVerify), smsCode)
	case types.ServiceRepeatedLoan:
		msg = fmt.Sprintf(i18n.GetMessageText(i18n.TextLoanSmsVerify), smsCode)
	default:
		msg = ""
	}

	return msg
}

func formatSmsContent4Indonesia(serviceType types.ServiceType, smsCode string) (message string) {
	switch serviceType {
	case types.ServiceRequestLogin:
		message = fmt.Sprintf(i18n.GetMessageText(i18n.TextLoginSmsVerify), smsCode)
	case types.ServiceRepeatedLoan:
		message = fmt.Sprintf(i18n.GetMessageText(i18n.TextLoanSmsVerify), smsCode)
	default:
		message = fmt.Sprintf(i18n.GetMessageText(i18n.TextDefSmsVerify), smsCode)
	}

	return
}

// SendVoiceAuthCode 发送语音验证码
func SendVoiceAuthCode(serviceType types.ServiceType, authCodeType types.AuthCodeType, phoneNumber string, ip string) bool {

	code := tools.GenerateMobileCaptcha(types.SMS_CODE_LEN)
	// 语音验证码内容
	codeFormat := nxtele.FormatAuthCode(code)
	content := fmt.Sprintf(i18n.GetMessageText(i18n.TextDefVoiceVerify), codeFormat)
	// 语音验证码播报三次
	content = fmt.Sprintf("%s,-%s,-%s", content, content, content)

	smsVerifyCodeID, _ := device.GenerateBizId(types.SmsVerifyCodeBiz)

	var status bool
	resp, err := nxtele.SendVoice(smsVerifyCodeID, phoneNumber, content)
	if err != nil || resp.Code != "0" {
		logs.Error("[SendVoiceAuthCode] send voice auth code failed, response:", resp, "error:", err)
		return false
	} else {
		logs.Info("[SendVoiceAuthCode] send voice auth code successed, response:", resp)
		status = true
	}

	o := orm.NewOrm()
	smsVerifyCode := new(models.SmsVerifyCode)
	smsVerifyCode.Id = smsVerifyCodeID
	smsVerifyCode.Mobile = phoneNumber
	smsVerifyCode.Code = code
	smsVerifyCode.AuthCodeType = authCodeType
	smsVerifyCode.Expires = types.SMS_EXPIRES
	smsVerifyCode.ServiceType = serviceType

	if status {
		smsVerifyCode.Status = types.VerifyCodeUnchecked
	} else {
		smsVerifyCode.Status = types.VerifyCodeSendFailed
	}
	smsVerifyCode.Ip = ip
	t := tools.GetUnixMillis()
	smsVerifyCode.Ctime = t
	smsVerifyCode.Utime = t

	o.Using(smsVerifyCode.Using())
	_, err = o.Insert(smsVerifyCode)
	if err != nil {
		return false
	}
	return status
}

func CheckSmsCode(phoneNumber string, smsCode string) bool {
	o := new(models.SmsVerifyCode)
	smsVerifyCode, err := o.GetSmsCode(phoneNumber)
	curMill := tools.GetUnixMillis()
	timeDiff := curMill - smsVerifyCode.Ctime

	if err == nil && smsVerifyCode.Code == smsCode && timeDiff <= types.SMS_EXPIRES {
		o.SetStatusUsed(smsVerifyCode)
		return true
	}

	return false
}

// 根据手机号 + 服务类型, 检查短信验证码/语音验证码
// 对于同一个手机号和服务类型, 最新验证码有效
func CheckSmsCodeV2(phoneNumber string, smsCode string, serviceType types.ServiceType) bool {
	o := new(models.SmsVerifyCode)
	smsVerifyCode, err := o.GetSmsCodeByPhoneAndServiceType(phoneNumber, serviceType)
	curMill := tools.GetUnixMillis()
	timeDiff := curMill - smsVerifyCode.Ctime

	if err == nil && smsVerifyCode.Code == smsCode && timeDiff <= types.SMS_EXPIRES {
		o.SetStatusUsed(smsVerifyCode)

		accountBase, _ := models.OneAccountBaseByMobile(phoneNumber)
		accountBase.LatestSmsVerifyTime = tools.GetUnixMillis()
		accountBase.Update("latest_sms_verify_time")

		return true
	}

	return false
}

// smsVerifyCodeListCond 用于后台查询条件生成
func smsVerifyCodeListCond(condCntr map[string]interface{}) (cond *orm.Condition) {
	cond = orm.NewCondition()
	// 生成查询条件
	if value, ok := condCntr["mobile"]; ok {
		cond = cond.And("mobile", value)
	}
	if value, ok := condCntr["status"]; ok {
		cond = cond.And("status", value)
	}
	if value, ok := condCntr["authcode_type"]; ok {
		cond = cond.And("authcode_type", value)
	}
	if value, ok := condCntr["expires_start_time"]; ok {
		cond = cond.And("ctime__gte", value)
	}
	if value, ok := condCntr["expires_end_time"]; ok {
		cond = cond.And("ctime__lt", value)
	}
	if value, ok := condCntr["ctime_start_time"]; ok {
		cond = cond.And("ctime__gte", value)
	}
	if value, ok := condCntr["ctime_end_time"]; ok {
		cond = cond.And("ctime__lt", value)
	}
	if value, ok := condCntr["utime_start_time"]; ok {
		cond = cond.And("utime__gte", value)
	}
	if value, ok := condCntr["utime_end_time"]; ok {
		cond = cond.And("utime__lt", value)
	}
	if value, ok := condCntr["ip"]; ok {
		cond = cond.And("ip", value)
	}
	return
}

// SmsVerifyCodeCount 返回符合条件的记录条数，
// 一般与 SmsVerifyCodeList 配合使用， 用于生成分页列表
func SmsVerifyCodeCount(condCntr map[string]interface{}) (count int64, err error) {
	obj := models.SmsVerifyCode{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	qs := o.QueryTable(obj.TableName())
	cond := smsVerifyCodeListCond(condCntr)

	count, err = qs.SetCond(cond).Count()

	return
}

// SmsVerifyCodeList 返回符合查询条件的所有记录
// 注：当前后台需返回所有 column
func SmsVerifyCodeList(condCntr map[string]interface{}, page, pagesize int) (list []models.SmsVerifyCode, num int64, err error) {
	obj := models.SmsVerifyCode{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	qs := o.QueryTable(obj.TableName())
	cond := smsVerifyCodeListCond(condCntr)

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	num, err = qs.SetCond(cond).OrderBy("-id").Limit(pagesize).Offset(offset).All(&list)

	return
}

// VerfiyCodeServiceTypeDisplay 返回 ServiceType 名称
func VerfiyCodeServiceTypeDisplay(lang string, serviceType types.ServiceType) (out string) {
	out = "未定义"

	if value, ok := types.ServiceTypeEnumMap()[serviceType]; ok {
		out = value
	}
	return i18n.T(lang, out)
}

// AuthCodeTypeDisplay 返回 AuthCodeType 名称
func AuthCodeTypeDisplay(lang string, authCodeType types.AuthCodeType) (out string) {
	out = "-"

	if value, ok := types.AuthCodeTypeMap()[authCodeType]; ok {
		out = value
	}
	return i18n.T(lang, out)
}

// ExpireTimeCalculate 根据过期时长（expire毫秒）和初始时间戳（ctime 毫秒），计算过期时间戳
func ExpireTimeCalculate(expire int, ctime int64) int64 {
	return ctime + int64(expire)
}

func SmsStatusList(condCntr map[string]interface{}, page, pagesize int) (count int64, list []models.Sms, num int64, err error) {
	obj := models.Sms{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	sqlCount := "SELECT COUNT(id) AS total"
	sqlQuery := "SELECT *"
	from := fmt.Sprintf("FROM `%s`", obj.TableName())

	where := "WHERE 1 = 1"
	if val, ok := condCntr["sms_service"]; ok {
		where += fmt.Sprintf(" AND sms_service = %d", val.(int))
	}

	if val, ok := condCntr["sms_type"]; ok {
		where += fmt.Sprintf(" AND service_type = %d", val.(int))
	}

	if val, ok := condCntr["sms_status"]; ok {
		where += fmt.Sprintf(" AND delivery_status = %d", val.(int))
	}

	if val, ok := condCntr["related_id"]; ok {
		where += fmt.Sprintf(" AND related_id = %d", val.(int64))
	}

	if val, ok := condCntr["send_start_time"]; ok {
		where += fmt.Sprintf(" AND ctime >= %d", val.(int64))
	}

	if val, ok := condCntr["send_end_time"]; ok {
		where += fmt.Sprintf(" AND ctime < %d", val.(int64))
	}

	orderBy := ""
	if v, ok := condCntr["field"]; ok {
		if vF, okF := smsFieldMap[v.(string)]; okF {
			orderBy = "ORDER BY " + vF
		} else {
			orderBy = "ORDER BY id"
		}
	} else {
		orderBy = "ORDER BY id"
	}

	if v, ok := condCntr["sort"]; ok {
		orderBy = fmt.Sprintf("%s %s", orderBy, v.(string))
	} else {
		orderBy = fmt.Sprintf("%s %s", orderBy, "DESC")
	}

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize
	limit := fmt.Sprintf("LIMIT %d, %d", offset, pagesize)

	sql := ""
	sql = fmt.Sprintf("%s %s %s", sqlCount, from, where)
	o.Raw(sql).QueryRow(&count)

	sql = fmt.Sprintf("%s %s %s %s %s", sqlQuery, from, where, orderBy, limit)
	num, err = o.Raw(sql).QueryRows(&list)

	return
}
