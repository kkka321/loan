package types

//! 作务退出命令号
const TaskExitCmd int64 = -111

type BigInt int64

// Undefined 未定义
const Undefined = "未定义"

type GenderEnum int

const (
	GenderSecrecy GenderEnum = -1
	GenderFemale  GenderEnum = 0
	GenderMale    GenderEnum = 1
)

var genderEnumMap = map[GenderEnum]string{
	GenderSecrecy: "未定义",
	GenderFemale:  "女",
	GenderMale:    "男",
}

func GenderEnumMap() map[GenderEnum]string {
	return genderEnumMap
}

// Robot 用于记录操作人时,非人为操作, 机器操作
const Robot = 0

// DefaultPagesize 后台分页列表中的默认单页条数
const DefaultPagesize = 15

// admin session keys config
const (
	SessAdminIsLogin  string = "AdminIsLogin"
	SessAdminUid      string = "AdminUid"
	SessAdminNickname string = "AdminNickname"
	SessAdminRoleID   string = "AdminRoleID"
	SessAdminRoleType string = "AdminRoleType"
	SessAdminRolePid  string = "AdminRolePid"
)

// orm 已经注册数据别名,需要有个`default`
const (
	OrmDataBaseAdmin            string = "default"
	OrmDataBaseApi              string = "api"
	OrmDataBaseAdminSlave       string = "adminSlave"
	OrmDataBaseApiSlave         string = "apiSlave"
	OrmDataBaseRiskMonitor      string = "riskMonitor"
	OrmDataBaseRiskMonitorSlave string = "riskMonitorSlave"
	OrmDataBaseMessage          string = "message"
)

// 资源使用标记
type ResourceUseMark int

const (
	Use2IdentityDetect        ResourceUseMark = 1  // 用于身份识别
	Use2FaceidVerify          ResourceUseMark = 2  // 用于活体识别
	Use2ReLoanHandHeldIdPhoto ResourceUseMark = 3  // 用于活体识别
	Use2FeedbackPhoto         ResourceUseMark = 4  // 用于反馈
	Use2Refund                ResourceUseMark = 5  // 用于退款凭证
	Use2PaymentVoucher        ResourceUseMark = 6  // 用于还款凭证
	Use2Advertisement         ResourceUseMark = 7  // 广告
	Use2Banner                ResourceUseMark = 8  // 广告
	Use2Pop                   ResourceUseMark = 9  // 广告
	Use2Float                 ResourceUseMark = 10 // 广告
	Use2AdPosition            ResourceUseMark = 11 // 广告位合作公司广告
)

var resourceUseMarkMap = map[ResourceUseMark]string{
	Use2IdentityDetect:        "身份证照片",
	Use2FaceidVerify:          "活体识别",
	Use2ReLoanHandHeldIdPhoto: "复贷手持身份证照片",
	Use2FeedbackPhoto:         "用户反馈",
	Use2PaymentVoucher:        "用于还款凭证",
}

func ResourceUseMarkMap() map[ResourceUseMark]string {
	return resourceUseMarkMap
}

const FaceidVerifyConfidence float64 = 75.0

// IDHoldingPhotoCheckResult 手持照片比对结果阈值
const IDHoldingPhotoCheckResult float64 = 60.0

// LivingBestAndReloanHandholdSimilar 活体最佳与复贷手持比对阈值
const LivingBestAndReloanHandholdSimilar float64 = 60.0

// 默认超管uid
const SuperAdminUID int64 = 1

const (
	DeletedNo  int = 0 // 标记为未删除
	DeletedYes int = 1 // 标记为已删除

	StatusInvalid int = 0 // 状态无效
	StatusValid   int = 1 // 状态是有效的
)

var statusMap = map[int]string{
	StatusInvalid: "无效",
	StatusValid:   "正常",
}

func StatusMap() map[int]string {
	return statusMap
}

// 客户的借款生命周期
const (
	LoanLifetimeExcept     int = 0 // 异常
	LoanLifetimeNormal     int = 1 // 正常状态,所有资料可以编辑
	LoanLifetimeInProgress int = 2 // 订单正在进行中,没有完结.不可创建新订单(不包括临时订单).处在此阶段的订单状态有:(已提交审核, 等待人工审核, 等待放款, 放款失败, 等待还款, 部分还款, 逾期)
	LoanLifetimeReject     int = 3 // 最后一次申请被拒绝,并且被拒不满7天.对应最后一条有效订单状态为: 审核拒绝
	LoanHitBlackList       int = 4 // 客户的手机号/IP地址命中内部黑名单
)

