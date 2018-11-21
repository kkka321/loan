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
}
type Orders struct {
	Id          int64
	CheckStatus int64
}

func main() {
	// 设置进程 title
	procTitle := "fix-order-isreloan"
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
	orderModel := models.Order{}
	o := orm.NewOrm()
	o.Using(accountBase.Using())
	oo := orm.NewOrm()
	oo.Using(orderModel.Using())

	for {
		var findData []FindData
		sql := fmt.Sprintf(`SELECT id AS account_id FROM %s WHERE id > %d ORDER BY id ASC LIMIT 100`,
			accountBase.TableName(), lastID)
		num, err := o.Raw(sql).QueryRows(&findData)
		if err != nil || num <= 0 {
			logs.Info("[%s] 没有更多待处理数据了...", procTitle)
			break
		}

		for _, accountData := range findData {

			lastID = accountData.AccountId

			logs.Info("处理该用户订单，accountID：", lastID)

			var orders []Orders

			//查询该用户订单,集中处理 ，没有订单跳过
			sql := fmt.Sprintf(`SELECT id,check_status FROM %s WHERE user_account_id = %d ORDER BY id ASC`,
				orderModel.TableName(), lastID)

			num, err := o.Raw(sql).QueryRows(&orders)
			if err != nil || num <= 0 {
				logs.Info("[%d] 此用户没有订单", lastID)
			} else {
				logs.Debug("订单：", orders)
				if len(orders) > 0 {

					re := false
					for _, v := range orders {

						if re == true {
							orderModel.Id = v.Id
							orderModel.IsReloan = 1

						} else {
							orderModel.Id = v.Id
							orderModel.IsReloan = 0

						}

						// logs.Debug("==修改：", orderModel)
						num, err := oo.Update(&orderModel, "is_reloan")

						if num == 1 && err == nil {
							logs.Debug("==修改成功：", orderModel)
						} else {
							logs.Error("==修改失败：", orderModel)
						}

						if v.CheckStatus == 8 {
							re = true
						}

					}

				}

			}

		}
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[%s] politeness exit.", procTitle)
}
