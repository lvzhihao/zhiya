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
		}
	]
`

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

		migrateSql(db.AutoMigrate(&models.RobotApplyCode{}).Error)

		migrateSql(db.AutoMigrate(&models.Robot{}).Error)

		migrateSql(db.AutoMigrate(&models.MyRobot{}).Error)

		migrateSql(db.AutoMigrate(&models.ChatRoom{}).Error)

		migrateSql(db.AutoMigrate(&models.ChatRoomTag{}).Error)

		migrateSql(db.AutoMigrate(&models.RobotChatRoom{}).Error)

		migrateSql(db.Model(&models.RobotChatRoom{}).AddUniqueIndex("idx_robot_no_chat_no", "robot_serial_no", "chat_room_serial_no").Error)

		migrateSql(db.AutoMigrate(&models.CmdType{}).Error)

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

		migrateSql(db.AutoMigrate(&models.ChatRoomCmd{}).Error)

		migrateSql(db.AutoMigrate(&models.MyCmd{}).Error)

		migrateSql(db.AutoMigrate(&models.SubCmd{}).Error)

		migrateSql(db.AutoMigrate(&models.TagCmd{}).Error)

		migrateSql(db.AutoMigrate(&models.MessageQueue{}).Error)

		migrateSql(db.AutoMigrate(&models.ChatRoomMember{}).Error)

		migrateSql(db.Model(&models.ChatRoomMember{}).AddUniqueIndex("idx_chat_no_member_no", "chat_room_serial_no", "wx_user_serial_no").Error)

		migrateSql(db.AutoMigrate(&models.MySubChatRoomConfig{}).Error)

		log.Println("db migrate success")

		// receive queue
		for k, v := range receiveQueueConfig {
			migrateQueue(k, viper.GetString("rabbitmq_receive_exchange_name"), v)
		}

		// message queue
		for k, v := range messageQueueConfig {
			migrateQueue(k, viper.GetString("rabbitmq_message_exchange_name"), v)
		}

		// command queue
		for _, v := range initCmdType {
			vname := "cmd." + v.TypeFlag
			migrateQueue(vname, viper.GetString("rabbitmq_command_exchange_name"), vname)
		}
	},
}

func migrateSql(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func migrateQueue(name, exchange, key string) {
	rmqApi := viper.GetString("rabbitmq_api")
	rmqUser := viper.GetString("rabbitmq_user")
	rmqPasswd := viper.GetString("rabbitmq_passwd")
	rmqVhost := viper.GetString("rabbitmq_vhost")
	err := RegisterQueue(
		rmqApi, rmqUser, rmqPasswd, rmqVhost,
		name, exchange, key,
	)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("queue create success: %s bind %s %s\n", name, exchange, key)
	}
}

func RegisterQueue(api, user, passwd, vhost, name, exchange, key string) error {
	client := &http.Client{}
	b := bytes.NewBufferString(`{"auto_delete":false, "durable":true, "arguments":[]}`)
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/queues/%s/%s", api, vhost, name), b)
	if err != nil {
		return err
	}
	// enusre queue
	req.SetBasicAuth(user, passwd)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("CreateQueue StatusError: %d, %v", resp.StatusCode, resp)
	}
	if exchange != "" && key != "" {
		b = bytes.NewBufferString(`{"routing_key":"` + key + `", "arguments":[]}`)
		// ensure binding
		req, err = http.NewRequest(
			"POST",
			fmt.Sprintf("%s/bindings/%s/e/%s/q/%s", api, vhost, exchange, name),
			b)
		req.SetBasicAuth(user, passwd)
		req.Header.Add("Content-Type", "application/json")
		resp, err = client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("BindRoutingKey StatusError: %d, %v", resp.StatusCode, resp)
		}
	}
	return nil
}

func init() {
	RootCmd.AddCommand(migrateCmd)
}
