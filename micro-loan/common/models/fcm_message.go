package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const FCM_MESSAGE_TABLENAME string = "fcm_message"

// Role 描述数据表结构与结构体的映射
type FcmMessage struct {
	Id          int64 `orm:"pk;"`
	AccountId   int64
	TaskId      int64
	MessageType int
	Title       string
	Body        string
	IsRead      int
	Mark        string
	SkipTo      int
	Ctime       int64
}

// TableName 返回当前模型对应的表名
func (r *FcmMessage) TableName() string {
	return FCM_MESSAGE_TABLENAME
}

// Using 返回当前模型的数据库
func (r *FcmMessage) Using() string {
	return types.OrmDataBaseMessage
}

func (r *FcmMessage) Add() (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Insert(r)
	if err != nil {
		return 0, err
	}
	r.Id = id
	return id, err
}

func (r *FcmMessage) Get(id int64) error {
	o := orm.NewOrm()
	o.Using(r.Using())
	err := o.QueryTable(r.TableName()).Filter("id", id).One(r)
	return err
}

func (r *FcmMessage) Update() (err error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	_, err = o.Update(r)
	if err != nil {
		return err
	}

	return err
}
