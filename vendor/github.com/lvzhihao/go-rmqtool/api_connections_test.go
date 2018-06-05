package rmqtool

import (
	"testing"

	"github.com/streadway/amqp"
)

func LinkTestConnection() (*amqp.Connection, error) {
	return amqp.Dial("amqp://guest:guest@localhost:6672")
}

func TestAPIListConnections(t *testing.T) {
	ret, err := GenerateTestClient().ListConnections()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIListVhostConnections(t *testing.T) {
	ret, err := GenerateTestClient().ListVhostConnections("/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIConnections(t *testing.T) {
	//null
}

func TestForceDeleteConnection(t *testing.T) {
	//null
}

func TestAPIListConnectionChannels(t *testing.T) {
	//null
}
