package service

import (
	"encoding/json"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"fmt"
	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type RiskRequestDetail struct {
	OpUid                     int64       `json:"op_uid"`
	ThirdName                 string      `json:"third_name"`
	Loan                      int64       `json:"loan"`
	Period                    int         `json:"period"`
	ServiceType               string      `json:"service_type"`
	PenaltyUtime              int64       `json:"penalty_utime"`
	CheckStatus               int         `json:"check_status"`
	ThirdVillage              string      `json:"third_village"`
	City                      string      `json:"city"`
	ApplyTime                 int64       `json:"apply_time"`
	ChildrenNumber            int         `json:"children_number"`
	CheckTime                 int64       `json:"check_time"`
	RiskCtlStatus             int         `json:"risk_ctl_status"`
	ServiceYears              int         `json:"service_years"`
	Platform                  string      `json:"platform"`
	AppVersionCode            string      `json:"app_version_code"`
	OcrRealname               string      `json:"ocr_realname"`
	Ip                        string      `json:"ip"`
	Latitude                  string      `json:"latitude"`
	Contact2Name              string      `json:"contact2_name"`
	AppVersion                string      `json:"app_version"`
	CompanyName               string      `json:"company_name"`
	BankName                  string      `json:"bank_name"`
	IsSimulator               int         `json:"is_simulator"`
	Realname                  string      `json:"realname"`
	RejectReason              int         `json:"reject_reason"`
	Brand                     string      `json:"brand"`
	PhoneVerifyAt             string      `json:"phone_verify_at"`
	HandPhotoQualityThreshold string      `json:"hand_photo_quality_threshold"`
	IdPhoto                   int64       `json:"id_photo"`
	LoanTime                  int64       `json:"loan_time"`
	HandHeldIdPhoto           int64       `json:"hand_held_id_photo"`
	ResidentAddress           string      `json:"resident_address"`
	Imei                      string      `json:"imei"`
	ProductId                 int64       `json:"product_id"`
	Relationship2             int         `json:"relationship2"`
	Relationship1             int         `json:"relationship1"`
	IsDeadDebt                int         `json:"is_dead_debt"`
	PhoneVerifyTime           int64       `json:"phone_verify_time"`
	Gender                    int         `json:"gender"`
	MaritalStatus             int         `json:"marital_status"`
	OrderId                   int64       `json:"order_id"`
	RiskCtlFinishTime         int64       `json:"risk_ctl_finish_time"`
	Longitude                 string      `json:"longitude"`
	FinishTime                int64       `json:"finish_time"`
	IdentityMessage           string      `json:"identity_message"`
	ThirdProvince             string      `json:"third_province"`
	FaceComparison            string      `json:"face_comparison"`
	IdCheckResult             string      `json:"id_check_result"`
	UserAccountId             int64       `json:"user_account_id"`
	OcrIdentity               string      `json:"ocr_identity"`
	LoanAt                    string      `json:"loan_at"`
	JobType                   int         `json:"job_type"`
	RandomValue               int         `json:"random_value"`
	IdPhotoQualityThreshold   string      `json:"id_photo_quality_threshold"`
	Contact1Name              string      `json:"contact1_name"`
	IdentityResult            string      `json:"identity_result"`
	Education                 int         `json:"education"`
	MonthlyIncome             int         `json:"monthly_income"`
	Network                   string      `json:"network"`
	HandPhotoQuality          string      `json:"hand_photo_quality"`
	LastLoginTime             int64       `json:"last_login_time"`
	AppsflyerId               string      `json:"appsflyer_id"`
	RegisterTime              int64       `json:"register_time"`
	ThirdCity                 string      `json:"third_city"`
	CompanyCity               string      `json:"company_city"`
	GoogleAdvertisingId       string      `json:"google_advertising_id"`
	CreatedDt                 string      `json:"created_dt"`
	IdCheckMessage            string      `json:"id_check_message"`
	IdPhotoQuality            string      `json:"id_photo_quality"`
	IdCheckSimilarity         float64     `json:"id_check_similarity"`
	BankNo                    string      `json:"bank_no"`
	Status                    int         `json:"status"`
	Tags                      int         `json:"tags"`
	RandomMark                int         `json:"random_mark"`
	ThirdId                   string      `json:"third_id"`
	CheckAt                   string      `json:"check_at"`
	ResidentCity              string      `json:"resident_city"`
	RiskCtlRegular            string      `json:"risk_ctl_regular"`
	Nickname                  string      `json:"nickname"`
	Identity                  string      `json:"identity"`
	IdHoldingPhotoCheck       string      `json:"id_holding_photo_check"`
	FaceQuality               float64     `json:"face_quality"`
	IsOverdue                 int         `json:"is_overdue"`
	Contact1                  string      `json:"contact1"`
	Contact2                  string      `json:"contact2"`
	RepayTime                 int64       `json:"repay_time"`
	Mobile                    string      `json:"mobile"`
	ApplyAt                   string      `json:"apply_at"`
	CreatedAt                 string      `json:"created_at"`
	FixedRandom               int         `json:"fixed_random"`
	TimeZone                  string      `json:"time_zone"`
	Amount                    int64       `json:"amount"`
	CompanyAddress            string      `json:"company_address"`
	IsTemporary               int         `json:"is_temporary"`
	Model                     string      `json:"model"`
	Os                        string      `json:"os"`
	ThirdDistrict             string      `json:"third_district"`
	UserAuthority             interface{} `json:"user_authority"`
}

type UserAuthorityInfo struct {
	Yys       int `json:"yys"`
	Facebook  int `json:"facebook"`
	GoJek     int `json:"go_jek"`
	Instagram int `json:"instagram"`
	Lazada    int `json:"lazada"`
	Linkedin  int `json:"linkedin"`
	Tokopedia int `json:"tokopedia"`
}

type RiskRequestInfo struct {
	Model   string              `json:"model"`
	Version string              `json:"version"`
	Data    []RiskRequestDetail `json:"data"`
}

type RiskResponseDetail struct {
	Score int `json:"score"`
	Extra struct {
		Prob      float64 `json:"prob"`
		AccountId int64   `json:"account_id"`
		Imei      string  `json:"imei"`
		Version   string  `json:"version"`
	} `json:"extra"`
}

type RiskResponse struct {
	Status int                  `json:"status"`
	Msg    string               `json:"msg"`
	Data   []RiskResponseDetail `json:"data"`
	ReqId  string               `json:"reqId"`
}

type FraudRequestDetail struct {
	Id          int64  `json:"id"`
	Mobile      string `json:"mobile"`
	ServiceType int    `json:"service_type"`
	RelatedId   int64  `json:"related_id"`
	Ip          string `json:"ip"`
	Os          string `json:"os"`
	Imei        string `json:"imei"`
	Model       string `json:"model"`
	Brand       string `json:"brand"`
	AppVersion  string `json:"app_version"`
	Longitude   string `json:"longitude"`
	Latitude    string `json:"latitude"`
	City        string `json:"city"`
	TimeZone    string `json:"time_zone"`
	Network     string `json:"network"`
	IsSimulator int    `json:"is_simulator"`
	Platform    string `json:"platform"`
	Ctime       int64  `json:"ctime"`
}

type FraudRequestInfo struct {
	Imei      string               `json:"imei"`
	AccountId int64                `json:"account_id"`
	OrderId   int64                `json:"order_id"`
	Data      []FraudRequestDetail `json:"data"`
}

type FraudResponse struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Data   struct {
		DistanceOfDevice                  float64 `json:"distance_of_device"`
		TimesOfDeviceRegistered           float64 `json:"times_of_device_registered"`
		AccountRegisteredDevice           int     `json:"account_registered_device"`
		AccountSameDeviceRegistered7Days  int     `json:"account_same_device_registered_7_days"`
		AccountSameDeviceRegistered30Days int     `json:"account_same_device_registered_30_days"`
		AccountSameDeviceLoginedOneday    int     `json:"account_same_device_logined_oneday"`
		AccountSameDeviceLogined7Days     int     `json:"account_same_device_logined_7_days"`
		AccountSameDeviceLogined30Days    int     `json:"account_same_device_logined_30_days"`
		AccountSameDeviceLoginedHistory   int     `json:"account_same_device_logined_history"`
		DistanceOfAccount                 float64 `json:"distance_of_account"`
		DeviceSameAccountLoginedOneday    int     `json:"device_same_account_logined_oneday"`
		DeviceSameAccountLoginedHistory   int     `json:"device_same_account_logined_history"`
		DeviceSameIpRegistered            int     `json:"device_same_ip_registered"`
		AccountsSameIpRegistered          int     `json:"accounts_same_ip_registered"`
	} `json:"data"`
	ReqId string `json:"reqId"`
}

