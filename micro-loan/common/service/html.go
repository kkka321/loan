package service

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/dao"
	"micro-loan/common/i18n"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/thirdparty/doku"
	"micro-loan/common/thirdparty/voip"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func GetStatusDescrip(status int) string {
	statusMap := types.StatusMap()
	if val, ok := statusMap[status]; ok {
		return val
	}
	return "未定义"
}

func StatusDisplay(lang string, status int) (html string) {
	value := GetStatusDescrip(status)
	value = i18n.T(lang, value)
	if status == types.StatusValid {
		html = fmt.Sprintf(`<span class="label label-success">%s</span>`, value)
	} else if status == types.StatusInvalid {
		html = fmt.Sprintf(`<span class="label label-danger">%s</span>`, value)
	} else {
		html = fmt.Sprintf(`<span class="label label-primary">%s</span>`, value)
	}
	return
}

func StatusDisplayProduct(lang string, status int) (html string) {
	value, ok := types.GetProductStatusMap()[types.ProductStatusEunm(status)]
	if !ok {
		value = "未定义"
	}
	value = i18n.T(lang, value)
	if status == int(types.ProductStatusValid) {
		html = fmt.Sprintf(`<span class="label label-success">%s</span>`, value)
	} else if status == int(types.ProductStatusInValid) {
		html = fmt.Sprintf(`<span class="label label-danger">%s</span>`, value)
	} else {
		html = fmt.Sprintf(`<span class="label label-primary">%s</span>`, value)
	}
	return
}

func TypeDisplayProduct(lang string, v interface{}) (out string) {
	out = "-"

	if vInt, ok := v.(int); ok {
		productTypeMap := types.GetProductTypeMap()
		if desc, ok := productTypeMap[types.ProductTypeEunm(vInt)]; ok {
			out = desc
		}
	}

	if vE, ok := v.(types.ProductTypeEunm); ok {
		productTypeMap := types.GetProductTypeMap()
		if desc, ok := productTypeMap[vE]; ok {
			out = desc
		}
	}

	return i18n.T(lang, out)
}

func StatusDisplayProductRepay(lang string, status int) (value string) {
	value, ok := types.LoanStatusMap()[types.LoanStatus(status)]
	if !ok {
		value = "-"
	}
	value = i18n.T(lang, value)
	return
}

func GetCustomerTags(lang string, tags types.CustomerTags) (out string) {
	out = "未定义"

	customerTagsMap := types.CustomerTagsMap()
	if v, ok := customerTagsMap[tags]; ok {
		out = v
	}

	return i18n.T(lang, out)
}

func GetResourceUserMark(lang string, mark types.ResourceUseMark) (out string) {
	out = "未定义"
	resourceUseMarkMap := types.ResourceUseMarkMap()
	if v, ok := resourceUseMarkMap[mark]; ok {
		out = v
	}
	return i18n.T(lang, out)
}

func GetIsReloan(lang string, isReloan types.IsReloanEnum) (out string) {

	isReloanMap := types.IsReloanMap()
	if v, ok := isReloanMap[isReloan]; ok {
		out = v
	}

	return i18n.T(lang, out)
}

func GetLoanStatusDesc(lang string, status types.LoanStatus) (out string) {
	out = "未定义"

	checkStatus := types.AllOrderStatusMap()
	if v, ok := checkStatus[status]; ok {
		out = v
	}

	return i18n.T(lang, out)
}

func GetGenderDisplay(lang string, gender types.GenderEnum) (out string) {
	out = "未知"
	conf := types.GenderEnumMap()
	if str, ok := conf[gender]; ok {
		out = str
	}

	return i18n.T(lang, out)
}

func GenImgHTML(resourceID int64) (html string) {
	src := BuildResourceUrl(resourceID)
	if len(src) <= 0 {
		return ""
	}

	html = fmt.Sprintf(`<img src="%s" />`, src)

	return
}

func GenImgHrefStr(resourceID int64) string {
	src := BuildResourceUrl(resourceID)
	if len(src) <= 0 {
		return ""
	}

	host := beego.AppConfig.String("domain_url") + src

	return host
}

func GenImgHref(resourceID int64) (html string) {
	host := GenImgHrefStr(resourceID)

	if host != "" {
		html = fmt.Sprintf(`<a href="%s" target="_blank">%d</a>`, host, resourceID)
	}

	return
}

func TplTimeNow() (html string) {
	html = tools.GetDateMHS(tools.TimeNow())
	return
}

// GetTagDisplay 获取多重
func GetThirdpartyNameForTemplate(lang string, num int) (out string) {
	out = "未定义项"
	thridpartyMap := models.ThirdpartyNameMap
	if desc, ok := thridpartyMap[num]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}

func RiskTypeDisplay(lang string, riskType types.RiskTypeEnum) (out string) {
	out = "未定义类型"

	riskTypeMap := types.RiskTypeMap()
	if desc, ok := riskTypeMap[riskType]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}

