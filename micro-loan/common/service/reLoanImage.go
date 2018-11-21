package service

import (
	"encoding/json"
	
	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
)

func AddReLoanImage(userAccountId int64, resourceId int64) (id int64, err error) {
	reLoanImage := models.ReLoanImage{}
	reLoanImage.UserAccountId = userAccountId
	reLoanImage.ReLoanPhoto = resourceId
	reLoanImage.Ctime = tools.GetUnixMillis()
	id, err = models.AddReLoanImage(reLoanImage)
	return
}

func UploadReLoanPhotoSuccess(accountId int64) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// 1. 更新数据库
	order, _ := dao.AccountLastTmpLoanOrder(accountId)
	order.Utime = tools.GetUnixMillis()
	order.IsUpHoldPhoto = 1
	order.Update("is_up_hold_photo", "utime")

	// 2. 更新redis
	keyName := dao.UserLatestOrderKey() + tools.Int642Str(accountId)
	//SET key value [EX seconds]
	orderJson, _ := json.Marshal(order)
	storageClient.Do("SET", keyName, orderJson, "EX", 3600*2)

	return
}