type GraphRequestInfo struct {
	Model          string `json:"model"`
	Version        string `json:"version"`
	Scene          string `json:"scene"`
	AccountId      int64  `json:"account_id"`
	OrderId        int64  `json:"order_id"`
	Imei           string `json:"imei"`
	Identity       string `json:"identity"`
	Realname       string `json:"realname"`
	Gender         int    `json:"gender"`
	Mobile         string `json:"mobile"`
	RegisterTime   int64  `json:"register_time"`
	MonthlyIncome  int    `json:"monthly_income"`
	Education      int    `json:"education"`
	MaritalStatus  int    `json:"marital_status"`
	ChildrenNumber int    `json:"children_number"`
	Contact1       string `json:"contact1"`
	Contact1Name   string `json:"contact1_name"`
	Relationship1  int    `json:"relationship1"`
	Contact2       string `json:"contact2"`
	Contact2Name   string `json:"contact2_name"`
	Relationship2  int    `json:"relationship2"`
	CompanyMobile  string `json:"company_mobile"`
	CompanyName    string `json:"company_name"`
	ServiceYears   int    `json:"service_years"`
	JobType        int    `json:"job_type"`
	BankNo         string `json:"bank_no"`
	BankName       string `json:"bank_name"`
	Ip             string `json:"ip"`
}

type GraphResponse struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Data   struct {
		Imei      string `imei`
		AccountId int64  `account_id`
		OrderId   int64  `order_id`
		Graph     struct {
			IpDeviceAllNum       int `json:"graph_ip_device_all_num"`
			IpAccountAllNum      int `json:"graph_ip_account_all_num"`
			DeviceAccountAllNum  int `json:"graph_device_account_all_num"`
			ContactAccountAllNum int `json:"graph_contact_account_all_num"`
			CompanyAccountAllNum int `json:"graph_company_account_all_num"`
			BanknoDeviceAllNum   int `json:"graph_bankno_device_all_num"`
			BanknoAccountAllNum  int `json:"graph_bankno_account_all_num"`
			AccountDeviceAllNum  int `json:"graph_account_device_all_num"`
		} `json:"graph"`
	} `json:"data"`
	ReqId string `json:"reqId"`
}

