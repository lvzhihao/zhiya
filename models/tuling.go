package models

import "github.com/jinzhu/gorm"

type TulingConfig struct {
	gorm.Model
	Name               string `gorm:size:200" json:"name"`
	ApiKey             string `gorm:size:100" json:"api_key"`
	ApiSecret          string `gorm:size:100" json:"api_secret"`
	UchatRobotSerialNo string `gorm:"size:100;unique_index" json:"robot_serial_no"` //关联uchat设备
	IsOpen             bool   `json:"is_open"`
}
