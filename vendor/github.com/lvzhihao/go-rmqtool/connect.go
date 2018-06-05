package rmqtool

import (
	"crypto/tls"
	"fmt"

	"github.com/streadway/amqp"
)

type ConnectConfig struct {
	Host     string                 `json:"host" yaml:"host"`         //127.0.0.1:5672
	Api      string                 `json:"api" yaml:"api"`           //http://127.0.0.1:15672
	User     string                 `json:"user" yaml:"user"`         //username
	Passwd   string                 `json:"passwd" yaml:"passwd"`     //passwd
	Vhost    string                 `json:"vhost" yaml:"vhost"`       //vhost
	MetaData map[string]interface{} `json:"metadata" yaml:"metadata"` //metadata
}

func (c *ConnectConfig) Scheme() string {
	return fmt.Sprintf("amqp://%s:%s@%s/%s", c.User, c.Passwd, c.Host, c.Vhost)
}

func NewConnect(config ConnectConfig) *Connect {
	return &Connect{
		config: config,
	}
}

type Connect struct {
	config ConnectConfig //queue config
}

//todo check link

func (c *Connect) Scheme() string {
	return c.config.Scheme()
}

func (c *Connect) Api() string {
	return c.config.Api
}

func (c *Connect) User() string {
	return c.config.User
}

func (c *Connect) Passwd() string {
	return c.config.Passwd
}

func (c *Connect) Vhost() string {
	return c.config.Vhost
}

func (c *Connect) Clone() *Connect {
	return &Connect{
		config: c.config,
	}
}

func (c *Connect) APIClient() *APIClient {
	return NewAPIClient(c.Api(), c.User(), c.Passwd())
}

func (c *Connect) Dial() (*amqp.Connection, error) {
	return amqp.Dial(c.Scheme())
}

func (c *Connect) DialConfig(config amqp.Config) (*amqp.Connection, error) {
	return amqp.DialConfig(c.Scheme(), config)
}

func (c *Connect) DialTLS(config *tls.Config) (*amqp.Connection, error) {
	return amqp.DialTLS(c.Scheme(), config)
}

func (c *Connect) QuickCreateExchange(name, kind string, durable bool) error {
	conn, err := c.Dial()
	if err != nil {
		return err
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return channel.ExchangeDeclare(name, kind, durable, false, false, false, nil)
}

func (c *Connect) QuickDeleteExchange(name string) error {
	conn, err := c.Dial()
	if err != nil {
		return err
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return channel.ExchangeDelete(name, true, false) //no force
}

func (c *Connect) QuickExchangeBind(destination, key, source string) error {
	conn, err := c.Dial()
	if err != nil {
		return err
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return channel.ExchangeBind(destination, key, source, false, nil)
}

func (c *Connect) QuickExchangeUnbind(destination, key, source string) error {
	conn, err := c.Dial()
	if err != nil {
		return err
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return channel.ExchangeUnbind(destination, key, source, false, nil)
}

func (c *Connect) QuickCreateQueue(name string, durable bool) error {
	conn, err := c.Dial()
	if err != nil {
		return err
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	_, err = channel.QueueDeclare(name, durable, false, false, false, nil)
	return err
}

func (c *Connect) QuickDeleteQueue(name string) error {
	conn, err := c.Dial()
	if err != nil {
		return err
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	_, err = channel.QueueDelete(name, true, true, false)
	return err
}

func (c *Connect) QuickPurgeQueue(name string) error {
	conn, err := c.Dial()
	if err != nil {
		return err
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	_, err = channel.QueuePurge(name, false)
	return err
}

func (c *Connect) QuickQueueBind(name, key, exchange string, args amqp.Table) error {
	conn, err := c.Dial()
	if err != nil {
		return err
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return channel.QueueBind(name, key, exchange, false, args)
}

func (c *Connect) QuickQueueUnbind(name, key, exchange string, args amqp.Table) error {
	conn, err := c.Dial()
	if err != nil {
		return err
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return channel.QueueUnbind(name, key, exchange, args)
}

func (c *Connect) ApplyQueue(name string) *Queue {
	return &Queue{
		conn: c,
		name: name,
	}
}

func (c *Connect) ApplyPublisher(exchange string, routeKeys []string) (*PublisherTool, error) {
	return NewPublisherTool(c.Scheme(), exchange, routeKeys)
}