// 是否是临时记录
const (
	IsTemporaryYes int = 1
	IsTemporaryNO  int = 0
)

// 是否坏账
const (
	IsDeadDebtYes int = 1
	IsDeadDebtNo  int = 0
)

const (
	ProductChargeFeeInterestBefore int = 0
	ProductChargeFeeInterestAfter  int = 1
)

const (
	// 旧版本表示已完成某步骤
	AccountInfoCompletePhaseNone    int = 0
	AccountInfoCompletePhaseBase    int = 1 //完成身份信息
	AccountInfoCompletePhaseLive    int = 2 //完成活体认证
	AccountInfoCompletePhaseWork    int = 3 //完成工作信息
	AccountInfoCompletePhaseContact int = 4 //完成联系人信息
	AccountInfoCompletePhaseDone    int = 5 //完成其他信息
	AccountInfoCompleteOptVerify    int = 6 //完成运营商授权
	AccountInfoCompleteAddition     int = 7 //完成其他授信

	AccountInfoCompletePhaseNoneReLoan         int = 100 //复贷 未填写任何信息，客户端会进入手持证件照页面
	AccountInfoCompletePhaseHoldReLoan         int = 101 //复贷 完成手持证件照，客户端会进入活体认证页面
	AccountInfoCompletePhaseLiveReLoan         int = 102 //复贷 完成活体认证，客户端会进入确认订单页
	AccountInfoCompletePhaseJumpToAuthoriation int = 103 //复贷 客户端会进入授权信息页

	// 表示要进行某步骤
	AccountInfoPhaseWork    int = 1001 //工作信息
	AccountInfoPhaseContact int = 1002 //联系人信息
	AccountInfoPhaseOther   int = 1003 //其他信息
	AccountInfoPhaseBase    int = 1004 //身份信息
	AccountInfoPhaseLive    int = 1005 //活体认证
	AccountInfoAddition     int = 1006 //其他授信
	AccountInfoComplete     int = 1007 //完成借贷流程

	DefaultLoanFlow string = "1001,1002,1003,1004,1005,1006" // 默认的借款流程配置
)

/**
1    手机号          08开头的10——12位数字                        长度错误
2    姓名            不允许输入数字，允许特殊字符，最小长度2     格式不符/长度错误
3    身份证号        16位数字                                    长度错误
4    公司名称        允许特殊字符，最小长度5                     格式不符/长度错误
5    公司详细地址    允许数字和特殊字符，最小长度10              长度错误
6    住址详细地址    允许数字和特殊字符，最小长度10              长度错误
7    银行卡号        不允许字符和特殊字符，仅允许数字，最小长度5 格式不符/长度错误
8    联系人号码      0开头的10——12位数字                         长度错误
*/
const (
	LimitMobile          int = 10
	LimitMobileCompany   int = 7
	LimitName            int = 1
	LimitIdentity        int = 16
	LimitCompanyName     int = 5
	LimitCompanyAddress  int = 10
	LimitResidentAddress int = 10
	LimitBankNo          int = 5
	LimitAge             int = 18 // 客户年龄不能低于此限制
	LimitSalaryDay       int = 0
)

const (
	PayTypeMoneyIn   int = 1
	PayTypeMoneyOut  int = 2
	PayTypeRefundIn  int = 3 // 退款给用户 入账（统计信息）
	PayTypeRefundOut int = 4 // 退款给用户 出帐（明细，具体退到那一块）
	PayTypeRollIn    int = 5 // 展期入账
	PayTypeRollOut   int = 6 // 展期出帐
	PayTypeTran      int = 7 // 余额结转
)

const (
	None                  int = 0
	Xendit                int = 1
	Bluepay               int = 2
	DoKu                  int = 3
	MobiPreInterest       int = 1001
	MobiReductionInterest int = 1002
	MobiReductionPenalty  int = 1003
	MobiRefundToOrder     int = 1004
	MobiFundTran          int = 1005
	MobiFundVirtual       int = 1006
	MobiCoupon            int = 2001
)

const (
	WorkType1 int = 1 //全职
	WorkType2 int = 2 //兼职
	WorkType3 int = 3 //个体户
	WorkType4 int = 4 //无业
)

var vaCodeNameMap = map[int]string{
	Xendit:  "Xendit",
	Bluepay: "Bluepay",
	DoKu:    "DoKu",
}

