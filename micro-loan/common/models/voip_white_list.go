package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

const VOIP_WHITE_LIST_TABLENAME string = "voip_white_list"

type VoipWhiteList struct {
	Id     int64 `orm:"pk"`
	Mobile string
	Ctime  int64
}

func (r *VoipWhiteList) TableName() string {
	return VOIP_WHITE_LIST_TABLENAME
}

// 当前模型的数据库
func (r *VoipWhiteList) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *VoipWhiteList) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *VoipWhiteList) Insert() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.Ctime = tools.GetUnixMillis()
	id, err := o.Insert(r)

	return id, err
}

// GetWhiteListByMobile 根据手机号查询白名单列表
func GetWhiteListByMobile(mobile string) (VoipWhiteList, error) {
	var whiteList VoipWhiteList

	o := orm.NewOrm()
	o.Using(whiteList.Using())
	err := o.QueryTable(whiteList.TableName()).Filter("mobile", mobile).One(&whiteList)

	return whiteList, err
}
