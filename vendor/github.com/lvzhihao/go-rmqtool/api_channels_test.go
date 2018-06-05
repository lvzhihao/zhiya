package rmqtool

import (
	"testing"

	"github.com/streadway/amqp"
)

func LinkTestChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	return conn.Channel()
}

func TestAPIListChannels(t *testing.T) {
	/*
		conn, err := LinkTestConnection()
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		channel, err := LinkTestChannel(conn)
		if err != nil {
			t.Fatal(err)
		}
		defer channel.Close()
		time.Sleep(5 * time.Second)
	*/
	ret, err := GenerateTestClient().ListChannels()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIListVhostChannels(t *testing.T) {
	ret, err := GenerateTestClient().ListVhostChannels("/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIChannel(t *testing.T) {
	//null
}

func TestAPIListConsumers(t *testing.T) {
	ret, err := GenerateTestClient().ListConsumers()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}
func TestAPIListVhostConsumers(t *testing.T) {
	ret, err := GenerateTestClient().ListVhostConsumers("/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}
