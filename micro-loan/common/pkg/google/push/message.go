package push

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/accesstoken"
	"micro-loan/common/thirdparty/fcmmsg"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

var messageMark = map[string]int{
	`"Akun Saya"`:                types.MessageSkipToAccount,
	`"CARA PEMBAYARAN"`:          types.MessageSkipToRepay,
	`"Umpan Balik"`:              types.MessageSkipToFeedback,
	`"kupon"`:                    types.MessageSkipToCoupon,
	`Segera daftar.`:             types.MessageSkipToApply,
	`Silakan ajukan sekali lagi`: types.MessageSkipToApply,
	`Menaikkan Skor Kredit`:      types.MessageSkipToNoCreditZRe,
}

const (
	MessageKeyTotal string = "total"
	MessageKeySucc  string = "succ"
	MessageKeyRead  string = "read"
)

func CheckMessageNewRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

func CheckMessageAllRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"type":   true,
		"offset": true,
	}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

func CheckMessageConfirmRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"id": true,
	}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

func AccountNewMessage(accountId int64) (list []models.FcmMessage, num int64, err error) {
	obj := models.FcmMessage{}

	o := orm.NewOrm()
	o.Using(obj.Using())

	qb, _ := orm.NewQueryBuilder(tools.DBDriver())
	qb.Select("*").
		From(obj.TableName()).
		Where("account_id = ?")

	// 导出 SQL 语句
	sql := qb.String()

	var alllist []models.FcmMessage
	size, err := o.Raw(sql, accountId).QueryRows(&alllist)

	sortMap := make(map[int]models.FcmMessage)
	for i := int64(0); i < size; i++ {
		old, ok := sortMap[alllist[i].MessageType]
		if !ok {
			sortMap[alllist[i].MessageType] = alllist[i]
		} else {
			if alllist[i].Id > old.Id {
				sortMap[alllist[i].MessageType] = alllist[i]
			}
		}
	}

	num = 0
	for _, v := range sortMap {
		list = append(list, v)
		num++
	}

	return
}

func AccountAllMessage(accountId, offset int64, msgtype int) (list []models.FcmMessage, num int64, err error) {
	obj := models.FcmMessage{}

	o := orm.NewOrm()
	o.Using(obj.Using())

	where := ""
	if offset > 0 {
		where = fmt.Sprintf("AND id < %d", offset)
	}

	qb, _ := orm.NewQueryBuilder(tools.DBDriver())
	qb.Select("*").
		From(obj.TableName()).
		Where("account_id = ? AND message_type = ?" + where).
		OrderBy("id").Desc().
		Limit(250)

	// 导出 SQL 语句
	sql := qb.String()

	num, err = o.Raw(sql, accountId, msgtype).QueryRows(&list)

	return
}

func AccountNewMessageSize(accountId int64) (num int64, err error) {
	obj := models.FcmMessage{}
	var list []models.FcmMessage

	o := orm.NewOrm()
	o.Using(obj.Using())
	num, err = o.QueryTable(obj.TableName()).Filter("account_id", accountId).Filter("is_read", types.MessageUnread).All(&list)

	return
}

func AccountConfirmMessage(ids string) (err error) {
	nowStr := tools.MDateMHSDate(tools.GetUnixMillis())

	idList := strings.Split(ids, ",")
	for _, v := range idList {
		id, _ := tools.Str2Int64(v)
		msg := models.FcmMessage{}
		err := msg.Get(id)
		if err != nil {
			continue
		}

		msg.IsRead = types.MessageRead
		msg.Update()

		cStr := tools.MDateMHSDate(msg.Ctime)
		if cStr == nowStr {
			if msg.TaskId != 0 {
				IncrReadCount(msg.TaskId, 1)
			}
		}
	}

	return
}

func IncrReadCount(taskId int64, num int) {
	nowStr := tools.MDateMHSDate(tools.GetUnixMillis())
	key := fmt.Sprintf("push:%d:%s", taskId, nowStr)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	storageClient.Do("HINCRBY", key, MessageKeyRead, num)
}

func BuildEmptyMessageData(data map[string]interface{}, withOffset bool) {
	data["size"] = 0
	var d = make([]string, 0)
	data["messages"] = d
	if withOffset {
		data["offset"] = "0"
	}
}

