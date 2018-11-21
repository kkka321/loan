package sms

import (
	"errors"
	"fmt"
	"micro-loan/common/lib/redis/cache"
	"micro-loan/common/types"
	"sort"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

// 更灵活的策略配置,待实现
// sms_service_strategy = {"default":{"sms253":1,"nexmo":2},"serviceType":[{"serviceType":2,"rule":{"sms253":1,"nexmo":2}}]}

// 获取策略中设置的发送者关键字 priority_and_failed_next
func priorityAndFailedNext(serviceType types.ServiceType, mobile string) (senderKey types.SmsServiceID, err error) {
	var next int

	if len(senderMap) == 1 {
		for k := range senderMap {
			senderKey = k
			return
		}
	}
	if serviceType != types.ServiceRequestLogin {
		next = strategySort[0]
	} else {
		lastFailed, _ := getMobileLastFailedSender(mobile)

		if s, ok := senderMap[lastFailed]; !ok {
			// 不存在 lastFailed 从起始开始
			next = strategySort[0]
		} else {
			index := sort.SearchInts(strategySort, s)
			if index == len(strategySort)-1 {
				// the last one
				// or err
				next = strategySort[0]
			} else {
				next = strategySort[index+1]
			}
		}
	}

	for k, v := range senderMap {
		if v == next {
			senderKey = k
			return
		}
	}
	err = errors.New("Not found next sender strategy")
	return
}

func getMobileLastFailedSender(mobile string) (lastFailed types.SmsServiceID, err error) {
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()
	key := beego.AppConfig.String("mobile_sms_failed_cache_prefix") + mobile

	val, err := cacheClient.Do("GET", key)
	if err != nil {
		return
	}
	if val == nil {
		return
	}

	temp, _ := strconv.Atoi(string(val.([]byte)))
	lastFailed = types.SmsServiceID(temp)
	fmt.Println(lastFailed)
	return
}

// setFailedCacheForStrategy 设置失败缓存
func setFailedCacheForStrategy(mobile string, sender types.SmsServiceID) error {
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()
	key := beego.AppConfig.String("mobile_sms_failed_cache_prefix") + mobile

	ex, err := beego.AppConfig.Int("sms_mobile_failed_sender_expire")
	if err != nil {
		logs.Error(err)
		logs.Error("[sms strategy failedNext required config] SetFailedForStrategy func will failed")
		return err
	}
	_, err = cacheClient.Do("SET", key, sender, "EX", ex*int(time.Hour/time.Second))
	if err != nil {
		logs.Error("[Redis Error]:", err)
	}
	return err
}
