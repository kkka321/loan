package thirdparty

import (
	"encoding/json"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

const (
	ApiCallResultSuccess = (1 << 0)
	ApiCallResultHit     = (1 << 1)
	split                = "+++"
	callCountSuffix      = "call_count"
	successCountSuffix   = "success_count"
	hitCountSuffix       = "hit_count"
)

func SetApiResultSuccess(result *int) int {
	*result = *result | ApiCallResultSuccess
	return *result
}

func ClearApiResultSuccess(result *int) int {
	*result = *result & (^ApiCallResultSuccess)
	return *result
}

func IsApiResultSuccess(result int) bool {
	return (result&ApiCallResultSuccess > 0)
}

func SetApiResultHit(result *int) int {
	*result = *result | ApiCallResultHit
	return *result
}

func ClearApiResultHit(result *int) int {
	*result = *result & (^ApiCallResultHit)
	return *result
}

func IsApiResultHit(result int) bool {
	return (result&ApiCallResultHit > 0)
}

type thirdPartyCalcInterface interface {
	// 可能一条调用对应多条费用 count对应条数
	CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error)
}

var CalcMap = map[string]thirdPartyCalcInterface{
	"https://api.253.com/open/i/witness/face-check": &API253{},

	"http://idtool.bluepay.asia//charge/express/npwpQuery": &Bluepay{},
	"http://120.76.101.146:21811/charge/express/npwpQuery": &Bluepay{},

	"http://intapi.253.com/send/json":      &Sms253{},
	"http://intapi.sgap.253.com/send/json": &Sms253{},

	"https://api-sgp.megvii.com/faceid/v1/detect": &Faceid{},
	"https://api-sgp.megvii.com/faceid/v2/verify": &Faceid{},

	"https://api.advance.ai/openapi/anti-fraud/v2/identity-check":        &Advance{},
	"https://api.advance.ai/openapi/anti-fraud/v3/identity-check":        &Advance{},
	"https://api.advance.ai/openapi/face-recognition/v2/check":           &Advance{},
	"https://api.advance.ai/openapi/face-recognition/v2/id-check":        &Advance{},
	"https://api.advance.ai/openapi/face-recognition/v2/ocr-check":       &Advance{},
	"https://api.advance.ai/openapi/default-detection/v3/multi-platform": &Advance{},
	"https://api.advance.ai/openapi/anti-fraud/v4/blacklist-check":       &Advance{},

	"https://api.xendit.co/callback_virtual_accounts": &Xendit{},
	"https://api.xendit.co/disbursements":             &Xendit{},
	"/xendit/virtual_account_callback/create":         &XenditVaCallback{},
	"/xendit/disburse_fund_callback/create":           &XenditVaCallback{},
	"/xendit/fva_receive_payment_callback/create":     &XenditPaymentCallback{},
	"/xendit/market_receive_payment_callback/create":  &XenditMarketPaymentCallback{},
	"/xendit/fix_payment_code_callback/create":        &XenditFixPaymentCodeCallback{},

	"/appsflyer/callback/install":                                                                         &AppsFlyer{},
	"https://api2.appsflyer.com/inappevent/com.loan.cash.credit.easy.kilat.cepat.pinjam.uang.dana.rupiah": &AppsFlyer{},
	"https://api2.appsflyer.com/inappevent/com.loan.cash.credit.pinjam.uang.dana.rapiah":                  &AppsFlyer{},
	"https://api2.appsflyer.com/inappevent/":                                                              &AppsFlyer{},

	"https://credit.akulaku.com/api/v2/credit_query": &Akulaku{},

	"https://rest.nexmo.com/sms/json": &Nexmo{},

	"/tongdun/callback/KTP":                                        &TongdunCallbackKTP{},
	"/tongdun/callback/YYS":                                        &TongdunCallbackYYS{},
	"/tongdun/callback/TRIP":                                       &TongdunCallbackYYS{},
	"https://talosapi.shujumohe.com/octopus/task.unify.create/v3":  &TongdunCreate{},
	"https://talosapi.shujumohe.com/octopus/task.unify.acquire/v3": &TongdunAcquire{},
	"https://talosapi.shujumohe.com/octopus/task.unify.query/v3":   &TongdunQuery{},
	"https://kirimdoku.com/v2/api/cashin/remit":                    &DoKuDisburseResponse{},
	"/doku/fva_receive_payment_callback/create":                    &DoKuDisburseResponse{},

	"https://gw.cmtelecom.com/v1.0/message": &Cmtelecom{},
}

