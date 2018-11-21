package models

import (
	"micro-loan/common/types"
)

const SYSTEM_CONFIG_TABLENAME string = "system_config"

type SystemConfig struct {
	Id          int64 `orm:"pk;"`
	ItemName    string
	Description string
	ItemType    types.SystemConfigItemType
	ItemValue   string
	Weight      int
	Version     int
	Status      int
	OnlineTime  int64
	OfflineTime int64
	OpUid       int64 `orm:"column(op_uid)"`
	Ctime       int64
	Utime       int64
}

func (r *SystemConfig) TableName() string {
	return SYSTEM_CONFIG_TABLENAME
}

func (r *SystemConfig) Using() string {
	return types.OrmDataBaseAdmin
}

func (r *SystemConfig) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}
