package models

import "github.com/jinzhu/gorm"

type Robot struct {
	gorm.Model
	SerialNo      string `gorm:"size:100;unique_index" json:"vcSerialNo"`
	ChatRoomCount int64  `gorm:"bigint(20)" json:"vcMerchantNo"`
	NickName      string `gorm:"size:255;index:idx_nickname" json:"vcNickName"`
	HeadImages    string `gorm:"size:1000" json:"vcHeadImages"`
	CodeImages    string `gorm:"size:1000" json:"vcCodeImages"`
	Status        int32  `gorm:"index:idx_status" json:"nStatus"`
}
