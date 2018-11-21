package thirdparty

import (
	"time"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

//***********************BusinessDetailStatistic 资金对账****************************************/
// 传入的时间会被强制格式化为前一天的 0：0：0 时间 如：传入 2018-03-22 的时间戳 会统计 2018-03-21 的数据
// 统一按照印尼的时间统计
func BusinessDetailStatistic(st int64) {
	logs.Debug("before startTime:%d ", st)

	// 获得印尼日期
	date := tools.MDateMHSDate(st - tools.MILLSSECONDADAY)

	//获得印尼日期 的时间戳
	st = tools.GetDateParseBackend(date) * 1000

	logs.Debug("after startTime:%d date:%s", st, date)

	// 2-在redis设置锁 防止统计时出现充值或提现操作
	// + 分布式锁
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := "lock:business_detail:" + tools.MDateMHSDate(st)
	tryTimes := 0
LOCK:
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", 60, "NX")
	if err != nil || lock == nil {
		if err != nil {
			logs.Error("[BusinessDetailStatistic] fatal error may redis out of service.tryTimes:%d err:%v lock:%v lockKey:%s unhandle data st:%#v", tryTimes, err, lock, lockKey, st)
			return
		}

		logs.Info("[BusinessDetailStatistic] process is working, so, I sleep to retry.  tryTimes :", tryTimes)
		time.Sleep(1 * time.Second)
		if tryTimes > 30 {
			logs.Warn("[BusinessDetailStatistic] process is working, so, but I can not wait.  tryTimes :", tryTimes)
			storageClient.Do("DEL", lockKey)
		}

		if tryTimes > 40 {
			//讲道理不会执行下边的条件
			logs.Error("[CustomerStatistic] fatal error. I can not believe it. tryTimes:%d err:%v lock:%v  lockKey:%s unhandle data st:%#v", tryTimes, err, lock, lockKey, st)
			return
		}
		tryTimes++
		goto LOCK
	}
	defer storageClient.Do("DEL", lockKey)
	// 分别统计各个第三方的数据
	thirdpartyStatistic(st)

	// 统计汇总数据
	totalStatistic(st)
}

func thirdpartyStatistic(startTime int64) {
	endTime := startTime + tools.MILLSSECONDADAY
	nameList, _ := dao.PaymentThirdpartyList()
	_ = types.Xendit
	for _, name := range nameList {
		if thirdparty, ok := types.ThirdpartyNameCodeMap[name]; ok {
			// id 为0 数据库无记录
			yesterday, _ := dao.OneByDateAndNameLastRecord(startTime, name)
			single, _ := dao.OneByDateAndName(startTime, name)
			if 0 == single.Id {
				single = models.BusinessDetail{
					PaymentName: name,
					RecordDate:  startTime,
					RecordDateS: tools.MDateMHSDate(startTime),
					RecordType:  types.RecordTypeSingle,
					Ctime:       tools.GetUnixMillis(),
				}
			}
			logs.Debug("single:%#v", single)
			//放款出账
			outAmount, err := dao.StatisticAmount(startTime, endTime, thirdparty.Code, types.PayTypeMoneyOut)
			if err != nil && err.Error() != types.EmptyOrmStr {
				logs.Warn("[thirdpartyStatistic] Statistic out Amount err:%s, st:%d et:%d code:%d ", err, startTime, endTime, thirdparty.Code)
			}
			single.PayOutAmount = outAmount

			//还款入账
			inAmount, err := dao.StatisticAmount(startTime, endTime, thirdparty.Code, types.PayTypeMoneyIn)
			if err != nil && err.Error() != types.EmptyOrmStr {
				logs.Warn("[thirdpartyStatistic] Statistic int Amount err:%s, st:%d et:%d code:%d ", err, startTime, endTime, thirdparty.Code)
			}
			single.PayInAmount = inAmount

			//放款手续费支出
			single.PayOutForFee = 0
			for _, outApi := range thirdparty.PayOutApiS {
				outFee, err := dao.StatisticFee(startTime, endTime, tools.Md5(outApi))
				if err != nil && err.Error() != types.EmptyOrmStr {
					logs.Warn("[thirdpartyStatistic] Statistic out Fee err:%s, st:%d et:%d api:%s ", err, startTime, endTime, outApi)
				}
				single.PayOutForFee += outFee
			}

			//还款款手续费支出
			single.PayInForFee = 0
			for _, inApi := range thirdparty.PayInApiS {
				inFee, err := dao.StatisticFee(startTime, endTime, tools.Md5(inApi))
				if err != nil && err.Error() != types.EmptyOrmStr {
					logs.Warn("[thirdpartyStatistic] Statistic in Fee err:%s, st:%d et:%d api:%s ", err, startTime, endTime, inApi)
				}
				single.PayInForFee += inFee
			}

			//三方账户余额=上一日三方账户余额+充值金额+还款入账-提现金额-放款出账-放款手续费支出-还款手续费支出
			single.AccountBalance = yesterday.AccountBalance + single.RechargeAmount + single.PayInAmount -
				single.WithdrawAmount - single.PayOutAmount - single.PayOutForFee - single.PayInForFee

			// update time
			single.Utime = tools.GetUnixMillis()

			// 更新数据库
			logs.Debug("single:%#v yesterday:%#v", single, yesterday)
			err = dao.AddOrUpdateBusinessDetailSingle(&single,
				"pay_out_amount",
				"pay_in_amount",
				"pay_out_for_fee",
				"pay_in_for_fee",
				"account_balance",
				"utime")
			if err != nil {
				logs.Error("[thirdpartyStatistic] AddOrUpdateBusinessDetailSingle err:%s, st:%d code:%d  single:%#v", err, startTime, thirdparty.Code, single)
			}
		} else {
			logs.Error("[thirdpartyStatistic] name not in ThirdpartyNameCodeMap. name:%s ThirdpartyNameCodeMap:%#v", name, types.ThirdpartyNameCodeMap)
		}
	}
}

func totalStatistic(startTime int64) {
	total := models.BusinessDetail{
		PaymentName: types.RecordTypeTotalName,
		RecordDate:  startTime,
		RecordDateS: tools.MDateMHSDate(startTime),
		RecordType:  types.RecordTypeTotal,
	}
	totalFromDB, _ := dao.OneByDateAndName(startTime, types.RecordTypeTotalName)

	// 为了保证幂等 如果有数据就清除
	total.Id = totalFromDB.Id
	total.RechargeAmount = totalFromDB.RechargeAmount
	total.WithdrawAmount = totalFromDB.WithdrawAmount
	total.Ctime = tools.ThreeElementExpression(totalFromDB.Id == 0, tools.GetUnixMillis(), totalFromDB.Ctime).(int64)
	total.Utime = tools.GetUnixMillis()
	// 读出单条数据
	logs.Debug("total:%#v", total)
	list, err := dao.BusinessDetailSingleList(startTime, types.RecordTypeSingle)
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[totalStatistic] BusinessDetailSingleList err:%s st:%d", err, startTime)
	}

	for _, one := range list {
		total.PayOutAmount += one.PayOutAmount
		total.PayOutForFee += one.PayOutForFee
		total.PayInAmount += one.PayInAmount
		total.PayInForFee += one.PayInForFee
		total.AccountBalance += one.AccountBalance
	}

	//在贷余额=所有客户的（应还本金-已还本金-减免本金）之和
	lendingBalance, err := dao.BusinessDetailLendingBalance()
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[totalStatistic] BusinessDetailLendingBalance err:%s st:%d", err, startTime)
	}
	total.LendingBalance = lendingBalance

	//服务费收入
	fee, err := dao.BusinessDetailInterestIncome(startTime, startTime+tools.MILLSSECONDADAY, "service_fee")
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[totalStatistic] BusinessDetailInterestIncome fee  err:%s current:%s", err, startTime)
	}
	total.FeeIncome = fee

	//利息收入
	interest, err := dao.BusinessDetailInterestIncome(startTime, startTime+tools.MILLSSECONDADAY, "pre_interest")
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[totalStatistic] BusinessDetailInterestIncome interest  err:%s current:%s", err, startTime)
	}
	total.InterestIncome = interest

	//宽限期利息收入
	grace, err := dao.BusinessDetailInterestIncome(startTime, startTime+tools.MILLSSECONDADAY, "grace_period_interest")
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[totalStatistic] BusinessDetailInterestIncome grace err:%s current:%s", err, startTime)
	}
	total.GraceInterestIncome = grace

	//罚息收入
	penalty, err := dao.BusinessDetailInterestIncome(startTime, startTime+tools.MILLSSECONDADAY, "penalty")
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[totalStatistic] BusinessDetailInterestIncome  penalty err:%s current:%s", err, startTime)
	}
	total.PenaltyInterestIncome = penalty

	// 更新数据库
	err = dao.AddOrUpdateBusinessDetailSingle(&total,
		"pay_out_amount",
		"pay_in_amount",
		"pay_out_for_fee",
		"pay_in_for_fee",
		"account_balance",
		"lending_balance",
		"fee_income",
		"interest_income",
		"grace_interest_income",
		"penalty_interest_income",
		"utime")
	if err != nil {
		logs.Error("[totalStatistic] AddOrUpdateBusinessDetailSingle err:%s, st:%d total:%#v", err, startTime, total)
	}
}
