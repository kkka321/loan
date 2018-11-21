package feedback

import "testing"

func TestGetTagDisplay(t *testing.T) {
	td := []struct {
		in  int
		out string
	}{
		{1, "Bug"},
		{4, "Complaints"},
		{2, "Suggest"},
		{2048, "Other"},
		{5, "Bug,Complaints"},
		{2048, "Other"},
	}
	for _, d := range td {
		if out := GetTagDisplay("english", d.in); out != d.out {
			t.Log(d, "out:", out)
			t.Fail()
		}
	}
}
