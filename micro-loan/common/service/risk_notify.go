package service

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/cerror"
	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// Notify 风控主动通知
func Notify(accessToken string, reqTime int64) (code cerror.ErrCode, err error) {

	code = cerror.CodeSuccess
	//判断请求时间
	timeNow := tools.TimeNow()
	logs.Debug("[Risk Notify] reqTime:", reqTime, "UTC Time:", timeNow)
	if (timeNow - reqTime) > 5 {
		//400119 无效请求数据
		code = cerror.InvalidRequestData
	}
	//获取用户信息
	accountToken, _ := models.GetAccessTokenInfo(accessToken)
	accountID := accountToken.AccountId
	if accountID == 0 {
		logs.Info("[Risk Notify] happend error:", err, "accessToken: ", accessToken, " AccountID: ", accountID)
		//400114 无效token
		code = cerror.InvalidAccessToken
	}
	//请求记录更新
	riskNotifyModel := models.RiskNotify{}
	riskNotifyModel.AccessToken = accessToken
	riskNotifyModel.ReqTime = reqTime
	riskNotifyModel.AccountID = accountID
	riskNotifyModel.Ctime = tools.GetUnixMillis()

	id, err := models.InsertRiskNotify(riskNotifyModel)
	if err != nil || id == 0 {
		logs.Error("[Risk Notify] InserRiskNotify happend error:", err, "AccountID", accountID)
		code = cerror.InvalidAccessToken
	}

	if accountID > 0 && code == 0 {
		// 写待处理队列
		storageClient := storage.RedisStorageClient.Get()
		defer storageClient.Close()
		queueName := beego.AppConfig.String("risk_notify")
		storageClient.Do("lpush", queueName, accountID)
		logs.Debug("[Risk Notify] write queue success, queueName:", queueName, "accountID:", accountID)
	}

	return
}

// QuotaConf 风控账户额度配置
func QuotaConf(accountID, quota, quotaVisable, accountPeriod, isPhoneVerify int64) (code cerror.ErrCode, err error) {
	code = cerror.CodeSuccess

	//如果存在某个accountID的数据，修改该为无效
	accountQuotaConfModel, _ := models.OneAccountQuotaConfByAccountID(accountID)
	if accountQuotaConfModel.ID > 0 {
		utime := tools.GetUnixMillis()
		obj := models.AccountQuotaConf{
			ID:     accountQuotaConfModel.ID,
			Status: 0,
			Utime:  utime,
		}
		cols := []string{"Status", "Utime"}
		models.OrmUpdate(&obj, cols)
	}
	accountQuotaConfInsert := models.AccountQuotaConf{
		AccountID:     accountID,
		Quota:         quota,
		QuotaVisable:  quotaVisable,
		AccountPeriod: accountPeriod,
		IsPhoneVerify: isPhoneVerify,
		Status:        1,
		IsDefault:     0,
		Ctime:         tools.GetUnixMillis(),
		Utime:         tools.GetUnixMillis(),
	}
	num, err := models.OrmInsert(&accountQuotaConfInsert)
	if err != nil && num == 0 {
		logs.Error("[QuotaConf] Insert happend error:", err)
		code = cerror.InvalidRequestData
		return
	}

	return
}

// InsertDefaultQuotaConf 判断是否需要写入复贷默认额度配置，需要则写入
func InsertDefaultQuotaConf(accountID int64) {
	accountQuotaConfModel, _ := models.OneAccountQuotaConfByAccountID(accountID)
	if accountQuotaConfModel.ID == 0 {
		models.InsertDefaultConf(accountID)
	}
}

// QueryThirdParty 查询第三方,目前只有同盾

func QueryThirdParty(accountID int64, sourceFrom, sourceCode string) (result string, serviceTime int64) {

	switch sourceFrom {
	case "tongdun":
		{
			tongdunModels, err := models.GetOneAC(accountID, sourceCode)
			if err != nil {
				return
			}
			result = tongdunModels.TaskData
			serviceTime = tongdunModels.NotifyTime
		}
	default:
		{
			return
		}
	}
	return
}

func QueryRiskValue(accountId, orderId int64, section string) (interface{}, error) {
	data := make(map[string]interface{})

	switch section {
	case "user_info":
		return QueryBaseInfo(accountId, orderId)
	case "black_info":
		return QueryBlacklistInfo(accountId, orderId)
	case "loan_info":
		return QueryLoanInfo(accountId, orderId)
	case "fraud":
		return QueryFraudInfo(accountId, orderId)
	}

	return data, fmt.Errorf("unexcept section")
}

func QueryBaseInfo(accountId, orderId int64) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	accountBase, err := models.OneAccountBaseByPkId(accountId)
	if err != nil {
		return data, fmt.Errorf("account not exist")
	}

	data["user_age"], _ = CustomerAge(accountBase.Identity)

	data["user_identity_check_invalid"] = tools.ThreeElementExpression(!IdentityVerify(accountId), 1, 0)

	unservicedAreaConf, _ := GetUnservicedAreaConf()
	data["user_identity_address_invalid"] = tools.ThreeElementExpression(unservicedAreaConf[accountBase.ThirdProvince], 1, 0)

	data["user_company_address_invalid"] = tools.ThreeElementExpression(unservicedAreaConf[accountBase.ThirdProvince], 1, 0)

	accountProfile, _ := dao.CustomerProfile(accountId)
	companyProvince, err := accountProfile.CompanyProvince()
	if err != nil || unservicedAreaConf[companyProvince] {
		data["user_company_address_invalid"] = 1
	} else {
		data["user_company_address_invalid"] = 0
	}

	return data, nil
}

