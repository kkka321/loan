package task

import (
	"sync"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func RunCouponExpireTask() error {
	logs.Info("[RunCouponExpireTask] start launch.")

	TimerWg.Add(1)
	defer TimerWg.Done()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("coupon_expire_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[RunCouponExpireTask] process is working, so, I will exit.")
		return nil
	}

	for {
		if cancelled() {
			logs.Info("[RunCouponExpireTask] receive exit cmd.")
			break
		}

		nowStr := tools.MDateMHSDate(tools.GetUnixMillis())
		nowZero := tools.GetDateParseBackend(nowStr) * 1000

		list, err := dao.QueryExpireAccountCoupon(nowZero)
		if err != nil {
			logs.Info("[RunCouponExpireTask] QueryExpireCoupon error:%v", err)
			break
		}

		if len(list) == 0 {
			break
		}

		for _, v := range list {
			consumeCoupon(&v)
		}

	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[RunCouponExpireTask] politeness exit.")

	return nil
}

// 消费逾期订单队列
func consumeCoupon(accountCoupon *models.AccountCoupon) {
	service.MakeAccountCouponInvalid(accountCoupon)
}

func RunAccountCouponReuseTask() error {
	logs.Info("[RunAccountCouponReuseTask] start launch.")

	TimerWg.Add(1)
	defer TimerWg.Done()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("account_coupon_reuse_lock")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[RunAccountCouponReuseTask] process is working, so, I will exit.")
		return nil
	}

	for {
		if cancelled() {
			logs.Info("[RunAccountCouponReuseTask] receive exit cmd.")
			break
		}

		// 生产队列,小批量处理
		qName := beego.AppConfig.String("account_coupon_reuse")
		qVal, err := storageClient.Do("LLEN", qName)
		if err == nil && qVal != nil && 0 == qVal.(int64) {
			logs.Info("[RunAccountCouponReuseTask] %s 队列为空,开始按条件生成.", qName)

			nowStr := tools.MDateMHSDate(tools.GetUnixMillis())
			nowZero := tools.GetDateParseBackend(nowStr) * 1000

			localStart := nowZero - 2*tools.MILLSSECONDADAY
			localEnd := nowZero - 1*tools.MILLSSECONDADAY

			list, _ := dao.QueryRejectOrderCoupon(localStart, localEnd, 100)

			// 如果没有满足条件的数据,work goroutine 也不用启动了
			if len(list) == 0 {
				break
			}

			for _, l := range list {
				storageClient.Do("LPUSH", qName, l.Id)
			}

			// 消费队列
			var wg sync.WaitGroup
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go consumeAccountCouponQueue(&wg, i)
			}

			// 主 goroutine,等待工作 goroutine 正常结束
			wg.Wait()
		}
	}

	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[RunAccountCouponReuseTask] politeness exit.")

	return nil
}

func consumeAccountCouponQueue(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	logs.Info("It will do consumeAccountCouponQueue, workerID:", workerID)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	qName := beego.AppConfig.String("account_coupon_reuse")
	for {
		if cancelled() {
			break
		}

		qValueByte, err := storageClient.Do("RPOP", qName)
		// 没有可供消费的数据,退出工作 goroutine
		if err != nil || qValueByte == nil {
			logs.Info("[consumeAccountCouponQueue] no data for consume, I will exit, workID:", workerID)
			break
		}

		id, _ := tools.Str2Int64(string(qValueByte.([]byte)))

		// 真正开始工作了
		addCurrentData(tools.Int642Str(id), "accountCouponId")
		handleAccountCoupon(id, workerID)
		removeCurrentData(tools.Int642Str(id))
	}
}

func handleAccountCoupon(id int64, workerID int) {
	logs.Info("[handleAccountCoupon] orderID:", id, ", workerID:", workerID)

	defer func() {
		if x := recover(); x != nil {
			logs.Error("[handleAccountCoupon] panic accountCouponId:%d, workId:%d, err:%v", id, workerID, x)
			logs.Error(tools.FullStack())
		}
	}()

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	accountCoupon, err := dao.GetAccountCouponById(id)
	if err != nil {
		logs.Warn("[handleAccountCoupon] GetAccountCouponById error id:%d, workId:%d, err:%v", id, workerID, err)
		return
	}

	if accountCoupon.Status != types.CouponStatusFrozen {
		return
	}

	_, err = dao.GetCouponById(accountCoupon.CouponId)
	if err != nil {
		logs.Warn("[handleAccountCoupon] GetCouponById error id:%d, workId:%d, err:%v", id, workerID, err)
		service.MakeAccountCouponInvalid(&accountCoupon)
		return
	}

	if accountCoupon.ValidEnd <= tools.GetUnixMillis() {
		service.MakeAccountCouponInvalid(&accountCoupon)
		return
	}

	service.MakeAccountCouponAvailable(&accountCoupon)
}
