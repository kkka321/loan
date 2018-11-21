package types

type AssignStatus int

const (
	ExtensionUnAssign AssignStatus = 0
	ExtensionAssign   AssignStatus = 1
)

const (
	SipHasAssigned          = "分机已被分配"
	SipNotAvailable         = "分机不可用"
	UpdateSipNumberInfoFail = "更新分机信息失败"
	UnAssignSuccess         = "取消分配成功"
	UserCanAssignNotFound   = "不存在可分配分机的用户"
	InvalidRequest          = "非法请求"
)
