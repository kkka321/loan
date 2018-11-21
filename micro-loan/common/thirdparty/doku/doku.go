// docs: https://dashboard.xendit.co/docs/introduction
// https://github.com/xendit

package doku

import (
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/satori/go.uuid"

	"micro-loan/common/dao"
	"micro-loan/common/lib/device"
	"micro-loan/common/lib/payment"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type DokuApi struct {
	payment.PaymentApi

	HandleDisburseCallback func(payType int, dataOrder *models.Order, bankCode string, isRoll bool) error
	HandleLoanFailCallback func(order *models.Order, err error)
}

type InquiryMainContent struct {
	XMLName          xml.Name `xml:"INQUIRY_RESPONSE"`
	PaymentCode      string   `xml:"PAYMENTCODE"`
	Amount           string   `xml:"AMOUNT"`
	PurchaseAmount   string   `xml:"PURCHASEAMOUNT"`
	MinAmount        string   `xml:"MINAMOUNT"`
	MaxAmount        string   `xml:"MAXAMOUNT"`
	TransidMerchant  string   `xml:"TRANSIDMERCHANT"`
	Words            string   `xml:"WORDS"`
	RequestDateTime  string   `xml:"REQUESTDATETIME"`
	Currency         string   `xml:"CURRENCY"`
	PurchaseCurrency string   `xml:"PURCHASECURRENCY"`
	SessionId        string   `xml:"SESSIONID"`
	Name             string   `xml:"NAME"`
	Email            string   `xml:"EMAIL"`
	Basket           string   `xml:"BASKET"`
	AdditionalData   string   `xml:"ADDITIONALDATA"`
	ResponseCode     string   `xml:"RESPONSECODE"`
}

type DokuRemitResp struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Remit   struct {
		PaymentData struct {
			MallId        string `json:"mallId"`
			AccountNumber string `json:"accountNumber"`
			AccountName   string `json:"accountName"`
			ChannelCode   string `json:"channelCode"`
			InquiryId     string `json:"inquiryId"`
			Currency      string `json:"currency"`
			Amount        string `json:"amount"`
			TrxCode       string `json:"trxCode"`
			ResponseCode  string `json:"responseCode"`
			ResponseMsg   string `json:"responseMsg"`
		} `json:"paymentData"`
		TransactionId string `json:"transactionId"`
	} `json:"remit"`
}

type DokuRemitReq struct {
	RequestId string `json:"requestId"`
	AgentKey  string `json:"agentKey"`
	Signature string `json:"signature"`
	SendType  string `json:"sendType"`
	//Auth1       string `json:"auth1"`
	BeneficiaryAmount int64 `json:"beneficiaryAmount"`
	Beneficiary       struct {
		Address string `json:"address"`
		Country struct {
			Code string `json:"code"`
		} `json:"country"`
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		PhoneNumber string `json:"phoneNumber"`
	} `json:"beneficiary"`
	BeneficiaryAccount struct {
		Address string `json:"address"`
		Bank    struct {
			Code        string `json:"code"`
			CountryCode string `json:"countryCode"`
			Id          string `json:"id"`
			Name        string `json:"name"`
		} `json:"bank"`
		City   string `json:"city"`
		Name   string `json:"name"`
		Number string `json:"number"`
	} `json:"beneficiaryAccount"`
	BeneficiaryCity    string `json:"beneficiaryCity"`
	BeneficiaryCountry struct {
		Code string `json:"code"`
	} `json:"beneficiaryCountry"`
	BeneficiaryCurrency struct {
		Code string `json:"code"`
	} `json:"beneficiaryCurrency"`
	Channel struct {
		Code string `json:"code"`
	} `json:"channel"`
	Inquiry struct {
		IdToken string `json:"idToken"`
	} `json:"inquiry"`
	Sender struct {
		Address   string `json:"address"`
		BirthDate string `json:"birthDate"`
		Country   struct {
			Code string `json:"code"`
		} `json:"country"`
		FirstName         string `json:"firstName"`
		Gender            string `json:"gender"`
		LastName          string `json:"lastName"`
		PersonalId        string `json:"personalId"`
		PersonalIdCountry struct {
			Code string `json:"code"`
		} `json:"personalIdCountry"`
		PersonalIdExpireDate string `json:"personalIdExpireDate"`
		PersonalIdIssueDate  string `json:"personalIdIssueDate"`
		PersonalIdType       string `json:"personalIdType"`
		PhoneNumber          string `json:"phoneNumber"`
	} `json:"sender"`
	//SenderAmount  string `json:"senderAmount"`
	SenderCountry struct {
		Code string `json:"code"`
	} `json:"senderCountry"`
	SenderCurrency struct {
		Code string `json:"code"`
	} `json:"senderCurrency"`
	SenderNote string `json:"senderNote"`
}

//key用xendit银行列表值
var commonBankNameCodeMap = map[string]string{
	"Bank Danamon":                "BDINIDJA",
	"Bank Permata":                "BBBAIDJA",
	"Bank Central Asia (BCA)":     "CENAIDJA",
	"Bank CIMB Niaga":             "BNIAIDJA",
	"Bank Mandiri":                "BMRIIDJA",
	"Bank Negara Indonesia (BNI)": "BNINIDJA",
	"Bank Rakyat Indonesia (BRI)": "BRINIDJA",
}

