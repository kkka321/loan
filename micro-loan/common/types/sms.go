package types

const (
	// DeliveryUnknown 已验证
	DeliveryUnknown = iota
	// DeliverySuccess 未验证
	DeliverySuccess
	// DeliveryFailed 已验证
	DeliveryFailed
)

// DeliveryStatusMap 送达状态 map
var DeliveryStatusMap = map[int]string{
	DeliveryUnknown: "未知",
	DeliverySuccess: "已送达",
	DeliveryFailed:  "未送达",
}

// SmsServiceID 数字ID 用于节省DB存储空间和REDIS内存使用
type SmsServiceID int

// SmsServiceName 名称用于配置可视化
type SmsServiceName string

const (
	// NexoID 短信服务商 Nexo 在本系统中的ID
	NexoID SmsServiceID = 1
	// NexoName 短信服务商 Nexo 在本系统中的name
	NexoName SmsServiceName = "nexmo"

	// Sms253ID 短信服务商 上海创蓝在本系统中ID , 因其域名为 253.com, 故简称 sms253
	Sms253ID SmsServiceID = 2
	// Sms253Name 短信服务商 sms253 在本系统的name
	Sms253Name SmsServiceName = "sms253"

	// textlocal
	TextlocalID   SmsServiceID   = 3
	TextlocalName SmsServiceName = "textlocal"

	BoomSmsID   SmsServiceID   = 4
	BoomSmsName SmsServiceName = "boomsms"

	CmtelcomSmsID    SmsServiceID   = 5
	CmtelecomSmsName SmsServiceName = "cmtelecom"
)

// SmsServiceMap 用于列表和转换
var SmsServiceMap = map[SmsServiceName]SmsServiceID{
	NexoName:         NexoID,
	Sms253Name:       Sms253ID,
	TextlocalName:    TextlocalID,
	BoomSmsName:      BoomSmsID,
	CmtelecomSmsName: CmtelcomSmsID,
}

var SmsServiceIdMap = map[SmsServiceID]string{
	NexoID:        string(NexoName),
	Sms253ID:      string(Sms253Name),
	TextlocalID:   string(TextlocalName),
	BoomSmsID:     string(BoomSmsName),
	CmtelcomSmsID: string(CmtelecomSmsName),
}
