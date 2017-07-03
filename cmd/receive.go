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
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/labstack/echo"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/uchat/uchat"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var receiveQueueConfig = map[string]string{
	"uchat.member.list":           uchat.ReceiveMQMemberList,
	"uchat.member.join":           uchat.ReceiveMQMemberJoin,
	"uchat.member.quit":           uchat.ReceiveMQMemberQuit,
	"uchat.member.message_sum":    uchat.ReceiveMQMemberMessageSum,
	"uchat.robot.chat.list":       uchat.ReceiveMQRobotChatList,
	"uchat.robot.chat.join":       uchat.ReceiveMQRobotJoinChat,
	"uchat.robot.message.private": uchat.ReceiveMQRobotPrivateMessage,
	"uchat.chat.create":           uchat.ReceiveMQChatCreate,
	"uchat.chat.message":          uchat.ReceiveMQChatMessage,
	"uchat.chat.keyword":          uchat.ReceiveMQChatKeyword,
	"uchat.chat.redpack":          uchat.ReceiveMQChatRedpack,
	// global
	"uchat.log": "uchat.#",
}

var receiveActConfig = map[string]string{
	"member_info":    uchat.ReceiveMQMemberList,
	"member_new":     uchat.ReceiveMQMemberJoin,
	"member_quit":    uchat.ReceiveMQMemberQuit,
	"robot_roomlist": uchat.ReceiveMQRobotChatList,
	"keyword":        uchat.ReceiveMQChatKeyword,
	"group_new":      uchat.ReceiveMQChatCreate,
	"group_msg":      uchat.ReceiveMQChatMessage,
	"robot_ingroup":  uchat.ReceiveMQRobotJoinChat,
	"saysum":         uchat.ReceiveMQMemberMessageSum,
	"msg":            uchat.ReceiveMQRobotPrivateMessage,
	"readpack":       uchat.ReceiveMQChatRedpack,
}

type receiveTool struct {
	channels map[string]*receiveChannel
}

func NewTool(url string) (*receiveTool, error) {
	tool := &receiveTool{
		channels: make(map[string]*receiveChannel, 0),
	}
	err := tool.conn(url)
	return tool, err
}

func (c *receiveTool) conn(url string) error {
	_, err := amqp.Dial(url)
	if err != nil {
		return err
	} //test link
	for _, route := range receiveActConfig {
		c.channels[route] = &receiveChannel{
			amqpUrl:  url,
			routeKey: route,
			Channel:  make(chan string, 2000),
		}
		go c.channels[route].Receive()
	}
	return nil
}

func (c *receiveTool) Publish(route, msg string) {
	c.channels[route].Channel <- msg
}

type receiveChannel struct {
	amqpUrl  string
	routeKey string
	Channel  chan string
}

func (c *receiveChannel) Receive() {
	for {
		conn, err := amqp.Dial(c.amqpUrl)
		if err != nil {
			Logger.Error("Channel Connection Error 1", zap.String("route", c.routeKey), zap.Error(err))
			time.Sleep(3 * time.Second)
			continue
		}
		defer conn.Close()
		channel, err := conn.Channel()
		if err != nil {
			Logger.Error("Channel Connection Error 2", zap.String("route", c.routeKey), zap.Error(err))
			time.Sleep(3 * time.Second)
			continue
		}
		err = channel.ExchangeDeclare(viper.GetString("rabbitmq_exchange_name"), "topic", true, false, false, false, nil)
		if err != nil {
			Logger.Error("Channel Connection Error 3", zap.String("route", c.routeKey), zap.Error(err))
			time.Sleep(3 * time.Second)
			continue
		}
		select {
		case str := <-c.Channel:
			if str == "quit" {
				Logger.Info("Channel Connection Quit", zap.String("route", c.routeKey))
				return
			} //quit
			msg := amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				Timestamp:    time.Now(),
				ContentType:  "application/json",
				Body:         []byte(str),
			}
			err := channel.Publish(viper.GetString("rabbitmq_exchange_name"), c.routeKey, false, false, msg)
			if err != nil {
				c.Channel <- str
				conn.Close()
				Logger.Error("Channel Connection Error 4", zap.String("route", c.routeKey), zap.Error(err))
				break
			}
		}
	}
}

// receiveCmd represents the receive command
var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		defer Logger.Sync()
		app := goutils.NewEcho()
		//app.Logger.SetLevel(log.INFO)
		client := uchat.NewClient(viper.GetString("merchant_no"), viper.GetString("merchant_secret"))
		tool, err := NewTool(fmt.Sprintf("amqp://%s:%s@%s/%s", viper.GetString("rabbitmq_user"), viper.GetString("rabbitmq_passwd"), viper.GetString("rabbitmq_host"), viper.GetString("rabbitmq_vhost")))
		if err != nil {
			Logger.Error("RabbitMQ Connect Error", zap.Error(err))
		}
		app.Any("/*", func(ctx echo.Context) error {
			act := ctx.QueryParam("act")
			if mqRoute, ok := receiveActConfig[act]; ok {
				str := ctx.FormValue("strContext")
				if strings.Compare(client.Sign(str), ctx.FormValue("strSign")) == 0 {
					tool.Publish(mqRoute, str)
					Logger.Debug("Receive Message", zap.String("route", mqRoute), zap.String("message", str))
				} else {
					Logger.Error("Error sign", zap.String("strSign", ctx.FormValue("strSign")), zap.String("checkSign", client.Sign(str)))
				}
			} else {
				Logger.Error("Unknow Action", zap.String("action", act))
			}
			return ctx.HTML(http.StatusOK, "SUCCESS")
		})
		goutils.EchoStartWithGracefulShutdown(app, ":8099")
	},
}

//func getConn() (*amqp.Connection, error) {
/*
	_, err = this.Conn.Channel()
	if err != nil {
		log.Printf("channel.open: %s", err)
		return nil, err
	}
	this.Channel, _ = this.Conn.Channel()
	if err := this.Channel.Qos(queueConf.PrefetchCount, 0, false); err != nil {
		log.Printf("channel.qos: %s", err)
		return nil, err
	}
	deliveries, err := this.Channel.Consume(queueConf.Queue, "logsticker", false, false, false, false, nil)
	if err != nil {
		log.Printf("base.consume: %v", err)
		return deliveries, err
	}
	return deliveries, nil
*/
//}

func init() {
	RootCmd.AddCommand(receiveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// receiveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// receiveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
