package uchat

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/uchat/models"
)

// support retry
func SyncRobots(client *UchatClient, db *gorm.DB) error {
	rst, err := client.RobotList()
	if err != nil {
		return err
	}
	for _, v := range rst {
		// ensure robot
		robot := models.Robot{
			SerialNo:       v["vcSerialNo"],
			ChatRoomCount:  goutils.ToInt(v["nChatRoomCount"]),
			NickName:       v["vcNickName"],
			Base64NickName: v["vcBase64NickName"],
			HeadImages:     v["vcHeadImages"],
			CodeImages:     v["vcCodeImages"],
			Status:         int32(goutils.ToInt(v["nStatus"])),
		}
		err := robot.Upsert(db)
		if err != nil {
			return err
		}
	}
	return nil
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
	chatRoomSerialNo, ok := rst["vcChatRoomSerialNo"]
	if !ok {
		return errors.New("empty vcChatRoomSerialNo")
	}
	chatRoomSerialNo = goutils.ToString(chatRoomSerialNo)
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	var list ChatRoomMembersList
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	log.Println(list)
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
		room := models.ChatRoom{
			ChatRoomSerialNo: v["vcChatRoomSerialNo"],
			WxUserSerialNo:   v["vcWxUserSerialNo"],
			Name:             v["vcName"],
			Base64Name:       v["vcBase64Name"],
		}
		err := room.Upsert(db)
		if err != nil {
			return err
		}
		// ensure robot chat link
		robotRoom := models.RobotChatRoom{
			RobotSerialNo:    RobotSerialNo,
			ChatRoomSerialNo: v["vcChatRoomSerialNo"],
			IsOpen:           true,
		}
		err = robotRoom.Upsert(db)
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
