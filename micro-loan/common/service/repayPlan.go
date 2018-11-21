package service

import (
	"reflect"
	"sort"

	"github.com/astaxie/beego/logs"

	"fmt"
	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/types"
)

func GetRepayAmount(orderId int64) (amount int64, err error) {
	repayPlan, err := models.GetLastRepayPlanByOrderid(orderId)
	if err != nil {
		logs.Error("There is not a repay plan for this order:", orderId)
		return
	}
	amount = (repayPlan.Amount - repayPlan.AmountPayed) + (repayPlan.GracePeriodInterest - repayPlan.GracePeriodInterestPayed) + (repayPlan.Penalty - repayPlan.PenaltyPayed)
	return
}

func GetBackendRepayPlanHistory(orderId int64) (list []models.RepayPlanHistory) {
	//1、根据order_id 取得所有的减免记录
	reduceRecordList, err := models.GetAllReduceRecordNew(orderId)
	if err != nil {
		logs.Error("[service.GetBackendRepayPlanHistory.GetAllReduceRecord] catch err:", err)
		return
	}

	//2、根据order_id 查询所有还款记录
	userETransList := models.GetOutETransByOrderId(orderId)

	//3、根据order_id 查询逾期记录 包括罚息、宽限息，如果订单为逾期
	repayPlanOverdueList, err := models.GetRepayPlanOverdueByOrderId(orderId)
	if err != nil {
		logs.Error("[service.GetBackendRepayPlanHistory.GetRepayPlanOverdueByOrderId] catch err:", err)
		return
	}

	//4、获取当前的还款计划
	repayPlan, err := models.GetLastRepayPlanByOrderid(orderId)
	if err != nil {
		logs.Error("[service.GetBackendRepayPlanHistory.GetLastRepayPlanByOrderid] catch err:", err)
		return
	}

	//5、整合所有记录生成 未按照生成时间排序的记录
	modifyMap := sortAllModify(&reduceRecordList, &userETransList, &repayPlanOverdueList)

	//6、通过记录和当前还款计划反推出之前的状态
	list = doGetHistory(repayPlan, modifyMap)
	return list
}

func sortAllModify(reduceList *[]models.ReduceRecordNew, eTransList *[]models.User_E_Trans, repayPlanOverdueList *[]models.RepayPlanOverdue) (modifyMap map[int64]interface{}) {

	modifyMap = make(map[int64]interface{})
	totalNum := 0
	if nil != reduceList {
		totalNum += len(*reduceList)
		for _, v := range *reduceList {
			org := modifyMap[v.Ctime]
			if org != nil {
				logs.Warn("[service.sortAllModify] range reduceList catch same ctime. org:%#v", org)
			}
			modifyMap[v.Ctime] = v
		}
	}

	if nil != eTransList {
		totalNum += len(*eTransList)
		for _, v := range *eTransList {
			org := modifyMap[v.Ctime]
			if org != nil {
				logs.Warn("[service.sortAllModify] range eTransList catch same ctime. org:%#v", org)
			}
			modifyMap[v.Ctime] = v
		}
	}

	if nil != repayPlanOverdueList {
		totalNum += len(*repayPlanOverdueList)
		len := len(*repayPlanOverdueList)
		for k, v := range *repayPlanOverdueList {
			org := modifyMap[v.Ctime]
			if org != nil {
				logs.Warn("[service.sortAllModify] range repayPlanOverdueList catch same ctime. org:%#v", org)
			}
			if k < len-1 {
				// logs.Warn("v.Penalty:%d (*repayPlanOverdueList)[k+1].Penalty:%d [v.Ctime]:%d", v.Penalty, (*repayPlanOverdueList)[k+1].Penalty, v.Ctime)
				if v.GracePeriodInterest > 0 {
					v.GracePeriodInterest -= (*repayPlanOverdueList)[k+1].GracePeriodInterest
				}

				if v.Penalty > 0 {
					v.Penalty -= (*repayPlanOverdueList)[k+1].Penalty
				}

			}
			modifyMap[v.Ctime] = v
		}
	}

	if len(modifyMap) != totalNum {
		logs.Warn("[service.sortAllModify] record count not equal. Get same ctime totalNum:", totalNum, " len(modifyMap):", len(modifyMap))
	}

	return
}

