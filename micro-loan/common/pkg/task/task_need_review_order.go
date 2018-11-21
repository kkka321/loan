package task

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/pkg/schema_task"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/tongdun"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// HitRiskRegularItem 命中风控规则项 struct
type HitRiskRegularItem struct {
	Regular      string
	RejectReason types.RejectReasonEnum
	HitTime      int64
	Status       int         // 默认为0,可不赋值
	Value        interface{} //附带的值, 可无值
}

type ReviewOrderTask struct {
}

// TaskHandleNeedReviewOrder 处理待审核订单 {{{
func (c *ReviewOrderTask) Start() {
	logs.Info("[TaskHandleNeedReviewOrder] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("need_review_order_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[TaskHandleNeedReviewOrder] process is working, so, I will exit. err:%v", err)
		// ***! // 很重要!
		close(done)
		return
	}

	qName := beego.AppConfig.String("need_review_order")
	for {
		if cancelled() {
			logs.Info("[TaskHandleNeedReviewOrder] receive exit cmd.")
			break
		}

		TaskHeartBeat(storageClient, lockKey)

		// 1. 创建任务队列
		logs.Info("[TaskHandleNeedReviewOrder] produceNeedReviewOrderQueue")
		qValueByte, err := storageClient.Do("LLEN", qName)
		logs.Debug("qValueByte:", qValueByte, ", err:", err)
		if err == nil && qValueByte != nil && 0 == qValueByte.(int64) {

			// 队列是空,需要生成了
			// 1. 取数据
			orderList, _ := service.OrderListLoanStatus4Review()
			// 2. 加队列
			if len(orderList) == 0 {
				logs.Info("[TaskHandleNeedReviewOrder] 生产待审核订单队列没有满足条件的数据,休眠500毫秒后将重试.")
				time.Sleep(500 * time.Millisecond)
				continue
			}

			for _, orderOne := range orderList {
				// 写队列
				storageClient.Do("LPUSH", qName, orderOne.Id)
			}
		} else if err != nil {
			logs.Error("[TaskHandleNeedReviewOrder] LLEN error. err:%v", err)
		}

		// 2. 消费队列
		logs.Info("[TaskHandleNeedReviewOrder] consume queue")
		var wg sync.WaitGroup
		// 可视情况加工作 goroutine 数,一期只开2个
		// 开启4个协程
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go consumeNeedReviewOrderQueue(&wg, i)
		}

		// 主 goroutine,等待工作 goroutine 正常结束
		wg.Wait()
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[TaskHandleNeedReviewOrder] politeness exit.")
}

func (c *ReviewOrderTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("need_review_order_lock")
	storageClient.Do("DEL", lockKey)
}

// OfflineHandleRiskReview 离线跑风控,估计跑不了多久了...
func OfflineHandleRiskReview(accountID int64) []HitRiskRegularItem {
	orderData := models.Order{
		UserAccountId: accountID,
	}

	return handleRiskReview(orderData, 0)
}

