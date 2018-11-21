// docs: https://dashboard.xendit.co/docs/introduction
// https://github.com/xendit

package xendit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/dao"
	"micro-loan/common/lib/device"
	"micro-loan/common/lib/payment"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/pkg/reduce"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type XenditApi struct {
	payment.PaymentApi
}

type XenditCreateVAccountResponse struct {
	OwnerId        string `json:"owner_id"`
	ExternalId     string `json:"external_id"`
	BankCode       string `json:"bank_code"`
	MerchantCode   string `json:"merchant_code"`
	Name           string `json:"name"`
	AccountNumber  string `json:"account_number"` //虚拟账户账号
	IsSingleUse    bool   `json:"is_single_use"`
	Status         string `json:"status"`
	ExpirationDate string `json:"expiration_date"`
	IsClosed       bool   `json:"is_closed"`
	Id             string `json:"id"` //FVA唯一id,可以用于创建invoice
	ErrorCode      string `json:"error_code"`
}

type XenditFVACallBackData struct {
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
	Created       string `json:"created"`
	Updated       string `json:"updated"`
	//ExpirationDate string `json:"expiration_date"`
	ExternalId   string `json:"external_id"`
	Id           string `json:"id"`
	IsClosed     bool   `json:"is_closed"`
	IsSingleUse  bool   `json:"is_single_use"`
	MerchantCode string `json:"merchant_code"`
	Name         string `json:"name"`
	OwnerId      string `json:"owner_id"`
	Status       string `json:"status"`
}

type XenditDisburseFundResponseData struct {
	UserId                  string `json:"user_id"`
	ExternalId              string `json:"external_id"`
	Amount                  int64  `json:"amount"`
	BankCode                string `json:"bank_code"`
	AccountHolderName       string `json:"account_holder_name"`
	DisbursementDescription string `json:"disbursement_description"`
	Status                  string `json:"status"`
	Id                      string `json:"id"`
	ErrorCode               string `json:"error_code"`
}

type XenditDisburseFundCallBackData struct {
	Id                      string `json:"id"`
	UserId                  string `json:"user_id"`
	ExternalId              string `json:"external_id"`
	Amount                  int64  `json:"Amount"`
	BankCode                string `json:"bank_code"`
	XenditFeeUserId         string `json:"xendit_fee_user_id"`
	XenditFeeAmount         string `json:"xendit_fee_amount"`
	AccountHolderName       string `json:"account_holder_name"`
	TransactionId           string `json:"transaction_id"`
	TransactionSequence     string `json:"transaction_sequence"`
	DisbursementDescription string `json:"disbursement_description"`
	FailureCode             string `json:"failure_code"`
	IsInstant               bool   `json:"is_instant"`
	Status                  string `json:"status"`
	Created                 string `json:"created"`
	Updated                 string `json:"updated"`
}

type XenditFVAReceivePaymentCallBackData struct {
	AccountNumber            string `json:"account_number"`
	BankCode                 string `json:"bank_code"`
	Amount                   int64  `json:"Amount"`
	ExternalId               string `json:"external_id"`
	Id                       string `json:"id"`
	CallbackVirtualAccountId string `json:"callback_virtual_account_id"`
	MerchantCode             string `json:"merchant_code"`
	PaymentId                string `json:"payment_id"`
	OwnerId                  string `json:"owner_id"`
	TransactionTimestamp     string `json:"transaction_timestamp"`
	Created                  string `json:"created"`
	Updated                  string `json:"updated"`
}

type XenditDisburseCorrectInquiryResp struct {
	UserId                  string `json:"user_id"`
	ExternalId              string `json:"external_id"`
	Amount                  int64  `json:"amount"`
	BankCode                string `json:"bank_code"`
	AccountHolderName       string `json:"account_holder_name"`
	DisbursementDescription string `json:"disbursement_description"`
	IsInstant               bool   `json:"is_instant"`
	Status                  string `json:"status"`
	Id                      string `json:"id"`
}

type XenditFixPaymentCode struct {
	OwnerId          string `json:"owner_id"`
	ExternalId       string `json:"external_id"`
	RetailOutletName string `json:"retail_outlet_name"`
	Prefix           string `json:"prefix"`
	name             string `json:"name"`
	PaymentCode      string `json:"payment_code"`
	Type             string `json:"type"`
	ExpectedAmount   int64  `json:"expected_amount"`
	IsSingleUse      bool   `json:"is_single_use"`
	ExpirationDate   string `json:"expiration_date"`
	Id               string `json:"id"`
	ErrorCode        string `json:"error_code"`
}

const PAYMENTCODELIMITAMOUNT = 10000

