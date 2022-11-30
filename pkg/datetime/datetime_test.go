package datetime

import "testing"

func TestDateTimeToInt64(t *testing.T) {
	var d, rev DateTime
	var res int64
	d = DateTime{3000, 12, 31, 0, 59}
	res = DateTimeToInt64(d)
	rev = Int64ToDateTime(res)
	t.Log("Input and Result:", d, rev)
	if rev != d {
		t.Errorf("Expected:%+v, received:%+v, %d", d, rev, res%31)
	}
	d = DateTime{-5, 1, 1, 23, 0}
	res = DateTimeToInt64(d)
	rev = Int64ToDateTime(res)
	t.Log("Input and Result:", d, rev)
	if rev != d {
		t.Errorf("Expected:%+v, received:%+v, %d", d, rev, res%31)
	}
}
