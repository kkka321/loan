package service

import (
	"encoding/json"
	"micro-loan/common/lib/device"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty/voip"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"

	"micro-loan/common/lib/gaws"

	"github.com/astaxie/beego/logs"
)

type ExtensionCallParams struct {
	Mobile     string
	OrderId    string
	CaseId     string
	TicketType string
}

func ExtensionCall(obj ExtensionCallParams) (isOk int, callRecordIDStr, msg string) {
	if len(obj.Mobile) <= 0 {
		isOk = 0
		msg = voip.ContactMobileIsBlank

		return
	}

	orderId, _ := tools.Str2Int64(obj.OrderId)
	caseId, _ := tools.Str2Int64(obj.CaseId)
	ticketType := obj.TicketType

	// 获取工单信息
	ticket, err := models.GetTicketByTicketTypeAndRelatedID(ticketType, caseId)
	if err != nil {
		isOk = 0
		msg = voip.GetTicketInfoFail

		return
	}
	ticketStr, _ := tools.JsonEncode(ticket)
	logs.Info("[ExtensionCall] Get ticket info:", ticketStr)
	if ticket.AssignUID == 0 {
		isOk = 0
		msg = voip.TicketUnAssign
		return
	}

	// 根据工单分配人获取分机信息
	sipInfo, err := models.GetSipInfoByAssignID(ticket.AssignUID)
	if err != nil {
		isOk = 0
		msg = voip.NotAssignExtension

		return
	}
	sipInfoStr, _ := tools.JsonEncode(sipInfo)
	logs.Info("[ExtensionCall] Get sip info:", sipInfoStr)

	// 获取分机状态
	sipStatusResp, err := voip.VoipSipCallStatus(sipInfo.ExtNumber)
	if err != nil {
		isOk = 0
		msg = voip.GetSipStatusFail

		return
	}
	sipStatusRespStr, _ := tools.JsonEncode(sipStatusResp)
	logs.Info("[ExtensionCall] Get sip status:", sipStatusRespStr)
	status := sipStatusResp.Data.Result[0].Status
	if status != voip.Call_Status_1201 {
		isOk = 0
		msg = voip.GetCallStatusVal(status)

		return
	}

	// 获取订单
	order, _ := models.GetOrder(orderId)
	orderStr, _ := tools.JsonEncode(order)
	logs.Info("[ExtensionCall] Get order data:", orderStr)

	// 分机呼叫
	voipCallRecordID, _ := device.GenerateBizId(types.VoipCallRecordBiz)
	makeCallReq := voip.MakeCallRequest{
		ExtNumber:  sipInfo.ExtNumber,
		DestNumber: obj.Mobile,

		UserID:     strconv.FormatInt(ticket.AssignUID, 10),
		MemberID:   strconv.FormatInt(voipCallRecordID, 10),
		CustomUUID: strconv.FormatInt(orderId, 10),
	}

	makeCallResp, err := voip.VoipMakeCall(makeCallReq)
	if err != nil || makeCallResp.GetMakeCallStatus() == voip.VoipStatusFailed {
		isOk = 0
		msg = err.Error()

		return
	}
	makeCallRespStr, _ := tools.JsonEncode(makeCallResp)
	logs.Info("[ExtensionCall] Get make call response:", makeCallRespStr)

	// 插入分机通话记录
	callRecordIDStr = tools.Int642Str(voipCallRecordID)
	sipCallRecord := models.SipCallRecord{
		Id:         voipCallRecordID,
		OrderId:    orderId,
		AssignId:   ticket.AssignUID,
		ItemId:     int64(ticket.ItemID),
		ExtNumber:  sipInfo.ExtNumber,
		DestNumber: obj.Mobile,
	}
	_, err = sipCallRecord.Insert()
	if err != nil {
		isOk = 0
		msg = voip.InsertCallRecordFail

		return
	}
	logs.Info("[ExtensionCall] Insert voip call record id:", voipCallRecordID)

	isOk = 1
	msg = voip.Calling

	return
}

