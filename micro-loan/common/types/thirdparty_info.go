package types

const (
	ChargeForCall        int = 1
	ChargeForCallSuccess int = 2
	ChargeForHit         int = 3
	ChargeForFree        int = 4
)

var chargeTypeMap = map[int]string{
	ChargeForCall:        "调用收费",
	ChargeForCallSuccess: "调用成功收费",
	ChargeForHit:         "命中收费",
	ChargeForFree:        "不收费",
}

func ChargeTypeMap() map[int]string {
	return chargeTypeMap
}

const (
	HTTPCodeSuccess = 200
	CodeSuccess     = "SUCCESS"
)

const (
	CallReaultSuccess = 1
	CallReaultHit     = 2
	CallReaultFailed  = 3
)

const (
	RecordTypeSingle    = 1
	RecordTypeTotal     = 2
	RecordTypeTotalName = ""
)

type Thirdparty struct {
	Code       int
	PayOutApiS []string // 放款统计接口
	PayInApiS  []string // 还款统计接口
}

var ThirdpartyNameCodeMap = map[string]Thirdparty{
	"xendit": {Xendit,
		[]string{
			"/xendit/disburse_fund_callback/create",
		},
		[]string{
			"/xendit/fva_receive_payment_callback/create",
			"/xendit/market_receive_payment_callback/create",
		},
	},
	"doku": {DoKu,
		[]string{
			"https://kirimdoku.com/v2/api/cashin/remit",
		},
		[]string{
			"/doku/fva_receive_payment_callback/create",
		},
	},
}

const EmptyOrmStr = "<QuerySeter> no row found"

var tableNameMap = map[int]string{
	1:  "201807",
	2:  "201808",
	3:  "201809",
	4:  "201810",
	5:  "201811",
	6:  "201812",
	7:  "201901",
	8:  "201902",
	9:  "201903",
	10: "201904",
	11: "201905",
	12: "201906",
	13: "201907",
	14: "201908",
	15: "201909",
	16: "201910",
	17: "201911",
	18: "201912",
}

func TableNameMap() map[int]string {
	return tableNameMap
}