func doGetHistory(repayPlan models.RepayPlan, modifyMap map[int64]interface{}) (list []models.RepayPlanHistory) {
	// jsonStr, _ := json.Marshal(repayPlan)
	// logs.Warn(string(jsonStr))
	// jsonStr, _ = json.Marshal(modifyMap)
	// logs.Warn(string(jsonStr))
	// logs.Warn("modifyMap:%d", len(modifyMap))

	sortedKeys := make([]float64, 0)
	for k := range modifyMap {
		sortedKeys = append(sortedKeys, float64(k))
	}
	sort.Float64s(sortedKeys)

	//按时间降序排列
	sort.Sort(sort.Reverse(sort.Float64Slice(sortedKeys)))

	//遍历排序后的记录
	i := 1
	for _, k := range sortedKeys {
		history := models.RepayPlanHistory{}
		int64Key := int64(k)
		v := modifyMap[int64Key]
		// logs.Warn("int64Key=%v, v=%#v\n", int64Key, v)

		vType := reflect.TypeOf(v)
		var err error
		flag := false
		switch vType.Name() {
		case "ReduceRecordNew":
			{
				reduceRecord := v.(models.ReduceRecordNew)
				flag, err = doReverseReduceRecord(&repayPlan, &history, &reduceRecord)
			}
		case "User_E_Trans":
			{
				eTrans := v.(models.User_E_Trans)
				flag, err = doReverseUserETrans(&repayPlan, &history, &eTrans)
			}
		case "RepayPlanOverdue":
			{
				repayPlanOverdue := v.(models.RepayPlanOverdue)
				flag, err = doReverseRepayPlanOverdue(&repayPlan, &history, &repayPlanOverdue)
			}
		default:
			{
				logs.Warn("[service.doGetHistory] unknow type: %#v", v)
			}
		}

		// append repayPlan
		if true == flag && nil == err {
			history.Id = i
			i++
			list = append(list, history)
		} else if nil != err {
			logs.Error("[service.doGetHistory] reverse err:%s v:%#v", err, v)
		}
	}

	//最后追加还款计划。
	// getOriginRepayPlan()
	history := models.RepayPlanHistory{
		Id: i,
		Plan: models.RepayPlan{
			Amount:      repayPlan.Amount,
			PreInterest: repayPlan.PreInterest,
			ServiceFee:  repayPlan.ServiceFee,
			RepayDate:   repayPlan.RepayDate,
		},
		PayOutTime: repayPlan.Ctime,
	}

	list = append(list, history)
	return
}

// doReverseReduceRecord ReduceRecord 用来生成用户的减免记录
func doReverseReduceRecord(repayPlan *models.RepayPlan, history *models.RepayPlanHistory, reduceRecord *models.ReduceRecordNew) (flag bool, err error) {
	err = nil
	flag = true
	history.PayInTime = reduceRecord.Ctime
	history.Plan = models.RepayPlan{
		AmountReduced:              reduceRecord.AmountReduced,
		GracePeriodInterestReduced: reduceRecord.GraceInterestReduced,
		PenaltyReduced:             reduceRecord.PenaltyReduced,
		RepayDate:                  repayPlan.RepayDate,
	}
	return
}

// doReverseUserETrans user_e_tran 用来生成用户的还款记录
func doReverseUserETrans(repayPlan *models.RepayPlan, history *models.RepayPlanHistory, eTrans *models.User_E_Trans) (flag bool, err error) {
	err = nil
	flag = true

	// 入账记录是还款总额，出帐记录才是实际还款的记录  减免记录不在这个表里处理
	if types.PayTypeMoneyIn == eTrans.PayType || types.MobiReductionPenalty == eTrans.VaCompanyCode {
		flag = false
		return
	}

	history.PayInTime = eTrans.Ctime
	history.Plan = models.RepayPlan{
		// Amount:                   repayPlan.Amount,
		AmountPayed: eTrans.Amount,
		// PreInterest:              repayPlan.PreInterest,
		PreInterestPayed: eTrans.PreInterest,
		// GracePeriodInterest:      repayPlan.GracePeriodInterest,
		GracePeriodInterestPayed: eTrans.GracePeriodInterest,
		// Interest:                 repayPlan.Interest,
		InterestPayed: eTrans.Interest,
		// ServiceFee:               repayPlan.ServiceFee,
		ServiceFeePayed: eTrans.ServiceFee,
		// Penalty:                  repayPlan.Penalty,
		PenaltyPayed: eTrans.Penalty,
		RepayDate:    repayPlan.RepayDate,
	}
	return
}

// doReverseReduceRecord ReduceRecord 用来生成用户的逾期记录
func doReverseRepayPlanOverdue(repayPlan *models.RepayPlan, history *models.RepayPlanHistory, repayPlanOverdue *models.RepayPlanOverdue) (flag bool, err error) {
	err = nil
	flag = true

	history.PayOutTime = repayPlanOverdue.Ctime
	history.Plan = models.RepayPlan{
		// Amount:              repayPlan.Amount,
		// PreInterest:         repayPlan.PreInterest,
		GracePeriodInterest: repayPlanOverdue.GracePeriodInterest,
		// Interest:            repayPlan.Interest,
		// ServiceFee:          repayPlan.ServiceFee,
		Penalty:   repayPlanOverdue.Penalty,
		RepayDate: repayPlan.RepayDate,
	}

	return
}

