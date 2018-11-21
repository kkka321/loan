package npwp

import (
	"encoding/json"
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"

	"micro-loan/common/service"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

//1:税号存在且一致  2:税号存在但不一致 3:税号不存在 4:服务端请求失败
const (
	VerifyResultExistUnanimous = 1 //税号存在且一致
	VerifyResultExistDiff      = 2 //税号存在但不一致
	VerifyResultNoExist        = 3 //税号不存在
	VerifyResultReqErr         = 4 //服务端请求失败
)

const (
	StatusExist   = 200
	StatusNoExist = 600
)

const (
	VerifySuccess = 1
	VerifyFailed  = 2
)

//var codeMap = map[int]int{
//	200: VerifyResultExistUnanimous,
//	401: VerifyResultErrEncrypt,
//	411: VerifyResultNoProduct,
//	600: VerifyResultExistUnanimous,
//}

type NpwpResp struct {
	Status       int    `json:"status"`
	Message      string `json:"message"`
	Npwp         string `json:"npwp"`
	CustomerName string `json:"customerName"`
}

func productId() string {
	return beego.AppConfig.String("bluepay_product_id")
}

func checkUrl() string {
	return beego.AppConfig.String("bluepay_npwp_url")
}

func key() string {
	return beego.AppConfig.String("bluepay_secret_key")
}

func NpwpVerify(accountId int64, npwpNo string) (resp NpwpResp, err error) {

	productId := productId()
	checkUrl := checkUrl()
	keyStr := key()

	encrypt := tools.Md5(fmt.Sprintf("productId=%s&npwp=%s%s", productId, npwpNo, keyStr))
	param := fmt.Sprintf("productId=%s&npwp=%s&encrypt=%s", productId, npwpNo, encrypt)
	reqUrl := checkUrl + "?" + param
	logs.Debug(reqUrl)

	//reqHeaders := map[string]string{}
	httpBody, httpCode, err := tools.SimpleHttpClient("GET", reqUrl, nil, "", tools.DefaultHttpTimeout())
	if err != nil {
		models.AddOneThirdpartyRecord(models.ThirdpartyBluepay, checkUrl, accountId, param, string(httpBody), 0, 0, httpCode)
		logs.Error("[NpwpVerify] SimpleHttpClient err:%v, accountId:%d npwpNo:%s", err, accountId, npwpNo)
		return
	}

	logs.Debug("httpBody:%v , httpCode:%d ", string(httpBody), httpCode)

	responstType, fee := thirdparty.CalcFeeByApi(checkUrl, param, param)
	models.AddOneThirdpartyRecord(models.ThirdpartyBluepay, checkUrl, accountId, param, string(httpBody), responstType, fee, httpCode)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: accountId,
		OrderId:       0,
		ApiMd5:        tools.Md5(checkUrl),
		Fee:           int64(fee),
		Result:        responstType,
	})

	if httpCode != 200 {
		err = fmt.Errorf("[NpwpVerify] bluepay npwp query httpCode is wrong [%d] accountId:%d", httpCode, accountId)
		logs.Error(err)
		return
	}
	return handelResp(httpBody, accountId, npwpNo)

}

func handelResp(httpBody []byte, accountId int64, npwpNo string) (resp NpwpResp, err error) {

	//resp := NpwpResp{}
	err = json.Unmarshal(httpBody, &resp)
	if err != nil {
		err = fmt.Errorf("[handelResp] npwp unmarshal failed, err is:%s. account:%d httpBody:%s", err.Error(), accountId, string(httpBody))
		logs.Error(err)
		return
	}
	logs.Debug("resp:%#v", resp)

	if resp.Status != StatusExist && resp.Status != StatusNoExist {
		logs.Error("[handelResp]  npwp status need check. resp:%#v", resp)
		return
	}

	tag := tools.GetUnixMillis()
	one := models.NpwpMobi{
		NpwpNo:       npwpNo,
		Status:       resp.Status,
		CustomerName: resp.CustomerName,
		Ctime:        tag,
		Utime:        tag,
	}

	_, err = models.OrmInsert(&one)
	if err != nil {
		logs.Error("[handelResp] OrmInsert err:%v accountId :%d npwpNo :%s resp:%#v", err, accountId, npwpNo, resp)
		return
	}
	return
}

func GiveReurn(accountId int64, npwpName string, npwpNo string, status int) int {

	switch status {
	case StatusExist:
		{
			account, _ := models.OneAccountBaseByPkId(accountId)
			if account.Realname == npwpName {
				aExt := saveResult(accountId, npwpNo, VerifySuccess)
				go service.IncreaseCreditByAuthoriation4Npwp(aExt, aExt.NpwpTime)
				return VerifyResultExistUnanimous
			} else {
				saveResult(accountId, npwpNo, VerifyFailed)
				logs.Info("[giveReurn] account.Realname：%s npwpName:%s accountId:%d", account.Realname, npwpName, accountId)
				return VerifyResultExistDiff
			}
		}
	case StatusNoExist:
		{
			return VerifyResultNoExist
		}
	default:
		{
			logs.Warn("[giveReurn] status:%d accountID:%d", status, accountId)
			return VerifyResultReqErr
		}
	}

}

func saveResult(accountId int64, npwpNo string, verifyResult int) models.AccountBaseExt {
	tag := tools.GetUnixMillis()
	accountExt, _ := models.OneAccountBaseExtByPkId(accountId)
	org := accountExt

	accountExt.NpwpNo = npwpNo
	accountExt.NpwpStatus = verifyResult
	accountExt.NpwpTime = tag
	accountExt.Utime = tag
	cols := []string{"npwp_no", "npwp_status", "npwp_time", "utime"}

	if accountExt.AccountId == 0 {
		accountExt.AccountId = accountId
		accountExt.Ctime = tag
		models.OrmInsert(&accountExt)
	} else {
		models.OrmUpdate(&accountExt, cols)
		models.OpLogWrite(accountId, accountId, models.OpCodeAccountBaseUpdate, accountExt.TableName(), org, accountExt)
	}
	return accountExt
}
