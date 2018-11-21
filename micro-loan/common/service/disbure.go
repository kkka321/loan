package service

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty/doku"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func CreateVirtualAccountAll(userAccountId int64, orderId int64) (err error) {

	//	1.准备基本信息
	accountBase, err := models.OneAccountBaseByPkId(userAccountId)
	if err != nil || accountBase.Id <= 0 {
		err = fmt.Errorf("[CreateVirtualAccountAll] user_account_id相关的account_base记录不存在！user_account_id is: %d", userAccountId)
		return err
	}
	accountProfile, err := dao.CustomerProfile(userAccountId)
	one, err := models.OneBankInfoByFullName(accountProfile.BankName)
	if err != nil {
		logs.Error("[CreateVirtualAccountAll] OneBankInfoByFullName err:%v. check bank name:%s userAccountId:%d", err, accountProfile.BankName, userAccountId)
		return
	}

	//2.创建需要的Va
	for companyType := range bankFiledNameMap {
		companyName := types.FundCodeNameMap()[companyType]
		//3. 检查是否已创建
		_, err = models.GetLastestActiveEAccountByVacompanyType(userAccountId, companyType)
		if err == nil {
			logs.Info("[CreateVirtualAccountAll] already has va. companyType:%d", companyType)
			continue
		}

		err = CreateVirtualAccountsV2(accountBase, *accountProfile, orderId, companyName, companyType, one)
		if err != nil {
			logs.Error("[CreateVirtualAccountAll] CreateVirtualAccountsV2 err:%v userAccountId:%d companyType:%d", err, userAccountId, companyType)
		}
	}
	return nil
}

func CreateVirtualAccountsV2(accountBase models.AccountBase, accountProfile models.AccountProfile, orderId int64, companyName string, vaCompanyType int, bankInfo models.BanksInfo) (err error) {

	var datas = make(map[string]interface{})
	datas["bank_name"] = accountProfile.BankName
	datas["account_id"] = accountBase.Id
	datas["account_name"] = accountBase.Realname
	datas["company_name"] = companyName
	datas["order_id"] = orderId
	datas["banks_info"] = bankInfo

	err = CreateVirtualAccountHandler(datas, vaCompanyType)

	return
}

func CreateVirtualAccountHandler(datas map[string]interface{}, vaCompanyType int) (err error) {
	payApi, err := CreatePaymentApi(vaCompanyType, datas)
	if err != nil {
		logs.Error("[CreateVirtualAccountsV2] CreatePaymentApi err:%v vaCompanyType:%d datas:%#v", err, vaCompanyType, datas)
		return err
	}

	resJson, err := payApi.CreateVirtualAccount(datas)
	if err != nil {
		logs.Error("[CreateVirtualAccountsV2] CreateVirtualAccount err:%v vaCompanyType:%d datas:%#v", err, vaCompanyType, datas)
		return err
	}

	err = payApi.CreateVirtualAccountResponse(resJson, datas)
	if err != nil {
		logs.Error("[CreateVirtualAccountsV2] CreateVirtualAccountResponse err:%v vaCompanyType:%d resJson:%#v datas:%#v", err, vaCompanyType, resJson, datas)
		return err
	}

	return err
}

func DisplayVAInfo(accountId int64) (eAccountDesc string) {

	eAccount, err := dao.GetActiveEaccountWithBankName(accountId)
	//eAccount, err := models.GetLastestActiveEAccount(accountId)
	if err == nil {
		bankCode := eAccount.BankCode
		if eAccount.VaCompanyCode == types.DoKu {
			bankCode = doku.DoKuVaBankCodeTransform(eAccount.BankCode)
		}
		eAccountDesc = fmt.Sprintf("%s %s", bankCode, eAccount.EAccountNumber)
	} else {
		logs.Warn("[DisplayVAInfo] GetActiveEaccountWithBankName err:%v accountID:%d", err, accountId)
	}

	return eAccountDesc
}

func DisplayBankCode(bankCode string) (bankCodeList []string) {
	for _, v := range types.MobileBankCodeMap() {
		if bankCode == v {
			continue
		}
		bankCodeList = append(bankCodeList, v)
	}
	return
}

func DisplayBankCodeV2(bankCode string) (bankCodeList []string) {
	for _, v := range types.MobileBankCodeMapV2() {
		if bankCode == v {
			continue
		}
		bankCodeList = append(bankCodeList, v)
	}
	return
}

func DisplayVAInfoV2(accountId int64) (bankCode, eAccountDesc string) {

	eAccount, err := dao.GetActiveUserEAccount(accountId)
	if err == nil {
		bankCode = eAccount.BankCode
		if eAccount.VaCompanyCode == types.DoKu {
			bankCode = doku.DoKuVaBankCodeTransform(eAccount.BankCode)
		}
		eAccountDesc = fmt.Sprintf("%s %s", bankCode, eAccount.EAccountNumber)
	} else {
		logs.Warn("[DisplayVAInfoV2] GetActiveUserEAccount err:%v accountID:%d", err, accountId)
	}

	return
}

