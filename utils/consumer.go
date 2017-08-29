package utils

import (
	"time"

	"github.com/lvzhihao/goutils"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type receiveConsumer struct {
	amqpUrl string
	conn    *amqp.Connection
}

func NewReceiveConsumer(url string) (*receiveConsumer, error) {
	c := &receiveConsumer{
		amqpUrl: url,
	}
	// first test dial
	_, err := amqp.Dial(url)
	return c, err
}

func (c *receiveConsumer) link(queue string, prefetchCount int) (<-chan amqp.Delivery, error) {
	var err error
	c.conn, err = amqp.Dial(c.amqpUrl)
	if err != nil {
		Logger.Error("amqp.open", zap.Error(err))
		return nil, err
	}
	_, err = c.conn.Channel()
	if err != nil {
		Logger.Error("channel.open", zap.Error(err))
		return nil, err
	}
	channel, _ := c.conn.Channel()
	if err := channel.Qos(prefetchCount, 0, false); err != nil {
		Logger.Error("channel.qos", zap.Error(err))
		return nil, err
	}
	deliveries, err := channel.Consume(queue, "ctag-"+goutils.RandomString(20), false, false, false, false, nil)
	if err != nil {
		Logger.Error("base.consume", zap.Error(err))
		return deliveries, err
	}
	return deliveries, nil
}

func (c *receiveConsumer) Consumer(queue string, prefetchCount int, handle func(msg amqp.Delivery)) {
	for {
		time.Sleep(3 * time.Second)
		deliveries, err := c.link(queue, prefetchCount)
		if err != nil {
			Logger.Error("Consumer Link Error", zap.Error(err))
			continue
		}
		for msg := range deliveries {
			go handle(msg)
		}
		c.conn.Close()
		Logger.Info("Consumer ReConnection After 3 Second")
	}
}
