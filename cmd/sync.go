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
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/uchatlib"
	"github.com/lvzhihao/zhiya/models"
	"github.com/lvzhihao/zhiya/uchat"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "同步工具",
	Long:  `手工同步设备、群及用户相关信息`,
	Run: func(cmd *cobra.Command, args []string) {
		defer Logger.Sync()
		sugar := Logger.Sugar()
		action, err := cmd.Flags().GetString("action")
		if err != nil {
			sugar.Fatal(err)
		}
		db, err := gorm.Open("mysql", viper.GetString("mysql_dns"))
		if err != nil {
			sugar.Fatal(err)
		}
		defer db.Close()
		//db.LogMode(true)
		gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
			return viper.GetString("table_prefix") + "_" + defaultTableName
		}
		client := uchatlib.NewClient(viper.GetString("merchant_no"), viper.GetString("merchant_secret"))
		switch action {
		case "robot":
			if err := uchat.SyncRobots(client, db); err == nil {
				sugar.Info("robots sync success")
			} else {
				sugar.Warn(err)
			}
		case "chat":
			var err error
			var objs []models.Robot
			if robots, _ := cmd.Flags().GetString("robots"); robots != "" {
				err = db.Where("serial_no in (?)", strings.Split(robots, ",")).Find(&objs).Error
			} else {
				err = db.Find(&objs).Error
			}
			if err != nil {
				sugar.Fatal(err)
			}
			for _, robot := range objs {
				err := uchat.SyncRobotChatRooms(robot.SerialNo, client, db)
				if err != nil {
					sugar.Fatal(err)
				} else {
					sugar.Infof("chatromms sync success: %s\n", robot.SerialNo)
				}
			}
		case "member":
			var err error
			var objs []models.ChatRoom
			if chats, _ := cmd.Flags().GetString("chats"); chats != "" {
				err = db.Where("chat_room_serial_no in (?)", strings.Split(chats, ",")).Find(&objs).Error
			} else {
				err = db.Where("robot_status = ?", 10).Where("robot_in_status = ?", 0).Find(&objs).Error
			}
			if err != nil {
				sugar.Fatal(err)
			}
			for _, chat := range objs {
				err := uchat.SyncChatRoomMembers(chat.ChatRoomSerialNo, client)
				if err != nil {
					sugar.Fatal(err)
				} else {
					sugar.Infof("members sync success: %s\n", chat.ChatRoomSerialNo)
				}
			}
		case "chatstatus":
			rows, err := db.Model(&models.ChatRoom{}).Select("chat_room_serial_no").Rows()
			if err != nil {
				sugar.Fatal(err)
			}
			defer rows.Close()
			num := 0
			for rows.Next() {
				var no string
				err := rows.Scan(&no)
				if err != nil {
					sugar.Error(err)
				} else {
					err := uchat.SyncChatRoomStatus(no, client, db)
					if err != nil {
						sugar.Error(err)
					} else {
						sugar.Infof("success: %s", no)
					}
				}
				num++
			}
			sugar.Infof("count: %d", num)
		case "openmessage":
			chats, err := cmd.Flags().GetString("chats")
			if err != nil {
				sugar.Fatal(err)
			}
			for _, chat := range strings.Split(chats, ",") {
				err := uchatlib.SetChatRoomOpenGetMessage(chat, client)
				if err != nil {
					sugar.Fatal(err)
				} else {
					sugar.Infof("open GetMessage success: %s\n", chat)
				}
			}
		case "closemessage":
			chats, err := cmd.Flags().GetString("chats")
			if err != nil {
				sugar.Fatal(err)
			}
			for _, chat := range strings.Split(chats, ",") {
				err := uchatlib.SetChatRoomCloseGetMessage(chat, client)
				if err != nil {
					sugar.Fatal(err)
				} else {
					sugar.Infof("close GetMessage success: %s\n", chat)
				}
			}
		case "checkqrcode":
			var chatRooms []models.ChatRoom
			var err error
			if robots, _ := cmd.Flags().GetString("robots"); robots != "" {
				err = db.Where("robot_serial_no in (?)", strings.Split(robots, ",")).Where("robot_status = ?", 10).Where("robot_in_status = ?", 0).Where("(qr_code is null OR qr_code_expired_date < NOW())").Find(&chatRooms).Error
			} else {
				err = db.Where("robot_status = ?", 10).Where("robot_in_status = ?", 0).Where("(qr_code is null OR qr_code_expired_date < NOW())").Limit(10).Find(&chatRooms).Error
			}
			if err != nil {
				sugar.Fatal(err)
			}
			//sugar.Info(chatRooms)
			for _, room := range chatRooms {
				err := uchatlib.ApplyChatRoomQrCode(room.ChatRoomSerialNo, client)
				if err != nil {
					sugar.Warn(err)
				} else {
					sugar.Infof("create new qrcode success: %s\n", room.ChatRoomSerialNo)
				}
			}
		default:
			sugar.Warn("only support robot/chat/member/chatstatus")
		}
	},
}

func init() {
	RootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	//syncCmd.PersistentFlags().String("action", "", "数据同步[robot/chat/member]")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	syncCmd.Flags().String("action", "", "数据同步[robot/chat/user]")
	syncCmd.Flags().String("robots", "", "同步设备关联，多个设备请用半角逗号分隔")
	syncCmd.Flags().String("chats", "", "同步群关联，多个群请用半角逗号分隔")
}
