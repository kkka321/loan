package entrust

import (
	"fmt"
	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/overdue"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/thirdparty/xendit"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"
	"sync"

	"micro-loan/common/service"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type OverdueCaseBaseInfo struct {
	OrderID           string  `json:"orderID"`           //订单编号
	ContractNumber    string  `json:"contractNumber"`    //合同编号
	ContractAmount    float64 `json:"contractAmount"`    //合同金额
	OverdueAmount     float64 `json:"overdueAmount"`     //逾期总⾦额
	OverdueCapital    float64 `json:"overdueCapital"`    //逾期本金
	OverDueInterest   float64 `json:"overDueInterest"`   //逾期利息
	OverdueFine       float64 `json:"overdueFine"`       //逾期罚息
	OverdueDelayFine  float64 `json:"overdueDelayFine"`  //逾期滞纳金
	Periods           int64   `json:"periods"`           //总期数
	OverduePeriods    int64   `json:"overduePeriods"`    //逾期期数
	OverdueDays       int64   `json:"overdueDays"`       //逾期天数
	OverDueDate       string  `json:"overDueDate"`       //逾期日期
	PerDueDate        string  `json:"perDueDate"`        //还款日期
	HasPayAmount      float64 `json:"hasPayAmount"`      //逾期已还款金额
	HasPayPeriods     int64   `json:"hasPayPeriods"`     //已还款期数
	LatelyPayDate     string  `json:"latelyPayDate"`     //最近还款日期
	LatelyPayAmount   float64 `json:"latelyPayAmount"`   //最近还款金额
	LoanDate          string  `json:"loanDate"`          //贷款日期
	OverdueManageFee  float64 `json:"overdueManageFee"`  //逾期管理费
	LeftCapital       float64 `json:"leftCapital"`       //剩余本金
	LeftInterest      float64 `json:"leftInterest"`      //剩余利息
	PersonalName      string  `json:"personalName"`      //客户姓名
	IDCard            string  `json:"idCard"`            //身份证号码
	MobileNo          string  `json:"mobileNo"`          //手机号
	IDCardAddress     string  `json:"idCardAddress"`     //身份证地址
	HomeAddress       string  `json:"homeAddress"`       //现居住地址
	ProductName       string  `json:"productName"`       //产品名
	ProductSeriesName string  `json:"productSeriesName"` //产品系列名称
	City              string  `json:"city"`              //业务所在城市
	DepositBank       string  `json:"depositBank"`       //客户还款卡银行
	CardNumber        string  `json:"cardNumber"`        //客户还款卡号
	CompanyName       string  `json:"companyName"`       //工作单位名称
	CompanyAddr       string  `json:"companyAddr"`       //工作单位地址
	CompanyPhone      string  `json:"companyPhone"`      //工作单位电话
	Province          string  `json:"province"`          //业务所在省份
	Married           string  `json:"married"`           //是否已婚
}

type RepayInfo struct {
	OrderID          string  `json:"orderID"`          //订单编号
	OverdueAmount    float64 `json:"overdueAmount"`    //逾期总⾦额
	OverdueCapital   float64 `json:"overdueCapital"`   //逾期本金
	OverDueInterest  float64 `json:"overDueInterest"`  //逾期利息
	OverdueFine      float64 `json:"overdueFine"`      //逾期罚息
	OverdueDelayFine float64 `json:"overdueDelayFine"` //逾期滞纳金
	LatelyPayDate    string  `json:"latelyPayDate"`    //最近还款日期
	LatelyPayAmount  float64 `json:"latelyPayAmount"`  //最近还款金额
	LeftCapital      float64 `json:"leftCapital"`      //剩余本金
	LeftInterest     float64 `json:"leftInterest"`     //剩余利息
	OverduePeriods   int64   `json:"overduePeriods"`   //逾期期数
	OverdueDays      int64   `json:"overdueDays"`      //逾期天数
	OverdueManageFee float64 `json:"overdueManageFee"` //逾期管理费
	IsEndCase        int64   `json:"isEndCase"`        //是否结清
}

type Contact struct {
	IDCard       string `json:"idCard"`       //联系人身份证
	ContactName  string `json:"contactName"`  //联系人姓名
	ContactPhone string `json:"contactPhone"` //联系人手机
}
type Contacts struct {
	CItem []Contact `json:"Citem"`
}

//roll tentative calculation   展期试算
type RollTC struct {
	OrderID         string  `json:"orderID"`         //订单ID
	IsCanRoll       int     `json:"isCanRoll"`       //是否允许展期
	MinRepay        float64 `json:"minRepay"`        //展期最小还款金额
	LatestRepayTime int64   `json:"latestRepayTime"` //最晚还款日期
	RollNeedRepay   float64 `json:"rollNeedRepay"`   //展期应该金额
	RollRepayTime   int64   `json:"rollRepayTime"`   //展期应还时间
}

type SPaymentCode struct {
	OrderID     string  `json:"orderID"`     //订单ID
	PaymentCode string  `json:"paymentCode"` //付款码
	Amount      float64 `json:"amount"`      //付款码金额
	Status      string  `json:"status"`      //状态
	PayCompany  string  `json:"payCompany"`  //支付公司
	Ctime       int64   `json:"ctime"`       //创建时间
	ExpiryDate  int64   `json:"expiryDate"`  //过期时间
}

func CheckOverdueBaseInfoRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"order_id_list": true,
	}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}
func CheckCallbackRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"order_id_list": true,
	}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

func CheckRepayStatusRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"order_id_list": true,
	}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

func CheckContactRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"id_card_list": true,
	}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

//获取可委外案件 ，满足>M2 & 入催 & 没被委外过
func GetEntrustList(orderIDStr, pname string) (caseList []OverdueCaseBaseInfo, count int64) {

	overdueCases, count, _ := GetEntrustOverdueCases(orderIDStr, pname, 100)
	logs.Debug("[GetEntrustList] overdue case count, unfiltered: %d", count)

	if len(overdueCases) > 0 {

		orderIDS := make([]int64, 0)
		for _, oc := range overdueCases {
			orderIDS = append(orderIDS, oc.OrderId)
		}
		//获取可被委外的订单集合
		logs.Debug("[GetEntrustList] orderIDs:", orderIDS)

		canEntustOrderIDSMapData := overdue.EdgeMultiOrdersFilterSelfUrge(orderIDS)
		logs.Debug("[GetEntrustList] overdue case count, filtered:", len(canEntustOrderIDSMapData), "data:", canEntustOrderIDSMapData)

		for _, overdueCase := range overdueCases {
			orderID := overdueCase.OrderId
			if _, ok := canEntustOrderIDSMapData[orderID]; ok {
				order, err := models.GetOrder(orderID)
				if err != nil {
					return
				}
				hasPayPeriods := int64(0)
				if order.CheckStatus == types.LoanStatusAlreadyCleared {
					hasPayPeriods = 1
				}
				repayPlan, _ := models.GetLastRepayPlanByOrderid(orderID)
				product, _ := models.GetProduct(order.ProductId)

				accountBase, _ := models.OneAccountBaseByPkId(order.UserAccountId)
				accountProfile, _ := models.OneAccountProfileByAccountID(order.UserAccountId)
				//thirdPartyPay, _, _ := service.PriorityThirdpartyPay(accountProfile.BankName)
				userEAccount, _ := dao.GetActiveEaccountWithBankName(order.UserAccountId)
				userETrans := models.GetLastInPayETransByOrderID(orderID)
				totalRepayAmout := repayplan.CaculateRepayTotalAmount(repayPlan.Amount, repayPlan.AmountPayed, repayPlan.AmountReduced,
					repayPlan.GracePeriodInterest, repayPlan.GracePeriodInterestPayed, repayPlan.GracePeriodInterestReduced,
					repayPlan.Penalty, repayPlan.PenaltyPayed, repayPlan.PenaltyReduced)
				totalAmount := repayplan.CaculateTotalAmount(repayPlan.Amount, repayPlan.AmountPayed, repayPlan.AmountReduced)
				totalGracPeriodInterest := repayplan.CaculateTotalGracePeriod(repayPlan.GracePeriodInterest, repayPlan.GracePeriodInterestPayed, repayPlan.GracePeriodInterestReduced)
				totalPenalty := repayplan.CaculateTotalPenalty(repayPlan.Penalty, repayPlan.PenaltyPayed, repayPlan.PenaltyReduced)
				var marital = ""
				maritalStatusConf := models.GetMaritalStatusConf()
				if v, ok := maritalStatusConf[accountProfile.MaritalStatus]; ok {
					marital = fmt.Sprintf(`%s`, v)
				}
				ocInfo := OverdueCaseBaseInfo{}
				ocInfo.OrderID = tools.Int642Str(orderID)
				ocInfo.ContractNumber = tools.Int642Str(orderID)                                                                   //    string  `json:"contractNumber"`    //合同编号
				ocInfo.ContractAmount = float64(repayPlan.Amount)                                                                  //    float64 `json:"contractAmount"`    //合同金额
				ocInfo.OverdueAmount = float64(totalRepayAmout)                                                                    //float64 `json:"overdueAmount"`     //逾期总⾦额
				ocInfo.OverdueCapital = float64(totalAmount)                                                                       //float64 `json:"overdueCapital"`    //逾期本金
				ocInfo.OverDueInterest = float64(totalGracPeriodInterest)                                                          //float64 `json:"overDueInterest"`   //逾期利息
				ocInfo.OverdueFine = float64(totalPenalty)                                                                         //float64 `json:"overdueFine"`       //逾期罚息
				ocInfo.OverdueDelayFine = 0                                                                                        //float64 `json:"overdueDelayFine"`  //逾期滞纳金
				ocInfo.Periods = 1                                                                                                 //int64   `json:"periods"`           //总期数
				ocInfo.OverduePeriods = 1                                                                                          //int64   `json:"overduePeriods"`    //逾期期数
				ocInfo.OverdueDays = int64(overdueCase.OverdueDays)                                                                //int64   `json:"overdueDays"`       //逾期天数
				ocInfo.OverDueDate = tools.MDateMHSDate(overdueCase.JoinUrgeTime)                                                  //string  `json:"overDueDate"`       //逾期日期
				ocInfo.PerDueDate = tools.MDateMHSDate(repayPlan.RepayDate)                                                        //string  `json:"perDueDate"`        //还款日期
				ocInfo.HasPayAmount = float64(repayPlan.AmountPayed + repayPlan.GracePeriodInterestPayed + repayPlan.PenaltyPayed) //float64 `json:"hasPayAmount"`      //逾期已还款金额
				ocInfo.HasPayPeriods = hasPayPeriods                                                                               //int64   `json:"hasPayPeriods"`     //已还款期数
				ocInfo.LatelyPayDate = tools.MDateMHSDate(userETrans.Ctime)                                                        //string  `json:"latelyPayDate"`     //最近还款日期
				ocInfo.LatelyPayAmount = float64(userETrans.Total)                                                                 //float64 `json:"latelyPayAmount"`   //最近还款金额
				ocInfo.LoanDate = tools.MDateMHSDate(repayPlan.Ctime)                                                              //string  `json:"loanDate"`          //贷款日期
				ocInfo.OverdueManageFee = 0                                                                                        //float64 `json:"overdueManageFee"`  //逾期管理费
				ocInfo.LeftCapital = float64(totalAmount)                                                                          //float64 `json:"leftCapital"`       //剩余本金
				ocInfo.LeftInterest = float64(totalGracPeriodInterest)                                                             //float64 `json:"leftInterest"`      //剩余利息
				ocInfo.PersonalName = accountBase.Realname                                                                         //string  `json:"personalName"`      //客户姓名
				ocInfo.IDCard = accountBase.Identity                                                                               //string  `json:"idCard"`            //身份证号码
				ocInfo.MobileNo = accountBase.Mobile                                                                               //string  `json:"mobileNo"`          //手机号
				ocInfo.IDCardAddress = ""                                                                                          //string  `json:"idCardAddress"`     //身份证地址
				ocInfo.HomeAddress = accountProfile.ResidentCity + " " + accountProfile.ResidentAddress                            //string  `json:"homeAddress"`       //现居住地址
				ocInfo.ProductName = product.Name                                                                                  //string  `json:"productName"`       //产品名
				ocInfo.ProductSeriesName = ""                                                                                      //string  `json:"productSeriesName"` //产品系列名称
				ocInfo.City = accountProfile.CompanyCity                                                                           //string  `json:"city"`              //业务所在城市
				ocInfo.DepositBank = userEAccount.BankCode                                                                         //string  `json:"depositBank"`       //客户还款卡银行
				ocInfo.CardNumber = userEAccount.EAccountNumber                                                                    //string  `json:"cardNumber"`        //客户还款卡号
				ocInfo.CompanyName = accountProfile.CompanyName                                                                    //string  `json:"companyName"`       //工作单位名称
				ocInfo.CompanyAddr = accountProfile.CompanyAddress                                                                 //string  `json:"companyAddr"`       //工作单位地址
				ocInfo.CompanyPhone = accountProfile.CompanyTelephone                                                              //string  `json:"companyPhone"`      //工作单位电话
				ocInfo.Province = ""                                                                                               //string  `json:"province"`          //业务所在省份
				ocInfo.Married = marital                                                                                           //string  `json:"married"`           //是否已婚
				caseList = append(caseList, ocInfo)
			}
		}
	}
	logs.Debug("[GetEntrustList] overdue case count, filtered caseList:", len(caseList), "data:", caseList)

	return
}

