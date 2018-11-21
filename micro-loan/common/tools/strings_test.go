package tools

import (
	"fmt"
	"testing"
)

func TestTrimRealName(t *testing.T) {
	origin := "  MELI AMALIA MUTAQIN,S.SOS T  -T         "
	after := "MELI AMALIA MUTAQIN S SOS T T"
	fmt.Printf("trim result: %s\n", TrimRealName(origin))
	if TrimRealName(origin) != after {
		t.Errorf(`TrimRealName 没达到预期, origin: %s, after: %s`, origin, after)
	}
}

func TestStrReplace(t *testing.T) {
	origin := "[W] [funcName]"
	after := "W funcName"
	find := []string{"[", "]"}
	fmt.Printf("test 4 StrReplace\n")
	if StrReplace(origin, find, "") != after {
		t.Errorf("StrReplace test fail, origin: `%s`, after: `%s`", origin, after)
	}
}

func TestIsValidIndonesiaMobile(t *testing.T) {
	mobile := "082228687124"
	ok, err := IsValidIndonesiaMobile(mobile)
	if !ok {
		t.Errorf("印尼电话号码不合法, mobile: %s, err: %v", mobile, err)
	} else {
		fmt.Printf("印尼电话号码测试通过\n")
	}
}
