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
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/uchat/uchat"
	"github.com/spf13/cobra"
)

var receiveActConfig = map[string]string{
	"member_info":    uchat.ReceiveMQMemberList,
	"member_new":     uchat.ReceiveMQMemberJoin,
	"member_quit":    uchat.ReceiveMQMemberQuit,
	"robot_roomlist": uchat.ReceiveMQRobotChatList,
	"keyword":        uchat.ReceiveMQChatKeyword,
	"group_new":      uchat.ReceiveMQChatCreate,
	"group_msg":      uchat.ReceiveMQChatMessage,
	"robot_ingroup":  uchat.ReceiveMQRobotJoinChat,
	"saysum":         uchat.ReceiveMQMemberMessageSum,
	"msg":            uchat.ReceiveMQRobotPrivateMessage,
	"readpack":       uchat.ReceiveMQChatRedpack,
}

// receiveCmd represents the receive command
var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		app := goutils.NewEcho()
		app.Logger.SetLevel(log.INFO)
		app.Any("/*", func(ctx echo.Context) error {
			act := ctx.QueryParam("act")
			if mqRoute, ok := receiveActConfig[act]; ok {
				str := ctx.FormValue("strContext")
				sign := ctx.FormValue("strSign")
				ctx.Logger().Info(mqRoute, str, sign)
				return ctx.HTML(http.StatusOK, "SUCCESS")
			} else {
				ctx.Logger().Errorf("Unknow Action: '%s'", act)
				return ctx.HTML(http.StatusOK, "SUCCESS")
			}
		})
		goutils.EchoStartWithGracefulShutdown(app, ":8099")
	},
}

func init() {
	RootCmd.AddCommand(receiveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// receiveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// receiveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
