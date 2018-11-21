package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// GetCustomerRiskItemVal 根据客户id, 风险项编号 RiskItemEnum , 获取用户风险项的值
func GetCustomerRiskItemVal(cid int64, riskItemNum types.RiskItemEnum) (map[types.RiskItemEnum]interface{}, error) {
	var riskItemMapVal = map[types.RiskItemEnum]interface{}{}
	notFoundCustomerRiskValErr := fmt.Errorf("Unexpected RiskItemNum: %d, cid: %d", int(riskItemNum), cid)

	// 顺便带出, 当前表中其他可用 itemVal的值, 减少冗余查询次数
	switch riskItemNum {
	case types.RiskItemMobile, types.RiskItemIdentity:
		// do account_base info query
		ab, err := models.OneAccountBaseByPkId(cid)
		if err != nil {
			return riskItemMapVal, err
		}
		if ab.Id <= 0 {
			return riskItemMapVal, notFoundCustomerRiskValErr
		}
		riskItemMapVal[types.RiskItemMobile] = ab.Mobile
		riskItemMapVal[types.RiskItemIdentity] = ab.Identity
		return riskItemMapVal, nil
	case types.RiskItemResidentAddress, types.RiskItemCompany, types.RiskItemCompanyAddress:
		// do account_profile info query
		ap, err := models.OneAccountProfileByAccountID(cid)
		if err != nil {
			return riskItemMapVal, err
		}
		riskItemResidentAddress := ""
		if ap.ResidentCity != "" && ap.ResidentAddress != "" {
			riskItemResidentAddress = ap.ResidentCity + "," + ap.ResidentAddress
		}
		riskItemCompanyAddress := ""
		if ap.CompanyCity != "" && ap.CompanyAddress != "" {
			riskItemCompanyAddress = ap.CompanyCity + "," + ap.CompanyAddress
		}
		riskItemMapVal[types.RiskItemResidentAddress] = riskItemResidentAddress
		riskItemMapVal[types.RiskItemCompany] = ap.CompanyName
		riskItemMapVal[types.RiskItemCompanyAddress] = riskItemCompanyAddress
		return riskItemMapVal, nil
	case types.RiskItemIMEI, types.RiskItemIP:
		c, err := models.OneLastClientInfoByRelatedID(cid)
		if err != nil {
			return riskItemMapVal, err
		}
		if c.Id >= 0 {
			return riskItemMapVal, notFoundCustomerRiskValErr
		}
		riskItemMapVal[types.RiskItemIMEI] = c.Imei
		riskItemMapVal[types.RiskItemIP] = c.IP
		return riskItemMapVal, nil
	default:
		return nil, fmt.Errorf("Unexpected RiskItemNum: %d", int(riskItemNum))
	}
}

// AddCustomerRisk 新增提报风险项
func AddCustomerRisk(cid, opUID int64, riskItem types.RiskItemEnum, riskType types.RiskTypeEnum, reason types.RiskReason, riskValue, remark string, riskStatus types.RiskStatusEnum, reviewTime int64, orderIds string, userAccoutIds string) (int64, error) {

	crisk := models.CustomerRisk{}
	o := orm.NewOrm()
	o.Using(crisk.Using())
	qs := o.QueryTable(crisk.TableName())
	cond := orm.NewCondition()
	cond = cond.And("risk_value", riskValue)
	cond = cond.And("is_deleted", 0)
	count, _ := qs.SetCond(cond).Count()

	if count > 0 {
		logs.Warn("[AddCustomerRisk] The cid:%d - riskVal:%s data already exist in blacklist", cid, riskValue)
		return 0, fmt.Errorf("[AddCustomerRisk] The cid:%d - riskVal:%s data already exist in blacklist", cid, riskValue)
	} else {
		obj := models.CustomerRisk{
			CustomerId:     cid,
			RiskItem:       riskItem,
			RiskType:       riskType,
			RiskValue:      riskValue,
			Reason:         reason,
			ReportRemark:   remark,
			OpUid:          opUID,
			Status:         riskStatus,
			Ctime:          tools.GetUnixMillis(),
			Utime:          tools.GetUnixMillis(),
			ReviewTime:     reviewTime,
			OrderIds:       orderIds,
			UserAccountIds: userAccoutIds,
		}
		o.Using(obj.Using())
		id, err := o.Insert(&obj)
		if err != nil {
			logs.Error("[AddCustomerRisk] insert has wrong. CustomerRisk:%v, err:%v", obj, err)
		}

		return id, err
	}
}

