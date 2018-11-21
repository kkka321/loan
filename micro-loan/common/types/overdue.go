package types

import (
	"fmt"
	"sort"
)

const (
	IsOverdueYes int = 1
	IsOverdueNo  int = 0
)

const (
	IsUrgeOutYes    int = 1
	IsUrgeOutNo     int = 0
	IsUrgeOutFrozen int = 2
)

type CollectionRemindDay int

const (
	CollectionRemindDef   CollectionRemindDay = 0
	CollectionRemindTwo   CollectionRemindDay = 2
	CollectionRemindFour  CollectionRemindDay = 4
	CollectionRemindEight CollectionRemindDay = 8
)

var urgeFilterMap = map[int]string{
	IsUrgeOutNo:     "入催",
	IsUrgeOutYes:    "出催",
	IsUrgeOutFrozen: "冻结",
}

func UrgeFilterMap() map[int]string {
	return urgeFilterMap
}

// 逾期订单类型: 首贷,复贷,展期
type UrgeOrderTypeEnum int

const (
	UrgeOrderTypeFirst  UrgeOrderTypeEnum = 1
	UrgeOrderTypeRepeat UrgeOrderTypeEnum = 2
	UrgeOrderTypeRoll   UrgeOrderTypeEnum = 3
)

var urgeOrderTypeMap = map[UrgeOrderTypeEnum]string{
	UrgeOrderTypeFirst:  "首贷",
	UrgeOrderTypeRepeat: "复贷",
	UrgeOrderTypeRoll:   "展单",
}

func UrgeOrderTypeMap() map[UrgeOrderTypeEnum]string {
	return urgeOrderTypeMap
}
func GetUrgeOrderTypeVal(key int) string {
	return urgeOrderTypeMap[UrgeOrderTypeEnum(key)]
}

// 逾期案件等级
const (
	OverdueLevelM11 = "M1-1" // 逾期2—12天，M1-1
	OverdueLevelM12 = "M1-2" // 逾期13—20天，M1-2
	OverdueLevelM13 = "M1-3" // 逾期21—30天，M1-3
	OverdueLevelM2  = "M2"   // 逾期31—60天，M2
	OverdueLevelM3  = "M3"   //逾期61—90天，M3
)

// OverdueOrderEndDays 第91天为坏账， 不再计息， M3案子也会自动过期
// 也就是逾期最大天数为90天，OverdueOrderEndDays - 1
const OverdueOrderEndDays = 91

var overdueLevelCreateDaysMap = map[string]int{
	OverdueLevelM11: OverdueLevelM11MinDay,
	OverdueLevelM12: OverdueLevelM12MinDay,
	OverdueLevelM13: OverdueLevelM13MinDay,
	OverdueLevelM2:  OverdueLevelM2MinDay,
	OverdueLevelM3:  OverdueLevelM3MinDay,
}

// OverdueLevelCreateDaysMap 返回案件创建是在逾期的哪一天配置列表
func OverdueLevelCreateDaysMap() map[string]int {
	return overdueLevelCreateDaysMap
}

// GetOverdueCaseExpireTime 返回案件创建是在逾期的哪一天配置列表
func GetOverdueCaseExpireTime(ticketItem TicketItemEnum, repayDate int64) int64 {
	var overdueLevel string
	for l, ti := range overdueLevelTicketItemMap {
		if ti == ticketItem {
			overdueLevel = l
			break
		}
	}
	createDays := overdueLevelCreateDaysMap[overdueLevel]
	//
	var willUpSlice []int
	for _, d := range overdueLevelCreateDaysMap {
		if d > createDays {
			willUpSlice = append(willUpSlice, d)
		}
	}
	if len(willUpSlice) == 0 {
		return repayDate + int64(OverdueOrderEndDays)*24*3600*1000
	}
	sort.Ints(willUpSlice)
	return repayDate + int64(willUpSlice[0])*24*3600*1000
}

