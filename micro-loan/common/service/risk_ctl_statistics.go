// 统计相关的函数

package service

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// 近1个月内第一联系人存在拒贷借款订单 实现的逻辑可能有问题,理论不会有多条记录...先这么着吧
func ContactHasRejectLoanOderInDays(mobile string, days int64) (yes bool, num int64, err error) {
	accountM := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(accountM.UsingSlave())

	var accountList []models.AccountBase
	num, err = o.QueryTable(accountM.TableName()).Filter("mobile", mobile).All(&accountList)
	if num <= 0 {
		// 没有相关数据
		return
	}

	var ids []string
	for _, account := range accountList {
		ids = append(ids, fmt.Sprintf("%d", account.Id))
	}

	orderM := models.Order{}
	sql := fmt.Sprintf("SELECT COUNT(id) AS total FROM `%s` WHERE user_account_id IN(%s) AND check_status = %d AND apply_time >= %d",
		orderM.TableName(), strings.Join(ids, ", "), types.LoanStatusReject, tools.GetUnixMillis()-3600000*24*days)
	err = o.Raw(sql).QueryRow(&num)
	if err != nil {
		logs.Debug("[ContactHasRejectLoanOderInLastMonth] err:", err)
	}
	if num > 0 {
		yes = true
	}

	return
}

// 第一联系人存在逾期中的借款订单
func ContactHasOverdueLoanOrder(mobile string) (yes bool, num int64, err error) {
	account, err := models.OneAccountBaseByMobile(mobile)
	if err != nil || account.Id <= 0 {
		// 不存在此用户
		return
	}

	orderM := models.Order{}
	o := orm.NewOrm()
	o.Using(orderM.UsingSlave())

	num, err = o.QueryTable(orderM.TableName()).Filter("user_account_id", account.Id).Filter("check_status", types.LoanStatusOverdue).Count()
	if err != nil || num <= 0 {
		// 不存在逾期订单
		return
	}

	yes = true

	return
}

// 近1个月内同一单位的申请人数
func SameCompanyApplyLoanOrderInLastMonth(company string) (num int64, err error) {
	if company == "" {
		return
	}

	accountProfile := models.AccountProfile{}
	o := orm.NewOrm()
	o.Using(accountProfile.UsingSlave())
	type profile struct {
		AccountId int64
	}
	var profileList []profile

	sql := fmt.Sprintf("SELECT account_id FROM `%s` WHERE company_name = '%s'", accountProfile.TableName(), tools.AddSlashes(company))
	num, err = o.Raw(sql).QueryRows(&profileList)
	if err != nil || num <= 0 {
		logs.Debug("[SameCompanyApplyLoanOrderInLastMonth] err:", err)
		return
	}

	var idsBox []string
	for _, pf := range profileList {
		idsBox = append(idsBox, fmt.Sprintf("%d", pf.AccountId))
	}

	orderM := models.Order{}
	sql = fmt.Sprintf("SELECT COUNT(DISTINCT(user_account_id)) AS total FROM `%s` WHERE user_account_id IN(%s) AND check_status > %d AND apply_time >= %d",
		orderM.TableName(), strings.Join(idsBox, ", "), types.LoanStatusSubmit, tools.GetUnixMillis()-3600000*24*30)
	err = o.Raw(sql).QueryRow(&num)
	if err != nil {
		logs.Debug("[SameCompanyApplyLoanOrderInLastMonth] err:", err)
	}

	return
}