type HyruleRequestInfo struct {
	BasicInfo struct {
		ChannelFrom string `json:"channel_from"`
		AccountId   int64  `json:"account_id"`
		OrderId     int64  `json:"order_id"`
		Imei        string `json:"imei"`
		Mobile      string `json:"mobile"`
		Identity    string `json:"identity"`
		Realname    string `json:"realname"`
	} `json:"basic_info"`

	UserInfo struct {
		AccountId                 int64  `json:"account_id"`
		AppsflyerId               string `json:"appsflyer_id"`
		BankName                  string `json:"bank_name"`
		BankNo                    string `json:"bank_no"`
		ChildrenNumber            int    `json:"children_number"`
		CompanyAddress            string `json:"company_address"`
		CompanyCity               string `json:"company_city"`
		CompanyMobile             string `json:"company_mobile"`
		CompanyName               string `json:"company_name"`
		Contact1                  string `json:"contact1"`
		Contact1Name              string `json:"contact1_name"`
		Contact2                  string `json:"contact2"`
		Contact2Name              string `json:"contact2_name"`
		Education                 int    `json:"education"`
		FaceComparison            string `json:"face_comparison"`
		Gender                    int    `json:"gender"`
		GoogleAdvertisingId       string `json:"google_advertising_id"`
		HandHeldIdPhoto           int64  `json:"hand_held_id_photo"`
		HandPhotoQuality          string `json:"hand_photo_quality"`
		HandPhotoQualityThreshold string `json:"hand_photo_quality_threshold"`
		IdHoldingPhotoCheck       string `json:"id_holding_photo_check"`
		IdPhoto                   int64  `json:"id_photo"`
		IdPhotoQuality            string `json:"id_photo_quality"`
		IdPhotoQualityThreshold   string `json:"id_photo_quality_threshold"`
		Identity                  string `json:"identity"`
		JobType                   int    `json:"job_type"`
		LastLoginTime             int64  `json:"last_login_time"`
		MaritalStatus             int    `json:"marital_status"`
		Mobile                    string `json:"mobile"`
		MonthlyIncome             int    `json:"monthly_income"`
		Nickname                  string `json:"nickname"`
		OcrIdentity               string `json:"ocr_identity"`
		OcrRealname               string `json:"ocr_realname"`
		Realname                  string `json:"realname"`
		RegisterTime              int64  `json:"register_time"`
		Relationship1             int    `json:"relationship1"`
		Relationship2             int    `json:"relationship2"`
		ResidentAddress           string `json:"resident_address"`
		ResidentCity              string `json:"resident_city"`
		SalaryDay                 string `json:"salary_day"`
		ServiceYears              int    `json:"service_years"`
		Status                    int    `json:"status"`
		Tags                      int    `json:"tags"`
		ThirdCity                 string `json:"third_city"`
		ThirdDistrict             string `json:"third_district"`
		ThirdId                   string `json:"third_id"`
		ThirdName                 string `json:"third_name"`
		ThirdProvince             string `json:"third_province"`
		ThirdVillage              string `json:"third_village"`
		AccuOverdueNum            int    `json:"accu_overdue_num"`
		ApplyOrderNum             int    `json:"apply_order_num"`
		MaxOverdueDays            int    `json:"max_overdue_days"`
		LoanOrderNum              int    `json:"loan_order_num"`
		TotalOverdueDays          int    `json:"total_overdue_days"`
		OverdueStatus             int    `json:"overdue_status"`
	} `json:"user_info"`

	OrderInfo struct {
		Amount            int64  `json:"amount"`
		ApplyTime         int64  `json:"apply_time"`
		CheckStatus       int    `json:"check_status"`
		CheckTime         int64  `json:"check_time"`
		Ctime             int64  `json:"ctime"`
		FinishTime        int64  `json:"finish_time"`
		FixedRandom       int    `json:"fixed_random"`
		IsDeadDebt        int    `json:"is_dead_debt"`
		IsOverdue         int    `json:"is_overdue"`
		IsReloan          int    `json:"is_reloan"`
		IsTemporary       int    `json:"is_temporary"`
		Loan              int64  `json:"loan"`
		LoanOrg           int64  `json:"loan_org"`
		LoanTime          int64  `json:"loan_time"`
		OpUid             int64  `json:"op_uid"`
		OrderId           int64  `json:"order_id"`
		PenaltyUtime      int64  `json:"penalty_utime"`
		Period            int    `json:"period"`
		PeriodOrg         int    `json:"period_org"`
		PhoneVerifyTime   int64  `json:"phone_verify_time"`
		ProductId         int64  `json:"product_id"`
		RandomMark        int    `json:"random_mark"`
		RandomValue       int    `json:"random_value"`
		RejectReason      int    `json:"reject_reason"`
		RepayTime         int64  `json:"repay_time"`
		RiskCtlFinishTime int64  `json:"risk_ctl_finish_time"`
		RiskCtlRegular    string `json:"risk_ctl_regular"`
		RiskCtlStatus     int    `json:"risk_ctl_status"`
	} `json:"order_info"`

	DeviceInfo struct {
		AppVersion     string `json:"app_version"`
		AppVersionCode string `json:"app_version_code"`
		Brand          string `json:"brand"`
		City           string `json:"city"`
		Ctime          int64  `json:"ctime"`
		Imei           string `json:"imei"`
		Ip             string `json:"ip"`
		IsSimulator    int    `json:"is_simulator"`
		Latitude       string `json:"latitude"`
		Longitude      string `json:"longitude"`
		Model          string `json:"model"`
		Network        string `json:"network"`
		Os             string `json:"os"`
		Platform       string `json:"platform"`
		RelatedId      int64  `json:"related_id"`
		ServiceType    int    `json:"service_type"`
	} `json:"device_info"`

	ExtraInfo struct {
		AccountId         int64   `json:"account_id"`
		FaceQuality       float64 `json:"face_quality"`
		IdCheckMessage    string  `json:"id_check_message"`
		IdCheckResult     string  `json:"id_check_result"`
		IdCheckSimilarity float64 `json:"id_check_similarity"`
		IdentityMessage   string  `json:"identity_message"`
		IdentityResult    string  `json:"identity_result"`
	} `json:"extra_info"`
}

