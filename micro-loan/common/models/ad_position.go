package models

// `admin`
import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const AD_POSITION_TABLENAME string = "ad_position"

type AdPosition struct {
	Id         int64  `orm:"pk;"`                 //'主键id',
	ResourceId int64  `orm:"column(resource_id)"` //'',
	LinkUrl    string `orm:"column(link_url)"`    //'链接url',
	CompanyId  int    `orm:"column(company_id)"`  //公司id
	Position   int    `orm:"column(position)"`    //广告位
	Ctime      int64  `orm:"column(ctime)"`       //'添加时间',
	Utime      int64  `orm:"column(utime)"`       //'更新时间',
}

// 当前模型对应的表名
func (r *AdPosition) TableName() string {
	return AD_POSITION_TABLENAME
}

// 当前模型的数据库
func (r *AdPosition) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *AdPosition) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *AdPosition) Insert() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

func (r *AdPosition) Updates(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}
func (r *AdPosition) Dels(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Delete(r, cols...)

	return
}

func GetAdPositionByCompanyIdAndPosition(companyId, position int) (data AdPosition, err error) {
	o := orm.NewOrm()
	obj := AdPosition{}
	o.Using(obj.Using())

	err = o.QueryTable(obj.TableName()).Filter("company_id", companyId).Filter("position", position).OrderBy("-id").One(&data)

	return
}

func GetMultiAdPosition() (data []AdPosition, err error) {
	o := orm.NewOrm()
	obj := AdPosition{}
	o.Using(obj.Using())

	_, err = o.QueryTable(obj.TableName()).OrderBy("company_id").
		All(&data)

	return
}