func RiskItemDisplay(lang string, riskItem types.RiskItemEnum) (out string) {
	out = "未定义项"

	riskItemMap := types.RiskItemMap()
	if desc, ok := riskItemMap[riskItem]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}

func RiskReasonDisplay(lang string, reason types.RiskReason) (out string) {
	out = "未定义原因"

	riskReportReasonMap := types.RiskReportReasonMap()

	if desc, ok := riskReportReasonMap[reason]; ok {
		out = desc
		return i18n.T(lang, out)
	}

	return
}

func RiskRelieveReasonDisplay(lang string, reason types.RiskRelieveReason) (out string) {
	out = "未定义原因"

	riskRelieveReason := types.RiskRelieveReasonMap()

	if desc, ok := riskRelieveReason[reason]; ok {
		out = desc
		i18n.T(lang, out)
	}

	return
}

func RiskStatusDisplay(lang string, status types.RiskStatusEnum) (out string) {
	out = "未定义类型"

	m := types.RiskStatusMap()
	if desc, ok := m[status]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}

func RiskCtlStatusDisplay(lang string, status types.RiskCtlEnum) (out string) {
	out = "-"
	riskCtlMap := types.RiskCtlMap()
	if desc, ok := riskCtlMap[status]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}

func SmsServiceTypeDisplay(lang string, status types.ServiceType) (out string) {
	out = "未定义类型"
	serviceTypeMap := types.ServiceTypeEnumMap()
	if desc, ok := serviceTypeMap[status]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}

func SmsServiceDisplay(lang string, status types.SmsServiceID) (out string) {
	out = "未知"
	if desc, ok := types.SmsServiceIdMap[status]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}

func SmsDeliveryStatusDisplay(lang string, status int) (out string) {
	out = "未知"
	if desc, ok := types.DeliveryStatusMap[status]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}

func RiskCtlOperateCmd(lang string, orderId int64, accountId int64, checkStatus types.LoanStatus, riskStatus types.RiskCtlEnum, phoneVerifyTime int64) (html string) {
	html = i18n.T(lang, "暂无")

	// 如果已经拒了或通过,展示电核结果
	if phoneVerifyTime > 0 {
		html = fmt.Sprintf(`<button type="button" class="btn btn bg-orange show-verify-result" data-id="%d" data-toggle="modal" data-target="#showVerifyResult">%s</button>`,
			orderId, i18n.T(lang, "查看电核结论"))
		return
	}

	if riskStatus == types.RiskCtlWaitPhoneVerify {
		html = fmt.Sprintf(`<a href="/riskctl/phone_verify?order_id=%d" target="_blank">%s</a>`, orderId, i18n.T(lang, "电核"))
		return
	}

	if riskStatus == types.RiskCtlThirdBlacklistDoing {
		html = fmt.Sprintf(`<button type="button" class="btn btn bg-orange show-verify-result check_blacklist" data-id="%d">%s</button>`,
			orderId, i18n.T(lang, "黑名单验证"))
		return
	}

	return
}

func buildFixedQ(lang string, htmlBox []string, accountBase models.AccountBase) []string {

	var items []PhoneVerifyQuestionItem

	if dao.IsRepeatLoan(accountBase.Id) {
		//复贷配置
		items = fixedReloanVerifyQuestionItems
	} else {
		//首贷配置
		items = fixedVerifyQuestionItems
	}

	for _, item := range items {

		var label string
		if item.QuestionSN == 15002 || item.QuestionSN == 17002 {
			label = fmt.Sprintf(`<input type="button" class="btn btn-success btn-xs check_photo" title="%s" value="%s">`, i18n.T(lang, "检查图片"), i18n.T(lang, "检查图片"))
		}
		// 问题标题
		htmlBox = append(htmlBox, fmt.Sprintf(`<tr>
		<td><p class="text-aqua">%s %s</p></td>`, PhoneVerifyQuestionItemTrans(lang, item.Question), label))

		switch item.Field {
		case "owner_mobile_status":
			htmlBox = append(htmlBox, fmt.Sprintf(`<td>%s</td>`, accountBase.Mobile))
		default:
			htmlBox = append(htmlBox, fmt.Sprintf(`<td></td>`))
		}

		// input标签
		htmlBox = append(htmlBox, fmt.Sprintf(`<td class="col-xs-2 col-md-2 col-lg-2">`))
		switch item.InputType {
		case "text":
			htmlBox = append(htmlBox, fmt.Sprintf(`<input name="%s" type="text" class="non-fixed" required/>`, item.Field))
			// 输入框的时候,状态也要有
			fallthrough
		case "radio":
			htmlBox = append(htmlBox, fmt.Sprintf(`
	<label><input type="radio" name="%s" value="1" class="non-fixed radio-normal" required />%s</label>
	<label><input type="radio" name="%s" value="2" class="non-fixed radio-abnormal" required />%s</label>
	`, item.Field, i18n.T(lang, `正常`), item.Field, i18n.T(lang, `异常`)))
		}
		htmlBox = append(htmlBox, fmt.Sprintf(`</td>`))

		//			htmlBox = append(htmlBox, `<td>
		//	<button type="button" class="btn btn-block btn-primary redirect-reject">直拒</button>
		//</td>`)
		htmlBox = append(htmlBox, `</tr>`)

	}
	return htmlBox
}

