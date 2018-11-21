package dao

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

//! 非临时订单的总数
func AccountLoanOrderCount(accountId int64) (count int64, err error) {
	order := models.Order{}

	o := orm.NewOrm()
	o.Using(order.Using())

	count, err = o.QueryTable(order.TableName()).
		Filter("user_account_id", accountId).
		Filter("is_temporary", types.IsTemporaryNO).
		Count()

	return
}

//! 取最后一条非临时订单
func AccountLastLoanOrder(accountId int64) (order models.Order, err error) {
	order.UserAccountId = accountId

	o := orm.NewOrm()
	o.Using(order.Using())

	err = o.QueryTable(order.TableName()).
		Filter("user_account_id", accountId).
		Filter("is_temporary", 0).
		OrderBy("-id").
		Limit(1).
		One(&order)

	return
}

//! 取最后一条非临时的逾期订单
func AccountLastOverdueLoanOrder(accountId int64) (order models.Order, err error) {
	order.UserAccountId = accountId

	o := orm.NewOrm()
	o.Using(order.Using())

	err = o.QueryTable(order.TableName()).
		Filter("user_account_id", accountId).
		Filter("is_temporary", 0).
		Filter("check_status", types.LoanStatusOverdue).
		OrderBy("-id").
		Limit(1).
		One(&order)

	return
}

//! 取最后一条结清订单
func AccountLastLoanClearOrder(accountId int64) (order models.Order, err error) {
	order.UserAccountId = accountId

	o := orm.NewOrm()
	o.Using(order.Using())

	err = o.QueryTable(order.TableName()).
		Filter("user_account_id", accountId).
		Filter("is_temporary", 0).
		Filter("check_status", 8).
		OrderBy("-id").
		Limit(1).
		One(&order)

	return
}

//! 取用户最后一条临时订单,按最后更新时间反序
func AccountLastTmpLoanOrder(accountId int64) (order models.Order, err error) {
	order.UserAccountId = accountId

	o := orm.NewOrm()
	o.Using(order.Using())

	err = o.QueryTable(order.TableName()).
		Filter("user_account_id", accountId).
		Filter("is_temporary", 1).
		OrderBy("-utime").
		Limit(1).
		One(&order)

	return
}

func AccountLastTmpLoanOrderByCond(accountId int64, loan int64, period int) (order models.Order, err error) {
	order.UserAccountId = accountId

	o := orm.NewOrm()
	o.Using(order.Using())

	err = o.QueryTable(order.TableName()).
		Filter("user_account_id", accountId).
		Filter("is_temporary", 1).
		Filter("loan", loan).
		Filter("period", period).
		Filter("check_status", types.LoanStatusSubmit).
		OrderBy("-utime").
		Limit(1).
		One(&order)

	return
}

// 带分页的账户历史订单列表,只显示已经结清的订单
func AccountHistoryLoanOrder(accountId, offset int64) (list []models.Order, num int64, err error) {
	obj := models.Order{}

	o := orm.NewOrm()
	o.Using(obj.Using())

	where := ""
	if offset > 0 {
		where = fmt.Sprintf("AND id < %d", offset)
	}

	qb, _ := orm.NewQueryBuilder(tools.DBDriver())
	qb.Select("*").
		From(fmt.Sprintf("`%s`", obj.TableName())).
		Where("user_account_id = ? AND is_temporary = ? AND check_status = ? " + where).
		OrderBy("id").Desc().
		Limit(250)

	// 导出 SQL 语句
	sql := qb.String()

	num, err = o.Raw(sql, accountId, types.IsTemporaryNO, types.LoanStatusAlreadyCleared).QueryRows(&list)

	return
}

// 带分页的账户历史订单列表,只显示[已经结清]或[展期结清]的订单
func AccountHistoryLoanOrderV2(accountId, offset int64) (list []models.Order, num int64, err error) {
	obj := models.Order{}

	o := orm.NewOrm()
	o.Using(obj.Using())

	where := ""
	if offset > 0 {
		where = fmt.Sprintf("AND id < %d", offset)
	}

	qb, _ := orm.NewQueryBuilder(tools.DBDriver())
	qb.Select("*").
		From(fmt.Sprintf("`%s`", obj.TableName())).
		Where("user_account_id = ? AND is_temporary = ? AND check_status IN (?, ?) " + where).
		OrderBy("id").Desc().
		Limit(250)

	// 导出 SQL 语句
	sql := qb.String()

	num, err = o.Raw(sql, accountId, types.IsTemporaryNO, types.LoanStatusAlreadyCleared, types.LoanStatusRollClear).QueryRows(&list)

	return
}

