package uchat

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/lvzhihao/goutils"
)

type ChatMessageStruct struct {
	MessageSerialNo    string
	ChatRoomSerialNo   string
	FromWxUserSerialNo string
	MsgTime            time.Time
	MsgType            int32
	Content            string
	VoiceTime          int32
	ShareTitle         string
	ShareDesc          string
	ShareUrl           string
}

// global
func FetchMapValue(m map[string]interface{}, key string) interface{} {
	if v, ok := m[key]; ok {
		return v
	} else {
		return nil
	}
}

func MessageCallbackUnmarshal(b []byte, structs *[]ChatMessageStruct) error {
	list, err := MessageCallback(b)
	if err != nil {
		return err
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	for _, v := range list {
		no := FetchMapValue(v, "vcSerialNo")
		if no == nil {
			return errors.New("vcSerialNo is nil, please check origin data")
		}
		s := ChatMessageStruct{}
		s.MessageSerialNo = goutils.ToString(no)
		s.ChatRoomSerialNo = goutils.ToString(FetchMapValue(v, "vcChatRoomSerialNo"))
		s.FromWxUserSerialNo = goutils.ToString(FetchMapValue(v, "vcFromWxUserSerialNo"))
		s.MsgTime, _ = time.ParseInLocation("2006-01-02 15:04:05", goutils.ToString(v["dtMsgTime"]), loc)
		s.MsgType = goutils.ToInt32(FetchMapValue(v, "nMsgType"))
		s.Content = goutils.ToString(FetchMapValue(v, "vcContent"))
		s.VoiceTime = goutils.ToInt32(FetchMapValue(v, "nVoiceTime"))
		s.ShareTitle = goutils.ToString(FetchMapValue(v, "vcShareTitle"))
		s.ShareDesc = goutils.ToString(FetchMapValue(v, "vcShareDesc"))
		s.ShareUrl = goutils.ToString(FetchMapValue(v, "vcShareUrl"))
		*structs = append(*structs, s)
	}
	return nil
}

func MessageCallback(b []byte) ([]map[string]interface{}, error) {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return nil, err
	}
	data, ok := rst["Data"]
	if !ok {
		return nil, errors.New("empty Data")
	}
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	return list, err
}
