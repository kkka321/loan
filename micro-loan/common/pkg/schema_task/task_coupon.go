package schema_task

import (
	"github.com/astaxie/beego/logs"
	"github.com/aws/aws-sdk-go/aws"

	"micro-loan/common/lib/gaws"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/coupon_event"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"
	"time"

	"micro-loan/common/dao"

	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
)

func DistributeCoupon(id int64) error {
	taskInfo, err := models.GetCouponTask(id)
	if err != nil {
		logs.Error("[DistributeCoupon] GetCouponTask return error taskId:%d, err:%v", id, err)
		return err
	}

	return runCouponTask(&taskInfo, 0)
}

func runCouponTask(task *models.CouponTask, param interface{}) error {
	if task.CouponTarget == types.CouponTargetCustom {
		distributeCustomCoupon(task, param)
	} else if task.CouponTarget == types.CouponTargetRegisterNoOrder {
		distributeRegisterNoOrder(task, param)
	} else if task.CouponTarget == types.CouponTargetRegisterTmpOrder {
		distributeRegisterTmpOrder(task, param)
	} else if task.CouponTarget == types.CouponTargetRepayClear {
		distributeRepayClear(task, param)
	} else if task.CouponTarget == types.CouponTargetRepayOverdue {
		distributeRepayOverdue(task, param)
	}

	return nil
}

func distributeCustomCoupon(task *models.CouponTask, param interface{}) {
	var b []byte
	w := aws.NewWriteAtBuffer(b)
	gaws.AwsDownload2Stream(task.CouponListPath, w)
	list := tools.ParseTargetList(string(w.Bytes()))

	eventParam := coupon_event.ManualEventParam{}
	eventParam.CouponId = task.CouponId
	eventParam.List = list

	coupon_event.StartCouponEvent(coupon_event.TriggerManual, eventParam)
}

func backupHistoryCouponData() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	setKey := beego.AppConfig.String("coupon_set") + tools.MDateMHSDate(tools.GetUnixMillis()-tools.MILLSSECONDADAY)

	num, _ := redis.Int(storageClient.Do("EXISTS", setKey))
	if num == 0 {
		return
	}

	keyList, _ := redis.Strings(storageClient.Do("SMEMBERS", setKey))

	count := 0
	for _, v := range keyList {
		list := strings.Split(v, ":")
		if len(list) < 3 {
			continue
		}

		id, _ := tools.Str2Int64(list[1])
		if id == 0 {
			continue
		}

		pushDate := tools.GetDateParseBackend(list[2]) * 1000

		totalNum, _ := redis.Int(storageClient.Do("HGET", v, coupon_event.CouponKeyTotal))
		succNum, _ := redis.Int(storageClient.Do("HGET", v, coupon_event.CouponKeySucc))
		usedNum, _ := redis.Int(storageClient.Do("HGET", v, coupon_event.CouponKeyUsed))

		count := int(dao.GetCouponTotalNumInfo(id))

		record := models.CouponDetail{}
		record.CouponId = id
		record.UsedNum = usedNum
		record.TotalNum = totalNum
		record.SuccNum = succNum
		record.CouponDate = pushDate
		record.Ctime = tools.GetUnixMillis()
		if count > 0 {
			record.SuccRate = record.SuccNum * 100 / count
			record.UsedRate = record.UsedNum * 100 / count
		}
		record.Insert()

		storageClient.Do("DEL", v)

		count++
	}

	logs.Info("[backupHistoryCouponData] backup history data success key:%s, count:%d", setKey, count)

	storageClient.Do("DEL", setKey)
}

func StartCouponBackup() {
	lockKey := beego.AppConfig.String("coupon_backup_lock")

	for {
		storageClient := storage.RedisStorageClient.Get()
		lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", 10*60, "NX")

		if err != nil || lock == nil {
			storageClient.Close()
			time.After(time.Hour)
			continue
		}

		backupHistoryCouponData()

		storageClient.Do("DEL", lockKey)

		storageClient.Close()

		time.Sleep(time.Second)
	}
}

