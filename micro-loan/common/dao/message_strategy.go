package dao

import (
	"fmt"

	"github.com/astaxie/beego/orm"

	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"github.com/astaxie/beego/logs"
)

func QueryRegisterNoOrderAccount(startTime int64, endTime int64, offsetId int64) ([]int64, error) {
	o := orm.NewOrm()
	order := models.Order{}
	account := models.AccountBase{}
	o.Using(order.UsingSlave())

	sql := fmt.Sprintf(`SELECT DISTINCT(a.id) 
FROM %s a LEFT JOIN %s o ON a.id = o.user_account_id 
WHERE a.id > %d AND a.register_time >= %d AND a.register_time <= %d AND ISNULL(o.id)
ORDER BY a.id 
LIMIT 100`,
		account.TableName(), order.TableName(),
		offsetId,
		startTime, endTime)

	list := make([]int64, 0)
	_, err := o.Raw(sql).QueryRows(&list)

	return list, err
}

func QueryRegisterOrderNoKtp(startTime int64, endTime int64, offsetId int64) ([]int64, error) {
	o := orm.NewOrm()
	order := models.Order{}
	account := models.AccountBase{}
	o.Using(order.UsingSlave())

	sql := fmt.Sprintf(`SELECT DISTINCT(a.id) 
FROM %s a LEFT JOIN %s o ON a.id = o.user_account_id
WHERE a.id > %d AND a.register_time >= %d AND a.register_time <= %d AND a.identity = "" AND o.id > 0  
ORDER BY a.id
LIMIT 100`,
		account.TableName(), order.TableName(),
		offsetId,
		startTime, endTime)

	list := make([]int64, 0)
	_, err := o.Raw(sql).QueryRows(&list)

	return list, err
}

func QueryNoRegister() ([]string, error) {
	uuidMd5List, _ := models.GetNeedRemindRegisterUUID()

	return uuidMd5List, nil
}

