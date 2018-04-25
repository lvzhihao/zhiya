package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

// 设备列表
type Robot struct {
	gorm.Model
	SerialNo       string `gorm:"size:100;unique_index" json:"serial_no"`           //设备编号
	WxId           string `gorm:"size:200" json:"wx_id"`                            //微信号
	ChatRoomCount  int32  `gorm:"index:idx_chat_room_count" json:"chat_room_count"` //设备开群个数
	NickName       string `gorm:"size:255" json:"nick_name"`                        //设备昵称
	Base64NickName string `gorm:"size:500" json:"base64_nick_name"`                 //设备昵称Base64
	HeadImages     string `gorm:"size:500" json:"head_images"`                      //设备头像
	CodeImages     string `gorm:"size:500" json:"code_images"`                      //设备二维码
	Status         int32  `gorm:"index:idx_status" json:"status"`                   //设备状态
	Used           bool   `gorm:"default:true" json:"used"`                         //商户下是否使用此设备
}

/*
  获取设备信息
  如果没有则初始化
*/
func (c *Robot) Ensure(db *gorm.DB, serialNo string) error {
	return db.Where(Robot{SerialNo: serialNo}).FirstOrInit(c).Error
}

// 设备好友
type RobotFriend struct {
	gorm.Model
	RobotSerialNo  string    `gorm:"size:100" json:"robot_serial_no"`   //设备编号
	WxUserSerialNo string    `gorm:"size:100" json:"wx_user_serial_no"` //会员编号
	NickName       string    `gorm:"size:255" json:"nick_name"`         //设备昵称
	Base64NickName string    `gorm:"size:500" json:"base64_nick_name"`  //设备昵称Base64
	HeadImages     string    `gorm:"size:500" json:"head_images"`       //设备头像
	AddDate        time.Time `json:"add_date"`                          //添加时间
}

/*
  供应商设备列表
*/
type MyRobot struct {
	gorm.Model
	RobotSerialNo string    `gorm:"size:100;unique_index" json:"robot_serial_no"` //Uchat设备号，一个设备号只能绑定一个第三方用户
	MyId          string    `gorm:"size:100;index:idx_my_id" json:"my_id"`        //第三方绑定用户标识
	SubId         string    `gorm:"size:100;index:idx_sub_id" json:"sub_id"`      //子商户标识，如果存在
	ExpireTime    time.Time `json:"expire_time"`                                  //拥有设备过期时间
}

/*
  供应设备开通日志
*/
type MyRobotRenew struct {
	gorm.Model
	Platform      string    `gorm:"size:50;not null;unique_index:uix_platform_payment_id" json:"platform"`    //支付平台，比如社群后台，并非真实支付端
	PaymentId     string    `gorm:"size:100;not null;unique_index:uix_platform_payment_id" json:"payment_id"` //支持单ID，支付平台&支付单ID唯一
	RenewDays     int32     `json:"renew_days"`                                                               //续费天数
	ExpireDate    time.Time `json:"expire_date"`                                                              //过期时间
	RobotSerialNo string    `gorm:"size:100" json:"robot_serial_no"`                                          //如果指定设备号，则先检测此设备是否为MyId或subId所有，如果不是则不能续费，如果为空则新开设备
	MyId          string    `gorm:"size:100" json:"my_id"`                                                    //第三方绑定用户标识
	SubId         string    `gorm:"size:100" json:"sub_id"`                                                   //子商户标识，如果存在
	Result        string    `gorm:"size:500" json:"result"`                                                   //处理结果
}

/*
  获取商家机器人
*/
func FindRobotByMyId(db *gorm.DB, myId string) (list []MyRobot, err error) {
	err = db.Where("my_id = ?", myId).Find(&list).Error
	return
}

/*
  查找供应商所拥有可用的设置列表
*/
func FindValidRobotByMyId(db *gorm.DB, myId string) (list []Robot, err error) {
	var myRobots []MyRobot
	myRobots, err = FindRobotByMyId(db, myId)
	if err != nil {
		return
	}
	var robotIds []string
	for _, myRobot := range myRobots {
		robotIds = append(robotIds, myRobot.RobotSerialNo)
	}
	if len(robotIds) == 0 {
		err = errors.New("no valid robots")
		return
	}
	err = db.Where("serial_no in (?)", robotIds).Where("chat_room_count < ?", 30).Find(&list).Error
	if len(list) == 0 {
		// 如果没有空的设备，同样返回错误
		err = errors.New("no empty robots")
	}
	return
}

func FindValidCodeRobotByMyId(db *gorm.DB, myId string, limit int) (list []Robot, err error) {
	robots, err := FindValidRobotByMyId(db, myId)
	if err != nil {
		return robots, err
	}
	for _, r := range robots {
		count := 0
		//limit used / day
		db.Model(&RobotApplyCode{}).Where("robot_serial_no = ?", r.SerialNo).Where("used = ?", 1).Where("DATE(created_at) = CURDATE()").Count(&count)
		if count < limit {
			list = append(list, r)
		}
	}
	if len(list) == 0 {
		// 如果没有空的设备，同样返回错误
		err = errors.New("no valid robots")
	}
	return
}
