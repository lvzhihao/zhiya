package uchat

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/uchatlib"
	"github.com/lvzhihao/zhiya/chatbot"
	"github.com/lvzhihao/zhiya/models"
	"github.com/lvzhihao/zhiya/shorten"
	"github.com/lvzhihao/zhiya/tuling"
	"github.com/lvzhihao/zhiya/utils"
	hashids "github.com/speps/go-hashids"
	"github.com/spf13/viper"
)

var (
	DefaultMemberJoinWelcome      string        = ""
	DefaultMemberJoinSendInterval time.Duration = 60 * time.Second
	HashID                        *hashids.HashID
	UseWorkTemplate               bool = false
)

//初始化hashids
func InitHashIds(salt string, minLen int) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = minLen
	HashID = hashids.NewWithData(hd)
}

/*
  同步设备列表
  支持重复调用
*/
func SyncRobots(client *uchatlib.UchatClient, db *gorm.DB) error {
	rst, err := client.RobotList()
	if err != nil {
		return err
	}
	for _, v := range rst {
		// ensure robot
		robot := models.Robot{}
		err := robot.Ensure(db, v["vcSerialNo"])
		if err != nil {
			return err
		}
		robot.WxId = goutils.ToString(v["vcWxAlias"])
		robot.ChatRoomCount = goutils.ToInt32(v["nChatRoomCount"])
		nickNameB, err := base64.StdEncoding.DecodeString(v["vcBase64NickName"])
		if err != nil {
			robot.NickName = goutils.ToString(v["vcNickName"])
		} else {
			robot.NickName = goutils.ToString(nickNameB)
		}
		robot.Base64NickName = v["vcBase64NickName"]
		robot.HeadImages = v["vcHeadImages"]
		robot.CodeImages = v["vcCodeImages"]
		robot.Status = goutils.ToInt32(v["nStatus"])
		err = db.Save(&robot).Error
		log.Printf("%+v\n", robot)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
  同步群会员信息
  支持重复调用
*/
func SyncChatRoomMembers(chatRoomSerialNo string, client *uchatlib.UchatClient) error {
	ctx := make(map[string]string, 0)
	ctx["vcChatRoomSerialNo"] = chatRoomSerialNo
	return client.ChatRoomUserInfo(ctx)
}

/*
  同步设备开群信息回调
  支持重复调用
*/
func SyncRobotChatRoomsCallback(b []byte, db *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	_, ok := rst["vcRobotSerialNo"]
	if !ok {
		return errors.New("empty vcRobotSerialNo")
	}
	robotSerialNo := goutils.ToString(rst["vcRobotSerialNo"])
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	/*
		var robotRooms []models.RobotChatRoom
		err = db.Where("robot_serial_no = ?", robotSerialNo).Where("is_open = ?", true).Find(&robotRooms).Error
		if err != nil {
			return err
		}
		rootSerialNos := make([]string, 0)
	*/
	for _, v := range list {
		if _, ok := v["vcChatRoomSerialNo"]; ok {
			chatRoomSerialNo := goutils.ToString(v["vcChatRoomSerialNo"])
			// ensure robot chat link
			robotRoom := models.RobotChatRoom{}
			err := robotRoom.Ensure(db, robotSerialNo, chatRoomSerialNo)
			if err != nil {
				return err
			}
			robotRoom.IsOpen = true
			err = db.Save(&robotRoom).Error
			if err != nil {
				return err
			}
			// last status for this robot
			//rootSerialNos = append(rootSerialNos, robotRoom.ChatRoomSerialNo)
		}
	}
	//set close history
	/*
		for _, room := range robotRooms {
			if goutils.InStringSlice(rootSerialNos, room.ChatRoomSerialNo) == false {
				err := room.Close(db)
				if err != nil {
					return err
				}
			}
		}
	*/
	return nil
}

/*
	同步群状态
	支持重复调用
*/
func SyncChatRoomStatus(chatRoomSerialNo string, client *uchatlib.UchatClient, db *gorm.DB) error {
	ctx := make(map[string]string, 0)
	ctx["vcChatRoomSerialNo"] = chatRoomSerialNo
	rst, err := client.ChatRoomStatus(ctx)
	if err != nil {
		if err.Error() == "群未开通" {
			var room models.ChatRoom
			err = db.Where("chat_room_serial_no = ?", chatRoomSerialNo).First(&room).Error
			if err != nil {
				return err
			}
			var robotChatRoom models.RobotChatRoom
			err = db.Where("robot_serial_no = ?", room.RobotSerialNo).Where("chat_room_serial_no = ?", chatRoomSerialNo).First(&robotChatRoom).Error
			if err == nil {
				robotChatRoom.Close(db)
			}
			room.Status = 0
			room.RobotInStatus = 1
			room.RobotStatus = 0
			return db.Save(&room).Error
		} else {
			return err
		}
	} else {
		serialNo := goutils.ToString(rst["vcSerialNo"])
		var room models.ChatRoom
		err = db.Where("chat_room_serial_no = ?", serialNo).First(&room).Error
		if err != nil {
			return err
		}
		var robotChatRoom models.RobotChatRoom
		err = db.Where("robot_serial_no = ?", goutils.ToString(rst["vcRobotSerialNo"])).Where("chat_room_serial_no = ?", serialNo).First(&robotChatRoom).Error
		if err == nil {
			if goutils.ToInt32(rst["nStatus"]) == 10 {
				robotChatRoom.Open(db)
			} else {
				robotChatRoom.Close(db)
			}
		}
		room.Name = goutils.ToString(rst["vcChatRoomName"])
		room.Base64Name = base64.StdEncoding.EncodeToString([]byte(room.Name))
		room.Status = goutils.ToInt32(rst["nStatus"])
		room.RobotInStatus = goutils.ToInt32(rst["nRobotInStatus"])
		room.RobotSerialNo = goutils.ToString(rst["vcRobotSerialNo"])
		room.RobotStatus = goutils.ToInt32(rst["nRobotStatus"])
		return db.Save(&room).Error
	}
}

/*
  同步设备开群信息
  支持重复调用
*/
func SyncRobotChatRooms(RobotSerialNo string, client *uchatlib.UchatClient, db *gorm.DB) error {
	ctx := make(map[string]string, 0)
	ctx["vcRobotSerialNo"] = RobotSerialNo
	rst, err := client.ChatRoomList(ctx)
	if err != nil {
		return err
	}
	// fetch exists with this robot
	var robotRooms []models.RobotChatRoom
	err = db.Where("robot_serial_no = ?", RobotSerialNo).Where("is_open = ?", true).Find(&robotRooms).Error
	if err != nil {
		return err
	}
	rootSerialNos := make([]string, 0)
	for _, v := range rst {
		// ensure chatroom
		room := models.ChatRoom{}
		err := room.Ensure(db, v["vcChatRoomSerialNo"])
		if err != nil {
			return err
		}
		room.WxUserSerialNo = v["vcWxUserSerialNo"]
		nameB, _ := base64.StdEncoding.DecodeString(v["vcBase64Name"])
		room.Name = goutils.ToString(nameB)
		room.Base64Name = v["vcBase64Name"]
		err = db.Save(&room).Error
		if err != nil {
			return err
		}
		// ensure robot chat link
		robotRoom := models.RobotChatRoom{}
		err = robotRoom.Ensure(db, RobotSerialNo, v["vcChatRoomSerialNo"])
		if err != nil {
			return err
		}
		robotRoom.IsOpen = true
		err = db.Save(&robotRoom).Error
		if err != nil {
			return err
		}
		// 同步群状态
		SyncChatRoomStatus(robotRoom.ChatRoomSerialNo, client, db)
		// last status for this robot
		rootSerialNos = append(rootSerialNos, robotRoom.ChatRoomSerialNo)
	}
	//set close history
	for _, room := range robotRooms {
		if goutils.InStringSlice(rootSerialNos, room.ChatRoomSerialNo) == false {
			err := room.Close(db)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/*
  同步群成员发言数量回调
  支持重复调用
*/
func SyncMemberMessageSumCallback(b []byte, db *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	_, ok := rst["vcChatRoomSerialNo"]
	if !ok {
		return errors.New("empty ChatRoomSerialNo")
	}
	chatRoomSerialNo := goutils.ToString(rst["vcChatRoomSerialNo"])
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	for _, v := range list {
		lastMsgDate, _ := time.ParseInLocation("2006-01-02 15:04:05.999", goutils.ToString(v["dtLastMsgDate"]), loc)
		err := db.Model(&models.ChatRoomMember{}).
			Where("chat_room_serial_no = ?", chatRoomSerialNo).
			Where("wx_user_serial_no = ?", goutils.ToString(v["vcWXSerialNo"])).
			Updates(map[string]interface{}{
				//"msg_count":     goutils.ToInt32(v["nMsgCount"]),  //这个接口会主动同步，msg_count是20分钟内的消息数，所以不能update
				"msg_count":     gorm.Expr("msg_count + ?", goutils.ToInt32(v["nMsgCount"])),
				"last_msg_date": lastMsgDate,
			}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

/*
  同步开群通知回调
  支持重复调用
*/
func SyncChatRoomCreateCallback(b []byte, client *uchatlib.UchatClient, db *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	for _, v := range list {
		room := models.ChatRoom{}
		chatRoomSerialNo := goutils.ToString(v["vcChatRoomSerialNo"])
		err = room.Ensure(db, chatRoomSerialNo)
		if err != nil {
			return err
		}
		room.WxUserSerialNo = goutils.ToString(v["vcWxUserSerialNo"])
		nameB, _ := base64.StdEncoding.DecodeString(goutils.ToString(v["vcBase64Name"]))
		room.Name = goutils.ToString(nameB)
		room.Base64Name = goutils.ToString(v["vcBase64Name"])
		err = db.Save(&room).Error
		if err != nil {
			return err
		}
		robotSerialNo := goutils.ToString(v["vcRobotSerialNo"])
		robotRoom := models.RobotChatRoom{}
		err = robotRoom.Ensure(db, robotSerialNo, chatRoomSerialNo)
		if err != nil {
			return err
		}
		applyCode, applyCodeerr := models.ApplyCodeUsed(db, goutils.ToString(v["vcApplyCodeSerialNo"]))
		if applyCodeerr == nil {
			robotRoom.MyId = applyCode.MyId
			robotRoom.SubId = applyCode.SubId
		} else { //如果有applyCode记录，可确定开群申请时的供应商和店铺身份
			var robotJoin models.RobotJoin
			err := db.Where("robot_serial_no = ?", goutils.ToString(v["vcRobotSerialNo"])).
				Where("chat_room_serial_no = ?", goutils.ToString(v["vcChatRoomSerialNo"])).
				Where("wx_user_serial_no = ?", goutils.ToString(v["vcWxUserSerialNo"])).
				Where("status = ?", 1).
				Where("UNIX_TIMESTAMP(updated_at) >= ?", time.Now().Add(-60*time.Minute).Unix()).
				First(&robotJoin).Error //1小时内有相同的开通记录
			if err == nil {
				robotRoom.MyId = robotJoin.MyId
			}
			//log.Fatal(robotJoin, err)
		}
		robotRoom.IsOpen = true
		// 默认试用期
		robotRoom.ExpiredDate = time.Now().Add(7 * 24 * time.Hour)
		err = db.Save(&robotRoom).Error
		if err != nil {
			return err
		}
		// sync member info
		SyncChatRoomMembers(chatRoomSerialNo, client)
		// sync chat qr code
		uchatlib.ApplyChatRoomQrCode(chatRoomSerialNo, client)
		// open get message
		log.Println(uchatlib.SetChatRoomOpenGetMessage(chatRoomSerialNo, client))
	}
	return nil
}

/*
  同步群成员退群通知回调
  支持重复调用
*/
func SyncMemberQuitCallback(b []byte, db *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	for _, v := range list {
		quitDate, _ := time.ParseInLocation("2006-01-02T15:04:05", goutils.ToString(v["dtCreateDate"]), loc)
		err := db.Model(&models.ChatRoomMember{}).
			Where("chat_room_serial_no = ?", goutils.ToString(v["vcChatRoomSerialNo"])).
			Where("wx_user_serial_no = ?", goutils.ToString(v["vcWxUserSerialNo"])).
			Updates(map[string]interface{}{
				"quit_date": quitDate,
				"is_active": false,
			}).Error
		if err != nil {
			return err
		}
		models.ApplyChatRoomMemberCount(db, goutils.ToString(v["vcChatRoomSerialNo"]))
	}
	return nil
}

/*
  同步群成员入群通知回调
  支持重复调用
*/
func SyncMemberJoinCallback(b []byte, db *gorm.DB, managerDB *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	for _, v := range list {
		member := models.ChatRoomMember{}
		err := member.Ensure(db, goutils.ToString(v["vcChatRoomSerialNo"]), goutils.ToString(v["vcWxUserSerialNo"]))
		if err != nil {
			return err
		}
		member.WxId = goutils.ToString(v["vcWxId"])
		//member.NickName = goutils.ToString(v["vcNickName"])
		nickNameB, _ := base64.StdEncoding.DecodeString(goutils.ToString(v["vcBase64NickName"]))
		member.NickName = goutils.ToString(nickNameB)
		member.Base64NickName = goutils.ToString(v["vcBase64NickName"])
		member.HeadImages = goutils.ToString(v["vcHeadImages"])
		member.JoinChatRoomType = goutils.ToInt32(v["nJoinChatRoomType"])
		member.FatherWxUserSerialNo = goutils.ToString(v["vcFatherWxUserSerialNo"])
		member.IsActive = true
		member.JoinDate, _ = time.ParseInLocation("2006-01-02T15:04:05.999", goutils.ToString(v["dtCreateDate"]), loc)
		err = db.Save(&member).Error
		if err != nil {
			return err
		}
		// send Message
		SendChatRoomMemberTextMessage(member.ChatRoomSerialNo, member.WxUserSerialNo, "", db, managerDB)
		// member count
		models.ApplyChatRoomMemberCount(db, goutils.ToString(v["vcChatRoomSerialNo"]))
	}
	return nil
}

func FetchChatRoomMemberJoinMessageTemplate(db *gorm.DB, chatRoomSerialNo string) (string, time.Duration) {
	template, err := GetChatRoomValidTemplate(db, chatRoomSerialNo, "member.join.welcome")
	if err != nil {
		return "", DefaultMemberJoinSendInterval
	}
	var params map[string]interface{}
	err = json.Unmarshal([]byte(template.CmdParams), &params)
	var interval = DefaultMemberJoinSendInterval
	if err == nil {
		if iter, ok := params["interval"]; ok {
			interval = time.Duration(goutils.ToInt32(iter)) * time.Second
		}
	}
	return strings.TrimSpace(template.CmdReply), interval
}

/*
  发送群内信息
*/
func SendChatRoomMemberTextMessage(charRoomSerialNo, wxSerialNo, msg string, db, managerDB *gorm.DB) error {
	log.Println("commit start")
	sendInterval := DefaultMemberJoinSendInterval
	if msg == "" {
		if UseWorkTemplate == true {
			msg, sendInterval = FetchChatRoomMemberJoinMessageTemplate(db, charRoomSerialNo)
		} else {
			msg = FetchChatRoomMemberJoinMessage(charRoomSerialNo, db)
		}
	}
	if msg == "" {
		return errors.New("no join content")
	}
	log.Println("commit start")
	message := &models.MessageQueue{}
	tx := db.Begin()
	err := tx.Where("chat_room_serial_no_list = ?", charRoomSerialNo).Where("send_type = 10").Where("send_status = 0").First(&message).Error
	if err == nil {
		log.Println("update join message")
		//update
		if strings.Contains(msg, "@新成员名称") {
			message.WeixinSerialNo = message.WeixinSerialNo + "," + wxSerialNo //添加@新成员
		}
	} else {
		log.Println("create join message")
		//new message
		message.ChatRoomSerialNoList = charRoomSerialNo
		message.ChatRoomCount = 1
		message.MsgType = "2001"
		message.IsHit = true
		if strings.Contains(msg, "@新成员名称") {
			message.MsgContent = strings.Replace(msg, "@新成员名称", "", -1) //有问题这里
			message.WeixinSerialNo = wxSerialNo
		} else {
			message.WeixinSerialNo = ""
			message.MsgContent = msg
		}
		if strings.Contains(msg, "{优惠链接}") {
			url, err := GenerateTuikeasyProductCouponSearchByChatRoomSerialNo(charRoomSerialNo, db, managerDB)
			if err != nil {
				message.MsgContent = strings.Replace(message.MsgContent, "{优惠链接}", "", -1)
			} else {
				message.MsgContent = strings.Replace(message.MsgContent, "{优惠链接}", ShortUrl(url), -1)
			}
		}
		message.SendType = 10 //第一个新成员加入后，30秒内所有加入成员一起发送，类型10
		message.SendStatus = 0
		message.SendTime = time.Now().Add(sendInterval)
		err := tx.Create(message).Error
		if err != nil {
			tx.Rollback()
			return err
		}
		message.QueueId, _ = HashID.Encode([]int{int(message.ID)})
	}
	log.Println(message)
	err = tx.Save(message).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit().Error
	log.Println("commit status:", err)
	return err
}

/*
  获取入群欢迎通知
*/
func FetchChatRoomMemberJoinMessage(charRoomSerialNo string, db *gorm.DB) string {
	var chatRoomCmd models.ChatRoomCmd
	db.Where("chat_room_serial_no = ?", charRoomSerialNo).Where("cmd_type = ?", "member.join.welcome").Where("is_open = ?", 1).First(&chatRoomCmd)
	if chatRoomCmd.ID > 0 {
		return chatRoomCmd.CmdReply
	}
	var robotChatRoom models.RobotChatRoom
	db.Where("chat_room_serial_no = ?", charRoomSerialNo).Where("is_open = ?", 1).Order("id desc").First(&robotChatRoom)
	if robotChatRoom.ID > 0 {
		var tagCmd models.TagCmd
		db.Where("tag_id = ?", robotChatRoom.TagId).Where("cmd_type = ?", "member.join.welcome").Where("is_open = ?", 1).First(&tagCmd)
		if tagCmd.ID > 0 {
			return tagCmd.CmdReply
		}
		var myCmd models.MyCmd
		db.Where("my_id = ?", robotChatRoom.MyId).Where("cmd_type = ?", "member.join.welcome").Where("is_open = ?", 1).First(&myCmd)
		if myCmd.ID > 0 {
			return myCmd.CmdReply
		}
		//如果绑过供应商且代应商也无欢迎信息，出系统默认
		if robotChatRoom.MyId != "" {
			return DefaultMemberJoinWelcome
		}
	}
	return ""
}

func FetchChatRoomIntelligentChatTemplate(db *gorm.DB, chatRoomSerialNo string, msgDate time.Time) (*models.WorkTemplate, error) {
	template, err := GetChatRoomValidTemplate(db, chatRoomSerialNo, "shop.intelligent.chatting")
	if err != nil {
		return nil, err
	}
	var params map[string]interface{}
	err = json.Unmarshal([]byte(template.CmdParams), &params)
	if err != nil {
		return nil, err
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", time.Now().Format("2006-01-02 ")+goutils.ToString(params["start_time"]), loc)
	if err != nil {
		return nil, err
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", time.Now().Format("2006-01-02 ")+goutils.ToString(params["end_time"]), loc)
	if err != nil {
		return nil, err
	}
	// 如果结束时间为第二天的，则规则是大于当天起始时间或小于当天结束时间
	if int64(endTime.Sub(startTime)/time.Second) < 0 {
		if int64(msgDate.Sub(startTime)/time.Second) > 0 || int64(msgDate.Sub(endTime)/time.Second) < 0 {
			// run
		} else {
			return nil, fmt.Errorf("time closed")
		}
	} else {
		if int64(msgDate.Sub(startTime)/time.Second) > 0 && int64(msgDate.Sub(endTime)/time.Second) < 0 {
			// run
		} else {
			return nil, fmt.Errorf("time closed")
		}
	}
	return template, nil
}

func SyncChatKeywordCallback(b []byte, db *gorm.DB, managerDB *gorm.DB, tool *utils.ReceiveTool, chatBotClient *chatbot.Client) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	for _, v := range list {
		var robotChatRoom models.RobotChatRoom
		db.Where("chat_room_serial_no = ?", goutils.ToString(v["vcChatRoomSerialNo"])).Where("is_open = ?", 1).Order("id desc").First(&robotChatRoom)
		//对像是该群机器人
		if robotChatRoom.ID > 0 && strings.Compare(robotChatRoom.RobotSerialNo, goutils.ToString(v["vcToWxUserSerialNo"])) == 0 {
			if UseWorkTemplate == true {
				// 目前关键词接口还没有返回时间
				/*
					loc, _ := time.LoadLocation("Asia/Shanghai")
					msgDate, err := time.ParseInLocation("2006-01-02T15:04:05", goutils.ToString(v["dtMsgTime"]), loc)
					if err != nil {
						continue
					}
				*/
				//template, err := FetchChatRoomIntelligentChatTemplate(db, robotChatRoom.ChatRoomSerialNo, msgDate)
				template, err := FetchChatRoomIntelligentChatTemplate(db, robotChatRoom.ChatRoomSerialNo, time.Now())
				if err != nil {
					// 如果没有匹配的运营模板，则忽略，读取下一条数据
					continue
				}
				data, err := chatBotClient.SendMessage(template.WorkTemplateId, "tuling", goutils.ToString(v["vcContent"]), robotChatRoom.ChatRoomSerialNo+"-"+goutils.ToString(v["vcFromWxUserSerialNo"]))
				//log.Fatal(data, err)
				if err != nil {
					log.Println(err)
				} else {
					switch data.Code {
					case 1000:
						// 文字
						rst := make(map[string]interface{}, 0)
						rst["MerchantNo"] = viper.GetString("merchant_no")
						rst["vcRelaSerialNo"] = "chatbot-" + goutils.RandomString(20)
						rst["vcChatRoomSerialNo"] = robotChatRoom.ChatRoomSerialNo
						rst["vcRobotSerialNo"] = robotChatRoom.RobotSerialNo
						rst["nIsHit"] = "1"
						rst["vcWeixinSerialNo"] = goutils.ToString(v["vcFromWxUserSerialNo"])
						rst["Data"] = []map[string]string{
							map[string]string{
								"nMsgType":   "2001",
								"msgContent": data.Text,
								"vcTitle":    "",
								"vcDesc":     "",
								"nVoiceTime": "0",
								"vcHref":     "",
							},
						}
						b, _ := json.Marshal(rst)
						tool.Publish("uchat.mysql.message.queue", goutils.ToString(b))
					}
				}
			} else {
				// old 规则
				var tulingConfig models.TulingConfig
				db.Where("uchat_robot_serial_no = ?", robotChatRoom.RobotSerialNo).Where("is_open = ?", 1).First(&tulingConfig)
				if tulingConfig.ApiKey != "" {
					data, err := FetchTulingResult(tulingConfig.ApiKey, tulingConfig.ApiSecret, map[string]interface{}{
						"info":   goutils.ToString(v["vcContent"]),
						"userid": goutils.ToString(v["vcFromWxUserSerialNo"]),
					}, robotChatRoom.ChatRoomSerialNo, db, managerDB)
					if err != nil {
						log.Println(err)
					} else {
						rst := make(map[string]interface{}, 0)
						rst["MerchantNo"] = viper.GetString("merchant_no")
						rst["vcRelaSerialNo"] = "tuling-" + goutils.RandomString(20)
						rst["vcChatRoomSerialNo"] = robotChatRoom.ChatRoomSerialNo
						rst["vcRobotSerialNo"] = robotChatRoom.RobotSerialNo
						rst["nIsHit"] = "1"
						rst["vcWeixinSerialNo"] = goutils.ToString(v["vcFromWxUserSerialNo"])
						rst["Data"] = data
						b, _ := json.Marshal(rst)
						tool.Publish("uchat.mysql.message.queue", goutils.ToString(b))
					}
				}
			}
		}
	}
	return nil
}

func FetchTulingResult(key, secret string, context map[string]interface{}, chat_room_serial_no string, db, managerDB *gorm.DB) ([]map[string]string, error) {
	c := tuling.NewTulingClient(tuling.TulingClientConfig{
		ApiUrl:         viper.GetString("tuling_api_url"),
		ApiKey:         key,
		ApiSecret:      secret,
		DefaultTimeout: 10 * time.Second,
	})
	data, err := c.Do(context)
	if err != nil {
		return nil, err
	}
	if data.IsError() {
		return nil, data.Error()
	}
	log.Println(data)
	switch data.Code {
	case 100000:
		return []map[string]string{
			map[string]string{
				"nMsgType":   "2001",
				"msgContent": data.Text,
				"vcTitle":    "",
				"vcDesc":     "",
				"nVoiceTime": "0",
				"vcHref":     "",
			},
		}, nil
	case 200000:
		var regex = regexp.MustCompile(`^亲，已帮你找到(.*)价格信息$`)
		strs := regex.FindStringSubmatch(data.Text)
		if len(strs) == 2 {
			keyword := strs[1]
			content, err := GenerateTuikeasyProductSearchContentByKeyword(chat_room_serial_no, keyword, db, managerDB)
			if err == nil {
				return []map[string]string{
					map[string]string{
						"nMsgType":   "2001",
						"msgContent": content,
						"vcTitle":    "",
						"vcDesc":     "",
						"nVoiceTime": "0",
						"vcHref":     "",
					},
				}, nil
			}
		}
		regex = regexp.MustCompile(`^亲，已找到在(.*)的(.*)价格行情$`)
		strs = regex.FindStringSubmatch(data.Text)
		if len(strs) == 3 {
			keyword := strs[2]
			content, err := GenerateTuikeasyProductSearchContentByKeyword(chat_room_serial_no, keyword, db, managerDB)
			if err == nil {
				return []map[string]string{
					map[string]string{
						"nMsgType":   "2001",
						"msgContent": content,
						"vcTitle":    "",
						"vcDesc":     "",
						"nVoiceTime": "0",
						"vcHref":     "",
					},
				}, nil
			}
		}
		return []map[string]string{
			map[string]string{
				"nMsgType":   "2001",
				"msgContent": data.Text + " " + ShortUrl(data.Url),
				"vcTitle":    "",
				"vcDesc":     "",
				"nVoiceTime": "0",
				"vcHref":     "",
			},
		}, nil
	case 308000:
		return []map[string]string{
			map[string]string{
				"nMsgType":   "2001",
				"msgContent": data.Text,
				"vcTitle":    "",
				"vcDesc":     "",
				"nVoiceTime": "0",
				"vcHref":     "",
			},
			map[string]string{
				"nMsgType":   "2005",
				"msgContent": data.List[0].Icon,
				"vcTitle":    data.List[0].Name,
				"vcDesc":     data.List[0].Info,
				"nVoiceTime": "0",
				"vcHref":     data.List[0].DetailUrl,
			},
		}, nil
	default:
		return nil, errors.New("未知类型")
	}
}

type CustomKeywordReplyTemplateParams struct {
	Keywords  []CustomKeywordReplyTemplateParamsKeyword `json:"keywords"`
	ReplyType string                                    `json:"reply_type"`
}

type CustomKeywordReplyTemplateParamsKeyword struct {
	Filter  string `json:"filter"`
	Keyword string `json:"keyword"`
}

/*
  获取商家自定义关键词回复模板
*/
func FetchChatRoomCustomKeywordReplyTemplate(db *gorm.DB, chatRoomSerialNo, msg string, org map[string]interface{}) ([]map[string]interface{}, error) {
	if goutils.ToString(org["nPlatformMsgType"]) == "12" {
		return nil, fmt.Errorf("robot message don't reply")
	}
	template, err := GetChatRoomValidTemplate(db, chatRoomSerialNo, "shop.custom.keyword.reply")
	if err != nil {
		return nil, err
	}
	var params CustomKeywordReplyTemplateParams
	err = json.Unmarshal([]byte(template.CmdParams), &params)
	if err != nil {
		return nil, err
	}
	msg = strings.TrimSpace(msg)
	hasRegex := false
	for _, keyword := range params.Keywords {
		switch strings.ToLower(keyword.Filter) {
		case "all":
			if strings.Compare(msg, keyword.Keyword) == 0 {
				hasRegex = true
				break
			}
		case "like":
			if strings.Index(msg, keyword.Keyword) >= 0 {
				hasRegex = true
				break
			}
		default:
		}
	}
	if hasRegex == true {
		var data []map[string]interface{}
		err := json.Unmarshal([]byte(template.CmdReply), &data)
		if err != nil {
			return nil, err
		} else {
			if strings.ToLower(params.ReplyType) == "all" {
				return data, nil
			} else {
				rnd := rand.Intn(len(data))
				return []map[string]interface{}{data[rnd]}, nil
			}
		}
	} else {
		return nil, fmt.Errorf("no regex msg")
	}
}

/*
  同步群成员时时消息回调
  支持重复调用
*/
func SyncChatMessageCallback(b []byte, db *gorm.DB, managerDB *gorm.DB, tool *utils.ReceiveTool) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	data, ok := rst["Data"]
	if !ok {
		return errors.New("empty Data")
	}
	var list []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(data)), &list)
	if err != nil {
		return err
	}
	for _, v := range list {
		if goutils.ToString(v["nMsgType"]) != "2001" {
			continue
		}
		b, err := base64.StdEncoding.DecodeString(goutils.ToString(v["vcContent"]))
		if err != nil {
			return err
		}
		content := string(b)
		if content == "" {
			continue
		}
		var robotChatRoom models.RobotChatRoom
		db.Where("chat_room_serial_no = ?", goutils.ToString(v["vcChatRoomSerialNo"])).Where("is_open = ?", 1).Order("id desc").First(&robotChatRoom)
		if UseWorkTemplate == true {
			data, err := FetchChatRoomCustomKeywordReplyTemplate(db, robotChatRoom.ChatRoomSerialNo, content, v)
			if err != nil {
				// 如果没有匹配的运营模板，则忽略，读取下一条数据
				continue
			}
			// 补充商户号和发送流水号
			rst := make(map[string]interface{}, 0)
			rst["MerchantNo"] = viper.GetString("merchant_no")
			rst["vcRelaSerialNo"] = "custom-keyword-reply-" + goutils.RandomString(20)
			rst["vcChatRoomSerialNo"] = robotChatRoom.ChatRoomSerialNo
			rst["vcRobotSerialNo"] = robotChatRoom.RobotSerialNo
			rst["vcWeixinSerialNo"] = ""
			rst["nIsHit"] = "1"
			rst["Data"] = data
			b, _ := json.Marshal(rst)
			log.Printf("%s\n", b)
			tool.Publish("uchat.mysql.message.queue", goutils.ToString(b))
		} else {
			if robotChatRoom.MyId != "" { //有绑定供应商
				err := SendShopCustomSearch(robotChatRoom.MyId, robotChatRoom.SubId, robotChatRoom.ChatRoomSerialNo, content, db)
				//如果没有自定义配置
				if err != nil {
					//pid, err := FetchAlimamaSearchPid(robotChatRoom.MyId, robotChatRoom.SubId, managerDB)
					domain, err := FetchTuikeasySearchDomain(robotChatRoom.MyId, robotChatRoom.SubId, managerDB)
					if err != nil {
						log.Println(err)
						continue
					}
					//err = SendAlimamProductSearch(robotChatRoom.MyId, pid, robotChatRoom.ChatRoomSerialNo, content, db)
					err = SendTuikeasyProductSearch(robotChatRoom.MyId, domain, robotChatRoom.ChatRoomSerialNo, content, db)
					//如果已经匹配关键字则不用再去搜索优惠链接
					if err != nil {
						//err = SendAlimamCouponSearch(robotChatRoom.MyId, pid, robotChatRoom.ChatRoomSerialNo, content, db)
						err = SendTuikeasyCouponSearch(robotChatRoom.MyId, domain, robotChatRoom.ChatRoomSerialNo, content, db)
						log.Println(err)
					}
				}
			}
		}
	}
	return nil
}

func SendShopCustomSearch(myId, subId, chatRoomSerialNo, content string, db *gorm.DB) error {
	var rst struct {
		gorm.Model
		CmdValue  string
		CmdReply  string
		CmdParams string
	}
	if subId != "" {
		db.Table(viper.GetString("table_prefix")+"_sub_cmds").Where("sub_id = ?", subId).Where("cmd_type = ?", "shop.custom.search").Where("is_open = 1").First(&rst)
	}
	if rst.ID == 0 {
		db.Table(viper.GetString("table_prefix")+"_my_cmds").Where("my_id = ?", myId).Where("cmd_type = ?", "shop.custom.search").Where("is_open = 1").First(&rst)
	}
	if rst.ID > 0 && strings.Index(strings.TrimSpace(content), strings.TrimSpace(rst.CmdValue)) == 0 {
		key := strings.TrimSpace(strings.Replace(content, rst.CmdValue, "", 1))
		keyList := strings.Split(strings.TrimSpace(rst.CmdParams), "|")
		if goutils.InStringSlice(keyList, key) {
			message := &models.MessageQueue{}
			message.ChatRoomSerialNoList = chatRoomSerialNo
			message.ChatRoomCount = 1
			message.MsgType = "2001"
			message.IsHit = true
			message.WeixinSerialNo = ""
			message.MsgContent = rst.CmdReply
			message.SendType = 1
			message.SendStatus = 0
			err := db.Create(message).Error
			if err != nil {
				return err
			}
			message.QueueId, _ = HashID.Encode([]int{int(message.ID)})
			return db.Save(message).Error
		}
	}
	return errors.New("no shop.custom.search config")
}

/*
  获取淘宝联盟广告位pid
*/
func FetchTuikeasySearchDomain(myId, subId string, db *gorm.DB) (string, error) {
	if subId != "" {
		pid := ""
		row := db.Table("sdb_maifou_promotion_detail").Where("platform = ?", "tuikeasy").Where("supplier_id = ?", myId).Where("shop_id = ?", subId).Select("pid").Row()
		row.Scan(&pid)
		if pid != "" {
			//return "mmzl" + subId, nil
			return pid, nil
		}
	} else {
		pid := ""
		row := db.Table("sdb_maifou_promotion_detail").Where("platform = ?", "tuikeasy").Where("supplier_id = ?", myId).Where("is_self = ?", true).Select("pid").Row()
		row.Scan(&pid)
		if pid != "" {
			//return "mmzlb" + utils.FakeIdEncode(goutils.ToInt64(myId)), nil
			return pid, nil
		}
	}
	return "", errors.New("no pid")
}

/*
  获取淘宝联盟广告位pid
*/
func FetchAlimamaSearchPid(myId, subId string, db *gorm.DB) (string, error) {
	if subId != "" {
		pid := ""
		row := db.Table("sdb_maifou_promotion_detail").Where("platform = ?", "taoke").Where("supplier_id = ?", myId).Where("shop_id = ?", subId).Select("pid").Row()
		row.Scan(&pid)
		if pid != "" {
			return pid, nil
		}
	} else {
		pid := ""
		row := db.Table("sdb_maifou_promotion_detail").Where("platform = ?", "taoke").Where("supplier_id = ?", myId).Where("is_self = ?", true).Select("pid").Row()
		row.Scan(&pid)
		if pid != "" {
			return pid, nil
		}
	}
	return "", errors.New("no pid")
}

/*
  获取淘宝联盟产品搜索信息
*/
func SendAlimamProductSearch(myId, pid, chatRoomSerialNo, content string, db *gorm.DB) error {
	return nil
	var cmd models.MyCmd
	db.Where("my_id = ?", myId).Where("cmd_type = ?", "alimama.product.search").Where("is_open = 1").First(&cmd)
	if cmd.ID > 0 && strings.Index(strings.TrimSpace(content), strings.TrimSpace(cmd.CmdValue)) == 0 {
		key := strings.Replace(content, cmd.CmdValue, "", 1)
		if key != "" {
			params := url.Values{}
			params.Set("nav", "0")
			params.Set("p", pid)
			params.Set("searchText", strings.TrimSpace(key))
			//销商配置
			searchConfig := viper.GetStringMap("taoke_search_config")
			if config, ok := searchConfig[myId]; ok {
				var p map[string]string
				err := json.Unmarshal([]byte(goutils.ToString(config)), &p)
				if err == nil {
					for k, v := range p {
						params.Set(k, v)
					}
				}
			}
			//end
			//temp change
			var url string
			if myId == "20546" {
				url = "http://m.xuanwonainiu.com/search?" + params.Encode()
			} else {
				url = "http://m.xuanwonainiu.com/sp-search?" + params.Encode()
			}
			pubContent := strings.Replace(cmd.CmdReply, "{搜索关键词}", strings.TrimSpace(key), -1)
			pubContent = strings.Replace(pubContent, "{优惠链接}", ShortUrl(url), -1)
			message := &models.MessageQueue{}
			message.ChatRoomSerialNoList = chatRoomSerialNo
			message.ChatRoomCount = 1
			message.MsgType = "2001"
			message.IsHit = true
			message.WeixinSerialNo = ""
			message.MsgContent = pubContent
			message.SendType = 1
			message.SendStatus = 0
			err := db.Create(message).Error
			if err != nil {
				return err
			}
			message.QueueId, _ = HashID.Encode([]int{int(message.ID)})
			return db.Save(message).Error
		}
	}
	//todo log
	return errors.New("no alimama.product.search config")
}

/*
  获取淘宝联盟优惠信息
*/
func SendAlimamCouponSearch(myId, pid, chatRoomSerialNo, content string, db *gorm.DB) error {
	return nil
	var cmd models.MyCmd
	db.Where("my_id = ?", myId).Where("cmd_type = ?", "alimama.coupon.search").Where("is_open = 1").First(&cmd)
	if cmd.ID > 0 && strings.Compare(strings.TrimSpace(content), strings.TrimSpace(cmd.CmdValue)) == 0 {
		params := url.Values{}
		params.Set("nav", "0")
		params.Set("p", pid)
		url := "http://m.xuanwonainiu.com/?" + params.Encode()
		pubContent := strings.Replace(cmd.CmdReply, "{优惠链接}", ShortUrl(url), -1)
		message := &models.MessageQueue{}
		message.ChatRoomSerialNoList = chatRoomSerialNo
		message.ChatRoomCount = 1
		message.MsgType = "2001"
		message.IsHit = true
		message.WeixinSerialNo = ""
		message.MsgContent = pubContent
		message.SendType = 1
		message.SendStatus = 0
		err := db.Create(message).Error
		if err != nil {
			return err
		}
		message.QueueId, _ = HashID.Encode([]int{int(message.ID)})
		return db.Save(message).Error
	}
	//todo log
	return errors.New("no alimama.coupon.search config")
}

func GenerateTuikeasyProductCouponSearchByChatRoomSerialNo(chat_room_serial_no string, db *gorm.DB, managerDB *gorm.DB) (string, error) {
	var robotChatRoom models.RobotChatRoom
	db.Where("chat_room_serial_no = ?", chat_room_serial_no).Where("is_open = ?", 1).Order("id desc").First(&robotChatRoom)
	if robotChatRoom.MyId != "" { //有绑定供应商
		//pid, err := FetchAlimamaSearchPid(robotChatRoom.MyId, robotChatRoom.SubId, managerDB)
		domain, err := FetchTuikeasySearchDomain(robotChatRoom.MyId, robotChatRoom.SubId, managerDB)
		if err != nil {
			return "", errors.New("no pid")
		}
		return "https://haschen.2mai2.com/index?pid=" + strings.TrimSpace(domain), nil
	}
	return "", errors.New("no tuikeasy.product.search config")

}

func GenerateTuikeasyProductSearchContentByKeyword(chat_room_serial_no, key string, db *gorm.DB, managerDB *gorm.DB) (string, error) {
	var robotChatRoom models.RobotChatRoom
	db.Where("chat_room_serial_no = ?", chat_room_serial_no).Where("is_open = ?", 1).Order("id desc").First(&robotChatRoom)
	if robotChatRoom.MyId != "" { //有绑定供应商
		//pid, err := FetchAlimamaSearchPid(robotChatRoom.MyId, robotChatRoom.SubId, managerDB)
		domain, err := FetchTuikeasySearchDomain(robotChatRoom.MyId, robotChatRoom.SubId, managerDB)
		if err != nil {
			return "", errors.New("no domain")
		}
		var cmd models.MyCmd
		db.Where("my_id = ?", robotChatRoom.MyId).Where("cmd_type = ?", "alimama.product.search").Where("is_open = 1").First(&cmd)
		if key != "" {
			url := GenerateTuikeasyProductSearchUrl(domain, key)
			content := strings.Replace(cmd.CmdReply, "{搜索关键词}", strings.TrimSpace(key), -1)
			content = strings.Replace(content, "{优惠链接}", ShortUrl(url), -1)
			return content, nil
		}
	}
	return "", errors.New("no tuikeasy.product.search config")
}

func GenerateTuikeasyProductSearchTextUrlByKeyword(chat_room_serial_no, key string, db *gorm.DB, managerDB *gorm.DB) (string, string, error) {
	var robotChatRoom models.RobotChatRoom
	db.Where("chat_room_serial_no = ?", chat_room_serial_no).Where("is_open = ?", 1).Order("id desc").First(&robotChatRoom)
	if robotChatRoom.MyId != "" { //有绑定供应商
		//pid, err := FetchAlimamaSearchPid(robotChatRoom.MyId, robotChatRoom.SubId, managerDB)
		domain, err := FetchTuikeasySearchDomain(robotChatRoom.MyId, robotChatRoom.SubId, managerDB)
		if err != nil {
			return "", "", errors.New("no domain")
		}
		var cmd models.MyCmd
		db.Where("my_id = ?", robotChatRoom.MyId).Where("cmd_type = ?", "alimama.product.search").Where("is_open = 1").First(&cmd)
		if key != "" {
			url := GenerateTuikeasyProductSearchUrl(domain, key)
			content := strings.Replace(cmd.CmdReply, "{搜索关键词}", strings.TrimSpace(key), -1)
			return content, url, nil
		}
	}
	return "", "", errors.New("no tuikeasy.product.search config")
}

func GenerateTuikeasyProductSearchUrl(domain, key string) string {
	//return "http://m.52jdyouhui.cn/" + url.QueryEscape(domain) + "/s/" + url.QueryEscape(strings.TrimSpace(key))
	return "https://haschen.2mai2.com/list?pid=" + strings.TrimSpace(domain) + "&kwd=" + url.QueryEscape(strings.TrimSpace(key))
}

func GenerateTuikeasyProductSearchContent(myId, domain, content string, db *gorm.DB) (string, error) {
	var cmd models.MyCmd
	db.Where("my_id = ?", myId).Where("cmd_type = ?", "alimama.product.search").Where("is_open = 1").First(&cmd)
	if cmd.ID > 0 && strings.Index(strings.TrimSpace(content), strings.TrimSpace(cmd.CmdValue)) == 0 {
		key := strings.Replace(content, cmd.CmdValue, "", 1)
		if key != "" {
			url := GenerateTuikeasyProductSearchUrl(domain, key)
			content := strings.Replace(cmd.CmdReply, "{搜索关键词}", strings.TrimSpace(key), -1)
			content = strings.Replace(content, "{优惠链接}", ShortUrl(url), -1)
			return content, nil
		}
	}
	return "", errors.New("no tuikeasy.product.search config")
}

func GenerateTuikeasyProductSearchTextUrl(myId, domain, content string, db *gorm.DB) (*models.MyCmd, string, string, error) {
	var cmd models.MyCmd
	db.Where("my_id = ?", myId).Where("cmd_type = ?", "alimama.product.search").Where("is_open = 1").First(&cmd)
	if cmd.ID > 0 && strings.Index(strings.TrimSpace(content), strings.TrimSpace(cmd.CmdValue)) == 0 {
		key := strings.Replace(content, cmd.CmdValue, "", 1)
		if key != "" {
			url := GenerateTuikeasyProductSearchUrl(domain, key)
			content := strings.Replace(cmd.CmdReply, "{搜索关键词}", strings.TrimSpace(key), -1)
			return &cmd, content, url, nil
		}
	}
	return nil, "", "", errors.New("no tuikeasy.product.search config")
}

func SendTuikeasyProductSearch(myId, domain, chatRoomSerialNo, content string, db *gorm.DB) error {
	cmd, text, link, err := GenerateTuikeasyProductSearchTextUrl(myId, domain, content, db)
	if err != nil {
		return err
	}
	message := &models.MessageQueue{}
	if CheckCmdUseWechatMini(cmd) == true {
		message = FixWechatMiniGroupMessageQueue(chatRoomSerialNo, strings.Replace(text, "{优惠链接}", "", -1), link, "http://zhiya-img.wdwdcdn.com/M_M5a6b0adc703a8.png")
	} else {
		message.ChatRoomSerialNoList = chatRoomSerialNo
		message.ChatRoomCount = 1
		message.MsgType = "2001"
		message.IsHit = true
		message.WeixinSerialNo = ""
		message.MsgContent = strings.Replace(text, "{优惠链接}", ShortUrl(link), -1)
		message.SendType = 1
		message.SendStatus = 0
	}
	err = db.Create(message).Error
	if err != nil {
		return err
	}
	message.QueueId, _ = HashID.Encode([]int{int(message.ID)})
	return db.Save(message).Error
}

func FixWechatMiniGroupContext(text, webviewUrl string) string {
	v := url.Values{}
	v.Set("url", webviewUrl)
	var mc []string
	mc = append(mc, "gh_af8953599a25@app")
	mc = append(mc, fmt.Sprintf("pages/event/webview.html?%s", v.Encode()))
	mc = append(mc, "wx8684cf95eb723443")
	mc = append(mc, "https://mp.weixin.qq.com/mp/waerrpageappid=wx8684cf95eb723443&type=upgrade&upgradetype=3#wechat_redirect")
	mc = append(mc, "http://mmbiz.qpic.cn/mmbiz_png/4YhLOcwTG8FTiaB6M5icrkwibciboO3QbdDcSHXichicZuor1KwyHZSU6ELKRxu9Np80CboX9sJ0yO3A8n308wrjd4pg/0?wx_fmt=png")
	mc = append(mc, text)
	mc = append(mc, text)
	mc = append(mc, "福利领取")
	for k, vv := range mc {
		mc[k] = base64.StdEncoding.EncodeToString([]byte(vv))
	}
	return strings.Join(mc, ";")
}

func FixWechatMiniGroupMessageQueue(chatRoomSerialNo, text, webviewUrl, img string) *models.MessageQueue {
	message := &models.MessageQueue{}
	message.ChatRoomSerialNoList = chatRoomSerialNo
	message.ChatRoomCount = 1
	message.MsgType = "2014"
	message.IsHit = true
	message.WeixinSerialNo = ""
	message.MsgContent = FixWechatMiniGroupContext(text, webviewUrl)
	message.Title = "福利领取"
	message.Href = img
	message.SendType = 1
	message.SendStatus = 0
	return message
}

func CheckCmdUseWechatMini(m *models.MyCmd) bool {
	var params map[string]interface{}
	err := json.Unmarshal([]byte(goutils.ToString(m.CmdParams)), &params)
	if err != nil {
		log.Println("CheckCmdUseWechatMini Error:", err)
		return false
	}
	if v, ok := params["use_mini"]; !ok {
		return false
	} else {
		switch v.(type) {
		case bool:
			return v.(bool)
		default:
			return false
		}
	}
}

func SendTuikeasyCouponSearch(myId, domain, chatRoomSerialNo, content string, db *gorm.DB) error {
	var cmd models.MyCmd
	db.Where("my_id = ?", myId).Where("cmd_type = ?", "alimama.coupon.search").Where("is_open = 1").First(&cmd)
	if cmd.ID > 0 && strings.Compare(strings.TrimSpace(content), strings.TrimSpace(cmd.CmdValue)) == 0 {
		//url := "http://m.52jdyouhui.cn/" + url.QueryEscape(domain) + "/"
		link := "https://haschen.2mai2.com/index?pid=" + strings.TrimSpace(domain)
		message := &models.MessageQueue{}
		if CheckCmdUseWechatMini(&cmd) == true {
			message = FixWechatMiniGroupMessageQueue(chatRoomSerialNo, strings.Replace(cmd.CmdReply, "{优惠链接}", "", -1), link, "http://zhiya-img.wdwdcdn.com/M_M5a6b0aab91245.png")
		} else {
			pubContent := strings.Replace(cmd.CmdReply, "{优惠链接}", ShortUrl(link), -1)
			message.ChatRoomSerialNoList = chatRoomSerialNo
			message.ChatRoomCount = 1
			message.MsgType = "2001"
			message.IsHit = true
			message.WeixinSerialNo = ""
			message.MsgContent = pubContent
			message.SendType = 1
			message.SendStatus = 0
		}
		err := db.Create(message).Error
		if err != nil {
			return err
		}
		message.QueueId, _ = HashID.Encode([]int{int(message.ID)})
		return db.Save(message).Error
	}
	//todo log
	return errors.New("no tuikeasy.coupon.search config")
}

var Shorten *shorten.T

func init() {
	Shorten = shorten.NewT(&shorten.TConfig{
		Source: "1998084992",
	})
}

/*
  生成新浪短链接
*/
func ShortUrl(link string) string {
	return Shorten.Short(link)
	/*
		p := url.Values{}
		p.Set("source", "1998084992")
		p.Set("url_long", link)
		req, err := http.NewRequest("GET", "https://api.weibo.com/2/short_url/shorten.json?"+p.Encode(), nil)
		if err != nil {
			return link
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return link
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return link
		}
		var rst map[string]interface{}
		err = json.Unmarshal(b, &rst)
		if err != nil {
			return link
		}
		if _, ok := rst["urls"]; !ok {
			return link
		}
		var links []map[string]interface{}
		err = json.Unmarshal([]byte(goutils.ToString(rst["urls"])), &links)
		if err != nil {
			return link
		}
		if len(links) == 0 {
			return link
		}
		if _, ok := links[0]["url_short"]; !ok {
			return link
		}
		return goutils.ToString(links[0]["url_short"])
	*/
}

func SyncChatOverCallback(b []byte, client *uchatlib.UchatClient) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	_, ok := rst["chat_room_serial_no"]
	if !ok {
		return errors.New("no chatRoomeSerialNo")
	}
	return uchatlib.SetChatRoomOver(goutils.ToString(rst["chat_room_serial_no"]), "pc user", client)
}

func SyncRobotFriendAddCallback(b []byte, db *gorm.DB) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	_, ok := rst["Data"]
	if !ok {
		return errors.New("no Data")
	}
	var data []map[string]interface{}
	err = json.Unmarshal([]byte(goutils.ToString(rst["Data"])), &data)
	if err != nil {
		return err
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	for _, v := range data {
		rec := &models.RobotFriend{}
		rec.RobotSerialNo = goutils.ToString(v["vcRobotSerialNo"])
		rec.WxUserSerialNo = goutils.ToString(v["vcUserSerialNo"])
		rec.NickName = goutils.ToString(v["vcNickName"])
		rec.Base64NickName = goutils.ToString(v["vcBase64NickName"])
		rec.HeadImages = goutils.ToString(v["vcHeadImgUrl"])
		rec.AddDate, _ = time.ParseInLocation("2006-01-02 15:04:05", goutils.ToString(v["dtAddDate"]), loc)
		err := db.Save(rec).Error
		if err != nil {
			return err
		}
	}
	return nil
}