// ReviewCustomerRisk 审核提报风险项
func ReviewCustomerRisk(id int64, s types.RiskStatusEnum, remark string, opUID int64) (affected int64, err error) {
	original := models.CustomerRisk{
		Id: id,
	}

	o := orm.NewOrm()
	o.Using(original.Using())

	err = o.Read(&original)
	if err != nil {
		logs.Error("Review customer risk:", err)
		return
	}

	t := tools.GetUnixMillis()
	obj := models.CustomerRisk{
		Id:           id,
		Status:       s,
		ReviewRemark: remark,
		Utime:        t,
		ReviewTime:   tools.GetUnixMillis(),
	}

	affected, err = o.Update(&obj, "Status", "ReviewRemark", "Utime", "ReviewTime")
	if err != nil {
		logs.Error("Review customer risk:", err)
	}
	models.OpLogWrite(opUID, id, models.OPCodeCustomerRisk, obj.TableName(), original, obj)

	return
}

// RelieveCustomerRisk 解除提报风险项
func RelieveCustomerRisk(id int64, reason types.RiskRelieveReason, remark string, opUID int64) (affected int64, err error) {

	original := models.CustomerRisk{
		Id: id,
	}

	o := orm.NewOrm()
	o.Using(original.Using())

	err = o.Read(&original)
	if err != nil {
		logs.Error("Relieve customer risk:", err)
		return
	}
	obj := models.CustomerRisk{
		Id:            id,
		RelieveReason: reason,
		IsDeleted:     1,
		RelieveRemark: remark,
		Utime:         tools.GetUnixMillis(),
		RelieveTime:   tools.GetUnixMillis(),
	}

	affected, err = o.Update(&obj, "RelieveReason", "IsDeleted", "RelieveRemark", "Utime", "RelieveTime")
	if err != nil {
		logs.Error("Relieve customer risk:", err)
	}
	models.OpLogWrite(opUID, id, models.OPCodeCustomerRisk, obj.TableName(), original, obj)

	return
}

func ImportBlacklist(typeList []int, dataList [][]string, opid int64) {

	type acctRiskInfo struct {
		accountId int64
		reason    int
		mark      string
	}
	logs.Info("[ImportBlacklist] typeSize:%d, dataSize:%d", len(typeList), len(dataList))

	m := models.CustomerRisk{}
	o := orm.NewOrm()
	o.Using(m.Using())

	var list []models.CustomerRisk
	o.QueryTable(m.TableName()).All(&list)

	var newDatas map[types.RiskItemEnum]map[string]acctRiskInfo = make(map[types.RiskItemEnum]map[string]acctRiskInfo)
	var oldDatas map[types.RiskItemEnum]map[string]bool = make(map[types.RiskItemEnum]map[string]bool)
	for _, v := range list {
		if v.IsDeleted == 1 {
			continue
		}

		if v.RiskType != 1 {
			continue
		}

		if oldDatas[v.RiskItem] == nil {
			oldDatas[v.RiskItem] = make(map[string]bool)
		}
		oldDatas[v.RiskItem][v.RiskValue] = true
	}

	for _, v := range dataList {
		if len(v) < 3 {
			continue
		}

		accountStr := strings.Trim(v[0], " ")
		accountId, _ := tools.Str2Int64(accountStr)

		reasonstr := strings.Trim(v[1], " ")
		reason, _ := tools.Str2Int(reasonstr)

		remark := strings.Trim(v[2], " ")
		remark = strings.Trim(remark, "\r\n")

		account, err := models.OneAccountBaseByPkId(accountId)
		if err != nil {
			continue
		}

		for _, typeId := range typeList {
			itemType, itemValue := getBlacklistValueFromAccount(&account, typeId)
			if itemValue == "" {
				continue
			}

			if _, ok := oldDatas[itemType][itemValue]; !ok {
				if newDatas[itemType] == nil {
					newDatas[itemType] = make(map[string]acctRiskInfo)
				}
				newDatas[itemType][itemValue] = acctRiskInfo{accountId, reason, remark}
			}
		}
	}

	timeNow := tools.GetUnixMillis()

	o.Using(m.Using())
	o.Begin()

	var count int = 0
	var err2 error
	for k, v := range newDatas {
		if err2 != nil {
			break
		}
		for k1, id := range v {
			c := models.CustomerRisk{}
			c.CustomerId = id.accountId
			c.RiskItem = k
			c.RiskType = 1
			c.RiskValue = k1
			c.Status = 1
			c.Reason = types.RiskReason(id.reason)
			c.Ctime = timeNow
			c.Utime = timeNow
			c.OpUid = opid
			c.ReviewTime = timeNow
			c.ReportRemark = id.mark

			_, err2 = o.Insert(&c)
			count++
			if err2 != nil {
				break
			}
		}
	}

	if err2 != nil {
		o.Rollback()
		logs.Error("[ImportBlacklist] Insert error err:%v", err2)
	} else {
		o.Commit()
		logs.Info("[ImportBlacklist] Insert done size:%d", count)
	}
}

