package monitor

import (
	"math"
	"fmt"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/tools"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
)

func getThirdpartyField(thirdparty int, success int) string {
	return fmt.Sprintf("%d_%d", thirdparty, success)
}

func GetThirdpartyStatistics(date int64) (list []models.ThirdpartyStatistics) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	key := getMonitorKey(thirdpartyPrefix, date)

	setsMem, err := redis.Values(storageClient.Do("HKEYS", key))
	if err != nil || setsMem == nil {
		return
	}

	strDate := tools.GetDate(date / 1000)
	mapData := map[int]models.ThirdpartyStatistics{}

	for _, m := range setsMem {
		field := string(m.([]byte))

		strVec := strings.Split(field, "_")
		if len(strVec) != 2 {
			logs.Error("[GetThirdpartyStatistics] wrong field:%s", field)
			continue
		}

		count := getCountFromCache(key, field)

		thirdparty, _ := tools.Str2Int(strVec[0])
		isSuccess, _ := tools.Str2Int(strVec[1])
		v, ok := mapData[thirdparty]

		var p models.ThirdpartyStatistics
		if ok {
			p = v
		} else {
			m := models.ThirdpartyStatistics{}
			m.Thirdparty = thirdparty
			m.Id = math.MaxInt64
			m.StatisticsDate = strDate
			p = m
		}

		if isSuccess == 0 {
			p.Fail = count
		} else {
			p.Success = count
		}

		mapData[thirdparty] = p
	}

	for _, v := range mapData  {
		list = append(list, v)
	}

	return
}

func GetThirdpartyMonitorKey(date int64) string {
	return getMonitorKey(thirdpartyPrefix, date)
}

func IsThirdpartyKeyExist(date int64) bool {
	return IsKeyExist(thirdpartyPrefix, date)
}

func DelThirdpartyKey(date int64) {
	DelKey(thirdpartyPrefix, date)
}

func IncrThirdpartyCount(thirdparty int, status int) {
	isSuccess := 0
	if status / 100 == 2 {
		isSuccess = 1
	}
	field := getThirdpartyField(thirdparty, isSuccess)

	incrCount(thirdpartyPrefix, tools.NaturalDay(0), field)
}