// GetEntrustOverdueCases 获取可推送给勤为的的案件
func GetEntrustOverdueCases(orderIDStr, pname string, limit int) (overdueCases []models.OverdueCase, num int64, err error) {
	entrustDay, err := config.ValidItemInt("outsource_day")
	if err != nil {
		entrustDay = types.EntrustDay
		logs.Warning("[GetEntrustOverdueCases] entrust day config losed:", entrustDay)
	}
	//如果指定订单，返回指定订单(包含已打勤为标记的订单)，否则返回可被勤为的订单
	and := "AND 1"
	if orderIDStr != "" {
		and = fmt.Sprintf(`AND oc.order_id IN(%s)`, orderIDStr)
	} else {
		and = fmt.Sprintf(`AND oe.is_entrust=%d and entrust_pname='%s'`, 0, pname)
	}
	overdueCase := models.OverdueCase{}
	orderExt := models.OrderExt{}
	o := orm.NewOrm()
	o.Using(overdueCase.UsingSlave())
	sql := fmt.Sprintf(`SELECT oc.* FROM %s oc
	LEFT JOIN %s oe on oc.order_id=oe.order_id
	WHERE oc.overdue_days>=%d
	AND oc.is_out = %d
	`+and+`
	LIMIT %d`,
		overdueCase.TableName(),
		orderExt.TableName(),
		entrustDay,
		types.IsOverdueNo,
		limit)
	num, err = o.Raw(sql).QueryRows(&overdueCases)
	return
}

