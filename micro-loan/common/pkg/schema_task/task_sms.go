package schema_task

import (
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gomodule/redigo/redis"

	"micro-loan/common/dao"
	"micro-loan/common/lib/gaws"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/lib/sms"
	"micro-loan/common/models"
	"micro-loan/common/pkg/google/push"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/thirdparty/doku"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

const (
	SmsVarSign  = '#'
	SmsFuncSign = `_`
)

const (
	SmsFuncNow         = "Now"
	SmsFuncNowDate     = "Date"
	SmsFuncNowTime     = "Time"
	SmsFuncNowDateTime = "DateTime"

	SmsFuncAuth     = "Auth"
	SmsFuncAuthCode = "Code"

	SmsFuncRepay       = "Repay"
	SmsFuncRepayAmount = "Amount"
	SmsFuncRepayDate   = "Date"
	SmsFuncRepayMD     = "MD"

	SmsFuncOrder              = "Order"
	SmsFuncOrderMinAmount     = "MinRepay"
	SmsFuncOrderApplyDateTime = "ApplyDateTime"

	SmsFuncEAccount       = "EAccount"
	SmsFuncEAccountBank   = "Bank"
	SmsFuncEAccountNumber = "Number"

	SmsFuncXendit        = "Xendit"
	SmsFuncXenditPaycode = "PayCode"
	SmsFuncXenditExpire  = "Expire"
	SmsFuncXenditAmount  = "Amount"

	SmsFuncInvite    = "Invite"
	SmsFuncInviteUrl = "Url"
)

var SmsVarHandler = map[string]func(subFunc map[string]bool, mobile string, param map[string]interface{}) map[string]string{
	SmsFuncNow:      smsFuncNow,
	SmsFuncAuth:     smsFuncAuth,
	SmsFuncRepay:    smsFuncRepay,
	SmsFuncOrder:    smsFuncOrder,
	SmsFuncEAccount: smsFuncEAccount,
	SmsFuncXendit:   smsFuncXendit,
	SmsFuncInvite:   smsFuncInvite,
}

var SmsVarComment = map[string]string{
	makeTplField(SmsFuncNow, SmsFuncNowDate):     "当前日期",
	makeTplField(SmsFuncNow, SmsFuncNowTime):     "当前时间",
	makeTplField(SmsFuncNow, SmsFuncNowDateTime): "当前时间日期",

	makeTplField(SmsFuncAuth, SmsFuncAuthCode): "验证码",

	makeTplField(SmsFuncRepay, SmsFuncRepayAmount): "还款金额",
	makeTplField(SmsFuncRepay, SmsFuncRepayDate):   "还款日期",
	makeTplField(SmsFuncRepay, SmsFuncRepayMD):     "还款月日",

	makeTplField(SmsFuncOrder, SmsFuncOrderMinAmount):     "展期最低还款金额",
	makeTplField(SmsFuncOrder, SmsFuncOrderApplyDateTime): "订单申请日期",

	makeTplField(SmsFuncEAccount, SmsFuncEAccountBank):   "va银行",
	makeTplField(SmsFuncEAccount, SmsFuncEAccountNumber): "va账号",

	makeTplField(SmsFuncXendit, SmsFuncXenditPaycode): "付款码",
	makeTplField(SmsFuncXendit, SmsFuncXenditExpire):  "付款码有效期",
	makeTplField(SmsFuncXendit, SmsFuncXenditAmount):  "付款码还款金额",

	makeTplField(SmsFuncInvite, SmsFuncInviteUrl): "邀请信息",
}

func makeTplField(c, f string) string {
	if c == "" {
		return f
	} else {
		return c + string(SmsFuncSign) + f
	}
}

func isMatchVar(b byte) bool {
	if (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		b == '_' {
		return true
	}

	return false
}

func displayVAInfoV2(accountId int64) (bankCode, eAccountDesc string) {
	eAccount, err := dao.GetActiveUserEAccount(accountId)
	if err == nil {
		bankCode = eAccount.BankCode
		if eAccount.VaCompanyCode == types.DoKu {
			bankCode = doku.DoKuVaBankCodeTransform(eAccount.BankCode)
		}
		eAccountDesc = eAccount.EAccountNumber
	}

	return
}

func smsFuncNow(subFunc map[string]bool, mobile string, param map[string]interface{}) map[string]string {
	now := tools.GetUnixMillis()
	ret := make(map[string]string)

	for k, _ := range subFunc {
		str := ""
		if k == SmsFuncNowDate {
			str = tools.MDateMHSDate(now)
		} else if k == SmsFuncNowTime {
			str = tools.MDateMHSHMS(now)
		} else if k == SmsFuncNowDateTime {
			str = tools.MDateMHS(now)
		}
		ret[k] = str
	}

	return ret
}

func smsFuncAuth(subFunc map[string]bool, mobile string, param map[string]interface{}) map[string]string {
	ret := make(map[string]string)

	for k, _ := range subFunc {
		str := ""
		if k == SmsFuncAuthCode {
			if v, ok := param["auth_code"]; ok {
				str, _ = v.(string)
			}
		}
		ret[k] = str
	}

	return ret
}

func smsFuncInvite(subFunc map[string]bool, mobile string, param map[string]interface{}) map[string]string {
	ret := make(map[string]string)

	for k, _ := range subFunc {
		str := ""
		if k == SmsFuncInviteUrl {
			if v, ok := param["url"]; ok {
				str, _ = v.(string)
			}
		}
		ret[k] = str
	}

	return ret
}

func smsFuncRepay(subFunc map[string]bool, mobile string, param map[string]interface{}) map[string]string {
	ret := make(map[string]string)

	for k, _ := range subFunc {
		ret[k] = ""
	}

	account, err := models.OneAccountBaseByMobile(mobile)
	if err != nil {
		return ret
	}

	order, err := dao.AccountLastLoanOrder(account.Id)
	if err != nil {
		return ret
	}

	rp, err := models.GetLastRepayPlanByOrderid(order.Id)
	if err != nil {
		return ret
	}

	for k, _ := range subFunc {
		str := ""
		if k == SmsFuncRepayAmount {
			repayMoney, _ := repayplan.CaculateRepayTotalAmountByRepayPlan(rp)
			str = tools.Int642Str(repayMoney)
		} else if k == SmsFuncRepayMD {
			str = tools.GetLocalDateFormat(rp.RepayDate, "02/01")
		} else if k == SmsFuncRepayDate {
			str = tools.MDateMHSDate(rp.RepayDate)
		}
		ret[k] = str
	}

	return ret
}

func smsFuncOrder(subFunc map[string]bool, mobile string, param map[string]interface{}) map[string]string {
	ret := make(map[string]string)

	for k, _ := range subFunc {
		ret[k] = ""
	}

	account, err := models.OneAccountBaseByMobile(mobile)
	if err != nil {
		return ret
	}

	order, err := dao.AccountLastLoanOrder(account.Id)
	if err != nil {
		return ret
	}

	for k, _ := range subFunc {
		str := ""
		if k == SmsFuncOrderMinAmount {
			str = tools.Int642Str(order.MinRepayAmount)
		} else if k == SmsFuncOrderApplyDateTime {
			str = tools.MDateMHS(order.ApplyTime)
		}
		ret[k] = str
	}

	return ret
}

func smsFuncEAccount(subFunc map[string]bool, mobile string, param map[string]interface{}) map[string]string {
	ret := make(map[string]string)

	for k, _ := range subFunc {
		ret[k] = ""
	}

	account, err := models.OneAccountBaseByMobile(mobile)
	if err != nil {
		return ret
	}

	bankCode, num := displayVAInfoV2(account.Id)

	for k, _ := range subFunc {
		str := ""
		if k == SmsFuncEAccountBank {
			str = bankCode
		} else if k == SmsFuncEAccountNumber {
			str = num
		}
		ret[k] = str
	}

	return ret
}

func smsFuncXendit(subFunc map[string]bool, mobile string, param map[string]interface{}) map[string]string {
	ret := make(map[string]string)

	for k, _ := range subFunc {
		ret[k] = ""
	}

	account, err := models.OneAccountBaseByMobile(mobile)
	if err != nil {
		return ret
	}

	order, err := dao.AccountLastLoanOrder(account.Id)
	if err != nil {
		return ret
	}

	paymentCode, err := models.OneFixPaymentCodeByUserAccountId(order.UserAccountId)
	if err != nil {
		return ret
	}

	for k, _ := range subFunc {
		str := ""
		if k == SmsFuncXenditPaycode {
			str = paymentCode.PaymentCode
		} else if k == SmsFuncXenditExpire {
			str = tools.MHSHMS(paymentCode.ExpirationDate)
		} else if k == SmsFuncXenditAmount {
			str = tools.Int642Str(paymentCode.ExpectedAmount)
		}
		ret[k] = str
	}

	return ret
}

func parseToken(token string, vars map[string]map[string]bool) {
	c := ""
	f := ""
	vec := strings.Split(token, string(SmsFuncSign))

	if len(vec) == 1 {
		f = vec[0]
	} else {
		c = vec[0]
		f = vec[1]
	}

	if _, ok := vars[c]; !ok {
		vars[c] = make(map[string]bool)
	}
	vars[c][f] = true
}

func parseMsgTemplate(body string, mobile string, param map[string]interface{}) string {
	size := len(body)
	isIn := false
	token := ""
	vars := make(map[string]map[string]bool)

	ret := body

	for i := 0; i < size; i++ {
		if isMatchVar(body[i]) {
			if isIn {
				token = token + string(body[i])
			}
		} else {
			if isIn {
				isIn = false

				parseToken(token, vars)

				token = ""
			}

			if body[i] == SmsVarSign {
				isIn = true
			}
		}
	}

	if isIn {
		parseToken(token, vars)
	}

	for k, subf := range vars {
		fields := make(map[string]string)
		f, ok := SmsVarHandler[k]
		if ok {
			fields = f(subf, mobile, param)
		}

		for f, v := range fields {
			token := ""
			if k == "" {
				token = string(SmsVarSign) + f
			} else {
				token = string(SmsVarSign) + k + string(SmsFuncSign) + f
			}
			ret = strings.Replace(ret, token, v, -1)
		}
	}

	return ret
}

func sendCustomMsg(task *models.SmsTask) (int, int) {
	total := 0
	succ := 0

	param := make(map[string]interface{})

	var b []byte
	w := aws.NewWriteAtBuffer(b)
	gaws.AwsDownload2Stream(task.SmsListPath, w)
	list := tools.ParseTargetList(string(w.Bytes()))
	for _, v := range list {
		msg := parseMsgTemplate(task.Body, v, param)
		status, err := sms.SendByKey(task.Sender, types.ServiceOthers, v, msg, task.Id)
		total += 1
		if status && err == nil {
			succ += 1
		}
	}

	return total, succ
}

func sendBusinessMsg(task *models.SmsTask, serviceType types.ServiceType, mobile string, param map[string]interface{}) (int, int) {
	total := 0
	succ := 0

	related_id := task.Id
	if v, ok := param["related_id"]; ok {
		if id, subok := v.(int64); subok {
			related_id = id
		}
	}
	msg := parseMsgTemplate(task.Body, mobile, param)

	status, err := sms.SendByKey(task.Sender, serviceType, mobile, msg, related_id)
	total += 1
	if status && err == nil {
		succ += 1
	}

	return total, succ
}

func StartSmsBackup() {
	lockKey := beego.AppConfig.String("sms_backup_lock")

	for {
		storageClient := storage.RedisStorageClient.Get()
		lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", 10*60, "NX")

		if err != nil || lock == nil {
			storageClient.Close()
			time.After(time.Hour)
			continue
		}

		backupHistorySmsData()

		storageClient.Do("DEL", lockKey)

		storageClient.Close()

		time.Sleep(time.Second)
	}
}

func backupHistorySmsData() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	setKey := beego.AppConfig.String("sms_set") + tools.MDateMHSDate(tools.GetUnixMillis()-tools.MILLSSECONDADAY)

	num, _ := redis.Int(storageClient.Do("EXISTS", setKey))
	if num == 0 {
		return
	}

	keyList, _ := redis.Strings(storageClient.Do("SMEMBERS", setKey))

	count := 0
	for _, v := range keyList {
		list := strings.Split(v, ":")
		if len(list) < 3 {
			continue
		}

		id, _ := tools.Str2Int64(list[1])
		if id == 0 {
			continue
		}

		pushDate := tools.GetDateParseBackend(list[2]) * 1000

		totalNum, _ := redis.Int(storageClient.Do("HGET", v, push.MessageKeyTotal))
		succNum, _ := redis.Int(storageClient.Do("HGET", v, push.MessageKeySucc))

		record := models.SmsTaskRecord{}
		record.TaskId = id
		record.TotalNum = totalNum
		record.SuccNum = succNum
		record.SendDate = pushDate
		record.Ctime = tools.GetUnixMillis()
		record.Insert()

		storageClient.Do("DEL", v)

		count++
	}

	logs.Info("[backupHistorySmsData] backup history data success key:%s, count:%d", setKey, count)

	storageClient.Do("DEL", setKey)
}

func sendRemindOrder2Msg(task *models.SmsTask, param map[string]interface{}) (int, int) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	total := 0
	succ := 0

	setsName := beego.AppConfig.String("collection_remind_sets_2")
	todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))
	yesterdaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(-1)))

	num, _ := storageClient.Do("EXISTS", yesterdaySetName)
	if num != nil && num.(int64) == 1 {
		storageClient.Do("DEL", yesterdaySetName)
	}

	qVal, err := storageClient.Do("EXISTS", todaySetName)
	// 初始化去重集合
	if err == nil && 0 == qVal.(int64) {
		storageClient.Do("SADD", todaySetName, 1)
	}

	var idsBox []string
	setsMem, err := redis.Values(storageClient.Do("SMEMBERS", todaySetName))
	if err != nil || setsMem == nil {
		logs.Warn("[sendRemindOrder2Msg] set is not empty err:%v, set:%v", err, setsMem)
		return total, succ
	}

	for _, m := range setsMem {
		idsBox = append(idsBox, string(m.([]byte)))
	}
	// 理论上不会出现
	if len(idsBox) == 0 {
		logs.Warn("[sendRemindOrder2Msg] idsbox is empty err:%v, set:%v", err, setsMem)
		return total, succ
	}

	// 获取订单列表
	collectionRemindDayInt := make([]types.CollectionRemindDay, 0)
	collectionRemindDayInt = append(collectionRemindDayInt, types.CollectionRemindTwo)
	orderList, _ := dao.GetCollectionRemindOrderList(idsBox, collectionRemindDayInt)

	if len(orderList) == 0 {
		logs.Info("[sendRemindOrder2Msg] orderList is empty")
		return total, succ
	}

	for _, v := range orderList {
		orderID := v
		qVal, err := storageClient.Do("SADD", todaySetName, orderID)
		// 说明有错,或已经处理过,忽略本次操作
		if err != nil || 0 == qVal.(int64) {
			logs.Warning("[sendRemindOrder2Msg] 此订单已经处理过, 忽略之. orderID:%d, err:%v", orderID, err)
			continue
		}

		order, orderErr := models.GetOrder(orderID)
		accountBase, abErr := dao.CustomerOne(order.UserAccountId)
		if orderErr != nil || abErr != nil {
			logs.Warning("[sendRemindOrder2Msg] get order or account error orderId:%d, ordererr:%v, aberr:%v", orderID, orderErr, abErr)
			continue
		}

		// 催收短信发送逻辑
		//sms.Send(types.ServiceCollectionRemind, accountBase.Mobile, smsContent, orderID)

		msg := parseMsgTemplate(task.Body, accountBase.Mobile, param)
		status, err := sms.SendByKey(task.Sender, types.ServiceCollectionRemind, accountBase.Mobile, msg, orderID)
		total += 1
		if status && err == nil {
			succ += 1
		}
	}

	logs.Info("[sendRemindOrder2Msg] done total:%d, succ:%d", total, succ)

	return total, succ
}

