package models

// `admin`
import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/tools"
	"micro-loan/common/types"
)

const ADVERTISEMENT_TABLENAME string = "advertisement"

type Advertisement struct {
	Id         int64  `orm:"pk;"`                //'主键id',
	ResourceId int64  `orm:"column(resource_id)` //'',
	LinkUrl    string `orm:"column(link_url)`    //'链接url',
	SourcePage int64  `orm:"column(source_page)` //'原生页面映射值，大于0->原生',
	StartTm    int64  `orm:"column(start_tm)`    // '展示的开始时间',
	EndTm      int64  `orm:"column(end_tm)`      //'展示的结束时间',
	IsShow     int64  `orm:"column(is_show)`     //'是否展示:1-展示，0-不展示',
	Ctime      int64  `orm:"column(ctime)`       //'添加时间',
	Utime      int64  `orm:"column(utime)`       //'更新时间',

}

// 当前模型对应的表名
func (r *Advertisement) TableName() string {
	return ADVERTISEMENT_TABLENAME
}

// 当前模型的数据库
func (r *Advertisement) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *Advertisement) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *Advertisement) Insert() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

func (r *Advertisement) Updates(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}
func (r *Advertisement) Dels(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Delete(r, cols...)

	return
}

func GetMultiAdvertisements() (data []Advertisement, err error) {
	o := orm.NewOrm()
	obj := Advertisement{}
	o.Using(obj.Using())

	_, err = o.QueryTable(obj.TableName()).OrderBy("id").
		All(&data)

	return
}

func OneAdvertisementByTm() (ad Advertisement, err error) {
	nowDate := tools.GetUnixMillis()
	o := orm.NewOrm()
	obj := Advertisement{}
	o.Using(obj.Using())

	err = o.QueryTable(obj.TableName()).
		Filter("start_tm__lte", nowDate).Filter("end_tm__gte", nowDate).Filter("is_show", 1).
		One(&ad)

	return
}