// 近1个月内同一居住地址的申请人数
func SameResidentAddressApplyLoanOrderInLastMonth(address string) (num int64, err error) {
	accountProfile := models.AccountProfile{}
	o := orm.NewOrm()
	o.Using(accountProfile.Using())
	type profile struct {
		AccountId int64
	}
	var profileList []profile

	sql := fmt.Sprintf("SELECT account_id FROM `%s` WHERE resident_address = '%s'", accountProfile.TableName(), tools.AddSlashes(address))
	num, err = o.Raw(sql).QueryRows(&profileList)
	if err != nil || num <= 0 {
		logs.Error("[SameResidentAddressApplyLoanOrderInLastMonth] sql:", sql, ", err:", err)
		return
	}

	var idsBox []string
	for _, pf := range profileList {
		idsBox = append(idsBox, fmt.Sprintf("%d", pf.AccountId))
	}

	orderM := models.Order{}
	sql = fmt.Sprintf("SELECT COUNT(DISTINCT(user_account_id)) AS total FROM `%s` WHERE user_account_id IN(%s) AND check_status > %d AND apply_time >= %d",
		orderM.TableName(), strings.Join(idsBox, ", "), types.LoanStatusSubmit, tools.GetUnixMillis()-3600000*24*30)
	err = o.Raw(sql).QueryRow(&num)
	if err != nil {
		logs.Error("[SameResidentAddressApplyLoanOrderInLastMonth] sql:", sql, ", err:", err)
	}

	return
}

// [联系人/客户]历史最大逾期天数
func ContactsMaxOverdueDaysInLoanHistory(mobile string) (days int64, err error) {
	accountBase, err := models.OneAccountBaseByMobile(mobile)
	if err != nil {
		logs.Warning("[ContactsMaxOverdueDaysInLoanHistory] no data, mobile:", mobile, ", err:", err)
		return
	}

	o := orm.NewOrm()
	o.Using(accountBase.UsingSlave())

	type itemT struct {
		OrderId int64
	}
	var itemList []itemT
	sql := fmt.Sprintf("SELECT id AS order_id FROM %s WHERE user_account_id = %d AND is_overdue = %d",
		models.ORDER_TABLENAME, accountBase.Id, types.IsOverdueYes)
	total, err := o.Raw(sql).QueryRows(&itemList)
	if err != nil || total <= 0 {
		logs.Warning("[ContactsMaxOverdueDaysInLoanHistory] 用户不存在逾期订单, accountID:", accountBase.Id, ", err:", err)
		return
	}

	var orderIdBox []string
	for _, item := range itemList {
		orderIdBox = append(orderIdBox, tools.Int642Str(item.OrderId))
	}

	sql = fmt.Sprintf("SELECT MAX(overdue_days) AS days FROM %s WHERE order_id IN(%s) LIMIT 1",
		models.OVERDUE_CASE_TABLENAME, strings.Join(orderIdBox, ", "))
	err = o.Raw(sql).QueryRow(&days)

	return
}

// 联系人/客户存在进行中的逾期订单统计
func ContactsOverdueLoanOrderStat(mobile string) (total int64, err error) {
	accountBase, err := models.OneAccountBaseByMobile(mobile)
	if err != nil {
		logs.Warning("[ContactsHasOverdueLoanOrder] no data, mobile:", mobile, ", err:", err)
		return
	}

	o := orm.NewOrm()
	o.Using(accountBase.UsingSlave())

	sql := fmt.Sprintf("SELECT COUNT(id) AS total FROM %s WHERE user_account_id = %d AND check_status = %d LIMIT 1",
		models.ORDER_TABLENAME, accountBase.Id, types.LoanStatusOverdue)
	err = o.Raw(sql).QueryRow(&total)

	return
}

// 重复代码抽一下下,通过一组`account_id`统计逾期的人数, 并返回用户id
func overdueStatByAccountIds(o orm.Ormer, idsBox []string) (accountID []int64, total int64, err error) {
	sql := fmt.Sprintf(`SELECT DISTINCT(user_account_id) FROM %s
WHERE user_account_id IN(%s) AND check_status = %d`,
		models.ORDER_TABLENAME, strings.Join(idsBox, ", "), types.LoanStatusOverdue)
	total, err = o.Raw(sql).QueryRows(&accountID)
	return
}

