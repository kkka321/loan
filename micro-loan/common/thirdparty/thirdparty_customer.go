package thirdparty

import (
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// 180627 01 015945 3982
func IsValiedId(value int64, targerType int) bool {
	checkNum := value / (1e17)
	if checkNum == 0 {
		return false
	}
	tmp := value / (1e10)

	return (tmp % 100) == int64(targerType)
}

func CustomerStatistic(e *evtypes.CustomerStatisticEv) (success bool, err error) {

	logs.Info("[CustomerStatistic] step into . e:%#v", e)
	defer logs.Info("[CustomerStatistic] step out")

	if nil == e {
		logs.Error("[CustomerStatistic] parm nil")
		return
	}

	if 0 == e.UserAccountId && 0 == e.OrderId {
		logs.Error("[CustomerStatistic] UserAccountId and OrderId both 0. e:%#v", e)
		return
	}

	if 0 == e.UserAccountId && 0 != e.OrderId {
		if !IsValiedId(e.OrderId, int(types.OrderSystem)) {
			account, err := models.OneAccountBaseByMobile(tools.Int642Str(e.OrderId))
			if err != nil {
				logs.Warn("[CustomerStatistic] inValied order id customer may not finish regist. e:%#v", e)
				return false, nil
			}
			e.UserAccountId = account.Id
		} else {
			order, _ := models.GetOrder(e.OrderId)
			e.UserAccountId = order.UserAccountId
		}
	}
	// 1- 准备数据
	accountBase, _ := models.OneAccountBaseByPkId(e.UserAccountId)
	thirdpartyInfo, err := models.GetThirdpartyInfoByApiMd5(e.ApiMd5)
	if err != nil || "" == thirdpartyInfo.ApiMd5 {
		logs.Error("[CustomerStatistic] GetThirdpartyInfoByApiMd5 err:%s e:%#v thirdpartyInfo:%#v", err, e, thirdpartyInfo)
		return
	}

	// 2-在redis设置锁
	// + 分布式锁
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()
	uId := tools.Int642Str(e.UserAccountId)
	lockKey := "thirdparty_customer_lock:" + uId
	tryTimes := 0

LOCK:
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", 60, "NX")
	if err != nil || lock == nil {
		if err != nil {
			logs.Error("[CustomerStatistic] fatal error may redis out of service.tryTimes:%d err:%v lock:%v lockKey:%s unhandle data e:%#v", tryTimes, err, lock, lockKey, e)
			return
		}

		logs.Info("[CustomerStatistic] process is working, so, I sleep to retry.  tryTimes :", tryTimes)
		time.Sleep(1 * time.Second)
		if tryTimes > 10 {
			logs.Warn("[CustomerStatistic] process is working, I do not want sleep because i sleep tryTimes :", tryTimes)
			storageClient.Do("DEL", lockKey)
		}

		if tryTimes > 20 {
			//讲道理不会执行下边的条件
			logs.Error("[CustomerStatistic] fatal error. I can not believe it. tryTimes:%d err:%v lock:%v  lockKey:%s unhandle data e:%#v", tryTimes, err, lock, lockKey, e)
			return
		}

		tryTimes++
		goto LOCK
	}
	// 4-解锁
	defer storageClient.Do("DEL", lockKey)

	// 3-读取数据库数据
	customerTotal, _ := models.GetThirdpartyStatisticCustomerByApiMd5AndUId("", e.UserAccountId, types.RecordTypeTotal)
	customerApi, _ := models.GetThirdpartyStatisticCustomerByApiMd5AndUId(e.ApiMd5, e.UserAccountId, types.RecordTypeSingle)
	newTotal := (customerTotal.Id == 0)
	newApi := (customerApi.Id == 0)

	// 数据库无此条记录说明是新数据需要创建
	if newTotal {
		customerTotal = models.ThirdpartyStatisticCustomer{
			UserAccountId: e.UserAccountId,
			OrderId:       e.OrderId,
			Mobile:        accountBase.Mobile,
			// MediaSource:   appsflyer.MediaSource,
			// Campaign:      appsflyer.Campaign,
			RecordType: types.RecordTypeTotal,
			Ctime:      tools.GetUnixMillis(),
		}
	}
	// 数据库无此条记录说明是新数据需要创建
	if newApi {
		customerApi = models.ThirdpartyStatisticCustomer{
			UserAccountId: e.UserAccountId,
			OrderId:       e.OrderId,
			Mobile:        accountBase.Mobile,
			// MediaSource:   appsflyer.MediaSource,
			// Campaign:      appsflyer.Campaign,
			Api:        thirdpartyInfo.Api,
			ApiMd5:     thirdpartyInfo.ApiMd5,
			RecordType: types.RecordTypeSingle,
			Ctime:      tools.GetUnixMillis(),
		}
	}

	// 3-写入数据库
	// 是短信且不是第一条记录 才去统计。如果时第一条记录直接执行下边的捞回就好啦，防止多次计算
	if false == e.MessageFlag {
		customerTotal.CutomerTotalCost += e.Fee
		customerTotal.CallCount++
		customerTotal.SuccessCallCount += tools.ThreeElementExpression(IsApiResultSuccess(e.Result), 1, 0).(int)
		customerTotal.HitCallCount += tools.ThreeElementExpression(IsApiResultHit(e.Result), 1, 0).(int)
		customerTotal.Utime = tools.GetUnixMillis()

		customerApi.ApiFee += e.Fee
		customerApi.CallCount++
		customerApi.SuccessCallCount += tools.ThreeElementExpression(IsApiResultSuccess(e.Result), 1, 0).(int)
		customerApi.HitCallCount += tools.ThreeElementExpression(IsApiResultHit(e.Result), 1, 0).(int)
		customerApi.Utime = tools.GetUnixMillis()

		err = dao.AddOrUpdateThirdpartyStatisticCustomer(&customerApi, &customerTotal)
		if err != nil {
			logs.Error("[CustomerStatistic]AddOrUpdateThirdpartyStatisticCustomer err:%s  e:%#v customerTotal:%#v customerApi:%#v", err, e, customerTotal, customerApi)
		}

	}

	logs.Info("newTotal:", newTotal, " newApi:", newApi)
	// 捞回当初注册时的花费, 以及登录时的花费
	// 如果此时 id为0 从新查下数据库 防止 上边已插入记录
	if 0 == customerTotal.Id {
		customerTmp, _ := models.GetThirdpartyStatisticCustomerByApiMd5AndUId("", e.UserAccountId, types.RecordTypeTotal)

		// 数据库有记录
		if 0 != customerTmp.Id {
			customerTotal = customerTmp
		}
	}
	err = SalvageAuthCodeSmsData(accountBase.Id, accountBase.Mobile, &customerTotal)

	// 5- 结束
	if err != nil {
		logs.Error("[CustomerStatistic] SalvageAuthCodeSmsData err:%s  e:%#v customerTotal:%#v customerApi:%#v", err, e, customerTotal, customerApi)
		return false, err
	}

	return true, nil
}

func SalvageAuthCodeSmsData(uId int64, mobile string,
	customerTotal *models.ThirdpartyStatisticCustomer) (err error) {

	iMobile, _ := tools.Str2Int64(strings.TrimSpace(mobile))
	if 0 == iMobile {
		logs.Error("[SalvageAuthCodeSmsData] parm err. uId:%d mobile:%s", uId, mobile)
		err = fmt.Errorf("[SalvageAuthCodeSmsData] parm err. uId:%d mobile:%s", uId, mobile)
		return
	}

	// types.RecordTypeTotal 类型的数据 ApiFee 字段用来存储已统计的短信的 最后id
	list, _ := models.GetAllByRelatedIdAndLastId(iMobile, customerTotal.ApiFee)
	if len(list) == 0 {
		logs.Info("[SalvageAuthCodeSmsData] no record in thirdparty_record iMobile:", iMobile)
		// err = fmt.Errorf("[SalvageAuthCodeSmsData] no record in thirdparty_record iMobile:", iMobile)
		return
	}

	cAPI, _ := models.GetThirdpartyStatisticCustomerByApiMd5AndUId(tools.Md5(list[0].Api), uId, types.RecordTypeSingle)
	customerTotal.ApiFee = list[0].Id
	newAPI := (cAPI.Id == 0)
	if newAPI {
		cAPI.UserAccountId = customerTotal.UserAccountId
		cAPI.Mobile = customerTotal.Mobile
		cAPI.Api = list[0].Api
		cAPI.ApiMd5 = tools.Md5(list[0].Api)
		// cAPI.MediaSource = customerTotal.MediaSource
		// cAPI.Campaign = customerTotal.Campaign
		cAPI.RecordType = types.RecordTypeSingle
		cAPI.Ctime = tools.GetUnixMillis()
	}

	for _, one := range list {

		cAPI.ApiFee += int64(one.FeeForCall)
		cAPI.CallCount++
		cAPI.SuccessCallCount += tools.ThreeElementExpression(IsApiResultSuccess(one.ResponseType), 1, 0).(int)
		cAPI.HitCallCount += tools.ThreeElementExpression(IsApiResultHit(one.ResponseType), 1, 0).(int)

		customerTotal.CutomerTotalCost += int64(one.FeeForCall)
		customerTotal.CallCount++
		customerTotal.SuccessCallCount += tools.ThreeElementExpression(IsApiResultSuccess(one.ResponseType), 1, 0).(int)
		customerTotal.HitCallCount += tools.ThreeElementExpression(IsApiResultHit(one.ResponseType), 1, 0).(int)
	}
	cAPI.Utime = tools.GetUnixMillis()
	customerTotal.Utime = tools.GetUnixMillis()

	err = dao.AddOrUpdateThirdpartyStatisticCustomer(&cAPI, customerTotal)
	if err != nil {
		logs.Error("[SalvageAuthCodeSmsData] AddOrUpdateThirdpartyStatisticCustomer err:%s  customerTotal:%#v cAPI:%#v", err, customerTotal, cAPI)
	}

	return err
}