func CalcFeeByApi(api string, request interface{}, reponse interface{}, apiCallTime ...int64) (result int, fee int) {
	logs.Info("[CalcFeeByApi] step into")
	defer logs.Info("[CalcFeeByApi] step out")
	logs.Debug("[CalcFeeByApi] api:", api, " request:%#v", request, " reponse:%#v", reponse)
	apis := strings.Split(api, "?")
	if v, ok := CalcMap[apis[0]]; ok {
		// 1、获得thirdparty
		thirdpartyInfo, err := models.GetThirdpartyInfoByApiMd5(tools.Md5(apis[0]))
		if nil != err {
			logs.Error("[CalcFeeByApi.Calc] models.GetThirdpartyInfoByApiMd5 err:%s  api:%s  request:%#v  reponse:%#v", err, api, request, reponse)
		}

		// 5、计算结果
		count := 1
		count, result, err = v.CalcFeeFunc(request, reponse, thirdpartyInfo)
		if nil != err {
			logs.Error("[CalcFeeByApi.Calc] err: %s  api:%s thirdpartyInfo:%#v  request:%#v reponse:%#v",
				err, api, thirdpartyInfo, request, reponse)
		}

		// 6、计算费率
		count = tools.ThreeElementExpression(count == 0, 1, count).(int)
		_, fee = getFeeByResult(result, count, &thirdpartyInfo)
		logs.Debug("[CalcFeeByApi.Calc] result:%d responseType:%d count:%d fee:%d ", result, result, count, fee)

		// 7、更新redis数据
		if len(apiCallTime) == 0 {
			UpdateThirdpartyStatisticFeeCache(apis[0], result)
		} else {
			UpdateThirdpartyStatisticFeeCacheForFixData(apis[0], result, apiCallTime...)
		}

	} else {
		logs.Warn("[CalcFeeByApi.Calc] do not get cal fun api:%s  request:%#v  reponse:%#v", api, request, reponse)
	}
	return
}

// 下面是第三方调用计算费率的各个实现。

func getResMap(reponse interface{}) map[string]interface{} {
	resMap := make(map[string]interface{})
	if _, ok := reponse.(map[string]interface{}); ok {
		resMap = reponse.(map[string]interface{})
	} else if str, ok := reponse.(string); ok {
		json.Unmarshal([]byte(str), &resMap)

		//这种情况可能是akulaku
		if len(resMap) == 0 {
			uStr := ""
			_ = json.Unmarshal([]byte(str), &uStr)
			_ = json.Unmarshal([]byte(uStr), &resMap)
		}
	} else {
		str, _ := json.Marshal(reponse)
		json.Unmarshal(str, &resMap)
		if len(resMap) == 0 {
			logs.Error("[getResMap] err decode:%#v", reponse)
		}
	}

	logs.Info("[getResMap] resMap:%#v", resMap)
	return resMap
}

func getFeeByResult(result int, count int, thirdpartyInfo *models.ThirdpartyInfo) (responseType int, fee int) {
	fee = 0
	switch thirdpartyInfo.ChargeType {
	case types.ChargeForCall:
		{
			fee = thirdpartyInfo.Price * count
		}
	case types.ChargeForCallSuccess:
		{
			if IsApiResultSuccess(result) {
				fee = thirdpartyInfo.Price * count
			}
		}
	case types.ChargeForHit:
		{
			if IsApiResultHit(result) {
				fee = thirdpartyInfo.Price * count
			}
		}
	case types.ChargeForFree:
		{
			fee = 0
		}
	}

	if IsApiResultHit(result) {
		responseType = types.CallReaultHit
	}

	if IsApiResultSuccess(result) {
		responseType = types.CallReaultSuccess
	}

	if !IsApiResultHit(result) && !IsApiResultSuccess(result) {
		responseType = types.CallReaultFailed
	}
	return
}

