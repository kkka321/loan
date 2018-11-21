package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/pkg/event/runner"
	"time"

	"github.com/astaxie/beego"
)

func main() {
	// event.Trigger(&evtypes.LoanSubmitEv{12321312312, tools.GetUnixMillis()})
	testConsumerEventPush()
}

func testConsumerEventPush() {
	qName := beego.AppConfig.String("event_queue")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()
	qValueByte, err := storageClient.Do("RPOP", qName)
	// 没有可供消费的数据,退出工作 goroutine
	if err != nil || qValueByte == nil {
		//logs.Info("[EventPush] no data for consume, I will exit after 500ms, workID:", workerID)
		time.Sleep(500 * time.Millisecond)
		return
	}

	// 真正开始工作了
	// EventParser.Run(qValueByte.([]byte))
	runner.ParseAndRun(qValueByte.([]byte))

}
