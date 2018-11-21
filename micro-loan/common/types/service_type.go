package types

type ServiceType int

const (
	ServiceRegisterOrLogin   ServiceType = 1  // 登录或注册
	ServiceRequestLogin      ServiceType = 2  // 请求登录或注册验证码
	ServiceCreateOrder       ServiceType = 3  // 创建订单
	ServiceRepeatedLoan      ServiceType = 4  // 复贷
	ServiceLogout            ServiceType = 5  // 注销
	ServiceDisburseSuccess   ServiceType = 6  // 放贷成功
	ServiceRegister          ServiceType = 7  // 注册
	ServiceLogin             ServiceType = 8  // 登录
	ServiceRepayRemind       ServiceType = 9  // 短信提醒还款
	ServiceCollectionRemind  ServiceType = 10 // 催收短信提醒
	ServiceAuthReport        ServiceType = 11 // 授权上报
	ServiceMonitor           ServiceType = 12 // 监控报警
	ServiceRollApplySuccess  ServiceType = 13 // 申请展期成功
	ServiceFindPassword      ServiceType = 14 // 找回密码
	ServiceConfirmOrder      ServiceType = 15 // 确认订单
	ServiceMarketPaymentCode ServiceType = 16 // 便利店支付码
	ServiceSales             ServiceType = 17 // 推广活动

	ServiceOthers ServiceType = 90 // 其他
)

var serviceTypeEnumMap = map[ServiceType]string{
	ServiceRegisterOrLogin:   "登录/注册",
	ServiceRequestLogin:      "请求登录/注册",
	ServiceCreateOrder:       "创建订单",
	ServiceRepeatedLoan:      "复贷",
	ServiceLogout:            "注销",
	ServiceDisburseSuccess:   "放贷成功",
	ServiceRegister:          "注册",
	ServiceLogin:             "登录",
	ServiceRepayRemind:       "短信提醒还款",
	ServiceCollectionRemind:  "催收短信提醒",
	ServiceMonitor:           "监控报警",
	ServiceRollApplySuccess:  "申请展期成功",
	ServiceFindPassword:      "找回密码",
	ServiceConfirmOrder:      "确认订单",
	ServiceMarketPaymentCode: "便利店支付码",
	ServiceSales:             "推广活动",
	ServiceOthers:            "其他",
}

// ServiceTypeEnumMap 返回 SMS Verify Code 的服务类型列表
func ServiceTypeEnumMap() map[ServiceType]string {
	return serviceTypeEnumMap
}
