package models

import "micro-loan/common/types"

// MENU_TABLENAME 表名
const MENU_TABLENAME string = "menu"

// Menu 描述数据表结构与结构体的映射
type Menu struct {
	Id          int64 `orm:"pk;"`
	Name        string
	Pid         int64
	Sort        int
	Path        string
	Class       string
	PrivilegeId int64 `orm:"column(privilege_id)"`
	Status      int
	Ctime       int64
	Utime       int64
}

// TableName 返回当前模型对应的表名
func (r *Menu) TableName() string {
	return MENU_TABLENAME
}

// Using 返回当前模型的数据库
func (r *Menu) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *Menu) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}