// ProcessedNotify 勤为已处理回调，收到回调，我方关闭催收工单，标记已勤为，打勤为标记订单不再升级
func ProcessedNotify(orderIDStr, pname string) (count int) {
	//  委外案件处理：  关闭工单，原因 已委外 案件不升级 但案件计息方式不变
	orderIDS := strings.Split(orderIDStr, ",")
	if len(orderIDS) > 0 {

		for _, orderIDStr := range orderIDS {

			orderID, _ := tools.Str2Int64(orderIDStr)
			if orderID > 0 {
				oneCase, err := dao.GetInOverdueCaseByOrderID(orderID)
				if err != nil {
					return
				}
				item := types.MustGetTicketItemIDByCaseName(oneCase.CaseLevel)
				//关闭工单
				ticket.CloseByRelatedID(oneCase.Id, item, types.TicketCloseReasonEntrust)
				//标记为已委外
				orderExt, err1 := models.GetOrderExt(orderID)
				if err1 != nil {
					return
				}
				unixtime := tools.GetUnixMillis()
				if orderExt.IsEntrust == 0 {
					orderExt.IsEntrust = 1
					orderExt.Utime = unixtime
					orderExt.EntrustTime = unixtime
					orderExt.EntrustPname = pname
					orderExt.Update()
				}
				count++
			}
		}
	}
	return
}