func BuildPhoneVerifyQuestionHtml(accountId, orderId int64, lang string) (html string) {
	orderData, _ := models.GetOrder(orderId)
	accountBase, _ := models.OneAccountBaseByPkId(accountId)
	accountProfile, _ := dao.CustomerProfile(accountId)
	liveVerify, _ := dao.CustomerLiveVerify(accountId)
	clientInfo, _ := OrderClientInfo(orderId)
	lastClearOrder, _ := dao.AccountLastLoanClearOrder(accountId)

	clearOrderClientInfo, _ := OrderClientInfo(lastClearOrder.Id)

	imeiMd5 := tools.Md5(clientInfo.Imei)
	esRes, _, _, _ := EsSearchById(imeiMd5)

	var htmlBox []string
	htmlBox = append(htmlBox, `<table class="table table-bordered table-striped">`)

	// 固定问题
	htmlBox = buildFixedQ(lang, htmlBox, accountBase)

	// 随机问题
	var qidsBox []string

	var modules []ModuleSN

	if dao.IsRepeatLoan(accountId) {
		modules = reloanChoiceModule()
	} else {
		modules = choiceRandomModule()
	}

	for _, moduleSN := range modules {
		questionList := phoneVerifyQuestionConfig[moduleSN]
		/*
			if !dao.IsRepeatLoan(accountId) {
				if moduleSN == ModuleOther {
					items[14002] = questionList[14002]
					//改版后其他问题只有14002：您的放款银行卡卡号是？
					logs.Debug("items 14002")
				} else {
					items = questionList
				}
			} else {
				items = questionList
			}

			logs.Debug("items are:", items)
		*/

		var i = 0
		for _, item := range questionList {

			/*
				if dao.IsRepeatLoan(accountId) {
					if i >= 1 {
						break
					}
				} else {
					if i >= 2 {
						break
					}
				}
			*/

			if moduleSN == ModuleOther {
				item = questionList[14002]
			}

			//目前只随机一个了
			if i >= 1 {
				break
			}

			qidsBox = append(qidsBox, fmt.Sprintf("%d", item.QuestionSN))

			// 问题
			htmlBox = append(htmlBox, fmt.Sprintf(`<tr>
<td>
	<p class="text-aqua">
		%s
	</p>
</td>`, PhoneVerifyQuestionItemTrans(lang, item.Question)))

			// 问题提示
			htmlBox = append(htmlBox, fmt.Sprintf(`<td>`))
			switch item.Field {
			case "identity":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountBase.Identity))
			case "id_photo":
				htmlBox = append(htmlBox, fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, BuildResourceUrl(accountProfile.IdPhoto), i18n.T(lang, "身份证照片")))
			case "image_best":
				htmlBox = append(htmlBox, fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, BuildResourceUrl(liveVerify.ImageBest), i18n.T(lang, "活体识别")))
			case "image_env":
				htmlBox = append(htmlBox, fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, BuildResourceUrl(liveVerify.ImageEnv), i18n.T(lang, "活体识别")))
			case "age":
				customerAge, err := CustomerAge(accountBase.Identity)
				if err != nil {
					logs.Warn("customerAge has wrong. err:", err)
				}
				htmlBox = append(htmlBox, fmt.Sprintf(`%d`, customerAge))
			case "gender":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, GetGenderDisplay(lang, accountBase.Gender)))
			case "loan":
				htmlBox = append(htmlBox, fmt.Sprintf(`Rp%d`, orderData.Loan))
			case "period":
				htmlBox = append(htmlBox, fmt.Sprintf(`%d`, orderData.Period))
			case "apply_time":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, tools.MDateMHS(orderData.ApplyTime)))
			case "brand":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, clientInfo.Brand))
			case "company_name":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.CompanyName))
			case "job_type":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.JobTypeHTML()))
			case "monthly_income":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.MonthlyIncomeHTML()))
			case "company_address":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.CompanyAddress))
			case "contact1":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.Contact1))
			case "contact1_name":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.Contact1Name))
			case "relationship1":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.RelationshipHTML(accountProfile.Relationship1)))
			case "es_contact1":
				var number int = 0
				if n, ok := esRes.Source.NumberOfCallsToFirstContact[accountProfile.Contact1]; ok {
					number = n
				}
				htmlBox = append(htmlBox, fmt.Sprintf(`c1: %s, mobile: %s, number: %d`, accountProfile.Contact1Name, accountProfile.Contact1, number))
			case "contact2":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.Contact2))
			case "contact2_name":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.Contact2Name))
			case "relationship2":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.RelationshipHTML(accountProfile.Relationship2)))
			case "es_contact2":
				var number int = 0
				if n, ok := esRes.Source.NumberOfCallsToFirstContact[accountProfile.Contact2]; ok {
					number = n
				}
				htmlBox = append(htmlBox, fmt.Sprintf(`c2: %s, mobile: %s, number: %d`, accountProfile.Contact2Name, accountProfile.Contact2, number))
			case "bank_name":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.BankName))
			case "bank_no":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.BankNo))
			case "marital_status":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.MaritalStatusHTML()))
			case "children_number":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.ChildrenNumberHTML()))
			case "resident_address":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.ResidentAddress))
			case "education":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s`, accountProfile.EducationHTML()))
			case "last_time_loan_amount":
				htmlBox = append(htmlBox, fmt.Sprintf(`%d`, lastClearOrder.Loan))
			case "last_time_cellphone_model":
				htmlBox = append(htmlBox, fmt.Sprintf(`%s %s`, clearOrderClientInfo.Brand, clearOrderClientInfo.Model))
			default:
				htmlBox = append(htmlBox, i18n.T(lang, `无提示`))
			}
			htmlBox = append(htmlBox, fmt.Sprintf(`</td>`))

			htmlBox = append(htmlBox, fmt.Sprintf(`<td>`))
			switch item.InputType {
			case "text":
				htmlBox = append(htmlBox, fmt.Sprintf(`<input name="qid_value_%d" type="text" class="non-fixed" required/>`, item.QuestionSN))
				// 输入框的时候,状态也要有
				fallthrough
			case "radio":

				if item.QuestionSN != 16003 {
					htmlBox = append(htmlBox, fmt.Sprintf(`
	<label><input type="radio" name="qid_status_%d" value="1" class="non-fixed radio-normal" required />%s</label>
	<label><input type="radio" name="qid_status_%d" value="2" class="non-fixed radio-abnormal" required />%s</label>