/******
advance
*****/
type Advance struct{}

// {"code":"SUCCESS","data":{"city":"KOTA BOGOR","district":"BOGOR BARAT","idNumber":"3271044808980023","name":"MUTIA GUSTIANI PUTRI","province":"JAWA BARAT","village":"PASIR JAYA"},"extra":null,"message":"OK"}

type AdvanceResStruct struct {
	Code string
}

func (r *Advance) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {

	resMap := getResMap(reponse)
	res := AdvanceResStruct{}
	str, err := json.Marshal(resMap)
	json.Unmarshal(str, &res)

	if types.CodeSuccess == res.Code {
		SetApiResultSuccess(&result)
		SetApiResultHit(&result)
	} else {
		logs.Info("[Advance.CalcFeeFunc] code not success:%#v ", reponse)
	}

	return
}

/******
Face++
*****/
type Faceid struct{}
type FaceidResStruct struct {
	Error        string
	ErrorMessage string
}

func (r *Faceid) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	resMap := getResMap(reponse)
	res := FaceidResStruct{}
	str, err := json.Marshal(resMap)
	json.Unmarshal(str, &res)

	SetApiResultSuccess(&result)
	SetApiResultHit(&result)

	if res.Error != "" {
		ClearApiResultSuccess(&result)
		ClearApiResultHit(&result)
	}

	if res.ErrorMessage != "" {
		ClearApiResultSuccess(&result)
		ClearApiResultHit(&result)
	}
	return
}

/******
AppsFlyer
*****/
type AppsFlyer struct{}
type AppsFlyerStruct struct {
	HTTPCode int64 `json:"HTTPCode"`
}

func (r *AppsFlyer) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	resMap := getResMap(reponse)
	res := AppsFlyerStruct{}
	str, err := json.Marshal(resMap)
	json.Unmarshal(str, &res)

	logs.Info("AppsFlyerStruct:%#v", res)

	SetApiResultSuccess(&result)
	SetApiResultHit(&result)

	if types.HTTPCodeSuccess != res.HTTPCode {
		ClearApiResultSuccess(&result)
		ClearApiResultHit(&result)
	}
	return
}

/******
TextLocal
*****/
type TextLocal struct{}

func (r *TextLocal) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	return
}

/******
API253
*****/
type API253 struct{}
type API253ResponseBody struct {
	Code string `json:"code"`
}

func (r *API253) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	resMap := getResMap(reponse)
	res := API253ResponseBody{}
	str, err := json.Marshal(resMap)
	json.Unmarshal(str, &res)

	logs.Debug("[API253 calcFeeFunc]resp:", res.Code)
	if res.Code == "200000" {
		SetApiResultSuccess(&result)
		SetApiResultHit(&result)
	}
	return
}

/******
Sms253
*****/
type Sms253 struct{}

// {"body":"{\"code\": \"0\", \"error\":\"\", \"msgid\":\"18061910431000435722\"}","httpCode":200,"httpErr":null}
type Sms253ResponseBody struct {
	Code  string `json:"code"`
	Error string `json:"error"`
	MsgID string `json:"msgid"`
}

func (r *Sms253) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	resMap := getResMap(reponse)

	if body, ok := resMap["body"]; ok {
		if str, ok := body.(string); ok {
			sBody := Sms253ResponseBody{}
			json.Unmarshal([]byte(str), &sBody)
			logs.Debug("[Sms253.CalcFeeFunc] sBody:", sBody, " reponse:", reponse)
			if "0" == sBody.Code {
				SetApiResultSuccess(&result)
				SetApiResultHit(&result)
			} else {
				logs.Warn("[Sms253.CalcFeeFunc] sBody:", sBody, " reponse:", reponse)
			}
		}
	}
	return
}

// Cmtelecom 计算收费结构体
type Cmtelecom struct{}

