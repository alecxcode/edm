package accs

import (
	"strings"
	"testing"
)

func someFunctionForTesting() string {
	name := CurrentFunction()
	return name
}

func TestCurrentFunction(t *testing.T) {
	name := someFunctionForTesting()
	if !strings.HasSuffix(name, "someFunctionForTesting") {
		t.Errorf("Expected:%+v, received:%+v", "someFunctionForTesting", name)
	}
}