//；；居住地址；单位地址；单位名称；；
func getBlacklistValueFromAccount(account *models.AccountBase, typeId int) (types.RiskItemEnum, string) {
	enumType := types.RiskItemEnum(typeId)

	if enumType == types.RiskItemMobile {
		return enumType, account.Mobile
	} else if enumType == types.RiskItemIdentity {
		return enumType, account.Identity
	} else if enumType == types.RiskItemIMEI ||
		enumType == types.RiskItemIP {
		order, err := dao.AccountLastLoanOrder(account.Id)
		if err != nil {
			return enumType, ""
		}
		aclientInfo, err := OrderClientInfo(order.Id)
		if err != nil {
			return enumType, ""
		}

		if enumType == types.RiskItemIMEI {
			return enumType, aclientInfo.Imei
		} else {
			return enumType, aclientInfo.IP
		}
	} else if enumType == types.RiskItemResidentAddress ||
		enumType == types.RiskItemCompany ||
		enumType == types.RiskItemCompanyAddress {
		accountProfile, err := dao.CustomerProfile(account.Id)
		if err != nil {
			return enumType, ""
		}
		if enumType == types.RiskItemResidentAddress {
			if accountProfile.ResidentCity == "" && accountProfile.ResidentAddress == "" {
				return enumType, ""
			} else {
				return enumType, accountProfile.ResidentCity + "," + accountProfile.ResidentAddress
			}
		} else if enumType == types.RiskItemCompanyAddress {
			if accountProfile.CompanyCity == "" && accountProfile.CompanyAddress == "" {
				return enumType, ""
			} else {
				return enumType, accountProfile.CompanyCity + "," + accountProfile.CompanyAddress
			}
		} else {
			return enumType, accountProfile.CompanyName
		}
	}

	return enumType, ""
}

func CanAddCustumetRecallScore(score int, order *models.Order) bool {
	value := false

	scoreDayN, _ := config.ValidItemInt("customer_recall_score_z002_N")
	scoreMin, _ := config.ValidItemInt("customer_recall_score_z002_M")

	logs.Debug("[CanAddCustumetRecallScore] score:%d scoreDayN:%d scoreMin:%d", score, scoreDayN, scoreMin)

	//1、获得是否是第一次反欺诈 第一次反欺诈拒绝时间<N
	list, num, _ := GetAllHitRegularRecordByOrderID(order.Id)

	//遍历命中列表 如果有不是z002的则说明不是唯一命中 可直接返回
	for _, v := range list {
		if v.HitRegular != types.RegularNameZ002 {
			value = false
			goto RETRUN
		}
	}

	if num > 0 {
		now := tools.GetUnixMillis()
		logs.Debug("[CanAddCustumetRecallScore] now:%d Ctime:%d ", now, list[0].Ctime)
		if (now-list[0].Ctime)/tools.MILLSSECONDADAY < int64(scoreDayN) {
			value = true
		} else {
			value = false
			goto RETRUN
		}
	} else {
		value = true
	}

	//2、获得系统配置 获得A卡评分
	//符合召回条件 给他打个标签.
	if score > scoreMin {
		value = true
	} else {
		value = false
		goto RETRUN
	}

RETRUN:
	return value
}