// OverdueItemEnum 工单优先级
type OverdueItemEnum int

// 逾期案件， 时间范围定义
// 只记录初始日期， 减少复杂性， 逾期案件为一个接着一个， 无日期跳跃
const (
	OverdueLevelM11MinDay = 1
	OverdueLevelM12MinDay = 5
	OverdueLevelM13MinDay = 13
	OverdueLevelM2MinDay  = 31
	OverdueLevelM3MinDay  = 61

	// max day范围暂时未被使用， 故先注释，
	// OverdueLevelM11MaxDay = 7
	// OverdueLevelM12MaxDay = 15
	// OverdueLevelM13MaxDay = 30
	// OverdueLevelM2MaxDay  = 60
	// OverdueLevelM3MaxDay  = 90
)

// Ticket 具体项
const (
	OverdueItemM11 OverdueItemEnum = 1
	OverdueItemM12 OverdueItemEnum = 2
	OverdueItemM13 OverdueItemEnum = 3
	OverdueItemM2  OverdueItemEnum = 4
	OverdueItemM3  OverdueItemEnum = 5
)

var overdueLevelItemMap = map[OverdueItemEnum]string{
	OverdueItemM11: OverdueLevelM11,
	OverdueItemM12: OverdueLevelM12,
	OverdueItemM13: OverdueLevelM13,
	OverdueItemM2:  OverdueLevelM2,
	OverdueItemM3:  OverdueLevelM3,
}

// OverdueLevelItemMap 读取逾期案件等级
func OverdueLevelItemMap() map[OverdueItemEnum]string {
	return overdueLevelItemMap
}

// OverdueLevelItemMap 读取逾期案件等级
func GetOverdueLevelItemVal(key int) string {
	return overdueLevelItemMap[OverdueItemEnum(key)]
}

var overdueLevelTicketItemMap = map[string]TicketItemEnum{
	OverdueLevelM11: TicketItemUrgeM11,
	OverdueLevelM12: TicketItemUrgeM12,
	OverdueLevelM13: TicketItemUrgeM13,
	OverdueLevelM2:  TicketItemUrgeM20,
	OverdueLevelM3:  TicketItemUrgeM30,
}

// OverdueLevelTicketItemMap 读取逾期案件等级与ticket类型的映射关系表
func OverdueLevelTicketItemMap() map[string]TicketItemEnum {
	return overdueLevelTicketItemMap
}

// GetOverdueLevelByTicketItem 根据工单类型, 获取催收case类型
func GetOverdueLevelByTicketItem(ticketItem TicketItemEnum) string {
	for overdueLevel, ti := range overdueLevelTicketItemMap {
		if ticketItem == ti {
			return overdueLevel
		}
	}
	return ""
}

var overdueConf = []string{
	OverdueLevelM11,
	OverdueLevelM12,
	OverdueLevelM13,
	OverdueLevelM2,
	OverdueLevelM3,
}

func OverdueConf() []string {
	return overdueConf
}

func GetPreviousOverdueLevel(level string) (preLevel string, err error) {
	var index int = -1
	for i, l := range overdueConf {
		if l == level {
			index = i - 1
			break
		}
	}

	if index >= 0 && index < len(overdueConf) {
		preLevel = overdueConf[index]
	} else {
		err = fmt.Errorf("can not find previous level for: %s", level)
	}

	return
}

type UrgeOutReasonEnum int

const (
	UrgeOutReasonCleared     UrgeOutReasonEnum = 1
	UrgeOutReasonAdjust      UrgeOutReasonEnum = 2
	UrgeOutReasonRollCleared UrgeOutReasonEnum = 3
)

var urgeOutReasonMap = map[UrgeOutReasonEnum]string{
	UrgeOutReasonCleared:     "已结清",
	UrgeOutReasonAdjust:      "案件级别调整",
	UrgeOutReasonRollCleared: "展期结清",
}

