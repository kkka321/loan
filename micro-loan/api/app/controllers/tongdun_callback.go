/**
* 同盾异步回调处理
* wudahai
* 2018-05-19
**/
package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty"
	"micro-loan/common/thirdparty/tongdun"
	"micro-loan/common/tools"
)

// TongdunCallbackController 同盾回调控制器
type TongdunCallbackController struct {
	beego.Controller
}

// CallBack 同盾回调入口
func (c *TongdunCallbackController) CallBack() {
	notifyEvent := c.GetString("notify_event")
	notifyDataJSON := c.GetString("notify_data")
	passbackParamsJSON := c.GetString("passback_params")
	notifyTime := c.GetString("notify_time")

	notifyData := tongdun.IdentityCheckCreateTask{}
	passbackParamsData := tongdun.PassbackParams{}
	if err := json.Unmarshal([]byte(notifyDataJSON), &notifyData); err != nil {
		logs.Error("[TongdunCallback] 同盾异步通知数据JSONdecode失败", "JSON:", notifyDataJSON, "ERROR:", err)
	}
	if err := json.Unmarshal([]byte(passbackParamsJSON), &passbackParamsData); err != nil {
		logs.Error("[TongdunCallback] 同盾透传参数 JSONdecode失败", "JSON:", passbackParamsJSON, "ERROR:", err)
	}

	//异步通知成功
	if tongdun.IsSuccess(notifyEvent) {
		service.RepirePassParams(&passbackParamsData)
		//记录及计费
		router := "/tongdun/callback/" + notifyData.Data.ChannelType
		responstType, fee := thirdparty.CalcFeeByApi(router, notifyDataJSON, "")
		models.AddOneThirdpartyRecord(models.ThirdpartyTongdun, router, passbackParamsData.AccountID, notifyDataJSON, "", responstType, fee, 200)
		event.Trigger(&evtypes.CustomerStatisticEv{
			UserAccountId: passbackParamsData.AccountID,
			OrderId:       0,
			ApiMd5:        tools.Md5(router),
			Fee:           int64(fee),
			Result:        responstType,
		})

		idCheckData, err := tongdun.QueryTask(passbackParamsData.AccountID, notifyData.TaskID)
		if err == nil {
			logs.Debug("[TongdunCallback] QueryTask idCheckData:%#v", idCheckData)

			// 社交和电商数据单独处理
			switch idCheckData.Data.ChannelType {
			case tongdun.IDSocialChannelType, tongdun.IDDSChannelType:
				{
					service.HandleTongdunSocialCallback(&idCheckData, passbackParamsData.AccountID, notifyTime)
				}
			default:
				{
					service.HandleTongdunNormalCallback(&idCheckData, passbackParamsData.AccountID, notifyTime)
				}
			}
		} else {
			logs.Error("[TongdunCallback] 更新TASK记录失败 ERROR:%v", err)
		}

		//返回正确响应
		c.SuccessResponse()
	} else {
		logs.Warn("[TongdunCallback] 异步通知,非SUCCESS状态不做处理", "EVENT:", notifyEvent, "CODE:", notifyData.Code)
	}
}

// SuccessResponse 给同盾返回成功响应
func (c *TongdunCallbackController) SuccessResponse() {
	// Response 同盾响应结构体
	var Response struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	Response.Message = "success"
	Response.Code = 0
	c.Data["json"] = &Response
	c.Ctx.Output.Status = 200
	c.ServeJSON()
}
