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
	"github.com/lvzhihao/zhiya/models"
	"github.com/lvzhihao/zhiya/uchat"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"github.com/vmihailenco/msgpack"
)

// broadcastCmd represents the broadcast command
var broadcastCmd = &cobra.Command{
	Use:   "broadcast",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		defer Logger.Sync()
		db, err := gorm.Open("mysql", viper.GetString("mysql_dns"))
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
			return viper.GetString("table_prefix") + "_" + defaultTableName
		}
		queuesList := make(map[string]BroadcastInstance, 0)
		var records []models.ChatBroadcast
		for {
			db.Where("is_open = ?", true).Find(&records)
			for _, v := range records {
				if v.ExpireTime.Unix() < time.Now().Unix() {
					v.IsOpen = false
					if db.Save(v).Error == nil {
						if q, ok := queuesList[v.BroadcastId]; ok {
							q.Close()
							delete(queuesList, v.BroadcastId)
						}
					}
				} else if _, ok := queuesList[v.BroadcastId]; !ok {
					instance := BroadcastInstance{
						Record:  v,
						amqpUrl: fmt.Sprintf("amqp://%s:%s@%s/%s", viper.GetString("rabbitmq_user"), viper.GetString("rabbitmq_passwd"), viper.GetString("rabbitmq_host"), viper.GetString("rabbitmq_vhost")),
					}
					instance.Start(db)
					queuesList[v.BroadcastId] = instance
				}
			}
			time.Sleep(5 * time.Second)
		}
	},
}

type BroadcastInstance struct {
	Record  models.ChatBroadcast
	Channel chan interface{}
	amqpUrl string
	conn    *amqp.Connection
	status  bool
}

func (c *BroadcastInstance) Start(db *gorm.DB) {
	c.status = true
	c.Channel = make(chan interface{}, 100)
	go c.Queue("broadcast.message."+c.Record.BroadcastId, 1)
	go c.Receive(db)
	Logger.Info("broadcast start", zap.String("queue", "broadcast.message."+c.Record.BroadcastId))
}

func (c *BroadcastInstance) link(queue string, prefetchCount int) (<-chan amqp.Delivery, error) {
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
		Logger.Error("broadcast.consume", zap.Error(err))
		return deliveries, err
	}
	return deliveries, nil
}

func (c *BroadcastInstance) Receive(db *gorm.DB) {
	var chatMsg uchat.ChatMessageStruct
	for {
		select {
		case msg := <-c.Channel:
			switch msg.(type) {
			case string:
				if msg.(string) == "quit" {
					return
				}
			case amqp.Delivery:
				err := msgpack.Unmarshal(msg.(amqp.Delivery).Body, &chatMsg)
				if err != nil {
					Logger.Error("msgpack unmarshal error", zap.Error(err), zap.Any("msg", msg))
				}
				err = uchat.SendBroadcast(c.Record.BroadcastChats, chatMsg, db)
				if err != nil {
					Logger.Error("create message error", zap.Error(err))
				}
			}
		}
	}
}

func (c *BroadcastInstance) Queue(queue string, prefetchCount int) {
	for {
		if c.status == false {
			//退出
			return
		}
		time.Sleep(3 * time.Second)
		deliveries, err := c.link(queue, prefetchCount)
		if err != nil {
			Logger.Error("Consumer Link Error", zap.Error(err))
			continue
		}
		for msg := range deliveries {
			c.Channel <- msg
			err := msg.Ack(false)
			if err != nil {
				break
			}
		}
		c.conn.Close()
		Logger.Info("Consumer ReConnection After 3 Second")
	}
}

func (c *BroadcastInstance) Close() {
	c.status = false
	c.Channel <- "quit"
	Logger.Info("broadcast close", zap.String("queue", "broadcast.message."+c.Record.BroadcastId))
}

func init() {
	RootCmd.AddCommand(broadcastCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// broadcastCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// broadcastCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