func sendRemindOrder4Msg(task *models.SmsTask, param map[string]interface{}) (int, int) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	total := 0
	succ := 0

	setsName := beego.AppConfig.String("collection_remind_sets_4")
	todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))
	yesterdaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(-1)))

	num, _ := storageClient.Do("EXISTS", yesterdaySetName)
	if num != nil && num.(int64) == 1 {
		storageClient.Do("DEL", yesterdaySetName)
	}

	qVal, err := storageClient.Do("EXISTS", todaySetName)
	// 初始化去重集合
	if err == nil && 0 == qVal.(int64) {
		storageClient.Do("SADD", todaySetName, 1)
	}

	var idsBox []string
	setsMem, err := redis.Values(storageClient.Do("SMEMBERS", todaySetName))
	if err != nil || setsMem == nil {
		logs.Warn("[sendRemindOrder4Msg] set is not empty err:%v, set:%v", err, setsMem)
		return total, succ
	}

	for _, m := range setsMem {
		idsBox = append(idsBox, string(m.([]byte)))
	}
	// 理论上不会出现
	if len(idsBox) == 0 {
		logs.Warn("[sendRemindOrder4Msg] idsbox is empty err:%v, set:%v", err, setsMem)
		return total, succ
	}

	// 获取订单列表
	collectionRemindDayInt := make([]types.CollectionRemindDay, 0)
	collectionRemindDayInt = append(collectionRemindDayInt, types.CollectionRemindFour)
	orderList, _ := dao.GetCollectionRemindOrderList(idsBox, collectionRemindDayInt)

	if len(orderList) == 0 {
		logs.Info("[sendRemindOrder4Msg] orderList is empty")
		return total, succ
	}

	for _, v := range orderList {
		orderID := v
		qVal, err := storageClient.Do("SADD", todaySetName, orderID)
		// 说明有错,或已经处理过,忽略本次操作
		if err != nil || 0 == qVal.(int64) {
			logs.Warning("[sendRemindOrder4Msg] 此订单已经处理过, 忽略之. orderID:%d, err:%v", orderID, err)
			continue
		}

		order, orderErr := models.GetOrder(orderID)
		accountBase, abErr := dao.CustomerOne(order.UserAccountId)
		if orderErr != nil || abErr != nil {
			logs.Warning("[sendRemindOrder4Msg] get order or account error orderId:%d, ordererr:%v, aberr:%v", orderID, orderErr, abErr)
			continue
		}

		// 催收短信发送逻辑
		//sms.Send(types.ServiceCollectionRemind, accountBase.Mobile, smsContent, orderID)

		msg := parseMsgTemplate(task.Body, accountBase.Mobile, param)
		status, err := sms.SendByKey(task.Sender, types.ServiceCollectionRemind, accountBase.Mobile, msg, orderID)
		total += 1
		if status && err == nil {
			succ += 1
		}
	}

	logs.Info("[sendRemindOrder4Msg] done total:%d, succ:%d", total, succ)

	return total, succ
}

