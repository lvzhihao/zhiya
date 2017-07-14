package uchat

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/zhiya/models"
	hashids "github.com/speps/go-hashids"
)

var DefaultMemberJoinWelcome string = ""

// support retry
func SyncRobots(client *UchatClient, db *gorm.DB) error {
	rst, err := client.RobotList()
	if err != nil {
		return err
	}
	for _, v := range rst {
		// ensure robot
		robot := models.Robot{}
		err := robot.Ensure(db, v["vcSerialNo"])
		if err != nil {
			return err
		}
		robot.ChatRoomCount = goutils.ToInt32(v["nChatRoomCount"])
		nickNameB, _ := base64.StdEncoding.DecodeString(v["vcBase64NickName"])
		robot.NickName = goutils.ToString(nickNameB)
		robot.Base64NickName = v["vcBase64NickName"]
		robot.HeadImages = v["vcHeadImages"]
		robot.CodeImages = v["vcCodeImages"]
		robot.Status = int32(goutils.ToInt(v["nStatus"]))
		err = db.Save(&robot).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func SetChatRoomOver(chatRoomSerialNo, comment string, client *UchatClient) error {
	ctx := make(map[string]string, 0)
	ctx["vcChatRoomSerialNo"] = chatRoomSerialNo
	ctx["vcComment"] = comment
	return client.ChatRoomOver(ctx)
}

func SetChatRoomOpenGetMessage(chatRoomSerialNo string, client *UchatClient) error {
	ctx := make(map[string]string, 0)
	ctx["vcChatRoomSerialNo"] = chatRoomSerialNo
	return client.ChatRoomOpenGetMessages(ctx)
}

func SetChatRoomCloseGetMessage(chatRoomSerialNo string, client *UchatClient) error {
	ctx := make(map[string]string, 0)
	ctx["vcChatRoomSerialNo"] = chatRoomSerialNo
	return client.ChatRoomCloseGetMessages(ctx)
}

// support retry
func SyncChatRoomMembers(chatRoomSerialNo string, client *UchatClient) error {
	ctx := make(map[string]string, 0)
	ctx["vcChatRoomSerialNo"] = chatRoomSerialNo
	return client.ChatRoomUserInfo(ctx)
}

type ChatRoomMembersList struct {
	ChatRoomUserData []map[string]string
}

// support retry
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
	for _, v := range list.ChatRoomUserData {
		// ensure chatroom
		member := models.ChatRoomMember{}
		err := member.Ensure(db, chatRoomSerialNo, goutils.ToString(v["vcSerialNo"]))
		if err != nil {
			return err
		}
		loc, _ := time.LoadLocation("Asia/Shanghai")
		//member.NickName = goutils.ToString(v["vcNickName"])
		nickNameB, _ := base64.StdEncoding.DecodeString(goutils.ToString(v["vcBase64NickName"]))
		member.NickName = goutils.ToString(nickNameB)
		member.Base64NickName = goutils.ToString(v["vcBase64NickName"])
		member.HeadImages = goutils.ToString(v["vcHeadImages"])
		member.JoinChatRoomType = goutils.ToInt32(v["nJoinChatRoomType"])
		member.FatherWxUserSerialNo = goutils.ToString(v["vcFatherWxUserSerialNo"])
		member.MsgCount = goutils.ToInt32(v["nMsgCount"])
		member.IsActive = true
		member.LastMsgDate, _ = time.ParseInLocation("2006/1/2 15:04:05", goutils.ToString(v["dtLastMsgDate"]), loc)
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
	return nil
}

func SyncRobotChatRoomsCallback(b []byte, db *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	_, ok := rst["vcRobotSerialNo"]
	if !ok {
		return errors.New("empty vcRobotSerialNo")
	}
	robotSerialNo := goutils.ToString(rst["vcRobotSerialNo"])
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	var robotRooms []models.RobotChatRoom
	err = db.Where("robot_serial_no = ?", robotSerialNo).Where("is_open = ?", true).Find(&robotRooms).Error
	if err != nil {
		return err
	}
	rootSerialNos := make([]string, 0)
	for _, v := range list {
		if _, ok := v["vcChatRoomSerialNo"]; ok {
			chatRoomSerialNo := goutils.ToString(v["vcChatRoomSerialNo"])
			// ensure robot chat link
			robotRoom := models.RobotChatRoom{}
			err := robotRoom.Ensure(db, robotSerialNo, chatRoomSerialNo)
			if err != nil {
				return err
			}
			robotRoom.IsOpen = true
			err = db.Save(&robotRoom).Error
			if err != nil {
				return err
			}
			// last status for this robot
			rootSerialNos = append(rootSerialNos, robotRoom.ChatRoomSerialNo)
		}
	}
	//set close history
	for _, room := range robotRooms {
		if goutils.InStringSlice(rootSerialNos, room.ChatRoomSerialNo) == false {
			err := room.Close(db)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// support retry
func SyncRobotChatRooms(RobotSerialNo string, client *UchatClient, db *gorm.DB) error {
	ctx := make(map[string]string, 0)
	ctx["vcRobotSerialNo"] = RobotSerialNo
	rst, err := client.ChatRoomList(ctx)
	if err != nil {
		return err
	}
	// fetch exists with this robot
	var robotRooms []models.RobotChatRoom
	err = db.Where("robot_serial_no = ?", RobotSerialNo).Where("is_open = ?", true).Find(&robotRooms).Error
	if err != nil {
		return err
	}
	rootSerialNos := make([]string, 0)
	for _, v := range rst {
		// ensure chatroom
		room := models.ChatRoom{}
		err := room.Ensure(db, v["vcChatRoomSerialNo"])
		if err != nil {
			return err
		}
		room.WxUserSerialNo = v["vcWxUserSerialNo"]
		nameB, _ := base64.StdEncoding.DecodeString(v["vcBase64Name"])
		room.Name = goutils.ToString(nameB)
		room.Base64Name = v["vcBase64Name"]
		err = db.Save(&room).Error
		if err != nil {
			return err
		}
		// ensure robot chat link
		robotRoom := models.RobotChatRoom{}
		err = robotRoom.Ensure(db, RobotSerialNo, v["vcChatRoomSerialNo"])
		if err != nil {
			return err
		}
		robotRoom.IsOpen = true
		err = db.Save(&robotRoom).Error
		if err != nil {
			return err
		}
		// last status for this robot
		rootSerialNos = append(rootSerialNos, robotRoom.ChatRoomSerialNo)
	}
	//set close history
	for _, room := range robotRooms {
		if goutils.InStringSlice(rootSerialNos, room.ChatRoomSerialNo) == false {
			err := room.Close(db)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func SyncMemberMessageSumCallback(b []byte, db *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	_, ok := rst["vcChatRoomSerialNo"]
	if !ok {
		return errors.New("empty ChatRoomSerialNo")
	}
	chatRoomSerialNo := goutils.ToString(rst["vcChatRoomSerialNo"])
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	for _, v := range list {
		lastMsgDate, _ := time.ParseInLocation("2006-01-02 15:04:05.999", goutils.ToString(v["dtLastMsgDate"]), loc)
		err := db.Model(&models.ChatRoomMember{}).
			Where("chat_room_serial_no = ?", chatRoomSerialNo).
			Where("wx_user_serial_no = ?", goutils.ToString(v["vcWXSerialNo"])).
			Updates(map[string]interface{}{
				"msg_count":     goutils.ToInt32(v["nMsgCount"]),
				"last_msg_date": lastMsgDate,
			}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func SyncChatRoomCreateCallback(b []byte, client *UchatClient, db *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	for _, v := range list {
		applyCode, err := models.ApplyCodeUsed(db, goutils.ToString(v["vcApplyCodeSerialNo"]))
		if err != nil {
			return err
		}
		room := models.ChatRoom{}
		chatRoomSerialNo := goutils.ToString(v["vcChatRoomSerialNo"])
		err = room.Ensure(db, chatRoomSerialNo)
		if err != nil {
			return err
		}
		room.WxUserSerialNo = goutils.ToString(v["vcWxUserSerialNo"])
		nameB, _ := base64.StdEncoding.DecodeString(goutils.ToString(v["vcBase64Name"]))
		room.Name = goutils.ToString(nameB)
		room.Base64Name = goutils.ToString(v["vcBase64Name"])
		err = db.Save(&room).Error
		if err != nil {
			return err
		}
		robotSerialNo := goutils.ToString(v["vcRobotSerialNo"])
		robotRoom := models.RobotChatRoom{}
		err = robotRoom.Ensure(db, robotSerialNo, chatRoomSerialNo)
		if err != nil {
			return err
		}
		robotRoom.MyId = applyCode.MyId
		robotRoom.SubId = applyCode.SubId
		robotRoom.IsOpen = true
		err = db.Save(&robotRoom).Error
		if err != nil {
			return err
		}
		// sync member info
		SyncChatRoomMembers(chatRoomSerialNo, client)
		// open get message
		SetChatRoomOpenGetMessage(chatRoomSerialNo, client)
	}
	return nil
}

func SyncMemberQuitCallback(b []byte, db *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	for _, v := range list {
		quitDate, _ := time.ParseInLocation("2006-01-02T15:04:05", goutils.ToString(v["dtCreateDate"]), loc)
		err := db.Model(&models.ChatRoomMember{}).
			Where("chat_room_serial_no = ?", goutils.ToString(v["vcChatRoomSerialNo"])).
			Where("wx_user_serial_no = ?", goutils.ToString(v["vcWxUserSerialNo"])).
			Updates(map[string]interface{}{
				"quit_date": quitDate,
				"is_active": false,
			}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func SyncMemberJoinCallback(b []byte, db *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	for _, v := range list {
		member := models.ChatRoomMember{}
		err := member.Ensure(db, goutils.ToString(v["vcChatRoomSerialNo"]), goutils.ToString(v["vcWxUserSerialNo"]))
		if err != nil {
			return err
		}
		//member.NickName = goutils.ToString(v["vcNickName"])
		nickNameB, _ := base64.StdEncoding.DecodeString(goutils.ToString(v["vcBase64NickName"]))
		member.NickName = goutils.ToString(nickNameB)
		member.Base64NickName = goutils.ToString(v["vcBase64NickName"])
		member.HeadImages = goutils.ToString(v["vcHeadImages"])
		member.JoinChatRoomType = goutils.ToInt32(v["nJoinChatRoomType"])
		member.FatherWxUserSerialNo = goutils.ToString(v["vcFatherWxUserSerialNo"])
		member.IsActive = true
		member.JoinDate, _ = time.ParseInLocation("2006-01-02T15:04:05.999", goutils.ToString(v["dtCreateDate"]), loc)
		err = db.Save(&member).Error
		if err != nil {
			return err
		}
		// send Message
		SendChatRoomMemberTextMessage(member.ChatRoomSerialNo, member.WxUserSerialNo, "", db)
	}
	return nil
}

func SendChatRoomMemberTextMessage(charRoomSerialNo, wxSerialNo, msg string, db *gorm.DB) error {
	if msg == "" {
		msg = FetchChatRoomMemberJoinMessage(charRoomSerialNo, db)
	}
	if msg == "" {
		return errors.New("no content")
	}
	message := &models.MessageQueue{}
	message.ChatRoomSerialNoList = charRoomSerialNo
	message.ChatRoomCount = 1
	message.MsgType = "2001"
	message.IsHit = true
	if strings.Contains(msg, "@新成员名称") {
		message.MsgContent = strings.Replace(msg, "@新成员名称", "", -1) //有问题这里
		message.WeixinSerialNo = wxSerialNo
	} else {
		message.WeixinSerialNo = ""
		message.MsgContent = msg
	}
	message.SendType = 1
	message.SendStatus = 0
	err := db.Create(message).Error
	if err != nil {
		return err
	}
	hd := hashids.NewData()
	hd.Salt = "test~~~llll"
	hd.MinLength = 16
	h := hashids.NewWithData(hd)
	message.QueueId, _ = h.Encode([]int{int(message.ID)})
	return db.Save(message).Error
}

func FetchChatRoomMemberJoinMessage(charRoomSerialNo string, db *gorm.DB) string {
	var chatRoomCmd models.ChatRoomCmd
	db.Where("chat_room_serial_no = ?", charRoomSerialNo).Where("cmd_type = ?", "member.join.welcome").Where("is_open = ?", 1).First(&chatRoomCmd)
	if chatRoomCmd.ID > 0 {
		return chatRoomCmd.CmdReply
	}
	var robotChatRoom models.RobotChatRoom
	db.Where("chat_room_serial_no = ?", charRoomSerialNo).Where("is_open = ?", 1).Order("id desc").First(&robotChatRoom)
	if robotChatRoom.ID > 0 {
		var tagCmd models.TagCmd
		db.Where("tag_id = ?", robotChatRoom.TagId).Where("cmd_type = ?", "member.join.welcome").Where("is_open = ?", 1).First(&tagCmd)
		if tagCmd.ID > 0 {
			return tagCmd.CmdReply
		}
		var myCmd models.MyCmd
		db.Where("my_id = ?", robotChatRoom.MyId).Where("cmd_type = ?", "member.join.welcome").Where("is_open = ?", 1).First(&myCmd)
		if myCmd.ID > 0 {
			return myCmd.CmdReply
		}
		//如果绑过供应商且代应商也无欢迎信息，出系统默认
		if robotChatRoom.MyId != "" {
			return DefaultMemberJoinWelcome
		}
	}
	return ""
}
