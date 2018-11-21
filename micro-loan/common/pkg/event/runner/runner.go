package runner

import (
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

// ParseAndRun 解析 eventBytes,并运行事件
// 将 eventBytes 解析成 event , 根据已注册事件,找到其对应的 异步运行方法, 并运行
func ParseAndRun(eventBytes []byte) (success bool, err error) {
	// 崩溃log记录
	defer func() {
		if x := recover(); x != nil {
			logs.Error("[Run] panic data:%s, err:%v", string(eventBytes), x)
			logs.Error(tools.FullStack())
		}
	}()

	return globalParser.Run(eventBytes)
}

// ParseAndRunSeperate 解析 eventBytes,并运行事件
// 将 eventBytes 解析成 event , 根据已注册事件,找到其对应的 异步运行方法, 并运行
func ParseAndRunSeperate(eventBytes []byte, eventName string) (success bool, err error) {
	// 崩溃log记录
	defer func() {
		if x := recover(); x != nil {
			logs.Error("[Run] panic data:%s, err:%v", string(eventBytes), x)
			logs.Error(tools.FullStack())
		}
	}()

	return globalParser.RunSepEvent(eventBytes, eventName)
}

// QueueConsumerGoroutineConfig 返回
func QueueConsumerGoroutineConfig(queueName string, def ...int) int {
	prefix := "goroutine_num_"
	globalDef := 1
	// 1
	num, err := config.ValidItemInt(prefix + queueName)
	if err != nil {
		if len(def) > 1 {
			num = def[0]
		} else {
			num = globalDef
		}
	}
	return num
}

// GetSeparateQueues
func GetSeparateQueues() []string {
	queues := []string{}
	seperateMap := evtypes.SeparateQueueMap()
	for k := range seperateMap {
		queues = append(queues, evtypes.GetQueueName(k))
	}
	return queues
}
