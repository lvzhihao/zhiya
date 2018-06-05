package rmqtool

import (
	"github.com/streadway/amqp"
)

type QueueConfig struct {
	Name     string       `yaml:"name"`
	Bindlist []*QueueBind `yaml:"bindlist"`
}

type QueueBind struct {
	Key       string     `yaml:"key"`
	Exchange  string     `yaml:"exchange"`
	Arguments amqp.Table `yaml:"arguments"`
}

type Queue struct {
	conn     *Connect
	name     string
	consumer *ConsumerTool
}

func (c *Queue) Clone(name string) *Queue {
	return &Queue{
		conn: c.conn,
		name: name,
	}
}

func (c *Queue) Scheme() string {
	return c.conn.Scheme()
}

func (c *Queue) Api() string {
	return c.conn.Api()
}

func (c *Queue) User() string {
	return c.conn.User()
}

func (c *Queue) Passwd() string {
	return c.conn.Passwd()
}

func (c *Queue) Vhost() string {
	return c.conn.Vhost()
}

func (c *Queue) Name() string {
	return c.name
}

func (c *Queue) ApplyConsumer() (*ConsumerTool, error) {
	return NewConsumerTool(c.Scheme(), c.Name())
}

func (c *Queue) Consume(prefetchCount int, handle func(amqp.Delivery)) error {
	var err error
	c.consumer, err = c.ApplyConsumer()
	if err != nil {
		return err
	}
	c.consumer.Consume(prefetchCount, handle)
	defer c.consumer.Close()
	return nil
}

func (c *Queue) Ensure(bindList []*QueueBind) error {
	err := c.Create()
	if err != nil {
		return err
	} else {
		return c.Binds(bindList)
	}
}

func (c *Queue) Create() error {
	return c.conn.QuickCreateQueue(c.Name(), true)
}

func (c *Queue) Purge() error {
	return c.conn.QuickPurgeQueue(c.Name())
}

func (c *Queue) Delete() error {
	return c.conn.QuickDeleteQueue(c.Name())
}

func (c *Queue) Close() {
	c.consumer.Close()
}

func (c *Queue) Binds(bindList []*QueueBind) error {
	for _, bind := range bindList {
		err := c.Bind(bind)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Queue) Bind(bind *QueueBind) error {
	return c.conn.QuickQueueBind(c.Name(), bind.Key, bind.Exchange, bind.Arguments)
}

func (c *Queue) Unbind(bind *QueueBind) error {
	return c.conn.QuickQueueUnbind(c.Name(), bind.Key, bind.Exchange, bind.Arguments)
}