var bankNameCodeMap = map[string]string{
	"Bank BRI":                   "BRINIDJA",
	"Bank Mandiri":               "BMRIIDJA",
	"Bank BNI":                   "BNINIDJA",
	"Bank Danamon":               "BDINIDJA",
	"Bank Permata":               "BBBAIDJA",
	"Bank Central Asia (BCA)":    "CENAIDJA",
	"BII Maybank":                "IBBKIDJA",
	"Bank Panin":                 "PINBIDJA",
	"CIMB Niaga":                 "BNIAIDJA",
	"Bank UOB Buana":             "BBIJIDJA",
	"Bank Arta Graha":            "ARTGIDJA",
	"ANZ Panin":                  "ANZBIDJX",
	"Bank Jabar Banten (BJB)":    "PDJBIDJA",
	"Bank Jatim":                 "PDJTIDJ1",
	"Bank Nusantara Parahyangan": "NUPAIDJ6",
	"Bank Muamalat Indonesia":    "MUABIDJA",
	"Bank Sinarmas":              "SBJKIDJA",
	"Bank BTN":                   "BTANIDJA",
	"Bank BTPN":                  "TAPEIDJ1",
	"Bank BRI Syariah":           "DJARIDJ1",
	"Bank BJB Syariah":           "SYJBIDJ1",
	"Bank Mega":                  "MEGAIDJA",
	"Bank BNI Syariah":           "SYNIIDJ1",
	"Bank Bukopin":               "BBUKIDJA",
	"Bank Syariah Mandiri":       "SYMDIDJ1",
	"Bank Hana":                  "HNBNIDJA",
	"Bank Syariah Mega":          "BUTGIDJ1",
	"Bank BCA Syariah":           "SYCAIDJ1",
	"Bank Centratama Nasional":   "CNBAIDJ1",
	"DOKU":                                                 "899",
	"Maybank Syariah":                                      "MBBEIDJA",
	"Bank Commonwealth":                                    "BICNIDJA",
	"(MA LAI XI YA DI QU) MA LAI XI YA ZHONG GUO YIN HANG": "989584028209",
	"xiong ya li di qu xiong ya li zhong guo yin hang":     "989584029009",
}

var bankNameBankIdMap = map[string]string{
	"Bank BRI":                   "002",
	"Bank Mandiri":               "008",
	"Bank BNI":                   "009",
	"Bank Danamon":               "011",
	"Bank Permata":               "013",
	"Bank Central Asia (BCA)":    "014",
	"BII Maybank":                "016",
	"Bank Panin":                 "019",
	"CIMB Niaga":                 "022",
	"Bank UOB Buana":             "023",
	"Bank Arta Graha":            "037",
	"ANZ Panin":                  "061",
	"Bank Jabar Banten (BJB)":    "110",
	"Bank Jatim":                 "114",
	"Bank Nusantara Parahyangan": "145",
	"Bank Muamalat Indonesia":    "147",
	"Bank Sinarmas":              "153",
	"Bank BTN":                   "200",
	"Bank BTPN":                  "213",
	"Bank BRI Syariah":           "422",
	"Bank BJB Syariah":           "425",
	"Bank Mega":                  "426",
	"Bank BNI Syariah":           "427",
	"Bank Bukopin":               "441",
	"Bank Syariah Mandiri":       "451",
	"Bank Hana":                  "484",
	"Bank Syariah Mega":          "506",
	"Bank BCA Syariah":           "536",
	"Bank Centratama Nasional":   "559",
	"DOKU":                                                 "899",
	"Maybank Syariah":                                      "947",
	"Bank Commonwealth":                                    "950",
	"(MA LAI XI YA DI QU) MA LAI XI YA ZHONG GUO YIN HANG": "989584028209",
	"xiong ya li di qu xiong ya li zhong guo yin hang":     "989584029009",
}

var bankXenditDokuBankNameMap = map[string]string{
	"Bank Rakyat Indonesia (BRI)":    "Bank BRI",
	"Bank Mandiri":                   "Bank Mandiri",
	"Bank Negara Indonesia (BNI)":    "Bank BNI",
	"Bank Danamon":                   "Bank Danamon",
	"Bank Permata":                   "Bank Permata",
	"Bank Central Asia (BCA)":        "Bank Central Asia (BCA)",
	"Bank Maybank":                   "BII Maybank",
	"Bank Panin":                     "Bank Panin",
	"Bank CIMB Niaga":                "CIMB Niaga",
	"Bank UOB Indonesia":             "Bank UOB Buana",
	"Bank Artha Graha International": "Bank Arta Graha",
	//"":"ANZ Panin",
	"Bank BJB": "Bank Jabar Banten (BJB)",
	//"":	"Jatim",
	"Bank Nusantara Parahyangan": "Bank Nusantara Parahyangan",
	"Bank Muamalat Indonesia":    "Bank Muamalat Indonesia",
	"Sinarmas":                   "Bank Sinarmas",
	"Bank Tabungan Negara (BTN)": "Bank BTN",
	//"": "Bank BTPN",
	"Bank Syariah BRI":                "Bank BRI Syariah",
	"Bank BJB Syariah":                "Bank BJB Syariah",
	"Bank Mega":                       "Bank Mega",
	"Bank BNI Syariah":                "Bank BNI Syariah",
	"Bank Bukopin":                    "Bank Bukopin",
	"Bank Syariah Mandiri":            "Bank Syariah Mandiri",
	"Bank Hana":                       "Bank Hana",
	"Bank Syariah Mega":               "Bank Syariah Mega",
	"Bank Central Asia (BCA) Syariah": "Bank BCA Syariah",
	"Centratama Nasional Bank":        "Bank Centratama Nasional",
	//"": "DOKU",
	"Bank Maybank Syariah Indonesia": "Maybank Syariah",
	"Bank Commonwealth":              "Bank Commonwealth",
	//"":"(MA LAI XI YA DI QU) MA LAI XI YA ZHONG GUO YIN HANG",
	//"":"xiong ya li di qu xiong ya li zhong guo yin hang"
}

