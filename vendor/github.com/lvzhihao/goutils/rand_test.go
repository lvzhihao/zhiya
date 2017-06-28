package goutils

import "testing"

func TestCompareRuntimeVersion(t *testing.T) {
	rst := CompareRuntimeVersion("1.8")
	t.Log(rst)
	rst = CompareRuntimeVersion("1.9")
	t.Log(rst)
	//todo
}
