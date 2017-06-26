package models

import "github.com/jinzhu/gorm"

type Robot struct {
	gorm.Model
	SerialNo       string `gorm:"size:100;unique_index" json:"serial_no"`
	ChatRoomCount  int64  `gorm:"bigint(20)" json:"chat_room_count"`
	NickName       string `gorm:"size:255;index:idx_nickname" json:"nick_name"`
	Base64NickName string `gorm:"size:500;index:idx_base64nickname" json:"base64_nick_name"`
	HeadImages     string `gorm:"size:1000" json:"head_images"`
	CodeImages     string `gorm:"size:1000" json:"code_images"`
	Status         int32  `gorm:"index:idx_status" json:"status"`
}

type MyRobot struct {
	gorm.Model
	RobotSerialNo string `gorm:"size:100;unique_index" json:"robot_serial_no"` //Uchat设备号，一个设备号只能绑定一个第三方用户
	MyId          string `gorm:"size:100;index:idx_my_id" json:"my_id"`        //第三方绑定用户标识
	SubId         string `gorm:"size:100;index:idx_sub_id" json:"sub_id"`      //子商户标识，如果存在
}

type ChatRoom struct {
	gorm.Model
	ChatRoomSerialNo string `gorm:"size:100;unique_index" json:"chat_room_serial_no"` //群编号
	WxUserSerialNo   string `gorm:"size:100" json:"wx_user_serial_no"`                //群主编号
	Name             string `gorm:"size:255;index:idx_name" json:"name"`              //群名称
	Base64Name       string `gorm:"size:500;index:idx_base64name" json:"base64_name"` //群名称
}

type RobotChatRoom struct {
	gorm.Model
	RobotSerialNo    string `gorm:"size:100" json:"serial_no"` //robotNo chatNo index
	ChatRoomSerialNo string `gorm:"size:100" json:"chat_room_serial_no"`
	IsOpen           bool   `gorm:"index:idx_is_open" json:"is_open"`
}

type ChatRoomCmd struct {
	gorm.Model
	ChatRoomSerialNo string `gorm:"size:100" json:"chat_room_serial_no"` //serialNo keyword index
	Cmd              string `gorm:"size:50" json:"cmd"`
}