`, item.QuestionSN, i18n.T(lang, `正常`), item.QuestionSN, i18n.T(lang, `异常`)))
				}

			}
			htmlBox = append(htmlBox, fmt.Sprintf(`</td>`))

			//			htmlBox = append(htmlBox, `<td>
			//	<button type="button" class="btn btn-block btn-primary redirect-reject">直拒</button>
			//</td>`)
			htmlBox = append(htmlBox, `</tr>`)

			i++
		}
	}

	htmlBox = append(htmlBox, `</table>`)
	// 所有本次选出来的问题qid

	boxLen := len(qidsBox)
	if boxLen < 6 {
		diff := 6 - boxLen
		for i := 0; i < diff; i++ {
			qidsBox = append(qidsBox, "0")
		}
	}

	htmlBox = append(htmlBox, fmt.Sprintf(`<input name="qids" type="hidden" value="%s">`, strings.Join(qidsBox, ",")))

	html = strings.Join(htmlBox, "\n")

	return
}

func infoReviewQuests() (questionsInt []int, questionsStr []string) {
	qs := config.ValidItemString("ticket_info_review_question")
	s := strings.Split(qs, ",")
	for _, v := range s {
		iV, _ := tools.Str2Int(v)
		if iV > 0 {
			questionsInt = append(questionsInt, iV)
		}
		questionsStr = append(questionsStr, v)
	}
	if len(questionsInt) == 0 {
		logs.Error("[infoReviewQuests] ticket_info_review_question config err. value:%s", qs)
	}
	return
}

func buildReasonsSelect(htmlBox []string, item PhoneVerifyQuestionItem, lang string) []string {
	htmlBox = append(htmlBox, fmt.Sprintf(`<td>
				<select name="qid_value_%d" id ="qid_value_%d" class="form-control reason-select" >
			`, item.QuestionSN, item.QuestionSN))

	htmlBox = append(htmlBox, fmt.Sprintf(`
				<option value="0">please select</option>`))

	for _, v := range item.Reasons {
		htmlBox = append(htmlBox, fmt.Sprintf(`
				<option value="%s">%s</option>`,
			v[ReasonSq], PhoneVerifyQuestionItemTrans(lang, v)))
	}
	htmlBox = append(htmlBox, "</select>")
	htmlBox = append(htmlBox, "</td>")

	return htmlBox
}

