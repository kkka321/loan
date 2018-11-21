package dao

import (
	"fmt"

	"github.com/astaxie/beego/orm"

	"micro-loan/common/models"
	"micro-loan/common/types"
)

func OneBusinessDetailByDateAndName(recordDate int64, paymentName string) (detail models.BusinessDetail, err error) {
	o := orm.NewOrm()
	o.Using(detail.Using())

	err = o.QueryTable(detail.TableName()).
		Filter("record_date", recordDate).
		Filter("payment_name", paymentName).
		One(&detail)
	return
}

func AddOrUpdateBusinessDetail(single *models.BusinessDetail, total *models.BusinessDetail) (err error) {

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

func AddOrUpdateBusinessDetailSingle(single *models.BusinessDetail, col ...string) (err error) {

	o := orm.NewOrm()
	o.Using(single.Using())

	// single
	if 0 == single.Id {
		_, err = o.Insert(single)
	} else {
		_, err = o.Update(single, col...)
	}
	return err
}

func PaymentThirdpartyList() (nameList []string, err error) {
	obj := models.ThirdpartyInfo{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	sql := "select distinct(name) from %s where is_payment_thirdparty = 1 order by name asc "

	sql = fmt.Sprintf(sql, obj.TableName())
	r := o.Raw(sql)

	_, err = r.QueryRows(&nameList)
	return
}

func OneByDateAndName(date int64, name string) (one models.BusinessDetail, err error) {
	o := orm.NewOrm()
	o.Using(one.UsingSlave())

	err = o.QueryTable(one.TableName()).
		Filter("record_date", date).
		Filter("payment_name", name).
		One(&one)
	return
}

func OneByDateAndNameLastRecord(startDate int64, name string) (one models.BusinessDetail, err error) {
	o := orm.NewOrm()
	o.Using(one.UsingSlave())

	err = o.QueryTable(one.TableName()).
		Filter("record_date__lt", startDate).
		Filter("payment_name", name).
		OrderBy("-record_date").
		One(&one)
	return
}

func StatisticAmount(startTime, endTime int64, companyCode int, payType int) (amount int64, err error) {
	obj := models.Payment{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	sql := "select sum(amount) as amount FROM %s where pay_type = %d and va_company_code = %d and ctime >= %d and ctime < %d"

	sql = fmt.Sprintf(sql, obj.TableName(), payType, companyCode, startTime, endTime)
	r := o.Raw(sql)

	err = r.QueryRow(&amount)
	return
}

func StatisticFee(startTime, endTime int64, api_md5 string) (fee int64, err error) {
	// 传入的时间是 印尼时间 而ThirdpartyStatisticFee保存的是utc时间
	obj := models.ThirdpartyStatisticFee{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	sql := "select total_price as fee FROM %s where api_md5 = '%s' and statistic_date >= %d and statistic_date < %d"

	sql = fmt.Sprintf(sql, obj.TableName(), api_md5, startTime, endTime)
	r := o.Raw(sql)

	err = r.QueryRow(&fee)
	return
}

func BusinessDetailSingleList(date int64, recordType int) (list []models.BusinessDetail, err error) {
	one := models.BusinessDetail{}
	o := orm.NewOrm()
	o.Using(one.UsingSlave())

	_, err = o.QueryTable(one.TableName()).
		Filter("record_date", date).
		Filter("record_type", recordType).
		All(&list)
	return
}

func BusinessDetailLendingBalance() (lendingBalance int64, err error) {
	//在贷余额=所有客户的（应还本金-已还本金-减免本金）之和
	repayPlan := models.RepayPlan{}
	order := models.Order{}
	o := orm.NewOrm()
	o.Using(repayPlan.UsingSlave())

	sql := "SELECT sum( amount - amount_payed - amount_reduced) AS lending_balance FROM %s WHERE order_id IN(SELECT id FROM %s WHERE check_status IN (%s))"
	status := fmt.Sprintf(" %d, %d,%d,%d",
		types.LoanStatusWaitRepayment,
		types.LoanStatusOverdue,
		types.LoanStatusPartialRepayment,
		types.LoanStatusRolling)

	sql = fmt.Sprintf(sql, repayPlan.TableName(), order.TableName(), status)

	r := o.Raw(sql)
	err = r.QueryRow(&lendingBalance)
	return
}

func BusinessDetailFeeIncome(startTime, endTime int64) (fee int64, err error) {
	repayPlan := models.RepayPlan{}
	o := orm.NewOrm()
	o.Using(repayPlan.UsingSlave())

	sql := "select sum(service_fee_payed) as fee from %s where ctime >= %d and ctime < %d"
	sql = fmt.Sprintf(sql, repayPlan.TableName(), startTime, endTime)

	r := o.Raw(sql)
	err = r.QueryRow(&fee)
	return
}

func BusinessDetailInterestIncome(startTime, endTime int64, interestName string) (interest int64, err error) {

	uTrans := models.User_E_Trans{}
	o := orm.NewOrm()
	o.Using(uTrans.UsingSlave())

	sql := "select sum(%s) as interest from %s where ctime >= %d and ctime < %d and pay_type = 2"
	sql = fmt.Sprintf(sql, interestName, uTrans.TableName(), startTime, endTime)

	r := o.Raw(sql)
	err = r.QueryRow(&interest)
	return
}

func BusinessDetailPenaltyIncomeOnlyInCurrent(current, yesterday string) (interest int64, err error) {

	overdue := models.RepayPlanOverdue{}
	o := orm.NewOrm()
	o.Using(overdue.UsingSlave())

	sql := "select sum(penalty) as interest from %s where order_id in (select order_id from %s where overdue_date = '%s' ) and order_id  not in (select order_id from %s where overdue_date = '%s') and penalty >0"
	sql = fmt.Sprintf(sql, overdue.TableName(), overdue.TableName(), current, overdue.TableName(), yesterday)

	r := o.Raw(sql)
	err = r.QueryRow(&interest)
	return
}

func BusinessDetailPenaltyIncomeBothIn(current, yesterday string) (interest int64, err error) {

	overdue := models.RepayPlanOverdue{}
	o := orm.NewOrm()
	o.Using(overdue.UsingSlave())

	interestCurrent := int64(0)
	interestYesterday := int64(0)
	sql := "select sum(penalty) from repay_plan_overdue where order_id in (select order_id from repay_plan_overdue where order_id in (select order_id from repay_plan_overdue where overdue_date = '%s' ) and order_id in (select order_id from repay_plan_overdue where overdue_date = '%s') and penalty >0 ) and overdue_date = '%s'"
	sqlCurrent := fmt.Sprintf(sql, current, yesterday, current)
	r := o.Raw(sqlCurrent)
	err = r.QueryRow(&interestCurrent)

	sqlYesterday := fmt.Sprintf(sql, current, yesterday, yesterday)
	r = o.Raw(sqlYesterday)
	err = r.QueryRow(&interestYesterday)

	interest = interestCurrent - interestYesterday
	return
}
