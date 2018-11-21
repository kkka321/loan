package service

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"micro-loan/common/thirdparty/voip"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type SipInfos struct {
	ExtNumber     string `orm:"column(extnumber)"`     //分机号码
	AssignId      int64  `orm:"column(assign_id)"`     //分配人员id
	CallStatus    int    `orm:"column(call_status)"`   //分机通话状态
	EnableStatus  int    `orm:"column(enable_status)"` //分配是否启用
	AssignStatus  int    `orm:"column(assign_status)"` //分机分配状态  0:未分配; 1:已分配
	Ctime         int64  //创建时间
	Utime         int64  //更新时间
	NickName      string `orm:"column(nickname)"`
	CallStatusStr string
}

func ListExtManageBackend(condCntr map[string]interface{}, page int, pagesize int) (lists []SipInfos, total int64, err error) {
	obj := models.SipInfo{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pagesize

	// 初始化查询条件
	where := whereExtManageBackend(condCntr)
	sqlCount := fmt.Sprintf("SELECT COUNT(extnumber) FROM `%s` %s", obj.TableName(), where)
	sqlList := fmt.Sprintf("SELECT assign_id,extnumber,call_status,enable_status,assign_status,nickname "+
		" FROM `%s` %s ORDER BY extnumber desc LIMIT %d,%d", obj.TableName(), where, offset, pagesize)
	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	var list []SipInfos
	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)
	for _, v := range list {
		tp := SipInfos{}
		tp.NickName = v.NickName
		tp.ExtNumber = v.ExtNumber
		tp.AssignId = v.AssignId
		tp.EnableStatus = v.EnableStatus
		tp.Ctime = v.Ctime
		callStatus, ok := voip.TagCallStatusMap()[v.CallStatus]
		if !ok {
			callStatus = ""
		}
		tp.CallStatusStr = callStatus
		tp.Utime = v.Utime
		tp.AssignStatus = v.AssignStatus
		lists = append(lists, tp)
	}

	return
}

func whereExtManageBackend(condCntr map[string]interface{}) string {
	// 初始化查询条件
	cond := []string{}

	//分机号码
	if v, ok := condCntr["extnumber"]; ok {
		cond = append(cond, fmt.Sprintf("microloan_admin.sip_info.extnumber=%s ", v.(string)))
	}

	//分机状态
	if v, ok := condCntr["call_status"]; ok {
		cond = append(cond, fmt.Sprintf("microloan_admin.sip_info.call_status=%d ", v))
	}

	//是否启用
	if v, ok := condCntr["enable_status"]; ok {
		cond = append(cond, fmt.Sprintf("microloan_admin.sip_info.enable_status=%d ", v))
	}

	//分配状态
	if v, ok := condCntr["assign_status"]; ok {
		cond = append(cond, fmt.Sprintf("microloan_admin.sip_info.assign_status =%d ", v))
	}

	//用户名称
	if v, ok := condCntr["name"]; ok {

		if len(cond) > 0 {
			return fmt.Sprintf("left join microloan_admin.admin on assign_id = microloan_admin.admin.id where microloan_admin.admin.nickname = '%v' AND ", v) +
				strings.Join(cond, " AND ")
		} else {
			return fmt.Sprintf("left join microloan_admin.admin on assign_id = microloan_admin.admin.id where microloan_admin.admin.nickname = '%v' ", v)
		}

	}

	if len(cond) > 0 {
		return "left join microloan_admin.admin on assign_id = microloan_admin.admin.id WHERE " + strings.Join(cond, " AND ")
	} else {
		return "left join microloan_admin.admin on assign_id = microloan_admin.admin.id "
	}
	return ""
}

func UpdateSipInfo() (err error) {
	sipNumberInfos, err := voip.VoipSipNumberInfo(voip.SipNumberInfoAll, "")
	if err != nil {
		logs.Warning("[UpdateExtInfo] Get sip numbers info, err:", err)
		return
	}

	results := sipNumberInfos.Data.Result
	for _, val := range results {
		extNumber := val.ExtNumber
		var enableStatus int
		if val.Status == voip.Sip_Status_1016 {
			enableStatus = 1
		} else if val.Status == voip.Sip_Status_1013 {
			enableStatus = 0
		}

		// 获取呼叫状态
		sipCallStatus, err := voip.VoipSipCallStatus(extNumber)
		if err != nil {
			continue
		}

		newSipInfo := models.SipInfo{
			ExtNumber:    extNumber,
			CallStatus:   sipCallStatus.Data.Result[0].Status,
			EnableStatus: enableStatus,
		}
		sipInfo, err := models.GetSipInfoByExtNumber(extNumber)
		if err != nil || len(sipInfo.ExtNumber) <= 0 {
			// 插入
			newSipInfoStr, _ := tools.JsonEncode(newSipInfo)
			logs.Info("[UpdateExtInfo] Insert sip numbers info:", newSipInfoStr)
			newSipInfo.Insert()
			continue
		}

		// 更新
		newSipInfo.Id = sipInfo.Id
		newSipInfo.Ctime = sipInfo.Ctime
		newSipInfo.AssignId = sipInfo.AssignId
		newSipInfo.AssignStatus = sipInfo.AssignStatus

		newSipInfoStr, _ := tools.JsonEncode(newSipInfo)
		logs.Info("[UpdateExtInfo] Insert sip numbers info:", newSipInfoStr)
		newSipInfo.Update()
	}

	return
}

