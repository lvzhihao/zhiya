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

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/uchatlib"
	"github.com/lvzhihao/zhiya/apis"
	"github.com/lvzhihao/zhiya/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "rust api 服务",
	Long:  `rust api 支持，仅限内网调用`,
	Run: func(cmd *cobra.Command, args []string) {
		defer Logger.Sync()
		//app.Logger.SetLevel(log.INFO)
		app := goutils.NewEcho()
		client := uchatlib.NewClient(viper.GetString("merchant_no"), viper.GetString("merchant_secret"))
		db, err := gorm.Open("mysql", viper.GetString("mysql_dns"))
		if err != nil {
			Logger.Sugar().Fatal(err)
		}
		defer db.Close()
		gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
			return viper.GetString("table_prefix") + "_" + defaultTableName
		}
		tool, err := utils.NewTool(
			fmt.Sprintf("amqp://%s:%s@%s/%s", viper.GetString("rabbitmq_user"), viper.GetString("rabbitmq_passwd"), viper.GetString("rabbitmq_host"), viper.GetString("rabbitmq_vhost")),
			viper.GetString("rabbitmq_message_exchange_name"),
			[]string{"uchat.mysql.message.queue"},
		)
		if err != nil {
			Logger.Error("RabbitMQ Connect Error", zap.Error(err))
		}
		utils.Logger = Logger
		apis.Logger = Logger
		apis.DB = db
		apis.Client = client
		apis.Tool = tool
		// action
		app.POST("/api/ping", func(ctx echo.Context) error {
			return ctx.String(http.StatusOK, "pong")
		})
		app.POST("/api/applycode", apis.ApplyCode)
		app.POST("/api/syncrobots", apis.SyncRobots)
		app.POST("/api/syncchats", apis.SyncChats)
		app.POST("/api/overchatroom", apis.OverChatRoom)
		app.POST("/api/welcome", apis.ChatRoomMemberJoinWelcome) //201706221050000271
		app.POST("/api/robotadduser", apis.RobotAddUser)
		app.POST("/api/sendmessage", apis.SendMessage)
		app.POST("/api/pyrobotloginqr", apis.PyRobotLoginQr)
		app.POST("/api/chatroomkicking", apis.ChatRoomKicking)
		app.POST("/api/applychatroomqrcode", apis.ApplyChatRoomQrCode)
		// api v2 support for prism
		app.POST("/api/v2/sendmessage", apis.SendMessageV2)
		app.GET("/api/v2/robot/join", apis.GetRobotJoinList)
		app.POST("/api/v2/robot/join/delete", apis.DeleteRobotJoin)
		app.PUT("/api/v2/robot/chatroom/open", apis.OpenChatRoom)
		app.POST("/api/v2/robot/info", apis.UpdateRobotInfo)
		app.POST("/api/v2/chatroom/over", apis.OverChatRoomV2)
		app.GET("/api/v2/cmd/type", apis.CmdTypeList)
		app.PUT("/api/v2/work/template", apis.CreateWorkTemplate)
		app.POST("/api/v2/work/template", apis.UpdateWorkTemplate)
		app.GET("/api/v2/work/template/list", apis.WorkTemplateList)
		app.GET("/api/v2/work/template", apis.WorkTemplate)
		app.POST("/api/v2/work/template/default", apis.SetWorkTemplateDefault)
		// graceful shutdown
		goutils.EchoStartWithGracefulShutdown(app, ":8079")
	},
}

func init() {
	RootCmd.AddCommand(apiCmd)
}
