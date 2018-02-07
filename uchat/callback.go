package uchat

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/uchatlib"
	"github.com/lvzhihao/zhiya/models"
)

type ChatRoomMembersList struct {
	ChatRoomUserData []map[string]string
}

/*
  群会员信息回调
  支持重复调用
*/
func SyncChatRoomMembersCallback(b []byte, db *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	_, ok := rst["vcChatRoomSerialNo"]
	if !ok {
		return errors.New("empty vcChatRoomSerialNo")
	}
	chatRoomSerialNo := goutils.ToString(rst["vcChatRoomSerialNo"])
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	var list ChatRoomMembersList
	err = json.Unmarshal([]byte(strings.TrimRight(strings.TrimLeft(goutils.ToString(data), "["), "]")), &list)
	if err != nil {
		return err
	}
	var members []models.ChatRoomMember
	err = db.Where("chat_room_serial_no = ?", chatRoomSerialNo).Where("is_active = ?", true).Find(&members).Error
	if err != nil {
		return err
	}
	userSerialNos := make([]string, 0)
	loc, _ := time.LoadLocation("Asia/Shanghai")
	for _, v := range list.ChatRoomUserData {
		// ensure chatroom
		member := models.ChatRoomMember{}
		err := member.Ensure(db, chatRoomSerialNo, goutils.ToString(v["vcSerialNo"]))
		if err != nil {
			return err
		}
		member.WxId = goutils.ToString(v["vcWeixinId"])
		//member.NickName = goutils.ToString(v["vcNickName"])
		nickNameB, _ := base64.StdEncoding.DecodeString(goutils.ToString(v["vcBase64NickName"]))
		member.NickName = goutils.ToString(nickNameB)
		member.Base64NickName = goutils.ToString(v["vcBase64NickName"])
		member.HeadImages = goutils.ToString(v["vcHeadImages"])
		member.JoinChatRoomType = goutils.ToInt32(v["nJoinChatRoomType"])
		member.FatherWxUserSerialNo = goutils.ToString(v["vcFatherWxUserSerialNo"])
		//member.MsgCount = goutils.ToInt32(v["nMsgCount"])
		//member.LastMsgDate, _ = time.ParseInLocation("2006/1/2 15:04:05", goutils.ToString(v["dtLastMsgDate"]), loc)
		member.IsActive = true
		member.JoinDate, _ = time.ParseInLocation("2006/1/2 15:04:05", goutils.ToString(v["dtCreateDate"]), loc)
		err = db.Save(&member).Error
		if err != nil {
			return err
		}
		// last status for this chatroome
		userSerialNos = append(userSerialNos, member.WxUserSerialNo)
	}
	//set close history
	for _, member := range members {
		if goutils.InStringSlice(userSerialNos, member.WxUserSerialNo) == false {
			err := member.Unactive(db)
			if err != nil {
				return err
			}
		}
	}
	// chatroom update
	var chatRoom models.ChatRoom
	err = db.Where("chat_room_serial_no = ?", goutils.ToString(chatRoomSerialNo)).First(&chatRoom).Error
	if err == nil && chatRoom.ID > 0 {
		if chatRoomId, ok := rst["vcChatRoomId"]; ok {
			chatRoom.ChatRoomId = goutils.ToString(chatRoomId)
			db.Save(&chatRoom)
		} // update chatroomId
		chatRoom.ApplyMemberCount(db) // update memberCount
	}
	return nil
}

func SyncChatQrCodeCallback(b []byte, db *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	chatRoomSerialNo, ok := rst["vcChatRoomSerialNo"]
	if !ok {
		return errors.New("empty vcChatRoomSerialNo")
	}
	var chatRoom models.ChatRoom
	err = db.Where("chat_room_serial_no = ?", goutils.ToString(chatRoomSerialNo)).First(&chatRoom).Error
	if err != nil {
		return err
	}
	//log.Fatal(chatRoom, "\n", rst)
	loc, _ := time.LoadLocation("Asia/Shanghai")
	chatRoom.QrCode = goutils.ToString(rst["vcChatRoomQRCode"])
	chatRoom.QrCodeExpiredDate, _ = time.ParseInLocation("2006-01-02 15:04:05", goutils.ToString(rst["dtExpireDateTime"]), loc)
	return db.Save(&chatRoom).Error
}

func SyncRobotChatJoinCallback(b []byte, db *gorm.DB) error {
	list, err := uchatlib.ConverUchatRobotChatJoin(b)
	if err != nil {
		return err
	}
	tx := db.Begin()
	for _, v := range list {
		var robotChatJoin models.RobotJoin
		var myRobot models.MyRobot
		err := db.Where("robot_serial_no = ?", v.RobotSerialNo).Where("expire_time > ?", time.Now().Unix()).First(&myRobot).Error
		if err != nil {
			robotChatJoin.MyId = ""
		} else {
			robotChatJoin.MyId = goutils.ToString(myRobot.MyId)
		}
		robotChatJoin.LogSerialNo = v.LogSerialNo
		robotChatJoin.RobotSerialNo = v.RobotSerialNo
		robotChatJoin.ChatRoomSerialNo = v.ChatRoomSerialNo
		robotChatJoin.ChatRoomNickName = v.ChatRoomNickName
		robotChatJoin.ChatRoomBase64NickName = v.ChatRoomBase64NickName
		robotChatJoin.WxUserSerialNo = v.WxUserSerialNo
		robotChatJoin.WxUserNickName = v.WxUserNickName
		robotChatJoin.WxUserBase64NickName = v.WxUserBase64NickName
		robotChatJoin.WxUserHeadImgUrl = v.WxUserHeadImgUrl
		robotChatJoin.JoinDate = v.JoinDate
		robotChatJoin.Status = 0
		err = tx.Create(&robotChatJoin).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