// CalcFeeFunc 计算收费方法
func (r *Cmtelecom) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	count = 1
	SetApiResultSuccess(&result)
	SetApiResultHit(&result)
	return
}

/******
Akulaku
*****/
type Akulaku struct{}
type AkulakuFlyerStruct struct {
	ErrMsg string
}

func (r *Akulaku) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	resMap := getResMap(reponse)
	res := AkulakuFlyerStruct{}
	str, err := json.Marshal(resMap)
	json.Unmarshal(str, &res)

	SetApiResultSuccess(&result)
	SetApiResultHit(&result)

	if res.ErrMsg != "" {
		ClearApiResultSuccess(&result)
		ClearApiResultHit(&result)
	}
	return
}

/******
Tongdun      https://talosapi.shujumohe.com/octopus/task.unify.create/v3
*****/
type TongdunCallbackKTP struct{}

type TongdunCallBackResStruct struct {
	Code int64
}

func (r *TongdunCallbackKTP) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	resMap := getResMap(reponse)
	res := TongdunCallBackResStruct{}
	str, err := json.Marshal(resMap)
	json.Unmarshal(str, &res)

	if res.Code == 0 {
		SetApiResultSuccess(&result)
		SetApiResultHit(&result)
	}

	return
}

type TongdunCallbackYYS struct{}

func (r *TongdunCallbackYYS) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	resMap := getResMap(reponse)
	res := TongdunCallBackResStruct{}
	str, err := json.Marshal(resMap)
	json.Unmarshal(str, &res)

	if res.Code == 0 {
		SetApiResultSuccess(&result)
		SetApiResultHit(&result)
	}

	return
}

/******
Tongdun      https://talosapi.shujumohe.com/octopus/task.unify.create/v3
*****/
type TongdunCreate struct{}

// identityCheckCreateTask 通用结构体
type tongdunInfo struct {
	Code    int64  `json:"code"`
	TaskID  string `json:"task_id"`
	Message string `json:"message"`
	Data    struct {
		ChannelCode  string `json:"channel_code"`
		ChannelType  string `json:"channel_type"`
		ChannelSrc   string `json:"channel_src"`
		ChannelAttr  string `json:"channel_attr"`
		CreateTime   string `json:"created_time"`
		IdentityCode string `json:"identity_code"`
		RealName     string `json:"real_name"`
		Mobile       string `json:"user_mobile"`
		TaskData     interface{}
	} `json:"data"`
}

func (r *TongdunCreate) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	resMap := getResMap(reponse)
	// code := getIntValue(resMap, "code", -111)
	res := TongdunCallBackResStruct{}
	str, err := json.Marshal(resMap)
	json.Unmarshal(str, &res)

	if res.Code == 0 {
		SetApiResultSuccess(&result)
		SetApiResultHit(&result)
	}

	return
}

/******
Tongdun      https://talosapi.shujumohe.com/octopus/task.unify.acquire/v3
*****/
type TongdunAcquire struct{}

func (r *TongdunAcquire) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	resMap := getResMap(reponse)
	res := TongdunCallBackResStruct{}
	str, err := json.Marshal(resMap)
	json.Unmarshal(str, &res)

	if res.Code == 0 {
		SetApiResultSuccess(&result)
		SetApiResultHit(&result)
	}
	return
}

/******
Tongdun      https://talosapi.shujumohe.com/octopus/task.unify.query/v3
*****/
type TongdunQuery struct{}

func (r *TongdunQuery) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {

	resMap := getResMap(reponse)
	res := TongdunCallBackResStruct{}
	str, err := json.Marshal(resMap)
	json.Unmarshal(str, &res)

	if res.Code == 0 || res.Code == 137 || res.Code == 105 || res.Code == 2007 {
		SetApiResultSuccess(&result)
		SetApiResultHit(&result)
	}
	return
}

/******
Boomsms
*****/
type Boomsms struct{}

func (r *Boomsms) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	return
}

/******
Xendit
*****/
type Xendit struct{}

type XenditStruct struct {
	ErrorCode string
	Status    string
}

