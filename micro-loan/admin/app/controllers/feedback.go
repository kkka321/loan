package controllers

import (
	"fmt"
	"micro-loan/common/i18n"
	"micro-loan/common/models"
	"micro-loan/common/pkg/feedback"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"
	"strings"

	"micro-loan/common/service"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils/pagination"
)

// FeedbackController 所有menu相关的控制器入口
type FeedbackController struct {
	BaseController
}

// Prepare 进入Action前的逻辑
func (c *FeedbackController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()
}

// List 列表
func (c *FeedbackController) List() {
	var condCntr = map[string]interface{}{}
	mobile := c.GetString("mobile")
	if len(mobile) > 0 {
		condCntr["mobile"] = mobile
	}
	c.Data["mobile"] = mobile

	//ID检索
	idCheck := c.GetString("id_check")
	if len(idCheck) > 0 {
		condCntr["id_check"] = idCheck
	}
	c.Data["id_check"] = idCheck

	accountID, _ := c.GetInt64("account_id")
	if accountID > 0 {
		condCntr["account_id"] = accountID
		c.Data["accountID"] = accountID
	}

	//App版本
	appVersion := c.GetString("app_version")
	if len(appVersion) > 0 {
		condCntr["app_version"] = appVersion
	}
	c.Data["app_version"] = appVersion

	//Api版本
	apiVersion := c.GetString("api_version")
	if len(apiVersion) > 0 {
		condCntr["api_version"] = apiVersion
	}
	c.Data["api_version"] = apiVersion

	//文本检索
	checkTxt := c.GetString("check_txt")
	if len(checkTxt) > 0 {
		condCntr["check_txt"] = checkTxt
	}
	c.Data["check_txt"] = checkTxt

	//字符数
	charNum := c.GetString("char_num")
	chars := 5
	if len(charNum) > 0 {
		chars, _ = strconv.Atoi(charNum)
	} else {
		charNum = "0"
		chars, _ = strconv.Atoi(charNum)
	}
	condCntr["char_num"] = chars
	c.Data["char_num"] = charNum

	//客户分类
	userTags := c.GetString("user_tags")
	if len(userTags) > 0 && userTags != "0" {
		condCntr["user_tags"] = userTags

	}
	selectedTagUserMap := map[interface{}]interface{}{}
	validTag, _ := strconv.Atoi(userTags)
	selectedTagUserMap[validTag] = nil
	c.Data["selectedTagUserMap"] = selectedTagUserMap

	//每页条数
	pageNumber, _ := c.GetInt("page_number", 1)

	selectedTagPageMap := map[interface{}]interface{}{}
	selectedTagPageMap[pageNumber] = nil
	c.Data["selectedTagPageMap"] = selectedTagPageMap

	//反馈分类
	tags := c.GetStrings("tags")
	if len(tags) > 0 {
		condCntr["tags"] = tags
	}
	selectedTagMap := map[interface{}]interface{}{}
	for _, tag := range tags {
		validTag, _ := strconv.Atoi(tag)
		selectedTagMap[validTag] = nil
	}
	c.Data["selectedTagMap"] = selectedTagMap

	splitSep := " - "
	//
	cTimeRange := c.GetString("ctime_range")
	if len(cTimeRange) > 16 {
		tr := strings.Split(cTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			fmt.Println("-------------------end", timeStart, timeEnd, cTimeRange)
			if timeStart > 0 && timeEnd > 0 {
				condCntr["ctime_start"] = timeStart
				condCntr["ctime_end"] = timeEnd
			}
		}
	}
	c.Data["ctimeRange"] = cTimeRange
	//ctimeRange := c.GetString("ctime_range")
	// c.Data["ctimeRange"] = ctimeRange
	// if start, end, err := tools.PareseDateRangeToMillsecond(ctimeRange); err == nil {
	// 	condCntr["ctime_start"], condCntr["ctime_end"] = start, end
	// }

	c.Data["tagMap"] = feedback.TagMap()
	c.Data["tagUserMap"] = feedback.TagsUserMap()
	c.Data["tagPageMap"] = feedback.TagPageMap()
	c.Data["domainURL"] = beego.AppConfig.String("domain_url")
	page, _ := c.GetInt("p")

	mp, ok := feedback.TagPageMap()[pageNumber]
	if !ok {
		mp = "15"
	}
	num, _ := strconv.Atoi(mp)
	pagesize := num

	list, count, _ := feedback.ListBackend(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list
	c.Data["exportURL"] = "/feedback/export?" + c.Ctx.Request.Form.Encode()
	// 获取指定

	c.Layout = "layout.html"
	c.TplName = "feedback/list.html"

	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "feedback/list_scripts.html"
}

func (c *FeedbackController) Export() {
	var condCntr = map[string]interface{}{}

	mobile := c.GetString("mobile")
	if len(mobile) > 0 {
		condCntr["mobile"] = mobile
	}
	c.Data["mobile"] = mobile

	tags := c.GetStrings("tags")
	if len(tags) > 0 {
		condCntr["tags"] = tags
	}
	selectedTagMap := map[interface{}]interface{}{}
	for _, tag := range tags {
		validTag, _ := strconv.Atoi(tag)
		selectedTagMap[validTag] = nil
	}
	c.Data["selectedTagMap"] = selectedTagMap

	splitSep := " - "
	//
	cTimeRange := c.GetString("ctime_range")
	if len(cTimeRange) > 16 {
		tr := strings.Split(cTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["ctime_start"] = timeStart
				condCntr["ctime_end"] = timeEnd
			}
		}
	}
	c.Data["cTimeRange"] = cTimeRange
	// ctimeRange := c.GetString("ctime_range")
	// c.Data["ctimeRange"] = ctimeRange
	// if start, end, err := tools.PareseDateRangeToMillsecond(ctimeRange); err == nil {
	// 	condCntr["ctime_start"], condCntr["ctime_end"] = start, end
	// }
	//

	host := beego.AppConfig.String("domain_url")

	list, _ := feedback.ExportXLSX(condCntr, c.LangUse, c.Ctx.ResponseWriter)
	fileName := fmt.Sprintf("feedback_%d.xlsx", tools.GetUnixMillis())
	lang := c.LangUse
	xlsx := excelize.NewFile()
	xlsx.SetCellValue("Sheet1", "A1", i18n.T(lang, "ID"))
	xlsx.SetCellValue("Sheet1", "B1", i18n.T(lang, "Mobile"))
	xlsx.SetCellValue("Sheet1", "C1", i18n.T(lang, "反馈分类"))
	xlsx.SetCellValue("Sheet1", "D1", i18n.T(lang, "客户分类"))
	xlsx.SetCellValue("Sheet1", "E1", i18n.T(lang, "Content"))
	xlsx.SetCellValue("Sheet1", "F1", i18n.T(lang, "API Version"))
	xlsx.SetCellValue("Sheet1", "G1", i18n.T(lang, "APP Version"))
	xlsx.SetCellValue("Sheet1", "H1", i18n.T(lang, "创建时间"))
	xlsx.SetCellValue("Sheet1", "I1", i18n.T(lang, "Image"))
	xlsx.SetCellValue("Sheet1", "J1", i18n.T(lang, "订单")+" ID")
	xlsx.SetCellValue("Sheet1", "K1", i18n.T(lang, "订单状态"))
	xlsx.SetCellValue("Sheet1", "L1", i18n.T(lang, "订单申请时间"))
	xlsx.SetCellValue("Sheet1", "M1", i18n.T(lang, "申请次数"))
	xlsx.SetCellValue("Sheet1", "N1", i18n.T(lang, "申请成功次数"))

	for i, d := range list {
		xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(i+2), d.Id)
		xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(i+2), d.Mobile)
		xlsx.SetCellValue("Sheet1", "C"+strconv.Itoa(i+2), feedback.GetTagDisplay(lang, d.Tags))
		xlsx.SetCellValue("Sheet1", "D"+strconv.Itoa(i+2), service.GetCustomerTags(lang, d.AccountTags))
		xlsx.SetCellValue("Sheet1", "E"+strconv.Itoa(i+2), d.Content)
		xlsx.SetCellValue("Sheet1", "F"+strconv.Itoa(i+2), d.ApiVersion)
		xlsx.SetCellValue("Sheet1", "G"+strconv.Itoa(i+2), d.AppVersion)
		xlsx.SetCellValue("Sheet1", "H"+strconv.Itoa(i+2), tools.MDateMHS(d.Ctime))
		if d.PhotoId1+d.PhotoId2+d.PhotoId3+d.PhotoId4 > 0 {
			url := host + "/feedback/image?id_check=" + tools.Int642Str(d.Id)
			xlsx.SetCellValue("Sheet1", "I"+strconv.Itoa(i+2), url)
		}
		xlsx.SetCellValue("Sheet1", "J"+strconv.Itoa(i+2), d.CurrentOrderID)
		xlsx.SetCellValue("Sheet1", "K"+strconv.Itoa(i+2), service.GetLoanStatusDesc(c.LangUse, d.CurrentOrderStatus))
		xlsx.SetCellValue("Sheet1", "L"+strconv.Itoa(i+2), tools.MDateMHS(d.CurrentOrderApplyTime))
		xlsx.SetCellValue("Sheet1", "M"+strconv.Itoa(i+2), d.ApplyOrderNum)
		xlsx.SetCellValue("Sheet1", "N"+strconv.Itoa(i+2), d.ApplyOrderSuccNum)
	}
	c.Ctx.Output.Header("Accept-Ranges", "bytes")
	c.Ctx.Output.Header("Content-Type", "application/octet-stream")
	c.Ctx.Output.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Ctx.Output.Header("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
	c.Ctx.Output.Header("Pragma", "no-cache")
	c.Ctx.Output.Header("Expires", "0")
	xlsx.Write(c.Ctx.ResponseWriter)
}

