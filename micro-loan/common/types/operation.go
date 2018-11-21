package types

// banner
const (
	// BannerTypeHomePage 首页
	BannerTypeHomePage = iota
	// BannerTypeInvitePage 邀请好友页
	BannerTypeInvitePage
)

// BannerTypeMap banner类型 map
var BannerTypeMap = map[int]string{
	BannerTypeHomePage:   "首页",
	BannerTypeInvitePage: "邀请好友页",
}

// 广告位管理
const (
	// AdPositionRejectPage 审核拒绝页
	AdPositionRejectPage = iota
	// AdPositionMyAccountPage 我的账户页
	AdPositionMyAccountPage
)

// AdPositionMap 广告位 map
var AdPositionMap = map[int]string{
	AdPositionRejectPage:    "审核拒绝页",
	AdPositionMyAccountPage: "我的账户页",
}
