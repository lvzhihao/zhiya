package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Robot struct {
	gorm.Model
	SerialNo       string `gorm:"size:100;unique_index" json:"serial_no"`
	ChatRoomCount  int32  `gorm:"index:idx_chat_room_count" json:"chat_room_count"`
	NickName       string `gorm:"size:255" json:"nick_name"`
	Base64NickName string `gorm:"size:500" json:"base64_nick_name"`
	HeadImages     string `gorm:"size:500" json:"head_images"`
	CodeImages     string `gorm:"size:500" json:"code_images"`
	Status         int32  `gorm:"index:idx_status" json:"status"`
}

func (c *Robot) Ensure(db *gorm.DB, serialNo string) error {
	return db.Where(Robot{SerialNo: serialNo}).FirstOrInit(c).Error
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
	Name             string `gorm:"size:255" json:"name"`                             //群名称
	Base64Name       string `gorm:"size:500" json:"base64_name"`                      //群名称
}

func (c *ChatRoom) Ensure(db *gorm.DB, chatRoomSerialNo string) error {
	return db.Where(ChatRoom{ChatRoomSerialNo: chatRoomSerialNo}).FirstOrInit(c).Error
}

type ChatRoomTag struct {
	gorm.Model
	TagId   string `gorm:size:100" json:"tag_id"`
	TagName string `gorm:"size:200" json:"tag_name"`
	Count   int32  `json:"count"`
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

type CmdType struct {
	gorm.Model
	TypeName string `gorm:"size:50" json:"type_name"`
	TypeFlag string `gorm:"size:100" json:"type_flag"` //定义标识，方便系统回调，命令进入回调队列
}

type ChatRoomCmd struct {
	gorm.Model
	ChatRoomSerialNo string `gorm:"size:100" json:"chat_room_serial_no"`
	CmdType          int32  `gorm:"index:idx_cmd_type" json:"cmd_type"`
	CmdValue         string `gorm:"size:100" json:"cmd_value"`
	CmdReply         string `gorm:"type:text(10000)" json:"cmd_reply"`
	IsOpen           bool   `gorm:"index:idx_is_open" json:"is_open"`
}

type MyCmd struct {
	gorm.Model
	MyId     string `gorm:"size:100" json:"my_id"`
	CmdType  int32  `gorm:"index:idx_cmd_type" json:"cmd_type"`
	CmdValue string `gorm:"size:100" json:"cmd_value"`
	CmdReply string `gorm:"type:text(10000)" json:"cmd_reply"`
	IsOpen   bool   `gorm:"index:idx_is_open" json:"is_open"`
}

type SubCmd struct {
	gorm.Model
	SubId    string `gorm:"size:100" json:"sub_id"`
	CmdType  int32  `gorm:"index:idx_cmd_type" json:"cmd_type"`
	CmdValue string `gorm:"size:100" json:"cmd_value"`
	CmdReply string `gorm:"type:text(10000)" json:"cmd_reply"`
	IsOpen   bool   `gorm:"index:idx_is_open" json:"is_open"`
}

type TagCmd struct {
	gorm.Model
	TagId    string `gorm:"size:100" json:"tag_id"`
	CmdType  int32  `gorm:"index:idx_cmd_type" json:"cmd_type"`
	CmdValue string `gorm:"size:100" json:"cmd_value"`
	CmdReply string `gorm:"type:text(10000)" json:"cmd_reply"`
	IsOpen   bool   `gorm:"index:idx_is_open" json:"is_open"`
}

type MessageQueue struct {
	gorm.Model
	QueueId              string `gorm:"size:100" json:"queue_id"`
	ChatRoomSerialNoList string `gorm:"type:text(10000)" json:"chat_room_serial_no_list"`
	ChatRoomCount        int32  `json:"chat_room_count"`
	IsHit                bool   `json:"is_hit"`
	WeixinSerialNo       string `gorm:"size:500" json:"weixin_serial_no"`
	MsgType              string `gorm:"size:20;index:idx_msg_type" json:"msg_type"`
	MsgContent           string `gorm:"type:text(10000)" json:"msg_content"`
	Title                string `gorm:"size:200" json:"title"`
	Description          string `gorm:"size:500" json:"description"`
	Href                 string `gorm:"size:500" json:"href"`
	VoiceTime            int32  `json:voice_time"`
}

type ChatRoomMember struct {
	gorm.Model
	ChatRoomSerialNo     string    `gorm:"size:100" json:"chat_room_serial_no"`
	WxUserSerialNo       string    `gorm:"size:100" json:"wx_user_serial_no"`
	NickName             string    `gorm:"size:255" json:"nick_name"`
	Base64NickName       string    `gorm:"size:500" json:"base64_nick_name"`
	HeadImages           string    `gorm:"size:500" json:"head_images"`
	JoinChatRoomType     int32     `gorm:"index:idx_join_chat_room_type" json:"join_chat_roome_type"`
	FatherWxUserSerialNo string    `gorm:"size:100" json:"father_wx_user_serial_no"`
	MsgCount             int32     `gorm:index:idx_message_count" json:"msg_count"`
	LastMsgDate          time.Time `json:"last_msg_date"`
	JoinDate             time.Time `json:"join_date"`
	IsActive             bool      `gorm:"index:idx_is_active" json:"is_active"`
}

func (c *ChatRoomMember) Ensure(db *gorm.DB, chatRoomSerialNo, wxUserSerialNo string) error {
	return db.Where(ChatRoomMember{ChatRoomSerialNo: chatRoomSerialNo, WxUserSerialNo: wxUserSerialNo}).FirstOrInit(c).Error
}

func (c *ChatRoomMember) Active(db *gorm.DB) error {
	return db.Model(c).Where("chat_room_serial_no = ?", c.ChatRoomSerialNo).Where("wx_user_serial_no = ?", c.WxUserSerialNo).Update("is_active", true).Error
}

func (c *ChatRoomMember) Unactive(db *gorm.DB) error {
	return db.Model(c).Where("chat_room_serial_no = ?", c.ChatRoomSerialNo).Where("wx_user_serial_no = ?", c.WxUserSerialNo).Update("is_active", false).Error
}
