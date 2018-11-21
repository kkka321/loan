package fantasy

import (
	"encoding/json"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

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

func FillFantasyFraudRequest(info *FraudRequestInfo, oo *models.Order, uu *models.AccountBase, aci models.ClientInfo) {
	info.Imei = aci.Imei
	info.AccountId = oo.UserAccountId
	info.OrderId = oo.Id

	list := dao.GetFantasyClientInfo(oo.UserAccountId)
	for _, v := range list {
		detail := FraudRequestDetail{}
		FillFantasyFraudDetail(&detail, oo, uu, v)
		info.Data = append(info.Data, detail)
	}
}

func FillFantasyFraudDetail(detail *FraudRequestDetail, oo *models.Order, uu *models.AccountBase, aci models.ClientInfo) {
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