// 还款消息提醒
func GetRepayMessageOrderList(timetag int64, limit int64) (list []int64, err error) {
	orderM := models.Order{}
	repayPlan := models.RepayPlan{}
	orderExt := models.OrderExt{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	beforeDay := tools.NaturalDay(-1)
	afterDay := tools.NaturalDay(1)
	sql := fmt.Sprintf(`SELECT o.id FROM %s o
LEFT JOIN %s r ON r.order_id = o.id
LEFT JOIN %s e ON e.order_id = o.id
WHERE o.check_status IN(%d, %d) AND (r.repay_date >= %d AND r.repay_date <= %d) AND (ISNULL(e.repay_msg_run_time) || e.repay_msg_run_time != %d)
LIMIT %d`,
		orderM.TableName(),
		repayPlan.TableName(),
		orderExt.TableName(),
		types.LoanStatusWaitRepayment, types.LoanStatusPartialRepayment, beforeDay, afterDay, timetag,
		limit)

	logs.Debug("[GetRepayMessageOrderList] sql:", sql)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

// 逾期消息提醒
func GetOverdueMessageOrderList(idsBox []string) (list []int64, err error) {
	if len(idsBox) <= 0 {
		logs.Warning("[GetOverdueMessageOrderList] 必要参数为空. idsBox:", idsBox)
		return
	}

	orderM := models.Order{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	sql := fmt.Sprintf(`SELECT o.id FROM %s o
WHERE o.id NOT IN(%s) AND o.is_overdue = 1 AND o.check_status IN(%d, %d, %d) AND o.is_dead_debt = %d`,
		orderM.TableName(),
		strings.Join(idsBox, ", "),
		types.LoanStatusWaitRepayment, types.LoanStatusPartialRepayment, types.LoanStatusOverdue,
		types.IsDeadDebtNo)

	logs.Debug("[GetOverdueMessageOrderList] sql:", sql)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

func GetAllAccountList(offsetId int64, limit int64) (list []int64, err error) {
	o := orm.NewOrm()
	account := models.AccountBase{}
	o.Using(account.UsingSlave())

	sql := fmt.Sprintf(`SELECT id
FROM %s
WHERE id > %d
ORDER BY id
LIMIT %d`,
		account.TableName(),
		offsetId,
		limit)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

func QueryRegisterTmpOrderAccount(startTime int64, endTime int64, offsetId int64) ([]int64, error) {
	o := orm.NewOrm()
	order := models.Order{}
	account := models.AccountBase{}
	o.Using(order.UsingSlave())

	sql := fmt.Sprintf(`SELECT a.id 
FROM %s a LEFT JOIN %s o ON a.id = o.user_account_id 
WHERE a.id > %d AND a.register_time >= %d AND a.register_time <= %d 
GROUP BY a.id 
HAVING AVG(o.is_temporary) = %d
ORDER BY a.id
LIMIT 100`,
		account.TableName(), order.TableName(),
		offsetId,
		startTime, endTime,
		types.IsTemporaryYes)

	list := make([]int64, 0)
	_, err := o.Raw(sql).QueryRows(&list)

	return list, err
}

func QueryRepayClearAccount(startTime int64, endTime int64, isOverdue int, offsetId int64) ([]int64, error) {
	o := orm.NewOrm()
	order := models.Order{}
	account := models.AccountBase{}
	o.Using(order.UsingSlave())

	sql := fmt.Sprintf(`SELECT a.id 
FROM %s a LEFT JOIN %s o ON a.id = o.user_account_id 
WHERE a.id > %d AND o.finish_time >= %d AND o.finish_time <= %d AND o.pre_order = 0 AND o.is_overdue = %d
ORDER BY a.id
LIMIT 100`,
		account.TableName(), order.TableName(),
		offsetId,
		startTime, endTime,
		isOverdue)

	list := make([]int64, 0)
	_, err := o.Raw(sql).QueryRows(&list)

	return list, err
}

// 催收短信提醒
func GetCollectionRemindOrderList(idsBox []string, collectionRemindDays []types.CollectionRemindDay) (list []int64, err error) {
	if len(idsBox) <= 0 {
		logs.Warning("[GetCollectionRemindOrderList] 必要参数为空. idsBox:", idsBox)
		return
	}

	orderM := models.Order{}
	repayPlan := models.RepayPlan{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	var remindDate []int64
	for _, val := range collectionRemindDays {
		remindDate = append(remindDate, tools.NaturalDay(int64(-1*val)))
	}

	remindDateStr := tools.ArrayToString(remindDate, ",")
	sql := fmt.Sprintf(`SELECT o.id FROM %s o
LEFT JOIN %s r ON r.order_id = o.id
WHERE o.id NOT IN(%s) AND o.check_status=%d AND r.repay_date IN (%s)`,
		orderM.TableName(),
		repayPlan.TableName(),
		strings.Join(idsBox, ", "), types.LoanStatusOverdue,
		remindDateStr,
	)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

// 还款提醒
func GetRepayRemindOrderList(idsBox []string) (list []int64, err error) {
	if len(idsBox) <= 0 {
		logs.Warning("[GetRepayRemindOrderList] 必要参数为空. idsBox:", idsBox)
		return
	}

	orderM := models.Order{}
	repayPlan := models.RepayPlan{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	beforeDay := tools.NaturalDay(1)
	afterDay := tools.NaturalDay(-1)
	sql := fmt.Sprintf(`SELECT o.id FROM %s o
LEFT JOIN %s r ON r.order_id = o.id
WHERE o.id NOT IN(%s) AND o.check_status IN(%d, %d, %d) AND (r.repay_date = %d OR r.repay_date = %d)`,
		orderM.TableName(),
		repayPlan.TableName(),
		strings.Join(idsBox, ", "), types.LoanStatusWaitRepayment, types.LoanStatusPartialRepayment, types.LoanStatusOverdue, beforeDay,
		afterDay)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}
