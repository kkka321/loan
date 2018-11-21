package models

import (
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

const ADMIN_LOGIN_LOG_TABLENAME string = "admin_login_log"

type AdminLoginLog struct {
	Id       int64  `orm:"pk;"`
	AdminUID int64  `orm:"column(admin_uid)"`
	IP       string `orm:"column(ip)"`
	Ctime    int64
}

func (*AdminLoginLog) TableName() string {
	return ADMIN_LOGIN_LOG_TABLENAME
}

func (*AdminLoginLog) Using() string {
	return types.OrmDataBaseAdmin
}

func (*AdminLoginLog) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func AddLoginLog(adminUID int64, ip string) (int64, error) {
	obj := AdminLoginLog{AdminUID: adminUID, IP: ip, Ctime: tools.GetUnixMillis()}

	return OrmInsert(&obj)
}
