package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// 开群验证码
type RobotApplyCode struct {
	gorm.Model
	RobotSerialNo    string    `gorm:"size:100;index:idx_robot_serial_no" json:"robot_serial_no"` //设备编号
	RobotNickName    string    `gorm:"size:200" json:"robot_nick_name"`                           //设备昵称
	Type             string    `gorm:"size:20" json:"type"`                                       //类型：开群验证码，冗余
	ChatRoomSerialNo string    `gorm:"size:100" json:"chat_room_serial_no"`                       //群编号
	ExpireTime       time.Time `json:"expire_time"`                                               //过期时间
	CodeSerialNo     string    `gorm:"size:100" json:"code_serial_no"`                            //验证码推送编号
	CodeValue        string    `gorm:"size:20;index:idx_code_value" json:"code_value"`            //验证码内容
	CodeImages       string    `gorm:"size:500" json:"code_images"`                               //验证码二维码
	CodeTime         time.Time `json:"code_time"`                                                 //验证码发送时间
	MyId             string    `gorm:"size:100;index:idx_my_id" json:"my_id"`                     //申请验证码，供应商
	SubId            string    `gorm:"size:100;index:idx_my_id" json:"sub_id"`                    //申请验证码，代理商
	Used             bool      `json:"used"`                                                      //是否使用
	UsedTime         time.Time `json:"used_time"`                                                 //使用时间
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
