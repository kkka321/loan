package models

// `admin`
import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/tools"
	"micro-loan/common/types"
)

const ADMIN_TABLENAME string = "admin"

type Admin struct {
	Id            int64 `orm:"pk;"`
	Email         string
	Mobile        string
	Nickname      string
	Password      string
	RoleID        int64 `orm:"column(role_id)"`
	Status        int
	WorkStatus    int
	ReducedQuota  int `orm:"column(reduced_quota)"`
	CreateUid     int64
	RegisterTime  int64 `orm:"column(register_time)"`
	LastLoginTime int64 `orm:"column(last_login_time)"`
}

// 此处声明为指针方法,并不会修改传入的对象,只是为了省去拷贝对象的开消

// 当前模型对应的表名
func (r *Admin) TableName() string {
	return ADMIN_TABLENAME
}

// 当前模型的数据库
func (r *Admin) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *Admin) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func OneAdminByUid(id int64) (Admin, error) {
	admin := &Admin{Id: id}
	o := orm.NewOrm()
	o.Using(admin.UsingSlave())

	err := o.Read(admin)

	return *admin, err
}

func OneAdminByNickName(nickName string) (Admin, error) {
	var admin Admin

	o := orm.NewOrm()
	o.Using(admin.UsingSlave())

	err := o.QueryTable(ADMIN_TABLENAME).Filter("nickname", nickName).One(&admin)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneAdminByNickName] sql error err:%v", err)
	}

	return admin, err
}

func OneAdminByEmail(email string) (Admin, error) {
	var admin Admin

	o := orm.NewOrm()
	o.Using(admin.UsingSlave())

	err := o.QueryTable(ADMIN_TABLENAME).Filter("email", email).One(&admin)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneAdminByEmail] sql error err:%v", err)
	}

	return admin, err
}

func CheckLoginIsValid(email string, password string) bool {
	admin, err := OneAdminByEmail(email)
	logs.Debug("admin:", admin)
	if err != nil || admin.Id <= 0 {
		logs.Warning("email and info does not exist:", email)
		return false
	}

	ciphertext := tools.PasswordEncrypt(password, admin.RegisterTime)
	if ciphertext == admin.Password {
		return true
	}

	logs.Warning("User information is incorrect, email:", email)
	return false
}

func UpdateLastLoginTime(id int64) {
	admin := Admin{
		Id:            id,
		LastLoginTime: tools.GetUnixMillis(),
	}
	o := orm.NewOrm()
	o.Using(admin.Using())

	o.Update(&admin, "last_login_time")
}

// 改
func Update(admin Admin) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(admin.Using())
	id, err = o.Update(&admin)
	if err != nil {
		logs.Error("model Admin UpdateRepayPlan failed.", err)
	}

	return
}

// GetUserIDsByRolePidFromDB 根据角色ID, 获取其下的用户ID slices
// 忽视工作状态, 封禁状态等
// 目前用于共享权限和共享工单给上级管理者
func GetUserIDsByRolePidFromDB(rolePid int64, container interface{}) (num int64, err error) {
	obj := Admin{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	sqlList := fmt.Sprintf("SELECT u.`id` FROM `%s` u LEFT JOIN `%s` r ON u.role_id=r.id  WHERE r.`pid` =%d",
		ADMIN_TABLENAME, ROLE_TABLENAME, rolePid)

	r := o.Raw(sqlList)
	num, err = r.QueryRows(container)
	return
}

// GetUserIDsByRoleIDStringsFromDB 根据角色ID, 获取其下的用户ID slices
// 用于ticket 自动分配列表, 直接获取在工作状态的就好
func GetUserIDsByRoleIDStringsFromDB(idStrings []string) (ids []int64, num int64, err error) {
	obj := Admin{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	sqlList := fmt.Sprintf("SELECT `id` FROM `%s` WHERE `role_id` in(%s) AND work_status=%d AND status=%d",
		ADMIN_TABLENAME, strings.Join(idStrings, ","), types.AdminWorkStatusNormal, 1)

	r := o.Raw(sqlList)
	num, err = r.QueryRows(&ids)
	return
}

// GetUsersByRoleIDStringsFromDB 根据角色ID, 获取其下的用户 slices
// 获取指定角色下的后台用户, 用于手动分配列表 和 下属工作状态管理
func GetUsersByRoleIDStringsFromDB(idStrings []string) (admins []Admin, num int64, err error) {
	obj := Admin{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	sqlList := fmt.Sprintf("SELECT `id`, `nickname`, `email`, `work_status` FROM `%s` WHERE `role_id` in(%s) AND status=%d",
		ADMIN_TABLENAME, strings.Join(idStrings, ","), 1)

	r := o.Raw(sqlList)
	num, err = r.QueryRows(&admins)
	return
}

// GetUsersByAssignIdFromDB 根据用户ID, 获取不包括这些ID的用户 slices
// 获取排除指定ID下的后台用户, 用于分配分机
func GetUsersByAssignIdFromDB(assignIds []string) (admins []Admin, num int64, err error) {
	obj := Admin{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	// 向数组中添加一个一定取不到的用户id
	assignIds = append(assignIds, "-1")
	sqlList := fmt.Sprintf("SELECT `id`, `nickname`, `email`, `work_status` FROM `%s` WHERE `id` not in(%s) AND status=%d AND work_status = %d",
		ADMIN_TABLENAME, strings.Join(assignIds, ","), 1, 1)

	r := o.Raw(sqlList)
	num, err = r.QueryRows(&admins)
	return
}
