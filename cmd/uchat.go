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
	"encoding/json"
	"fmt"
	"log"

	"go.uber.org/zap"

	"github.com/jinzhu/gorm"
	rmqtool "github.com/lvzhihao/go-rmqtool"
	"github.com/lvzhihao/uchatlib"
	"github.com/lvzhihao/zhiya/chatbot"
	"github.com/lvzhihao/zhiya/uchat"
	"github.com/lvzhihao/zhiya/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

// uchatCmd represents the uchat command
var uchatCmd = &cobra.Command{
	Use:   "uchat",
	Short: "小U机器人队列调用",
	Long:  `根据队列名执行相应消费程序`,
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
		consumer, err := utils.NewReceiveConsumer(fmt.Sprintf("amqp://%s:%s@%s/%s", viper.GetString("rabbitmq_user"), viper.GetString("rabbitmq_passwd"), viper.GetString("rabbitmq_host"), viper.GetString("rabbitmq_vhost")))
		if err != nil {
			sugar.Fatal(err)
		}
		utils.Logger = Logger
		switch queue {
		case "uchat.member.list":
			consumer.Consumer("uchat.member.list", 20, shell.MemberList)
		case "uchat.robot.chat.list":
			consumer.Consumer("uchat.robot.chat.list", 20, shell.ChatRoomList)
		case "uchat.member.message_sum":
			consumer.Consumer("uchat.member.message_sum", 20, shell.MemberMessageSum)
		case "uchat.chat.create":
			consumer.Consumer("uchat.chat.create", 20, shell.ChatRoomCreate)
		case "uchat.member.quit":
			consumer.Consumer("uchat.member.quit", 20, shell.MemberQuit)
		case "uchat.member.join":
			consumer.Consumer("uchat.member.join", 20, shell.MemberJoin)
		case "uchat.mysql.message.queue":
			//consumer.Consumer("uchat.mysql.message.queue", 20, shell.SendMessage)
			shell.SendChatMessage(queue) //
		case "uchat.chat.message":
			consumer.Consumer("uchat.chat.message", 20, shell.ChatMessage)
		case "uchat.chat.keyword":
			consumer.Consumer("uchat.chat.keyword", 20, shell.ChatKeyword)
		case "uchat.web.unbindrooms":
			consumer.Consumer("uchat.web.unbindrooms", 20, shell.ChatOver)
		case "uchat.chat.qrcode":
			consumer.Consumer("uchat.chat.qrcode", 20, shell.ChatQrCode)
		case "uchat.robot.friend.add":
			consumer.Consumer("uchat.robot.friend.add", 20, shell.RobotFriendAdd)
		case "uchat.robot.chat.join":
			consumer.Consumer("uchat.robot.chat.join", 20, shell.RobotChatJoin)
		default:
			sugar.Fatal("Please input current queue name")
		}
	},
}

type consumerShell struct {
	db            *gorm.DB
	managerDB     *gorm.DB
	client        *uchatlib.UchatClient
	sendTool      *utils.ReceiveTool
	chatBotClient *chatbot.Client
	conn          *rmqtool.Connect
	publisher     *rmqtool.PublisherTool
}

func (c *consumerShell) Init() (err error) {
	c.db, err = gorm.Open("mysql", viper.GetString("mysql_dns"))
	if err != nil {
		return
	}
	c.managerDB, err = gorm.Open("mysql", viper.GetString("manager_mysql_dns"))
	if err != nil {
		return
	}
	c.client = uchatlib.NewClient(viper.GetString("merchant_no"), viper.GetString("merchant_secret"))
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return viper.GetString("table_prefix") + "_" + defaultTableName
	}
	c.sendTool, err = utils.NewTool(
		fmt.Sprintf("amqp://%s:%s@%s/%s", viper.GetString("rabbitmq_user"), viper.GetString("rabbitmq_passwd"), viper.GetString("rabbitmq_host"), viper.GetString("rabbitmq_vhost")),
		viper.GetString("rabbitmq_message_exchange_name"),
		[]string{"uchat.mysql.message.queue"},
	)
	if err != nil {
		return
	}
	c.chatBotClient = chatbot.NewClient(&chatbot.ClientConfig{
		ApiHost:        viper.GetString("chatbot_api_host"),
		MerchantNo:     viper.GetString("chatbot_merchant_no"),
		MerchantSecret: viper.GetString("chatbot_merchant_secret"),
	})
	c.conn = rmqtool.NewConnect(rmqtool.ConnectConfig{
		Host:     viper.GetString("rabbitmq_host"),
		Api:      viper.GetString("rabbitmq_api"),
		User:     viper.GetString("rabbitmq_user"),
		Passwd:   viper.GetString("rabbitmq_passwd"),
		Vhost:    viper.GetString("rabbitmq_vhost"),
		MetaData: nil,
	})
	err = c.conn.QuickCreateExchange(viper.GetString("rabbitmq_event_exchange_name"), "topic", true)
	if err != nil {
		return
	}
	c.publisher, err = c.conn.ApplyPublisher(viper.GetString("rabbitmq_event_exchange_name"), []string{
		"event",
	})
	if err != nil {
		return
	}
	return
}