//GetRepayStatus 获取订单支付状态
func GetRepayStatus(orderIDStr string) (repayInfoList []RepayInfo, count int) {

	orderIDS := strings.Split(orderIDStr, ",")
	orderCount := len(orderIDS)

	succCounter := struct {
		count int
		mutex sync.Mutex
	}{}
	if orderCount > 0 {
		ch := make(chan int64, len(orderIDS))
		for _, orderIDStr := range orderIDS {
			orderID, _ := tools.Str2Int64(orderIDStr)
			ch <- orderID
		}
		logs.Notice("[GetRepayStatus] len :", len(ch))
		for {
			if len(ch) == 0 {
				logs.Debug("[GetRepayStatus] complete ,i'll quit")
				break
			}
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				if len(ch) == 0 {
					logs.Debug("[GetRepayStatus] complete ,i'll quit")
					break
				}
				wg.Add(1)
				orderID := <-ch
				logs.Notice("[GetRepayStatus] orderID:", orderID, " workID:", i)
				go func(orderID int64) {
					defer wg.Done()

					order, err := models.GetOrder(orderID)
					if err != nil {
						logs.Error("[GetRepayStatus] err:", err)
						return
					}

					rpInfo := RepayInfo{}
					inEndCase := int64(0)
					if order.CheckStatus == types.LoanStatusAlreadyCleared ||
						order.CheckStatus == types.LoanStatusRollClear {
						userETrans := models.GetLastInPayETransByOrderID(orderID)
						oneCase, _ := dao.GetInOverdueCaseByOrderID(orderID)
						rpInfo.OrderID = tools.Int642Str(orderID)
						rpInfo.LatelyPayDate = tools.MDateMHSDate(userETrans.Ctime) //string  `json:"latelyPayDate"`    //最近还款日期
						rpInfo.LatelyPayAmount = float64(userETrans.Total)          //float64 `json:"latelyPayAmount"`  //最近还款金额
						rpInfo.OverduePeriods = 1                                   //int64   `json:"overduePeriods"`   //逾期期数
						rpInfo.OverdueDays = int64(oneCase.OverdueDays)             //int64   `json:"overdueDays"`      //逾期天数
						rpInfo.IsEndCase = 1
					} else {
						oneCase, _ := dao.GetInOverdueCaseByOrderID(orderID)
						repayPlan, err2 := models.GetLastRepayPlanByOrderid(orderID)
						if err2 != nil {
							logs.Error("[GetRepayStatus] err2:", err2)
							return
						}
						userETrans := models.GetLastInPayETransByOrderID(orderID)
						totalRepayAmout := repayplan.CaculateRepayTotalAmount(repayPlan.Amount, repayPlan.AmountPayed, repayPlan.AmountReduced,
							repayPlan.GracePeriodInterest, repayPlan.GracePeriodInterestPayed, repayPlan.GracePeriodInterestReduced,
							repayPlan.Penalty, repayPlan.PenaltyPayed, repayPlan.PenaltyReduced)
						totalAmount := repayplan.CaculateTotalAmount(repayPlan.Amount, repayPlan.AmountPayed, repayPlan.AmountReduced)
						totalGracPeriodInterest := repayplan.CaculateTotalGracePeriod(repayPlan.GracePeriodInterest, repayPlan.GracePeriodInterestPayed, repayPlan.GracePeriodInterestReduced)
						totalPenalty := repayplan.CaculateTotalPenalty(repayPlan.Penalty, repayPlan.PenaltyPayed, repayPlan.PenaltyReduced)

						rpInfo.OrderID = tools.Int642Str(orderID)                   //string  `json:"orderID"`          //订单编号
						rpInfo.OverdueAmount = float64(totalRepayAmout)             //float64 `json:"overdueAmount"`    //逾期总⾦额
						rpInfo.OverdueCapital = float64(totalAmount)                //float64 `json:"overdueCapital"`   //逾期本金
						rpInfo.OverDueInterest = float64(totalGracPeriodInterest)   //float64 `json:"overDueInterest"`  //逾期利息
						rpInfo.OverdueFine = float64(totalPenalty)                  //float64 `json:"overdueFine"`      //逾期罚息
						rpInfo.OverdueDelayFine = 0                                 //float64 `json:"overdueDelayFine"` //逾期滞纳金
						rpInfo.LatelyPayDate = tools.MDateMHSDate(userETrans.Ctime) //string  `json:"latelyPayDate"`    //最近还款日期
						rpInfo.LatelyPayAmount = float64(userETrans.Total)          //float64 `json:"latelyPayAmount"`  //最近还款金额
						rpInfo.LeftCapital = float64(totalAmount)                   //float64 `json:"leftCapital"`      //剩余本金
						rpInfo.LeftInterest = float64(totalGracPeriodInterest)      //float64 `json:"leftInterest"`     //剩余利息
						rpInfo.OverduePeriods = 1                                   //int64   `json:"overduePeriods"`   //逾期期数
						rpInfo.OverdueDays = int64(oneCase.OverdueDays)             //int64   `json:"overdueDays"`      //逾期天数
						rpInfo.OverdueManageFee = 0                                 //float64 `json:"overdueManageFee"` //逾期管理费
						rpInfo.IsEndCase = inEndCase                                //int64   `json:"isEndCase"`        //是否结清
					}
					succCounter.mutex.Lock()
					repayInfoList = append(repayInfoList, rpInfo)
					succCounter.mutex.Unlock()
					succCounter.count++
				}(orderID)
			}
			wg.Wait()
			count = succCounter.count
		}
	}
	return
}