func sendRemindOrder8Msg(task *models.SmsTask, param map[string]interface{}) (int, int) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	total := 0
	succ := 0

	setsName := beego.AppConfig.String("collection_remind_sets_8")
	todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))
	yesterdaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(-1)))

	num, _ := storageClient.Do("EXISTS", yesterdaySetName)
	if num != nil && num.(int64) == 1 {
		storageClient.Do("DEL", yesterdaySetName)
	}

	qVal, err := storageClient.Do("EXISTS", todaySetName)
	// 初始化去重集合
	if err == nil && 0 == qVal.(int64) {
		storageClient.Do("SADD", todaySetName, 1)
	}

	var idsBox []string
	setsMem, err := redis.Values(storageClient.Do("SMEMBERS", todaySetName))
	if err != nil || setsMem == nil {
		logs.Warn("[sendRemindOrder8Msg] set is not empty err:%v, set:%v", err, setsMem)
		return total, succ
	}

	for _, m := range setsMem {
		idsBox = append(idsBox, string(m.([]byte)))
	}
	// 理论上不会出现
	if len(idsBox) == 0 {
		logs.Warn("[sendRemindOrder8Msg] idsbox is empty err:%v, set:%v", err, setsMem)
		return total, succ
	}

	// 获取订单列表
	collectionRemindDayInt := make([]types.CollectionRemindDay, 0)
	collectionRemindDayInt = append(collectionRemindDayInt, types.CollectionRemindEight)
	orderList, _ := dao.GetCollectionRemindOrderList(idsBox, collectionRemindDayInt)

	if len(orderList) == 0 {
		logs.Info("[sendRemindOrder8Msg] orderList is empty")
		return total, succ
	}

	for _, v := range orderList {
		orderID := v
		qVal, err := storageClient.Do("SADD", todaySetName, orderID)
		// 说明有错,或已经处理过,忽略本次操作
		if err != nil || 0 == qVal.(int64) {
			logs.Warning("[sendRemindOrder8Msg] 此订单已经处理过, 忽略之. orderID:%d, err:%v", orderID, err)
			continue
		}

		order, orderErr := models.GetOrder(orderID)
		accountBase, abErr := dao.CustomerOne(order.UserAccountId)
		if orderErr != nil || abErr != nil {
			logs.Warning("[sendRemindOrder8Msg] get order or account error orderId:%d, ordererr:%v, aberr:%v", orderID, orderErr, abErr)
			continue
		}

		// 催收短信发送逻辑
		//sms.Send(types.ServiceCollectionRemind, accountBase.Mobile, smsContent, orderID)

		msg := parseMsgTemplate(task.Body, accountBase.Mobile, param)
		status, err := sms.SendByKey(task.Sender, types.ServiceCollectionRemind, accountBase.Mobile, msg, orderID)
		total += 1
		if status && err == nil {
			succ += 1
		}
	}

	logs.Info("[sendRemindOrder8Msg] done total:%d, succ:%d", total, succ)

	return total, succ
}

