package voip

import (
	"encoding/json"
	"fmt"
	"micro-loan/common/lib/redis/cache"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/lib/sms/areacode"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"
)

var (
	reqUrl, appid, accessKey string
)

func getVoipReqUrl() {
	reqUrl = config.ValidItemString("voip_address")
}

func getVoipAccess() {
	getVoipReqUrl()
	appid = config.ValidItemString("voip_appid")
	accessKey = config.ValidItemString("voip_accesskey")
}

func getRet(body []byte) (ret int) {
	var voipRet VoipRet
	err := json.Unmarshal(body, &voipRet)
	if err != nil {
		logs.Error("[getRet] Get ret value, parse body failed")
		return
	}
	logs.Info("[getRet] Get ret value, voipRet:", voipRet)
	return voipRet.Ret
}

// 获取token
func GetVoipToken() (token string, err error) {

	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	cValue, err := cacheClient.Do("GET", VoipCacheTokenKey)
	if err != nil || cValue == nil {
		token, err = VoipLogin()
		return
	}

	token = string(cValue.([]byte))
	logs.Info("[GetVoipToken] Voip token value:", token)

	return
}

// 登录验证，并将token存入redis
func VoipLogin() (token string, err error) {
	var authLoginResp AuthLoginResponse

	getVoipAccess()

	reqBody := map[string]string{
		"service":   AuthLoginApi,
		"appid":     appid,
		"accesskey": accessKey,
	}

	var reqBox []string
	for k, v := range reqBody {
		reqBox = append(reqBox, fmt.Sprintf("%s=%s", k, v))
	}

	reqParma := strings.Join(reqBox, "&")
	logs.Info("[GetVoipToken] Voip login request url:", reqUrl, ", reqParma:", reqParma)
	reqHeaders := map[string]string{
		"User-Agent":   "curl/7.54.0",
		"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
	}

	body, _, err := tools.SimpleHttpClient("POST", reqUrl, reqHeaders, reqParma, tools.DefaultHttpTimeout())
	if err != nil {
		logs.Error("[GetVoipToken] Voip login request has wrong. reqUrl:", reqUrl, ", err:", err)
		return
	}

	ret := getRet(body)
	if ret != VoipRespRetSuccessed {
		err = fmt.Errorf("[GetVoipToken] Voip login request failed")
		return
	}
	err = json.Unmarshal(body, &authLoginResp)
	if err != nil {
		logs.Error("[GetVoipToken] Voip login request, parse body failed")
		return
	}

	authLoginRespStr, _ := tools.JsonEncode(authLoginResp)
	logs.Info("[GetVoipToken] Voip login response:", authLoginRespStr)

	token = authLoginResp.Data.Result.Token
	logs.Info("[GetVoipToken] Voip login token:", token)

	// 将token存入到redis
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	cacheClient.Do("SET", VoipCacheTokenKey, string(token), "PX", VoipTokenExpire)

	return
}

// 获取分机通话状态
// extnumber可以是：多个分机号用','连接的情况，不可以为空
func VoipSipCallStatus(extNumber string) (sipCallStatusResp SipCallStatusResponse, err error) {

	getVoipReqUrl()

	if len(extNumber) <= 0 {
		err = fmt.Errorf(ParamsError)
		logs.Error("[VoipSipCallStatus] has wrong, err:", err)
		return
	}

	for i := 0; i < 3; i++ {
		token, err1 := GetVoipToken()
		if err1 != nil {
			err = fmt.Errorf(GetTokenFail)
			logs.Error("[VoipSipCallStatus] Get voip token failed, err:", err1)
			return
		}

		reqBody := map[string]string{
			"service":   SipCallStatusApi,
			"token":     token,
			"extnumber": extNumber,
		}

		var reqBox []string
		for k, v := range reqBody {
			reqBox = append(reqBox, fmt.Sprintf("%s=%s", k, v))
		}

		reqParma := strings.Join(reqBox, "&")
		logs.Info("[VoipSipCallStatus] Voip sip call status request url:", reqUrl, ", reqParma:", reqParma)
		reqHeaders := map[string]string{
			"User-Agent":   "curl/7.54.0",
			"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
		}

		body, _, err1 := tools.SimpleHttpClient("POST", reqUrl, reqHeaders, reqParma, tools.DefaultHttpTimeout())
		if err1 != nil {
			err = fmt.Errorf(RequestFail)
			logs.Error("[VoipSipCallStatus] Voip sip call status request has wrong. reqUrl:", reqUrl, ", err:", err1)
			return
		}

		ret := getRet(body)
		if ret == Voip_Ret_600 {
			cacheClient := cache.RedisCacheClient.Get()
			defer cacheClient.Close()

			cacheClient.Do("DEL", VoipCacheTokenKey)
			continue
		} else if ret != VoipRespRetSuccessed {
			err = fmt.Errorf(GetVoipRetVal(ret))
			return
		}
		err = json.Unmarshal(body, &sipCallStatusResp)
		if err != nil {
			err = fmt.Errorf(GetSipStatusFail)
			logs.Error("[VoipSipCallStatus] Voip sip call status request, parse body failed")
			return
		}

		sipCallStatusRespStr, _ := tools.JsonEncode(sipCallStatusResp)
		logs.Info("[VoipSipCallStatus] Voip sip call status response:", sipCallStatusRespStr)
		break
	}

	return
}

