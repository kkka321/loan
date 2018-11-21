package service

import (
	"micro-loan/common/models"
	"micro-loan/common/thirdparty/nxtele"
	"micro-loan/common/types"
)

type AutoCallResult struct {
	StartTime int64
	IsDial    int
}

// 获取最新的自动呼叫结果
func GetLatestAutoCallResult(mobile string) (result AutoCallResult) {
	mobile = nxtele.MobileFormat(mobile)

	voiceRemind, err := models.GetLatestVoiceRemindByMobile(mobile)
	if err != nil {
		return
	}

	result.StartTime = voiceRemind.Ctime
	result.IsDial = types.PhoneNotConnected
	if voiceRemind.Duration > 0 {
		result.IsDial = types.PhoneConnected
	}

	return
}

// 获取所有的自动呼叫结果
func GetAllAutoCallResult(mobile string) (results []AutoCallResult) {
	mobile = nxtele.MobileFormat(mobile)

	voiceReminds, err := models.GetAllVoiceRemindByMobile(mobile)
	if err != nil {
		return
	}

	for _, v := range voiceReminds {
		var r AutoCallResult
		r.StartTime = v.Ctime
		r.IsDial = types.PhoneNotConnected
		if v.Duration > 0 {
			r.IsDial = types.PhoneConnected
		}
		results = append(results, r)
	}

	return
}
