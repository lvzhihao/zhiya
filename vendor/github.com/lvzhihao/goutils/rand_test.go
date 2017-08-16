package goutils

import (
	"strings"
	"testing"
)

func TestCompareRuntimeVersion(t *testing.T) {
	r1 := CompareVersion("1.8.1.3", "1.8.1.10", 2)
	if strings.Compare(r1, "<") != 0 {
		t.Error("1.8.1.3 < 1.8.1.10 Then Result:", r1)
	}
	r2 := CompareVersion("1.8.1", "1.8.1", 2)
	if strings.Compare(r2, "=") != 0 {
		t.Error("1.8.1 = 1.8.1 Then Result:", r2)
	}
	r3 := CompareVersion("1.8.101", "1.8.102", 3)
	if strings.Compare(r3, "<") != 0 {
		t.Error("1.8.101 < 1.8.102 Then Result:", r3)
	}
}

func TestInStringSlice(t *testing.T) {
	slist := []string{
		"111",
		"222",
		"333",
		"444",
	}
	if !InStringSlice(slist, "111") {
		t.Error("111 in list Then Result: false")
	}
	if InStringSlice(slist, "1111") {
		t.Error("1111 not in list Then Result: true")
	}
}
