package models

import (
	"encoding/json"
	"fmt"

	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const THIRDPARTY_RECORD_TABLENAME string = "thirdparty_record"

// 第三方服务编号
const (
	ThirdpartyAdvance      int = 1
	ThirdpartyFaceid       int = 2
	ThirdpartyAppsFlyer    int = 3
	ThirdpartyTextLocal    int = 4
	ThirdpartySms253       int = 5
	ThirdpartyAkulaku      int = 6
	ThirdpartyTongdun      int = 7
	ThirdpartyBoomsms      int = 8
	ThirdpartyXendit       int = 9
	ThirdpartyBluepay      int = 10
	ThirdpartyNexmo        int = 11
	ThirdpartyDoKu         int = 12
	ThirdpartyAPI253       int = 13
	ThirdpartySmsPhone     int = 14
	ThirdpartySmsCmtelecom int = 15
)

const (
	ThirdpartyNameFaceid       string = "faceid"
	ThirdpartyNameAdvance      string = "advance"
	ThirdpartyNameAppsFlyer    string = "appsflyer"
	ThirdpartyNameTextLocal    string = "textlocal"
	ThirdpartyNameSms253       string = "sms253"
	ThirdpartyNameAkulaku      string = "akulaku"
	ThirdpartyNameTongdun      string = "tongdun"
	ThirdpartyNameBoomsms      string = "boomsms"
	ThirdpartyNameXendit       string = "xendit"
	ThirdpartyNameBluepay      string = "bluepay"
	ThirdpartyNameNexmo        string = "nexmo"
	ThirdpartyNameDoKu         string = "doku"
	ThirdpartyNameAPI253       string = "api253"
	ThirdpartyNameSmsPhone     string = "sms_phone"
	ThirdpartyNameSmsCmtelecom string = "cmtelecom"
)

var ThirdpartyNameMap = map[int]string{
	ThirdpartyAdvance:      ThirdpartyNameAdvance,
	ThirdpartyFaceid:       ThirdpartyNameFaceid,
	ThirdpartyAppsFlyer:    ThirdpartyNameAppsFlyer,
	ThirdpartyTextLocal:    ThirdpartyNameTextLocal,
	ThirdpartySms253:       ThirdpartyNameSms253,
	ThirdpartyAkulaku:      ThirdpartyNameAkulaku,
	ThirdpartyTongdun:      ThirdpartyNameTongdun,
	ThirdpartyBoomsms:      ThirdpartyNameBoomsms,
	ThirdpartyXendit:       ThirdpartyNameXendit,
	ThirdpartyBluepay:      ThirdpartyNameBluepay,
	ThirdpartyNexmo:        ThirdpartyNameNexmo,
	ThirdpartyDoKu:         ThirdpartyNameDoKu,
	ThirdpartyAPI253:       ThirdpartyNameAPI253,
	ThirdpartySmsPhone:     ThirdpartyNameSmsPhone,
	ThirdpartySmsCmtelecom: ThirdpartyNameSmsCmtelecom,
}

type ThirdpartyRecord struct {
	Id               int64 `orm:"pk;"`
	Thirdparty       int
	RelatedId        int64 `orm:"column(related_id)"`
	Api              string
	Request          string
	Response         string
	ResponseType     int
	FeeForCall       int
	Ctime            int64
	HttpResponseCode int `orm:"column(http_response_code)"`
}

func (r *ThirdpartyRecord) OriTableName() string {
	return THIRDPARTY_RECORD_TABLENAME
}

func (r *ThirdpartyRecord) TableName() string {
	timetag := tools.TimeNow()
	return r.TableNameByMonth(timetag)
}

func (r *ThirdpartyRecord) TableNameByMonth(month int64) string {
	date := tools.GetDateFormat(month, "200601")
	return THIRDPARTY_RECORD_TABLENAME + "_" + date
}

func (r *ThirdpartyRecord) Using() string {
	return types.OrmDataBaseApi
}

func (r *ThirdpartyRecord) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func GetThirpartyRecordById(Id int64) (data ThirdpartyRecord, err error) {
	o := orm.NewOrm()
	data = ThirdpartyRecord{}
	o.Using(data.Using())

	sql := "select * from %s where id = %d"
	sql = fmt.Sprintf(sql, data.TableName(), Id)

	err = o.Raw(sql).QueryRow(&data)

	//err = o.QueryTable(data.TableName()).Filter("id", Id).One(&data)

	return data, err
}

func AddOneThirdpartyRecord(thirdparty int, api string, relatedId int64, request interface{}, response interface{}, responseType int, fee int, resCode int) (id int64, err error) {

	requestJSON, _ := json.Marshal(request)
	responseJSON, _ := json.Marshal(response)

	obj := ThirdpartyRecord{
		Thirdparty:       thirdparty,
		RelatedId:        relatedId,
		Api:              api,
		Request:          string(requestJSON),
		Response:         string(responseJSON),
		ResponseType:     responseType,
		FeeForCall:       fee,
		HttpResponseCode: resCode,
		Ctime:            tools.GetUnixMillis(),
	}

	o := orm.NewOrm()
	o.Using(obj.Using())
	id, err = o.Insert(&obj)
	return
}

func (r *ThirdpartyRecord) UpdateFee() (id int64, err error) {

	o := orm.NewOrm()
	o.Using(r.Using())
	id, err = o.Update(r, "fee_for_call", "response_type")

	return
}

//func GetAllByRelatedId(relate int64) (list []ThirdpartyRecord, err error) {
//	obj := ThirdpartyRecord{}
//	o := orm.NewOrm()
//	o.Using(obj.Using())
//
//	_, err = o.QueryTable(obj.TableName()).
//		Filter("related_id", relate).
//		OrderBy("-id").
//		All(&list)
//	return
//}

func GetAllByRelatedIdAndLastId(relate int64, lastId int64) (list []ThirdpartyRecord, err error) {
	obj := ThirdpartyRecord{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	sql := "select * from %s where related_id = %d and id>%d order by id desc "
	sql = fmt.Sprintf(sql, obj.TableName(), relate, lastId)

	_, err = o.Raw(sql).QueryRows(&list)

	//_, err = o.QueryTable(obj.TableName()).
	//	Filter("related_id", relate).
	//	Filter("id__gt", lastId).
	//	OrderBy("-id").
	//	All(&list)
	return
}
