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
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/zhiya/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var messageQueueConfig = map[string]string{
	"uchat.mysql.message.queue": "uchat.mysql.message.queue",
	// global
}

// mqueueCmd represents the mqueue command
var mqueueCmd = &cobra.Command{
	Use:   "mqueue",
	Short: "读取消息数据进入发入发送队列",
	Long:  `读取消息数据进入发入发送队列`,
	Run: func(cmd *cobra.Command, args []string) {
		defer Logger.Sync()
		//sugar := Logger.Sugar()
		db, err := gorm.Open("mysql", viper.GetString("mysql_dns"))
		if err != nil {
			Logger.Sugar().Fatal(err)
		}
		defer db.Close()
		gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
			return viper.GetString("table_prefix") + "_" + defaultTableName
		}
		MQueue(db)
	},
}

func MQueue(db *gorm.DB) {
RetryConnect:
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/%s", viper.GetString("rabbitmq_user"), viper.GetString("rabbitmq_passwd"), viper.GetString("rabbitmq_host"), viper.GetString("rabbitmq_vhost")))
	if err != nil {
		Logger.Error("Channel Connection Error 1", zap.Error(err))
		time.Sleep(3 * time.Second)
		goto RetryConnect
	}
	channel, err := conn.Channel()
	if err != nil {
		Logger.Error("Channel Connection Error 2", zap.Error(err))
		conn.Close()
		time.Sleep(3 * time.Second)
		goto RetryConnect
	}
	err = channel.ExchangeDeclare(viper.GetString("rabbitmq_message_exchange_name"), "topic", true, false, false, false, nil)
	if err != nil {
		Logger.Error("Channel Connection Error 3", zap.Error(err))
		conn.Close()
		time.Sleep(3 * time.Second)
		goto RetryConnect
	}
BreakFor:
	for {
		var msg1 []models.MessageQueue
		err := db.Where("send_type = 1").Where("send_status = 0").Order("id asc").Limit(100).Find(&msg1).Error
		if err != nil {
			Logger.Error("Load Message Error 1", zap.Error(err))
			continue
		}
		var msg2 []models.MessageQueue
		err = db.Where("send_type = 2").Where("send_status = 0").Where("send_time <= ?", time.Now()).Order("id asc").Limit(100).Find(&msg2).Error
		if err != nil {
			Logger.Error("Load Message Error 2", zap.Error(err))
		}
		msg1 = append(msg1, msg2...)
		for _, msg := range msg1 {
			chatRooms := strings.Split(msg.ChatRoomSerialNoList, ",")
			msg.SendStatus = 1
			msg.SendStatusTime = time.Now()
			err := db.Save(&msg).Error
			if err != nil {
				Logger.Error("Load Message Error 4", zap.Error(err))
			} else {
				Logger.Info("Load Message Success", zap.Any("msg", msg))
				for _, chatRoom := range chatRooms {
					chatRoom := strings.TrimSpace(chatRoom)
					if chatRoom == "" {
						continue
					}
					robotChatRoom, err := models.FindRobotChatRoomByChatRoom(db, chatRoom)
					if err != nil {
						Logger.Error("Load Message Error 3", zap.Error(err))
						continue
					}
					rst := make(map[string]interface{}, 0)
					rst["vcRelaSerialNo"] = msg.QueueId
					rst["vcChatRoomSerialNo"] = chatRoom
					rst["vcRobotSerialNo"] = robotChatRoom.RobotSerialNo
					if msg.IsHit {
						rst["nIsHit"] = "1"
					} else {
						rst["nIsHit"] = "0"
					}
					rst["vcWeixinSerialNo"] = msg.WeixinSerialNo
					//优化
					datas := make([]map[string]string, 0)
					data := make(map[string]string, 0)
					data["nMsgType"] = msg.MsgType
					data["msgContent"] = msg.MsgContent
					data["vcTitle"] = msg.Title
					data["vcDesc"] = msg.Description
					data["vcHref"] = msg.Href
					data["nVoiceTime"] = goutils.ToString(msg.VoiceTime)
					datas = append(datas, data)
					rst["Data"] = datas
					b, _ := json.Marshal(rst)
					mqMsg := amqp.Publishing{
						DeliveryMode: amqp.Persistent,
						Timestamp:    time.Now(),
						ContentType:  "application/json",
						Body:         b,
					}
					err = channel.Publish(viper.GetString("rabbitmq_message_exchange_name"), "uchat.mysql.message.queue", false, false, mqMsg)
					if err != nil {
						Logger.Error("Channel Connection Error 4", zap.String("route", "uchat.mysql.message.queue"), zap.Error(err))
						// retry
						break BreakFor
					}
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
	goto RetryConnect
}

func init() {
	RootCmd.AddCommand(mqueueCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mqueueCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mqueueCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
