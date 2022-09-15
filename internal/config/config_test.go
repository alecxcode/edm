package config

import (
	"strings"
	"testing"
)

func TestReadiniForComments(t *testing.T) {
	testmap, err := readini("testini.cfg")
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