// CanAssignUsers 获取指定 extension 可分配用户列表
func CanExtensionAssignUsers(extNumber string) (admins []models.Admin, num int64, err error) {

	// 获取已分配分机的用户id
	var assignIds []string
	sipInfos, err := models.GetAssignedSipInfo(extNumber)
	for _, v := range sipInfos {
		assignIds = append(assignIds, tools.Int642Str(v.AssignId))
	}

	admins, num, err = models.GetUsersByAssignIdFromDB(assignIds)

	return
}

// 更改分机分配状态
func ManualAssignOperate(assignID int64, extNumber string, isAssign int) (isSuccessed bool, err error) {
	flagSipInfo := false
	flagSipHistory := false
	assignStatus := 0

	// 获取本地分机信息
	sipInfo, err := models.GetSipInfoByExtNumber(extNumber)
	if err != nil {
		//logs.Error("[ManualAssignOperate] GetSipInfoByExtNumbers err:", err)
		return false, fmt.Errorf(voip.GetSipStatusVal(voip.Sip_Status_1012))
	}
	// 分机被分配之后,不可再分配
	if sipInfo.AssignStatus == 1 {
		return false, fmt.Errorf(types.SipHasAssigned)
	}

	// 获取分机状态信息
	sipNumberInfo, err := voip.VoipSipNumberInfo(voip.SipNumberInfoAll, extNumber)
	if err != nil {
		return false, err
	}
	status := sipNumberInfo.Data.Result[0].Status
	if status == voip.Sip_Status_1019 {
		return false, fmt.Errorf(voip.GetSipStatusVal(voip.Sip_Status_1019))
	}

	// 获取分机通话状态信息
	sipCallStatus, err := voip.VoipSipCallStatus(extNumber)
	if err != nil {
		return false, err
	}
	status = sipCallStatus.Data.Result[0].Status
	if status != voip.Call_Status_1201 && status != voip.Sip_Status_1014 {
		return false, fmt.Errorf(types.SipNotAvailable)
	}

	// 更新本地分机信息
	assignStatus = 1
	tmpSipInfo := models.SipInfo{}
	tmpSipInfo.Id = sipInfo.Id
	tmpSipInfo.AssignStatus = assignStatus
	tmpSipInfo.ExtNumber = extNumber
	tmpSipInfo.AssignId = assignID
	tmpSipInfo.CallStatus = status
	tmpSipInfo.Utime = tools.GetUnixMillis()

	num, err := tmpSipInfo.Updates("extnumber", "assign_status", "assign_id", "call_status", "utime")
	if num > 0 && err == nil {
		flagSipInfo = true
	}

	// 更新分机分配历史
	tmpSipAssignHistory := models.SipAssignHistory{}

	tmpSipAssignHistory.AssignId = assignID
	tmpSipAssignHistory.ExtNumber = extNumber
	tmpSipAssignHistory.Ctime = tools.GetUnixMillis()
	tmpSipAssignHistory.AssignTime = tools.GetUnixMillis()
	tmpSipAssignHistory.Utime = tools.GetUnixMillis()
	num, err = tmpSipAssignHistory.Insert()
	if num > 0 && err == nil {
		flagSipHistory = true
	} else {
		logs.Error("[ManualAssignOperate] tmpSipAssignHistory.Insert, err:", err, "is_assign :", 1)
	}

	if flagSipHistory && flagSipInfo {
		return true, nil
	}
	return false, nil
}

// 更改分机分配状态
func ManualUnAssignOperate(assignID int64, extNumber string, isAssign int) (isSuccessed bool, err error) {
	flagSipInfo := false
	flagSipHistory := false
	assignStatus := 0

	// 获取分机通话状态信息
	sipCallStatus, err := voip.VoipSipCallStatus(extNumber)
	if err != nil {
		return false, err
	}
	status := sipCallStatus.Data.Result[0].Status
	if status != voip.Call_Status_1201 && status != voip.Sip_Status_1014 {
		return false, fmt.Errorf(types.SipNotAvailable)
	}

	sipInfo, _ := models.GetSipInfoByExtNumber(extNumber)

	tmpSipInfo := models.SipInfo{}
	tmpSipInfo.Id = sipInfo.Id
	tmpSipInfo.AssignStatus = assignStatus
	tmpSipInfo.ExtNumber = extNumber
	tmpSipInfo.AssignId = 0
	tmpSipInfo.CallStatus = status
	tmpSipInfo.Utime = tools.GetUnixMillis()

	num, err := tmpSipInfo.Updates("extnumber", "assign_status", "assign_id", "call_status", "utime")
	if num > 0 && err == nil {
		flagSipInfo = true
	}

	// 更新分机分配历史
	tmpSipAssignHistory := models.SipAssignHistory{}

	sipHistory, err := models.GetSipAssignHistory(extNumber, assignID)
	if err != nil {
		logs.Error("[ManualAssignOperate] GetSipAssignHistory, err,is_assign :", err, isAssign)
	}
	tmpSipAssignHistory.Utime = tools.GetUnixMillis()
	tmpSipAssignHistory.UnAssignTime = tools.GetUnixMillis()
	tmpSipAssignHistory.Id = sipHistory.Id
	tmpSipAssignHistory.AssignId = assignID
	tmpSipAssignHistory.ExtNumber = extNumber
	num, err = tmpSipAssignHistory.Updates("utime", "unassign_time", "id")
	if num > 0 && err == nil {
		flagSipHistory = true
	} else {
		logs.Error("[ManualAssignOperate] tmpSipAssignHistory.Update, err,is_assign :", err, 0)
	}

	if flagSipHistory && flagSipInfo {
		return true, nil
	}
	return false, nil
}
