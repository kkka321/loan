package feedback

// 反馈标签索引定义
const (
	TagBug        = 1
	TagSuggest    = 2
	TagComplaints = 4
	TagOther      = 2048
)

var tagMap = map[int]string{
	TagBug:        "Bug",
	TagSuggest:    "Suggest",
	TagComplaints: "Complaints",
	TagOther:      "Other",
}

// TagMap 返回标签的map
func TagMap() map[int]string {
	return tagMap
}

// 客户标签
const (
	NoChoice                = 0 // no choice
	CustomerTagsPotential   = 1 // 潜在客户 ：已完成注册，但未进行身份认证客户
	CustomerTagsTarget      = 2 // 目标客户：身份认证通过但未提交过一笔借款申请的客户
	CustomerTagsProspective = 3 // 准客户：未完成首贷，但存在进行中的借款申请（审核中/等待还款/审核拒绝）
	CustomerTagsDeal        = 4 // 成交客户：首贷完成的客户 （完成指已经结清）
	CustomerTagsLoyal       = 5 // 忠实客户：复贷完成的客户
)

var tagUserMap = map[int]string{
	NoChoice:                "Nothing selected",
	CustomerTagsPotential:   "潜在客户",
	CustomerTagsTarget:      "目标客户",
	CustomerTagsProspective: "准客户",
	CustomerTagsDeal:        "成交客户",
	CustomerTagsLoyal:       "忠实客户",
}

func TagsUserMap() map[int]string {
	return tagUserMap
}

//每页条数
const (
	Tag_10  = 1  //10行
	Tag_25  = 2  //25行
	Tag_50  = 4  //50行
	Tag_100 = 8  //100行
	Tag_200 = 16 //200行
)

var tagPageMap = map[int]string{
	Tag_10:  "10",
	Tag_25:  "25",
	Tag_50:  "50",
	Tag_100: "100",
	Tag_200: "200",
}

// TagMap 返回标签的map
func TagPageMap() map[int]string {
	return tagPageMap
}

// 客户标签
const (
	Normal          = 0
	HomePageTags    = 1 //
	SubmitOrderTags = 2 //
)

var tagFloatingMap = map[int]string{
	Normal:          "no choice",
	HomePageTags:    "首页位置",
	SubmitOrderTags: "确认订单位置",
}

func TagsFloatingMap() map[int]string {
	return tagFloatingMap
}
