package sms

// lib 独立的sms集成器
// 被 (尽量仅由)service 或者 controller 调用
// 1. 短信发送服务商选择策略
// 2. 短信发送
// 3. TODO 短信回执接收
// 4. sms log记录

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/sms/api"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/thirdparty/boomsms"
	"micro-loan/common/thirdparty/cmtelecom"
	"micro-loan/common/thirdparty/nexmo"
	"micro-loan/common/thirdparty/sms253"
	"micro-loan/common/thirdparty/textlocal"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// TODO 将这三个包级全局变量， 封装到一个结构体里， 确保数据安全性
var senderMap map[types.SmsServiceID]int
var singleInit sync.Once
var strategySort []int

const smsCallbackKeyPrefix = "SMS-"

func parseSenderStrategyConf() {
	jsonStrategyConf := config.ValidItemString("sms_strategy")
	var senderNameMap map[types.SmsServiceName]int
	json.Unmarshal([]byte(jsonStrategyConf), &senderNameMap)

	if len(senderNameMap) == 0 {
		logs.Error("No sms sender config, Emergency Error")
		return
	}
	senderMap = make(map[types.SmsServiceID]int)
	for k, v := range senderNameMap {
		strategySort = append(strategySort, v)
		senderMap[types.SmsServiceMap[k]] = v
	}
	sort.Ints(strategySort)
}

func initSender(serviceType types.ServiceType, mobile string, msg string) (api.API, error) {
	singleInit.Do(parseSenderStrategyConf)
	// if len(senderMap) == 0 {
	// 	// 策略配置解析
	// 	parseSenderStrategyConf()
	// }

	// 获取策略发送者
	senderKey, err := priorityAndFailedNext(serviceType, mobile)
	if err != nil {
		return nil, err
	}
	switch senderKey {
	case types.Sms253ID:
		return &sms253.Sender{Mobile: mobile, Msg: msg}, nil
	case types.NexoID:
		return &nexmo.Sender{Mobile: mobile, Msg: msg}, nil
	case types.TextlocalID:
		return &textlocal.Sender{Mobile: mobile, Msg: msg}, nil
	case types.BoomSmsID:
		return &boomsms.Sender{Mobile: mobile, Msg: msg}, nil
	case types.CmtelcomSmsID:
		return &cmtelecom.Sender{Mobile: mobile, Msg: msg}, nil
	default:
		return nil, errors.New("Sms Unexpected configure")
	}

}

// Send 发送短信
func Send(serviceType types.ServiceType, mobile string, msg string, relatedID int64) (status bool, err error) {
	status = false

	logs.Debug(">>> H5_Send:", serviceType, mobile, msg, relatedID)
	sender, err := initSender(serviceType, mobile, msg)
	if err != nil {
		logs.Error("SMS Send error occured, and send operation not triggered", err)
		return
	}

	return doSend(sender, serviceType, mobile, msg, relatedID)
}

func SendByKey(senderKey types.SmsServiceID, serviceType types.ServiceType, mobile string, msg string, relatedID int64) (status bool, err error) {
	var sender api.API

	switch senderKey {
	case types.Sms253ID:
		sender = &sms253.Sender{Mobile: mobile, Msg: msg}
	case types.NexoID:
		sender = &nexmo.Sender{Mobile: mobile, Msg: msg}
	case types.TextlocalID:
		sender = &textlocal.Sender{Mobile: mobile, Msg: msg}
	case types.BoomSmsID:
		sender = &boomsms.Sender{Mobile: mobile, Msg: msg}
	case types.CmtelcomSmsID:
		sender = &cmtelecom.Sender{Mobile: mobile, Msg: msg}

	default:
		return false, errors.New("Sms Unexpected configure")
	}

	return doSend(sender, serviceType, mobile, msg, relatedID)
}

func doSend(sender api.API, serviceType types.ServiceType, mobile string, msg string, relatedID int64) (status bool, err error) {
	resp, originalResp, err := sender.Send()
	status = resp.IsSuccess()

	{
		receipt, _ := json.Marshal(originalResp)

		//if !status {
		if serviceType == types.ServiceRequestLogin ||
			serviceType == types.ServiceLogin || serviceType == types.ServiceRegister {
			setFailedCacheForStrategy(mobile, sender.GetID())
		}
		//}
		smsM := models.Sms{
			MsgID:       resp.GetMsgID(),
			Content:     msg,
			Mobile:      mobile,
			SendStatus:  boolToIntStatus(status),
			RelatedID:   relatedID,
			SmsService:  sender.GetID(),
			Receipt:     string(receipt),
			ServiceType: serviceType,
		}
		smsM.AddSms()
	}

	return
}

// HandleDelivery 处理送达通知
func HandleDelivery(smsEncryptKey string, req *http.Request) {
	handler, _ := getHandlerByEncryptKey(smsEncryptKey)
	msgID, deliveryStatus, callbackContent, err := handler.Delivery(req)
	//
	if err != nil {
		logs.Error(err)
		return
	}
	models.UpdateSmsByMsgID(msgID, deliveryStatus, callbackContent)

	//fmt.Println(succ)
}

func getHandlerByEncryptKey(smsEncryptKey string) (api.API, error) {
	switch smsEncryptKey {
	case generateSmsCallbackKey(types.Sms253ID):
		return &sms253.Sender{}, nil
	case generateSmsCallbackKey(types.NexoID):
		return &nexmo.Sender{}, nil
	default:
		return nil, errors.New("Sms Unexpected callback")
	}
}

func generateSmsCallbackKey(id types.SmsServiceID) string {
	return tools.Md5(smsCallbackKeyPrefix + strconv.Itoa(int(id)))
}

func boolToIntStatus(b bool) int {
	if b {
		return 1
	}
	return 0
}
