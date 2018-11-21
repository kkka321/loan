package voip

const (
	VoipCacheTokenKey = "voip_access_token"
	VoipTokenExpire   = 12 * 60 * 60 * 1000

	AuthLoginApi     = "App.Sip_Auth.Login"              // 登录验证Api
	SipCallStatusApi = "App.Sip_Status.GetSipCallStatus" // 分机通话状态Api
	SipNumberInfoApi = "App.Sip_Sipnum.GetSipnumberInfo" // 分机管理Api
	MakeCallApi      = "App.Sip_Call.MakeCall"           // 发起呼叫Api
	BillApi          = "App.Sip_Cdr.GetBill"             // 获取话单Api
	RecodeFileApi    = "App.Sip_Cdr.GetRecodeFile"       // 获取下载录音地址Api

	VoipWhiteListSetName = "set:voip-white-list" // voip白名单redis中的集合名称
)

type SipNumberInfoStatus int

var (
	SipNumberInfoEnable  SipNumberInfoStatus = 1 // 启用分机
	SipNumberInfoDisable SipNumberInfoStatus = 2 // 禁用分机
	SipNumberInfoAll     SipNumberInfoStatus = 3 // 全部分机
)

const (
	// 呼叫方向。reverse则先呼B(客户)侧再呼A(员工分机)侧，positive则相反
	VoipMakeCallMethodReverse  = "reverse"
	VoipMakeCallMethodPositive = "positive" // 使用该项

	// 双呼。yes则是，no则否
	VoipMakeDoubleCallYes = "yes"
	VoipMakeDoubleCallNo  = "no" // 使用该项
)

const (
	// 响应返回码
	VoipRespRetSuccessed int = 200 // 成功

	// 响应状态码
	VoipStatusSuccessed int = 0
	VoipStatusFailed    int = 1
)

// 提供给第三方的回调相关数据
const (

	// 消息通知时的“呼叫方向”数据
	VoipCallOutStr = "callout"
	VoipCallInStr  = "callin"

	// 通话记录消息通知后，数据响应（回调响应）
	VoipBillMessageSuccess = "1"
	VoipBillMessageFail    = "0"

	// 1：获取未查询过的记录(默认)，2：获取已查询过的记录，3：获取全部记录
	VoipSyncflagUnQuery int = 1
	VoipSyncflagQuery   int = 2
	VoipSyncflagAll     int = 3
)

const (
	ContactMobileIsBlank = "联系人通讯方式为空"
	GetTicketInfoFail    = "获取工单信息失败"
	TicketUnAssign       = "工单未分配"
	NotAssignExtension   = "员工未分配分机号"
	GetSipStatusFail     = "获取分机状态失败"
	InsertCallRecordFail = "插入分机通话记录失败"
	SendCallRequestFail  = "呼叫请求发送失败"
	Calling              = "正在呼叫中"

	ParamsError          = "请求参数错误"
	GetTokenFail         = "获取token失败"
	RequestFail          = "请求发送失败"
	GetSipCallStatusFail = "获取分机通话状态失败"
	GetSipNumberInfoFail = "获取分机信息失败"
	MakeCallFail         = "分机呼叫失败"
	GetCallListFail      = "获取通话详单失败"
	GetRecordFileURLFail = "获取录音文件下载地址失败"
	HitVoipWhiteList     = "此电话号码不允许做外呼操作，如有疑问，请询问leader"
)

// 数据库存储的“电话拨打状态”
const (
	DBCallFail    int = 0
	DBCallSuccess int = 1
)

var VoipCallDialStatusMap = map[int]string{
	DBCallFail:    "未接通",
	DBCallSuccess: "接通",
}

// 数据库存储的“呼叫方向”
const (
	DBCallOutInt int = 0
	DBCallInInt  int = 1
)

var VoipCallDirectionMap = map[int]string{
	DBCallOutInt: "呼出",
	DBCallInInt:  "呼入",
}

const (
	// 获取通话详单时的参数callmethod设置值
	VoipCallMethodAll           int = 0 // 全部
	VoipCallMethodSipMutualCall int = 1 // 分机互拨
	VoipCallMethodSipDirectCall int = 2 // 分机直拨
	VoipCallMethodSipCall       int = 3 // api 呼叫
	VoipCallMethodDoubleCall    int = 4 // 双呼

	// 5 是我们自己定义的,和第三方没有关系,标识"还款提醒/催收"记录中的"手工"呼叫
	VoipCallManual int = 5
)

var VoipCallMethodMap = map[int]string{
	VoipCallMethodAll:           "api呼叫",
	VoipCallMethodSipMutualCall: "分机直拨",
	VoipCallMethodSipDirectCall: "分机直拨",
	VoipCallMethodSipCall:       "api呼叫",
	VoipCallMethodDoubleCall:    "分机直拨",

	VoipCallManual: "手工",
}

// 返回呼叫方法值
func GetVoipCallMethodVal(voipCallMethod int) string {
	return VoipCallMethodMap[voipCallMethod]
}

