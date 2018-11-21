package advance

// api: https://doc.advance.ai

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/cerror"
	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
)

type ResponseData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		//ApiFaceComparison ,ApiIDCheck
		Similarity float64
		// ApiOCR -2  ApiIdentityCheck -all
		IDNumber string `json:"idNumber"`
		Name     string `json:"name"`
		Province string `json:"province"`
		City     string `json:"city"`
		District string `json:"district"`
		Village  string `json:"village"`

		// ApiCompanyCheck
		LegalCompanyInfoList  []Company `json:"legalCompanyInfoList"`
		GoogleCompanyInfoList []Company `json:"googleCompanyInfoList"`

		// ApiMultiPlatform
		Records    []MultiRecords    `json:"records"`
		Statistics []MultiStatistics `json:"statistics"`

		//ApiBlacklistCheck
		Recommendation    string      `json:"recommendation"`
		DefaultListResult []BlackList `json:"defaultListResult"`
	}
	Extra interface{} `json:"extra,omitempty"`
}

type BlackList struct {
	EventTime   string `json:"eventTime"`
	HitReason   string `json:"hitReason"`
	ProductType string `json:"productType"`
	ReasonCode  string `json:"reasonCode"`
}

type MultiRecords struct {
	Type       string   `json:"type"`
	QueryCount int      `json:"queryCount"`
	QueryDates []string `json:"queryDates"`
}

type MultiStatistics struct {
	QueryCount int    `json:"queryCount"`
	TimePeriod string `json:"timePeriod"`
}

type Company struct {
	CompanyName string `json:"companyName"`
	Address     string `json:"address"`
}

const (
	ApiHost           string = "https://api.advance.ai"
	ApiFaceComparison string = "/openapi/face-recognition/v2/check"
	ApiIDCheck        string = "/openapi/face-recognition/v2/id-check"
	ApiOCR            string = "/openapi/face-recognition/v2/ocr-check"
	ApiIdentityCheck  string = "/openapi/anti-fraud/v3/identity-check" //"/openapi/anti-fraud/v2/identity-check (已升级)"
	ApiCompanyCheck   string = "/openapi/anti-fraud/v3/company-check"
	ApiMultiPlatform  string = "/openapi/default-detection/v3/multi-platform" // "/openapi/default-detection/v2/multi-platform(已弃用)"
	ApiBlacklistCheck string = "/openapi/anti-fraud/v4/blacklist-check"
)

func genBoundary() string {
	millis := tools.GetUnixMillis()
	boundary := fmt.Sprintf("%s%s.%08d", "----AD1238MJL7", tools.SubString(tools.Md5(tools.Int642Str(millis)), 0, 25), millis)
	return boundary
}

func genTimeNow() string {
	now := time.Now()
	local, err := time.LoadLocation("GMT")
	if err != nil {
		return ""
	}

	timeNow := now.In(local).Format(time.RFC1123)
	return timeNow
}

