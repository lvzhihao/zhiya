package uchat

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/zhiya/models"
)

/*
  同步设备列表
  支持重复调用
*/
func SyncRobots(client *UchatClient, db *gorm.DB) error {
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
func SyncChatRoomMembers(chatRoomSerialNo string, client *UchatClient) error {
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
	var robotRooms []models.RobotChatRoom
	err = db.Where("robot_serial_no = ?", robotSerialNo).Where("is_open = ?", true).Find(&robotRooms).Error
	if err != nil {
		return err
	}
	rootSerialNos := make([]string, 0)
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
			rootSerialNos = append(rootSerialNos, robotRoom.ChatRoomSerialNo)
		}
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
  同步设备开群信息
  支持重复调用
*/
func SyncRobotChatRooms(RobotSerialNo string, client *UchatClient, db *gorm.DB) error {
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
func SyncChatRoomCreateCallback(b []byte, client *UchatClient, db *gorm.DB) error {
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
		applyCode, err := models.ApplyCodeUsed(db, goutils.ToString(v["vcApplyCodeSerialNo"]))
		if err != nil {
			return err
		}
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
		robotRoom.MyId = applyCode.MyId
		robotRoom.SubId = applyCode.SubId
		robotRoom.IsOpen = true
		err = db.Save(&robotRoom).Error
		if err != nil {
			return err
		}
		// sync member info
		SyncChatRoomMembers(chatRoomSerialNo, client)
		// open get message
		log.Println(SetChatRoomOpenGetMessage(chatRoomSerialNo, client))
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
	if msg == "" {
		msg = FetchChatRoomMemberJoinMessage(charRoomSerialNo, db)
	}
	if msg == "" {
		return errors.New("no content")
	}
	message := &models.MessageQueue{}
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
	message.SendType = 1
	message.SendStatus = 0
	err := db.Create(message).Error
	if err != nil {
		return err
	}
	message.QueueId, _ = HashID.Encode([]int{int(message.ID)})
	return db.Save(message).Error
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
			pid, err := FetchAlimamaSearchPid(robotChatRoom.MyId, robotChatRoom.SubId, managerDB)
			if err != nil {
				log.Println(err)
				continue
			}
			err = SendAlimamProductSearch(robotChatRoom.MyId, pid, robotChatRoom.ChatRoomSerialNo, content, db)
			//如果已经匹配关键字则不用再去搜索优惠链接
			if err != nil {
				err = SendAlimamCouponSearch(robotChatRoom.MyId, pid, robotChatRoom.ChatRoomSerialNo, content, db)
				log.Println(err)
			}
		}
	}
	return nil
}

/*
  获取淘宝联盟广告位pid
*/
func FetchAlimamaSearchPid(myId, subId string, db *gorm.DB) (string, error) {
	if subId != "" {
		pid := ""
		row := db.Table("sdb_maifou_promotion_detail").Where("supplier_id = ?", myId).Where("shop_id = ?", subId).Select("pid").Row()
		row.Scan(&pid)
		if pid != "" {
			return pid, nil
		}
	} else {
		pid := ""
		row := db.Table("sdb_maifou_promotion_detail").Where("supplier_id = ?", myId).Where("is_self = ?", true).Select("pid").Row()
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
	var cmd models.MyCmd
	db.Where("my_id = ?", myId).Where("cmd_type = ?", "alimama.product.search").Where("is_open = 1").First(&cmd)
	if cmd.ID > 0 && strings.Index(strings.TrimSpace(content), strings.TrimSpace(cmd.CmdValue)) == 0 {
		key := strings.Replace(content, cmd.CmdValue, "", 1)
		if key != "" {
			params := url.Values{}
			params.Set("nav", "0")
			params.Set("p", pid)
			params.Set("searchText", strings.TrimSpace(key))
			url := "http://m.xuanwonainiu.com/search?" + params.Encode()
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

func SyncChatOverCallback(b []byte, client *UchatClient) error {
	var rst map[string]interface{}
	err := json.Unmarshal(b, &rst)
	if err != nil {
		return err
	}
	_, ok := rst["chat_room_serial_no"]
	if !ok {
		return errors.New("no chatRoomeSerialNo")
	}
	return SetChatRoomOver(goutils.ToString(rst["chat_room_serial_no"]), "pc user", client)
}
