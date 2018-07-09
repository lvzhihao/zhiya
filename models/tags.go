package models

import "github.com/jinzhu/gorm"

// 群TAG分组
type ChatRoomTag struct {
	gorm.Model
	TagId    string `gorm:"size:100" json:"tag_id"`               //TagId
	TagName  string `gorm:"size:200" json:"tag_name"`             //名称
	MyId     string `gorm:"size:100" json:"my_id"`                //供应商ID
	Count    int32  `json:"count"`                                //个数
	IsActive bool   `gorm:"index:idx_is_active" json:"is_active"` //是否启用
}

func GetChatRoomTagByMyId(db *gorm.DB, my_id string) (rst []ChatRoomTag, err error) {
	rst = make([]ChatRoomTag, 0)
	err = db.Where("my_id = ?", my_id).Where("is_active = ?", true).Find(&rst).Error
	return
}
