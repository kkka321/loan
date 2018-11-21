package api253

// api: https://api.253.com/#/api/interface/list

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
)

type ResponseData struct {
	ChargesStatus int64  `json:"chargesStatus"` //是否收费
	Code          string `json:"code"`          //响应code码。200000：成功，其他失败。请对照状态码
	Message       string `json:"message"`       //响应消息内容
	Data          struct {
		TradeNo     string `json:"tradeNo"`     //交易号，唯一
		CheckStatus string `json:"checkStatus"` //检测结果
		Remark      string `json:"remark"`      //检测说明
		Score       string `json:"score"`       //活体检测的分值，大于87分可判断为活体
		Code        string `json:"code"`        //活体检测返回码，0表示成功，其他为失败
	}
}

const (
	APIHost      string = "https://api.253.com"
	APIFaceCheck string = "/open/i/witness/face-check"
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
	var requestUrl string = APIHost + apiName
	var requestPostBody string = ""
	var requestHeaders = map[string]string{}
	var requestBodyBox []string
	if len(file) > 0 {

		for k, fn := range file {
			if _, err := os.Stat(fn.(string)); os.IsNotExist(err) {
				logs.Warning("file does not exist:", fn.(string))
				return requestUrl, requestPostBody, requestHeaders, err
			}
			buf, _ := ioutil.ReadFile(fn.(string))
			imageBase64 := tools.Base64Encode(buf)
			requestBodyBox = append(requestBodyBox, fmt.Sprintf("%s=%s", k, tools.UrlEncode(imageBase64)))
		}

	} else {
		jsonByte, _ := json.Marshal(param)
		requestPostBody = string(jsonByte)
	}

	requestBody := map[string]string{
		"appId":     "UEls146F",
		"appKey":    "cLwvstWI",
		"imageType": "BASE64",
	}
	requestHeaders["Content-Type"] = "application/x-www-form-urlencoded"
	requestHeaders["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36"
	for k, v := range requestBody {
		requestBodyBox = append(requestBodyBox, fmt.Sprintf("%s=%s", k, tools.UrlEncode(v)))
	}

	requestPostBody = strings.Join(requestBodyBox, "&")

	return requestUrl, requestPostBody, requestHeaders, nil
}

func Request(relatedId int64, apiName string, param map[string]interface{}, file map[string]interface{}) ([]byte, ResponseData, error) {
	var original []byte
	resData := ResponseData{}
	reqUrl, reqBody, reqHeaders, err := prepare(apiName, param, file)
	// logs.Debug("reqUrl:", reqUrl, ", reqBody:", reqBody, ", reqHeaders:", reqHeaders)
	logs.Debug("reqUrl:", reqUrl, ", reqHeaders:", reqHeaders)

	if err != nil {
		return original, resData, err
	}

	httpBody, code, err := tools.SimpleHttpClient("POST", reqUrl, reqHeaders, reqBody, tools.DefaultHttpTimeout())

	monitor.IncrThirdpartyCount(models.ThirdpartyAPI253, code)

	if err != nil {
		return original, resData, err
	}

	API253appId := beego.AppConfig.String("API253appId")
	API253appKey := beego.AppConfig.String("API253appKey")

	requestMap := map[string]interface{}{
		"appId":        API253appId,
		"appKey":       API253appKey,
		"query_string": param,
		"files":        file,
	}
	resMap := map[string]interface{}{}
	json.Unmarshal(httpBody, &resMap)

	responstType, fee := thirdparty.CalcFeeByApi(reqUrl, requestMap, resMap)
	models.AddOneThirdpartyRecord(models.ThirdpartyAPI253, reqUrl, relatedId, requestMap, resMap, responstType, fee, code)
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

// FaceCheck 活体检测
func FaceCheck(accountID int64, photo1 string) (score float64, err error) {
	fileHC := map[string]interface{}{
		"image": photo1,
	}
	_, resultData, err := Request(accountID, APIFaceCheck, map[string]interface{}{}, fileHC)
	if IsSuccess(resultData.Code) {
		score, _ = tools.Str2Float64(resultData.Data.Score)
	} else {
		logs.Warning("[FaceCheck] api253 facecheck ERROR CODE:", resultData.Code, "Message:", resultData.Message)
	}
	return
}

func IsSuccess(code string) bool {
	return "200000" == code
}
