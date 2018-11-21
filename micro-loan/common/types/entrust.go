package types

// 逾期天数>=12的可以被委外
const EntrustDay = 12

const (
	AgreeYes int = 1
	AgreeNo  int = 0
)

var agreeEnumMap = map[int]string{
	AgreeYes: "同意",
	AgreeNo:  "拒绝",
}

func AgreeEnumMap() map[int]string {
	return agreeEnumMap
}

//催收类型

const (
	UrgeTypeEntrust int = 1
	UrgeTypeSelf    int = 0
)

var urgeTypeEnumMap = map[int]string{
	UrgeTypeEntrust: "委外",
	UrgeTypeSelf:    "自催",
}

func UrgeTypeEnumMap() map[int]string {
	return urgeTypeEnumMap
}

//委外状态

const (
	EntrustYes int = 1
	EntrustNo  int = 0
)

var entrustEnumMap = map[int]string{
	EntrustYes: "已委外",
	EntrustNo:  "委外中",
}

func EntrustEnumMap() map[int]string {
	return entrustEnumMap
}

//委外公司
const (
	QinweiGroup string = "qinweigroup"
	JuceGroup   string = "jucegroup"
	DachuiGroup string = "dachuigroup"
	MBAGroup    string = "mbagroup"
)

var entrustCompanyMap = map[string]string{
	QinweiGroup: "qinwei",
	JuceGroup:   "juce",
	DachuiGroup: "dachui",
	MBAGroup:    "MBA",
}

func EntrustCompanyMap() map[string]string {
	return entrustCompanyMap
}
