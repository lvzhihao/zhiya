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
	"github.com/lvzhihao/uchat/models"
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
		err = db.AutoMigrate(&models.RobotApplyCode{}).Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.AutoMigrate(&models.Robot{}).Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.AutoMigrate(&models.MyRobot{}).Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.AutoMigrate(&models.ChatRoom{}).Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.AutoMigrate(&models.ChatRoomTag{}).Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.AutoMigrate(&models.RobotChatRoom{}).Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.Model(&models.RobotChatRoom{}).AddUniqueIndex("idx_robot_no_chat_no", "robot_serial_no", "chat_room_serial_no").Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.AutoMigrate(&models.CmdType{}).Error
		if err != nil {
			log.Fatal(err)
		}
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
		err = db.AutoMigrate(&models.ChatRoomCmd{}).Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.AutoMigrate(&models.MyCmd{}).Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.AutoMigrate(&models.SubCmd{}).Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.AutoMigrate(&models.TagCmd{}).Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.AutoMigrate(&models.MessageQueue{}).Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.AutoMigrate(&models.ChatRoomMember{}).Error
		if err != nil {
			log.Fatal(err)
		}
		err = db.Model(&models.ChatRoomMember{}).AddUniqueIndex("idx_chat_no_member_no", "chat_room_serial_no", "wx_user_serial_no").Error
		if err != nil {
			log.Fatal(err)
		}
		log.Println("db migrate success")

		for k, v := range receiveQueueConfig {
			client := &http.Client{}
			b := bytes.NewBufferString("{\"auto_delete\":false,\"durable\":true,\"arguments\":[]}")
			req, err := http.NewRequest("PUT", fmt.Sprintf("%s/queues/%s/%s", viper.GetString("rabbitmq_api"), viper.GetString("rabbitmq_vhost"), k), b)
			if err != nil {
				log.Fatal(err)
			}
			// enusre queue
			req.SetBasicAuth(viper.GetString("rabbitmq_user"), viper.GetString("rabbitmq_passwd"))
			req.Header.Add("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			if resp.StatusCode != http.StatusNoContent {
				log.Fatal(resp)
			}
			b = bytes.NewBufferString("{\"routing_key\":\"" + v + "\",\"arguments\":[]}")
			// ensure binding
			req, err = http.NewRequest("POST", fmt.Sprintf("%s/bindings/%s/e/%s/q/%s", viper.GetString("rabbitmq_api"), viper.GetString("rabbitmq_vhost"), viper.GetString("rabbitmq_exchange_name"), k), b)
			req.SetBasicAuth(viper.GetString("rabbitmq_user"), viper.GetString("rabbitmq_passwd"))
			req.Header.Add("Content-Type", "application/json")
			resp, err = client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			if resp.StatusCode != http.StatusCreated {
				log.Fatal(resp)
			}
			log.Printf("queue create success: %s bind %s %s\n", k, viper.GetString("rabbitmq_exchange_name"), v)
		}
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)
}
