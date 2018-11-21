package cerror

import (
	"encoding/json"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/tools"
)

// ErrCode represents a specific error type in a error class.
// Same error code can be used in different error classes.
type ErrCode int

type ApiResponse struct {
	Code ErrCode     `json:"code"`
	Data interface{} `json:"data"`
}

type ApiResponseV2 struct {
	Code    ErrCode     `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type APIEntrustResponse struct {
	Code   ErrCode     `json:"code"`
	Result interface{} `json:"result"`
}

type APIAdminResponse struct {
	Code   ErrCode     `json:"code"`
	Result interface{} `json:"result"`
}

const (
	// CodeUnknown is for errors of unknown reason.
	CodeUnknown ErrCode = -1

	CodeSuccess ErrCode = 0

	// 500 服务不可用,存在严重问题的特殊码
	ServiceIsDown ErrCode = 500111 // 服务完全不可用,后端服务发生严重问题,需要时间恢复,让客户提示用户稍后再试

	// 800800
	AppForceUpgrade ErrCode = 800800 // 客户端强制升级 印尼>=12起支持

	// 4xx 接口相关
	RequestExceedsLimit           ErrCode = 400109 // 请求量超出限制（用于列表）
	LostRequiredParameters        ErrCode = 400110 // 缺少必要参数
	SignatureVerifyFail           ErrCode = 400111 // 验证签名失败
	LostAccessToken               ErrCode = 400112 // 缺少 token
	AccessTokenExpired            ErrCode = 400113 // token 过期
	InvalidAccessToken            ErrCode = 400114 // 无效的 token
	InvalidMobile                 ErrCode = 400115 // 无效的手机号
	InvalidAuthCode               ErrCode = 400116 // 无效的验证码
	RequestApiTooMore             ErrCode = 400117 // 请求接口过于频繁
	ApiNotFound                   ErrCode = 400118 // 接口不存在
	InvalidRequestData            ErrCode = 400119 // 无效的请求数据
	LimitStrategyMobile           ErrCode = 400120 // 手机号使用受到限制,请求验证码达到上限,目前是每种类型24小时内6次
	SMSServiceUnavailable         ErrCode = 400121 // 短信服务不可用
	ServiceUnavailable            ErrCode = 400122 // 服务不可用
	FileTypeUnsupported           ErrCode = 400123 // 文件类型不支持
	UploadResourceFail            ErrCode = 400124 // 上传资源操作失败
	PermissionDenied              ErrCode = 400125 // 操作系统文件权限不足
	InvalidAccount                ErrCode = 400126 // 无效账户
	MismatchRepeatLoan            ErrCode = 400127 // 不满足复贷条件
	ProductDoesNotExist           ErrCode = 400128 // 产品不存在
	UnsettledOrders               ErrCode = 400129 // 存在未结清订单
	InvalidParameterValue         ErrCode = 400130 // 无效的参数值
	CreateOrderFail               ErrCode = 400131 // 创建或确认订单失败
	IdentityBindRepeated          ErrCode = 400132 // 身份证号和手机号重复绑定
	OriginHandHeldIdPhotoNotExist ErrCode = 400133 // 原有手持照片不存在
	TwoImagesCompareError         ErrCode = 400134 // 照片比对错误(未识别人脸之类的)

	HandPhotoCheckLessThanDefine ErrCode = 400136 // 手持照片比对结果少于阈值(手持照片比对结果<0.7)
	IdentityVerifyNotPass        ErrCode = 400137 // 同盾，Advance身份检查都未通过
	OcrIdentifyError             ErrCode = 400138 // OCR识别错误
	CreateRollOrderFail          ErrCode = 400139 // 创建展期订单失败
	SMSRequestFrequencyTooHigh   ErrCode = 400140 // 获取短信请求过于频繁
	MobileHasRegistered          ErrCode = 400141 // 手机号已被注册
	MobileNotRegistered          ErrCode = 400142 // 手机号未被注册
	InvalidPassword              ErrCode = 400143 // 密码无效
	InvalidOldPassword           ErrCode = 400144 // 旧密码无效
	AccountLocked                ErrCode = 400145 // 账号被锁定
	AccountPasswordUnset         ErrCode = 400146 // 找回密码时，账号未设置密码
	OrderDoesNotExist            ErrCode = 400147 // order不存在
	PaymentGenerateErr           ErrCode = 400148 // 付款码不存在
	RollOrderNotSupport          ErrCode = 400149 // 付款码不存在
	VoiceAuthCodeServiceFail     ErrCode = 400150 // 语音验证码服务不可用
	RiskReCheckError             ErrCode = 400151 // 重新反欺诈审核失败
	LimitStrategyVoiceAuthCode   ErrCode = 400152 // 手机号使用受到限制,请求语音验证码达到上限,目前是每种类型24小时内6次
	ModifyBankFail               ErrCode = 400153 // 修改银行信息失败
	PaymentVoucherCode           ErrCode = 400154 // 上传凭证失败
	ModifyRepayBankAndGetVAFail  ErrCode = 400155 // 修改银行，并获取VA失败
	AccountGetVAFail             ErrCode = 400156 // 获取用户所有va失败
	OcrServiceError              ErrCode = 400157 // OCR服务调用失败

	TongdunIDCheckCreateTaskFail ErrCode = 400200 //同盾身份检查任务创建失败(身份证，姓名格式错误，多半是渠道信息有误)
	//NoMoreData 			   ErrCode = 400130 // 没有更多数据. 注意:并非接口真的出错了,而是服务端没有满足条件的数据

	TongdunAcquireCodeFail ErrCode = 400300 //同盾获取运营商授权验证码失败
	TongdunVerifyCodeFail  ErrCode = 400301 //同盾验证运营商授权验证码失败

	AdvertisementGetCodeFail ErrCode = 400401 //获取广告位图片失败

	BannerGetCodeFail ErrCode = 400501 //获取banner失败

	PopGetCodeFail      ErrCode = 400601 //获取pop window失败
	FloatingGetCodeFail ErrCode = 400602 //获取floating window失败

	AdPositionGetCodeFail ErrCode = 400701 //获取广告位失败
)

func BuildApiResponse(code ErrCode, data interface{}) ApiResponse {
	r := ApiResponse{
		Code: code,
		Data: "",
	}
	if code == CodeSuccess || code == AppForceUpgrade {
		jsonByte, err := json.Marshal(data)
		if err != nil {
			return r
		}

		// 打印响应体主数据,以供联调排查问题
		logs.Debug(">>> ResponseJSON:", string(jsonByte))

		encryptData, err := tools.AesEncryptCBC(string(jsonByte), tools.AesCBCKey, tools.AesCBCIV)
		if err != nil {
			return r
		}

		r.Data = encryptData
	} else {
		logs.Debug(">>> ResponseCode:", code, ", >>> ResponseJSON: {}")
	}

	return r
}

func BuildApiResponseV2(code ErrCode, message string, data interface{}) ApiResponseV2 {
	r := ApiResponseV2{
		Code:    code,
		Message: message,
		Data:    "",
	}
	if code == CodeSuccess || code == AppForceUpgrade {
		jsonByte, err := json.Marshal(data)
		if err != nil {
			return r
		}

		// 打印响应体主数据,以供联调排查问题
		logs.Debug(">>> ResponseJSON:", string(jsonByte))

		encryptData, err := tools.AesEncryptCBC(string(jsonByte), tools.AesCBCKey, tools.AesCBCIV)
		if err != nil {
			return r
		}

		r.Data = encryptData
	} else {
		logs.Debug(">>> ResponseCode:", code, ", >>> ResponseJSON: {}")
	}

	return r
}

func BuildEntrustApiResponse(code ErrCode, data interface{}) APIEntrustResponse {
	r := APIEntrustResponse{
		Code:   code,
		Result: "",
	}
	if code == CodeSuccess {
		// jsonByte, err := json.Marshal(data)
		// if err != nil {
		// 	return r
		// }
		// // 打印响应体主数据,以供联调排查问题
		// logs.Debug(">>> ResponseJSON:", string(jsonByte))
		r.Result = data //string(jsonByte)
	} else {
		logs.Debug(">>> ResponseCode:", code, ", >>> ResponseJSON: {}")
	}

	return r
}

func BuildAdminApiResponse(code ErrCode, data interface{}) APIAdminResponse {
	r := APIAdminResponse{
		Code:   code,
		Result: "",
	}
	if code == CodeSuccess {
		// jsonByte, err := json.Marshal(data)
		// if err != nil {
		// 	return r
		// }
		// // 打印响应体主数据,以供联调排查问题
		// logs.Debug(">>> ResponseJSON:", string(jsonByte))
		r.Result = data //string(jsonByte)
	} else {
		logs.Debug(">>> ResponseCode:", code, ", >>> ResponseJSON: {}")
	}

	return r
}