var bankNameCodeMap = map[string]string{
	"BPD Aceh":                                          "ACEH",
	"BPD Aceh UUS":                                      "ACEH_UUS",
	"Bank Agris":                                        "AGRIS",
	"Bank Agroniaga":                                    "AGRONIAGA",
	"Bank Andara":                                       "OKE",
	"Anglomas International Bank":                       "AMAR",
	"Bank Antar Daerah":                                 "CCB",
	"Bank ANZ Indonesia":                                "ANZ",
	"Bank Arta Niaga Kencana":                           "ARTA_NIAGA_KENCANA",
	"Bank Artha Graha International":                    "ARTHA",
	"Bank Artos Indonesia":                              "ARTOS",
	"BPD Bali":                                          "BALI",
	"Bank of America Merill-Lynch":                      "BAML",
	"Bangkok Bank":                                      "BANGKOK",
	"Bank Central Asia (BCA)":                           "BCA",
	"Bank Central Asia (BCA) Syariah":                   "BCA_SYR",
	"BPD Bengkulu":                                      "BENGKULU",
	"Bank Maybank":                                      "MAYBANK",
	"Bank Bisnis Internasional":                         "BISNIS_INTERNASIONAL",
	"Bank BJB":                                          "BJB",
	"Bank BJB Syariah":                                  "BJB_SYR",
	"Bank Negara Indonesia (BNI)":                       "BNI",
	"Bank BNI Syariah":                                  "BNI_SYR",
	"Bank BNP Paribas":                                  "BNP_PARIBAS",
	"Bank of China (BOC)":                               "BOC",
	"Bank Rakyat Indonesia (BRI)":                       "BRI",
	"Bank Syariah BRI":                                  "BRI_SYR",
	"Bank Tabungan Negara (BTN)":                        "BTN",
	"Bank Tabungan Negara (BTN) UUS":                    "BTN_UUS",
	"Bank Bukopin":                                      "BUKOPIN",
	"Bank Syariah Bukopin":                              "BUKOPIN_SYR",
	"Bank Bumi Arta":                                    "BUMI_ARTA",
	"Bank Capital Indonesia":                            "CAPITAL",
	"Centratama Nasional Bank":                          "CENTRATAMA",
	"Bank Chinatrust Indonesia":                         "CHINATRUST",
	"Bank CIMB Niaga":                                   "CIMB",
	"Bank CIMB Niaga UUS":                               "CIMB_UUS",
	"Citibank":                                          "CITIBANK",
	"Bank Commonwealth":                                 "COMMONWEALTH",
	"BPD Daerah Istimewa Yogyakarta (DIY)":              "DAERAH_ISTIMEWA",
	"BPD Daerah Istimewa Yogyakarta (DIY) UUS":          "DAERAH_ISTIMEWA_UUS",
	"Bank Danamon":                                      "DANAMON",
	"Bank Danamon UUS":                                  "DANAMON_UUS",
	"Bank DBS Indonesia":                                "DBS",
	"Deutsche Bank":                                     "DEUTSCHE",
	"Bank Dinar Indonesia":                              "DINAR_INDONESIA",
	"Bank DKI":                                          "DKI",
	"Bank DKI UUS":                                      "DKI_UUS",
	"Bank Ekonomi Raharja":                              "HSBC",
	"Bank Ekspor Indonesia":                             "EXIMBANK",
	"Bank Fama International":                           "FAMA",
	"Bank Ganesha":                                      "GANESHA",
	"Bank Hana":                                         "HANA",
	"Bank Harda Internasional":                          "HARDA_INTERNASIONAL",
	"Bank Himpunan Saudara 1906":                        "WOORI_SAUDARA",
	"Hongkong and Shanghai Bank Corporation (HSBC)":     "HSBC",
	"Hongkong and Shanghai Bank Corporation (HSBC) UUS": "HSBC_UUS",
	"Bank ICBC Indonesia":                               "ICBC",
	"Bank Ina Perdania":                                 "INA_PERDANA",
	"Bank Index Selindo":                                "INDEX_SELINDO",
	"Bank of India Indonesia":                           "INDIA",
	"BPD Jambi":                                         "JAMBI",
	"BPD Jambi UUS":                                     "JAMBI_UUS",
	"Bank Jasa Jakarta":                                 "JASA_JAKARTA",
	"BPD Jawa Tengah":                                   "JAWA_TENGAH",
	"BPD Jawa Tengah UUS":                               "JAWA_TENGAH_UUS",
	"BPD Jawa Timur":                                    "JAWA_TIMUR",
	"BPD Jawa Timur UUS":                                "JAWA_TIMUR_UUS",
	"JP Morgan Chase Bank":                              "JPMORGAN",
	"BPD Kalimantan Barat":                              "KALIMANTAN_BARAT",
	"BPD Kalimantan Barat UUS":                          "KALIMANTAN_BARAT_UUS",
	"BPD Kalimantan Selatan":                            "KALIMANTAN_SELATAN",
	"BPD Kalimantan Selatan UUS":                        "KALIMANTAN_SELATAN_UUS",
	"BPD Kalimantan Tengah":                             "KALIMANTAN_TENGAH",
	"BPD Kalimantan Timur":                              "KALIMANTAN_TIMUR",
	"BPD Kalimantan Timur UUS":                          "KALIMANTAN_TIMUR_UUS",
	"Bank Kesejahteraan Ekonomi":                        "KESEJAHTERAAN_EKONOMI",
	"BPD Lampung":                                       "LAMPUNG",
	"BPD Maluku":                                        "MALUKU",
	"Bank Mandiri":                                      "MANDIRI",
	"Bank Syariah Mandiri":                              "MANDIRI_SYR",
	"Bank Maspion Indonesia":                            "MASPION",
	"Bank Mayapada International":                       "MAYAPADA",
	"Bank Maybank Syariah Indonesia":                    "MAYBANK_SYR",
	"Bank Mayora":                                       "MAYORA",
	"Bank Mega":                                         "MEGA",
	"Bank Syariah Mega":                                 "MEGA_SYR",
	"Bank Mestika Dharma":                               "MESTIKA_DHARMA",
	"Bank Metro Express":                                "SHINHAN",
	"Bank Mitra Niaga":                                  "MITRA_NIAGA",
	"Bank Sumitomo Mitsui Indonesia":                    "MITSUI",
	"Bank Mizuho Indonesia":                             "MIZUHO",
	"Bank MNC Internasional":                            "MNC_INTERNASIONAL",
	"Bank Muamalat Indonesia":                           "MUAMALAT",
	"Bank Multi Arta Sentosa":                           "MULTI_ARTA_SENTOSA",
	"Bank Mutiara":                                      "JTRUST",
	"Bank Nationalnobu":                                 "NATIONALNOBU",
	"BPD Nusa Tenggara Barat":                           "NUSA_TENGGARA_BARAT",
	"BPD Nusa Tenggara Barat UUS":                       "NUSA_TENGGARA_BARAT_UUS",
	"BPD Nusa Tenggara Timur":                           "NUSA_TENGGARA_TIMUR",
	"Bank Nusantara Parahyangan":                        "NUSANTARA_PARAHYANGAN",
	"Bank OCBC NISP":                                    "OCBC",
	"Bank OCBC NISP UUS":                                "OCBC_UUS",
	"Bank Panin":                                        "PANIN",
	"Bank Panin Syariah":                                "PANIN_SYR",
	"BPD Papua":                                         "PAPUA",
	"Bank Permata":                                      "PERMATA",
	"Bank Permata UUS":                                  "PERMATA_UUS",
	"Prima Master Bank":                                 "PRIMA_MASTER",
	"Bank Pundi Indonesia":                              "BANTEN",
	"Bank QNB Kesawan":                                  "QNB_INDONESIA",
	"Bank Rabobank International Indonesia":             "RABOBANK",
	"Royal Bank of Scotland (RBS)":                      "RBS",
	"Bank Resona Perdania":                              "RESONA",
	"BPD Riau Dan Kepri":                                "RIAU_DAN_KEPRI",
	"BPD Riau Dan Kepri UUS":                            "RIAU_DAN_KEPRI_UUS",
	"Bank Royal Indonesia":                              "ROYAL",
	"Bank Sahabat Purba Danarta":                        "SAHABAT_PURBA_DANARTA",
	"Bank Sahabat Sampoerna":                            "SAHABAT_SAMPOERNA",
	"Bank SBI Indonesia":                                "SBI_INDONESIA",
	"Bank Sinar Harapan Bali":                           "MANDIRI_TASPEN",
	"Sinarmas":                                          "SINARMAS",
	"Standard Charted Bank":                             "STANDARD_CHARTERED",
	"BPD Sulawesi Tengah":                               "SULAWESI",
	"BPD Sulawesi Tenggara":                             "SULAWESI_TENGGARA",
	"BPD Sulselbar":                                     "SULSELBAR",
	"BPD Sulselbar UUS":                                 "SULSELBAR_UUS",
	"BPD Sulut":                                         "SULUT",
	"BPD Sumatera Barat":                                "SUMATERA_BARAT",
	"BPD Sumatera Barat UUS":                            "SUMATERA_BARAT_UUS",
	"BPD Sumsel Dan Babel":                              "SUMSEL_DAN_BABEL",
	"BPD Sumsel Dan Babel UUS":                          "SUMSEL_DAN_BABEL_UUS",
	"BPD Sumut":                                         "SUMUT",
	"BPD Sumut UUS":                                     "SUMUT_UUS",
	"Bank Tabungan Pensiunan Nasional":                  "TABUNGAN_PENSIUNAN_NASIONAL",
	"Bank Tabungan Pensiunan Nasional UUS":              "TABUNGAN_PENSIUNAN_NASIONAL_UUS",
	"Bank of Tokyo Mitsubishi UFJ":                      "TOKYO",
	"Bank UOB Indonesia":                                "UOB",
	"Bank Victoria Internasional":                       "VICTORIA_INTERNASIONAL",
	"Bank Victoria Syariah":                             "VICTORIA_SYR",
	"Bank Windu Kentjana Int":                           "WINDU",
	"Bank Woori Indonesia":                              "WOORI",
	"Bank Yudha Bhakti":                                 "YUDHA_BHAKTI",
}