func (r *Xendit) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	resMap := getResMap(reponse)
	res := XenditStruct{}
	str, err := json.Marshal(resMap)
	json.Unmarshal(str, &res)

	SetApiResultSuccess(&result)
	SetApiResultHit(&result)

	if res.ErrorCode != "" {
		ClearApiResultSuccess(&result)
		ClearApiResultHit(&result)
	}
	return
}

/******
Xendit    "/xendit/virtual_account_callback/create"
*****/
type XenditVaCallback struct{}

func (r *XenditVaCallback) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	resMap := getResMap(request)
	res := XenditStruct{}
	str, err := json.Marshal(resMap)
	json.Unmarshal(str, &res)

	if res.Status == "ACTIVE" || res.Status == "COMPLETED" {
		SetApiResultSuccess(&result)
		SetApiResultHit(&result)
	}
	return
}

type DoKuDisburseResponse struct{}

func (r *DoKuDisburseResponse) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {

	SetApiResultSuccess(&result)
	SetApiResultHit(&result)

	return
}

type DoKuPaymentCallbak struct{}

func (r *DoKuPaymentCallbak) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {

	SetApiResultSuccess(&result)
	SetApiResultHit(&result)

	return
}

type XenditPaymentCallback struct{}

func (r *XenditPaymentCallback) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {

	SetApiResultSuccess(&result)
	SetApiResultHit(&result)

	return
}

type XenditMarketPaymentCallback struct{}

func (r *XenditMarketPaymentCallback) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {

	SetApiResultSuccess(&result)
	SetApiResultHit(&result)

	return
}

type XenditFixPaymentCodeCallback struct{}

func (r *XenditFixPaymentCodeCallback) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {

	SetApiResultSuccess(&result)
	SetApiResultHit(&result)

	return
}

/******
Bluepay
*****/
type Bluepay struct{}

func (r *Bluepay) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {

	SetApiResultSuccess(&result)
	SetApiResultHit(&result)
	return
}

/******
Nexmo
*****/
type Nexmo struct{}

type nexmoResponseBody struct {
	MessageCount string    `json:"message-count"`
	Messages     []message `json:"messages"`
}

type message struct {
	To           string `json:"to"`
	MessageId    string `json:"message-id"`
	Status       string `json:"status"`
	MessagePrice string `json:"message-price"`
}

func (r *Nexmo) CalcFeeFunc(request interface{}, reponse interface{}, thirdpartyInfo models.ThirdpartyInfo) (count int, result int, err error) {
	resMap := getResMap(reponse)

	if body, ok := resMap["body"]; ok {
		if str, ok := body.(string); ok {
			sBody := nexmoResponseBody{}
			json.Unmarshal([]byte(str), &sBody)
			logs.Debug("[Nexmo.CalcFeeFunc] sBody:", sBody, " reponse:", reponse)
			count, _ = tools.Str2Int(sBody.MessageCount)
			if 0 < count && sBody.Messages[0].Status == "0" {
				SetApiResultSuccess(&result)
				SetApiResultHit(&result)
			} else {
				logs.Warn("[Nexmo.CalcFeeFunc] sBody:", sBody, " reponse:", reponse)
			}
		} else {
			logs.Error("body not string")
		}
	}
	return
}

//**********************************redis 相关**********************************/
func HashThirdpartyStatisticFeeKey() string {
	return beego.AppConfig.String("thirdpart_info_statistic_hash")
}

func UpdateThirdpartyStatisticFeeCache(api string, result int) {
	logs.Debug("[UpdateThirdpartyStatisticFeeCache] api:", api, "result:", result)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	hashName := HashThirdpartyStatisticFeeKey()
	hashName += (split + tools.MDateMHSDate(tools.GetUnixMillis()))
	logs.Debug("[UpdateThirdpartyStatisticFeeCache] hashName:", hashName)

	// 2、保存新数据
	_, err := storageClient.Do("HINCRBY", hashName, api+split+callCountSuffix, 1)
	if err != nil {
		logs.Error("[UpdateThirdpartyStatisticFeeCache] HINCRBY callCount err:", err, " api:", api, " hashName:", hashName)
	}

	if IsApiResultSuccess(result) {
		_, err = storageClient.Do("HINCRBY", hashName, api+split+successCountSuffix, 1)
		if err != nil {
			logs.Error("[UpdateThirdpartyStatisticFeeCache] HINCRBY successCount err:", err, " api:", api, " hashName:", hashName)
		}
	}
	if IsApiResultHit(result) {
		_, err = storageClient.Do("HINCRBY", hashName, api+split+hitCountSuffix, 1)
		if err != nil {
			logs.Error("[UpdateThirdpartyStatisticFeeCache] HINCRBY hitCount err:", err, " api:", api, " hashName:", hashName)
		}
	}

	return
}