func BuildInfoReviewQuestionHtml(accountId, orderId int64, lang string) (html string) {
	// 先读取配置
	qs, qidsBox := infoReviewQuests()
	if len(qs) == 0 {
		return
	}
	var htmlBox []string

	htmlBox = append(htmlBox, `<table class="table table-bordered">`)
	for _, qId := range qs {
		if item, ok := phoneVerifyQuestionConfig[ModuleInfoReview][qId]; !ok {
			logs.Warn("[BuildInfoReviewQuestionHtml] qId:%d", qId)
			continue
		} else {
			// 问题
			htmlBox = append(htmlBox, fmt.Sprintf(`<tr>
			<td>
				<p class="text-aqua">
					%s
				</p>
			</td>`, PhoneVerifyQuestionItemTrans(lang, item.Question)))

			// 问题结果
			htmlBox = append(htmlBox, fmt.Sprintf(`<td>`))
			switch item.InputType {
			case "radio":
				htmlBox = append(htmlBox, fmt.Sprintf(`
					<label><input type="radio" name="qid_status_%d" id="%d" value="1" class="non-fixed radio-normal" />%s</label>
					<label><input type="radio" name="qid_status_%d" id="%d" value="2" class="non-fixed radio-abnormal"/>%s</label>
				`, item.QuestionSN, item.QuestionSN, i18n.T(lang, `正常`), item.QuestionSN, item.QuestionSN, i18n.T(lang, `异常`)))
			}
			htmlBox = append(htmlBox, fmt.Sprintf(`</td>`))

			// 拒绝原因
			htmlBox = buildReasonsSelect(htmlBox, item, lang)

			htmlBox = append(htmlBox, `</tr>`)
		}
	}
	htmlBox = append(htmlBox, `</table>`)

	// 所有本次选出来的问题qid

	boxLen := len(qidsBox)
	if boxLen < 6 {
		diff := 6 - boxLen
		for i := 0; i < diff; i++ {
			qidsBox = append(qidsBox, "0")
		}
	}
	htmlBox = append(htmlBox, fmt.Sprintf(`<input name="qids" type="hidden" value="%s">`, strings.Join(qidsBox, ",")))
	html = strings.Join(htmlBox, "\n")

	return
}

func GetPayTypeDesc(lang string, payType int) string {

	desc := "未定义"

	if payType == 1 {
		desc = "入账"
	} else if payType == 2 {
		desc = "出账"
	} else if payType == 3 {
		desc = "退款入账"
	} else if payType == 4 {
		desc = "退款出账"
	} else if payType == 5 {
		desc = "展期入账"
	} else if payType == 6 {
		desc = "展期出账"
	}
	return i18n.T(lang, desc)
}

func GetVaCompanyTypeDesc(lang string, vaCode int) string {

	desc := "未定义"

	if vaCode == 1 {
		desc = "摩比神奇"
	} else if vaCode == 2 {
		desc = "xendit"
	}

	return i18n.T(lang, desc)
}

func GetOpLoggerCodeDesc(vaCode models.OpCodeEnum) string {
	desc := "未定义"

	if val, ok := models.OpCodeList[vaCode]; ok {
		desc = val
	}
	return desc
}

// SmsVerifyCodeStatusDisplay 获取验证码状态
func SmsVerifyCodeStatusDisplay(lang string, status int) string {
	val := types.Undefined

	if v, ok := types.SmsVerifyCodeStatusMap[status]; ok {
		val = v
	}

	switch status {
	case types.VerifyCodeChecked:
		return i18n.T(lang, val)
	case types.VerifyCodeCheckFailed, types.VerifyCodeSendFailed:
		return i18n.T(lang, val)
	default:
		return i18n.T(lang, val)
	}
}

func UrgeOutReasonDisplay(lang string, reason types.UrgeOutReasonEnum) (str string) {
	urgeOutReasonMap := types.UrgeOutReasonMap()
	if r, ok := urgeOutReasonMap[reason]; ok {
		return i18n.T(lang, r)
	}

	return i18n.T(lang, "-")
}

func PhoneConnectDisplay(lang string, v int) (out string) {
	out = "-"
	conf := types.PhoneConnectMap()
	if str, ok := conf[v]; ok {
		return i18n.T(lang, str)
	}
	return i18n.T(lang, out)
}

func RepayInclinationDisplay(lang string, v int) (out string) {
	out = "-"
	conf := types.RepayInclinationMap()
	if str, ok := conf[v]; ok {
		return i18n.T(lang, str)
	}
	return i18n.T(lang, out)
}

func UnconnectReasonDisplay(lang string, v int) (out string) {
	out = "-"
	conf := types.UnconnectReasonMap()
	if str, ok := conf[v]; ok {
		return i18n.T(lang, str)
	}
	return i18n.T(lang, out)
}

func OverdueReasonItemDisplay(lang string, v types.OverdueReasonItemEnum) (out string) {
	out = "-"
	conf := types.OverdueReasonItemMap()
	if str, ok := conf[v]; ok {
		return i18n.T(lang, str)
	}
	return i18n.T(lang, out)
}

func SystemConfigItemTypeDisplay(lang string, itemType types.SystemConfigItemType) (str string) {
	systemConfigItemTypeMap := types.SystemConfigItemTypeMap()
	if d, ok := systemConfigItemTypeMap[itemType]; ok {
		return i18n.T(lang, d)
	}

	return i18n.T(lang, "-")
}

// GetRoleNameDisplay 获取后台展示角色名
func GetRoleNameDisplay(lang string, rt types.RoleTypeEnum, name string) (out string) {
	m := types.RoleTypeMap()

	if desc, ok := m[rt]; ok {
		out = desc
	}

	return i18n.T(lang, out) + "-" + i18n.T(lang, name)
}

