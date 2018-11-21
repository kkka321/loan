package dao

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/models"
	"micro-loan/common/types"
)

func ListActiveProductByType(productType int) (list []models.Product, err error) {
	p := models.Product{}
	o := orm.NewOrm()
	o.Using(p.UsingSlave())

	_, err = o.QueryTable(p.TableName()).
		Filter("product_type", productType).
		Filter("status", types.ProductStatusValid).
		OrderBy("-id").
		All(&list)

	return
}

func GetProductApp(r *models.Product, accountID int64) (dst models.ProductReturnApp) {
	dst.Id = r.Id
	dst.DayInterestRate = r.DayInterestRate
	dst.DayFeeRate = r.DayFeeRate
	dst.ChargeInterestType = r.ChargeInterestType
	dst.ChargeFeeType = r.ChargeFeeType
	dst.MinAmount = r.MinAmount
	dst.MaxAmount = r.MaxAmount
	dst.MinPeriod = r.MinPeriod
	dst.MaxPeriod = r.MaxPeriod
	dst.CeilWay = r.CeilWay
	dst.CeilWayUnit = r.CeilWayUnit
	dst.CustomerVisible = int(r.CustomerVisible)

	//复贷用户展示 最大可贷额度
	if r.ProductType == int(types.ProductTypeReLoan) {
		quotaConfModel, err := GetLastAccountQuotaConf(accountID)
		if err != nil {
			logs.Error("[GetProductApp] GetLastAccountQuotaConf err:%v accountID:%d", err, accountID)
			return
		}

		// 防止异常情况
		if quotaConfModel.QuotaVisable > 0 {
			dst.MaxAmount = quotaConfModel.QuotaVisable
		}
	}
	return
}
