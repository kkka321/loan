package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/service"

	"github.com/astaxie/beego/logs"
)

func main() {

	orderId := int64(180716020192591708)
	accountId := int64(180627010158727345)
	//accountIdStr := tools.Int642Str(accountId)
	bankName := "Bank Permata"
	accountHolderName := "YOLANDA MEYSHA ZULFILIA MAHMUDAH"
	accountNumber := "1234567890"
	//desc := datas["desc"].(string)
	amount := int64(90000)

	err := service.ThirdPartyDisburse(orderId, accountId, bankName, accountHolderName, accountNumber, amount, 3, 1)
	logs.Debug(err)

	//code := doku.DoKuVaBankCodeTransform("BMRIIDJA1")
	//logs.Debug(code)

	//hash, err := doku.CheckVAWords("3461", "8856067010000002", "3736b0443a98f8329e0fde5004a3671cf5bd952a")
	//logs.Debug(hash)
	//logs.Debug(err)

	/*
		str := "abc,cde"
		arr := strings.Split(str, "1")

		logs.Debug(arr[0])
	*/

	//str := `{"user_id": "5785e6334d7b410667d355c4","external_id": "disbursement_12345","amount": 500000,"bank_code": "BCA","account_holder_name": "Rizky","disbursement_description": "Custom description","status": "PENDING","id": "57c9010f5ef9e7077bcb96b6"}`

	/*
		resp := xendit.XenditDisburseCorrectInquiryResp{}

		resp.Id = "test"
		resp.Status = "COMPLETED"
		resp.DisbursementDescription = "123123,234234"
		resp.ExternalId = "123123123"
		resp.Amount = 23423
		resp.AccountHolderName = "chester"
		resp.UserId = "chester_user_id"
		resp.IsInstant = true
		resp.BankCode = "BNI"

		byteStr, _ := json.Marshal(resp)

		xendit.SimulateDisburse(string(byteStr))

		//accounts := service.GetEaccounts(180627010161048114)
		//logs.Debug(accounts)

		var arr []string
		arr = append(arr, "123123")
		arr = append(arr, "343434")

		logs.Debug(arr)
	*/

}
