package service

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/tools"
)

// EsResponseSource 描述 ES 中 具体设备的源信息
// 索引 imeiMd5
type esResponseSource struct {
	IsAll                          int                `json:"is_all"`
	MorningCallMinutes             float64            `json:"morning_call_minutes"`
	NotObtainedMessage             int                `json:"not_obtained_message"`
	NotObtainedAddressList         int                `json:"not_obtained_address_list"`
	NumberOfContacts               int                `json:"number_of_contacts"`
	NotObtainedCallRecord          int                `json:"not_obtained_call_record"`
	NotObtainedGpsInfo             int                `json:"not_obtained_gps_info"`
	NotObtainedDeviceInfo          int                `json:"not_obtained_device_info"`
	PhoneDevice                    int                `json:"phone_device"`
	DevicePhone                    int                `json:"device_phone"`
	ProportionOfTelephone          float64            `json:"proportion_of_telephone"`
	NoCallRecordDays               int                `json:"no_call_record_days"`
	ProportionOfCallRecord         float64            `json:"proportion_of_call_record"`
	NumberOfMessagesContainKeyword int                `json:"number_of_messages_contain_keyword"`
	NumberOfCallsToFirstContact    map[string]int     `json:"number_of_calls_to_first_contact"`
	MinutesOfCallsToFirstContact   map[string]float64 `json:"minutes_of_calls_to_first_contact"`
	DistanceOfDevice               float64            `json:"distance_of_device"`
	TimesOfDeviceRegistered        int                `json:"times_of_device_registered"` // 1天内，同一设备注册时间间隔
	SameNumberInOut3Months         int                `json:"same_number_in_out_3_months"`
	LastCallDays                   int                `json:"last_call_days"`
	NoSmsRecordDays                int                `json:"no_sms_record_days"`
	LastSmsDays                    int                `json:"last_sms_days"`
	PhoneOverdueSmsNum             int                `json:"phone_overdue_sms_num"`
	PhoneOverdueOneSmsNum          int                `json:"phone_overdue_one_sms_num"`
	PhoneOverdueTwoSmsNum          int                `json:"phone_overdue_two_sms_num"`
	PhoneTuiguangSmsDayDiff        int                `json:"phone_tuiguang_sms_day_diff"`
}

// EsResponseAccountSource 描述 ES 中 具体用户的源信息
// 索引 accountID
type esResponseAccountSource struct {
	IsAll int `json:"is_all"`
}

// EsResponseIPSource 描述 ES 中 具体用户的源信息
// 索引 IP
type esResponseIPSource struct {
	IsAll int `json:"is_all"`
}

// esResponseACardSource 描述 ES 中 A卡的源信息
// 索引 accountID
type esResponseACardSource struct {
	IsAll             int     `json:"is_all"`
	AccountId         string  `json:"account_id"`
	IMEI              string  `json:"imei"`
	Prob              float64 `json:"prob"`
	AppCount          int     `json:"dev_apps_like_paydayloan_num"`
	RangeAppCount     int     `json:"dev_apps_like_paydayloan_in_mdays_num"`
	RangeContactCount int     `json:"dev_contact_create_in_mdays_num"`
}

// EsResponse 描述 ES 设备源信息返回结构
type EsResponse struct {
	Found  bool             `json:"found"`
	Source esResponseSource `json:"_source"`
}

// EsAccountResponse 描述 ES 账户源信息返回结构
type EsAccountResponse struct {
	Found  bool                    `json:"found"`
	Source esResponseAccountSource `json:"_source"`
}

// EsIPResponse 描述 ES 账户源信息返回结构
type EsIPResponse struct {
	Found  bool               `json:"found"`
	Source esResponseIPSource `json:"_source"`
}

// EsACardResponse 描述 ES A卡信息返回结构
type EsACardResponse struct {
	Found  bool                  `json:"found"`
	Source esResponseACardSource `json:"_source"`
}

const esIsAll int = 1

func (r *EsResponse) IsAll() (yes bool) {
	yes = r.Source.IsAll == esIsAll
	return
}

func (r *EsAccountResponse) IsAll() (yes bool) {
	yes = r.Source.IsAll == esIsAll
	return
}

func (r *EsIPResponse) IsAll() (yes bool) {
	yes = r.Source.IsAll == esIsAll
	return
}

func (r *EsACardResponse) IsAll() (yes bool) {
	yes = r.Source.IsAll == esIsAll
	return
}

// EsSearchById 根据设备ID imeiMd5 返回设备信息
func EsSearchById(id string) (esRes EsResponse, esIndex string, restByte []byte, err error) {
	esHost := beego.AppConfig.String("es_host")
	esIndex = beego.AppConfig.String("es_index")
	esType := beego.AppConfig.String("es_type")

	esApiUrl := fmt.Sprintf("%s/%s/%s/%s", esHost, esIndex, esType, id)
	logs.Debug("esApiUrl: %s", esApiUrl)
	reqHeaders := map[string]string{}
	restByte, _, err = tools.SimpleHttpClient("GET", esApiUrl, reqHeaders, "", tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[EsSearch] has wrong. esApiUrl:", esApiUrl, ", err:", err)
		return
	}

	json.Unmarshal(restByte, &esRes)

	return
}

// EsSearchByAccountId 根据用户 ID 返回用户数据
func EsSearchByAccountId(id int64) (esRes EsAccountResponse, esIndex string, restByte []byte, err error) {
	esHost := beego.AppConfig.String("es_host")
	// 添加新配置 ES_INDEX microloan_account
	esIndex = beego.AppConfig.String("es_account_index")
	esType := beego.AppConfig.String("es_type")

	esApiUrl := fmt.Sprintf("%s/%s/%s/%s", esHost, esIndex, esType, strconv.FormatInt(id, 10))
	reqHeaders := map[string]string{}
	restByte, _, err = tools.SimpleHttpClient("GET", esApiUrl, reqHeaders, "", tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[EsSearch] has wrong. esApiUrl:", esApiUrl, ", err:", err)
		return
	}

	json.Unmarshal(restByte, &esRes)

	return
}

// EsSearchByIP 根据IP 返回数据
func EsSearchByIP(ip string) (esRes EsIPResponse, esIndex string, restByte []byte, err error) {
	esHost := beego.AppConfig.String("es_host")
	// 添加新配置 ES_INDEX microloan_account
	esIndex = beego.AppConfig.String("es_ip_index")
	esType := beego.AppConfig.String("es_type")

	esApiUrl := fmt.Sprintf("%s/%s/%s/%s", esHost, esIndex, esType, ip)
	reqHeaders := map[string]string{}
	restByte, _, err = tools.SimpleHttpClient("GET", esApiUrl, reqHeaders, "", tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[EsSearch] has wrong. esApiUrl:", esApiUrl, ", err:", err)
		return
	}

	json.Unmarshal(restByte, &esRes)

	return
}

// EsSearchByIP 根据IP 返回数据
func EsSearchACardByImei(md5Imei string) (esRes EsACardResponse, esIndex string, restByte []byte, err error) {
	esHost := beego.AppConfig.String("es_host")
	esIndex = beego.AppConfig.String("es_acard_index")
	esType := beego.AppConfig.String("es_type")

	esApiUrl := fmt.Sprintf("%s/%s/%s/%s", esHost, esIndex, esType, md5Imei)
	reqHeaders := map[string]string{}
	restByte, _, err = tools.SimpleHttpClient("GET", esApiUrl, reqHeaders, "", tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[EsSearch] has wrong. esApiUrl:", esApiUrl, ", err:", err)
		return
	}

	json.Unmarshal(restByte, &esRes)

	return
}