/**
* 全部账户历史订单列表,根据订单状态集
*
**/
func AccountHistory(accountID int64, status []string) (list []models.Order, num int64, err error) {
	obj := models.Order{}

	o := orm.NewOrm()
	o.Using(obj.Using())

	statusStr := strings.Join(status, ",")
	statusStr = "(" + statusStr + ")"
	qb, _ := orm.NewQueryBuilder(tools.DBDriver())
	qb.Select("*").
		From(fmt.Sprintf("`%s`", obj.TableName())).
		Where("user_account_id = ? AND is_temporary = ? AND check_status in " + statusStr + "").
		OrderBy("id").Desc()

	// 导出 SQL 语句
	sql := qb.String()
	num, err = o.Raw(sql, accountID, types.IsTemporaryNO).QueryRows(&list)

	return
}

func AccountAllOrders(accountID int64) (list []models.Order, num int64, err error) {
	obj := models.Order{}

	o := orm.NewOrm()
	o.Using(obj.Using())

	qb, _ := orm.NewQueryBuilder(tools.DBDriver())
	qb.Select("*").
		From(fmt.Sprintf("`%s`", obj.TableName())).
		Where("user_account_id = ? AND is_temporary = ?").
		OrderBy("id").Desc()

	// 导出 SQL 语句
	sql := qb.String()
	num, err = o.Raw(sql, accountID, types.IsTemporaryNO).QueryRows(&list)

	return
}

// 查找客户同条件的临时订单
func OneTemporaryLoanOrder(accountId, loan int64, period int) (order models.Order, err error) {
	o := orm.NewOrm()
	o.Using(order.Using())

	err = o.QueryTable(order.TableName()).
		Filter("user_account_id", accountId).
		Filter("loan", loan).
		Filter("period", period).
		Filter("check_status", types.LoanStatusSubmit).
		Filter("is_temporary", types.IsTemporaryYes).
		OrderBy("-utime").
		Limit(1).
		One(&order)

	return
}

func HashRepeatLoanKey() string {
	return beego.AppConfig.String("repeat_loan")
}

// 先从redis中查询用户是否时复贷用户，若未查到再查数据库
func HasAlreadyClearedLoanOrder(accountId int64) (num int64, has bool, err error) {

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	hashName := HashRepeatLoanKey()
	hValue, err := storageClient.Do("HGET", hashName, accountId)
	if err != nil {
		logs.Error("[HasAlreadyClearedLoanOrder] no data, hashName: %s, accountId: %d, err: %#v", hashName, accountId, err)
		return
	} else if hValue == nil {
		// redis中没有,从db拿一份,再放到redis
		num, _ = models.GetClearedOrderNumByAccountId(accountId)
		logs.Info("[HasAlreadyClearedLoanOrder], redis no data, hashName: %s, accountId: %d, num: %d", hashName, accountId, num)
		if num > 0 {
			has = true
			storageClient.Do("HSET", hashName, accountId, num)
		}
	} else {
		n, _ := tools.Str2Int(string(hValue.([]byte)))
		logs.Info("[HasAlreadyClearedLoanOrder], redis hava data, hashName: %s, accountId: %d, num: %d", hashName, accountId, n)
		if n > 0 {
			has = true
		}
	}

	return
}

// 查查用户是否满足复贷条件
func IsRepeatLoan(accountId int64) (yes bool) {
	_, has, _ := HasAlreadyClearedLoanOrder(accountId)
	if has {
		yes = true
	}

	return
}

// IsOverdueAccount 用户是否有逾期记录
func IsOverdueAccount(accountId int64) bool {
	m := models.Order{}
	o := orm.NewOrm()
	o.Using(m.Using())

	num, _ := o.QueryTable(m.TableName()).Filter("user_account_id", accountId).Filter("is_overdue", types.IsOverdueYes).Count()
	if num > 0 {
		return true
	}
	return false
}

func CustomerOne(id int64) (obj models.AccountBase, err error) {
	obj.Id = id
	o := orm.NewOrm()
	o.Using(obj.Using())

	err = o.Read(&obj)

	return
}

func CustomerRiskOne(id int64) (obj models.CustomerRisk, err error) {
	obj.Id = id
	o := orm.NewOrm()
	o.Using(obj.Using())

	err = o.Read(&obj)

	return
}

//! 此方法返回模型的指针,因为模板层有几个方法是绑定在模型的指针上的,请调用者多加注意
func CustomerProfile(cid int64) (p *models.AccountProfile, err error) {
	obj := models.AccountProfile{AccountId: cid}
	o := orm.NewOrm()
	o.Using(obj.Using())

	err = o.Read(&obj)
	p = &obj

	return
}

func GetThirdpartyOne(tableName string, id int64) (obj models.ThirdpartyRecord, err error) {
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE id = %d", tableName, id)

	r := o.Raw(sql)
	r.QueryRow(&obj)

	return
}

