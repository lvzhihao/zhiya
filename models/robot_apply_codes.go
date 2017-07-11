package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type RobotApplyCode struct {
	gorm.Model
	RobotSerialNo    string    `gorm:"size:100;index:idx_robot_serial_no" json:"robot_serial_no"`
	RobotNickName    string    `gorm:"size:200" json:"robot_nick_name"`
	Type             string    `gorm:"size:20" json:"type"`
	ChatRoomSerialNo string    `gorm:"size:100" json:"chat_room_serial_no"`
	ExpireTime       time.Time `json:"expire_time"`
	CodeSerialNo     string    `gorm:"size:100" json:"code_serial_no"`
	CodeValue        string    `gorm:"size:20;index:idx_code_value" json:"code_value"`
	CodeImages       string    `gorm:"size:500" json:"code_images"`
	CodeTime         time.Time `json:"code_time"`
	MyId             string    `gorm:"size:100;index:idx_my_id" json:"my_id"`
	SubId            string    `gorm:"size:100;index:idx_my_id" json:"sub_id"`
	Used             bool      `json:"used"`
	UsedTime         time.Time `json:"used_time"`
}

func FindVaildApplyCodeByMyId(db *gorm.DB, myId, subId string) (list []RobotApplyCode, err error) {
	err = db.Where("expire_time >= ?", time.Now()).Where("my_id = ?", myId).Where("sub_id = ?", subId).Order("expire_time desc").Find(&list).Error
	return
}