func BankNameCodeMap() map[string]string {
	return bankNameCodeMap
}

func AllBankListStr() (ret string) {
	banks := BankNameCodeMap()
	keys := make([]string, 0, len(banks))
	for k := range banks {
		keys = append(keys, k)
	}
	ret = strings.Join(keys, ",")
	return ret
}

func AllBankList() []string {
	banks := BankNameCodeMap()
	keys := make([]string, 0, len(banks))
	for k := range banks {
		keys = append(keys, k)
	}
	logs.Debug(keys[0])
	return keys
}

func BankName2Code(name string) (code string, err error) {
	bankNameCodeMap := BankNameCodeMap()
	if v, ok := bankNameCodeMap[name]; ok {
		code = v
		return
	}

	err = fmt.Errorf("bank code undefined")

	return
}

func BankName2DisbureCode(name string) (code string, err error) {
	one, err := models.OneBankInfoByFullName(name)
	if err != nil {
		logs.Error("[BankName2DisbureCode] OneBankInfoByFullName err:%v. Xendit unsport bank name:%s", err, name)
		return
	}
	code = one.XenditBrevityName
	return
}

func VaBankCode(info models.BanksInfo) (code string, err error) {
	conf := map[string]bool{
		"MANDIRI": true,
		"BRI":     true,
		"BNI":     true,
	}

	// 还款没有不支持的 情况

	//if len(info.XenditBrevityName) == 0 {
	//	err = fmt.Errorf("[BankName2VaCode]  XenditBrevityName err. info:%#v", info)
	//	logs.Error(err)
	//	return
	//}

	if !conf[info.XenditBrevityName] {
		code = "BNI"
	} else {
		code = info.XenditBrevityName
	}
	return
}

func (c *XenditApi) CreateVirtualAccount(datas map[string]interface{}) (res []byte, err error) {
	//bankName := datas["bank_name"].(string)
	name := datas["account_name"].(string)
	externalId := datas["account_id"].(int64)
	bankInfo := datas["banks_info"].(models.BanksInfo)

	//bankCode, err := BankName2VaCode(bankName)
	bankCode, err := VaBankCode(bankInfo)
	if err != nil {
		return []byte{}, err
	}

	secretKey := beego.AppConfig.String("secret_key")
	virtualAccounts := beego.AppConfig.String("xendit_create_virtual_accounts")

	paramStr := fmt.Sprintf("%s%d%s%s%s%s",
		"external_id=", externalId,
		"&bank_code=", bankCode,
		"&name=", name)

	reqHeader := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + tools.BasicAuth(secretKey, ""),
	}

	httpBody, httpCode, err := tools.SimpleHttpClient("POST", virtualAccounts, reqHeader, paramStr, tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[CreateVirtualAccount] SimpleHttpClient error url:%s, params:%s, err:%s", virtualAccounts, paramStr, err.Error())
	}

	monitor.IncrThirdpartyCount(models.ThirdpartyXendit, httpCode)

	responstType, fee := thirdparty.CalcFeeByApi(virtualAccounts, paramStr, string(httpBody))
	models.AddOneThirdpartyRecord(models.ThirdpartyXendit, virtualAccounts, externalId, paramStr, string(httpBody), responstType, fee, httpCode)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: externalId,
		OrderId:       0,
		ApiMd5:        tools.Md5(virtualAccounts),
		Fee:           int64(fee),
		Result:        responstType,
	})

	return httpBody, err
}

func (c *XenditApi) CheckVirtualAccount(datas map[string]interface{}) (res []byte, err error) {
	return []byte{}, err
}

func invokeStatusByHttpCode(httpCode int) int {
	switch httpCode {
	case 200:
		{
			return types.DisbureStatusCallSuccess
		}
	case 0:
		{
			return types.DisbureStatusCallUnknow
		}
	default:
		{
			return types.DisbureStatusCallFailed
		}

	}
}

