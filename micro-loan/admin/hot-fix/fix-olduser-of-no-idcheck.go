package main

import (
	"flag"
	"fmt"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"github.com/erikdubbelboer/gspt"
)

var maxCheckedID int64

func init() {
}

func main() {

	flag.Int64Var(&maxCheckedID, "id", 0, "Max fixed User admin id.不包括当前用户ID")
	flag.Parse()
	procTitle := "fix-olduser-idcheck"
	gspt.SetProcTitle(procTitle)
	logs.Info("[%s] start launch.", procTitle)
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// 分布式锁
	lockKey := fmt.Sprintf("lock:%s", procTitle)
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}
	defer storageClient.Do("DEL", lockKey)

	accountBase := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(accountBase.Using())

	queueName := beego.AppConfig.String("account_identity_detect")

	for {
		var findData []int64
		fmt.Println("query maxcheckedID:", maxCheckedID)

		sql := fmt.Sprintf("SELECT id FROM %s WHERE id>=180530000000000000 AND id < %d AND realname <> '' AND identity<>'' AND third_id='' ORDER BY id DESC LIMIT 500", accountBase.TableName(), maxCheckedID)
		num, err := o.Raw(sql).QueryRows(&findData)
		if err != nil || num == 0 {
			logs.Info("[%s] 没有更多待处理数据了, 停止任务...", procTitle)
			break
		}
		logs.Info("正在进入队列:", findData)
		for _, id := range findData {
			lenQ, err := storageClient.Do("LPUSH", queueName, id)
			if err != nil && lenQ.(int) <= 0 {
				logs.Error("[%s] 插入队列不成功, ", procTitle, id)
				fmt.Println("MinCheckedID:", maxCheckedID)
				break
			}
			maxCheckedID = id
		}

		// qLen, err := storageClient.Do("LPUSH", queueName, findData...)
		//
		// if err != nil || qLen.(int) <= 0 {
		// 	logs.Error("[%s] 插入队列不成功, ", procTitle, findData)
		// }
	}

}

// // before 180406010094164047
// func getUserOfNoIDCheck() {
//
// }
