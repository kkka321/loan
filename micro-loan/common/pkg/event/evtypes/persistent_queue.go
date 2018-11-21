package evtypes

// QueueVal 描述触发事件在队列里的值
// 持久化参数结构体名, 会作为 EventName 存入队列, 方便获取
type QueueVal struct {
	EventName string `json:"n"`
	Data      []byte `json:"d"`
}
