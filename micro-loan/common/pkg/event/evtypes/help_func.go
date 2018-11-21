package evtypes

import (
	"reflect"
)

// GetStructName 获取 struct name , 以生成 map
func GetStructName(m interface{}) string {
	val := reflect.ValueOf(m)
	name := reflect.Indirect(val).Type().Name()
	//fmt.Println(name)
	return name
}