func ModifyRepayBankAndVA(accountId int64, repayBankCode string) (eAccountNumber string, err error) {

	// 由"还款银行简码"获取"还款银行"对应的第三方支付公司编码
	banksInfo, err := models.OneBankInfoByXenditBrevity(repayBankCode)
	if err != nil {
		logs.Error("[ModifyRepayBankAndVA] OneBankInfoByXenditBrevity repayBankCode no valid. accountId:%d, xenditBrevityName:%s, err:%s", accountId, repayBankCode, err)
		return
	}

	repayVaCompanyCode := types.RepayVaCompanyCodeMap()[repayBankCode]

	// 第三个支付如果是doku，需要将上传的‘还款银行简码’更改一下（上传的银行简码都是xendit的）
	if repayVaCompanyCode == types.DoKu {
		repayBankCode = doku.XenditVaBankCodeTransform(repayBankCode)
	}

	var userEAccount models.User_E_Account
	userEAccount, err = models.GetLastestActiveEAccountByRepayBankAndVacompanyType(accountId, repayBankCode, repayVaCompanyCode)
	if err != nil {
		userEAccount, err = models.GetLastestActiveEAccountByBankAndVacompanyType(accountId, repayBankCode, repayVaCompanyCode)
		if err != nil {
			userEAccount, err = GenerateVaAndSave(accountId, repayBankCode, repayVaCompanyCode, banksInfo)
			if err != nil {
				return
			}
		} else {
			if len(userEAccount.RepayBankCode) <= 0 {
				userEAccount.RepayBankCode = repayBankCode
				userEAccount.Utime = tools.GetUnixMillis()
				userEAccount.UpdateEAccount(&userEAccount)
			} else {
				userEAccount, err = GenerateVaAndSave(accountId, repayBankCode, repayVaCompanyCode, banksInfo)
				if err != nil {
					return
				}
			}
		}
	}

	eAccountNumber = userEAccount.EAccountNumber

	// 升级profile
	accountProfile, _ := models.OneAccountProfileByAccountID(accountId)
	origin := accountProfile

	accountProfile.AccountId = accountId
	accountProfile.RepayBankCode = repayBankCode
	accountProfile.Utime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(accountProfile.Using())

	num, _ := o.Update(&accountProfile, "repay_bank_code")
	if num != 1 {
		logs.Error("Update account profile repay bank code has wrong. accountId:", accountId, ", err:", err)
	} else {
		// 写操作日志
		models.OpLogWrite(accountId, accountId, models.OpUserInfoUpdate, accountProfile.TableName(), origin, accountProfile)
	}

	return
}

func GenerateVaAndSave(accountId int64, repayBankCode string, repayVaCompanyCode int, banksInfo models.BanksInfo) (userEAccount models.User_E_Account, err error) {

	var datas = make(map[string]interface{})
	datas["account_id"] = accountId
	accountBase, _ := models.OneAccountBaseByPkId(accountId)
	datas["account_name"] = accountBase.Realname
	datas["banks_info"] = banksInfo
	datas["company_name"] = types.FundCodeNameMap()[repayVaCompanyCode]
	datas["account_id"] = accountId
	order, _ := dao.AccountLastLoanOrder(accountId)
	datas["order_id"] = order.Id

	err = CreateVirtualAccountHandler(datas, repayVaCompanyCode)
	if err != nil {
		logs.Error("[GenerateVaAndSave] CreateVirtualAccountHandler userAccountId:%d, companyType:%d, err:%v ", accountId, repayVaCompanyCode, err)
		return
	}

	for i := 0; i < 5; i++ {
		time.Sleep(1000 * time.Millisecond)
		userEAccount, err = models.GetLastestActiveEAccountByRepayBankAndVacompanyType(accountId, repayBankCode, repayVaCompanyCode)
		if err == nil {
			break
		}
	}
	if err != nil {
		logs.Warn("[GenerateVaAndSave] GetLastestActiveEAccountByRepayBankAndVacompanyType userAccountId:%d, companyType:%d, err:%v ", accountId, repayVaCompanyCode, err)
	}

	return
}

func IsCompanySupport(companyType int, info models.BanksInfo) bool {

	ret := false
	switch companyType {
	case types.Xendit:
		{
			if len(info.XenditBrevityName) > 0 {
				ret = true
			}
		}
	case types.Bluepay:
		{
			if len(info.BluepayBrevityName) > 0 {
				ret = true
			}
		}
	case types.DoKu:
		{
			if len(info.DokuFullName) > 0 &&
				len(info.DokuBrevityName) > 0 &&
				len(info.DokuBrevityId) > 0 {
				ret = true
			}
		}
	default:
		{
			logs.Error("[isCompanySupport] unknow type:%d", companyType)
		}
	}

	return ret
}

func LoanCompany(orderId int64, bankName string) (current int, support []int) {
	orderExt, _ := models.GetOrderExt(orderId)
	bankInfo, err := models.OneBankInfoByFullName(bankName)
	if err != nil {
		logs.Error("[LoanCompany] OneBankInfoByFullName. err:%v bankName:%s orderId:%d", err, bankName, orderId)
		return
	}

	if orderExt.SpecialLoanCompany > 0 {
		current = orderExt.SpecialLoanCompany
	} else {
		current = bankInfo.LoanCompanyCode
	}

	if len(bankInfo.XenditBrevityName) > 0 {
		support = append(support, types.Xendit)
	}

	if len(bankInfo.DokuBrevityName) > 0 &&
		len(bankInfo.DokuFullName) > 0 &&
		len(bankInfo.DokuBrevityId) > 0 {
		support = append(support, types.DoKu)
	}

	return
}
func GetAccountVas(accountId int64) ([]string, error) {
	vas, err := models.GetEAccountNumberByAccountId(accountId)
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[GetAccountVas] models GetEAccountNumberByAccountId err :%v ", err)
		return nil, err
	}
	vass := make([]string, 0)
	for _, v := range vas {
		vaStr := fmt.Sprintf("%s %s", v.BankCode, v.EAccountNumber)
		vass = append(vass, vaStr)
	}
	if len(vass) == 0 {
		vass = []string{}
	}
	return vass, nil
}
