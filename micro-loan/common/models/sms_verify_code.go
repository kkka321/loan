package models

// `product`
import (
	//"github.com/astaxie/beego/logs"

	"github.com/astaxie/beego/orm"

	//"micro-loan/common/tools"
	//"fmt"

	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
)

const SMS_VERIFY_CODE_TABLENAME string = "sms_verify_code"

type SmsVerifyCode struct {
	Id           int64             `orm:"pk;"`
	ServiceType  types.ServiceType `orm:"column(service_type)"`
	Mobile       string
	Code         string
	AuthCodeType types.AuthCodeType `orm:"column(authcode_type)"`
	Expires      int
	Ip           string
	Status       int
	Ctime        int64
	Utime        int64
}

// 当前模型对应的表名
func (r *SmsVerifyCode) TableName() string {
	return SMS_VERIFY_CODE_TABLENAME
}

// 当前模型的数据库
func (r *SmsVerifyCode) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *SmsVerifyCode) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *SmsVerifyCode) AddSms(smsVerifyCode *SmsVerifyCode) (int64, error) {

	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(smsVerifyCode)

	return id, err
}

func (r *SmsVerifyCode) GetSmsCode(phoneNumber string) (*SmsVerifyCode, error) {

	var smsVerifyCode SmsVerifyCode

	o := orm.NewOrm()
	o.Using(r.Using())
	err := o.QueryTable("sms_verify_code").Filter("mobile", phoneNumber).Filter("status", types.VerifyCodeUnchecked).OrderBy("-id").Limit(1).One(&smsVerifyCode)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[GetSmsCode] sql error err:%v", err)
	}

	return &smsVerifyCode, err

}

func (r *SmsVerifyCode) GetSmsCodeByPhoneAndServiceType(phoneNumber string, serviceType types.ServiceType) (*SmsVerifyCode, error) {

	var smsVerifyCode SmsVerifyCode
	o := orm.NewOrm()
	o.Using(r.Using())
	err := o.QueryTable("sms_verify_code").Filter("mobile", phoneNumber).Filter("status", types.VerifyCodeUnchecked).
		Filter("service_type", int8(serviceType)).OrderBy("-id").Limit(1).One(&smsVerifyCode)

	return &smsVerifyCode, err

}

func (r *SmsVerifyCode) GetSmsCodeByServiceTypeAndAuthCodeType(phoneNumber string, serviceType types.ServiceType, authCodeType types.AuthCodeType) (*SmsVerifyCode, error) {

	var smsVerifyCode SmsVerifyCode
	o := orm.NewOrm()
	o.Using(r.Using())
	err := o.QueryTable("sms_verify_code").Filter("mobile", phoneNumber).Filter("status", types.VerifyCodeUnchecked).
		Filter("service_type", int8(serviceType)).Filter("authcode_type", int8(authCodeType)).OrderBy("-id").Limit(1).One(&smsVerifyCode)

	return &smsVerifyCode, err

}

func (r *SmsVerifyCode) SetStatusUsed(smsVerifyCode *SmsVerifyCode) {
	o := orm.NewOrm()
	o.Using(r.Using())
	smsVerifyCode.Status = types.VerifyCodeChecked
	o.Update(smsVerifyCode)
}
