package task

import (
	"fmt"
	"sync"

	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"

	"micro-loan/common/tools"
)

//! see: https://github.com/adonovan/gopl.io/blob/master/ch8/du4/main.go

// 多进程工作基本原则: 如果生产数据和消费数据依赖同一个状态值,则要先生产再消费;如果生产消费完全独立,生产和消费可以并行.
// 主要是解决多进程生产和消费数据竞争的问题

// 不带参的工作函数
type TaskWork0 interface {
	Start()
	Cancel()
}

var taskWork0Map = map[string]TaskWork0{
	"identity_detect":             new(IdentityDetectTask),
	"need_review_order":           new(ReviewOrderTask),
	"wait4loan_order":             new(Wait4LoanOrderTask),
	"invalid_order":               new(InvalidOrderTask),
	"overdue_order":               new(OverdueOrderTask),
	"repay_voice_order":           new(RepayVoiceOrderTask),
	"overdue_auto_call":           new(OverdueAutoCallTask),
	"timer_task":                  new(TimerTask),
	"event_push":                  new(EventPushTask),
	"monitor":                     new(MonitorTask),
	"ticket_realtime_assign_task": new(TicketRealtimeAssignTask),
	"auto_reduce_order":           new(AutoReduceOrderTask),
	"bigdata_contact":             new(BigdataContactTask),
	"customer_recall":             new(CustomerRecallTask),
	"register_remind":             new(RegisterRemindTask),
	"info_review_auto_call_task":  new(InfoReviewAutoCallTask),
	"author_status_check":         new(AuthoriationStatusCheck),
}

func TaskWork0Map() map[string]TaskWork0 {
	return taskWork0Map
}

/** 从 golang 圣经里面抄来的代码,用于广播事件 */

//!+1
var done = make(chan struct{})

func cancelled() bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}

var mutex sync.Mutex
var currentDatas map[string]interface{} = make(map[string]interface{})

func addCurrentData(key string, value interface{}) {
	mutex.Lock()

	currentDatas[key] = value

	mutex.Unlock()
}

func removeCurrentData(key string) {
	mutex.Lock()

	delete(currentDatas, key)

	mutex.Unlock()
}

func GetCurrentData() string {
	mutex.Lock()

	str := fmt.Sprintf("%v", currentDatas)

	mutex.Unlock()

	return str
}

var lastTimetag int64

func TaskHeartBeat(coon redis.Conn, lockKey string) {
	nowT := tools.GetUnixMillis()

	if lastTimetag == 0 {
		lastTimetag = nowT
		return
	}

	if nowT-lastTimetag < int64(1000*60*10) {
		return
	}

	lastTimetag = nowT

	_, err := coon.Do("SET", lockKey, nowT)
	if err != nil {
		logs.Error("[TaskHeartBeat] set key error time:%d, err:%v", nowT, err)
	}
}

//!-1
