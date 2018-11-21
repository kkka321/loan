package rbac

import (
	"errors"
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"strconv"

	"github.com/astaxie/beego/orm"
)

// ListPrivilegeGroup 根据条件返回 PrivilegeGroup 列表
func ListPrivilegeGroup(condContr map[string]interface{}) (list []models.PrivilegeGroup, err error) {
	// 初始化 orm,
	obj := models.PrivilegeGroup{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	// if page < 1 {
	// 	page = 1
	// }
	// if pagesize < 1 {
	// 	pagesize = service.Pagesize
	// }
	// offset := (page - 1) * pagesize

	// 初始化查询条件
	cond := "1=1"
	if v, ok := condContr["name"]; ok {
		cond += "  AND `name` LIKE('%" + tools.Escape(v.(string)) + "%')"
	}
	if v, ok := condContr["status"]; ok {
		cond += fmt.Sprintf(" AND `status` = %d", v)
	}
	//sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` WHERE %s", models.PRIVILEGE_GROUP_TABLENAME, cond)
	//sqlList := fmt.Sprintf("SELECT * FROM `%s` WHERE %s ORDER BY `id` desc LIMIT %d,%d", models.PRIVILEGE_GROUP_TABLENAME, cond, offset, pagesize)
	sqlList := fmt.Sprintf("SELECT * FROM `%s` WHERE %s ORDER BY `id` desc", models.PRIVILEGE_GROUP_TABLENAME, cond)
	// 查询符合条件的所有条数
	// r := o.Raw(sqlCount)
	// r.QueryRow(&total)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

// PrivilegeGroupNameCheck 角色名合法性检查
func PrivilegeGroupNameCheck(name string) error {
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

// PrivilegeGroupNameUnique 角色名唯一性校验
func PrivilegeGroupNameUnique(name string) error {
	// 唯一性校验
	// 初始化 orm,
	obj := models.PrivilegeGroup{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` WHERE name='%s' LIMIT 1", models.PRIVILEGE_GROUP_TABLENAME, name)
	// 查询符合条件的所有条数
	var total int
	r := o.Raw(sqlCount)
	r.QueryRow(&total)
	if total > 0 {
		return errors.New("Name already exist")
	}
	return nil
}

// AddOnePrivilegeGroup 新增角色
// 含属性校验, 外部可直接调用
// 内部自动初始化, Ctime 与 Utime
func AddOnePrivilegeGroup(m *models.PrivilegeGroup) (id int64, err error) {
	err = PrivilegeGroupNameCheck(m.Name)
	if err != nil {
		return
	}
	err = PrivilegeGroupNameUnique(m.Name)
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

// UpdateOnePrivilegeGroup 更新指定角色的属性
// 不含属性校验
// 内部自动更新 Utime
func UpdateOnePrivilegeGroup(m *models.PrivilegeGroup, cols ...string) (num int64, err error) {
	if len(m.Name) > 0 {
		if err = PrivilegeGroupNameCheck(m.Name); err != nil {
			return 0, err
		}
		if err = UpdatePrivilegeGroupNameUnique(m.Name, m.Id); err != nil {
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

// UpdatePrivilegeGroupNameUnique 角色名唯一性校验
func UpdatePrivilegeGroupNameUnique(name string, id int64) error {
	// 唯一性校验
	// 初始化 orm,
	obj := models.PrivilegeGroup{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` WHERE name='%s' AND id<>%d LIMIT 1", models.PRIVILEGE_GROUP_TABLENAME, name, id)
	// 查询符合条件的所有条数
	var total int
	r := o.Raw(sqlCount)
	r.QueryRow(&total)
	if total > 0 {
		return errors.New("Name already exist")
	}
	return nil
}

// GetOnePrivilegeGroup 获取指定ID的角色信息
func GetOnePrivilegeGroup(id int64) (data models.PrivilegeGroup, err error) {
	obj := models.PrivilegeGroup{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	qs := o.QueryTable(obj.TableName())

	err = qs.Filter("id", id).One(&data)

	return
}
