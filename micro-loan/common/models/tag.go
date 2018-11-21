package models

import "micro-loan/common/types"

const TAG_TABLENAME = "tag"

// Tag 描述映射表 tag
type Tag struct {
	Id        int64 `orm:"pk;"`
	Name      string
	Type      int
	IsDeleted int
	Ctime     int64
	Utime     int64
}

// TableName 当前模型对应的表名
func (r *Tag) TableName() string {
	return TAG_TABLENAME
}

// Using 当前模型的主库
func (r *Tag) Using() string {
	return types.OrmDataBaseAdmin
}

// UsingSlave 返回从库
func (r *Tag) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}