const (
	// xendit
	PERMATA = "PERMATA" // BBBAIDJA (doku)
	DANAMON = "DANAMON" // BDINIDJA (doku)
	CIMB    = "CIMB"    // BNIAIDJA (doku)
	BCA     = "BCA"

	// doku
	BBBAIDJA = "BBBAIDJA" // PERMATA (xendit)
	BDINIDJA = "BDINIDJA" // DANAMON (xendit)
	BNIAIDJA = "BNIAIDJA" // CIMB (xendit)
	CENAIDJA = "CENAIDJA" //BCA
)

var doKuVaBankCodeToXenditMap = map[string]string{
	"BBBAIDJA": "PERMATA",
	"BDINIDJA": "DANAMON",
	"BNIAIDJA": "CIMB",
	"CENAIDJA": "BCA",
}

func DoKuVaBankCodeToXenditMap() map[string]string {
	return doKuVaBankCodeToXenditMap
}

//将doku简码转换为xendit简码,这样用户看起来
func DoKuVaBankCodeTransform(dokuBankCode string) (xenditBankCode string) {
	xenditCode := DoKuVaBankCodeToXenditMap()
	if v, ok := xenditCode[dokuBankCode]; ok {
		xenditBankCode = v
		return
	}
	return
}

var xenditVaBankCodeToDokuMap = map[string]string{
	PERMATA: BBBAIDJA,
	DANAMON: BDINIDJA,
	CIMB:    BNIAIDJA,
	BCA:     CENAIDJA,
}

func XenditVaBankCodeToDoKuMap() map[string]string {
	return xenditVaBankCodeToDokuMap
}

//将xendit简码转换为doku简码
func XenditVaBankCodeTransform(xenditBankCode string) (dokuBankCode string) {
	doku := XenditVaBankCodeToDoKuMap()
	if v, ok := doku[xenditBankCode]; ok {
		dokuBankCode = v
		return
	}
	return
}

//func CommonBankNameCodeMap() map[string]string {
//	return commonBankNameCodeMap
//}

func BankNameCodeMap() map[string]string {
	return bankNameCodeMap
}

func BankName2Code(name string) (code string, err error) {
	bankNameCodeMap := BankNameCodeMap()
	if v, ok := bankNameCodeMap[name]; ok {
		code = v
		return
	}

	err = fmt.Errorf("doku BankName2Code bank code undefined")

	return
}

//func CommonBankName2Code(name string) (code string, err error) {
//	commonBankNameCodeMap := CommonBankNameCodeMap()
//	if v, ok := commonBankNameCodeMap[name]; ok {
//		code = v
//		return
//	}
//
//	err = fmt.Errorf("doku bank code undefined")
//
//	return
//}

func GetBankXenditDokuBankMap() map[string]string {
	return bankXenditDokuBankNameMap
}

func GetBankXenditDokuBandIdMap() map[string]string {
	return bankNameBankIdMap
}

//func BankXenditDokuName2Code(name string) (code string, err error) {
//	bankNameCodeMap := GetBankXenditDokuBankMap()
//	if v, ok := bankNameCodeMap[name]; ok {
//		code = v
//		return
//	}
//
//	err = fmt.Errorf("doku bank code undefined")
//
//	return
//}

//
//func BankXenditDokuId2Code(name string) (id string, err error) {
//	dokuBankCodeMap := GetBankXenditDokuBankMap()
//	dokuBankIdMap := GetBankXenditDokuBandIdMap()
//
//	if v, ok := dokuBankCodeMap[name]; ok {
//		if id, ok = dokuBankIdMap[v]; ok {
//			return
//		}
//	}
//
//	err = fmt.Errorf("doku bank id undefined")
//
//	return
//}

//
//func GetDoKuVABankCode(bankName string) (bankCode string, err error) {
//
//	defaultVAconf := map[string]bool{
//		"Bank Central Asia (BCA)": true,
//	}
//
//	if defaultVAconf[bankName] {
//		bankName = "Bank Permata"
//	}
//
//	conf := map[string]bool{
//		"Bank Danamon": true,
//		"Bank Permata": true,
//		"CIMB Niaga":   true,
//	}
//
//	bankName, err = BankXenditDokuName2Code(bankName)
//
//	if err != nil {
//		return
//	}
//
//	if !conf[bankName] {
//		err = fmt.Errorf("not in doku bank priority.")
//		logs.Info(err)
//		return
//	}
//
//	bankCode, err = BankName2Code(bankName)
//	if err != nil {
//		logs.Info(err)
//	}
//
//	return
//}

