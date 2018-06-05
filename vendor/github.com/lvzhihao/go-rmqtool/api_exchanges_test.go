package rmqtool

import "testing"

func TestAPIListExchange(t *testing.T) {
	ret, err := GenerateTestClient().ListExchanges()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIListVhostExchange(t *testing.T) {
	ret, err := GenerateTestClient().ListVhostExchanges("/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIExhcnage(t *testing.T) {
	testName := "test.test"
	err := GenerateTestClient().CreateExchange("/", testName, map[string]interface{}{
		"type": "direct",
	})
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Create Test Exchange Success: ", testName)
	}
	exchange, err := GenerateTestClient().Exchange("/", testName)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(exchange)
	}
	err = GenerateTestClient().DeleteExchange("/", exchange["name"].(string))
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Delete Test Exchange Success")
	}
}

func TestAPIListExchangeSourceBindings(t *testing.T) {
	ret, err := GenerateTestClient().ListExchangeSourceBindings("/", "amq.topic")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIListExchangeDestinationBindings(t *testing.T) {
	ret, err := GenerateTestClient().ListExchangeDestinationBindings("/", "amq.topic")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIExchangePublish(t *testing.T) {
	ret, err := GenerateTestClient().ExchangePublish("/", "amq.topic", map[string]interface{}{
		"properties":       struct{}{},
		"routing_key":      "my key",
		"payload":          "my body",
		"payload_encoding": "string",
	})
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Publish Success: ", ret)
	}
}

func TestAPIExchangeBindings(t *testing.T) {
	name, err := GenerateTestClient().ExchangeCreateBinding("/", "amq.topic", "amq.direct", "aaab", map[string]interface{}{
		"abcdefg": "xx",
	})
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Binding Key Success: ", name)
	}
	ret, err := GenerateTestClient().ExchangeBindings("/", "amq.topic", "amq.direct")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Bindings List: ", ret)
		obj, err := GenerateTestClient().ExchangeBinding("/", "amq.topic", "amq.direct", ret[0]["properties_key"].(string))
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log(obj)
			err = GenerateTestClient().ExchangeDeleteBinding("/", "amq.topic", "amq.direct", obj["properties_key"].(string))
			if err != nil {
				t.Fatal(err)
			} else {
				t.Log("Delete Binding Success: ", obj)
			}
		}
	}
}
