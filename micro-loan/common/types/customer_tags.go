package types

// 客户账号失效时的手机号前缀
const (
	CustomerAccountInvalidSuffix = "_m"
)

// 客户标签
type CustomerTags int

const (
	CustomerTagsPotential   CustomerTags = 1 // 潜在客户 ：已完成注册，但未进行身份认证客户
	CustomerTagsTarget      CustomerTags = 2 // 目标客户：身份认证通过但未提交过一笔借款申请的客户
	CustomerTagsProspective CustomerTags = 3 // 准客户：未完成首贷，但存在进行中的借款申请（审核中/等待还款/审核拒绝）
	CustomerTagsDeal        CustomerTags = 4 // 成交客户：首贷完成的客户 （完成指已经结清）
	CustomerTagsLoyal       CustomerTags = 5 // 忠实客户：复贷完成的客户
)

var customerTagsMap = map[CustomerTags]string{
	CustomerTagsPotential:   "潜在客户",
	CustomerTagsTarget:      "目标客户",
	CustomerTagsProspective: "准客户",
	CustomerTagsDeal:        "成交客户",
	CustomerTagsLoyal:       "忠实客户",
}

func CustomerTagsMap() map[CustomerTags]string {
	return customerTagsMap
}

// 客户类型
const (
	CustomerTypeFirstLoan   = "a" // 首贷客户(非当天注册，并且没有已结清订单的客户)
	CustomerTypeRepeatLoan  = "b" // 复贷客户(非当天注册，有已结清订单的客户)
	CustomerTypeNewRegister = "c" // 新注册客户(当天注册的用户)
	CustomerTypeAll         = "d" // 所有客户
)

var customerTypesMap = map[string]string{
	CustomerTypeFirstLoan:   "首贷客户",
	CustomerTypeRepeatLoan:  "复贷客户",
	CustomerTypeNewRegister: "新注册客户",
	CustomerTypeAll:         "所有客户",
}

func CustomerTypesMap() map[string]string {
	return customerTypesMap
}

const (
	PlatformMark_No    int64 = 0
	PlatformMark_Gojek int64 = 0x1
	PlatformMark_Max
)

var platformMarkMap = map[int64]string{
	PlatformMark_Gojek: "gojek",
}

func PlatformMarkMap() map[int64]string {
	return platformMarkMap
}

func GetPlatformMarkDesc(platform int64) string {
	v, ok := platformMarkMap[platform]

	if ok {
		return v
	}

	return ""
}