func consumeNeedReviewOrderQueue(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	logs.Info("It will do consumeNeedReviewOrderQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	qName := beego.AppConfig.String("need_review_order")
	for {
		if cancelled() {
			logs.Info("[consumeNeedReviewOrderQueue] receive exit cmd, workID:", workerID)
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		if err != nil {
			logs.Error("[consumeNeedReviewOrderQueue] RPOP error workID:%d, err:%v", workerID, err)
		}

		// 没有可供消费的数据
		if qValueByte == nil {
			logs.Info("[consumeNeedReviewOrderQueue] no data for consume, so exit, workID:", workerID)
			break
		}

		orderID, _ := tools.Str2Int64(string(qValueByte.([]byte)))
		if orderID == types.TaskExitCmd {
			logs.Info("[consumeNeedReviewOrderQueue] receive exit cmd, I will exit after jobs done. workID:", workerID, ", orderID:", orderID)
			// ***! // 很重要!
			close(done)
			break
		}

		// 真正开始工作了
		addCurrentData(tools.Int642Str(orderID), "orderId")
		newHandleNeedReviewOrder(orderID, workerID)
		removeCurrentData(tools.Int642Str(orderID))
	}
}

// newHandleNeedReviewOrder (新)处理需要审核的订单
func newHandleNeedReviewOrder(orderID int64, workerID int) {
	// 记录日志
	logs.Info("[handleNeedReviewOrder] orderID:", orderID, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[newHandleNeedReviewOrder] panic orderID:%d, workerID:%d, err:%v", orderID, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	// 获取订单完整数据, 备用, 为值类型
	orderData, err := models.GetOrder(orderID)
	if err != nil || orderData.CheckStatus != types.LoanStatus4Review {
		orderDataJSON, _ := tools.JsonEncode(orderData)
		logs.Error("[handleNeedReviewOrder] 订单状态不正确,请检查. orderDataJSON:", orderDataJSON, ", workerID", workerID, ", err:", err)
		return
	}

	// 浅拷贝,简单值复制
	originOrder := orderData

	// 全量运行反欺诈规则, 并获取命中规则列表
	hitRiskRegularBox := handleRiskReview(orderData, workerID)

	// 根据命中规则和用户需要验证规则进行决策, 并将状态赋值给 orderData
	makeDecisionByRiskCtlRegular(&orderData, &hitRiskRegularBox)
	// 保存命中规则, 此时命中规则状态status已经被赋值是否审核
	saveHitRiskRegular(hitRiskRegularBox, orderID, orderData.UserAccountId)

	// 如果 反欺诈通过, 则查看是否触发熔断
	if orderData.CheckStatus == types.LoanStatusWaitManual {
		if isTriggerDayLoanOrdersNumLimit() {
			// 若触发熔断,则修改订单为反欺诈拒绝
			orderData.CheckStatus = types.LoanStatusReject
			orderData.RiskCtlStatus = types.RiskCtlAFReject
			// 此处修正值代表被熔断拒绝,
			// TODO 更好的处理, 比如CheckStatus 增加 熔断拒绝
			orderData.FixedRandom = service.FixedDayLimitOrdersReject
		}
	}

	monitor.IncrOrderCount(orderData.CheckStatus)

	if orderData.CheckStatus == types.LoanStatusWait4Loan {
		//第三方黑名单验证
		service.CheckThirdBlacklist(&orderData)
	}

	if orderData.CheckStatus == types.LoanStatusWait4Loan {
		//第三方黑名单通过后增加人脸比对
		service.CompareAfterBlackList(&orderData)
	}

	if orderData.CheckStatus == types.LoanStatusWait4Loan {
		monitor.IncrOrderCount(orderData.CheckStatus)
		schema_task.PushBusinessMsg(types.PushTargetReviewPass, orderData.UserAccountId)
	} else if orderData.CheckStatus == types.LoanStatusReject {
		schema_task.PushBusinessMsg(types.PushTargetReviewReject, orderData.UserAccountId)
	}

	makeRandomMark(&orderData)

	// 添加必更新字段
	orderData.CheckTime = tools.GetUnixMillis()
	orderData.Utime = orderData.CheckTime
	orderData.RiskCtlFinishTime = orderData.CheckTime
	models.UpdateOrder(&orderData)
	// 若订单更为等待电核状态, 则触发<工单创建>event
	logs.Debug("[newHandleNeedReviewOrder] start judge, whether need to trigger phone verify ticket, order id:", orderData.Id)
	if orderData.CheckStatus == types.LoanStatusWaitManual {
		logs.Debug("[newHandleNeedReviewOrder] trigger to create ticket, order id:", orderData.Id)
		ticket.CreateAfterRisk(orderData)
		// event.Trigger(&evtypes.TicketCreateEv{
		// 	Item:       types.TicketItemPhoneVerify,
		// 	CreateUID:  types.Robot,
		// 	RelatedID:  orderData.Id,
		// 	OrderID:    orderData.Id,
		// 	CustomerID: orderData.UserAccountId,
		// 	Data:       nil})
	}

	// 校验是否va全部生成
	service.CreateVirtualAccountAll(orderData.UserAccountId, orderData.Id)
	// 添加操作日志
	models.OpLogWrite(0, orderData.Id, models.OpCodeOrderUpdate, orderData.TableName(), originOrder, orderData)
}

// 保存命中规则
func saveHitRiskRegular(hitRiskRegularBox []HitRiskRegularItem, orderID, accountID int64) {
	if len(hitRiskRegularBox) > 0 {
		// 1. 记录所有命中的规则
		for _, item := range hitRiskRegularBox {
			oneRiskRegular := models.RiskRegularRecord{
				OrderId:    orderID,
				AccountId:  accountID,
				HitRegular: item.Regular,
				Status:     item.Status,
				Ctime:      item.HitTime,
			}
			models.AddOneRiskRegularRecord(oneRiskRegular)
		}
	}
}

func makeRandomMark(orderData *models.Order) {
	if service.IsLevel1Random(orderData.RandomValue) {
		orderData.RandomMark = 1
	} else if service.IsLevel2Random(orderData.RandomValue) {
		orderData.RandomMark = 2
	}
}

// 根据命中规则进行决策
// 显示声明 *[]HitRiskRegularItem 为引用类型的原因
// 1. 告诉调用者, 方法内部必然会会对变量产生变动
// 2. slice的引用指针, 再发生扩容时,会失效
func makeDecisionByRiskCtlRegular(orderData *models.Order, hitRiskRegularBoxRef *[]HitRiskRegularItem) {
	// 决策
	logs.Debug("Start Make Decision")

	if len(*hitRiskRegularBoxRef) <= 0 {
		// 若未命中反欺诈规则, 则反欺诈通过, 暂时适配任何情况
		orderData.CheckStatus = types.LoanStatusWaitManual
		orderData.RiskCtlStatus = types.RiskCtlWaitPhoneVerify

		processQuotaConf(orderData)
		return
	}

	accountBase, _ := models.OneAccountBaseByPkId(orderData.UserAccountId)
	// 获取需检查反欺诈规则命中列表
	hitNeedCheckRiskRegularBox, err := getNeedCheckHitRiskRegularList(accountBase, hitRiskRegularBoxRef)
	// 未配置风控列表, 让task空转,不修改此类订单状态 等待alert被发现.........
	// TODO 此处需要更好的报警策略, 如直接邮件提醒, 直接短信提醒
	if err != nil {
		return
	}
	if len(hitNeedCheckRiskRegularBox) > 0 {

		// 2. 随机数据等级策略
		orderData.CheckStatus = types.LoanStatusReject
		orderData.RiskCtlStatus = types.RiskCtlAFReject
		orderData.RiskCtlRegular = hitNeedCheckRiskRegularBox[0].Regular
		orderData.RejectReason = hitNeedCheckRiskRegularBox[0].RejectReason

		if orderData.IsReloan == 0 && accountBase.IsPlatformMark(types.PlatformMark_Gojek) {
			//pass
		} else if service.IsLevel1Random(orderData.RandomValue) {
			isHit1, fixValue := isHitRiskCtlLevel1(hitNeedCheckRiskRegularBox)

			// 一类随机数,反欺诈中了也让过,走电核
			orderData.CheckStatus = types.LoanStatusWaitManual
			orderData.RiskCtlStatus = types.RiskCtlWaitPhoneVerify

			// 如果命中一类反欺诈规则,则随机值失效
			if isHit1 {
				orderData.CheckStatus = types.LoanStatusReject
				orderData.RiskCtlStatus = types.RiskCtlAFReject
				orderData.RejectReason = types.RejectReasonHitBlackList // 借款订单命中C开头的反欺诈规则时，拒绝原因展示：命中黑名单
				orderData.FixedRandom = fixValue
			}
		} else if service.IsLevel2Random(orderData.RandomValue) {
			isHit2, fixValue := isHitRiskCtlLevel2(hitNeedCheckRiskRegularBox)

			// 二类随机值,默认为等特电核
			orderData.CheckStatus = types.LoanStatusWaitManual
			orderData.RiskCtlStatus = types.RiskCtlWaitPhoneVerify

			// 如果命中二类反欺诈规则,则随机值失效
			if isHit2 {
				orderData.CheckStatus = types.LoanStatusReject
				orderData.RiskCtlStatus = types.RiskCtlAFReject
				orderData.FixedRandom = fixValue
			}
		}

		// 仅因为Z002被风控拒绝的订单需要 打标签
		if orderData.RiskCtlStatus == types.RiskCtlAFReject {
			addTagCustomerRecall(hitNeedCheckRiskRegularBox, orderData)
		} else if orderData.RiskCtlStatus == types.RiskCtlWaitPhoneVerify || orderData.CheckStatus == types.LoanStatusWait4Loan {
			tryCancleCustomerTagScore(orderData)
		}
	} else {
		orderData.CheckStatus = types.LoanStatusWaitManual
		orderData.RiskCtlStatus = types.RiskCtlWaitPhoneVerify

		processQuotaConf(orderData)
	}
}

// 添加召回标签 评分模型需要召回. 目前包括:仅因为Z002被反欺诈拒的用户
// 传入参数须为 真正生效的规则名字。
func addTagCustomerRecall(hitRiskRegularItemInReview []HitRiskRegularItem, order *models.Order) {

	// 拒绝列表仅有一个 Z002 的用户需要打标签
	if len(hitRiskRegularItemInReview) == 1 &&
		hitRiskRegularItemInReview[0].Regular == types.RegularNameZ002 {

		if score, ok := hitRiskRegularItemInReview[0].Value.(int); ok {
			// 是否需要打标签 M N 配置
			if service.CanAddCustumetRecallScore(score, order) {
				err := service.ChangeCustomerRecall(order.UserAccountId, order.Id, types.RecallTagScore, types.RemarkTagNone)
				if err != nil {
					logs.Error("[addTagCustomerRecall] Z002 ChangeCustomerRecall accountId:%d, err:%v  hitItem:%#v", order.UserAccountId, err, hitRiskRegularItemInReview[0])
				}
			}
		} else {
			logs.Error("[addTagCustomerRecall] Z002_score   assert error. hitItem:%#v ", hitRiskRegularItemInReview[0])
		}
	}
}

func tryCancleCustomerTagScore(order *models.Order) {
	accountExt, _ := models.OneAccountBaseExtByPkId(order.UserAccountId)
	if accountExt.AccountId == 0 ||
		accountExt.RecallTag == types.RecallTagNone {
		logs.Info("[tryCancleCustomerTagScore] no need to cancle tag. accountExt:%#v orderId:%d", accountExt, order.Id)
		return
	}

	err := service.ChangeCustomerRecall(order.UserAccountId, order.Id, types.RecallTagNone, types.RemarkTagNone)
	if err != nil {
		logs.Error("[tryCancleCustomerTagScore] Z002 ChangeCustomerRecall accountId:%d, err:%v  ", order.UserAccountId, err)
	}
	return
}

func processQuotaConf(orderData *models.Order) {
	//拉取风控配置的账号额度账期
	accountQuotaConf, err := dao.GetLastAccountQuotaConf(orderData.UserAccountId)
	if err != nil {
		logs.Error("[processQuotaConf] GetLastAccountQuotaConf hanppend err:", err, "AccountID:", orderData.UserAccountId, "orderID:", orderData.Id)
	}
	if orderData.IsReloan == 0 {
		return
	}
	similarVal, compareType := service.SaveLoanIDHeadAndLivingEnvCompare(orderData.UserAccountId, orderData.Id)
	var configSimilarVal float64
	if compareType == "firstenv_reloanenv_similar" {
		configSimilarVal, _ = config.ValidItemFloat64("firstenv_reloanenv_similar")
	}
	if compareType == "first_idhand_reloanenv_similar" {
		configSimilarVal, _ = config.ValidItemFloat64("first_idhand_reloanenv_similar")
	}
	if configSimilarVal == 0 {
		logs.Debug("[processQuotaConf] configSimilarVal 阈值配置：", configSimilarVal)
		configSimilarVal = types.LivingBestAndReloanHandholdSimilar
	}
	orderData.LivingbestReloanhandSimilar = tools.Float642Str(similarVal)
	logs.Debug("[processQuotaConf] SaveLoanIDHeadAndLivingEnvCompare orderID:%d, accountID:%d, similarVal:%g, compareType:%s,configSimilarVal:%g", orderData.Id, orderData.UserAccountId, similarVal, compareType, configSimilarVal)

	//如果风控账号配置不需要电核并且订单未命中一二级随机数并且活体最佳与复贷手持比对结果大于阈值
	//直接跳过电核等待放款
	if accountQuotaConf.IsPhoneVerify == 0 && similarVal >= configSimilarVal {
		monitor.IncrOrderCount(orderData.CheckStatus)
		orderData.CheckStatus = types.LoanStatusWait4Loan
		orderData.RiskCtlStatus = types.RiskCtlPhoneVerifyPass
	} else {
		if accountQuotaConf.IsPhoneVerify == 0 {
			logs.Error("[processQuotaConf] similarVal:", similarVal, "orderID:", orderData.Id, "accountID:", orderData.UserAccountId)
		}
	}
	logs.Debug("[processQuotaConf]riskQuotaConf isPhoneVerify:", accountQuotaConf.IsPhoneVerify, " similarVal:", similarVal)
}

func getNeedCheckHitRiskRegularList(accountBase models.AccountBase, hitRiskRegularBoxRef *[]HitRiskRegularItem) (hitRiskRegularItemInReview []HitRiskRegularItem, err error) {
	var checkMap map[string]bool
	isRandomMarkAccount := service.IsRandomMarkAccountByAccountBase(&accountBase)
	//isOverdueAccount := dao.IsOverdueAccount(accountBase.Id)

	if dao.IsRepeatLoan(accountBase.Id) {
		// 复贷
		if isRandomMarkAccount {

			//2018年11月06日19:49:39 去掉此不过风控的逻辑
			// 复贷-随机数[用户]
			//if !isOverdueAccount {
			//	// 复贷-随机数-无逾期[用户]
			//	// 跳过反欺诈,直接进入电核环节
			//	logs.Debug("[func:getNeedCheckHitRiskRegularList]复贷-随机数-无逾期, 无反欺诈审查项,直接通过, accountID: %d", accountBase.Id)
			//	return
			//}
			// 复贷-随机数-有逾期
			logs.Debug("[func:getNeedCheckHitRiskRegularList]复贷-随机数-有逾期, 获取待审核反欺诈项, accountID: %d", accountBase.Id)
			checkMap, err = getRiskMapByRegularName(types.ReloanWithRandomMarkRiskRegularList)
		} else {
			// 复贷-非随机数[用户]
			// 针对复贷的借款订单，如客户非测试客户，跳过D009的反欺诈规则 。其余环节与首贷流程一致
			// 意味着此规则列表中, 应该不包含D009
			logs.Debug("[func:getNeedCheckHitRiskRegularList]复贷-非随机数, 获取待审核反欺诈项, accountID: %d", accountBase.Id)
			checkMap, err = getRiskMapByRegularName(types.ReloanWithoutRandomMarkRiskRegularList)
		}
	} else if accountBase.IsPlatformMark(types.PlatformMark_Gojek) {
		checkMap, err = getRiskMapByRegularName(types.LoanGojekRiskRegularList)
	} else {
		// 首贷
		logs.Debug("[func:getNeedCheckHitRiskRegularList]首贷, 获取待审核反欺诈项, accountID: %d", accountBase.Id)
		checkMap, err = getRiskMapByRegularName(types.FirstLoanRiskRegularList)
	}

	// status 状态
	for i, r := range *hitRiskRegularBoxRef {
		if _, ok := checkMap[r.Regular]; ok {
			(*hitRiskRegularBoxRef)[i].Status = types.RiskCtlRegularReviewed
			hitRiskRegularItemInReview = append(hitRiskRegularItemInReview, (*hitRiskRegularBoxRef)[i])
		}
	}
	logs.Debug("[func:getNeedCheckHitRiskRegularList] Hit and Reviewed: %v; accountID: %d", hitRiskRegularItemInReview, accountBase.Id)
	logs.Debug("[func:getNeedCheckHitRiskRegularList] Run and Hit: %v; accountID: %d", *hitRiskRegularBoxRef, accountBase.Id)
	return
}

// getRiskMapByRegularName 获取指定类型用户的风控检查列表
// 此处返回 err 为恶性 err, 风控列表配置为空,很可能是后台无配置,或者错误操作引起
// 检查到此处err, 应该停止此订单检查, 跳过,或者置特殊状态
func getRiskMapByRegularName(listName types.RiskRegularListName) (m map[string]bool, err error) {
	m = make(map[string]bool)

	originConf := config.ValidItemString(string(listName))
	logs.Debug("[func:getRiskMapByRegularName]获取待审核反欺诈项: %s", originConf)

	if len(originConf) <= 0 {
		// 此配置不应该为空, 为空会引起恶劣影响
		err = fmt.Errorf("[RiskRegular] required config, check please, config name: %s ;此错误会造成任务空转,订单状态不改变,及特定类别待审核订单暴涨", listName)
		logs.Alert("err:", err)
		return
	}

	regularList := strings.Split(originConf, ",")
	for _, regular := range regularList {
		m[strings.Trim(regular, " ")] = true
	}
	return
}

// isTriggerDayLoanOrdersNumLimit 是否触发,日贷款订单上限
// 感觉此处日订单熔断应该放在, 放款前的最后一关(此时是电核)上更好
// 熔断状态, 应该给类似 99 这样的特殊状态, 以增加可扩展性
func isTriggerDayLoanOrdersNumLimit() bool {
	dayLoanOrderLimit, err := config.ValidItemInt64("day_loan_order_limit")
	if err != nil {
		logs.Error("day_loan_order_limit SystemConfigValidItemInt64 Error:", err)
	}
	loanTotal, _ := service.GetTodayLoanOrderTotal()
	if loanTotal >= dayLoanOrderLimit {
		return true
	}
	return false
}

// TODO: 此方法巨长无比,明显不符合软件工程思想,优化...
// 已拆出取订单,日熔断和订单更新逻辑约30行....
// 已拆出决策逻辑
// 剩余逻辑: 取待审核数据(用户基本数据,profile, 大数据等), 风控规则审核
func handleRiskReview(orderData models.Order, workerID int) []HitRiskRegularItem {
	orderID := orderData.Id
	var hitRiskRegularBox []HitRiskRegularItem
	accountID := orderData.UserAccountId
	accountBase, _ := models.OneAccountBaseByPkId(accountID)
	isReloan := dao.IsRepeatLoan(accountID)
	/** A001 身份证在内部白名单内 */

	multiResp := service.AdvanceMultiPlatform(orderData.Id, orderData.UserAccountId, accountBase.Identity)

	// -----Bxxx-----
	/** B001 客户年龄不符合要求 */
	age, _ := service.CustomerAge(accountBase.Identity)
	if age < types.LimitAge {
		logs.Warn("[handleNeedReviewOrder] 客户年龄不符合要求, age: %d, orderID: %d, accountID: %d, age: %d, workerID: %d",
			age, orderID, accountBase.Id, age, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "B001",
			RejectReason: types.RejectReasonAge,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	accountProfile, _ := dao.CustomerProfile(orderData.UserAccountId)
	var clientInfo models.ClientInfo
	var err error
	if orderID > 0 {
		clientInfo, err = service.OrderClientInfo(orderID)
	} else {
		clientInfo, err = service.LastClientInfo(orderData.UserAccountId)
	}

	/** B002 GPS 所在区域不符 */

	/** B003 客户当前贷款已逾期 */

	/** B004 客户近3月贷款最大逾期天数>=15天 */

	///** B005 联系人1与联系人2手机号重复 */

	// B005 身份信息检查结果返回为“未知”
	//身份检查V3更新，先判断同盾，再判断acvance

	identityVerify := service.IdentityVerify(accountID)
	if !identityVerify {
		hitItem := HitRiskRegularItem{
			Regular:      "B005",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	///** B006 工作地与居住地不在同一城市 */

	unservicedAreaConf, _ := service.GetUnservicedAreaConf()

	/** B008：身份证归属地所在区域不符 */
	if unservicedAreaConf[accountBase.ThirdProvince] {
		logs.Warn("[handleNeedReviewOrder] [B008] 身份证归属地所在区域不符, ThirdProvince: %s, orderID: %d, accountID: %d, workerID: %d",
			accountBase.ThirdProvince, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "B008",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/*
		    // B009：居住地址所在区域不符
			residentProvince, err := accountProfile.ResidentProvince()
			if err != nil || unservicedAreaConf[residentProvince] {
				logs.Warn("[handleNeedReviewOrder] [B009] 居住地址所在区域不符, ResidentCity: [%s], orderID: %d, accountID: %d, workerID: %d, err: %#v",
					accountProfile.ResidentCity, orderID, accountBase.Id, workerID, err)

				hitItem := HitRiskRegularItem{
					Regular:      "B009",
					RejectReason: types.RejectReasonLackCredit,
					HitTime:      tools.GetUnixMillis(),
				}
				hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
			}
	*/

	/** B010：单位地址所在区域不符 */
	companyProvince, err := accountProfile.CompanyProvince()
	if companyProvince != "" && unservicedAreaConf[companyProvince] {
		logs.Warn("[handleNeedReviewOrder] [B010] 单位地址所在区域不符, CompanyCity: [%s], orderID: %d, accountID: %d, workerID: %d, err: %#v",
			accountProfile.CompanyCity, orderID, accountBase.Id, workerID, err)

		hitItem := HitRiskRegularItem{
			Regular:      "B010",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	//! 以下三段代码是不是可以用反射来简化呢?留给你来解决了^_*
	// -----Cxxx-----

	/** C001 身份证号在内部黑名单内 */
	yes, _ := models.IsBlacklistIdentity(accountBase.Identity)
	if yes {
		logs.Warn("[handleNeedReviewOrder] [C001] 身份证号在内部黑名单内, Identity: %s, orderID: %d, accountID: %d, workerID: %d",
			accountBase.Identity, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "C001",
			RejectReason: types.RejectReasonHitBlackList,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/** C002 手机号在内部黑名单内 */
	yes, _ = models.IsBlacklistMobile(accountBase.Mobile)
	if yes {
		logs.Warn("[handleNeedReviewOrder] [C002] 手机号在内部黑名单内, Mobile: %s, orderID: %d, accountID: %d, workerID: %d",
			accountBase.Mobile, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "C002",
			RejectReason: types.RejectReasonHitBlackList,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/** C003 第一联系人在内部黑名单内 */
	yes, _ = models.IsBlacklistMobile(accountProfile.Contact1)
	if yes {
		logs.Warn("[handleNeedReviewOrder] [C003] 第一联系人在内部黑名单内, Contact1: %s",
			accountProfile.Contact1, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "C003",
			RejectReason: types.RejectReasonHitBlackList,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/** C004 居住地址在内部黑名单内 */
	address := accountProfile.ResidentCity + "," + accountProfile.ResidentAddress
	yes, _ = models.IsBlacklistItem(types.RiskItemResidentAddress, address)
	if yes {
		logs.Warn("[handleNeedReviewOrder] [C004] RiskItemResidentAddress, ResidentAddress: %s, orderID: %d, accountID: %d, workerID: %d",
			address, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "C004",
			RejectReason: types.RejectReasonHitBlackList,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/** C005 单位名称在内部黑名单内 */
	if accountProfile.CompanyName != "" {
		yes, _ = models.IsBlacklistItem(types.RiskItemCompany, accountProfile.CompanyName)
		if yes {
			logs.Warn("[handleNeedReviewOrder] [C005] RiskItemCompany, CompanyName: %s, orderID: %d, accountID: %d, workerID: %d",
				accountProfile.CompanyName, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "C005",
				RejectReason: types.RejectReasonHitBlackList,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	}

	/** C006 单位地址在内部黑名单内 */
	companyAddress := accountProfile.CompanyCity + "," + accountProfile.CompanyAddress
	yes, _ = models.IsBlacklistItem(types.RiskItemCompanyAddress, companyAddress)
	if yes {
		logs.Warn("[handleNeedReviewOrder] [C006] RiskItemCompanyAddress, CompanyAddress: %s, orderID: %d, accountID: %d, workerID: %d",
			companyAddress, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "C006",
			RejectReason: types.RejectReasonHitBlackList,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/** C007 设备号在内部黑名单内 */
	yes, _ = models.IsBlacklistItem(types.RiskItemIMEI, clientInfo.Imei)
	if yes {
		logs.Warn("[handleNeedReviewOrder] [C007] RiskItemIMEI, IMEI: %s, orderID: %d, accountID: %d, workerID: %d",
			clientInfo.Imei, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "C007",
			RejectReason: types.RejectReasonHitBlackList,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/** C008 IP在内部黑名单内 */
	yes, _ = models.IsBlacklistItem(types.RiskItemIP, clientInfo.IP)
	if yes {
		logs.Warn("[handleNeedReviewOrder] [C008] RiskItemIP, IP: %s, orderID: %d, accountID: %d, workerID: %d",
			clientInfo.IP, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "C008",
			RejectReason: types.RejectReasonHitBlackList,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/** 大数据服务 */
	// -----Dxxx-----

	/** D000 未上报设备信息 */
	if len(clientInfo.Imei) <= 0 {
		logs.Warn("[handleNeedReviewOrder] [D000] 未抓取到设备信息IMEI, orderID: %d, accountID: %d, workerID: %d",
			orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "D000",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	imeiMd5 := tools.Md5(clientInfo.Imei)
	esRes, router, rawData, _ := service.EsSearchById(imeiMd5)
	/** D001 未抓取到设备信息 */
	if !esRes.Found || !esRes.IsAll() {
		logs.Warn("[handleNeedReviewOrder] [D001] 未抓取到设备信息, orderID: %d, accountID: %d, workerID: %d",
			orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "D001",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	} else {
		// 有大数据时才跑
		const esDataNo = 1

		dao.SaveEsData(orderID, accountBase.Id, router, string(rawData))

		/** D002 未抓取到通话记录 */
		if esRes.Source.NotObtainedCallRecord == esDataNo {
			logs.Warn("[handleNeedReviewOrder] [D002] 未抓取到通话记录, NotObtainedCallRecord: %d, orderID: %d, accountID: %d, workerID: %d",
				esRes.Source.NotObtainedCallRecord, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D002",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D003 未抓取到通讯录 */
		if esRes.Source.NotObtainedAddressList == esDataNo {
			logs.Warn("[handleNeedReviewOrder] [D003] 未抓取到通讯录, orderID: %d, accountID: %d, workerID: %d",
				orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D003",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D004 未抓取到短信记录 */
		if esRes.Source.NotObtainedMessage == esDataNo {
			logs.Warn("[handleNeedReviewOrder] [D004] 未抓取到短信记录, orderID: %d, accountID: %d, workerID: %d",
				orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D004",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D005 未抓取到GPS信息 */
		if esRes.Source.NotObtainedGpsInfo == esDataNo {
			logs.Warn("[handleNeedReviewOrder] [D005] 未抓取到GPS信息, orderID: %d, accountID: %d, workerID: %d",
				orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D005",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		///** D006 近1个月设备号登录的手机账号>=3 */

		///** D007 近1个月手机号登录的设备号>=3 */

		///** D008 3个月内00:00——4:00通话时长>=60m */

		// 针对复贷的借款订单，如客户非测试客户，跳过D009的反欺诈规则 。其余环节与首贷流程一致
		// 此特殊情况,已移动至,决策的条件过滤中
		/** D009 3个月内无通话记录天数 >= 45天 */
		riskCtlD009, _ := config.ValidItemInt("risk_ctl_D009")
		if esRes.Source.NoCallRecordDays == 0 || esRes.Source.NoCallRecordDays >= riskCtlD009 {
			logs.Warn("[handleNeedReviewOrder] [D009] 3个月内无通话记录天数 >= %d 天, NoCallRecordDays: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD009, esRes.Source.NoCallRecordDays, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D009",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		///** D010 3个月内与第一联系人通话次数<=3 */

		///** D011 3个月内与第一联系人通话时长<=10m */

		/** D012 通讯录个数<=30 */
		riskCtlD012, _ := config.ValidItemInt("risk_ctl_D012")
		if esRes.Source.NumberOfContacts <= riskCtlD012 {
			logs.Warn("[handleNeedReviewOrder] [D012] 通讯录个数 <= %d, NumberOfContacts: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD012, esRes.Source.NumberOfContacts, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D012",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		///** D013 通讯录中固定电话占比>=50% */

		///** D014 通话记录名单占通讯录的比例<=40% */

		/** D015 包含以下关键字“逾期”“贷款”的短信数量>=5 */
		riskCtlD015, _ := config.ValidItemInt("risk_ctl_D015")
		if esRes.Source.NumberOfMessagesContainKeyword >= riskCtlD015 {
			logs.Warn("[handleNeedReviewOrder] [D015] 包含以下关键字“逾期”“贷款”的短信数量 >= %d. NumberOfMessagesContainKeyword: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD015, esRes.Source.NumberOfMessagesContainKeyword, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D015",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** ReD015 包含以下关键字“逾期”“贷款”的短信数量>=5 */
		riskCtlReD015, _ := config.ValidItemInt("risk_ctl_ReD015")
		if esRes.Source.NumberOfMessagesContainKeyword >= riskCtlReD015 {
			logs.Warn("[handleNeedReviewOrder] [ReD015] 包含以下关键字“逾期”“贷款”的短信数量 >= %d. NumberOfMessagesContainKeyword: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD015, esRes.Source.NumberOfMessagesContainKeyword, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "ReD015",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D017： 1小时内，设备移动距离≥1000Km */
		riskCtlD017, _ := config.ValidItemFloat64("risk_ctl_D017")
		if esRes.Source.DistanceOfDevice >= riskCtlD017 {
			logs.Warn("[handleNeedReviewOrder] [D017] 设备移动距离 >= %d. DistanceOfDevice: %f, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD017, esRes.Source.DistanceOfDevice, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D017",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D018：1天内，同一设备注册时间间隔<30s */
		riskCtlD018, _ := config.ValidItemInt("risk_ctl_D018")
		// 先进行 > 0 判断, 防止未取到大数据字段, 而被go int 默认置为0
		logs.Debug("D018 start, 规则: 0 < 获取大数据值: %d < 限定配置值: %d", esRes.Source.TimesOfDeviceRegistered, riskCtlD018)
		if esRes.Source.TimesOfDeviceRegistered > 0 && esRes.Source.TimesOfDeviceRegistered < riskCtlD018 {
			logs.Warn("[handleNeedReviewOrder] [D018] 同一设备最小注册时间间隔 < %d. TimesOfDeviceRegistered: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD018, esRes.Source.TimesOfDeviceRegistered, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D018",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D026：3个月内呼入与呼出前10的重叠个数≤1 */
		riskCtlD026, _ := config.ValidItemInt("risk_ctl_D026")
		if esRes.Source.SameNumberInOut3Months <= riskCtlD026 {
			logs.Warn("[handleNeedReviewOrder] [D026] 3个月内呼入与呼出前10的重叠个数 SameNumberInOut3Months: %d ≤ %d, orderID: %d, accountID: %d, workerID: %d",
				esRes.Source.SameNumberInOut3Months, riskCtlD026, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D026",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** ReD026：3个月内呼入与呼出前10的重叠个数≤1 */
		riskCtlReD026, _ := config.ValidItemInt("risk_ctl_ReD026")
		if esRes.Source.SameNumberInOut3Months <= riskCtlReD026 {
			logs.Warn("[handleNeedReviewOrder] [ReD026] 3个月内呼入与呼出前10的重叠个数 SameNumberInOut3Months: %d ≤ %d, orderID: %d, accountID: %d, workerID: %d",
				esRes.Source.SameNumberInOut3Months, riskCtlReD026, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "ReD026",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D031：最近通话时间距现在天数>7 */
		riskCtlD031, _ := config.ValidItemInt("risk_ctl_D031")
		if esRes.Source.LastCallDays > riskCtlD031 {
			logs.Warn("[handleNeedReviewOrder] [D031] LastCallDays: %d > %d, orderID: %d, accountID: %d, workerID: %d",
				esRes.Source.LastCallDays, riskCtlD031, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D031",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** ReD031：最近通话时间距现在天数>7 */
		riskCtlReD031, _ := config.ValidItemInt("risk_ctl_ReD031")
		if esRes.Source.LastCallDays > riskCtlReD031 {
			logs.Warn("[handleNeedReviewOrder] [ReD031] LastCallDays: %d > %d, orderID: %d, accountID: %d, workerID: %d",
				esRes.Source.LastCallDays, riskCtlReD031, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "ReD031",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D032：3个月内无短信记录天数>=45 */
		riskCtlD032, _ := config.ValidItemInt("risk_ctl_D032")
		if esRes.Source.NoSmsRecordDays >= riskCtlD032 {
			logs.Warn("[handleNeedReviewOrder] [D032] NoSmsRecordDays: %d ≥ %d, orderID: %d, accountID: %d, workerID: %d",
				esRes.Source.NoSmsRecordDays, riskCtlD032, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D032",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D033：最近发短信时间距离现在天数>30 */
		riskCtlD033, _ := config.ValidItemInt("risk_ctl_D033")
		if esRes.Source.LastSmsDays > riskCtlD033 {
			logs.Warn("[handleNeedReviewOrder] [D033] LastSmsDays: %d > %d, orderID: %d, accountID: %d, workerID: %d",
				esRes.Source.LastSmsDays, riskCtlD033, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D033",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D037：匹配逾期短信数量 >5 */
		riskCtlD037, _ := config.ValidItemInt("risk_ctl_D037")
		if esRes.Source.PhoneOverdueSmsNum > riskCtlD037 {
			logs.Warn("[handleNeedReviewOrder] [D037] PhoneOverdueSmsNum: %d > %d, orderID: %d, accountID: %d, workerID: %d",
				esRes.Source.PhoneOverdueSmsNum, riskCtlD037, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D037",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D038：匹配逾期短信(排除keterlambatan)数量 >3 */
		riskCtlD038, _ := config.ValidItemInt("risk_ctl_D038")
		if esRes.Source.PhoneOverdueOneSmsNum > riskCtlD038 {
			logs.Warn("[handleNeedReviewOrder] [D038] PhoneOverdueOneSmsNum: %d > %d, orderID: %d, accountID: %d, workerID: %d",
				esRes.Source.PhoneOverdueOneSmsNum, riskCtlD038, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D038",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D039：匹配逾期短信(排除keterlambatan_jika_terlambat)数量 >1 */
		riskCtlD039, _ := config.ValidItemInt("risk_ctl_D039")
		if esRes.Source.PhoneOverdueTwoSmsNum > riskCtlD039 {
			logs.Warn("[handleNeedReviewOrder] [D039] PhoneOverdueTwoSmsNum: %d > %d, orderID: %d, accountID: %d, workerID: %d",
				esRes.Source.PhoneOverdueTwoSmsNum, riskCtlD033, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D039",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D040：匹配keterlambatan_jika_terlambat短信距今天数 >5 */
		riskCtlD040, _ := config.ValidItemInt("risk_ctl_D040")
		if esRes.Source.PhoneTuiguangSmsDayDiff > riskCtlD040 {
			logs.Warn("[handleNeedReviewOrder] [D040] PhoneTuiguangSmsDayDiff: %d > %d, orderID: %d, accountID: %d, workerID: %d",
				esRes.Source.PhoneTuiguangSmsDayDiff, riskCtlD040, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D040",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	}

	fraudReq := service.FraudRequestInfo{}
	service.FillFantasyFraudRequest(&fraudReq, &orderData, &accountBase, &clientInfo)
	rawData, router, fraudRes, err := service.GetFantasyFraud(fraudReq)
	if err != nil || !fraudRes.IsSuccess() {
		logs.Warn("[handleNeedReviewOrder] [D999] 大数据没有抓取到账户信息, orderID:%d, accountID:%d, workerID:%d, data:%s",
			orderID, accountBase.Id, workerID, string(rawData))

		hitItem := HitRiskRegularItem{
			Regular:      "D999",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	} else {
		dao.SaveEsData(orderID, accountBase.Id, router, string(rawData))

		/** D016： 1小时内，账户移动距离≥1000Km */
		riskCtlD016, _ := config.ValidItemFloat64("risk_ctl_D016")
		if fraudRes.Data.DistanceOfAccount >= riskCtlD016 {
			logs.Warn("[handleNeedReviewOrder] [D016] 1小时内，账户移动距离 ≥ %d. DistanceOfAccount: %f, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD016, fraudRes.Data.DistanceOfAccount, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D016",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D019：1天内，同一设备注册的账号数≥3 */
		riskCtlD019, _ := config.ValidItemInt("risk_ctl_D019")
		if fraudRes.Data.AccountRegisteredDevice >= riskCtlD019 {
			logs.Warn("[handleNeedReviewOrder] [D019] 1天内，同一设备注册的账号数 ≥ %d. AccountRegisteredDevice: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD019, fraudRes.Data.AccountRegisteredDevice, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D019",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D020：1天内，同一设备登录账户号≥3 */
		riskCtlD020, _ := config.ValidItemInt("risk_ctl_D020")
		if fraudRes.Data.AccountSameDeviceLoginedOneday >= riskCtlD020 {
			logs.Warn("[handleNeedReviewOrder] [D020] 1天内，同一设备注册的账号数 ≥ %d. AccountSameDeviceLoginedOneday: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD020, fraudRes.Data.AccountSameDeviceLoginedOneday, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D020",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D021：历史同一设备号登录账号≥5 */
		riskCtlD021, _ := config.ValidItemInt("risk_ctl_D021")
		if fraudRes.Data.AccountSameDeviceLoginedHistory >= riskCtlD021 {
			logs.Warn("[handleNeedReviewOrder] [D021] 历史同一设备号登录账号 ≥ %d. AccountSameDeviceLoginedHistory: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD021, fraudRes.Data.AccountSameDeviceLoginedHistory, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D021",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D022：1天内，同一账号登录的设备数≥3 */
		riskCtlD022, _ := config.ValidItemInt("risk_ctl_D022")
		if fraudRes.Data.DeviceSameAccountLoginedOneday >= riskCtlD022 {
			logs.Warn("[handleNeedReviewOrder] [D022] 1天内，同一账号登录的设备数 ≥ %d. DeviceSameAccountLoginedOneday: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD022, fraudRes.Data.DeviceSameAccountLoginedOneday, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D022",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D023：历史同一账号登录的设备数≥5 */
		riskCtlD023, _ := config.ValidItemInt("risk_ctl_D023")
		if fraudRes.Data.DeviceSameAccountLoginedHistory >= riskCtlD023 {
			logs.Warn("[handleNeedReviewOrder] [D023] 历史同一账号登录的设备数 ≥ %d. DeviceSameAccountLoginedHistory: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD023, fraudRes.Data.DeviceSameAccountLoginedHistory, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D023",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D024：同一IP，1小时内注册设备数≥10 */
		riskCtlD024, _ := config.ValidItemInt("risk_ctl_D024")
		if fraudRes.Data.DeviceSameIpRegistered >= riskCtlD024 {
			logs.Warn("[handleNeedReviewOrder] [D024] 同一IP，1小时内注册设备数 ≥ %d. DeviceSameIPRegistered: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD024, fraudRes.Data.DeviceSameIpRegistered, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D024",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D025：同一IP，1小时内注册账号数≥10 */
		riskCtlD025, _ := config.ValidItemInt("risk_ctl_D025")
		if fraudRes.Data.AccountsSameIpRegistered >= riskCtlD025 {
			logs.Warn("[handleNeedReviewOrder] [D025] 同一IP，1小时内注册账号数 ≥ %d. AccountsSameIPRegistered: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlD025, fraudRes.Data.AccountsSameIpRegistered, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D025",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D027：7天内，同一设备注册的账号数≥5 */
		riskCtlD027, _ := config.ValidItemInt("risk_ctl_D027")
		if fraudRes.Data.AccountSameDeviceRegistered7Days >= riskCtlD027 {
			logs.Warn("[handleNeedReviewOrder] [D027] 7天内，同一设备注册的账号数 AccountSameDeviceRegistered7Days: %d ≥ %d, orderID: %d, accountID: %d, workerID: %d",
				fraudRes.Data.AccountSameDeviceRegistered7Days, riskCtlD027, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D027",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D028：30天内，同一设备注册的账号数≥10 */
		riskCtlD028, _ := config.ValidItemInt("risk_ctl_D028")
		if fraudRes.Data.AccountSameDeviceRegistered30Days >= riskCtlD028 {
			logs.Warn("[handleNeedReviewOrder] [D028] 30天内，同一设备注册的账号数 AccountSameDeviceRegistered30Days: %d ≥ %d, orderID: %d, accountID: %d, workerID: %d",
				fraudRes.Data.AccountSameDeviceRegistered30Days, riskCtlD028, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D028",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D029：7天内，同一设备登录账户号≥5 */
		riskCtlD029, _ := config.ValidItemInt("risk_ctl_D029")
		if fraudRes.Data.AccountSameDeviceLogined7Days >= riskCtlD029 {
			logs.Warn("[handleNeedReviewOrder] [D029] 7天内，同一设备登录账户号 AccountSameDeviceLogined7Days: %d ≥ %d, orderID: %d, accountID: %d, workerID: %d",
				fraudRes.Data.AccountSameDeviceLogined7Days, riskCtlD029, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D029",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D030：30天内，同一设备登录账户号≥10 */
		riskCtlD030, _ := config.ValidItemInt("risk_ctl_D030")
		if fraudRes.Data.AccountSameDeviceLogined30Days >= riskCtlD030 {
			logs.Warn("[handleNeedReviewOrder] [D030] 30天内，同一设备登录账户号 AccountSameDeviceLogined30Days: %d ≥ %d, orderID: %d, accountID: %d, workerID: %d",
				fraudRes.Data.AccountSameDeviceLogined30Days, riskCtlD030, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D030",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

	}

	// -----Exxx-----

	/** E001 近1个月内第一联系人存在拒贷借款订单 */
	riskCtlE001, _ := config.ValidItemInt("risk_ctl_E001")
	yes, _, err = service.ContactHasRejectLoanOderInDays(accountProfile.Contact1, int64(riskCtlE001))
	if yes {
		logs.Warn("[handleNeedReviewOrder] [E001] 近 %d 天内第一联系人存在拒贷借款订单. orderID: %d, accountID: %d, workerID: %d", riskCtlE001, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "E001",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	riskCtlReE001, _ := config.ValidItemInt("risk_ctl_ReE001")
	yes, _, err = service.ContactHasRejectLoanOderInDays(accountProfile.Contact1, int64(riskCtlReE001))
	if yes {
		logs.Warn("[handleNeedReviewOrder] [ReE001] 近 %d 天内第一联系人存在拒贷借款订单. orderID: %d, accountID: %d, workerID: %d", riskCtlReE001, orderID, accountBase.Id, workerID)

		hitItem1 := HitRiskRegularItem{
			Regular:      "ReE001",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem1)
	}

	/** E002 第一联系人存在逾期中的借款订单 */
	yes, _, err = service.ContactHasOverdueLoanOrder(accountProfile.Contact1)
	if yes {
		logs.Warn("[handleNeedReviewOrder] [E002] 第一联系人存在逾期中的借款订单. orderID: %d, accountID: %d, workerID: %d", orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "E002",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	if accountProfile.CompanyName != "" {
		/** E003 近1个月内同一单位的申请人数>=10 */
		riskCtlE003, _ := config.ValidItemInt64("risk_ctl_E003")
		count, _ := service.SameCompanyApplyLoanOrderInLastMonth(accountProfile.CompanyName)
		if count >= riskCtlE003 {
			logs.Warn("[handleNeedReviewOrder] [E003] 近1个月内同一单位的申请人数 ≥ %d. count: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlE003, count, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "E003",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	}

	/*
		// E004 近1个月内同一居住地址的申请人数>=5
		riskCtlE004, _ := config.ValidItemInt64("risk_ctl_E004")
		count, err = service.SameResidentAddressApplyLoanOrderInLastMonth(accountProfile.ResidentAddress)
		if count >= riskCtlE004 {
			logs.Warn("[handleNeedReviewOrder] [E004] 近1个月内同一居住地址的申请人数 ≥ %d. count: %d, orderID: %d, accountID: %d, workerID: %d",
				riskCtlE004, count, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "E004",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	*/

	/** E005 近1个月同联系人在我司申请人数≥5 */
	riskCtlE005, _ := config.ValidItemInt64("risk_ctl_E005")
	if pass, num, _, _ := service.SameContactApplyLoanOrderInLastMonth(riskCtlE005, accountProfile.Contact1, accountProfile.Contact2); !pass {
		logs.Warn("[handleNeedReviewOrder] [E005] 近1个月同联系人在我司申请人数 %d ≥ %d. orderID: %d, accountID: %d, workerID: %d",
			num, riskCtlE005, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "E005",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/** E006：近3个月同联系人在我司申请人数≥10 */
	riskCtlE006, _ := config.ValidItemInt64("risk_ctl_E006")
	if pass, num, _, _ := service.SameContactApplyLoanOrderInLast3Month(riskCtlE006, accountProfile.Contact1, accountProfile.Contact2); !pass {
		logs.Warn("[handleNeedReviewOrder] [E006] 近3个月同联系人在我司申请人数 %d ≥ %d. orderID: %d, accountID: %d, workerID: %d", num, riskCtlE006, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "E006",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/**
	// E007：近3个月内同一居住地址的申请人数≥8
	riskCtlE007, _ := config.ValidItemInt64("risk_ctl_E007")
	if pass, num, _ := service.SameResidentAddressApplyLoanOrderInLast3Month(riskCtlE007, accountProfile.ResidentAddress); !pass {
		logs.Warn("[handleNeedReviewOrder] [E007] 近3个月内同一居住地址的申请人数 %d ≥ %d. orderID: %d, accountID: %d, workerID: %d", num, riskCtlE007, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "E007",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}
	*/

	/*
		// E008：历史同一居住地址的申请人数≥10
		riskCtlE008, _ := config.ValidItemInt64("risk_ctl_E008")
		if pass, num, _ := service.SameResidentAddressApplyLoanOrderInHistory(riskCtlE008, accountProfile.ResidentAddress); !pass {
			logs.Warn("[handleNeedReviewOrder] [E008] 历史同一居住地址的申请人数 %d ≥ %d. orderID: %d, accountID: %d, workerID: %d", num, riskCtlE008, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "E008",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	*/

	if accountProfile.CompanyName != "" {
		/** E009：近3月内同单位名称在我司申请人数≥20 */
		riskCtlE009, _ := config.ValidItemInt64("risk_ctl_E009")
		if pass, num, _ := service.SameCompanyApplyLoanOrderInLast3Month(riskCtlE009, accountProfile.CompanyName); !pass {
			logs.Warn("[handleNeedReviewOrder] [E009] 近3月内同单位名称在我司申请人数 %d ≥ %d. orderID: %d, accountID: %d, workerID: %d", num, riskCtlE009, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "E009",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	}

	/** E010 第一联系人在我司贷款历史最大逾期天数≥15天 */
	riskCtlE010, _ := config.ValidItemInt64("risk_ctl_E010")
	num, err := service.ContactsMaxOverdueDaysInLoanHistory(accountProfile.Contact1)
	if num >= riskCtlE010 {
		logs.Warn("[handleNeedReviewOrder] [E010] 第一联系人在我司贷款历史最大逾期天数 %d ≥ %d. orderID: %d, accountID: %d, workerID: %d",
			num, riskCtlE010, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "E010",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/** E011 第二联系人存在逾期中的借款订单 */
	total, _ := service.ContactsOverdueLoanOrderStat(accountProfile.Contact2)
	if total > 0 {
		logs.Warn("[handleNeedReviewOrder] [E011] 第二联系人存在逾期中的借款订单, total: %d. orderID: %d, accountID: %d, workerID: %d",
			total, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "E011",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/** E012 第二联系人在我司贷款历史最高逾期天数≥15天 */
	riskCtlE012, _ := config.ValidItemInt64("risk_ctl_E012")
	num, _ = service.ContactsMaxOverdueDaysInLoanHistory(accountProfile.Contact2)
	if num >= riskCtlE012 {
		logs.Warn("[handleNeedReviewOrder] [E012] 第二联系人在我司贷款历史最高逾期天数 %d ≥ %d. orderID: %d, accountID: %d, workerID: %d",
			num, riskCtlE012, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "E012",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	/** E013 同联系人我司申请人当前逾期人数≥3 */
	riskCtlE013, _ := config.ValidItemInt64("risk_ctl_E013")
	accountIDs, total, _ := service.SameContactsCustomerOverdueStat(accountProfile.Contact1, accountProfile.Contact2, accountProfile.AccountId)
	if total >= riskCtlE013 {
		logs.Warn("[handleNeedReviewOrder] [E013] 同联系人我司申请人当前逾期人数 %d ≥ %d, orderID: %d, accountID: %d, workerID: %d",
			total, riskCtlE013, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "E013",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)

		//命中E013规则，触发加入黑名单事件，系统自动加入黑名单（命中客户的手机，身份证，联系人手机）
		event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemMobile, accountBase.Mobile, types.RiskReasonLiar, "E013"})
		event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemIdentity, accountBase.Identity, types.RiskReasonLiar, "E013"})

		//联系人加入黑名单需要命中公共联系人规则
		commonContact := service.FindCommonContact(accountIDs)
		for _, contact := range commonContact {
			event.Trigger(&evtypes.BlacklistEv{0, types.RiskItemMobile, contact, types.RiskReasonLiar, "E013"})
		}
		//连带一起命中规则的其他账户
		for _, accountID := range accountIDs {
			accountBase, _ := models.OneAccountBaseByPkId(accountID)
			event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemMobile, accountBase.Mobile, types.RiskReasonLiar, "E013"})
			event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemIdentity, accountBase.Identity, types.RiskReasonLiar, "E013"})
		}

	}

	/*
		//E014 同居住地址的申请人当前逾期人数≥3
		riskCtlE014, _ := config.ValidItemInt64("risk_ctl_E014")
		accountIDs, total, _ = service.SameResidenceOverdueStat(accountProfile.ResidentCity, accountProfile.ResidentAddress)
		if total >= riskCtlE014 {
			logs.Warn("[handleNeedReviewOrder] [E014] 同居住地址的申请人当前逾期人数 %d ≥ %d, orderID: %d, accountID: %d, workerID: %d",
				total, riskCtlE014, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "E014",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)

			//命中E014规则，触发加入黑名单事件，系统自动加入黑名单（命中客户的手机，身份证，居住地址）
			homeAddress := accountProfile.ResidentCity + "," + accountProfile.ResidentAddress
			event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemMobile, accountBase.Mobile, types.RiskReasonLiar, "E014"})
			event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemIdentity, accountBase.Identity, types.RiskReasonLiar, "E014"})
			event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemResidentAddress, homeAddress, types.RiskReasonLiar, "E014"})
			//连带一起命中规则的其他账户
			for _, accountID := range accountIDs {
				accountBase, _ := models.OneAccountBaseByPkId(accountID)
				event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemMobile, accountBase.Mobile, types.RiskReasonLiar, "E014"})
				event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemIdentity, accountBase.Identity, types.RiskReasonLiar, "E014"})
			}

		}
	*/

	if accountProfile.CompanyName != "" {
		/** E015 同单位名称我司申请人当前逾期人数≥3 */
		riskCtlE015, _ := config.ValidItemInt64("risk_ctl_E015")
		accountIDs, total, _ = service.SameCompanyOverdueStat(accountProfile.CompanyName)
		if total >= riskCtlE015 {
			logs.Warn("[handleNeedReviewOrder] [E015] 同单位名称我司申请人当前逾期人数 %d ≥ %d, orderID: %d, accountID: %d, workerID: %d",
				total, riskCtlE015, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "E015",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	}

	/** E017 同一银行账号关联客户数>=2 */
	riskCtlE017, _ := config.ValidItemInt64("risk_ctl_E017")
	accountIDs, total, _ = service.SameBankNoStat(accountProfile.BankNo)
	if total >= riskCtlE017 {
		logs.Warn("[handleNeedReviewOrder] [E017] 同一银行账号关联客户数 %d ≥ %d, orderID: %d, accountID: %d, workerID: %d",
			total, riskCtlE017, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "E017",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	if accountProfile.CompanyName != "" {
		/** E018 同一公司正在逾期人数/同一公司所有有过在贷用户> N1 且 该公司下的所有有过在贷用户的数量>N2    命中此规则 触发组团骗贷*/
		riskCtlE018 := service.GetE018Config()
		//同一公司正在逾期人数
		_, totalOverdue, _ := service.SameCompanyOverdueStat(accountProfile.CompanyName)
		//该公司下的所有有过在贷用户的数量
		_, n2, _ := service.SameCompanyAllOrderStat(accountProfile.CompanyName)

		n1 := float64(0)
		if n2 > 0 {
			n1 = float64(totalOverdue) / float64(n2)
		}

		logs.Info("[handleNeedReviewOrder] check E018 n1:%v N1:%v n2:%v N2:%v", n1, riskCtlE018.N1, n2, riskCtlE018.N2)

		if n1 > riskCtlE018.N1 && n2 > int64(riskCtlE018.N2) {
			logs.Warn("[handleNeedReviewOrder] [E018] 同一公司正在逾期人数/同一公司所有有过在贷用户> N1 且该公司下的所有有过在贷用户的数量>N2 n1:%v > N1: %v && n2:%v >N2:%v, orderID: %d, accountID: %d, workerID: %d",
				n1, riskCtlE018.N1, n2, riskCtlE018.N2, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "E018",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)

			//命中E018规则，触发加入黑名单事件，系统自动加入黑名单（命中客户的手机，身份证，单位名称）
			event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemMobile, accountBase.Mobile, types.RiskReasonLiar, "E018"})
			event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemIdentity, accountBase.Identity, types.RiskReasonLiar, "E018"})
			event.Trigger(&evtypes.BlacklistEv{accountBase.Id, types.RiskItemCompany, accountProfile.CompanyName, types.RiskReasonLiar, "E018"})
		}
	}

	// -----Fxxx-----

	/** F001 三个月内累计逾期订单数≥3单 */
	riskCtlF001, _ := config.ValidItemInt64("risk_ctl_F001")
	condBox := map[string]interface{}{
		"is_overdue":    true,                    // 历史逾期
		"last_3_months": true,                    // 最近三个月
		"account_id":    orderData.UserAccountId, // 单个用户
	}
	total, _ = service.CustomerOverdueTotalStat(condBox)
	if total >= riskCtlF001 {
		logs.Warn("[handleNeedReviewOrder] [F001] 三个月内累计逾期订单数 %d ≥ %d, orderID: %d, accountID: %d, workerID: %d",
			total, riskCtlF001, orderID, accountBase.Id, workerID)

		hitItem := HitRiskRegularItem{
			Regular:      "F001",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	}

	// -----Gxxx-----
	tdData, telData, err := tongdun.GetPurchaseDate(accountBase.Id)
	if err != nil {
		//pass
	} else {
		/** G001 运营商入网时长 <N */
		riskCtlG001, _ := config.ValidItemInt("risk_ctl_G001")
		purchaseDate, _ := tools.GetTimeParseWithFormat(telData.PurchaseDate.FirstPurchasedate, "01/02/2006")
		if purchaseDate > 0 {
			dataRange := tools.GetDateRange(purchaseDate, tdData.CreateTime)
			if int(dataRange) < riskCtlG001 {
				logs.Warn("[handleNeedReviewOrder] [G001] 运营商入网时长 %d < %d, orderID: %d, accountID: %d, workerID: %d",
					dataRange, riskCtlG001, orderID, accountBase.Id, workerID)

				hitItem := HitRiskRegularItem{
					Regular:      "G001",
					RejectReason: types.RejectReasonLackCredit,
					HitTime:      tools.GetUnixMillis(),
				}
				hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
			}
		} else {
			logs.Warn("[handleNeedReviewOrder] [G001] 运营商入网时长 skip FirstPurchasedate: %d, orderID: %d, accountID: %d, workerID: %d",
				purchaseDate, orderID, accountBase.Id, workerID)
		}

		/** G002 当前距套餐到日期天数 telkomsel_today_to_active_until < N telkomsel_today_to_active_until = active_until - 服务器爬取时间 */
		riskCtlG002_days, _ := config.ValidItemInt("risk_ctl_G002_days")
		activeUntil, err1 := tools.GetTimeParseWithFormat(telData.AccountIndo.ActiveUntil, "01/02/2006")

		/** G002  telkomsel积分  telkomsel_poin < N */
		riskCtlG002_poin, _ := config.ValidItemInt("risk_ctl_G002_poin")
		telkomselNum, err2 := tools.Str2Int(telData.AccountIndo.TelkomselPoin)

		/** G002 剩余信用点  telkomsel_remaining_credits <N, telkomsel_remaining_credits = remaining_credits */
		riskCtlG002_credit, _ := config.ValidItemInt("risk_ctl_G002_credit")
		remainingCredits, err3 := tools.Str2Int(telData.AccountIndo.RemainingCredits)

		/** G002 运营商入网时长 */
		riskCtlG002_online, _ := config.ValidItemInt("risk_ctl_G002_online")
		online, err4 := tools.GetTimeParseWithFormat(telData.PurchaseDate.FirstPurchasedate, "01/02/2006")

		if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
			dataRange := tools.GetDateRange(tdData.CreateTime, activeUntil)
			onlineRange := tools.GetDateRange(online, tdData.CreateTime)
			if int(dataRange) < riskCtlG002_days && telkomselNum < riskCtlG002_poin && remainingCredits < riskCtlG002_credit && int(onlineRange) < riskCtlG002_online {
				logs.Warn("[handleNeedReviewOrder] [G002] %d < %d, %d < %d, %d < %d, %d < %d orderID: %d, accountID: %d, workerID: %d",
					dataRange, riskCtlG002_days, telkomselNum, riskCtlG002_poin, remainingCredits, riskCtlG002_credit, onlineRange, riskCtlG002_online, orderID, accountBase.Id, workerID)

				hitItem := HitRiskRegularItem{
					Regular:      "G002",
					RejectReason: types.RejectReasonLackCredit,
					HitTime:      tools.GetUnixMillis(),
				}
				hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
			}
		} else {
			logs.Warn("[handleNeedReviewOrder] [G002] G002 data error skip orderID: %d, accountID: %d, workerID: %d, err1: %v, err2: %v, err3: %v, err4: %v",
				orderID, accountBase.Id, workerID, err1, err2, err3, err4)
		}
	}

	// G003 首贷 根据advance ai 多头接口返回的数据做规则判断 命中一个即拒绝
	if !isReloan {
		sum := 0
		g003Config := service.GetRiskCtlG034Config("risk_ctl_G003")
		for _, statist := range multiResp.Data.Statistics {
			info, ok := service.RespodColNameMap[statist.TimePeriod]
			if !ok {
				logs.Warn("[handleNeedReviewOrder] [G003] respons Name not in map. statist:%#v orderID: %d, accountID: %d, workerID: %d", statist, orderID, accountBase.Id, workerID)
				continue
			}
			if statist.TimePeriod == "1-90d" ||
				statist.TimePeriod == "90+d" {
				sum += statist.QueryCount
			}

			configValue := service.GetConfigValueByColNameV2(g003Config, info.FiledName)
			if statist.QueryCount > configValue {
				regularName := "G003" + "-" + info.Index

				logs.Warn("[handleNeedReviewOrder] [%s] statist:%#v configValue:%d orderID: %d, accountID: %d, workerID: %d",
					regularName, statist, configValue, orderID, accountBase.Id, workerID)

				hitItem := HitRiskRegularItem{
					Regular:      regularName,
					RejectReason: types.RejectReasonLackCredit,
					HitTime:      tools.GetUnixMillis(),
				}
				hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
			}
		}

		// 检查 和是否超过配置
		configValue := g003Config.Sum
		if sum > configValue {
			regularName := "G003" + "-" + "8"

			logs.Warn("[handleNeedReviewOrder] [%s] sum:%#v configValue:%d orderID: %d, accountID: %d, workerID: %d",
				regularName, sum, configValue, orderID, accountBase.Id, workerID)
			hitItem := HitRiskRegularItem{
				Regular:      regularName,
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	}

	// G004复贷 根据advance ai 多头接口返回的数据做规则判断 命中一个即拒绝
	if isReloan {
		sum := 0
		g004Config := service.GetRiskCtlG034Config("risk_ctl_G004")
		for _, statist := range multiResp.Data.Statistics {
			info, ok := service.RespodColNameMap[statist.TimePeriod]
			if !ok {
				logs.Warn("[handleNeedReviewOrder] [G004] respons Name not in map. statist:%#v orderID: %d, accountID: %d, workerID: %d", statist, orderID, accountBase.Id, workerID)
				continue
			}

			if statist.TimePeriod == "1-90d" ||
				statist.TimePeriod == "90+d" {
				sum += statist.QueryCount
			}

			configValue := service.GetConfigValueByColNameV2(g004Config, info.FiledName)
			if statist.QueryCount > configValue {
				regularName := "G004" + "-" + info.Index

				logs.Warn("[handleNeedReviewOrder] [%s] statist:%#v configValue:%d orderID: %d, accountID: %d, workerID: %d",
					regularName, statist, configValue, orderID, accountBase.Id, workerID)
				hitItem := HitRiskRegularItem{
					Regular:      regularName,
					RejectReason: types.RejectReasonLackCredit,
					HitTime:      tools.GetUnixMillis(),
				}
				hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
			}
		}

		// 检查 和是否超过配置
		configValue := g004Config.Sum
		if sum > configValue {
			regularName := "G004" + "-" + "8"

			logs.Warn("[handleNeedReviewOrder] [%s] sum:%#v configValue:%d orderID: %d, accountID: %d, workerID: %d",
				regularName, sum, configValue, orderID, accountBase.Id, workerID)
			hitItem := HitRiskRegularItem{
				Regular:      regularName,
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

	}

	// -----Zxxx-----
	riskReq := service.RiskRequestInfo{}
	service.FillFantasyRiskRequest(&riskReq, &orderData, &accountBase, accountProfile, &clientInfo)

	riskReq.Model = "ascore"
	riskReq.Version = "v1"
	rawData, router, riskA, err := service.GetFantasyRisk(riskReq)
	if err != nil || !riskA.IsSuccess() {
		//...
		logs.Warn("[handleNeedReviewOrder] [D997] 大数据没有抓取到A数据, orderID:%d, accountID:%d, workerID:%d, reqid:%s, data:%s",
			orderID, accountBase.Id, workerID, riskA.ReqId, string(rawData))

		hitItem := HitRiskRegularItem{
			Regular:      "D997",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	} else {
		dao.SaveEsData(orderID, accountBase.Id, router, string(rawData))

		/** Z002 系统评分不足 为 A卡评分<600 */
		riskCtlZ002, _ := config.ValidItemInt("risk_ctl_Z002")
		if riskA.Data[0].Score < riskCtlZ002 {
			logs.Warn("[handleNeedReviewOrder] [Z002] 系统评分不足, score: %d, orderID: %d, accountID: %d, workerID: %d",
				riskA.Data[0].Score, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      types.RegularNameZ002,
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
				Value:        riskA.Data[0].Score,
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	}

	riskReq.Model = "ascore"
	riskReq.Version = "v2"
	rawData, router, riskA2, err := service.GetFantasyRisk(riskReq)
	if err != nil || !riskA2.IsSuccess() {
		//...
		logs.Warn("[handleNeedReviewOrder] [D997] 大数据没有抓取到A数据V2, orderID:%d, accountID:%d, workerID:%d, reqid:%s, data:%s",
			orderID, accountBase.Id, workerID, riskA2.ReqId, string(rawData))

		hitItem := HitRiskRegularItem{
			Regular:      "D997",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	} else {
		dao.SaveEsData(orderID, accountBase.Id, router, string(rawData))

		/** Z003 系统评分不足 为 A卡评分<600  2.0 */
		riskCtlZ003, _ := config.ValidItemInt("risk_ctl_Z003")
		if riskA2.Data[0].Score < riskCtlZ003 {
			logs.Warn("[handleNeedReviewOrder] [Z003] 系统评分不足V2, score: %d, orderID: %d, accountID: %d, workerID: %d",
				riskA2.Data[0].Score, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "Z003",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	}

	riskReq.Model = "bscore"
	riskReq.Version = "v1"
	rawData, router, riskB, err := service.GetFantasyRisk(riskReq)
	if err != nil || !riskB.IsSuccess() {
		//...
		logs.Warn("[handleNeedReviewOrder] [D997] 大数据没有抓取到B数据, orderID:%d, accountID:%d, workerID:%d, reqid:%s, data:%s",
			orderID, accountBase.Id, workerID, riskB.ReqId, string(rawData))

		hitItem := HitRiskRegularItem{
			Regular:      "D997",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	} else {
		dao.SaveEsData(orderID, accountBase.Id, router, string(rawData))

		/** Z004 系统评分不足 为 B卡评分<600 */
		riskCtlZ004, _ := config.ValidItemInt("risk_ctl_Z004")
		if riskB.Data[0].Score < riskCtlZ004 {
			logs.Warn("[handleNeedReviewOrder] [Z004] 系统评分不足, score: %d, orderID: %d, accountID: %d, workerID: %d",
				riskB.Data[0].Score, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "Z004",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	}

	esACardRes, router, rawData, _ := service.EsSearchACardByImei(tools.Md5(clientInfo.Imei))
	if !esACardRes.Found || !esACardRes.IsAll() {
		//...
		logs.Warn("[handleNeedReviewOrder] [D996] 大数据没有抓取到流数据, orderID:%d, accountID:%d, workerID:%d, imei:%s, data:%s",
			orderID, accountBase.Id, workerID, clientInfo.Imei, string(rawData))

		hitItem := HitRiskRegularItem{
			Regular:      "D996",
			RejectReason: types.RejectReasonLackCredit,
			HitTime:      tools.GetUnixMillis(),
		}
		hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
	} else {

		dao.SaveEsData(orderID, accountBase.Id, router, string(rawData))

		/** D034 用户现金贷类app安装数>9 */
		riskCtlD034, _ := config.ValidItemInt("risk_ctl_D034")
		if esACardRes.Source.AppCount > riskCtlD034 {
			logs.Warn("[handleNeedReviewOrder] [D034] 用户现金贷类app安装数%d>%d, orderID: %d, accountID: %d, workerID: %d",
				esACardRes.Source.AppCount, riskCtlD034, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D034",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D035 用户现金贷类app安装数>9 */
		riskCtlD035, _ := config.ValidItemInt("risk_ctl_D035")
		if esACardRes.Source.RangeAppCount > riskCtlD035 {
			logs.Warn("[handleNeedReviewOrder] [D035] 近一段时间用户安装现金贷app个数%d>%d, orderID: %d, accountID: %d, workerID: %d",
				esACardRes.Source.RangeAppCount, riskCtlD035, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D035",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** D036 用户现金贷类app安装数>9 */
		riskCtlD036, _ := config.ValidItemInt("risk_ctl_D036")
		if esACardRes.Source.RangeContactCount > riskCtlD036 {
			logs.Warn("[handleNeedReviewOrder] [D036] 近一段时间创建通讯录联系人个数%d>%d, orderID: %d, accountID: %d, workerID: %d",
				esACardRes.Source.RangeContactCount, riskCtlD036, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "D036",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	}

	graphReq := service.GraphRequestInfo{}
	service.FillFantasyGraphRequest(&graphReq, &orderData, &accountBase, accountProfile, &clientInfo)

	graphReq.Model = "graph"
	graphReq.Version = "v1"
	graphReq.Scene = "deal"
	graphRawData, rgraphRuter, graphRes, err := service.GetFantasyGraph(graphReq)
	if err != nil || !graphRes.IsSuccess() {
		//...pass
		logs.Warn("[handleNeedReviewOrder] GetFantasyGraph data wrong, orderID:%d, accountID:%d, workerID:%d, data:%s",
			orderID, accountBase.Id, workerID, string(graphRawData))
	} else {
		dao.SaveEsData(orderID, accountBase.Id, rgraphRuter, string(graphRawData))

		/** Z005 同一IP历史对应的设备数量>9 */
		riskCtlZ005, _ := config.ValidItemInt("risk_ctl_Z005")
		if graphRes.Data.Graph.IpDeviceAllNum > riskCtlZ005 {
			logs.Warn("[handleNeedReviewOrder] [Z005] 同一IP历史对应的设备数量%d>%d, orderID: %d, accountID: %d, workerID: %d",
				graphRes.Data.Graph.IpDeviceAllNum, riskCtlZ005, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "Z005",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** Z006 同一IP历史对应的账户数量>9 */
		riskCtlZ006, _ := config.ValidItemInt("risk_ctl_Z006")
		if graphRes.Data.Graph.IpAccountAllNum > riskCtlZ006 {
			logs.Warn("[handleNeedReviewOrder] [Z006] 同一IP历史对应的账户数量%d>%d, orderID: %d, accountID: %d, workerID: %d",
				graphRes.Data.Graph.IpAccountAllNum, riskCtlZ006, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "Z006",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** Z007 同一设备历史对应的账户数量>9 */
		riskCtlZ007, _ := config.ValidItemInt("risk_ctl_Z007")
		if graphRes.Data.Graph.DeviceAccountAllNum > riskCtlZ007 {
			logs.Warn("[handleNeedReviewOrder] [Z007] 同一设备历史对应的账户数量%d>%d, orderID: %d, accountID: %d, workerID: %d",
				graphRes.Data.Graph.DeviceAccountAllNum, riskCtlZ007, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "Z007",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** Z008 同一联系人历史对应的账户数量>9 */
		riskCtlZ008, _ := config.ValidItemInt("risk_ctl_Z008")
		if graphRes.Data.Graph.ContactAccountAllNum > riskCtlZ008 {
			logs.Warn("[handleNeedReviewOrder] [Z008] 同一联系人历史对应的账户数量%d>%d, orderID: %d, accountID: %d, workerID: %d",
				graphRes.Data.Graph.ContactAccountAllNum, riskCtlZ008, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "Z008",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** Z009 同一公司历史对应的账户数量>9 */
		riskCtlZ009, _ := config.ValidItemInt("risk_ctl_Z009")
		if graphRes.Data.Graph.CompanyAccountAllNum > riskCtlZ009 {
			logs.Warn("[handleNeedReviewOrder] [Z009] 同一公司历史对应的账户数量%d>%d, orderID: %d, accountID: %d, workerID: %d",
				graphRes.Data.Graph.CompanyAccountAllNum, riskCtlZ009, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "Z009",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** Z010 同一银行卡历史对应的设备数量>9 */
		riskCtlZ010, _ := config.ValidItemInt("risk_ctl_Z010")
		if graphRes.Data.Graph.BanknoDeviceAllNum > riskCtlZ010 {
			logs.Warn("[handleNeedReviewOrder] [Z010] 同一银行卡历史对应的设备数量%d>%d, orderID: %d, accountID: %d, workerID: %d",
				graphRes.Data.Graph.BanknoDeviceAllNum, riskCtlZ010, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "Z010",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** Z011 同一银行卡历史对应的账户数量>9 */
		riskCtlZ011, _ := config.ValidItemInt("risk_ctl_Z011")
		if graphRes.Data.Graph.BanknoAccountAllNum > riskCtlZ011 {
			logs.Warn("[handleNeedReviewOrder] [Z011] 同一银行卡历史对应的账户数量%d>%d, orderID: %d, accountID: %d, workerID: %d",
				graphRes.Data.Graph.BanknoAccountAllNum, riskCtlZ011, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "Z011",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}

		/** Z012 同一账户历史对应的设备数量>9 */
		riskCtlZ012, _ := config.ValidItemInt("risk_ctl_Z012")
		if graphRes.Data.Graph.AccountDeviceAllNum > riskCtlZ012 {
			logs.Warn("[handleNeedReviewOrder] [Z012] 同一账户历史对应的设备数量%d>%d, orderID: %d, accountID: %d, workerID: %d",
				graphRes.Data.Graph.AccountDeviceAllNum, riskCtlZ012, orderID, accountBase.Id, workerID)

			hitItem := HitRiskRegularItem{
				Regular:      "Z012",
				RejectReason: types.RejectReasonLackCredit,
				HitTime:      tools.GetUnixMillis(),
			}
			hitRiskRegularBox = append(hitRiskRegularBox, hitItem)
		}
	}

	return hitRiskRegularBox
}

func isHitRiskCtlLevel1(riskItems []HitRiskRegularItem) (bool, int) {
	riskCtlRegularLevel1, _ := service.RiskCtlRegularLevel1()
	isHit1 := false
	for _, item := range riskItems {
		if riskCtlRegularLevel1[item.Regular] {
			isHit1 = true
			break
		}
	}

	if isHit1 {
		return isHit1, service.FixedRiskCtlRegularRandom1
	} else {
		return isHit1, 0
	}
}

func isHitRiskCtlLevel2(riskItems []HitRiskRegularItem) (bool, int) {
	isHit2, fixValue := isHitRiskCtlLevel1(riskItems)
	if isHit2 {
		return isHit2, fixValue
	}

	riskCtlRegularLevel2, _ := service.RiskCtlRegularLevel2()
	for _, item := range riskItems {
		if riskCtlRegularLevel2[item.Regular] {
			isHit2 = true
			break
		}
	}

	if isHit2 {
		return isHit2, service.FixedRiskCtlRegularRandom2
	} else {
		return isHit2, 0
	}
}