//GetContact 获取联系人信息
func GetContact(IDCards string) (contactList []Contacts, count int) {

	IDCardSlice := strings.Split(IDCards, ",")
	if len(IDCardSlice) > 0 {
		for _, IDcardStr := range IDCardSlice {
			if IDcardStr != "" {
				accountBase, err := models.OneAccountBaseByIdentity(IDcardStr)
				if err != nil {
					return
				}
				bigdataContact, num, err1 := models.OneAccountBigdataContactByAccountID(accountBase.Id)
				if err1 != nil {
					return
				}
				contacts := Contacts{}
				if num > 0 {
					for _, v := range bigdataContact {
						contact := Contact{}
						contact.IDCard = accountBase.Identity //string `json:"idCard"`          //联系人身份证
						contact.ContactName = v.ContactName   //string `json:"contactName"`     //联系人姓名
						contact.ContactPhone = v.Mobile
						contacts.CItem = append(contacts.CItem, contact)
					}
				}
				//string `json:"contactTIme"`     //通话时长
				contactList = append(contactList, contacts)
				count++
			}
		}
	}

	return
}

func GetRepayList(pname string) (orderIDs []int64) {
	prefix := beego.AppConfig.String("entrust_notify_repay_queue_prefix")
	if prefix != "" {
		pnameKey := prefix + pname
		logs.Debug("[EntrustRepayList] pnamekey:", pnameKey)
		storageClient := storage.RedisStorageClient.Get()
		defer storageClient.Close()

		for {
			qValueByte, err := storageClient.Do("RPOP", pnameKey)
			// 没有可供消费的数据,退出工作 goroutine
			if err != nil || qValueByte == nil {
				logs.Info("[GetRepayList] no data for consume, I will exit")
				break
			}
			id, _ := tools.Str2Int64(string(qValueByte.([]byte)))
			orderIDs = append(orderIDs, id)
		}

	}
	return
}

