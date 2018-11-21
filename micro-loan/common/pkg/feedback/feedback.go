package feedback

import (
	"micro-loan/common/cerror"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// CheckCreateRequired 检查
func CheckCreateRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"tags":             true,
		"content":          true,
		"app_version":      true,
		"app_version_code": true,
		"ui_version":       true,
	}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

// CreateByCustomer 用户创建反馈
func CreateByCustomer(accountID int64, data map[string]interface{}, photoIds []int64) (id int64, errCode cerror.ErrCode) {
	errCode = cerror.CodeSuccess
	tagsString, okTag := data["tags"].(string)
	content, okContent := data["content"].(string)
	if okTag && len(tagsString) <= 0 && okContent && len(content) <= 0 {
		errCode = cerror.InvalidRequestData
		return
	}

	accountBase, err := models.OneAccountBaseByPkId(accountID)
	if err != nil {
		errCode = cerror.InvalidAccount
		return
	}

	m := models.Feedback{}
	m.AccountID = accountID
	m.Content = content
	m.Mobile = accountBase.Mobile
	tagsStringSlice := strings.Split(tagsString, ",")
	m.Tags = caculateTagsIntByStringSlice(tagsStringSlice)

	m.AppVersion = data["app_version"].(string)
	m.AppVersionCode, _ = strconv.Atoi(data["app_version_code"].(string))
	m.UIVersion = data["app_version_code"].(string)
	m.TaskVersion = types.TaskVersion
	m.ApiVersion = types.AppVersion
	m.Ctime = tools.GetUnixMillis()
	m.Status = types.StatusValid
	m.PhotoId1 = photoIds[0]
	m.PhotoId2 = photoIds[1]
	m.PhotoId3 = photoIds[2]
	m.PhotoId4 = photoIds[3]

	order, orderErr := models.GetUserLastOrder(m.AccountID)
	if orderErr == nil {
		m.CurrentOrderID = order.Id
		m.CurrentOrderStatus = order.CheckStatus
		m.CurrentOrderApplyTime = order.ApplyTime
	}
	m.ApplyOrderNum = models.GetUserOrderNum(m.AccountID)
	m.ApplyOrderSuccNum = models.GetUserApplySuccOrderNum(m.AccountID)

	o := orm.NewOrm()
	o.Using(m.Using())

	id, err = o.Insert(&m)
	if err != nil {
		logs.Error("[feedback.CreateByCustomer] insert failed by", err)
		errCode = cerror.CodeUnknown
	}
	return
}

func caculateTagsIntByStringSlice(tagsStringSlice []string) (tags int) {
	for _, s := range tagsStringSlice {
		if tagInt, err := strconv.Atoi(strings.Trim(s, " ")); err == nil {
			logs.Warn("tagInt", tagInt)

			if _, ok := tagMap[tagInt]; ok {
				tags = tagInt | tags
			} else {
				logs.Error("[caculateTagsIntByStringSlice] not found tag:", tagInt)
			}
		}
	}
	return
}
