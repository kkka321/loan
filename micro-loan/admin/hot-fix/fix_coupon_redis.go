package main

import (
	_ "micro-loan/common/lib/db/mysql"

	"fmt"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/types"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

func QueryConpon() (list []models.Coupon, err error) {
	o := orm.NewOrm()
	m := models.Coupon{}
	o.Using(m.UsingSlave())

	sql := "SELECT * FROM coupon"

	r := o.Raw(sql)
	_, err = r.QueryRows(&list)

	return
}

type CouponUsedInfo struct {
	CouponId int64
	Num      int64
}

func GetCouponNumInfo(couponId int64) (list []CouponUsedInfo, err error) {
	o := orm.NewOrm()
	m := models.AccountCoupon{}
	o.Using(m.UsingSlave())

	sql := ""
	if couponId == 0 {
		sql = fmt.Sprintf(`SELECT coupon_id, count(amount) as num FROM %s
group by coupon_id`,
			m.TableName())
	} else {
		sql = fmt.Sprintf(`SELECT coupon_id, count(amount) as num FROM %s
WHERE coupon_id = %d
group by coupon_id`,
			m.TableName(),
			couponId)
	}

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

func GetCouponUsedNumInfo(couponId int64) (list []CouponUsedInfo, err error) {
	o := orm.NewOrm()
	m := models.AccountCoupon{}
	o.Using(m.UsingSlave())

	sql := ""
	if couponId == 0 {
		sql = fmt.Sprintf(`SELECT coupon_id, count(amount) as num FROM %s
WHERE status = %d
group by coupon_id`,
			m.TableName(),
			types.CouponStatusUsed)
	} else {
		sql = fmt.Sprintf(`SELECT coupon_id, count(amount) as num FROM %s
WHERE coupon_id = %d AND status = %d
group by coupon_id`,
			m.TableName(),
			couponId,
			types.CouponStatusUsed)
	}

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

func GetCouponUsedAmountInfo(couponId int64) (list []CouponUsedInfo, err error) {
	o := orm.NewOrm()
	m := models.AccountCoupon{}
	o.Using(m.UsingSlave())

	sql := ""
	if couponId == 0 {
		sql = fmt.Sprintf(`SELECT coupon_id, sum(amount) as num FROM %s
WHERE status = %d
group by coupon_id`,
			m.TableName(),
			types.CouponStatusUsed)
	} else {
		sql = fmt.Sprintf(`SELECT coupon_id, sum(amount) as num FROM %s
WHERE coupon_id = %d AND status = %d
group by coupon_id`,
			m.TableName(),
			couponId,
			types.CouponStatusUsed)
	}

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

func main() {
	list, _ := QueryConpon()

	for {
		infos, err := GetCouponUsedNumInfo(0)
		if err != nil {
			break
		}

		tmpMap := make(map[int64]CouponUsedInfo)
		for _, v := range infos {
			tmpMap[v.CouponId] = v
		}

		for i, _ := range list {
			info, ok := tmpMap[list[i].Id]
			if !ok {
				continue
			}

			list[i].UsedNum = info.Num
		}

		break
	}

	for {
		infos, err := GetCouponUsedAmountInfo(0)
		if err != nil {
			break
		}

		tmpMap := make(map[int64]CouponUsedInfo)
		for _, v := range infos {
			tmpMap[v.CouponId] = v
		}

		for i, _ := range list {
			info, ok := tmpMap[list[i].Id]
			if !ok {
				continue
			}

			list[i].UsedAmount = info.Num
		}

		break
	}

	for {
		infos, err := GetCouponNumInfo(0)
		if err != nil {
			break
		}

		tmpMap := make(map[int64]CouponUsedInfo)
		for _, v := range infos {
			tmpMap[v.CouponId] = v
		}

		for i, _ := range list {
			info, ok := tmpMap[list[i].Id]
			if !ok {
				continue
			}

			list[i].DistributeAll = info.Num
		}

		break
	}

	totalKey := beego.AppConfig.String("coupon_total")
	amountKey := beego.AppConfig.String("coupon_amount")
	usedKey := beego.AppConfig.String("coupon_used")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	for i := 0; i < len(list); i++ {
		fmt.Sprintln("id:%d, total:%d, used:%d, amount:%d", list[i].Id, list[i].DistributeAll, list[i].UsedNum, list[i].UsedAmount)

		storageClient.Do("HSET", totalKey, list[i].Id, list[i].DistributeAll)
		storageClient.Do("HSET", usedKey, list[i].Id, list[i].UsedNum)
		storageClient.Do("HSET", amountKey, list[i].Id, list[i].UsedAmount)
	}
}
