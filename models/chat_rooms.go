package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

// 代理商开群设置
type MySubChatRoomConfig struct {
	gorm.Model
	MyId  string `gorm:"size:100;index:idx_my_id" json:"my_id"`   //供应商
	SubId string `gorm:"size:100;index:idx_sub_id" json:"sub_id"` //代理商
	Num   int32  `json:"num"`                                     //开群上限
}

// 群记录
type ChatRoom struct {
	gorm.Model
	ChatRoomSerialNo string `gorm:"size:100;unique_index" json:"chat_room_serial_no"` //群编号
	WxUserSerialNo   string `gorm:"size:100" json:"wx_user_serial_no"`                //群主编号
	Name             string `gorm:"size:255" json:"name"`                             //群名称
	Base64Name       string `gorm:"size:500" json:"base64_name"`                      //群名称
}

/*
  获取群记录
  如没有则初始化
*/
func (c *ChatRoom) Ensure(db *gorm.DB, chatRoomSerialNo string) error {
	return db.Where(ChatRoom{ChatRoomSerialNo: chatRoomSerialNo}).FirstOrInit(c).Error
}

// 设备开群记录
type RobotChatRoom struct {
	gorm.Model
	RobotSerialNo    string `gorm:"size:100" json:"robot_serial_no"` //robotNo chatNo index
	ChatRoomSerialNo string `gorm:"size:100" json:"chat_room_serial_no"`
	IsOpen           bool   `gorm:"index:idx_is_open" json:"is_open"`
	MyId             string `gorm:"size:100;index:idx_my_id" json:"my_id"`
	SubId            string `gorm:"size:100;index:idx_sub_id" json:"sub_id"`
	TagId            string `gorm:"size:100;index:idx_tag_id" json:"tag_id"`
}

/*
  获取设备开群记录
  如果没有则初始化
*/
func (c *RobotChatRoom) Ensure(db *gorm.DB, robotSerialNo, chatRoomSerialNo string) error {
	return db.Where(RobotChatRoom{RobotSerialNo: robotSerialNo, ChatRoomSerialNo: chatRoomSerialNo}).FirstOrInit(c).Error
}

/*
  开群
*/
func (c *RobotChatRoom) Open(db *gorm.DB) error {
	return db.Model(c).Where("robot_serial_no = ?", c.RobotSerialNo).Where("chat_room_serial_no = ?", c.ChatRoomSerialNo).Update("is_open", true).Error
}

/*
  关群
*/
func (c *RobotChatRoom) Close(db *gorm.DB) error {
	return db.Model(c).Where("robot_serial_no = ?", c.RobotSerialNo).Where("chat_room_serial_no = ?", c.ChatRoomSerialNo).Update("is_open", false).Error
}

/*
  按群编号查找已开通的设备开群记录
*/
func FindRobotChatRoomByChatRoom(db *gorm.DB, chatRoomSerialNo string) (robotChatRoom RobotChatRoom, err error) {
	err = db.Where("chat_room_serial_no = ?", chatRoomSerialNo).Where("is_open = 1").First(&robotChatRoom).Error
	if robotChatRoom.ID == 0 {
		err = errors.New("no found")
	}
	return
}
