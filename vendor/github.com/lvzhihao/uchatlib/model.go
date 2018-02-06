package uchatlib

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/lvzhihao/goutils"
)

func GetKey(v map[string]interface{}, key string) interface{} {
	if data, ok := v[key]; ok {
		return data
	} else {
		return nil
	}
}

func GetString(v map[string]interface{}, key string) string {
	return goutils.ToString(GetKey(v, key))
}

func GetInt32(v map[string]interface{}, key string) int32 {
	return goutils.ToInt32(GetKey(v, key))
}

type UchatMessage struct {
	MerchantNo       string      // 商户号
	MsgId            string      //【微信】消息ID
	LogSerialNo      string      // [开放平台]消息SN
	RobotSerialNo    string      // 设备号
	RobotWxId        string      // 设备微信ID
	ChatRoomSerialNo string      // 群号
	ChatRoomId       string      // 群微信ID
	WxUserSerialNo   string      // 发送人号
	WeixinId         string      // 发送人微信ID
	ToWxUserSerialNo string      // 被@用户号
	ToWeixinId       string      // 被@用户微信ID
	MsgTime          time.Time   // 发送时间
	MsgType          int32       // 消息类型 2001 文字 2002 图片 2003 语音 2004 视频 2005 链接 2006 名片 2007 动态表情 2013 小程序
	Content          string      // 文字、图文链接、名片、小程序(会进行Base64编码)；图片、语音、视频、动态表情(不经过Base64编码)
	VoiceTime        int32       // 语音时长
	ShareTitle       string      // 链接标题(或小程序名称)不经过Base64编码
	ShareDesc        string      // 链接描述(或小程序icon链接)不经过Base64编码
	ShareUrl         string      // 链接URL(或小程序图片链接)不经过Base64编码
	AppId            string      // 私聊信息里有，未知
	PlatformMsgType  int32       // 区别用户发言还是机器人发言，10：普通人 12：机器人
	IsHit            int32       // 是否艾特所有人 (0 艾特群内所有人 1 艾特单个人或者不艾特人)
	ExtraData        interface{} //补充数据，并非接口返回
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
	for _, v := range list {
		msg := &UchatMessage{}
		msg.MerchantNo = goutils.ToString(merchantNo)
		msg.MsgId = GetString(v, "vcMsgId")
		msg.LogSerialNo = GetString(v, "vcSerialNo")
		msg.RobotSerialNo = GetString(v, "vcRobotSerialNo")
		msg.RobotWxId = GetString(v, "vcRobotWxId")
		msg.ChatRoomSerialNo = GetString(v, "vcChatRoomSerialNo")
		msg.ChatRoomId = GetString(v, "vcChatRoomId")
		msg.WxUserSerialNo = GetString(v, "vcFromWxUserSerialNo")
		msg.WeixinId = GetString(v, "vcFromWeixinId")
		msg.ToWxUserSerialNo = GetString(v, "vcToWxUserSerialNo")
		msg.ToWeixinId = GetString(v, "vcToWxId")
		msg.MsgTime, _ = time.ParseInLocation("2006-01-02 15:04:05", GetString(v, "dtMsgTime"), UchatTimeLocation)
		msg.MsgType = GetInt32(v, "nMsgType")
		content, err := base64.StdEncoding.DecodeString(GetString(v, "vcContent"))
		if err != nil {
			msg.Content = GetString(v, "vcContent")
		} else {
			msg.Content = goutils.ToString(content)
		}
		msg.VoiceTime = GetInt32(v, "nVoiceTime")
		msg.ShareTitle = GetString(v, "vcShareTitle")
		msg.ShareDesc = GetString(v, "vcShareDesc")
		msg.ShareUrl = GetString(v, "vcShareUrl")
		msg.AppId = GetString(v, "vcAppId")
		msg.PlatformMsgType = GetInt32(v, "nPlatformMsgType")
		msg.IsHit = GetInt32(v, "nIsHit")
		ret = append(ret, msg)
	}
	return ret, nil
}

// keyword model
type UchatKeyword struct {
	MerchantNo         string
	LogSerialNo        string
	ChatRoomSerialNo   string
	FromWxUserSerialNo string
	ToWxUserSerialNo   string
	Content            string
	ExtraData          interface{} //补充数据，并非接口返回
}

func ConvertUchatKeyword(b []byte) ([]*UchatKeyword, error) {
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
	ret := make([]*UchatKeyword, 0)
	for _, v := range list {
		key := &UchatKeyword{}
		key.MerchantNo = goutils.ToString(merchantNo)
		key.LogSerialNo = goutils.ToString(v["vcSerialNo"])
		key.ChatRoomSerialNo = goutils.ToString(v["vcChatRoomSerialNo"])
		key.FromWxUserSerialNo = goutils.ToString(v["vcFromWxUserSerialNo"])
		key.ToWxUserSerialNo = goutils.ToString(v["vcToWxUserSerialNo"])
		key.Content = goutils.ToString(v["vcContent"])
		ret = append(ret, key)
	}
	return ret, nil
}

