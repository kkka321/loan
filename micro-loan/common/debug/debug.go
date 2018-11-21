package main

import (
	"encoding/hex"
	"fmt"
	"time"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/device"
	"micro-loan/common/lib/redis/cache"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"micro-loan/common/models"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
)

func main() {
	logs.Debug("debug...")
	// redis cache 使用例子
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()

	cKey := "test"
	cacheClient.Do("SET", cKey, fmt.Sprintf("time:%d", tools.GetUnixMillis()), "EX", 30)
	cValue, err := cacheClient.Do("GET", cKey)
	fmt.Printf("just test cache redis\n")
	if err == nil {
		fmt.Printf("cKey: %s\n", cKey)
		fmt.Printf("%s: %s\n", cKey, cValue)
	} else {
		fmt.Printf("%v %t\n", err, err == nil)
	}

	check, err := cacheClient.Do("EXISTS", "nonexistent")
	fmt.Printf("EXISTS nonexistent: check: %v, err: %v\n", check, err)

	cValue, err = cacheClient.Do("GET", "nonexistent")
	fmt.Printf(">>> GET nonexistent -> value: %T, err: %v\n", cValue, err)

	// redis storage 例子
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	fmt.Printf("just test storage redis\n")
	qName := "message"
	storageClient.Do("lpush", qName, tools.GetUnixMillis())
	qValue, _ := storageClient.Do("rpop", qName)
	fmt.Printf("qName: %s, qValue: %s\n", qName, qValue)

	qVal, _ := redis.Values(storageClient.Do("SMEMBERS", "test_sets"))
	fmt.Printf("test_sets: %#v\n", qVal)
	for _, v := range qVal {
		fmt.Printf("    %s\n", v.([]byte))
	}

	bizId, _ := device.GenerateBizId(types.OrderSystem)
	fmt.Printf("bizId->order: %d\n", bizId)

	randomStr := tools.GenerateRandomStr(8)
	fmt.Printf("randomStr: %s\n", randomStr)

	fmt.Printf("env: %s, IsProductEnv: %v\n", tools.GetCurrentEnv(), tools.IsProductEnv())
	fmt.Printf("mobile captcha: %s\n", tools.GenerateMobileCaptcha(6))

	ip := "117.100.140.129"
	fmt.Printf("ip: %s, ISOCountryCode: %s, CityEn: %s, CityZhCN: %s\n", ip, tools.GeoipISOCountryCode(ip), tools.GeoipCityEn(ip), tools.GeoipCityZhCN(ip))
	fmt.Printf("ip Location: %v\n", tools.GeoipLocation(ip))

	pwd := "10ve@9o9MM"
	encryptPwd := tools.PasswordEncrypt(pwd, 1516710209872)
	fmt.Printf("pwd: %s, encryptPwd: %s\n", pwd, encryptPwd)

	fmt.Printf("now: %s\n", tools.GetDateFormat(time.Now().Unix(), "2018 08:58:54"))

	testUrl := "http://127.0.0.1/post.php"
	reqHeaders := map[string]string{
		"Connection":       "keep-alive",
		"Content-Type":     "application/x-www-form-urlencoded",
		"User-Agent":       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
		"X-Encrypt-Method": "AES",
	}

	httpBody, httpStatusCode, err := tools.SimpleHttpClient("POST", testUrl, reqHeaders, "model=debug", tools.DefaultHttpTimeout())
	fmt.Printf("httpBody: %s, httpStatusCode: %d, err: %v\n", httpBody, httpStatusCode, err)

	aesKey := tools.AesCBCKey
	key, _ := hex.DecodeString(aesKey)
	fmt.Printf("key: %v\n", key)
	aesIV := tools.AesCBCIV
	//iv, _ := hex.DecodeString(aesIV)
	plaintext := "hello"
	ciphertext, _ := tools.AesEncryptCBC(plaintext, aesKey, aesIV)
	fmt.Printf("plaintext: %s, ciphertext: %s\n", plaintext, ciphertext)
	//decryptText, _ := tools.AesDecryptCBC(ciphertext, aesKey, aesIV)
	decryptText, _ := tools.AesDecryptUrlCode(ciphertext, aesKey, aesIV)
	fmt.Printf("decryptText: %s\n", decryptText)

	params := map[string]interface{}{
		"noise":        "123456789",
		"access_token": "abc",
		"app_version":  "1.0.0.0",
		"platform":     "android",
	}
	signature := tools.Signature(params, tools.GetSignatureSecret())
	fmt.Printf("params: %v, signature: %s\n", params, signature)

	var list []models.AccountBase
	var ab models.AccountBase
	o := orm.NewOrm()
	o.Using(ab.Using())
	//sql := fmt.Sprintf(`SELECT * FROM %s WHERE id > 0 LIMIT 1`, ab.TableName())
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE id > 0 LIMIT 1; SELECT * FROM orders `, ab.TableName())
	num, err := o.Raw(sql).QueryRows(&list)
	fmt.Printf("num: %d, err: %#v, list: %#v\n", num, err, list)

	err = o.QueryTable(ab.TableName()).Filter("id", 0).One(&ab)
	fmt.Printf("ab: %#v, err: %s\n", ab, err.Error())
}
