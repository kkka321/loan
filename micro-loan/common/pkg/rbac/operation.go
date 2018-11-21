package rbac

import (
	"errors"
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"github.com/astaxie/beego/orm"
)

// ListOperation 返回
func ListOperation(condCntr map[string]interface{}, page int, pagesize int) (list []models.Operation, total int64, err error) {
	obj := models.Operation{}
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
		cond += "  AND `name` LIKE('%" + tools.Escape(v.(string)) + "%')"
	}

	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` WHERE %s", models.OPERATION_TABLENAME, cond)
	sqlList := fmt.Sprintf("SELECT * FROM `%s` WHERE %s ORDER BY `id` desc LIMIT %d,%d", models.OPERATION_TABLENAME, cond, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

// OperationNameUnique 名称唯一性校验
func OperationNameUnique(name string) error {
	// 唯一性校验
	// 初始化 orm,
	obj := models.Operation{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` WHERE name='%s' LIMIT 1", models.OPERATION_TABLENAME, name)
	// 查询符合条件的所有条数
	var total int
	r := o.Raw(sqlCount)
	r.QueryRow(&total)
	if total > 0 {
		return errors.New("Name already exist")
	}
	return nil
}

// AddOneOperation 新增权限
// 含属性校验, 外部可直接调用
// 内部自动初始化, Ctime 与 Utime
func AddOneOperation(m *models.Operation) (id int64, err error) {
	err = OperationNameUnique(m.Name)
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

// UpdateOneOperation 更新指定角色的属性
// 不含属性校验
// 内部自动更新 Utime
func UpdateOneOperation(m *models.Operation, cols ...string) (num int64, err error) {
	if len(m.Name) > 0 {
		if err = UpdateOperationNameUnique(m.Name, m.Id); err != nil {
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

// UpdateOperationNameUnique 名称唯一性校验
func UpdateOperationNameUnique(name string, id int64) error {
	// 唯一性校验
	// 初始化 orm,
	obj := models.Operation{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` WHERE name='%s' AND id<>%d LIMIT 1", models.OPERATION_TABLENAME, name, id)
	// 查询符合条件的所有条数
	var total int
	r := o.Raw(sqlCount)
	r.QueryRow(&total)
	if total > 0 {
		return errors.New("Name already exist")
	}
	return nil
}

// GetOneOperation 获取指定ID的角色信息
func GetOneOperation(id int64) (data models.Operation, err error) {
	obj := models.Operation{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	qs := o.QueryTable(obj.TableName())

	err = qs.Filter("id", id).One(&data)

	return
}

// GetOperationIDsByNames 根据名称返回指定的 operationIDs
// 此处给获取
func GetOperationIDsByNames(name []string) (operationIDs []int64, err error) {

	var p models.Operation
	o := orm.NewOrm()
	o.Using(p.UsingSlave())

	nameStr := strings.Join(name, "','")

	sql := fmt.Sprintf("SELECT `id` FROM `%s` WHERE name in('%s')", models.OPERATION_TABLENAME, nameStr)

	r := o.Raw(sql)
	_, err = r.QueryRows(&operationIDs)

	return
}
