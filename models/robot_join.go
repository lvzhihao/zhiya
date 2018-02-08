package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type RobotJoin struct {
	gorm.Model
	MyId                   string    `gorm:"type:varchar(100);index;not null" json:"my_id"`               //设备拥有者
	LogSerialNo            string    `gorm:"type:varchar(30);unique_index;not null" json:"log_serial_no"` //入群日志ID
	RobotSerialNo          string    `gorm:"type:varchar(100);index;not null" json:"robot_serial_no"`     //设备号
	ChatRoomSerialNo       string    `gorm:"type:varchar(100);index;not null" json:"chat_room_serial_no"` //群编号
	ChatRoomNickName       string    `gorm:"type:varchar(255)" json:"chat_room_nick_name"`                //群昵称
	ChatRoomBase64NickName string    `gorm:"type:varchar(500)" json:"chat_room_base64_nick_name"`         //群昵称Base64
	WxUserSerialNo         string    `gorm:"type:varchar(100)" json:"wx_user_serial_no"`                  //拉群用户的微信编号
	WxUserNickName         string    `gorm:"type:varchar(255)" json:"wx_user_nick_name"`                  //用户昵称
	WxUserBase64NickName   string    `gorm:"type:varchar(500)" json:"wx_user_base64_nick_name"`           //base64编码后的用户昵称
	WxUserHeadImgUrl       string    `gorm:"type:varchar(1000)" json:"wx_user_head_img_url"`              //用户头像
	JoinDate               time.Time `json:"join_date"`                                                   //入群时间
	Status                 int32     `gorm:"index;default:0" json:"status"`                               //状态: 0 未处理 1 点击开通 2 逻辑删除
}

func (c *RobotJoin) SetStatusOpen(db *gorm.DB) error {
	if c.ID > 0 && c.Status == 0 {
		c.Status = 1
		return db.Save(c).Error
	} else {
		return fmt.Errorf("no record")
	}
}

func (c *RobotJoin) SetStatusDelete(db *gorm.DB) error {
	if c.ID > 0 && c.Status == 0 {
		c.Status = 2
		return db.Save(c).Error
	} else {
		return fmt.Errorf("no record")
	}
}