type HyruleResponse struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Data   struct {
		CreditLimit struct {
			NonCashableLimit int `json:"non_cashable_limit"`
			CashableLimit    int `json:"cashable_limit"`
			CreditType       int `json:"credit_type"`
			TotalLimit       int `json:"total_limit"`
		} `json:"credit_limit"`

		PeriodLimit struct {
			PeriodValue string `json:"period_value"`
			PeriodType  string `json:"period_type"`
		} `json:"period_limit"`

		RateLimit struct {
			RateValue float64 `json:"rate_value"`
			RateType  string  `json:"rate_type"`
		} `json:"rate_limit"`

		HitRule       string      `json:"hit_rule"`
		LastDiagramId string      `json:"last_diagram_id"`
		FinalDecision string      `json:"final_decision"`
		DataUsed      interface{} `json:"data_used"`
		Plugin        interface{} `json:"plugin"`
	} `json:"data"`
	ReqId string `json:"req_id"`
}

func (c *RiskResponse) IsSuccess() bool {
	if c.Status == 0 && strings.ToLower(c.Msg) == "success" && len(c.Data) > 0 {
		return true
	}

	return false
}

func (c *FraudResponse) IsSuccess() bool {
	if c.Status == 0 && strings.ToLower(c.Msg) == "success" {
		return true
	}

	return false
}

func (c *GraphResponse) IsSuccess() bool {
	if c.Status == 0 {
		return true
	}

	return false
}

func (c *HyruleResponse) IsSuccess() bool {
	if c.Status == 0 {
		return true
	}

	return false
}

func GetFantasyRisk(req RiskRequestInfo) (restByte []byte, index string, res RiskResponse, err error) {
	host := beego.AppConfig.String("fantasy_host")
	url := host + "/risk"
	index = "fantasy/risk"

	reqHeader := map[string]string{
		"Content-Type": "application/json",
	}

	bytesData, err := json.Marshal(req)

	reqBody := string(bytesData)

	logs.Debug("[GetFantasyRisk] req:", reqBody)

	restByte, code, err := tools.SimpleHttpClient("POST", url, reqHeader, reqBody, tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[GetFantasyRisk] has wrong. url:", url, ", err:", err)
		return
	}

	if code != 200 {
		logs.Error("[GetFantasyRisk] code wrong. url:", url, ", data:", string(restByte))
	}

	json.Unmarshal(restByte, &res)

	return
}

func GetFantasyFraud(req FraudRequestInfo) (restByte []byte, index string, res FraudResponse, err error) {
	host := beego.AppConfig.String("fantasy_host")
	url := host + "/fraud"
	index = "fantasy/fraud"

	bytesData, err := json.Marshal(req)

	reqBody := string(bytesData)

	reqHeader := map[string]string{
		"Content-Type": "application/json",
	}

	logs.Debug("[GetFantasyFraud] req:", reqBody)

	restByte, code, err := tools.SimpleHttpClient("POST", url, reqHeader, reqBody, tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[GetFantasyFraud] has wrong. url:", url, ", err:", err)
		return
	}

	if code != 200 {
		logs.Error("[GetFantasyFraud] code wrong. url:", url, ", data:", string(restByte))
	}

	json.Unmarshal(restByte, &res)

	return
}

func GetFantasyGraph(req GraphRequestInfo) (restByte []byte, index string, res GraphResponse, err error) {
	host := beego.AppConfig.String("fantasy_host")
	url := host + "/graph/detect"
	index = "fantasy/graph/detect"

	bytesData, err := json.Marshal(req)

	reqBody := string(bytesData)

	reqHeader := map[string]string{
		"Content-Type": "application/json",
	}

	restByte, code, err := tools.SimpleHttpClient("POST", url, reqHeader, reqBody, tools.DefaultHttpTimeout())
	logs.Debug("[GetFantasyGraph] req:%s, res:%s", reqBody, string(restByte))
	if err != nil {
		logs.Error("[GetFantasyGraph] has wrong. url:", url, ", err:", err)
		return
	}

	if code != 200 {
		logs.Error("[GetFantasyGraph] code wrong. url:", url, ", data:", string(restByte))
	}

	json.Unmarshal(restByte, &res)

	return
}