func (c *consumerShell) Close() {
	defer c.db.Close()
	defer c.managerDB.Close()
}

// 群成员列表
func (c *consumerShell) MemberList(msg amqp.Delivery) {
	err := uchat.SyncChatRoomMembersCallback(msg.Body, c.db)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.member.list"), zap.Error(err), zap.Any("msg", msg))
		msg.Ack(false)
	} else {
		Logger.Info("process success", zap.String("queue", "uchat.member.list"), zap.Any("msg", msg))
		msg.Ack(false)
	}
	//msg.Nack(false, true)
}

func (c *consumerShell) ChatRoomList(msg amqp.Delivery) {
	err := uchat.SyncRobotChatRoomsCallback(msg.Body, c.db)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.robot.chat.list"), zap.Error(err), zap.Any("msg", msg))
		msg.Ack(false)
	} else {
		Logger.Info("process success", zap.String("queue", "uchat.robot.chat.list"), zap.Any("msg", msg))
		msg.Ack(false)
	}
}

func (c *consumerShell) MemberMessageSum(msg amqp.Delivery) {
	err := uchat.SyncMemberMessageSumCallback(msg.Body, c.db)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.member.message_sum"), zap.Error(err), zap.Any("msg", msg))
		msg.Ack(false)
	} else {
		Logger.Info("process success", zap.String("queue", "uchat.member.message_sum"), zap.Any("msg", msg))
		msg.Ack(false)
	}
}

func (c *consumerShell) ChatRoomCreate(msg amqp.Delivery) {
	err := uchat.SyncChatRoomCreateCallback(msg.Body, c.client, c.db, c.publisher)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.chat.create"), zap.Error(err), zap.Any("msg", msg))
		//msg.Nack(false, true)
		//开通错误需要重试，不成功则需要人工干预，扔回队列
		msg.Ack(false)
	} else {
		Logger.Info("process success", zap.String("queue", "uchat.chat.create"), zap.Any("msg", msg))
		msg.Ack(false)
	}
}

func (c *consumerShell) MemberQuit(msg amqp.Delivery) {
	err := uchat.SyncMemberQuitCallback(msg.Body, c.db)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.member.quit"), zap.Error(err), zap.Any("msg", msg))
		msg.Ack(false)
	} else {
		Logger.Info("process success", zap.String("queue", "uchat.member.quit"), zap.Any("msg", msg))
		msg.Ack(false)
	}
}

func (c *consumerShell) MemberJoin(msg amqp.Delivery) {
	err := uchat.SyncMemberJoinCallback(msg.Body, c.db, c.managerDB)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.member.join"), zap.Error(err), zap.Any("msg", msg))
		msg.Ack(false)
	} else {
		Logger.Info("process success", zap.String("queue", "uchat.member.join"), zap.Any("msg", msg))
		msg.Ack(false)
	}
}

func (c *consumerShell) SendMessage(msg amqp.Delivery) {
	var rst map[string]interface{}
	err := json.Unmarshal(msg.Body, &rst)
	if err != nil {
		Logger.Error("process error json unmarshal", zap.String("queue", "uchat.mysql.message.queue"), zap.Error(err), zap.Any("msg", msg))
		msg.Ack(false)
	} else {
		err := c.client.SendMessage(rst)
		if err != nil {
			Logger.Error("process error send error", zap.String("queue", "uchat.mysql.message.queue"), zap.Error(err), zap.Any("msg", msg))
			//msg.Nack(false, true)
			msg.Ack(false)
		} else {
			Logger.Error("process success", zap.String("queue", "uchat.mysql.message.queue"), zap.Error(err), zap.Any("msg", msg))
			msg.Ack(false)
		}
	}
}

