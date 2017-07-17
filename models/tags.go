package models

import "github.com/jinzhu/gorm"

// 群TAG分组
type ChatRoomTag struct {
	gorm.Model
	TagId    string `gorm:"size:100" json:"tag_id"`
	TagName  string `gorm:"size:200" json:"tag_name"`
	MyId     string `gorm:"size:100" json:"my_id"`
	Count    int32  `json:"count"`
	IsActive bool   `gorm:"index:idx_is_active" json:"is_active"`
}
