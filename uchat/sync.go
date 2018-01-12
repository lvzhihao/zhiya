package uchat

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/uchatlib"
	"github.com/lvzhihao/zhiya/models"
	"github.com/lvzhihao/zhiya/tuling"
	"github.com/lvzhihao/zhiya/utils"
	hashids "github.com/speps/go-hashids"
	"github.com/spf13/viper"
)

var (
	DefaultMemberJoinWelcome      string        = ""
	DefaultMemberJoinSendInterval time.Duration = 60 * time.Second
	HashID                        *hashids.HashID
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
		robot.ChatRoomCount = goutils.ToInt32(v["nChatRoomCount"])
		nickNameB, _ := base64.StdEncoding.DecodeString(v["vcBase64NickName"])
		robot.NickName = goutils.ToString(nickNameB)
		robot.Base64NickName = v["vcBase64NickName"]
		robot.HeadImages = v["vcHeadImages"]
		robot.CodeImages = v["vcCodeImages"]
		robot.Status = goutils.ToInt32(v["nStatus"])
		err = db.Save(&robot).Error
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
		applyCode, applyCodeerr := models.ApplyCodeUsed(db, goutils.ToString(v["vcApplyCodeSerialNo"]))
		/*
			if err != nil {
				return err
			}
		*/
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
		if applyCodeerr == nil {
			robotRoom.MyId = applyCode.MyId
			robotRoom.SubId = applyCode.SubId
		} //如果有applyCode记录，可确定开群申请时的供应商和店铺身份
		robotRoom.IsOpen = true
		robotRoom.ExpiredDate = time.Now().Add(7 * 24 * time.Hour)
		err = db.Save(&robotRoom).Error
		if err != nil {
			return err
		}
		// sync member info
		SyncChatRoomMembers(chatRoomSerialNo, client)
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
	}
	return nil
}

/*
  同步群成员入群通知回调
  支持重复调用
*/
func SyncMemberJoinCallback(b []byte, db *gorm.DB) error {
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
		SendChatRoomMemberTextMessage(member.ChatRoomSerialNo, member.WxUserSerialNo, "", db)
	}
	return nil
}

/*
  发送群内信息
*/
func SendChatRoomMemberTextMessage(charRoomSerialNo, wxSerialNo, msg string, db *gorm.DB) error {
	log.Println("commit start")
	message := &models.MessageQueue{}
	tx := db.Begin()
	err := tx.Where("chat_room_serial_no_list = ?", charRoomSerialNo).Where("send_type = 10").Where("send_status = 0").First(&message).Error
	if err == nil {
		log.Println("update join message")
		//update
		message.WeixinSerialNo = message.WeixinSerialNo + "," + wxSerialNo //添加@新成员
	} else {
		log.Println("create join message")
		//new message
		if msg == "" {
			msg = FetchChatRoomMemberJoinMessage(charRoomSerialNo, db)
		}
		if msg == "" {
			tx.Rollback()
			return errors.New("no content")
		}
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
		message.SendType = 10 //第一个新成员加入后，30秒内所有加入成员一起发送，类型10
		message.SendStatus = 0
		message.SendTime = time.Now().Add(DefaultMemberJoinSendInterval)
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

func SyncChatKeywordCallback(b []byte, db *gorm.DB, managerDB *gorm.DB, tool *utils.ReceiveTool) error {
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
		db.Where("chat_room_serial_no = ?", goutils.ToString(v["vcChatRoomSerialNo"])).Where("is_open = ?", 1).Where("open_tuling = ?", 1).Order("id desc").First(&robotChatRoom)
		//对像是该群机器人
		if robotChatRoom.ID > 0 && strings.Compare(robotChatRoom.RobotSerialNo, goutils.ToString(v["vcToWxUserSerialNo"])) == 0 {
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

/*
  同步群成员时时消息回调
  支持重复调用
*/
func SyncChatMessageCallback(b []byte, db *gorm.DB, managerDB *gorm.DB) error {
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
			return "mmzl" + subId, nil
		}
	} else {
		pid := ""
		row := db.Table("sdb_maifou_promotion_detail").Where("platform = ?", "tuikeasy").Where("supplier_id = ?", myId).Where("is_self = ?", true).Select("pid").Row()
		row.Scan(&pid)
		if pid != "" {
			return "mmzlb" + utils.FakeIdEncode(goutils.ToInt64(myId)), nil
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

func GenerateTuikeasyProductSearchUrl(domain, key string) string {
	return "http://m.52jdyouhui.cn/" + url.QueryEscape(domain) + "/s/" + url.QueryEscape(strings.TrimSpace(key))
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

func SendTuikeasyProductSearch(myId, domain, chatRoomSerialNo, content string, db *gorm.DB) error {
	content, err := GenerateTuikeasyProductSearchContent(myId, domain, content, db)
	if err != nil {
		return err
	}
	message := &models.MessageQueue{}
	message.ChatRoomSerialNoList = chatRoomSerialNo
	message.ChatRoomCount = 1
	message.MsgType = "2001"
	message.IsHit = true
	message.WeixinSerialNo = ""
	message.MsgContent = content
	message.SendType = 1
	message.SendStatus = 0
	err = db.Create(message).Error
	if err != nil {
		return err
	}
	message.QueueId, _ = HashID.Encode([]int{int(message.ID)})
	return db.Save(message).Error
}

func SendTuikeasyCouponSearch(myId, domain, chatRoomSerialNo, content string, db *gorm.DB) error {
	return nil
	var cmd models.MyCmd
	db.Where("my_id = ?", myId).Where("cmd_type = ?", "alimama.coupon.search").Where("is_open = 1").First(&cmd)
	if cmd.ID > 0 && strings.Compare(strings.TrimSpace(content), strings.TrimSpace(cmd.CmdValue)) == 0 {
		url := "http://m.52jdyouhui.cn/" + url.QueryEscape(domain) + "/"
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
	return errors.New("no tuikeasy.coupon.search config")
}

/*
  生成新浪短链接
*/
func ShortUrl(link string) string {
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
