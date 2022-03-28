package main

import (
	"strings"
	"testing"
)

func TestReadiniForComments(t *testing.T) {
	testmap, err := readini("./testfiles/testini.cfg")
	if err != nil {
		t.Errorf("Expected map of settings, received:%v", err)
	}
	for k, v := range testmap {
		t.Log("pair:" + k + ":" + v)
		if strings.Contains(k, "#commented") {
			t.Errorf("Expected omit comments, received:%s:%s", k, v)
		}
	}
}

func TestDateToInt64(t *testing.T) {
	var d, rev Date
	var res int64
	d = Date{1999, 1, 1}
	res = dateToInt64(d)
	rev = int64ToDate(res)
	t.Log("Input and Result:", d, rev)
	if rev != d {
		t.Errorf("Expected:%+v, received:%+v, %d", d, rev, res%31)
	}
	d = Date{0, 12, 31}
	res = dateToInt64(d)
	rev = int64ToDate(res)
	t.Log("Input and Result:", d, rev)
	if rev != d {
		t.Errorf("Expected:%+v, received:%+v", d, rev)
	}
}

func TestDateTimeToInt64(t *testing.T) {
	var d, rev DateTime
	var res int64
	d = DateTime{3000, 12, 31, 0, 59}
	res = dateTimeToInt64(d)
	rev = int64ToDateTime(res)
	t.Log("Input and Result:", d, rev)
	if rev != d {
		t.Errorf("Expected:%+v, received:%+v, %d", d, rev, res%31)
	}
	d = DateTime{-5, 1, 1, 23, 0}
	res = dateTimeToInt64(d)
	rev = int64ToDateTime(res)
	t.Log("Input and Result:", d, rev)
	if rev != d {
		t.Errorf("Expected:%+v, received:%+v, %d", d, rev, res%31)
	}
}
