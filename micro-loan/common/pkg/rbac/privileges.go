package rbac

import (
	"errors"
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/orm"
)

// GroupPrivilege 分组的GroupPrivilege
type GroupPrivilege struct {
	Id        int64 `orm:"pk;"`
	Name      string
	GroupID   int64 `orm:"column(group_id);"`
	Ctime     int64
	Utime     int64
	GroupName string `orm:"column(group_name);"`
}

// PrivilegeList 根据条件返回 Privilege 列表
func PrivilegeList(condContr map[string]interface{}) (list []GroupPrivilege, err error) {
	// 初始化 orm,
	obj := models.Privilege{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	cond := "1=1"
	if v, ok := condContr["name"]; ok {
		cond += fmt.Sprintf(" AND  L.`name` LIKE('%s')", "%"+tools.Escape(v.(string))+"%")
	}

	sqlList := fmt.Sprintf("SELECT L.id, L.name,L.group_id, L.ctime, L.utime, R.name as group_name FROM `%s` L LEFT JOIN %s R ON L.group_id=R.id WHERE %s ORDER BY `id` desc",
		models.PRIVILEGE_TABLENAME, models.PRIVILEGE_GROUP_TABLENAME, cond)
	// 查询符合条件的所有条数
	// r := o.Raw(sqlCount)
	// r.QueryRow(&total)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

// ListPrivilege 返回
func ListPrivilege(condCntr map[string]interface{}, page int, pagesize int) (list []GroupPrivilege, total int64, err error) {
	obj := models.Privilege{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = types.DefaultPagesize
	}
	offset := (page - 1) * pagesize

	// 初始化查询条件
	cond := "1=1"
	if v, ok := condCntr["name"]; ok {
		cond += fmt.Sprintf(" AND  L.`name` LIKE('%s')", "%"+tools.Escape(v.(string))+"%")
	}
	if v, ok := condCntr["group_id"]; ok {
		cond += fmt.Sprintf("  AND L.`group_id` = %d", v)
	}

	sqlList := fmt.Sprintf("SELECT L.id, L.name,L.group_id, L.ctime, L.utime, R.name as group_name FROM `%s`  L LEFT JOIN %s R ON L.group_id=R.id WHERE %s ORDER BY L.`id` desc LIMIT %d,%d",
		models.PRIVILEGE_TABLENAME, models.PRIVILEGE_GROUP_TABLENAME, cond, offset, pagesize)

	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` L WHERE %s", models.PRIVILEGE_TABLENAME, cond)
	//sqlList := fmt.Sprintf("SELECT * FROM `%s` WHERE %s ORDER BY `id` desc LIMIT %d,%d", models.PRIVILEGE_TABLENAME, cond, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

// PrivilegeNameUnique 名称唯一性校验
func PrivilegeNameUnique(name string) error {
	// 唯一性校验
	// 初始化 orm,
	obj := models.Privilege{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` WHERE name='%s' LIMIT 1", models.PRIVILEGE_TABLENAME, name)
	// 查询符合条件的所有条数
	var total int
	r := o.Raw(sqlCount)
	r.QueryRow(&total)
	if total > 0 {
		return errors.New("Name already exist")
	}
	return nil
}

// AddOnePrivilege 新增权限
// 含属性校验, 外部可直接调用
// 内部自动初始化, Ctime 与 Utime
func AddOnePrivilege(m *models.Privilege) (id int64, err error) {
	// err = PrivilegeNameUnique(m.Name)
	// if err != nil {
	// 	return
	// }

	o := orm.NewOrm()
	o.Using(m.Using())
	//
	m.Ctime = tools.GetUnixMillis()
	m.Utime = m.Ctime
	id, err = o.Insert(m)

	return
}

// UpdateOnePrivilege 更新指定角色的属性
// 不含属性校验
// 内部自动更新 Utime
func UpdateOnePrivilege(m *models.Privilege, cols ...string) (num int64, err error) {
	// if len(m.Name) > 0 {
	// 	if err = UpdatePrivilegeNameUnique(m.Name, m.Id); err != nil {
	// 		return 0, err
	// 	}
	// }
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

// UpdatePrivilegeNameUnique 名称唯一性校验
func UpdatePrivilegeNameUnique(name string, id int64) error {
	// 唯一性校验
	// 初始化 orm,
	obj := models.Privilege{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` WHERE name='%s' and id!=%d LIMIT 1", models.PRIVILEGE_TABLENAME, name, id)
	// 查询符合条件的所有条数
	var total int
	r := o.Raw(sqlCount)
	r.QueryRow(&total)
	if total > 0 {
		return errors.New("Name already exist")
	}
	return nil
}

// GetOnePrivilege 获取指定ID的角色信息
func GetOnePrivilege(id int64) (data models.Privilege, err error) {
	obj := models.Privilege{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	qs := o.QueryTable(obj.TableName())

	err = qs.Filter("id", id).One(&data)

	return
}

// PrivilegeGroupList 根据条件返回 Privilege 列表
func PrivilegeGroupList() (list []models.PrivilegeGroup, err error) {
	// 初始化 orm,
	obj := models.PrivilegeGroup{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	sqlList := fmt.Sprintf("SELECT * FROM `%s` ORDER BY `id` desc", models.PRIVILEGE_GROUP_TABLENAME)
	// 查询符合条件的所有条数
	// r := o.Raw(sqlCount)
	// r.QueryRow(&total)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRows(&list)

	return
}