func GetAccountProfile(cid int64) (obj models.AccountProfile, err error) {
	o := orm.NewOrm()
	o.Using(obj.Using())
	obj.AccountId = cid
	err = o.Read(&obj)
	return
}

// 最后一次活体认证
func CustomerLiveVerify(cid int64) (obj models.LiveVerify, err error) {
	o := orm.NewOrm()
	o.Using(obj.Using())

	err = o.QueryTable(obj.TableName()).Filter("account_id", cid).OrderBy("-id").Limit(1).One(&obj)

	return
}

// 上一次有效订单活体认证
func CustomerPrevLiveVerify(cid int64) (obj models.LiveVerify, err error) {

	order, _ := AccountLastLoanClearOrder(cid)
	o := orm.NewOrm()
	o.Using(obj.Using())
	err = o.QueryTable(obj.TableName()).Filter("account_id", cid).Filter("order_id", order.Id).OrderBy("id").Limit(1).One(&obj)
	return
}

//最后一次复贷照片
func GetLastReLoanImage(accountID int64) (obj models.ReLoanImage, err error) {
	o := orm.NewOrm()
	o.Using(obj.Using())
	err = o.QueryTable(obj.TableName()).Filter("user_account_id", accountID).OrderBy("-id").One(&obj)
	return
}

//最后一次风控针对复贷用户配置的额度账期
func GetLastAccountQuotaConf(accountID int64) (obj models.AccountQuotaConf, err error) {
	o := orm.NewOrm()
	o.Using(obj.Using())
	err = o.QueryTable(obj.TableName()).Filter("account_id", accountID).Filter("status", 1).OrderBy("-id").One(&obj)
	return
}

func SaveEsData(orderId int64, accountId int64, index string, data string) (int64, error) {
	m := models.EsData{
		OrderId:   orderId,
		AccountId: accountId,
		EsIndex:   index,
		Data:      data,
		Ctime:     tools.GetUnixMillis(),
		Utime:     tools.GetUnixMillis(),
	}

	o := orm.NewOrm()
	o.Using(m.Using())

	id, err := o.Insert(&m)

	return id, err
}

func GetMultiOptRecordByProductId(productId int64) (list []models.ProductOptRecord, err error) {
	o := orm.NewOrm()
	obj := models.ProductOptRecord{}
	o.Using(obj.Using())
	_, err = o.QueryTable(obj.TableName()).Filter("product_id", productId).
		OrderBy("-id").
		All(&list)
	return
}

func UserLatestOrderKey() string {
	return beego.AppConfig.String("latest_order")
}

// 先从redis中查询用户最新的订单以及订单是否上传了证件照
func IsUploadHoldPhoto(accountId int64) (has bool) {
	has = false
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	keyName := UserLatestOrderKey() + tools.Int642Str(accountId)
	hValue, err := storageClient.Do("GET", keyName)
	if err != nil {
		logs.Error("[dao.IsUploadHoldPhoto] no data, keyName: %s, accountId: %d, err: %#v", keyName, accountId, err)
		return
	} else if hValue == nil {
		// redis中没有,从db拿一份,再放到redis
		order, err := AccountLastTmpLoanOrder(accountId)
		if err != nil {
			logs.Warn("[dao.IsUploadHoldPhoto.AccountLastTmpLoanOrder] no data, accountId: %d, err: %#v", accountId, err)
			return
		}

		if 1 == order.IsUpHoldPhoto {
			has = true
		}

		//SET key value [EX seconds]
		orderJson, _ := json.Marshal(order)
		storageClient.Do("SET", keyName, orderJson, "EX", 3600*2)
	} else {
		order := models.Order{}
		err := json.Unmarshal(hValue.([]byte), &order)

		if err != nil {
			logs.Error("[dao.IsUploadHoldPhoto.json.Unmarshal]  accountId: %d, err: %#v", accountId, err)
			return
		}

		if 1 == order.IsUpHoldPhoto {
			has = true
		}
	}

	return
}

func GetFrozenTrans(orderId int64) (list []models.User_E_Trans, err error) {
	o := orm.NewOrm()
	obj := models.User_E_Trans{}
	o.Using(obj.Using())
	_, err = o.QueryTable(obj.TableName()).Filter("order_id", orderId).Filter("is_frozen", types.PayMoneyFrozen).Filter("pay_type", types.PayTypeMoneyIn).
		OrderBy("id").
		All(&list)
	return
}