// 重新审核
func RiskReCheck(accountId int64) (err error) {
	order, _ := dao.AccountLastLoanOrder(accountId)
	// 1.取消标签
	accountExt, _ := models.OneAccountBaseExtByPkId(accountId)
	if accountExt.RecallTag != types.RecallTagScore {
		err = fmt.Errorf("[RiskReCheck] accountExt recall type not score . accountExt:%#v orderId:%d ", accountExt, order.Id)
		logs.Error(err)
		return err
	}

	//err = ChangeCustomerRecall(accountId, order.Id, types.RecallTagNone, types.RemarkTagNone)
	//if err != nil {
	//	logs.Error("[RiskReCheck] ChangeCustomerRecall  accountId:%d orderId:%d err:%v", accountId, order.Id, err)
	//	return err
	//}

	// 2.检查订单状态
	if order.CheckStatus != types.LoanStatusReject || order.RiskCtlStatus != types.RiskCtlAFReject {
		errStr := fmt.Sprintf("[RiskReCheck] RiskCtlStatus status err. order:%#v", order)
		err = errors.New(errStr)
		logs.Warning(err)
		return nil
	}

	old := order
	order.CheckStatus = types.LoanStatus4Review
	order.RiskCtlStatus = types.RiskCtlAFDoing
	order.RiskCtlRegular = ""
	order.RejectReason = types.RejectReasonEnum(0)
	order.CheckTime = 0
	order.RiskCtlFinishTime = 0
	order.Utime = tools.GetUnixMillis()
	order.Update("check_status", "risk_ctl_status", "check_time", "reject_reason", "risk_ctl_regular", "risk_ctl_finish_time", "utime")
	models.OpLogWrite(0, order.Id, models.OpCodeOrderUpdate, order.TableName(), old, order)

	return nil
}

//电核拒绝重新召回操作
func PhoneVrifyRefuseRecall(accountId int64, reVerify int, nextCallTime string) (err error) {
	order, _ := dao.AccountLastLoanOrder(accountId)
	// 1.取消标签
	err = ChangeCustomerRecall(accountId, order.Id, types.RecallTagNone, types.RemarkTagNone)
	if err != nil {
		logs.Error("[RiskReCheck] ChangeCustomerRecall  accountId:%d orderId:%d err:%v", accountId, order.Id, err)
		return
	}

	// 2.检查订单状态
	if order.CheckStatus != types.LoanStatusReject ||
		(order.RiskCtlStatus != types.RiskCtlPhoneVerifyReject && order.RiskCtlStatus != types.RiskCtlAutoCallReject) {
		logs.Warning("[RiskReCheck] RiskCtlStatus status err. order:", order)
		return
	}

	if reVerify == 1 {
		old := order
		order.CheckStatus = types.LoanStatusWaitManual
		order.RiskCtlStatus = types.RiskCtlWaitPhoneVerify
		order.PhoneVerifyTime = 0
		order.RejectReason = types.RejectReasonEnum(0)
		if old.FixedRandom == FixedPhoneVerifyRandom {
			order.FixedRandom = 0
		}
		order.Utime = tools.GetUnixMillis()
		order.Update("check_status", "risk_ctl_status", "phone_verify_time", "fixed_random", "utime")
		models.OpLogWrite(accountId, order.Id, models.OpCodeOrderUpdate, order.TableName(), old, order)

		//重启工单
		//ticket.ReopenTicket(order.Id, types.TicketItemPhoneVerify)
		ticket.ReopenPhoneVerifyOrInfoReviewByRelatedID(order.Id, nextCallTime)
		//ticket.ReopenByRelatedID(order.Id, types.TicketItemPhoneVerify, nextCallTime)

	}

	return
}

func HandleRecallCancleScore(accountId int64) {

	//当前时间-此订单第一次反欺诈拒绝时间>=N 的客户，取消“评分模型需召回的客户”的 标签。
	order, _ := dao.AccountLastLoanOrder(accountId)
	scoreMin, _ := config.ValidItemInt("customer_recall_score_z002_M")

	// >= N 取消标签
	if !CanAddCustumetRecallScore(scoreMin+1, &order) {
		ChangeCustomerRecall(accountId, order.Id, types.RecallTagNone, types.RemarkTagNone)
	} else {
		logs.Info("[HandleRecallCancleScore] no need to cancle account:%d", accountId)
	}
}
