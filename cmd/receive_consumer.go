// Copyright © 2017 edwin <edwin.lzh@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"

	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/uchat/uchat"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

// receive_consumerCmd represents the receive_consumer command
var receive_consumerCmd = &cobra.Command{
	Use:   "receive_consumer",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		defer Logger.Sync()
		sugar := Logger.Sugar()
		queue, err := cmd.Flags().GetString("queue")
		if err != nil {
			sugar.Fatal(err)
		}
		shell := &consumerShell{}
		err = shell.Init()
		if err != nil {
			sugar.Fatal(err)
		}
		defer shell.Close()
		consumer, err := NewReceiveConsumer(fmt.Sprintf("amqp://%s:%s@%s/%s", viper.GetString("rabbitmq_user"), viper.GetString("rabbitmq_passwd"), viper.GetString("rabbitmq_host"), viper.GetString("rabbitmq_vhost")))
		if err != nil {
			sugar.Fatal(err)
		}
		switch queue {
		case "uchat.member.list":
			consumer.Consumer("uchat.member.list", 20, shell.MemberList)
		case "uchat.robot.chat.list":
			consumer.Consumer("uchat.robot.chat.list", 20, shell.ChatRoomList)
		default:
			sugar.Fatal("Please input current queue name")
		}
	},
}

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
		log.Printf("amqp.open: %s", err)
		return nil, err
	}
	_, err = c.conn.Channel()
	if err != nil {
		log.Printf("channel.open: %s", err)
		return nil, err
	}
	channel, _ := c.conn.Channel()
	if err := channel.Qos(prefetchCount, 0, false); err != nil {
		log.Printf("channel.qos: %s", err)
		return nil, err
	}
	deliveries, err := channel.Consume(queue, "ctag-"+goutils.RandomString(20), false, false, false, false, nil)
	if err != nil {
		log.Printf("base.consume: %v", err)
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

type consumerShell struct {
	db *gorm.DB
}

func (c *consumerShell) Init() (err error) {
	c.db, err = gorm.Open("mysql", viper.GetString("mysql_dns"))
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return viper.GetString("table_prefix") + "_" + defaultTableName
	}
	return
}

func (c *consumerShell) Close() {
	c.db.Close()
}

// 群成员列表
func (c *consumerShell) MemberList(msg amqp.Delivery) {
	err := uchat.SyncChatRoomMembersCallback(msg.Body, c.db)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.member.list"), zap.Error(err), zap.Any("msg", msg))
	}
	msg.Ack(false)
	//msg.Nack(false, true)
}

func (c *consumerShell) ChatRoomList(msg amqp.Delivery) {
	err := uchat.SyncRobotChatRoomsCallback(msg.Body, c.db)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.member.list"), zap.Error(err), zap.Any("msg", msg))
	}
	msg.Ack(false)
	//msg.Nack(false, true)
}

func init() {
	RootCmd.AddCommand(receive_consumerCmd)

	receive_consumerCmd.Flags().String("queue", "", "队列名称")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// receive_consumerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// receive_consumerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