func (c *FeedbackController) Image() {
	c.Layout = "layout.html"
	c.TplName = "feedback/image.html"

	var condCntr = map[string]interface{}{}

	//ID检索
	idCheck := c.GetString("id_check")
	if len(idCheck) == 0 {
		c.Data["Content"] = ""
		return
	}

	condCntr["id_check"] = idCheck
	c.Data["id_check"] = idCheck
	list, count, _ := feedback.ListBackend(condCntr, 0, 10)
	if count != 1 {
		c.Data["Content"] = ""
		return
	}

	content := ""

	data := list[0]
	if data.PhotoId1 != 0 {
		str := service.GenImgHTML(data.PhotoId1)
		content += str
	}

	if data.PhotoId2 != 0 {
		str := service.GenImgHTML(data.PhotoId2)
		content += "<br /><br /><br />"
		content += str
	}

	if data.PhotoId3 != 0 {
		str := service.GenImgHTML(data.PhotoId3)
		content += "<br /><br /><br />"
		content += str
	}

	if data.PhotoId4 != 0 {
		str := service.GenImgHTML(data.PhotoId4)
		content += "<br /><br /><br />"
		content += str
	}

	c.Data["Content"] = beego.Str2html(content)
}

// List 列表
func (c *FeedbackController) PaymentVocherList() {
	var condCntr = map[string]interface{}{}

	//订单id
	idCheck := c.GetString("order_id")
	if len(idCheck) > 0 {
		condCntr["order_id"] = idCheck
	}
	c.Data["orderId"] = idCheck

	//客户id
	accountID, _ := c.GetInt64("account_id")
	if accountID > 0 {
		condCntr["account_id"] = accountID
		c.Data["accountID"] = accountID
	}

	//mobile
	mobile, _ := c.GetInt64("mobile")
	if mobile > 0 {
		condCntr["mobile"] = mobile
		c.Data["mobile"] = mobile
	}

	//订单状态
	checkStatusMulti := c.GetStrings("check_status")
	if len(checkStatusMulti) > 0 {
		condCntr["check_status"] = checkStatusMulti
	}

	//
	remibChannel, _ := c.GetInt("remib_tags")
	if remibChannel > 0 {
		condCntr["remib_tags"] = remibChannel
	}
	selectedTagRemibMap := map[interface{}]interface{}{}
	selectedTagRemibMap[remibChannel] = nil
	c.Data["selectedTagRemibMap"] = selectedTagRemibMap
	//time
	splitSep := " - "
	cTimeRange := c.GetString("ctime_range")
	if len(cTimeRange) > 16 {
		tr := strings.Split(cTimeRange, splitSep)
		if len(tr) == 2 {
			timeStart := tools.GetDateParseBackend(tr[0]) * 1000
			timeEnd := tools.GetDateParseBackend(tr[1])*1000 + 3600*24*1000
			if timeStart > 0 && timeEnd > 0 {
				condCntr["ctime_start"] = timeStart
				condCntr["ctime_end"] = timeEnd
			}
		}
	}

	c.Data["ctimeRange"] = cTimeRange
	c.Data["tagRemibMap"] = types.TagsRemibMap()
	c.Data["check_status"] = types.OrderStatusMap()
	c.Data["statusSelectMultiBox"] = service.BuildJsVar("statusSelectMultiBox", checkStatusMulti)
	page, _ := tools.Str2Int(c.GetString("p"))
	pagesize := 15

	list, count, _ := service.GetPaymentVocherList(condCntr, page, pagesize)
	paginator := pagination.SetPaginator(c.Ctx, pagesize, count)

	c.Data["paginator"] = paginator
	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "feedback/payment_vocher_list.html"

	c.Data["tagResultMap"] = types.TagsResultMap()
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "feedback/list_scripts.html"
}