// 有相同联系人的客户中,存在逾期的人数
func SameContactsCustomerOverdueStat(ct1, ct2 string, exclude int64) (accountIDs []int64, total int64, err error) {
	obj := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	type itemT struct {
		AccountId int64
	}
	var itemList []itemT
	sql := fmt.Sprintf(`SELECT account_id FROM %s
WHERE (contact1 = '%s' OR contact2 = '%s' OR contact1 = '%s' OR contact2 = '%s') AND account_id != %d`,
		models.ACCOUNT_PROFILE_TABLENAME,
		tools.AddSlashes(ct1), tools.AddSlashes(ct1), tools.AddSlashes(ct2), tools.AddSlashes(ct2), exclude)
	num, err := o.Raw(sql).QueryRows(&itemList)
	if err != nil || num <= 0 {
		logs.Warning("[SameContactsCustomerOverdueStat] has NO same contacts customer, accountID:", exclude, "ct1:", ct1, ", ct2:", ct2)
		return
	}

	var idsBox []string
	for _, item := range itemList {
		idsBox = append(idsBox, tools.Int642Str(item.AccountId))
	}
	accountIDs, total, err = overdueStatByAccountIds(o, idsBox)

	return
}

//FindCommonContact 找出n个客户的共有联系人
func FindCommonContact(accountIDs []int64) (commonContact []string) {
	obj := models.AccountProfile{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	type itemT struct {
		Contact1 string
		Contact2 string
	}
	var itemList []itemT

	//弥补strings.Join 指定是[]string的不足
	accountIDstr := tools.ArrayToString(accountIDs, ",")

	sql := fmt.Sprintf(`SELECT contact1, contact2 FROM %s WHERE account_id in( %s)`,
		models.ACCOUNT_PROFILE_TABLENAME, accountIDstr)
	num, err := o.Raw(sql).QueryRows(&itemList)
	if err != nil || num <= 0 {
		logs.Warning("[SameContactsCustomerOverdueStat] Can't find common contact , AccountIDs:", accountIDstr)
		return
	}
	var contact []interface{}
	for _, v := range itemList {
		contact = append(contact, v.Contact1)
		contact = append(contact, v.Contact2)
	}
	//获取切片中重复的值
	repeatArr := tools.SliceRepeatVal(contact)
	for _, v := range repeatArr {
		commonContact = append(commonContact, v.(string))
	}
	return
}

// 同居住地址的申请人当前逾期人数
func SameResidenceOverdueStat(residentCity, residentAddress string) (accountIDs []int64, total int64, err error) {
	obj := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	type itemT struct {
		AccountId int64
	}
	var itemList []itemT

	sql := fmt.Sprintf(`SELECT account_id FROM %s
WHERE resident_city = '%s' AND resident_address = '%s'`,
		models.ACCOUNT_PROFILE_TABLENAME,
		tools.AddSlashes(residentCity), tools.AddSlashes(residentAddress))
	num, err := o.Raw(sql).QueryRows(&itemList)
	if err != nil || num <= 0 {
		logs.Warning("[SameResidenceOverdueStat] has NO data, residentCity:", residentCity, ", residentAddress:", residentAddress, ", err:", err)
		return
	}

	var idsBox []string
	for _, item := range itemList {
		idsBox = append(idsBox, tools.Int642Str(item.AccountId))
	}
	accountIDs, total, err = overdueStatByAccountIds(o, idsBox)

	return
}

// 同单位申请人当前逾期人数
func SameCompanyOverdueStat(companyName string) (accountIDs []int64, total int64, err error) {
	if companyName == "" {
		return
	}

	obj := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	type itemT struct {
		AccountId int64
	}
	var itemList []itemT

	sql := fmt.Sprintf(`SELECT account_id FROM %s
WHERE company_name = '%s'`,
		models.ACCOUNT_PROFILE_TABLENAME, tools.AddSlashes(companyName))
	num, err := o.Raw(sql).QueryRows(&itemList)
	if err != nil || num <= 0 {
		logs.Warning("[SameCompanyOverdueStat] has NO data, companyName:", companyName, ", err:", err)
		return
	}

	var idsBox []string
	for _, item := range itemList {
		idsBox = append(idsBox, tools.Int642Str(item.AccountId))
	}
	accountIDs, total, err = overdueStatByAccountIds(o, idsBox)

	return
}

// 通过指定条件统计逾期订单数统计
func CustomerOverdueTotalStat(condBox map[string]interface{}) (total int64, err error) {
	obj := models.Order{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	sqlCount := fmt.Sprintf(`SELECT COUNT(id) AS total FROM %s`, models.ORDER_TABLENAME)
	//where := fmt.Sprintf("WHERE is_overdue = %d", types.IsOverdueYes)
	where := fmt.Sprintf("WHERE 1 = 1")

	// 累计逾期
	if _, ok := condBox["is_overdue"]; ok {
		where = fmt.Sprintf("%s AND is_overdue = %d", where, types.IsOverdueYes)
	}
	// 当前逾期
	if _, ok := condBox["check_status"]; ok {
		where = fmt.Sprintf("%s AND check_status = %d", where, types.LoanStatusOverdue)
	}
	// 最近三个月
	if _, ok := condBox["last_3_months"]; ok {
		last3Months := tools.NaturalDay(-90)
		where = fmt.Sprintf("%s AND apply_time >= %d", where, last3Months)
	}
	// 单个用户
	if v, ok := condBox["account_id"]; ok {
		where = fmt.Sprintf("%s AND user_account_id = %d", where, v.(int64))
	}
	// 一组用户
	if v, ok := condBox["account_ids_box"]; ok {
		where = fmt.Sprintf("%s AND user_account_id IN(%s)", where, strings.Join(v.([]string), ", "))
	}

	sql := fmt.Sprintf("%s %s LIMIT 1", sqlCount, where)
	err = o.Raw(sql).QueryRow(&total)

	return
}

// 同一银行账号关联客户数
func SameBankNoStat(bankNo string) (accountIDs []int64, total int64, err error) {
	obj := models.AccountProfile{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	var itemList []models.AccountProfile

	sql := fmt.Sprintf(`SELECT account_id FROM %s
WHERE bank_no = '%s'`,
		models.ACCOUNT_PROFILE_TABLENAME, bankNo)
	total, err = o.Raw(sql).QueryRows(&itemList)
	if err != nil || total <= 0 {
		return
	}

	for _, item := range itemList {
		accountIDs = append(accountIDs, item.AccountId)
	}

	return
}

// 同单位申请人 所有有过在贷用户的数量
func SameCompanyAllOrderStat(companyName string) (accountIDs []int64, total int64, err error) {
	if companyName == "" {
		return
	}

	obj := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	type itemT struct {
		AccountId int64
	}
	var itemList []itemT

	sql := fmt.Sprintf(`SELECT account_id FROM %s
WHERE company_name = '%s'`,
		models.ACCOUNT_PROFILE_TABLENAME, tools.AddSlashes(companyName))
	num, err := o.Raw(sql).QueryRows(&itemList)
	if err != nil || num <= 0 {
		logs.Warning("[SameCompanyOverdueStat] has NO data, companyName:", companyName, ", err:", err)
		return
	}

	var idsBox []string
	for _, item := range itemList {
		idsBox = append(idsBox, tools.Int642Str(item.AccountId))
	}

	sql = fmt.Sprintf(`SELECT DISTINCT(orders.user_account_id) FROM %s LEFT JOIN repay_plan
		on orders.id=repay_plan.order_id WHERE repay_plan.id >0 and orders.user_account_id IN (%s)`,
		models.ORDER_TABLENAME, strings.Join(idsBox, ","))

	logs.Info("[SameCompanyAllOrderStat] sql:%s", sql)
	total, err = o.Raw(sql).QueryRows(&accountIDs)
	return
}
