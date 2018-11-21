package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/pkg/accesstoken"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

var customerFieldMap = map[string]string{
	"Id":           "id",
	"RegisterTime": "register_time",
}

func customerListCond(condCntr map[string]interface{}) (cond *orm.Condition) {
	cond = orm.NewCondition()

	cond = cond.And("is_deleted__lte", 0)
	// 创造条件
	if value, ok := condCntr["mobile"]; ok {
		cond = cond.And("mobile__icontains", tools.Escape(value.(string)))
	}
	if value, ok := condCntr["realname"]; ok {
		cond = cond.And("realname__icontains", value)
	}
	if value, ok := condCntr["tags"]; ok {
		tags, _ := value.(types.CustomerTags)
		cond = cond.And("tags", tags)
	}
	if value, ok := condCntr["idCheckStatus"]; ok {
		if value.(int) == 1 {
			cond = cond.AndNot("third_id", "")
		} else {
			cond = cond.And("third_id", "")
		}
	}
	if value, ok := condCntr["register_time_start"]; ok {
		cond = cond.And("register_time__gte", strconv.FormatInt(value.(int64), 10))
	}
	if value, ok := condCntr["register_time_end"]; ok {
		cond = cond.And("register_time__lt", strconv.FormatInt(value.(int64), 10))
	}
	if value, ok := condCntr["user_account_id"]; ok {
		cond = cond.And("id", value.(int64))
	}
	return
}

// AccountBaseWithOrigin 声明包含来源信息的用户基本信息结构体
type AccountBaseWithOrigin struct {
	models.AccountBase
	OriginID    int64 `orm:"column(origin_id)"`
	Balance     int64
	MediaSource string
	Campaign    string
}

// CustomerList 根据查询条件返回后台查询客户查询列表, 包括来源数据, 来自 appsflyer_source 表
func CustomerList(condCntr map[string]interface{}, page, pagesize int) (list []AccountBaseWithOrigin, num int64, err error) {
	if len(condCntr) == 0 {
		return
	}
	obj := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = types.DefaultPagesize
	}
	offset := (page - 1) * pagesize

	// 初始化查询条件
	where, leftJoinAppsflyer := customerWhereBackend(condCntr)

	orderBy := ""
	if v, ok := condCntr["field"]; ok {
		if vF, okF := customerFieldMap[v.(string)]; okF {
			orderBy += "p." + vF
		} else {
			orderBy += "p.id"
		}
	} else {
		orderBy += "p.id"
	}

	if v, ok := condCntr["sort"]; ok && v == "ASC" {
		orderBy += " ASC"
	} else {
		orderBy += " DESC"
	}
	var sqlCount string
	if leftJoinAppsflyer {
		sqlCount = fmt.Sprintf("SELECT COUNT(p.`id`) FROM `%s` AS p LEFT JOIN %s AS s ON p.appsflyer_id = s.appsflyer_id %s", obj.TableName(), models.APPSFLYER_SOURCE_TABLENAME, where)
	} else {
		sqlCount = fmt.Sprintf("SELECT COUNT(p.`id`) FROM `%s` AS p  %s", obj.TableName(), where)
	}

	sqlList := fmt.Sprintf("SELECT p.*, s.id as `origin_id`, s.media_source, s.campaign FROM `%s` AS p LEFT JOIN %s AS s ON p.appsflyer_id = s.appsflyer_id %s ORDER BY %s LIMIT %d, %d",
		obj.TableName(), models.APPSFLYER_SOURCE_TABLENAME, where, orderBy, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&num)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

func CustomerAddBalance(list []AccountBaseWithOrigin) {
	// 一般情况下 list里保存的是当页要展示的客户 15个
	if len(list) == 0 {
		logs.Warn("[CustomerAddBalance] list len == 0")
		return
	}

	//1
	ids := []string{}
	for k := range list {
		ids = append(ids, tools.Int642Str(list[k].Id))
	}
	strIds := strings.Join(ids, ",")

	//2
	obj := models.AccountBalance{}
	sql := "SELECT * FROM %s WHERE account_id in ( %s )"
	sql = fmt.Sprintf(sql, obj.TableName(), strIds)
	logs.Info("[CustomerAddBalance] sql:%s", sql)

	//3
	balanceList := []models.AccountBalance{}

	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	r := o.Raw(sql)
	r.QueryRows(&balanceList)

	for k := range list {
		ids = append(ids, tools.Int642Str(list[k].Id))

		for _, v := range balanceList {
			if v.AccountId == list[k].Id {
				list[k].Balance = v.Balance
			}
		}
	}
}