func VaCodeNameMap() map[int]string {
	return vaCodeNameMap
}

var fundCodeNameMap = map[int]string{
	Xendit: "Xendit",
	//Bluepay: "Bluepay",
	DoKu: "DoKu",
}

func FundCodeNameMap() map[int]string {
	return fundCodeNameMap
}

var failureCodeMap = map[int]string{
	1:  "API_VALIDATION_ERROR",
	2:  "BANK_CODE_NOT_SUPPORTED_ERROR",
	3:  "INSUFFICIENT_BALANCE",
	4:  "INVALID_DESTINATION",
	5:  "RECIPIENT_ACCOUNT_NUMBER_ERROR",
	6:  "SWITCHING_NETWORK_ERROR",
	7:  "TEMPORARY_BANK_NETWORK_ERROR",
	8:  "TEMPORARY_TRANSFER_ERROR",
	9:  "Transfer Decline",
	10: "Transfer Inquiry Decline",
	11: "Name_Contain_Number",
	12: "SERVER_ERROR",
	13: "Inquiry Timeout Fail",
	14: "Exceed Balance",
	15: "",
	16: "UNKNOWN_BANK_NETWORK_ERROR",
}

func FailureCodeMap() map[int]string {
	return failureCodeMap
}

const (
	XenditFiledName  = "xendit_brevity_name"
	BluepayFiledName = "bluepay_brevity_name"
	DokuFiledName    = "doku_brevity_name"
)

const (
	LoanRepayTypeLoan  int = 1
	LoanRepayTypeRepay int = 2
)

const (
	PayeeCountryIDId    string = "ID"
	PayeeMsisdnID       int    = 62
	PayeeTypePersonal   string = "NORMAL"
	PayeeTypeCompany    string = "company"
	PayeeTypeIDCurrency string = "IDR"
)

const (
	XenditMarketPaymentBankCode string = "ALFAMART"
	XenditFixPaymentCode        string = "FIX_ALFAMART"
)

const (
	MobileBankCodeBRI     int = 1
	MobileBankCodeBNI     int = 2
	MobileBankCodeMANDIRI int = 3
	MobileBankCodeCIMB    int = 4
	MobileBankCodePERMATA int = 5
	MobileBankCodeDANAMON int = 6
	MobileBankCodeBCA     int = 7

	MobileBankCodeStrBRI     = "BRI"
	MobileBankCodeStrBNI     = "BNI"
	MobileBankCodeStrMANDIRI = "MANDIRI"
	MobileBankCodeStrCIMB    = "CIMB"
	MobileBankCodeStrPERMATA = "PERMATA"
	MobileBankCodeStrDANAMON = "DANAMON"
	MobileBankCodeStrBCA     = "BCA"
)

var mobileBankCodeMap = map[int]string{
	MobileBankCodeBRI:     MobileBankCodeStrBRI,
	MobileBankCodeBNI:     MobileBankCodeStrBNI,
	MobileBankCodeMANDIRI: MobileBankCodeStrMANDIRI,
	MobileBankCodeCIMB:    MobileBankCodeStrCIMB,
	MobileBankCodePERMATA: MobileBankCodeStrPERMATA,
	MobileBankCodeDANAMON: MobileBankCodeStrDANAMON,
}

func MobileBankCodeMap() map[int]string {
	return mobileBankCodeMap
}

var mobileBankCodeMapV2 = map[int]string{
	MobileBankCodeBRI:     MobileBankCodeStrBRI,
	MobileBankCodeBNI:     MobileBankCodeStrBNI,
	MobileBankCodeMANDIRI: MobileBankCodeStrMANDIRI,
	MobileBankCodeCIMB:    MobileBankCodeStrCIMB,
	MobileBankCodePERMATA: MobileBankCodeStrPERMATA,
	MobileBankCodeDANAMON: MobileBankCodeStrDANAMON,
	MobileBankCodeBCA:     MobileBankCodeStrBCA,
}

func MobileBankCodeMapV2() map[int]string {
	return mobileBankCodeMapV2
}

var repayVaCompanyCodeMap = map[string]int{
	MobileBankCodeStrBRI:     Xendit,
	MobileBankCodeStrBNI:     Xendit,
	MobileBankCodeStrMANDIRI: Xendit,
	MobileBankCodeStrCIMB:    DoKu,
	MobileBankCodeStrPERMATA: DoKu,
	MobileBankCodeStrDANAMON: DoKu,
	MobileBankCodeStrBCA:     DoKu,
}

