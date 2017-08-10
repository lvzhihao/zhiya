package utils

import (
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var Logger *zap.Logger

type ReceiveTool struct {
	channels map[string]*receiveChannel
}

func NewTool(url, exchange string, routeKeys []string) (*ReceiveTool, error) {
	tool := &ReceiveTool{
		channels: make(map[string]*receiveChannel, 0),
	}
	err := tool.conn(url, exchange, routeKeys)
	return tool, err
}

func (c *ReceiveTool) conn(url, exchange string, routeKeys []string) error {
	_, err := amqp.Dial(url)
	if err != nil {
		return err
	} //test link
	for _, route := range routeKeys {
		c.channels[route] = &receiveChannel{
			amqpUrl:  url,
			exchange: exchange,
			routeKey: route,
			Channel:  make(chan string, 2000),
		}
		go c.channels[route].Receive()
	}
	return nil
}

func (c *ReceiveTool) Publish(route, msg string) {
	c.channels[route].Channel <- msg
}

type receiveChannel struct {
	amqpUrl  string
	exchange string
	routeKey string
	Channel  chan string
}

func (c *receiveChannel) Receive() {
RetryConnect:
	conn, err := amqp.Dial(c.amqpUrl)
	if err != nil {
		Logger.Error("Channel Connection Error 1", zap.String("route", c.routeKey), zap.Error(err))
		time.Sleep(3 * time.Second)
		goto RetryConnect
	}
	channel, err := conn.Channel()
	if err != nil {
		Logger.Error("Channel Connection Error 2", zap.String("route", c.routeKey), zap.Error(err))
		conn.Close()
		time.Sleep(3 * time.Second)
		goto RetryConnect
	}
	err = channel.ExchangeDeclare(c.exchange, "topic", true, false, false, false, nil)
	if err != nil {
		Logger.Error("Channel Connection Error 3", zap.String("route", c.routeKey), zap.Error(err))
		conn.Close()
		time.Sleep(3 * time.Second)
		goto RetryConnect
	}
BreakFor:
	for {
		select {
		case str := <-c.Channel:
			if str == "quit" {
				Logger.Info("Channel Connection Quit", zap.String("route", c.routeKey))
				conn.Close()
				return
			} //quit
			msg := amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				Timestamp:    time.Now(),
				ContentType:  "application/json",
				Body:         []byte(str),
			}
			err := channel.Publish(c.exchange, c.routeKey, false, false, msg)
			if err != nil {
				c.Channel <- str
				conn.Close()
				Logger.Error("Channel Connection Error 4", zap.String("route", c.routeKey), zap.Error(err))
				break BreakFor
			}
		}
	}
	goto RetryConnect
}