// 获取分机信息
// extnumber可以是：多个分机号用','连接, 也可以为空(extNumber 为空时,查询所有分机信息)
func VoipSipNumberInfo(status SipNumberInfoStatus, extNumber string) (sipNumberInfoResp SipNumberInfoResponse, err error) {

	getVoipReqUrl()

	if status <= 0 {
		err = fmt.Errorf(ParamsError)
		logs.Error("[VoipSipNumberInfo] has wrong, err:", err)
		return
	}

	for i := 0; i < 3; i++ {
		token, err1 := GetVoipToken()
		if err1 != nil {
			err = fmt.Errorf(GetTokenFail)
			logs.Error("[VoipSipNumberInfo] Get voip token failed, err:", err1)
			return
		}

		reqBody := map[string]string{
			"service": SipNumberInfoApi,
			"token":   token,
			"status":  tools.Int2Str(int(status)),
		}
		if len(extNumber) > 0 {
			reqBody["extnumber"] = extNumber
		}

		var reqBox []string
		for k, v := range reqBody {
			reqBox = append(reqBox, fmt.Sprintf("%s=%s", k, v))
		}

		reqParma := strings.Join(reqBox, "&")
		logs.Info("[VoipSipNumberInfo] Voip sip number info request url:", reqUrl, ", reqParma:", reqParma)
		reqHeaders := map[string]string{
			"User-Agent":   "curl/7.54.0",
			"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
		}

		body, _, err1 := tools.SimpleHttpClient("POST", reqUrl, reqHeaders, reqParma, tools.DefaultHttpTimeout())
		if err1 != nil {
			err = fmt.Errorf(RequestFail)
			logs.Error("[VoipSipNumberInfo] Voip sip number info request has wrong. reqUrl:", reqUrl, ", err:", err1)
			return
		}

		ret := getRet(body)
		if ret == Voip_Ret_600 {
			cacheClient := cache.RedisCacheClient.Get()
			defer cacheClient.Close()

			cacheClient.Do("DEL", VoipCacheTokenKey)
			continue
		} else if ret != VoipRespRetSuccessed {
			err = fmt.Errorf(GetVoipRetVal(ret))
			return
		}
		err = json.Unmarshal(body, &sipNumberInfoResp)
		if err != nil {
			err = fmt.Errorf(GetSipNumberInfoFail)
			logs.Error("[VoipSipNumberInfo] Voip sip number info request, parse body failed")
			return
		}

		sipNumberInfoRespStr, _ := tools.JsonEncode(sipNumberInfoResp)
		logs.Info("[VoipSipNumberInfo] Voip sip number info response:", sipNumberInfoRespStr)
		break
	}

	return
}

func (r *MakeCallResponse) GetMakeCallStatus() int {
	return r.Data.Status
}

func isHitWhiteList(mobile string) (isHit bool) {

	/*
		// 规则匹配
		for _, v := range voipWhiteListPre {
			if strings.HasPrefix(mobile, v) {
				isHit = true
				return
			}
		}
	*/

	str := areacode.PhoneWithoutServiceRegionCode(mobile)
	for _, v := range voipWhiteListContain {
		if strings.Contains(str, v) {
			isHit = true
			return
		}
	}

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// 说明有错,或已经处理过,忽略本次操作
	isMember, err := redis.Bool(storageClient.Do("SISMEMBER", VoipWhiteListSetName, str))
	if err != nil {
		logs.Error("[isHitWhiteList] redis err:", err)
	}
	isHit = isMember

	return
}