func (c *XenditApi) Disburse(datas map[string]interface{}) (res []byte, err error) {
	orderId := datas["order_id"].(int64)
	accountId := datas["account_id"].(int64)
	invokeId := datas["invoke_id"].(int64)
	externId := fmt.Sprintf("%d,%d", accountId, invokeId)
	bankName := datas["bank_name"].(string)
	accountHolderName := datas["account_name"].(string)
	accountNumber := datas["account_num"].(string)
	desc := datas["desc"].(string)
	amount := datas["amount"].(int64)
	bankCode, err := BankName2DisbureCode(bankName)
	if err != nil {
		return []byte{}, err
	}

	secretKey := beego.AppConfig.String("secret_key")
	virtualAccounts := beego.AppConfig.String("xendit_disburse_fund")

	paramStr := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%d",
		"external_id=", externId,
		"&bank_code=", bankCode,
		"&account_holder_name=", accountHolderName,
		"&account_number=", accountNumber,
		"&description=", desc,
		"&amount=", amount,
	)

	reqHeader := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + tools.BasicAuth(secretKey, ""),
	}

	httpBody, httpCode, err := tools.SimpleHttpClient("POST", virtualAccounts, reqHeader, paramStr, tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[Disburse] SimpleHttpClient error url:%s, params:%s, err:%s", virtualAccounts, paramStr, err.Error())
	}

	// save invoke result
	invoke, _ := models.OneDisburseInvorkLogByPkId(invokeId)
	invoke.DisbureStatus = invokeStatusByHttpCode(httpCode)
	invoke.HttpCode = httpCode
	invoke.Utime = tools.GetUnixMillis()
	cols := []string{"disbure_status", "http_code", "utime"}
	models.OrmUpdate(&invoke, cols)

	monitor.IncrThirdpartyCount(models.ThirdpartyXendit, httpCode)

	responstType, fee := thirdparty.CalcFeeByApi(virtualAccounts, paramStr, string(httpBody))
	models.AddOneThirdpartyRecord(models.ThirdpartyXendit, virtualAccounts, orderId, paramStr, string(httpBody), responstType, fee, httpCode)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: accountId,
		OrderId:       orderId,
		ApiMd5:        tools.Md5(virtualAccounts),
		Fee:           int64(fee),
		Result:        responstType,
	})

	return httpBody, err
}

func (c *XenditApi) CreateVirtualAccountResponse(jsonData []byte, datas map[string]interface{}) (err error) {
	var resp XenditCreateVAccountResponse = XenditCreateVAccountResponse{}
	err = json.Unmarshal(jsonData, &resp)
	if err != nil {
		logs.Error("[CreateVirtualAccountResponse] json.Unmarshal err:%s, json:%s", err.Error(), string(jsonData))
		return err
	}

	if resp.ErrorCode != "" {
		err = fmt.Errorf("%s%s", "Xendit CreateVirtualAccountResponse err", string(jsonData))
		return
	}

	accountId := datas["account_id"].(int64)
	//bankName := datas["bank_name"].(string)
	accountHolderName := datas["account_name"].(string)
	bankInfo := datas["banks_info"].(models.BanksInfo)
	bankCode, _ := VaBankCode(bankInfo)

	//bankCode, _ := BankName2VaCode(bankName)

	respAccountId, _ := tools.Str2Int64(resp.ExternalId)
	if respAccountId != accountId || bankCode != resp.BankCode || accountHolderName != resp.Name {
		errStr := fmt.Sprintf("[CreateVirtualAccountResponse] data not matched [%d],[%d] [%s],[%s] [%s],[%s]", accountId, respAccountId, bankCode, resp.BankCode, accountHolderName, resp.Name)
		err := fmt.Errorf(errStr)
		return err
	}

	//_, err = models.GetEAccount(accountId, types.Xendit)
	_, err = models.GetLastestActiveEAccountByRepayBankAndVacompanyType(accountId, resp.BankCode, types.Xendit)
	if err != nil {
		//不存在则创建
		eAccount := models.User_E_Account{}
		eAccount.Id, _ = device.GenerateBizId(types.UserEAccountBiz)
		eAccount.UserAccountId = accountId
		eAccount.VaCompanyCode = types.Xendit
		eAccount.EAccountNumber = resp.AccountNumber
		eAccount.BankCode = resp.BankCode
		eAccount.RepayBankCode = resp.BankCode
		eAccount.Status = resp.Status
		eAccount.Ctime = tools.GetUnixMillis()
		eAccount.Utime = tools.GetUnixMillis()
		if resp.IsClosed {
			eAccount.IsClosed = 1
		} else {
			eAccount.IsClosed = 0
		}
		_, err = eAccount.AddEAccount(&eAccount)
	}

	return err
}

func (c *XenditApi) DisburseResponse(jsonData []byte, datas map[string]interface{}) (err error) {
	var resp XenditDisburseFundResponseData = XenditDisburseFundResponseData{}
	err = json.Unmarshal(jsonData, &resp)
	if err != nil {
		logs.Error("[DisburseResponse] json.Unmarshal, err:%s, json:%s", err.Error(), string(jsonData))
		return err
	}

	invokeId := datas["invoke_id"].(int64)
	invoke, _ := models.OneDisburseInvorkLogByPkId(invokeId)

	if resp.ErrorCode != "" {
		invoke.Utime = tools.GetUnixMillis()
		invoke.DisbureStatus = types.DisbureStatusCallFailed
		invoke.FailureCode = resp.ErrorCode
		cols := []string{"failure_code", "disbure_status", "utime"}
		models.OrmUpdate(&invoke, cols)
		err = fmt.Errorf("%s%s", "Xendit DisburseResponse err", string(jsonData))
		return
	}

	accountId := datas["account_id"].(int64)
	accountHolderName := datas["account_name"].(string)
	amount := datas["amount"].(int64)

	externId := fmt.Sprintf("%d,%d", accountId, invokeId)

	if resp.ExternalId != externId || resp.AccountHolderName != accountHolderName || resp.Amount != amount {
		//response数据如果和请求的不一致，直接报警
		errStr := fmt.Sprintf("[DisburseResponse] response error, [%s],[%d] [%s],[%s] [%d],[%d]",
			resp.ExternalId, accountId, resp.AccountHolderName, accountHolderName, resp.Amount, amount)
		logs.Error(errStr)
		err = fmt.Errorf(errStr)
		return err
	}

	//userAccountId, _ := tools.Str2Int64(resp.ExternalId)
	//dataOrder, err := dao.AccountLastLoanOrder(userAccountId)
	//if err != nil {
	//	return err
	//}

	o := models.Mobi_E_Trans{}

	o.UserAcccountId = accountId
	o.VaCompanyCode = types.Xendit
	o.Amount = amount
	//向上取整，百位取整
	o.PayType = types.PayTypeMoneyOut
	o.BankCode = resp.BankCode
	o.AccountHolderName = resp.AccountHolderName
	o.DisbursementDescription = resp.DisbursementDescription
	o.DisbursementId = resp.Id
	o.Status = resp.Status
	o.Utime = tools.GetUnixMillis()
	o.Ctime = tools.GetUnixMillis()
	_, err = o.AddMobiEtrans(&o)
	//添加mobi_e_trans记录

	// 更新 调用记录
	//invoke, err = models.OneDisburseInvorkLogByPkId(invokeId)
	invoke.DisbursementId = resp.Id
	invoke.DisbureStatus = types.DisbureStatusCallSuccess
	invoke.Utime = tools.GetUnixMillis()
	cols := []string{"disbursement_id", "disbure_status", "utime"}
	models.OrmUpdate(&invoke, cols)

	return err
}

