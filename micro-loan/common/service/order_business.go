package service

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/cerror"
	"micro-loan/common/dao"
	"micro-loan/common/i18n"
	"micro-loan/common/lib/device"
	"micro-loan/common/lib/gaws"
	"micro-loan/common/lib/redis/cache"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/coupon_event"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/pkg/schema_task"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/thirdparty/advance"
	"micro-loan/common/thirdparty/doku"
	"micro-loan/common/thirdparty/xendit"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// api接口层专用方法 {{{

type orderDataSet map[string]interface{}

type BorrowCreditBackendData struct {
	Id              int64
	PreOrder        int64
	UserAccountId   int64
	Realname        string
	Loan            int64
	Amount          int64
	Period          int
	CheckStatus     types.LoanStatus
	RiskCtlStatus   types.RiskCtlEnum
	ApplyTime       int64
	CheckTime       int64
	RepayTime       int64
	PhoneVerifyTime int64
	LoanTime        int64
	FinishTime      int64
	RandomMark      int64
	Ctime           int64
	Mobile          string
	UiVersion       string
	AppVersionCode  int
	CTemp           int
	IsTemporary     int
	IsOverdue       int
	IsDeadDebt      int
	OrderRandomMark int
	IsReloan        int
}

type GiveoutCreditBackendData struct {
	Id              int64
	AccountId       int64
	EAccountNumber  string
	Realname        string
	Mobile          int64
	BankName        string
	BankNo          string
	Loan            int64
	Amount          int64
	AmountPayed     int64
	Period          int
	CheckStatus     types.LoanStatus
	ApplyTime       int64
	CheckTime       int64
	RepayDate       int64
	PhoneVerifyTime int64
	LoanTime        int64
	FinishTime      int64
	DisbursementId  string
	LoanCompany     int
}

// RepayBackendData 还款计划后台数据列表
type RepayBackendData struct {
	Id                         int64
	UserAccountId              int64
	EAccountNumber             string
	Realname                   string
	Loan                       int64
	TotalRepay                 int64
	TotalRepayPayed            int64
	Amount                     int64
	AmountPayed                int64
	AmountReduced              int64
	GracePeriodInterest        int64
	GracePeriodInterestPayed   int64
	GracePeriodInterestReduced int64
	Penalty                    int64
	PenaltyPayed               int64
	PenaltyReduced             int64
	Period                     int
	CheckStatus                types.LoanStatus
	ApplyTime                  int64
	CheckTime                  int64
	RepayDate                  int64
	PhoneVerifyTime            int64
	LoanTime                   int64
	RepayTime                  int64 // 实际还款日期 orders.repay_time
	FinishTime                 int64 // 结清时间 orders.finish_time
	ReduceTotal                int64 // 总的减免额
	RepayBalanceAmount         int64
	UserEAccounts              [][]string
}

// RepayVaDisplay 还款计划VA展示
type RepayVaDisplay struct {
	UserAccountId int64
	RealName      string
	OrderId       int64
	Mobile        string
	Code          string
	ApplyTime     int64
	ExpireTime    int64
	CompanyCode   int
	Amount        int64
}

// OrderIdReduce id 和减免额的 struct
type OrderIdReduce struct {
	OrderId     int64 `orm:"order_id"`
	ReduceTotal int64 `orm:"reduce_total"`
}

type UserEtransBackendData struct {
	Id            int64
	UserAccountId int64
	VaCompanyCode int
	OrderId       int64
	Amount        int64
	Interest      int64
	ServiceFee    int64
	Penalty       int64
	PayType       int
	Ctime         int64
	Utime         int64
}

type OverdueCaseListItem struct {
	models.OverdueCase
	AccountId                  int64
	Realname                   string
	Mobile                     string
	TotalRepay                 int64
	TotalRepayPayed            int64
	Amount                     int64
	AmountPayed                int64
	AmountReduced              int64
	GracePeriodInterest        int64
	GracePeriodInterestPayed   int64
	GracePeriodInterestReduced int64
	Penalty                    int64
	PenaltyPayed               int64
	PenaltyReduced             int64
	RepayDate                  int64
	SalaryDay                  string
	CompanyTelephone           string
	PromiseRepayTime           int64
	PhoneTime                  int64
}

type OctwoCaseListItem struct {
	models.OverdueCase
	AccountId                  int64
	Realname                   string
	Mobile                     string
	TotalRepay                 int64
	TotalRepayPayed            int64
	Amount                     int64
	AmountPayed                int64
	AmountReduced              int64
	GracePeriodInterest        int64
	GracePeriodInterestPayed   int64
	GracePeriodInterestReduced int64
	Penalty                    int64
	PenaltyPayed               int64
	PenaltyReduced             int64
	RepayDate                  int64
	SalaryDay                  string
	CompanyTelephone           string
	PromiseRepayTime           int64
	PhoneTime                  int64
	EntrustPname               string `orm:"column(entrust_pname)"`
	IsEntrust                  int    `orm:"column(is_entrust)"`
	IsTicket                   int64
}

var overdueFieldMap = map[string]string{
	"OrderId":      "c.id",
	"RepayDate":    "r.repay_date",
	"TotalRepay":   "(r.amount + r.grace_period_interest + r.penalty)",
	"JoinUrgeTime": "c.join_urge_time",
	"OverdueDays":  "c.overdue_days",
}

var orderFieldMap = map[string]string{
	"Id":         "orders.id",
	"Amount":     "orders.amount",
	"Loan":       "orders.loan",
	"Period":     "orders.period",
	"ApplyTime":  "orders.apply_time",
	"CheckTime":  "orders.check_time",
	"LoanTime":   "orders.loan_time",
	"FinishTime": "orders.finish_time",
	"Ctime":      "orders.ctime",
}

var loanFieldMap = map[string]string{
	"Id":              "orders.id",
	"Amount":          "orders.amount",
	"Loan":            "orders.loan",
	"Period":          "orders.period",
	"PhoneVerifyTime": "orders.phone_verify_time",
	"LoanTime":        "orders.loan_time",
	"FinishTime":      "orders.finish_time",
}

var repayFieldMap = map[string]string{
	"Id":              "orders.id",
	"RepayDate":       "repay_plan.repay_date",
	"RepayTime":       "orders.repay_time",
	"FinishTime":      "orders.finish_time",
	"Loan":            "orders.loan",
	"TotalRepay":      "(repay_plan.amount + repay_plan.grace_period_interest + repay_plan.penalty)",
	"TotalRepayPayed": "(repay_plan.amount_payed + repay_plan.grace_period_interest_payed + repay_plan.penalty_payed)",
	"ReduceTotal":     "(repay_plan.amount_reduced + repay_plan.grace_period_interest_reduced + repay_plan.penalty_reduced)",
}

func BuildEmptyOrderData(data map[string]interface{}) {
	data["size"] = 0
	data["offset"] = "0"
	var orderData = make([]orderDataSet, 0)
	data["order_data"] = orderData
}

func BuildEmptyHomeOrderData() (data map[string]interface{}) {

	data = map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}

	return
}

func BuildOrderData(data map[string]interface{}, orderList []models.Order) {
	size := len(orderList)
	data["size"] = size
	data["offset"] = tools.Int642Str(orderList[size-1].Id)

	var orderData []orderDataSet
	for _, order := range orderList {
		repayPlan, _ := models.GetLastRepayPlanByOrderid(order.Id)

		var displayAmount int64
		// 逾期,待还款,部分还款,显示金额展示还需要还多少
		if order.CheckStatus == types.LoanStatusWaitRepayment || order.CheckStatus == types.LoanStatusOverdue || order.CheckStatus == types.LoanStatusPartialRepayment {
			displayAmount = repayPlan.Amount + repayPlan.Penalty + repayPlan.GracePeriodInterest - repayPlan.AmountPayed - repayPlan.PenaltyPayed - repayPlan.GracePeriodInterestPayed
		} else {
			displayAmount = order.Amount
		}

		// 订单状态为'等待自动外呼'时，显示给客户端为'等待人工审核'
		if order.CheckStatus == types.LoanStatusWaitAutoCall {
			order.CheckStatus = types.LoanStatusWaitManual
		}
		var displayTime int64
		if order.CheckStatus == types.LoanStatusSubmit ||
			order.CheckStatus == types.LoanStatus4Review ||
			order.CheckStatus == types.LoanStatusReject ||
			order.CheckStatus == types.LoanStatusWaitManual ||
			order.CheckStatus == types.LoanStatusLoanFail ||
			order.CheckStatus == types.LoanStatusWait4Loan ||
			order.CheckStatus == types.LoanStatusIsDoing ||
			order.CheckStatus == types.LoanStatusThirdBlacklistIsDoing ||
			order.CheckStatus == types.LoanStatusWaitPhotoCompare {
			displayTime = order.ApplyTime
		} else if order.CheckStatus == types.LoanStatusWaitRepayment || order.CheckStatus == types.LoanStatusAlreadyCleared || order.CheckStatus == types.LoanStatusOverdue || order.CheckStatus == types.LoanStatusPartialRepayment {
			displayTime = repayPlan.RepayDate
		}

		_, eAccountDesc := DisplayVAInfoV2(order.UserAccountId)

		subSet := map[string]interface{}{
			"id":                order.Id,
			"loan":              order.Loan,
			"period":            order.Period,
			"apply_time":        order.ApplyTime,
			"status":            order.CheckStatus,
			"e_account_number":  eAccountDesc,
			"repay_date":        order.ApplyTime,
			"display_time":      displayTime,
			"reason":            order.RejectReason,
			"repayment":         repayPlan.Amount - repayPlan.AmountPayed,
			"fine":              repayPlan.Penalty - repayPlan.PenaltyPayed,
			"actual_repay_time": order.FinishTime,
			"display_amount":    displayAmount,
		}

		orderData = append(orderData, subSet)
	}

	data["order_data"] = orderData
}

// 所有订单数据
func BuildOrderDataV2(data map[string]interface{}, orderList []models.Order, isExtension bool) {
	size := len(orderList)
	data["size"] = size
	data["offset"] = tools.Int642Str(orderList[size-1].Id)

	var orderData []orderDataSet
	for _, order := range orderList {
		var totalGrace int64
		repayPlan, err := models.GetLastRepayPlanByOrderid(order.Id)
		if err == nil {
			totalGrace, _ = repayplan.CaculateTotalGracePeriodAndPenaltyByRepayPlan(repayPlan)
		}

		var displayAmount int64
		var extensionRefund int64
		var clearReduced int64
		var repaymentBeforeReduce int64 // 结清减免之前的金额
		var isClearReduced bool
		// 待还款,部分还款,以及等待展期,显示金额展示还需要还多少
		if order.CheckStatus == types.LoanStatusWaitRepayment ||
			order.CheckStatus == types.LoanStatusPartialRepayment || order.CheckStatus == types.LoanStatusRolling {
			displayAmount, _ = repayplan.CaculateRepayTotalAmountByRepayPlan(repayPlan)
		} else if order.CheckStatus == types.LoanStatusOverdue { // 逾期结清减免显示，减去结清减免额的数值
			displayAmount, _ = repayplan.CaculateRepayTotalAmountByRepayPlan(repayPlan)
			clearReducedVal, err := repayplan.CaculatePenaltyClearReducedByOrderId(order.Id)
			if err == nil { // 结清减免, 并且未生效
				repaymentBeforeReduce = displayAmount
				displayAmount = displayAmount - clearReducedVal
				clearReduced = clearReducedVal
				isClearReduced = true
			}
		} else if order.CheckStatus == types.LoanStatusRollApply {
			// 展期申请中,显示金额为最低还款额(父订单)
			preOrder, _ := models.GetOrder(order.PreOrder)
			displayAmount = preOrder.MinRepayAmount
			extensionRefund = order.Amount
		} else {
			displayAmount = order.Amount
		}

		accountCoupon, err := dao.GetAccountFrozenCouponByOrder(order.UserAccountId, order.Id)
		if err == nil {
			displayAmount = displayAmount - accountCoupon.Amount
		}

		// 订单状态为'等待自动外呼'时，显示给客户端为'等待人工审核'
		if order.CheckStatus == types.LoanStatusWaitAutoCall {
			order.CheckStatus = types.LoanStatusWaitManual
		}

		var displayTime int64
		if order.CheckStatus == types.LoanStatusSubmit ||
			order.CheckStatus == types.LoanStatus4Review ||
			order.CheckStatus == types.LoanStatusReject ||
			order.CheckStatus == types.LoanStatusWaitManual ||
			order.CheckStatus == types.LoanStatusLoanFail ||
			order.CheckStatus == types.LoanStatusWait4Loan ||
			order.CheckStatus == types.LoanStatusAlreadyCleared ||
			order.CheckStatus == types.LoanStatusIsDoing ||
			order.CheckStatus == types.LoanStatusThirdBlacklistIsDoing ||
			order.CheckStatus == types.LoanStatusRollClear ||
			order.CheckStatus == types.LoanStatusWaitPhotoCompare {
			displayTime = order.ApplyTime
		} else if order.CheckStatus == types.LoanStatusWaitRepayment ||
			order.CheckStatus == types.LoanStatusOverdue ||
			order.CheckStatus == types.LoanStatusPartialRepayment ||
			order.CheckStatus == types.LoanStatusRolling {
			displayTime = repayPlan.RepayDate
		} else if order.CheckStatus == types.LoanStatusRollApply {
			displayTime = tools.GetIDNCurrDayLastSecond() // 当前日期的最后一秒
		}

		_, eAccountDesc := DisplayVAInfoV2(order.UserAccountId)

		subSet := map[string]interface{}{
			"id":                      order.Id,
			"loan":                    order.Loan,
			"period":                  order.Period,
			"apply_time":              order.ApplyTime,
			"status":                  order.CheckStatus,
			"e_account_number":        eAccountDesc,
			"repay_date":              order.ApplyTime,
			"display_time":            displayTime,
			"reason":                  order.RejectReason,
			"repayment":               repayPlan.Amount - repayPlan.AmountPayed,
			"fine":                    totalGrace,
			"actual_repay_time":       order.FinishTime,
			"display_amount":          displayAmount,
			"repayment_before_reduce": repaymentBeforeReduce,
			"extension_refund":        extensionRefund,
			"is_clear_reduced":        isClearReduced,
			"clear_reduced":           clearReduced,
			"is_extension":            isExtension,
		}

		adPosition, _ := GetAdPositionDisplay(order.UserAccountId, types.AdPositionMyAccountPage)
		ApiDataAddAdPosition(adPosition, subSet)

		orderData = append(orderData, subSet)
	}

	data["order_data"] = orderData
}

// 客户端首页订单数据
func BuildHomeOrderData(order models.Order) (data map[string]interface{}) {

	var totalGrace int64
	repayPlan, err := models.GetLastRepayPlanByOrderid(order.Id)
	if err == nil {
		totalGrace, _ = repayplan.CaculateTotalGracePeriodAndPenaltyByRepayPlan(repayPlan)
	}

	var repayment int64
	// 逾期,待还款,部分还款,以及展期,剩余应还金额展示
	if order.CheckStatus == types.LoanStatusWaitRepayment || order.CheckStatus == types.LoanStatusOverdue ||
		order.CheckStatus == types.LoanStatusPartialRepayment || order.CheckStatus == types.LoanStatusRolling {
		repayment, _ = repayplan.CaculateRepayTotalAmountByRepayPlan(repayPlan)
	}

	var amount int64
	var repaymentBeforeReduce int64
	var activeTag string
	var activeValue int64
	var overdueDayNum int
	var isClearReduced bool
	var isExtension bool

	amount, _ = repayplan.CaculateTotalAmountByRepayPlan(repayPlan)
	activeTag = i18n.GetMessageText(i18n.HomeOrderTagPenalty)
	activeValue = totalGrace
	if order.CheckStatus == types.LoanStatusWaitRepayment || order.CheckStatus == types.LoanStatusPartialRepayment {
		activeTag = i18n.GetMessageText(i18n.HomeOrderTagPayedAmount)
		activeValue, _ = repayplan.CaculateTotalPayedByRepayPlan(repayPlan)
	} else if order.CheckStatus == types.LoanStatusOverdue {
		overdueDayNum = int((tools.GetUnixMillis() - repayPlan.RepayDate) / (3600 * 24 * 1000))
		clearReduced, err := repayplan.CaculatePenaltyClearReducedByOrderId(order.Id)
		if err == nil { // 结清减免, 并且未生效
			repaymentBeforeReduce = repayment
			repayment = repayment - clearReduced
			activeTag = i18n.GetMessageText(i18n.HomeOrderTagReducedPenalty)
			activeValue = clearReduced
			isClearReduced = true
		} else { // 是否满足展期
			isExtension = IsOrderExtension(order)
		}
	} else if order.CheckStatus == types.LoanStatusRolling {
		activeTag = i18n.GetMessageText(i18n.HomeOrderTagMixPayAmount)
		activeValue = order.MinRepayAmount
	}

	_, eAccountDesc := DisplayVAInfoV2(order.UserAccountId)

	var paymentCode string
	var paymentCodeExpire int64
	if order.CheckStatus != types.LoanStatusRolling {
		now := time.Now().Unix() * 1000

		marketPayment, err := models.OneFixPaymentCodeByUserAccountId(order.UserAccountId)
		if err == nil && marketPayment.ExpirationDate > now {
			paymentCode = marketPayment.PaymentCode
			paymentCodeExpire, _ = tools.GetTimeParseWithFormat("2019-12-31 23:59:59", "2006-01-02 15:04:05")
			paymentCodeExpire = 1000 * paymentCodeExpire
		} else {
			marketPaymentOld, err := models.GetMarketPaymentByOrderId(order.Id)
			if err == nil && marketPaymentOld.ExpiryDate > now {
				paymentCode = marketPaymentOld.PaymentCode
				paymentCodeExpire = marketPaymentOld.ExpiryDate
			}
		}
	}

	data = map[string]interface{}{
		"server_time":             tools.GetUnixMillis(),
		"id":                      order.Id,
		"loan":                    order.Loan,
		"period":                  order.Period,
		"apply_time":              order.ApplyTime,
		"status":                  order.CheckStatus,
		"e_account_number":        eAccountDesc,
		"repay_date":              repayPlan.RepayDate,
		"repayment":               repayment,
		"repayment_before_reduce": repaymentBeforeReduce,
		"amount":                  amount,
		"overdue_day_num":         overdueDayNum,
		"payment_code":            paymentCode,
		"paymentcode_expire":      paymentCodeExpire,
		"active_tag":              activeTag,
		"active_value":            activeValue,
		"is_clear_reduced":        isClearReduced,
		"is_extension":            isExtension,
	}

	return
}

// MenuControlByOrderStatus 只有临时订单 或者 失效订单，隐藏菜单
func MenuControlByOrderStatus(accountID int64) (show bool) {

	show = false
	order := models.Order{}
	o := orm.NewOrm()
	o.Using(order.Using())
	// 2,3,4,5,6,7,8,9,11,12,13,14,15,16,17
	var orderCount int64
	sql := fmt.Sprintf(`SELECT count(*) FROM %s WHERE user_account_id= %d AND is_deleted=0 AND check_status in (%d,%d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d)`,
		order.TableName(), accountID,
		types.LoanStatus4Review,
		types.LoanStatusReject,
		types.LoanStatusWaitManual,
		types.LoanStatusWait4Loan,
		types.LoanStatusLoanFail,
		types.LoanStatusWaitRepayment,
		types.LoanStatusAlreadyCleared,
		types.LoanStatusOverdue,
		types.LoanStatusPartialRepayment,
		types.LoanStatusIsDoing,
		types.LoanStatusThirdBlacklistIsDoing,
		types.LoanStatusRolling,
		types.LoanStatusRollClear,
		types.LoanStatusRollApply,
		types.LoanStatusRollFail,
		types.LoanStatusWaitAutoCall,
		types.LoanStatusWaitPhotoCompare,
	)

	err := o.Raw(sql).QueryRow(&orderCount)
	if err != nil {
		logs.Error("[MenuControlByOrderStatus] 获取数据失败， ERROR：", err)
	}

	logs.Debug("[MenuControlByOrderStatus]满足条件的订单数：", orderCount)
	if orderCount >= 1 {
		show = true
	}
	return
}

// MenuControlByOrderStatusV2 客户存在"在贷订单"时,显示便利店付款码;并且借款订单提交审核后,未结清前，不允许编辑资料
// 在2018-08-22之后的版本使用
func MenuControlByOrderStatusV2(accountID int64) (show bool) {

	show = false
	order := models.Order{}
	o := orm.NewOrm()
	o.Using(order.Using())
	// 7,9,11,14,16
	var orderCount int64
	sql := fmt.Sprintf(`SELECT count(*) FROM %s WHERE user_account_id= %d AND is_deleted=0 AND check_status in (%d, %d, %d, %d, %d)`,
		order.TableName(), accountID,
		types.LoanStatusWaitRepayment,
		types.LoanStatusOverdue,
		types.LoanStatusPartialRepayment,
		types.LoanStatusRolling,
		types.LoanStatusRollApply,
	)

	err := o.Raw(sql).QueryRow(&orderCount)
	if err != nil {
		logs.Error("[MenuControlByOrderStatusV2] 获取数据失败， ERROR：", err)
	}

	logs.Debug("[MenuControlByOrderStatusV2]满足条件的订单数：", orderCount)
	if orderCount >= 1 {
		show = true
	}
	return
}

// 是否在首页显示再贷借款订单(等待还款, 部分还款, 逾期, 展期)
func GetHomeOrderType(accountId int64) (orderType types.HomeOrderType) {
	order, err := dao.AccountLastLoanOrder(accountId)
	if err != nil {
		// 没有有效订单
		return
	}

	if order.CheckStatus == types.LoanStatusWaitRepayment || order.CheckStatus == types.LoanStatusPartialRepayment ||
		order.CheckStatus == types.LoanStatusOverdue || order.CheckStatus == types.LoanStatusRolling ||
		order.CheckStatus == types.LoanStatusRollApply {
		orderType = types.HomeOrderTypeLoaning
	}

	// 订单为＂等待人工审核＂
	if order.CheckStatus == types.LoanStatusWaitManual || order.CheckStatus == types.LoanStatusWaitAutoCall {
		orderType = types.HomeOrderTypePhoneVerify
	}

	if order.CheckStatus == types.LoanStatusReject {
		accountExt, _ := models.OneAccountBaseExtByPkId(accountId)
		if accountExt.RecallTag == int(types.HomeOrderTypeCustomerRecallScore) {
			orderType = types.HomeOrderType(accountExt.RecallTag)
		}
	}

	if order.CheckStatus == types.LoanStatusLoanFail {
		accountExt, _ := models.OneAccountBaseExtByPkId(accountId)
		if accountExt.RecallTag == int(types.HomeOrderTypeModifyBank) &&
			CanUpdateBankInfo() {
			orderType = types.HomeOrderType(accountExt.RecallTag)
		}
	}

	return
}

// 取用户的借款生命周期
func GetLoanLifetime(accountId int64) (loanLifetime int) {
	order, err := dao.AccountLastLoanOrder(accountId)
	if err != nil {
		// 没有有效订单,状态为初始
		loanLifetime = types.LoanLifetimeNormal

		return
	}

	// 最后一条有效订单是正常还款或无效的订单
	if order.CheckStatus == types.LoanStatusAlreadyCleared || order.CheckStatus == types.LoanStatusInvalid || order.CheckStatus == types.LoanStatusRollClear {
		loanLifetime = types.LoanLifetimeNormal

		return
	}

	// 审核被拒7天后，可以下单
	if order.CheckStatus == types.LoanStatusReject {
		var limitTime int64 = 3600 * 24 * 7 * 1000 // 7天
		if tools.GetUnixMillis()-order.CheckTime < limitTime {
			loanLifetime = types.LoanLifetimeReject
		} else {
			loanLifetime = types.LoanLifetimeNormal
		}

		return
	}

	// 有正在进行中订单
	if order.CheckStatus != types.LoanStatusAlreadyCleared &&
		order.CheckStatus != types.LoanStatusInvalid {

		loanLifetime = types.LoanLifetimeInProgress
		return
	}

	// 如果真的能走到这里,说明逻辑有漏洞,那么,需要修修补补了...Oo.
	orderJSON, _ := tools.JsonEncode(order)
	logs.Error("[GetLoanLifetime] order status has exception. order:", orderJSON)

	return
}