// 去除空格，并且不能加拨0或62
func voipMobileFormat(mobile string) string {
	// 去除空格
	str := strings.Replace(mobile, " ", "", -1)

	if strings.HasPrefix(str, "08") {
		str = strings.Replace(str, "08", "8", 1)
	}

	if strings.HasPrefix(str, "628") {
		str = strings.Replace(str, "628", "8", 1)
	}

	return str
}

// 呼叫
func VoipMakeCall(makeCallReq MakeCallRequest) (makeCallResp MakeCallResponse, err error) {

	getVoipReqUrl()

	if len(makeCallReq.ExtNumber) <= 0 || len(makeCallReq.DestNumber) <= 0 {
		err = fmt.Errorf(ParamsError)
		logs.Error("[VoipMakeCall] has wrong, err:", err)
		return
	}

	// 检查外呼电话是否命中白名单
	isHit := isHitWhiteList(makeCallReq.DestNumber)
	if isHit {
		err = fmt.Errorf(HitVoipWhiteList)
		logs.Error("[VoipMakeCall] hit white list, makeCallReq.DestNumber: %s, err: %v", makeCallReq.DestNumber, err)
		return
	}

	for i := 0; i < 3; i++ {
		token, err1 := GetVoipToken()
		if err1 != nil {
			err = fmt.Errorf(GetTokenFail)
			logs.Error("[VoipMakeCall] Make call failed, err:", err1)
			return
		}

		mobile := voipMobileFormat(makeCallReq.DestNumber)
		reqBody := map[string]string{
			"service":    MakeCallApi,
			"token":      token,
			"extnumber":  makeCallReq.ExtNumber,
			"destnumber": mobile,
			"callmethod": VoipMakeCallMethodPositive,
			"doublecall": VoipMakeDoubleCallNo,
			"userid":     makeCallReq.UserID,
			"memberid":   makeCallReq.MemberID,
			"chengshudu": makeCallReq.Ripeness,
			"customuuid": makeCallReq.CustomUUID,
		}

		var reqBox []string
		for k, v := range reqBody {
			reqBox = append(reqBox, fmt.Sprintf("%s=%s", k, v))
		}

		reqParma := strings.Join(reqBox, "&")
		logs.Info("[VoipMakeCall] Make call request url:", reqUrl, ", reqParma:", reqParma)
		reqHeaders := map[string]string{
			"User-Agent":   "curl/7.54.0",
			"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
		}

		body, _, err1 := tools.SimpleHttpClient("POST", reqUrl, reqHeaders, reqParma, tools.DefaultHttpTimeout())
		if err1 != nil {
			err = fmt.Errorf(SendCallRequestFail)
			logs.Error("[VoipMakeCall] Make call request has wrong. reqUrl:", reqUrl, ", err:", err1)
			return
		}

		ret := getRet(body)
		if ret == Voip_Ret_600 {
			cacheClient := cache.RedisCacheClient.Get()
			defer cacheClient.Close()

			cacheClient.Do("DEL", VoipCacheTokenKey)
			continue
		} else if ret != VoipRespRetSuccessed {
			logs.Error("[VoipMakeCall] Make call request failed, err: ", GetVoipRetVal(ret))
			err = fmt.Errorf("%s", GetVoipRetVal(ret))
			return
		}
		err = json.Unmarshal(body, &makeCallResp)
		if err != nil {
			err = fmt.Errorf(MakeCallFail)
			logs.Error("[VoipMakeCall] Make call request, parse body failed")
			return
		}

		makeCallRespStr, _ := tools.JsonEncode(makeCallResp)
		logs.Info("[VoipMakeCall] Make call response:", makeCallRespStr)
		break
	}

	return
}

