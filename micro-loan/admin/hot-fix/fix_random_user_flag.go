package main

import (
	"encoding/json"
	"fmt"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/erikdubbelboer/gspt"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
)

type findData struct {
	AccountID int64 `orm:"column(user_account_id)"`
	OrderID   int64 `orm:"column(id)"`
}

func main() {
	// 设置进程 title
	procTitle := "fix_random_user_flag"
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
	defer storageClient.Do("DEL", lockKey)

	var lastID int64
	for {
		order := models.Order{}
		o := orm.NewOrm()
		o.Using(order.Using())

		var fds []findData
		// 存量数据较少, 所以不做分批处理
		sql := fmt.Sprintf(`SELECT * FROM %s WHERE loan_time>0 AND risk_ctl_status=%d AND risk_ctl_regular="" AND id>%d ORDER BY id ASC LIMIT 100`,
			order.TableName(), types.RiskCtlPhoneVerifyReject, lastID)
		num, err := o.Raw(sql).QueryRows(&fds)
		if err != nil || num <= 0 {
			logs.Info("[%s] 没有更多待处理数据了...", procTitle)
			break
		}

		for _, fd := range fds {
			account := models.AccountBase{}
			account.Id = fd.AccountID
			err := o.Read(&account)
			if err != nil {
				logs.Error("[fix_random_user_flag] cannot find user", account, ";error:", err, ";find data:", fd)
				continue
			}
			if account.RandomMark == 0 {
				// 说明有特殊字符被过滤掉了
				originData, _ := json.Marshal(account)
				account.RandomMark = fd.OrderID
				num, err = o.Update(&account, "RandomMark")
				dataJSON, _ := tools.JsonEncode(account)
				if err != nil || num != 1 {
					logs.Error("[%s] updata RandomMark hos wrong, data: %s", procTitle, dataJSON)
				} else {
					logs.Info("[%s] updata RandomMark success. origin: %s, after: %s", procTitle, originData, dataJSON)
				}
			} else {
				logs.Warning("User[%d]Already marked as : %d, ignore it", account.Id, account.RandomMark)
			}

			lastID = fd.OrderID
		}
	}

	// -1 正常退出时,释放锁
	logs.Info("[%s] politeness exit.", procTitle)
}
