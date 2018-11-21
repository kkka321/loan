// 模型层的接口定义
package models

type IModels interface {
	// 取模型对应的表名
	TableName() string
	// 选择数据库
	Using() string

	// 通过主键取一条数据
	OneByPkID() (error)
	// 通过主键增加一条数据
	AddOne() (id int64, err error)
	// 更新一条数据,可指定字段
	Update(cols ...string) (id int64, err error)
}
