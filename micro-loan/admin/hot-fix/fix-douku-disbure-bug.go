package main

import (
	// 数据库初始化
	"encoding/json"
	"flag"
	"fmt"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/erikdubbelboer/gspt"
)

var fixOrders = []int64{}

var douKuDisApi = "https://kirimdoku.com/v2/api/cashin/remit"

var response = "\"{\\\"status\\\":0,\\\"message\\\":\\\"Remit accepted\\\",\\\"remit\\\":{\\\"paymentData\\\":{\\\"mallId\\\":\\\"2\\\",\\\"chainMallId\\\":null,\\\"accountNumber\\\":\\\"0000000899\\\",\\\"accountName\\\":\\\"DOKU\\\",\\\"channelCode\\\":\\\"07\\\",\\\"inquiryId\\\":\\\"I031152764843200\\\",\\\"currency\\\":\\\"IDR\\\",\\\"amount\\\":\\\"1000000.00\\\",\\\"trxCode\\\":\\\"6238c40900af1b1a2f89917b126da6737d8d0d8c\\\",\\\"responseCode\\\":\\\"00\\\",\\\"responseMsg\\\":\\\"Transfer Approve\\\"},\\\"transactionId\\\":\\\"DK01292409\\\"}}\""

func modifyResp(resp string) string {

	ret := strings.Replace(resp, "\\\"", "\"", len(resp)-1)
	ret = strings.Replace(ret, "\"{", "{", len(ret)-1)
	ret = strings.Replace(ret, "}\"", "}", len(ret)-1)
	return ret
}

var tableName string

func getInvoke(orderId int64) (one models.DisburseInvokeLog, err error) {

	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).Filter("order_id", orderId).OrderBy("-id").One(&one)
	return
}
func remitResp(httpBody []byte) (int, string, string) {
	status := -1
	var dokuRemitResp struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Remit   struct {
			TransactionId string `json:"transactionId"`
			PaymentData   struct {
				ResponseCode string `json:"responseCode"`
				ResponseMsg  string `json:"responseMsg"`
			} `json:"paymentData"`
		} `json:"remit"`
	}

	err := json.Unmarshal(httpBody, &dokuRemitResp)
	if err != nil {
		err = fmt.Errorf("remit response json.Unmarshal err, err is %s", err.Error())
		logs.Error(err)
	}
	status = dokuRemitResp.Status
	respCode := dokuRemitResp.Remit.PaymentData.ResponseCode
	responseMsg := dokuRemitResp.Remit.PaymentData.ResponseMsg
	return status, respCode, responseMsg
}

func fix(orderId int64) {
	logs.Info("start fix orderId:%d", orderId)

	order, _ := models.GetOrder(orderId)
	if order.CheckStatus != types.LoanStatusIsDoing {
		logs.Error("[fix] order status:%d orderId:%d", order.CheckStatus, orderId)
		return
	}

	// 1\qurey thiridparty
	obj := models.ThirdpartyRecord{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	sql := "select * from %s where related_id = %d and api = '%s' order by id desc limit 1"
	sql = fmt.Sprintf(sql, tableName, orderId, douKuDisApi)

	err := o.Raw(sql).QueryRow(&obj)
	if err != nil {
		logs.Error("[fix] QueryRow err:%v orderId:%d", err, orderId)
		return
	}

	//2\ modify resp
	logs.Warn("response is :%s", obj.Response)
	resp := modifyResp(obj.Response)
	logs.Warn("after modify response is :%s", resp)

	logs.Warn("Request is :%s", obj.Request)
	req := modifyResp(obj.Request)
	logs.Warn("after modify Request is :%s", req)

	//3 modify invoke
	invoke, err := getInvoke(orderId)
	if err != nil {
		logs.Error("getInvoke err:%v orderId:%d", err, orderId)
		return
	}

	httpBody := []byte(resp)
	remitUrl := douKuDisApi
	status, respCode, remitResp := remitResp(httpBody)
	if status == 0 && respCode == "00" {
		//成功才计费
		thirdPartyData, _ := models.GetThirpartyRecordById(obj.Id)
		responstType, fee := thirdparty.CalcFeeByApi(remitUrl, req, string(httpBody))
		event.Trigger(&evtypes.CustomerStatisticEv{
			UserAccountId: invoke.UserAccountId,
			OrderId:       orderId,
			ApiMd5:        tools.Md5(remitUrl),
			Fee:           int64(fee),
			Result:        responstType,
		})
		thirdPartyData.ResponseType = responstType
		thirdPartyData.FeeForCall = fee
		thirdPartyData.UpdateFee()
		invoke.FailureCode = ""
		invoke.DisbureStatus = types.DisbureStatusCallSuccess
	} else {
		err = fmt.Errorf("[DoKuDisburse Remit RespCode err], the body is: %s the remitResp is:%s", string(httpBody), respCode)
		logs.Error(err)
		invoke.FailureCode = remitResp
		invoke.DisbureStatus = types.DisbureStatusCallFailed
	}
	invoke.Utime = tools.GetUnixMillis()
	cols := []string{"disbure_status", "http_code", "utime", "failure_code"}
	models.OrmUpdate(&invoke, cols)

	// 4 modify repayplain

	jsonData := []byte(resp)
	datas := map[string]interface{}{}

	accountProfile, _ := models.OneAccountProfileByAccountID(invoke.UserAccountId)
	accountBase, _ := models.OneAccountBaseByPkId(invoke.UserAccountId)

	datas["account_id"] = invoke.UserAccountId
	datas["order_id"] = orderId
	datas["invoke_id"] = invoke.Id
	datas["bank_name"] = accountProfile.BankName
	datas["account_name"] = accountBase.Realname
	datas["amount"] = order.Loan
	datas["company_name"] = "doku"

	bankInfo, err := models.OneBankInfoByFullName(accountProfile.BankName)
	if err != nil {
		logs.Error("[ThirdPartyDisburse] OneBankInfoByFullName err:%v. unsport bank name:%s accountId:%d orderId：%d", err, accountProfile.BankName, invoke.UserAccountId, orderId)
		return
	}
	datas["banks_info"] = bankInfo

	doukuApi, _ := service.CreatePaymentApi(types.DoKu, datas)
	err = doukuApi.DisburseResponse(jsonData, datas)
	if err != nil {
		logs.Error("DisburseResponse err:%v order:%d", err, orderId)
		return
	}

	logs.Notice("ok")
}

func main() {

	// get flag
	ids := flag.String("ids", "", "ids")
	tableNameP := flag.String("tn", "", "table name")
	flag.Parse()
	logs.Warn("input ids:%v tableName:%v", *ids, *tableNameP)

	tableName = *tableNameP
	if len(*ids) == 0 || len(tableName) == 0 {
		flag.Usage()
		logs.Error("please input the ids to fixed")
		return
	}

	idStrs := strings.Split(*ids, ",")
	for _, v := range idStrs {
		id, _ := tools.Str2Int64(v)
		if id > 0 {
			fixOrders = append(fixOrders, id)
		}
	}

	if len(fixOrders) == 0 {
		logs.Error("fixOrders:%#v", fixOrders)
		return
	}
	logs.Info("input fixOrders:%v", fixOrders)

	procTitle := "fix-doku-disbure-timeout"
	gspt.SetProcTitle(procTitle)
	logs.Info("[%s] start launch.", procTitle)

	// lock
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()
	lockKey := fmt.Sprintf("lock:%s", procTitle)
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}
	defer storageClient.Do("DEL", lockKey)

	for _, v := range fixOrders {
		fix(v)
	}
}
