package datetime

import "testing"

func TestDateToInt64(t *testing.T) {
	var d, rev Date
	var res int64
	d = Date{1999, 1, 1}
	res = DateToInt64(d)
	rev = Int64ToDate(res)
	t.Log("Input and Result:", d, rev)
	if rev != d {
		t.Errorf("Expected:%+v, received:%+v, %d", d, rev, res%31)
	}
	d = Date{0, 12, 31}
	res = DateToInt64(d)
	rev = Int64ToDate(res)
	t.Log("Input and Result:", d, rev)
	if rev != d {
		t.Errorf("Expected:%+v, received:%+v", d, rev)
	}
}
