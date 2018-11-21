package limit

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/i18n"
	"micro-loan/common/lib/redis/cache"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type SmsStrategyType int

const (
	SmsTimesTooMore     SmsStrategyType = 1
	SmsFrequencyTooHigh SmsStrategyType = 2
)

// 限制策略
var limitStrategyServiceMap = map[types.ServiceType]string{
	types.ServiceRequestLogin: "register-login",
	types.ServiceRepeatedLoan: "repeated-loan",
	types.ServiceRegister:     "register",
	types.ServiceLogin:        "login",
	types.ServiceFindPassword: "find-password",
	types.ServiceConfirmOrder: "confirm-order",
}

var limitStrategyMobile = map[string]int64{
	"frequency": 6,
	"interval":  86400000, // 银行级别的,24小时内只能试6次
	"coolTime":  60000,    // 两次短信的间隔最短时间
}

var limitStrategyPassword = map[string]int64{
	"notify":   3,        // 密码输错3次后，提示错误信息
	"lock":     6,        // 密码输错6次后，锁定用户
	"interval": 86400000, // 用户锁定后，24小时内自动解锁
}

type AuthCodeStrategy struct {
	Strategy        string
	Mobile          string
	ServiceType     types.ServiceType
	AuthCodeTypeVal types.AuthCodeType
}

func buildCacheKey(key string, serviceType types.ServiceType, authCodeType types.AuthCodeType) string {
	prefix := beego.AppConfig.String("limit_strategy_prefix")
	cKey := fmt.Sprintf("%s:%s:%s:%s", prefix, limitStrategyServiceMap[serviceType], types.AuthCodeTypeEnMap[authCodeType], key)
	logs.Debug("buildCacheKey -> cKey:", cKey)
	return cKey
}

// 返回为 true 说明中了限制策略
// 判断24小时内验证码次数
func commonStrategy(authCodeStrategy AuthCodeStrategy) bool {
	strategy := authCodeStrategy.Strategy
	mobile := authCodeStrategy.Mobile
	serviceType := authCodeStrategy.ServiceType
	authCodeType := authCodeStrategy.AuthCodeTypeVal

	// 考虑以后加更多限制策略的...
	if strategy != "mobile" {
		return false
	}

	// 想扩展,就扩展它
	var cfg = limitStrategyMobile

	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	cKey := buildCacheKey(mobile, serviceType, authCodeType)
	cValue, _ := cacheClient.Do("GET", cKey)
	if cValue == nil {
		cacheClient.Do("SET", cKey, "1", "PX", cfg["interval"])
	} else {
		// 好个坑
		value := string(cValue.([]byte))
		num, _ := tools.Str2Int64(value)
		num++
		//cacheClient.Do("SET", cKey, tools.Int642Str(num))
		cacheClient.Do("INCR", cKey)
		if num > cfg["frequency"] {
			logs.Warning("hit limit strategy, key:", cKey, ", value:", num)
			return true
		}
	}

	return false // 不使用限制策略
}

func MobileStrategy(mobile string, serviceType types.ServiceType, authCodeType types.AuthCodeType) bool {
	authCodeStrategy := AuthCodeStrategy{
		Strategy:        "mobile",
		Mobile:          mobile,
		ServiceType:     serviceType,
		AuthCodeTypeVal: authCodeType,
	}
	return commonStrategy(authCodeStrategy)
}

// 短信验证码频率检查
func smsFrequencyStrategy(authCodeStrategy AuthCodeStrategy) bool {
	strategy := authCodeStrategy.Strategy
	mobile := authCodeStrategy.Mobile
	serviceType := authCodeStrategy.ServiceType
	//authCodeType := authCodeStrategy.AuthCodeTypeVal

	// 考虑以后加更多限制策略的...
	if strategy != "mobile" {
		return false
	}

	var cfg = limitStrategyMobile

	o := new(models.SmsVerifyCode)
	smsVerifyCode, err := o.GetSmsCodeByPhoneAndServiceType(mobile, serviceType)
	curMill := tools.GetUnixMillis()
	timeDiff := curMill - smsVerifyCode.Ctime
	if err == nil && timeDiff <= cfg["coolTime"] {
		return true
	}

	return false
}

// (短信验证码/语音验证码)一天只能发6次，并且(短信验证码/语音验证码)发送间隔大于60秒
func MobileStrategyV2(mobile string, serviceType types.ServiceType, authCodeType types.AuthCodeType) (smsStrategy SmsStrategyType) {

	authCodeStrategy := AuthCodeStrategy{
		Strategy:        "mobile",
		Mobile:          mobile,
		ServiceType:     serviceType,
		AuthCodeTypeVal: authCodeType,
	}

	if smsFrequencyStrategy(authCodeStrategy) {
		smsStrategy = SmsFrequencyTooHigh
		return
	}

	if commonStrategy(authCodeStrategy) {
		smsStrategy = SmsTimesTooMore
		return
	}

	return
}

func BuildPwdCacheKey(mobile string) string {
	prefix := beego.AppConfig.String("limit_strategy_prefix")
	cKey := fmt.Sprintf("%s:%s:%s", prefix, "password", mobile)
	logs.Debug("BuildPwdCacheKey -> cKey:", cKey)

	return cKey
}

func ClearPwdCache(mobile string) {
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	cKey := BuildPwdCacheKey(mobile)
	cacheClient.Do("DEL", cKey)
}

func GetPwdCache(mobile string) (num int64) {
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	cKey := BuildPwdCacheKey(mobile)
	cValue, _ := cacheClient.Do("GET", cKey)
	if cValue == nil {
		return
	}

	value := string(cValue.([]byte))
	num, _ = tools.Str2Int64(value)

	return
}

func IsAccountLocked(mobile string) bool {
	var cfg = limitStrategyPassword
	if GetPwdCache(mobile) >= cfg["lock"] {
		return true
	}

	return false
}

// 密码登录错误次数限制
func PasswordStrategy(mobile string) (msgErr string) {

	var cfg = limitStrategyPassword

	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	cKey := BuildPwdCacheKey(mobile)
	cValue, _ := cacheClient.Do("GET", cKey)
	msgErr = i18n.GetMessageText(i18n.MsgPasswordErr)
	if cValue == nil {
		cacheClient.Do("SET", cKey, "1", "PX", cfg["interval"])
	} else {
		value := string(cValue.([]byte))
		num, _ := tools.Str2Int64(value)
		num++

		if num <= cfg["lock"] {
			cacheClient.Do("SET", cKey, tools.Int642Str(num), "PX", cfg["interval"])
		}
		if num >= cfg["lock"] {
			logs.Warning("hit password limit strategy-lock, key:", cKey, ", value:", num)
			msgErr = fmt.Sprintf(i18n.GetMessageText(i18n.MsgUserLocked), 6)
		} else if num >= cfg["notify"] {
			logs.Warning("hit password limit strategy-notify, key:", cKey, ", value:", num)
			msgErr = fmt.Sprintf(i18n.GetMessageText(i18n.MsgUserLocking), 6-num)
		}
	}

	return
}
