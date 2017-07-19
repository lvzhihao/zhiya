package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// 消息发送表
type MessageQueue struct {
	gorm.Model
	QueueId              string    `gorm:"size:100" json:"queue_id"`                         //发送标识
	MyId                 string    `gorm:"size:100;index:idx_my_id" json:"my_id"`            //供应商ID
	SubId                string    `gorm:"size:100;index:idx_sub_id" json:"sub_id"`          //代理商ID
	ChatRoomSerialNoList string    `gorm:"type:text(10000)" json:"chat_room_serial_no_list"` //推送群信息
	ChatRoomCount        int32     `json:"chat_room_count"`                                  //推送群总数
	IsHit                bool      `json:"is_hit"`                                           //是否@用户
	WeixinSerialNo       string    `gorm:"size:500" json:"weixin_serial_no"`                 //@用户的编号
	MsgType              string    `gorm:"size:20;index:idx_msg_type" json:"msg_type"`       //消息类型
	MsgContent           string    `gorm:"type:text(10000)" json:"msg_content"`              //消息内容
	Title                string    `gorm:"size:200" json:"title"`                            //卡片标题
	Description          string    `gorm:"size:500" json:"description"`                      //卡片介绍
	Href                 string    `gorm:"size:500" json:"href"`                             //卡片链接
	VoiceTime            int32     `json:"voice_time"`                                       //音频时间
	ProductId            string    `gorm:"size:100" json:"product_id"`                       //商品ID
	SendType             int8      `gorm:"type:tinyint(8)" json:"send_type"`                 //发送类型
	SendTime             time.Time `json:"send_time"`                                        //发送时间
	SendStatus           int8      `gorm:"type:tinyint(8)" json:"send_status"`               //发送状态
	SendStatusTime       time.Time `json:"send_status_time"`                                 //发送状态更新时间
}
