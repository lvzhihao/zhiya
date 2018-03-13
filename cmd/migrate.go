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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/lvzhihao/zhiya/models"
	"github.com/lvzhihao/zhiya/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var InitCmdTypeValues = `
	[
		{
			"type_flag": "member.join.welcome",
			"type_name": "入群欢迎语"
		},
		{
			"type_flag": "alimama.product.search",
			"type_name": "淘客商品搜索"
		},
		{
			"type_flag": "alimama.coupon.search",
			"type_name": "淘客优惠搜索"
		},
		{
			"type_flag": "shop.custom.search",
			"type_name": "客户自定义搜索"
		}
	]
`

var InitDBModel = []interface{}{
	&models.RobotApplyCode{},       //机器人开通验证码，通过这里的记录验证开通的群属于谁
	&models.Robot{},                //设备列表
	&models.MyRobot{},              //设备归属关系
	&models.ChatRoom{},             //微信群列表
	&models.ChatRoomTag{},          //微信群TAG分组
	&models.RobotChatRoom{},        //微信群归属关系
	&models.CmdType{},              //命令类型
	&models.ChatRoomCmd{},          //微信群命令
	&models.MyCmd{},                //供应商命令
	&models.SubCmd{},               //代理商命令
	&models.TagCmd{},               //TAG分组命令
	&models.MessageQueue{},         //发送消息队列
	&models.ChatRoomMember{},       //微信群用户
	&models.MySubChatRoomConfig{},  //代理商限制配置
	&models.TulingConfig{},         //图灵机器人配置
	&models.RobotFriend{},          //设备好友信息
	&models.MyRobotRenew{},         //设备续费日志
	&models.WorkTemplate{},         //指令模板
	&models.ChatRoomWorkTemplate{}, //群指令配置
	&models.RobotJoin{},            //机器人入群信息
}

var receiveQueueConfig = map[string]string{
	"uchat.member.list":         "uchat.member.list",
	"uchat.member.join":         "uchat.member.join",
	"uchat.member.quit":         "uchat.member.quit",
	"uchat.member.message_sum":  "uchat.member.message.sum",
	"uchat.chat.create":         "uchat.chat.create",
	"uchat.chat.message":        "uchat.chat.message",
	"uchat.chat.keyword":        "uchat.chat.keyword",
	"uchat.chat.redpack":        "uchat.chat.redpack",
	"uchat.robot.chat.list":     "uchat.robot.chat.list",
	"uchat.robot.chat.join":     "uchat.robot.chat.join",
	"uchat.robot.deny":          "uchat.robot.deny",
	"uchat.send.messages.error": "uchat.send.messages.error",
	"uchat.chat.qrcode":         "uchat.chat.qrcode",
	"uchat.robot.friend.add":    "uchat.robot.friend.add",
	//"uchat.robot.message.private": "uchat.robot.message.private",
	"uchat.log":      "uchat.#",
	"uchat.messages": "uchat.chat.message",
}

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "初始化及更新操作",
	Long:  `更新数据结构并创建相应回调对列，可重复调用`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := gorm.Open("mysql", viper.GetString("mysql_dns"))
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
			return viper.GetString("table_prefix") + "_" + defaultTableName
		}

		// migrate db table
		for _, m := range InitDBModel {
			log.Printf("%s migrate ... \n", db.NewScope(m).TableName())
			migrateSql(db.AutoMigrate(m).Error)
		}

		migrateSql(db.Model(&models.RobotChatRoom{}).AddUniqueIndex("idx_robot_no_chat_no", "robot_serial_no", "chat_room_serial_no").Error)
		migrateSql(db.Model(&models.ChatRoomMember{}).AddUniqueIndex("idx_chat_no_member_no", "chat_room_serial_no", "wx_user_serial_no").Error)
		log.Println("db migrate success")

		// ensure cmd type data
		var initCmdType []models.CmdType
		err = json.Unmarshal([]byte(InitCmdTypeValues), &initCmdType)
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range initCmdType {
			err := v.Ensure(db, v.TypeFlag)
			if err != nil {
				log.Fatal(err)
			}
			db.Save(&v)
		}
		log.Println("data init success")

		// exchange
		migrateExchange(viper.GetString("rabbitmq_message_exchange_name"))

		// receive queue
		for k, v := range receiveQueueConfig {
			migrateQueue(k, viper.GetString("rabbitmq_receive_exchange_name"), v)
		}

		// message queue
		for k, v := range messageQueueConfig {
			migrateQueue(k, viper.GetString("rabbitmq_message_exchange_name"), v)
		}
	},
}

func migrateSql(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func migrateExchange(exchange string) {
	api := viper.GetString("rabbitmq_api")
	user := viper.GetString("rabbitmq_user")
	passwd := viper.GetString("rabbitmq_passwd")
	vhost := viper.GetString("rabbitmq_vhost")
	client := &http.Client{}
	b := bytes.NewBufferString(`{"type":"topic","auto_delete":false,"durable":true,"internal":false,"arguments":[]}`)
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/exchanges/%s/%s", api, vhost, exchange), b)
	if err != nil {
		log.Fatal(err)
	}
	// enusre queue
	req.SetBasicAuth(user, passwd)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if (resp.StatusCode == http.StatusNoContent) || (resp.StatusCode == http.StatusCreated) {
		log.Printf("exchage create success: %s\n", exchange)
	} else {
		log.Fatal(fmt.Sprintf("CreateExchange StatusError: %d, %v", resp.StatusCode, resp))
	}
}

func migrateQueue(name, exchange, key string) {
	rmqApi := viper.GetString("rabbitmq_api")
	rmqUser := viper.GetString("rabbitmq_user")
	rmqPasswd := viper.GetString("rabbitmq_passwd")
	rmqVhost := viper.GetString("rabbitmq_vhost")
	err := utils.RegisterQueue(
		rmqApi, rmqUser, rmqPasswd, rmqVhost,
		name, exchange, key,
	)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("queue create success: %s bind %s %s\n", name, exchange, key)
	}
}

func init() {
	RootCmd.AddCommand(migrateCmd)
}