/**
 * 获取电子账户列表
 */
func XenditGetVirtualBanks() string {

	secretKey := beego.AppConfig.String("secret_key")
	virtualBanks := beego.AppConfig.String("xendit_virtual_banks")

	client := &http.Client{}
	req, err := http.NewRequest("GET", virtualBanks, nil)
	if err != nil {
		logs.Error("Xendit get virtual banks http.NewRequest err: ", err)
		return ""
	}

	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Set("Cookie", "name=anny")
	req.SetBasicAuth(secretKey, "")
	resp, err := client.Do(req)

	if err != nil {
		logs.Error("Xendit get virtual banks client.Do err: ", err)
		return ""
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logs.Error("Xendit get virtual banks ioutil.ReadAll err: ", err)
		return ""
	}

	return string(body)
}

/**
 * 获取电子账户详情
 */
func XenditGetEAccount(id string) (XenditCreateVAccountResponse, error) {

	secretKey := beego.AppConfig.String("secret_key")
	virtualBanks := beego.AppConfig.String("xendit_e_account_info")

	virtualBanks = fmt.Sprintf("%s%s", virtualBanks, id)

	resJson := XenditCreateVAccountResponse{}
	client := &http.Client{}
	req, err := http.NewRequest("GET", virtualBanks, nil)
	if err != nil {
		logs.Error("Xendit get eAccount http.NewRequest err: ", err)
		return resJson, err
	}

	req.SetBasicAuth(secretKey, "")
	resp, err := client.Do(req)

	if err != nil {
		logs.Error("Xendit get eAccount info client.Do err: ", err)
		return resJson, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logs.Error("Xendit get eAccount info ioutil.ReadAll err: ", err)
		return resJson, err
	}

	err = json.Unmarshal(body, &resJson)
	if err != nil {
		logs.Error("Xendit get eAccount info json.Unmarshal err: ", err)
	}

	return resJson, err
}

func CreateVirtualAccountCallback(jsonData []byte, accountId *int64) error {
	resp := XenditFVACallBackData{}
	err := json.Unmarshal(jsonData, &resp)
	if err != nil {
		logs.Error("[CreateVirtualAccountCallback] Json Unmarshal err:%s data:%s", err, string(jsonData))
		return err
	}

	account_id, _ := tools.Str2Int64(resp.ExternalId)
	*accountId = account_id

	account, err := models.OneAccountBaseByPkId(account_id)
	if err != nil {
		errStr := fmt.Sprintf("[CreateVirtualAccountCallback] account not exist account_id:%d, err:%s", account_id, err.Error())
		logs.Error(errStr)
		return fmt.Errorf(errStr)
	}

	account_profile, err := dao.CustomerProfile(account_id)
	if err != nil {
		errStr := fmt.Sprintf("[CreateVirtualAccountCallback] account_profile not exist account_id:%d, err:%s", account_id, err.Error())
		logs.Error(errStr)
		return fmt.Errorf(errStr)
	}

	one, err := models.OneBankInfoByFullName(account_profile.BankName)
	if err != nil {
		logs.Error("[CreateVirtualAccountCallback] OneBankInfoByFullName err:%v. check bank name:%s userAccountId:%d", err, account_profile.BankName, account_profile.AccountId)
		return err
	}

	bankCode, err := VaBankCode(one)
	if err != nil {
		errStr := fmt.Sprintf("[CreateVirtualAccountCallback] BankName2VaCode err bankName:%s, err:%s", account_profile.BankName, err.Error())
		logs.Error(errStr)
		return fmt.Errorf(errStr)
	}

	userEAccount, _ := models.GetEAccount(account_id, types.Xendit)
	eAccountNumber := userEAccount.EAccountNumber
	//o, err := models.GetEAccount(account_id, types.Xendit)
	o, err := models.GetLastestActiveEAccountByRepayBankAndVacompanyType(account_id, resp.BankCode, types.Xendit)
	if resp.BankCode != bankCode || resp.Name != account.Realname || resp.AccountNumber != eAccountNumber {
		//errStr := fmt.Sprintf("Xendit Account Create Callback data not matched bancode [%s],[%s], realname [%s],[%s], e_account_number [%s],[%s]", data.BankCode, bankCode, data.Name, account.Realname, data.AccountNumber, eAccountNumber)
		//logs.Error(errStr)
		//return fmt.Errorf(errStr)
		//TODO 暂时关闭
	}

	if err != nil {
		eAccount := models.User_E_Account{}
		eAccount.Id, _ = device.GenerateBizId(types.UserEAccountBiz)
		eAccount.EAccountNumber = resp.AccountNumber
		eAccount.BankCode = resp.BankCode
		eAccount.RepayBankCode = resp.BankCode
		accountId, _ := tools.Str2Int64(resp.ExternalId)
		eAccount.UserAccountId = accountId
		eAccount.VaCompanyCode = types.Xendit
		eAccount.Status = resp.Status
		eAccount.CallbackJson = string(jsonData)
		eAccount.Ctime = tools.GetUnixMillis()
		eAccount.Utime = tools.GetUnixMillis()
		if resp.IsClosed {
			eAccount.IsClosed = 1
		} else {
			eAccount.IsClosed = 0
		}
		_, err := eAccount.AddEAccount(&eAccount)
		if err != nil {
			logs.Error("[CreateVirtualAccountCallback] AddEAccount error err:", err)
		}
	} else {
		o.Status = resp.Status
		o.CallbackJson = string(jsonData)
		_, err := o.UpdateEAccount(&o)
		if err != nil {
			logs.Error("[CreateVirtualAccountCallback] UpdateEAccount error err:", err)
		}
	}

	return err
}

