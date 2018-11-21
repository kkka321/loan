package types

// RiskRegularListName 风控列表名
type RiskRegularListName string

const (
	// FirstLoanRiskRegularList 首贷 反欺诈规则列表
	FirstLoanRiskRegularList RiskRegularListName = "first_loan_risk_regular_list"
	// ReloanWithRandomMarkRiskRegularList 复贷-随机[用户]反欺诈规则列表
	ReloanWithRandomMarkRiskRegularList RiskRegularListName = "reloan_with_random_mark_risk_regular_list"
	// ReloanWithoutRandomMarkRiskRegularList 复贷-非随机[用户]反欺诈规则列表
	ReloanWithoutRandomMarkRiskRegularList RiskRegularListName = "reloan_without_random_mark_risk_regular_list"

	LoanGojekRiskRegularList RiskRegularListName = "gojek_risk_regular_list"
)

const (
	// RiskCtlRegularJustRun 风控规则仅运行
	RiskCtlRegularJustRun = 0
	// RiskCtlRegularReviewed 风控规则在审核列表,且已审核
	RiskCtlRegularReviewed = 1
)