func distributeRegisterNoOrder(task *models.CouponTask, param interface{}) {
	accountList := make([]string, 0)

	startStr := tools.MDateMHSDate(tools.NaturalDay(-1))
	start := tools.GetDateParseBackend(startStr) * 1000
	end := tools.GetUnixMillis()

	maxId := int64(0)
	for {
		list, _ := dao.QueryRegisterNoOrderAccount(start, end, maxId)
		if len(list) == 0 {
			break
		}

		for _, v := range list {
			if v > maxId {
				maxId = v
			}

			accountList = append(accountList, tools.Int642Str(v))
		}
	}

	eventParam := coupon_event.ManualEventParam{}
	eventParam.CouponId = task.CouponId
	eventParam.List = accountList

	coupon_event.StartCouponEvent(coupon_event.TriggerManual, eventParam)
}

func distributeRegisterTmpOrder(task *models.CouponTask, param interface{}) {
	accountList := make([]string, 0)

	startStr := tools.MDateMHSDate(tools.NaturalDay(-1))
	start := tools.GetDateParseBackend(startStr) * 1000
	end := tools.GetUnixMillis()

	maxId := int64(0)
	for {
		list, _ := dao.QueryRegisterTmpOrderAccount(start, end, maxId)
		if len(list) == 0 {
			break
		}

		for _, v := range list {
			if v > maxId {
				maxId = v
			}

			accountList = append(accountList, tools.Int642Str(v))
		}
	}

	eventParam := coupon_event.ManualEventParam{}
	eventParam.CouponId = task.CouponId
	eventParam.List = accountList

	coupon_event.StartCouponEvent(coupon_event.TriggerManual, eventParam)
}

func distributeRepayClear(task *models.CouponTask, param interface{}) {
	accountList := make([]string, 0)

	startStr := tools.MDateMHSDate(tools.NaturalDay(-1))
	start := tools.GetDateParseBackend(startStr) * 1000
	end := tools.GetUnixMillis()

	maxId := int64(0)
	for {
		list, _ := dao.QueryRepayClearAccount(start, end, types.IsOverdueNo, maxId)
		if len(list) == 0 {
			break
		}

		for _, v := range list {
			if v > maxId {
				maxId = v
			}

			order, err := dao.AccountLastLoanOrder(v)
			if err != nil {
				continue
			}

			if order.CheckStatus != types.LoanStatusAlreadyCleared {
				continue
			}

			status := []string{
				tools.Int642Str(int64(types.LoanStatusRollClear)),
			}
			list, _, _ := dao.AccountHistory(v, status)
			if len(list) > 0 {
				continue
			}

			accountList = append(accountList, tools.Int642Str(v))
		}
	}

	eventParam := coupon_event.ManualEventParam{}
	eventParam.CouponId = task.CouponId
	eventParam.List = accountList

	coupon_event.StartCouponEvent(coupon_event.TriggerManual, eventParam)
}

func distributeRepayOverdue(task *models.CouponTask, param interface{}) {
	accountList := make([]string, 0)

	startStr := tools.MDateMHSDate(tools.NaturalDay(-1))
	start := tools.GetDateParseBackend(startStr) * 1000
	end := tools.GetUnixMillis()

	maxId := int64(0)
	for {
		list, _ := dao.QueryRepayClearAccount(start, end, types.IsOverdueYes, maxId)
		if len(list) == 0 {
			break
		}

		for _, v := range list {
			if v > maxId {
				maxId = v
			}

			order, err := dao.AccountLastLoanOrder(v)
			if err != nil {
				continue
			}

			account, err := models.OneAccountBaseByPkId(v)
			if err != nil {
				continue
			}

			if order.CheckStatus != types.LoanStatusAlreadyCleared {
				continue
			}

			yes, _ := models.IsBlacklistMobile(account.Mobile)
			if yes {
				continue
			}

			yes, _ = models.IsBlacklistIdentity(account.Identity)
			if yes {
				continue
			}

			status := []string{
				tools.Int642Str(int64(types.LoanStatusRollClear)),
			}
			list, _, _ := dao.AccountHistory(v, status)
			if len(list) > 0 {
				continue
			}

			accountList = append(accountList, tools.Int642Str(v))
		}
	}

	eventParam := coupon_event.ManualEventParam{}
	eventParam.CouponId = task.CouponId
	eventParam.List = accountList

	coupon_event.StartCouponEvent(coupon_event.TriggerManual, eventParam)
}
