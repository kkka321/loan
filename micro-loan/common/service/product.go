package service

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/dao"
	pt "micro-loan/common/lib/product"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

var productFieldMap = map[string]string{
	"Id":              "Id",
	"DayInterestRate": "day_interest_rate",
	"DayFeeRate":      "day_fee_rate",
	"MinAmount":       "min_amount",
	"MaxAmount":       "max_amount",
	"MinPeriod":       "min_period",
	"MaxPeriod":       "max_period",
	"Ctime":           "ctime",
}

func ListProduct(condCntr map[string]interface{}, page int, pagesize int) (list []models.Product, num int64, err error) {
	obj := models.Product{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	// 构建查询对象
	qb, _ := orm.NewQueryBuilder(tools.DBDriver())

	where := "1 = 1"
	if val, ok := condCntr["id"]; ok {
		where += fmt.Sprintf(" AND id = %d", val.(int64))
	}

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = Pagesize
	}
	offset := (page - 1) * pagesize

	qb.Select("*").
		From(obj.TableName()).
		Where(where)

	// 导出 SQL 语句
	sql := qb.String()

	orderBy := ""
	if v, ok := condCntr["field"]; ok {
		if vF, okF := productFieldMap[v.(string)]; okF {
			orderBy = "ORDER BY " + vF
		} else {
			orderBy = "ORDER BY id"
		}
	} else {
		orderBy = "ORDER BY id"
	}

	if v, ok := condCntr["sort"]; ok {
		orderBy = fmt.Sprintf("%s %s", orderBy, v.(string))
	} else {
		orderBy = fmt.Sprintf("%s %s", orderBy, "DESC")
	}

	sql = fmt.Sprintf("%s %s LIMIT %d, %d", sql, orderBy, offset, pagesize)

	num, err = o.Raw(sql).QueryRows(&list)

	return
}

func ListProductOptRecord(productId int64) (list []models.ProductOptRecord, err error) {
	list, err = dao.GetMultiOptRecordByProductId(productId)

	return
}

// 写产品操作日志  edited 一定要存储正确的值
func ProductOptRecordWrite(opUid int64, opName string, opType types.ProductOptTypeEunm, orginal, edited *models.Product) {
	//如果时未定义状态 不需要记录他的修改日志
	if edited.Status == int(types.ProductStatusNever) && opType == types.ProductOptTypeModify {
		return
	}

	// 记录操作流水
	productOptRecord := models.ProductOptRecord{
		ProductId:   edited.Id,
		ProductName: edited.Name,
		Nickname:    opName,
		OpType:      int(opType),
		OpUid:       opUid,
		Original:    "",
		Edited:      "",
		Ctime:       tools.GetUnixMillis()}

	// 根据不同的类型修正 Original 和 Edited 的值，防止保存多余数据
	switch opType {
	case types.ProductOptTypeCreate:
		{
			editedJSON, _ := json.Marshal(edited)
			productOptRecord.Edited = string(editedJSON)
		}
	case types.ProductOptTypeModify:
		{
			orignalJSON, _ := json.Marshal(struct{ Remarks string }{Remarks: orginal.Remarks})
			editedJSON, _ := json.Marshal(struct{ Remarks string }{Remarks: edited.Remarks})
			productOptRecord.Original = string(orignalJSON)
			productOptRecord.Edited = string(editedJSON)
		}
	case types.ProductOptTypeUp:
		{
			productOptRecord.Original = ""
			productOptRecord.Edited = ""
		}
	case types.ProductOptTypeDown:
		{
			productOptRecord.Original = ""
			productOptRecord.Edited = ""
		}
	}

	productOptRecord.AddRecord()

}

// FindOneProductByPeriod 此方法只为印尼旧版本使用，通过冗余字段 period_loan 获取product 其他版本请不要调用
func FindOneProductByPeriod(period int) (p models.Product, err error) {
	o := orm.NewOrm()
	o.Using(p.UsingSlave())

	err = o.QueryTable(p.TableName()).Filter("period_loan", period).Limit(1).One(&p)

	return
}

