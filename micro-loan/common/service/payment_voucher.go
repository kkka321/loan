package service

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"github.com/astaxie/beego/orm"
)

type PaymentVocherResp struct {
	Id           int64
	AccountId    int64
	OrderId      int64
	ResourceId   int64
	TotalPayment int64
	Ctime        int64
	Mobile       string
	Status       int64
	CheckStatus  types.LoanStatus
	ReimbMeans   string
	ReimbChannel int
}

func AddOnePaymentVoucherResource(record map[string]interface{}) {
	obj := models.PaymentVoucher{
		AccountId:  record["account_id"].(int64),
		OrderId:    record["order_id"].(int64),
		ResourceId: record["resource_id"].(int64),
		ReimbMeans: record["reimb_means"].(string),
		OpUid:      0,
		Status:     0,
		Ctime:      tools.GetUnixMillis(),
		Utime:      tools.GetUnixMillis(),
	}

	o := orm.NewOrm()
	o.Using(obj.Using())
	o.Insert(&obj)
}

func GetUserLastPaymentVoucher(userAccountId int64) (voucher models.PaymentVoucher, err error) {
	r := models.PaymentVoucher{}
	o := orm.NewOrm()
	o.Using(r.UsingSlave())
	err = o.QueryTable(r.TableName()).Filter("order_id", userAccountId).OrderBy("-id").One(&voucher)
	return
}

func GetPaymentVocherList(contr map[string]interface{}, page int, pagesize int) (lists []PaymentVocherResp, total int64, err error) {

	obj := models.PaymentVoucher{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pagesize

	// 初始化查询条件
	where := whereMaymentVocherBackend(contr)
	sqlCount := fmt.Sprintf(`SELECT COUNT(payment_voucher.id) FROM %s %s`, obj.TableName(), where)
	sqlList := fmt.Sprintf(`SELECT  payment_voucher.id,payment_voucher.status,payment_voucher.reimb_means,account_base.mobile,
	 payment_voucher.account_id, payment_voucher.order_id, orders.check_status, payment_voucher.ctime  FROM %s %s ORDER BY payment_voucher.ctime  desc LIMIT %d,%d`, obj.TableName(), where, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	list := []PaymentVocherResp{}
	r = o.Raw(sqlList)
	r.QueryRows(&list)
	for _, v := range list {
		tp := PaymentVocherResp{}
		tp.AccountId = v.AccountId
		tp.Ctime = v.Ctime
		tp.Id = v.Id
		tp.ResourceId = v.ResourceId
		total, err := repayplan.CaculateRepayTotalAmountByOrderID(v.OrderId)
		if err != nil {
			total = -1
		}
		tp.TotalPayment = total
		tp.OrderId = v.OrderId
		tp.Mobile = v.Mobile
		tp.CheckStatus = v.CheckStatus
		tp.ReimbMeans = v.ReimbMeans
		tp.Status = v.Status
		channels := strings.Split(v.ReimbMeans, " ")
		var tmp int
		tmp, ok := types.RemibChannel[channels[0]]
		if !ok {
			if len(tp.ReimbMeans) <= 0 {
				tmp = 3
			} else {
				tmp = 2
			}
		}

		tp.ReimbChannel = tmp
		tag, ok := contr["remib_tags"]
		if ok {
			if tag.(int) == tmp {
				lists = append(lists, tp)
			}
		} else {
			lists = append(lists, tp)
		}
	}

	if _, ok := contr["remib_tags"]; ok {
		total = int64(len(lists))
	}
	return
}

func whereMaymentVocherBackend(condStr map[string]interface{}) string {
	// 初始化查询条件
	cond := []string{}

	//借款id
	if v, ok := condStr["order_id"]; ok {
		cond = append(cond, fmt.Sprintf(" payment_voucher.order_id=%v", v))
	}
	//phone
	if v, ok := condStr["mobile"]; ok {
		cond = append(cond, fmt.Sprintf(" account_base.mobile=%v", v))

	}
	//
	if v, ok := condStr["account_id"]; ok {
		cond = append(cond, fmt.Sprintf(" payment_voucher.account_id=%v", v))
	}
	//工单类型
	if f, ok := condStr["check_status"]; ok {
		checkStatusArr := f.([]string)
		if len(checkStatusArr) > 0 {
			cond = append(cond, fmt.Sprintf(" orders.check_status IN(%s)", strings.Join(checkStatusArr, ", ")))
		}
	} else {
		cond = append(cond, fmt.Sprintf(" orders.check_status IN(7,9 )"))
	}

	//时间筛选
	if v, ok := condStr["ctime_start"]; ok {
		callEndTime := condStr["ctime_end"]
		cond = append(cond, fmt.Sprintf("payment_voucher.ctime > %v AND payment_voucher.ctime < %v", v.(int64), callEndTime))
	}

	if len(cond) > 0 {
		return "left join  microloan.account_base  on payment_voucher.account_id = account_base.id left join microloan.orders " +
			"on payment_voucher.account_id = orders.user_account_id WHERE " + strings.Join(cond, " AND ")
	} else {
		return "left join  microloan.account_base  on payment_voucher.account_id = account_base.id left join microloan.orders " +
			"on payment_voucher.account_id = orders.user_account_id where orders.check_status IN(7,9)"
	}

}
