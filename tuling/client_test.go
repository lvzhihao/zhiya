package tuling

import (
	"testing"
	"time"
)

func Test(t *testing.T) {
	c := NewTulingClient(TulingClientConfig{
		ApiUrl: "http://www.tuling123.com/openapi/api",
		ApiKey: "70aba412ec2841718038ad3988566487",
		//ApiSecret:      "8e74f8e7bca79eb1",
		DefaultTimeout: 10 * time.Second,
	})
	t.Log(
		c.Do(map[string]interface{}{
			"info":   "北京到上海的飞机",
			"userid": "testUtil",
		}),
	)
	t.Log(
		c.Do(map[string]interface{}{
			"info":   "北京到东京的飞机",
			"userid": "testUtil-1",
		}),
	)
	t.Log(
		c.Do(map[string]interface{}{
			"info":   "明天去",
			"userid": "testUtil",
		}),
	)
}