func DisburseCallback(jsonData []byte, accountId *int64, bankCode *string, status *types.LoanStatus, isMatch *bool, callBackOrderId *int64, amount *int64) (models.Mobi_E_Trans, error) {
	resp := XenditDisburseFundCallBackData{}
	err := json.Unmarshal(jsonData, &resp)
	if err != nil {
		logs.Error("[DisburseCallback] callback Json Unmarshal err:", err)
		return models.Mobi_E_Trans{}, err
	}
	*amount = resp.Amount
	*callBackOrderId, _ = tools.Str2Int64(resp.DisbursementDescription)
	*isMatch = true
	*bankCode = resp.BankCode
	if resp.Status != "COMPLETED" {
		logs.Error("[DisburseCallback] status err, json:%s", string(jsonData))
		*status = types.LoanStatusLoanFail
	}

	ids := strings.Split(resp.ExternalId, ",")

	if len(ids) > 1 {
		invokeId, _ := tools.Str2Int64(ids[1])
		invoke, _ := models.OneDisburseInvorkLogByPkId(invokeId)
		if resp.Status == "COMPLETED" {
			invoke.DisbureStatus = types.DisbureStatusCallBackSuccess
		} else {
			invoke.DisbureStatus = types.DisbureStatusCallBackFailed
			invoke.FailureCode = resp.FailureCode
		}
		invoke.Utime = tools.GetUnixMillis()
		cols := []string{"disbure_status", "failure_code", "utime"}
		models.OrmUpdate(&invoke, cols)
	}

	tranData, err := models.GetMobiEtrans(resp.Id)
	if err != nil {
		//目前发现这种超时的情况，如果回调显示COMPLETED,我们直接向表中插入数据，避免查询再次超时
		if resp.Status == "COMPLETED" {
			accountId, _ := tools.Str2Int64(ids[0])

			mobiEtrans := &models.Mobi_E_Trans{}
			mobiEtrans.UserAcccountId = accountId
			mobiEtrans.VaCompanyCode = types.Xendit
			mobiEtrans.Amount = resp.Amount
			//向上取整，百位取整
			mobiEtrans.PayType = types.PayTypeMoneyOut
			mobiEtrans.BankCode = resp.BankCode
			mobiEtrans.AccountHolderName = resp.AccountHolderName
			mobiEtrans.DisbursementDescription = resp.DisbursementDescription
			mobiEtrans.DisbursementId = resp.Id
			mobiEtrans.Status = "COMPLETED"
			mobiEtrans.CallbackJson = string(jsonData)
			mobiEtrans.Utime = tools.GetUnixMillis()
			mobiEtrans.Ctime = tools.GetUnixMillis()
			_, err = mobiEtrans.AddMobiEtrans(mobiEtrans)

			*tranData = *mobiEtrans
		} else {
			errStr := fmt.Sprintf("[DisburseCallback] status err, json:%s", string(jsonData))
			logs.Error(errStr)
			*isMatch = false
			return *tranData, nil
		}
	}

	*accountId = tranData.UserAcccountId

	if tranData.DisbursementId != resp.Id ||
		tranData.AccountHolderName != resp.AccountHolderName ||
		tranData.BankCode != resp.BankCode {
		errStr := fmt.Sprintf("[DisburseCallback] data not matched [%s],[%s] [%s],[%s] [%s],[%s]", tranData.DisbursementId, resp.Id, tranData.AccountHolderName, resp.AccountHolderName, tranData.BankCode, resp.BankCode)
		logs.Error(errStr)
		*isMatch = false
		return *tranData, nil
	}

	tranData.Status = resp.Status
	tranData.CallbackJson = string(jsonData)

	return *tranData, nil
}

func ReceivePaymentCallback(jsonData []byte, accountId *int64, amount *int64, bankCode *string) (error, XenditFVAReceivePaymentCallBackData) {
	resp := XenditFVAReceivePaymentCallBackData{}
	err := json.Unmarshal(jsonData, &resp)
	if err != nil {
		logs.Error("[ReceivePaymentCallback] Json Unmarshal err:", err)
		return err, resp
	}

	*accountId, _ = tools.Str2Int64(resp.ExternalId)
	*amount = resp.Amount
	*bankCode = resp.BankCode

	return err, resp
}

func DisburseInquiry(accountId int64, orderId int64) (resp XenditDisburseCorrectInquiryResp, err error, httpCode int) {
	inquiryUrl := beego.AppConfig.String("xendit_disburse_inquiry")
	secretKey := beego.AppConfig.String("secret_key")
	inquiryUrl = fmt.Sprintf("%s%d", inquiryUrl, accountId)

	auth := tools.BasicAuth(secretKey, "")
	reqHeaders := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + auth,
	}

	//logs.Debug(inquiryUrl)
	httpBody, httpCode, err := tools.SimpleHttpClient("GET", inquiryUrl, reqHeaders, "", tools.DefaultHttpTimeout())
	logs.Debug(string(httpBody))
	if err != nil {
		logs.Error("[DoKuDisburseInquiryError] httpBody: %s, httpStatusCode: %d, err: %v\n", httpBody, httpCode, err)
		return
	}

	if httpCode != 200 {
		err = fmt.Errorf("[DoKuDisburseInquiryError] httpCode is wrong, httpCode is %d, httpBody is %s", httpCode, string(httpBody))
		logs.Error(err)
		return
	}

	//debug start
	//httpString := `[{"user_id":"5a743292ea1830b877710ed2","external_id":"180813018257657471","amount":600000,"bank_code":"BNI","account_holder_name":"RINI NURILMI","disbursement_description":"180814028391632797","is_instant":true,"status":"COMPLETED","id":"5b727531c8d0f31000b79c79"}]`
	//httpBody = []byte(httpString)
	//debug end
	var inquiryResp []XenditDisburseCorrectInquiryResp
	err = json.Unmarshal(httpBody, &inquiryResp)

	if err != nil {
		err = fmt.Errorf("[DoKuDisburseInquiryError] json.Unmarshal err, err is ", err.Error())
		logs.Warn(err)
		return
	}

	for i := 0; i < len(inquiryResp); i++ {
		disbursementDescription, _ := tools.Str2Int64(inquiryResp[i].DisbursementDescription)
		//logs.Debug(inquiryResp[i])
		if disbursementDescription == orderId && inquiryResp[i].Status == "COMPLETED" {
			//disbursementDescription是我方服务器在放款时候传入的orderId
			//查询的时候对方服务器会将此参数传回，方便我们查询
			//此处验证时，只要保证订单是一样的，并且其中一单是完成的，证明对方已经成功放款过
			resp = inquiryResp[i]
			break
		}
	}
	return

}