//GetRollTC 获取订单展期信息
func GetRollTC(orderIDStr string) (rollTCList []RollTC, count int) {

	orderIDS := strings.Split(orderIDStr, ",")
	orderCount := len(orderIDS)

	succCounter := struct {
		count int
		mutex sync.Mutex
	}{}
	if orderCount > 0 {
		ch := make(chan int64, len(orderIDS))
		for _, orderIDStr := range orderIDS {
			orderID, _ := tools.Str2Int64(orderIDStr)
			ch <- orderID
		}
		logs.Notice("[GetRollTC] len :", len(ch))
		for {
			if len(ch) == 0 {
				logs.Debug("[GetRollTC]for complete ,i'll quit")
				break
			}
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				if len(ch) == 0 {
					logs.Debug("[GetRollTC]goruntine complete ,i'll quit")
					break
				}
				wg.Add(1)
				orderID := <-ch
				logs.Notice("[GetRollTC] orderID:", orderID, " workID:", i)
				go func(orderID int64) {
					defer wg.Done()
					order, err := models.GetOrder(orderID)
					if err != nil {
						return
					}
					period, minRepay, _, err1 := service.CalcRollRepayAmount(order)
					logs.Notice("[GetRollTC] period:", period, "orderID:", orderID)
					if err1 != nil {
						return
					}
					isCanRoll := 0
					if service.IsOrderExtension(order) {
						isCanRoll = 1
					}
					rollTC := RollTC{}
					rollTC.OrderID = tools.Int642Str(orderID)                //string  `json:"order_id"`          //订单ID
					rollTC.IsCanRoll = isCanRoll                             //int     `json:"is_can_roll"`       //是否允许展期
					rollTC.MinRepay = float64(minRepay)                      //float64 `json:"min_repay"`         //展期最小还款金额
					rollTC.LatestRepayTime = tools.GetIDNCurrDayLastSecond() //int64   `json:"latest_repay_time"` //最晚还款日期
					rollTC.RollNeedRepay = float64(order.Amount)             //float64 `json:"roll_need_pay"`     //展期应该金额
					rollTC.RollRepayTime = tools.NaturalDay(int64(period))   //int64   `json:"roll_repay_time"`   //展期应还时间
					succCounter.mutex.Lock()
					rollTCList = append(rollTCList, rollTC)
					succCounter.mutex.Unlock()
					succCounter.count++
				}(orderID)
			}
			wg.Wait()
			count = succCounter.count
		}
	}
	return
}

//GetSPaymentCode 超市付款码
func GetSPaymentCode(orderIDStr string) (sPaymentCodeList []SPaymentCode, count int) {

	orderIDS := strings.Split(orderIDStr, ",")
	orderCount := len(orderIDS)

	succCounter := struct {
		count int
		mutex sync.Mutex
	}{}
	if orderCount > 0 {
		ch := make(chan int64, len(orderIDS))
		for _, orderIDStr := range orderIDS {
			orderID, _ := tools.Str2Int64(orderIDStr)
			ch <- orderID
		}
		logs.Notice("[GetSPaymentCode] len :", len(ch))
		for {
			if len(ch) == 0 {
				logs.Debug("[GetSPaymentCode]for complete ,i'll quit")
				break
			}
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				if len(ch) == 0 {
					logs.Debug("[GetSPaymentCode]goruntine complete ,i'll quit")
					break
				}
				wg.Add(1)
				orderID := <-ch
				logs.Notice("[GetSPaymentCode] orderID:", orderID, " workID:", i)
				go func(orderID int64) {
					defer wg.Done()

					err, marketPayment, _ := xendit.MarketPaymentCodeGenerate(orderID, 0)
					if err != nil {
						return
					}
					sPaymentCode := SPaymentCode{}
					sPaymentCode.OrderID = tools.Int642Str(orderID)             //string  `json:"order_id"`     //订单ID
					sPaymentCode.PaymentCode = marketPayment.PaymentCode        //string  `json:"payment_code"` //付款码
					sPaymentCode.Amount = float64(marketPayment.ExpectedAmount) //float64 `json:"amount"`       //付款码金额
					//sPaymentCode.Status = marketPayment.Status           //string  `json:"status"`       //状态
					sPaymentCode.PayCompany = "Xendit"
					sPaymentCode.Ctime = marketPayment.Ctime
					sPaymentCode.ExpiryDate = marketPayment.ExpirationDate //int64  `json:"expiry_date"`  //过期时间
					succCounter.mutex.Lock()
					sPaymentCodeList = append(sPaymentCodeList, sPaymentCode)
					succCounter.mutex.Unlock()
					succCounter.count++
				}(orderID)
			}
			wg.Wait()
			count = succCounter.count
		}
	}
	return
}
