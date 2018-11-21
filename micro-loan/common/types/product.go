package types

const ProductFeeBase = 10000

//产品类型
type ProductTypeEunm int

const (
	ProductTypeFirst    ProductTypeEunm = 1 // 1、 首贷
	ProductTypeReLoan   ProductTypeEunm = 2 // 2、 复贷
	ProductTypeRollLoan ProductTypeEunm = 3 // 3、 展期
)

var productTypeMap = map[ProductTypeEunm]string{
	ProductTypeFirst:    "首贷",
	ProductTypeReLoan:   "复贷",
	ProductTypeRollLoan: "展期",
}

// GetProductTypeMap 导出函数
func GetProductTypeMap() map[ProductTypeEunm]string {
	return productTypeMap
}

//产品状态
type ProductStatusEunm int

const (
	ProductStatusNever   ProductStatusEunm = 0 // 0、未使用过
	ProductStatusInValid ProductStatusEunm = 1 // 1、下架
	ProductStatusValid   ProductStatusEunm = 2 // 2、上架
)

var productStatusMap = map[ProductStatusEunm]string{
	ProductStatusNever:   "未生效",
	ProductStatusInValid: "下架",
	ProductStatusValid:   "上架",
}

// GetProductStatusMap 导出函数
func GetProductStatusMap() map[ProductStatusEunm]string {
	return productStatusMap
}

//利息收取时间类型
type ProductChargeInterestTypeEnum int

const (
	ProductChargeInterestTypeHeadCut  ProductChargeInterestTypeEnum = 0 // 0、放款时扣取
	ProductChargeInterestTypeByStages ProductChargeInterestTypeEnum = 1 // 1、分期还
)

var productChargeInterestTypeMap = map[ProductChargeInterestTypeEnum]string{
	ProductChargeInterestTypeHeadCut:  "放款时扣取",
	ProductChargeInterestTypeByStages: "分期还",
}

// GetPsroductChargeInterestTypeMap 导出函数
func GetProductChargeInterestTypeMap() map[ProductChargeInterestTypeEnum]string {
	return productChargeInterestTypeMap
}

//费用收取时间类型
// type ProductChargeFeeTypeEnum int
//
// const (
// 	ProductChargeFeeTypeHeadCut  ProductChargeFeeTypeEnum = 0 // 0、放款时扣取
// 	ProductChargeFeeTypeByStages ProductChargeFeeTypeEnum = 1 // 1、分期还
//
// )

var productChargeFeeTypeMap = map[int]string{
	ProductChargeFeeInterestBefore: "放款时扣取",
	ProductChargeFeeInterestAfter:  "分期还",
}

// GetProductStatusMap 导出函数
func GetProductChargeFeeTypeMap() map[int]string {
	return productChargeFeeTypeMap
}

// 还款方式
type ProductRepayTypeEunm int

const (
	ProductRepayTypeOnce                       ProductRepayTypeEunm = 0 // 0、一次性还本付息
	ProductRepayTypeByMonth                    ProductRepayTypeEunm = 1 // 1、按月付息到期还本
	ProductRepayTypeAverageCapitalPlusInterest ProductRepayTypeEunm = 2 // 2、等额本息
	ProductRepayTypeNoInterest                 ProductRepayTypeEunm = 3 // 3、等本等息
)

var productRepayTypeMap = map[ProductRepayTypeEunm]string{
	ProductRepayTypeOnce:                       "一次性还本付息",
	ProductRepayTypeByMonth:                    "按月付息到期还本",
	ProductRepayTypeAverageCapitalPlusInterest: "等额本息",
	ProductRepayTypeNoInterest:                 "等本等息",
}

// GetProductRepayTypeMap 导出函数
func GetProductRepayTypeMap() map[ProductRepayTypeEunm]string {
	return productRepayTypeMap
}

//期限单位
type ProductPeriodEunm int

const (
	ProductPeriodOfDay        ProductPeriodEunm = 1  // 0、日
	ProductPeriodOfWeek       ProductPeriodEunm = 7  // 1、周
	ProductPeriodOfWeekDouble ProductPeriodEunm = 14 // 2、双周
	ProductPeriodOfMonth      ProductPeriodEunm = 30 // 3、月
)

var productPeriodMap = map[ProductPeriodEunm]string{
	ProductPeriodOfDay:        "日",
	ProductPeriodOfWeek:       "周",
	ProductPeriodOfWeekDouble: "双周",
	ProductPeriodOfMonth:      "月",
}

// GetProductRepayTypeMap 导出函数
func GetProductPeriodMap() map[ProductPeriodEunm]string {
	return productPeriodMap
}

//取整方式
type ProductCeilWayEunm int

const (
	ProductCeilWayUp ProductCeilWayEunm = 0 // 0、向上取整
	ProductCeilWayNo ProductCeilWayEunm = 1 // 1、不取整

)