func QueryBlacklistInfo(accountId, orderId int64) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	accountBase, err := models.OneAccountBaseByPkId(accountId)
	if err != nil {
		return data, fmt.Errorf("account not exist")
	}

	accountProfile, _ := dao.CustomerProfile(accountId)

	yes, _ := models.IsBlacklistIdentity(accountBase.Identity)
	data["user_identity_hit_black"] = tools.ThreeElementExpression(yes, 1, 0)

	yes, _ = models.IsBlacklistMobile(accountBase.Mobile)
	data["user_mobile_hit_black"] = tools.ThreeElementExpression(yes, 1, 0)

	yes, _ = models.IsBlacklistMobile(accountProfile.Contact1)
	data["user_contact_mobile_hit_black"] = tools.ThreeElementExpression(yes, 1, 0)

	address := accountProfile.ResidentCity + "," + accountProfile.ResidentAddress
	yes, _ = models.IsBlacklistItem(types.RiskItemResidentAddress, address)
	data["user_resident_address_hit_black"] = tools.ThreeElementExpression(yes, 1, 0)

	yes, _ = models.IsBlacklistItem(types.RiskItemCompany, accountProfile.CompanyName)
	data["user_company_hit_black"] = tools.ThreeElementExpression(yes, 1, 0)

	companyAddress := accountProfile.CompanyCity + "," + accountProfile.CompanyAddress
	yes, _ = models.IsBlacklistItem(types.RiskItemCompanyAddress, companyAddress)
	data["user_company_address_hit_black"] = tools.ThreeElementExpression(yes, 1, 0)

	clientInfo, _ := OrderClientInfo(orderId)
	yes, _ = models.IsBlacklistItem(types.RiskItemIMEI, clientInfo.Imei)
	data["user_device_hit_black"] = tools.ThreeElementExpression(yes, 1, 0)

	yes, _ = models.IsBlacklistItem(types.RiskItemIP, clientInfo.IP)
	data["user_ip_hit_black"] = tools.ThreeElementExpression(yes, 1, 0)

	return data, nil
}

func QueryLoanInfo(accountId, orderId int64) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	_, err := models.OneAccountBaseByPkId(accountId)
	if err != nil {
		return data, fmt.Errorf("account not exist")
	}

	accountProfile, _ := dao.CustomerProfile(accountId)

	riskCtlE001, _ := config.ValidItemInt("risk_ctl_E001")
	_, num, _ := ContactHasRejectLoanOderInDays(accountProfile.Contact1, int64(riskCtlE001))
	data["user_contact1_exists_reject_orders_1m"] = num

	_, num, _ = ContactHasOverdueLoanOrder(accountProfile.Contact1)
	data["user_contact1_exists_overdoing_orders"] = num

	num, _ = SameCompanyApplyLoanOrderInLastMonth(accountProfile.CompanyName)
	data["user_company_apply_orders_1m"] = num

	_, num, _, _ = SameContactApplyLoanOrderInLastMonth(1, accountProfile.Contact1, accountProfile.Contact2)
	data["user_contact_apply_orders_1m"] = num

	_, num, _, _ = SameContactApplyLoanOrderInLast3Month(1, accountProfile.Contact1, accountProfile.Contact2)
	data["user_contact_apply_orders_3m"] = num

	_, num, _ = SameCompanyApplyLoanOrderInLast3Month(1, accountProfile.CompanyName)
	data["user_company_apply_orders_3m"] = num

	num, _ = ContactsMaxOverdueDaysInLoanHistory(accountProfile.Contact1)
	data["user_contact1_max_overdue"] = num

	num, _ = ContactsOverdueLoanOrderStat(accountProfile.Contact2)
	data["user_contact2_exists_overdoing_orders"] = num

	num, _ = ContactsMaxOverdueDaysInLoanHistory(accountProfile.Contact2)
	data["user_contact2_max_overdue"] = num

	_, num, _ = SameContactsCustomerOverdueStat(accountProfile.Contact1, accountProfile.Contact2, accountProfile.AccountId)
	data["user_contact_apply_overdue_accounts"] = num

	_, num, _ = SameCompanyOverdueStat(accountProfile.CompanyName)
	data["user_company_apply_overdue_accounts"] = num

	_, num, _ = SameBankNoStat(accountProfile.BankNo)
	data["user_bank_related_accounts"] = num

	condBox := map[string]interface{}{
		"is_overdue":    true,
		"last_3_months": true,
		"account_id":    accountId,
	}
	num, _ = CustomerOverdueTotalStat(condBox)
	data["user_overdue_orders_3m"] = num

	return data, nil
}

func QueryFraudInfo(accountId, orderId int64) (interface{}, error) {
	data := FraudRequestInfo{}

	orderData, err := models.GetOrder(orderId)
	if err != nil {
		return data, fmt.Errorf("order not exist")
	}

	accountBase, err := models.OneAccountBaseByPkId(accountId)
	if err != nil {
		return data, fmt.Errorf("account not exist")
	}

	clientInfo, _ := OrderClientInfo(orderId)

	FillFantasyFraudRequest(&data, &orderData, &accountBase, &clientInfo)

	return data, nil
}
