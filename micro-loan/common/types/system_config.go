package types

type SystemConfigItemType int

const (
	SystemConfigItemTypeInt     = 1
	SystemConfigItemTypeInt64   = 2
	SystemConfigItemTypeFloat64 = 3
	SystemConfigItemTypeBool    = 4
	SystemConfigItemTypeString  = 5
)

var systemConfigItemTypeMap = map[SystemConfigItemType]string{
	SystemConfigItemTypeInt:     "int",
	SystemConfigItemTypeInt64:   "int64",
	SystemConfigItemTypeFloat64: "float64",
	SystemConfigItemTypeBool:    "bool",
	SystemConfigItemTypeString:  "string",
}

func SystemConfigItemTypeMap() map[SystemConfigItemType]string {
	return systemConfigItemTypeMap
}
