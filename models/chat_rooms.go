package models

import "github.com/jinzhu/gorm"

type ChatRoom struct {
	gorm.Model
	ChatRoomSerialNo string `gorm:"size:100;unique_index" json:"chat_room_serial_no"` //群编号
	WxUserSerialNo   string `gorm:"size:100" json:"wx_user_serial_no"`                //群主编号
	Name             string `gorm:"size:255" json:"name"`                             //群名称
	Base64Name       string `gorm:"size:500" json:"base64_name"`                      //群名称
}

func (c *ChatRoom) Ensure(db *gorm.DB, chatRoomSerialNo string) error {
	return db.Where(ChatRoom{ChatRoomSerialNo: chatRoomSerialNo}).FirstOrInit(c).Error
}

type RobotChatRoom struct {
	gorm.Model
	RobotSerialNo    string `gorm:"size:100" json:"robot_serial_no"` //robotNo chatNo index
	ChatRoomSerialNo string `gorm:"size:100" json:"chat_room_serial_no"`
	IsOpen           bool   `gorm:"index:idx_is_open" json:"is_open"`
	MyId             string `gorm:"size:100;index:idx_my_id" json:"my_id"`
	SubId            string `gorm:"size:100;index:idx_sub_id" json:"sub_id"`
	TagId            string `gorm:"size:100;index:idx_tag_id" json:"tag_id"`
}

func (c *RobotChatRoom) Ensure(db *gorm.DB, robotSerialNo, chatRoomSerialNo string) error {
	return db.Where(RobotChatRoom{RobotSerialNo: robotSerialNo, ChatRoomSerialNo: chatRoomSerialNo}).FirstOrInit(c).Error
}

func (c *RobotChatRoom) Open(db *gorm.DB) error {
	return db.Model(c).Where("robot_serial_no = ?", c.RobotSerialNo).Where("chat_room_serial_no = ?", c.ChatRoomSerialNo).Update("is_open", true).Error
}

func (c *RobotChatRoom) Close(db *gorm.DB) error {
	return db.Model(c).Where("robot_serial_no = ?", c.RobotSerialNo).Where("chat_room_serial_no = ?", c.ChatRoomSerialNo).Update("is_open", false).Error
}
