package area

import (
	"errors"
	"testing"
)

var (
	//	3322: "KAB. SEMARANG",
	// 33 "JAWA TENGAH"
	//	332217: "Kaliwungu",

	testName2AreaCodeData1 = map[string]string{
		"province": "JAWA TENGAH",
		"city":     "KAB. SEMARANG",
		"area":     "Kaliwungu",
	}
)

func TestGetAreaJSONData(t *testing.T) {
	_, err := GetAreaJSONData()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestName2AreaCode(t *testing.T) {
	code, err := Name2AreaCode(testName2AreaCodeData1["province"], testName2AreaCodeData1["city"], testName2AreaCodeData1["area"])
	if err != nil {
		t.Error(err)
	} else if code != 332217 {
		t.Error(errors.New("wrong query"))
	}
}
