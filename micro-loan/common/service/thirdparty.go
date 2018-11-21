package service

import (
	"encoding/json"
	"fmt"
	"strings"

	//"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/models"
	"micro-loan/common/thirdparty"
	"micro-loan/common/thirdparty/doku"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func ListThirdpartyStatisticFee(condStr map[string]interface{}, page, pagesize int) (
	list []models.ThirdpartyStatisticFee, total int,
	totalCount int64, totalSuccessCount int64, totalChargeAmout int64,
	err error) {

	//将缓存中数据写入数据库
	thirdparty.MoveOutThirdpartyStatisticFeeFromCache()

	o := orm.NewOrm()
	fee := models.ThirdpartyStatisticFee{}
	o.Using(fee.UsingSlave())

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	where := fmt.Sprintf("WHERE 1 = 1 ")
	if f, ok := condStr["name"]; ok {
		where = fmt.Sprintf("%s AND name = '%s'", where, tools.Escape(f.(string)))
	}
	if f, ok := condStr["statistic_start"]; ok {
		where = fmt.Sprintf("%s AND statistic_date >= '%d'", where, f.(int64))
	}
	if f, ok := condStr["statistic_end"]; ok {
		where = fmt.Sprintf("%s AND statistic_date <= '%d'", where, f.(int64))
	}
	if f, ok := condStr["api_url"]; ok {
		where = fmt.Sprintf("%s AND api_md5 = '%s'", where, tools.Md5(f.(string)))
	}
	if f, ok := condStr["charge_type"]; ok {
		where = fmt.Sprintf("%s AND charge_type = '%d'", where, f.(int))
	}

	sqlCount := "SELECT COUNT(id) AS total"
	sqlStatistic := "SELECT SUM(call_count) AS totalCount, SUM(success_call_count) AS totalSuccessCount, SUM(total_price) AS totalChargeAmout"
	sqlSelect := "SELECT id, name, api, charge_type, price, total_price, call_count, success_call_count, hit_call_count, statistic_date_s "
	from := fmt.Sprintf(`FROM %s `, fee.TableName())

	// count
	sql := fmt.Sprintf(`%s %s %s`, sqlCount, from, where)
	r := o.Raw(sql)
	err = r.QueryRow(&total)
	if err != nil {
		return
	}

	// statistic
	sql = fmt.Sprintf(`%s %s %s`, sqlStatistic, from, where)
	r = o.Raw(sql)
	err = r.QueryRow(&totalCount, &totalSuccessCount, &totalChargeAmout)
	if err != nil {
		return
	}

	// data
	orderBy := " order by statistic_date_s desc , name desc "
	limit := fmt.Sprintf(`LIMIT %d, %d`, offset, pagesize)
	sql = fmt.Sprintf(`%s %s %s %s %s`, sqlSelect, from, where, orderBy, limit)
	r = o.Raw(sql)
	_, err = r.QueryRows(&list)
	return
}

func ListThirdpartyStatisticCustomer(condStr map[string]interface{}, page, pagesize int) (
	list []models.ThirdpartyStatisticCustomerInfo, total int,
	err error) {

	accountBase := models.AccountBase{}
	appsflyer := models.AppsflyerSource{}

	o := orm.NewOrm()
	customer := models.ThirdpartyStatisticCustomer{}
	o.Using(customer.UsingSlave())

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize
	where := fmt.Sprintf("WHERE 1 = 1 ")
	where = fmt.Sprintf("%s AND c.record_type = '%d'", where, types.RecordTypeTotal)
	if f, ok := condStr["user_account_id"]; ok {
		where = fmt.Sprintf("%s AND c.user_account_id = '%d'", where, f.(int64))
	}
	if f, ok := condStr["mobile"]; ok {
		where = fmt.Sprintf("%s AND a.mobile = '%s'", where, f.(string))
	}
	if f, ok := condStr["media_source"]; ok {
		where = fmt.Sprintf("%s AND s.media_source = '%s'", where, f.(string))
	}
	if f, ok := condStr["campaign"]; ok {
		where = fmt.Sprintf("%s AND s.campaign = '%s'", where, f.(string))
	}

	sqlCount := "SELECT COUNT(user_account_id) AS total"
	sqlSelect := `SELECT c.user_account_id, c.mobile, s.media_source, s.campaign, c.call_count, 
		c.success_call_count, c.hit_call_count, c.cutomer_total_cost, a.tags , a.realname`
	from := fmt.Sprintf(`FROM %s c `, customer.TableName())
	leftJoin := fmt.Sprintf(` left join  %s a on c.user_account_id = a.id left join %s s on s.appsflyer_id = a.appsflyer_id`, accountBase.TableName(), appsflyer.TableName())

	// count
	sql := fmt.Sprintf(`%s %s %s %s`, sqlCount, from, leftJoin, where)
	r := o.Raw(sql)
	err = r.QueryRow(&total)
	if err != nil {
		return
	}

	// data
	orderBy := " order by c.id desc "
	limit := fmt.Sprintf(`LIMIT %d, %d`, offset, pagesize)
	sql = fmt.Sprintf(`%s %s %s %s %s %s`, sqlSelect, from, leftJoin, where, orderBy, limit)
	r = o.Raw(sql)
	_, err = r.QueryRows(&list)

	// logs.Warn("%#v", list)
	return
}

func ListThirdpartyStatisticCustomerDetail(id int64) (
	list []models.ThirdpartyStatisticCustomer, err error) {

	o := orm.NewOrm()
	customer := models.ThirdpartyStatisticCustomer{}
	o.Using(customer.UsingSlave())

	_, err = o.QueryTable(customer.TableName()).
		Filter("user_account_id", id).
		Filter("record_type", types.RecordTypeSingle).OrderBy("api").All(&list)

	// logs.Warn("list ", list, " dd")
	return
}

// ThirdpartyListBackend 返回
func ThirdpartyListBackend(condCntr map[string]interface{}, page int, pagesize int) (list []models.ThirdpartyRecord, total int64, err error) {
	//logs.Debug("condCntr:", condCntr)
	if len(condCntr) <= 0 {
		return
	}

	obj := models.ThirdpartyRecord{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}
	// if pagesize < 1 {
	// 	pagesize = types.DefaultPagesize
	// }
	offset := (page - 1) * pagesize

	tableName := obj.TableName()
	if v, ok := condCntr["month"]; ok {
		if v.(int64) == 1 {
			tableName = obj.OriTableName()
		} else {
			tableName = obj.TableNameByMonth(v.(int64))
		}
	}

	// 初始化查询条件
	where := whereBackend(condCntr)
	sqlCount := fmt.Sprintf("SELECT COUNT(id) FROM `%s` %s", tableName, where)
	sqlList := fmt.Sprintf("SELECT * FROM `%s` %s ORDER BY id desc LIMIT %d,%d", tableName, where, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

func whereBackend(condCntr map[string]interface{}) string {
	// 初始化查询条件
	cond := []string{}
	if v, ok := condCntr["thirdparty"]; ok {
		cond = append(cond, fmt.Sprintf("thirdparty=%d", v.(int)))
	}

	if v, ok := condCntr["id_check"]; ok {
		cond = append(cond, fmt.Sprintf("id=%s", v.(string)))
	}
	if v, ok := condCntr["related_id"]; ok {
		cond = append(cond, fmt.Sprintf("related_id=%s", v.(string)))
	}

	if v, ok := condCntr["api"]; ok {
		cond = append(cond, fmt.Sprintf("api like '%%%s%%'", v.(string)))
	}

	if v, ok := condCntr["request"]; ok {
		cond = append(cond, fmt.Sprintf("request like '%%%s%%'", v.(string)))
	}

	if v, ok := condCntr["response"]; ok {
		cond = append(cond, fmt.Sprintf("response like '%%%s%%'", v.(string)))
	}

	if v, ok := condCntr["ctime_start"]; ok {
		cond = append(cond, fmt.Sprintf("ctime>=%d", v))
	}

	if v, ok := condCntr["ctime_end"]; ok {
		cond = append(cond, fmt.Sprintf("ctime<%d", v))
	}

	if len(cond) > 0 {
		return "WHERE " + strings.Join(cond, " AND ")
	}
	return ""
}

func modifyResp(resp string) string {

	ret := strings.Replace(resp, "\\\"", "\"", len(resp)-1)
	ret = strings.Replace(ret, "\"{", "{", len(ret)-1)
	ret = strings.Replace(ret, "}\"", "}", len(ret)-1)
	return ret
}

// 补单时用 其他时候慎用
func FixThirdParty(resp doku.DokuRemitResp, orderId int64, tableName string, recordId int, opUid int64) error {
	tr := models.ThirdpartyRecord{}
	o := orm.NewOrm()
	o.Using(tr.Using())

	name := "thirdparty_record_" + tableName
	sql := "select * from %s where id = %d"
	sql = fmt.Sprintf(sql, name, recordId)

	err := o.Raw(sql).QueryRow(&tr)
	if err != nil {
		return fmt.Errorf("查询第三方记录出错. err:%v sql:%v", err, sql)
	}

	if tr.RelatedId != orderId {
		return fmt.Errorf("订单Id不匹配. tr:%#v", tr)
	}

	reqStr := modifyResp(tr.Request)
	req := doku.DokuRemitReq{}
	err = json.Unmarshal([]byte(reqStr), &req)
	if err != nil {
		return fmt.Errorf("请求解析出错, err:%v tr:%#v", err, tr)
	}

	if req.Inquiry.IdToken != resp.Remit.PaymentData.InquiryId {
		return fmt.Errorf("InquiryId 不一致。 tr:%#v", tr)
	}

	if tr.Response != "" && tr.Response != "\"\"" {
		return fmt.Errorf("Response 不为空不需要补单 tr：%#v", tr)
	}

	old := tr
	jsStr, _ := json.Marshal(resp)
	tr.Response = string(jsStr)

	upSql := "update %s set response = '%s' where id=%d"
	upSql = fmt.Sprintf(upSql, name, jsStr, recordId)
	err = o.Raw(upSql).QueryRow(&tr)
	if err.Error() != types.EmptyOrmStr {
		return fmt.Errorf(" 更新数据库失败 err：%#v upSql:%v", err, upSql)
	}
	models.OpLogWrite(opUid, orderId, models.OpCodeSupplementOrder, name, old, tr)

	return nil

}