func genSign(str string, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func prepare(apiName string, param map[string]interface{}, file map[string]interface{}) (string, string, map[string]string, error) {
	var requestUrl string = ApiHost + apiName
	var requestPostBody string = ""
	var requestHeaders = map[string]string{}
	var contentType string = ""
	var eof string = "\r\n"

	if len(file) > 0 {
		boundary := genBoundary()
		contentType = "multipart/form-data; boundary=" + boundary

		for k, v := range param {
			requestPostBody += fmt.Sprintf("--%s%s", boundary, eof)
			requestPostBody += fmt.Sprintf("Content-Disposition: form-data; name=\"%s\"%s", k, eof)
			requestPostBody += fmt.Sprintf("%s%s%s", eof, v, eof)
		}

		for k, fn := range file {
			if _, err := os.Stat(fn.(string)); os.IsNotExist(err) {
				logs.Warning("file does not exist:", fn.(string))
				return requestUrl, requestPostBody, requestHeaders, err
			}

			baseName := path.Base(fn.(string))
			mimeType := "application/octet-stream"
			buf, err := ioutil.ReadFile(fn.(string))
			if err != nil {
				logs.Warning("Failed to read the contents of the file:", fn.(string))
				return requestUrl, requestPostBody, requestHeaders, err
			}
			requestPostBody += fmt.Sprintf("--%s%s", boundary, eof)
			requestPostBody += fmt.Sprintf("Content-Disposition: form-data; name=\"%s\"; filename=\"%s\"%s", k, baseName, eof)
			requestPostBody += fmt.Sprintf("Content-Type: %s%s", mimeType, eof)
			requestPostBody += fmt.Sprintf("%s%s%s", eof, string(buf), eof)
		}

		requestPostBody += fmt.Sprintf("--%s--", boundary)
	} else {
		contentType = "application/json"
		jsonByte, _ := json.Marshal(param)
		requestPostBody = string(jsonByte)
	}

	now := genTimeNow()
	//now := "Wed, 24 Jan 2018 10:06:11 GMT"
	requestHeaders["Content-Type"] = contentType
	requestHeaders["Date"] = now

	separator := "$"
	signStr := fmt.Sprintf("POST%s", separator)
	signStr += apiName + separator
	signStr += contentType + separator
	signStr += now + separator

	fmt.Printf("signStr: %s\n", signStr)

	accessKey := beego.AppConfig.String("advance_access_key")
	secretKey := beego.AppConfig.String("advance_secret_key")
	logs.Debug("accessKey:", accessKey, ", secretKey:", secretKey)
	authorization := fmt.Sprintf("%s:%s", accessKey, genSign(signStr, secretKey))

	requestHeaders["Authorization"] = authorization

	return requestUrl, requestPostBody, requestHeaders, nil
}

func Request(relatedId int64, apiName string, param map[string]interface{}, file map[string]interface{}) ([]byte, ResponseData, error) {
	var original []byte
	resData := ResponseData{}
	reqUrl, reqBody, reqHeaders, err := prepare(apiName, param, file)
	//logs.Debug("reqUrl:", reqUrl, ", reqBody:", reqBody, ", reqHeaders:", reqHeaders)
	logs.Debug("reqUrl:", reqUrl, ", reqHeaders:", reqHeaders)

	if err != nil {
		return original, resData, err
	}

	httpBody, code, err := tools.SimpleHttpClient("POST", reqUrl, reqHeaders, reqBody, tools.DefaultHttpTimeout())

	monitor.IncrThirdpartyCount(models.ThirdpartyAdvance, code)

	if err != nil {
		return original, resData, err
	}

	requestMap := map[string]interface{}{
		"query_string": param,
		"files":        file,
	}
	resMap := map[string]interface{}{}
	json.Unmarshal(httpBody, &resMap)

	responstType, fee := thirdparty.CalcFeeByApi(reqUrl, requestMap, resMap)
	models.AddOneThirdpartyRecord(models.ThirdpartyAdvance, reqUrl, relatedId, requestMap, resMap, responstType, fee, code)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: relatedId,
		OrderId:       0,
		ApiMd5:        tools.Md5(reqUrl),
		Fee:           int64(fee),
		Result:        responstType,
	})

	err = json.Unmarshal(httpBody, &resData)
	if err != nil {
		logs.Warning("API data has wrong:", httpBody)
		return original, resData, err
	}

	return httpBody, resData, nil
}

