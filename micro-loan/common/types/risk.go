package types

// 风险上报/解除原因

type RiskTypeEnum int

const (
	RiskUndefined RiskTypeEnum = 0
	RiskBlacklist RiskTypeEnum = 1
	RiskGraylist  RiskTypeEnum = 2
)

var riskTypeMap = map[RiskTypeEnum]string{
	RiskBlacklist: "黑名单",
	RiskGraylist:  "灰名单",
}

func RiskTypeMap() map[RiskTypeEnum]string {
	return riskTypeMap
}

type RiskReason int

const (
	RiskReasonFake        RiskReason = 1
	RiskReasonBroker      RiskReason = 2
	RiskReasonLiar        RiskReason = 3
	RiskReasonHighRisk    RiskReason = 4
	RiskReasonLostContact RiskReason = 5
	RiskReasonOther       RiskReason = 6
	RiskReasonAkulaku     RiskReason = 7
	RiskReasonAdvance     RiskReason = 8
)

var riskReportReasonMap = map[RiskReason]string{
	RiskReasonFake:        "伪冒申请",
	RiskReasonBroker:      "中介代办",
	RiskReasonLiar:        "组团骗贷",
	RiskReasonHighRisk:    "贷后高风险",
	RiskReasonLostContact: "失联客户",
	RiskReasonOther:       "其它",
	RiskReasonAkulaku:     "Akulaku黑名单",
	RiskReasonAdvance:     "Advance黑名单",
}

func RiskReportReasonMap() map[RiskReason]string {
	return riskReportReasonMap
}

type RiskRelieveReason int

const (
	RiskRelieveReasonError RiskRelieveReason = 1
	RiskRelieveReasonClear RiskRelieveReason = 2
	RiskRelieveReasonOther RiskRelieveReason = 3
)

var riskRelieveReasonMap = map[RiskRelieveReason]string{
	RiskRelieveReasonError: "风险识别错误",
	RiskRelieveReasonClear: "负面信息消除",
	RiskRelieveReasonOther: "其它",
}

// RiskRelieveReasonMap 返回解除黑名单原因列表
func RiskRelieveReasonMap() map[RiskRelieveReason]string {
	return riskRelieveReasonMap
}

type RiskItemEnum int

const (
	RiskItemMobile          RiskItemEnum = 1
	RiskItemIdentity        RiskItemEnum = 2
	RiskItemResidentAddress RiskItemEnum = 3
	RiskItemCompany         RiskItemEnum = 4
	RiskItemCompanyAddress  RiskItemEnum = 5
	RiskItemIMEI            RiskItemEnum = 6
	RiskItemIP              RiskItemEnum = 7
)

var riskItemMap = map[RiskItemEnum]string{
	RiskItemMobile:          "手机号码",
	RiskItemIdentity:        "身份证号码",
	RiskItemResidentAddress: "居住地址",
	RiskItemCompany:         "单位名称",
	RiskItemCompanyAddress:  "单位地址",
	RiskItemIMEI:            "设备号",
	RiskItemIP:              "IP",
}

func RiskItemMap() map[RiskItemEnum]string {
	return riskItemMap
}

// 风控相关 {
type RiskCtlEnum int

const (
	//! AF = AntiFraud
	RiskCtlUndefined            RiskCtlEnum = 0  // 未定义,此状态值的订单风险控制不用关心
	RiskCtlAFDoing              RiskCtlEnum = 1  // 1、反欺诈处理中：借款订单提交审核成功后，风控状态为反欺诈处理中
	RiskCtlAFReject             RiskCtlEnum = 2  // 2、反欺诈拒绝：借款订单由于客户反欺诈规则未通过，则状态更新为反欺诈拒绝
	RiskCtlAFPassDirect         RiskCtlEnum = 3  // 3、反欺诈直批：借款订单由于反欺诈通过，且在内部白名单，则状态更新为反欺诈直批
	RiskCtlWaitPhoneVerify      RiskCtlEnum = 4  // 4、等待电核：反欺诈通过，但未在内部白名单，风控状态更新为等待电核
	RiskCtlPhoneVerifyDoing     RiskCtlEnum = 5  // 5、电核处理中：电核人员开始电核处理，但未给出电核意见时，风控状态为电核处理中
	RiskCtlPhoneVerifyPass      RiskCtlEnum = 6  // 6、电核通过：电核人员完成电核，且电核意见为通过时，风控状态为电核通过
	RiskCtlPhoneVerifyReject    RiskCtlEnum = 7  // 7、电核拒绝：电核人员完成电核，且电核意见为拒绝时，风控状态为电核拒绝
	RiskCtlThirdBlacklistDoing  RiskCtlEnum = 8  // 8、第三方黑名单处理：第三方黑名单处理中
	RiskCtlThirdBlacklistPass   RiskCtlEnum = 9  // 9、第三方黑名单通过：第三方黑名单启用并通过
	RiskCtlThirdBlacklistReject RiskCtlEnum = 10 // 10、第三方黑名单拒绝：第三方黑名单启用并在黑名单中
	RiskCtlWaitAutoCall         RiskCtlEnum = 11 // 11、等待自动外呼：infoReview工单在审核通过，A卡分在某区间后的风控状态
	RiskCtlAutoCallPass         RiskCtlEnum = 12 // 12、自动外呼通过：infoReview工单在审核通过，A卡分在某区间，并且手机能拨通后的风控状态
	RiskCtlAutoCallReject       RiskCtlEnum = 13 // 13、等待自动拒绝：infoReview工单在审核通过，A卡分在某区间，但是手机不能拨通的风控状态
	RiskCtlWaitPhotoCompare     RiskCtlEnum = 14 //等待人脸比对 （第三方黑名单后如果打开人脸比对开关进入该状态）
	RiskCtlPhotoComparePass     RiskCtlEnum = 15 //人脸比对成功
	RiskCtlPhotoCompareFail     RiskCtlEnum = 16 //人脸比对失败
)

