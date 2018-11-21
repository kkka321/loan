package types

// 业务编号
type BizSN int

const (
	AccountSystem       BizSN = 1  // 帐户系统
	OrderSystem         BizSN = 2  // 订单系统
	FinancialProduct    BizSN = 3  // 金融产品
	AccessTokenBiz      BizSN = 4  // token
	FaceidBiz           BizSN = 5  // faceid 接口调用
	AdvanceBiz          BizSN = 6  // advance 接口调用
	UploadResourceBiz   BizSN = 7  // 上传资源
	UserEAccountBiz     BizSN = 8  //电子账户
	UserETransBiz       BizSN = 9  //电子账户交易明细
	PaymentBiz          BizSN = 10 //放款打款
	RepayPlanBiz        BizSN = 11 //还款计划
	SmsVerifyCodeBiz    BizSN = 12 //短信验证码
	VoipCallRecordBiz   BizSN = 13 //Voip 通话记录
	DokuTransIdMerchant BizSN = 14 //DoKu VA inquiry for Receive Payment
	RefundBiz           BizSN = 15 //退款
	ReduceRecordBiz     BizSN = 16 //减免
	DisburseRecordBiz   BizSN = 17 //放款
	CouponBiz           BizSN = 18 //优惠券
    SalesBiz            BizSN = 19 //活动码
)
