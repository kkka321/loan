package main

import (
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"

	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"micro-loan/common/lib/redis/cache"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty/xendit"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

var xenditDisburseApi = "https://api.xendit.co/disbursements"
var xenditDisburseCallbackApi = "/xendit/disburse_fund_callback/create"

func getOrders(lasteOrderId int64) (list []models.Order, err error) {
	obj := models.Order{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	_, err = o.QueryTable(obj.TableName()).
		Filter("id__gt", lasteOrderId).
		Filter("check_status", types.LoanStatusIsDoing).
		OrderBy("id").
		Limit(100).
		All(&list)
	return
}

func checkThirdPartyPass(orderId int64) bool {
	third := models.ThirdpartyRecord{}
	o := orm.NewOrm()
	o.Using(third.Using())

	disburseList := []models.ThirdpartyRecord{}
	disburseCallbackList := []models.ThirdpartyRecord{}

	// call
	_, err := o.QueryTable(third.TableName()).
		Filter("api", xenditDisburseApi).
		Filter("related_id", orderId).
		OrderBy("-id").
		All(&disburseList)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("checkThirdPartyPass xenditDisburseApi err")
		return false
	}

	//callback
	_, err = o.QueryTable(third.TableName()).
		Filter("api", xenditDisburseCallbackApi).
		Filter("related_id", orderId).
		OrderBy("-id").
		All(&disburseCallbackList)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("checkThirdPartyPass xenditDisburseApi err")
		return false
	}

	// 发生了调用
	if len(disburseList) > 0 {
		if len(disburseList) != 1 || len(disburseCallbackList) != 1 {
			logs.Error("order may no need to fix. len(disburseList) :%d len(disburseCallbackList) :%d", len(disburseList), len(disburseCallbackList))
			return false
		}

		logs.Info("respone:%#v", disburseCallbackList[0].Request)

		// check response
		resp := xendit.XenditDisburseFundCallBackData{}
		uStr := ""
		_ = json.Unmarshal([]byte(disburseCallbackList[0].Request), &uStr)
		_ = json.Unmarshal([]byte(uStr), &resp)

		logs.Info("resp after Unmarshal :%#v", resp)
		if resp.Status == "FAILED" && resp.FailureCode == "INSUFFICIENT_BALANCE" {
			return true
		} else {
			logs.Error("resp check not pass. resp:%#v", resp)
			return false
		}
	} else {
		//logs.Warn("len(disburseList) == 0 未发生调用暂不修复")
		return checkDisburseLog(orderId)
	}
	// 未发生调用
	return false
}

func checkDisburseLog(orderID int64) bool {
	maxTime := int64(1536033600000)
	dLog := models.DisburseInvokeLog{}
	o := orm.NewOrm()
	o.Using(dLog.Using())

	list := []models.DisburseInvokeLog{}
	o.QueryTable(dLog.TableName()).Filter("order_id", orderID).All(&list)

	if len(list) != 1 {
		logs.Error("checkDisburseLog len(list) :%d", len(list))
		return false
	}

	dLog = list[0]

	if dLog.Ctime > maxTime {
		logs.Error("checkDisburseLog Ctime:%d", dLog.Ctime)
		return false
	}

	if dLog.DisbursementId != "" || dLog.DisbureStatus != 0 || dLog.HttpCode != 0 {
		logs.Error("checkDisburseLog err dLog:%#v", dLog)
		return false
	}

	return true

}
func fixOrder(order models.Order) (err error) {

	if order.CheckStatus != types.LoanStatusIsDoing {
		return fmt.Errorf("fixOrder checkstatus err. status:%d", order.CheckStatus)
	}

	if !checkThirdPartyPass(order.Id) {
		return fmt.Errorf("checkThirdPartyPass false")
	}

	// 清锁
	cacheClient := cache.RedisCacheClient.Get()
	defer cacheClient.Close()
	keyPrefix := beego.AppConfig.String("disburse_order_lock")
	key := fmt.Sprintf("%s%d", keyPrefix, order.Id)
	cacheClient.Do("DEL", key)

	originOrder := order
	order.CheckStatus = types.LoanStatusWait4Loan
	order.Utime = tools.GetUnixMillis()
	models.UpdateOrder(&order)

	// 写操作日志
	models.OpLogWrite(9999, order.Id, models.OpCodeOrderUpdate, order.TableName(), originOrder, order)
	return nil
}

func main() {
	// 设置进程 title
	// +1 分布式锁
	// -1 正常退出时,释放锁
	procTitle := "fix-xendit-balance"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := fmt.Sprintf("lock:%s", procTitle)
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}
	defer storageClient.Do("DEL", lockKey)

	lasteOrderId := int64(0)

	for {
		list, err := getOrders(lasteOrderId)
		if len(list) == 0 || err != nil {
			logs.Warn("fix over break . len(list):%d, err:%v", len(list), err)
			break
		}

		for _, v := range list {
			logs.Info("handel orderID:%d", v.Id)
			lasteOrderId = v.Id
			err := fixOrder(v)

			if err != nil {
				logs.Error("[range]fixOrder err:%v id:%d", err, v.Id)
			}
		}
	}

	logs.Warn("fix ok")
	logs.Info("[%s] politeness exit.", procTitle)
}