func VABankCode(info models.BanksInfo) (bankCode string, err error) {
	conf := map[string]bool{
		"BDINIDJA": true,
		"BBBAIDJA": true,
		"BNIAIDJA": true,
		"CENAIDJA": true,
	}

	// 还款时才用到va 而还款不分支持不支持
	//if len(info.DokuBrevityName) == 0 {
	//	err = fmt.Errorf("[VABankCode]  DokuBrevityName err. info:%#v", info)
	//	logs.Error(err)
	//	return
	//}

	if !conf[info.DokuBrevityName] {
		bankCode = "BBBAIDJA" //Bank Permata
	} else {
		bankCode = info.DokuBrevityName
	}
	return
}

/**
此方法仅用于放款
*/
//func GetDoKuDisburseBankCode(bankName string) (bankCode string, err error) {
//
//	conf := map[string]bool{
//		"Bank Danamon": true,
//		"Bank Permata": true,
//		"CIMB Niaga":   true,
//	}
//
//	bankName, err = BankXenditDokuName2Code(bankName)
//
//	if err != nil {
//		return
//	}
//
//	if !conf[bankName] {
//		err = fmt.Errorf("not in doku bank priority.")
//		logs.Info(err)
//		return
//	}
//
//	bankCode, err = BankName2Code(bankName)
//	if err != nil {
//		logs.Info(err)
//	}
//
//	return
//}

func GetDokuVAPrefix(bankCode string) (prefix string) {

	prefix = beego.AppConfig.String("doku_va_permata_prefix")
	//默认使用permata bank code

	switch bankCode {
	case "BBBAIDJA":
		//PERMATA
		prefix = beego.AppConfig.String("doku_va_permata_prefix")
	case "BDINIDJA":
		//DANAMON
		prefix = beego.AppConfig.String("doku_va_danamon_prefix")
	case "BNIAIDJA":
		//CIMB
		prefix = beego.AppConfig.String("doku_va_cimb_prefix")
	case "CENAIDJA":
		//bca
		prefix = beego.AppConfig.String("doku_va_bca_prefix")
	}

	return
}

func GenerateDoKuVA(bankCode string) (va string) {

	prefix := GetDokuVAPrefix(bankCode)
	bank := DoKuVaBankCodeTransform(bankCode)
	dokuVANumberKey := beego.AppConfig.String("doku_va_number")
	dokuBankVAKey := fmt.Sprintf("%s:%s", dokuVANumberKey, bank)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	dokuVANumber, err := storageClient.Do("INCR", dokuBankVAKey)
	if err != nil || dokuVANumber == nil {
		logs.Error("Generate VA sequence error %v", err.Error())
		return
	}

	va = fmt.Sprintf("%s%d", prefix, dokuVANumber)
	logs.Debug("doku va is ", va)
	return
}

func CheckDoKuVAExist(va string) (eAccountNumber models.User_E_Account, err error) {
	eAccountNumber, err = models.GetEAccountByENumber(va)
	return
}

func CheckVAWords(mallId string, paymentCode string, words string) (hash string, err error) {

	shardkey := beego.AppConfig.String("doku_shared_key")
	str := fmt.Sprintf("%s%s%s", mallId, shardkey, paymentCode)
	hash = tools.Sha1(str)

	if hash != words {
		err = fmt.Errorf("DoKu VA words mismatches.")
	}
	return
}

func CheckRepayVAWords(amount string, transIdMerchant string, resultMsg string, verifyStatus string, words string) (hash string, err error) {
	mallId := beego.AppConfig.String("doku_mallid")
	sharedKey := beego.AppConfig.String("doku_shared_key")

	str := fmt.Sprintf("%s%s%s%s%s%s", amount, mallId, sharedKey, transIdMerchant, resultMsg, verifyStatus)
	hash = tools.Sha1(str)

	if hash != words {
		logs.Error("DoKu RepayVA words mismatches.")
	}
	return
}

func CheckValidOrder(va string) (order models.Order, err error) {
	userEAccount, err := models.GetEAccountByENumber(va)
	if err != nil {
		logs.Error("DoKu userEAccount does not exist. orderId is %d", order.Id)
		return
	}

	order, err = dao.AccountLastLoanOrder(userEAccount.UserAccountId)
	if err != nil {
		logs.Error("DoKu LastLoanOrder does not exist. orderId is %d", order.Id)
	}
	return
}

