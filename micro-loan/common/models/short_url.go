package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// SHORT_URL_TABLENAME 表名
const SHORT_URL_TABLENAME string = "short_url"

// SHORT_URL_TABLENAME 描述数据表结构与结构体的映射
type ShortUrl struct {
	Id       int64 `orm:"pk;"`
	Url      string
	UrlMd5   string
	ShortUrl string
	Ctime    int64
}

// TableName 返回当前模型对应的表名
func (r *ShortUrl) TableName() string {
	return SHORT_URL_TABLENAME
}

// Using 返回当前模型的数据库
func (r *ShortUrl) Using() string {
	return types.OrmDataBaseApi
}

func (r *ShortUrl) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func GetShortUrl(url string) (data ShortUrl, err error) {
	obj := ShortUrl{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	qs := o.QueryTable(obj.TableName())

	err = qs.Filter("short_url", url).One(&data)

	return
}

func GetShortUrlByMd5(md5 string) (data ShortUrl, err error) {
	obj := ShortUrl{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	qs := o.QueryTable(obj.TableName())

	err = qs.Filter("url_md5", md5).One(&data)

	return
}

func (r *ShortUrl) Insert() error {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err := o.Insert(r)
	return err
}
