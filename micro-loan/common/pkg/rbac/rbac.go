package rbac

import (
	"fmt"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"encoding/json"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type singleDynamicRes struct {
	name string //unique
	key  string
}

type dynamicResPool struct {
}

type dynamicRes struct {
	authorizedResCode string
	name              string
}

// dynamic res 动态资源管理
type authorizedRes interface {
	IsAuthorized(resName string) bool
	showAllRes() []singleDynamicRes
}

// PrivilegeForRole 指定权限是否分配给角色
type PrivilegeForRole struct {
	Id         int64  `json:"id"`
	GroupName  string `json:"groupName",orm:"column(group_name)"`
	Name       string `json:"name"`
	IsAssigned int    `json:"isAssigned",orm:"column(is_assigned);"`
}

// OperationForPrivilege 指定权限是否分配给角色
type OperationForPrivilege struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	IsAssigned int    `json:"isAssigned",orm:"column(is_assigned);"`
}

// GetRoleOperationIDMap 查询角色拥有的所有操作ID为键的map
// 先查询 Redis Hash
func GetRoleOperationIDMap(roleID int64) (pidMap map[int64]bool, err error) {
	// get RoleOpearation From redis
	redisStorageClient := storage.RedisStorageClient.Get()
	defer redisStorageClient.Close()
	hashKey := beego.AppConfig.String("rbac_roles_operations")
	valBytes, err := redisStorageClient.Do("HGET", hashKey, roleID)

	if err == nil && valBytes != nil {
		json.Unmarshal(valBytes.([]byte), &pidMap)
		return
	}

	pidMap, err = GetRoleOperationIDMapFromDB(roleID)
	jsonBytes, _ := json.Marshal(pidMap)
	redisStorageClient.Do("HSET", hashKey, roleID, jsonBytes)
	return
}

func clearRedisRoleOperations(roleID int64) {
	redisStorageClient := storage.RedisStorageClient.Get()
	defer redisStorageClient.Close()
	hashKey := beego.AppConfig.String("rbac_roles_operations")
	_, err := redisStorageClient.Do("HDEL", hashKey, roleID)
	if err != nil {
		logs.Error(err)
	}
}

func clearAllRedisRoleOperations() {
	redisStorageClient := storage.RedisStorageClient.Get()
	defer redisStorageClient.Close()
	hashKey := beego.AppConfig.String("rbac_roles_operations")
	_, err := redisStorageClient.Do("DEL", hashKey)
	if err != nil {
		logs.Error(err)
	}
}

