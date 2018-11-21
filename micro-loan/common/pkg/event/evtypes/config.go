package evtypes

import "github.com/astaxie/beego"

var separateQueueMap = map[string]bool{
	"FixPaymentCodeEv": true,
}

// GetQueueName 根据事件名获取对应的统一队列名或者独立队列名
func GetQueueName(eventName string) string {
	globalQueueName := beego.AppConfig.String("event_queue")

	if _, ok := separateQueueMap[eventName]; ok {
		globalQueueName += "_" + eventName
	}

	return globalQueueName
}

func SeparateQueueMap() map[string]bool {
	return separateQueueMap
}

func IsSeparate(eventName string) bool {
	if _, ok := separateQueueMap[eventName]; ok {
		return true
	}
	return false
}