// 通话详单
func VoipCallList(callListReq CallListRequest) (callListResp CallListResponse, err error) {

	getVoipReqUrl()

	if len(callListReq.StartTime) <= 0 || len(callListReq.EndTime) <= 0 {
		err = fmt.Errorf(ParamsError)
		logs.Error("[VoipCallList] has wrong, err:", err)
		return
	}

	for i := 0; i < 3; i++ {
		token, err1 := GetVoipToken()
		if err1 != nil {
			err = fmt.Errorf(GetTokenFail)
			logs.Error("[VoipCallList] Get call list failed, err:", err1)
			return
		}

		reqBody := map[string]string{
			"service":   BillApi,
			"token":     token,
			"starttime": callListReq.StartTime,
			"endtime":   callListReq.EndTime,
			"syncflag":  tools.Int2Str(VoipSyncflagAll),
			"memberid":  callListReq.MemberID,
		}

		var reqBox []string
		for k, v := range reqBody {
			reqBox = append(reqBox, fmt.Sprintf("%s=%s", k, v))
		}

		reqParma := strings.Join(reqBox, "&")
		logs.Info("[VoipCallList] Get call list request url:", reqUrl, ", reqParma:", reqParma)
		reqHeaders := map[string]string{
			"User-Agent":   "curl/7.54.0",
			"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
		}

		body, _, err1 := tools.SimpleHttpClient("POST", reqUrl, reqHeaders, reqParma, tools.DefaultHttpTimeout())
		if err1 != nil {
			err = fmt.Errorf(RequestFail)
			logs.Error("[VoipCallList] Get call list request has wrong. reqUrl:", reqUrl, ", err:", err1)
			return
		}

		ret := getRet(body)
		if ret == Voip_Ret_600 {
			cacheClient := cache.RedisCacheClient.Get()
			defer cacheClient.Close()

			cacheClient.Do("DEL", VoipCacheTokenKey)
			continue
		} else if ret != VoipRespRetSuccessed {
			err = fmt.Errorf(GetVoipRetVal(ret))
			return
		}
		err = json.Unmarshal(body, &callListResp)
		if err != nil {
			err = fmt.Errorf(GetCallListFail)
			logs.Error("[VoipCallList] Get call list request, parse body failed")
			return
		}

		callListRespStr, _ := tools.JsonEncode(callListResp)
		logs.Info("[VoipCallList] Get call list response:", callListRespStr)
		break
	}

	return
}

func (r *RecodeFileResponse) GetRecordFileUrl() string {
	return r.Data.Result.DownURL
}

// 获取录音文件下载链接
func VoipRecordFileURL(fileName string) (recodeFileResp RecodeFileResponse, err error) {

	getVoipReqUrl()

	if len(fileName) <= 0 {
		err = fmt.Errorf(ParamsError)
		logs.Error("[VoipRecordFileURL] has wrong, err:", err)
		return
	}

	for i := 0; i < 3; i++ {
		token, err1 := GetVoipToken()
		if err1 != nil {
			err = fmt.Errorf(GetTokenFail)
			logs.Error("[VoipRecordFileURL] Get record file url failed, err:", err1)
			return
		}

		reqBody := map[string]string{
			"service":  RecodeFileApi,
			"token":    token,
			"filename": fileName,
		}

		var reqBox []string
		for k, v := range reqBody {
			reqBox = append(reqBox, fmt.Sprintf("%s=%s", k, v))
		}

		reqParma := strings.Join(reqBox, "&")
		logs.Info("[VoipRecordFileURL] Get record file url request url:", reqUrl, ", reqParma:", reqParma)
		reqHeaders := map[string]string{
			"User-Agent":   "curl/7.54.0",
			"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
		}

		body, _, err1 := tools.SimpleHttpClient("POST", reqUrl, reqHeaders, reqParma, tools.DefaultHttpTimeout())
		if err != nil {
			err = fmt.Errorf(RequestFail)
			logs.Error("[VoipRecordFileURL] Get record file url request has wrong. reqUrl:", reqUrl, ", err:", err1)
			return
		}

		ret := getRet(body)
		if ret == Voip_Ret_600 {
			cacheClient := cache.RedisCacheClient.Get()
			defer cacheClient.Close()

			cacheClient.Do("DEL", VoipCacheTokenKey)
			continue
		} else if ret != VoipRespRetSuccessed {
			err = fmt.Errorf(GetVoipRetVal(ret))
			return
		}
		err = json.Unmarshal(body, &recodeFileResp)
		if err != nil {
			err = fmt.Errorf(GetRecordFileURLFail)
			logs.Error("[VoipRecordFileURL] Get record file url request, parse body failed")
			return
		}

		recodeFileRespStr, _ := tools.JsonEncode(recodeFileResp)
		logs.Info("[VoipRecordFileURL] Get record file url response:", recodeFileRespStr)
		break
	}

	return
}