func UrgeOutReasonMap() map[UrgeOutReasonEnum]string {
	return urgeOutReasonMap
}

const (
	PhoneConnected    = 1
	PhoneNotConnected = 0
)

var phoneConnectMap = map[int]string{
	PhoneConnected:    "接通",
	PhoneNotConnected: "未接通",
}

func PhoneConnectMap() map[int]string {
	return phoneConnectMap
}

const (
	ConnectWillingToPay    = 1
	ConnectNotWillingToPay = 2
	ConnectNotTheCustomer  = 3
	ConnectHangUp          = 4
)

var repayInclinationMap = map[int]string{
	ConnectWillingToPay:    "有还款意愿",
	ConnectNotWillingToPay: "无还款意愿",
	ConnectNotTheCustomer:  "不是客户本人接听",
	ConnectHangUp:          "接听后挂断",
}

func RepayInclinationMap() map[int]string {
	return repayInclinationMap
}

const (
	UnconnectReasonNullOutOfTheServiceArea  = 1
	UnconnectReasonNullTheNetworkIsBusy     = 2
	UnconnectReasonNullRegect               = 3
	UnconnectReasonNullNotRegistered        = 4
	UnconnectReasonNullBlockAllIncomingCall = 5
	UnconnectReasonNullBackToMainScreen     = 6
	UnconnectReasonNullContinueRinging      = 7
	UnconnectReasonEmptyNumber              = 8
)

var unconnectReasonMap = map[int]string{
	UnconnectReasonNullOutOfTheServiceArea:  "不在服务区",
	UnconnectReasonNullTheNetworkIsBusy:     "用户忙",
	UnconnectReasonNullRegect:               "拒接",
	UnconnectReasonNullNotRegistered:        "用户不存在",
	UnconnectReasonNullBlockAllIncomingCall: "客户设置拒接所有来电",
	UnconnectReasonNullBackToMainScreen:     "拨打后返回主页面",
	UnconnectReasonNullContinueRinging:      "无人接听",
	UnconnectReasonEmptyNumber:              "空号",
}

func UnconnectReasonMap() map[int]string {
	return unconnectReasonMap
}

const (
	PhoneObjectSelf        = 1
	PhoneObjectContact1    = 2
	PhoneObjectContact2    = 3
	PhoneObjectCompany     = 4
	PhoneObjectContactList = 5
	PhoneObjectOther       = 6
)

var phoneObjectMap = map[int]string{
	PhoneObjectSelf:        "本人",
	PhoneObjectContact1:    "联系人1",
	PhoneObjectContact2:    "联系人2",
	PhoneObjectCompany:     "公司电话",
	PhoneObjectContactList: "通讯录",
	PhoneObjectOther:       "其他联系人",
}

func PhoneObjectMap() map[int]string {
	return phoneObjectMap
}

const (
	UrgeTypeMobile   = 1
	UrgeTypeWhatsapp = 2
)

var urgeTypeMap = map[int]string{
	UrgeTypeMobile:   "电话",
	UrgeTypeWhatsapp: "Whatsapp",
}

func UrgeTypeMap() map[int]string {
	return urgeTypeMap
}

const (
	ClearReducedInvalidReasonCaseUp        string = "案件升级"
	ClearReducedInvalidReasonNotClear      string = "未还够预减免应还总额"
	ClearReducedInvalidReasonAmountInvalid string = "可减免金额为0，结清减免失效"
	ReduceInvalidReasonPenalty             string = "减免值大于可减免罚息"
	ReduceInvalidReasonGrace               string = "减免值大于可减免宽限息"
	ReduceInvalidReasonAmount              string = "减免值大于可减免本金"
	ReduceInvalidReasonOrdersStatus        string = "订单状态不允许减免"
)

// OverdueReasonItemEnum 还款原因项
type OverdueReasonItemEnum int

