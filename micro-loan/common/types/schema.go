package types

type SchemaMode int

const (
	SchemaModeManual   SchemaMode = 1 //手动执行
	SchemaModeAuto     SchemaMode = 2 //自动执行
	SchemaModeBusiness SchemaMode = 3 //业务触发
)

var SchemaModeMap = map[SchemaMode]string{
	SchemaModeManual:   "手动执行",
	SchemaModeAuto:     "自动执行",
	SchemaModeBusiness: "业务触发",
}

var CouponSchemaMode = map[SchemaMode]string{
	SchemaModeManual: "手动执行",
	SchemaModeAuto:   "自动执行",
}

type SchemaStatus int

const (
	SchemaStatusOn      SchemaStatus = 1 //开启
	SchemaStatusOff     SchemaStatus = 2 //关闭
	SchemaStatusError   SchemaStatus = 3 //错误
	SchemaStatusRunning SchemaStatus = 4 //运行中
)

var SchemaStatusMap = map[SchemaStatus]string{
	SchemaStatusOn:      "正常",
	SchemaStatusOff:     "已停止",
	SchemaStatusError:   "异常",
	SchemaStatusRunning: "运行中",
}

type PushTarget int

const (
	PushTargetRegister         PushTarget = 1
	PushTargetReviewPass       PushTarget = 2
	PushTargetReviewReject     PushTarget = 3
	PushTargetLoanSuccess      PushTarget = 4
	PushTargetLoanFail         PushTarget = 5
	PushTargetWaitRepayment    PushTarget = 6
	PushTargetClear            PushTarget = 7
	PushTargetOverdue          PushTarget = 8
	PushTargetRollApplySuccess PushTarget = 9
	PushTargetRollSuccess      PushTarget = 10

	PushTargetInvalidFog            PushTarget = 11
	PushTargetInvalidHoldFog        PushTarget = 12
	PushTargetInvalidHoldNoFace     PushTarget = 13
	PushTargetInvalidHoldNoIdentify PushTarget = 14

	PushTargetCreditIncrease   PushTarget = 15
	PushTargetReCreditIncrease PushTarget = 16

	PushTargetCustom PushTarget = 90

	PushTargetRegisterNoOrder    PushTarget = 101
	PushTargetRegisterOrderNoKtp PushTarget = 102
	PushTargetNoRegister         PushTarget = 103
	PushTargetAllAccount         PushTarget = 104
)

var PushTargetMap = map[PushTarget]string{
	PushTargetRegister:         "注册成功",
	PushTargetReviewPass:       "审核通过",
	PushTargetReviewReject:     "审核拒绝",
	PushTargetLoanSuccess:      "放款成功",
	PushTargetLoanFail:         "放款失败",
	PushTargetWaitRepayment:    "等待还款",
	PushTargetClear:            "结清",
	PushTargetOverdue:          "逾期",
	PushTargetRollApplySuccess: "展期申请成功",
	PushTargetRollSuccess:      "展期成功",

	PushTargetInvalidFog:            "身份证照片模糊",
	PushTargetInvalidHoldFog:        "手持证件照模糊",
	PushTargetInvalidHoldNoFace:     "手持证件照中缺少人脸",
	PushTargetInvalidHoldNoIdentify: "手持证件照中缺少身份证",

	PushTargetCustom: "导入",

	PushTargetCreditIncrease:   "提额成功",
	PushTargetReCreditIncrease: "复贷提额成功",

	PushTargetRegisterNoOrder:    "注册未下单",
	PushTargetRegisterOrderNoKtp: "下单未填写资料",
	PushTargetNoRegister:         "未注册",
	PushTargetAllAccount:         "全量用户",
}

const (
	PushWayAccount int = 1
	PushWayImei    int = 2
)

var PushWayMap = map[int]string{
	PushWayAccount: "账号",
	PushWayImei:    "设备号",
}

const (
	MessageTypeReview   int = 1 //审核消息
	MessageTypeLoan     int = 2 //放款消息
	MessageTypeDisburse int = 3 //还款消息
	MessageTypeOther    int = 4 //其它消息
	MessageTypeTip      int = 5 //提醒消息
	MessageTypeSales    int = 6 //活动消息
)

var MessageTypeMap = map[int]string{
	MessageTypeReview:   "审核消息",
	MessageTypeLoan:     "放款消息",
	MessageTypeDisburse: "还款消息",
	MessageTypeOther:    "其它消息",
	MessageTypeTip:      "提醒消息",
	MessageTypeSales:    "活动消息",
}

const (
	MessageUnread int = 0 //未读
	MessageRead   int = 1 //已读
)

const (
	SkipToMessageCenter = "message_center"
	SkipToRegisterPage  = "register_page"
)

