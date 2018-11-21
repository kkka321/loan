package main

import (
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"

	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

func getOneInvalidOrder(lastId int64) (order models.Order) {
	o := orm.NewOrm()
	o.Using(order.Using())

	o.QueryTable(order.TableName()).
		Filter("fixed_random", service.FixedPhoneVerifySet2Invalid).
		Filter("id__gt", lastId).
		OrderBy("id").
		One(&order)
	return
}

func getAllFirstApplyOrders(order models.Order) (orders []models.Order) {
	o := orm.NewOrm()
	o.Using(order.Using())

	o.QueryTable(order.TableName()).
		Filter("user_account_id", order.UserAccountId).
		Filter("is_reloan", 0).
		Filter("apply_time__gte", order.ApplyTime).
		OrderBy("id").
		All(&orders)

	return
}

func fixTag(order models.Order) {
	logs.Info("fixTag handel order:%d", order.Id)
	aExt, _ := models.OneAccountBaseExtByPkId(order.UserAccountId)
	if aExt.PhyInvalidTag == types.PhoneVerifyInvalidTag {
		logs.Warn("no need add account:%d tag", order.UserAccountId)
		return
	}

	logs.Info("add order:%d account :%d tag", order.Id, order.UserAccountId)
	service.AddAccountBaseExtPhyInvalidTag(order.UserAccountId, 0)
	orders := getAllFirstApplyOrders(order)
	for _, v := range orders {
		logs.Info("add order:%d tag", v.Id)
		service.AddOrdersExtPhyInvalidTag(v.Id, 0)
	}
}

func main() {

	procTitle := "fix_phy_invalid_order_tag"
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
	// -1 正常退出时,释放锁
	defer storageClient.Do("DEL", lockKey)

	lastId := int64(0)

	for {
		logs.Info("====================>>>>>>>>>>")
		logs.Info("lastId:%d", lastId)
		order := getOneInvalidOrder(lastId)
		if order.Id == 0 {
			logs.Warn("order.id ==0, so break")
			break
		}
		lastId = order.Id
		fixTag(order)
	}

	logs.Notice("[%s] politeness exit.", procTitle)
}
