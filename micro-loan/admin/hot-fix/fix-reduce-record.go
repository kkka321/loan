package main

import (
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	// 数据库初始化
	"github.com/astaxie/beego/orm"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/device"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"reflect"
	"sort"
)

func QueryAllReduce() (list []models.ReduceRecord) {
	reduce := models.ReduceRecord{}
	o := orm.NewOrm()
	o.Using(reduce.UsingSlave())

	_, err := o.QueryTable(reduce.TableName()).All(&list)
	if err != nil {
		logs.Error("[QueryAllReduce] err:%v", err)
		return
	}

	return
}
func QueryAllPreReduce() (list []models.AdminPrereduced) {
	reduce := models.AdminPrereduced{}
	o := orm.NewOrm()
	o.Using(reduce.UsingSlave())

	_, err := o.QueryTable(reduce.TableName()).All(&list)
	if err != nil {
		logs.Error("[QueryAllPreReduce] err:%v", err)
		return
	}

	return
}

func sortRecord(list []models.ReduceRecord, listPre []models.AdminPrereduced) (result map[int64]interface{}) {
	//keys := []int64
	result = make(map[int64]interface{})
	for _, v := range list {
		result[v.Ctime] = v
	}

	for _, v := range listPre {
		result[v.Ctime] = v
	}

	return result
}

func InsertRecord(result map[int64]interface{}, key int64) error {
	if v, ok := result[key]; ok {
		vType := reflect.TypeOf(v)
		logs.Info("key [%d] v:%#v", key, v)
		logs.Info("vType:%s", vType.Name())
		switch vType.Name() {
		case "ReduceRecord":
			{
				if vR, ok := v.(models.ReduceRecord); ok {
					insertReduceRecord(vR)
				} else {
					logs.Error("好奇怪，断言失败。 key:%d", key)
					err := fmt.Errorf("好奇怪，断言失败。")
					return err
				}

			}
		case "AdminPrereduced":
			{
				if vR, ok := v.(models.AdminPrereduced); ok {
					insertAdminPreReduce(vR)
				} else {
					logs.Error("好奇怪，断言失败。 key:%d", key)
					err := fmt.Errorf("好奇怪，断言失败。")
					return err
				}
			}
		}
	} else {
		logs.Error("key[%v] not in map:%#v", key, result)
		err := fmt.Errorf("key[%v] not in map:%#v", key, result)
		return err
	}

	return nil
}

func insertReduceRecord(record models.ReduceRecord) {

	order, _ := models.GetOrder(record.OrderId)
	caseOver, _ := models.OneOverdueCaseByOrderID(record.OrderId)
	id, _ := device.GenerateBizId(types.ReduceRecordBiz)

	typ := types.ReduceTypeManual
	if record.ReduceType == 1 {
		typ = types.ReduceTypeAuto
	}
	recordNew := models.ReduceRecordNew{
		Id:                   id,
		OrderId:              record.OrderId,
		UserAccountId:        order.UserAccountId,
		ApplyUid:             record.OpUid,
		ConfirmUid:           record.OpUid,
		AmountReduced:        record.AmountReduced,
		PenaltyReduced:       record.PenaltyReduced,
		GraceInterestReduced: record.InterestReduced,
		ReduceType:           typ,
		ReduceStatus:         types.ReduceStatusValid,
		OpReason:             record.OpReason,
		ConfirmRemark:        "",
		ApplyTime:            record.Ctime,
		ConfirmTime:          record.Ctime,
		CaseID:               caseOver.Id,
		Ctime:                record.Ctime,
		Utime:                record.Utime,
	}
	models.OrmInsert(&recordNew)
}

func insertAdminPreReduce(record models.AdminPrereduced) {

	order, _ := models.GetOrder(record.OrderID)
	//caseOver, _ := models.OneOverdueCaseByOrderID(record.OrderId)
	//AmountReduced :=

	status := 0
	switch record.PrereducedStatus {
	case types.ClearReducedNotValid:
		{
			status = types.ReduceStatusNotValid
		}
	case types.ClearReducedInValid:
		{
			status = types.ReduceStatusInvalid
		}
	case types.ClearReducedValid:
		{
			status = types.ReduceStatusValid
		}
	}

	id, _ := device.GenerateBizId(types.ReduceRecordBiz)
	recordNew := models.ReduceRecordNew{
		Id:                            id,
		OrderId:                       order.Id,
		UserAccountId:                 order.UserAccountId,
		ApplyUid:                      record.Opuid,
		ConfirmUid:                    record.Opuid,
		AmountReduced:                 0,
		PenaltyReduced:                record.PenaltyPrereduced,
		GraceInterestReduced:          record.GracePeriodInterestPrededuced,
		ReduceType:                    types.ReduceTypePrereduced,
		ReduceStatus:                  status,
		CaseID:                        record.CaseID,
		DerateRatio:                   record.DerateRatio,
		GracePeriodInterestPrededuced: record.GracePeriodInterestPrededuced,
		PenaltyPrereduced:             record.PenaltyPrereduced,
		InvalidReason:                 record.InvalidReason,
		ApplyTime:                     record.Ctime,
		ConfirmTime:                   record.Utime,
		Ctime:                         record.Ctime,
		Utime:                         record.Utime,
	}
	models.OrmInsert(&recordNew)
}

func main() {
	// 设置进程 title
	procTitle := "fix-reduce-record"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := fmt.Sprintf("lock:%s", procTitle)
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}
	// -1 正常退出时,释放锁
	defer storageClient.Do("DEL", lockKey)

	// all record
	list := QueryAllReduce()
	if len(list) == 0 {
		logs.Error("list is empty")
		return
	}
	logs.Debug("len(list) :%d", len(list))
	// all pre record
	preList := QueryAllPreReduce()
	if len(preList) == 0 {
		logs.Error("preList is empty")
		return
	}
	logs.Debug("len(preList) :%d", len(preList))

	// sort record
	all := sortRecord(list, preList)
	logs.Debug("len(all) :%d", len(all))
	sortedKeys := make([]float64, 0)
	for k := range all {
		sortedKeys = append(sortedKeys, float64(k))
	}
	sort.Float64s(sortedKeys)
	//logs.Info("sortedKeys:%#v", sortedKeys)

	for k, v := range sortedKeys {
		logs.Warn("[%d][%d]", k, int64(v))
		err := InsertRecord(all, int64(v))
		if err != nil {
			logs.Error("[InsertRecord] err:%v v:%d", err, int64(v))
			break
		}
	}

	logs.Warn("statistic ok")
	logs.Info("[%s] politeness exit.", procTitle)
}