func (c *DokuApi) CreateVirtualAccount(datas map[string]interface{}) (res []byte, err error) {
	//bankName := datas["bank_name"].(string)
	//name := datas["account_name"].(string)
	accountId := datas["account_id"].(int64)
	bankInfo := datas["banks_info"].(models.BanksInfo)

	//bankCode, err := GetDoKuVABankCode(bankName)
	bankCode, err := VABankCode(bankInfo)
	if err != nil {
		return
	}

	paymentCode := GenerateDoKuVA(bankCode)

	_, err = models.GetEAccountByENumber(paymentCode)

	if err != nil {
		//不存在则创建
		eAccount := models.User_E_Account{}
		eAccount.Id, _ = device.GenerateBizId(types.UserEAccountBiz)
		eAccount.UserAccountId = accountId
		eAccount.VaCompanyCode = types.DoKu
		eAccount.EAccountNumber = paymentCode
		eAccount.Status = "ACTIVE"
		eAccount.BankCode = bankCode
		eAccount.RepayBankCode = bankCode
		//DoKu 流程和其他流程不一样
		//对方服务器只有在用户还款的时候，才来回调我们
		//所以默认状体都是ACTIVE,不然我们无法放款
		eAccount.Ctime = tools.GetUnixMillis()
		eAccount.Utime = tools.GetUnixMillis()
		eAccount.IsClosed = 0
		_, err = eAccount.AddEAccount(&eAccount)
	}
	return
}

func (c *DokuApi) CheckVirtualAccount(datas map[string]interface{}) (res []byte, err error) {
	return []byte{}, err
}