// voip电话结束后，消息推送
func SipBillMessageCallBack(reqBody []byte) (msg string, err error) {

	var billMsg voip.BillMessage
	err = json.Unmarshal(reqBody, &billMsg)
	if err != nil {
		logs.Error("[SipBillMessage] Unmarshal json decode request failed, data:", string(reqBody))
		msg = voip.VoipBillMessageFail
		return
	}
	billMsgStr, _ := tools.JsonEncode(billMsg)
	logs.Info("[SipBillMessage] Get sip bill message:", billMsgStr)

	// 存在录音文件时,下载录音文件,并且保存到aws
	if len(billMsg.RecordFileName) > 0 {
		msg, err = AudioFileDownAndUpAws(billMsg)
		if err != nil {
			return
		}
	}

	// 更新到数据库
	id, _ := tools.Str2Int64(billMsg.MemberID)
	isDial := voip.DBCallFail
	if billMsg.BillSec > 0 {
		isDial = voip.DBCallSuccess
	}

	calldir := voip.DBCallOutInt
	if billMsg.Type == voip.VoipCallInStr {
		calldir = voip.DBCallInInt
	}

	_, err = models.GetSipCallRecordById(id)
	if id > 0 && err == nil {
		sipCallRecord := models.SipCallRecord{
			Id:              id,
			CallId:          billMsg.CallId,
			ExtNumber:       billMsg.ExtNumber,
			DisNumber:       billMsg.DisNumber,
			DestNumber:      billMsg.DestNumber,
			CallDirection:   calldir,
			StartTime:       billMsg.StartTime,
			AnswerTime:      billMsg.AnswerTime,
			EndTime:         billMsg.EndTime,
			StartTimestamp:  tools.GetDateParseBeijing(billMsg.StartTime) * 1000,
			AnswerTimestamp: tools.GetDateParseBeijing(billMsg.AnswerTime) * 1000,
			EndTimestamp:    tools.GetDateParseBeijing(billMsg.EndTime) * 1000,
			IsDial:          isDial,
			BillSec:         billMsg.BillSec,
			Duration:        billMsg.Duration,
			HangupDirection: billMsg.HangupDirection,
			HangupCause:     billMsg.HangupCause,
			AudioRecordName: billMsg.RecordFileName,
			CallMethod:      billMsg.CallMethod,
		}

		sipCallRecord.Updates("id", "call_id", "disnumber", "destnumber", "call_direction", "start_time", "answer_time", "end_time",
			"start_timestamp", "answer_timestamp", "end_timestamp", "is_dial", "billsec", "duration", "hangup_direction", "hangup_cause",
			"audio_record_name", "call_method")
	} else {
		// 插入分机通话记录
		voipCallRecordID, _ := device.GenerateBizId(types.VoipCallRecordBiz)
		sipCallRecord := models.SipCallRecord{
			Id:              voipCallRecordID,
			CallId:          billMsg.CallId,
			ExtNumber:       billMsg.ExtNumber,
			DisNumber:       billMsg.DisNumber,
			DestNumber:      billMsg.DestNumber,
			CallDirection:   calldir,
			StartTime:       billMsg.StartTime,
			AnswerTime:      billMsg.AnswerTime,
			EndTime:         billMsg.EndTime,
			StartTimestamp:  tools.GetDateParseBeijing(billMsg.StartTime) * 1000,
			AnswerTimestamp: tools.GetDateParseBeijing(billMsg.AnswerTime) * 1000,
			EndTimestamp:    tools.GetDateParseBeijing(billMsg.EndTime) * 1000,
			IsDial:          isDial,
			BillSec:         billMsg.BillSec,
			Duration:        billMsg.Duration,
			HangupDirection: billMsg.HangupDirection,
			HangupCause:     billMsg.HangupCause,
			AudioRecordName: billMsg.RecordFileName,
			CallMethod:      billMsg.CallMethod,
		}
		sipCallRecordStr, _ := tools.JsonEncode(sipCallRecord)

		_, err = sipCallRecord.Insert()
		if err != nil {
			logs.Error("[SipBillMessage] Insert voip call record failed. sipCallRecordStr:", sipCallRecordStr, ", err:", err)
			msg = voip.VoipBillMessageFail
			return
		}
		logs.Info("[SipBillMessage] Insert voip call record. sipCallRecordStr:", sipCallRecordStr)

	}

	msg = voip.VoipBillMessageSuccess

	return
}

func AudioFileDownAndUpAws(billMsg voip.BillMessage) (msg string, err error) {
	recordFileResp, _ := voip.VoipRecordFileURL(billMsg.RecordFileName)
	recordFileUrl := recordFileResp.GetRecordFileUrl()

	fileName := tools.CreateVoipFileName(billMsg.RecordFileName)
	realFileName, err := tools.FileDownload(billMsg.RecordFileName, recordFileUrl)
	if err == nil {
		logs.Info("[SipBillMessage] Record file download success.")

		// 上传到aws
		_, err = gaws.AwsUpload(realFileName, fileName)
		defer tools.Remove(realFileName)
		if err != nil {
			logs.Error("[SipBillMessage] Upload record file to aws fail. file:", fileName, ", err:", err)
			msg = voip.VoipBillMessageFail
			return
		}
		logs.Info("[SipBillMessage] Upload record file to aws success. file:", fileName)

	} else {
		logs.Error("[SipBillMessage] Record file download fail. file:", fileName, ", err:", err)
		msg = voip.VoipBillMessageFail
	}

	return
}
