package controllers

import (
	"micro-loan/common/cerror"
	"micro-loan/common/pkg/entrust"
	"micro-loan/common/tools"
	"strings"

	"github.com/astaxie/beego/logs"
)

// EntrustController 勤为
type EntrustController struct {
	APIBaseEntrustController
}

func (c *EntrustController) Prepare() {
	// 调用上一级的 Prepare 方
	c.APIBaseEntrustController.Prepare()

	// 统一将 ip 加到 RequestJSON 中
	c.RequestJSON["ip"] = c.Ctx.Input.IP()
	c.RequestJSON["related_id"] = int64(0)
}

func (c *EntrustController) GetRepayList() {

	if v, ok := c.RequestJSON["pname"]; ok {
		pname := v.(string)
		orderIDs := entrust.GetRepayList(pname)
		logs.Debug("[GetRepayList] orderIDS:", orderIDs)
		data := map[string]interface{}{
			"repay_orderid_list": orderIDs,
			"server_time":        tools.GetUnixMillis(),
		}
		c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.CodeSuccess, data)
		c.ServeJSON()
		return
	}
	c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.LostRequiredParameters, "")
	c.ServeJSON()
}

func (c *EntrustController) BaseInfo() {

	if !entrust.CheckOverdueBaseInfoRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	orderIDs := c.RequestJSON["order_id_list"].(string)
	pname := c.RequestJSON["pname"].(string)

	if orderIDs != "" {
		orderIDErr := false
		orderIDSlices := strings.Split(orderIDs, ",")
		for _, orderIDStr := range orderIDSlices {
			orderID, _ := tools.Str2Float64(orderIDStr)
			if orderID == 0 {
				orderIDErr = true
				break
			}
		}

		if orderIDErr || len(orderIDSlices) > 100 {
			data := map[string]interface{}{
				"server_time": tools.GetUnixMillis(),
			}
			c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.InvalidParameterValue, data)
			c.ServeJSON()
			return
		}
	}

	orders, num := entrust.GetEntrustList(orderIDs, pname)

	data := map[string]interface{}{
		"count":          num,
		"case_base_info": orders,
		"server_time":    tools.GetUnixMillis(),
	}

	// resp := cerror.BuildEntrustApiResponse(cerror.CodeSuccess, data)
	// JSONResp, _ := json.Marshal(resp)
	// logs.Debug("resp:", string(JSONResp))

	c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// ProcessedNotify用于勤为收到订单后处理回调，只有回调后才正式标记为勤为
func (c *EntrustController) ProcessedCallback() {

	if !entrust.CheckCallbackRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}
	// orderIDs := c.RequestJSON["order_id_list"].(string)
	orderIDs := c.RequestJSON["order_id_list"].(string)
	pname := c.RequestJSON["pname"].(string)
	orderIDErr := false
	orderIDSlices := strings.Split(orderIDs, ",")
	for _, orderIDStr := range orderIDSlices {
		orderID, _ := tools.Str2Float64(orderIDStr)
		if orderID == 0 {
			orderIDErr = true
			break
		}
	}

	if orderIDErr || len(orderIDSlices) > 100 {
		data := map[string]interface{}{
			"server_time": tools.GetUnixMillis(),
		}
		c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.InvalidParameterValue, data)
		c.ServeJSON()
		return
	}
	count := entrust.ProcessedNotify(orderIDs, pname)

	data := map[string]interface{}{
		"count":       count,
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()

}

func (c *EntrustController) RepayStatus() {

	if !entrust.CheckRepayStatusRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}
	orderIDs := c.RequestJSON["order_id_list"].(string)
	// pageSize, _ := tools.Str2Int(c.RequestJSON["page_size"].(string))

	if orderIDs != "" {
		orderIDErr := false
		orderIDSlices := strings.Split(orderIDs, ",")
		for _, orderIDStr := range orderIDSlices {
			orderID, _ := tools.Str2Float64(orderIDStr)
			if orderID == 0 {
				orderIDErr = true
				break
			}
		}

		if orderIDErr || len(orderIDSlices) > 100 {
			data := map[string]interface{}{
				"server_time": tools.GetUnixMillis(),
			}
			c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.InvalidParameterValue, data)
			c.ServeJSON()
			return
		}
	}

	orders, num := entrust.GetRepayStatus(orderIDs)
	data := map[string]interface{}{
		"number":       num,
		"repay_status": orders,
		"server_time":  tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

func (c *EntrustController) Contacts() {

	if !entrust.CheckContactRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}
	IDCards := c.RequestJSON["id_card_list"].(string)

	if IDCards != "" {
		IDCardErr := false
		IDCardSlices := strings.Split(IDCards, ",")
		for _, IDCardStr := range IDCardSlices {
			IDCard, _ := tools.Str2Float64(IDCardStr)
			if IDCard == 0 {
				IDCardErr = true
				break
			}
		}

		if IDCardErr || len(IDCardSlices) > 100 {
			data := map[string]interface{}{
				"server_time": tools.GetUnixMillis(),
			}
			c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.InvalidParameterValue, data)
			c.ServeJSON()
			return
		}
	}

	contacts, num := entrust.GetContact(IDCards)
	data := map[string]interface{}{
		"number":      num,
		"contacts":    contacts,
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// RollTC roll tentative calculation 展期试算
func (c *EntrustController) RollTC() {

	if !entrust.CheckRepayStatusRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}
	orderIDs := c.RequestJSON["order_id_list"].(string)
	// pageSize, _ := tools.Str2Int(c.RequestJSON["page_size"].(string))

	if orderIDs != "" {
		orderIDErr := false
		orderIDSlices := strings.Split(orderIDs, ",")
		for _, orderIDStr := range orderIDSlices {
			orderID, _ := tools.Str2Float64(orderIDStr)
			if orderID == 0 {
				orderIDErr = true
				break
			}
		}

		if orderIDErr || len(orderIDSlices) > 100 {
			data := map[string]interface{}{
				"server_time": tools.GetUnixMillis(),
			}
			c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.InvalidParameterValue, data)
			c.ServeJSON()
			return
		}
	}

	rollTC, num := entrust.GetRollTC(orderIDs)
	data := map[string]interface{}{
		"number":      num,
		"rollTC":      rollTC,
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}

// SPaymentCode 超市付款码
func (c *EntrustController) SPaymentCode() {

	if !entrust.CheckRepayStatusRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}
	orderIDs := c.RequestJSON["order_id_list"].(string)
	// pageSize, _ := tools.Str2Int(c.RequestJSON["page_size"].(string))

	if orderIDs != "" {
		orderIDErr := false
		orderIDSlices := strings.Split(orderIDs, ",")
		for _, orderIDStr := range orderIDSlices {
			orderID, _ := tools.Str2Float64(orderIDStr)
			if orderID == 0 {
				orderIDErr = true
				break
			}
		}

		if orderIDErr || len(orderIDSlices) > 100 {
			data := map[string]interface{}{
				"server_time": tools.GetUnixMillis(),
			}
			c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.InvalidParameterValue, data)
			c.ServeJSON()
			return
		}
	}

	paymentCode, num := entrust.GetSPaymentCode(orderIDs)
	data := map[string]interface{}{
		"number":             num,
		"spayment_code_info": paymentCode,
		"server_time":        tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildEntrustApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}
