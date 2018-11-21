package tools

import (
	"errors"
	"testing"
)

func TestPareseDateRangeToMillsecond(t *testing.T) {
	tds := []struct {
		in    string
		start string
		end   string
		err   error
	}{
		{"2018-06-90 - 2018-09-11", "2018-06-05", "2018-09-11", errors.New("any err,start format error")},
		{"2018-03-02 - 2018-07-11", "2018-03-02", "2018-07-11", nil},
		{"2018-06-05- 2018-09-11", "2018-06-05", "2018-09-11", errors.New("any err,seprate err")},
		{"2018-06-05  2018-09-11", "2018-06-07", "2018-09-11", errors.New("any err,start and end not match")},
	}
	for _, d := range tds {
		start, end, err := PareseDateRangeToMillsecond(d.in)
		expectedStart := GetDateParseBackend(d.start) * 1000
		expectedEnd := GetDateParseBackend(d.end)*1000 + 3600*24*1000

		if err != nil || d.err != nil {
			if (err != d.err) && (d.err == nil || err == nil) {
				t.Error("Expected err and acutal err is same = nil, or any error match ")
			}
			continue
		}

		if start != expectedStart || end != expectedEnd {
			t.Error("Not expected", "in:", d, "expectedStart:", expectedStart,
				"actual start:", start, "expectedEnd:", expectedEnd, "actual end:", end)

		}

	}
}
