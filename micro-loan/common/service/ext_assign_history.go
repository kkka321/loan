package service

import (
	"fmt"
	"micro-loan/common/models"
	"strings"

	"github.com/astaxie/beego/orm"
)

type SipAssignHistorys struct {
	Id           int64  `orm:"pk;"` //通话记录id
	NickName     string `orm:"column(nickname)"`
	ExtNumber    string `orm:"column(extnumber)"`     //分机号码
	AssignId     int64  `orm:"column(assign_id)"`     //分配人员id
	AssignTime   int64  `orm:"column(assign_time)"`   //分配时间
	UnAssignTime int64  `orm:"column(unassign_time)"` //未分配时间
	Ctime        int64  `orm:"column(ctime)"`         //创建时间
}

func ListAssignHistoryBackend(condCntr map[string]interface{}, page int, pagesize int) (lists []SipAssignHistorys, total int64, err error) {
	obj := models.SipAssignHistory{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pagesize

	// 初始化查询条件
	where := whereAssignHistoryBackend(condCntr)
	sqlCount := fmt.Sprintf("SELECT COUNT(sip_assign_history.id) FROM `%s` %s", obj.TableName(), where)
	sqlList := fmt.Sprintf("SELECT sip_assign_history.id,extnumber,nickname,assign_time,unassign_time FROM `%s` %s ORDER BY sip_assign_history.id asc LIMIT %d,%d", obj.TableName(), where, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&lists)

	return
}

func whereAssignHistoryBackend(condCntr map[string]interface{}) string {
	// 初始化查询条件
	cond := []string{}

	//分机号码
	if _, ok := condCntr["extnumber"]; ok {

		cond = append(cond, fmt.Sprintf("extnumber=%s", condCntr["extnumber"].(string)))
	}

	if len(cond) > 0 {
		return fmt.Sprintf("%s%s%s", " left join microloan_admin.admin on assign_id = microloan_admin.admin.id ", "WHERE ", strings.Join(cond, " "))
	}

	return ""
}
