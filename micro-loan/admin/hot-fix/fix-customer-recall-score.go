package main

import (
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/fantasy"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

func getOrderDatas(latesId int64, limitTime int64) (list []models.Order) {
	order := models.Order{}
	o := orm.NewOrm()

	o.Using(order.Using())

	o.QueryTable(order.TableName()).
		Filter("check_status", types.LoanStatusReject).
		Filter("risk_ctl_status", types.RiskCtlAFReject).
		Filter("risk_ctl_regular", types.RegularNameZ002).
		Filter("id__gt", latesId).
		OrderBy("id").
		Limit(100).
		All(&list)

	return
}

func addScoreByOrder(order models.Order) {
	logs.Info("[addScoreByOrder] handle order :%d", order.Id)

	risk := fantasy.NewSingleRequestByOrderPt(&order)
	//logs.Info("risk:%#v", risk)

	score, _ := risk.GetAScoreV1()
	logs.Info("[addScoreByOrder] handle order . score:%d", score)

	if service.CanAddCustumetRecallScore(score, &order) {
		err := service.ChangeCustomerRecall(order.UserAccountId, order.Id, types.RecallTagScore, types.RemarkTagNone)
		if err != nil {
			logs.Error("[addScoreByOrder]  addScoreTag accountId:%d err:%v", order.UserAccountId, err)
		}
	}
}

func main() {
	// 设置进程 title
	procTitle := "fix-customer-recall-score"
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

	scoreDayN, _ := config.ValidItemInt("customer_recall_score_z002_N")
	latesId := int64(0)

	limitTime := tools.GetUnixMillis() - int64(scoreDayN)*tools.MILLSSECONDADAY
	logs.Info("limitTime:%d", limitTime)

	for {
		list := getOrderDatas(latesId, limitTime)
		if len(list) == 0 {
			logs.Info(" len(list) =0")
			break
		}

		for _, v := range list {
			latesId = v.Id
			addScoreByOrder(v)
		}
	}

	// all record
	logs.Warn("fix-customer-recall-score ok")
	logs.Info("[%s] politeness exit.", procTitle)
}
