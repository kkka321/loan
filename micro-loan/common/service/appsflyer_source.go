package service

import (
	"micro-loan/common/models"
	"micro-loan/common/thirdparty/appsflyer"
	"time"

	"github.com/astaxie/beego/logs"
)

// CreateAccountOriginByAppsflyerPush 根据appsflyer push数据设置归因
func CreateAccountOriginByAppsflyerPush(origin *appsflyer.Origin) bool {
	appsflyerSource, _ := models.OneAppsflyerSourceByAppsflyerID(origin.AppsflyerDeviceID)
	if appsflyerSource.Id > 0 {
		logs.Error("[SetAccountOriginByAppsflyerPush]appsflyerSource already exist,ignore it;want to set Origin: ", origin)
		return false
	}

	appsflyerSource.AppsflyerID = origin.AppsflyerDeviceID
	appsflyerSource.MediaSource = origin.MediaSource
	appsflyerSource.Campaign = origin.Campaign
	appsflyerSource.GoogleAdvertisingID = origin.GoogleAdvertisingID
	appsflyerSource.AppVersion = origin.AppVersion
	appsflyerSource.City = origin.City
	appsflyerSource.DeviceModel = origin.DeviceModel
	unixTime, _ := time.Parse("2006-01-02 15:04:05", origin.InstallUnixTime)
	appsflyerSource.InstallTime = unixTime.Unix() * 1000
	id, err := appsflyerSource.Add()
	if err != nil {
		logs.Error("[CreateAccountOriginByAppsflyerPush]", err)
	}
	if id > 0 {
		return true
	}

	return false
}