func BuildMessageData(data map[string]interface{}, list []models.FcmMessage, withOffset bool) {
	size := len(list)
	data["size"] = size
	if withOffset {
		data["offset"] = tools.Int642Str(list[size-1].Id)
	}

	var dataSet [](map[string]interface{})
	for _, msg := range list {
		subSet := map[string]interface{}{
			"id":       msg.Id,
			"type":     msg.MessageType,
			"title":    msg.Title,
			"body":     msg.Body,
			"is_ready": msg.IsRead,
			"mark":     msg.Mark,
			"skip_to":  msg.SkipTo,
			"ctime":    msg.Ctime,
		}

		dataSet = append(dataSet, subSet)
	}

	data["messages"] = dataSet
}

func SendFmsMessageV2(taskId int64, accountId int64, title, body string, msgType int, mark string, skipTo int, version string) (total int, succ int) {
	if version != "" {
		info, _ := models.OneLastClientInfoByRelatedID(accountId)
		if !isMatchVersion(info.AppVersionCode, version) {
			total = 0
			succ = 0

			logs.Info("[SendFmsMessageV2] skip not match version accountId:%s, uuid_ver:%d, version:%s", accountId, info.AppVersionCode, version)
			return
		}
	}

	num, tokens, _ := models.AccountValidToken(accountId)

	validTokens := make([]string, 0)
	for i := int64(0); i < num; i++ {
		if isValid, _ := accesstoken.IsValidAccessToken(types.PlatformAndroid, tokens[i].AccessToken); isValid {
			if tokens[i].FcmToken == "" {
				continue
			}
			validTokens = append(validTokens, tokens[i].FcmToken)
		}
	}

	msg := models.FcmMessage{}
	msg.Ctime = tools.GetUnixMillis()
	msg.IsRead = types.MessageUnread
	msg.Body = body
	msg.Title = title
	msg.MessageType = msgType
	msg.AccountId = accountId
	msg.Mark = mark
	msg.SkipTo = skipTo
	msg.TaskId = taskId
	msg.Add()

	total = 1
	succ = 0

	if len(validTokens) == 0 {
		logs.Warn("[SendFmsMessageV2] fcm token empty accountId:%d", accountId)

		return
	}

	n, err := fcmmsg.SendMessage(validTokens, title, body, types.SkipToMessageCenter)
	if err != nil {
		logs.Error("[SendFmsMessageV2] SendMessage error accountId:%d, err:%v", accountId, err)
	}

	succ = n
	return
}

func SendFmsMessageViaUuidV2(uuidMd5 string, title, body string, version string) (total int, succ int) {
	clientInfoOpenApp, err := models.GetClientInfoOpenAppByUUIDMd5(uuidMd5)

	if !isMatchVersion(clientInfoOpenApp.AppVersionCode, version) {
		total = 0
		succ = 0

		logs.Info("[SendFmsMessageViaUuidV2] skip not match version uuid:%s, uuid_ver:%d, version:%s", uuidMd5, clientInfoOpenApp.AppVersionCode, version)
		return
	}

	total = 1
	succ = 0

	if err != nil || clientInfoOpenApp.FcmToken == "" {
		return
	}

	validTokens := make([]string, 0)
	validTokens = append(validTokens, clientInfoOpenApp.FcmToken)

	num, err := fcmmsg.SendMessage(validTokens, title, body, types.SkipToRegisterPage)
	if err != nil {
		logs.Error("[SendFmsMessageViaUuidV2] SendMessage error uuidMd5:%d, err:%v", uuidMd5, err)
	}

	succ = num

	return
}

func isMatchVersion(cur_ver int, version string) bool {
	if version == "" {
		return true
	}

	verStr := tools.Int2Str(cur_ver)

	vec := strings.Split(version, ",")
	for _, v := range vec {
		str := strings.TrimSpace(v)
		if !strings.Contains(str, "-") {
			if verStr == str {
				return true
			}
		} else {
			ranVec := strings.Split(str, "-")
			if len(ranVec) != 2 {
				continue
			}

			fValue, fErr := tools.Str2Int(strings.TrimSpace(ranVec[0]))
			sValue, sErr := tools.Str2Int(strings.TrimSpace(ranVec[1]))
			if fErr != nil || sErr != nil {
				continue
			}

			if cur_ver >= fValue && cur_ver <= sValue {
				return true
			}
		}
	}

	return false
}
