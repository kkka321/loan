package faceid

// docs: https://faceid.com/pages/documents

import (
	"encoding/json"
	"fmt"

	"micro-loan/common/models"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"

	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"

	"github.com/astaxie/beego"
)

// 对image参数启用图片旋转检测功能
const (
	MultiOrientedDetectionYes string = "1"
	MultiOrientedDetectionNo  string = "0"

	ComparisonTypeDefault string = "0"
	FaceImageTypeDefault  string = "meglive"
)

// 接口定义
const (
	faceidHost string = "https://api-sgp.megvii.com"
	detectApi  string = "faceid/v1/detect"
	verifyApi  string = "faceid/v2/verify"
)

// 人脸识别接口响应结构体 {
type ResponseDetectFaces struct {
	Quality          float64 `json:"quality"`
	QualityThreshold float64 `json:"quality_threshold"`
}

type ResponseDetect struct {
	Faces []ResponseDetectFaces `json:"faces"`
}

// }

func getAppKey() string {
	return beego.AppConfig.String("faceid_app_key")
}

func getAppSecret() string {
	return beego.AppConfig.String("faceid_app_secret")
}

func getApiUrl(api string) (url string) {
	if api == "detect" {
		url = fmt.Sprintf("%s/%s", faceidHost, detectApi)
	} else if api == "verify" {
		url = fmt.Sprintf("%s/%s", faceidHost, verifyApi)
	}
	return
}

func Detect(relatedId int64, image string, detection string) (originRes []byte, httCode int, err error) {
	apiName := "detect"
	var reqUrl = getApiUrl(apiName)
	queryString := map[string]string{
		"api_key":                  getAppKey(),
		"api_secret":               getAppSecret(),
		"multi_oriented_detection": detection,
	}
	reqHeaders := map[string]string{}
	files := map[string]string{
		"image": image,
	}

	originRes, httCode, err = tools.MultipartClient(reqUrl, queryString, reqHeaders, files, tools.DefaultHttpTimeout())

	monitor.IncrThirdpartyCount(models.ThirdpartyFaceid, httCode)

	logger(reqUrl, relatedId, queryString, files, originRes, httCode)

	return
}

// Verify ...
// files: image_ref1, image_ref2, image_ref3, image_best, image_env
func Verify(relatedId int64, comparisonType string, faceImageType string, files map[string]string, delta string) (originRes []byte, httCode int, err error) {
	apiName := "verify"
	var reqUrl = getApiUrl(apiName)
	queryString := map[string]string{
		"api_key":         getAppKey(),
		"api_secret":      getAppSecret(),
		"comparison_type": comparisonType,
		"face_image_type": faceImageType,
		"uuid":            tools.GetGuid(),
		"delta":           delta,
	}
	reqHeaders := map[string]string{}

	originRes, httCode, err = tools.MultipartClient(reqUrl, queryString, reqHeaders, files, tools.DefaultHttpTimeout())

	logger(reqUrl, relatedId, queryString, files, originRes, httCode)

	return
}

// 记录下每次请求的参数和结果,后面自数据时,是一块高质量的数据集
func logger(reqUrl string, relatedId int64, queryString map[string]string, files map[string]string, originRes []byte, httpCode int) {
	// logger 4 api request.
	delete(queryString, "api_key")
	delete(queryString, "api_secret")
	delete(queryString, "delta") // 此数据为SDK生成,对原始数据而言,无参考价值,不必入库
	reqMap := map[string]interface{}{
		"query_string": queryString,
		"files":        files,
	}
	resMap := map[string]interface{}{}
	json.Unmarshal(originRes, &resMap)

	responstType, fee := thirdparty.CalcFeeByApi(reqUrl, reqMap, resMap)
	models.AddOneThirdpartyRecord(models.ThirdpartyFaceid, reqUrl, relatedId, reqMap, resMap, responstType, fee, httpCode)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: relatedId,
		OrderId:       0,
		ApiMd5:        tools.Md5(reqUrl),
		Fee:           int64(fee),
		Result:        responstType,
	})
}