// IDHoldingPhotoCheck 比对手持照片，如果结果少于定义阈值则返回错误码
func IDHoldingPhotoCheck(accountID int64, handHeldIDPhotoTmp, IDphoneTmp string) (code cerror.ErrCode, err error) {

	//手持比对开关
	compareSwitch, _ := config.ValidItemBool("firstloan_idhand_switch")

	logs.Debug("[IDHoldingPhotoCheck] compareSwitch: %v", compareSwitch)
	if compareSwitch == true {
		configSimilary, _ := config.ValidItemFloat64("first_idhand_idcard_similar")
		profile, _ := dao.CustomerProfile(accountID)
		fileHC := map[string]interface{}{
			"idHoldingImage": handHeldIDPhotoTmp,
		}
		_, faceHoldData, _ := Request(accountID, ApiIDCheck, map[string]interface{}{}, fileHC)
		if IsSuccess(faceHoldData.Code) {
			similarity := faceHoldData.Data.Similarity
			profile.SaveHoldCheck(similarity)
			logs.Debug("[ IDHoldingPhotoCheck ]手持比对结果（手持身份证中的头像与本人比对）：", similarity, " 本地阈值：", configSimilary)
			//如果小于阈值， 让我们再对比一波身份证与手持比对结果
			if similarity < configSimilary {
				similarity2, _ := FaceComparison(accountID, IDphoneTmp, handHeldIDPhotoTmp)
				profile.SaveHoldAndIDComparison(similarity2)
				logs.Debug("[ IDHoldingPhotoCheck ]二次检查身份证与手持比对结果：", similarity2, " 本地阈值：", configSimilary)
				if similarity2 < configSimilary {
					code = cerror.HandPhotoCheckLessThanDefine
				}
			} else {
				code = cerror.CodeSuccess
			}

		} else {
			similarity2, _ := FaceComparison(accountID, IDphoneTmp, handHeldIDPhotoTmp)
			profile.SaveHoldAndIDComparison(similarity2)
			logs.Debug("[ IDHoldingPhotoCheck ]二次检查身份证与手持比对结果：", similarity2, " 本地阈值：", configSimilary)
			if similarity2 < configSimilary {
				code = cerror.HandPhotoCheckLessThanDefine
			} else {
				code = cerror.CodeSuccess
			}
			logs.Warn("[IDHoldingPhotoCheck] advance 手持检查 ERROR CODE:", faceHoldData.Code, "Message:", faceHoldData.Message, "二次检查身份证与手持比对结果：", similarity2, " 本地阈值：", configSimilary)
		}
	} else {
		code = cerror.CodeSuccess
		logs.Debug("[IDHoldingPhotoCheck] 手持比对开关关闭 ，直接返回，不再调用第三方接口")
	}

	return
}

// FaceComparison 两张面部照片比较，返回0-100相识度值
func FaceComparison(accountID int64, photo1, photo2 string) (similarity float64, err error) {
	fileHC := map[string]interface{}{
		"firstImage":  photo1,
		"secondImage": photo2,
	}
	_, resultData, err := Request(accountID, ApiFaceComparison, map[string]interface{}{}, fileHC)

	if IsSuccess(resultData.Code) {
		similarity = resultData.Data.Similarity
	} else {
		logs.Warning("[FaceComparison] advance face comparison ERROR CODE:", resultData.Code, "Message:", resultData.Message)
	}
	return
}

// IdentiryCheck Advance身份检查
func IdentiryCheck(accountID int64, name, identity string) (resp ResponseData, err error) {
	param := map[string]interface{}{
		"name":     name,
		"idNumber": identity,
	}
	_, resp, err = Request(accountID, ApiIdentityCheck, param, map[string]interface{}{})
	return

}

func IsSuccess(code string) bool {
	return "SUCCESS" == code
}

func BlacklistCheck(accountId int64, name, identity, countryCode, mobile string) (body []byte, resp ResponseData, err error) {
	subParam := map[string]interface{}{
		"countryCode": countryCode,
		"areaCode":    "",
		"number":      mobile,
	}

	param := map[string]interface{}{
		"name":        name,
		"idNumber":    identity,
		"phoneNumber": subParam,
	}
	body, resp, err = Request(accountId, ApiBlacklistCheck, param, map[string]interface{}{})
	return
}

func MultiRecordsCheck(accountId int64, identity string) (body []byte, resp ResponseData, err error) {
	param := map[string]interface{}{
		"idNumber": identity,
	}
	body, resp, err = Request(accountId, ApiMultiPlatform, param, map[string]interface{}{})
	return
}

func BalcklistPass(advance *models.AccountAdvance) (string, map[string]bool, error) {
	res := ResponseData{}
	err := json.Unmarshal([]byte(advance.Response), &res)
	if err != nil {
		return "", map[string]bool{}, err
	}

	reasonCodeMap := map[string]bool{}

	for _, blackList := range res.Data.DefaultListResult {
		reasonCodeMap[blackList.ReasonCode] = true
	}

	return res.Data.Recommendation, reasonCodeMap, nil
}
