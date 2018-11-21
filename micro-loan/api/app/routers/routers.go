package routers

import (
	"github.com/astaxie/beego"

	"micro-loan/api/app/controllers"
	"micro-loan/common/cprof"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/ping", &controllers.MainController{}, "*:Ping")

	beego.Router("/api/v1/upload_client_info", &controllers.AccountController{}, "post:SaveClientInfo")

	beego.Router("/api/v1/request_login_auth_code", &controllers.AccountController{}, "post:RequestLoginAuthCode")
	beego.Router("/api/v2/request_login_auth_code", &controllers.AccountController{}, "post:RequestLoginAuthCodeV2")

	beego.Router("/api/v1/request_voice_auth_code", &controllers.AccountController{}, "post:RequestVoiceAuthCode")

	beego.Router("/api/v1/register", &controllers.AccountController{}, "post:Register")

	beego.Router("/api/v1/login", &controllers.AccountController{}, "post:Login")
	beego.Router("/api/v1/login/sms", &controllers.AccountController{}, "post:SmsLogin")
	beego.Router("/api/v1/login/password", &controllers.AccountController{}, "post:PwdLogin")

	beego.Router("/api/v1/sms/verify", &controllers.AccountController{}, "post:SmsVerify")
	beego.Router("/api/v1/password/find", &controllers.AccountController{}, "post:FindPassword")
	beego.Router("/api/v1/password/set", &controllers.AccountController{}, "post:SetPassword")
	beego.Router("/api/v1/password/modify", &controllers.AccountController{}, "post:ModifyPassword")

	beego.Router("/api/v1/auth_report", &controllers.AccountController{}, "post:AuthReport")
	beego.Router("/api/v1/logout", &controllers.AccountController{}, "post:Logout")
	beego.Router("/api/v1/account/info", &controllers.AccountController{}, "post:AccountInfo")
	beego.Router("/api/v2/account/info", &controllers.AccountController{}, "post:AccountInfoV2")
	beego.Router("/api/v1/account/va_all", &controllers.AccountController{}, "post:AccountVas")
	beego.Router("/api/v1/identity/detect", &controllers.AccountController{}, "post:IdentityDetect")
	beego.Router("/api/v1/account/verify", &controllers.AccountController{}, "post:AccountVerify")
	beego.Router("/api/v2/account/verify", &controllers.AccountController{}, "post:AccountVerifyV2")
	beego.Router("/api/v1/account/verify_cl", &controllers.AccountController{}, "post:AccountVerifyCL")
	beego.Router("/api/v1/account/operator_acquire_code", &controllers.AccountController{}, "post:OperatorAcquireCode")
	beego.Router("/api/v1/account/operator_verify_code", &controllers.AccountController{}, "post:OperatorVerifyCode")
	beego.Router("/api/v2/account/operator_verify_code", &controllers.AccountController{}, "post:OperatorVerifyCodeV2")
	beego.Router("/api/v1/account/u/base", &controllers.AccountController{}, "post:UpdateBase")
	beego.Router("/api/v2/account/u/base", &controllers.AccountController{}, "post:UpdateBaseV2")
	beego.Router("/api/v1/account/u/work", &controllers.AccountController{}, "post:UpdateWorkInfo")
	beego.Router("/api/v2/account/u/work", &controllers.AccountController{}, "post:UpdateWorkInfoV2")
	beego.Router("/api/v1/account/u/contact", &controllers.AccountController{}, "post:UpdateContactInfo")
	beego.Router("/api/v1/account/u/other", &controllers.AccountController{}, "post:UpdateOtherInfo")
	beego.Router("/api/v1/account/u/repay_bank", &controllers.AccountController{}, "post:ModifyRepayBank")
	beego.Router("/api/v2/account/u/repay_bank", &controllers.AccountController{}, "post:ModifyRepayBankV2")
	beego.Router("/api/v1/account/visual_account", &controllers.AccountController{}, "post:AccountVAInfo")
	beego.Router("/api/v2/account/visual_account", &controllers.AccountController{}, "post:AccountVAInfoV2")
	beego.Router("/api/v1/account/update/bank/info", &controllers.AccountController{}, "post:UpdateBankInfo")
	beego.Router("/api/v1/account/risk/recheck", &controllers.AccountController{}, "post:RiskReCheck")
	beego.Router("/api/v1/account/risk/recall_hint", &controllers.AccountController{}, "post:PhoneVeiryRefuseRecallHint")
	beego.Router("/api/v1/account/risk/recall", &controllers.AccountController{}, "post:PhoneVeiryRefuseRecall")
	beego.Router("/api/v1/account/npwp/verify", &controllers.AccountController{}, "post:NpwpVerify")
	beego.Router("/api/v1/tongdun/invoke/record", &controllers.AccountController{}, "post:TongdunInvokeRecord")
	beego.Router("/api/v1/account/auth_list", &controllers.AccountController{}, "post:AuthList")

	// 获取配置信息
	beego.Router("/api/v1/config/not_login", &controllers.AccountController{}, "post:ConfigNotLogin")
	beego.Router("/api/v1/config/login", &controllers.AccountController{}, "post:ConfigLogin")

	beego.Router("/api/v1/order/confirm_auth_code", &controllers.LoanOrderController{}, "post:ConfirmLoanAuthCode")
	beego.Router("/api/v1/order/confirm_voice_authcode", &controllers.LoanOrderController{}, "post:ConfirmLoanVoiceAuthCode")
	beego.Router("/api/v1/order/repeat_auth_code", &controllers.LoanOrderController{}, "post:RepeatLoanAuthCode")
	beego.Router("/api/v2/order/repeat_auth_code", &controllers.LoanOrderController{}, "post:RepeatLoanAuthCodeV2")
	beego.Router("/api/v1/order/repeat_verify", &controllers.LoanOrderController{}, "post:RepeatLoanVerify")
	beego.Router("/api/v2/order/repeat_verify", &controllers.LoanOrderController{}, "post:RepeatLoanVerifyV2")
	beego.Router("/api/v1/order/confirm", &controllers.LoanOrderController{}, "post:Confirm")
	beego.Router("/api/v1/order/current", &controllers.LoanOrderController{}, "post:Current")
	beego.Router("/api/v1/order/all", &controllers.LoanOrderController{}, "post:All")
	beego.Router("/api/v1/order/re_loan_upload_photo", &controllers.ReLoanController{}, "post:ReLoanUploadHandHeldInPhoto")
	beego.Router("/api/v1/order/extension_trialcal", &controllers.LoanOrderController{}, "post:ExtensionTrialCal")
	beego.Router("/api/v1/order/extension_confirm", &controllers.LoanOrderController{}, "post:ExtensionConfirm")
	beego.Router("/api/v1/order/home_order", &controllers.LoanOrderController{}, "post:HomeOrder")
	beego.Router("/api/v1/order/payment_voucher", &controllers.LoanOrderController{}, "post:PaymentVoucher")

	beego.Router("/api/v1/feedback/create", &controllers.FeedbackController{}, "post:Create")
	beego.Router("/api/v1/product/info", &controllers.ProductController{}, "post:ProductInfoV1")

	beego.Router("/api/v2/identity/detect", &controllers.AccountController{}, "post:IdentityDetectV2")
	beego.Router("/api/v2/order/create", &controllers.LoanOrderController{}, "post:CreateOrderV2")
	beego.Router("/api/v2/order/current", &controllers.LoanOrderController{}, "post:CurrentV2")
	beego.Router("/api/v2/order/all", &controllers.LoanOrderController{}, "post:AllV2")

	beego.Router("/api/v3/identity/detect", &controllers.AccountController{}, "post:IdentityDetectV3")
	beego.Router("/api/v3/identity/verify", &controllers.AccountController{}, "post:IdentityVerifyV3")
	// 修改身份证对应手机号（解绑之前手机号+绑定当前手机号）
	beego.Router("/api/v1/identity/modify", &controllers.AccountController{}, "post:IdentityModify")

	beego.Router("/sms/callback/delivery/:smsEncryptKey:string", &controllers.SmsCallbackController{}, "*:Delivery")

	beego.Router("/xendit/virtual_account_callback/create", &controllers.XenditCallbackController{}, "*:VirtualAccountCreate")
	beego.Router("/xendit/disburse_fund_callback/create", &controllers.XenditCallbackController{}, "*:DisburseFundCreate")
	beego.Router("/xendit/fva_receive_payment_callback/create", &controllers.XenditCallbackController{}, "*:FVAReceivePaymentCreate")
	beego.Router("/xendit/market_receive_payment_callback/create", &controllers.XenditCallbackController{}, "*:MarketReceivePaymentCreate")
	beego.Router("/xendit/fix_payment_code_callback/create", &controllers.XenditCallbackController{}, "*:FixPaymentcodeCreate")

	//callback
	beego.Router("/bluepay/callback", &controllers.BluePayCallbackController{}, "*:CallBack")
	beego.Router("/tongdun/callback", &controllers.TongdunCallbackController{}, "*:CallBack")

	beego.Router("/api/v1/order/loan_quota", &controllers.LoanOrderController{}, "*:LoanQuota")
	beego.Router("/api/v2/order/loan_quota", &controllers.LoanOrderController{}, "*:LoanQuotaV2")
	beego.Router("/api/v2/order/confirm", &controllers.LoanOrderController{}, "post:ConfirmV2")
	beego.Router("/api/v3/order/confirm", &controllers.LoanOrderController{}, "post:ConfirmV3")
	beego.Router("/api/v4/order/confirm", &controllers.LoanOrderController{}, "post:ConfirmV4")
	beego.Router("/api/v1/order/xendit_paymentcode", &controllers.LoanOrderController{}, "post:XenditPaymentCode")

	beego.Router("/api/v1/message/new", &controllers.MessageController{}, "post:New")
	beego.Router("/api/v1/message/all", &controllers.MessageController{}, "post:All")
	beego.Router("/api/v1/message/confirm", &controllers.MessageController{}, "post:Confirm")

	//dot
	beego.Router("/api/v1/dot/dot1", &controllers.DotController{}, "post:Dot1")
	beego.Router("/api/v1/dot/dot2", &controllers.DotController{}, "post:Dot2")

	// appsflyer CallBack
	// appsflyer install callback url
	beego.Router("/appsflyer/callback/install", &controllers.AppsflyerController{}, "*:Install")

	//entrust 勤为
	beego.Router("/outsource/v1/case/sync/base_info", &controllers.EntrustController{}, "post:BaseInfo")
	beego.Router("/outsource/v1/case/sync/processed_callback", &controllers.EntrustController{}, "post:ProcessedCallback")
	beego.Router("/outsource/v1/case/sync/repaylist", &controllers.EntrustController{}, "post:GetRepayList")
	beego.Router("/outsource/v1/case/sync/repay_status", &controllers.EntrustController{}, "post:RepayStatus")
	beego.Router("/outsource/v1/case/contacts", &controllers.EntrustController{}, "post:Contacts")
	beego.Router("/outsource/v1/case/rolltc", &controllers.EntrustController{}, "post:RollTC")
	beego.Router("/outsource/v1/case/spayment_code", &controllers.EntrustController{}, "post:SPaymentCode")

	//doku第三方支付
	beego.Router("/doku/virtual_account_callback/create", &controllers.DoKuCallbackController{}, "*:VirtualAccountCreate")
	beego.Router("/doku/disburse_fund_callback/create", &controllers.DoKuCallbackController{}, "*:DisburseFundCreate")
	beego.Router("/doku/fva_receive_payment_callback/create", &controllers.DoKuCallbackController{}, "*:FVAReceivePaymentCreate")
	beego.Router("/doku/identify/create", &controllers.DoKuCallbackController{}, "*:IdentifyCreate")

	// web api
	beego.Router("/webapi/v1/is_login", &controllers.WebApiController{}, "post:IsLogin")
	beego.Router("/webapi/v1/request_login_auth_code", &controllers.WebApiController{}, "post:RequestLoginAuthCode")
	beego.Router("/webapi/v1/login", &controllers.WebApiController{}, "post:Login")

	beego.Router("/webapi/v2/request_login_auth_code", &controllers.WebApiController{}, "post:RequestLoginAuthCodeV2")
	beego.Router("/webapi/v2/login", &controllers.WebApiController{}, "post:LoginV2")

	// voip callback
	beego.Router("/extension/sip_bill_msg", &controllers.VoipCallbackController{}, "*:SipBillMessageCB")

	beego.Router("/api/v1/coupon/list", &controllers.CouponController{}, "*:List")
	beego.Router("/api/v1/coupon/active", &controllers.CouponController{}, "*:Active")
	beego.Router("/api/v1/coupon/has_new", &controllers.CouponController{}, "*:HasNew")
	beego.Router("/api/v1/coupon/mark_new", &controllers.CouponController{}, "*:MarkNew")

	beego.Router("/api/v2/coupon/list", &controllers.CouponController{}, "*:ListV2")
	beego.Router("/api/v2/coupon/active", &controllers.CouponController{}, "*:ActiveV2")

	// new loan flow
	beego.Router("/api/loan_flow/v1/register", &controllers.AccountController{}, "post:RegisterTwo")
	beego.Router("/api/loan_flow/v1/login", &controllers.AccountController{}, "post:LoginTwo")
	beego.Router("/api/loan_flow/v1/login/sms", &controllers.AccountController{}, "post:SmsLoginTwo")
	beego.Router("/api/loan_flow/v1/login/password", &controllers.AccountController{}, "post:PwdLoginTwo")
	beego.Router("/api/loan_flow/v1/account/info", &controllers.AccountController{}, "post:AccountInfoTwo")
	beego.Router("/api/loan_flow/v1/account/verify", &controllers.AccountController{}, "post:AccountVerifyTwo")
	beego.Router("/api/loan_flow/v2/account/verify", &controllers.AccountController{}, "post:AccountVerifyTwoV2")
	beego.Router("/api/loan_flow/v1/account/verify_cl", &controllers.AccountController{}, "post:AccountVerifyCLTwo")
	beego.Router("/api/loan_flow/v1/account/operator_verify_code", &controllers.AccountController{}, "post:OperatorVerifyCodeTwo")
	beego.Router("/api/loan_flow/v1/account/u/base", &controllers.AccountController{}, "post:UpdateBaseTwo")
	beego.Router("/api/loan_flow/v1/account/u/work", &controllers.AccountController{}, "post:UpdateWorkInfoTwo")
	beego.Router("/api/loan_flow/v1/account/u/contact", &controllers.AccountController{}, "post:UpdateContactInfoTwo")
	beego.Router("/api/loan_flow/v1/account/u/other", &controllers.AccountController{}, "post:UpdateOtherInfoTwo")
	beego.Router("/api/loan_flow/v1/order/create", &controllers.LoanOrderController{}, "post:CreateOrderTwo")
	beego.Router("/api/loan_flow/v1/identity/detect", &controllers.AccountController{}, "post:IdentityDetectTwo")
	beego.Router("/api/loan_flow/v1/identity/verify", &controllers.AccountController{}, "post:IdentityVerifyTwo")
	beego.Router("/api/loan_flow/v1/order/confirm", &controllers.LoanOrderController{}, "post:ConfirmTwo")
	beego.Router("/api/loan_flow/v2/identity/detect", &controllers.AccountController{}, "post:IdentityDetectTwoV2")

	beego.Router("/api/sales/v1/invite_info", &controllers.SalesController{}, "post:InviteInfo")
	beego.Router("/api/sales/v1/invite", &controllers.SalesController{}, "post:Invite")
	beego.Router("/api/sales/v1/invite_list", &controllers.SalesController{}, "post:InviteList")

	beego.Router("/t/:u", &controllers.ShortUrlController{}, "*:Access")

	beego.Router("/api/advertisement/v1/get", &controllers.AdvertisementController{}, "post:GetAdvertisement")
	beego.Router("/api/banner/v1/get", &controllers.BannerController{}, "post:GetBanners")
	beego.Router("/api/adposition/v1/get", &controllers.AdPositionController{}, "post:GetAdPosition")

	// pprof
	beego.Router("/debug/pprof", &cprof.ProfController{}, "*:Get")
	beego.Router(`/debug/pprof/:pp([\w]+)`, &cprof.ProfController{}, "*:Get")

	beego.Router("/api/activity/v1/get_popoversor", &controllers.ActivityController{}, "post:GetPopoversor")
	beego.Router("/api/activity/v1/get_floating", &controllers.ActivityController{}, "post:GetFloating")

	beego.Router("/api/v1/log/boot", &controllers.LogController{}, "post:Boot")
}
