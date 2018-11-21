package coupon_event

import (
	"strings"
	"sync"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
)

type ManualEvent struct {
}

type ManualEventParam struct {
	CouponId int64
	List     []string
}

func (c *ManualEvent) HandleEvent(trigger CouponEventTrigger, data interface{}) {
	if trigger != TriggerManual {
		return
	}

	if data == nil {
		return
	}

	param, ok := data.(ManualEventParam)
	if !ok || len(param.List) == 0 {
		return
	}

	coupon, err := dao.GetCouponById(param.CouponId)
	if err != nil {
		logs.Error("[ManualEvent] GetCouponById error err:%v", err)
		return
	}

	if !IsCouponIsPushing(&coupon) {
		return
	}

	count := 10
	var wg sync.WaitGroup
	preSize := len(param.List)/count + 1

	for i := 0; i < count; i++ {
		startIndex := i * preSize
		endIndex := startIndex + preSize

		if startIndex >= len(param.List) {
			break
		}

		if endIndex > len(param.List) {
			endIndex = len(param.List)
		}

		wg.Add(1)
		go batchAddCoupon(&wg, param.List[startIndex:endIndex], &coupon)
	}

	wg.Wait()
}

func batchAddCoupon(wg *sync.WaitGroup, list []string, coupon *models.Coupon) {
	defer wg.Done()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	totalKey := beego.AppConfig.String("coupon_total")

	count := 0
	for _, v := range list {
		str := strings.Trim(v, " ")
		str = strings.Trim(str, "\r")
		if str == "" {
			continue
		}

		accountId, _ := tools.Str2Int64(str)
		_, err := models.OneAccountBaseByPkId(accountId)
		if err != nil {
			logs.Error("[batchAddCoupon] OneAccountBaseByPkId error accountId:%d, err:%v", accountId, err)
			continue
		}

		AddAccountCoupon(accountId, coupon)

		count++
	}

	storageClient.Do("HINCRBY", totalKey, coupon.Id, count)

	IncrCouponDailyTotal(coupon.Id, count)
}
