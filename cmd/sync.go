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
	"github.com/lvzhihao/uchat/models"
	"github.com/lvzhihao/uchat/uchat"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger, _ := zap.NewProduction()
		defer logger.Sync() // flushes buffer, if any
		sugar := logger.Sugar()
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
		client := uchat.NewClient(viper.GetString("merchant_no"), viper.GetString("merchant_secret"))
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
		case "user":
		default:
			sugar.Warn("only support obot/chat/user")
		}
	},
}

func init() {
	RootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	//syncCmd.PersistentFlags().String("action", "", "数据同步[robots/chatroom/users]")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	syncCmd.Flags().String("action", "", "数据同步[robot/chat/user]")
	syncCmd.Flags().String("robots", "", "同步设备关联，多个设备请用半角逗号分隔")

}