// GetRoleTypeDisplay 角色类型或者部门展示
func GetRoleTypeDisplay(lang string, rt types.RoleTypeEnum) (out string) {
	m := types.RoleTypeMap()

	if desc, ok := m[rt]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}

func GetTicketItemDisplay(lang string, itemID types.TicketItemEnum) (out string) {
	m := types.TicketItemMap()

	if desc, ok := m[itemID]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}

func GetTicketStatusDisplay(lang string, status types.TicketStatusEnum) (out string) {
	m := types.TicketStatusMap()

	if desc, ok := m[status]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}

func PhoneObjectDisplay(lang string, v int) (out string) {
	out = "-"
	phoneObjectMap := types.PhoneObjectMap()
	if desc, ok := phoneObjectMap[v]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}
func UrgeCallTypeDisplay(lang string, v int) (out string) {
	out = "-"
	urgeTypeMap := types.UrgeTypeMap()
	if desc, ok := urgeTypeMap[v]; ok {
		out = desc
	}

	return i18n.T(lang, out)
}

func ChargeFeeTypeDisplay(lang string, v interface{}) (out string) {
	out = "-"
	if vInt, ok := v.(int); ok {
		chargeFeeMap := types.GetProductChargeFeeTypeMap()
		if desc, ok := chargeFeeMap[vInt]; ok {
			out = desc
		}
	}
	if vE, ok := v.(types.ProductChargeInterestTypeEnum); ok {
		chargeInterMap := types.GetProductChargeInterestTypeMap()
		if desc, ok := chargeInterMap[vE]; ok {
			out = desc
		}
	}

	return i18n.T(lang, out)
}

func CeilWayDisplay(lang string, v interface{}) (out string) {
	out = "-"
	if vE, ok := v.(types.ProductCeilWayEunm); ok {
		wayMap := types.GetProductCeilWayMap()
		if desc, ok := wayMap[vE]; ok {
			out = desc
		}
	}

	return i18n.T(lang, out)
}

func ProductOptTypeDisplay(lang string, v interface{}) (out string) {
	out = "-"
	if vInt, ok := v.(int); ok {
		vE := types.ProductOptTypeEunm(vInt)
		wayMap := types.GetProductOptTypMap()
		if desc, ok := wayMap[vE]; ok {
			out = desc
		}
	}

	return i18n.T(lang, out)
}

func ChargeTypeDisplay(lang string, v interface{}) (out string) {
	out = "-"
	if vInt, ok := v.(int); ok {
		wayMap := types.ChargeTypeMap()
		if desc, ok := wayMap[vInt]; ok {
			out = desc
		}
	}

	return i18n.T(lang, out)
}

func GetThirdpartyName(lang string, v interface{}) (out string) {
	out = ""
	if vInt, ok := v.(int); ok {
		if v, ok := models.ThirdpartyNameMap[vInt]; ok {
			out = v
		}
	}

	return i18n.T(lang, out)
}

func GetCommunicationWayDisplay(lang string, k int) (out string) {
	out = "-"
	if v, ok := types.CommnicationWayMap()[k]; ok {
		return i18n.T(lang, v)
	}
	return i18n.T(lang, out)
}

func GetIsEmptyDisplay(lang string, k int) (out string) {
	out = "否"
	if k == 1 {
		out = "是"
	}
	return i18n.T(lang, out)
}

func IsOutDisplay(lang string, v interface{}) (out string) {
	out = ""

	if vInt, ok := v.(int); ok {
		maps := types.UrgeFilterMap()
		if v, ok := maps[vInt]; ok {
			out = v
		}
	}

	return i18n.T(lang, out)
}

func UrgeTypeDisplay(lang string, v interface{}) (out string) {
	out = ""

	if vInt, ok := v.(int); ok {
		maps := types.UrgeTypeEnumMap()
		if v, ok := maps[vInt]; ok {
			out = v
		}
	}
	return i18n.T(lang, out)
}

func EntrustStatusDisplay(lang string, v interface{}) (out string) {
	out = ""

	if vInt, ok := v.(int); ok {
		maps := types.EntrustEnumMap()
		if v, ok := maps[vInt]; ok {
			out = v
		}
	}
	return i18n.T(lang, out)
}

func EntrustCompanyDisplay(lang string, v interface{}) (out string) {
	out = ""
	if vstring, ok := v.(string); ok {
		maps := types.EntrustCompanyMap()
		if v, ok := maps[vstring]; ok {
			out = v
		}
	}
	return i18n.T(lang, out)
}

func BuildJsVar(valName string, valValue interface{}) (html string) {
	valBSON, _ := tools.JsonEncode(valValue)
	html = fmt.Sprintf(`<script>var %s = %s;</script>`, valName, string(valBSON))

	return
}

// IsInMap template辅助方法 key 是否在map中
func IsInMap(m map[interface{}]interface{}, key interface{}) bool {
	if _, ok := m[key]; ok {
		return true
	}
	return false
}

