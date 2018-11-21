package main

import (
	"fmt"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/service"
	"micro-loan/common/tools"
)

var procTitle = "fix-roll"

func main() {
	setProcTitle()
	doFixRoll()
}

func setProcTitle() {
	// 设置进程 title
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)
}

func doFixRoll() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := fmt.Sprintf("lock:%s", procTitle)
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}

	var orderID int64 = 180627022898492305
	service.HandleRollOrder(orderID)

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[%s] politeness exit.", procTitle)
}
