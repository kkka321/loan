package appsflyer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty"
	"net/http"

	"github.com/astaxie/beego/logs"
)

// Origin 代表用户
type Origin struct {
	AppsflyerDeviceID   string `json:"appsflyer_device_id"`
	GoogleAdvertisingID string `json:"advertising_id"`
	MediaSource         string `json:"media_source"`
	Campaign            string `json:"campaign"`
	EventType           string `json:"event_type"`
	AppID               string `json:"app_id"`
	// like 2018-06-06 08:30:37 Unix
	InstallUnixTime string `json:"install_time"`
	AppVersion      string `json:"app_version"`
	City            string `json:"city"`
	DeviceModel     string `json:"device_model"`
}

const (
	// if event_type = install , it's valid install post back
	installEventType = "install"
	// media_source:Organic
	organicTag = "Organic"
)

// ParseOrigin 从 appsflyer request 中解析出用户来源
func ParseOrigin(r *http.Request) (*Origin, error) {
	r.ParseForm()
	var bytesData []byte
	var err error
	data := &Origin{}

	bytesData, err = ioutil.ReadAll(r.Body)
	if err != nil {
		logs.Error("[ParseOrigin] read body err:", err)
		return data, err
	}
	logs.Debug("[ParseOrigin]read body:", string(bytesData))

	err = json.Unmarshal(bytesData, data)
	if err != nil {
		logs.Error("[Appsflyer][ParseOrigin]", err)
		return data, err
	}

	if data == nil {
		err = fmt.Errorf("[Appsflyer][ParseOrigin] cannot get valid data Origin{} from :%s", string(bytesData))
		return data, err
	}

	if r.Method != http.MethodPost && !checkInstallCallback(data) {
		err = fmt.Errorf("[Appsflyer][ParseOrigin] check failed, not invalid request")
		return data, err
	}
	// record valid push
	responstType, fee := thirdparty.CalcFeeByApi(r.RequestURI, string(bytesData), "")
	models.AddOneThirdpartyRecord(models.ThirdpartyAppsFlyer, r.RequestURI, 0, string(bytesData), nil, responstType, fee, 200)
	// event.Trigger(&evtypes.CustomerStatisticEv{
	// 	UserAccountId: 0,
	// 	OrderId:       0,
	// 	ApiMd5:        tools.Md5(r.RequestURI),
	// 	Fee:           int64(fee),
	// 	Result:        responstType,
	// })

	return data, err
}

func checkInstallCallback(data *Origin) bool {
	logs.Debug("[Appsflyer][checkInstallCallback] start data check,postData:", data)

	if len(data.AppID) <= 0 {
		logs.Error("[Appsflyer][checkInstallCallback], no app_id parameter, data:", data)
		return false
	}

	// if data.AppID != appID {
	// 	logs.Error("[Appsflyer][checkInstallCallback],request appID(%s) is not the appID(%s) in configure", data.AppID, appID)
	// 	return false
	// }

	// event_type 校验
	if data.EventType != installEventType {
		logs.Error("[Appsflyer][checkInstallCallback], no valid event_type parameter, want %s, post data:%v", installEventType, data)
		return false
	}

	// event_type 校验
	if len(data.AppsflyerDeviceID) <= 0 {
		logs.Error("[Appsflyer][checkInstallCallback], appsflyer_device_id is invalid, post data:", data)
		return false
	}

	return true
}
