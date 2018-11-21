package models

// `product`
import (
	//"github.com/astaxie/beego/logs"
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/tools"
	"micro-loan/common/types"
	//"fmt"
)

const SMS_TABLENAME string = "sms"

type Sms struct {
	Id              int64              `orm:"pk;"`
	MsgID           string             `orm:"column(msg_id)"`
	RelatedID       int64              `orm:"column(related_id)"`
	SmsService      types.SmsServiceID `orm:"column(sms_service)"`
	ServiceType     types.ServiceType  `orm:"column(service_type)"`
	Mobile          string
	ip              string
	Content         string
	SendStatus      int `orm:"column(send_status)"`
	DeliveryStatus  int `orm:"column(delivery_status)"`
	CallbackContent int `orm:"column(callback_content)"`
	Receipt         string
	Ctime           int64
	Utime           int64
	DeliveryTime    int64 `orm:"column(delivery_time)"`
}

// 当前模型对应的表名
func (r *Sms) TableName() string {
	return SMS_TABLENAME
}

// 当前模型的数据库
func (r *Sms) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *Sms) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// AddSms 添加sms log记录, *Sms 无需初始化Ctime, 内部已自动初始化
func (r *Sms) AddSms() (int64, error) {
	r.Ctime = tools.GetUnixMillis()
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

// UpdateSmsByMsgID 更新送达状态
func UpdateSmsByMsgID(msgID string, deliveryStatus int, callbackContent interface{}) (int64, error) {
	if len(msgID) <= 0 {
		return 0, fmt.Errorf("[Sms delivery update], msgID(val:%s) can't be empty ", msgID)
	}
	bc, _ := json.Marshal(callbackContent)
	o := orm.NewOrm()

	s := &Sms{}
	o.Using(s.Using())

	t := tools.GetUnixMillis()
	sql := fmt.Sprintf("UPDATE sms SET delivery_status=%d, callback_content='%s', utime=%d",
		deliveryStatus, string(bc), t)
	if deliveryStatus == types.DeliverySuccess {
		sql += fmt.Sprintf(" ,delivery_time=%d", t)
	}
	sql += fmt.Sprintf("  WHERE msg_id = '%s'", msgID)

	rs := o.Raw(sql)
	sqlResult, err := rs.Exec()
	if err != nil {
		logs.Error(err)
		return 0, err
	}
	aff, err := sqlResult.RowsAffected()
	return aff, err
}
