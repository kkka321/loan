package main

import (
	"encoding/json"
	"fmt"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/types"
	"strings"

	"micro-loan/common/tools"

	"github.com/astaxie/beego/orm"
)

type xenditData struct {
	Id                      string `json:"id"`
	UserId                  string `json:"user_id"`
	ExternalId              string `json:"external_id"`
	Amount                  int64  `json:"Amount"`
	BankCode                string `json:"bank_code"`
	XenditFeeUserId         string `json:"xendit_fee_user_id"`
	XenditFeeAmount         string `json:"xendit_fee_amount"`
	AccountHolderName       string `json:"account_holder_name"`
	TransactionId           string `json:"transaction_id"`
	TransactionSequence     string `json:"transaction_sequence"`
	DisbursementDescription string `json:"disbursement_description"`
	FailureCode             string `json:"failure_code"`
	IsInstant               bool   `json:"is_instant"`
	Status                  string `json:"status"`
	Created                 string `json:"created"`
	Updated                 string `json:"updated"`
}

func queryOrderData() (list []int64, err error) {
	orderM := models.Order{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	sql := `SELECT related_id FROM thirdparty_record_201808 WHERE ctime >= 1534204800000 and response = '""' and thirdparty = 9 and api = 'https://api.xendit.co/disbursements';`

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

func queryThirdParty(orderId int64) (list []models.ThirdpartyRecord, err error) {
	orderM := models.ThirdpartyRecord{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	sql := fmt.Sprintf(`SELECT * FROM thirdparty_record_201808 WHERE related_id = %d and thirdparty = 9 and api = '/xendit/disburse_fund_callback/create' order by id desc;`, orderId)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

func handleCallback(list []models.ThirdpartyRecord) (bool, string) {
	for _, v := range list {
		str := strings.Replace(v.Request, "\\", "", -1)
		str = strings.Trim(str, "\"")
		data := xenditData{}
		ok := json.Unmarshal([]byte(str), &data)
		if ok != nil {
			continue
		}

		if data.Status == "COMPLETED" {
			return true, str
		}
	}

	return false, ""
}

func main() {
	ids, err := queryOrderData()
	if err != nil {
		fmt.Println(fmt.Sprintf("[ERROR]queryOrderData error %v", err))
		return
	}

	for _, v := range ids {
		order, err := models.GetOrder(v)
		if err != nil {
			fmt.Println(fmt.Sprintf("[ERROR]GetOrder wrong id:%d, error:%v", order.Id, err))
			continue
		}

		if order.CheckStatus != types.LoanStatusLoanFail {
			fmt.Println(fmt.Sprintf("[WARN]status skip id:%d, status:%d", order.Id, order.CheckStatus))
			continue
		}

		records, err := queryThirdParty(v)
		if err != nil {
			fmt.Println(fmt.Sprintf("[ERROR]queryThirdParty error orderId:%d, err:%v", order.Id, err))
			continue
		}

		if len(records) == 0 {
			fmt.Println(fmt.Sprintf("[WARN]queryThirdParty data empty orderId:%d", order.Id))
			continue
		}

		succ, str := handleCallback(records)
		if !succ {
			fmt.Println(fmt.Sprintf("[ERROR]handleCallback data empty orderId:%d", order.Id))
			continue
		}

		oldOrder := order
		order.CheckStatus = types.LoanStatusIsDoing
		order.Utime = tools.GetUnixMillis()

		models.UpdateOrder(&order)
		models.OpLogWrite(0, order.Id, models.OpCodeOrderUpdate, order.TableName(), oldOrder, order)

		fmt.Println(fmt.Sprintf("[INFO]change status to doing and callback orderId:%d, str:%s", order.Id, str))

		err = service.XenditDisburseCallback("/xendit/disburse_fund_callback/create", []byte(str))
		if err != nil {
			fmt.Println(fmt.Sprintf("[ERROR]XenditDisburseCallback error orderId:%d, err:%v", order.Id, err))
		}
	}
}