func (c *DokuApi) Disburse(datas map[string]interface{}) (res []byte, err error) {

	orderId := datas["order_id"].(int64)
	accountId := datas["account_id"].(int64)
	invokeId := datas["invoke_id"].(int64)
	//accountIdStr := tools.Int642Str(accountId)
	bankName := datas["bank_name"].(string)
	accountHolderName := datas["account_name"].(string)
	accountNumber := datas["account_num"].(string)
	//desc := datas["desc"].(string)
	amount := datas["amount"].(int64)
	invoke, _ := models.OneDisburseInvorkLogByPkId(invokeId)

	bankInfo := datas["banks_info"].(models.BanksInfo)
	if len(bankInfo.DokuBrevityName) == 0 ||
		len(bankInfo.DokuFullName) == 0 ||
		len(bankInfo.DokuBrevityId) == 0 {
		err = fmt.Errorf("[Disburse] doku unsport bank. accountId:%d datas:%#v", accountId, datas)
		return
	}

	bankCode := bankInfo.DokuBrevityName
	dokuBankName := bankInfo.DokuFullName
	bankId := bankInfo.DokuBrevityId

	inquiryUrl := beego.AppConfig.String("doku_disburse_inquiry")
	remitUrl := beego.AppConfig.String("doku_disburse_remit")
	agentKey := beego.AppConfig.String("doku_agent_key")
	dokuEncryptionKsey := beego.AppConfig.String("doku_encryption_key")
	country := beego.AppConfig.String("doku_indonesia_country")
	currency := beego.AppConfig.String("doku_indonesia_currency")
	channelCode := beego.AppConfig.String("doku_indonesia_channel_code")
	defaultCity := beego.AppConfig.String("doku_default_city")
	companyName := beego.AppConfig.String("doku_company_name")
	companyBirday := beego.AppConfig.String("doku_company_birthday")
	companyMobile := beego.AppConfig.String("doku_company_mobile")
	personalId := beego.AppConfig.String("doku_company_personal_id")
	personalIdType := beego.AppConfig.String("doku_personal_id_type")
	senderType := beego.AppConfig.String("doku_sender_type")
	//auth := beego.AppConfig.String("doku_auth1")
	uuidStr := uuid.Must(uuid.NewV4())
	requestId := uuidStr.String()
	signature := fmt.Sprintf("%s%s", agentKey, requestId)
	signature = tools.AesEncryptECBUrlEncode(signature, dokuEncryptionKsey)
	reqFormat := "agentKey=%s&requestId=%s&signature=%s&senderCountry.code=%s&senderCurrency.code=%s"
	reqFormat += "&beneficiaryCountry.code=%s&beneficiaryCurrency.code=%s&channel.code=%s&senderAmount=%d"
	reqFormat += "&beneficiaryAccount.bank.code=%s&beneficiaryAccount.bank.countryCode=%s"
	reqFormat += "&beneficiaryAccount.bank.id=%s&beneficiaryAccount.bank.name=%s"
	reqFormat += "&beneficiaryAccount.city=%s&beneficiaryAccount.name=%s&beneficiaryAccount.number=%s"

	reqData := fmt.Sprintf(reqFormat, agentKey, requestId, signature, country, currency, country, currency, channelCode, amount, bankCode, country, bankId, bankName, defaultCity, accountHolderName, accountNumber)
	logs.Debug(reqData)
	reqHeaders := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	httpBody, httpStatusCode, err := tools.SimpleHttpClient("POST", inquiryUrl, reqHeaders, reqData, tools.DefaultHttpTimeout())

	monitor.IncrThirdpartyCount(models.ThirdpartyDoKu, httpStatusCode)

	models.AddOneThirdpartyRecord(models.ThirdpartyDoKu, inquiryUrl, orderId, reqData, string(httpBody), 0, 0, httpStatusCode)

	if err != nil {
		//inquiry超时处理，直接置为放款失败
		logs.Error("[DoKuRemitInquiryError] httpBody: %s, httpStatusCode: %d, err: %v\n", httpBody, httpStatusCode, err)
		invoke.FailureCode = "Inquiry Timeout Fail"
		invoke.DisbureStatus = types.DisbureStatusCallFailed
		invoke.Utime = tools.GetUnixMillis()
		cols := []string{"disbure_status", "utime", "failure_code"}
		models.OrmUpdate(&invoke, cols)
		return
	}

	var dokuDisburseInquiryResponse struct {
		Status  interface{} `json:"status"`
		Message string      `json:"message"`
		Inquiry struct {
			IdToken string `json:"idToken"`
		}
	}

	errJsonUnmarsha := json.Unmarshal(httpBody, &dokuDisburseInquiryResponse)

	if errJsonUnmarsha != nil {
		err = fmt.Errorf("[DokuDisburse err InquiryResponse json.Unmarshal]: %s", errJsonUnmarsha.Error())
		return
	}

	if _, ok := dokuDisburseInquiryResponse.Status.(float64); !ok {
		//我草你妈，我记住你了
		invoke.Utime = tools.GetUnixMillis()
		invoke.DisbureStatus = types.DisbureStatusCallFailed
		invoke.HttpCode = httpStatusCode
		invoke.FailureCode = dokuDisburseInquiryResponse.Message
		cols := []string{"disbure_status", "http_code", "utime", "failure_code"}
		models.OrmUpdate(&invoke, cols)
		err = fmt.Errorf("[DokuDisburse err InquiryResponse] status err %s", string(httpBody))
		return
	}

	var dokuRemitReq struct {
		RequestId string `json:"requestId"`
		AgentKey  string `json:"agentKey"`
		Signature string `json:"signature"`
		SendType  string `json:"sendType"`
		//Auth1       string `json:"auth1"`
		BeneficiaryAmount int64 `json:"beneficiaryAmount"`
		Beneficiary       struct {
			Address string `json:"address"`
			Country struct {
				Code string `json:"code"`
			} `json:"country"`
			FirstName   string `json:"firstName"`
			LastName    string `json:"lastName"`
			PhoneNumber string `json:"phoneNumber"`
		} `json:"beneficiary"`
		BeneficiaryAccount struct {
			Address string `json:"address"`
			Bank    struct {
				Code        string `json:"code"`
				CountryCode string `json:"countryCode"`
				Id          string `json:"id"`
				Name        string `json:"name"`
			} `json:"bank"`
			City   string `json:"city"`
			Name   string `json:"name"`
			Number string `json:"number"`
		} `json:"beneficiaryAccount"`
		BeneficiaryCity    string `json:"beneficiaryCity"`
		BeneficiaryCountry struct {
			Code string `json:"code"`
		} `json:"beneficiaryCountry"`
		BeneficiaryCurrency struct {
			Code string `json:"code"`
		} `json:"beneficiaryCurrency"`
		Channel struct {
			Code string `json:"code"`
		} `json:"channel"`
		Inquiry struct {
			IdToken string `json:"idToken"`
		} `json:"inquiry"`
		Sender struct {
			Address   string `json:"address"`
			BirthDate string `json:"birthDate"`
			Country   struct {
				Code string `json:"code"`
			} `json:"country"`
			FirstName         string `json:"firstName"`
			Gender            string `json:"gender"`
			LastName          string `json:"lastName"`
			PersonalId        string `json:"personalId"`
			PersonalIdCountry struct {
				Code string `json:"code"`
			} `json:"personalIdCountry"`
			PersonalIdExpireDate string `json:"personalIdExpireDate"`
			PersonalIdIssueDate  string `json:"personalIdIssueDate"`
			PersonalIdType       string `json:"personalIdType"`
			PhoneNumber          string `json:"phoneNumber"`
		} `json:"sender"`
		//SenderAmount  string `json:"senderAmount"`
		SenderCountry struct {
			Code string `json:"code"`
		} `json:"senderCountry"`
		SenderCurrency struct {
			Code string `json:"code"`
		} `json:"senderCurrency"`
		SenderNote string `json:"senderNote"`
	}

	accountInfo, _ := dao.CustomerOne(accountId)
	logs.Debug(accountInfo)

	dokuRemitReq.SendType = senderType //TODO
	//dokuRemitReq.Auth1 = auth          //TODO
	dokuRemitReq.Beneficiary.Address = defaultCity
	dokuRemitReq.BeneficiaryAmount = amount
	dokuRemitReq.Beneficiary.Country.Code = country
	dokuRemitReq.BeneficiaryCity = defaultCity
	dokuRemitReq.Beneficiary.FirstName = accountHolderName //realname
	dokuRemitReq.Beneficiary.LastName = accountHolderName  //realname
	dokuRemitReq.Beneficiary.PhoneNumber = accountInfo.Mobile
	dokuRemitReq.BeneficiaryAccount.Address = defaultCity //default
	dokuRemitReq.BeneficiaryAccount.Bank.Code = bankCode
	dokuRemitReq.BeneficiaryAccount.Bank.CountryCode = country
	dokuRemitReq.BeneficiaryAccount.Bank.Id = bankId
	dokuRemitReq.BeneficiaryAccount.Bank.Name = dokuBankName
	dokuRemitReq.BeneficiaryAccount.City = defaultCity //default
	dokuRemitReq.BeneficiaryAccount.Name = accountHolderName
	dokuRemitReq.BeneficiaryAccount.Number = accountNumber
	dokuRemitReq.BeneficiaryCountry.Code = country
	dokuRemitReq.BeneficiaryCurrency.Code = currency
	dokuRemitReq.Channel.Code = channelCode
	//dokuRemitReq.Sender.Address = ""
	dokuRemitReq.Sender.BirthDate = companyBirday
	dokuRemitReq.Sender.Country.Code = country
	dokuRemitReq.Sender.FirstName = companyName //company name
	dokuRemitReq.Sender.LastName = companyName  //company name
	//dokuRemitReq.Sender.Gender = ""
	dokuRemitReq.Sender.PersonalId = personalId
	dokuRemitReq.Sender.PersonalIdCountry.Code = country
	dokuRemitReq.Sender.Gender = "FEMALE"
	dokuRemitReq.Sender.Address = defaultCity
	//dokuRemitReq.Sender.PersonalIdExpireDate = ""
	//dokuRemitReq.Sender.PersonalIdIssueDate = ""
	dokuRemitReq.Sender.PersonalIdType = personalIdType
	dokuRemitReq.Sender.PhoneNumber = companyMobile
	//dokuRemitReq.SenderAmount = tools.Int642Str(amount)
	dokuRemitReq.SenderCountry.Code = country
	dokuRemitReq.SenderCurrency.Code = currency
	dokuRemitReq.SenderNote = tools.Int642Str(orderId) // TODO
	dokuRemitReq.Inquiry.IdToken = dokuDisburseInquiryResponse.Inquiry.IdToken

	uuidStr = uuid.Must(uuid.NewV4())
	requestId = uuidStr.String()
	signature = fmt.Sprintf("%s%s", agentKey, requestId)
	signature = tools.AesEncryptECB(signature, dokuEncryptionKsey)
	reqHeaders = map[string]string{
		"Content-Type": "application/json",
		//"requestId":    requestId,
		//"agentKey":     agentKey,
		//"signature":    signature,
	}

	dokuRemitReq.RequestId = requestId
	dokuRemitReq.AgentKey = agentKey
	dokuRemitReq.Signature = signature

	reqRemitJson, _ := json.Marshal(&dokuRemitReq)
	reqData = string(reqRemitJson)

	logs.Debug(remitUrl)
	logs.Debug(reqData)

	httpBody, httpStatusCode, err = tools.SimpleHttpClient("POST", remitUrl, reqHeaders, reqData, tools.DefaultHttpTimeout())
	invoke.HttpCode = httpStatusCode
	if err != nil {
		logs.Error("[DoKuRemitError] httpBody: %s, httpStatusCode: %d, err: %v\n", httpBody, httpStatusCode, err)
		//return
	}

	monitor.IncrThirdpartyCount(models.ThirdpartyDoKu, httpStatusCode)
	id, _ := models.AddOneThirdpartyRecord(models.ThirdpartyDoKu, remitUrl, orderId, reqData, string(httpBody), 0, 0, httpStatusCode)

	if httpStatusCode == 200 {
		status, respCode, remitResp := remitResp(httpBody)
		if status == 0 && respCode == "00" {
			//成功才计费
			thirdPartyData, _ := models.GetThirpartyRecordById(id)
			responstType, fee := thirdparty.CalcFeeByApi(remitUrl, reqData, string(httpBody))
			event.Trigger(&evtypes.CustomerStatisticEv{
				UserAccountId: accountId,
				OrderId:       orderId,
				ApiMd5:        tools.Md5(remitUrl),
				Fee:           int64(fee),
				Result:        responstType,
			})
			thirdPartyData.ResponseType = responstType
			thirdPartyData.FeeForCall = fee
			thirdPartyData.UpdateFee()
			invoke.FailureCode = ""
			invoke.DisbureStatus = types.DisbureStatusCallSuccess
		} else {
			err = fmt.Errorf("[DoKuDisburse Remit RespCode err], the body is: %s the remitResp is:%s", string(httpBody), respCode)
			logs.Error(err)
			invoke.FailureCode = remitResp
			invoke.DisbureStatus = types.DisbureStatusCallFailed
		}
	} else {
		if httpStatusCode == 0 {
			//超时的请求
			err = fmt.Errorf("[DoKuRemitTimeout] the entire reqStr is: ", reqData)
			logs.Error(err)
			invoke.FailureCode = ""
			invoke.DisbureStatus = types.DisbureStatusCallUnknow
		} else {
			//其他类型的错误
			err = fmt.Errorf("[DoKuRemit HttpStatusCode wrong] the entire reqStr is ", reqData)
			logs.Error(err)
			invoke.FailureCode = ""
			invoke.DisbureStatus = types.DisbureStatusCallFailed
		}
	}
	invoke.Utime = tools.GetUnixMillis()
	cols := []string{"disbure_status", "http_code", "utime", "failure_code"}
	models.OrmUpdate(&invoke, cols)

	res = httpBody
	return
}