// ProductTrialCalc 建议每种还款方式 单独创建文件
func ProductTrialCalc(trialIn types.ProductTrialCalcIn, product models.Product) (trialResults []types.ProductTrialCalcResult, err error) {
	switch product.RepayType {
	case types.ProductRepayTypeOnce:
		{
			trialResults, err = pt.TrialCalcRepayTypeOnce(trialIn, product)
		}
	case types.ProductRepayTypeByMonth:
		{
			// TODO:
		}
	case types.ProductRepayTypeAverageCapitalPlusInterest:
		{
			// TODO:
		}
	case types.ProductRepayTypeNoInterest:
		{
			// TODO:
		}
	}
	return
}

func IsProductCanActice(productId int64) (conflictId int64, err error) {

	product, err := models.GetProduct(productId)

	list, err := dao.ListActiveProductByType(product.ProductType)
	if err != nil {
		return
	}

	// 展期产品只能有一个处于上架状态
	if product.ProductType == int(types.ProductTypeRollLoan) &&
		len(list) > 0 {
		conflictId = list[0].Id
		err = fmt.Errorf("[IsProductCanActice]. product type conflict. now:%#v p:%#v", product, list[0])
		logs.Error(err)
		return
	}

	for _, v := range list {

		// 没有冲突
		if product.MinPeriod > v.MaxPeriod || product.MaxPeriod < v.MinPeriod {
			continue
		}

		conflictId = v.Id
		err = fmt.Errorf("[IsProductCanActice]. product Period conflict. now:%#v p:%#v", product, v)
		logs.Error(err)
		break
	}
	return
}

// 根据用户id返回适用的产品
func productSuitables(accountId int64) (list []models.Product) {
	isReloan := false
	if 0 != accountId {
		isReloan = dao.IsRepeatLoan(accountId)
	}

	productType := tools.ThreeElementExpression(isReloan, types.ProductTypeReLoan, types.ProductTypeFirst).(types.ProductTypeEunm)

	list, err := dao.ListActiveProductByType(int(productType))
	if err != nil {
		logs.Error("[productSuitables.ListActiveProductByType] err:%s accountId:%d  isReloan:%d",
			err, accountId, isReloan)
	}
	return list
}

// 根据用户id返回适用的产品
func ProductSuitablesForApp(accountId int64) (list []models.ProductReturnApp) {
	products := productSuitables(accountId)

	for _, v := range products {
		list = append(list, dao.GetProductApp(&v, accountId))
	}
	return list
}

// 根据用户id和期限返回适用的产品
func ProductSuitablesByPeriod(accountId int64, period int, loan int64) (product models.Product, err error) {
	list := productSuitables(accountId)

	if len(list) == 0 {
		err = fmt.Errorf("[ProductSuitablesByPeriod].ProductSuitables no suit products. accountId:%d", accountId)
		logs.Error(err)
		return product, err
	}

	for _, p := range list {
		if period >= p.MinPeriod &&
			period <= p.MaxPeriod &&
			loan >= p.MinAmount &&
			loan <= p.MaxAmount {
			return p, nil
		}
	}

	// 没找到返回错误
	err = fmt.Errorf("[ProductSuitablesByPeriod].ProductSuitables no find suit product. accountId:%d list:%#v", accountId, list)
	logs.Error(err)
	return product, err
}

// 根据期限返回 已上架的展期产品
func ProductRollSuitables() (product models.Product, err error) {
	list, err := dao.ListActiveProductByType(int(types.ProductTypeRollLoan))
	if err != nil {
		logs.Error("[ProductRollSuitables.ListActiveProductByType] err:%s ", err)
	}

	if len(list) == 0 {
		err = fmt.Errorf("[ProductRollSuitables]  no suit roll products.")
		logs.Error(err)
		return product, err
	}

	// 讲道理的话 只能取出一条展期的产品，如果大于1的话取id大的
	if len(list) > 1 {
		logs.Warn("[ProductRollSuitables] len(list)>0. list:%#v", list)
	}

	return list[0], nil
}
