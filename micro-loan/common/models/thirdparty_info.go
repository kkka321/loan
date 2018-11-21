package models

import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const THIRDPARTY_INFO_TABLENAME string = "thirdparty_info"

type ThirdpartyInfo struct {
	Id                  int64 `orm:"pk;"`
	Index               int
	Name                string
	Api                 string
	ApiMd5              string
	ChargeType          int
	Price               int
	IsPaymentThirdparty int
	Remarks             string
	Ctime               int64
	Utime               int64
}

func (r *ThirdpartyInfo) TableName() string {
	return THIRDPARTY_INFO_TABLENAME
}

func (r *ThirdpartyInfo) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *ThirdpartyInfo) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

// func AddOneThirdpartyRecord(thirdparty int, api string, relatedId int64, request interface{}, response interface{}) {
// 	requestJSON, _ := json.Marshal(request)
// 	responseJSON, _ := json.Marshal(response)

// 	obj := ThirdpartyRecord{
// 		Thirdparty: thirdparty,
// 		RelatedId:  relatedId,
// 		Api:        api,
// 		Request:    string(requestJSON),
// 		Response:   string(responseJSON),
// 		Ctime:      tools.GetUnixMillis(),
// 	}

// 	o := orm.NewOrm()
// 	o.Using(obj.Using())
// 	o.Insert(&obj)
// }

func (r *ThirdpartyInfo) Add() (id int64, err error) {

	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Insert(r)
	return
}

func (r *ThirdpartyInfo) Upadte(cols ...string) (err error) {

	o := orm.NewOrm()
	o.Using(r.Using())
	_, err = o.Update(r, cols...)
	return
}

func GetThirdpartyInfoByPkId(id int64) (one ThirdpartyInfo, err error) {
	obj := ThirdpartyInfo{}

	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	err = o.QueryTable(obj.TableName()).Filter("id", id).One(&one)
	return

}

// func GetThirdpartyInfoByApi(api string) (one ThirdpartyInfo, err error) {
// 	obj := ThirdpartyInfo{}

// 	o := orm.NewOrm()
// 	o.Using(obj.UsingSlave())
// 	err = o.QueryTable(obj.TableName()).Filter("api", api).One(&one)
// 	return
// }

func GetThirdpartyInfoByApiMd5(apiMd5 string) (one ThirdpartyInfo, err error) {
	obj := ThirdpartyInfo{}

	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	err = o.QueryTable(obj.TableName()).Filter("api_md5", apiMd5).One(&one)
	return
}

func ThirdpartyInfoList() (list []ThirdpartyInfo, err error) {
	obj := ThirdpartyInfo{}

	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	_, err = o.QueryTable(obj.TableName()).All(&list)
	return
}
