package accesstoken

import (
	"encoding/json"
	"fmt"

	"micro-loan/common/lib/redis/cache"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// 缓存相关的 token

func buildTokeCacheKey(platform, token string) (key string) {
	switch platform {
	case types.PlatformH5:
		key = fmt.Sprintf("%s%s", beego.AppConfig.String("webapi_token_prefix"), token)
	default:
		key = fmt.Sprintf("%s%s", beego.AppConfig.String("account_token_prefix"), token)
	}

	return
}

// 调用方只用关心是否有效,不用关心具体原因
func IsValidAccessToken(platform, token string) (bool, int64) {
	cKey := buildTokeCacheKey(platform, token)
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	//check, err := cacheClient.Do("EXISTS", cKey)
	//if err != nil || 1 != check.(int) {
	//	logs.Warning("access token key DOES NOT EXISTS, cKey:", cKey)
	//	return false
	//}

	cValue, err := cacheClient.Do("GET", cKey)
	//logs.Debug("cValue:", cValue, ", err:", err)
	if err != nil || cValue == nil {
		logs.Warning("access token key DOES NOT EXISTS, cKey:", cKey)
		return false, 0
	}

	var tokenInfo models.AccountToken
	err = json.Unmarshal(cValue.([]byte), &tokenInfo)
	if err != nil {
		// 说明有缓存数据,但内容有问题,消除之
		CleanTokenCache(platform, token)
		logs.Warning("json decode has wrong, please checkout. cKey:", cKey, ", cValue:", string(cValue.([]byte)))
		return false, 0
	}

	if tokenInfo.AccountId <= 0 || tokenInfo.Status != types.StatusValid || tokenInfo.Expires < tools.GetUnixMillis() {
		// 无效数据,消除之
		CleanTokenCache(platform, token)
		logs.Warning("cache data is invalid, please checkout. cKey:", cKey, ", cValue:", string(cValue.([]byte)))
		return false, 0
	}

	return true, tokenInfo.AccountId
}

func GenTokenWithCache(accountId int64, platform string, ip string, fcmToken string) (string, error) {
	token, err := models.GenerateAccountToken(accountId, platform, ip, fcmToken)
	if err != nil {
		logs.Error("can NOT create account_token. accountId:", accountId, ", platform:", platform, ", ip:", ip)
		return "", err
	}

	tokenInfo, err := models.GetAccessTokenInfo(token)
	if err != nil {
		logs.Error("can NOT find token info. token:", token)
		return "", err
	}

	cKey := buildTokeCacheKey(platform, token)
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	bson, err := json.Marshal(tokenInfo)
	expires := tokenInfo.Expires - tools.GetUnixMillis()
	if err != nil || expires < 0 {
		logs.Error("json encode model data has wrong. tokenInfo:", tokenInfo, ", expires:", expires)
		return "", err
	}

	// 安全忽略
	cacheClient.Do("SET", cKey, string(bson), "PX", expires)

	return token, nil
}

func GenTokenWithCacheV2(accountId int64, platform string, ip string, fcmToken string) (string, error) {

	// 在redis中删除，该账号的所有token
	num, accountTokens, _ := models.AccountValidToken(accountId)
	for i := int64(0); i < num; i++ {
		CleanTokenCache(platform, accountTokens[i].AccessToken)
	}
	// 将数据表account_token中，与该账号相关的token状态置位无效
	models.UpdateAccessTokenStatusByAccountId(accountId, types.StatusInvalid)

	// 生成新的token
	token, err := models.GenerateAccountToken(accountId, platform, ip, fcmToken)
	if err != nil {
		logs.Error("can NOT create account_token. accountId:", accountId, ", platform:", platform, ", ip:", ip)
		return "", err
	}

	tokenInfo, err := models.GetAccessTokenInfo(token)
	if err != nil {
		logs.Error("can NOT find token info. token:", token)
		return "", err
	}

	cKey := buildTokeCacheKey(platform, token)
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	bson, err := json.Marshal(tokenInfo)
	expires := tokenInfo.Expires - tools.GetUnixMillis()
	if err != nil || expires < 0 {
		logs.Error("json encode model data has wrong. tokenInfo:", tokenInfo, ", expires:", expires)
		return "", err
	}

	// 在redis中添加token信息
	cacheClient.Do("SET", cKey, string(bson), "PX", expires)

	return token, nil
}

// UpdateFcmToken 更新fcm_token
func UpdateFcmToken(token, fcmtoken string) {
	accountToken := models.AccountToken{}
	sql := fmt.Sprintf(`UPDATE %s SET fcm_token = '%s' WHERE access_token = '%s'`,
		accountToken.TableName(), fcmtoken, token)

	o := orm.NewOrm()
	o.Using(accountToken.Using())
	res, err := o.Raw(sql).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		logs.Debug("[UpdateFcmToken] update count: ", num, "token:", token, "fcm_token", fcmtoken)
	}
}

// 登出操作
func CleanTokenCache(platform, token string) {
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	cKey := buildTokeCacheKey(platform, token)
	// 理论不会出错,直接忽略返回
	cacheClient.Do("DEL", cKey)
}