// UpdateThirdpartyStatisticFeeCacheForFixData apiCallTime参数是为了修复数据而是用，实际应用中请勿填写具体值
func UpdateThirdpartyStatisticFeeCacheForFixData(api string, result int, apiCallTime ...int64) {
	logs.Debug("[UpdateThirdpartyStatisticFeeCacheForFixData] api:", api, "result:", result)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// 真实api不传调用时间  ，数据修复时传入调用时间
	if len(apiCallTime) == 0 {
		logs.Error("[UpdateThirdpartyStatisticFeeCacheForFixData] len(time) == 0")
		return
	}
	time := apiCallTime[0]

	hashName := HashThirdpartyStatisticFeeKey()
	hashName += (split + tools.MDateMHSDate(time))
	logs.Debug("[UpdateThirdpartyStatisticFeeCacheForFixData] hashName:", hashName)

	// 2、保存新数据
	_, err := storageClient.Do("HINCRBY", hashName, api+split+callCountSuffix, 1)
	if err != nil {
		logs.Error("[UpdateThirdpartyStatisticFeeCacheForFixData] HINCRBY callCount err:", err, " api:", api, " hashName:", hashName)
	}

	if IsApiResultSuccess(result) {
		_, err = storageClient.Do("HINCRBY", hashName, api+split+successCountSuffix, 1)
		if err != nil {
			logs.Error("[UpdateThirdpartyStatisticFeeCacheForFixData] HINCRBY successCount err:", err, " api:", api, " hashName:", hashName)
		}
	}
	if IsApiResultHit(result) {
		_, err = storageClient.Do("HINCRBY", hashName, api+split+hitCountSuffix, 1)
		if err != nil {
			logs.Error("[UpdateThirdpartyStatisticFeeCacheForFixData] HINCRBY hitCount err:", err, " api:", api, " hashName:", hashName)
		}
	}
	return
}

