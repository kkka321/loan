package sales

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/dao"
	"micro-loan/common/i18n"
	"micro-loan/common/lib/device"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/coupon_event"
	"micro-loan/common/pkg/schema_task"
	"micro-loan/common/pkg/short_url"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

var (
	hostUrl   = ""
	webUrl    = ""
	mobileSet = ""
)

const (
	urlInvite = "invite"
	urlOp     = "op"
	urlType   = "type"
)

const (
	//已到了今天匿名邀请上限，去立即邀请看看
	maxAccountLimit = "Telah mencapai batas undangan anonim hari ini, saya akan mengundang anda untuk melihatnya"

	//今天匿名邀请活动参与人数已满，明天早些来，去立即邀请看看
	maxInviteLimit = "Hari ini batas limit mengundang anonim telah penuh, Besok pagi anda bisa kembali melihat dan mengundang."

	//有人匿名送您大红包
	anonymousMsg = "Seseorang mengirimi Anda amplop merah besar secara anonim"

	shareMsg     = ""
	share1018Msg = "Rupiah Cepat:Selamat, nomor Anda mendapatkan dana sebesar ini dari teman Anda, klik untuk info lebih lanjut"
)

func init() {
	hostUrl = beego.AppConfig.String("host_url")

	webUrl = beego.AppConfig.String("invite_web_host")

	mobileSet = beego.AppConfig.String("invite_mobile")
}

func GetInviteInfo(accountId int64, inviteType int) models.SaleInvite {
	logs.Info("[GetInviteInfo] begin accountId:%d, inviteType:%d", accountId, inviteType)

	data, err := models.GetSaleInvite(accountId)

	id := data.Id
	if err != nil {
		id, _ = device.GenerateBizId(types.SalesBiz)
	}

	shareUrl := short_url.GenerateShortUrl(fmt.Sprintf("%s?%s=%d&%s=%d&%s=%d", webUrl, urlInvite, id, urlOp, types.InviteShare, urlType, inviteType), hostUrl)
	anonymousUrl := short_url.GenerateShortUrl(fmt.Sprintf("%s?%s=%d&%s=%d&%s=%d", webUrl, urlInvite, id, urlOp, types.InviteAnonymous, urlType, inviteType), hostUrl)

	if shareUrl == "" {
		shareUrl = fmt.Sprintf("%s?%s=%d&%s=%d", webUrl, urlInvite, id, urlOp, types.InviteShare)
		logs.Error("[GetInviteInfo] GenerateShortUrl wrong use origin url inviteId:%d, url:%s", id, shareUrl)
	}

	if err != nil {
		data = models.SaleInvite{}
		data.Id = id
		data.AccountId = accountId
		data.Ctime = tools.GetUnixMillis()
		data.ShareUrl = shareUrl
		data.AnonymousUrl = anonymousUrl
		data.Insert()

		return data
	}

	if data.ShareUrl == shareUrl && data.AnonymousUrl == anonymousUrl {
		return data
	}

	logs.Info("[GetInviteInfo] update short url inviteId:%d, urlId:%d", id, data.Id)

	data.ShareUrl = shareUrl
	data.AnonymousUrl = anonymousUrl
	data.Utime = tools.GetUnixMillis()
	data.Update()

	return data
}

