package main

import (
	"fmt"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/erikdubbelboer/gspt"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
)

type FindData struct {
	AccountId int64
	Realname  string
}

func main() {
	// 设置进程 title
	procTitle := "fix-realname"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := fmt.Sprintf("lock:%s", procTitle)
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}

	var lastID int64
	accountBase := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(accountBase.Using())

	for {
		var findData []FindData
		sql := fmt.Sprintf(`SELECT id AS account_id, realname FROM %s WHERE id > %d ORDER BY id ASC LIMIT 100`,
			accountBase.TableName(), lastID)
		num, err := o.Raw(sql).QueryRows(&findData)
		if err != nil || num <= 0 {
			logs.Info("[%s] 没有更多待处理数据了...", procTitle)
			break
		}

		for _, accountData := range findData {
			realname := tools.TrimRealName(accountData.Realname)
			if realname != accountData.Realname {
				// 说明有特殊字符被过滤掉了
				accountBase.Id = accountData.AccountId
				accountBase.Realname = realname
				num, err = o.Update(&accountBase, "realname")
				dataJSON, _ := tools.JsonEncode(accountData)
				if err != nil || num != 1 {
					logs.Error("[%s] updata realname hos wrong, data: %s", procTitle, dataJSON)
				} else {
					logs.Info("[%s] updata realname success. origin: %s, after: %s", procTitle, dataJSON, realname)
				}
			}

			lastID = accountData.AccountId
		}
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[%s] politeness exit.", procTitle)
}
