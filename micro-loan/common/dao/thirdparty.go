package dao

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/models"
)

func AddOrUpdateThirdpartyStatisticCustomer(single *models.ThirdpartyStatisticCustomer, total *models.ThirdpartyStatisticCustomer) (err error) {

	o := orm.NewOrm()
	o.Using(single.Using())

	o.Begin()
	// single
	if 0 == single.Id {
		_, err = o.Insert(single)
	} else {
		_, err = o.Update(single)
	}
	if err != nil {
		o.Rollback()
		return err
	}

	// total
	if 0 == total.Id {
		_, err = o.Insert(total)
	} else {
		_, err = o.Update(total)
	}
	if err != nil {
		o.Rollback()
		return err
	}
	o.Commit()

	return nil
}

// InsertOrUpdate 更新数据库
func InsertOrUpdateTongdunManual(tongdunModel models.AccountTongdun) (id int64, err error) {
	if tongdunModel.ID == 0 {
		_, err = models.InsertTongdun(tongdunModel)
		if err != nil {
			logs.Error("[InsertOrUpdateTongdunManual] InsertTongdun err:%v tongdun:%#v", err, tongdunModel)
		}
	} else {
		_, err = models.UpdateTongdun(tongdunModel)
		if err != nil {
			logs.Error("[InsertOrUpdateTongdunManual] UpdateTongdun err:%v tongdun:%#v", err, tongdunModel)
		}
	}
	return
}

func GetThirdparthStatisticFeeByMd5(apimd5 string, startDate, endDate int64) (thirdparthStatisticFee models.ThirdpartyStatisticFee, err error) {

	o := orm.NewOrm()
	o.Using(thirdparthStatisticFee.UsingSlave())
	err = o.QueryTable(thirdparthStatisticFee.TableName()).
		Filter("api_md5", apimd5).
		Filter("statistic_date__gte", startDate).
		Filter("statistic_date__lte", endDate).
		OrderBy("-id").
		Limit(1).One(&thirdparthStatisticFee)

	return
}