func calcRepayInfo(loan int64, product models.Product) (int64, int64, int64) {
	var total, interest, serviceFee int64
	if product.ChargeFeeType == types.ProductChargeFeeInterestBefore {
		//如果是砍头息
		total = int64(math.Ceil(float64(loan)/((1-0.014*float64(product.Period))*100)) * 100)
		interest = (total * 1 / 100) * int64(product.Period)
		serviceFee = total - loan - interest
	} else {
		interest = int64(loan / 100 * int64(product.Period))
		serviceFee = int64(loan / 1000 * 4 * int64(product.Period))
		total = loan + interest + serviceFee
	}
	return total, interest, serviceFee
}

// CompareIDPhotoAndLivingEnv 身份证与活体全景比对
func CompareIDPhotoAndLivingEnv(accountID int64) (similarity float64) {
	//最新活体认证全景照片
	livingModel, _ := dao.CustomerLiveVerify(accountID)
	livingPhotoTmp := gaws.BuildTmpFilename(livingModel.ImageEnv)
	livingResource, _ := OneResource(livingModel.ImageEnv)
	_, err1 := gaws.AwsDownload(livingResource.HashName, livingPhotoTmp)
	defer tools.Remove(livingPhotoTmp)
	if err1 != nil {
		logs.Error("[CompareIDPhotoAndLivingEnv] Dowload resource from aws has wrong. livingResource:", livingResource, "AccountID:", accountID)
		return
	}
	//身份证照片
	accountProfile, _ := dao.GetAccountProfile(accountID)
	IdPhotoResource, _ := OneResource(accountProfile.IdPhoto)
	IDPhotoTmp := gaws.BuildTmpFilename(accountProfile.IdPhoto)
	_, err2 := gaws.AwsDownload(IdPhotoResource.HashName, IDPhotoTmp)
	defer tools.Remove(IDPhotoTmp)

	if err2 != nil {
		logs.Error("[SaveLoanIDHeadAndLivingEnvCompare] Dowload resource from aws has wrong. IdPhotoResource:", livingResource, "AccountID:", accountID)
		return
	}
	//比对照片
	similarity, _ = advance.FaceComparison(accountID, livingPhotoTmp, IDPhotoTmp)
	return
}

// SaveLoanIDHeadAndLivingEnvCompare 比对活体全景照与首贷手持照并保持结果
// 比对结果在风控配置不需要过电核的基础上判断到底是否需要过电核的标准之一，与随机数一起

// 11.07 因首贷手持可以随意传，所以修改为复贷全景与首贷全景比对
func SaveLoanIDHeadAndLivingEnvCompare(accountID, orderID int64) (similarityF float64, compareType string) {
	order, _ := models.GetOrder(orderID)
	org := order
	if order.IsReloan == 1 {
		// similarityF = 0.01
		//最新活体认证全景照片
		livingModel, _ := dao.CustomerLiveVerify(accountID)
		livingPhotoTmp := gaws.BuildTmpFilename(livingModel.ImageEnv)
		livingResource, _ := OneResource(livingModel.ImageEnv)
		_, err1 := gaws.AwsDownload(livingResource.HashName, livingPhotoTmp)

		if err1 != nil {
			logs.Error("[SaveLoanIDHeadAndLivingEnvCompare] Dowload resource from aws has wrong. livingResource:", livingResource, "orderID:", orderID, "AccountID:", accountID)
			tools.Remove(livingPhotoTmp)
			return
		}

		//上一次有效订单全景照片
		firstLivingModel, err := dao.CustomerPrevLiveVerify(accountID)

		if firstLivingModel.Id > 0 && err == nil {
			compareType = "firstenv_reloanenv_similar"
			firstLivingPhotoTmp := gaws.BuildTmpFilename(firstLivingModel.ImageEnv)
			firstLivingResource, _ := OneResource(firstLivingModel.ImageEnv)
			_, err1 = gaws.AwsDownload(firstLivingResource.HashName, firstLivingPhotoTmp)

			if err1 != nil {
				logs.Error("[SaveLoanIDHeadAndLivingEnvCompare] Dowload resource from aws has wrong. firstLivingResource:", livingResource, "orderID:", orderID, "AccountID:", accountID)
				tools.Remove(livingPhotoTmp)
				tools.Remove(firstLivingPhotoTmp)
				return
			}
			//比对照片
			similarityF, _ = advance.FaceComparison(accountID, livingPhotoTmp, firstLivingPhotoTmp)
			//删除资源文件
			tools.Remove(livingPhotoTmp)
			tools.Remove(firstLivingPhotoTmp)
		} else {
			compareType = "first_idhand_reloanenv_similar"
			//首贷手持
			accountProfile, _ := dao.GetAccountProfile(accountID)
			handIdPhotoResource, _ := OneResource(accountProfile.HandHeldIdPhoto)
			handPhotoTmp := gaws.BuildTmpFilename(accountProfile.HandHeldIdPhoto)
			_, err2 := gaws.AwsDownload(handIdPhotoResource.HashName, handPhotoTmp)

			if err2 != nil {
				logs.Error("[SaveLoanIDHeadAndLivingEnvCompare] Dowload resource from aws has wrong. handIdPhotoResource:", handIdPhotoResource, "orderID:", orderID, "AccountID:", accountID)
				tools.Remove(livingPhotoTmp)
				tools.Remove(handPhotoTmp)
				return
			}
			//比对照片
			similarityF, _ = advance.FaceComparison(accountID, livingPhotoTmp, handPhotoTmp)
			//删除资源文件
			tools.Remove(livingPhotoTmp)
			tools.Remove(handPhotoTmp)

		}

		order.LivingbestReloanhandSimilar = tools.Float642Str(similarityF)
		order.Utime = tools.GetUnixMillis()
		// similarityF += 0.02
		models.UpdateOrder(&order)
		logs.Debug("[SaveLoanIDHeadAndLivingEnvCompare] ---happend err when update: order:%v ,order.LivingbestReloanhandSimilar:%s,similarityF:%g ,orderID:%d,accountid:%d", order, order.LivingbestReloanhandSimilar, similarityF, orderID, accountID)
		models.OpLogWrite(accountID, orderID, models.OpCodeOrderUpdate, order.TableName(), org, order)
		// similarityF += 0.02

	}
	similarityF += 0.022
	return
}

func CreateOrder(accountId, productId int64, loan int64, period, isTemporary int) (code cerror.ErrCode, orderId int64, err error) {
	// 正式上线,需要多检查一步
	product, err := models.GetProduct(productId)
	if err != nil {
		err = fmt.Errorf("can not find product: %d, accountId: %d", productId, accountId)
		logs.Error("[CreateOrder has wrong] err:", err)
		code = cerror.ProductDoesNotExist
		return
	}

	total, _, _ := repayplan.CalcRepayInfoV2(loan, product, period)

	//创建订单冗余标识是否为复贷
	isReloan := dao.IsRepeatLoan(accountId)
	reloan := 0
	if isReloan {
		reloan = 1
	}

	orderNew := models.Order{
		UserAccountId: accountId,
		ProductId:     productId,
		ProductIdOrg:  productId,
		Loan:          loan,
		LoanOrg:       loan,
		Amount:        total,
		Period:        period,
		PeriodOrg:     period,
		IsReloan:      reloan,
		CheckStatus:   types.LoanStatusSubmit,
		IsTemporary:   isTemporary,
		Ctime:         tools.GetUnixMillis(),
		Utime:         tools.GetUnixMillis(),
	}

	// 取临时订单之前加分布式锁，防止压测时重复创建临时订单
	// +1 分布式锁
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	lockKey := beego.AppConfig.String("create_order_lock")
	lockKey = fmt.Sprintf("%s:%d", lockKey, accountId)
	lock, err := cacheClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", 60, "NX") // 60秒分布式锁
	if err != nil || lock == nil {
		logs.Error("[CreateOrder] process is working, so return. accountId:%d", accountId)
		err = fmt.Errorf("[CreateOrder] lock err:%v", err)
		return cerror.CreateOrderFail, 0, err
	}
	defer cacheClient.Do("DEL", lockKey)

	orderData, err := dao.AccountLastLoanOrder(accountId)
	if err != nil {
		// 之前从来没有过有效订单,则创建新订单
		goto CREATE_ORDER
	}

	// 最后有效订单被拒超过7天
	if orderData.CheckStatus == types.LoanStatusReject && tools.GetUnixMillis()-orderData.CheckTime >= 360000*24*7 {
		goto CREATE_ORDER
	}

	// 有未完结订单,不能再创建新订单,哪怕是临时订单
	if orderData.CheckStatus != types.LoanStatusAlreadyCleared && orderData.CheckStatus != types.LoanStatusInvalid {
		code = cerror.UnsettledOrders
		orderJSON, _ := tools.JsonEncode(orderData)
		err = fmt.Errorf("there are unsettled orders. orderData: %s", orderJSON)
		logs.Warn("[CreateOrder has wrong] err:", err)
		return
	}

CREATE_ORDER:
	// 有完结订单的复贷情况
	////! 如果存在同金额/周期的临时订单,则利用之前的,不生成新的订单,仅修改`utime`
	////! 后续会根据最后修改时间来确定将哪个订单改为正式有效订单
	o := orm.NewOrm()
	o.Using(orderNew.Using())

	oldOrder, err := dao.OneTemporaryLoanOrder(accountId, loan, period)
	// 找到同条件历史临时订单,更改`utime`
	if err == nil {
		code = cerror.CodeSuccess
		orderId = oldOrder.Id
		oldOrder.Utime = tools.GetUnixMillis()
		o.Update(&oldOrder, "utime")
		return
	}

	orderId, err = device.GenerateBizId(types.OrderSystem)
	orderNew.Id = orderId
	_, err = orderNew.AddOrder(&orderNew)
	if err != nil {
		code = cerror.ServiceUnavailable
		logs.Error("[orderNew.AddOrder CreateOrder has wrong] err:", err)
		return
	}
	code = cerror.CodeSuccess
	// 用户借款订单申请事件触发
	event.Trigger(&evtypes.OrderApplyEv{
		OrderID:   orderId,
		AccountID: accountId,
		Time:      tools.GetUnixMillis(),
	})

	monitor.IncrOrderCount(orderNew.CheckStatus)

	//更新账号额度配置，防止极端复贷情况下没有额度配置
	//额度配置由风控每天更新，如果用户当日复贷则没有额度配置，所以在创建订单时更新默认额度配置
	go InsertDefaultQuotaConf(accountId)

	return
}

func CreateRollOrder(accountId int64) error {
	orderData, err := dao.AccountLastOverdueLoanOrder(accountId)
	if err != nil {
		return err
	}

	originOrder := orderData

	if orderData.CheckStatus != types.LoanStatusOverdue {
		str := fmt.Sprintf("[CreateRollOrder] order check status wrong orderId:%d, status:%d", orderData.Id, orderData.CheckStatus)
		logs.Error(str)
		return fmt.Errorf(str)
	}

	if !IsOrderCanRoll(orderData) {
		str := fmt.Sprintf("[CreateRollOrder] order can not roll orderId:%d", orderData.Id)
		logs.Error(str)
		return fmt.Errorf(str)
	}

	p, err := ProductRollSuitables()
	if err != nil {
		logs.Error("[CreateRollOrder] ProductSuitablesByPeriod error, orderid:%d, err:%v", orderData.Id, err)
		return err
	}

	period := p.MinPeriod
	_, minRepay, total, err := CalcRollRepayAmount(orderData)
	logs.Info("[CreateRollOrder] CalcRollRepayAmount orderid:%d, minRepay:%d, total:%d", orderData.Id, minRepay, total)
	if err != nil {
		logs.Error("[CreateRollOrder] CalcRollRepayAmount error, orderid:%d, err:%v", orderData.Id, err)
		return err
	}

	orderNew := models.Order{
		UserAccountId: accountId,
		ProductId:     p.Id,
		// Loan:          orderData.Loan,
		Amount:      total,
		Period:      period,
		PreOrder:    originOrder.Id,
		RollTimes:   originOrder.RollTimes + 1,
		CheckStatus: types.LoanStatusRollApply,
		IsTemporary: types.IsTemporaryYes,
		IsReloan:    orderData.IsReloan,
		ApplyTime:   tools.GetUnixMillis(),
		Ctime:       tools.GetUnixMillis(),
		Utime:       tools.GetUnixMillis(),
	}

	o := orm.NewOrm()
	o.Using(orderNew.Using())

	orderId, _ := device.GenerateBizId(types.OrderSystem)
	orderNew.Id = orderId

	orderData.MinRepayAmount = minRepay
	orderData.CheckStatus = types.LoanStatusRolling
	orderData.Utime = tools.GetUnixMillis()

	// 开始事务
	o.Begin()
	_, err = o.Insert(&orderNew)
	if err != nil {
		o.Rollback()
		logs.Error("[CreateRollOrder] Create roll order failed.", err)
		return err
	}

	_, err = o.Update(&orderData)
	if err != nil {
		o.Rollback()
		logs.Error("[CreateRollOrder] Update origin order failed.", err)
		return err
	}

	err = o.Commit()
	if err != nil {
		o.Rollback()
		logs.Error("[CreateRollOrder] Commit modify order failed.", err)
		return err
	}

	HandleOverdueCase(orderData.Id)

	monitor.IncrOrderCount(orderNew.CheckStatus)

	monitor.IncrOrderCount(orderData.CheckStatus)
	models.OpLogWrite(0, orderData.Id, models.OpCodeOrderUpdate, orderData.TableName(), originOrder, orderData)

	accountBase, _ := models.OneAccountBaseByPkId(accountId)

	param := make(map[string]interface{})
	param["related_id"] = orderData.Id
	schema_task.SendBusinessMsg(types.SmsTargetRollApplySuccess, types.ServiceRollApplySuccess, accountBase.Mobile, param)

	schema_task.PushBusinessMsg(types.PushTargetRollApplySuccess, orderData.UserAccountId)

	/*
		//获取VA账户信息
		userEAccount, eAccountErr := dao.GetActiveEaccountWithBankName(orderData.UserAccountId)
		if eAccountErr == nil {
			//xenditCallback := models.GetXenditCallBack(userEAccount.CallbackJson)

			smsContent := fmt.Sprintf(i18n.GetMessageText(i18n.TextRollApplySuccess), orderData.MinRepayAmount, userEAccount.BankCode, userEAccount.EAccountNumber)

			sms.Send(types.ServiceRollApplySuccess, accountBase.Mobile, smsContent, orderData.Id)
		}
	*/
	return nil
}

//后台借款管理
func OrderListBackend(condStr map[string]interface{}, page, pagesize int) (maps []BorrowCreditBackendData, total int64) {
	o := orm.NewOrm()
	order := models.Order{}
	o.Using(order.UsingSlave())
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	cond := "1 = 1"

	if f, ok := condStr["id"]; ok {
		cond = fmt.Sprintf("%s AND orders.id = %d", cond, f.(int64))
	}
	if f, ok := condStr["order_type"]; ok {
		var isTemporaryBox []string
		var isReloanBox []string
		for _, typeKey := range f.([]string) {
			// 正常普通订单
			if typeKey == "normal" {
				isTemporaryBox = append(isTemporaryBox, fmt.Sprintf("%d", types.IsTemporaryNO))
			}
			// 临时订单
			if typeKey == "temporary" {
				isTemporaryBox = append(isTemporaryBox, fmt.Sprintf("%d", types.IsTemporaryYes))
			}
			// 首/复贷
			if typeKey == "first" {
				isReloanBox = append(isReloanBox, fmt.Sprintf("%d", 0))
			}
			if typeKey == "repeat" {
				isReloanBox = append(isReloanBox, fmt.Sprintf("%d", 1))
			}
			// 展单
			if typeKey == "roll" {
				cond = fmt.Sprintf("%s AND orders.pre_order > 0", cond)
			}
			// 历史逾期
			if typeKey == "overdue" {
				cond = fmt.Sprintf("%s AND orders.is_overdue = 1", cond)
			}
			// 坏帐
			if typeKey == "dead_debt" {
				cond = fmt.Sprintf("%s AND orders.is_dead_debt = 1", cond)
			}
		}
		if len(isTemporaryBox) > 0 {
			cond = fmt.Sprintf("%s AND orders.is_temporary IN (%s)", cond, strings.Join(isTemporaryBox, ", "))
		}
		if len(isReloanBox) > 0 {
			cond = fmt.Sprintf("%s AND orders.is_reloan IN (%s)", cond, strings.Join(isReloanBox, ", "))
		}
	}
	if f, ok := condStr["realname"]; ok {
		cond = fmt.Sprintf("%s AND account_base.realname LIKE '%%%s%%'", cond, tools.Escape(f.(string)))
	}
	if f, ok := condStr["check_status"]; ok {
		checkStatusArr := f.([]string)
		if len(checkStatusArr) > 0 {
			cond = fmt.Sprintf("%s  AND orders.check_status IN(%s)", cond, strings.Join(checkStatusArr, ", "))
		}
	}
	if f, ok := condStr["apply_start_time"]; ok {
		cond = fmt.Sprintf("%s AND orders.apply_time >= %d", cond, f.(int64))
	}
	if f, ok := condStr["apply_end_time"]; ok {
		cond = fmt.Sprintf("%s AND orders.apply_time < %d", cond, f.(int64))
	}
	if f, ok := condStr["ctime_start_time"]; ok {
		cond = fmt.Sprintf("%s AND orders.ctime >= %d", cond, f.(int64))
	}
	if f, ok := condStr["ctime_end_time"]; ok {
		cond = fmt.Sprintf("%s AND orders.ctime < %d", cond, f.(int64))
	}
	if f, ok := condStr["user_account_id"]; ok {
		cond = fmt.Sprintf("%s AND orders.user_account_id = %d", cond, f.(int64))
	}
	if f, ok := condStr["mobile"]; ok {
		cond = fmt.Sprintf("%s AND account_base.mobile = '%s'", cond, f.(string))
	}
	// 软删除功能,金管局专用,先下线
	//cond = fmt.Sprintf("%s AND orders.is_deleted = 0 And account_base.is_deleted = 0", cond)

	sql := `SELECT COUNT(orders.id) FROM orders
LEFT JOIN account_base ON orders.user_account_id = account_base.id
WHERE ` + cond

	sqlOrder := fmt.Sprintf(`SELECT orders.id, orders.user_account_id, orders.amount,
orders.loan, orders.period, orders.check_status, orders.apply_time, orders.check_time,
orders.repay_time, orders.loan_time, orders.finish_time, orders.ctime, orders.is_temporary,
orders.random_mark AS order_random_mark, orders.is_reloan, orders.risk_ctl_status,
orders.pre_order, orders.is_overdue, orders.is_dead_debt,
account_base.realname, account_base.random_mark, account_base.mobile
FROM orders LEFT JOIN account_base ON orders.user_account_id = account_base.id
WHERE %s`, cond)

	orderBy := ""
	if v, ok := condStr["field"]; ok {
		if vF, okF := orderFieldMap[v.(string)]; okF {
			orderBy = "ORDER BY " + vF
		} else {
			orderBy = "ORDER BY orders.id"
		}
	} else {
		orderBy = "ORDER BY orders.id"
	}

	if v, ok := condStr["sort"]; ok {
		orderBy = fmt.Sprintf("%s %s", orderBy, v.(string))
	} else {
		orderBy = fmt.Sprintf("%s %s", orderBy, "DESC")
	}

	if len(condStr) == 0 {
		return
	}

	limit := fmt.Sprintf("LIMIT %d, %d", offset, pagesize)
	sqlOrder = fmt.Sprintf("%s %s %s", sqlOrder, orderBy, limit)

	r := o.Raw(sql)
	r.QueryRow(&total)

	r = o.Raw(sqlOrder)
	r.QueryRows(&maps)

	return
}

