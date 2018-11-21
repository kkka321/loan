package tongdun

import (
	"micro-loan/common/cerror"
	"micro-loan/common/types"
	"time"

	"github.com/astaxie/beego/logs"
	"micro-loan/common/models"
)

// Channel ...
type Channel struct {
	Name                 string
	Code                 string
	IdentifyingCodeCount int // 验证码的位数 XL、Indonesia这两家运营商的短信验证码为6位，而Tolksen的短信验证码为4位 。
}

// ChannelCodeMap ...
var ChannelCodeMap = map[string]*Channel{
	"0811": &Channel{Code: "102001", Name: "Telkosmel", IdentifyingCodeCount: 4},
	"0812": &Channel{Code: "102001", Name: "Telkosmel", IdentifyingCodeCount: 4},
	"0813": &Channel{Code: "102001", Name: "Telkosmel", IdentifyingCodeCount: 4},
	"0821": &Channel{Code: "102001", Name: "Telkosmel", IdentifyingCodeCount: 4},
	"0822": &Channel{Code: "102001", Name: "Telkosmel", IdentifyingCodeCount: 4},
	"0823": &Channel{Code: "102001", Name: "Telkosmel", IdentifyingCodeCount: 4},
	"0852": &Channel{Code: "102001", Name: "Telkosmel", IdentifyingCodeCount: 4},

	// "0817": &Channel{Code: "102002", Name: "XL", IdentifyingCodeCount: 6},
	// "0859": &Channel{Code: "102002", Name: "XL", IdentifyingCodeCount: 6},
	// "0818": &Channel{Code: "102002", Name: "XL", IdentifyingCodeCount: 6},
	// "0819": &Channel{Code: "102002", Name: "XL", IdentifyingCodeCount: 6},
	// "0877": &Channel{Code: "102002", Name: "XL", IdentifyingCodeCount: 6},
	// "0878": &Channel{Code: "102002", Name: "XL", IdentifyingCodeCount: 6},

	"0814": &Channel{Code: "102003", Name: "Indoset", IdentifyingCodeCount: 6},
	"0815": &Channel{Code: "102003", Name: "Indoset", IdentifyingCodeCount: 6},
	"0816": &Channel{Code: "102003", Name: "Indoset", IdentifyingCodeCount: 6},
	"0855": &Channel{Code: "102003", Name: "Indoset", IdentifyingCodeCount: 6},
	"0856": &Channel{Code: "102003", Name: "Indoset", IdentifyingCodeCount: 6},
}

const (
	TongdunRetUnknow    = 1    // 未知错误 结束爬取,请反馈错误信息给技术支支持
	TongdunRetWorking   = 100  // 爬取任务正在处理理,请稍后查询结束 调用用登录验证接口口查询任务状态
	TongdunRetWaitInput = 105  // 请输入入手手机验证码 等待用用户输入入,继续调用用接口口提交验证码
	TongdunRetReqFail   = 108  // 请求验证码失败 调用用重试验证码接口口,重新发送
	TongdunRetUsrPswErr = 112  // 账号或密码错误 检查账号密码后,调用用INIT阶段的爬取API
	TongdunRetLogFail   = 113  // 登录失败,请稍后再试 结束爬取,检查账号密码是否可用用后重试
	TongdunRetInRain    = 124  // 手手机验证码错误或过期 调用用重发验证码接口口
	TongdunRetTimeLimit = 126  // 请求验证码时间受限 等待一一分钟,调用用重发验证码接口口
	TongdunRetSubmitted = 137  // 任务已成功提交 完成登录验证,等待爬取任务回调通知
	TongdunRetErrFmt    = 180  // 号码格式错误 结束爬取 魔盒会做格式修正处理理
	TongdunRetUnRegit   = 181  // 手手机号未注册,请注册你的手手机号 结束爬取
	TongdunRetLocked    = 182  // 账户已被锁定 结束爬取
	TongdunRetTimeOut   = 2006 // 任务已超时 结束爬取,稍后重试
	TongdunRetDone      = 2007 // 任务已完成,请通过查询接口口查询结果 调用用查询任务结果接口口
)

// GetChannelByMobile 根据手机号获取type类型
func GetChannelByMobile(mobile string) (channelType, code string, count int) {
	channelType = ""
	code = ""
	if len(mobile) <= 4 {
		return
	}

	prefix := mobile[0:4]

	chann := ChannelCodeMap[prefix]
	if chann != nil {
		code = chann.Code
		channelType = IDYYSChannelType
		count = chann.IdentifyingCodeCount
	}
	return
}

