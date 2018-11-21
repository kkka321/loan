package admin

import (
	"errors"
	"fmt"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/tools"

	"micro-loan/common/models"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/types"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type AdminForDisplay struct {
	models.Admin
	RoleName string             `orm:"column(role_name)"`
	RoleType types.RoleTypeEnum `orm:"column(role_type)"`
}

func List(condCntr map[string]interface{}, page int, pagesize int) (list []AdminForDisplay, num int64, err error) {

	obj := models.Admin{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	// 构建查询对象

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = types.DefaultPagesize
	}
	offset := (page - 1) * pagesize

	sqlList := fmt.Sprintf("SELECT L.*,R.name as role_name,R.type as role_type FROM `%s`  L LEFT JOIN %s R ON L.role_id=R.id ORDER BY L.`id` ASC LIMIT %d,%d",
		models.ADMIN_TABLENAME, models.ROLE_TABLENAME, offset, pagesize)

	num, err = o.Raw(sqlList).QueryRows(&list)

	return
}

func LowPrivilegeList(condCntr map[string]interface{}, page int, pagesize int) (list []AdminForDisplay, num int64, err error) {

	obj := models.Admin{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	// 构建查询对象

	lowPriviRoleString, _ := tools.IntsSliceToWhereInString(types.LowPrivilegeRoleTypeContainer())
	cond := "1=1"
	cond += fmt.Sprintf(" AND R.`type` in(%s)", lowPriviRoleString)
	cond += fmt.Sprintf(" AND R.`pid` !=%d", types.RoleSuperPid)

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = types.DefaultPagesize
	}
	offset := (page - 1) * pagesize

	sqlList := fmt.Sprintf("SELECT L.*,R.name as role_name,R.type as role_type FROM `%s`  L LEFT JOIN %s R ON L.role_id=R.id WHERE %s ORDER BY L.`id` ASC LIMIT %d,%d",
		models.ADMIN_TABLENAME, models.ROLE_TABLENAME, cond, offset, pagesize)

	num, err = o.Raw(sqlList).QueryRows(&list)

	return
}

func Add(admin *models.Admin) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(admin.Using())

	admin.WorkStatus = types.AdminWorkStatusNormal
	id, err = o.Insert(admin)

	return
}

func GetReducedQuotaConf(id int64) (quota int) {
	adminModel, _ := models.OneAdminByUid(id)
	quota = adminModel.ReducedQuota
	return
}

