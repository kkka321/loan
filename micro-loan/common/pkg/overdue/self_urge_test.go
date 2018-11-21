package overdue

import (
	"testing"
)

func TestIsEdgeOrBeyond(t *testing.T) {
	td := []struct {
		in  int
		out bool
	}{
		{1, false},
		{4, false},
		{10, false},
		{20, true},
		{30, true},
		{13, true},
	}
	for _, d := range td {
		if out := IsEdgeOrBeyond(d.in); out != d.out {
			t.Log(d, "out:", out)
			t.Fail()
		}
	}
}
