package main

import (
	"flag"
	"time"

	_ "micro-loan/common/lib/db/mysql"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/service"
	"micro-loan/common/tools"
)

var freq int
var interval int64

func init() {
	flag.IntVar(&freq, "freq", 10, "freq (minute)")
	flag.Int64Var(&interval, "interval", 10, "interval (minute)")

}

var monitorTask = []string{"identity_detect", "need_review_order", "wait4loan_order", "event_push", "monitor", "timer_task", "ticket_realtime_assign_task"}

func main() {
	flag.Parse()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	timetag := tools.GetUnixMillis()

	title := "task monitor"

	for _, v := range monitorTask {
		lockstr := "lock:" + v
		dateKey := "monitor:" + v

		hValue, err := storageClient.Do("GET", lockstr)
		if err != nil || hValue == nil {
			service.SendNotification(dateKey, freq, title, v+" task not exist")
			continue
		}

		iValue, err := tools.Str2Int64(string(hValue.([]byte)))
		if err != nil {
			service.SendNotification(dateKey, freq, title, v+" task not exist")
			continue
		}

		if timetag-iValue > int64(time.Minute)*interval {
			service.SendNotification(dateKey, freq, title, v+" task don't beat")
			continue
		}
	}
}
