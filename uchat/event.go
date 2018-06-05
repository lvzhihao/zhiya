package uchat

import (
	"encoding/json"
	"fmt"
	"time"

	rmqtool "github.com/lvzhihao/go-rmqtool"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/zhiya/models"
	"github.com/streadway/amqp"
)

func AllowKeysInterfaceMarshal(input interface{}, keys []string) ([]byte, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return b, err
	}
	var rst map[string]interface{}
	err = json.Unmarshal(b, &rst)
	if err != nil {
		return b, err
	}
	for k, _ := range rst {
		if goutils.InStringSlice(keys, k) == false {
			delete(rst, k)
		}
	}
	return json.Marshal(rst)
}

func EventChatRoomCreate(v *models.RobotChatRoom, publisher *rmqtool.PublisherTool) error {
	b, err := AllowKeysInterfaceMarshal(v, []string{
		"robot_serial_no",     // 设备号
		"chat_room_serial_no", // 群号
		"wx_user_serial_no",   // 开群人
		"my_id",               // 主帐户
		"sub_id",              // 子帐户
	})
	if err == nil {
		ext := fmt.Sprintf(".chat.create.%s.%s.%s", v.RobotSerialNo, v.ChatRoomSerialNo, v.WxUserSerialNo)
		publisher.PublishExt("event", ext, amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			ContentType:  "application/json",
			Body:         b,
		})
	}
	return err
}