// AcquireCode 获取验证码
func AcquireCode(accountID int64, taskID string, mobile string) (code cerror.ErrCode, err error) {
	// accountBase := models.OneAccountBaseByPkId(accountId)

	params := map[string]interface{}{
		"task_id":      taskID,
		"user_name":    mobile,
		"task_stage":   "INIT",
		"request_type": "submit",
		"login_type":   "0",
	}

	_, idCheckData, err := Request(accountID, AcquireCodeURL, params, map[string]interface{}{})

	//请求失败重新请求
	if TongdunRetReqFail == idCheckData.Code {
		logs.Warn("[AcquireCode] 第一次获取验证码失败，重新获取 ")
		_, idCheckData, err = Request(accountID, RetryURL, params, map[string]interface{}{})
	}

	// 返回100 继续请求  尝试9次后返回 一共等待45秒
	for i := 0; TongdunRetWorking == idCheckData.Code && i < types.OperatorMaxAttmpTimes; i++ {
		logs.Warn("[AcquireCode] 正在执行任务，查询任务状态 ")
		time.Sleep(time.Second * time.Duration(i+1))
		idCheckData, err = QueryTask(accountID, taskID)
	}

	updateIdCheckData(idCheckData, IsMatchAcquire, SourceQueryTask)
	if TongdunRetWaitInput == idCheckData.Code {
		logs.Info("[AcquireCode] 获取验证码成功:%#v", idCheckData)
		code = cerror.CodeSuccess
	} else {
		logs.Warn("[AcquireCode] 获取验证码失败，返回信息:%#v", idCheckData)
		code = cerror.TongdunAcquireCodeFail
	}
	return
}

// VerifyCode 验证客户输入的验证码
func VerifyCode(accountID int64, codeVerify, taskID string, mobile string) (code cerror.ErrCode, err error) {

	params := map[string]interface{}{
		"task_id":      taskID,
		"sms_code":     codeVerify,
		"task_stage":   "SEND_SMS",
		"request_type": "submit",
	}
	_, idCheckData, err := Request(accountID, AcquireCodeURL, params, map[string]interface{}{})
	logs.Warn("[VerifyCode] 验证手机验证码 第一次返回信息:%#v", idCheckData)

	// 返回100 继续请求 尝试3次后返回
	for i := 0; TongdunRetWorking == idCheckData.Code && i < types.OperatorMaxAttmpTimes; i++ {
		logs.Warn("[VerifyCode] 正在执行任务，查询任务状态 ")
		time.Sleep(time.Millisecond * 1500)
		idCheckData, err = QueryTask(accountID, taskID)
	}

	updateIdCheckData(idCheckData, IsMatchVerify, SourceQueryTask)
	if TongdunRetSubmitted == idCheckData.Code ||
		TongdunRetDone == idCheckData.Code {
		logs.Warn("[VerifyCode] 验证手机验证码成功")
		code = cerror.CodeSuccess
	} else {
		logs.Warn("[VerifyCode] 验证手机验证码失败，返回信息:%#v", idCheckData)
		code = cerror.TongdunVerifyCodeFail
	}

	return
}

// GetChannelCodeByTypeAndMobile 根据不同的通道类型获得通道码
func GetChannelCodeByTypeAndMobile(channelType, mobile string) (channelCode string) {
	switch channelType {
	case IDYYSChannelType:
		{
			_, channelCode, _ = GetChannelByMobile(mobile)
			if "" == channelCode {
				logs.Warn("[GetChannelCodeByTypeAndMobile] mobile  channelType :%s mobile:%s", channelType, mobile)
			}
		}
	case IDGoJekChannelType:
		{
			channelCode = ChannelCodeGoJek
		}
	default:
		{
			logs.Warn("[GetChannelCodeByTypeAndMobile] unknow channelType :%s mobile:%s", channelType, mobile)
		}
	}
	return
}

func updateIdCheckData(idCheckData IdentityCheckCreateTask, matchCodeStat string, source int64) {
	accountTongdun, err := models.GetOneByCondition("task_id", idCheckData.TaskID)
	if err != nil {
		logs.Error("[updateIdCheckData] GetOneByCondition err:%v idCheckData:%#v", err, idCheckData)
		return
	}
	if accountTongdun.CheckCode != IDCheckCodeYes {

		accountTongdun.CheckCode = idCheckData.Code
		accountTongdun.Message = idCheckData.Message
		accountTongdun.IsMatch = matchCodeStat
		accountTongdun.Source = source
		models.UpdateTongdun(accountTongdun, "check_code", "message", "is_match", "source")
	}
}
