package apis

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/uchatlib"
	"github.com/lvzhihao/zhiya/models"
	"github.com/lvzhihao/zhiya/uchat"
	"github.com/lvzhihao/zhiya/utils"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"github.com/spf13/viper"
)

var (
	DB                        *gorm.DB              //数据库
	Logger                    *zap.Logger           //日志
	Client                    *uchatlib.UchatClient //uchat客户端
	Tool                      *utils.ReceiveTool
	DefaultApplyCodeAddMinute string = "10" //默认验证码有效时间
)

type Result struct {
	Code  string      `json:"code"`
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
}

func init() {

}

// 获取验证码
func FetchApplyCode(db *gorm.DB, myId, subId, robotSerialNo, limit string) (*models.RobotApplyCode, error) {
	// myId 不得为空
	if myId == "" {
		return nil, errors.New("error my_id")
	}
	if robotSerialNo == "" {
		// 获取已经存在的验证码，如果有则直接返回
		exists, err := models.FindVaildApplyCodeByMyId(DB, myId, subId)
		if err == nil && len(exists) > 0 {
			return &exists[0], nil
		}
	} else {
		// 获取已经存在的验证码，如果有则直接返回
		exists, err := models.FindVaildApplyCodeByMyIdAndRobot(DB, myId, subId, robotSerialNo)
		if err == nil && len(exists) > 0 {
			return &exists[0], nil
		}
	}
	// todo check subid chatroom limit

	limitNum := 10
	if limit != "" {
		limitNum = int(goutils.ToInt(limit))
	}

	// 查找此用户可用当天加群机器人
	robots, err := models.FindValidCodeRobotByMyId(DB, myId, limitNum)
	if err != nil {
		return nil, err
	}

	//随机选取一个机器人
	numn := rand.Intn(len(robots))
	sendRobotSerialNo := robots[numn].SerialNo
	for _, r := range robots {
		//如果有，则用指定的robotserialNo
		if r.SerialNo == robotSerialNo {
			sendRobotSerialNo = r.SerialNo
			break
		}
	}

	params := map[string]string{
		"vcRobotSerialNo":    sendRobotSerialNo,
		"nType":              "10",
		"vcChatRoomSerialNo": "",
		"nCodeCount":         "1",
		"nAddMinute":         DefaultApplyCodeAddMinute,
		"nIsNotify":          "0",
		"vcNotifyContent":    "",
	}
	// 获取开通验证码
	datas, err := Client.ApplyCodeList(params)
	if err != nil {
		return nil, err
	}

	// 暂时只支持单条验证码获取
	codeData := datas[0]
	loc, _ := time.LoadLocation("Asia/Shanghai")
	applyCode := &models.RobotApplyCode{}
	applyCode.RobotSerialNo = codeData["vcRobotSerialNo"]
	applyCode.RobotNickName = codeData["vcNickName"]
	applyCode.Type = "10"
	applyCode.ChatRoomSerialNo = codeData["vcChatRoomSerialNo"]
	applyCode.ExpireTime, _ = time.ParseInLocation("2006/1/2 15:04:05", codeData["dtEndDate"], loc)
	applyCode.CodeSerialNo = codeData["vcSerialNo"]
	applyCode.CodeValue = codeData["vcCode"]
	applyCode.CodeImages = codeData["vcCodeImages"]
	applyCode.CodeTime, _ = time.ParseInLocation("2006/1/2 15:04:05", codeData["dtCreateDate"], loc)
	applyCode.MyId = myId
	applyCode.SubId = subId
	applyCode.Used = false
	err = DB.Create(applyCode).Error
	return applyCode, err
}

/*
RUSTAPI ApplyCode

`curl -vvv "http://host:prot/applycode" -X POST -d "my_id=xxx&sub_id=xxx"`
*/
func ApplyCode(ctx echo.Context) error {
	myId := ctx.FormValue("my_id")
	subId := ctx.FormValue("sub_id")
	robotSerialNo := ctx.FormValue("robot_serial_no")
	limit := ctx.FormValue("limit")

	applyCode, err := FetchApplyCode(DB, myId, subId, robotSerialNo, limit)
	if err != nil {
		return ctx.JSON(http.StatusOK, Result{
			Code:  "000001",
			Error: err.Error(),
		})
	} else {
		return ctx.JSON(http.StatusOK, Result{
			Code: "000000",
			Data: applyCode,
		})
	}
}

func SyncRobots(ctx echo.Context) error {
	err := uchat.SyncRobots(Client, DB)
	if err != nil {
		return ctx.JSON(http.StatusOK, Result{
			Code:  "000002",
			Error: err.Error(),
		})
	} else {
		return ctx.JSON(http.StatusOK, Result{
			Code: "000000",
			Data: "success",
		})
	}
}

func OverChatRoom(ctx echo.Context) error {
	serialNo := ctx.FormValue("serial_no")
	comment := ctx.FormValue("comment")
	err := uchatlib.SetChatRoomOver(serialNo, comment, Client)
	if err != nil {
		return ctx.JSON(http.StatusOK, Result{
			Code:  "000003",
			Error: err.Error(),
		})
	} else {
		return ctx.JSON(http.StatusOK, Result{
			Code: "000000",
			Data: "success",
		})
	}
}

func ChatRoomMemberJoinWelcome(ctx echo.Context) error {
	serialNo := ctx.FormValue("serial_no")
	msg := uchat.FetchChatRoomMemberJoinMessage(serialNo, DB)
	return ctx.JSON(http.StatusOK, Result{
		Code: "000000",
		Data: msg,
	})
}

