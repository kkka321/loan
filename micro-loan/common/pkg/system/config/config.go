package config

// config 为当前系统 config,
// 存储在数据库中, redis 做 hash 缓存
// 需先在数据库中配置, 否则取出为当前取出类型的默认值, 如 int 为 0

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

// OneByPkID 根据主键ID获取配置
func OneByPkID(id int64) (one models.SystemConfig, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	one.Id = id
	err = o.Read(&one)

	if err != nil {
		logs.Error("[SystemConfig-OneByPkID] has wrong, id:", id, ", err:", err)
	}

	return
}

// List 获取配置列表
func List(condBox map[string]interface{}) (list []models.SystemConfig, num int64, err error) {
	one := models.SystemConfig{}
	o := orm.NewOrm()
	o.Using(one.Using())

	where := "WHERE 1 = 1"

	if v, ok := condBox["status"]; ok {
		where = fmt.Sprintf("%s AND status = %d", where, v.(int))
	}
	if v, ok := condBox["item_name"]; ok {
		where = fmt.Sprintf("%s AND item_name LIKE '%%%s%%'", where, v.(string))
	}

	orderBy := fmt.Sprintf("ORDER BY weight ASC, id ASC")
	sql := fmt.Sprintf("SELECT * FROM %s %s %s", one.TableName(), where, orderBy)

	num, err = o.Raw(sql).QueryRows(&list)

	return
}

// Create 创建配置
func Create(itemName, itemValue string, itemType types.SystemConfigItemType, weight int, description string, opid int64) (id int64, err error) {
	item := models.SystemConfig{}
	o := orm.NewOrm()
	o.Using(item.Using())

	// 1. 查看是否存在旧的有效配置,则将其置为下线状态
	originID, originValue, err := getValidItemValueFromDB(itemName)
	if originID > 0 {
		if originValue == itemValue {
			// 存在相同值的有效配置,则什么也不做
			logs.Warning("[SystemConfigCreate] 存在相同值的有效配置, itemName:", itemName, ", itemValue:", itemValue, ", originID:", originID)
			id = originID
			return
		}
		item, err = OneByPkID(originID)
		if err != nil {
			return
		}

		// 下线
		item.Status = types.StatusInvalid
		item.OfflineTime = tools.GetUnixMillis()
		item.Utime = item.OfflineTime
		o.Update(&item)
	}

	// 2. 新增配置并写redis hash
	version, _ := getItemVersion(itemName)
	version++
	newItem := models.SystemConfig{
		ItemName:    itemName,
		Description: description,
		ItemValue:   itemValue,
		ItemType:    itemType,
		Weight:      weight,
		Version:     version,
		Status:      types.StatusValid,
		OpUid:       opid,
		OnlineTime:  tools.GetUnixMillis(),
	}
	newItem.Ctime = newItem.OnlineTime
	newItem.Utime = newItem.Ctime

	id, err = o.Insert(&newItem)
	if err != nil {
		logs.Warning("[SystemConfigCreate] 创建新配置失败, newItem: %#v, err: %#v", newItem, err)
		return
	}

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	hashName := hashKey()
	storageClient.Do("HSET", hashName, itemName, itemValue)

	return
}

// ValidItemInt 获取 int 配置项
func ValidItemInt(itemName string) (v int, err error) {
	value, err := getValidItemValue(itemName, types.SystemConfigItemTypeInt)
	if err == nil {
		v = value.(int)
	}

	return
}

// ValidItemInt64 获取 int64 配置项
func ValidItemInt64(itemName string) (v int64, err error) {
	value, err := getValidItemValue(itemName, types.SystemConfigItemTypeInt64)
	if err == nil {
		v = value.(int64)
	}

	return
}

// ValidItemFloat64 获取 float64 配置项
func ValidItemFloat64(itemName string) (v float64, err error) {
	value, err := getValidItemValue(itemName, types.SystemConfigItemTypeFloat64)
	if err == nil {
		v = value.(float64)
	}

	return
}

