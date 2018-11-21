package models

// `admin`
import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const ADMIN_PREREDUCED_TABLENAME string = "admin_prereduced"

type AdminPrereduced struct {
	Id                            int64   `orm:"pk;"`
	Opuid                         int64   `orm:"column(Opuid)"`
	OrderID                       int64   `orm:"column(order_id)"`
	CaseID                        int64   `orm:"column(case_id)"`
	DerateRatio                   float64 `orm:"column(derate_ratio)"`
	GracePeriodInterestPrededuced int64   `orm:"column(grace_period_interest_prereduced)"`
	PenaltyPrereduced             int64   `orm:"column(penalty_prereduced)"`
	PrereducedStatus              int     `orm:"column(prereduced_status)"`
	InvalidReason                 string  `orm:"column(invalid_reason)"`
	Ctime                         int64   `orm:"column(ctime)"`
	Utime                         int64   `orm:"column(utime)"`
}

// 此处声明为指针方法,并不会修改传入的对象,只是为了省去拷贝对象的开消

// 当前模型对应的表名
func (r *AdminPrereduced) TableName() string {
	return ADMIN_PREREDUCED_TABLENAME
}

// 当前模型的数据库
func (r *AdminPrereduced) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *AdminPrereduced) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func OneAdminPrereducedByUid(id int64) (AdminPrereduced, error) {
	obj := AdminPrereduced{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	err := o.QueryTable(obj.TableName()).Filter("id", id).One(&obj)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneAdminPrereducedByUid] sql error err:%v", err)
	}
	return obj, err
}