// 包含主表和关联表查询条件
// 主表 简称 "p" 新增条件, 写在函数上方, 无需加表名前缀, where column>0....
// 子表在函数最下面 s.column=....
func customerWhereBackend(condCntr map[string]interface{}) (where string, leftJoinAppsflyer bool) {
	// 初始化查询条件
	cond := []string{}
	// 主表 where 条件
	/** 去掉金管局条件 2018.08.09 */
	//cond = append(cond, "is_deleted=0")
	if v, ok := condCntr["mobile"]; ok {
		cond = append(cond, fmt.Sprintf("mobile LIKE('%%%s%%')", v.(string)))
	}
	if v, ok := condCntr["realname"]; ok {
		cond = append(cond, fmt.Sprintf("realname LIKE('%%%s%%')", v))
	}
	if v, ok := condCntr["tags"]; ok {
		cond = append(cond, fmt.Sprintf("tags=%d", v))
	}

	if v, ok := condCntr["idCheckStatus"]; ok {
		if v.(int) == 1 {
			cond = append(cond, "third_id<>''")
		} else {
			cond = append(cond, "third_id=''")
		}
	}
	if v, ok := condCntr["register_time_start"]; ok {
		cond = append(cond, fmt.Sprintf("register_time>=%d", v))
	}
	if v, ok := condCntr["register_time_end"]; ok {
		cond = append(cond, fmt.Sprintf("register_time<%d", v))
	}
	if v, ok := condCntr["user_account_id"]; ok {
		cond = append(cond, fmt.Sprintf("id=%d", v))
	}
	if v, ok := condCntr["identity"]; ok {
		cond = append(cond, fmt.Sprintf("identity=%d", v))
	}
	if v, ok := condCntr["generalize"]; ok {
		cond = append(cond, fmt.Sprintf("channel = '%s'", v.(string)))
	}
	for i, condition := range cond {
		cond[i] = "p." + condition
	}

	// 表 s 查询条件
	if v, ok := condCntr["media_source"]; ok {
		leftJoinAppsflyer = true
		if v.(string) == "-1" {
			//  -1 代表未识别
			cond = append(cond, fmt.Sprintf("s.id IS NULL"))
		} else {
			cond = append(cond, fmt.Sprintf("s.media_source = '%s'", v))
		}
	}
	if v, ok := condCntr["campaign"]; ok {
		leftJoinAppsflyer = true
		cond = append(cond, fmt.Sprintf("s.campaign = '%s'", v))
	}

	if len(cond) > 0 {
		where = "WHERE " + strings.Join(cond, " AND ")
	}

	return
}

