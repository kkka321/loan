package ticket

import "github.com/astaxie/beego/logs"

// PollWatchRoleOnlineUser 队列池监察角色中角色上线
// 用于给角色新增用户,或者后台用户请假归来, 或者以后轮班
func PollWatchRoleOnlineUser(roleID, adminUID int64) {
	// 不直接删除 queue重建, 是为维持, 当前分配序列, 确保,高频修改下, 分配均匀

	ticketItems := getWorkTicketItemsByRoleID(roleID)
	logs.Debug("[PollWatchRoleOnlineUser] own ticket items:", ticketItems, "role id:", roleID)
	for _, ticketItem := range ticketItems {
		workerAssignStrategy := getItemWorkerStrategy(ticketItem)
		workerAssignStrategy.UserStartWork(adminUID, roleID)
	}
}

// PollWatchRoleOfflineUser 队列池监察角色中角色离线
// 用于给角色删除后台用户,或者后台用户请假, 或者以后轮班
func PollWatchRoleOfflineUser(roleID, adminUID int64) {
	// 不直接删除 queue重建, 是为维持, 当前分配序列, 确保,高频修改下, 分配均匀
	logs.Debug("[PollWatchRoleOfflineUser] roleID %d, adminUID %d", roleID, adminUID)
	ticketItems := getWorkTicketItemsByRoleID(roleID)
	for _, ticketItem := range ticketItems {
		workerAssignStrategy := getItemWorkerStrategy(ticketItem)
		workerAssignStrategy.UserStopWork(adminUID, roleID)
	}
}
