package feedback

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/astaxie/beego/orm"

	"micro-loan/common/i18n"
	"micro-loan/common/models"
	"micro-loan/common/types"
)

type BorrowFeedbackData struct {
	models.Feedback
	AccountTags types.CustomerTags
}

// GetTagDisplay 获取多重
func GetTagDisplay(lang string, tags int) (out string) {
	sortIndex := []int{}
	for i := range tagMap {
		sortIndex = append(sortIndex, i)
	}
	sort.Ints(sortIndex)

	for _, tag := range sortIndex {
		if tag&tags == tag {
			out += i18n.T(lang, tagMap[tag]) + ","
		}
	}
	if len(out) > 0 {
		out = strings.TrimSuffix(out, ",")
	}
	return
}

// ListBackend 返回
func ListBackend(condCntr map[string]interface{}, page int, pagesize int) (list []BorrowFeedbackData, total int64, err error) {
	obj := models.Feedback{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}
	// if pagesize < 1 {
	// 	pagesize = types.DefaultPagesize
	// }
	offset := (page - 1) * pagesize

	// 初始化查询条件
	where := whereBackend(condCntr)
	sqlCount := fmt.Sprintf("SELECT COUNT(feedback.id) FROM `%s` %s", obj.TableName(), where)
	sqlList := fmt.Sprintf("SELECT feedback.*, account_base.tags as account_tags FROM feedback left join account_base on feedback.account_id=account_base.id %s ORDER BY feedback.id desc LIMIT %d,%d", where, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

// ExportXLSX 返回
//func ExportXLSX(condCntr map[string]interface{}, lang string, rw http.ResponseWriter, ctx context.Context) (err error) {
func ExportXLSX(condCntr map[string]interface{}, lang string, rw http.ResponseWriter) (list []BorrowFeedbackData, err error) {
	obj := models.Feedback{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	// 初始化查询条件
	where := whereBackend(condCntr)

	sqlList := fmt.Sprintf("SELECT feedback.*, account_base.tags as account_tags FROM feedback left join account_base on feedback.account_id=account_base.id %s ORDER BY feedback.id desc", where)

	//list := []models.Feedback{}
	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

func whereBackend(condCntr map[string]interface{}) string {
	// 初始化查询条件
	cond := []string{}
	if v, ok := condCntr["mobile"]; ok {
		cond = append(cond, fmt.Sprintf("mobile=%s", v.(string)))
	}

	if v, ok := condCntr["account_id"]; ok {
		cond = append(cond, fmt.Sprintf("account_id=%d", v.(int64)))
	}

	//反馈分类
	if v, ok := condCntr["tags"]; ok {
		if tags, ok := v.([]string); ok && len(tags) > 0 {
			for _, tag := range tags {
				cond = append(cond, fmt.Sprintf("microloan.feedback.tags & %s = %s", tag, tag))
			}
		}
	}

	if v, ok := condCntr["ctime_start"]; ok {
		cond = append(cond, fmt.Sprintf("ctime>=%d", v))
	}

	if v, ok := condCntr["ctime_end"]; ok {
		cond = append(cond, fmt.Sprintf("ctime<%d", v))
	}

	//id check
	if v, ok := condCntr["id_check"]; ok {
		cond = append(cond, fmt.Sprintf("microloan.feedback.id =%s", v.(string)))
	}

	//app
	if v, ok := condCntr["app_version"]; ok {
		cond = append(cond, fmt.Sprintf("app_version = '%s'", v.(string)))
	}

	//api
	if v, ok := condCntr["api_version"]; ok {
		cond = append(cond, fmt.Sprintf("api_version = '%s'", v.(string)))
	}

	//模糊搜索的文本
	if v, ok := condCntr["check_txt"]; ok {
		cond = append(cond, fmt.Sprintf("content like '%%%s%%'", v))
	}

	if v, ok := condCntr["char_num"]; ok {
		cond = append(cond, fmt.Sprintf("length(content) > %d ", v.(int)))
	}
	//用户分类
	if v, ok := condCntr["user_tags"]; ok {
		cond = append(cond, fmt.Sprintf("account_base.tags=%v ", v))
	}

	if len(cond) > 0 {
		return "WHERE " + strings.Join(cond, " AND ")
	}
	return ""
}
