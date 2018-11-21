package models

import (
	"fmt"

	"github.com/astaxie/beego/orm"
)

// 辅助方法, 为减少 orm的初始化 和 m.Using 操作
// 减除冗余代码

// OrmModelPt ...
type OrmModelPt interface {
	Using() string
}

// OrmInsert orm 插入
// m 必须为 model struct 的指针
// id 为自增主键的值
func OrmInsert(m OrmModelPt) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(m.Using())

	id, err = o.Insert(m)
	return
}

// OrmAllUpdate 全字段更新
// 若要使用全字段更新, m 必须为刚读出来的即时数据, 否则容易出现其他并发更新被覆盖
// m 必须为 model struct 的指针
// num 为 Affected Rows
func OrmAllUpdate(m OrmModelPt) (num int64, err error) {
	o := orm.NewOrm()
	o.Using(m.Using())

	num, err = o.Update(m)
	return
}

// OrmUpdate 特定字段更新
// 部分更新, 若是全字段更新, 请用 OrmAllUpdate
func OrmUpdate(m OrmModelPt, cols []string) (num int64, err error) {
	if len(cols) == 0 {
		// 部分更新, 必须指明字段, 防止错误更新
		err = fmt.Errorf("[OrmUpdate] can't do update with empty cols, %v", m)
		return
	}
	o := orm.NewOrm()
	o.Using(m.Using())

	num, err = o.Update(m, cols...)
	return
}

// OrmDelete 删除对象
func OrmDelete(m OrmModelPt) (num int64, err error) {
	o := orm.NewOrm()
	o.Using(m.Using())

	num, err = o.Delete(m)
	return
}