func sendRepayRemindMsg(task *models.SmsTask, param map[string]interface{}) (int, int) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	total := 0
	succ := 0

	setsName := beego.AppConfig.String("repay_remind_sets")
	todaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(0)))
	yesterdaySetName := fmt.Sprintf("%s:%s", setsName, tools.MDateMHSLocalDate(tools.NaturalDay(-1)))

	num, _ := storageClient.Do("EXISTS", yesterdaySetName)
	if num != nil && num.(int64) == 1 {
		//如果存在就干掉
		storageClient.Do("DEL", yesterdaySetName)
	}

	qVal, err := storageClient.Do("EXISTS", todaySetName)
	// 初始化去重集合
	if err == nil && 0 == qVal.(int64) {
		storageClient.Do("SADD", todaySetName, 1)
	}

	var idsBox []string
	setsMem, err := redis.Values(storageClient.Do("SMEMBERS", todaySetName))
	if err != nil || setsMem == nil {
		logs.Warn("[sendRepayRemindMsg] set is not empty err:%v, set:%v", err, setsMem)
		return total, succ
	}

	for _, m := range setsMem {
		idsBox = append(idsBox, string(m.([]byte)))
	}
	// 理论上不会出现
	if len(idsBox) == 0 {
		logs.Warn("[sendRepayRemindMsg] idsbox is empty err:%v, set:%v", err, setsMem)
		return total, succ
	}

	orderList, _ := dao.GetRepayRemindOrderList(idsBox)
	if len(orderList) == 0 {
		logs.Info("[sendRepayRemindMsg] orderList is empty")
		return total, succ
	}

	for _, v := range orderList {
		orderID := v

		qVal, err := storageClient.Do("SADD", todaySetName, orderID)
		// 说明有错,或已经处理过,忽略本次操作
		if err != nil || 0 == qVal.(int64) {
			logs.Warning("[sendRepayRemindMsg] 此订单已经处理过, 忽略之. orderID:%d, err:%v", orderID, err)
			continue
		}

		order, orderErr := models.GetOrder(orderID)
		accountBase, abErr := dao.CustomerOne(order.UserAccountId)
		if orderErr != nil || abErr != nil {
			logs.Warning("[sendRepayRemindMsg] get order or account error orderId:%d, ordererr:%v, aberr:%v", orderID, orderErr, abErr)
			continue
		}

		//smsContent := fmt.Sprintf(i18n.GetMessageText(i18n.TextRepayRemind), date, repayMoney, vaAccountNumber)
		//sms.Send(types.ServiceRepayRemind, accountBase.Mobile, smsContent, orderID)

		msg := parseMsgTemplate(task.Body, accountBase.Mobile, param)
		status, err := sms.SendByKey(task.Sender, types.ServiceRepayRemind, accountBase.Mobile, msg, orderID)
		total += 1
		if status && err == nil {
			succ += 1
		}
	}

	logs.Info("[sendRepayRemindMsg] done total:%d, succ:%d", total, succ)

	return total, succ
}

