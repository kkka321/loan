package repayremind

import (
	"fmt"
	"micro-loan/common/models"
	"strings"

	"github.com/astaxie/beego/orm"
)

// CaseBackend 描述单个还款提醒案件数据
type CaseBackend struct {
	models.RepayRemindCase
	Realname                   string
	Mobile                     string
	TotalRepay                 int64
	TotalRepayPayed            int64
	Amount                     int64
	AmountPayed                int64
	AmountReduced              int64
	GracePeriodInterest        int64
	GracePeriodInterestPayed   int64
	GracePeriodInterestReduced int64
	Penalty                    int64
	PenaltyPayed               int64
	PenaltyReduced             int64
	RepayDate                  int64
}

// ListBackend 返回后台查询列表
func ListBackend(condCntr map[string]interface{}, page int, pagesize int) (list []CaseBackend, total int64, err error) {
	obj := models.RepayRemindCase{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}
	// if pagesize < 1 {
	// 	pagesize = types.DefaultPagesize
	// }
	offset := (page - 1) * pagesize

	// 初始化查询条件
	where := whereBackend(condCntr)

	sqlCount := fmt.Sprintf("SELECT COUNT(t1.id) FROM %s t1 LEFT JOIN %s t4 ON t1.user_account_id = t4.id %s",
		obj.TableName(), models.ACCOUNT_BASE_TABLENAME, where)
	sqlList := fmt.Sprintf(`SELECT t1.*, t4.realname, t4.mobile, t2.amount,
    t2.amount_payed,t2.amount_reduced, t2.repay_date,t2.grace_period_interest,
     t2.grace_period_interest_payed,t2.grace_period_interest_reduced, t2.penalty, t2.penalty_payed,t2.penalty_reduced
    FROM %s t1
     LEFT JOIN %s t2 ON t1.order_id = t2.order_id
     LEFT JOIN %s t3 ON t1.order_id = t3.id
     LEFT JOIN %s t4 ON t1.user_account_id = t4.id
     %s ORDER BY id desc LIMIT %d,%d`,
		obj.TableName(), models.REPAY_PLAN_TABLENAME, models.ORDER_TABLENAME, models.ACCOUNT_BASE_TABLENAME,
		where, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

func whereBackend(condCntr map[string]interface{}) string {
	// 初始化查询条件
	cond := []string{}
	// 主表 where 条件

	if v, ok := condCntr["id"]; ok {
		cond = append(cond, fmt.Sprintf("id=%d", v.(int64)))
	}
	if v, ok := condCntr["order_id"]; ok {
		cond = append(cond, fmt.Sprintf("order_id=%d", v.(int64)))
	}
	if v, ok := condCntr["ctime_start"]; ok {
		cond = append(cond, fmt.Sprintf("ctime>=%d", v))
	}
	if v, ok := condCntr["ctime_end"]; ok {
		cond = append(cond, fmt.Sprintf("ctime<%d", v))
	}
	if v, ok := condCntr["account_id"]; ok {
		cond = append(cond, fmt.Sprintf("user_account_id=%d", v))
	}
	if v, ok := condCntr["level"]; ok {
		cond = append(cond, fmt.Sprintf("level='%s'", v))
	}

	for i, condition := range cond {
		cond[i] = "t1." + condition
	}

	// 表 account_base 查询条件
	if v, ok := condCntr["mobile"]; ok {
		cond = append(cond, fmt.Sprintf("t4.mobile LIKE('%%%s%%')", v.(string)))
	}
	if v, ok := condCntr["realname"]; ok {
		cond = append(cond, fmt.Sprintf("t4.realname LIKE('%%%s%%')", v))
	}

	// 组织sql语句
	if len(cond) > 0 {
		return "WHERE " + strings.Join(cond, " AND ")
	}
	return ""
}
