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

	clientInfo := models.ClientInfo{}
	o := orm.NewOrm()
	o.Using(clientInfo.Using())

	for {
		var data []ClientInfoTmp
		sql := fmt.Sprintf(`SELECT id, imei FROM %s WHERE imei_md5 = '' ORDER BY id ASC LIMIT 100`,
			clientInfo.TableName())
		num, err := o.Raw(sql).QueryRows(&data)
		if err != nil || num <= 0 {
			logs.Info("[%s] 没有更多待处理数据了...", procTitle)
			break
		}

		for _, v := range data {

			clientInfo.Id = v.Id
			clientInfo.Imei = v.Imei
			logs.Info("处理客户端数据(client_info), id：", clientInfo.Id)

			clientInfo.ImeiMd5 = tools.Md5(v.Imei)
			num, err := o.Update(&clientInfo, "imei_md5")
			if num == 1 && err == nil {
				logs.Debug("----- 客户端数据Imei_Md5修改成功：", clientInfo)
			} else {
				logs.Error("----- 客户端数据Imei_Md5修改失败：", clientInfo)
			}
		}
	}

	storageClient.Do("DEL", lockKey)
	logs.Info("[%s] politeness exit.", procTitle)
}