func CustomerRiskList(condCntr map[string]interface{}, page, pagesize int) (count int64, list []models.CustomerRisk, num int64, err error) {
	obj := models.CustomerRisk{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	qs := o.QueryTable(obj.TableName())
	cond := orm.NewCondition()

	if value, ok := condCntr["risk_type"]; ok {
		cond = cond.And("risk_type", value)
	}
	if value, ok := condCntr["risk_value"]; ok {
		cond = cond.And("risk_value__icontains", value)
	}
	if value, ok := condCntr["ctime_start"]; ok {
		cond = cond.And("ctime__gte", value)
	}
	if value, ok := condCntr["ctime_end"]; ok {
		cond = cond.And("ctime__lt", value)
	}
	if value, ok := condCntr["review_time_start"]; ok {
		cond = cond.And("review_time__gte", value)
	}
	if value, ok := condCntr["review_time_end"]; ok {
		cond = cond.And("review_time__lt", value)
	}
	if value, ok := condCntr["status"]; ok {
		cond = cond.And("status", value)
	}
	if value, ok := condCntr["is_deleted"]; ok {
		cond = cond.And("is_deleted", value)
	}
	if value, ok := condCntr["source"]; ok {
		if value.(int) >= 1 {
			cond = cond.And("op_uid__gte", value)
		} else {
			cond = cond.And("op_uid", value)
		}
	}

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	count, _ = qs.SetCond(cond).Count()
	num, err = qs.SetCond(cond).OrderBy("-id").Limit(pagesize).Offset(offset).All(&list)

	return
}

type AccDupOrderNo struct {
	Mobile          string
	Contact1        string
	Contact2        string
	ResidentAddress string
	CompanyName     string
	CompanyAddress  string
	Imei            string
	IP              string
}

func GetDupOrderNo(baseInfo *models.AccountBase, profile *models.AccountProfile, clientInfo *models.ClientInfo) (dupOrderNo AccDupOrderNo) {

	var contact []string
	if baseInfo != nil && len(baseInfo.Mobile) > 0 {
		contact = append(contact, baseInfo.Mobile)

	}
	if profile != nil {
		if len(profile.Contact1) > 0 {
			contact = append(contact, profile.Contact1)
		}
		if len(profile.Contact2) > 0 {
			contact = append(contact, profile.Contact2)
		}
	}
	contactStr := tools.ArrayToString(contact, ",")
	contactStr = fmt.Sprintf("(%s)", contactStr)

	if baseInfo != nil {
		dupOrderNo.Mobile = getDupOrderNo(baseInfo.Id, models.ACCOUNT_BASE_TABLENAME, contactStr, "mobile")
	}

	if profile != nil {
		dupOrderNo.Contact1 = getDupOrderNo(profile.AccountId, models.ACCOUNT_PROFILE_TABLENAME, contactStr, "contact1")
		dupOrderNo.Contact2 = getDupOrderNo(profile.AccountId, models.ACCOUNT_PROFILE_TABLENAME, contactStr, "contact2")
		dupOrderNo.ResidentAddress = getDupOrderNo(profile.AccountId, models.ACCOUNT_PROFILE_TABLENAME, profile.ResidentAddress, "resident_address")
		dupOrderNo.CompanyName = getDupOrderNo(profile.AccountId, models.ACCOUNT_PROFILE_TABLENAME, profile.CompanyName, "company_name")
		dupOrderNo.CompanyAddress = getDupOrderNo(profile.AccountId, models.ACCOUNT_PROFILE_TABLENAME, profile.CompanyAddress, "company_address")
	}

	if clientInfo != nil {
		dupOrderNo.Imei = getDupOrderNo(clientInfo.RelatedId, models.CLIENT_INFO_TABLENAME, clientInfo.ImeiMd5, "imei_md5")
		dupOrderNo.IP = getDupOrderNo(clientInfo.RelatedId, models.CLIENT_INFO_TABLENAME, clientInfo.IP, "ip")
	}

	return
}

func getDupOrderNo(accountId int64, table, value, key string) (dupOrderNo string) {
	if value == "" {
		return
	}

	o := orm.NewOrm()

	cond := "1=1"
	var sqlOrder string

	switch table {
	case models.ACCOUNT_BASE_TABLENAME:
		m := models.AccountBase{}
		o.Using(m.UsingSlave())
		cond = fmt.Sprintf("%s%s%s%s%s", cond, " AND account_base.", key, " IN ", value)
		sqlOrder = fmt.Sprintf(`SELECT distinct account_base.id FROM account_base WHERE %s`, cond)
	case models.ACCOUNT_PROFILE_TABLENAME:
		m := models.AccountProfile{}
		o.Using(m.UsingSlave())
		if key == "contact1" || key == "contact2" {
			cond = fmt.Sprintf("%s%s%s%s%s", cond, " AND account_profile.", key, " IN ", value)
		} else {
			cond = fmt.Sprintf("%s%s%s%s%s%s", cond, " AND account_profile.", key, "='", value, "'")
		}
		sqlOrder = fmt.Sprintf(`SELECT distinct account_profile.account_id FROM account_profile WHERE %s`, cond)
	case models.CLIENT_INFO_TABLENAME:
		m := models.ClientInfo{}
		o.Using(m.UsingSlave())
		cond = fmt.Sprintf("%s%s%s%s%s%s", cond, " AND client_info.", key, "='", value, "'")
		sqlOrder = fmt.Sprintf(`SELECT distinct client_info.related_id FROM client_info WHERE %s`, cond)
	}

	var accountIds []int64
	if len(sqlOrder) > 0 {
		r := o.Raw(sqlOrder)
		r.QueryRows(&accountIds)
	}

	orderIds := make([]int64, 0)
	for _, v := range accountIds {
		if v == accountId {
			continue
		}

		o, e := models.GetOrderByAccountId(v)
		if e != nil {
			continue
		}

		orderIds = append(orderIds, o.Id)
	}

	dupOrderNo = ArrayToParagraphString(orderIds)
	return
}

/*
func getDupOrderNo2(accountId int64, table, value, key string) (dupOrderNo string) {
	if value == "" {
		return
	}

	o := orm.NewOrm()
	order := models.Order{}
	o.Using(order.UsingSlave())

	cond := "1=1"
	var sqlOrder string
	switch table {
	case models.ACCOUNT_BASE_TABLENAME:
		cond = fmt.Sprintf("%s%s%d", cond, " AND account_base.Id<>", accountId)
		cond = fmt.Sprintf("%s%s%s%s%s", cond, " AND account_base.", key, " IN ", value)

		sqlOrder = fmt.Sprintf(`SELECT distinct orders.id FROM orders LEFT JOIN account_base ON orders.user_account_id=account_base.id WHERE %s`, cond)
	case models.ACCOUNT_PROFILE_TABLENAME:
		cond = fmt.Sprintf("%s%s%d", cond, " AND account_profile.account_id<>", accountId)
		if key == "contact1" || key == "contact2" {
			cond = fmt.Sprintf("%s%s%s%s%s", cond, " AND account_profile.", key, " IN ", value)
		} else {
			cond = fmt.Sprintf("%s%s%s%s%s%s", cond, " AND account_profile.", key, "='", value, "'")
		}
		sqlOrder = fmt.Sprintf(`SELECT distinct orders.id FROM orders LEFT JOIN account_profile ON orders.user_account_id=account_profile.account_id WHERE %s`, cond)
	case models.CLIENT_INFO_TABLENAME:
		cond = fmt.Sprintf("%s%s%d", cond, " AND client_info.related_id<>", accountId)
		cond = fmt.Sprintf("%s%s%s%s%s%s", cond, " AND client_info.", key, "='", value, "'")

		sqlOrder = fmt.Sprintf(`SELECT distinct orders.id FROM orders LEFT JOIN client_info ON orders.user_account_id=client_info.related_id WHERE %s`, cond)
	}

	var orderIdArr []int64
	if len(sqlOrder) > 0 {
		r := o.Raw(sqlOrder)
		r.QueryRows(&orderIdArr)
	}

	dupOrderNo = ArrayToParagraphString(orderIdArr)
	return
}
*/

func AddCustomerFollow(cid, opUid, followTime int64, content, remark string) (id int64, err error) {
	obj := models.CustomerFollow{
		CustomerId: cid,
		FollowTime: followTime,
		OpUid:      opUid,
		Content:    content,
		Remark:     remark,
		Ctime:      tools.GetUnixMillis(),
	}

	o := orm.NewOrm()
	o.Using(obj.Using())

	id, err = o.Insert(&obj)

	return
}

// AjaxAccountBaseModify 针对 ajax固定模式修改，提供filed,value,主键ID
func AjaxAccountBaseModify(ID int64, filed, value string) (num int64, err error) {
	accountbase := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(accountbase.Using())

	sql := fmt.Sprintf(`UPDATE %s SET %s = '%s' WHERE id = %d `, accountbase.TableName(), filed, value, ID)

	res, err := o.Raw(sql).Exec()
	if err == nil {
		num, _ = res.RowsAffected()
	}
	return
}

func CustomerFollowList(cid int64) (list []models.CustomerFollow, num int64, err error) {
	obj := models.CustomerFollow{}

	o := orm.NewOrm()
	o.Using(obj.Using())

	num, err = o.QueryTable(obj.TableName()).Filter("customer_id", cid).OrderBy("id").All(&list)

	return
}

func CustomerAge(identity string) (age int, err error) {
	if len(identity) != types.LimitIdentity {
		err = fmt.Errorf("identity length is not enough, identity: %s", identity)
		return
	}

	birthYearStr := fmt.Sprintf("19%s", tools.SubString(identity, 10, 2))
	birthYear, _ := tools.Str2Int(birthYearStr)
	if birthYear <= 0 {
		err = fmt.Errorf("can't find birth year from identity: %s", identity)
		return
	}

	age = time.Now().Year() - birthYear
	if age <= 0 {
		age = 0
		err = fmt.Errorf("identity age out of range. identity: %s, age: %d", identity, age)
		return
	}

	// 兼容2000年以后出生的人
	if age > 100 {
		age = age - 100
	}

	age = age + 1 // 周年向上加一年

	return
}

// CustomerTags 计算出账号的客户标签
func CustomerTags(accountID int64) (tags int64) {

	// CustomerTagsPotential   CustomerTags = 1 // 潜在客户 ：已完成注册，但未进行身份认证客户
	// CustomerTagsTarget      CustomerTags = 2 // 目标客户：身份认证通过但未提交过一笔借款申请的客户
	// CustomerTagsProspective CustomerTags = 3 // 准客户：未完成首贷，但存在进行中的借款申请（审核中/等待还款/审核拒绝）
	// CustomerTagsDeal        CustomerTags = 4 // 成交客户：首贷完成的客户 （完成指已经结清）
	// CustomerTagsLoyal       CustomerTags = 5 // 忠实客户：复贷完成的客户
	clearOrderNum, _ := models.GetClearedOrderNumByAccountId(accountID)
	if clearOrderNum >= 2 {
		tags = int64(types.CustomerTagsLoyal)
	}
	if clearOrderNum == 1 {
		tags = int64(types.CustomerTagsDeal)
	}
	if clearOrderNum == 0 {
		order, _ := models.GetOrderByAccountId(accountID)
		if order.Id > 0 {
			tags = int64(types.CustomerTagsProspective)
		} else {
			if IdentityVerify(accountID) {
				tags = int64(types.CustomerTagsTarget)
			}
		}
	}

	return
}

// CustomerWaitingTag 等待打标签客户 tag为 1，2，3，4的客户都需要
func CustomerWaitingTag() (list []models.AccountBase, err error) {
	accountBase := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(accountBase.Using())

	sqlOrder := "SELECT * from " + models.ACCOUNT_BASE_TABLENAME + " WHERE tags in (1,2,3,4) "
	// sqlOrder = fmt.Sprintf("%s%d", sqlOrder)
	o.Raw(sqlOrder).QueryRows(&list)
	return list, err
}

// UpdateCustomer 更新账号表
func UpdateCustomer(accountID, tags int64) (num int64, err error) {
	accountBase := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(accountBase.Using())

	sql := fmt.Sprintf(`UPDATE %s SET tags = %d WHERE id = %d`, accountBase.TableName(), tags, accountID)
	res, err := o.Raw(sql).Exec()
	if err == nil {
		num, _ = res.RowsAffected()
	}
	return
}

func DeleteCustomer(cid int64) {
	o := orm.NewOrm()

	accountbase := models.AccountBase{}
	o.Using(accountbase.Using())
	sql := fmt.Sprintf(`UPDATE %s SET is_deleted = 1 WHERE id = %d `, accountbase.TableName(), cid)
	o.Raw(sql).Exec()

	order := models.Order{}
	o.Using(order.Using())
	sql = fmt.Sprintf(`UPDATE %s SET is_deleted = 1 WHERE user_account_id = %d `, order.TableName(), cid)
	o.Raw(sql).Exec()
}

func SuperDeleteCustomer(opUid, cid int64) error {
	accountBase, _ := models.OneAccountBaseByPkId(cid)
	oldAccountBase := accountBase
	_, err := accountBase.Delete()
	if err != nil {
		err = fmt.Errorf("[ModifyMobile] delete oldAccountBase err.  opUid:%d accountId:%v err:%v ", opUid, cid, err)
		logs.Error(err)
		return err
	}
	models.OpLogWrite(opUid, accountBase.Id, models.OpCodeAccountBaseDelete, accountBase.TableName(), oldAccountBase, "")

	// 强力删除,保留订单,只删除account_base
	/*
		order, _ := models.GetOrderByAccountId(cid)
		oldOrder := order
		_, err = order.Delete()
		if err != nil {
			err = fmt.Errorf("[ModifyMobile] delete oldOrder err.  opUid:%d accountId:%v err:%v opUid:%d", opUid, cid, err)
			logs.Error(err)
			return err
		}
		models.OpLogWrite(opUid, order.UserAccountId, models.OpCodeOrderDelete, order.TableName(), oldOrder, "")
	*/
	return nil
}

func ModifyMobile(opUid int64, accountId int64, mobileNew string) error {

	//1、取原帐号信息
	account, err := models.OneAccountBaseByPkId(accountId)
	if err != nil {
		err = fmt.Errorf("[ModifyMobile] query account err:%v accountId:%d mobile:%s opUid:%d", err, accountId, mobileNew, opUid)
		logs.Error(err)
		return err
	}
	if account.Mobile == mobileNew {
		return nil
	}

	//2、检查新手机号是否注册
	err = UpdateOneAccountByMobild(opUid, mobileNew)
	if err != nil {
		err = fmt.Errorf("[ModifyMobile] UpdateOneAccountByMobild err:%v  accountId:%d mobile:%s opUid:%d", err, accountId, mobileNew, opUid)
		logs.Error(err)
		return err
	}

	//3、更新旧表
	old := account
	account.Mobile = mobileNew
	_, err = account.Update("mobile")
	if err != nil {
		logs.Error("[ModifyMobile] update err:%v accountId:%#v newMobile:%v opUid:%d", err, old, mobileNew, opUid)
		return err
	}
	//opUid int64, opCode OpCodeEnum, opTable string, original interface{}, edited interface{}
	models.OpLogWrite(opUid, account.Id, models.OpCodeAccountBaseUpdate, account.TableName(), old, account)

	// 手机号修改历史
	num, errs := models.GetAccountMobileModifyNum(old.Id)
	if errs != nil {
		return errs
	}
	if num == 0 {
		err = AddAccountMobileHistory(old.Id, old.Mobile)
		if err != nil {
			logs.Error("[ ModifyMobile ] insert accountMobileHistory failed, err is", err, ", mobileOld:", old.Mobile)
			return err
		}
	}

	err = AddAccountMobileHistory(old.Id, mobileNew)
	if err != nil {
		logs.Error("[ ModifyMobile ] insert accountMobileHistory failed, err is", err, ", mobileNew:", mobileNew)
		return err
	}

	return nil
}

func UpdateOneAccountByMobild(opUid int64, mobile string) error {
	accountNew, err := models.OneAccountBaseByMobile(mobile)
	if err == orm.ErrNoRows {
		logs.Debug("no new account. mobild:%s opUid:%d", mobile, opUid)
		return nil
	}

	// 是否有正式订单
	count, _ := dao.AccountLoanOrderCount(accountNew.Id)
	if count > 0 {
		err = fmt.Errorf("[ModifyMobile] new mobile:%s account:%#v AccountLoanOrderCount:%d", mobile, accountNew.Id, count)
		logs.Error(err)
		return err
	}

	//用户已新注册清除掉新用户
	logs.Warn("[UpdateOneAccountByMobild] delete account:%#v opUid:%d", accountNew, opUid)

	//accontToken
	accontToken := models.LatestToken(accountNew.Id)
	if accontToken == "" {
		logs.Warn("[UpdateOneAccountByMobild] LatestToken empty. accountNew:%#v opUid:%d", accountNew, opUid)
	}
	accesstoken.CleanTokenCache(types.PlatformAndroid, accontToken)

	//accontpro
	//accountPro, _ := models.OneAccountProfileByAccountID(accountNew.Id)
	//oldPro := accountPro
	//_, err = accountPro.Delete()
	//if err != nil {
	//	err = fmt.Errorf("[ModifyMobile] delete accountPro err. account:%#v err:%v opUid:%d", accountNew, err, opUid)
	//	logs.Error(err)
	//	return err
	//}
	//models.OpLogWrite(opUid, accountPro.AccountId, models.OpCodeAccountBaseDelete, accountPro.TableName(), oldPro, "")

	//accountbase
	oldAccount := accountNew
	accountNew.Mobile = mobile + types.CustomerAccountInvalidSuffix
	_, err = accountNew.Update("mobile")
	if err != nil {
		err = fmt.Errorf("[UpdateOneAccountByMobild] update accountbase err. account:%#v err:%v opUid:%d", accountNew, err, opUid)
		logs.Error(err)
		return err
	}
	models.OpLogWrite(opUid, oldAccount.Id, models.OpCodeAccountBaseDelete, oldAccount.TableName(), oldAccount, "")
	return nil
}

// CustomerRecallTag 给客户打召回标签
func CustomerRecallTag(timetag int64) {
	N, _ := config.ValidItemInt("customer_recall_phone_verify")
	if N > 0 {
		timeUnix := tools.GetUnixMillis()
		beforeNday := timeUnix - int64(N)*tools.MILLSSECONDADAY
		logs.Debug("[CustomerRecallTag] now:", timeUnix, "N:", N, " beforeNday", beforeNday)
		records, _ := models.GetRefuseRecordByPhoneStatus(beforeNday)
		logs.Debug("[CustomerRecallTag] records:", records)

		if len(records) > 0 {
			for _, v := range records {
				refuseDay := (timetag - v.Ctime) / tools.MILLSSECONDADAY
				logs.Debug("[CustomerRecallTag] refuseDay:", refuseDay, "N:", N)
				orderData, _ := models.GetOrder(v.OrderId)

				if orderData.Id > 0 {
					accountId := orderData.UserAccountId
					recallTagLog, _ := models.OneRecallPhoneVerifyTagLogByAOID(accountId, orderData.Id)
					//打电核拒绝召回标记，是第一次标记或者有重新标记标识的
					if refuseDay < int64(N) && (recallTagLog.Remark == types.RemarkTagYes || recallTagLog.Id == 0) {
						ChangeCustomerRecall(accountId, orderData.Id, types.RecallTagPhoneVerify, types.RemarkTagNone)
					}
					if refuseDay >= int64(N) {
						ChangeCustomerRecall(accountId, orderData.Id, types.RecallTagNone, types.RemarkTagNone)
					}
				} else {
					logs.Debug("[CustomerRecallTag] order data is nil")
				}

			}
		}
		cancelOrders := GetCancelOrders(beforeNday)
		if len(cancelOrders) > 0 {
			for _, v := range cancelOrders {
				logs.Notice("[CustomerRecallTag] cancel tag orderID:%d,accountID:%d", v.OrderId, v.AccountId)
				ChangeCustomerRecall(v.AccountId, v.OrderId, types.RecallTagNone, types.RemarkTagNone)
			}
		}

	}

}

type CancelOrders struct {
	OrderId   int64
	AccountId int64
}

func GetCancelOrders(beforeNday int64) (cancelOrders []CancelOrders) {

	accountBaseExt := models.AccountBaseExt{}
	customerRecallTagChangeLog := models.CustomerRecallTagChangeLog{}
	o := orm.NewOrm()
	o.Using(accountBaseExt.UsingSlave())
	sql := fmt.Sprintf(`SELECT l.account_id, l.order_id FROM %s l
	LEFT JOIN %s a ON l.account_id = a.account_id
	WHERE l.edit_recall_tag = %d
	AND a.recall_tag = %d
	AND l.ctime < %d`,
		customerRecallTagChangeLog.TableName(),
		accountBaseExt.TableName(),
		types.RecallTagPhoneVerify,
		types.RecallTagPhoneVerify,
		beforeNday)
	o.Raw(sql).QueryRows(&cancelOrders)
	return
}

func ChangeCustomerRecall(accountID, orderID int64, recallTag, remark int) (err error) {
	accountExt, _ := models.OneAccountBaseExtByPkId(accountID)
	if accountExt.AccountId == 0 && recallTag == types.RecallTagNone {
		err := fmt.Errorf("[ChangeCustomerRecall] do not have account ext so no need to cancle. account:%#v order:%d", accountExt, orderID)
		logs.Error(err)
		return err
	}

	if accountExt.AccountId > 0 && accountExt.RecallTag == recallTag {
		logs.Warn("[ChangeCustomerRecall] no need to change tag. accountExt:%#v remark:%v", accountExt, remark)
		return nil
	}

	tag := tools.GetUnixMillis()
	org := accountExt.RecallTag
	accountExt.RecallTag = recallTag
	accountExt.Utime = tag
	if accountExt.AccountId == 0 {
		accountExt.AccountId = accountID
		accountExt.Ctime = tag
		err = accountExt.InsertWithNoReturn()
	} else {
		cols := []string{"recall_tag", "utime"}
		err = accountExt.UpdateWithNoReturn(cols)
	}

	// 记录操作日志
	opLog := models.CustomerRecallTagChangeLog{
		AccountId:       accountExt.AccountId,
		OrderId:         orderID,
		OrgionRecallTag: org,
		EditRecallTag:   recallTag,
		Remark:          remark,
		Ctime:           tag,
		Utime:           tag,
	}

	models.OrmInsert(&opLog)

	return err
}