// DisplayLimitText template辅助方法 key 是否在map中
func DisplayLimitText(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	runeS := []rune(s)
	return string(runeS[0:limit-2]) + "..."
}

func ArrayToParagraphString(a []int64) (html string) {
	for _, val := range a {
		html = fmt.Sprintf("%s%s%s%s", html, "<p>", fmt.Sprint(val), "</p>")
	}
	return
}

func OverdueDaysDisplay(orderID int64) (days int) {
	overdueCase, err := models.OneOverdueCaseByOrderID(orderID)
	if err != nil {
		return
	}

	if overdueCase.OverdueDays > 0 {
		days = overdueCase.OverdueDays
	}

	return days
}

func PaymentCodeDisplay(userAccountId int64) (paymentCode string) {
	fixPaymentCode, err := models.OneFixPaymentCodeByUserAccountId(userAccountId)
	if err != nil {
		return
	}
	paymentCode = fixPaymentCode.PaymentCode
	/*
		expireFlag := MarketPaymentCodeGenerateButton(orderID)
		if expireFlag == true {
			paymentCode = fmt.Sprintf("%s(expired)", paymentCode)
		}
	*/
	return
}
func PlatformMarkDisplay(lang string, mark int64) (display string) {
	a := models.AccountBase{}
	a.PlatformMark = mark

	isFirst := true
	for i := types.PlatformMark_No + 1; i <= types.PlatformMark_Max; i = i << 1 {
		if !a.IsPlatformMark(i) {
			continue
		}

		str := types.GetPlatformMarkDesc(i)
		if str == "" {
			continue
		}

		str = i18n.T(lang, str)
		if isFirst {
			display = str
			isFirst = false
		} else {
			display = display + " | " + str
		}
	}

	return
}

func MarketPaymentCodeGenerateButton(orderId int64) (expireFlag bool) {

	marketPayment, _ := models.GetMarketPaymentByOrderId(orderId)

	if marketPayment.PaymentCode == "" {
		//html = fmt.Sprintf(`<input id="generate_paymentcode" type="button" value="generate payment code" />`)
		expireFlag = true
	} else {
		//将过期时间和当前时间做比较

		now := time.Now().Unix() * 1000

		if marketPayment.ExpiryDate < now {
			//已过期
			//html = fmt.Sprintf(`<input id="generate_paymentcode" type="button" value="generate payment code" />`)
			expireFlag = true
		} else {
			//html = fmt.Sprintf("PaymentCode is: [%s]. ExpireDate is [%d]", marketPayment.PaymentCode, expireDateTime)
			expireFlag = false
		}
	}

	return
}

func CheckPromiseIsToday(timestamp int64) (isToday bool) {

	today := tools.GetUnixMillis()
	todayStr := tools.MDateMHSDate(today)
	todayZero, _ := tools.GetTimeParseWithFormat(todayStr, "2006-01-02")

	if timestamp/1000-todayZero >= 0 && timestamp/1000-todayZero <= 24*60*60 {
		isToday = true
	}

	return
}

func ReduceTypeDisplay(lang string, in int) (out string) {
	out = "-"

	if desc, ok := types.ReduceTypeMap[in]; ok {
		out = desc
	}

	out = i18n.T(lang, out)

	switch in {
	case types.ReduceTypeManual:
		{
			out = fmt.Sprintf(`<span class="label label-success">%s</span>`, out)
		}
	case types.ReduceTypeAuto:
		{
			out = fmt.Sprintf(`<span class="label label-warning">%s</span>`, out)
		}
	case types.ReduceTypePrereduced:
		{
			out = fmt.Sprintf(`<span class="label label-primary">%s</span>`, out)
		}

	}

	return out
}

func ReduceStatusDisplay(lang string, in int) (out string) {
	out = "-"

	if desc, ok := types.ReduceStatusMap[in]; ok {
		out = desc
	}
	return i18n.T(lang, out)
}

func VoipCallMothodDisplay(lang string, v int) (out string) {
	out = "-"
	conf := voip.VoipCallMethodMap
	if str, ok := conf[v]; ok {
		return i18n.T(lang, str)
	}
	return i18n.T(lang, out)
}

func VoipCallDirectionDisplay(lang string, v int) (out string) {
	out = "-"
	conf := voip.VoipCallDirectionMap
	if str, ok := conf[v]; ok {
		return i18n.T(lang, str)
	}
	return i18n.T(lang, out)
}

func VoipIsDialDisplay(lang string, v int) (out string) {
	out = "-"
	conf := voip.VoipCallDialStatusMap
	if str, ok := conf[v]; ok {
		return i18n.T(lang, str)
	}
	return i18n.T(lang, out)
}

func VoipHangupDisplay(lang string, v int) (out string) {
	out = "-"
	conf := voip.VoipSipHangupMap
	if str, ok := conf[v]; ok {
		return i18n.T(lang, str)
	}
	return i18n.T(lang, out)
}

