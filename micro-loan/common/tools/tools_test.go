package tools

import (
	"errors"
	"testing"
)

func TestIntsSliceToWhereInString(t *testing.T) {
	anyError := errors.New("any error")
	td := []struct {
		in  interface{}
		out string
		err error
	}{
		{
			[]int64{1, 2, 3, 4},
			"1,2,3,4",
			nil,
		},
		{
			[]interface{}{int64(1), int64(2), int64(3), int64(4)},
			"1,2,3,4",
			nil,
		},
		{
			nil,
			"",
			anyError,
		},
		{
			[]interface{}{},
			"",
			anyError,
		},
	}
	for _, d := range td {
		actualOut, err := IntsSliceToWhereInString(d.in)
		if d.out != actualOut || (err != nil && d.err == nil) || (err == nil && d.err != nil) {
			t.Errorf("want out:%s, actual out: %s, input:%v,actual err: %v", d.out, actualOut, d.in, err)
		}
	}

}
