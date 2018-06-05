package rmqtool

import (
	"strings"
	"time"

	"github.com/streadway/amqp"
)

var (
	DefaultConsumerRetryTime time.Duration = 3 * time.Second
	DefaultConsumerToolName  string        = "golang.rmqtool"
)

func GenerateConsumerName(name string) string {
	return strings.Join([]string{
		name,
		time.Now().Format(time.RFC3339),
	}, ".")
}

type ConsumerTool struct {
	amqpUrl   string
	queue     string
	conn      *amqp.Connection
	name      string
	RetryTime time.Duration
	isClosed  bool
	//todo tmux.sync
}

func NewConsumerTool(url, queue string) (*ConsumerTool, error) {
	c := &ConsumerTool{
		amqpUrl:   url,                      //rmq link
		RetryTime: DefaultConsumerRetryTime, //default retry
		isClosed:  false,
		name:      DefaultConsumerToolName,
		queue:     queue,
	}
	// first test dial
	testConn, err := amqp.Dial(url)
	if testConn != nil {
		go testConn.Close()
	} // close test conn
	return c, err
}

func (c *ConsumerTool) Name() string {
	return c.name
}

func (c *ConsumerTool) SetName(name string) string {
	c.name = name
	return c.Name()
}

func (c *ConsumerTool) Link(prefetchCount int) (<-chan amqp.Delivery, error) {
	var err error
	c.conn, err = amqp.Dial(c.amqpUrl)
	if err != nil {
		if c.conn != nil {
			c.conn.Close()
		}
		return nil, err
	}
	channel, err := c.conn.Channel()
	if err != nil {
		c.conn.Close()
		return nil, err
	}
	if err := channel.Qos(prefetchCount, 0, false); err != nil {
		c.conn.Close()
		return nil, err
	}
	deliveries, err := channel.Consume(c.queue, GenerateConsumerName(c.name), false, false, false, false, nil)
	if err != nil {
		c.conn.Close()
		return nil, err
	}
	return deliveries, nil
}

func (c *ConsumerTool) Close() {
	// close
	if c.isClosed == false {
		if c.conn != nil {
			c.conn.Close()
		}
		c.isClosed = true
	}
}

func (c *ConsumerTool) Consume(prefetchCount int, handle func(amqp.Delivery)) {
	defer c.Close()
	for {
		if c.isClosed == true {
			Log.Error("Consumer Link Closed, Quit...", c.queue)
			break
		}
		time.Sleep(c.RetryTime)
		deliveries, err := c.Link(prefetchCount)
		if err != nil {
			Log.Error("Consumer Link Error", err)
			continue
		}
		for msg := range deliveries {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						Log.Error("Consumer Recover", r)
					}
				}()
				handle(msg)
			}()
		}
		c.conn.Close()
		Log.Debug("Consumer ReConnection After RetryTime", c.RetryTime)
	}
}
