package coupon_event

import (
	"micro-loan/common/models"
	"micro-loan/common/types"

	"micro-loan/common/dao"
	"micro-loan/common/lib/device"
	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

type InviteV3Event struct {
}

type InviteV3Param struct {
	AccountId int64
	InviteId  int64
	TaskType  types.AccountTask
}

func (c *InviteV3Event) HandleEvent(trigger CouponEventTrigger, data interface{}) {
	logs.Debug("[InviteV3Event] HandleEvent trigger:%d, data:%v", trigger, data)

	if trigger != TriggerInviteV3 {
		return
	}

	if data == nil {
		logs.Warn("[InviteV3Event] HandleEvent data nil data:%v", data)
		return
	}

	param, ok := data.(InviteV3Param)
	if !ok {
		logs.Warn("[InviteV3Event] format data error data:%v", data)
		return
	}

	if param.AccountId == 0 {
		return
	}

	if param.TaskType == types.AccountTaskRegister {
		handleRegisterTask(&param)
	} else if param.TaskType == types.AccountTaskLogin {
		handleLoginTask(&param)
	} else if param.TaskType == types.AccountTaskApply {
		handleApplyTask(&param)
	} else if param.TaskType == types.AccountTaskRepay {
		handleRepayTask(&param)
	}
}

func handleRegisterTask(param *InviteV3Param) {
	task, err := dao.GetAccountTask(param.AccountId, types.AccountTaskRegister)
	if err == nil && task.TaskStatus == types.AccountTaskStatusDone {
		logs.Debug("[handleRegisterTask] task already done param:%+v, data:%+v", param, task)
		return
	}

	inviteInfo, err := models.GetSaleInviteById(param.InviteId)
	if err != nil {
		logs.Debug("[handleRegisterTask] GetSaleInviteById not invite account param:%+v", param)
		return
	}

	IncrAccountInviteNum(inviteInfo.AccountId, 1, types.InviteV3)
	ids := distributeCoupon("好友注册成功", inviteInfo.AccountId)
	IncrAccountCouponNum(inviteInfo.AccountId, len(ids), types.InviteV3)

	info := models.SaleInvite{}
	info.Id, _ = device.GenerateBizId(types.SalesBiz)
	info.InviterId = inviteInfo.AccountId
	info.AccountId = param.AccountId
	info.Ctime = tools.GetUnixMillis()
	info.Insert()

	createOrUpdateAccountTask(param.AccountId, inviteInfo.AccountId, 0, types.AccountTaskRegister)
}

func handleLoginTask(param *InviteV3Param) {
	inviteInfo, err := models.GetSaleInvite(param.AccountId)
	if err != nil || inviteInfo.InviterId == 0 {
		logs.Debug("[handleLoginTask] GetSaleInvite not invite account param:%+v", param)
		return
	}

	task, err := dao.GetAccountTask(param.AccountId, types.AccountTaskLogin)
	if err == nil && task.TaskStatus == types.AccountTaskStatusDone {
		logs.Debug("[handleLoginTask] task already done param:%+v, data:%+v", param, task)
		return
	}

	ids := distributeCoupon("好友首登", inviteInfo.InviterId)
	IncrAccountCouponNum(inviteInfo.InviterId, len(ids), types.InviteV3)

	ids = distributeCoupon("好友首登后奖励被邀请人", param.AccountId)
	id := int64(0)
	if len(ids) > 0 {
		id = ids[0]
	}

	createOrUpdateAccountTask(param.AccountId, inviteInfo.InviterId, id, types.AccountTaskLogin)
}

func handleApplyTask(param *InviteV3Param) {
	inviteInfo, err := models.GetSaleInvite(param.AccountId)
	if err != nil || inviteInfo.InviterId == 0 {
		logs.Debug("[handleApplyTask] GetSaleInvite not invite account param:%+v", param)
		return
	}

	task, err := dao.GetAccountTask(param.AccountId, types.AccountTaskApply)
	if err == nil && task.TaskStatus == types.AccountTaskStatusDone {
		logs.Debug("[handleApplyTask] task already done param:%+v, data:%+v", param, task)
		return
	}

	ids := distributeCoupon("好友首单申请", inviteInfo.InviterId)
	IncrAccountCouponNum(inviteInfo.InviterId, len(ids), types.InviteV3)

	createOrUpdateAccountTask(param.AccountId, inviteInfo.InviterId, 0, types.AccountTaskApply)
}

func handleRepayTask(param *InviteV3Param) {
	inviteInfo, err := models.GetSaleInvite(param.AccountId)
	if err != nil || inviteInfo.InviterId == 0 {
		logs.Debug("[handleRepayTask] GetSaleInvite not invite account param:%+v", param)
		return
	}

	task, err := dao.GetAccountTask(param.AccountId, types.AccountTaskRepay)
	if err == nil && task.TaskStatus == types.AccountTaskStatusDone {
		logs.Debug("[handleRepayTask] task already done param:%+v, data:%+v", param, task)
		return
	}

	ids := distributeCoupon("好友首单还款", inviteInfo.InviterId)
	IncrAccountCouponNum(inviteInfo.InviterId, len(ids), types.InviteV3)

	ids = distributeCoupon("好友首单还款奖励被邀请人", param.AccountId)
	id := int64(0)
	if len(ids) > 0 {
		id = ids[0]
	}

	createOrUpdateAccountTask(param.AccountId, inviteInfo.InviterId, id, types.AccountTaskRepay)
}

func createOrUpdateAccountTask(accountId, inviterId, couponId int64, taskType types.AccountTask) {
	task, err := dao.GetAccountTask(accountId, taskType)
	if err == nil {
		task.TaskStatus = types.AccountTaskStatusDone
		task.CouponId = couponId
		dao.UpdateAccountTask(&task)
	} else {
		task = models.AccountTask{}
		task.TaskStatus = types.AccountTaskStatusDone
		task.AccountId = accountId
		task.InviterId = inviterId
		task.TaskType = taskType
		task.CouponId = couponId
		task.DoneTime = tools.GetUnixMillis()
		dao.AddAccountTask(&task)
	}
}
