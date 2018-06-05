package rmqtool

import "testing"

var TestClient *APIClient

func GenerateTestClient() *APIClient {
	if TestClient == nil {
		TestClient = NewAPIClient("http://localhost:16672/api", "guest", "guest")
	}
	return TestClient
}

func TestAPIClientCheck(t *testing.T) {
	err := GenerateTestClient().Check()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Http Self Check Success")
	}
}
