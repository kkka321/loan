package service

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func BusinessDetailList(condCntr map[string]interface{}, page, pagesize int) (list []models.BusinessDetail, count int64, err error) {
	obj := models.BusinessDetail{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	qs := o.QueryTable(obj.TableName())
	cond := orm.NewCondition()

	if value, ok := condCntr["register_time_start"]; ok {
		cond = cond.And("record_date__gte", value)
	}
	if value, ok := condCntr["register_time_end"]; ok {
		cond = cond.And("record_date__lt", value)
	}
	cond = cond.And("record_type", types.RecordTypeTotal)
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	count, _ = qs.SetCond(cond).Count()
	_, err = qs.SetCond(cond).OrderBy("-id").Limit(pagesize).Offset(offset).All(&list)

	return
}

func DoSaveRecharge(date string, paymentName string, amount int64) (err error) {

	// 1、检查日期合法性
	rechargeDate := tools.GetDateParseBackend(date) * 1000
	if rechargeDate <= 0 {
		err = fmt.Errorf("[service.DoSaveRecharge] date err. date :%s", date)
		return
	}

	// 2-在redis设置锁 防止并发
	// + 分布式锁
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := "lock:business_detail:" + date
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", 60, "NX")
	if err != nil || lock == nil {
		err := fmt.Errorf("[DoSaveRecharge] redis lock. err:%v lock:%v lockKey:%s", err, lock, lockKey)
		logs.Warn(err)
		return err
	}
	defer storageClient.Do("DEL", lockKey)

	//3.更新数据库
	recordSingle, _ := dao.OneBusinessDetailByDateAndName(rechargeDate, paymentName)
	recordTotal, _ := dao.OneBusinessDetailByDateAndName(rechargeDate, types.RecordTypeTotalName)

	if 0 == recordSingle.Id {
		recordSingle.RecordDate = rechargeDate
		recordSingle.RecordDateS = date
		recordSingle.PaymentName = paymentName
		recordSingle.RecordType = types.RecordTypeSingle
		recordSingle.Ctime = tools.GetUnixMillis()
	}
	recordSingle.Utime = tools.GetUnixMillis()
	recordSingle.RechargeAmount += amount
	recordSingle.AccountBalance += amount

	if 0 == recordTotal.Id {
		recordTotal.RecordDate = rechargeDate
		recordTotal.RecordDateS = date
		recordTotal.PaymentName = types.RecordTypeTotalName
		recordTotal.RecordType = types.RecordTypeTotal
		recordTotal.Ctime = tools.GetUnixMillis()
	}
	recordTotal.Utime = tools.GetUnixMillis()
	recordTotal.RechargeAmount += amount
	recordTotal.AccountBalance += amount
	return dao.AddOrUpdateBusinessDetail(&recordSingle, &recordTotal)
}

func DoSaveWithdraw(date string, paymentName string, amount int64) (err error) {

	// 1、检查日期合法性
	rechargeDate := tools.GetDateParseBackend(date) * 1000
	if rechargeDate <= 0 {
		err = fmt.Errorf("[service.DoSaveRecharge] date err. date :%s", date)
		return
	}

	// 2-在redis设置锁 防止并发
	// + 分布式锁
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := "lock:business_detail:" + date
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "EX", 60, "NX")
	if err != nil || lock == nil {
		err := fmt.Errorf("[DoSaveWithdraw] redis lock. err:%v lock:%v lockKey:%s", err, lock, lockKey)
		logs.Warn(err)
		return err
	}
	defer storageClient.Do("DEL", lockKey)

	//3.更新数据库
	recordSingle, _ := dao.OneBusinessDetailByDateAndName(rechargeDate, paymentName)
	recordTotal, _ := dao.OneBusinessDetailByDateAndName(rechargeDate, types.RecordTypeTotalName)

	if 0 == recordSingle.Id {
		recordSingle.RecordDate = rechargeDate
		recordSingle.RecordDateS = date
		recordSingle.PaymentName = paymentName
		recordSingle.RecordType = types.RecordTypeSingle
		recordSingle.Ctime = tools.GetUnixMillis()
	}
	recordSingle.Utime = tools.GetUnixMillis()
	recordSingle.WithdrawAmount += amount
	recordSingle.AccountBalance -= amount

	if 0 == recordTotal.Id {
		recordTotal.RecordDate = rechargeDate
		recordTotal.RecordDateS = date
		recordTotal.PaymentName = types.RecordTypeTotalName
		recordTotal.RecordType = types.RecordTypeTotal
		recordTotal.Ctime = tools.GetUnixMillis()
	}
	recordTotal.Utime = tools.GetUnixMillis()
	recordTotal.WithdrawAmount += amount
	recordTotal.AccountBalance -= amount
	return dao.AddOrUpdateBusinessDetail(&recordSingle, &recordTotal)
}

func ListRecordByDate(date string) (list []models.BusinessDetail, err error) {

	// 1、检查日期合法性
	recordDate := tools.GetDateParseBackend(date) * 1000
	if recordDate <= 0 {
		err = fmt.Errorf("[service.DoSaveRecharge] date err. date :%s", date)
		return
	}
	detail := models.BusinessDetail{}

	o := orm.NewOrm()
	o.Using(detail.UsingSlave())

	_, err = o.QueryTable(detail.TableName()).
		Filter("record_date", recordDate).
		Filter("record_type", types.RecordTypeSingle).
		OrderBy("payment_name").
		All(&list)
	return
}
