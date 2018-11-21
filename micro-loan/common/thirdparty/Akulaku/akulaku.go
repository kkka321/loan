package Akulaku

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/astaxie/beego"

	"encoding/json"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/monitor"
	"micro-loan/common/thirdparty"
	"micro-loan/common/tools"
	"strings"

	"github.com/astaxie/beego/logs"
)

type akulakuSourceData struct {
	CreditResult int `json:"creditresult"`
	//不是标准json格式，暂时不解析了
	//RiskType     []string `json:"risktype"`
	Related int `json:"related"`
	//不是标准json格式，暂时不解析了
	//RelatedRisks []string `json:"relatedrisks"`
	DataNo string `json:"dataNo"`
}

type akulakuData struct {
	Success string            `json:"success"`
	Data    akulakuSourceData `json:"data"`
	SysTime int64             `json:"sysTime"`
	ErrMsg  string            `json:"errMsg"`
}

func fixName(name string) string {
	b := []byte(name)
	var str string = ""
	for _, v := range b {
		if v >= 'A' && v <= 'Z' {
			str += string(v)
		} else if v >= 'a' && v <= 'z' {
			str += string(v)
		} else {
			continue
		}
	}

	return str
}

func GetRiskResult(releatdId int64, name string, ktp string) (int, error) {
	url := beego.AppConfig.String("akulaku_url")
	secretKey := beego.AppConfig.String("akulaku_secret_key")
	appkey := beego.AppConfig.String("akulaku_app_key")

	name = fixName(name)
	md5val := tools.Md5(fmt.Sprintf("%sappkey%sktp%sname%s", secretKey, appkey, ktp, name))
	paramStr := fmt.Sprintf("%s?name=%s&ktp=%s&appkey=%s&sign=%s", url, name, ktp, appkey, md5val)

	logs.Info("[GetRiskResult] orderId:%d, param:%s", releatdId, paramStr)

	client := &http.Client{}
	req, err := http.NewRequest("GET", paramStr, nil)
	if err != nil {
		logs.Error("[GetRiskResult] NewRequest err:%v", err)
		return -1, err
	}

	resp, err := client.Do(req)
	if err != nil {
		monitor.IncrThirdpartyCount(models.ThirdpartyAkulaku, 0)
		logs.Error("[GetRiskResult] do req err:%v", err)
		return -1, err
	}

	monitor.IncrThirdpartyCount(models.ThirdpartyAkulaku, resp.StatusCode)

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logs.Error("[GetRiskResult] request error status:%d, param:%s", resp.StatusCode, paramStr)
		return -1, fmt.Errorf("request param error")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("[GetRiskResult] ReadAll err:%v", err)
		return -1, err
	}

	bodyStr := string(body)
	logs.Debug("[GetRiskResult] orderId:%d, res:%s", releatdId, bodyStr)

	responstType, fee := thirdparty.CalcFeeByApi(url, paramStr, bodyStr)
	models.AddOneThirdpartyRecord(models.ThirdpartyAkulaku, url, releatdId, paramStr, bodyStr, responstType, fee, resp.StatusCode)
	event.Trigger(&evtypes.CustomerStatisticEv{
		UserAccountId: 0,
		OrderId:       releatdId,
		ApiMd5:        tools.Md5(url),
		Fee:           int64(fee),
		Result:        responstType,
	})

	data := akulakuData{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		logs.Error("[GetRiskResult] Unmarshal err:%v", err)
		return -1, nil
	}

	if strings.ToUpper(data.Success) != "TRUE" {
		return -1, err
	}

	if data.Data.CreditResult == 100 {
		return -1, err
	}

	return data.Data.CreditResult, err
}
