package credit

import (
	"encoding/json"
	"reflect"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/models"
	"micro-loan/common/thirdparty/tongdun"
	"micro-loan/common/tools"
)

// 根据channelcode 对应去风控请求字段
var CreditCodeMap = map[string]string{
	tongdun.ChannelCodeTelkomsel: "Yys",
	tongdun.ChannelCodeXI:        "Yys",
	tongdun.ChannelCodeIndosat:   "Yys",
	tongdun.ChannelCodeGoJek:     "GoJek",
	tongdun.ChannelCodeLazada:    "Lazada",
	tongdun.ChannelCodeTokopedia: "Tokopedia",
	tongdun.ChannelCodeFacebook:  "Facebook",
	tongdun.ChannelCodeInstagram: "Instagram",
	tongdun.ChannelCodeLinkedin:  "Linkedin",
}

var host = "http://10.9.175.198:8000"
var increaseRoute = "/riskquota/increasecredit/"

type RequestIncreaseCredit struct {
	Version  string `json:"version"`
	UserData struct {
		AccountId       int64       `json:"account_id"`
		MaxDisplayQuota int64       `json:"max_display_quota"`
		MaxLoanQuota    int64       `json:"max_loan_quota"`
		OrderId         interface{} `json:"order_id"`
		LoanOrg         interface{} `json:"loan_org"`
		PeriodOrg       interface{} `json:"period_org"`
	} `json:"user_data"`
	Data struct {
		GoJek     interface{} `json:"gojek"`
		Yys       interface{} `json:"yys"`
		Lazada    interface{} `json:"lazada"`
		Tokopedia interface{} `json:"tokopedia"`
		Facebook  interface{} `json:"facebook"`
		Instagram interface{} `json:"instagram"`
		Linkedin  interface{} `json:"linkedin"`
		Npwp      interface{} `json:"npwp"`
	} `json:"data"`
}

type RequestItem struct {
	ChannelCode  string `json:"channel_code"`
	CallbackData string `json:"callback_data"`
}

type RespondIncreaseCredit struct {
	Version string `json:"version"`
	Code    int    `json:"code"`
	Data    struct {
		AccountId        int64 `json:"account_id"`
		IncreaseQuotaSum int64 `json:"increase_quota_sum"`
		GoJekQuota       int64 `json:"gojek_quota"`
		YysQuota         int64 `json:"yys_quota"`
		LazadaQuota      int64 `json:"lazada_quota"`
		TokopediaQuota   int64 `json:"tokopedia_quota"`
		FacebookQuota    int64 `json:"facebook_quota"`
		InstagramQuota   int64 `json:"instagram_quota"`
		LinkedinQuota    int64 `json:"linkedin_quota"`
		NpwpQuota        int64 `json:"npwp_quota"`
	} `json:"data"`
}

func init() {
	host = beego.AppConfig.String("credit_host")
	increaseRoute = beego.AppConfig.String("credit_increase_quota_route")
}

// NewSingleRequestIncreaseCreditByAccount 返回请求 struct
//func NewRequestIncreaseCreditByAccount(accountId int64) RequestIncreaseCredit {
//	accountQuotaConf, _ := models.OneAccountQuotaConfByAccountID(accountId)
//	Gojek, _ := models.GetLatestAC(accountId, "TRIP")
//	Yys, _ := models.GetLatestAC(accountId, "YYS")
//
//	tongdunNameModel := map[string]models.AccountTongdun{}
//	tongdunNameModel["GoJek"] = Gojek
//	tongdunNameModel["Yys"] = Yys
//
//	return NewRequestIncreaseCredit(accountQuotaConf, &tongdunNameModel)
//}