func GetRollBackDetail(orderId int64) (totalPayed int64) {
	order, err := models.GetOrder(orderId)
	if err != nil {
		logs.Error("[GetRollBackDetail] GetOrder err:%v id:%d", err, orderId)
		return
	}

	if order.CheckStatus == types.LoanStatusRolling ||
		order.CheckStatus == types.LoanStatusRollClear ||
		order.CheckStatus == types.LoanStatusRollApply ||
		order.PreOrder > 0 {
		logs.Error("[GetRollBackDetail] order status:%d no support id:%d", order.CheckStatus, orderId)
		return
	}

	repayPlan := GetBackendRepayPlan(orderId)
	if repayPlan.Id == 0 {
		logs.Error("[GetRollBackDetail] GetBackendRepayPlan id:%d", orderId)
		return
	}
	totalPayed, _ = repayplan.CaculateTotalPayedByRepayPlan(repayPlan)

	//1 是否有过优惠券记录
	coupon, _ := dao.GetAccountCouponByOrderAndStatus(order.UserAccountId, order.Id, int(types.CouponStatusUsed))

	// 减去优惠券的金额
	return totalPayed - coupon.Amount
}

func DoRollBackRepayPlan(opUid, orderId, rollBackTotal int64) (err error) {
	order, err := models.GetOrder(orderId)
	if err != nil {
		err = fmt.Errorf("[GetRollBackDetail] GetOrder err:%v id:%d", err, orderId)
		logs.Error(err)
		return
	}

	//清除非优惠券 非砍头息的内容
	repayPlan := GetBackendRepayPlan(orderId)
	if repayPlan.Id == 0 {
		err = fmt.Errorf("[DoRollBackRepayPlan] GetBackendRepayPlan id:%d", orderId)
		logs.Error(err)
		return
	}

	totalPayed, _ := repayplan.CaculateTotalPayedByRepayPlan(repayPlan)
	coupon, _ := dao.GetAccountCouponByOrderAndStatus(order.UserAccountId, order.Id, int(types.CouponStatusUsed))
	totalPayed = totalPayed - coupon.Amount
	if rollBackTotal != totalPayed ||
		totalPayed == 0 {
		err = fmt.Errorf("[DoRollBackRepayPlan] data errr. id:%d rollBackTotal:%d", orderId, rollBackTotal)
		logs.Error(err)
		return
	}

	// 1\ 更新还款计划
	old := repayPlan
	repayPlan.AmountPayed = coupon.Amount
	repayPlan.GracePeriodInterestPayed = 0
	repayPlan.PenaltyPayed = 0

	models.OrmAllUpdate(&repayPlan)
	models.OpLogWrite(opUid, orderId, models.OpCodeRepayPlanUpdate, repayPlan.TableName(), old, repayPlan)

	// 2 增余额
	err = IncreaseBalanceByRefund(order.UserAccountId, totalPayed)
	if err != nil {
		logs.Error("[DoRollBackRepayPlan] IncreaseBalanceByRefund err:%v UserAccountId:%d totalPayed:%d", err, order.UserAccountId, totalPayed)
	}

	// 3 更新订单状态
	oldOrder := order
	if order.IsOverdue == 1 {

		//将最后一条出崔改为如催
		if order.CheckStatus == types.LoanStatusAlreadyCleared {
			clearOverdueCase(opUid, orderId)
		}
		order.CheckStatus = types.LoanStatusOverdue
	} else if coupon.Amount > 0 {
		order.CheckStatus = types.LoanStatusPartialRepayment
	} else {
		order.CheckStatus = types.LoanStatusWaitRepayment
	}
	models.OrmAllUpdate(&order)
	models.OpLogWrite(opUid, orderId, models.OpCodeOrderUpdate, order.TableName(), oldOrder, order)

	//4 清空user_e_trans
	deleteUserTrans(opUid, orderId, types.Xendit)
	deleteUserTrans(opUid, orderId, types.DoKu)
	deleteUserTrans(opUid, orderId, types.Bluepay)
	return
}

func clearOverdueCase(opUid, orderId int64) {

	oneCase, err := models.OneOverdueCaseByOrderID(orderId)
	if err != nil {
		logs.Error("[clearOverdueCase] OneOverdueCaseByOrderID:%d err:%v", orderId, err)
		return
	}

	if oneCase.IsOut == types.IsUrgeOutNo {
		return
	}

	old := oneCase
	oneCase.IsOut = types.IsUrgeOutNo
	models.OrmAllUpdate(&oneCase)
	models.OpLogWrite(opUid, orderId, models.OpOverdueCaseUpdate, old.TableName(), old, oneCase)
}

func deleteUserTrans(opUid, orderId int64, company int) {
	trans := models.GetAllETransByCompany(orderId, company)
	for _, v := range trans {
		models.OrmDelete(&v)
		models.OpLogWrite(opUid, orderId, models.OpUserEtransModDel, v.TableName(), v, "")
	}
}
