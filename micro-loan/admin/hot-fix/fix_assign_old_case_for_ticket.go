package main

import (
	"flag"
	"fmt"
	"micro-loan/common/pkg/ticket"
	"time"

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

type FindData struct {
	Id        int64
	CaseLevel string
	OrderId   int64
	UrgeUid   int64
}

func main() {
	var maxCheckedID int64

	flag.Int64Var(&maxCheckedID, "id", 0, "Max fixed User admin id.不包括当前用户ID")
	flag.Parse()
	// 设置进程 title
	procTitle := "fix_assign_old_case_for_ticket"
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

	overdueCase := models.OverdueCase{}
	ocOrm := orm.NewOrm()
	ocOrm.Using(overdueCase.Using())

	var findData []FindData
	sql := fmt.Sprintf(`select case_level, urge_uid,order_id, id from %s where is_out=0 and join_urge_time<1530532800000 and urge_uid>0 and id<=%d`,
		overdueCase.TableName(), maxCheckedID)
	num, err := ocOrm.Raw(sql).QueryRows(&findData)
	if err != nil || num <= 0 {
		logs.Info("[%s] 没有更多待处理数据了...", procTitle)
		return
	}

	for _, fd := range findData {

		logs.Warn("Start case id:", fd.Id, "case level:", fd.CaseLevel, "Order ID:", fd.OrderId)
		order, _ := models.GetOrder(fd.OrderId)
		ticketItem := types.OverdueLevelTicketItemMap()[fd.CaseLevel]
		oldTicket, _ := models.GetTicketByItemAndRelatedID(ticketItem, fd.Id)
		if oldTicket.Id > 0 && int(oldTicket.Status) < 5 {
			logs.Warn("No change, exists ticket:", oldTicket.Id, ";status:", oldTicket.Status)
			continue
		}
		id, _ := ticket.CreateTicket(types.OverdueLevelTicketItemMap()[fd.CaseLevel], fd.Id, types.Robot, fd.OrderId, order.UserAccountId, nil)
		logs.Warn("End, ticketID:", id)

		time.Sleep(time.Millisecond * 500)
	}
}