// ret值
const (
	Voip_Ret_200 = 200 // 操作成功
	Voip_Ret_400 = 400 // 非法请求
	Voip_Ret_500 = 500 // 服务器错误
	Voip_Ret_600 = 600 // token无效
	Voip_Ret_601 = 601 // appid未授权
	Voip_Ret_602 = 602 // appid授权已到期
	Voip_Ret_603 = 603 // 模块未授权
	Voip_Ret_604 = 604 // 非法IP访问
)

var voipRetMap = map[int]string{
	Voip_Ret_200: "操作成功",
	Voip_Ret_400: "非法请求",
	Voip_Ret_500: "服务器错误",
	Voip_Ret_600: "token无效",
	Voip_Ret_601: "appid未授权",
	Voip_Ret_602: "appid授权已到期",
	Voip_Ret_603: "模块未授权",
	Voip_Ret_604: "非法IP访问",
}

// 返回ret状态值
func GetVoipRetVal(voipRet int) string {
	return voipRetMap[voipRet]
}

//分机状态
const (
	Sip_Status_1001 = 1001 // 服务器连接失败
	Sip_Status_1002 = 1002 // 操作异常，一般为校验异常
	Sip_Status_1003 = 1003 // 操作失败，一般是授权失败、注销失败、命令发送失败、服务器器连接异常等
	Sip_Status_1010 = 1010 // 分机异常，可能是新加的分机,需要重新登陆获取新token
	Sip_Status_1011 = 1011 // 非法分机，非本公司所有
	Sip_Status_1012 = 1012 // 分机不存在
	Sip_Status_1013 = 1013 // 分机已停用
	Sip_Status_1014 = 1014 // 分机未注册
	Sip_Status_1015 = 1015 // 分机不在通话中
	Sip_Status_1016 = 1016 // 分机已启用
	Sip_Status_1017 = 1017 // 分机已注册
	Sip_Status_1018 = 1018 // 号码已启用
	Sip_Status_1019 = 1019 // 号码已禁用
	Sip_Status_1020 = 1020 // 号码不存在
	Sip_Status_1021 = 1021 // 非法号码，非本公司所有
	Sip_Status_1024 = 1024 // 任务不存在
	Sip_Status_1025 = 1025 // 未开始
	Sip_Status_1026 = 1026 // 进行中
	Sip_Status_1027 = 1027 // 暂停
	Sip_Status_1028 = 1028 // 已结束
	Sip_Status_1029 = 1029 // 文件不存在
	Sip_Status_1031 = 1031 // 策略设置异常
	Sip_Status_1032 = 1032 // 违反策略规则
	Sip_Status_1033 = 1033 // 无策略资源
)

var tagSipStatusMap = map[int]string{
	Sip_Status_1012: "分机不存在",
	Sip_Status_1013: "分机已停用",
	Sip_Status_1014: "分机未注册",
	Sip_Status_1015: "分机不在通话中",
	Sip_Status_1016: "分机已启用",
	Sip_Status_1017: "分机已注册",
	Sip_Status_1018: "号码已启用",
	Sip_Status_1019: "号码已禁用",
}

// 返回呼叫状态值
func GetSipStatusVal(sipStatus int) string {
	return tagSipStatusMap[sipStatus]
}

//分机通话状态
const (
	Call_Status_0    = -1
	Call_Status_1201 = 1201 //空闲
	Call_Status_1202 = 1202 //振铃
	Call_Status_1203 = 1203 //摘机
	Call_Status_1204 = 1204 //通话中
	Call_Status_1208 = 1208 //其他
	Call_Status_1209 = 1209 //⾮通话中不能语⾳评分
	Call_Status_1210 = 1210 //队列异常
	Call_Status_1211 = 1211 //⾮法队列
	Call_Status_1212 = 1212 //未接听
	Call_Status_1213 = 1213 //等待中
	Call_Status_1214 = 1214 //接收中
	Call_Status_1215 = 1215 //已接听
	Call_Status_1216 = 1216 //拒接
	Call_Status_1217 = 1217 //暂停
)

var tagCallStatusMap = map[int]string{
	Call_Status_0:    "请选择",
	Sip_Status_1014:  "分机未注册",
	Call_Status_1201: "空闲",
	Call_Status_1202: "振铃",
	Call_Status_1203: "摘机",
	Call_Status_1204: "通话中",
	Call_Status_1208: "其他",
	Call_Status_1209: "⾮通话中不能语⾳评分",
	Call_Status_1210: "队列异常",
	Call_Status_1211: "⾮法队列",
	Call_Status_1212: "未接听",
	Call_Status_1213: "等待中",
	Call_Status_1214: "接收中",
	Call_Status_1215: "已接听",
	Call_Status_1216: "拒接",
	Call_Status_1217: "暂停",
}

// 返回标签的map
func TagCallStatusMap() map[int]string {
	return tagCallStatusMap
}

// 返回呼叫状态值
func GetCallStatusVal(hangupStatus int) string {
	return tagCallStatusMap[hangupStatus]
}

