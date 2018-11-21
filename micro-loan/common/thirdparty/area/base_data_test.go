package area

import (
	"strconv"
	"testing"
)

//  10< province < 100
//  10< city < 100
func TestProvinceCodeCheck(t *testing.T) {
	for k := range provinceCodeMap {
		if k > MaxProvinceCode {
			t.Error("Province Code:", strconv.Itoa(k), "Bigger then max province code")
		}
		if k < MinProvinceCode {
			t.Error("Province Code:", strconv.Itoa(k), "Little then max province code")
		}
	}
}

func TestCityCodeCheck(t *testing.T) {
	for k := range cityCodeMap {
		if k > MaxCityCode {
			t.Error("City Code:", strconv.Itoa(k), "Bigger then max City code")
		}
		if k < MinCityCode {
			t.Error("City Code:", strconv.Itoa(k), "Little then max City code")
		}
	}
}

func TestAreaCodeCheck(t *testing.T) {
	for k := range areaCodeMap {
		if k > MaxAreaCode {
			t.Error("Area Code:", strconv.Itoa(k), "Bigger then max Area code")
		}
		if k < MinAreaCode {
			t.Error("Area Code:", strconv.Itoa(k), "Little then max Area code")
		}
	}
}

// wait todo check 省市区从属关系
