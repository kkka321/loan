package models

import (
	"encoding/json"

	//"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/tools"
	"micro-loan/common/types"
)

const OP_LOGGER_TABLENAME string = "op_logger"

type OpCodeEnum int

const (
	// op_logger 操作码

	OpCodeProductEdit              OpCodeEnum = 100 // 编辑/修改
	OpCodeProductPublish           OpCodeEnum = 101 // 发布
	OpCodeProductSetOff            OpCodeEnum = 102 // 下线
	OpCodeProductDelete            OpCodeEnum = 103 // 删除
	OPCodeCustomerRisk             OpCodeEnum = 104 // 客户风险操作
	OpCodeOrderUpdate              OpCodeEnum = 105 // 修改订单数据
	OpCodeRepayPlanUpdate          OpCodeEnum = 106 // 修改还款计划
	OpUserInfoUpdate               OpCodeEnum = 107 // 修改用户信息
	OpReductionInterestUpdate      OpCodeEnum = 108 // 减免用户利息
	OpOverdueCaseUpdate            OpCodeEnum = 109 // 逾期案件
	OpRepayRemindCaseUpdate        OpCodeEnum = 110 // 还款提醒案件
	OpAutoReduction                OpCodeEnum = 111 // 自动减免用户利息
	OpCodeAccountBaseDelete        OpCodeEnum = 112 // 用户基本信息删除
	OpCodeOrderDelete              OpCodeEnum = 113 // 用户订单删除
	OpCodeTicketAssign             OpCodeEnum = 114 // 工单分配
	OpDelectForRollBackLoan        OpCodeEnum = 115 // 为了放款回退删除数据
	OpCodeAccountBaseUpdate        OpCodeEnum = 120 // 用户基本信息修改
	OpCodeWorkerOnlineStatusUpdate OpCodeEnum = 121 // 后台用户上线状态
	OpCodeSupplementOrder          OpCodeEnum = 122 // 补单操作
	OpPhoneVerifyCaseUpdate        OpCodeEnum = 123 // 电核案件
	OpUserEtransModDel             OpCodeEnum = 124 // user_e_tran 删除及修改
	OpCodeAuthoriationUpdate       OpCodeEnum = 125 // 用户授权信息变化
)

// OpCodeList 描述 opCode 与操作对应关系表
var OpCodeList = map[OpCodeEnum]string{

	OpCodeProductEdit:              "编辑",
	OpCodeProductPublish:           "发布",
	OpCodeProductSetOff:            "下线",
	OpCodeProductDelete:            "删除",
	OPCodeCustomerRisk:             "客户风险",
	OpCodeOrderUpdate:              "修改订单",
	OpCodeRepayPlanUpdate:          "修改还款计划",
	OpUserInfoUpdate:               "编辑用户资料",
	OpReductionInterestUpdate:      "减免用户利息",
	OpOverdueCaseUpdate:            "逾期案件",
	OpCodeAccountBaseUpdate:        "客户基本信息更新",
	OpAutoReduction:                "自动减免",
	OpCodeAccountBaseDelete:        "客户基本信息删除",
	OpCodeWorkerOnlineStatusUpdate: "员工工作状态更新",
	OpCodeSupplementOrder:          "补单操作",
	OpPhoneVerifyCaseUpdate:        "电核案件",
	OpCodeTicketAssign:             "工单分配",
	OpDelectForRollBackLoan:        "回滚放款删除数据",
}

// OpLogger 描述对应表单行数据结构，及字段映射关系
type OpLogger struct {
	Id        int64 `orm:"pk;"`
	OpUid     int64 `orm:"column(op_uid)"`
	RelatedId int64
	OpCode    OpCodeEnum `orm:"column(op_code)"`
	OpTable   string     `orm:"column(op_table)"`
	Original  string
	Edited    string
	Ctime     int64
}

func (r *OpLogger) OriTableName() string {
	return OP_LOGGER_TABLENAME
}

func (r *OpLogger) TableName() string {
	timetag := tools.TimeNow()
	return r.TableNameByMonth(timetag)
}

func (r *OpLogger) TableNameByMonth(month int64) string {
	date := tools.GetDateFormat(month, "200601")
	return OP_LOGGER_TABLENAME + "_" + date
}

func (r *OpLogger) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *OpLogger) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// 用于记录一些关键数据被修改的日志,由业务来决定那些需要记录
func OpLogWrite(opUid int64, relatedId int64, opCode OpCodeEnum, opTable string, original interface{}, edited interface{}) {
	originalJson, _ := json.Marshal(original)
	editedJson, _ := json.Marshal(edited)

	opLogIns := OpLogger{
		OpUid:     opUid,
		RelatedId: relatedId,
		OpCode:    opCode,
		OpTable:   opTable,
		Original:  string(originalJson),
		Edited:    string(editedJson),
		Ctime:     tools.GetUnixMillis(),
	}

	o := orm.NewOrm()
	o.Using(opLogIns.Using())
	o.Insert(&opLogIns)
}
