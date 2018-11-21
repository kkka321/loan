package main

import (
	"fmt"
	"strings"

	// 数据库初始化
	"micro-loan/common/dao"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/models"
	"micro-loan/common/service"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/tools"
)

type ClientInfoTmp struct {
	Id   int64
	Imei string
}

func main() {
	// 设置进程 title
	procTitle := "fix-clientinfo-imeimd5"
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

	var idsBox []string
	idsBox = append(idsBox, "1")
	orderList, _ := service.GetRepayVoiceRemindOrderList(idsBox, 1)

	var mobileArr []string
	for _, orderID := range orderList {
		order, _ := models.GetOrder(orderID)
		accountBase, _ := dao.CustomerOne(order.UserAccountId)
		if len(accountBase.Mobile) > 0 {
			mobile := accountBase.Mobile
			if mobile[0:2] == "08" {
				mobile = strings.Replace(mobile, "08", "628", 1)
			}
			mobileArr = append(mobileArr, mobile)
		}
	}
	fmt.Println("====== mobile:", mobileArr)
	mobileStr := strings.Join(mobileArr, ",")
	fmt.Println("====== mobileStr:", mobileStr)

	storageClient.Do("DEL", lockKey)
	logs.Info("[%s] politeness exit.", procTitle)
}
