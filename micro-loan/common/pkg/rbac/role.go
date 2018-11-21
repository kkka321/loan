package rbac

import (
	"encoding/json"
	"errors"
	"fmt"
	"micro-loan/common/lib/redis/cache"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
)

// RoleTreeNode describe a Role node , 包含子节点slice, 用于角色管理
type RoleTreeNode struct {
	models.Role
	SubList  []RoleTreeNode
	IsActive bool
}

// RoleList 根据条件返回 Role 列表
func RoleList(condContr map[string]interface{}) (list []models.Role, err error) {
	// 初始化 orm,
	obj := models.Role{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	// 初始化查询条件
	cond := "1=1"
	if v, ok := condContr["name"]; ok {
		cond += "  AND `name` LIKE('%" + tools.Escape(v.(string)) + "%')"
	}
	if v, ok := condContr["status"]; ok {
		cond += fmt.Sprintf(" AND `status` = %d", v)
	}
	sqlList := fmt.Sprintf("SELECT * FROM `%s` WHERE %s ORDER BY `pid` DESC, id ASC", models.ROLE_TABLENAME, cond)

	r := o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

// LowPrivilegeRoleList 根据条件返回 Role 列表
func LowPrivilegeRoleList(condContr map[string]interface{}) (list []models.Role, err error) {
	// 初始化 orm,
	obj := models.Role{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	lowPriviRoleString, _ := tools.IntsSliceToWhereInString(types.LowPrivilegeRoleTypeContainer())

	// 初始化查询条件
	cond := "1=1"
	cond += fmt.Sprintf(" AND `type` in(%s)", lowPriviRoleString)
	cond += fmt.Sprintf(" AND `pid` !=%d", types.RoleSuperPid)

	if v, ok := condContr["name"]; ok {
		cond += "  AND `name` LIKE('%" + tools.Escape(v.(string)) + "%')"
	}
	if v, ok := condContr["status"]; ok {
		cond += fmt.Sprintf(" AND `status` = %d", v)
	}
	sqlList := fmt.Sprintf("SELECT * FROM `%s` WHERE %s ORDER BY `pid` DESC, id ASC", models.ROLE_TABLENAME, cond)

	r := o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

// RoleNameCheck 角色名合法性检查
func RoleNameCheck(name string) error {
	const MaxLength = 30
	const MinLength = 3
	if len(name) > MaxLength {
		return errors.New("Name length must less then " + strconv.Itoa(MaxLength))
	}
	if len(name) < MinLength {
		return errors.New("Name length must less then " + strconv.Itoa(MaxLength))
	}
	return nil
}

// RoleNameUnique 角色名唯一性校验
func RoleNameUnique(name string) error {
	// 唯一性校验
	// 初始化 orm,
	obj := models.Role{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` WHERE name='%s' LIMIT 1", models.ROLE_TABLENAME, name)
	// 查询符合条件的所有条数
	var total int
	r := o.Raw(sqlCount)
	r.QueryRow(&total)
	if total > 0 {
		return errors.New("Name already exist")
	}
	return nil
}

// AddOneRole 新增角色
// 含属性校验, 外部可直接调用
// 内部自动初始化, Ctime 与 Utime
func AddOneRole(m *models.Role) (id int64, err error) {
	err = RoleNameCheck(m.Name)
	if err != nil {
		return
	}
	err = RoleNameUnique(m.Name)
	if err != nil {
		return
	}

	o := orm.NewOrm()
	o.Using(m.Using())
	//
	m.Ctime = tools.GetUnixMillis()
	m.Utime = m.Ctime
	id, err = o.Insert(m)

	return
}

// UpdateOneRole 更新指定角色的属性
// 不含属性校验
// 内部自动更新 Utime
func UpdateOneRole(m *models.Role, cols ...string) (num int64, err error) {
	role, _ := models.GetOneRole(m.Id)
	if len(m.Name) > 0 && role.Name != m.Name {
		if err = RoleNameCheck(m.Name); err != nil {
			return 0, err
		}
		if err = UpdateRoleNameUnique(m.Name, m.Id); err != nil {
			return 0, err
		}

	}
	if m.Id <= 0 {
		err = errors.New("Update ID must exist and >0")
		return
	}

	m.Utime = tools.GetUnixMillis()
	cols = append(cols, "Utime")
	o := orm.NewOrm()
	o.Using(m.Using())
	num, err = o.Update(m, cols...)
	return
}

// UpdateRoleNameUnique 角色名唯一性校验
func UpdateRoleNameUnique(name string, id int64) error {
	// 唯一性校验
	// 初始化 orm,
	obj := models.Role{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` WHERE name='%s' AND id<>%d LIMIT 1", models.ROLE_TABLENAME, name, id)
	// 查询符合条件的所有条数
	var total int
	r := o.Raw(sqlCount)
	r.QueryRow(&total)
	if total > 0 {
		return errors.New("Name already exist")
	}
	return nil
}

// GetOneRole 获取指定ID的角色信息
func GetOneRole(id int64) (data models.Role, err error) {
	obj := models.Role{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	qs := o.QueryTable(obj.TableName())

	err = qs.Filter("id", id).One(&data)

	return
}

// RoleTree 返回一个完整的角色树
func RoleTree() (treeList []RoleTreeNode, err error) {
	redisCache := cache.RedisCacheClient.Get()
	defer redisCache.Close()
	// cacheKey := beego.AppConfig.String("rbac_role_tree")
	// valBytes, err := redisCache.Do("GET", cacheKey)
	// if err == nil && valBytes != nil {
	// 	json.Unmarshal(valBytes.([]byte), &treeList)
	// 	return
	// }

	// 从数据库获取menu数据, 并
	{
		allList, _ := RoleList(nil)
		logs.Debug("pretty print: %#v", allList)
		mapSameLayerList := map[int64][]RoleTreeNode{}
		for _, v := range allList {
			var node RoleTreeNode
			if subList, ok := mapSameLayerList[v.Id]; ok {
				// subList 存在, 则将 subList 合入父亲节点
				node = RoleTreeNode{v, subList, false}
				delete(mapSameLayerList, v.Id)
			} else {
				node = RoleTreeNode{v, nil, false}
			}
			if set, ok := mapSameLayerList[v.Pid]; ok {
				mapSameLayerList[v.Pid] = append(set, node)
			} else {
				mapSameLayerList[v.Pid] = []RoleTreeNode{node}
			}
		}
		if len(mapSameLayerList) > 1 {
			err = errors.New("should be one tree")
			return
		}

		if v, ok := mapSameLayerList[types.RoleSuperPid]; ok {
			treeList = v
		} else {
			err = errors.New("No root node")
		}
	}

	// // 执行到此处说明已经生成 menu tree
	// jsonBytes, _ := json.Marshal(treeList)
	// redisCache.Do("SET", cacheKey, jsonBytes)

	return
}

// ActiveRoleTreeForUpdate 返回已授权的 menu tree
// active node 已经被标记
func ActiveRoleTreeForUpdate(id int64) (wholeTree []RoleTreeNode, err error) {
	wholeTree, _ = RoleTree()
	for i, v := range wholeTree {
		// 含有子节点
		if id == v.Id {
			wholeTree[i].IsActive = true
			return
		}
		if len(v.SubList) > 0 {
			// 深度拷贝 v 至 topNode

			for secI, secV := range v.SubList {
				// 三级节点
				if id == secV.Id {
					wholeTree[i].SubList[secI].IsActive = true
					wholeTree[i].IsActive = true
					return
				}
				if len(secV.SubList) > 0 {
					for thirdI, thirdV := range secV.SubList {
						if id == thirdV.Id {
							wholeTree[i].SubList[secI].SubList[thirdI].IsActive = true
							wholeTree[i].SubList[secI].IsActive = true
							wholeTree[i].IsActive = true
							return
						}
					}

				}
			}

		}
	}
	return
}

// RoleCache 角色缓存数据描述
type RoleCache struct {
	RoleLevel int                `json:"l"`
	Name      string             `json:"n"`
	Pid       int64              `json:"p"`
	Type      types.RoleTypeEnum `json:"t"`
}

// IsLeaderRoleAndBeyond 是不是超过leader角色， 目前包括 leader和超管
func IsLeaderRoleAndBeyond(roleID int64, rolePid int64) bool {
	//
	r := GetRoleCache(roleID)
	if r.RoleLevel == types.RoleSuper || r.RoleLevel == types.RoleLeader {
		return true
	}
	return false
}

// GetRoleLevel 是不是超过leader角色， 目前包括 leader和超管
func GetRoleLevel(roleID int64) (roleLevel int) {
	r := GetRoleCache(roleID)
	roleLevel = r.RoleLevel
	return
}

// GetRoleName 是不是超过leader角色， 目前包括 leader和超管
func GetRoleName(roleID int64) (name string) {
	r := GetRoleCache(roleID)
	name = r.Name
	return
}

// GetRoleCache 获取角色cache
func GetRoleCache(roleID int64) (r RoleCache) {
	redisStorageClient := storage.RedisStorageClient.Get()
	defer redisStorageClient.Close()
	hashKey := beego.AppConfig.String("rbac_role_level_hash")
	reply, errReply := redis.Bytes(redisStorageClient.Do("HGET", hashKey, roleID))
	if errReply == nil {
		err := json.Unmarshal(reply, &r)
		if err == nil {
			return
		}
	}

	if errReply != redis.ErrNil {
		logs.Error("[GetRole] redis err:", errReply)
	}

	// get leader result from db
	// then set and return
	//
	role, errQ := models.GetOneRole(roleID)
	if errQ != nil {
		err := fmt.Errorf("[GetRoleCache]cannot find role by id:%d, err:%v", roleID, errQ)
		logs.Error(err)
		return
	}
	r.Name = role.Name
	r.Pid = role.Pid
	r.Type = role.Type
	if role.Pid == types.SuperRolePid {
		r.RoleLevel = types.RoleSuper
	} else {
		pRole, dbErr := models.GetOneRole(role.Pid)
		if dbErr != nil {
			err := fmt.Errorf("[GetRoleCache]cannot find parent role by pid:%d, err:%v", role.Pid, dbErr)
			logs.Error(err)
			return
		}
		if pRole.Pid == types.SuperRolePid {
			r.RoleLevel = types.RoleLeader
		} else {
			r.RoleLevel = types.RoleEmployee
		}
	}
	rc, _ := json.Marshal(r)

	redisStorageClient.Do("HSET", hashKey, roleID, rc)
	return
}

// GetChildRoles 获取所有孩子节点
func GetChildRoles(pid int64) (ids []int64) {
	obj := models.Role{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	qs := o.QueryTable(obj.TableName())
	var list []models.Role
	qs.Filter("pid", pid).All(&list, "id")
	for _, role := range list {
		ids = append(ids, role.Id)
	}
	return
}

type RoleUserStru struct {
	RoleID    int64
	RoleLevel int
	LeaderUID int64
	Members   []int64
}

func GetUserRoleMapByRoleType() {

}
