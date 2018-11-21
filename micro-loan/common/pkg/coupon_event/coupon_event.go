package coupon_event

import (
	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"

	"micro-loan/common/tools"
	"micro-loan/common/types"

	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"
)

const (
	CouponKeyTotal string = "total"
	CouponKeyUsed  string = "used"
	CouponKeySucc  string = "succ"
)

type CouponEventTrigger int

const (
	TriggerManual          CouponEventTrigger = 1
	TriggerRegister        CouponEventTrigger = 2
	TriggerLogin           CouponEventTrigger = 3
	TriggerCreateOrder     CouponEventTrigger = 4
	TriggerConfirmOrder    CouponEventTrigger = 5
	TriggerDisburse        CouponEventTrigger = 6
	TriggerRepay           CouponEventTrigger = 7
	TriggerClear           CouponEventTrigger = 8
	TriggerTimer           CouponEventTrigger = 9
	TriggerWebRegister     CouponEventTrigger = 10
	TriggerWeb1018Register CouponEventTrigger = 11
	TriggerInviteV3        CouponEventTrigger = 12
)

type CouponEvent interface {
	HandleEvent(trigger CouponEventTrigger, data interface{})
}

var couponEventList = make([]CouponEvent, 0)

func init() {
	couponEventList = append(couponEventList, new(ManualEvent))
	//couponEventList = append(couponEventList, new(RegisterEvent))
	couponEventList = append(couponEventList, new(InviteEvent))
	//couponEventList = append(couponEventList, new(Invite1018Event))
	couponEventList = append(couponEventList, new(InviteV3Event))
}

func StartCouponEvent(trigger CouponEventTrigger, data interface{}) {
	for _, v := range couponEventList {
		v.HandleEvent(trigger, data)
	}
}

func distributeCoupon(userType string, accountId int64) []int64 {
	ids := make([]int64, 0)

	list, err := dao.GetCouponByUserType(userType)
	if err != nil {
		logs.Error("[distributeCoupon] GetCouponByKey error err:%v", err)
		return ids
	}

	_, err = models.OneAccountBaseByPkId(accountId)
	if err != nil {
		logs.Error("[distributeCoupon] OneAccountBaseByPkId error accountId:%d, err:%v", accountId, err)
		return ids
	}

	totalKey := beego.AppConfig.String("coupon_total")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	for _, v := range list {
		if !IsCouponIsPushing(&v) {
			logs.Warn("[distributeCoupon] coupon out of date couponId:%d", v.Id)
			continue
		}

		num := dao.GetCouponTotalNumInfo(v.Id)
		if v.DistributeSize != 0 && num >= v.DistributeSize {
			logs.Error("[distributeCoupon] coupon distribute max couponId:%d num:%d, max_num:%d", v.Id, num, v.DistributeSize)
			continue
		}

		num++

		id, _ := AddAccountCoupon(accountId, &v)

		if id > 0 {
			ids = append(ids, id)
			storageClient.Do("HINCRBY", totalKey, v.Id, 1)

			IncrCouponDailyTotal(v.Id, 1)
		}
	}

	return ids
}

func IsCouponIsPushing(coupon *models.Coupon) bool {
	timetag := tools.GetUnixMillis()
	if timetag > coupon.DistributeStart && timetag < coupon.DistributeEnd && coupon.IsAvailable > 0 {
		return true
	}

	return false
}

func IncrCouponDailyTotal(id int64, num int) {
	incrCouponDailyNum(CouponKeyTotal, id, num)
}

func IncrCouponDailyUsed(id int64, num int) {
	incrCouponDailyNum(CouponKeyUsed, id, num)
}

func IncrCouponDailySucc(id int64, num int) {
	incrCouponDailyNum(CouponKeySucc, id, num)
}

func incrCouponDailyNum(numType string, id int64, num int) {
	nowStr := tools.MDateMHSDate(tools.GetUnixMillis())
	key := fmt.Sprintf("coupon:%d:%s", id, nowStr)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	exist, _ := redis.Int(storageClient.Do("HSETNX", key, numType, 0))
	if exist > 0 {
		setKey := beego.AppConfig.String("coupon_set") + nowStr
		storageClient.Do("SADD", setKey, key)
	}

	storageClient.Do("HINCRBY", key, numType, num)
}

func AddAccountCoupon(accountId int64, coupon *models.Coupon) (int64, error) {
	accountCoupon := models.AccountCoupon{}
	accountCoupon.Status = types.CouponStatusAvailable
	accountCoupon.UserAccountId = accountId
	accountCoupon.CouponId = coupon.Id
	accountCoupon.IsNew = types.CouponUnread
	accountCoupon.Ctime = tools.GetUnixMillis()
	if coupon.ValidStart > 0 {
		accountCoupon.ValidStart = coupon.ValidStart
	} else {
		validStart := tools.MDateMHSDate(tools.GetUnixMillis())
		timeStart := tools.GetDateParseBackend(validStart) * 1000
		accountCoupon.ValidStart = timeStart
	}
	if coupon.ValidEnd > 0 {
		accountCoupon.ValidEnd = coupon.ValidEnd
	} else {
		accountCoupon.ValidEnd = accountCoupon.ValidStart + int64(3600*(coupon.ValidDays+1)*24*1000) - 1000
	}

	id, err := dao.AddAccountCoupon(&accountCoupon)

	return id, err
}