func GetInOverdueCaseByOrderID(orderID int64) (oneCase models.OverdueCase, err error) {
	o := orm.NewOrm()
	o.Using(oneCase.Using())

	err = o.QueryTable(oneCase.TableName()).Filter("order_id", orderID).
		// Exclude("is_out", types.IsUrgeOutYes). ============此条产生NOT IN 效果极差，不介意再用此函数
		Filter("is_out__in", types.IsUrgeOutNo, types.IsUrgeOutFrozen).
		OrderBy("-id").
		Limit(1).
		One(&oneCase)

	return
}

//获取一条可操作的预减免申请
func GetPrereducedByOrderAndCaseID(orderID, caseID int64) (preReduced models.ReduceRecordNew, err error) {
	o := orm.NewOrm()
	o.Using(preReduced.Using())
	err = o.QueryTable(preReduced.TableName()).Filter("order_id", orderID).
		Filter("case_id", caseID).
		Filter("reduce_status", types.ReduceStatusNotValid).
		Filter("reduce_type", types.ReduceTypePrereduced).
		OrderBy("-id").
		Limit(1).
		One(&preReduced)
	return
}

func GetLastPrereducedByOrderid(orderId int64) (preReduced models.ReduceRecordNew, err error) {
	o := orm.NewOrm()
	o.Using(preReduced.Using())
	err = o.QueryTable(preReduced.TableName()).Filter("order_id", orderId).
		Filter("reduce_status", types.ReduceStatusNotValid).
		Filter("reduce_type", types.ReduceTypePrereduced).
		OrderBy("-id").Limit(1).One(&preReduced)
	return
}

// 按照银行名字查找va账户
func GetActiveEaccountWithBankName(accountId int64) (userEAccount models.User_E_Account, err error) {

	profile, _ := models.OneAccountProfileByAccountID(accountId)
	if len(profile.BankName) == 0 {
		err = fmt.Errorf("[GetActiveEaccountWithBankName] bankname empty. accountId:%d", accountId)
		return
	}

	vaCompanyType, err := PriorityThirdpartyVACreate(profile.BankName)
	if err != nil {
		logs.Error("[GetActiveEaccountWithBankName] PriorityThirdpartyVACreate err.:%v accountId:%d profile:%#v", err, accountId, profile)
	}
	userEAccount, err = models.GetLastestActiveEAccountByVacompanyType(accountId, vaCompanyType)
	if err == nil {
		return
	} else if err == orm.ErrNoRows {
		// 老用户可能没有 新切过来的va
		userEAccount, err = models.GetLastestActiveEAccount(accountId)
	} else {
		logs.Error("[GetActiveEaccountWithBankName] GetLastestActiveEAccountByVacompanyType err:%v accountId:%d", err, accountId)
	}
	return
}

// 查找va账户(先查找最新的va, 然后在对比是否满足)
func GetActiveUserEAccount(accountId int64) (userEAccount models.User_E_Account, err error) {

	profile, _ := models.OneAccountProfileByAccountID(accountId)
	if len(profile.BankName) == 0 {
		err = fmt.Errorf("[GetActiveUserEAccount] bankname empty. accountId:%d", accountId)
		return
	}

	if len(profile.RepayBankCode) <= 0 {
		userEAccount, err = models.GetLastestActiveEAccount(accountId)
	} else {
		userEAccount, err = models.GetLastestActiveEAccountByRepayBank(accountId, profile.RepayBankCode)
	}
	if err != nil {
		//logs.Error("[GetActiveUserEAccount] GetLastestActiveEAccount err:%v accountId:%d", err, accountId)
		return
	}

	vaCompanyType, err := PriorityThirdpartyVACreate(profile.BankName)
	if err != nil {
		logs.Error("[GetActiveUserEAccount] PriorityThirdpartyVACreate err.:%v accountId:%d profile:%#v", err, accountId, profile)
	}

	var userEAccountTmp models.User_E_Account
	userEAccountTmp, err = models.GetLastestActiveEAccountByVacompanyType(accountId, vaCompanyType)
	if err == nil {
		if userEAccount.BankCode == userEAccountTmp.BankCode && userEAccount.VaCompanyCode != userEAccountTmp.VaCompanyCode {
			// 当银行简码相同并且第三方支付不相同时，更新 userEAccount(只有在过风控生成va时，会生成两个va，选择一个合适的)
			userEAccount = userEAccountTmp
		}
	} else {
		err = nil
	}

	return
}

/**
优先va创建逻辑
*/
func PriorityThirdpartyVACreate(bankName string) (thirdPartyPay int, err error) {
	one, err := models.OneBankInfoByFullName(bankName)
	if err != nil {
		logs.Error("[PriorityThirdpartyVACreate] OneBankInfoByFullName err:%v. check bank name:%s", err, bankName)
		return
	}
	thirdPartyPay = one.RepayCompanyCode
	logs.Debug("The priority Repay bank info:", thirdPartyPay)
	return

}
