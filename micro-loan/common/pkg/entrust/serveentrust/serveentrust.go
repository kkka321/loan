package serveentrust

import (
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"

	"github.com/astaxie/beego/logs"

	"github.com/astaxie/beego"
)

// EntrustRepayList 还款时加入还款list，供催收公司抓取
func EntrustRepayList(orderID int64) (ok bool) {
	if orderID == 0 {
		ok = false
		return
	}
	ordersExt, err := models.GetOrderExt(orderID)
	if err == nil && ordersExt.IsEntrust == 1 && ordersExt.EntrustPname != "" {
		prefix := beego.AppConfig.String("entrust_notify_repay_queue_prefix")
		if prefix != "" {
			pnameKey := prefix + ordersExt.EntrustPname
			logs.Debug("[EntrustRepayList] pnamekey:", pnameKey)
			storageClient := storage.RedisStorageClient.Get()
			defer storageClient.Close()
			storageClient.Do("LPUSH", pnameKey, orderID)
			ok = true
		}
	}
	return
}