const (
	MessageSkipToNo           int = 0  //不跳转
	MessageSkipToAccount      int = 1  //我的账户
	MessageSkipToRepay        int = 2  //还款指引
	MessageSkipToFeedback     int = 3  //反馈
	MessageSkipToCoupon       int = 4  //优惠券
	MessageSkipToApply        int = 5  //借款首页（首贷）
	MessageSkipToApplyRe      int = 6  //借款首页（复贷）
	MessageSkipToRegiste      int = 7  //验证码快速注册
	MessageSkipToLogin        int = 8  //登录页面
	MessageSkipToAlfamart     int = 9  //Alfa Mart
	MessageSkipToVoucherUp    int = 10 //还款凭证上传
	MessageSkipToCenter       int = 11 //客服中心
	MessageSkipToNoCredit     int = 12 //补充授信（首贷）
	MessageSkipToNoCreditZRe  int = 13 //补充授信（复贷）
	MessageSkipToWorkInfo     int = 14 //工作信息
	MessageSkipToContactsInfo int = 15 //联系人信息
	MessageSkipToOtherInfo    int = 16 //其它信息
	MessageSkipToRepayHome    int = 17 //还款首页（有在贷）
	MessageSkipToInvite       int = 18 //邀请好友
)

var MessageSkipMap = map[int]string{
	MessageSkipToNo:           "不跳转",
	MessageSkipToAccount:      "我的账户",
	MessageSkipToRepay:        "还款指引",
	MessageSkipToFeedback:     "反馈",
	MessageSkipToCoupon:       "优惠券",
	MessageSkipToApply:        "借款首页(首贷)",
	MessageSkipToApplyRe:      "借款首页(复贷)",
	MessageSkipToRegiste:      "验证码快速注册",
	MessageSkipToLogin:        "登录页面",
	MessageSkipToAlfamart:     "Alfa Mart",
	MessageSkipToVoucherUp:    "还款凭证上传",
	MessageSkipToCenter:       "客服中心",
	MessageSkipToNoCredit:     "补充授信(首贷)",
	MessageSkipToNoCreditZRe:  "补充授信(复贷)",
	MessageSkipToWorkInfo:     "工作信息",
	MessageSkipToContactsInfo: "联系人信息",
	MessageSkipToOtherInfo:    "其它信息",
	MessageSkipToRepayHome:    "还款首页(有在贷)",
	MessageSkipToInvite:       "邀请好友",
}

const (
	InvalidIdentifyFog            = 1
	InvalidIdentifyHoldFog        = 2
	InvalidIdentifyHoldNoFace     = 3
	InvalidIdentifyHoldNoIdentify = 4
)

type CouponTarget int

const (
	CouponTargetCustom CouponTarget = 90

	CouponTargetRegisterNoOrder  CouponTarget = 101
	CouponTargetRegisterTmpOrder CouponTarget = 102
	CouponTargetRepayClear       CouponTarget = 103
	CouponTargetRepayOverdue     CouponTarget = 104
)

var CouponTargetMap = map[CouponTarget]string{
	CouponTargetCustom:           "导入",
	CouponTargetRegisterNoOrder:  "注册未申请用户(未创建临时订单)",
	CouponTargetRegisterTmpOrder: "注册未申请用户(已创建临时订单)",
	CouponTargetRepayClear:       "还款完成离开用户(未逾期)",
	CouponTargetRepayOverdue:     "还款完成离开用户(逾期但未入黑用户)",
}

type SmsTarget int

const (
	SmsTargetH5Register            SmsTarget = 1
	SmsTargetAuthCode              SmsTarget = 2
	SmsTargetRequestLogin          SmsTarget = 3
	SmsTargetDisburseSuccess       SmsTarget = 4
	SmsTargetRefundDisburseSuccess SmsTarget = 5
	SmsTargetRepeatedLoan          SmsTarget = 6
	SmsTargetRollApplySuccess      SmsTarget = 7
	SmsTargetRollSuccess           SmsTarget = 8
	SmsTargetPaymentCode           SmsTarget = 9
	SmsTargetInvite                SmsTarget = 10

	SmsTargetCustom SmsTarget = 90

	SmsTargetRemindOrder2 SmsTarget = 101
	SmsTargetRemindOrder4 SmsTarget = 102
	SmsTargetRemindOrder8 SmsTarget = 103
	SmsTargetRepayRemind  SmsTarget = 104
)

var SmsTargetMap = map[SmsTarget]string{
	SmsTargetH5Register:            "H5注册消息",
	SmsTargetAuthCode:              "验证码",
	SmsTargetRequestLogin:          "登录验证码",
	SmsTargetDisburseSuccess:       "放款成功",
	SmsTargetRefundDisburseSuccess: "重新放款成功",
	SmsTargetRepeatedLoan:          "复贷验证码",
	SmsTargetRollApplySuccess:      "展期申请成功",
	SmsTargetRollSuccess:           "展期成功",
	SmsTargetPaymentCode:           "付款码",
	SmsTargetInvite:                "邀请好友",

	SmsTargetCustom: "导入",

	SmsTargetRemindOrder2: "逾期2天还款提醒",
	SmsTargetRemindOrder4: "逾期4天还款提醒",
	SmsTargetRemindOrder8: "逾期8天还款提醒",
	SmsTargetRepayRemind:  "D-1,D1 还款提醒",
}
