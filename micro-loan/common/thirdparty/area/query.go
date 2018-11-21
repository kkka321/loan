package area

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sort"
)

var (
	errUndefinedArea     = errors.New("Undefined Area")
	errUndefinedCity     = errors.New("Undefined City")
	errUndefinedProvince = errors.New("Undefined Province")

	errAreaDismatchAnyCity     = errors.New("Area dismatched any city")
	errCityDismatchAnyProvince = errors.New("City dismatched any Province")
)

// Name2AreaCode 根据省市区名称查询code
func Name2AreaCode(province, city, area string) (code int, err error) {
	provinceCode, err := provinceName2Code(province)
	if nil != err {
		return
	}
	cityCode, err := cityName2CodeWithProvinceCode(provinceCode, city)
	if nil != err {
		return
	}
	code, err = areaName2CodeWithCityCode(cityCode, area)
	return
}

// provinceName2Code 根据省名查询 code
func provinceName2Code(provinceName string) (code int, err error) {
	for k, v := range provinceCodeMap {
		if v == provinceName {
			code = k
			return
		}
	}
	err = errUndefinedProvince
	return
}

// cityName2CodeWithProvinceCode 利用 provinceCode 与 cityCode 的关系，将循环数缩减到最大 100次
func cityName2CodeWithProvinceCode(provinceCode int, cityName string) (code int, err error) {
	maxCode := (provinceCode+1)*100 - 1
	for i := provinceCode * 100; i <= maxCode; i++ {
		if name, ok := cityCodeMap[i]; ok && cityName == name {
			return i, nil
		}
	}
	err = errUndefinedCity
	return
}

// areaName2CodeWithCityCode 利用 areaCode 与 cityCode 的关系，将循环数缩减到最大 100次
func areaName2CodeWithCityCode(cityCode int, areaName string) (code int, err error) {
	maxCode := (cityCode+1)*100 - 1
	for i := cityCode * 100; i <= maxCode; i++ {
		if name, ok := areaCodeMap[i]; ok && areaName == name {
			return i, nil
		}
	}
	err = errUndefinedArea
	return
}

// City 描述 City 结构，主要用于JSON数据中的数据生成
type City struct {
	ID   int      `json:"-"`
	Name string   `json:"name"`
	Area []string `json:"area"`
}

// Province 描述 Province 结构，主要用于JSON数据中的数据生成
type Province struct {
	ID   int    `json:"-"`
	Name string `json:"name"`
	City []City `json:"city"`
}

// generate a treeMap for client
func generateAreaMapTree() (provinceMap map[int]Province, err error) {
	// 生成map[int]City 备用
	cityMap := make(map[int]City)
	provinceMap = make(map[int]Province)

	for k, v := range cityCodeMap {
		if _, ok := provinceCodeMap[k/100]; ok {
			cityMap[k] = City{ID: k, Name: v}

			//areaMapTree[provinceCode]["city"] = v
		} else {
			err = errCityDismatchAnyProvince
			return
		}
	}

	// 按照id顺序遍历 area， 并写入 cityMap中对应的City.Area
	var keys []int
	for k := range areaCodeMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		// 校验 area 所属city是否存在
		if _, ok := cityMap[k/100]; ok {
			c := cityMap[k/100]
			c.Area = append(c.Area, areaCodeMap[k])
			cityMap[k/100] = c
		} else {
			err = errAreaDismatchAnyCity
			return
		}
	}

	// 生成 ProvinceMap 备用
	for k, v := range provinceCodeMap {
		provinceMap[k] = Province{ID: k, Name: v}
	}

	// 顺序遍历 cityMap, 将 cityMap 写入指定 Province.City 中
	var cKeys []int
	for k := range cityMap {
		cKeys = append(cKeys, k)
	}
	sort.Ints(cKeys)
	for _, k := range cKeys {
		if _, ok := provinceMap[k/100]; ok {
			p := provinceMap[k/100]
			p.City = append(p.City, cityMap[k])
			provinceMap[k/100] = p
		} else {
			err = errCityDismatchAnyProvince
			return
		}
	}

	return
}

// GetAreaTreeSlice 返回一个三级
func GetAreaTreeSlice() (provinceSlice []Province, err error) {
	provinceMap, err := generateAreaMapTree()
	if err != nil {
		return
	}
	// 将 provinceMap 按key即province id转成 slice
	var pKeys []int
	for k := range provinceMap {
		pKeys = append(pKeys, k)
	}
	sort.Ints(pKeys)
	for _, k := range pKeys {
		provinceSlice = append(provinceSlice, provinceMap[k])
	}
	return
}

// GetAreaJSONData 返回 AreaJsonData 给前端
func GetAreaJSONData() ([]byte, error) {
	provinceSlice, err := GetAreaTreeSlice()
	if err != nil {
		return nil, err
	}
	provinceJSON, err := json.Marshal(provinceSlice)
	if err != nil {
		return nil, err
	}
	return provinceJSON, nil
}

func createJSONFile() error {
	provinceJSON, err := GetAreaJSONData()
	if err != nil {
		return err
	}
	ioutil.WriteFile("./area.json", provinceJSON, os.ModeAppend)
	//fmt.Println(string(provinceJson))
	return nil
}
