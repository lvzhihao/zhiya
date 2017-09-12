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

	"go.uber.org/zap"

	"github.com/lvzhihao/uchatlib"
	"github.com/lvzhihao/zhiya/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

// messageCmd represents the message command
var messageCmd = &cobra.Command{
	Use:   "message",
	Short: "预处理群内聊天记录",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		defer Logger.Sync()
		sugar := Logger.Sugar()
		consumer, err := utils.NewReceiveConsumer(
			fmt.Sprintf("amqp://%s:%s@%s/%s", viper.GetString("rabbitmq_user"), viper.GetString("rabbitmq_passwd"), viper.GetString("rabbitmq_host"), viper.GetString("rabbitmq_vhost")))
		if err != nil {
			sugar.Fatal(err)
		}
		utils.Logger = Logger
		consumer.Consumer("uchat.messages.msgpack.process", 1, processMessage) //尽量保证聊天记录的时序，以api回调接口收到消息进入receive队列为准
	},
}

func processMessage(msg amqp.Delivery) {
	ret, err := uchatlib.ConvertUchatMessage(msg.Body)
	if err != nil {
		msg.Ack(false)
		Logger.Error("process error", zap.Error(err), zap.Any("msg", msg))
	} else {
		for _, v := range ret {
			log.Fatal(v)
		}
	}
	/*
		if err != nil {
			Logger.Error("process error", zap.String("queue", "uchat.robot.chat.list"), zap.Error(err), zap.Any("msg", msg))
			msg.Ack(false)
		} else {
			Logger.Info("process success", zap.String("queue", "uchat.robot.chat.list"), zap.Any("msg", msg))
			msg.Ack(false)
		}
	*/
}

func init() {
	RootCmd.AddCommand(messageCmd)
}