func RepayVaCompanyCodeMap() map[string]int {
	return repayVaCompanyCodeMap
}

const (
	OperatorCatchFlag   string = "operator_catch_flag"
	AdditionalCatchFlag string = "additional_authorize_flag"

	OperatorCatchFlagValueSkip  int = 1 // 跳过抓取 运营商数据和补充授信公用此标志
	OperatorCatchFlagValueCatch int = 2 // 需要抓取 运营商数据和补充授信公用此标志

	OperatorVerifyStatusInvaild int = 0
	OperatorVerifyStatusSkip    int = 1
	OperatorVerifyStatusFailed  int = 2
	OperatorVerifyStatusSuccess int = 3

	OperatorMaxAttmpTimes  int = 9
	OperatorAcquireCodeMax int = 6

	IndonesiaAppUIVersion                string = "10048" // 旧版本app 不会传输ui_version，通过客户端是否传输该字段区分是否是新版本app
	IndonesiaAppRipeVersionCode          int    = 11      // 新版本低于该版本号的app贷款流程还是原来的五步，不增加数据抓取等流程(用于运营商数据抓取判断)
	IndonesiaAppRipeVersionCodeAddition  int    = 14      // 新版本低于该版本号的app贷款流程还是原来的五步，不增加数据抓取等流程(用于补充授信抓取)
	IndonesiaAppRipeVersionModifyCode    int    = 14      // 14版本之前（11、12）的运营商只抓取 Telkosmel的数据，而14（包括）之后抓取 三个运营商的数据（Telkosmel、XL、Indoset）
	IndonesiaAppRipeVersionLiveVerify    int    = 18      // 18版本之前的用户不做活体认证有效期 可配值的判断。
	IndonesiaAppRipeVersionXlCatchCode   int    = 22      // 22版本之前因xl验证码不能输入字母验证码，所以不抓xl  22（包含）之后的版本需要抓取
	IndonesiaAppRipeVersionSalaryDay     int    = 28      // 28版本以后增加发薪日
	IndonesiaAppRipeVersionNewLoanFlowT  int    = 69      // dev仅69版本走新的借款流程
	IndonesiaAppRipeVersionNewLoanFlow   int    = 70      // 线上仅70版本走新的借款流程
	IndonesiaAppRipeVersionNewReloanStep int    = 80      // 线上仅80版本走新的复贷流程，增加授信信息授信
)

//用户授权项状态
const (
	AuthorizeStatusSuccess       = 1
	AuthorizeStatusFailed        = 2
	AuthorizeStatusExpired       = 3
	AuthorizeStatusCrawleSuccess = 4
)

// 平台定义
const (
	PlatformH5      = "h5"
	PlatformAndroid = "android"
)

const (
	LiveVerifyInterval string = "live_verify_interval" //活体认证时间间隔配置项 name
)

const (
	PayMoneyAvaliable int = 0
	PayMoneyFrozen    int = 1
)

const (
//LoanOrderStatusUnknow = 1
//LoanOrderStatuSuccess = 2
//LoanOrderStatuFaild   = 3
)

var DayMap = map[string]int{
	"MON": 1,
	"TUE": 2,
	"WED": 3,
	"THU": 4,
	"FRI": 5,
	"SAT": 6,
	"SUN": 7,
}

var RemibChannel = map[string]int{
	"MANDIRI": 1,
	"BRI":     1,
	"BNI":     1,
	"Alfa":    1,
}

const (
	UUIDUnRegistered int = 0
	UUIDRegistered   int = 1
)

const (
	NoChoice   = 0 // no choice
	TagsXendit = 1 //
	TagsDoKu   = 2 //

)

var tagRemibMap = map[int]string{
	NoChoice:   "Nothing selected",
	TagsDoKu:   "DoKu",
	TagsXendit: "Xendit",
}

func TagsRemibMap() map[int]string {
	return tagRemibMap
}

const (
	TagsSend      = 1 //
	TagsTwoUpdate = 2 //
	TagsRefund    = 3 //
	TagsCallBack  = 4 //

)

var tagResultMap = map[int]string{
	NoChoice:      "Nothing selected",
	TagsSend:      "已发送跟进",
	TagsTwoUpdate: "待二次更新",
	TagsRefund:    "退款",
	TagsCallBack:  "重新回调",
}

func TagsResultMap() map[int]string {
	return tagResultMap
}

// AB测试分流标记
const (
	ABTestDividerFlagA = "A"
	ABTestDividerFlagB = "B"
)
