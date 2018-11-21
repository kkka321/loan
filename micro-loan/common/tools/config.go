// 针对不同服务区域的配置,通过一个服务区域来进行代码级别的配置,简化配置文件的切换

package tools

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

const (
	ServiceRegionIndonesia = "IDN"   // 印度尼西亚
	ServiceRegionIndia     = "INDIA" // 印度,和印尼长的太像,为了区分,采用全称
)

func GetServiceRegion() (serviceRegion string) {
	serviceRegion = beego.AppConfig.String("service_region")
	return
}

// 服务区域的时区
func GetServiceTimezone() (timezone string) {
	serviceRegion := GetServiceRegion()

	switch serviceRegion {
	case ServiceRegionIndonesia:
		timezone = "Asia/Jakarta"
	case ServiceRegionIndia:
		timezone = "Asia/Kolkata"
	}

	return
}

// 服务区域的代币
func GetServiceCurrency() (serviceCurrency string) {
	serviceRegion := GetServiceRegion()

	switch serviceRegion {
	case ServiceRegionIndonesia:
		serviceCurrency = "IDR"
	case ServiceRegionIndia:
		serviceCurrency = "INR"
	}

	return
}

// 不同服务区域的参数签名盐不一样
func GetSignatureSecret() (secret string) {
	serviceRegion := GetServiceRegion()

	switch serviceRegion {
	case ServiceRegionIndonesia:
		secret = "hy0le#GML0k"
	case ServiceRegionIndia:
		secret = "hy0kle#GM/mb"
	default:
		logs.Error("no service region, please check it out.")
	}
	//logs.Debug("serviceRegion:", serviceRegion, ", secret:", secret)
	return
}

// GetEntrustSignatureSecret 勤为接口不同服务区域的参数签名盐不一样
func GetEntrustSignatureSecret(pname string) (secret string) {
	switch pname {
	case "qinweigroup":
		secret = "MobiMG6bd8QWGefae"
	case "jucegroup":
		secret = "MobiMG3cHk5OkC488F"
	case "dachuigroup":
		secret = "MMc3H7lL411NK1fU8E"
	case "mbagroup":
		secret = "ckIDD978jVNSL12D4k"
	default:
		logs.Error("no service region, please check it out.")
	}
	//logs.Debug("serviceRegion:", serviceRegion, ", secret:", secret)
	return
}
