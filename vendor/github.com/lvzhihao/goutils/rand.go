package goutils

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func RandStr(strlen int) string {
	return RandomString(strlen)
}

func RandomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	rand.Seed(time.Now().UTC().UnixNano())

	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}

/*
 * runtime version >= version; return true
 * runtime version <  version; return false
 */
func CompareRuntimeVersion(version string) bool {
	switch CompareVersion(strings.Replace(runtime.Version(), "go", "", -1), strings.Replace(version, "go", "", -1), 2) {
	case "<":
		return false
	default:
		return true
	}
}

func CompareVersion(v1, v2 string, width int) string {
	v1s := parseVersion(v1, width)
	v2s := parseVersion(v2, width)
	if v1s > v2s {
		return ">"
	} else if v1s < v2s {
		return "<"
	} else {
		return "="
	}
}

func parseVersion(version string, width int) int64 {
	vList := strings.Split(version, ".")
	format := fmt.Sprintf("%%s%%0%ds", width)
	v := ""
	for _, value := range vList {
		v = fmt.Sprintf(format, v, value)
	}
	rst, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	} else {
		return rst
	}
}

func InStringSlice(slice []string, s string) bool {
	for _, v := range slice {
		if strings.Compare(v, s) == 0 {
			return true
		}
	}
	return false
}
