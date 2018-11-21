package credit

import (
	"encoding/json"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/pkg/system/config"
	"micro-loan/common/thirdparty/tongdun"
)

type ConfigInfo struct {
	Period  int `json:"period"`
	IsCatch int `json:"is_catch"`
}

type AuthConfigBackend struct {
	AuthList map[string]ConfigInfo `json:"auth_list"`
}

type AuthorInfo struct {
	IndonesiaName       string //客户端展示名字
	TongdunChannelCodes []string
	StatusColName       string
	CrawTimeColName     string
	QuotaColName        string
	BackendCode         string
	ReqColName          string
	RespondColName      string
}

const (
	BackendCodeYys       = "1001"
	BackendCodeGoJek     = "1002"
	BackendCodeLazada    = "1003"
	BackendCodeTokopedia = "1004"
	BackendCodeFacebook  = "1005"
	BackendCodeInstagram = "1006"
	BackendCodeLinkedin  = "1007"
	BackendCodeNpwp      = "1008"
)

var backendCodeMap = map[string]AuthorInfo{
	BackendCodeYys: {
		"Verifikasi operator",
		[]string{
			tongdun.ChannelCodeTelkomsel,
			tongdun.ChannelCodeXI,
			tongdun.ChannelCodeIndosat,
		},
		"AuthorizeStatusYys",
		"AuthorizeCrawleTimeYys",
		"QuotaRaiseYys",
		BackendCodeYys,
		"Yys",
		"YysQuota",
	},

	BackendCodeGoJek: {
		"GoJek",
		[]string{
			tongdun.ChannelCodeGoJek,
		},
		"AuthorizeStatusGoJek",
		"AuthorizeCrawleTimeGoJek",
		"QuotaRaiseGoJek",
		BackendCodeGoJek,
		"GoJek",
		"GoJekQuota",
	},
	BackendCodeLazada: {
		"Lazada",
		[]string{
			tongdun.ChannelCodeLazada,
		},
		"AuthorizeStatusLazada",
		"AuthorizeCrawleTimeLazada",
		"QuotaRaiseLazada",
		BackendCodeLazada,
		"Lazada",
		"LazadaQuota",
	},
	BackendCodeTokopedia: {
		"Tokopedia",
		[]string{
			tongdun.ChannelCodeTokopedia,
		},
		"AuthorizeStatusTokopedia",
		"AuthorizeCrawleTimeTokopedia",
		"QuotaRaiseTokopedia",
		BackendCodeTokopedia,
		"Tokopedia",
		"TokopediaQuota",
	},
	BackendCodeFacebook: {
		"Facebook",
		[]string{
			tongdun.ChannelCodeFacebook,
		},
		"AuthorizeStatusFacebook",
		"AuthorizeCrawleTimeFacebook",
		"QuotaRaiseFacebook",
		BackendCodeFacebook,
		"Facebook",
		"FacebookQuota",
	},
	BackendCodeInstagram: {
		"Instagram",
		[]string{
			tongdun.ChannelCodeInstagram,
		},
		"AuthorizeStatusInstagram",
		"AuthorizeCrawleTimeInstagram",
		"QuotaRaiseInstagram",
		BackendCodeInstagram,
		"Instagram",
		"InstagramQuota",
	},
	BackendCodeLinkedin: {
		"Linkedin",
		[]string{
			tongdun.ChannelCodeLinkedin,
		},
		"AuthorizeStatusLinkedin",
		"AuthorizeCrawleTimeLinkedin",
		"QuotaRaiseLinkedin",
		BackendCodeLinkedin,
		"Linkedin",
		"LinkedinQuota",
	},
	BackendCodeNpwp: {
		"NPWP",
		[]string{},
		"NpwpStatus",
		"NpwpTime",
		"QuotaRaiseNpwp",
		BackendCodeNpwp,
		"Npwp",
		"NpwpQuota",
	},
}

func AuthorizeValidityCatchList(isReloan bool) (items []string) {
	acb := AuthorizeValidityConfigAll(isReloan)

	for k, v := range acb.AuthList {
		if v.IsCatch == 1 {
			items = append(items, k)
		}
	}
	return
}

func AuthorizeValidityConfigAll(isReloan bool) AuthConfigBackend {
	v := ""
	if !isReloan {
		v = config.ValidItemString("additional_authorize_items_first")
	} else {
		v = config.ValidItemString("additional_authorize_items_reloan")
	}

	acb := AuthConfigBackend{}
	json.Unmarshal([]byte(v), &acb)
	return acb
}

func AuthorizeValidityPeriod(backendCode string, isReloan bool) (period int, isCatch int) {
	acb := AuthorizeValidityConfigAll(isReloan)
	if v, ok := acb.AuthList[backendCode]; ok {
		period = v.Period
		isCatch = v.IsCatch
	}

	if period == 0 {
		logs.Warn("[authorizeValidityPeriod] backendCode:%d period==0", backendCode)
		period = 9999
	}
	return
}

func AuthorInfoByTongdunChannelCode(channelCode string) (ret AuthorInfo, ok bool) {
	for _, v := range backendCodeMap {
		cs := v.TongdunChannelCodes
		for _, code := range cs {
			if code == channelCode {
				return v, true
			}
		}
	}
	return
}

func AuthorInfoByBackendCode(backendCode string) (ret AuthorInfo, ok bool) {
	ret, ok = backendCodeMap[backendCode]
	return
}

func BackendCodeMap() map[string]AuthorInfo {
	return backendCodeMap
}