// GetRoleOperationIDMapFromDB 从数据库查询该角色拥有的所有操作ID为键的map
// 自动附加基础权限
func GetRoleOperationIDMapFromDB(roleID int64) (pidMap map[int64]bool, err error) {
	// 初始化 orm,
	obj := models.Privilege{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	sqlList := fmt.Sprintf("SELECT DISTINCT R.`operation_id` FROM `%s` L LEFT JOIN `%s` R ON L.`privilege_id` = R.`privilege_id` WHERE L.role_id=%d",
		models.ROLE_PRIVILEGE_TABLENAME, models.PRIVILEGE_OPERATION_TABLENAME, roleID)
	//select DISTINCT R.operation_id from role_privilege L left join privilege_operation R on L.privilege_id=R.privilege_id where L.role_id = 3;

	// 查询指定页
	var list []int64
	r := o.Raw(sqlList)
	r.QueryRows(&list)
	pidMap = map[int64]bool{}
	for _, v := range list {
		pidMap[v] = true
	}

	baseOperationsList, err := GetOperationIDsByNames(types.RBACBaseOpeartionList)
	for _, v := range baseOperationsList {
		pidMap[v] = true
	}

	return
}

// GetRolePrivileges 获取角色所有权限列表
func GetRolePrivileges(roleID int64) (list []models.Privilege, err error) {
	// 初始化 orm,
	obj := models.Privilege{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	sqlList := fmt.Sprintf("SELECT S.`id`, S.`name` FROM `%s` L LEFT JOIN `%s` R ON L.`privilege_id` = R.`id` WHERE L.role_id=%d",
		models.ROLE_PRIVILEGE_TABLENAME, models.PRIVILEGE_TABLENAME, roleID)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

// AllPrivilegesForRole 获取所有权限对该角色的分配情况, 已分配/未分配
// 此处读取发生在分配请求中, 先保留主库读取
func AllPrivilegesForRole(roleID int64) (list []PrivilegeForRole, err error) {
	// 初始化 orm,
	obj := models.Privilege{}
	o := orm.NewOrm()
	// 此处读取发生在分配请求中, 先保留主库读取
	o.Using(obj.Using())

	sqlList := fmt.Sprintf("SELECT L.`id`, R2.`name` as group_name, L.`name`, IF(R.id>0, 1, 0) as is_assigned FROM `%s` L",
		models.PRIVILEGE_TABLENAME)
	sqlList += fmt.Sprintf(" LEFT JOIN `%s` R ON R.`privilege_id` = L.`id` AND R.role_id=%d  LEFT JOIN `%s` R2 ON R2.id=L.group_id",
		models.ROLE_PRIVILEGE_TABLENAME, roleID, models.PRIVILEGE_GROUP_TABLENAME)
	sqlList += " ORDER BY R2.name desc, L.name desc"

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

// AssignPrivilegesToRole 分配权限给角色
func AssignPrivilegesToRole(privileges []string, roleID int64) (iNum int64, existNum int64, err error) {
	var existIds []int64
	existIds, existNum, err = getAllExistPrivilegesID(privileges)

	if err != nil {
		return
	}

	var rolePrivileges []models.RolePrivilege
	ctime := tools.GetUnixMillis()
	for _, v := range existIds {
		rolePrivileges = append(rolePrivileges, models.RolePrivilege{RoleID: roleID, PrivilegeID: v, Ctime: ctime})
	}

	obj := models.Privilege{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	iNum, err = o.InsertMulti(len(rolePrivileges), rolePrivileges)
	if iNum > 0 {
		clearRedisRoleOperations(roleID)
	}
	return
}

// RevokePrivilegesFromRole 从角色中移除权限
func RevokePrivilegesFromRole(privileges []string, roleID int64) (affectedRows int64, err error) {
	// var existIds []int64
	// existIds, existNum, err = getAllExistOperationsID(operations)

	if err != nil {
		return
	}

	privilegeIDStr := strings.Join(privileges, ",")

	obj := models.RolePrivilege{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	sql := fmt.Sprintf("DELETE FROM %s WHERE `role_id`=%d and `privilege_id` IN(%s)",
		models.ROLE_PRIVILEGE_TABLENAME, roleID, privilegeIDStr)
	r := o.Raw(sql)

	sqlResult, _ := r.Exec()
	affectedRows, err = sqlResult.RowsAffected()
	if affectedRows > 0 {
		clearRedisRoleOperations(roleID)
	}
	return
}

func getAllExistPrivilegesID(privileges []string) (ids []int64, num int64, err error) {

	privilegesIDStr := strings.Join(privileges, ",")

	// 初始化 orm,
	obj := models.Privilege{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	sqlList := fmt.Sprintf("SELECT `id` FROM `%s` WHERE id IN(%s)", models.PRIVILEGE_TABLENAME, privilegesIDStr)

	// 查询指定页
	r := o.Raw(sqlList)
	num, err = r.QueryRows(&ids)

	return
}

// AllOperationsForPrivilege 获取所有权限对该角色的分配情况, 已分配/未分配
// 建议走主库, 此处查询部分时候会发生在插入之后瞬间发生, 若主从有延迟, 会出现数据不一致
func AllOperationsForPrivilege(privilegeID int64) (list []OperationForPrivilege, err error) {
	// 初始化 orm,
	obj := models.Privilege{}
	o := orm.NewOrm()
	// 建议走主库, 此处查询部分时候会发生在插入之后瞬间发生, 若主从有延迟, 会出现数据不一致
	o.Using(obj.Using())

	sqlList := fmt.Sprintf("SELECT L.`id`, L.`name`, IF(R.id>0, 1, 0) as is_assigned FROM `%s` L LEFT JOIN `%s` R ON R.`operation_id` = L.`id` AND R.privilege_id=%d ORDER BY name ASC",
		models.OPERATION_TABLENAME, models.PRIVILEGE_OPERATION_TABLENAME, privilegeID)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

// AssignOperationsToPrivilege 分配权限给角色
func AssignOperationsToPrivilege(operations []string, privilegeID int64) (iNum int64, existNum int64, err error) {
	var existIds []int64
	existIds, existNum, err = getAllExistOperationsID(operations)

	if err != nil {
		return
	}

	var privilegeOperations []models.PrivilegeOperation
	ctime := tools.GetUnixMillis()
	for _, v := range existIds {
		privilegeOperations = append(privilegeOperations, models.PrivilegeOperation{PrivilegeID: privilegeID, OperationID: v, Ctime: ctime})
	}

	obj := models.PrivilegeOperation{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	iNum, err = o.InsertMulti(len(privilegeOperations), privilegeOperations)
	if iNum > 0 {
		clearAllRedisRoleOperations()
	}
	return
}

// RevokeOperationsFromPrivilege 从权限中移除操作
func RevokeOperationsFromPrivilege(operations []string, privilegeID int64) (affectedRows int64, err error) {
	// var existIds []int64
	// existIds, existNum, err = getAllExistOperationsID(operations)

	if err != nil {
		return
	}

	operationIDStr := strings.Join(operations, ",")

	obj := models.PrivilegeOperation{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	sql := fmt.Sprintf("DELETE FROM %s WHERE `privilege_id`=%d and `operation_id` IN(%s)",
		models.PRIVILEGE_OPERATION_TABLENAME, privilegeID, operationIDStr)
	r := o.Raw(sql)

	sqlResult, _ := r.Exec()
	affectedRows, err = sqlResult.RowsAffected()
	if affectedRows > 0 {
		clearAllRedisRoleOperations()
	}
	return
}

func getAllExistOperationsID(operations []string) (ids []int64, num int64, err error) {

	operationIDStr := strings.Join(operations, ",")

	// 初始化 orm,
	obj := models.Privilege{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	sqlList := fmt.Sprintf("SELECT `id` FROM `%s` WHERE id IN(%s)", models.OPERATION_TABLENAME, operationIDStr)

	// 查询指定页
	r := o.Raw(sqlList)
	num, err = r.QueryRows(&ids)

	return
}