func (c *consumerShell) ChatMessage(msg amqp.Delivery) {
	err := uchat.SyncChatMessageCallback(msg.Body, c.db, c.managerDB, c.sendTool)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.chat.message"), zap.Error(err), zap.Any("msg", msg))
		msg.Ack(false)
	} else {
		Logger.Info("process success", zap.String("queue", "uchat.chat.message"), zap.Any("msg", msg))
		msg.Ack(false)
	}
}

func (c *consumerShell) ChatKeyword(msg amqp.Delivery) {
	err := uchat.SyncChatKeywordCallback(msg.Body, c.db, c.managerDB, c.sendTool, c.chatBotClient)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.chat.keyword"), zap.Error(err), zap.Any("msg", msg))
		msg.Ack(false)
	} else {
		Logger.Info("process success", zap.String("queue", "uchat.chat.keyword"), zap.Any("msg", msg))
		msg.Ack(false)
	}
}

func (c *consumerShell) ChatOver(msg amqp.Delivery) {
	err := uchat.SyncChatOverCallback(msg.Body, c.client)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.web.unbindrooms"), zap.Error(err), zap.Any("msg", msg))
		msg.Ack(false)
	} else {
		Logger.Info("process success", zap.String("queue", "uchat.web.unbindrooms"), zap.Any("msg", msg))
		msg.Ack(false)
	}
}

func (c *consumerShell) ChatQrCode(msg amqp.Delivery) {
	err := uchat.SyncChatQrCodeCallback(msg.Body, c.db)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.chat.qrcode"), zap.Error(err), zap.Any("msg", msg))
		msg.Ack(false)
	} else {
		Logger.Info("process success", zap.String("queue", "uchat.chat.qrcode"), zap.Any("msg", msg))
		msg.Ack(false)
	}
}

func (c *consumerShell) RobotFriendAdd(msg amqp.Delivery) {
	err := uchat.SyncRobotFriendAddCallback(msg.Body, c.db)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.robot.friend.add"), zap.Error(err), zap.Any("msg", msg))
		msg.Ack(false)
	} else {
		Logger.Info("process success", zap.String("queue", "uchat.robot.friend.add"), zap.Any("msg", msg))
		msg.Ack(false)
	}
}

func (c *consumerShell) RobotChatJoin(msg amqp.Delivery) {
	err := uchat.SyncRobotChatJoinCallback(msg.Body, c.db)
	if err != nil {
		Logger.Error("process error", zap.String("queue", "uchat.robot.chat.join"), zap.Error(err), zap.Any("msg", msg))
		msg.Ack(false)
	} else {
		Logger.Info("process success", zap.String("queue", "uchat.robot.chat.join"), zap.Any("msg", msg))
		msg.Ack(false)
	}
}

func (c *consumerShell) SendChatMessage(name string) {
	queue := c.conn.ApplyQueue(name)
	err := queue.Ensure([]*rmqtool.QueueBind{
		&rmqtool.QueueBind{
			Key:       "uchat.mysql.message.queue", // old
			Exchange:  viper.GetString("rabbitmq_message_exchange_name"),
			Arguments: nil,
		},
		&rmqtool.QueueBind{
			Key:       "uchat.chat.message.#", // new
			Exchange:  viper.GetString("rabbitmq_message_exchange_name"),
			Arguments: nil,
		},
	})
	if err != nil {
		Logger.Fatal(name+" error", zap.Error(err))
	}
	queue.Consume(20, func(msg amqp.Delivery) {
		var rst map[string]interface{}
		err := json.Unmarshal(msg.Body, &rst)
		log.Fatal(rst, err)
		if err != nil {
			Logger.Error("process error json unmarshal", zap.String("queue", name), zap.Error(err), zap.Any("msg", msg))
			msg.Ack(false)
		} else {
			err := c.client.SendMessage(rst)
			if err != nil {
				Logger.Error("process error send error", zap.String("queue", name), zap.Error(err), zap.Any("msg", msg))
				//msg.Nack(false, true)
				msg.Ack(false)
			} else {
				Logger.Error("process success", zap.String("queue", name), zap.Error(err), zap.Any("msg", msg))
				msg.Ack(false)
			}
		}
	})
}

func init() {
	RootCmd.AddCommand(uchatCmd)

	uchatCmd.Flags().String("queue", "", "队列名称")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uchatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uchatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
