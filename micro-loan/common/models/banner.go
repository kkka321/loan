package models

// `admin`
import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const BANNER_TABLENAME string = "banner"

type Banner struct {
	Id          int64  `orm:"pk;"`                 //'主键id',
	ResourceId  int64  `orm:"column(resource_id)"` //'',
	LinkUrl     string `orm:"column(link_url)"`    //'链接url',
	SourcePage  int64  `orm:"column(source_page)"` //'原生页面映射值，大于0->原生',
	Postion     int64  `orm:"column(postion)"`
	BannerType  int    `orm:"column(type)"`          // banner类型
	StartTime   int64  `orm:"column(start_time)"`    // 开始时间
	EndTime     int64  `orm:"column(end_time)"`      // 结束时间
	Content     string `orm:"column(content)"`       // 图片上的文字内容
	FontColor   string `orm:"column(font_color)"`    // 图片上的文字颜色
	FontLinkUrl string `orm:"column(font_link_url)"` // 图片上的文字跳转链接
	Ctime       int64  `orm:"column(ctime)"`         //'添加时间',
	Utime       int64  `orm:"column(utime)"`         //'更新时间',
}

// 当前模型对应的表名
func (r *Banner) TableName() string {
	return BANNER_TABLENAME
}

// 当前模型的数据库
func (r *Banner) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *Banner) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *Banner) Insert() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

func (r *Banner) Updates(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}
func (r *Banner) Dels(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Delete(r, cols...)

	return
}

func GetMultiBanners() (data []Banner, err error) {
	o := orm.NewOrm()
	obj := Banner{}
	o.Using(obj.Using())

	_, err = o.QueryTable(obj.TableName()).OrderBy("postion").
		All(&data)

	return
}

func GetMultiBannersByType(bannerType int) (data []Banner, err error) {
	o := orm.NewOrm()
	obj := Banner{}
	o.Using(obj.Using())

	_, err = o.QueryTable(obj.TableName()).Filter("type", bannerType).OrderBy("postion").
		All(&data)

	return
}