// 逾期原因枚举项, 不要破坏原有顺序
// 逾期原因：alasan-alasan customer tidak mau bayar
// -没有钱   -belum punya uang
// -小孩子要上学- karna kebutuhan anak sekolah
// -没发工资- belum menerima gaji
// -去医院（生病、出事故、家里有人生病）-karna sakit,kecelakaan, musibah dalam keluarga
// -忘记有借款- karena lupa  dia ada pinjaman di rupiah cepat
// -不知道应该如何操作还款/ATM尝试还款但失败-customer tidak tahu cara melakukan pembayaran/selalu gagal waktu pembayaran di ATM
const (
	OverdueReasonNoMoney OverdueReasonItemEnum = iota + 1
	OverdueReasonChildGoSchool
	OverdueReasonNoWage
	OverdueReasonSick
	OverdueReasonForgetLoan
	OverdueReasonOperationUnknown

	// 上面插入新增原因
	OverdueReasonOther OverdueReasonItemEnum = 100
)

var overdueReasonItemMap = map[OverdueReasonItemEnum]string{
	OverdueReasonNoMoney:          "没有钱",
	OverdueReasonChildGoSchool:    "小孩子要上学",
	OverdueReasonNoWage:           "没发工资",
	OverdueReasonSick:             "去医院(生病、出事故、家里有人生病)",
	OverdueReasonForgetLoan:       "忘记有借款",
	OverdueReasonOperationUnknown: "不知道应该如何操作还款/ATM尝试还款但失败",
	OverdueReasonOther:            "其他",
}

// OverdueReasonItemMap 返回逾期原因map
func OverdueReasonItemMap() map[OverdueReasonItemEnum]string {
	return overdueReasonItemMap
}

var urgeResultMap = map[int]string{
	1:  "承诺还款",
	2:  "无还款意愿",
	3:  "非借款人",
	4:  "振铃不接",
	5:  "不在服务区",
	6:  "暂时无法接通",
	7:  "呼叫转接",
	8:  "用户忙",
	9:  "被拉黑",
	10: "空号",
	11: "未收到放款",
	12: "已付款未入系统",
	13: "已留言联系人或通讯录",
	14: "意愿展期",
	15: "意愿部分还款",
	16: "没声音",
}

// UrgeResultMap 返回催收结果map
func UrgeResultMap() map[int]string {
	return urgeResultMap
}

var notRepayReasonMap = map[int]string{
	1: "回调延迟",
	2: "客户未收到放款",
	3: "其他",
}

// NotRepayReasonMap 返回未还原因map
func NotRepayReasonMap() map[int]string {
	return notRepayReasonMap
}

const (
	ReduceTypeManual     int = 1
	ReduceTypeAuto       int = 2
	ReduceTypePrereduced int = 3

	ClearReducedNotValid int = 0 //未生效
	ClearReducedValid    int = 1 //已生效
	ClearReducedInValid  int = 2 //失效

	ReduceStatusNotValid = 1
	ReduceStatusApplyed  = 2
	ReduceStatusRejected = 3
	ReduceStatusInvalid  = 4
	ReduceStatusValid    = 5

	//通过/拒绝
	ReduceConfirmOptionPass   = 1
	ReduceConfirmOptionReject = 2
)

var ReduceTypeMap = map[int]string{
	ReduceTypeManual:     "普通",
	ReduceTypeAuto:       "自动",
	ReduceTypePrereduced: "结清",
}

var ReduceStatusMap = map[int]string{
	ReduceStatusNotValid: "未生效",
	ReduceStatusApplyed:  "等待审核",
	ReduceStatusRejected: "审核拒绝",
	ReduceStatusInvalid:  "失效",
	ReduceStatusValid:    "生效",
}

var ReduceConfirmOptionMap = map[int]string{
	ReduceConfirmOptionPass:   "通过",
	ReduceConfirmOptionReject: "拒绝",
}
