package types

const SMS_CODE_LEN = 4     //短信验证码长度
const SMS_EXPIRES = 600000 //短信失效时间60000*10 10分钟

const (
	// VerifyCodeSendFailed 验证码发送失败
	VerifyCodeSendFailed = iota
	// VerifyCodeUnchecked 未验证
	VerifyCodeUnchecked
	// VerifyCodeChecked 已验证
	VerifyCodeChecked
	// VerifyCodeCheckFailed 验证失败, 多次尝试之后无果
	VerifyCodeCheckFailed
)

// SmsVerifyCodeStatusMap 验证状态 map
var SmsVerifyCodeStatusMap = map[int]string{
	VerifyCodeSendFailed:  "发送失败",
	VerifyCodeUnchecked:   "未验证",
	VerifyCodeChecked:     "已验证",
	VerifyCodeCheckFailed: "验证失败",
}

type AuthCodeType int

const (
	AuthCodeTypeText  AuthCodeType = 1 // 短信验证码/文本验证码
	AuthCodeTypeVoice AuthCodeType = 2 // 语音验证码
)

var AuthCodeTypeEnMap = map[AuthCodeType]string{
	AuthCodeTypeText:  "text",
	AuthCodeTypeVoice: "voice",
}

var authCodeTypeChMap = map[AuthCodeType]string{
	AuthCodeTypeText:  "短信验证码",
	AuthCodeTypeVoice: "语音验证码",
}

// AuthCodeTypeMap 返回 auth Code 的类型
func AuthCodeTypeMap() map[AuthCodeType]string {
	return authCodeTypeChMap
}
