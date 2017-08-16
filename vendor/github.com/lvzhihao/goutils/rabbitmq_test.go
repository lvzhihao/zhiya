package goutils

import "testing"

var (
	TestRMQHttpApi *RMQHttpApi
)

func init() {
	TestRMQHttpApi, _ = NewRMQHttpApi("http://mq.sandbox.wdwd.com/", "uchat", "uchat")
}

func TestRMQHttpApi_000_Pre(t *testing.T) {
	status, json, err := TestRMQHttpApi.Scan(TestRMQHttpApi.GET("api/overview"))
	t.Logf("status: %d\njson: %s\nerr: %v\n", status, json, err)
}

func TestRMQHttpApi_001_WhoAmI(t *testing.T) {
	t.Log(TestRMQHttpApi.WhoAmI())
}

func TestRMQHttpApi_002_Permissions(t *testing.T) {
	t.Log(TestRMQHttpApi.Permissions())
}
