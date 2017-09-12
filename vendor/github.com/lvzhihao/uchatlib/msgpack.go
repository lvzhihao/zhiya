package uchatlib

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/lvzhihao/goutils"
	"github.com/vmihailenco/msgpack"
)

type UchatMessage struct {
	MerchantNo       string
	LogSerialNo      string
	ChatRoomSerialNo string
	WxUserSerialNo   string
	MsgTime          time.Time
	MsgType          int32
	Content          string
	VoiceTime        int32
	ShareTitle       string
	ShareDesc        string
	ShareUrl         string
	ExtraData        UchatMessageExtraData //补充数据，并非接口返回
}

type UchatMessageExtraData struct {
	RobotSerialNo    string
	RobotNickName    string
	ChatRoomName     string
	WxUserNickName   string
	WxUserHeadImages string
}

func FetchUchatMessageExtraData([]*UchatMessage) error {
	return nil
}

func ConvertUchatMessage(b []byte) ([]*UchatMessage, error) {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return nil, err
	}
	merchantNo, ok := rst["vcMerchantNo"]
	if !ok {
		return nil, errors.New("empty merchantNo")
	}
	data, ok := rst["Data"]
	if !ok {
		return nil, errors.New("empty Data")
	}
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return nil, err
	}
	ret := make([]*UchatMessage, 0)
	loc, _ := time.LoadLocation("Asia/Shanghai")
	for _, v := range list {
		msg := &UchatMessage{}
		msg.MerchantNo = goutils.ToString(merchantNo)
		msg.LogSerialNo = goutils.ToString(v["vcSerialNo"])
		msg.ChatRoomSerialNo = goutils.ToString(v["vcChatRoomSerialNo"])
		msg.WxUserSerialNo = goutils.ToString(v["vcFromWxUserSerialNo"])
		msg.MsgTime, _ = time.ParseInLocation("2006-01-02 15:04:05", goutils.ToString(v["dtMsgTime"]), loc)
		msg.MsgType = goutils.ToInt32(v["nMsgType"])
		content, err := base64.StdEncoding.DecodeString(goutils.ToString(v["vcContent"]))
		if err != nil {
			msg.Content = goutils.ToString(v["vcContent"])
		} else {
			msg.Content = goutils.ToString(content)
		}
		msg.VoiceTime = goutils.ToInt32(v["nVoiceTime"])
		msg.ShareTitle = goutils.ToString(v["vcShareTitle"])
		msg.ShareDesc = goutils.ToString(v["vcShareDesc"])
		msg.ShareUrl = goutils.ToString(v["vcShareUrl"])
		ret = append(ret, msg)
	}
	return ret, nil
}

/*
 转换Uchat Message数据为Msgpack格式
*/
func ConvertUchatMessageToMsgpack(b []byte) ([][]byte, error) {
	msgs, err := ConvertUchatMessage(b)
	if err != nil {
		return nil, err
	}
	ret := make([][]byte, 0)
	for _, msg := range msgs {
		b, err := msgpack.Marshal(msg)
		if err != nil {
			return nil, err
		}
		ret = append(ret, b)
	}
	return ret, nil
}
