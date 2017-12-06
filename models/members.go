package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// 群成员
type ChatRoomMember struct {
	gorm.Model
	ChatRoomSerialNo     string    `gorm:"size:100" json:"chat_room_serial_no"`                       //群编号
	WxId                 string    `gorm:"size:100" json:"wx_id"`                                     //微信原生ID号
	WxUserSerialNo       string    `gorm:"size:100" json:"wx_user_serial_no"`                         //会员编号
	NickName             string    `gorm:"size:255" json:"nick_name"`                                 //昵称
	Base64NickName       string    `gorm:"size:500" json:"base64_nick_name"`                          //昵称Base64
	HeadImages           string    `gorm:"size:500" json:"head_images"`                               //会员头像
	JoinChatRoomType     int32     `gorm:"index:idx_join_chat_room_type" json:"join_chat_roome_type"` //入群方式
	FatherWxUserSerialNo string    `gorm:"size:100" json:"father_wx_user_serial_no"`                  //推荐人
	MsgCount             int32     `gorm:"index:idx_message_count" json:"msg_count"`                  //发言总数
	LastMsgDate          time.Time `json:"last_msg_date"`                                             //最后发言时间
	JoinDate             time.Time `json:"join_date"`                                                 //入群时间
	QuitDate             time.Time `json:"quit_date"`                                                 //退群时间
	IsActive             bool      `gorm:"index:idx_is_active" json:"is_active"`                      //是否活跃
}

/*
  获取群成员信息
  如果没有则初始化
*/
func (c *ChatRoomMember) Ensure(db *gorm.DB, chatRoomSerialNo, wxUserSerialNo string) error {
	return db.Where(ChatRoomMember{ChatRoomSerialNo: chatRoomSerialNo, WxUserSerialNo: wxUserSerialNo}).FirstOrInit(c).Error
}

/*
  群成员活跃
*/
func (c *ChatRoomMember) Active(db *gorm.DB) error {
	return db.Model(c).Where("chat_room_serial_no = ?", c.ChatRoomSerialNo).Where("wx_user_serial_no = ?", c.WxUserSerialNo).Update("is_active", true).Error
}

/*
  群成员不活跃，即退群，回调
*/
func (c *ChatRoomMember) Unactive(db *gorm.DB) error {
	return db.Model(c).Where("chat_room_serial_no = ?", c.ChatRoomSerialNo).Where("wx_user_serial_no = ?", c.WxUserSerialNo).Update("is_active", false).Error
}