func (c *FeedbackController) PaymentVocherImage() {
	//客户id
	orderId, _ := c.GetInt64("cid")

	list, _ := models.GetMultiPaymentByOrderId(orderId)

	c.Data["List"] = list

	c.Layout = "layout.html"
	c.TplName = "feedback/payment_vocher_image.html"

	c.LayoutSections = make(map[string]string)

}

func (c *FeedbackController) UpdatePaymentVocher() {
	action := ""
	url := "/feedback/paymentvocher"
	id, _ := c.GetInt64("ids")

	comment := c.GetString("comment")

	result, _ := c.GetInt64("result_tags")
	if result <= 0 {
		c.commonError(action, url, " must choice deal result")
		return
	}
	selectedTagResultMap := map[interface{}]interface{}{}
	selectedTagResultMap[result] = nil
	c.Data["selectedTagResultMap"] = selectedTagResultMap

	update := models.PaymentVoucher{
		Id:      id,
		OpUid:   c.AdminUid,
		Comment: comment,
		Status:  result,
		Utime:   tools.GetUnixMillis(),
	}
	_, errs := update.Updates("id", "op_uid", "comment", "status", "utime")
	if errs != nil {
		logs.Error("[UpdatePaymentVocher] update failed, err is", errs)
		c.commonError(action, url, "payment vocher update failed")
		return
	}

	c.Redirect("/feedback/paymentvocher", 302)
}