func CouponTypeDisplay(lang string, v types.CouponType) (out string) {
	out = "-"
	conf := types.CouponTypeMap
	if str, ok := conf[v]; ok {
		return i18n.T(lang, str)
	}
	return i18n.T(lang, out)
}

func CouponAvaliableDisplay(lang string, v int) (out string) {
	out = "-"
	conf := types.CouponMap
	if str, ok := conf[v]; ok {
		return i18n.T(lang, str)
	}
	return i18n.T(lang, out)
}

func CouponDistributeDisplay(lang string, v int, startDate int64, endDate int64) string {
	now := tools.GetUnixMillis()
	str := "发放中"

	if v == types.CouponInvalid {
		str = "已停止"
	} else {
		if now < startDate {
			str = "未开始"
		} else if now > endDate {
			str = "已停止"
		}
	}

	return i18n.T(lang, str)
}

func CouponStatusDisplay(lang string, v types.CouponStatus) (out string) {
	out = "-"
	conf := types.CouponStatusMap
	if str, ok := conf[v]; ok {
		return i18n.T(lang, str)
	}

	return i18n.T(lang, out)
}

func BannerTypeDisplay(lang string, v int) (out string) {
	out = "-"
	conf := types.BannerTypeMap
	if str, ok := conf[v]; ok {
		return i18n.T(lang, str)
	}
	return i18n.T(lang, out)
}

func AdPositionDisplay(lang string, position int) (out string) {
	out = "-"
	conf := types.AdPositionMap
	if str, ok := conf[position]; ok {
		return i18n.T(lang, str)
	}
	return i18n.T(lang, out)
}

func VaDisplayBankCode(eAccount models.User_E_Account) (out string) {
	if eAccount.VaCompanyCode == types.DoKu {
		out = doku.DoKuVaBankCodeTransform(eAccount.BankCode)
		return
	}
	return eAccount.BankCode
}

func CompanyDisplay(lang string, bankName string) (display string) {
	display = "-"

	one, _ := models.OneBankInfoByFullName(bankName)
	switch one.LoanCompanyCode {
	case types.Xendit:
		{
			display = "Xendit"
		}
	case types.Bluepay:
		{
			display = "Bluepay"
		}
	case types.DoKu:
		{
			display = "DoKu"
		}
	}

	return
}

func RepaymentSourceDisplay(lang string, paymentId int64, payType int, vaCompanyCode int) string {
	if payType == types.PayTypeMoneyOut || payType == types.PayTypeRefundOut || payType == types.PayTypeRollOut || payType == types.PayTypeTran {
		return ""
	}

	if vaCompanyCode == types.MobiCoupon {
		return i18n.T(lang, "优惠券")
	}

	if vaCompanyCode == types.MobiFundVirtual {
		return i18n.T(lang, "虚拟还款")
	}

	if vaCompanyCode == types.MobiRefundToOrder {
		return i18n.T(lang, "余额")
	}

	if vaCompanyCode == types.MobiPreInterest {
		if payType == types.PayTypeRefundIn {
			return i18n.T(lang, "余额")
		} else {
			return i18n.T(lang, "砍头收取")
		}
	}

	if paymentId == 0 {
		return ""
	}

	payment, _ := models.GetPaymentById(paymentId)

	return payment.VaCode
}

func CompanyDisplayByCode(lang string, companyType int) (display string) {
	display = "-"

	if v, ok := types.FundCodeNameMap()[companyType]; ok {
		display = v
	}
	return
}

func PhoneVerifyCallResult(lang string, in int) (out string) {
	out = "-"

	if desc, ok := types.PhoneVerifyTypeMap[in]; ok {
		out = desc
	}
	return i18n.T(lang, out)
}

func PushTargetDisplay(lang string, pushTarget types.PushTarget) string {
	if v, ok := types.PushTargetMap[pushTarget]; ok {
		return v
	}

	return ""
}

func MessageTypeDisplay(lang string, messageType int) string {
	if v, ok := types.MessageTypeMap[messageType]; ok {
		return v
	}

	return ""
}

func PushWayDisplay(lang string, pushWay int) string {
	if v, ok := types.PushWayMap[pushWay]; ok {
		return v
	}

	return ""
}

func SchemaModeDisplay(lang string, schemaMode types.SchemaMode) string {
	if v, ok := types.SchemaModeMap[schemaMode]; ok {
		return v
	}

	return ""
}

func SchemaStatusDisplay(lang string, schemaStatus types.SchemaStatus) string {
	if v, ok := types.SchemaStatusMap[schemaStatus]; ok {
		return v
	}

	return ""
}

func CouponTargetDisplay(lang string, couponTarget types.CouponTarget) string {
	if v, ok := types.CouponTargetMap[couponTarget]; ok {
		return i18n.T(lang, v)
	}

	return ""
}

func SmsTargetDisplay(lang string, smsTarget types.SmsTarget) string {
	if v, ok := types.SmsTargetMap[smsTarget]; ok {
		return i18n.T(lang, v)
	}

	return ""
}
