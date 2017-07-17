package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// 开群验证码
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

/*
  根据供应商查找当前可用的验证码
*/
func FindVaildApplyCodeByMyId(db *gorm.DB, myId, subId string) (list []RobotApplyCode, err error) {
	err = db.Where("expire_time >= ?", time.Now()).Where("my_id = ?", myId).Where("sub_id = ?", subId).Where("used = ?", 0).Order("expire_time desc").Find(&list).Error
	return
}

/*
  使用验证码
*/
func ApplyCodeUsed(db *gorm.DB, codeSerialNo string) (code RobotApplyCode, err error) {
	err = db.Where("code_serial_no = ?", codeSerialNo).Find(&code).Error
	if err != nil {
		return
	}
	err = db.Model(&code).
		Updates(map[string]interface{}{
			"used":      true,
			"used_time": time.Now(),
		}).Error
	return
}