var productCeilWayMap = map[ProductCeilWayEunm]string{
	ProductCeilWayUp: "向上取整",
	ProductCeilWayNo: "不取整",
}

// GetProductCeilWayMap 导出函数
func GetProductCeilWayMap() map[ProductCeilWayEunm]string {
	return productCeilWayMap
}

//取整单位
type ProductCeilWayUnitEunm int

const (
	ProductCeilWayUnitOne ProductCeilWayUnitEunm = 1    //
	ProductCeilWayUnitTen ProductCeilWayUnitEunm = 10   //
	ProductCeilWayUnitHun ProductCeilWayUnitEunm = 100  //
	ProductCeilWayUnitTho ProductCeilWayUnitEunm = 1000 //
)

var productCeilWayUnitMap = map[ProductCeilWayUnitEunm]string{
	ProductCeilWayUnitOne: "1",
	ProductCeilWayUnitTen: "10",
	ProductCeilWayUnitHun: "100",
	ProductCeilWayUnitTho: "1000",
}

// GetProductCeilWayMap 导出函数
func GetProductCeilWayUnitMap() map[ProductCeilWayUnitEunm]string {
	return productCeilWayUnitMap
}

//操作类型
type ProductOptTypeEunm int

const (
	ProductOptTypeCreate ProductOptTypeEunm = 1 // 1、 创建产品
	ProductOptTypeModify ProductOptTypeEunm = 2 // 2、 修改产品
	ProductOptTypeUp     ProductOptTypeEunm = 3 // 3、 上架产品
	ProductOptTypeDown   ProductOptTypeEunm = 4 // 4、 下架产品
)

var productOptTypMap = map[ProductOptTypeEunm]string{
	ProductOptTypeCreate: "创建产品",
	ProductOptTypeModify: "修改备注内容",
	ProductOptTypeUp:     "上架产品",
	ProductOptTypeDown:   "下架产品",
}

// GetProductOptTypMap 导出函数
func GetProductOptTypMap() map[ProductOptTypeEunm]string {
	return productOptTypMap
}

//客户是否可见
type CustomerVisibleTypeEunm int

const (
	CustomerVisibleTypeInVisible CustomerVisibleTypeEunm = 0 // 0、 不可见
	CustomerVisibleTypeVisible   CustomerVisibleTypeEunm = 1 // 1、 可见
)

var customerVisibleTypeMap = map[CustomerVisibleTypeEunm]string{
	CustomerVisibleTypeInVisible: "不可见",
	CustomerVisibleTypeVisible:   "可见",
}

// customerVisibleTypeMap 导出函数
func GetCustomerVisibleTypeMap() map[CustomerVisibleTypeEunm]string {
	return customerVisibleTypeMap
}

// 试算
type ProductTrialCalcIn struct {
	ID           int64
	Loan         int64
	Amount       int64
	Period       int
	LoanDate     string
	CurrentDate  string
	RepayDate    string
	RepayedTotal int64
}

type ProductTrialCalcResult struct {
	ID                       int64  // id
	Name                     string // 产品名称
	NumberOfPeriods          int    // 期数
	RepayDateShould          int64  // 应还日期
	Loan                     int64  // 放款金额
	OverdueDays              int    // 逾期天数
	RepayStatus              int    // 还款状态
	RepayTotalShould         int64  // 应还总额
	RepayAmountShould        int64  // 应还本金
	RepayInterestShould      int64  // 应还利息
	RepayFeeShould           int64  // 应还服务费
	RepayGraceInterestShould int64  // 应还款限期利息
	RepayPenaltyShould       int64  // 应还罚息
	ForfeitPenalty           int64  // 应还滞纳金
	RepayedDate              int64  // 实还日期
	RepayedTotal             int64  // 已还总额
	RepayedAmount            int64  // 已还本金
	RepayedInterest          int64  // 已还利息
	RepayedGraceInterest     int64  // 已还款限期利息
	RepayedFee               int64  // 已还服务费
	RepayedPenalty           int64  // 已还罚息
	RepayedForfeitPenalty    int64  // 已还滞纳金
}

type ProductTrialCalcStatus struct {
	LoanDate          int64
	CurrentDate       int64
	RepayDate         int64
	RepayDateShould   int64
	GraceInterestDate int64
	InterestTotal     int64
	FeeTotal          int64
	LoanTotal         int64
	AmountTotal       int64 // 应还本金
	AmountRepayed     int64 // 已还本金
	RepayOrder        []string
}

const (
	ProductOrderAmount         = "p"
	ProductOrderInterest       = "i"
	ProductOrderFee            = "s"
	ProductOrderGraceInterest  = "g"
	ProductOrderPenalty        = "f"
	ProductOrderForfeitPenalty = "z"

//	DayPenaltyRate             = 100
)