var riskCtlMap = map[RiskCtlEnum]string{
	RiskCtlAFDoing:              "反欺诈处理中",
	RiskCtlAFReject:             "反欺诈拒绝",
	RiskCtlAFPassDirect:         "反欺诈直批",
	RiskCtlWaitPhoneVerify:      "等待电核",
	RiskCtlPhoneVerifyDoing:     "电核处理中",
	RiskCtlPhoneVerifyPass:      "电核通过",
	RiskCtlPhoneVerifyReject:    "电核拒绝",
	RiskCtlThirdBlacklistDoing:  "第三方黑名单处理中",
	RiskCtlThirdBlacklistPass:   "第三方黑名单通过",
	RiskCtlThirdBlacklistReject: "第三方黑名单拒绝",
	RiskCtlWaitAutoCall:         "等待自动外呼",
	RiskCtlAutoCallPass:         "自动外呼通过",
	RiskCtlAutoCallReject:       "自动外呼拒绝",
	RiskCtlWaitPhotoCompare:     "等待人脸比对",
	RiskCtlPhotoComparePass:     "人脸比对通过",
	RiskCtlPhotoCompareFail:     "人脸比对失败",
}

func RiskCtlMap() map[RiskCtlEnum]string {
	return riskCtlMap
}

// 是否复贷 {
type IsReloanEnum int

const (
	IsReloanNo  IsReloanEnum = 0
	IsReloanYes IsReloanEnum = 1
)

var isReloanMap = map[IsReloanEnum]string{
	IsReloanNo:  "首贷",
	IsReloanYes: "复贷",
}

func IsReloanMap() map[IsReloanEnum]string {
	return isReloanMap
}

// }

// 订单被拒绝的原因
type RejectReasonEnum int

const (
	RejectReasonAge          RejectReasonEnum = 1
	RejectReasonGPS          RejectReasonEnum = 2
	RejectReasonHasOverdue   RejectReasonEnum = 3
	RejectReasonLackCredit   RejectReasonEnum = 4
	RejectReasonVerifyFail   RejectReasonEnum = 5
	RejectReasonHitBlackList RejectReasonEnum = 6
)

var rejectReasonMap = map[RejectReasonEnum]string{
	RejectReasonAge:          "客户年龄不符合要求",
	RejectReasonGPS:          "GPS所在区域不符",
	RejectReasonHasOverdue:   "客户当前贷款已逾期",
	RejectReasonLackCredit:   "信用评分不足",
	RejectReasonVerifyFail:   "审查拒绝", // 电核拒绝的通用原因
	RejectReasonHitBlackList: "命中黑名单",
}

func RejectReasonMap() map[RejectReasonEnum]string {
	return rejectReasonMap
}

// RiskStatusEnum 风险审核状态枚举
type RiskStatusEnum int

const (
	// RiskWaitReview 待审核
	RiskWaitReview RiskStatusEnum = 0
	// RiskReviewPass 审核通过
	RiskReviewPass RiskStatusEnum = 1
	// RiskReviewReject 审核拒绝
	RiskReviewReject RiskStatusEnum = 2
)

var riskStatusMap = map[RiskStatusEnum]string{
	RiskWaitReview:   "等待审核",
	RiskReviewPass:   "审核通过",
	RiskReviewReject: "审核拒绝",
}

// RiskStatusMap 返回解除黑名单原因列表
func RiskStatusMap() map[RiskStatusEnum]string {
	return riskStatusMap
}

const (
	PhoneVerifyPass    int = 1
	PhoneVerifyReject  int = 2
	PhoneVerifyInvalid int = 3
)

var PhoneVerifyTypeMap = map[int]string{
	PhoneVerifyPass:    "电核通过",
	PhoneVerifyReject:  "电核拒绝",
	PhoneVerifyInvalid: "置为失效",
}

const (
	RegularNameZ002 = "Z002"
)

const (
	RemarkTagNone        = 0 // 不需要重新打标记
	RemarkTagYes         = 1 // 需要重新打标记
	RecallTagNone        = 0 // 无特殊标记
	RecallTagScore       = 3 // 评分模型需召回的客户
	RecallTagPhoneVerify = 4 // 电核环节需召回的客户
	RecallTagModifyBank  = 5 // 因账号原因放款失败的的用户
)

const (
	ModifyBankAllow     = 1
	ModifyBankForbidden = 2
)

const (
	PhoneVerifyInvalidTag = 1 // 电核阶段置为失效标签
)

// A/B卡分数在0-1500之内，定义不可能取到的边界值
const (
	RiskCtlAOrBScoreLower = -1
	RiskCtlAOrBScoreUpper = 2000
)
