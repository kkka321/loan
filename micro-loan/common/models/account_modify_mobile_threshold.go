package models

// `admin`
import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const ACCOUNT_MODIFY_MOBILE_THRESHOLD_TABLENAME string = "account_modify_mobile_threshold"

type AccountModifyMobileThreshold struct {
	Id                       int64  `orm:"pk;"`                                 // 主键id
	AccountId                int64  `orm:"column(account_id)"`                  // 账号ID
	IdPhoto                  int64  `orm:"column(id_photo)"`                    // 身份证照资源ID
	IdPhotoThreshold         string `orm:"column(id_photo_threshold)"`          // 上传身份证与历史身份证对比阈值
	HandPhoto                int64  `orm:"column(hand_photo)"`                  // 上传手持身份证照资源ID
	HandPhotoThreshold       string `orm:"column(hand_photo_threshold)"`        // 上传手持身份证照脸部识别对比阈值
	HandPhotoRecopyThreshold string `orm:"column(hand_photo_recopy_threshold)"` // 上传手持身份证照翻拍对比阈值
	Ctime                    int64  `orm:"column(ctime)"`                       // 添加时间
	Utime                    int64  `orm:"column(utime)"`                       // 更新时间
}

// 当前模型对应的表名
func (r *AccountModifyMobileThreshold) TableName() string {
	return ACCOUNT_MODIFY_MOBILE_THRESHOLD_TABLENAME
}

// 当前模型的数据库
func (r *AccountModifyMobileThreshold) Using() string {
	return types.OrmDataBaseApi
}

func (r *AccountModifyMobileThreshold) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *AccountModifyMobileThreshold) Insert() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

func (r *AccountModifyMobileThreshold) Updates(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}
