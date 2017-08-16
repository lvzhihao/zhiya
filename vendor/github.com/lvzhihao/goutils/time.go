package goutils

import "time"

func CurrentTimeMills() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