// IsExistAdminPrereduced 判断逾期案件是否可以申请结清减免
func IsExistPrereduced(caseID, orderID int64) (exist bool) {
	obj := models.ReduceRecordNew{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	num := 0

	sql := fmt.Sprintf("SELECT COUNT(id) FROM %s  WHERE case_id=%d AND order_id=%d and reduce_status in(%d,%d) and reduce_type=%d",
		obj.TableName(), caseID, orderID, types.ReduceStatusNotValid, types.ReduceStatusValid, types.ReduceTypePrereduced)
	o.Raw(sql).QueryRow(&num)

	if num > 0 {
		exist = true
	}
	return
}

// GetReducedQuotaToday 获取管理员当日结清减免额度
func GetReducedQuotaToday(id int64) (quotaToday int64) {
	obj := models.ReduceRecordNew{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	date := tools.MDateUTC(tools.GetUnixMillis())

	logs.Debug("date:=====", date)

	beginDate := date + " 00:00:00"
	endDate := date + " 23:59:59"

	logs.Debug("beginDate:=====", beginDate)
	logs.Debug("endDate:=====", endDate)

	beginTimeStamp, _ := tools.GetTimeParseWithFormat(beginDate, "2006-01-02 15:04:05")
	endTimeStamp, _ := tools.GetTimeParseWithFormat(endDate, "2006-01-02 15:04:05")
	logs.Debug("beginTimeStamp:=====", beginTimeStamp)
	logs.Debug("endTimeStamp:=====", endTimeStamp)

	sql := fmt.Sprintf("SELECT COUNT(id) FROM %s  WHERE opuid=%d AND ctime>=%d  AND  ctime<=%d AND reduce_type=%d",
		obj.TableName(), id, beginTimeStamp*1000, endTimeStamp*1000, types.ReduceTypePrereduced)
	o.Raw(sql).QueryRow(&quotaToday)

	return
}

func UpdateStatus(adminId int64, status int) (num int64, err error) {
	if adminId <= 1 {
		err = errors.New("参数不正确")
		return
	}

	obj := models.Admin{
		Id:     adminId,
		Status: status,
	}
	o := orm.NewOrm()
	o.Using(obj.Using())
	num, err = o.Update(&obj, "status")
	if num > 0 {
		newAdmin, _ := models.OneAdminByUid(adminId)
		if obj.Status == types.StatusInvalid {
			ticket.PollWatchRoleOfflineUser(newAdmin.RoleID, adminId)
		} else {
			ticket.PollWatchRoleOnlineUser(newAdmin.RoleID, adminId)
		}
	}
	return
}

// Update 更新指定角色的属性
// 不含属性校验
// 内部自动更新 Utime
func Update(m *models.Admin, om *models.Admin, cols []string) (num int64, err error) {

	if m.Id <= 0 {
		err = errors.New("Update ID must exist and >0")
		return
	}

	o := orm.NewOrm()
	o.Using(m.Using())

	num, err = o.Update(m, cols...)
	if num > 0 {
		if m.Nickname != om.Nickname {
			ClearNameCache(m.Id)
		}
		if m.RoleID != om.RoleID {
			ticket.PollWatchRoleOnlineUser(m.RoleID, m.Id)
			ticket.PollWatchRoleOfflineUser(om.RoleID, om.Id)
			// 工单功能正式上线可移除
		}
		if m.WorkStatus != om.WorkStatus {
			if m.WorkStatus == types.AdminWorkStatusNormal {
				ticket.PollWatchRoleOnlineUser(m.RoleID, m.Id)
			} else {
				ticket.PollWatchRoleOfflineUser(om.RoleID, om.Id)
			}
		}
	}
	return
}

// GetNameByID 根据adminID 获取用户名
func GetNameByID(adminID int64) string {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	var name string
	hashKey := beego.AppConfig.String("operator_name")
	valueByte, err := storageClient.Do("HGET", hashKey, adminID)
	//logs.Debug("valueByte:", valueByte, ", err:", err)
	if err == nil && valueByte != nil {
		name = string(valueByte.([]byte))
	} else {
		admin, err := models.OneAdminByUid(adminID)
		if err != nil {
			return "无效的操作员"
		}

		name = admin.Nickname
		storageClient.Do("HSET", hashKey, adminID, name)
	}

	return name
}

// ClearNameCache 清除用户名缓存
func ClearNameCache(adminID int64) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()
	hashKey := beego.AppConfig.String("operator_name")
	storageClient.Do("HDEL", hashKey, adminID)
}

// Worker 工作人员管理页面, 描述工人属性
type Worker struct {
	Id            int64
	Email         string
	Nickname      string
	RoleID        int64 `orm:"column(role_id)"`
	Status        int
	WorkStatus    int
	OnlineStatus  bool
	LastLoginTime int64
	ReducedQuota  int
	IsTodayLogin  bool // 今天用户来过,也就是上过线, 才可以被操作为继续接单 will handle on controller
}

// GetUsersByType 根据角色ID, 获取其下的用户 slices
// 获取指定角色下的后台用户, 用于手动分配列表 和 下属工作状态管理
func GetUsersByType(roleType types.RoleTypeEnum, cond map[string]interface{}, page int, pagesize int) (admins []Worker, num int64, err error) {
	obj := models.Admin{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = types.DefaultPagesize
	}

	offset := (page - 1) * pagesize

	todayStartTime, _ := tools.GetTodayTimestampByLocalTime("00:00")
	sqlList := fmt.Sprintf("SELECT u.`id`, u.`nickname`, u.`email`, u.`work_status`, u.`reduced_quota`,u.`last_login_time`, IF(u.`last_login_time`>%d, 1, 0) as is_today_login FROM `%s` u LEFT JOIN `%s` r ON u.role_id=r.id  WHERE u.status=%d",
		todayStartTime*1000, models.ADMIN_TABLENAME, models.ROLE_TABLENAME, 1)
	sqlListCnt := fmt.Sprintf("SELECT count(u.`id`) as total FROM `%s` u LEFT JOIN `%s` r ON u.role_id=r.id  WHERE u.status=%d", models.ADMIN_TABLENAME, models.ROLE_TABLENAME, 1)

	where := ""

	if v, ok := cond["op_uid"]; ok {
		//sqlList += fmt.Sprintf(" AND u.id = %d", v)
		where = fmt.Sprintf("%s AND u.id = %d", where, v)
	}

	if v, ok := cond["status"]; ok {
		//sqlList += fmt.Sprintf(" AND u.work_status = %d", v)
		where = fmt.Sprintf("%s AND u.work_status = %d", where, v)
	}
	if v, ok := cond["leader_role_id"]; ok {
		//sqlList += fmt.Sprintf(" AND u.work_status = %d", v)
		where = fmt.Sprintf("%s AND (r.pid = %d OR u.id=%d)", where, v, cond["leader_user_id"])
	}

	if roleType > 0 {
		sqlList += fmt.Sprintf(" AND r.`type` =%d ", roleType)
	}

	limitSql := fmt.Sprintf(" limit %d, %d", offset, pagesize)
	sqlList = fmt.Sprintf("%s%s%s", sqlList, where, limitSql)
	sqlListCnt = fmt.Sprintf("%s%s", sqlListCnt, where)

	r := o.Raw(sqlList)
	num, err = r.QueryRows(&admins)

	r = o.Raw(sqlListCnt)
	var total int64
	r.QueryRow(&total)

	return admins, total, err
}

// GetLeaderManageUsers 获取 leader 管理的子用户列表
// 第二个可选参数, 如果传入 leaderUIDs 意味着管理者用户也会被放入其中
func GetLeaderManageUsers(roleID int64, leaderUIDs ...int64) (manageIDs []int64) {
	obj := models.Admin{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	sqlList := fmt.Sprintf("SELECT u.`id` FROM `%s` u LEFT JOIN `%s` r ON u.role_id=r.id  WHERE u.status=%d",
		models.ADMIN_TABLENAME, models.ROLE_TABLENAME, types.StatusValid)

	sqlList += fmt.Sprintf(" AND r.pid = %d", roleID)
	r := o.Raw(sqlList)
	_, err := r.QueryRows(&manageIDs)
	if err != nil {
		logs.Error(err)
	}
	if len(leaderUIDs) > 0 {
		manageIDs = append(manageIDs, leaderUIDs...)
	}
	return
}

// OperatorName 取操作员的名字
func OperatorName(opUid int64) string {
	if 0 == opUid {
		return "-"
	}

	return GetNameByID(opUid)
}
