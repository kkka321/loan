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
	"micro-loan/common/service"
	"micro-loan/common/tools"
)

var procTitle = "mark-customer"

func main() {
	setProcTitle()
	doMarkCustomer()
}

func setProcTitle() {
	// 设置进程 title
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)
}

func doMarkCustomer() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := fmt.Sprintf("lock:%s", procTitle)
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}

	order := models.Order{}
	o := orm.NewOrm()
	o.Using(order.Using())

	var lastID int64
	for {
		var orderList []models.Order
		sql := fmt.Sprintf(`SELECT o.* FROM %s o
LEFT JOIN %s a ON a.id = o.user_account_id 
WHERE o.loan_time > 0 AND o.risk_ctl_regular != "" AND a.random_mark = 0 AND o.id > %d
ORDER BY o.id LIMIT 100`,
			models.ORDER_TABLENAME,
			models.ACCOUNT_BASE_TABLENAME,
			lastID)

		num, err := o.Raw(sql).QueryRows(&orderList)
		if err != nil || num <= 0 {
			logs.Warning("[%s] 没有更多数据了. sql: %s", procTitle, sql)
			break
		}

		for _, findData := range orderList {
			lastID = findData.Id
			service.MarkCustomerIfHitRandom(findData)
		}
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[%s] politeness exit.", procTitle)
}
