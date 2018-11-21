package main

import (
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	"github.com/astaxie/beego/orm"

	"micro-loan/common/dao"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"micro-loan/common/models"
)

func main() {

	procTitle := "fix_ticket_close_byentust"
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

	for {

		ordersExt := models.OrderExt{}
		o := orm.NewOrm()
		o.Using(ordersExt.UsingSlave())
		sql := fmt.Sprintf("select * from orders_ext where is_entrust=1 and order_id>%d order by order_id asc limit 100 ", lastID)

		var ordersExts = []models.OrderExt{}
		num, err := o.Raw(sql).QueryRows(&ordersExts)

		if err != nil || num <= 0 {
			logs.Info("[%s] 没有更多待处理数据了...", procTitle)
			break
		}
		logs.Debug("[fix_ticket_close_byentust] ordersExts:", ordersExts)

		for _, c := range ordersExts {
			lastID = c.OrderId
			logs.Debug("[fix_ticket_close_byentust] is_entrust:", c.IsEntrust)

			oneCase, _ := dao.GetInOverdueCaseByOrderID(c.OrderId)
			item := types.MustGetTicketItemIDByCaseName(oneCase.CaseLevel)
			//关闭工单
			ticket.CloseByRelatedID(oneCase.Id, item, types.TicketCloseReasonEntrust)

		}
	}
	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[%s] politeness exit.", procTitle)

}