// ValidItemBool 获取 bool 配置项
func ValidItemBool(itemName string) (v bool, err error) {
	value, err := getValidItemValue(itemName, types.SystemConfigItemTypeBool)
	if err == nil {
		v = value.(bool)
	}

	return
}

// ValidItemString 获取 string 配置项
func ValidItemString(itemName string) (v string) {
	value, err := getValidItemValue(itemName, types.SystemConfigItemTypeString)
	if err == nil {
		v = value.(string)
	}

	return
}

// 按指定类型取某一配置项的有效值
func getValidItemValue(itemName string, itemType types.SystemConfigItemType) (itemValue interface{}, err error) {
	origin, err := getValidItemValueFromCache(itemName)
	if err != nil {
		return
	}

	switch itemType {
	case types.SystemConfigItemTypeInt:
		s2i, err := tools.Str2Int(origin)
		if err != nil {
			logs.Error("[getValidItemValue] value:", origin, ", type:", itemType, "error:", err)
		}
		itemValue = s2i
	case types.SystemConfigItemTypeInt64:
		s2i64, err := tools.Str2Int64(origin)
		if err != nil {
			logs.Error("[getValidItemValue] value:", origin, ", type:", itemType, "error:", err)
		}
		itemValue = s2i64
	case types.SystemConfigItemTypeFloat64:
		s2f64, err := tools.Str2Float64(origin)
		if err != nil {
			logs.Error("[getValidItemValue] value:", origin, ", type:", itemType, "error:", err)
		}
		itemValue = s2f64
	case types.SystemConfigItemTypeBool:
		var bl = false
		if origin != "0" && strings.ToLower(origin) != "false" {
			bl = true
		}
		itemValue = bl
	default:
		itemValue = origin
	}

	return
}

func hashKey() string {
	return beego.AppConfig.String("system_config")
}

// 从cache取配置项
func getValidItemValueFromCache(itemName string) (itemValue string, err error) {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	hashName := hashKey()
	hValue, err := storageClient.Do("HGET", hashName, itemName)
	if err != nil {
		logs.Error("[getValidItemValueFromCache] no data, hashName: %s, itemName: %s, err: %#v", hashName, itemName, err)
		return
	} else if hValue == nil {
		// redis中没有,从db拿一份,再放到redis
		_, dbValue, errDB := getValidItemValueFromDB(itemName)
		if errDB != nil {
			err = errDB
			return
		}

		itemValue = dbValue

		storageClient.Do("HSET", hashName, itemName, itemValue)
	} else {
		itemValue = string(hValue.([]byte))
	}

	return
}

// 从DB取某一配置项的有效值
func getValidItemValueFromDB(itemName string) (id int64, itemValue string, err error) {
	sysConf := models.SystemConfig{}
	o := orm.NewOrm()
	o.Using(sysConf.Using())

	sql := fmt.Sprintf("SELECT id, item_value FROM %s WHERE item_name = '%s' AND status = %d ORDER BY id DESC LIMIT 1",
		sysConf.TableName(), itemName, types.StatusValid)
	err = o.Raw(sql).QueryRow(&sysConf)

	if err != nil {
		logs.Error("[getValidItemValueFromDB] has wrong, itemName: %s, err: %v", itemName, err)
		return
	}

	id = sysConf.Id
	itemValue = sysConf.ItemValue

	return
}

// 取某一配置项最新的version
func getItemVersion(item string) (version int, err error) {
	sysConf := models.SystemConfig{}
	o := orm.NewOrm()
	o.Using(sysConf.Using())

	sql := fmt.Sprintf("SELECT version FROM %s WHERE item_name = '%s' ORDER BY id DESC LIMIT 1",
		sysConf.TableName(), item)

	err = o.Raw(sql).QueryRow(&version)

	return
}
