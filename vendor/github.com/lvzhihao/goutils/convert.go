package goutils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

func ToFloat(v interface{}) float64 {
	f, err := strconv.ParseFloat(ToString(v), 64)
	if err != nil {
		return 0
	} else {
		return f
	}
}

func ToString(v interface{}) (s string) {
	switch v.(type) {
	case nil:
		return ""
	case string:
		s = v.(string)
	case []byte:
		s = string(v.([]byte))
	case io.Reader:
		b, _ := ioutil.ReadAll(v.(io.Reader))
		s = string(b)
	case error:
		s = v.(error).Error()
	case NullString:
		if v.(NullString).Valid {
			s = v.(NullString).String
		} else {
			s = ""
		}
	case sql.NullString:
		if v.(sql.NullString).Valid {
			s = v.(sql.NullString).String
		} else {
			s = ""
		}
	default:
		b, err := json.Marshal(v)
		if err == nil {
			s = string(b)
		} else {
			s = fmt.Sprintf("%s", b)
		}
	}
	return
}

func ToBool(v interface{}) bool {
	switch v.(type) {
	case bool:
		return v.(bool)
	default:
		if strings.Compare(strings.ToLower(ToString(v)), "true") == 0 {
			return true
		} else {
			return false
		}
	}
}

func ToInt32(v interface{}) int32 {
	return int32(ToInt(v))
}

func ToInt64(v interface{}) int64 {
	return ToInt(v)
}

func ToInt(v interface{}) int64 {
	i, err := strconv.ParseInt(ToString(v), 10, 64)
	if err != nil {
		return 0
	} else {
		return i
	}
}

func ToPrice(f float64) int64 {
	s := strings.Replace(fmt.Sprintf("%.02f", f), ".", "", 1)
	rst, _ := strconv.ParseInt(s, 10, 64)
	return rst
}
