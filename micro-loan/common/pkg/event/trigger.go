package event

import (
	"encoding/json"
	"fmt"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/pkg/event/evtypes"

	"github.com/astaxie/beego/logs"
)

// Trigger 触发事件
// persistentParam 必须是在 event/evtypes/persistent_param.go
// 异步运行方法必须定义在 runevent包中
// Event trigger, will be import by anywhere to trigger event
// calls events
func Trigger(persistentParam interface{}) (ok bool, err error) {
	if persistentParam == nil {
		err = fmt.Errorf("[event.Trigger]persistentParam can not be nil, persistentParam:%v", persistentParam)
		logs.Error(err)
		return
	}

	ok = false
	var eqv evtypes.QueueVal
	eqv.EventName = evtypes.GetStructName(persistentParam)
	eqv.Data, _ = json.Marshal(persistentParam)

	// 如果配置, 即时触发, 可以直接此处 run event

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	//key := beego.AppConfig.String("event_queue")
	key := evtypes.GetQueueName(eqv.EventName)
	if evtypes.IsSeparate(eqv.EventName) {
		// 独立队列不需要再存储 外部事件名结构
		_, err = storageClient.Do("LPUSH", key, eqv.Data)
	} else {
		data, _ := json.Marshal(eqv)
		_, err = storageClient.Do("LPUSH", key, data)
	}

	if err != nil {
		logs.Error("[event.Trigger]Event Queue, LPUSH", err, "; Event: ", persistentParam)
		return
	}
	ok = true
	return
}
