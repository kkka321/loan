package coupon_event

import (
	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/device"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type Invite1018Event struct {
}

type InviteEvent1018Param struct {
	NewAccountId int64
	InviteId     int64
	InviteType   int
}

func (c *Invite1018Event) HandleEvent(trigger CouponEventTrigger, data interface{}) {
	logs.Debug("[Invite1018Event] HandleEvent trigger:%d, data:%v", trigger, data)

	if trigger != TriggerWeb1018Register {
		return
	}

	if data == nil {
		logs.Warn("[Invite1018Event] HandleEvent data nil data:%v", data)
		return
	}

	param, ok := data.(InviteEvent1018Param)
	if !ok {
		logs.Warn("[Invite1018Event] format data error data:%v", data)
		return
	}

	inviteInfo, _ := models.GetSaleInviteById(param.InviteId)

	if param.NewAccountId == 0 || inviteInfo.AccountId == 0 {
		logs.Warn("[Invite1018Event] data wrong NewAccountId:%d, InviterId:%d", param.NewAccountId, inviteInfo.AccountId)
		return
	}

	if param.InviteType == types.InviteShare {
		handle1018ShareAccount(param.NewAccountId, inviteInfo.AccountId)
	} else if param.InviteType == types.InviteAnonymous {
		handle1018AnonymousAccount(param.NewAccountId, inviteInfo.AccountId)
	} else {
		logs.Warn("[Invite1018Event] InviteType wrong InviteType:%d", param.InviteType)
	}
}

func handle1018ShareAccount(newAccountId int64, inviterId int64) {
	IncrAccountInviteNum(inviterId, 1, types.Invite1018)
	inviteNum := GetAccountInviteNum(inviterId, types.Invite1018)

	logs.Info("[handle1018ShareAccount] newAccountId:%d, inviterId:%d, num:%d", newAccountId, inviterId, inviteNum)

	if inviteNum != 0 && inviteNum%3 == 0 {
		ids := distributeCoupon("邀请人-qnj", inviterId)
		if len(ids) > 0 {
			IncrAccountCouponNum(inviterId, len(ids), types.Invite1018)
		}
	}

	distributeCoupon("被邀请人-qnj", newAccountId)

	info := models.SaleInvite{}
	info.Id, _ = device.GenerateBizId(types.SalesBiz)
	info.InviterId = inviterId
	info.AccountId = newAccountId
	info.Ctime = tools.GetUnixMillis()
	info.Insert()
}

func handle1018AnonymousAccount(newAccountId int64, inviterId int64) {
	IncrAccountInviteNum(inviterId, 1, types.Invite1018)
	inviteNum := GetAccountInviteNum(inviterId, types.Invite1018)

	logs.Info("[handle1018AnonymousAccount] newAccountId:%d, inviterId:%d, num:%d", newAccountId, inviterId, inviteNum)

	if inviteNum != 0 && inviteNum%3 == 0 {
		ids := distributeCoupon("匿名邀请人-qnj", inviterId)
		if len(ids) > 0 {
			IncrAccountCouponNum(inviterId, len(ids), types.Invite1018)
		}
	}

	distributeCoupon("匿名被邀请人-qnj", newAccountId)

	info := models.SaleInvite{}
	info.Id, _ = device.GenerateBizId(types.SalesBiz)
	info.InviterId = inviterId
	info.AccountId = newAccountId
	info.Ctime = tools.GetUnixMillis()
	info.Insert()
}
