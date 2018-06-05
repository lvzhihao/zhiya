package rmqtool

import "testing"

func TestAPIClientOverview(t *testing.T) {
	ret, err := GenerateTestClient().Overview()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIListNodes(t *testing.T) {
	ret, err := GenerateTestClient().ListNodes()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPINode(t *testing.T) {
	ret, err := GenerateTestClient().Node("rabbit@test-rabbit", map[string]string{"memory": "true", "binary": "true"})
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIListExtensions(t *testing.T) {
	ret, err := GenerateTestClient().ListExtensions()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIDefinitions(t *testing.T) {
	ret, err := GenerateTestClient().Definitions()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIVhostDefinitions(t *testing.T) {
	ret, err := GenerateTestClient().VhostDefinitions("/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}