func NewRequestIncreaseCreditByTongdunModel(tongdunModels models.AccountTongdun) (r RequestIncreaseCredit) {
	accountQuotaConf, _ := models.OneAccountQuotaConfByAccountID(tongdunModels.AccountID)
	name, ok := CreditCodeMap[tongdunModels.ChannelCode]
	if !ok {
		logs.Error("NewRequestIncreaseCreditByTongdunModel: channel code err. %#v", tongdunModels)
		return
	}

	tongdunNameModel := map[string]models.AccountTongdun{}
	tongdunNameModel[name] = tongdunModels
	return NewRequestIncreaseCredit(accountQuotaConf, &tongdunNameModel)
}

func NewRequestIncreaseCreditByTNpwp(aExt models.AccountBaseExt) (r RequestIncreaseCredit) {
	accountQuotaConf, _ := models.OneAccountQuotaConfByAccountID(aExt.AccountId)
	return NewRequestIncreaseCreditNpwp(accountQuotaConf, aExt)
}

func NewRequestIncreaseCredit(accountQuota models.AccountQuotaConf, tongdunModels *map[string]models.AccountTongdun) (r RequestIncreaseCredit) {
	r.Version = "v0"
	r.UserData.AccountId = accountQuota.AccountID
	r.UserData.MaxDisplayQuota = accountQuota.QuotaVisable
	r.UserData.MaxLoanQuota = accountQuota.Quota

	for k, v := range *tongdunModels {
		item := RequestItem{
			ChannelCode:  v.ChannelCode,
			CallbackData: v.TaskData,
		}

		r = fillModel(r, k, item)
	}

	//logs.Info("before return r.version:%s", r.Version)
	return
}

func NewRequestIncreaseCreditNpwp(accountQuota models.AccountQuotaConf, aExt models.AccountBaseExt) (r RequestIncreaseCredit) {
	r.Version = "v0"
	r.UserData.AccountId = accountQuota.AccountID
	r.UserData.MaxDisplayQuota = accountQuota.QuotaVisable
	r.UserData.MaxLoanQuota = accountQuota.Quota
	r = fillModel(r, "Npwp", aExt.NpwpNo)

	return
}

func fillModel(r RequestIncreaseCredit, name string, itemValue interface{}) RequestIncreaseCredit {
	logs.Info("handel:%s", name)
	rv := reflect.ValueOf(&r)
	rv = rv.Elem()
	if !rv.IsValid() {
		logs.Error("[fillModel]  rv.IsValid :false. account:%d", r.UserData.AccountId)
		return r
	}

	data := rv.FieldByName("Data")
	item := data.FieldByName(name)
	if !item.IsValid() {
		logs.Error("[fillModel]  item.IsValid :false. account:%d name:%s", r.UserData.AccountId, name)
		return r
	}
	item.Set(reflect.ValueOf(itemValue))
	return r
}

func GetIncreaseCredit(req RequestIncreaseCredit) (res RespondIncreaseCredit) {
	url := host + increaseRoute

	bytesData, err := json.Marshal(req)

	reqBody := string(bytesData)

	reqHeader := map[string]string{
		"Content-Type": "application/json",
	}

	restByte, code, err := tools.SimpleHttpClient("POST", url, reqHeader, reqBody, tools.DefaultHttpTimeout())
	logs.Debug("[GetIncreaseCredit] req:%s, res:%s", reqBody, string(restByte))
	if err != nil {
		logs.Error("[GetIncreaseCredit] has wrong. url:", url, ", err:", err)
		return
	}

	if code != 200 {
		logs.Error("[GetIncreaseCredit] code wrong. url:", url, ", data:", string(restByte))
	}

	json.Unmarshal(restByte, &res)

	return
}

func DefaultRet(accountId int64) (ret RespondIncreaseCredit) {
	ret.Version = "v0"
	ret.Code = 200
	ret.Data.AccountId = accountId
	ret.Data.GoJekQuota = 10000
	ret.Data.YysQuota = 10000
	ret.Data.TokopediaQuota = 10000
	ret.Data.FacebookQuota = 10000
	ret.Data.InstagramQuota = 10000
	ret.Data.LinkedinQuota = 10000
	ret.Data.NpwpQuota = 10000

	return
}
