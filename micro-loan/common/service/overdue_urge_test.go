package service

import (
	"errors"
	"micro-loan/common/tools"
	"testing"

	"github.com/astaxie/beego/logs"
)

func TestCalculateOverdueLevel(t *testing.T) {
	todayMillTimeStamp := tools.NaturalDay(0)
	var oneDayMill int64 = 24 * 3600 * 1000
	anyError := errors.New("any error")
	dt := []struct {
		in   int64
		out1 string
		out2 int64
		err  error
	}{
		{
			todayMillTimeStamp + 1*oneDayMill, "", 0, anyError,
		},
		{
			todayMillTimeStamp - 1*oneDayMill, "", 1, anyError,
		},
		{
			todayMillTimeStamp - 2*oneDayMill, "M1-1", 2, nil,
		},
		{
			todayMillTimeStamp - 5*oneDayMill, "M1-1", 5, nil,
		},
		{
			todayMillTimeStamp - 7*oneDayMill, "M1-1", 7, nil,
		},
		{
			todayMillTimeStamp - 8*oneDayMill, "M1-1", 8, nil,
		},
		{
			todayMillTimeStamp - 9*oneDayMill, "M1-2", 9, nil,
		},
		{
			todayMillTimeStamp - 10*oneDayMill, "M1-2", 10, nil,
		},
		{
			todayMillTimeStamp - 16*oneDayMill, "M1-2", 16, nil,
		},
		{
			todayMillTimeStamp - 17*oneDayMill, "M1-2", 17, nil,
		},
		{
			todayMillTimeStamp - 18*oneDayMill, "M1-2", 18, nil,
		},
		{
			todayMillTimeStamp - 19*oneDayMill, "M1-2", 19, nil,
		},
		{
			todayMillTimeStamp - 100*oneDayMill, "M3", 100, nil,
		},
	}

	for _, d := range dt {
		o1, o2, err := CalculateOverdueLevel(d.in)
		if o1 != d.out1 || o2 != d.out2 || (err == nil && d.err != nil) || (err != nil && d.err == nil) {
			logs.Error("TestCalculateOverdueLevel", "want:", d, "acutal out:", o1, o2, err)
			t.Fail()
		}
	}
}
