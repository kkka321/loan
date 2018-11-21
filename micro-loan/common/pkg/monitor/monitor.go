package monitor

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/tools"
)

var orderHashPrefix string
var thirdpartyPrefix string

func init() {
	orderHashPrefix = beego.AppConfig.String("monitor_order")
	thirdpartyPrefix = beego.AppConfig.String("monitor_thirdparty")
}

func getMonitorKey(prefix string, date int64) string {
	return fmt.Sprintf("%s:%s", prefix, tools.GetDate(date/1000))
}

func IsKeyExist(prefix string, date int64) bool {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	key := getMonitorKey(prefix, date)

	num, _ := storageClient.Do("EXISTS", key)
	if num != nil && num.(int64) == 1 {
		return true
	} else {
		return false
	}
}

func DelKey(prefix string, date int64) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	key := getMonitorKey(prefix, date)

	storageClient.Do("DEL", key)
}

func incrCount(prefix string, date int64, field interface{}) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	key := getMonitorKey(prefix, date)

	storageClient.Do("HSETNX", key, field, 0)

	storageClient.Do("HINCRBY", key, field, 1)
}

func getCountFromCache(key string, field interface{}) int {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	hValue, err := storageClient.Do("HGET", key, field)
	if err != nil {
		logs.Warn("[getCountFromCache] HGET error key:%s, field:%v, err:%v", key, field, err)
		return 0
	}

	if hValue == nil {
		logs.Warn("[getCountFromCache] value is nil key:%s, field:%v", key, field)
		return 0
	}

	iValue, err := tools.Str2Int(string(hValue.([]byte)))
	if err != nil {
		logs.Warn("[getCountFromCache] unexcept value key:%s, field:%v, value:%s", key, field, string(hValue.([]byte)))
		return 0
	}

	return iValue
}