// 挂机原因/挂机方向
const (
	Sip_Hangup_10001 = 10001 // 正常挂断
	Sip_Hangup_10002 = 10002 // 呼叫取消
	Sip_Hangup_10003 = 10003 // 拒绝接听
	Sip_Hangup_10004 = 10004 // 外呼通道线路失败
	Sip_Hangup_10005 = 10005 // 用户操作未接听
	Sip_Hangup_10006 = 10006 // 用户忙
	Sip_Hangup_10007 = 10007 // 服务器器端挂断
	Sip_Hangup_10008 = 10008 // 分机未注册
	Sip_Hangup_10009 = 10009 // 目标不可达
	Sip_Hangup_10011 = 10011 // 定时器超时
	Sip_Hangup_10012 = 10012 // 呼入时回调接口错误
	Sip_Hangup_10013 = 10013 // 分机不存在
	Sip_Hangup_10014 = 10014 // 未发现
	Sip_Hangup_10015 = 10015 // 请求超时
	Sip_Hangup_10016 = 10016 // 无人接听
	Sip_Hangup_10017 = 10017 // 呼叫失效
	Sip_Hangup_10019 = 10019 // 归属地未知
	Sip_Hangup_10020 = 10020 // 其他原因
	Sip_Hangup_10024 = 10024 // 错误请求
	Sip_Hangup_10025 = 10025 // 呼叫被禁止
	Sip_Hangup_10027 = 10027 // 号码被改变
	Sip_Hangup_10028 = 10028 // 呼叫拦截
	Sip_Hangup_10031 = 10031 // 未知
	Sip_Hangup_10040 = 10040 // 主叫挂机
	Sip_Hangup_10041 = 10041 // 被叫挂机
)

var VoipSipHangupMap = map[int]string{
	Sip_Hangup_10001: "正常挂断",
	Sip_Hangup_10002: "呼叫取消",
	Sip_Hangup_10003: "拒绝接听",
	Sip_Hangup_10004: "外呼通道线路失败",
	Sip_Hangup_10005: "用户操作未接听",
	Sip_Hangup_10006: "用户忙",
	Sip_Hangup_10007: "服务器器端挂断",
	Sip_Hangup_10008: "分机未注册",
	Sip_Hangup_10009: "目标不可达",
	Sip_Hangup_10011: "定时器超时",
	Sip_Hangup_10012: "呼入时回调接口错误",
	Sip_Hangup_10013: "分机不存在",
	Sip_Hangup_10014: "未发现",
	Sip_Hangup_10015: "请求超时",
	Sip_Hangup_10016: "无人接听",
	Sip_Hangup_10017: "呼叫失效",
	Sip_Hangup_10019: "归属地未知",
	Sip_Hangup_10020: "其他原因",
	Sip_Hangup_10024: "错误请求",
	Sip_Hangup_10025: "呼叫被禁止",
	Sip_Hangup_10027: "号码被改变",
	Sip_Hangup_10028: "呼叫拦截",
	Sip_Hangup_10031: "未知",
	Sip_Hangup_10040: "主叫挂机",
	Sip_Hangup_10041: "被叫挂机",
}

// 返回挂机原因/挂机方向
func GetSipHangeupVal(hangupStatus int) string {
	return VoipSipHangupMap[hangupStatus]
}

//分机是否启用
const (
	Is_Use_0    = -1
	Is_Use_1013 = 0 //分机已停用
	Is_Use_1016 = 1 //分机已启用
)

var tagExtIsUseMap = map[int]string{
	Is_Use_0:    "请选择",
	Is_Use_1013: "分机已停用",
	Is_Use_1016: "分机已启用",
}

// 返回标签的map
func TagExtIsUseMap() map[int]string {
	return tagExtIsUseMap
}

//分机分配状态
const (
	ExtStatus_No = -1
	ExtStatus_0  = 0 //未分配
	ExtStatus_1  = 1 //已分配
)

var tagExtStatusMap = map[int]string{
	ExtStatus_No: "请选择",
	ExtStatus_0:  "未分配",
	ExtStatus_1:  "已分配",
}

// 返回标签的map
func TagExtStatusMap() map[int]string {
	return tagExtStatusMap
}

// voip白名单前缀
var voipWhiteListPre = map[int]string{
	1:  "62815",
	2:  "62816",
	3:  "62814",
	4:  "628588",
	5:  "62855",
	6:  "62817",
	7:  "62818",
	8:  "62819",
	9:  "62811",
	10: "0815",
	11: "0816",
	12: "0814",
	13: "08588",
	14: "0855",
	15: "0817",
	16: "0818",
	17: "0819",
	18: "0811",
}

// voip白名单包含字符
var voipWhiteListContain = map[int]string{
	1:  "000",
	2:  "111",
	3:  "222",
	4:  "333",
	5:  "444",
	6:  "555",
	7:  "666",
	8:  "777",
	9:  "888",
	10: "999",
	11: "123",
	12: "456",
	13: "789",
	14: "8080",
	15: "5050",
	16: "6060",
	17: "1212",
	18: "8081",
	19: "1010",
	20: "8008",
	21: "0880",
}
