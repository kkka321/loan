package types

type CouponType int

const (
	CouponTypeRedPacket CouponType = 1
	CouponTypeDiscount  CouponType = 2
	CouponTypeInterest  CouponType = 3
	CouponTypeLimit     CouponType = 4
)

var CouponTypeMap = map[CouponType]string{
	CouponTypeRedPacket: "红包券",
	CouponTypeDiscount:  "折扣券",
	CouponTypeInterest:  "利息券",
	CouponTypeLimit:     "额度券",
}

type CouponStatus int

const (
	CouponStatusAvailable CouponStatus = 1
	CouponStatusFrozen    CouponStatus = 2
	CouponStatusUsed      CouponStatus = 3
	CouponStatusInvalid   CouponStatus = 4
)

var CouponStatusMap = map[CouponStatus]string{
	CouponStatusAvailable: "可用",
	CouponStatusFrozen:    "冻结",
	CouponStatusUsed:      "已使用",
	CouponStatusInvalid:   "无效",
}

const (
	CouponInvalid   int = 0
	CouponAvailable int = 1
)

var CouponMap = map[int]string{
	CouponInvalid:   "无效",
	CouponAvailable: "有效",
}

var DistributeStatusMap = map[int]string{
	1: "未开始",
	2: "发放中",
	3: "已停止",
}

const (
	CouponUnread int = 0 //未读
	CouponRead   int = 1 //已读
)

const (
	InviteShare     int = 0 //分享
	InviteAnonymous int = 1 //匿名分享
)

const (
	InviteNormal int = 0
	Invite1018   int = 1
	InviteV3     int = 2
)

type AccountTask int

const (
	AccountTaskRegister AccountTask = 1
	AccountTaskLogin    AccountTask = 2
	AccountTaskApply    AccountTask = 3
	AccountTaskRepay    AccountTask = 4
)

type AccountTaskStatus int

const (
	AccountTaskStatusCreate AccountTaskStatus = 1
	AccountTaskStatusRun    AccountTaskStatus = 2
	AccountTaskStatusDone   AccountTaskStatus = 3
)

var AccountTaskMap = map[AccountTask]map[AccountTaskStatus]string{
	AccountTaskRegister: {
		AccountTaskStatusDone: "注册成功",
	},
	AccountTaskLogin: {
		AccountTaskStatusDone: "登录成功",
	},
	AccountTaskApply: {
		AccountTaskStatusDone: "申请成功",
	},
	AccountTaskRepay: {
		AccountTaskStatusDone: "还款成功",
	},
}
