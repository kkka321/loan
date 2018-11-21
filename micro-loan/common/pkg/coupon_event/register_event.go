package coupon_event

import (
	"github.com/astaxie/beego/logs"
)

type RegisterEvent struct {
}

func (c *RegisterEvent) HandleEvent(trigger CouponEventTrigger, data interface{}) {
	logs.Debug("[RegisterEvent] HandleEvent trigger:%d, data:%v", trigger, data)

	if trigger != TriggerRegister {
		return
	}

	if data == nil {
		return
	}

	accountId, ok := data.(int64)
	if !ok || accountId == 0 {
		logs.Debug("[RegisterEvent] accountId empty data:%v", data)
		return
	}

	distributeCoupon("新注册用户", accountId)
}
