package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const UPLOAD_RESOURCE_TABLENAME string = "upload_resource"

type UploadResource struct {
	Id         int64  `orm:"pk;"`
	OpUid      int64  `orm:"column(op_uid)"`
	ContentMd5 string `orm:"column(content_md5)"`
	HashName   string `orm:"column(hash_name)"`
	Extension  string
	Mime       string
	UseMark    types.ResourceUseMark
	Ctime      int64
}

func (r *UploadResource) TableName() string {
	return UPLOAD_RESOURCE_TABLENAME
}

func (r *UploadResource) Using() string {
	return types.OrmDataBaseApi
}
func (r *UploadResource) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

//获取上传资源中的所有字段
func GetMultiPicShow(op_id int64) ([]UploadResource, error) {
	uploadResource := UploadResource{}

	var data []UploadResource
	o := orm.NewOrm()
	o.Using(uploadResource.UsingSlave())
	_, err := o.QueryTable(uploadResource.TableName()).Filter("op_uid", op_id).OrderBy("-id").All(&data)

	return data, err
}

//获取上传资源中的最近两个记录
func GetLatestSecondResource(op_id int64) ([]UploadResource, error) {
	uploadResource := UploadResource{}

	var data []UploadResource
	o := orm.NewOrm()
	o.Using(uploadResource.UsingSlave())
	_, err := o.QueryTable(uploadResource.TableName()).Filter("op_uid", op_id).OrderBy("-id").Limit(2).All(&data)

	return data, err
}

func GetHashNameByResourceId(id int64) (UploadResource, error) {
	uploadResource := UploadResource{}

	var data UploadResource
	o := orm.NewOrm()
	o.Using(uploadResource.UsingSlave())
	err := o.QueryTable(uploadResource.TableName()).Filter("id", id).One(&data)

	return data, err
}