func MoveOutThirdpartyStatisticFeeFromCache() {
	logs.Info("MoveOutThirdpartyStatisticFeeFromCache")
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	hashName := HashThirdpartyStatisticFeeKey()
	hashNameCurrentDay := hashName + split + tools.MDateMHSDate(tools.GetUnixMillis())

	// +1 分布式锁
	lockKey := "lock:" + hashName
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", 30*60, "NX")
	if err != nil || lock == nil {
		logs.Warn("[MoveOutThirdpartyStatisticFeeFromCache] process is working, so, I reutrn.")
		if err != nil {
			logs.Error("[MoveOutThirdpartyStatisticFeeFromCache] may redis err:%v. lockKey:%s", err, lockKey)
		}
		return
	}
	defer storageClient.Do("DEL", lockKey)

	keys := getSringKeys(&storageClient, "KEYS", hashName+"*")
	logs.Info("[MoveOutThirdpartyStatisticFeeFromCache] keys:", keys)

	// 内存缓存一份 第三方数据的信息
	var thirdpartyInfoMap = make(map[string]models.ThirdpartyInfo)
	for _, key := range keys {
		// 将不是当天的 数据移除缓存

		//1、获取时间
		hashKeys := strings.Split(key, split)
		date := int64(0)
		if len(hashKeys) >= 2 {
			date = tools.GetDateParse(hashKeys[1]) * 1000
		} else {
			logs.Error("[MoveOutThirdpartyStatisticFeeFromCache] split hash key err: hashkey:", key)
			continue
		}

		// 获取所有的fileds
		fileds := getSringKeys(&storageClient, "HKEYS", key)
		for _, filed := range fileds {
			apis := strings.Split(filed, split)
			// 通过filed 反解出 api 和 类型
			if len(apis) >= 2 {
				api := apis[0]
				suffix := apis[1]

				//2、获取记录
				statisticFee, _ := models.GetThirdpartyStatisticFeeByApiAndDate(api, date)
				statisticFee.StatisticDate = date
				statisticFee.StatisticDateS = hashKeys[1]

				//3、填充结构体
				v, _ := tools.Str2Int(getHashFiledVal(&storageClient, key, filed))
				switch suffix {
				case callCountSuffix:
					{
						statisticFee.CallCount = v
					}
				case successCountSuffix:
					{
						statisticFee.SuccessCallCount = v
					}
				case hitCountSuffix:
					{
						statisticFee.HitCallCount = v
					}
				}

				//4、检查是否缓存了 第三方数据的信息
				if info, ok := thirdpartyInfoMap[api]; ok {
					statisticFee.Api = info.Api
					statisticFee.ApiMd5 = info.ApiMd5
					statisticFee.Name = info.Name
					statisticFee.Price = info.Price
					statisticFee.ChargeType = info.ChargeType
				} else {
					info, err := models.GetThirdpartyInfoByApiMd5(tools.Md5(api))
					if err != nil {
						logs.Error("[GetThirdpartyInfoByApiMd5] err:", err, " api:", api)
						continue
					}
					statisticFee.Api = info.Api
					statisticFee.ApiMd5 = info.ApiMd5
					statisticFee.Name = info.Name
					statisticFee.Price = info.Price
					statisticFee.ChargeType = info.ChargeType
					thirdpartyInfoMap[api] = info
				}

				//5、计算服务费
				switch statisticFee.ChargeType {
				case types.ChargeForCall:
					{
						statisticFee.TotalPrice = int64(statisticFee.CallCount) * int64(statisticFee.Price)
					}
				case types.ChargeForCallSuccess:
					{
						statisticFee.TotalPrice = int64(statisticFee.SuccessCallCount) * int64(statisticFee.Price)
					}
				case types.ChargeForHit:
					{
						statisticFee.TotalPrice = int64(statisticFee.HitCallCount) * int64(statisticFee.Price)
					}
				}

				//6、保存
				statisticFee.Ctime = tools.GetUnixMillis()
				if statisticFee.Id > 0 {
					_, err := statisticFee.Update()
					if err != nil {
						logs.Error("[MoveOutThirdpartyStatisticFeeFromCache] statisticFee.Update err:", err)
					}
				} else {
					_, err := statisticFee.Add()
					if err != nil {
						logs.Error("[MoveOutThirdpartyStatisticFeeFromCache] statisticFee.Add err:", err)
					}
				}

			} else {
				logs.Error("[MoveOutThirdpartyStatisticFeeFromCache] len(apis): ", len(apis), " apis:", apis, " filed:", filed, " key:", key)
				continue
			}
		}

		if key != hashNameCurrentDay {
			// 删除hash key
			storageClient.Do("DEL", key)
		}
	}
}

func getSringKeys(storageClient *redis.Conn, commond string, prifx string) (keys []string) {

	res, err := (*storageClient).Do(commond, prifx)
	if err != nil {
		logs.Error("[getSringKeys] commond: %s err:", commond, err, " prifx:", prifx)
	} else {
		keysB := res.([]interface{})
		for _, v := range keysB {
			if vBytes, ok := v.([]byte); ok {
				// logs.Debug("string(vBytes)--->%s", string(vBytes))
				keys = append(keys, string(vBytes))
			}
		}
	}
	return
}

func getHashFiledVal(storageClient *redis.Conn, hashKey string, filed string) (val string) {

	v, err := (*storageClient).Do("HGET", hashKey, filed)
	if err != nil {
		logs.Error("[getSringKeys] commond: %s err:", hashKey, err, " filed:", filed)
	} else {
		if vBytes, ok := v.([]byte); ok {
			val = string(vBytes)
		}
	}

	return
}
