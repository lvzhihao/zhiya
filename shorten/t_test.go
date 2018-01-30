package shorten

import (
	"flag"
	"testing"
)

var (
	testTClient  *T
	testTSource  string
	testBaseLink string = "http://www.baidu.com"
)

func init() {
	flag.StringVar(&testTSource, "source", "", "appkey")
	flag.Parse()
	testTClient = NewT(&TConfig{
		Source: testTSource,
	})
}

func Test_001_Shorts(t *testing.T) {
	i := 0
	for {
		short := testTClient.Shorts([]string{testBaseLink})
		t.Log(short)
		i++
		if i > 10 {
			break
		}
	}
}
