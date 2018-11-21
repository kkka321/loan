package areacode

import (
	"testing"
)

func TestPhoneWithServiceRegionCode(t *testing.T) {
	td := []struct {
		in string
		sr string
		ex string
	}{
		{
			"088988090809",
			"IDN",
			"6288988090809",
		},
		{
			"8988090809",
			"IDN",
			"628988090809",
		},
		{
			"621111111111",
			"IDN",
			"621111111111",
		},
		{
			"18518027928",
			"CHN",
			"8618518027928",
		},
		{
			"0000000000",
			"UExpectedRegion",
			"0000000000",
		},
	}
	for _, d := range td {
		defaultServiceRegion = d.sr
		if d.ex != PhoneWithServiceRegionCode(d.in) {
			t.Error(d, "unexpected result:", normalPareseAndWrapCountryCode(d.in))
			t.Fail()
		}
	}
}
