package runner

import (
	"encoding/json"
	"micro-loan/common/pkg/event/evtypes"
	"reflect"

	"github.com/astaxie/beego/logs"
	"github.com/mohae/deepcopy"
)

// 主要分为 EventParser(事件解析器) 和 XxxxEv(一类事件)

// 触发某事件
// 初始化 e1 := XxxxEv{}
// Trigger(e1)

// 事件运行
// 任务自动从事件队列中取出事件, 由事件解析器, 解析并运行
// 从redis 获取 二进制 val []byte
// EventParser.Run(val)

// 此处 regEvent 作为一个注册事件结构定义
type regEvent struct {
	// RunFunc 为异步事件运行时调用的方法
	RunFunc func(persistentParam interface{}) (success bool, err error)
	// PersistentParam 为异步事件触发时定义并在运行时的持久化数据
	// 因 PersistentParam 会在任意pkg触发位置被import, 并在运行位置调用任意pkg
	// 极易造成循环引用, 因此此处做分离设计
	// runFunc 被 task触发, 然后调用任意位置的pkg满足运行条件
	// PersistentParam 被任意位置调用, 而不引入系统任何包, 作为types存在
	// 为兼容之前event, 此处仍然使用此struct name 作为事件唯一识别键
	// 所以不同事件, 请保证简短而唯一, 实际目前设计, 应使用 func name为佳
	PersistentParam interface{} // must be put in events
}

// EventParser 全局事件解析器, 注册解析事件, 自动保存当前事件Map
var globalParser *parser

func init() {
	// 注册事件
	globalParser = new(parser)
	globalParser.Register(
		//用户激活事件--用户行为
		regEvent{userActiveEv, new(evtypes.UserActiveEv)},
		//用户提交借款申请事件--用户行为
		regEvent{orderApplyEv, new(evtypes.OrderApplyEv)},
		//用户提交审核事件[都归属于未完成一笔贷款]（申请之后传各种资料后再次确认进入审核）--用户行为
		regEvent{orderAuditEv, new(evtypes.OrderAuditEv)},
		//贷款申请事件
		regEvent{loanSubmitEv, new(evtypes.LoanSubmitEv)},
		//放款成功事件
		regEvent{loanSuccessEv, new(evtypes.LoanSuccessEv)},
		//用户还款成功事件 --用户行为
		regEvent{repaySuccessEv, new(evtypes.RepaySuccessEv)},
		//订单失效事件
		regEvent{orderInvalidEv, new(evtypes.OrderInvalidEv)},
		//黑名单事件
		regEvent{blacklistEv, new(evtypes.BlacklistEv)},
		//工单创建事件
		regEvent{ticketCreateEv, new(evtypes.TicketCreateEv)},
		//api统计事件发生
		regEvent{customerStatisticEv, new(evtypes.CustomerStatisticEv)},
		//员工后台登录异步事件
		regEvent{workerDailyFirstOnlineEv, new(evtypes.WorkerDailyFirstOnlineEv)},
		//逾期更新固定付款码金额
		regEvent{updateFixPaymentCodeAmount, new(evtypes.FixPaymentCodeEv)},
		//推送注册事件给appsflyer
		regEvent{registerTrackEv, new(evtypes.RegisterTrackEv)},
	)
}

// parser 描述解析器类型
type parser struct {
	// 保存所有注册事件映射
	//EventMap  map[string]EventInterface
	RegisteredEventMap map[string]regEvent
}

// Register 注册事件
func (p *parser) Register(dbrs ...regEvent) {
	if p.RegisteredEventMap == nil {
		p.RegisteredEventMap = make(map[string]regEvent)
	}
	for _, v := range dbrs {
		name := evtypes.GetStructName(v.PersistentParam)
		if _, ok := p.RegisteredEventMap[name]; !ok {
			p.RegisteredEventMap[name] = v
		}
	}
}

// Run 解析器解析并运行事件
func (p *parser) Run(d []byte) (success bool, err error) {
	success = false

	logs.Debug("[event][runner] string queue value", string(d))
	var eql evtypes.QueueVal
	json.Unmarshal(d, &eql)
	if v, ok := p.RegisteredEventMap[eql.EventName]; ok {
		// 此处拷贝的指针类型
		realParam := deepcopy.Copy(v.PersistentParam)
		json.Unmarshal(eql.Data, realParam)
		// 此处传递的也是指针类型的 struct
		success, err = v.RunFunc(realParam)
		if err != nil {
			logs.Error("[event.Parser.Run]", err)
		}
	} else {
		logs.Error("[event.Run] unregistered event:", string(d))
	}
	return
}

// RunSepEvent 解析器解析并运行事件
func (p *parser) RunSepEvent(d []byte, eventName string) (success bool, err error) {
	success = false

	logs.Debug("[event][runner] string queue value", string(d), "eventName:", eventName)

	if v, ok := p.RegisteredEventMap[eventName]; ok {
		// 此处拷贝的指针类型
		realParam := deepcopy.Copy(v.PersistentParam)
		json.Unmarshal(d, realParam)
		// 此处传递的也是指针类型的 struct
		success, err = v.RunFunc(realParam)
		if err != nil {
			logs.Error("[event.Parser.Run]", err)
		}
	} else {
		logs.Error("[event.Run] unregistered event:", string(d), "eventName:", eventName, p.RegisteredEventMap)
	}
	return
}

// newByCopy a interface by a original copy
// return a value type
func newByCopy(o interface{}) interface{} {
	if o == nil {
		return nil
	}
	ov := reflect.ValueOf(o)

	cpy := reflect.New(ov.Type()).Elem()
	return cpy.Interface()
}
