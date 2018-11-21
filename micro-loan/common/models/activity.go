package models

// ``
import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/tools"
	"micro-loan/common/types"
)

const ACTIVITY_TABLENAME string = "activity"

type Activity struct {
	Id         int64  `orm:"pk;"`
	Etype      int64  `orm:"column(etype)"`
	FPostion   int64  `orm:"column(f_postion)"`
	ResourceId int64  `orm:"column(resource_id)"`
	LinkUrl    string `orm:"column(link_url)"`
	SourcePage int64  `orm:"column(source_page)"`
	StartTm    int64  `orm:"column(start_tm)"`
	EndTm      int64  `orm:"column(end_tm)"`
	IsShow     int64  `orm:"column(is_show)"`
	Ctime      int64  `orm:"column(ctime)"`
	Utime      int64  `orm:"column(utime)"`
}

// 此处声明为指针方法,并不会修改传入的对象,只是为了省去拷贝对象的开消

// 当前模型对应的表名
func (r *Activity) TableName() string {
	return ACTIVITY_TABLENAME
}

// 当前模型的数据库
func (r *Activity) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *Activity) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}
func (r *Activity) Insert() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)

	return id, err
}

// Add
func (r *Activity) Add() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.Ctime = tools.GetUnixMillis()
	r.Utime = r.Ctime

	id, err := o.Insert(r)

	return id, err
}

func (r *Activity) Updates(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}
func (r *Activity) Dels(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Delete(r, cols...)

	return
}

func GetAllByEtype(etype int64) (data []Activity, err error) {
	var r = Activity{}
	o := orm.NewOrm()
	o.Using(r.UsingSlave())

	_, err = o.QueryTable(r.TableName()).Filter("etype", etype).All(&data)
	return
}

func GetOneByEtypeAndPostionFloating(etype, postion int64) (data Activity, err error) {
	var r = Activity{}
	nowDate := tools.GetUnixMillis()
	o := orm.NewOrm()
	o.Using(r.UsingSlave())

	err = o.QueryTable(r.TableName()).Filter("etype", etype).Filter("f_postion", postion).Filter("start_tm__lte", nowDate).
		Filter("end_tm__gte", nowDate).Filter("is_show", 1).One(&data)
	return
}

func GetOneByEtypeAndPostionPopWindow(etype, postion int64) (data []Activity, err error) {
	var r = Activity{}
	nowDate := tools.GetUnixMillis()
	o := orm.NewOrm()
	o.Using(r.UsingSlave())

	_, err = o.QueryTable(r.TableName()).Filter("etype", etype).Filter("f_postion", postion).Filter("start_tm__lte", nowDate).
		Filter("end_tm__gte", nowDate).Filter("is_show", 1).All(&data)
	return
}

func GetOneByEtypePopWindow(etype int64) (data []Activity, err error) {
	var r = Activity{}

	o := orm.NewOrm()
	o.Using(r.UsingSlave())

	_, err = o.QueryTable(r.TableName()).Filter("etype", etype).All(&data)
	return
}
