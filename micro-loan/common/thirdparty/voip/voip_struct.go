package voip

type VoipRet struct {
	Ret int `json:"ret"`
}

// 登录认证响应
type AuthLoginResponse struct {
	Ret  int    `json:"ret"`
	Msg  string `json:"msg"`
	Data struct {
		Status  int    `json:"status"`  // 状态码， 0: 成功，1: 失败
		Desc    string `json:"desc"`    // 状态码描述
		ReqTime int64  `json:"reqtime"` // 请求时间
		RspTime int64  `json:"rsptime"` // 响应时间
		Errors  struct {
			Code    int    `json:"code"`    // 错误码
			CodeMsg string `json:"codemsg"` // 错误信息
		} `json:"errors"`
		Result struct {
			CompanyCode string `json:"companycode"` // 公司编号
			CompanyName string `json:"companyname"` // 公司名称
			Token       string `json:"token"`       // token，有效期12小时
			AuthTime    string `json:"authtime"`    // 授权到期时间
			AuthModel   string `json:"authmodel"`   // 授权模块
			IPWhite     string `json:"ipwhite"`     // 授权访问的IP
		} `json:"result"`
	} `json:"data"`
}

// 分机状态数据
type SipCallStatusResult struct {
	ExtNumber    string `json:"extnumber"`    // 分机号
	Status       int    `json:"status"`       // 分机状态
	Caller       string `json:"caller"`       // 主叫号码
	Callee       string `json:"callee"`       // 被叫号码
	DisNumber    string `json:"disnumber"`    // 外显号码
	Direction    string `json:"direction"`    // 呼叫方向  callout为呼出
	Channel      string `json:"channel"`      // 当前通道
	IdleDuration int64  `json:"idleduration"` // 空闲时长，单位秒
	CallDuration int    `json:"callduration"` // 通话时长，单位秒
}

// 获取分机状态响应
type SipCallStatusResponse struct {
	Ret  int    `json:"ret"`
	Msg  string `json:"msg"`
	Data struct {
		Status  int    `json:"status"`  // 状态码， 0: 成功，1: 失败
		Desc    string `json:"desc"`    // 状态码描述
		ReqTime int64  `json:"reqtime"` // 请求时间
		RspTime int64  `json:"rsptime"` // 响应时间
		Errors  struct {
			Code    int    `json:"code"`    // 错误码
			CodeMsg string `json:"codemsg"` // 错误信息
		} `json:"errors"`
		Result []SipCallStatusResult `json:"result"` // 成功时，返回的结果集
	} `json:"data"`
}

// 分机号数据
type SipNumberInfoResult struct {
	ExtNumber string `json:"extnumber"` // 分机号
	Password  string `json:"password"`  // 分机密码
	Status    int    `json:"status"`    // 分机状态
	DisNumber string `json:"disnumber"` // 外显号码
}

// 获取所有分机号
type SipNumberInfoResponse struct {
	Ret  int    `json:"ret"`
	Msg  string `json:"msg"`
	Data struct {
		Status  int    `json:"status"`  // 状态码， 0: 成功，1: 失败
		Desc    string `json:"desc"`    // 状态码描述
		ReqTime int64  `json:"reqtime"` // 请求时间
		RspTime int64  `json:"rsptime"` // 响应时间
		Errors  struct {
			Code    int    `json:"code"`    // 错误码
			CodeMsg string `json:"codemsg"` // 错误信息
		} `json:"errors"`
		Result []SipNumberInfoResult `json:"result"` // 成功时，返回的结果集
	} `json:"data"`
}

// sip呼叫请求
type MakeCallRequest struct {
	Token      string `json:"token"`      // token
	ExtNumber  string `json:"extnumber"`  // 分机号
	DestNumber string `json:"destnumber"` // 目标号码

	DisNumber  string `json:"disnumber"`  // 主叫号码
	CallMethod string `json:"callMethod"` // 呼叫方向
	DoubleCall string `json:"doublecall"` // 双呼
	UserID     string `json:"userid"`     // 用户ID（分配人员ID）
	MemberID   string `json:"memberid"`   // 通话记录ID
	Ripeness   string `json:"chengshudu"`
	CustomUUID string `json:"customuuid"` // 借款ID
}

// sip呼叫响应
type MakeCallResponse struct {
	Ret  int    `json:"ret"`
	Msg  string `json:"msg"`
	Data struct {
		Status  int    `json:"status"`  // 状态码， 0: 成功，1: 失败
		Desc    string `json:"desc"`    // 状态码描述
		ReqTime int64  `json:"reqtime"` // 请求时间
		RspTime int64  `json:"rsptime"` // 响应时间
		Errors  struct {
			Code    int    `json:"code"`    // 错误码
			CodeMsg string `json:"codemsg"` // 错误信息
		} `json:"errors"`
	} `json:"data"`
}

// sip通话列表详单请求
type CallListRequest struct {
	Token     string `json:"token"`     // token
	StartTime string `json:"starttime"` // 开始时间
	EndTime   string `json:"endtime"`   // 结束时间

	Syncflag     int    `json:"syncflag"`     // 1：获取未查询过的记录(默认)，2：获取已查询过的记录，3：获取全部记录
	ExtNumber    string `json:"extnumber"`    // 分机号
	DestNumber   string `json:"destnumber"`   // 目标号码
	UserID       string `json:"userid"`       // 用户ID（分配人员ID）
	MemberID     string `json:"memberid"`     // 通话记录ID
	Ripeness     string `json:"chengshudu"`   // 成熟度
	CustomUUID   string `json:"customuuid"`   // 借款ID
	Direction    int    `json:"direction"`    // 呼叫方向
	CallMethod   int    `json:"callmethod"`   // 呼叫方法
	CurrentPage  int    `json:"currentpage"`  // 第几页
	ItemsPerPage int    `json:"itemsperpage"` // 每页数量
}