func remitResp(httpBody []byte) (int, string, string) {
	status := -1
	var dokuRemitResp struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Remit   struct {
			TransactionId string `json:"transactionId"`
			PaymentData   struct {
				ResponseCode string `json:"responseCode"`
				ResponseMsg  string `json:"responseMsg"`
			} `json:"paymentData"`
		} `json:"remit"`
	}

	err := json.Unmarshal(httpBody, &dokuRemitResp)
	if err != nil {
		err = fmt.Errorf("remit response json.Unmarshal err, err is %s", err.Error())
		logs.Error(err)
	}
	status = dokuRemitResp.Status
	respCode := dokuRemitResp.Remit.PaymentData.ResponseCode
	responseMsg := dokuRemitResp.Remit.PaymentData.ResponseMsg
	return status, respCode, responseMsg
}

func (c *DokuApi) CreateVirtualAccountResponse(jsonData []byte, datas map[string]interface{}) (err error) {
	return
}

func (c *DokuApi) DisburseResponse(jsonData []byte, datas map[string]interface{}) (err error) {

	accountId := datas["account_id"].(int64)
	accountHolderName := datas["account_name"].(string)
	amount := datas["amount"].(int64)
	orderId := datas["order_id"].(int64)
	invokeId := datas["invoke_id"].(int64)

	bankInfo := datas["banks_info"].(models.BanksInfo)
	if len(bankInfo.DokuBrevityName) == 0 ||
		len(bankInfo.DokuFullName) == 0 ||
		len(bankInfo.DokuBrevityId) == 0 {
		err = fmt.Errorf("[DisburseResponse] doku unsport bank. accountId:%d datas:%#v", accountId, datas)
		return
	}
	bankCode := bankInfo.DokuBrevityName

	var dokuRemitResp struct {
		Status  int
		Message string
		Remit   struct {
			TransactionId string `json:"transactionId"`
		}
	}

	err = json.Unmarshal(jsonData, &dokuRemitResp)
	if err != nil {
		logs.Debug(err)
		return
	}

	if dokuRemitResp.Status != 0 {
		err = fmt.Errorf("[dokuRemitResponse] err: %s", string(jsonData))
		logs.Debug(err)

		// 更新 调用记录
		invoke, _ := models.OneDisburseInvorkLogByPkId(invokeId)
		invoke.FailureCode = dokuRemitResp.Message
		invoke.DisbureStatus = types.DisbureStatusCallBackFailed
		invoke.DisbursementId = dokuRemitResp.Remit.TransactionId
		invoke.Utime = tools.GetUnixMillis()
		cols := []string{"disbursement_id", "failure_code", "disbure_status", "utime"}
		models.OrmUpdate(&invoke, cols)

		return
	}

	o := models.Mobi_E_Trans{}

	o.UserAcccountId = accountId
	o.VaCompanyCode = types.DoKu
	o.Amount = amount
	//向上取整，百位取整
	o.PayType = types.PayTypeMoneyOut
	o.BankCode = bankCode
	o.AccountHolderName = accountHolderName
	o.DisbursementDescription = tools.Int642Str(orderId)
	o.DisbursementId = dokuRemitResp.Remit.TransactionId
	o.Status = "COMPLETED"
	o.CallbackJson = string(jsonData)
	o.Utime = tools.GetUnixMillis()
	o.Ctime = tools.GetUnixMillis()
	_, err = o.AddMobiEtrans(&o)

	// 更新 调用记录
	invoke, err := models.OneDisburseInvorkLogByPkId(invokeId)
	invoke.DisbursementId = dokuRemitResp.Remit.TransactionId
	invoke.DisbureStatus = types.DisbureStatusCallSuccess
	invoke.Utime = tools.GetUnixMillis()
	cols := []string{"disbursement_id", "disbure_status", "utime"}
	models.OrmUpdate(&invoke, cols)

	err = DoKuDisburseCallback(c, accountId, bankCode, datas)

	return
}

