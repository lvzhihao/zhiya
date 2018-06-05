package models

import (
	"errors"
	"time"

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
	ChatRoomSerialNo  string    `gorm:"size:100;unique_index" json:"chat_room_serial_no"` //群编号
	WxUserSerialNo    string    `gorm:"size:100" json:"wx_user_serial_no"`                //群主编号
	Name              string    `gorm:"size:255" json:"name"`                             //群名称
	Base64Name        string    `gorm:"size:500" json:"base64_name"`                      //群名称
	Status            int32     `gorm:"index:idx_status" json:"status"`                   //群状态 10:开启 11:注销
	RobotInStatus     int32     `json:"robot_in_status"`                                  //机器人是否在群内 0:在群内 1:不在群内
	RobotSerialNo     string    `gorm:"size:100" json:"robot_serial_no"`                  //设备号
	RobotStatus       int32     `gorm:"index:idx_robot_status" json:"robot_status"`       //设备状态
	QrCode            string    `gorm:"size:500" json:"qr_code"`                          //群二维码
	QrCodeExpiredDate time.Time `json:"qr_code_expired_date"`                             //群二维码失效时间
	ChatRoomId        string    `gorm:"size:100" json:"chat_room_id"`                     //微信原生ID
	MemberCount       int32     `gorm:"default:0" json:"member_count"`                    //群有效人数
}

/*
  获取群记录
  如没有则初始化
*/
func (c *ChatRoom) Ensure(db *gorm.DB, chatRoomSerialNo string) error {
	return db.Where(ChatRoom{ChatRoomSerialNo: chatRoomSerialNo}).FirstOrInit(c).Error
}

/*
  群有效人数
*/
func (c *ChatRoom) ApplyMemberCount(db *gorm.DB) int32 {
	var count int32
	err := db.Model(&ChatRoomMember{}).Where("chat_room_serial_no = ?", c.ChatRoomSerialNo).Where("is_active = ?", true).Count(&count).Error
	if err == nil {
		c.MemberCount = count
		db.Save(c)
	}
	return count
}

/*
  按群编号查找已开通的设备开群记录
*/
func FindChatRoom(db *gorm.DB, chatRoomSerialNo string) (chatRoom ChatRoom, err error) {
	err = db.Where("chat_room_serial_no = ?", chatRoomSerialNo).First(&chatRoom).Error
	if chatRoom.ID == 0 {
		err = errors.New("no found")
	}
	return
}

func ApplyChatRoomMemberCount(db *gorm.DB, chatRoomSerialNo string) (int32, error) {
	var chatRoom ChatRoom
	err := db.Where("chat_room_serial_no = ?", chatRoomSerialNo).First(&chatRoom).Error
	if err != nil {
		return 0, err
	}
	return chatRoom.ApplyMemberCount(db), nil
}

// 设备开群记录
type RobotChatRoom struct {
	gorm.Model
	ExpiredDate      time.Time `json:"expired_date"`                                          //失效时间, 微选社群增加的，暂时还没有实际用处，todo
	RobotSerialNo    string    `gorm:"size:100" json:"robot_serial_no"`                       //设备号
	ChatRoomSerialNo string    `gorm:"size:100" json:"chat_room_serial_no"`                   //群编号
	WxUserSerialNo   string    `gorm:"size:100" json:"wx_user_serial_no"`                     //开群用户号
	IsOpen           bool      `gorm:"index:idx_is_open" json:"is_open"`                      //是否开启
	MyId             string    `gorm:"size:100;index:idx_my_id" json:"my_id"`                 //供应商ID
	SubId            string    `gorm:"size:100;index:idx_sub_id" json:"sub_id"`               //代理商ID
	TagId            string    `gorm:"size:100;index:idx_tag_id" json:"tag_id"`               //分组TAG
	OpenTuling       bool      `gorm:"index:idx_open_tuling" json:"open_tuling"`              //是否启用图片机器人
	OpenSignin       bool      `gorm:"default:true;index:idx_open_signin" json:"open_singin"` //是否启用签到功能
	WorkTempalteId   string    `gorm:"size:50;not null" json:"work_template_id"`              //指令模板ID
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
