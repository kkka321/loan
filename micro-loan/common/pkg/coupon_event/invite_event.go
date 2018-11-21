package coupon_event

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/device"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type InviteEvent struct {
}

const (
	accountCoupon = "coupon"
	accountInvite = "invite"
)

type InviteEventParam struct {
	NewAccountId int64
	InviteId     int64
	InviteType   int
}

func (c *InviteEvent) HandleEvent(trigger CouponEventTrigger, data interface{}) {
	logs.Debug("[InviteEvent] HandleEvent trigger:%d, data:%v", trigger, data)

	if trigger != TriggerWebRegister {
		return
	}

	if data == nil {
		logs.Warn("[InviteEvent] HandleEvent data nil data:%v", data)
		return
	}

	param, ok := data.(InviteEventParam)
	if !ok {
		logs.Warn("[InviteEvent] format data error data:%v", data)
		return
	}

	inviteInfo, _ := models.GetSaleInviteById(param.InviteId)

	if param.NewAccountId == 0 || inviteInfo.AccountId == 0 {
		logs.Warn("[InviteEvent] data wrong NewAccountId:%d, InviterId:%d", param.NewAccountId, inviteInfo.AccountId)
		return
	}

	if param.InviteType == types.InviteShare {
		handleShareAccount(param.NewAccountId, inviteInfo.AccountId)
	} else if param.InviteType == types.InviteAnonymous {
		handleAnonymousAccount(param.NewAccountId, inviteInfo.AccountId)
	} else {
		logs.Warn("[InviteEvent] InviteType wrong InviteType:%d", param.InviteType)
	}
}

func handleShareAccount(newAccountId int64, inviterId int64) {
	IncrAccountInviteNum(inviterId, 1, types.InviteNormal)

	ids := distributeCoupon("邀请人", inviterId)
	if len(ids) > 0 {
		IncrAccountCouponNum(inviterId, len(ids), types.InviteNormal)
	}

	distributeCoupon("被邀请邀人", newAccountId)

	info := models.SaleInvite{}
	info.Id, _ = device.GenerateBizId(types.SalesBiz)
	info.InviterId = inviterId
	info.AccountId = newAccountId
	info.Ctime = tools.GetUnixMillis()
	info.Insert()
}

func handleAnonymousAccount(newAccountId int64, inviterId int64) {
	IncrAccountInviteNum(inviterId, 1, types.InviteNormal)

	ids := distributeCoupon("匿名邀请人", inviterId)
	if len(ids) > 0 {
		IncrAccountCouponNum(inviterId, len(ids), types.InviteNormal)
	}

	distributeCoupon("匿名被邀请邀人", newAccountId)

	info := models.SaleInvite{}
	info.Id, _ = device.GenerateBizId(types.SalesBiz)
	info.InviterId = inviterId
	info.AccountId = newAccountId
	info.Ctime = tools.GetUnixMillis()
	info.Insert()
}

func GetInviteEvent() int {
	startDateStr := "2018-10-18 00:00:00"
	endDateStr := "2018-10-29 00:00:00"

	startDate := tools.GetDateParseBackends(startDateStr) * 1000
	endDate := tools.GetDateParseBackends(endDateStr) * 1000

	now := tools.GetUnixMillis()

	if now >= startDate && now <= endDate {
		return types.Invite1018
	} else {
		return types.InviteNormal
	}
}

func getDailyHash(inviteType int) string {
	if inviteType == types.InviteNormal {
		return beego.AppConfig.String("invite_daily_hash")
	} else {
		return beego.AppConfig.String("invite_daily_hash_1018")
	}
}

func getAccountHash(inviteType int) string {
	if inviteType == types.Invite1018 {
		return beego.AppConfig.String("invite_account_hash_1018")
	}

	return beego.AppConfig.String("invite_account_hash")
}

func incrValueFromStorage(hash string, field string, num int) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	storageClient.Do("HINCRBY", hash, field, num)
}

func GetAccountCouponNum(accountId int64, inviteType int) int {
	hash := fmt.Sprintf("%s%d", getAccountHash(inviteType), accountId)

	return getValueFromStorage(hash, accountCoupon, "")
}

func IncrAccountCouponNum(accountId int64, num int, inviteType int) {
	hash := fmt.Sprintf("%s%d", getAccountHash(inviteType), accountId)

	incrValueFromStorage(hash, accountCoupon, num)
}

func GetAccountInviteNum(accountId int64, inviteType int) int {
	hash := fmt.Sprintf("%s%d", getAccountHash(inviteType), accountId)

	return getValueFromStorage(hash, accountInvite, "")
}

func IncrAccountInviteNum(accountId int64, num int, inviteType int) {
	hash := fmt.Sprintf("%s%d", getAccountHash(inviteType), accountId)

	incrValueFromStorage(hash, accountInvite, num)
}

func GetAccountDailyInvite(accountId int64, inviteType int) int {
	now := tools.GetUnixMillis()

	date := tools.MDateMHSDate(now)
	lastDate := tools.MDateMHSDate(now - tools.MILLSSECONDADAY)

	hash := fmt.Sprintf("%s%s", getDailyHash(inviteType), date)
	lastHash := fmt.Sprintf("%s%s", getDailyHash(inviteType), lastDate)
	num := getValueFromStorage(hash, tools.Int642Str(accountId), lastHash)

	return num
}

func IncrAccountDailyInvite(accountId int64, num int, inviteType int) {
	now := tools.GetUnixMillis()

	date := tools.MDateMHSDate(now)

	hash := fmt.Sprintf("%s%s", getDailyHash(inviteType), date)

	incrValueFromStorage(hash, tools.Int642Str(accountId), num)
}

func getValueFromStorage(hash string, field string, delHash string) int {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	num := 0
	hValue, err := storageClient.Do("HGET", hash, field)
	if err != nil {
		logs.Error("[getValueFromStorage] get value error hashName:%s, field:%s, err:%v", hash, field, err)
	} else if hValue == nil {
		storageClient.Do("HSETNX", hash, field, 0)
	} else {
		num, err = tools.Str2Int(string(hValue.([]byte)))
		if err != nil {
			logs.Error("[getValueFromStorage] Str2Int error hashName:%s, field:%s, data:%s, err:%v", hash, field, string(hValue.([]byte)), err)
		}
	}

	if delHash != "" {
		num, _ := storageClient.Do("EXISTS", delHash)
		if num != nil && num.(int64) == 1 {
			storageClient.Do("DEL", delHash)
		}
	}

	return num
}