func DoKuDisburseCallback(c *DokuApi, accountId int64, bankCode string, datas map[string]interface{}) (err error) {
	order, err := dao.AccountLastLoanOrder(accountId)
	if err != nil {
		logs.Error("[DoKuDisburse] order nil err:%s, accountId:", err, accountId)
		return
	}

	if order.CheckStatus != types.LoanStatusIsDoing {
		//DoKu没回调, 放款后 应该在放款中
		err = fmt.Errorf("[DoKuDisburse] status error status:%d, orderid:%d ", int(order.CheckStatus), order.Id)

		//c.HandleLoanFailCallback(&order, err)

		return
	}

	err = c.HandleDisburseCallback(types.DoKu, &order, bankCode, false)

	if err != nil {
		logs.Error("[DoKuDisburse] status error err:%s, orderid:%d", err, order.Id)

		//c.HandleLoanFailCallback(&order, err)

		return
	}

	invokeId := datas["invoke_id"].(int64)
	invoke, err := models.OneDisburseInvorkLogByPkId(invokeId)
	invoke.DisbureStatus = types.DisbureStatusCallBackSuccess
	invoke.Utime = tools.GetUnixMillis()
	cols := []string{"disbursement_id", "disbure_status", "utime"}
	models.OrmUpdate(&invoke, cols)

	return
}

func CreateVirtualAccountCallback(jsonData []byte, accountId *int64) (err error) {
	return
}

func DisburseCallback(jsonData []byte, accountId *int64, bankCode *string, status *types.LoanStatus, isMatch *bool, tranData *models.Mobi_E_Trans) (err error) {
	return
}

func ReceivePaymentCallback(jsonData []byte, accountId *int64, amount *int64, bankCode *string) (err error) {
	return
}