// MemberJoin model
// todo add more chatroom info
type UchatMemberJoin struct {
	MerchantNo           string
	ChatRoomSerialNo     string
	WxUserSerialNo       string
	FatherWxUserSerialNo string
	WxId                 string
	NickName             string
	HeadImages           string
	JoinChatRoomType     int32
	JoinDate             time.Time
	ExtraData            interface{} //补充数据，并非接口返回
}

func ConvertUchatMemberJoin(b []byte) ([]*UchatMemberJoin, error) {
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
	ret := make([]*UchatMemberJoin, 0)
	for _, v := range list {
		key := &UchatMemberJoin{}
		key.MerchantNo = goutils.ToString(merchantNo)
		key.ChatRoomSerialNo = goutils.ToString(v["vcChatRoomSerialNo"])
		key.WxUserSerialNo = goutils.ToString(v["vcWxUserSerialNo"])
		key.FatherWxUserSerialNo = goutils.ToString(v["vcFatherWxUserSerialNo"])
		key.WxId = goutils.ToString(v["vcWxId"])
		nickName, err := base64.StdEncoding.DecodeString(goutils.ToString(v["vcBase64NickName"]))
		if err != nil {
			key.NickName = goutils.ToString(v["vcNickName"])
		} else {
			key.NickName = goutils.ToString(nickName)
		}
		key.HeadImages = goutils.ToString(v["vcHeadImages"])
		key.JoinChatRoomType = goutils.ToInt32(v["nJoinChatRoomType"])
		key.JoinDate, _ = time.ParseInLocation("2006-01-02T15:04:05.999", goutils.ToString(v["dtCreateDate"]), UchatTimeLocation)
		ret = append(ret, key)
	}
	return ret, nil
}

// MemberQuit model
type UchatMemberQuit struct {
	MerchantNo       string
	ChatRoomSerialNo string
	WxUserSerialNo   string
	WxId             string
	NickName         string
	QuitDate         time.Time
	ExtraData        interface{} //补充数据，并非接口返回
}

func ConvertUchatMemberQuit(b []byte) ([]*UchatMemberQuit, error) {
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
	ret := make([]*UchatMemberQuit, 0)
	for _, v := range list {
		key := &UchatMemberQuit{}
		key.MerchantNo = goutils.ToString(merchantNo)
		key.ChatRoomSerialNo = goutils.ToString(v["vcChatRoomSerialNo"])
		key.WxUserSerialNo = goutils.ToString(v["vcWxUserSerialNo"])
		key.WxId = goutils.ToString(v["vcWxId"])
		key.NickName = goutils.ToString(v["vcNickName"])
		key.QuitDate, _ = time.ParseInLocation("2006-01-02T15:04:05", goutils.ToString(v["dtCreateDate"]), UchatTimeLocation)
		ret = append(ret, key)
	}
	return ret, nil
}

// 机器人入群model
type UchatRobotChatJoin struct {
	MerchantNo             string
	LogSerialNo            string
	RobotSerialNo          string
	ChatRoomSerialNo       string
	ChatRoomNickName       string
	ChatRoomBase64NickName string
	WxUserSerialNo         string
	WxUserNickName         string
	WxUserBase64NickName   string
	WxUserHeadImgUrl       string
	JoinDate               time.Time
	ExtraData              interface{} //补充数据，并非接口返回
}

func ConverUchatRobotChatJoin(b []byte) ([]*UchatRobotChatJoin, error) {
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
	ret := make([]*UchatRobotChatJoin, 0)
	for _, v := range list {
		key := &UchatRobotChatJoin{}
		key.MerchantNo = goutils.ToString(merchantNo)
		key.LogSerialNo = GetString(v, "vcSerialNo")
		key.RobotSerialNo = GetString(v, "vcRobotSerialNo")
		key.ChatRoomSerialNo = GetString(v, "vcChatRoomSerialNo")
		key.ChatRoomNickName = GetString(v, "vcName")
		key.ChatRoomBase64NickName = GetString(v, "vcBase64Name")
		key.WxUserSerialNo = GetString(v, "vcWxUserSerialNo")
		key.WxUserNickName = GetString(v, "vcNickName")
		key.WxUserBase64NickName = GetString(v, "vcBase64NickName")
		key.WxUserHeadImgUrl = GetString(v, "vcHeadImgUrl")
		key.JoinDate, _ = time.ParseInLocation("2006-01-02 15:04:05", GetString(v, "dtCreateDate"), UchatTimeLocation)
		ret = append(ret, key)
	}
	return ret, nil
}