func GetHyruleResult(req HyruleRequestInfo) (restByte []byte, index string, res HyruleResponse, err error) {
	host := beego.AppConfig.String("hyrule_host")
	country := beego.AppConfig.DefaultString("hyrule_country", "indonesia")
	scene := beego.AppConfig.DefaultString("hyrule_scene", "app_deal")

	url := fmt.Sprintf(`%s/pangu/partner/%s/scene/%s/user/%d`, host, country, scene, req.BasicInfo.AccountId)
	index = "pangu/partner"

	bytesData, err := json.Marshal(req)

	reqBody := string(bytesData)

	reqHeader := map[string]string{
		"Content-Type": "application/json",
	}

	restByte, code, err := tools.SimpleHttpClient("POST", url, reqHeader, reqBody, tools.DefaultHttpTimeout())
	logs.Debug("[GetHyruleResult] req:%s, res:%s", reqBody, string(restByte))
	if err != nil {
		logs.Error("[GetHyruleResult] has wrong. url:", url, ", err:", err)
		return
	}

	if code != 200 {
		logs.Error("[GetHyruleResult] code wrong. url:", url, ", data:", string(restByte))
	}

	json.Unmarshal(restByte, &res)

	return
}

func FillFantasyRiskRequest(info *RiskRequestInfo, oo *models.Order, uu *models.AccountBase, ap *models.AccountProfile, aci *models.ClientInfo) {
	detail := RiskRequestDetail{}
	FillFantasyRiskDetail(&detail, oo, uu, ap, aci)
	info.Data = append(info.Data, detail)
}

func FillFantasyRiskDetail(detail *RiskRequestDetail, oo *models.Order, uu *models.AccountBase, ap *models.AccountProfile, aci *models.ClientInfo) {
	adv := dao.GetFantasyAdvanceResponse("identity-check", oo.UserAccountId)
	adv_s := dao.GetFantasyAdvanceResponse("id-check", oo.UserAccountId)
	fd := dao.GetFantasyFaceidResponse("detect", oo.UserAccountId)
	ext, _ := models.OneAccountBaseExtByPkId(oo.UserAccountId)

	detail.OpUid = oo.OpUid
	detail.Loan = oo.Loan
	detail.Period = oo.Period
	detail.PenaltyUtime = oo.PenaltyUtime
	detail.CheckStatus = int(oo.CheckStatus)
	detail.ApplyTime = oo.ApplyTime
	detail.CheckTime = oo.CheckTime
	detail.RiskCtlStatus = int(oo.RiskCtlStatus)
	detail.RejectReason = int(oo.RejectReason)
	detail.PhoneVerifyAt = tools.GetDateMHS(oo.PhoneVerifyTime / 1000)
	detail.LoanTime = oo.LoanTime
	detail.ProductId = oo.ProductId
	detail.IsDeadDebt = oo.IsDeleted
	detail.PhoneVerifyTime = oo.PhoneVerifyTime
	detail.OrderId = oo.Id
	detail.RiskCtlFinishTime = oo.RiskCtlFinishTime
	detail.FinishTime = oo.FinishTime
	detail.UserAccountId = oo.UserAccountId
	detail.LoanAt = tools.GetDateMHS(oo.LoanTime / 1000)
	detail.RandomValue = oo.RandomValue
	detail.RandomMark = oo.RandomMark
	detail.CheckAt = tools.GetDateMHS(oo.CheckTime / 1000)
	detail.RiskCtlRegular = oo.RiskCtlRegular
	detail.IsOverdue = oo.IsOverdue
	detail.RepayTime = oo.RepayTime
	detail.ApplyAt = tools.GetDateMHS(oo.ApplyTime / 1000)
	detail.CreatedAt = tools.GetDateMHS(oo.Ctime / 1000)
	detail.FixedRandom = oo.FixedRandom
	detail.Amount = oo.Amount
	detail.IsTemporary = oo.IsTemporary

	detail.ThirdName = uu.ThirdName
	detail.ThirdVillage = uu.ThirdVillage
	detail.OcrRealname = uu.OcrRealname
	detail.Realname = uu.Realname
	detail.Gender = int(uu.Gender)
	detail.ThirdProvince = uu.ThirdProvince
	detail.OcrIdentity = uu.OcrIdentity
	detail.LastLoginTime = uu.LastLoginTime
	detail.AppsflyerId = uu.AppsflyerID
	detail.RegisterTime = uu.RegisterTime
	detail.ThirdCity = uu.ThirdCity
	detail.GoogleAdvertisingId = uu.GoogleAdvertisingID
	detail.Status = uu.Status
	detail.Tags = int(uu.Tags)
	detail.ThirdId = uu.ThirdID
	detail.Nickname = uu.Nickname
	detail.Identity = uu.Identity
	detail.ThirdDistrict = uu.ThirdDistrict
	detail.Mobile = uu.Mobile

	detail.ServiceType = tools.Int2Str(int(aci.ServiceType))
	detail.City = aci.City
	detail.Platform = aci.Platform
	detail.AppVersionCode = tools.Int2Str(aci.AppVersionCode)
	detail.Ip = aci.IP
	detail.Latitude = aci.Latitude
	detail.AppVersion = aci.AppVersion
	detail.IsSimulator = aci.IsSimulator
	detail.Brand = aci.Brand
	if aci.Imei != "" {
		detail.Imei = tools.Md5(aci.Imei)
	}
	detail.Longitude = aci.Longitude
	detail.Network = aci.Network
	detail.CreatedDt = tools.GetDateMHS(aci.Ctime / 1000)
	detail.TimeZone = aci.TimeZone
	detail.Model = aci.Model
	detail.Os = aci.OS

	detail.ResidentCity = ap.ResidentCity
	detail.CompanyCity = ap.CompanyCity
	detail.ChildrenNumber = ap.ChildrenNumber
	detail.ServiceYears = ap.ServiceYears
	detail.Contact2Name = ap.Contact2Name
	detail.CompanyName = ap.CompanyName
	detail.BankName = ap.BankName
	if ap.HandPhotoQualityThreshold == "" {
		detail.HandPhotoQualityThreshold = "0"
	} else {
		detail.HandPhotoQualityThreshold = ap.HandPhotoQualityThreshold
	}
	detail.IdPhoto = ap.IdPhoto
	detail.HandHeldIdPhoto = ap.HandHeldIdPhoto
	detail.ResidentAddress = ap.ResidentAddress
	detail.Relationship2 = ap.Relationship2
	detail.Relationship1 = ap.Relationship1
	detail.MaritalStatus = ap.MaritalStatus
	if ap.FaceComparison == "" {
		detail.FaceComparison = "0"
	} else {
		detail.FaceComparison = ap.FaceComparison
	}
	if ap.IdPhotoQualityThreshold == "" {
		detail.IdPhotoQualityThreshold = "0"
	} else {
		detail.IdPhotoQualityThreshold = ap.IdPhotoQualityThreshold
	}
	detail.Contact1Name = ap.Contact1Name
	detail.Education = ap.Education
	detail.MonthlyIncome = ap.MonthlyIncome
	if ap.HandPhotoQuality == "" {
		detail.HandPhotoQuality = "0"
	} else {
		detail.HandPhotoQuality = ap.HandPhotoQuality
	}
	if ap.IdPhotoQuality == "" {
		detail.IdPhotoQuality = "0"
	} else {
		detail.IdPhotoQuality = ap.IdPhotoQuality
	}
	detail.BankNo = ap.BankNo

	if ap.IdHoldingPhotoCheck == "" {
		detail.IdHoldingPhotoCheck = "0"
	} else {
		detail.IdHoldingPhotoCheck = ap.IdHoldingPhotoCheck
	}
	detail.JobType = ap.JobType
	detail.Contact1 = ap.Contact1
	detail.Contact2 = ap.Contact2
	detail.CompanyAddress = ap.CompanyAddress

	detail.IdentityResult = adv.Code
	detail.IdentityMessage = adv.Message

	detail.IdCheckResult = adv_s.Code
	detail.IdCheckMessage = adv_s.Message
	detail.IdCheckSimilarity = adv_s.Data.Similarity

	if len(fd.Data) > 0 {
		detail.FaceQuality = fd.Data[0].Quality
	}

	//UserAuthority
	if ext.RecallTag == types.RecallTagScore {
		author := UserAuthorityInfo{
			GoJek:     ext.AuthorizeStatusGoJek,
			Yys:       ext.AuthorizeStatusYys,
			Facebook:  ext.AuthorizeStatusFacebook,
			Lazada:    ext.AuthorizeStatusLazada,
			Linkedin:  ext.AuthorizeStatusLinkedin,
			Tokopedia: ext.AuthorizeStatusTokopedia,
			Instagram: ext.AuthorizeStatusInstagram,
		}
		detail.UserAuthority = author
	}
}