// 通话订单详情
type BillResult struct {
	Id              int64  `json:"id"`              // 话单id
	ExtNumber       string `json:"extnumber"`       // 分机号
	DestNumber      string `json:"destnumber"`      // 目标号码
	DisNumber       string `json:"displaynumber"`   // 主叫号码
	StartTime       string `json:"starttime"`       // 开始时间
	AnswerTime      string `json:"answertime"`      // 接通时间
	EndTime         string `json:"endtime"`         // 结束时间
	Duration        int    `json:"duration"`        // 接通前等待时长
	Billsec         int    `json:"billsec"`         // 呼叫时长
	Direction       string `json:"direction"`       // 呼叫方向 callin为呼入，callout呼出
	CallMethod      int    `json:"callmethod"`      // 呼叫方法 1：分机互拨，2：分机直拔，3：API呼叫，4：双呼
	UserID          string `json:"userid"`          // 用户ID（分配人员ID）
	MemberID        string `json:"memberid"`        // 通话记录ID
	Ripeness        string `json:"chengshudu"`      // 成熟度
	CustomUUID      string `json:"customuuid"`      // 借款ID
	RecordFileName  string `json:"recordfilename"`  // 录音文件名
	DownloadIP      string `json:"downloadip"`      // 录音服务器地址
	HangupDirection int    `json:"hangupdirection"` // 挂机方向
	HangupCause     int    `json:"hangupcause"`     // 挂机原因
}

// 分机号数据
type CallListResult struct {
	TatalItems   int64        `json:"totalitems"`   // 分机号
	CurrentPage  int          `json:"currentpage"`  // 分机密码
	ItemsPerPage int          `json:"itemsperpage"` // 分机状态
	Bills        []BillResult `json:"bills"`
}

// sip通话列表详单响应
type CallListResponse struct {
	Ret  int    `json:"ret"`
	Msg  string `json:"msg"`
	Data struct {
		Status  int    `json:"status"`  // 状态码， 0: 成功，1: 失败
		Desc    string `json:"desc"`    // 状态码描述
		ReqTime int64  `json:"reqtime"` // 请求时间
		RspTime int64  `json:"rsptime"` // 响应时间
		Errors  struct {
			Code    int    `json:"code"`    // 错误码
			CodeMsg string `json:"codemsg"` // 错误信息
		} `json:"errors"`
		Result CallListResult `json:"result"` // 成功时，返回的结果集
	} `json:"data"`
}

// 录音文件数据
type RecodeFileResult struct {
	DownURL     string `json:"downurl"`     // 录音文件下载地址
	ExpiredTime string `json:"expiredtime"` // 录音文件失效时间
}

// 获取录音文件响应
type RecodeFileResponse struct {
	Ret  int    `json:"ret"`
	Msg  string `json:"msg"`
	Data struct {
		Status  int    `json:"status"`  // 状态码， 0: 成功，1: 失败
		Desc    string `json:"desc"`    // 状态码描述
		ReqTime int64  `json:"reqtime"` // 请求时间
		RspTime int64  `json:"rsptime"` // 响应时间
		Errors  struct {
			Code    int    `json:"code"`    // 错误码
			CodeMsg string `json:"codemsg"` // 错误信息
		} `json:"errors"`
		Result RecodeFileResult `json:"result"` // 成功时，返回的结果集
	} `json:"data"`
}

// 通话详单数据推送
type BillMessage struct {
	MemberID        string `json:"memberid"`        // 通话记录ID
	CallId          int64  `json:"id"`              // 话单id
	ExtNumber       string `json:"extnumber"`       // 分机号
	DestNumber      string `json:"destnumber"`      // 目标号码
	DisNumber       string `json:"disnumber"`       // 主叫号码
	StartTime       string `json:"starttime"`       // 开始时间
	AnswerTime      string `json:"answertime"`      // 接通时间
	EndTime         string `json:"endtime"`         // 结束时间
	Duration        int64  `json:"duration"`        // 接通前等待时长
	BillSec         int64  `json:"billsec"`         // 通话时长
	UserID          string `json:"crmid"`           // 员工工号（分配人员ID）
	Type            string `json:"type"`            // 呼叫方向 callin为呼入，callout呼出
	CallMethod      int    `json:"callmethod"`      // 呼叫方法 1：分机互拨，2：分机直拔，3：API呼叫，4：双呼
	Ripeness        string `json:"chengshudu"`      // 成熟度
	CustomUUID      string `json:"customuuid"`      // 借款ID
	RecordFileName  string `json:"recordfilename"`  // 录音文件名
	DownloadIP      string `json:"downloadip"`      // 录音服务器地址
	HangupDirection int    `json:"hangupdirection"` // 挂机方向
	HangupCause     int    `json:"hangupcause"`     // 挂机原因
	CompanyCode     int64  `json:"companycode"`     // 公司ID
	UUID            string `json:"uuid"`            // 通话uuid
	UserKey         string `json:"userkey"`         // 用户按键
}