//后台放款管理
func LoanListBackend(condStr map[string]interface{}, page, pagesize int) (*[]GiveoutCreditBackendData, int64, int64, int64) {
	o := orm.NewOrm()
	order := models.Order{}
	o.Using(order.UsingSlave())
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	// 金管局SQL
	//cond := "1=1 AND orders.is_deleted = 0 AND account_base.is_deleted = 0"
	cond := "1=1"
	joinCount := ""
	joinList := " LEFT JOIN account_base ON orders.user_account_id = account_base.id LEFT JOIN account_profile ON orders.user_account_id = account_profile.account_id"

	if f, ok := condStr["id"]; ok {
		cond = fmt.Sprintf("%s%s%s%s", cond, " AND orders.id = '", tools.Escape(f.(string)), "'")
	}
	if f, ok := condStr["account_id"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.user_account_id = ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["realname"]; ok {
		cond = fmt.Sprintf("%s%s%s%s", cond, " AND account_base.realname like '%", tools.Escape(f.(string)), "%'")
		if !strings.Contains(joinCount, "account_base") {
			joinCount = fmt.Sprintf("%s%s", joinCount, " LEFT JOIN account_base ON orders.user_account_id = account_base.id")
		}
	}
	if f, ok := condStr["mobile"]; ok {
		cond = fmt.Sprintf("%s%s%s%s", cond, " AND account_base.mobile = '", f.(string), "'")
		if !strings.Contains(joinCount, "account_base") {
			joinCount = fmt.Sprintf("%s%s", joinCount, " LEFT JOIN account_base ON orders.user_account_id = account_base.id")
		}
	}
	if f, ok := condStr["bankname"]; ok {
		cond = fmt.Sprintf("%s%s%s%s", cond, " AND account_profile.bank_name = '", f.(string), "'")
		if !strings.Contains(joinCount, "account_profile") {
			joinCount = fmt.Sprintf("%s%s", joinCount, " LEFT JOIN account_profile ON orders.user_account_id = account_profile.account_id")
		}
	}
	if f, ok := condStr["loan_channel"]; ok {
		//names := loanBankFullNameList(f.(int))
		//cond = fmt.Sprintf("%s%s%s%s", cond, " AND account_profile.bank_name in(", names, ")")
		//if !strings.Contains(joinCount, "account_profile") {
		//	joinCount = fmt.Sprintf("%s%s", joinCount, " LEFT JOIN account_profile ON orders.user_account_id = account_profile.account_id")
		//}

		qId := `SELECT order_id FROM disburse_invoke_log WHERE id IN ( SELECT max(id) FROM disburse_invoke_log GROUP BY order_id ) AND va_company_code = '%d' `
		qId = fmt.Sprintf(qId, f.(int))
		cond = fmt.Sprintf("%s%s%s%s", cond, " AND orders.id IN (", qId, ")")
	}
	if f, ok := condStr["failed_code"]; ok {
		str := types.FailureCodeMap()[f.(int)]
		qId := `SELECT order_id FROM disburse_invoke_log WHERE id IN ( SELECT max(id) FROM disburse_invoke_log WHERE failure_code != '' and order_id in( select id from orders where check_status = 6) GROUP BY order_id ) AND failure_code = '%s' `
		qId = fmt.Sprintf(qId, str)
		cond = fmt.Sprintf("%s%s%s%s", cond, " AND orders.id IN (", qId, ")")
		//if !strings.Contains(joinCount, "disburse_invoke_log") {
		//	joinCount = fmt.Sprintf("%s%s", joinCount, " LEFT JOIN disburse_invoke_log ON orders.id = disburse_invoke_log.order_id")
		//	joinList = fmt.Sprintf("%s%s", joinList, " LEFT JOIN disburse_invoke_log ON orders.id = disburse_invoke_log.order_id")
		//}
	}

	loanStatusMap := types.LoanStatusMap()
	keys := make([]int, 0, len(loanStatusMap))
	for k := range loanStatusMap {
		keys = append(keys, int(k))
	}
	orderStatusStr := tools.ArrayToString(keys, ",")
	orderStatusStr = fmt.Sprintf("(%s)", orderStatusStr)

	if f, ok := condStr["check_status"]; ok {
		checkStatusArr := f.([]string)
		checkStatusStr := tools.ArrayToString(checkStatusArr, ",")
		checkStatusStr = fmt.Sprintf("(%s)", checkStatusStr)
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.check_status IN ", checkStatusStr)
	} else {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.check_status IN ", orderStatusStr)
	}

	if f, ok := condStr["apply_start_time"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.apply_time >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["apply_end_time"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.apply_time < ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["loan_start_time"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.loan_time >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["loan_end_time"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.loan_time < ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["finish_start_time"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.finish_time >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["finish_end_time"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.finish_time < ", strconv.FormatInt(f.(int64), 10))
	}

	sql := "SELECT COUNT(orders.id) FROM orders "
	sqlOrder := "SELECT orders.*,account_profile.account_id,account_base.realname,account_profile.bank_name,account_profile.bank_no FROM orders "

	orderBy := ""
	if v, ok := condStr["field"]; ok {
		if vF, okF := loanFieldMap[v.(string)]; okF {
			orderBy = "ORDER BY " + vF
		} else {
			orderBy = "ORDER BY orders.id"
		}
	} else {
		orderBy = "ORDER BY orders.id"
	}

	if v, ok := condStr["sort"]; ok {
		orderBy = fmt.Sprintf("%s %s", orderBy, v.(string))
	} else {
		orderBy = fmt.Sprintf("%s %s", orderBy, "DESC")
	}

	sql = fmt.Sprintf("%s %s WHERE %s", sql, joinCount, cond)
	sqlOrder = fmt.Sprintf("%s %s WHERE %s %s LIMIT ", sqlOrder, joinList, cond, orderBy)
	sqlOrder = fmt.Sprintf("%s%d%s%d", sqlOrder, offset, ",", pagesize)
	// 放款总金额  统计状态为 6:放款失败 7:等待还款 8:已结清 9:逾期 11:部分还款
	//sqlLoanTotal := "SELECT SUM(orders.loan) FROM orders LEFT JOIN account_base ON orders.user_account_id=account_base.id WHERE " + cond + " AND orders.check_status in (6, 7, 8, 9, 11)"
	// 放款成功总金额  统计状态为 7:等待还款 8:已结清 9:逾期 11:部分还款
	//sqlLoanTotalSuccess := "SELECT SUM(orders.loan) FROM orders LEFT JOIN account_base ON orders.user_account_id=account_base.id WHERE " + cond + " AND orders.check_status in ( 7, 8, 9, 11)"

	r := o.Raw(sql)
	var total int64
	r.QueryRow(&total)

	r = o.Raw(sqlOrder)
	var maps []GiveoutCreditBackendData
	r.QueryRows(&maps)

	//r = o.Raw(sqlLoanTotal)
	var loanTotal int64
	//r.QueryRow(&loanTotal)
	//
	//r = o.Raw(sqlLoanTotalSuccess)
	var loanTotalSuccess int64
	//r.QueryRow(&loanTotalSuccess)

	return &maps, total, loanTotal, loanTotalSuccess
}

// RepayListBackend 后台还款管理列表页面
func RepayListBackend(condStr map[string]interface{}, page, pagesize int) (maps []RepayBackendData, total int64, totalRepay int64, totalRepayPayed int64, totalRepayReduce int64) {
	o := orm.NewOrm()
	order := models.Order{}
	o.Using(order.UsingSlave())
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	// 金管局SQL
	//cond := "1=1 AND orders.is_deleted = 0 AND account_base.is_deleted = 0"
	cond := "1 = 1"

	repayStatusMap := types.RepayStatusMap()
	keys := make([]int, 0, len(repayStatusMap))
	for k := range repayStatusMap {
		keys = append(keys, int(k))
	}
	orderStatusStr := tools.ArrayToString(keys, ",")
	orderStatusStr = fmt.Sprintf("(%s)", orderStatusStr)

	if f, ok := condStr["id"]; ok {
		cond = fmt.Sprintf("%s%s%s%s", cond, " AND orders.id = '", tools.Escape(f.(string)), "'")
	}
	if f, ok := condStr["account_id"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND account_base.id = ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["realname"]; ok {
		cond = fmt.Sprintf("%s%s%s%s", cond, " AND account_base.realname like '%", tools.Escape(f.(string)), "%'")
	}
	if f, ok := condStr["check_status"]; ok {
		if checks, ok := f.([]string); ok && len(checks) > 0 {
			for k, check := range checks {
				if k == 0 {
					cond = fmt.Sprintf("%s%s%s", cond, " AND orders.check_status IN (", check)
				}
				if k == len(checks)-1 {
					cond = fmt.Sprintf("%s%s%s%s", cond, ", ", check, ")")
				} else {
					cond = fmt.Sprintf("%s%s%s", cond, ", ", check)
				}
			}
		}

	} else {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.check_status IN ", orderStatusStr)
	}

	if f, ok := condStr["apply_start_time"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.apply_time >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["apply_end_time"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.apply_time < ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["repay_start_date"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND repay_plan.repay_date >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["repay_end_date"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND repay_plan.repay_date < ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["repay_time_start"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.repay_time >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["repay_time_end"]; ok {
		cond = fmt.Sprintf("%s%s%s", cond, " AND orders.repay_time < ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["finish_time_start"]; ok {
		cond = cond + fmt.Sprintf(" AND orders.finish_time >=%d ", f)
	}
	if f, ok := condStr["finish_time_end"]; ok {
		cond = cond + fmt.Sprintf(" AND orders.finish_time <%d ", f)
	}
	if f, ok := condStr["mobile"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND account_base.mobile = ", f.(int64))
	}
	if f, ok := condStr["left_amount"]; ok {
		cond = fmt.Sprintf("%s%s%d%s", cond, " AND ((repay_plan.amount + repay_plan.grace_period_interest + repay_plan.penalty - repay_plan.amount_payed - repay_plan.amount_reduced - repay_plan.grace_period_interest_payed - repay_plan.grace_period_interest_reduced - repay_plan.penalty_payed - repay_plan.penalty_reduced) between 1 and ", f.(int64), ")")
	}

	leftJoin := " LEFT JOIN account_base ON orders.user_account_id=account_base.id"
	// 表名太长，user_e_account 已加别名 uea
	//leftJoin += fmt.Sprintf(" LEFT JOIN user_e_account uea ON uea.user_account_id = account_base.id AND uea.va_company_code = 1 AND uea.status='ACTIVE'")
	leftJoin += " LEFT JOIN repay_plan ON repay_plan.order_id = orders.id "

	sql := "SELECT COUNT(orders.id) FROM orders"
	sql += leftJoin + " WHERE " + cond
	sqlOrder := `SELECT orders.id,orders.user_account_id,
		account_base.realname,repay_plan.amount,repay_plan.amount_payed,repay_plan.amount_reduced,
		repay_plan.grace_period_interest,repay_plan.grace_period_interest_payed,
		repay_plan.grace_period_interest_reduced,repay_plan.penalty,repay_plan.penalty_payed,
		repay_plan.penalty_reduced,orders.loan,orders.period,orders.check_status,orders.apply_time,orders.check_time,
		repay_plan.repay_date,orders.loan_time,orders.repay_time,orders.finish_time FROM orders`
	sqlOrder += leftJoin + " WHERE " + cond

	orderBy := ""
	if v, ok := condStr["field"]; ok {
		if vF, okF := repayFieldMap[v.(string)]; okF {
			orderBy = "ORDER BY " + vF
		} else {
			orderBy = "ORDER BY orders.id"
		}
	} else {
		orderBy = "ORDER BY orders.id"
	}

	if v, ok := condStr["sort"]; ok {
		orderBy = fmt.Sprintf("%s %s", orderBy, v.(string))
	} else {
		orderBy = fmt.Sprintf("%s %s", orderBy, "DESC")
	}

	if len(condStr) == 0 {
		return
	}

	sqlOrder = fmt.Sprintf("%s %s LIMIT ", sqlOrder, orderBy)

	sqlOrder = fmt.Sprintf("%s%d%s%d", sqlOrder, offset, ",", pagesize)
	sqlTotalRepay := "SELECT SUM(repay_plan.amount + repay_plan.grace_period_interest + repay_plan.penalty) FROM orders"
	sqlTotalRepay += leftJoin + " WHERE " + cond
	sqlTotalRepayPayed := "SELECT SUM(repay_plan.amount_payed + repay_plan.grace_period_interest_payed + repay_plan.penalty_payed) FROM orders"
	sqlTotalRepayPayed += leftJoin + " WHERE " + cond

	r := o.Raw(sql)
	r.QueryRow(&total)

	r = o.Raw(sqlOrder)
	r.QueryRows(&maps)

	r = o.Raw(sqlTotalRepay)
	r.QueryRow(&totalRepay)

	r = o.Raw(sqlTotalRepayPayed)
	r.QueryRow(&totalRepayPayed)

	// 查询所有满足条件的记录 的减免总额
	sqlOrder = "SELECT orders.id FROM orders"
	sqlOrder += leftJoin + " WHERE " + cond
	sqlOrder += " GROUP BY orders.id"
	sqlTotalRepayRecuce := "SELECT	sum(repay_plan.amount_reduced + repay_plan.penalty_reduced + repay_plan.grace_period_interest_reduced	) as reduce_total FROM repay_plan  where order_id IN "
	sqlTotalRepayRecuce += "("
	sqlTotalRepayRecuce += sqlOrder
	sqlTotalRepayRecuce += ")"

	r = o.Raw(sqlTotalRepayRecuce)
	r.QueryRow(&totalRepayReduce)

	return
}

func eA2RepayVaDisplay(eA models.User_E_Account) (one RepayVaDisplay) {
	one.UserAccountId = eA.UserAccountId
	one.Code = eA.BankCode + " " + eA.EAccountNumber
	one.ApplyTime = eA.Ctime
	one.CompanyCode = eA.VaCompanyCode
	if one.CompanyCode == types.DoKu {
		one.Code = doku.DoKuVaBankCodeTransform(eA.BankCode) + " " + eA.EAccountNumber
	}
	return
}

func mP2RepayVaDisplay(mP models.MarketPayment) (one RepayVaDisplay) {
	one.UserAccountId = mP.UserAccountId
	one.OrderId = mP.OrderId
	one.Code = mP.PaymentCode
	one.ApplyTime = mP.Ctime
	one.ExpireTime = mP.ExpiryDate
	one.CompanyCode = types.Xendit
	one.Amount = mP.Amount
	return
}

func FpRepayVaDisplay(fP models.FixPaymentCode) (one RepayVaDisplay) {
	one.UserAccountId = fP.UserAccountId
	one.OrderId = fP.OrderId
	one.Code = fP.PaymentCode
	one.ApplyTime = fP.Utime
	one.ExpireTime = fP.ExpirationDate
	one.CompanyCode = types.Xendit
	one.Amount = fP.ExpectedAmount
	return
}

func vaSearchSingle(condStr map[string]interface{}) (maps []RepayVaDisplay) {

	if f, ok := condStr["va_code"]; ok {
		eA, err := models.GetEAccountByENumber(f.(string))
		if err != nil {
			logs.Error("[vaSearchSingle] GetEAccountByENumber err:%v condStr:%#v", err, condStr)
			return
		}
		one := eA2RepayVaDisplay(eA)
		maps = append(maps, one)
	}

	if f, ok := condStr["payment_code"]; ok {
		oFp, _ := models.OneFixPaymentCodeByPaymentCode(f.(string))
		if oFp.UserAccountId != 0 {
			//如果数据库存在记录
			oneRepayVaDisplay := FpRepayVaDisplay(oFp)
			maps = append(maps, oneRepayVaDisplay)
		} else {
			mP, _ := models.GetMarketPaymentByPaymentCode(f.(string))
			if mP.UserAccountId > 0 {
				one := mP2RepayVaDisplay(mP)
				maps = append(maps, one)
			}
		}
	}

	return
}

func condByMap(condStr map[string]interface{}) *orm.Condition {
	cond := orm.NewCondition()
	if f, ok := condStr["account_id"]; ok {
		cond = cond.And("user_account_id", f.(int64))
	}
	if f, ok := condStr["apply_start_time"]; ok {
		cond = cond.And("ctime__gte", f.(int64))
	}
	if f, ok := condStr["apply_end_time"]; ok {
		cond = cond.And("ctime__lt", f.(int64))
	}
	if f, ok := condStr["mobile"]; ok {
		aB, err := models.OneAccountBaseByMobile(f.(string))
		if err == nil {
			cond = cond.And("user_account_id", aB.Id)
		}
	}
	return cond
}

func allVa(condStr map[string]interface{}) (all []models.User_E_Account) {
	eA := models.User_E_Account{}
	o := orm.NewOrm()
	o.Using(eA.UsingSlave())

	cond := condByMap(condStr)
	if f, ok := condStr["id"]; ok {
		order, err := models.GetOrder(f.(int64))
		if err == nil {
			cond = cond.And("user_account_id", order.UserAccountId)
		}
	}

	if cond.IsEmpty() {
		return
	}

	cond = cond.And("status", "ACTIVE")
	o.QueryTable(eA.TableName()).SetCond(cond).OrderBy("-id").All(&all)
	return
}

func allMarketPay(condStr map[string]interface{}) (all []models.MarketPayment) {
	mP := models.MarketPayment{}
	o := orm.NewOrm()
	o.Using(mP.UsingSlave())

	cond := condByMap(condStr)
	if f, ok := condStr["id"]; ok {
		cond = cond.And("order_id", f.(int64))
	}

	if cond.IsEmpty() {
		return
	}

	o.QueryTable(mP.TableName()).SetCond(cond).OrderBy("-id").All(&all)
	return
}

func oneFixPaymentCode(condStr map[string]interface{}) (one models.FixPaymentCode) {
	fP := models.FixPaymentCode{}
	o := orm.NewOrm()
	o.Using(fP.UsingSlave())

	cond := condByMap(condStr)
	if f, ok := condStr["id"]; ok {
		cond = cond.And("order_id", f.(int64))
	}

	if cond.IsEmpty() {
		return
	}

	o.QueryTable(fP.TableName()).SetCond(cond).One(&one)
	return
}

func vaSearchMulti(last []RepayVaDisplay, condStr map[string]interface{}, repayType int) []RepayVaDisplay {
	eAs := []models.User_E_Account{}
	mPs := []models.MarketPayment{}
	oFp := models.FixPaymentCode{}
	switch repayType {
	case types.RepayTypeVa:
		{
			eAs = allVa(condStr)
			logs.Info("len(eAs):%d", len(eAs))
			for _, ea := range eAs {
				one := eA2RepayVaDisplay(ea)
				last = append(last, one)
			}
		}
	case types.RepayTypePaymentCode:
		{
			oFp = oneFixPaymentCode(condStr)
			if oFp.UserAccountId != 0 {
				//如果数据库存在记录
				oneRepayVaDisplay := FpRepayVaDisplay(oFp)
				last = append(last, oneRepayVaDisplay)
			}

			mPs = allMarketPay(condStr)
			logs.Info("len(mPs):%d", len(mPs))
			for _, mp := range mPs {
				one := mP2RepayVaDisplay(mp)
				last = append(last, one)
			}
		}
	}

	return last
}

// RepayListBackend 后台还款管理列表页面
func RepayVaSearch(condStr map[string]interface{}) []RepayVaDisplay {
	if _, ok := condStr["va_code"]; ok {
		return vaSearchSingle(condStr)
	}
	if _, ok := condStr["payment_code"]; ok {
		return vaSearchSingle(condStr)
	}

	var maps []RepayVaDisplay
	if f, ok := condStr["repay_type"]; ok {
		if f.(int) != 0 {
			return vaSearchMulti(maps, condStr, f.(int))
		}
	}

	maps = vaSearchMulti(maps, condStr, types.RepayTypeVa)
	logs.Info("after va len(maps):%d", len(maps))
	maps = vaSearchMulti(maps, condStr, types.RepayTypePaymentCode)
	logs.Info("after mp len(maps):%d", len(maps))
	return maps
}

// 后台逾期管理
func OverdueListBackend(adminUID int64, condStr map[string]interface{}, page, pagesize int) (list []OverdueCaseListItem, total int64, err error) {
	o := orm.NewOrm()
	order := models.Order{}
	o.Using(order.UsingSlave())

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	//where := fmt.Sprintf("WHERE c.is_out = %d", types.IsUrgeOutNo)
	where := fmt.Sprintf("WHERE 1 = 1 AND a.is_deleted = 0 AND o.is_deleted = 0")
	if f, ok := condStr["realname"]; ok {
		where = fmt.Sprintf("%s AND a.realname LIKE '%%%s%%'", where, tools.Escape(f.(string)))
	}
	if f, ok := condStr["mobile"]; ok {
		where = fmt.Sprintf("%s AND a.mobile = '%s'", where, tools.Escape(f.(string)))
	}
	if f, ok := condStr["id"]; ok {
		where = fmt.Sprintf("%s AND c.id= %d", where, f.(int64))
	}
	if f, ok := condStr["order_id"]; ok {
		where = fmt.Sprintf("%s AND c.order_id = %d", where, f.(int64))
	}
	if f, ok := condStr["account_id"]; ok {
		where = fmt.Sprintf("%s AND a.id = %d", where, f.(int64))
	}
	if f, ok := condStr["filter"]; ok {
		where = fmt.Sprintf("%s AND c.is_out = %d", where, f.(int))
	}
	// 非超管,只能查看分配给他,或没有指派的案件
	if adminUID != types.SuperAdminUID {
		where = fmt.Sprintf("%s AND assign_uid IN(0, %d)", where, adminUID)
	}
	if f, ok := condStr["join_urge_time_start"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.join_urge_time >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["join_urge_time_end"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.join_urge_time < ", strconv.FormatInt(f.(int64), 10))
	}

	if f, ok := condStr["out_urge_time_start"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.out_urge_time >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["out_urge_time_end"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.out_urge_time < ", strconv.FormatInt(f.(int64), 10))
	}

	if f, ok := condStr["case_level"]; ok {
		where = fmt.Sprintf("%s AND c.case_level = '%s'", where, tools.Escape(f.(string)))
	}

	if f, ok := condStr["order_type"]; ok {
		var isReloanBox []string
		val := f.(string)
		switch val {
		case types.GetUrgeOrderTypeVal(int(types.UrgeOrderTypeFirst)):
			// 首贷
			isReloanBox = append(isReloanBox, fmt.Sprintf("%d", 0))
		case types.GetUrgeOrderTypeVal(int(types.UrgeOrderTypeRepeat)):
			// 复贷
			isReloanBox = append(isReloanBox, fmt.Sprintf("%d", 1))
		case types.GetUrgeOrderTypeVal(int(types.UrgeOrderTypeRoll)):
			// 展单
			where = fmt.Sprintf("%s AND o.pre_order > 0", where)

		}
		if len(isReloanBox) > 0 {
			where = fmt.Sprintf("%s AND o.is_reloan IN (%s)", where, strings.Join(isReloanBox, ", "))
		}
	}

	if f, ok := condStr["overdue_days_start"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.overdue_days >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["overdue_days_end"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.overdue_days <= ", strconv.FormatInt(f.(int64), 10))
	}

	if f, ok := condStr["left_amount"]; ok {
		where = fmt.Sprintf("%s AND ((r.amount + r.grace_period_interest + r.penalty - r.amount_payed - r.amount_reduced - r.grace_period_interest_payed - r.grace_period_interest_reduced - r.penalty_payed - r.penalty_reduced) between 1 and %d)", where, f.(int64))
	}

	accountBase := models.AccountBase{}
	overdueCase := models.OverdueCase{}
	repayPlan := models.RepayPlan{}
	accountProfile := models.AccountProfile{}

	sqlCount := "SELECT COUNT(c.id) AS total"
	sqlSelect := "SELECT c.*, a.id AS account_id, a.realname, a.mobile, r.amount, r.amount_payed,r.amount_reduced, r.repay_date, r.grace_period_interest, r.grace_period_interest_payed, r.grace_period_interest_reduced,r.penalty, r.penalty_payed,r.penalty_reduced, p.company_telephone, p.salary_day "

	/*
			// 添加承诺还款时间, 最近一次催收时间
			sqlSelect = fmt.Sprintf("%s%s", sqlSelect, ", tocd.promise_repay_time, tocd.phone_time ")

			// 获取逾期案例详单中,每个order_id相关的最新项
			ocdSql := `(select a.order_id, a.promise_repay_time, a.phone_time from overdue_case_detail as a, (select max(id) as maxid,order_id from overdue_case_detail group by order_id) as b where a.id = b.maxid)`
			from :=
				fmt.Sprintf(`FROM %s c
		LEFT JOIN %s r ON r.order_id = c.order_id
		LEFT JOIN %s o ON o.id = c.order_id
		LEFT JOIN %s a ON a.id = o.user_account_id
		LEFT JOIN %s p ON p.account_id = o.user_account_id
		LEFT JOIN %s as tocd ON tocd.order_id = c.order_id`, overdueCase.TableName(), repayPlan.TableName(),
					order.TableName(), accountBase.TableName(), accountProfile.TableName(), ocdSql)

	*/

	// 获取逾期案例详单中,每个order_id相关的最新项
	from :=
		fmt.Sprintf(`FROM %s c
LEFT JOIN %s r ON r.order_id = c.order_id
LEFT JOIN %s o ON o.id = c.order_id
LEFT JOIN %s a ON a.id = o.user_account_id
LEFT JOIN %s p ON p.account_id = o.user_account_id`, overdueCase.TableName(), repayPlan.TableName(),
			order.TableName(), accountBase.TableName(), accountProfile.TableName())

	sql := fmt.Sprintf(`%s %s %s`, sqlCount, from, where)
	r := o.Raw(sql)
	err = r.QueryRow(&total)
	if err != nil {
		return
	}

	orderBy := ""
	if v, ok := condStr["field"]; ok {
		if vF, okF := overdueFieldMap[v.(string)]; okF {
			orderBy = "ORDER BY " + vF
		} else {
			orderBy = "ORDER BY c.id"
		}
	} else {
		orderBy = "ORDER BY c.id"
	}

	if v, ok := condStr["sort"]; ok {
		orderBy = fmt.Sprintf("%s %s", orderBy, v.(string))
	} else {
		orderBy = fmt.Sprintf("%s %s", orderBy, "DESC")
	}

	limit := fmt.Sprintf(`LIMIT %d, %d`, offset, pagesize)

	sql = fmt.Sprintf(`%s %s %s %s %s`, sqlSelect, from, where, orderBy, limit)
	r = o.Raw(sql)
	_, err = r.QueryRows(&list)

	return
}

func OverdueCO2caseListBackend(adminUID int64, condStr map[string]interface{}, page, pagesize int) (list []OctwoCaseListItem, total int64, err error) {

	caseLevel := "M1-2"

	o := orm.NewOrm()
	order := models.Order{}
	o.Using(order.UsingSlave())

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	//where := fmt.Sprintf("WHERE c.is_out = %d", types.IsUrgeOutNo)
	where := fmt.Sprintf("WHERE 1 = 1 AND a.is_deleted = 0 AND o.is_deleted = 0")
	if f, ok := condStr["realname"]; ok {
		where = fmt.Sprintf("%s AND a.realname LIKE '%%%s%%'", where, tools.Escape(f.(string)))
	}
	if f, ok := condStr["mobile"]; ok {
		where = fmt.Sprintf("%s AND a.mobile = '%s'", where, tools.Escape(f.(string)))
	}
	if f, ok := condStr["id"]; ok {
		where = fmt.Sprintf("%s AND c.id= %d", where, f.(int64))
	}
	if f, ok := condStr["order_id"]; ok {
		where = fmt.Sprintf("%s AND c.order_id = %d", where, f.(int64))
	}
	if f, ok := condStr["account_id"]; ok {
		where = fmt.Sprintf("%s AND a.id = %d", where, f.(int64))
	}
	if f, ok := condStr["filter"]; ok {
		where = fmt.Sprintf("%s AND c.is_out = %d", where, f.(int))
	}
	// 非超管,只能查看分配给他,或没有指派的案件
	if adminUID != types.SuperAdminUID {
		where = fmt.Sprintf("%s AND assign_uid IN(0, %d)", where, adminUID)
	}
	if f, ok := condStr["join_urge_time_start"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.join_urge_time >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["join_urge_time_end"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.join_urge_time < ", strconv.FormatInt(f.(int64), 10))
	}

	if f, ok := condStr["out_urge_time_start"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.out_urge_time >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["out_urge_time_end"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.out_urge_time < ", strconv.FormatInt(f.(int64), 10))
	}

	where = fmt.Sprintf("%s AND c.case_level = '%s'", where, caseLevel)

	if f, ok := condStr["order_type"]; ok {
		var isReloanBox []string
		val := f.(string)
		switch val {
		case types.GetUrgeOrderTypeVal(int(types.UrgeOrderTypeFirst)):
			// 首贷
			isReloanBox = append(isReloanBox, fmt.Sprintf("%d", 0))
		case types.GetUrgeOrderTypeVal(int(types.UrgeOrderTypeRepeat)):
			// 复贷
			isReloanBox = append(isReloanBox, fmt.Sprintf("%d", 1))
		case types.GetUrgeOrderTypeVal(int(types.UrgeOrderTypeRoll)):
			// 展单
			where = fmt.Sprintf("%s AND o.pre_order > 0", where)

		}
		if len(isReloanBox) > 0 {
			where = fmt.Sprintf("%s AND o.is_reloan IN (%s)", where, strings.Join(isReloanBox, ", "))
		}
	}

	if f, ok := condStr["overdue_days_start"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.overdue_days >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["overdue_days_end"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.overdue_days <= ", strconv.FormatInt(f.(int64), 10))
	}

	if f, ok := condStr["left_amount"]; ok {
		where = fmt.Sprintf("%s AND ((r.amount + r.grace_period_interest + r.penalty - r.amount_payed - r.amount_reduced - r.grace_period_interest_payed - r.grace_period_interest_reduced - r.penalty_payed - r.penalty_reduced) between 1 and %d)", where, f.(int64))
	}

	accountBase := models.AccountBase{}
	overdueCase := models.OverdueCase{}
	repayPlan := models.RepayPlan{}
	accountProfile := models.AccountProfile{}
	ordersExt := models.OrderExt{}

	sqlCount := "SELECT COUNT(c.id) AS total"
	sqlSelect := "SELECT c.*, a.id AS account_id, a.realname, a.mobile, r.amount, r.amount_payed,r.amount_reduced, r.repay_date, r.grace_period_interest, r.grace_period_interest_payed, r.grace_period_interest_reduced,r.penalty, r.penalty_payed,r.penalty_reduced, p.company_telephone, p.salary_day,oe.entrust_pname,oe.is_entrust"

	// 获取逾期案例详单中,每个order_id相关的最新项
	from :=
		fmt.Sprintf(`FROM %s c
LEFT JOIN %s r ON r.order_id = c.order_id
LEFT JOIN %s o ON o.id = c.order_id
LEFT JOIN %s a ON a.id = o.user_account_id
LEFT JOIN %s p ON p.account_id = o.user_account_id
LEFT JOIN %s oe on c.order_id=oe.order_id`,
			overdueCase.TableName(), repayPlan.TableName(),
			order.TableName(), accountBase.TableName(), accountProfile.TableName(), ordersExt.TableName())

	sql := fmt.Sprintf(`%s %s %s`, sqlCount, from, where)
	r := o.Raw(sql)
	err = r.QueryRow(&total)
	if err != nil {
		return
	}

	orderBy := ""
	if v, ok := condStr["field"]; ok {
		if vF, okF := overdueFieldMap[v.(string)]; okF {
			orderBy = "ORDER BY " + vF
		} else {
			orderBy = "ORDER BY c.id"
		}
	} else {
		orderBy = "ORDER BY c.id"
	}

	if v, ok := condStr["sort"]; ok {
		orderBy = fmt.Sprintf("%s %s", orderBy, v.(string))
	} else {
		orderBy = fmt.Sprintf("%s %s", orderBy, "DESC")
	}

	limit := fmt.Sprintf(`LIMIT %d, %d`, offset, pagesize)

	sql = fmt.Sprintf(`%s %s %s %s %s`, sqlSelect, from, where, orderBy, limit)
	r = o.Raw(sql)
	_, err = r.QueryRows(&list)

	if len(list) > 0 {
		for k, v := range list {

			logs.Info("entrust_pname:", v.EntrustPname)

			oneTicket, _ := models.GetTicketByItemAndRelatedID(types.MustGetTicketItemIDByCaseName(v.CaseLevel), v.Id)
			if oneTicket.Id > 0 {
				list[k].IsTicket = oneTicket.Id
			}
			logs.Info("istiecket:", list[k].IsTicket)

		}
	}

	return
}

func EntrustApprovalListBackend(adminUID int64, condStr map[string]interface{}, page, pagesize int) (list []OctwoCaseListItem, total int64, err error) {
	// caseLevel := "M1-2"
	o := orm.NewOrm()
	ticketModel := models.Ticket{}
	entrustApprovalRecord := models.EntrustApprovalRecord{}
	o.Using(ticketModel.UsingSlave())

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize
	orders := make([]int64, 0)
	needStatus := true

	ffrom := fmt.Sprintf(`FROM %s t LEFT JOIN %s ear ON t.order_id = ear.order_id`,
		ticketModel.TableName(), entrustApprovalRecord.TableName())

	fwhere := fmt.Sprintf("WHERE 1 = 1")

	if f, ok := condStr["isAgree"]; ok {
		needStatus = false
		fwhere = fmt.Sprintf("%s AND ear.is_agree = %d", fwhere, f.(int))
	}

	if f, ok := condStr["pname"]; ok {
		needStatus = false
		logs.Notice("pname:", condStr["pname"])
		fwhere = fmt.Sprintf("%s AND ear.pname = '%s'", fwhere, f.(string))
	}

	if f, ok := condStr["entrust_apply_start"]; ok {
		fwhere = fmt.Sprintf("%s%s%s", fwhere, " AND t.apply_entrust_time >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["entrust_apply_end"]; ok {
		fwhere = fmt.Sprintf("%s%s%s", fwhere, " AND t.apply_entrust_time <= ", strconv.FormatInt(f.(int64), 10))
	}
	if needStatus {
		fwhere = fmt.Sprintf("%s AND t.`status` = '%d'", fwhere, types.TicketStatusWaitingEntrust)
	}

	countsql := fmt.Sprintf("select count(t.order_id) %s %s ", ffrom, fwhere)
	sql := fmt.Sprintf("select t.order_id %s %s order by t.id desc", ffrom, fwhere)
	r := o.Raw(countsql)
	// err = r.QueryRow(&total)
	// if err != nil {
	// 	return
	// }
	r = o.Raw(sql)
	_, err = r.QueryRows(&orders)

	ordersStr := tools.ArrayToString(orders, ",")
	logs.Debug("[EntrustApprovalListBackend] orders:", ordersStr)

	oo := orm.NewOrm()
	order := models.Order{}
	oo.Using(order.UsingSlave())
	//where := fmt.Sprintf("WHERE c.is_out = %d", types.IsUrgeOutNo)
	where := fmt.Sprintf("WHERE 1 = 1 AND a.is_deleted = 0 AND o.is_deleted = 0 AND c.is_out=0")

	if f, ok := condStr["order_id"]; ok {
		where = fmt.Sprintf("%s AND c.order_id = %d", where, f.(int64))
	}
	if f, ok := condStr["account_id"]; ok {
		where = fmt.Sprintf("%s AND a.id = %d", where, f.(int64))
	}
	if f, ok := condStr["filter"]; ok {
		where = fmt.Sprintf("%s AND c.is_out = %d", where, f.(int))
	}

	if f, ok := condStr["isEntrust"]; ok {
		where = fmt.Sprintf("%s AND oe.is_entrust = %d", where, f.(int))
	}

	where = fmt.Sprintf("%s AND c.order_id in(%s) ", where, ordersStr)
	// where = fmt.Sprintf("%s AND c.case_level = '%s'", where, caseLevel)

	if f, ok := condStr["order_type"]; ok {
		var isReloanBox []string
		val := f.(string)
		switch val {
		case types.GetUrgeOrderTypeVal(int(types.UrgeOrderTypeFirst)):
			// 首贷
			isReloanBox = append(isReloanBox, fmt.Sprintf("%d", 0))
		case types.GetUrgeOrderTypeVal(int(types.UrgeOrderTypeRepeat)):
			// 复贷
			isReloanBox = append(isReloanBox, fmt.Sprintf("%d", 1))
		case types.GetUrgeOrderTypeVal(int(types.UrgeOrderTypeRoll)):
			// 展单
			where = fmt.Sprintf("%s AND o.pre_order > 0", where)

		}
		if len(isReloanBox) > 0 {
			where = fmt.Sprintf("%s AND o.is_reloan IN (%s)", where, strings.Join(isReloanBox, ", "))
		}
	}

	if f, ok := condStr["overdue_days_start"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.overdue_days >= ", strconv.FormatInt(f.(int64), 10))
	}
	if f, ok := condStr["overdue_days_end"]; ok {
		where = fmt.Sprintf("%s%s%s", where, " AND c.overdue_days <= ", strconv.FormatInt(f.(int64), 10))
	}

	if f, ok := condStr["left_amount"]; ok {
		where = fmt.Sprintf("%s AND ((r.amount + r.grace_period_interest + r.penalty - r.amount_payed - r.amount_reduced - r.grace_period_interest_payed - r.grace_period_interest_reduced - r.penalty_payed - r.penalty_reduced) between 1 and %d)", where, f.(int64))
	}

	accountBase := models.AccountBase{}
	overdueCase := models.OverdueCase{}
	repayPlan := models.RepayPlan{}
	accountProfile := models.AccountProfile{}
	ordersExt := models.OrderExt{}

	sqlCount := "SELECT COUNT(c.id) AS total"
	sqlSelect := "SELECT c.*, a.id AS account_id, a.realname, a.mobile, r.amount, r.amount_payed,r.amount_reduced, r.repay_date, r.grace_period_interest, r.grace_period_interest_payed, r.grace_period_interest_reduced,r.penalty, r.penalty_payed,r.penalty_reduced, p.company_telephone, p.salary_day,oe.entrust_pname,oe.is_entrust"

	// 获取逾期案例详单中,每个order_id相关的最新项
	from :=
		fmt.Sprintf(`FROM %s c
	LEFT JOIN %s r ON r.order_id = c.order_id
	LEFT JOIN %s o ON o.id = c.order_id
	LEFT JOIN %s a ON a.id = o.user_account_id
	LEFT JOIN %s p ON p.account_id = o.user_account_id
	LEFT JOIN %s oe on c.order_id=oe.order_id`,

			overdueCase.TableName(), repayPlan.TableName(),
			order.TableName(), accountBase.TableName(), accountProfile.TableName(), ordersExt.TableName())

	csql := fmt.Sprintf(`%s %s %s`, sqlCount, from, where)
	r = oo.Raw(csql)
	err = r.QueryRow(&total)
	if err != nil {
		return
	}

	orderBy := ""
	if v, ok := condStr["field"]; ok {
		if vF, okF := overdueFieldMap[v.(string)]; okF {
			orderBy = "ORDER BY " + vF
		} else {
			orderBy = "ORDER BY c.id"
		}
	} else {
		orderBy = "ORDER BY c.id"
	}

	if v, ok := condStr["sort"]; ok {
		orderBy = fmt.Sprintf("%s %s", orderBy, v.(string))
	} else {
		orderBy = fmt.Sprintf("%s %s", orderBy, "DESC")
	}

	limit := fmt.Sprintf(`LIMIT %d, %d`, offset, pagesize)

	sql = fmt.Sprintf(`%s %s %s %s %s`, sqlSelect, from, where, orderBy, limit)
	r = oo.Raw(sql)
	_, err = r.QueryRows(&list)

	// if len(list) > 0 {
	// 	for k, v := range list {

	// 		logs.Info("entrust_pname:", v.EntrustPname)

	// 		oneTicket, _ := models.GetTicketByItemAndRelatedID(types.MustGetTicketItemIDByCaseName(v.CaseLevel), v.Id)
	// 		if oneTicket.Id > 0 {
	// 			list[k].IsTicket = oneTicket.Id
	// 		}
	// 		logs.Info("istiecket:", list[k].IsTicket)

	// 	}
	// }

	return
}

/**
再次放款展示信息
*/
func DisbureseAgainDetailBackend(orderId int64) GiveoutCreditBackendData {
	sql := "SELECT account_base.id,account_base.realname,account_profile.bank_name,account_profile.bank_no,orders.loan,orders.check_status FROM orders LEFT JOIN account_base ON orders.user_account_id = account_base.id LEFT JOIN account_profile ON account_profile.account_id = account_base.id WHERE orders.id = '"
	sql = fmt.Sprintf("%s%d%s", sql, orderId, "'")
	o := orm.NewOrm()
	order := models.Order{}
	o.Using(order.Using())
	r := o.Raw(sql)
	var maps GiveoutCreditBackendData
	r.QueryRow(&maps)
	return maps
}

func CanDisbureAgain(opUid int64, orderId int64, newStautus int) bool {

	if newStautus != int(types.LoanStatusWait4Loan) {
		return true
	}

	invork, _ := models.GetLastestDisburseInvorkLogByPkOrderId(orderId)
	if invork.FailureCode == "INVALID_DESTINATION" ||
		invork.FailureCode == "Transfer Inquiry Decline" ||
		invork.FailureCode == "RECIPIENT_ACCOUNT_NUMBER_ERROR" ||
		invork.FailureCode == "SWITCHING_NETWORK_ERROR" ||
		invork.FailureCode == "Name_Contain_Number" {
		return true
	}

	if opUid == types.SuperAdminUID {
		return true
	}

	return false
}

func DoDisbureseAgainBackend(adminId int64, orderId int64, checkStatus types.LoanStatus) error {
	order := models.Order{}
	order, err := models.GetOrder(orderId)
	if err != nil {
		err = fmt.Errorf("DoDisbureseAgainBackend 查不出订单，orderid is: %d", orderId)
		logs.Error(err)
		return err
	}

	return DoDisbureseAgainBackendV2(adminId, order, checkStatus)
}

func DoDisbureseAgainBackendV2(adminId int64, order models.Order, checkStatus types.LoanStatus) error {
	origin := order
	if order.CheckStatus == types.LoanStatusLoanFail {
		/*
			if checkStatus == types.LoanStatusWait4Loan {
				inquriyResp, err, httpCode := xendit.DisburseInquiry(order.UserAccountId, order.Id)
				//对于失败的订单再次放款前，我们要向xendit反查一遍，确认此订单确实没有放款过，然后再对此单放款，避免重复放款
				if httpCode == 0 {
					//超时处理
					//如果请求超时，直接返回，让操作员继续再后台尝试放款
					return err
				} else {
					if httpCode == 200 {
						desc := inquriyResp.DisbursementDescription
						if desc == tools.Int642Str(order.Id) && inquriyResp.Status == "COMPLETED" {
							//证明此订单已经放出过一笔了
							order.CheckStatus = types.LoanStatusIsDoing
							models.UpdateOrder(&order)
							inquiryB, _ := json.Marshal(inquriyResp)
							err = xendit.SimulateDisburse(string(inquiryB))
							if err != nil {
								err = fmt.Errorf("[xendit.SimulateDisburse] failed.")
								return err
							}
						} else {
							//查询不到放款记录，可以重新放款
							updateReDisburseOrder(&order, checkStatus)
						}
					} else if httpCode == 404 {
						//{"error_code":"DIRECT_DISBURSEMENT_NOT_FOUND_ERROR","message":"Direct disbursement not fond"}
						updateReDisburseOrder(&order, checkStatus)
					}
					//其他httpCode情况不处理，表示本次调用失败，让订单维持放款失败状态
				}
			} else {
				//将订单变成"失效"状态
				updateReDisburseOrder(&order, checkStatus)
			}
		*/
		updateReDisburseOrder(&order, checkStatus)
	} else {
		err := fmt.Errorf("DoDisbureseAgainBackendV2 订单状态不是放款失败，无法更改状态。orderid is: %d status:%d", order.Id, order.CheckStatus)
		logs.Error(err)
		return err
	}

	models.OpLogWrite(adminId, order.Id, models.OpCodeOrderUpdate, order.TableName(), origin, order)
	return nil
}

func updateReDisburseOrder(order *models.Order, checkStatus types.LoanStatus) {
	order.CheckStatus = checkStatus
	order.CheckTime = tools.GetUnixMillis()
	order.Utime = tools.GetUnixMillis()
	models.UpdateOrder(order)
	//删除锁后，脚本才能正常执行
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()
	keyPrefix := beego.AppConfig.String("disburse_order_lock")
	key := fmt.Sprintf("%s%d", keyPrefix, order.Id)
	cacheClient.Do("DEL", key)
	//修改成失效的时候出发失效事件
	if checkStatus == types.LoanStatusInvalid {
		// 订单失效事件触发
		accountCoupon, err := dao.GetAccountFrozenCouponByOrder(order.UserAccountId, order.Id)
		if err == nil {
			MakeAccountCouponAvailable(&accountCoupon)
		}

		event.Trigger(&evtypes.OrderInvalidEv{
			OrderID:   order.Id,
			AccountID: order.UserAccountId,
			Time:      tools.GetUnixMillis(),
		})
	}
}

func ConfirmOrder(accountId int64, loan int64, period int) (orderId int64, err error) {
	// 1. 查看资料完成阶段
	// clientInfo, err := models.OneLastClientInfoByRelatedID(accountId)
	// phase := ProfileCompletePhase(accountId, clientInfo.UiVersion, clientInfo.AppVersionCode)
	ok, phase := IsUserPerfectInformation(accountId, loan, period)
	if !ok {
		// 如果用户资料没有提交完成,则不能完成订单转换
		err = fmt.Errorf("user information is incomplete. accountId: %d, phase: %d", accountId, phase)
		logs.Warning("[ConfirmOrder] has wrong. err:", err)

		return
	}

	// 2. 查看借款生命周期
	loanLifetime := GetLoanLifetime(accountId)
	if loanLifetime != types.LoanLifetimeNormal {
		err = fmt.Errorf("loan lifetime has wrong, please checkout it. accountId: %d, loanLifetime: %d", accountId, loanLifetime)
		logs.Warning("[ConfirmOrder] err:", err)

		return
	}

	// 3. 转为正常订单
	tmpOrder, err := dao.AccountLastTmpLoanOrderByCond(accountId, loan, period)
	if err != nil {
		logs.Warning("[ConfirmOrder] customer has no temporary order. accountId:", accountId, ", err:", err)
		return
	}

	orderId = tmpOrder.Id

	tmpOrder.CheckStatus = types.LoanStatus4Review
	tmpOrder.IsTemporary = types.IsTemporaryNO
	tmpOrder.ApplyTime = tools.GetUnixMillis()
	tmpOrder.Utime = tools.GetUnixMillis()
	tmpOrder.RandomValue = tools.GenerateRandom(1, 101) // 用户确认订单时生成随机数
	tryAddOrderPhyInvalidTag(accountId, tmpOrder.Id)

	o := orm.NewOrm()
	o.Using(tmpOrder.Using())
	num, err := o.Update(&tmpOrder, "check_status", "is_temporary", "apply_time", "random_value", "utime")
	if err != nil || num != 1 {
		orderJSON, _ := tools.JsonEncode(tmpOrder)
		logs.Error("[ConfirmOrder] update order has wrong. order:", orderJSON, ", num:", num, ", err:", err)
	}

	// 用户借款订单审核事件触发
	event.Trigger(&evtypes.OrderAuditEv{
		OrderID:   orderId,
		AccountID: accountId,
		Time:      tools.GetUnixMillis(),
	})

	monitor.IncrOrderCount(tmpOrder.CheckStatus)

	// 事件触发
	event.Trigger(&evtypes.LoanSubmitEv{OrderID: tmpOrder.Id, Time: tools.GetUnixMillis()})

	// 4. 将此用户的其他临时订单全部置为无效订单
	ConvertTemporaryOrder2Invalid(accountId)

	return
}

type LoanOrderCond struct {
	Loan           int64
	LoanNew        int64
	ContractAmount int64
	Period         int
	PeriodNew      int
}

func ConfirmOrderV2(accountId, productId int64, loanOrderCond LoanOrderCond) (orderId int64, err error) {
	// 1. 查看资料完成阶段
	// clientInfo, err := models.OneLastClientInfoByRelatedID(accountId)
	// phase := ProfileCompletePhase(accountId, clientInfo.UiVersion, clientInfo.AppVersionCode)
	ok, phase := IsUserPerfectInformation(accountId, loanOrderCond.Loan, loanOrderCond.Period)
	if !ok {
		// 如果用户资料没有提交完成,则不能完成订单转换
		err = fmt.Errorf("user information is incomplete. accountId: %d, phase: %d", accountId, phase)
		logs.Warning("[ConfirmOrderV2] has wrong. err:", err)

		return
	}

	// 2. 查看借款生命周期
	loanLifetime := GetLoanLifetime(accountId)
	if loanLifetime != types.LoanLifetimeNormal {
		err = fmt.Errorf("loan lifetime has wrong, please checkout it. accountId: %d, loanLifetime: %d", accountId, loanLifetime)
		logs.Warning("[ConfirmOrderV2] err:", err)

		return
	}

	// 3. 转为正常订单
	tmpOrder, err := dao.AccountLastTmpLoanOrderByCond(accountId, loanOrderCond.Loan, loanOrderCond.Period)
	if err != nil {
		logs.Warning("[ConfirmOrderV2] customer has no temporary order. accountId:", accountId, ", err:", err)
		return
	}

	orderId = tmpOrder.Id

	tmpOrder.CheckStatus = types.LoanStatus4Review
	tmpOrder.IsTemporary = types.IsTemporaryNO
	tmpOrder.ApplyTime = tools.GetUnixMillis()
	tmpOrder.Utime = tools.GetUnixMillis()
	tmpOrder.RandomValue = tools.GenerateRandom(1, 101) // 用户确认订单时生成随机数
	tmpOrder.Loan = loanOrderCond.LoanNew
	tmpOrder.Amount = loanOrderCond.ContractAmount
	tmpOrder.Period = loanOrderCond.PeriodNew
	tmpOrder.ProductId = productId
	tryAddOrderPhyInvalidTag(accountId, tmpOrder.Id)

	o := orm.NewOrm()
	o.Using(tmpOrder.Using())
	num, err := o.Update(&tmpOrder, "check_status", "is_temporary", "apply_time", "random_value", "utime", "loan", "amount", "period", "product_id")
	if err != nil || num != 1 {
		orderJSON, _ := tools.JsonEncode(tmpOrder)
		logs.Error("[ConfirmOrderV2] update order has wrong. order:", orderJSON, ", num:", num, ", err:", err)
	}

	// 用户借款订单审核事件触发
	event.Trigger(&evtypes.OrderAuditEv{
		OrderID:   orderId,
		AccountID: accountId,
		Time:      tools.GetUnixMillis(),
	})

	monitor.IncrOrderCount(tmpOrder.CheckStatus)

	// 事件触发
	event.Trigger(&evtypes.LoanSubmitEv{OrderID: tmpOrder.Id, Time: tools.GetUnixMillis()})

	// 4. 将此用户的其他临时订单全部置为无效订单
	ConvertTemporaryOrder2Invalid(accountId)

	return
}

func ConfirmOrderV3(accountId, productId int64, loanOrderCond LoanOrderCond, couponId int64) (phase int, orderId int64, couErr error, err error) {
	// 1. 查看资料完成阶段
	// clientInfo, err := models.OneLastClientInfoByRelatedID(accountId)
	// phase := ProfileCompletePhase(accountId, clientInfo.UiVersion, clientInfo.AppVersionCode)
	ok, phase := IsUserPerfectInformation(accountId, loanOrderCond.Loan, loanOrderCond.Period)
	if !ok {
		// 如果用户资料没有提交完成,则不能完成订单转换
		err = fmt.Errorf("user information is incomplete. accountId: %d, phase: %d", accountId, phase)
		logs.Warning("[ConfirmOrderV3] has wrong. err:", err)

		return
	}

	// 2. 查看借款生命周期
	loanLifetime := GetLoanLifetime(accountId)
	if loanLifetime != types.LoanLifetimeNormal {
		err = fmt.Errorf("loan lifetime has wrong, please checkout it. accountId: %d, loanLifetime: %d", accountId, loanLifetime)
		logs.Warning("[ConfirmOrderV3] err:", err)

		return
	}

	// 3. 转为正常订单
	tmpOrder, err := dao.AccountLastTmpLoanOrderByCond(accountId, loanOrderCond.Loan, loanOrderCond.Period)
	if err != nil {
		logs.Warning("[ConfirmOrderV3] customer has no temporary order. accountId:", accountId, ", err:", err)
		return
	}

	orderId = tmpOrder.Id

	tmpOrder.CheckStatus = types.LoanStatus4Review
	tmpOrder.IsTemporary = types.IsTemporaryNO
	tmpOrder.ApplyTime = tools.GetUnixMillis()
	tmpOrder.Utime = tools.GetUnixMillis()
	tmpOrder.RandomValue = tools.GenerateRandom(1, 101) // 用户确认订单时生成随机数
	tmpOrder.Loan = loanOrderCond.LoanNew
	tmpOrder.Amount = loanOrderCond.ContractAmount
	tmpOrder.Period = loanOrderCond.PeriodNew
	tmpOrder.ProductId = productId
	tryAddOrderPhyInvalidTag(accountId, tmpOrder.Id)

	o := orm.NewOrm()
	o.Using(tmpOrder.Using())
	num, err := o.Update(&tmpOrder, "check_status", "is_temporary", "apply_time", "random_value", "utime", "loan", "amount", "period", "product_id")
	if err != nil || num != 1 {
		orderJSON, _ := tools.JsonEncode(tmpOrder)
		logs.Error("[ConfirmOrderV3] update order has wrong. order:", orderJSON, ", num:", num, ", err:", err)
	}

	if couponId != 0 {
		_, _, _, couErr = OrderUseCoupon(loanOrderCond.LoanNew, loanOrderCond.ContractAmount, loanOrderCond.PeriodNew, &tmpOrder, couponId)
	}

	// 用户借款订单审核事件触发
	event.Trigger(&evtypes.OrderAuditEv{
		OrderID:   orderId,
		AccountID: accountId,
		Time:      tools.GetUnixMillis(),
	})
	// 事件触发
	event.Trigger(&evtypes.LoanSubmitEv{OrderID: tmpOrder.Id, Time: tools.GetUnixMillis()})

	// 4. 将此用户的其他临时订单全部置为无效订单
	ConvertTemporaryOrder2Invalid(accountId)

	return
}

func ConfirmOrderTwo(accountId, productId int64, loanOrderCond LoanOrderCond, couponId int64) (phase int, orderId int64, couErr error, err error) {
	// 1. 查看资料完成阶段
	// clientInfo, err := models.OneLastClientInfoByRelatedID(accountId)
	// phase := ProfileCompletePhase(accountId, clientInfo.UiVersion, clientInfo.AppVersionCode)
	ok, phase := IsUserPerfectInformationTwo(accountId, loanOrderCond.Loan, loanOrderCond.Period)
	if !ok {
		// 如果用户资料没有提交完成,则不能完成订单转换
		err = fmt.Errorf("user information is incomplete. accountId: %d, phase: %d", accountId, phase)
		logs.Warning("[ConfirmOrderV3] has wrong. err:", err)

		return
	}

	// 2. 查看借款生命周期
	loanLifetime := GetLoanLifetime(accountId)
	if loanLifetime != types.LoanLifetimeNormal {
		err = fmt.Errorf("loan lifetime has wrong, please checkout it. accountId: %d, loanLifetime: %d", accountId, loanLifetime)
		logs.Warning("[ConfirmOrderV3] err:", err)

		return
	}

	// 3. 转为正常订单
	tmpOrder, err := dao.AccountLastTmpLoanOrderByCond(accountId, loanOrderCond.Loan, loanOrderCond.Period)
	if err != nil {
		logs.Warning("[ConfirmOrderTwo] customer has no temporary order. accountId:", accountId, ", err:", err)
		return
	}

	orderId = tmpOrder.Id

	var newAmount, newLoan int64
	var coutonType types.CouponType
	if couponId != 0 {
		newAmount, newLoan, coutonType, couErr = OrderUseCoupon(loanOrderCond.LoanNew, loanOrderCond.ContractAmount, loanOrderCond.PeriodNew, &tmpOrder, couponId)
	}

	tmpOrder.CheckStatus = types.LoanStatus4Review
	tmpOrder.IsTemporary = types.IsTemporaryNO
	tmpOrder.ApplyTime = tools.GetUnixMillis()
	tmpOrder.Utime = tools.GetUnixMillis()
	tmpOrder.RandomValue = tools.GenerateRandom(1, 101) // 用户确认订单时生成随机数
	tmpOrder.Loan = loanOrderCond.LoanNew
	tmpOrder.Amount = loanOrderCond.ContractAmount
	tmpOrder.Period = loanOrderCond.PeriodNew
	tmpOrder.ProductId = productId
	if couErr == nil && coutonType == types.CouponTypeLimit && newLoan > tmpOrder.Loan && newAmount > tmpOrder.Amount {
		tmpOrder.Loan = newLoan
		tmpOrder.Amount = newAmount
	}

	tryAddOrderPhyInvalidTag(accountId, tmpOrder.Id)

	o := orm.NewOrm()
	o.Using(tmpOrder.Using())
	num, err := o.Update(&tmpOrder, "check_status", "is_temporary", "apply_time", "random_value", "utime", "loan", "amount", "period", "product_id")
	if err != nil || num != 1 {
		orderJSON, _ := tools.JsonEncode(tmpOrder)
		logs.Error("[ConfirmOrderTwo] update order has wrong. order:", orderJSON, ", num:", num, ", err:", err)
	}

	// 用户借款订单审核事件触发
	event.Trigger(&evtypes.OrderAuditEv{
		OrderID:   orderId,
		AccountID: accountId,
		Time:      tools.GetUnixMillis(),
	})
	// 事件触发
	event.Trigger(&evtypes.LoanSubmitEv{OrderID: tmpOrder.Id, Time: tools.GetUnixMillis()})

	// 4. 将此用户的其他临时订单全部置为无效订单
	ConvertTemporaryOrder2Invalid(accountId)

	if tmpOrder.IsReloan == int(types.IsReloanNo) {
		param := coupon_event.InviteV3Param{}
		param.AccountId = tmpOrder.UserAccountId
		param.TaskType = types.AccountTaskApply
		HandleCouponEvent(coupon_event.TriggerInviteV3, param)
	}

	return
}

func WriteOrdersQuota(orderId int64, quota int64) {
	timetag := tools.GetUnixMillis()
	orderExt, err := models.GetOrderExt(orderId)
	if err != nil {
		orderExt = models.OrderExt{}
		orderExt.OrderId = orderId
		orderExt.QuotaIncreased = quota
		orderExt.Ctime = timetag
		orderExt.Utime = timetag
		orderExt.Add()
	} else {
		orderExt.QuotaIncreased = quota
		orderExt.Utime = timetag
		orderExt.Update()
	}
}

func ConvertTemporaryOrder2Invalid(accountId int64) (num int64, err error) {
	order := models.Order{}
	o := orm.NewOrm()
	o.Using(order.Using())

	sql := fmt.Sprintf(`UPDATE %s
SET check_status = %d, utime = %d
WHERE user_account_id = %d AND check_status = %d AND is_temporary = %d`, order.TableName(),
		types.LoanStatusInvalid, tools.GetUnixMillis(),
		accountId, types.LoanStatusSubmit, types.IsTemporaryYes)
	res, err := o.Raw(sql).Exec()
	if err == nil {
		num, _ = res.RowsAffected()
	}

	return
}

// 获取“借款管理/业务流水”所需数据
func GetLoanOrderBusiness(orderId int64) (orderLoanBusiness models.OrderLoanBusiness) {
	order, _ := models.GetOrder(orderId)
	accountBase, _ := models.OneAccountBaseByPkId(order.UserAccountId)

	orderLoanBusiness.OpUid = order.OpUid
	orderLoanBusiness.ApplyTime = order.ApplyTime
	if orderLoanBusiness.ApplyTime > 0 {
		orderLoanBusiness.ApplyOperator = accountBase.Realname
	} else {
		orderLoanBusiness.ApplyOperator = "-"
	}

	orderLoanBusiness.CheckTime = order.CheckTime
	if orderLoanBusiness.CheckTime > 0 {
		orderLoanBusiness.CheckOperator = accountBase.Realname
	} else {
		orderLoanBusiness.CheckOperator = "-"
	}

	orderLoanBusiness.RiskCtlFinishTime = order.RiskCtlFinishTime

	orderLoanBusiness.PhoneVerifyTime = order.PhoneVerifyTime
	if orderLoanBusiness.PhoneVerifyTime > 0 {
		orderLoanBusiness.PhoneVerifyOperator = accountBase.Realname
	} else {
		orderLoanBusiness.PhoneVerifyOperator = "-"
	}

	switch order.CheckStatus {
	case types.LoanStatusLoanFail:
		orderLoanBusiness.LoanStatus = "放款失败"
		orderLoanBusiness.PayOperator = "-"
	case types.LoanStatusWaitRepayment:
		orderLoanBusiness.LoanStatus = "放款成功"
		orderLoanBusiness.LoanTime = order.LoanTime
	case types.LoanStatusAlreadyCleared:
		orderLoanBusiness.LoanStatus = "放款成功"
		orderLoanBusiness.LoanTime = order.LoanTime
		orderLoanBusiness.PayStatus = "已结清"
		orderLoanBusiness.PayTime = order.FinishTime
		orderLoanBusiness.PayOperator = accountBase.Realname
	case types.LoanStatusOverdue:
		orderLoanBusiness.LoanStatus = "放款成功"
		orderLoanBusiness.LoanTime = order.LoanTime
		orderLoanBusiness.PayStatus = "逾期"
		orderLoanBusiness.PayTime = order.PenaltyUtime
		orderLoanBusiness.PayOperator = accountBase.Realname
	case types.LoanStatusPartialRepayment:
		orderLoanBusiness.LoanStatus = "放款成功"
		orderLoanBusiness.LoanTime = order.LoanTime
		orderLoanBusiness.PayStatus = "部分还款"
		orderLoanBusiness.PayTime = order.RepayTime
		orderLoanBusiness.PayOperator = accountBase.Realname
	default:
		orderLoanBusiness.LoanStatus = "-"
		orderLoanBusiness.PayStatus = "-"
		orderLoanBusiness.PayOperator = "-"
	}

	return
}

func GetBackendRepayPlan(orderId int64) models.RepayPlan {
	repayPlan, _ := models.GetLastRepayPlanByOrderid(orderId)
	return repayPlan
}

func GetAccountBankInfo(accountId int64) GiveoutCreditBackendData {

	o := orm.NewOrm()
	accountProfile := models.AccountProfile{}
	o.Using(accountProfile.Using())

	sql := fmt.Sprintf("%s%d", "SELECT account_profile.account_id,account_base.realname, account_profile.bank_name, account_profile.bank_no FROM account_base LEFT JOIN account_profile ON account_base.id = account_profile.account_id WHERE account_base.id = ", accountId)
	r := o.Raw(sql)
	var maps GiveoutCreditBackendData
	r.QueryRow(&maps)

	return maps
}

func OrderClientInfo(orderId int64) (clientInfo models.ClientInfo, err error) {
	if orderId <= 0 {
		logs.Warning("[OrderClientInfo] invalid orderId:", orderId)
		err = fmt.Errorf("invalid orderId: %d", orderId)
		return
	}

	o := orm.NewOrm()
	o.Using(clientInfo.UsingSlave())

	err = o.QueryTable(clientInfo.TableName()).Filter("related_id", orderId).Filter("service_type", types.ServiceCreateOrder).OrderBy("-id").Limit(1).One(&clientInfo)

	return
}

func LastClientInfo(relatedID int64) (clientInfo models.ClientInfo, err error) {
	if relatedID <= 0 {
		logs.Warning("[OrderClientInfo] invalid related_id:", relatedID)
		err = fmt.Errorf("invalid related_id: %d", relatedID)
		return
	}

	o := orm.NewOrm()
	o.Using(clientInfo.UsingSlave())

	err = o.QueryTable(clientInfo.TableName()).
		Filter("related_id", relatedID).
		OrderBy("-id").Limit(1).One(&clientInfo)

	return
}

func GetBackendUserETrans(orderId int64) []models.User_E_Trans {
	list := models.GetETransByOrderId(orderId)
	return list
}

// 后台逾期管理
func ReductionListBackend(condStr map[string]interface{}, page, pagesize int) (list []models.ReduceRecordListItem, total int64, err error) {
	o := orm.NewOrm()
	reduce := models.ReduceRecordNew{}
	o.Using(reduce.UsingSlave())

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	where := fmt.Sprintf("WHERE 1 = 1")
	if f, ok := condStr["case_id"]; ok {
		where = fmt.Sprintf("%s AND case_id = %d", where, f.(int64))
	}
	if f, ok := condStr["order_id"]; ok {
		where = fmt.Sprintf("%s AND order_id = %d", where, f.(int64))
	}
	if f, ok := condStr["account_id"]; ok {
		where = fmt.Sprintf("%s AND user_account_id = %d", where, f.(int64))
	}
	if f, ok := condStr["reduce_type"]; ok {
		where = fmt.Sprintf("%s AND reduce_type = %d", where, f.(int))
	}
	if f, ok := condStr["reduce_status"]; ok {
		where = fmt.Sprintf("%s AND reduce_status = %d", where, f.(int))
	}

	sqlCount := "SELECT COUNT(id) AS total"
	sqlSelect := "SELECT * "

	// 获取逾期案例详单中,每个order_id相关的最新项
	from := fmt.Sprintf("FROM %s", reduce.TableName())

	sql := fmt.Sprintf(`%s %s %s`, sqlCount, from, where)
	r := o.Raw(sql)
	err = r.QueryRow(&total)
	if err != nil {
		return
	}
	orderBy := "ORDER BY id DESC"

	limit := fmt.Sprintf(`LIMIT %d, %d`, offset, pagesize)

	sql = fmt.Sprintf(`%s %s %s %s %s`, sqlSelect, from, where, orderBy, limit)
	r = o.Raw(sql)
	_, err = r.QueryRows(&list)

	return
}

func ReducePenaltyApply(orderId int64, opUid int64, reductionAmount int64, reduction_penalty int64, reduction_grace_period_interest int64, opReason string) error {

	obj, err := models.GetLastRepayPlanByOrderid(orderId)
	if err != nil {
		logs.Error("[ReducePenalty] GetLastRepayPlanByOrderid err:%v orderId:%d", err, orderId)
		return err
	}

	order, err := models.GetOrder(orderId)
	if err != nil {
		logs.Error("[ReducePenalty] GetOrder err:%v orderId：%d", err, orderId)
		return err
	}

	diffAmount := obj.Amount - obj.AmountPayed - obj.AmountReduced
	if diffAmount < reductionAmount {
		err = fmt.Errorf("减免罚息金额大于剩余本金, 剩余应还本金%d , 欲减本金%d", diffAmount, reductionAmount)
		return err
	}

	diffPenalty := obj.Penalty - obj.PenaltyPayed - obj.PenaltyReduced
	if diffPenalty < reduction_penalty {
		err = fmt.Errorf("减免罚息金额大于剩余罚息, 剩余罚息%d , 欲减罚息%d", diffPenalty, reduction_penalty)
		return err
	}

	diffGracePeriodInterest := obj.GracePeriodInterest - obj.GracePeriodInterestPayed - obj.GracePeriodInterestReduced
	if diffGracePeriodInterest < reduction_grace_period_interest {
		err = fmt.Errorf("减免宽限期利息金额大于剩余宽限期利息, 剩余宽限期利息%d , 欲减宽限期利息%d", diffGracePeriodInterest, reduction_grace_period_interest)
		return err
	}

	// save record to reduce_record
	caseOver, _ := models.OneOverdueCaseByOrderID(order.Id)
	id, _ := device.GenerateBizId(types.ReduceRecordBiz)
	tag := tools.GetUnixMillis()
	record := models.ReduceRecordNew{
		Id:                   id,
		ApplyUid:             opUid,
		ApplyTime:            tag,
		UserAccountId:        order.UserAccountId,
		CaseID:               caseOver.Id,
		OrderId:              orderId,
		OpReason:             opReason,
		ReduceType:           types.ReduceTypeManual,
		ReduceStatus:         types.ReduceStatusApplyed,
		AmountReduced:        reductionAmount,
		PenaltyReduced:       reduction_penalty,
		GraceInterestReduced: reduction_grace_period_interest,
		Ctime:                tag,
		Utime:                tag,
	}
	_, err = models.OrmInsert(&record)
	if err != nil {
		logs.Error("[ReducePenalty] OrmInsert err:%v reduce:%#v", err, record)
		return err
	}
	return nil
}

func ReductionConfirmSave(reduceId int64, opUid int64, confirmOption int, opRemark string) error {

	one, err := dao.GetReduceById(reduceId)
	if err != nil {
		logs.Error("[ReductionConfirmSave] GetReduceById err:%v reduceId:%d opUid:%d confirmOption:%d", err, reduceId, opUid, confirmOption)
		return err
	}

	// 只有手工减免需要确认
	if one.ReduceType != types.ReduceTypeManual || one.ReduceStatus != types.ReduceStatusApplyed {
		err = fmt.Errorf("[ReductionConfirmSave] reducetype err. reduce:%#v opUid:%d", one, opUid)
		return err
	}

	// 审核拒绝直接更新
	if confirmOption == types.ReduceConfirmOptionReject {
		return updateReduce(one, opUid, opRemark, types.ReduceStatusRejected, "")
	}

	// 再次确认是否可以减免
	obj, err := models.GetLastRepayPlanByOrderid(one.OrderId)
	if err != nil {
		logs.Error("[ReductionConfirmSave] GetLastRepayPlanByOrderid err:%v reduce:%#v", err, one)
		return err
	}
	origin := obj
	order, err := models.GetOrder(one.OrderId)
	if err != nil {
		logs.Error("[ReductionConfirmSave] GetOrder err:%v reduce:%#v", err, one)
		return err
	}
	oldOrder := order
	orderId := one.OrderId

	if !orderStatusAllowReduce(&order) {
		reason := fmt.Sprintf("%s:%d", types.ReduceInvalidReasonOrdersStatus, order.CheckStatus)
		updateReduce(one, opUid, opRemark, types.ReduceStatusInvalid, reason)

		err = fmt.Errorf("[ReductionConfirmSave] %s", reason)
		return err
	}

	diffAmount := obj.Amount - obj.AmountPayed - obj.AmountReduced
	if diffAmount < one.AmountReduced {
		logs.Error("[ReductionConfirmSave] 减免罚息金额大于剩余本金, 剩余应还本金%d , 欲减本金%d", diffAmount, one.AmountReduced)
		err = fmt.Errorf("[ReductionConfirmSave] 减免罚息金额大于剩余本金, 剩余应还本金%d , 欲减本金%d", diffAmount, one.AmountReduced)
		reason := fmt.Sprintf("%s:剩余应还本金%d ", types.ReduceInvalidReasonAmount, diffAmount)
		updateReduce(one, opUid, opRemark, types.ReduceStatusInvalid, reason)
		return err
	}
	diffAmount -= one.AmountReduced

	diffPenalty := obj.Penalty - obj.PenaltyPayed - obj.PenaltyReduced
	if diffPenalty < one.PenaltyReduced {
		logs.Error("[ReductionConfirmSave] 减免罚息金额大于剩余罚息, 剩余罚息%d , 欲减罚息%d", diffPenalty, one.PenaltyReduced)
		reason := fmt.Sprintf("%s:剩余罚息%d", types.ReduceInvalidReasonPenalty, diffPenalty)
		updateReduce(one, opUid, opRemark, types.ReduceStatusInvalid, reason)
		err = fmt.Errorf("[ReductionConfirmSave] 减免罚息金额大于剩余罚息, 剩余罚息%d , 欲减罚息%d", diffPenalty, one.PenaltyReduced)
		return err
	}
	diffPenalty -= one.PenaltyReduced

	diffGracePeriodInterest := obj.GracePeriodInterest - obj.GracePeriodInterestPayed - obj.GracePeriodInterestReduced
	if diffGracePeriodInterest < one.GraceInterestReduced {
		logs.Error("[ReductionConfirmSave] 减免宽限期利息金额大于剩余宽限期利息, 剩余宽限期利息%d , 欲减宽限期利息%d", diffGracePeriodInterest, one.GraceInterestReduced)
		err = fmt.Errorf("[ReductionConfirmSave] 减免宽限期利息金额大于剩余宽限期利息, 剩余宽限期利息%d , 欲减宽限期利息%d", diffGracePeriodInterest, one.GraceInterestReduced)
		reason := fmt.Sprintf("%s:剩余宽限期利息%d", types.ReduceInvalidReasonGrace, diffGracePeriodInterest)
		updateReduce(one, opUid, opRemark, types.ReduceStatusInvalid, reason)
		return err
	}
	diffGracePeriodInterest -= one.GraceInterestReduced

	// 减免可以生效
	err = updateReduce(one, opUid, opRemark, types.ReduceStatusValid, "")
	if err != nil {
		logs.Error("[ReductionConfirmSave] updateReduce err:%d one:%#v", err, one)
		return err
	}

	// 更新还款计划
	obj.AmountReduced += one.AmountReduced
	obj.PenaltyReduced += one.PenaltyReduced
	obj.GracePeriodInterestReduced += one.GraceInterestReduced

	o := orm.NewOrm()
	o.Using(obj.Using())
	_, err = o.Update(&obj)
	if err != nil {
		logs.Error("[ReductionConfirmSave] reduce Update err:%v ", err)
		return err
	}

	models.OpLogWrite(opUid, obj.Id, models.OpReductionInterestUpdate, obj.TableName(), origin, obj)
	models.AddReductionPenalty(orderId, order.UserAccountId, one.AmountReduced, one.PenaltyReduced, one.GraceInterestReduced)

	// 剩余应还的 本金 罚息 宽限息 都为0时更新orders表的订单状态为 已结清
	if 0 == diffAmount && 0 == diffPenalty && 0 == diffGracePeriodInterest {
		UpdateOrderToAlreadyCleared(&order, oldOrder)
		//order.CheckStatus = types.LoanStatusAlreadyCleared
		//order.FinishTime = tools.GetUnixMillis()
		//order.Utime = tools.GetUnixMillis()
		//models.UpdateOrder(&order)
		//
		//monitor.IncrOrderCount(order.CheckStatus)
		//
		//HandleOverdueCase(order.Id)
	}

	// 自动减免逻辑  此时还的钱已存在于 repayPlan内
	if ok, reduce := CanAutoReduce(&order, &obj, 0); ok {
		logs.Info("[ReductionConfirmSave] do auto reduce")
		origin = obj
		err = DoAutoReduce(&order, &obj, reduce)
		if err == nil {
			// order.CheckStatus = types.LoanStatusAlreadyCleared
			// order.FinishTime = tools.GetUnixMillis()
			// order.Utime = tools.GetUnixMillis()
			//
			// models.UpdateOrder(&order)
			// //更新订单状态
			// monitor.IncrOrderCount(order.CheckStatus)
			//
			// HandleOverdueCase(order.Id)
			UpdateOrderToAlreadyCleared(&order, oldOrder)

			diffPenalty = 0
			diffGracePeriodInterest = 0
			models.UpdateRepayPlan(&obj)
			//更新还款计划

			models.OpLogWrite(0, obj.Id, models.OpAutoReduction, obj.TableName(), origin, obj)

		} else {
			logs.Warn("[ReductionConfirmSave] DoAutoReduce err:%v repayPlan:%#v", err, obj)
		}
	}

	//更新结清减免数据
	// TODO 下次更新, 删除此处查case, 此处只为写debug数据, 线上属于空查询, 对应下方 logs.Debug
	// err 也未使用
	oneCase, err := dao.GetInOverdueCaseByOrderID(orderId)
	prereduced, _ := dao.GetLastPrereducedByOrderid(orderId)
	derateRatio := prereduced.DerateRatio
	logs.Debug("[ReductionConfirmSave]结清减免=case:", oneCase.CaseLevel, "取已申请减免的derateRatio:", derateRatio)
	canReductionPenalty := repayplan.CaculateCanReducedAmount(diffPenalty, derateRatio)
	canReductionGracePeriodInterest := repayplan.CaculateCanReducedAmount(diffGracePeriodInterest, derateRatio)
	prereduced.GracePeriodInterestPrededuced = canReductionGracePeriodInterest
	prereduced.PenaltyPrereduced = canReductionPenalty
	prereduced.GraceInterestReduced = canReductionGracePeriodInterest
	prereduced.PenaltyReduced = canReductionPenalty
	prereduced.Utime = tools.GetUnixMillis()
	if diffPenalty == 0 && diffGracePeriodInterest == 0 {
		prereduced.ReduceStatus = types.ReduceStatusInvalid
		prereduced.ConfirmTime = tools.GetUnixMillis()
		prereduced.InvalidReason = types.ClearReducedInvalidReasonAmountInvalid
	}
	cols := []string{"grace_period_interest_prereduced", "penalty_prereduced", "reduce_status", "confirm_time", "grace_interest_reduced", "penalty_reduced", "invalid_reason", "Utime"}
	models.OrmUpdate(&prereduced, cols)

	return nil
}

func orderStatusAllowReduce(order *models.Order) bool {
	switch order.CheckStatus {
	case types.LoanStatusWaitRepayment, types.LoanStatusOverdue, types.LoanStatusPartialRepayment:
		{
			return true
		}
	}
	return false
}

func updateReduce(one models.ReduceRecordNew, opUid int64, opRemark string, status int, invalidReason string) (err error) {
	tag := tools.GetUnixMillis()
	one.ConfirmUid = opUid
	one.ReduceStatus = status
	one.ConfirmTime = tag
	one.ConfirmRemark = opRemark
	one.Utime = tag
	one.InvalidReason = invalidReason
	cols := []string{"confirm_uid", "reduce_status", "confirm_time", "confirm_remark", "invalid_reason", "utime"}
	_, err = models.OrmUpdate(&one, cols)
	if err != nil {
		logs.Error("[updateReduce] OrmUpdate err:%v reduce:%#v", err, one)
		return err
	}
	return nil
}

func HandleDisburse(payType int, dataOrder *models.Order, bankCode string, isRoll bool) error {
	//将打款记录更新为完成
	repayPlan, _ := models.GetLastRepayPlanByOrderid(dataOrder.Id)
	if repayPlan.Id > 0 {
		//防止对方服务器多次回调
		err := fmt.Errorf("[HandleDisburse] This order has been disbursed already! we must stop this disbursement. orderid:%d", dataOrder.Id)
		return err
	}

	oPay := models.Payment{}
	payId, err := oPay.GetDisburseOrder(dataOrder.Id)

	if err == nil {
		errStr := fmt.Sprintf("Disburse payment record exists %d", payId)
		logs.Error(errStr)
		return fmt.Errorf(errStr)
	}

	dataProduct, err := models.GetProduct(dataOrder.ProductId)
	if err != nil {
		errStr := fmt.Sprintf("DisburseFund GetProduct does not have a record: %d", dataOrder.ProductId)
		logs.Error(errStr)
		return fmt.Errorf(errStr)
	}

	var total, interest, serviceFee int64
	if dataOrder.PreOrder > 0 {
		// 展期订单使用不同的试算逻辑
		total, interest, serviceFee = repayplan.CalcRepayInfoV3(dataOrder.Amount, dataProduct, dataOrder.Period)
	} else {
		total, interest, serviceFee = repayplan.CalcRepayInfoV2(dataOrder.Loan, dataProduct, dataOrder.Period)
	}

	logs.Debug("orderid is %d, total is %d, interst is %d, serviceFee is %d", dataOrder.Id, total, interest, serviceFee)
	oPay.Id, _ = device.GenerateBizId(types.PaymentBiz)
	oPay.OrderId = dataOrder.Id
	if !isRoll {
		oPay.Amount = dataOrder.Loan
		oPay.PayType = types.PayTypeMoneyOut
		oPay.VaCompanyCode = payType
	} else {
		oPay.Amount = total
		oPay.PayType = types.PayTypeRollOut
		oPay.VaCompanyCode = types.MobiFundVirtual
	}
	oPay.UserAccountId = tools.Int642Str(dataOrder.UserAccountId)
	oPay.UserBankCode = bankCode
	oPay.Ctime = tools.GetUnixMillis()
	oPay.Utime = tools.GetUnixMillis()
	oPay.AddPayment(&oPay)

	if dataProduct.ChargeFeeType == types.ProductChargeFeeInterestBefore || dataProduct.ChargeInterestType == types.ProductChargeInterestTypeHeadCut {
		//如果是砍头息，需要向还款记录中添加数据
		userEtrans := &models.User_E_Trans{}
		userEtrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
		userEtrans.UserAccountId = dataOrder.UserAccountId
		userEtrans.OrderId = dataOrder.Id
		if dataProduct.ChargeFeeType == types.ProductChargeFeeInterestBefore {
			userEtrans.Total += serviceFee
		}
		if dataProduct.ChargeInterestType == types.ProductChargeInterestTypeHeadCut {
			userEtrans.Total += interest
		}
		if !isRoll {
			userEtrans.VaCompanyCode = types.MobiPreInterest
			userEtrans.PayType = types.PayTypeMoneyIn
		} else {
			userEtrans.VaCompanyCode = types.MobiPreInterest
			userEtrans.PayType = types.PayTypeRefundIn
		}
		userEtrans.Ctime = tools.GetUnixMillis()
		userEtrans.Utime = tools.GetUnixMillis()
		userEtrans.AddEtrans(userEtrans)
		//进账
		userEtrans = &models.User_E_Trans{}
		userEtrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
		userEtrans.UserAccountId = dataOrder.UserAccountId
		userEtrans.OrderId = dataOrder.Id
		if dataProduct.ChargeFeeType == types.ProductChargeFeeInterestBefore {
			userEtrans.ServiceFee = serviceFee
		}
		if dataProduct.ChargeInterestType == types.ProductChargeInterestTypeHeadCut {
			userEtrans.PreInterest = interest
		}
		if !isRoll {
			userEtrans.VaCompanyCode = types.MobiPreInterest
			userEtrans.PayType = types.PayTypeMoneyOut
		} else {
			userEtrans.VaCompanyCode = types.MobiPreInterest
			userEtrans.PayType = types.PayTypeRefundOut
		}
		userEtrans.Ctime = tools.GetUnixMillis()
		userEtrans.Utime = tools.GetUnixMillis()
		userEtrans.AddEtrans(userEtrans)
		//出账
	}

	rp := repayplan.CreateRepayPlan(total, interest, serviceFee, dataOrder, &dataProduct)
	xendit.MarketPaymentCodeGenerate(dataOrder.Id, 0)

	dataOrder.Amount = rp.Amount
	UpdateOrderToWaitRepayment(dataOrder)

	if !isRoll {
		schema_task.PushBusinessMsg(types.PushTargetLoanSuccess, dataOrder.UserAccountId)

		accountBase, _ := dao.CustomerOne(dataOrder.UserAccountId)
		//date := tools.MDateMHS(dataOrder.ApplyTime)
		//content := fmt.Sprintf(i18n.GetMessageText(i18n.TextSmsDisburseSuccess), date)
		// service.SendSmsContent(types.ServiceDisburseSuccess, accountBase.Mobile, content)
		// 发短信新统一入口
		//sms.Send(types.ServiceDisburseSuccess, accountBase.Mobile, content, dataOrder.Id)
		param := make(map[string]interface{})
		param["related_id"] = dataOrder.Id
		schema_task.SendBusinessMsg(types.SmsTargetDisburseSuccess, types.ServiceDisburseSuccess, accountBase.Mobile, param)
	}

	accountCoupon, err := HandleCoupon(dataOrder)
	if err != nil || accountCoupon.Amount == 0 {
		logs.Info("[HandleDisburse] HandleCoupon no coupon to handle orderId:%d, err:%v", dataOrder.Id, err)
		return nil
	}

	couponM, _ := dao.GetCouponById(accountCoupon.CouponId)
	if couponM.CouponType == types.CouponTypeLimit {
		return nil
	}

	timetag := tools.GetUnixMillis()

	eInTrans := models.User_E_Trans{}
	eInTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
	eInTrans.OrderId = dataOrder.Id
	eInTrans.UserAccountId = dataOrder.UserAccountId
	eInTrans.VaCompanyCode = types.MobiCoupon
	eInTrans.Total = accountCoupon.Amount
	eInTrans.PayType = types.PayTypeMoneyIn
	eInTrans.CallbackJson = ""
	eInTrans.Ctime = timetag
	eInTrans.Utime = timetag
	eInTrans.AddEtrans(&eInTrans)

	eTrans := models.User_E_Trans{}
	eTrans.Id, _ = device.GenerateBizId(types.UserETransBiz)
	eTrans.OrderId = dataOrder.Id
	eTrans.UserAccountId = dataOrder.UserAccountId
	eTrans.VaCompanyCode = types.MobiCoupon
	eTrans.PayType = types.PayTypeMoneyOut
	eTrans.Ctime = timetag
	eTrans.Utime = timetag
	generateRepayTrans(dataOrder, &eTrans, &rp, accountCoupon.Amount)
	eTrans.AddEtrans(&eTrans)

	models.UpdateRepayPlan(&rp)

	oldOrder := *dataOrder

	dataOrder.CheckStatus = types.LoanStatusPartialRepayment
	dataOrder.Utime = tools.GetUnixMillis()

	models.UpdateOrder(dataOrder)

	models.OpLogWrite(0, dataOrder.Id, models.OpCodeOrderUpdate, dataOrder.TableName(), oldOrder, *dataOrder)

	monitor.IncrOrderCount(types.LoanStatusPartialRepayment)

	return nil
}

// UpdateOrderToAlreadyCleared 更新订单至结清, 唯一方法, 所有的结清必须调用此方法
// 如果需要额外更新,其他的order 信息, 可在外部set
func UpdateOrderToAlreadyCleared(order *models.Order, oldOrder models.Order) (num int64, err error) {
	if order.Id <= 0 {
		err = fmt.Errorf("[UpdateOrderToAlreadyCleared] wrong order info; order want Update to: %v, oldOrder: %v",
			*order, oldOrder)
		logs.Error(err)
		return
	}

	// 如果是首贷 结清。需将授权状态置为过期
	reloan := dao.IsRepeatLoan(order.UserAccountId)
	if !reloan {
		ExpireAllAuthorStatus(order.UserAccountId)
	}

	t := tools.GetUnixMillis()
	order.CheckStatus = types.LoanStatusAlreadyCleared
	order.FinishTime = t
	order.Utime = t
	num, err = models.UpdateOrder(order)
	if err != nil {
		logs.Error("[UpdateOrderToAlreadyCleared] sql err:", err, order, oldOrder)
		return
	}
	if num != 1 {
		err = fmt.Errorf("[UpdateOrderToAlreadyCleared] update err, affected rows:%d; order want Update to: %v, oldOrder: %v",
			num, *order, oldOrder)
		logs.Error(err)
		return
	}
	// 更新至结清成功, 写更新日志
	models.OpLogWrite(0, order.Id, models.OpCodeOrderUpdate, order.TableName(), oldOrder, *order)

	// 结清发短信
	schema_task.PushBusinessMsg(types.PushTargetClear, order.UserAccountId)

	// 用户还款事件触发 - 内含更新是否是复贷
	event.Trigger(&evtypes.RepaySuccessEv{
		OrderID:   order.Id,
		AccountID: order.UserAccountId,
		Time:      t,
	})

	// 更新成功后触发, 结清之后的事件
	monitor.IncrOrderCount(order.CheckStatus)
	// 结清触发逾期案件处理
	// TODO 建议放至异步事件里处理
	HandleOverdueCase(order.Id)

	return
}

func UpdateOrderToLoanFail(order *models.Order, err error) {
	logs.Error(err)

	oldOrder := *order

	order.CheckTime = tools.GetUnixMillis()
	order.Utime = tools.GetUnixMillis()
	order.CheckStatus = types.LoanStatusLoanFail
	models.UpdateOrder(order)

	schema_task.PushBusinessMsg(types.PushTargetLoanFail, order.UserAccountId)

	monitor.IncrOrderCount(types.LoanStatusLoanFail)

	models.OpLogWrite(0, order.Id, models.OpCodeOrderUpdate, order.TableName(), oldOrder, *order)
}

func RecordNewInvork(order *models.Order, accountProfile *models.AccountProfile, disbureStatus int, disburseCompany int, failedCode string) int64 {
	// 记录调用
	tag := tools.GetUnixMillis()
	invokeId, _ := device.GenerateBizId(types.DisburseRecordBiz)
	disburse := models.DisburseInvokeLog{
		Id:             invokeId,
		OrderId:        order.Id,
		UserAccountId:  order.UserAccountId,
		VaCompanyCode:  disburseCompany,
		BankName:       accountProfile.BankName,
		BankNo:         accountProfile.BankNo,
		DisbursementId: "",
		DisbureStatus:  disbureStatus,
		FailureCode:    failedCode,
		Ctime:          tag,
		Utime:          tag,
	}
	_, err := models.OrmInsert(&disburse)
	if err != nil {
		logs.Error("[CreateDisburse] OrmInsert err:%v disburse:%#v", err, disburse)
		return 0
	}
	return invokeId

}

func UpdateOrderToWaitRepayment(order *models.Order) {
	oldOrder := *order

	order.CheckStatus = types.LoanStatusWaitRepayment
	order.Utime = tools.GetUnixMillis()
	order.LoanTime = tools.GetUnixMillis()

	models.UpdateOrder(order)

	models.OpLogWrite(0, order.Id, models.OpCodeOrderUpdate, order.TableName(), oldOrder, *order)

	// 事件触发
	event.Trigger(&evtypes.LoanSuccessEv{OrderID: order.Id, Time: tools.GetUnixMillis()})

	monitor.IncrOrderCount(types.LoanStatusWaitRepayment)
}

// 跑批任务相关方法 {{{
//// 给大数据计算留5分钟的时间
func OrderListLoanStatus4Review() (list []models.Order, err error) {
	order := models.Order{}

	o := orm.NewOrm()
	o.Using(order.Using())

	//先查询风控主动通知队列
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()
	queueName := beego.AppConfig.String("risk_notify")

	queueNotifyByte, _ := storageClient.Do("LLEN", queueName)
	queueNotifyNum := queueNotifyByte.(int64)
	var limitNum int64
	if queueNotifyNum > 0 && queueNotifyNum > 100 {
		limitNum = 100
	}
	if queueNotifyNum > 0 && queueNotifyNum < 100 {
		limitNum = queueNotifyNum
	}
	if limitNum > 0 {
		accountIDSince := make([]string, 0)
		for i := int64(0); i < limitNum; i++ {
			qValueByte, _ := storageClient.Do("RPOP", queueName)
			orderID := string(qValueByte.([]byte))
			accountIDSince = append(accountIDSince, orderID)
		}
		accountIDstr := strings.Join(accountIDSince, ",")
		sql := fmt.Sprintf(`SELECT * FROM %s WHERE check_status = %d AND user_account_id in (%s) LIMIT %d`,
			order.TableName(), types.LoanStatus4Review, accountIDstr, 100)

		logs.Debug("[OrderListLoanStatus4Review] risk_notify: ", sql)
		_, err = o.Raw(sql).QueryRows(&list)
	} else {
		sql := fmt.Sprintf(`SELECT * FROM %s WHERE check_status = %d AND apply_time > 0 AND apply_time < %d LIMIT %d`,
			order.TableName(), types.LoanStatus4Review, tools.GetUnixMillis()-240000*60, 100)
		logs.Debug("[OrderListLoanStatus4Review] apply_time < 5 min : ", sql)
		_, err = o.Raw(sql).QueryRows(&list)
	}
	return
}

// 上一次查询最大值
func OrderList4ReviewAutoReduce(lastedOrderId int64) (list []models.Order, err error) {
	order := models.Order{}

	o := orm.NewOrm()
	o.Using(order.UsingSlave())

	cond := orm.NewCondition()
	cond = cond.And("check_status__in", types.LoanStatusOverdue, types.LoanStatusPartialRepayment, types.LoanStatusRolling)
	cond = cond.And("id__gt", lastedOrderId)

	_, err = o.QueryTable(order.TableName()).
		SetCond(cond).
		OrderBy("id").
		Limit(100).
		All(&list)
	return
}

func OrderListByStatus(status types.LoanStatus) (list []models.Order, err error) {
	order := models.Order{}

	o := orm.NewOrm()
	o.Using(order.Using())

	_, err = o.QueryTable(order.TableName()).Filter("check_status", status).OrderBy("id").Limit(100).All(&list)

	return
}

// 30天内没有操作过的临时已提交订单
func InvalidOrderList() (list []models.Order, err error) {
	order := models.Order{}

	o := orm.NewOrm()
	o.Using(order.Using())

	_, err = o.QueryTable(order.TableName()).
		Filter("check_status", types.LoanStatusSubmit).
		//Filter("is_temporary", types.IsTemporaryYes).
		Filter("ctime__lt", tools.GetUnixMillis()-3600000*24*30).
		OrderBy("id").Limit(100).All(&list)

	return
}

// 取给定条数的逾期订单
func GetOverdueOrderIDList(timetag int64, limit int64) (list []int64, err error) {
	orderM := models.Order{}
	repayPlan := models.RepayPlan{}
	orderExt := models.OrderExt{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	todayNatural := tools.NaturalDay(0)
	sql := fmt.Sprintf(`SELECT o.id FROM %s o
		LEFT JOIN %s r ON r.order_id = o.id
		LEFT JOIN %s e ON e.order_id = o.id
		WHERE o.check_status IN(%d, %d, %d) AND r.repay_date < %d AND o.is_dead_debt = %d
        AND (ISNULL(e.overdue_run_time) || e.overdue_run_time != %d)  ORDER BY r.repay_date DESC
		LIMIT %d`,
		orderM.TableName(),
		repayPlan.TableName(),
		orderExt.TableName(),
		types.LoanStatusWaitRepayment, types.LoanStatusPartialRepayment, types.LoanStatusOverdue,
		todayNatural, types.IsDeadDebtNo, timetag,
		limit)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

func GetRollApplyOrderList(idsBox []string, limit int64) (list []models.Order, err error) {
	if len(idsBox) <= 0 {
		logs.Warning("[GetRollApplyOrderList] 必要参数为空. idsBox:", idsBox)
		return
	}

	orderM := models.Order{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	sql := fmt.Sprintf(`SELECT o.* FROM %s o
WHERE o.id NOT IN(%s) AND o.check_status = %d AND o.is_dead_debt = %d
LIMIT %d`,
		orderM.TableName(),
		strings.Join(idsBox, ", "),
		types.LoanStatusRolling,
		types.IsDeadDebtNo,
		limit)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

// 还款提醒
func GetRepayRemindOrderList(idsBox []string) (list []int64, err error) {
	if len(idsBox) <= 0 {
		logs.Warning("[GetRepayRemindOrderList] 必要参数为空. idsBox:", idsBox)
		return
	}

	orderM := models.Order{}
	repayPlan := models.RepayPlan{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	beforeDay := tools.NaturalDay(1)
	afterDay := tools.NaturalDay(-1)
	sql := fmt.Sprintf(`SELECT o.id FROM %s o
LEFT JOIN %s r ON r.order_id = o.id
WHERE o.id NOT IN(%s) AND o.check_status IN(%d, %d) AND (r.repay_date = %d OR r.repay_date = %d)`,
		orderM.TableName(),
		repayPlan.TableName(),
		strings.Join(idsBox, ", "), types.LoanStatusWaitRepayment, types.LoanStatusPartialRepayment, beforeDay,
		afterDay)
	_, err = o.Raw(sql).QueryRows(&list)

	return
}

// GetRepayRemindCaseOrderList 还款提醒
func GetRepayRemindCaseOrderList(idsBox []string) (list []int64, err error) {
	if len(idsBox) <= 0 {
		logs.Warning("[GetRepayRemindCaseOrderList] 必要参数为空. idsBox:", idsBox)
		return
	}

	orderM := models.Order{}
	repayPlan := models.RepayPlan{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	// beforeDay := tools.NaturalDay(-1)
	// afterDay := tools.NaturalDay(1)
	today := tools.NaturalDay(0)
	sql := fmt.Sprintf(`SELECT o.id FROM %s o
LEFT JOIN %s r ON r.order_id = o.id
WHERE o.id NOT IN(%s) AND o.check_status IN(%d, %d) AND (r.repay_date = %d)`,
		orderM.TableName(),
		repayPlan.TableName(),
		strings.Join(idsBox, ", "), types.LoanStatusWaitRepayment, types.LoanStatusPartialRepayment, today)
	_, err = o.Raw(sql).QueryRows(&list)

	return
}

/*
// GetRepayRemindUnpaidAmount 还款提醒case未还金额,还款提醒总的未还本金
func GetRepayRemindUnpaidAmount() (unpaidCasePrincipal, unpaidPrincipal int64, err error) {

	startTime, _ := tools.GetTimeParseWithFormat(tools.MDateMHSDate(tools.GetUnixMillis()), "2006-01-02")
	startTime = startTime * 1000
	endTime := startTime + tools.MILLSSECONDADAY - 1

	var userOrderContainer []string

	ticket := models.Ticket{}
	o := orm.NewOrm()
	o.Using(ticket.UsingSlave())

	where := fmt.Sprintf("WHERE item_id=%d AND ctime>%d AND ctime<%d)",
		types.TicketItemRM0, startTime, endTime)
	sql := fmt.Sprintf("SELECT order_id FROM `%s` %s ", models.TICKET_TABLENAME, where)

	r := o.Raw(sql)
	r.QueryRows(&userOrderContainer)

	orderM := models.Order{}
	repayPlan := models.RepayPlan{}
	o.Using(orderM.Using())
	today := tools.NaturalDay(0)

	// 获取rm0 case总的应还金额
	sql = fmt.Sprintf(`select sum(r.amount-r.amount_payed) as unpaid_case_principal FROM %s r
	WHERE r.order_id IN(%s) `,
		repayPlan.TableName(),
		strings.Join(userOrderContainer, ", "))

	err = o.Raw(sql).QueryRow(&unpaidCasePrincipal)
	if err != nil {
		logs.Error("[GetRepayRemindUnpaidAmount] query unpaid amount should be ok, but err:", err)
	}

	// 获取rm0的总的应还金额
	sqlSum := fmt.Sprintf(`select sum(r.amount-r.amount_payed) as unpaid_principal FROM %s o
		LEFT JOIN %s r ON r.order_id = o.id
		WHERE o.check_status IN(%d, %d) AND (r.repay_date = %d)`,
		orderM.TableName(),
		repayPlan.TableName(),
		types.LoanStatusWaitRepayment, types.LoanStatusPartialRepayment, today)

	err = o.Raw(sqlSum).QueryRow(&unpaidPrincipal)
	if err != nil {
		logs.Error("[GetRepayRemindUnpaidAmount] query total unpaid amount should be ok, but err:", err)
	}

	return
}
*/

// 还款语音提醒
func GetRepayVoiceRemindOrderList(idsBox []string, t int) (list []int64, err error) {
	if len(idsBox) <= 0 {
		logs.Warning("[GetRepayVoiceRemindOrderList] 必要参数为空. idsBox:", idsBox)
		return
	}

	orderM := models.Order{}
	repayPlan := models.RepayPlan{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	day := tools.NaturalDay(int64(t))
	sql := fmt.Sprintf(`SELECT o.id FROM %s o
LEFT JOIN %s r ON r.order_id = o.id
WHERE o.id NOT IN(%s) AND o.check_status IN (%d, %d) AND r.repay_date = %d`,
		orderM.TableName(),
		repayPlan.TableName(),
		strings.Join(idsBox, ", "),
		types.LoanStatusWaitRepayment, types.LoanStatusPartialRepayment,
		day)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

// 催收短信提醒
func GetCollectionRemindOrderList(idsBox []string, collectionRemindDays []types.CollectionRemindDay) (list []int64, err error) {
	if len(idsBox) <= 0 {
		logs.Warning("[GetCollectionRemindOrderList] 必要参数为空. idsBox:", idsBox)
		return
	}

	orderM := models.Order{}
	repayPlan := models.RepayPlan{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	var remindDate []int64
	for _, val := range collectionRemindDays {
		remindDate = append(remindDate, tools.NaturalDay(int64(-1*val)))
	}

	remindDateStr := tools.ArrayToString(remindDate, ",")
	sql := fmt.Sprintf(`SELECT o.id FROM %s o
LEFT JOIN %s r ON r.order_id = o.id
WHERE o.id NOT IN(%s) AND o.check_status=%d AND r.repay_date IN (%s)`,
		orderM.TableName(),
		repayPlan.TableName(),
		strings.Join(idsBox, ", "), types.LoanStatusOverdue,
		remindDateStr,
	)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

// 风控状态为'等待自动外呼'的订单
func GetWaitAutoCallOrderList(idsBox []string) (list []int64, err error) {
	if len(idsBox) <= 0 {
		logs.Warning("[GetWaitAutoCallOrderList] 必要参数为空. idsBox:", idsBox)
		return
	}

	orderM := models.Order{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	sql := fmt.Sprintf(`SELECT o.id FROM %s o
WHERE o.id NOT IN(%s) AND o.risk_ctl_status = %d`,
		orderM.TableName(),
		strings.Join(idsBox, ", "),
		types.RiskCtlWaitAutoCall)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

// 逾期自动外呼
func GetOverdueOrderListByDays(idsBox []string, days string) (list []int64, err error) {
	if len(idsBox) <= 0 {
		logs.Warning("[GetOverdueOrderList] 必要参数为空. idsBox:", idsBox)
		return
	}

	overdueCase := models.OverdueCase{}
	o := orm.NewOrm()
	o.Using(overdueCase.Using())

	sql := fmt.Sprintf(`SELECT o.order_id FROM %s o
WHERE o.is_out = %d AND o.overdue_days IN(%s)`,
		overdueCase.TableName(),
		types.IsUrgeOutNo,
		days)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

/**
bluepay退出历史舞台
*/
func CreateVirtualAccounts(userAccountId int64, amount int64, orderId int64, mobile string, vaCompanyType int) (err error) {
	if vaCompanyType == types.Xendit {
		err = XenditCreateVirtualAccounts(userAccountId, orderId)
	} else if vaCompanyType == types.DoKu {
		err = DokuCreateVirtualAccounts(userAccountId, orderId)
	} else {
		err = fmt.Errorf("CreateVirtualAccounts不支持这家银行.vaCompany is %d, orderId is %d", vaCompanyType, orderId)
	}
	return
}

/**
Xendit 创建虚拟账户
*/
func XenditCreateVirtualAccounts(userAccountId int64, orderId int64) error {
	accountBase, err := models.OneAccountBaseByPkId(userAccountId)
	if err != nil || accountBase.Id <= 0 {
		err = fmt.Errorf("user_account_id相关的account_base记录不存在！user_account_id is: %d", userAccountId)
		return err
	}
	accountProfile, err := dao.CustomerProfile(userAccountId)
	one, err := models.OneBankInfoByFullName(accountProfile.BankName)
	if err != nil {
		logs.Error("[XenditCreateVirtualAccounts] OneBankInfoByFullName err:%v. check bank name:%s userAccountId:%d", err, accountProfile.BankName, userAccountId)
		return err
	}

	var datas = make(map[string]interface{})
	datas["bank_name"] = accountProfile.BankName
	datas["account_id"] = userAccountId
	datas["account_name"] = accountBase.Realname
	datas["company_name"] = "xendit"
	datas["order_id"] = orderId
	datas["banks_info"] = one

	payApi, err := CreatePaymentApi(types.Xendit, datas)
	if err != nil {
		logs.Error("[XenditCreateVirtualAccounts] err:", err)
		return err
	}

	resJson, err := payApi.CreateVirtualAccount(datas)
	if err != nil {
		logs.Error("[XenditCreateVirtualAccounts] err:", err)
		return err
	}

	err = payApi.CreateVirtualAccountResponse(resJson, datas)
	return err
}

/**
bluepay创建虚拟账户
*/
func BluepayCreateVirtualAccounts(userAccountId int64, price int64, orderId int64, mobile string) error {
	accountBase, err := models.OneAccountBaseByPkId(userAccountId)
	if err != nil || accountBase.Id <= 0 {
		err = fmt.Errorf("user_account_id相关的account_base记录不存在！user_account_id is: %d", userAccountId)
		return err
	}

	accountProfile, err := dao.CustomerProfile(userAccountId)
	one, err := models.OneBankInfoByFullName(accountProfile.BankName)
	if err != nil {
		logs.Error("[XenditCreateVirtualAccounts] OneBankInfoByFullName err:%v. check bank name:%s userAccountId:%d", err, accountProfile.BankName, userAccountId)
		return err
	}

	var datas = make(map[string]interface{})
	datas["amount"] = price
	datas["order_id"] = orderId
	datas["mobile"] = mobile
	datas["bank_name"] = accountProfile.BankName
	datas["account_id"] = userAccountId
	datas["company_name"] = "bluepay"
	datas["banks_info"] = one

	payApi, err := CreatePaymentApi(types.Bluepay, datas)
	if err != nil {
		logs.Error("[BluepayCreateVirtualAccounts] err:", err)
		return err
	}

	resJson, err := payApi.CreateVirtualAccount(datas)
	if err != nil {
		logs.Error("[BluepayCreateVirtualAccounts] err:", err)
		return err
	}

	err = payApi.CreateVirtualAccountResponse(resJson, datas)
	return err
}

/**
doku创建虚拟账户
*/
func DokuCreateVirtualAccounts(userAccountId int64, orderId int64) error {
	accountBase, err := models.OneAccountBaseByPkId(userAccountId)
	if err != nil || accountBase.Id <= 0 {
		err = fmt.Errorf("user_account_id相关的account_base记录不存在！user_account_id is: %d", userAccountId)
		return err
	}
	accountProfile, err := dao.CustomerProfile(userAccountId)
	one, err := models.OneBankInfoByFullName(accountProfile.BankName)
	if err != nil {
		logs.Error("[XenditCreateVirtualAccounts] OneBankInfoByFullName err:%v. check bank name:%s userAccountId:%d", err, accountProfile.BankName, userAccountId)
		return err
	}

	var datas = make(map[string]interface{})
	datas["bank_name"] = accountProfile.BankName
	datas["account_id"] = userAccountId
	datas["account_name"] = accountBase.Realname
	datas["company_name"] = "xendit"
	datas["order_id"] = orderId
	datas["banks_info"] = one

	payApi, err := CreatePaymentApi(types.DoKu, datas)
	if err != nil {
		logs.Error("[DoKuCreateVirtualAccounts] err:", err)
		return err
	}

	_, err = payApi.CreateVirtualAccount(datas)
	if err != nil {
		logs.Error("[DoKuCreateVirtualAccounts] err:", err)
	}
	//doku 创建VA账户不需要向doku服务器发起请求，这个流程和xendit不一样
	return err

}

func GetEaccountsDesc(accountId int64) (eAccountDesc []string) {
	accounts, _ := models.GetMultiEAccounts(accountId)
	if len(accounts) > 0 {
		for i := 0; i < len(accounts); i++ {
			bankCode := accounts[i].BankCode
			if accounts[i].VaCompanyCode == types.DoKu {
				bankCode = doku.DoKuVaBankCodeTransform(accounts[i].BankCode)
			}
			eAccountDesc = append(eAccountDesc, fmt.Sprintf("%s %s", bankCode, accounts[i].EAccountNumber))
			//accounts[i].BankCode = eAccountDesc
		}
	}
	return
}

/**
优先选择第三方付款
*/
func PriorityThirdpartyDisburse(bankName string) (thirdPartyPay int, err error) {
	one, err := models.OneBankInfoByFullName(bankName)
	if err != nil {
		logs.Error("[PriorityThirdpartyDisburse] OneBankInfoByFullName err:%v. check bank name:%s", err, bankName)
		return
	}
	thirdPartyPay = one.LoanCompanyCode
	logs.Debug("The priority loan bank info:", thirdPartyPay)
	return
}

func checkAndAcquireEaccount(accountBase *models.AccountBase, accountProfile *models.AccountProfile, loan int64, relatedId int64, vaCompanyType int) error {

	//userEAccount, err := models.GetEAccount(orderData.UserAccountId, vaCompanyType)
	userEAccount, err := models.GetLastestActiveEAccountByVacompanyType(accountBase.Id, vaCompanyType)
	if err != nil {
		//此时需要创建虚拟账户
		err = CreateVirtualAccounts(accountBase.Id, loan, relatedId, accountBase.Mobile, vaCompanyType)
		if err != nil {
			return err
		}

		VAStartTime := time.Now().Unix()
		//等待回调，先等待30秒
		//创建虚拟账户是异步操作，需要再次查询数据库确认虚拟账户创建成功
		for {
			time.Sleep(time.Second)
			userEAccount, err = models.GetLastestActiveEAccountByVacompanyType(accountBase.Id, vaCompanyType)
			VAEndTime := time.Now().Unix()
			if err == nil {
				break
			}

			if VAEndTime-VAStartTime <= 60 {
				continue
			}

			errStr := fmt.Sprintf("[CreateDisburse] CreateVirtualAccounts timeout account_id is: %d, userEAccount is %#v", accountBase.Id, userEAccount)
			err = fmt.Errorf(errStr)
			return err
		}

	} else {
		_, err = xendit.BankName2Code(accountProfile.BankName)
		if err != nil {
			err = fmt.Errorf("the user's bankBankName is not in the valid list, somebody changes it! bankName is: %s", accountProfile.BankName)
			return err
		}
		//上述银行列表，如果在客户端上做更改了，就需要去掉
		//目前用xendit列表去查是因为目前客户端列表都是xendit的银行列表
		vaCompanyType = userEAccount.VaCompanyCode
	}

	if userEAccount.Status != "ACTIVE" {
		errStr := fmt.Sprintf("[CreateDisburse] CreateVirtualAccounts account not active account_id is: %d, userEAccount is %#v", accountBase.Id, userEAccount)
		err = fmt.Errorf(errStr)
		return err
	}

	return nil
}

/**
Xendit 放款操作
*/
func CreateDisburse(orderId int64) (invokeId int64, err error) {
	orderData, err := models.GetOrder(orderId)
	if err != nil {
		errStr := fmt.Sprintf("order不存在, order状态不是等待放款. order id is : %d", orderId)
		err = fmt.Errorf(errStr)
		return 0, err
	}
	accountBase, err := models.OneAccountBaseByPkId(orderData.UserAccountId)
	if err != nil || accountBase.Id <= 0 {
		errStr := fmt.Sprintf("orderid相关的acccount_base记录不存在,orderid is: %d", orderId)
		err = fmt.Errorf(errStr)
		return 0, err
	}

	accountProfile, err := dao.CustomerProfile(orderData.UserAccountId)
	thirdPartyVACreate, _ := dao.PriorityThirdpartyVACreate(accountProfile.BankName)
	vaCompanyType := thirdPartyVACreate
	thirdPartyDisburse, _ := LoanCompany(orderId, accountProfile.BankName)

	// 名字包含数字
	if tools.ContainNumber(accountBase.Realname) {

		err = fmt.Errorf("[CreateDisburse] ContainNumber err. orderID:%d Realname:%s", orderId, accountBase.Realname)
		UpdateOrderToLoanFail(&orderData, err)
		RecordNewInvork(&orderData, accountProfile, types.DisbureStatusCallFailed, thirdPartyDisburse, "Name_Contain_Number")
		return
	}

	// 银行名字不在列表里
	one, err := models.OneBankInfoByFullName(accountProfile.BankName)
	if err != nil {
		//logs.Error("[CreateDisburse] OneBankInfoByFullName err:%v. check bank name:%s userAccountId:%d", err, accountProfile.BankName, accountProfile.AccountId)
		err = fmt.Errorf("[CreateDisburse] OneBankInfoByFullName err:%v. check bank name:%s userAccountId:%d", err, accountProfile.BankName, accountProfile.AccountId)
		UpdateOrderToLoanFail(&orderData, err)
		RecordNewInvork(&orderData, accountProfile, types.DisbureStatusCallFailed, thirdPartyDisburse, "bank_name_not_support")
		return
	}

	// 可能由于正在切换放款路由
	if !IsCompanySupport(thirdPartyDisburse, one) {
		thirdPartyDisburse = one.LoanCompanyCode
		if thirdPartyDisburse == 0 {
			err = fmt.Errorf("[CreateDisburse] check IsCompanySupport ret false. thirdPartyDisburse:%d one:%#v  userAccountId:%d", thirdPartyDisburse, one, accountProfile.AccountId)
			logs.Error(err)
			return
		}
	}

	err = CreateVirtualAccountAll(orderData.UserAccountId, orderData.Id)
	if err != nil {
		logs.Warn("[CreateDisburse] CreateVirtualAccountAll err:%v orderId:%d", err, orderId)
	}

	// 检查是否有指定的va
	err = checkAndAcquireEaccount(&accountBase, accountProfile, orderData.Loan, orderData.Id, vaCompanyType)
	if err != nil {
		logs.Warn("[CreateDisburse] checkAndAcquireEaccount err:%v orderId:%d", err, orderId)
		//return 0, err
	}

	// 是否至少有一个可用va
	eAccount, err := dao.GetActiveEaccountWithBankName(accountBase.Id)
	if eAccount.Id == 0 {
		err = fmt.Errorf("[CreateDisburse] checkAndAcquireEaccount err:%v orderId:%d . no active va .", err, orderId)
		UpdateOrderToLoanFail(&orderData, err)
		RecordNewInvork(&orderData, accountProfile, types.DisbureStatusCallFailed, thirdPartyDisburse, "no_active_va")
		return
	}

	//防止多协程同时放款
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()
	keyPrefix := beego.AppConfig.String("disburse_order_lock")
	key := fmt.Sprintf("%s%d", keyPrefix, orderData.Id)
	re, err := cacheClient.Do("SET", key, tools.GetUnixMillis(), "EX", 24*60*60, "NX")
	if err != nil || re == nil {
		logs.Error("[CreateDisburse] 防止多协程同时放款 orderId:%d err:%v", orderId, err)
		return
	}

	invokeId = RecordNewInvork(&orderData, accountProfile, 0, thirdPartyDisburse, "")

	//订单置为放款中
	tag := tools.GetUnixMillis()
	originOrder := orderData
	orderData.CheckStatus = types.LoanStatusIsDoing
	orderData.CheckTime = tag
	orderData.Utime = tag
	models.UpdateOrder(&orderData)
	monitor.IncrOrderCount(orderData.CheckStatus)
	// 写操作日志
	models.OpLogWrite(0, orderData.Id, models.OpCodeOrderUpdate, orderData.TableName(), originOrder, orderData)

	// 调用第三方放款
	err = ThirdPartyDisburse(orderId, accountBase.Id, accountProfile.BankName, accountBase.Realname, accountProfile.BankNo, orderData.Loan, thirdPartyDisburse, invokeId)

	return invokeId, err
}

// 退款统一使用xendit
func CreateRefund(refund *models.Refund) (invokeId int64, err error) {
	if refund == nil {
		errStr := fmt.Sprintf("[CreateRefund] parm nil.")
		return 0, fmt.Errorf(errStr)
	}
	accountBase, err := models.OneAccountBaseByPkId(refund.UserAccountId)
	if err != nil || accountBase.Id <= 0 {
		errStr := fmt.Sprintf("[CreateRefund]refund相关的acccount_base记录不存在,refund is: %#v", refund)
		err = fmt.Errorf(errStr)
		return 0, err
	}

	accountProfile, err := dao.CustomerProfile(refund.UserAccountId)
	err = checkAndAcquireEaccount(&accountBase, accountProfile, refund.Amount, refund.Id, types.Xendit)
	if err != nil {
		logs.Error("[CreateRefund] checkAndAcquireEaccount err:%v refundId:%d", err, refund.Id)
		return 0, err
	}

	// 记录调用
	tag := tools.GetUnixMillis()
	invokeId, _ = device.GenerateBizId(types.DisburseRecordBiz)
	disburse := models.DisburseInvokeLog{
		Id:             invokeId,
		OrderId:        refund.Id,
		UserAccountId:  accountBase.Id,
		VaCompanyCode:  types.Xendit,
		BankName:       accountProfile.BankName,
		BankNo:         accountProfile.BankNo,
		DisbursementId: "",
		Ctime:          tag,
		Utime:          tag,
	}
	_, err = models.OrmInsert(&disburse)
	if err != nil {
		logs.Error("[CreateDisburse] OrmInsert err:%v disburse:%#v", err, disburse)
		return 0, err
	}

	err = ThirdPartyDisburse(refund.Id, accountBase.Id, accountProfile.BankName, accountBase.Realname, accountProfile.BankNo, refund.Amount, types.Xendit, invokeId)
	if err != nil {
		logs.Error("[CreateRefund] ThirdPartyDisburse err:%s refund:%#v", err, refund)
		return invokeId, err
	}
	return invokeId, err
}

func ThirdPartyDisburse(orderId int64, userAccountId int64, bankName string, realname string, bankNo string, amount int64, vaCompanyCode int, invokeId int64) (err error) {
	desc := tools.Int642Str(orderId)

	bankInfo, err := models.OneBankInfoByFullName(bankName)
	if err != nil {
		logs.Error("[ThirdPartyDisburse] OneBankInfoByFullName err:%v. unsport bank name:%s accountId:%d orderId：%d", err, bankName, userAccountId, orderId)
		return
	}
	var datas = make(map[string]interface{})
	datas["bank_name"] = bankName
	datas["account_id"] = userAccountId
	datas["order_id"] = orderId
	datas["account_name"] = realname
	datas["account_num"] = bankNo
	datas["amount"] = amount
	datas["desc"] = desc
	datas["invoke_id"] = invokeId
	datas["banks_info"] = bankInfo

	if vaCompanyCode == types.DoKu {
		datas["company_name"] = "doku"
	} else {
		datas["company_name"] = "xendit"
	}

	payApi, err := CreatePaymentApi(vaCompanyCode, datas)
	if err != nil {
		return err
	}

	resJson, err := payApi.Disburse(datas)
	if err != nil {
		return err
	}

	err = payApi.DisburseResponse(resJson, datas)
	if err != nil {
		return err
	}

	return err
}

// 逾期订单更新还款计划表
func OverdueUpdatePenalty(orderId int64) error {
	order, err := models.GetOrder(orderId)
	if err != nil {
		logs.Error("[OverdueUpdatePenalty] models.GetOrder, orderId: %d, err: %v", orderId, err)
		return err
	}

	repayPlan, err := models.GetLastRepayPlanByOrderid(orderId)
	if err != nil {
		logs.Error("[OverdueUpdatePenalty] models.GetLastRepayPlanByOrderid, orderId: %d err: %v", orderId, err)
		return err
	}

	product, err := models.GetProduct(order.ProductId)
	if err != nil {
		logs.Error("[OverdueUpdatePenalty]= models.GetProduct , orderId: %d  product_id %d err: %v", orderId, order.ProductId, err)
		return err
	}

	originOrder := order
	originRepay := repayPlan

	// 逾期超过90天,就认为坏帐了,不再更新还款计划表
	if repayPlan.RepayDate+90*3600*1000*24 < tools.GetUnixMillis() {
		order.IsDeadDebt = types.IsDeadDebtYes
		_, err = models.UpdateOrder(&order)
		if err != nil {
			logs.Error("[OverdueUpdatePenalty] order.UpdateOrder, orderId: %d, err: %v", orderId, err)
		}
		//坏账后不再计算罚息了
		return err
	}

	if order.CheckStatus == types.LoanStatusWaitRepayment ||
		order.CheckStatus == types.LoanStatusPartialRepayment ||
		order.CheckStatus == types.LoanStatusOverdue {
		today := tools.NaturalDay(0)

		if today > repayPlan.RepayDate && today > order.PenaltyUtime {
			//当前时间大于还款时间同时要大于最后更新的时间(防止一天内更新多次)

			//例子：如果还款期限是7，3.11号是放款时间，那么正常还款时间为3.18(包含3.18)之前，宽限期为3.19，
			// 大于3.19号的时间(单位为天)就是逾期，例如3.20就已经逾期了

			gracePeriodDay := tools.BaseDayOffset(repayPlan.RepayDate, int64(product.GracePeriod))

			if today <= gracePeriodDay {
				//宽限期内
				diffAmount := repayPlan.Amount - repayPlan.AmountPayed - repayPlan.AmountReduced
				gracePeriodInterestFloat := float64(diffAmount) * float64(product.DayGraceRate) / float64(types.ProductFeeBase)
				gracePeriodInterest, _ := tools.Str2Float64(fmt.Sprintf("%f", gracePeriodInterestFloat))
				repayPlan.GracePeriodInterest += int64(gracePeriodInterest) //此利息和产品商量过目前就定成这个
				repayPlan.Utime = tools.GetUnixMillis()

				repayPlanOverdueObj := &models.RepayPlanOverdue{}
				repayPlanOverdueObj.GracePeriodInterest = repayPlan.GracePeriodInterest
				t := time.Now()
				date := tools.GetDate(t.Unix())
				baseUm := tools.GetDateParse(date) * 1000
				repayPlanOverdueObj.OverdueDate = tools.MDateMHSDateNumber(baseUm)
				repayPlanOverdueObj.OrderId = orderId
				repayPlanOverdueObj.Ctime = tools.GetUnixMillis()
				repayPlanOverdueObj.Utime = tools.GetUnixMillis()
				models.AddRepayPlanOverdue(repayPlanOverdueObj)
				//宽限期的记录增加明细

				_, err = models.UpdateRepayPlan(&repayPlan)
				if err != nil {
					logs.Error("[OverdueUpdatePenalty] models.UpdateRepayPlan, orderId: %d, err: %v", orderId, err)
					return err
				}
				order.PenaltyUtime = today

				//更新结清减免预减免宽限期利息
				prereduced, _ := dao.GetLastPrereducedByOrderid(orderId)
				canReducedAmount := repayplan.CaculateCanReducedAmount(int64(gracePeriodInterest), prereduced.DerateRatio)
				prereduced.GracePeriodInterestPrededuced += canReducedAmount
				prereduced.GraceInterestReduced += canReducedAmount
				prereduced.Utime = tools.GetUnixMillis()
				cols := []string{"grace_period_interest_prereduced", "grace_interest_reduced", "Utime"}
				models.OrmUpdate(&prereduced, cols)

			} else {
				//大于宽限期就表示逾期
				diffAmount := repayPlan.Amount - repayPlan.AmountPayed - repayPlan.AmountReduced
				todayPenalty := float64(diffAmount) * float64(product.DayPenaltyRate) / float64(types.ProductFeeBase)
				penaltyFloat := float64(repayPlan.Penalty) + todayPenalty
				penalty, _ := tools.Str2Float64(fmt.Sprintf("%f", penaltyFloat)) //四舍五入
				repayPlan.Penalty = int64(penalty)
				repayPlan.Utime = tools.GetUnixMillis()

				repayPlanOverdueObj := &models.RepayPlanOverdue{}
				repayPlanOverdueObj.Penalty = int64(penalty)
				t := time.Now()
				date := tools.GetDate(t.Unix())
				baseUm := tools.GetDateParse(date) * 1000
				repayPlanOverdueObj.OverdueDate = tools.MDateMHSDateNumber(baseUm)
				repayPlanOverdueObj.OrderId = orderId
				repayPlanOverdueObj.Ctime = tools.GetUnixMillis()
				repayPlanOverdueObj.Utime = tools.GetUnixMillis()
				models.AddRepayPlanOverdue(repayPlanOverdueObj)
				//逾期的记录增加明细

				_, err = models.UpdateRepayPlan(&repayPlan)
				if err != nil {
					logs.Error("[OverdueUpdatePenalty] models.UpdateRepayPlan, orderId: %d, err: %v", orderId, err)
					return err
				}
				order.PenaltyUtime = today

				//更新结清减免的预减免罚息
				prereduced, _ := dao.GetLastPrereducedByOrderid(orderId)
				canReducedAmount := repayplan.CaculateCanReducedAmount(int64(todayPenalty), prereduced.DerateRatio)
				prereduced.PenaltyPrereduced += canReducedAmount
				prereduced.PenaltyReduced += canReducedAmount
				prereduced.Utime = tools.GetUnixMillis()
				cols := []string{"penalty_prereduced", "penalty_reduced", "Utime"}
				models.OrmUpdate(&prereduced, cols)
			}
			_, err = models.UpdateOrder(&order)

			if err != nil {
				logs.Error("[OverdueUpdatePenalty] order.UpdateOrder, orderId: %d, err: %v", orderId, err)
			}

			// 写入修改日志
			models.OpLogWrite(0, order.Id, models.OpCodeOrderUpdate, order.TableName(), originOrder, order)
			models.OpLogWrite(0, repayPlan.Id, models.OpCodeRepayPlanUpdate, repayPlan.TableName(), originRepay, repayPlan)
		}
	}

	//更新付款码
	event.Trigger(&evtypes.FixPaymentCodeEv{OrderID: orderId})
	//xendit.MarketPaymentCodeGenerate(orderId, false, 0)

	return err
}

// SameContactApplyLoanOrderInLastMonth 近1个月同联系人在我司申请人数
func SameContactApplyLoanOrderInLastMonth(limit int64, contacts ...string) (pass bool, applyNum int64, issueContact string, err error) {
	lastPeriodTimeNode := tools.GetUnixMillis() - 3600000*24*30
	pass = true
	for _, contact := range contacts {
		pass, applyNum, err = sameContactApplyLoanOrderInLastPeriod(contact, lastPeriodTimeNode, limit)
		if !pass {
			issueContact = contact
			return
		}
	}
	return
}

// SameContactApplyLoanOrderInLast3Month 近3个月同联系人在我司申请人数
func SameContactApplyLoanOrderInLast3Month(limit int64, contacts ...string) (pass bool, applyNum int64, issueContact string, err error) {
	lastPeriodTimeNode := tools.GetUnixMillis() - 3600000*24*30*3
	pass = true
	for _, contact := range contacts {
		pass, applyNum, err = sameContactApplyLoanOrderInLastPeriod(contact, lastPeriodTimeNode, limit)
		if !pass {
			issueContact = contact
			return
		}
	}
	return
}

// sameContactApplyLoanOrderInLastPeriod 最近一段时期lastPeriod同联系人在我司申请人数
// lastPeriod : 一段时间前的毫秒时间戳, 一个月前, 三个月前
func sameContactApplyLoanOrderInLastPeriod(contact string, lastPeriodTimeNode int64, limit int64) (pass bool, applyNum int64, err error) {
	pass = true
	contact = tools.Escape(contact)
	accountProfile := models.AccountProfile{}
	o := orm.NewOrm()
	o.Using(accountProfile.UsingSlave())

	var strAccountIDs []string

	sql := fmt.Sprintf("SELECT `account_id` FROM `%s` WHERE contact1 = '%s' OR contact2= '%s'", accountProfile.TableName(), contact, contact)
	num, err := o.Raw(sql).QueryRows(&strAccountIDs)
	if err != nil || num <= 0 {
		logs.Debug("[SameContactApplyLoanOrderInLastMonth] err:", err)
		return
	}
	// 如果联系人下的子用户数不小于 limit 则总申请人数必然小于 limit
	if num < limit {
		return
	}

	var loanStatusBox []string
	for _, v := range types.ProcessingLoanStatus {
		loanStatusBox = append(loanStatusBox, fmt.Sprintf("%d", v))
	}

	orderM := models.Order{}
	sql = fmt.Sprintf("SELECT COUNT(DISTINCT `user_account_id`) AS total FROM `%s` WHERE user_account_id IN(%s) AND check_status IN(%s) AND apply_time >= %d",
		orderM.TableName(), strings.Join(strAccountIDs, ", "), strings.Join(loanStatusBox, ", "), lastPeriodTimeNode)
	err = o.Raw(sql).QueryRow(&applyNum)
	if err != nil {
		logs.Debug("[SameContactApplyLoanOrderInLastMonth] err:", err)
	}
	if applyNum >= limit {
		pass = false
	}

	return
}

// SameResidentAddressApplyLoanOrderInLast3Month 近1个月内同一居住地址的申请人数
func SameResidentAddressApplyLoanOrderInLast3Month(limit int64, address string) (pass bool, applyNum int64, err error) {
	lastPeriodTimeNode := tools.GetUnixMillis() - 3600000*24*30*3
	pass, applyNum, err = sameResidentAddressApplyLoanOrderInLastPeriod(address, lastPeriodTimeNode, limit)
	return
}

// SameResidentAddressApplyLoanOrderInHistory 历史同一居住地址的申请人数
func SameResidentAddressApplyLoanOrderInHistory(limit int64, address string) (pass bool, applyNum int64, err error) {
	pass, applyNum, err = sameResidentAddressApplyLoanOrderInLastPeriod(address, 0, limit)
	return
}

func sameResidentAddressApplyLoanOrderInLastPeriod(address string, lastPeriodTimeNode int64, limit int64) (pass bool, applyNum int64, err error) {
	// 默认通过
	pass = true
	accountProfile := models.AccountProfile{}
	o := orm.NewOrm()
	o.Using(accountProfile.Using())

	var strAccountIDs []string

	sql := fmt.Sprintf("SELECT account_id FROM `%s` WHERE resident_address = '%s'", accountProfile.TableName(), tools.Escape(address))
	num, err := o.Raw(sql).QueryRows(&strAccountIDs)
	if err != nil || num <= 0 {
		logs.Debug("[sameResidentAddressApplyLoanOrderInLastPeriod] err:", err)
		return
	}
	// 如果联系人下的子用户数不小于 limit 则总申请人数必然小于 limit
	if num < limit {
		return
	}

	var loanStatusBox []string
	for _, v := range types.ProcessingLoanStatus {
		loanStatusBox = append(loanStatusBox, fmt.Sprintf("%d", v))
	}

	orderM := models.Order{}
	sql = fmt.Sprintf("SELECT COUNT(DISTINCT `user_account_id`) AS total FROM `%s` WHERE user_account_id IN(%s) AND check_status IN(%s)",
		orderM.TableName(), strings.Join(strAccountIDs, ", "), strings.Join(loanStatusBox, ", "))
	if lastPeriodTimeNode > 0 {
		sql += fmt.Sprintf(" AND apply_time >= %d", lastPeriodTimeNode)
	}

	err = o.Raw(sql).QueryRow(&applyNum)
	if err != nil {
		logs.Debug("[SameContactApplyLoanOrderInLastMonth] err:", err)
	}
	if applyNum >= limit {
		pass = false
	}

	return
}

func UpdateIsRepeatUser(accountId int64) {

	var num int64
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	hashName := dao.HashRepeatLoanKey()
	// 从db获取已结清的订单数,再放到redis
	num, err := models.GetClearedOrderNumByAccountId(accountId)
	logs.Info("[RecordRepeatUser], redis no data, hashName: %s, accountId: %d, num: %d", hashName, accountId, num)
	if err != nil {
		logs.Error("[RecordRepeatUser], get already cleared order num fail, accountId: %d, err: %#v", hashName, accountId, err)
		return
	}

	storageClient.Do("HSET", hashName, accountId, num)

	return
}

// SameCompanyApplyLoanOrderInLast3Month 近1个月内同一居住地址的申请人数
func SameCompanyApplyLoanOrderInLast3Month(limit int64, company string) (pass bool, applyNum int64, err error) {
	if company == "" {
		pass = true
		return
	}

	lastPeriodTimeNode := tools.GetUnixMillis() - 3600000*24*30*3
	pass, applyNum, err = sameCompanyApplyLoanOrderInLastPeriod(company, lastPeriodTimeNode, limit)
	return
}

func sameCompanyApplyLoanOrderInLastPeriod(company string, lastPeriodTimeNode int64, limit int64) (pass bool, applyNum int64, err error) {
	// 默认通过
	pass = true

	accountProfile := models.AccountProfile{}
	o := orm.NewOrm()
	o.Using(accountProfile.UsingSlave())

	var strAccountIDs []string

	sql := fmt.Sprintf("SELECT account_id FROM `%s` WHERE company_name = '%s'", accountProfile.TableName(), tools.AddSlashes(company))
	num, err := o.Raw(sql).QueryRows(&strAccountIDs)
	if err != nil || num <= 0 {
		logs.Debug("[sameCompanyApplyLoanOrderInLastPeriod] err:", err)
		return
	}
	// 如果联系人下的子用户数不小于 limit 则总申请人数必然小于 limit
	if num < limit {
		return
	}

	var loanStatusBox []string
	for _, v := range types.ProcessingLoanStatus {
		loanStatusBox = append(loanStatusBox, fmt.Sprintf("%d", v))
	}

	orderM := models.Order{}
	sql = fmt.Sprintf("SELECT COUNT(DISTINCT `user_account_id`) AS total FROM `%s` WHERE user_account_id IN(%s) AND check_status IN(%s)",
		orderM.TableName(), strings.Join(strAccountIDs, ", "), strings.Join(loanStatusBox, ", "))
	if lastPeriodTimeNode > 0 {
		sql += fmt.Sprintf(" AND apply_time >= %d", lastPeriodTimeNode)
	}

	err = o.Raw(sql).QueryRow(&applyNum)
	if err != nil {
		logs.Debug("[SameContactApplyLoanOrderInLastMonth] err:", err)
	}
	if applyNum >= limit {
		pass = false
	}

	return
}

func IsUserPerfectInformation(accountId int64, loan int64, period int) (ok bool, step int) {
	tmpOrder, _ := dao.AccountLastTmpLoanOrderByCond(accountId, loan, period)
	clientInfo, _ := models.OneLastClientInfoByRelatedID(tmpOrder.Id)
	phase := ProfileCompletePhase(accountId, clientInfo.UiVersion, clientInfo.AppVersionCode)

	logs.Debug("tmpOrder:%#v", tmpOrder, " clientInfo:%#v", clientInfo, " phase:%d", phase)
	return types.AccountInfoCompleteAddition == phase ||
		types.AccountInfoCompletePhaseDone == phase ||
		types.AccountInfoCompletePhaseLiveReLoan == phase, phase

}

// IsUserPerfectInformationTwo（首贷借贷流程变化）
func IsUserPerfectInformationTwo(accountId int64, loan int64, period int) (ok bool, step int) {
	tmpOrder, _ := dao.AccountLastTmpLoanOrderByCond(accountId, loan, period)
	clientInfo, _ := models.OneLastClientInfoByRelatedID(tmpOrder.Id)
	_, phase := ProfileCompletePhaseTwo(accountId, clientInfo.UiVersion, clientInfo.AppVersionCode)

	logs.Debug("tmpOrder:%#v", tmpOrder, " clientInfo:%#v", clientInfo, " phase:%d", phase)
	return types.AccountInfoComplete == phase ||
		types.AccountInfoAddition == phase ||
		types.AccountInfoCompletePhaseLiveReLoan == phase ||
		types.AccountInfoCompletePhaseJumpToAuthoriation == phase, phase
}

// 根据配置和当前时间算出周末的开始时间戳
func currentWeekendStartTs() int64 {
	conf := config.ValidItemString("first_loan_weekend_start_time")
	s := strings.Split(conf, "--")
	if len(s) < 2 {
		logs.Error("[CurrentWeekendStartTs] config split err. conf:%s  s:%#v", conf, s)
		return tools.GetUnixMillis() + 1
	}

	if indexConf, ok := types.DayMap[strings.ToUpper(s[0])]; ok {

		// 获取当前时周几
		nowDay := tools.GetNowWeekDayConf()
		offset := indexConf - nowDay

		thatTime := tools.GetUnixMillis() + int64(offset)*tools.MILLSSECONDADAY
		thatDay := tools.GetLocalDateFormat(thatTime, "2006-01-02")

		thatStr := thatDay + " " + s[1]

		restlt, _ := tools.GetTimeParseWithFormat(thatStr, "2006-01-02 15:04:05")
		logs.Debug("[CurrentWeekendStartTs] thatTime:%d offset:%d thatStr:%s restlt:%d", thatTime, offset, thatStr, restlt)

		return restlt * 1000
	} else {
		logs.Error("[CurrentWeekendStartTs] index err. conf:%s s:%#v", conf, s)
		return tools.GetUnixMillis() + 1
	}

}

// 根据时区算出周日的23：59：59 时间戳
func currentWeekendEndTs() int64 {

	conf := config.ValidItemString("first_loan_weekend_end_time")
	s := strings.Split(conf, "--")
	if len(s) < 2 {
		logs.Error("[currentWeekendEndTs] config split err. conf:%s  s:%#v", conf, s)
		return tools.GetUnixMillis() + 1
	}

	if indexConf, ok := types.DayMap[strings.ToUpper(s[0])]; ok {

		// 获取当前时周几
		nowDay := tools.GetNowWeekDayConf()
		offset := indexConf - nowDay

		thatTime := tools.GetUnixMillis() + int64(offset)*tools.MILLSSECONDADAY
		thatDay := tools.GetLocalDateFormat(thatTime, "2006-01-02")

		thatStr := thatDay + " " + s[1]

		restlt, _ := tools.GetTimeParseWithFormat(thatStr, "2006-01-02 15:04:05")
		logs.Debug("[currentWeekendEndTs] thatTime:%d offset:%d thatStr:%s restlt:%d", thatTime, offset, thatStr, restlt)

		return restlt * 1000
	} else {
		logs.Error("[currentWeekendEndTs] index err. conf:%s s:%#v", conf, s)
		return tools.GetUnixMillis() + 1
	}
}

func IsWeekend() bool {
	st := currentWeekendStartTs()
	et := currentWeekendEndTs() // 加1秒代表周一零点
	nt := tools.GetUnixMillis()

	logs.Debug("[IsWeekend] st:%d et:%d nt:%d", st, et, nt)
	if et <= st {
		// 终值 小于 起点
		//   |×××true××××××××终止--------false-------起点×××××××true×××××|
		if nt <= st && nt > et {
			return false
		} else {
			return true
		}
	} else {
		// 终值 大于 起点
		//  |×××false××××××××起点---------true--------终止×××××××false×××××|
		if nt >= st && nt <= et {
			return true
		} else {
			return false
		}
	}

}

func CheckAndDoAutoReduce(orderId int64) {
	order, _ := models.GetOrder(orderId)
	obj, _ := models.GetLastRepayPlanByOrderid(orderId)

	// 自动减免逻辑  此时还的钱已存在于 repayPlan内
	if ok, reduce := CanAutoReduce(&order, &obj, 0); ok {
		logs.Info("[CheckAndDoAutoReduce] do auto reduce")
		origin := obj
		oldOrder := order
		err := DoAutoReduce(&order, &obj, reduce)
		if err == nil {
			UpdateOrderToAlreadyCleared(&order, oldOrder)

			models.UpdateRepayPlan(&obj)
			//更新还款计划

			models.OpLogWrite(0, obj.Id, models.OpAutoReduction, obj.TableName(), origin, obj)

			// TODO 所有其他的减免置为失效

		} else {
			logs.Warn("[CheckAndDoAutoReduce] DoAutoReduce err:%v repayPlan:%#v", err, obj)
		}
	}
}

func loanBankFullNameList(loanCompany int) (ret string) {
	list, err := models.BankListByCompanyType(loanCompany, types.LoanRepayTypeLoan)

	if err != nil {
		logs.Error("[loanBankFullNameList] BankListByCompanyType err:%v", err)
		return
	}

	names := []string{}
	for _, one := range list {
		v := "'" + one.FullName + "'"
		names = append(names, v)
	}
	ret = strings.Join(names, ",")
	return
}

func RollBackOrder(opUid int64, orderId int64) (err error) {
	order, err := models.GetOrder(orderId)
	if err != nil {
		logs.Error("[RollBackOrder] orderId is not valid:%d", orderId)
		return
	}

	if order.CheckStatus != types.LoanStatusWaitRepayment &&
		order.CheckStatus != types.LoanStatusOverdue &&
		order.CheckStatus != types.LoanStatusPartialRepayment {
		err = fmt.Errorf("[RollBackOrder] order status not 7 or 9. status:%d", order.CheckStatus)
		return
	}

	repayPlan, _ := models.GetLastRepayPlanByOrderid(orderId)
	if repayPlan.Id > 0 {
		models.OrmDelete(&repayPlan)
		models.OpLogWrite(opUid, orderId, models.OpDelectForRollBackLoan, repayPlan.TableName(), repayPlan, "")
	}

	payment, _ := models.GetPaymentByOrderIdPayType(orderId, types.PayTypeMoneyOut)
	if payment.Id > 0 {
		//删除payment放款记录
		models.OrmDelete(&payment)
		models.OpLogWrite(opUid, orderId, models.OpDelectForRollBackLoan, payment.TableName(), payment, "")
	}

	//userEtrans1, _ := models.GetEtranByOrderIdPayTypeVaCompanyCode(orderId, 1, 1001)
	//userEtrans2, _ := models.GetEtranByOrderIdPayTypeVaCompanyCode(orderId, 2, 1001)

	trans := models.GetETransByOrderId(orderId)
	for _, v := range trans {
		//删除砍头息进账出账
		if v.Id > 0 && v.VaCompanyCode == types.MobiPreInterest {
			models.OrmDelete(&v)
			models.OpLogWrite(opUid, orderId, models.OpDelectForRollBackLoan, v.TableName(), v, "")
		} else {
			logs.Error("[RollBackOrder] user trans err. trans:%#v orderID:%d opUid:%d", v, order.Id, opUid)
		}
	}

	overdueList, _ := models.GetRepayPlanOverdueByOrderId(orderId)
	for _, v := range overdueList {
		if v.Id > 0 {
			models.OrmDelete(&v)
			models.OpLogWrite(opUid, orderId, models.OpDelectForRollBackLoan, v.TableName(), v, "")
		}
		//删除所有的逾期罚息记录
	}

	desc := tools.Int642Str(orderId)
	mobiEtrans, _ := models.GetMobiEtransByAccountIdDescription(order.UserAccountId, desc)
	if mobiEtrans.Id > 0 {
		models.OrmDelete(&mobiEtrans)
		models.OpLogWrite(opUid, orderId, models.OpDelectForRollBackLoan, mobiEtrans.TableName(), mobiEtrans, "")
		//删除mobi放款记录
	}

	// 删除ticket
	rollBackTicket(orderId)

	//删除overdue case
	rollBackOverdueCase(orderId)

	//删除 repay remaind
	rollBackRepayRemaind(orderId)

	old := order
	order.CheckStatus = types.LoanStatusLoanFail
	order.Utime = tools.GetUnixMillis()
	order.UpdateOrder(&order)
	models.OpLogWrite(opUid, orderId, models.OpDelectForRollBackLoan, order.TableName(), old, order)

	logs.Debug("orderId has been cleared data.", orderId)
	return nil
}

func rollBackTicket(orderId int64) {
	ticket := models.Ticket{}
	o := orm.NewOrm()
	o.Using(ticket.Using())

	sql := "delete  from %s where order_id = %d and item_id > 2"

	sql = fmt.Sprintf(sql,
		ticket.TableName(),
		orderId)

	err := o.Raw(sql).QueryRow(&ticket)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[rollBackTicket] orderId:%d err:%v",
			orderId, err)
	}

	return
}

func rollBackOverdueCase(orderId int64) {
	t := models.OverdueCase{}
	o := orm.NewOrm()
	o.Using(t.Using())

	sql := "delete from %s where order_id = %d"

	sql = fmt.Sprintf(sql,
		t.TableName(),
		orderId)

	err := o.Raw(sql).QueryRow(&t)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[rollBackOverdueCase] orderId:%d err:%v",
			orderId, err)
	}

	return
}

func rollBackRepayRemaind(orderId int64) {
	t := models.RepayRemindCase{}
	o := orm.NewOrm()
	o.Using(t.Using())

	sql := "delete from %s where order_id = %d"

	sql = fmt.Sprintf(sql,
		t.TableName(),
		orderId)

	err := o.Raw(sql).QueryRow(&t)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[rollBackRepayRemaind] orderId:%d err:%v",
			orderId, err)
	}

	return
}
