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
	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/zhiya/apis"
	"github.com/lvzhihao/zhiya/uchat"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		client := uchat.NewClient(viper.GetString("merchant_no"), viper.GetString("merchant_secret"))
		/*
			tool, err := NewTool(fmt.Sprintf("amqp://%s:%s@%s/%s", viper.GetString("rabbitmq_user"), viper.GetString("rabbitmq_passwd"), viper.GetString("rabbitmq_host"), viper.GetString("rabbitmq_vhost")))
			if err != nil {
				Logger.Error("RabbitMQ Connect Error", zap.Error(err))
			}
		*/
		db, err := gorm.Open("mysql", viper.GetString("mysql_dns"))
		if err != nil {
			Logger.Sugar().Fatal(err)
		}
		defer db.Close()
		gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
			return viper.GetString("table_prefix") + "_" + defaultTableName
		}
		apis.Logger = Logger
		apis.DB = db
		apis.Client = client
		// action
		app.POST("/api/applycode", apis.ApplyCode)
		// graceful shutdown
		goutils.EchoStartWithGracefulShutdown(app, ":8079")
	},
}

func init() {
	RootCmd.AddCommand(apiCmd)
}