func SendBusinessMsg(target types.SmsTarget, serviceType types.ServiceType, mobile string, param map[string]interface{}) (error, int) {
	logs.Debug("[SendBusinessMsg] start target:%d, mobile:%v", target, mobile)

	succ := 0

	list, _ := models.GetSmsTaskByTarget(target)
	for _, v := range list {
		_, s := runSmsTask(&v, serviceType, mobile, param)
		succ += s
	}

	return nil, succ
}

func runSmsTask(task *models.SmsTask, serviceType types.ServiceType, mobile string, param map[string]interface{}) (error, int) {
	total := 0
	succ := 0

	if task.SmsTarget == types.SmsTargetCustom {
		total, succ = sendCustomMsg(task)
	} else if task.SmsTarget == types.SmsTargetRemindOrder2 {
		total, succ = sendRemindOrder2Msg(task, param)
	} else if task.SmsTarget == types.SmsTargetRemindOrder4 {
		total, succ = sendRemindOrder4Msg(task, param)
	} else if task.SmsTarget == types.SmsTargetRemindOrder8 {
		total, succ = sendRemindOrder8Msg(task, param)
	} else if task.SmsTarget == types.SmsTargetRepayRemind {
		total, succ = sendRepayRemindMsg(task, param)
	} else {
		total, succ = sendBusinessMsg(task, serviceType, mobile, param)
	}

	IncrSmsCount(task.Id, total, succ)

	logs.Info("[runSmsTask] push msg taskId:%d, target:%d, param:%v, total:%d, succ:%d", task.Id, task.SmsTarget, param, total, succ)

	return nil, succ
}

func IncrSmsCount(id int64, total int, succ int) {
	nowStr := tools.MDateMHSDate(tools.GetUnixMillis())
	key := fmt.Sprintf("sms:%d:%s", id, nowStr)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	exist, _ := redis.Int(storageClient.Do("HSETNX", key, push.MessageKeyTotal, 0))
	if exist > 0 {
		setKey := beego.AppConfig.String("sms_set") + nowStr
		storageClient.Do("SADD", setKey, key)
	}

	storageClient.Do("HINCRBY", key, push.MessageKeyTotal, total)
	storageClient.Do("HINCRBY", key, push.MessageKeySucc, succ)
}

func SmsMessage(id int64) error {
	smsInfo, err := models.GetSmsTask(id)
	if err != nil {
		logs.Error("[SendMessage] GetSmsTask return error taskId:%d, err:%v", id, err)
		return err
	}

	param := make(map[string]interface{})
	err, _ = runSmsTask(&smsInfo, types.ServiceOthers, "", param)

	return err
}
