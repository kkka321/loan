package areacode

import (
	"strings"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/tools"
)

var defaultServiceRegion string

var phoneRegionCodeMap = map[string]string{
	tools.ServiceRegionIndonesia: "62",
	tools.ServiceRegionIndia:     "91",
	"MMR": "95",
	"CHN": "86",
}

func init() {
	defaultServiceRegion = tools.GetServiceRegion()
	if len(defaultServiceRegion) == 0 {
		logs.Alert("[Config Error] service_region must be configured")
	}
}

// PhoneWithServiceRegionCode 给电话号码附加国家电话编码前缀
func PhoneWithServiceRegionCode(mobile string) string {
	switch defaultServiceRegion {
	case tools.ServiceRegionIndonesia:
		if strings.HasPrefix(mobile, "08") {
			return strings.Replace(mobile, "08", "628", 1)
		}
		return mobile
	default:
		return normalPareseAndWrapCountryCode(mobile)
	}
}

// PhoneWithoutServiceRegionCode 将电话号码的国家前缀去掉
func PhoneWithoutServiceRegionCode(mobile string) string {
	switch defaultServiceRegion {
	case tools.ServiceRegionIndonesia:
		if strings.HasPrefix(mobile, "628") {
			return strings.Replace(mobile, "628", "08", 1)
		}
		return mobile
	default:
		return normalPareseAndWrapCountryCode(mobile)
	}
}

func normalPareseAndWrapCountryCode(mobile string) string {
	if v, ok := phoneRegionCodeMap[defaultServiceRegion]; ok {
		if strings.HasPrefix(mobile, v) {
			return mobile
		}
		return v + mobile
	}
	logs.Error("Cannot find the phone region code of service region: %s", defaultServiceRegion)
	return mobile
}

func GetRegionCode() string {
	v, ok := phoneRegionCodeMap[defaultServiceRegion]
	if ok {
		return v
	}

	return ""
}