func MarketPaymentCodeGenerate(orderId int64, balance int64) (err error, fixPaymentCode models.FixPaymentCode, amount int64) {
	order, err := models.GetOrder(orderId)
	if err != nil {
		logs.Error("[MarketPaymentCodeGenerate Xendit Invoice Create Response] Order does not exist! Order id is:[%d], err is:[%s]", orderId, err.Error())
		return
	}
	fixPaymentCode, err = XenditAddMarketPayment(order, balance)
	if err != nil {
		return
	}
	/*
		if smsFlag {
			//并不是所有情况都要发短信，目前只有后台逾期页面生成超市还款码时需要发短信
			account, _ := dao.CustomerOne(fixPaymentCode.UserAccountId)
			expireTime := tools.MHSHMS(fixPaymentCode.ExpirationDate)
			msg := fmt.Sprintf("CustYth, silahkan bayar via Alfamart dgn nama merchant Xendit, kode %s. Nominal Rp%d berlaku sampai %s[RUPIAH CEPAT]", fixPaymentCode.PaymentCode, fixPaymentCode.ExpectedAmount, expireTime)
			amount = fixPaymentCode.ExpectedAmount
			sms.Send(types.ServiceMarketPaymentCode, account.Mobile, msg, orderId)
		}
	*/
	return
}

func UpdatePaymentCodeExpireDate(repayPlan models.RepayPlan) (ExpiryDate int64) {
	now := tools.GetUnixMillis()
	today := tools.GetLocalDateFormat(now, "2006-01-02")
	tomorrow := tools.GetLocalDateFormat(now+tools.MILLSSECONDADAY, "2006-01-02")

	startDate := today + " " + "00:00:00"
	endDate := today + " " + "07:00:00"
	tomorrowEnd := tomorrow + " " + "07:00:00"

	startUnix, _ := tools.GetTimeParseWithFormat(startDate, "2006-01-02 15:04:05")
	endUnix, _ := tools.GetTimeParseWithFormat(endDate, "2006-01-02 15:04:05")
	tomorrowUnix, _ := tools.GetTimeParseWithFormat(tomorrowEnd, "2006-01-02 15:04:05")

	startUnix = startUnix * 1000
	endUnix = endUnix * 1000
	tomorrowUnix = tomorrowUnix * 1000

	if now >= repayPlan.RepayDate {
		//如果当前时间大于应还日期
		if now >= startUnix && now < endUnix {
			//同时当前时间大于当天印尼时间0:00, 小于当天印尼时间7:00
			ExpiryDate = endUnix
		} else {
			ExpiryDate = tomorrowUnix
		}
	}
	return
}

func XenditHeaderAuth() (reqHeaders map[string]string) {
	secretKey := beego.AppConfig.String("secret_key")
	auth := tools.BasicAuth(secretKey, "")
	reqHeaders = map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + auth,
	}
	return reqHeaders
}

func XenditInvokeGenerateFixPaymentCode(fixPaymentCodeUrl string, reqHeaders map[string]string, paramStr string) (obj XenditFixPaymentCode, httpBody []byte, err error) {
	httpBody, httpCode, err := tools.SimpleHttpClient("POST", fixPaymentCodeUrl, reqHeaders, paramStr, tools.DefaultHttpTimeout())
	monitor.IncrThirdpartyCount(models.ThirdpartyXendit, httpCode)
	err = json.Unmarshal(httpBody, &obj)
	if err != nil {
		err = fmt.Errorf("[XenditInvokeGenerateFixPaymentCode] XenditFixPaymentCode json.Unmarshal err:%s, json:%s", err.Error(), string(httpBody))
	}
	if obj.ErrorCode != "" {
		err = fmt.Errorf(string(httpBody))
	}
	return
}

func XenditInvokeUpdateFixPaymentCode(fixPaymentCodeUrl string, reqHeaders map[string]string, paramStr string) (obj XenditFixPaymentCode, httpBody []byte, err error) {
	logs.Debug(fixPaymentCodeUrl)
	logs.Debug(reqHeaders)
	logs.Debug(paramStr)
	httpBody, httpCode, err := tools.SimpleHttpClient("PATCH", fixPaymentCodeUrl, reqHeaders, paramStr, tools.DefaultHttpTimeout())
	monitor.IncrThirdpartyCount(models.ThirdpartyXendit, httpCode)
	err = json.Unmarshal(httpBody, &obj)
	if err != nil {
		err = fmt.Errorf("[XenditInvokeUpdateFixPaymentCode] XenditFixPaymentCode json.Unmarshal err:%s, json:%s", err.Error(), string(httpBody))
	}
	if obj.ErrorCode != "" {
		err = fmt.Errorf(string(httpBody))
	}
	return
}

func AddFixPaymentCode(fixPaymentCodeResp XenditFixPaymentCode, order models.Order, httpBody string) (fixPaymentCode models.FixPaymentCode, err error) {
	fixPaymentCode.Id = fixPaymentCodeResp.Id
	fixPaymentCode.UserAccountId = order.UserAccountId
	fixPaymentCode.OrderId = order.Id
	fixPaymentCode.PaymentCode = fixPaymentCodeResp.PaymentCode
	fixPaymentCode.ExpirationDate = tools.RFC3339TimeTransfer(fixPaymentCodeResp.ExpirationDate)
	fixPaymentCode.ExpectedAmount = fixPaymentCodeResp.ExpectedAmount
	fixPaymentCode.ResponseJson = string(httpBody)
	_, err = models.AddFixPaymentCode(&fixPaymentCode)
	return
}