func SendInviteMessage(accountId int64, clientTag int, mobiles []string) string {
	inviteType := types.InviteNormal
	if clientTag == 1 {
		inviteType = coupon_event.GetInviteEvent()
	}

	logs.Debug("[SendInviteMessage] begin accountId:%d, mobiles:%v, inviteType:%d", accountId, mobiles, inviteType)

	if len(mobiles) == 0 {
		return ""
	}

	info := GetInviteInfo(accountId, inviteType)
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	num := coupon_event.GetAccountDailyInvite(accountId, inviteType)
	if num >= 8 {
		logs.Warn("[SendInviteMessage] account daily count > 8 accountId:%d, count:%d", accountId, num)
		return maxAccountLimit
	}

	totalLimit, _ := config.ValidItemInt("invite_daily_limit")
	totalNum1 := coupon_event.GetAccountDailyInvite(0, types.InviteNormal)
	totalNum2 := coupon_event.GetAccountDailyInvite(0, types.Invite1018)
	totalNum := totalNum1 + totalNum2
	if totalLimit > 0 && totalNum >= totalLimit {
		logs.Warn("[SendInviteMessage] total daily count > %d accountId:%d, count1:%d, count2:%d", totalLimit, accountId, totalNum1, totalNum2)
		return maxInviteLimit
	}

	sendCount := 0
	for _, v := range mobiles {
		qVal, err := storageClient.Do("SADD", mobileSet, v)
		if err == nil && 0 == qVal.(int64) {
			logs.Info("[SendInviteMessage] mobile already send msg accountId:%d, set:%s, mobile:%s", accountId, mobileSet, v)
			continue
		}

		_, err = models.OneAccountBaseByMobile(v)
		if err == nil {
			logs.Info("[SendInviteMessage] mobile registered accountId:%d, set:%s, mobile:%s", accountId, mobileSet, v)
			continue
		}

		//msg := fmt.Sprintf("%s %s", anonymousMsg, info.AnonymousUrl)
		//sms.Send(types.ServiceSales, v, msg, accountId)
		param := make(map[string]interface{})
		param["url"] = info.AnonymousUrl
		schema_task.SendBusinessMsg(types.SmsTargetInvite, types.ServiceSales, v, param)

		sendCount++
		if num+sendCount >= 8 {
			break
		}

		if totalLimit > 0 && totalNum+sendCount >= totalLimit {
			break
		}
	}

	logs.Debug("[SendInviteMessage] send sms done accountId:%d, count:%d", accountId, sendCount)

	if sendCount > 0 {
		coupon_event.IncrAccountDailyInvite(accountId, sendCount, inviteType)
		coupon_event.IncrAccountDailyInvite(0, sendCount, inviteType)
	}

	return ""
}

func CheckInviteInfoRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

func CheckInviteRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"mobile_list": true,
	}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

func CheckInviteListRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

func QueryAccountInviteInfo(accountId int64, clientTag int, data map[string]interface{}) {
	inviteType := types.InviteNormal
	if clientTag == 1 {
		inviteType = coupon_event.GetInviteEvent()
	}

	info := GetInviteInfo(accountId, inviteType)

	couponNum := coupon_event.GetAccountCouponNum(accountId, inviteType)

	inviteNum := coupon_event.GetAccountInviteNum(accountId, inviteType)

	data["invite_num"] = inviteNum
	data["coupon_num"] = couponNum
	data["link_url"] = info.ShareUrl
	data["invite_type"] = inviteType
	if inviteType == types.Invite1018 {
		data["link_msg"] = share1018Msg
	} else {
		data["link_msg"] = ""
	}
}

func QueryAccountInviteList(accountId int64, clientTag int, data map[string]interface{}) {
	inviteType := types.InviteNormal
	if clientTag == 1 {
		inviteType = coupon_event.GetInviteEvent()
	}

	couponNum := coupon_event.GetAccountCouponNum(accountId, inviteType)

	inviteNum := coupon_event.GetAccountInviteNum(accountId, inviteType)

	data["invite_num"] = inviteNum
	data["coupon_num"] = couponNum

	list, err := dao.QueryAccountTaskByInviter(accountId)
	if err != nil {
		logs.Debug("[QueryAccountInviteList] QueryAccountTaskByInviter return error accountId:%d, err:%v", accountId, err)
	}
	var dataSet [](map[string]interface{})
	for _, l := range list {
		subMap, ok := types.AccountTaskMap[l.TaskType]
		if !ok {
			continue
		}

		str, ok := subMap[l.TaskStatus]
		if !ok {
			continue
		}

		account, err := models.OneAccountBaseByPkId(l.AccountId)
		if err != nil {
			continue
		}

		mobile := tools.MobileDesensitization(account.Mobile)

		subSet := map[string]interface{}{
			"status": i18n.T(i18n.LangIdID, str),
			"name":   mobile,
		}

		dataSet = append(dataSet, subSet)
	}

	data["data"] = dataSet
}