func ChatRoomKicking(ctx echo.Context) error {
	post := make(map[string]string, 0)
	post["vcChatRoomSerialNo"] = ctx.FormValue("chat_room_serial_no")
	post["vcXxUserSerialNo"] = ctx.FormValue("wx_user_serial_no")
	post["vcComment"] = ctx.FormValue("comment")
	err := uchatlib.ChatRoomKicking(post)
	if err != nil {
		return ctx.JSON(http.StatusOK, Result{
			Code:  "000009",
			Error: err.Error(),
		})
	} else {
		return ctx.JSON(http.StatusOK, Result{
			Code: "000000",
			Data: "success",
		})
	}
}

func RobotAddUser(ctx echo.Context) error {
	robotSerialNo := ctx.FormValue("robot_serial_no")
	userWeixinId := ctx.FormValue("user_weixin_id")
	err := uchatlib.ApplyRobotAddUser(robotSerialNo, userWeixinId, Client)
	if err != nil {
		return ctx.JSON(http.StatusOK, Result{
			Code:  "000004",
			Error: err.Error(),
		})
	} else {
		return ctx.JSON(http.StatusOK, Result{
			Code: "000000",
			Data: "success",
		})
	}
}

func SendMessage(ctx echo.Context) error {
	var rst map[string]interface{}
	err := json.Unmarshal([]byte(ctx.FormValue("context")), &rst)
	if err != nil {
		return ctx.JSON(http.StatusOK, Result{
			Code:  "000005",
			Error: err.Error(),
		})
	}
	// 补充商户号和发送流水号
	rst["MerchantNo"] = viper.GetString("merchant_no")
	rst["vcRelaSerialNo"] = "api-" + goutils.RandomString(20)
	b, _ := json.Marshal(rst)
	Tool.Publish("uchat.mysql.message.queue", goutils.ToString(b))
	return ctx.JSON(http.StatusOK, Result{
		Code: "000000",
		Data: "success",
	})
}

type PyRobotLoginQrInput struct {
	RobotName string `json:"robot_name"`
	QrCode    string `json:"qr_code"`
}

func PyRobotLoginQr(ctx echo.Context) error {
	b, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return ctx.JSON(http.StatusOK, Result{
			Code:  "000006",
			Error: err.Error(),
		})
	}
	var input PyRobotLoginQrInput
	err = json.Unmarshal(b, &input)
	if err != nil {
		return ctx.JSON(http.StatusOK, Result{
			Code:  "000007",
			Error: err.Error(),
		})
	}
	rst := make(map[string]interface{}, 0)
	if len(input.QrCode) > 0 {
		data, err := base64.StdEncoding.DecodeString(input.QrCode)
		if err != nil {
			return ctx.JSON(http.StatusOK, Result{
				Code:  "000008",
				Error: err.Error(),
			})
		}
		filename := fmt.Sprintf("qrcode/%x-%d.png", md5.Sum([]byte(input.RobotName)), time.Now().UnixNano())
		_, err = UploadQiuniu(filename, data)
		if err != nil {
			return ctx.JSON(http.StatusOK, Result{
				Code:  "000009",
				Error: err.Error(),
			})
		}
		qrUrl := fmt.Sprintf("https://%s/%s", viper.GetString("qiniu_zhiya_domain"), filename)
		rst["MerchantNo"] = viper.GetString("merchant_no")
		rst["vcRelaSerialNo"] = "qrcode-" + goutils.RandomString(20)
		rst["vcChatRoomSerialNo"] = viper.GetString("py_robot_send_chat")
		rst["vcRobotSerialNo"] = viper.GetString("py_robot_send_robot")
		rst["nIsHit"] = 1
		rst["vcWeixinSerialNo"] = ""
		rst["Data"] = []map[string]string{
			map[string]string{
				"nMsgType":   "2001",
				"msgContent": "小程序机器人( " + input.RobotName + " )需要重新登录，请用手机识别下方二维码登录",
				"vcTitle":    "",
				"vcDesc":     "",
				"nVoiceTime": "0",
				"vcHref":     "",
			},
			map[string]string{
				"nMsgType":   "2002",
				"msgContent": qrUrl,
				"vcTitle":    "",
				"vcDesc":     "",
				"nVoiceTime": "0",
				"vcHref":     "",
			},
		}
	} else {
		rst["MerchantNo"] = viper.GetString("merchant_no")
		rst["vcRelaSerialNo"] = "qrcode-" + goutils.RandomString(20)
		rst["vcChatRoomSerialNo"] = viper.GetString("py_robot_send_chat")
		rst["vcRobotSerialNo"] = viper.GetString("py_robot_send_robot")
		rst["nIsHit"] = 1
		rst["vcWeixinSerialNo"] = ""
		rst["Data"] = []map[string]string{
			map[string]string{
				"nMsgType":   "2001",
				"msgContent": "小程序机器人( " + input.RobotName + " )需要重新登录，请去手机上确认登录",
				"vcTitle":    "",
				"vcDesc":     "",
				"nVoiceTime": "0",
				"vcHref":     "",
			},
		}
	}
	b, _ = json.Marshal(rst)
	Tool.Publish("uchat.mysql.message.queue", goutils.ToString(b))
	return ctx.JSON(http.StatusOK, Result{
		Code: "000000",
		Data: "success",
	})
}

func UploadQiuniu(filename string, data []byte) (storage.PutRet, error) {
	putPolicy := storage.PutPolicy{
		Scope: viper.GetString("qiniu_zhiya_bucket"),
	}
	mac := qbox.NewMac(viper.GetString("qiniu_access_key"), viper.GetString("qiniu_secret_key"))
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.ZoneHuadong
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{
	/*
		Params: map[string]string{
			"x:name": "",
		},
	*/
	}
	dataLen := int64(len(data))
	err := formUploader.Put(context.Background(), &ret, upToken, filename, bytes.NewReader(data), dataLen, &putExtra)
	return ret, err
}