func AddFixPaymentCodeOrder(fixPaymentCodeResp XenditFixPaymentCode, order models.Order) (fixPaymentCodeOrder models.FixPaymentCodeOrder, err error) {
	fixPaymentCodeOrder.UserAccountId = order.UserAccountId
	fixPaymentCodeOrder.OrderId = order.Id
	fixPaymentCodeOrder.PaymentCode = fixPaymentCodeResp.PaymentCode
	fixPaymentCodeOrder.ExpectedAmount = fixPaymentCodeResp.ExpectedAmount
	_, err = models.AddFixPaymentCodeOrder(&fixPaymentCodeOrder)
	return
}

func XenditAddMarketPayment(order models.Order, balance int64) (fixPaymentCode models.FixPaymentCode, err error) {

	if balance < PAYMENTCODELIMITAMOUNT && balance != 0 {
		err = fmt.Errorf("[XenditAddMarketPayment] parameter balance is not valid. order is %#v balance is %d", order, balance)
		logs.Error(err)
		return
	}

	repayPlanObj, err := models.GetLastRepayPlanByOrderid(order.Id)
	if err != nil {
		err = fmt.Errorf("[XenditAddMarketPayment] repayPlanObj does not exist. repayPlanObj is %#v, order is %#v", repayPlanObj, order)
		logs.Error(err)
		return
	}

	if balance == 0 {
		//来自app或task的请求,需要重新计算应还金额
		balance, err = reduce.RepayLowestMoney4ClearReduce(order, repayPlanObj)
		if err != nil {
			logs.Error("[XenditAddMarketPayment] RepayLowestMoney4ClearReduce amount errs, Order id is:[%d], err is:[%s]", order.Id, err.Error())
			return
		}
	}

	fixPaymentCodeUrl := beego.AppConfig.String("xendit_fix_paymentcode")
	reqHeaders := XenditHeaderAuth()
	customer, _ := dao.CustomerOne(order.UserAccountId)
	paymentCode, err := models.OneFixPaymentCodeByUserAccountId(order.UserAccountId)

	if err != nil {
		//没有则生成付款码
		paramStr := fmt.Sprintf("%s%d%s%s%s%s%d", "external_id=", order.UserAccountId, "&retail_outlet_name=ALFAMART", "&name=", customer.Realname, "&expected_amount=", balance)
		fixPaymentCodeResp, httpBody, err1 := XenditInvokeGenerateFixPaymentCode(fixPaymentCodeUrl, reqHeaders, paramStr)
		if err1 != nil {
			logs.Error(err1)
			err = err1
			return
		}
		fixPaymentCode, err = AddFixPaymentCode(fixPaymentCodeResp, order, string(httpBody))
		_, err = AddFixPaymentCodeOrder(fixPaymentCodeResp, order)
		return
	}

	if paymentCode.ExpectedAmount < PAYMENTCODELIMITAMOUNT {
		//如果已经存在付款码，金额又小于10000，则是脏数据！！！
		err = fmt.Errorf("[XenditAddMarketPayment] paymentCode.amount is not valid. paymentCode is %#v", paymentCode)
		return
	}

	if paymentCode.ExpectedAmount != balance {
		//更新付款码金额，后台操作
		paramStr := fmt.Sprintf("%s%d", "expected_amount=", balance)
		fixPaymentCodeUrl = fmt.Sprintf("%s/%s", fixPaymentCodeUrl, paymentCode.Id)
		fixPaymentCodeResp, _, err2 := XenditInvokeUpdateFixPaymentCode(fixPaymentCodeUrl, reqHeaders, paramStr)
		if err2 != nil {
			logs.Error(err2)
			err = err2
			return
		}
		cols := []string{"order_id", "expected_amount", "utime"}
		paymentCode.ExpectedAmount = balance
		paymentCode.OrderId = order.Id
		paymentCode.Utime = tools.GetUnixMillis()
		_, err = models.UpdateFixPaymentCode(&paymentCode, cols)
		_, err = AddFixPaymentCodeOrder(fixPaymentCodeResp, order)
		return
	} else {
		//已经存在了，有效的同时还款金额又没有变化，就直接返回
		fixPaymentCode = paymentCode
		logs.Warning("[XenditAddMarketPayment] previous paymentcode is valid, you don't need to generate again.", fixPaymentCode)
		return
	}

	return
}

func SimulateDisburse(req string) (err error) {
	disburseCallbackUrl := beego.AppConfig.String("xendit_disburse_callback_url")
	//disburseCallbackUrl = "http://localhost:8700/xendit/disburse_fund_callback/create"
	reqHeaders := map[string]string{
		"Content-Type": "application/json",
	}
	httpBody, httpCode, err := tools.SimpleHttpClient("POST", disburseCallbackUrl, reqHeaders, req, tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[XenditSimulateDisburse] httpBody: %s, httpStatusCode: %d, err: %v\n", httpBody, httpCode, err)
		return
	}
	if httpCode != 200 {
		err = fmt.Errorf("httpCode is not correct.")
		logs.Error("[XenditSimulateDisburse] httpBody: %s, httpStatusCode: %d, err: %v\n", httpBody, httpCode, err)
		return
	}
	return

	/*
			curl --include \
		     --request POST \
		     --header "Content-Type: application/json" \
		     --data-binary "{
		    \"id\": \"xendit_id_180813027644158910\",
		    \"user_id\": \"xendit_user_id\",
		    \"external_id\": \"180813017644055873\",
		    \"amount\": 600000,
		    \"bank_code\": \"BNI\",
		    \"account_holder_name\": \"FATHUL WAHDI\",
		    \"disbursement_description\": \"180813027644158910\",
		    \"is_instant\": true,
		    \"status\": \"COMPLETED\",
		    \"updated\": \"2018-08-14T08:15:03.404Z\",
		    \"created\": \"2018-08-14T08:15:03.404Z\"
		}" \
		'https://api.rupiahcepatweb.com/xendit/disburse_fund_callback/create'
	*/
}
