package chatbot

import (
	"flag"
	"testing"
)

var (
	testClient         *Client
	testHost           string
	testMerchantNo     string
	testMerchantSecret string
)

func init() {
	flag.StringVar(&testHost, "host", "", "host")
	flag.StringVar(&testMerchantNo, "no", "", "merchantNo")
	flag.StringVar(&testMerchantSecret, "secret", "", "merchantSecret")
	flag.Parse()
	testClient = NewClient(&ClientConfig{
		ApiHost:        testHost,
		MerchantNo:     testMerchantNo,
		MerchantSecret: testMerchantSecret,
	})
}

func Test_001_Api_ApplyRule(t *testing.T) {
	rst, err := testClient.ApplyRule("testId", "tuling")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rst)
}

func Test_001_Api_SendMessage(t *testing.T) {
	rst, err := testClient.SendMessage("testId", "tuling", "上海天气", "testUserId")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rst)
}