func FillFantasyFraudRequest(info *FraudRequestInfo, oo *models.Order, uu *models.AccountBase, aci *models.ClientInfo) {
	info.Imei = aci.Imei
	info.AccountId = oo.UserAccountId
	info.OrderId = oo.Id

	list := dao.GetFantasyClientInfo(oo.UserAccountId)
	for _, v := range list {
		detail := FraudRequestDetail{}
		FillFantasyFraudDetail(&detail, oo, uu, &v)
		info.Data = append(info.Data, detail)
	}
}

func FillFantasyFraudDetail(detail *FraudRequestDetail, oo *models.Order, uu *models.AccountBase, aci *models.ClientInfo) {
	detail.Id = aci.Id
	detail.Mobile = aci.Mobile
	detail.ServiceType = int(aci.ServiceType)
	detail.RelatedId = aci.RelatedId
	detail.Ip = aci.IP
	detail.Os = aci.OS
	detail.Imei = aci.Imei
	detail.Model = aci.Model
	detail.Brand = aci.Brand
	detail.AppVersion = aci.AppVersion
	detail.Longitude = aci.Longitude
	detail.Latitude = aci.Latitude
	detail.City = aci.City
	detail.TimeZone = aci.TimeZone
	detail.Network = aci.Network
	detail.IsSimulator = aci.IsSimulator
	detail.Platform = aci.Platform
	detail.Ctime = aci.Ctime
}

func FillFantasyGraphRequest(detail *GraphRequestInfo, oo *models.Order, uu *models.AccountBase, ap *models.AccountProfile, aci *models.ClientInfo) {
	detail.AccountId = uu.Id
	detail.OrderId = oo.Id
	detail.Imei = aci.ImeiMd5
	detail.Identity = uu.Identity
	detail.Realname = uu.Realname
	detail.Gender = int(uu.Gender)
	detail.Mobile = uu.Mobile
	detail.RegisterTime = uu.RegisterTime
	detail.MonthlyIncome = ap.MonthlyIncome
	detail.Education = ap.Education
	detail.MaritalStatus = ap.MaritalStatus
	detail.ChildrenNumber = ap.ChildrenNumber
	detail.Contact1 = ap.Contact1
	detail.Contact1Name = ap.Contact1Name
	detail.Relationship1 = ap.Relationship1
	detail.Contact2 = ap.Contact2
	detail.Contact2Name = ap.Contact2Name
	detail.Relationship2 = ap.Relationship2
	detail.CompanyMobile = ap.CompanyTelephone
	detail.CompanyName = ap.CompanyName
	detail.ServiceYears = ap.ServiceYears
	detail.JobType = ap.JobType
	detail.BankNo = ap.BankNo
	detail.BankName = ap.BankName
	detail.Ip = aci.IP
}

