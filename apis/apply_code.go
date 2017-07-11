package apis

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/lvzhihao/uchat/models"
	"github.com/lvzhihao/uchat/uchat"
)

var (
	DB                        *gorm.DB                  //数据库
	Logger                    *zap.Logger               //日志
	Client                    *uchat.UchatClient        //uchat客户端
	DefaultApplyCodeAddMinute string             = "10" //默认验证码有效时间
)

func init() {
	//todo
}

// 获取验证码
func FetchApplyCode(db *gorm.DB, myId, subId string) (*models.RobotApplyCode, error) {
	// myId 不得为空
	if myId == "" {
		return nil, errors.New("error my_id")
	}
	// 获取已经存在的验证码，如果有则直接返回
	exists, err := models.FindVaildApplyCodeByMyId(DB, myId, subId)
	if err == nil && len(exists) > 0 {
		return &exists[0], nil
	}
	// todo check subid chatroom limit

	// 查找此用户可用机器人
	robots, err := models.FindValidRobotByMyId(DB, myId)
	if err != nil {
		return nil, err
	}

	//随机选取一个机器人
	numn := rand.Intn(len(robots) - 1)
	params := map[string]string{
		"vcRobotSerialNo":    robots[numn].SerialNo, //todo
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

// RUSTAPI ApplyCode
// method POST
// Params my_id 供应商ID
// Params sub_id 商户ID
func ApplyCode(ctx echo.Context) error {
	myId := ctx.FormValue("my_id")
	subId := ctx.FormValue("sub_id")

	applyCode, err := FetchApplyCode(DB, myId, subId)
	if err != nil {
		return ctx.HTML(http.StatusOK, "Error"+err.Error())
	} else {
		return ctx.JSON(http.StatusOK, applyCode)
	}
}