func SendFantasyGraphReq(accountId int64, scene string) {
	accountBase, _ := models.OneAccountBaseByPkId(accountId)
	accountProfile, _ := dao.CustomerProfile(accountId)
	order, _ := dao.AccountLastLoanOrder(accountId)
	clientInfo, _ := models.OneLastClientInfoByRelatedID(accountId)

	graphReq := GraphRequestInfo{}
	FillFantasyGraphRequest(&graphReq, &order, &accountBase, accountProfile, &clientInfo)
	graphReq.Model = "graph"
	graphReq.Version = "v1"
	graphReq.Scene = scene
	GetFantasyGraph(graphReq)
}

func FillHyruleRequestInfo(detail *HyruleRequestInfo, oo *models.Order, uu *models.AccountBase, ap *models.AccountProfile, aci *models.ClientInfo) {
	adv := dao.GetFantasyAdvanceResponse("identity-check", oo.UserAccountId)
	adv_s := dao.GetFantasyAdvanceResponse("id-check", oo.UserAccountId)
	fd := dao.GetFantasyFaceidResponse("detect", oo.UserAccountId)

	detail.BasicInfo.ChannelFrom = "app"
	detail.BasicInfo.AccountId = oo.UserAccountId
	detail.BasicInfo.OrderId = oo.Id
	if aci.Imei != "" {
		detail.BasicInfo.Imei = tools.Md5(aci.Imei)
	}
	detail.BasicInfo.Mobile = uu.Mobile
	detail.BasicInfo.Identity = uu.Identity
	detail.BasicInfo.Realname = uu.Realname

	detail.UserInfo.AccountId = oo.UserAccountId
	detail.UserInfo.AppsflyerId = uu.AppsflyerID
	detail.UserInfo.BankName = ap.BankName
	detail.UserInfo.BankNo = ap.BankNo
	detail.UserInfo.ChildrenNumber = ap.ChildrenNumber
	detail.UserInfo.CompanyAddress = ap.CompanyAddress
	detail.UserInfo.CompanyCity = ap.CompanyCity
	detail.UserInfo.CompanyMobile = ap.CompanyTelephone
	detail.UserInfo.CompanyName = ap.CompanyName
	detail.UserInfo.Contact1 = ap.Contact1
	detail.UserInfo.Contact1Name = ap.Contact1Name
	detail.UserInfo.Contact2 = ap.Contact2
	detail.UserInfo.Contact2Name = ap.Contact2Name
	detail.UserInfo.Education = ap.Education
	if ap.FaceComparison == "" {
		detail.UserInfo.FaceComparison = "0"
	} else {
		detail.UserInfo.FaceComparison = ap.FaceComparison
	}
	detail.UserInfo.Gender = int(uu.Gender)
	detail.UserInfo.GoogleAdvertisingId = uu.GoogleAdvertisingID
	detail.UserInfo.HandHeldIdPhoto = ap.HandHeldIdPhoto
	if ap.HandPhotoQuality == "" {
		detail.UserInfo.HandPhotoQuality = "0"
	} else {
		detail.UserInfo.HandPhotoQuality = ap.HandPhotoQuality
	}
	if ap.HandPhotoQualityThreshold == "" {
		detail.UserInfo.HandPhotoQualityThreshold = "0"
	} else {
		detail.UserInfo.HandPhotoQualityThreshold = ap.HandPhotoQualityThreshold
	}
	if ap.IdHoldingPhotoCheck == "" {
		detail.UserInfo.IdHoldingPhotoCheck = "0"
	} else {
		detail.UserInfo.IdHoldingPhotoCheck = ap.IdHoldingPhotoCheck
	}
	detail.UserInfo.IdPhoto = ap.IdPhoto
	if ap.IdPhotoQuality == "" {
		detail.UserInfo.IdPhotoQuality = "0"
	} else {
		detail.UserInfo.IdPhotoQuality = ap.IdPhotoQuality
	}
	if ap.IdPhotoQualityThreshold == "" {
		detail.UserInfo.IdPhotoQualityThreshold = "0"
	} else {
		detail.UserInfo.IdPhotoQualityThreshold = ap.IdPhotoQualityThreshold
	}
	detail.UserInfo.Identity = uu.Identity
	detail.UserInfo.JobType = ap.JobType
	detail.UserInfo.LastLoginTime = uu.LastLoginTime
	detail.UserInfo.MaritalStatus = ap.MaritalStatus
	detail.UserInfo.Mobile = uu.Mobile
	detail.UserInfo.MonthlyIncome = ap.MonthlyIncome
	detail.UserInfo.Nickname = uu.Nickname
	detail.UserInfo.OcrIdentity = uu.OcrIdentity
	detail.UserInfo.OcrRealname = uu.OcrRealname
	detail.UserInfo.Realname = uu.Realname
	detail.UserInfo.RegisterTime = uu.RegisterTime
	detail.UserInfo.Relationship1 = ap.Relationship1
	detail.UserInfo.Relationship2 = ap.Relationship2
	detail.UserInfo.ResidentAddress = ap.ResidentAddress
	detail.UserInfo.ResidentCity = ap.ResidentCity
	detail.UserInfo.SalaryDay = ap.SalaryDay
	detail.UserInfo.ServiceYears = ap.ServiceYears
	detail.UserInfo.Status = uu.Status
	detail.UserInfo.Tags = int(uu.Tags)
	detail.UserInfo.ThirdCity = uu.ThirdCity
	detail.UserInfo.ThirdDistrict = uu.ThirdDistrict
	detail.UserInfo.ThirdId = uu.ThirdID
	detail.UserInfo.ThirdName = uu.ThirdName
	detail.UserInfo.ThirdProvince = uu.ThirdProvince
	detail.UserInfo.ThirdVillage = uu.ThirdVillage

	detail.OrderInfo.Amount = oo.Amount
	detail.OrderInfo.ApplyTime = oo.ApplyTime
	detail.OrderInfo.CheckStatus = int(oo.CheckStatus)
	detail.OrderInfo.CheckTime = oo.CheckTime
	detail.OrderInfo.Ctime = oo.Ctime
	detail.OrderInfo.FinishTime = oo.FinishTime
	detail.OrderInfo.FixedRandom = oo.FixedRandom
	detail.OrderInfo.IsDeadDebt = oo.IsDeadDebt
	detail.OrderInfo.IsOverdue = oo.IsOverdue
	detail.OrderInfo.IsReloan = oo.IsReloan
	detail.OrderInfo.IsTemporary = oo.IsTemporary
	detail.OrderInfo.Loan = oo.Loan
	detail.OrderInfo.LoanOrg = oo.LoanOrg
	detail.OrderInfo.LoanTime = oo.LoanTime
	detail.OrderInfo.OpUid = oo.OpUid
	detail.OrderInfo.OrderId = oo.Id
	detail.OrderInfo.PenaltyUtime = oo.PenaltyUtime
	detail.OrderInfo.Period = oo.Period
	detail.OrderInfo.PeriodOrg = oo.PeriodOrg
	detail.OrderInfo.PhoneVerifyTime = oo.PhoneVerifyTime
	detail.OrderInfo.ProductId = oo.ProductId
	detail.OrderInfo.RandomMark = oo.RandomMark
	detail.OrderInfo.RandomValue = oo.RandomValue
	detail.OrderInfo.RejectReason = int(oo.RejectReason)
	detail.OrderInfo.RepayTime = oo.RepayTime
	detail.OrderInfo.RiskCtlFinishTime = oo.RiskCtlFinishTime
	detail.OrderInfo.RiskCtlRegular = oo.RiskCtlRegular
	detail.OrderInfo.RiskCtlStatus = int(oo.RiskCtlStatus)

	detail.DeviceInfo.AppVersion = aci.AppVersion
	detail.DeviceInfo.AppVersionCode = tools.Int2Str(aci.AppVersionCode)
	detail.DeviceInfo.Brand = aci.Brand
	detail.DeviceInfo.City = aci.City
	detail.DeviceInfo.Ctime = aci.Ctime
	if aci.Imei != "" {
		detail.DeviceInfo.Imei = tools.Md5(aci.Imei)
	}
	detail.DeviceInfo.Ip = aci.IP
	detail.DeviceInfo.IsSimulator = aci.IsSimulator
	detail.DeviceInfo.Latitude = aci.Latitude
	detail.DeviceInfo.Longitude = aci.Longitude
	detail.DeviceInfo.Model = aci.Model
	detail.DeviceInfo.Network = aci.Network
	detail.DeviceInfo.Os = aci.OS
	detail.DeviceInfo.Platform = aci.Platform
	detail.DeviceInfo.RelatedId = aci.RelatedId
	detail.DeviceInfo.ServiceType = int(aci.ServiceType)

	detail.ExtraInfo.AccountId = oo.UserAccountId
	if len(fd.Data) > 0 {
		detail.ExtraInfo.FaceQuality = fd.Data[0].Quality
	}
	detail.ExtraInfo.IdCheckMessage = adv_s.Message
	detail.ExtraInfo.IdCheckResult = adv_s.Code
	detail.ExtraInfo.IdCheckSimilarity = adv_s.Data.Similarity
	detail.ExtraInfo.IdentityMessage = adv.Message
	detail.ExtraInfo.IdentityResult = adv.Code

	orders, _, _ := dao.AccountAllOrders(oo.UserAccountId)
	for _, v := range orders {
		if v.IsTemporary == int(types.IsTemporaryYes) {
			continue
		}

		if v.RollTimes == 0 {
			detail.UserInfo.ApplyOrderNum++
			if v.LoanTime > 0 {
				detail.UserInfo.LoanOrderNum++
			}
		}

		overdueDays := int(CalculateOverdueDays(&v))
		if overdueDays > 0 {
			detail.UserInfo.AccuOverdueNum++
		}

		if overdueDays > detail.UserInfo.MaxOverdueDays {
			detail.UserInfo.MaxOverdueDays = overdueDays
		}

		detail.UserInfo.TotalOverdueDays += overdueDays
		if v.CheckStatus == types.LoanStatusOverdue {
			detail.UserInfo.OverdueStatus = 1
		}
	}
}
