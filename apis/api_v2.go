package apis

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	rmqtool "github.com/lvzhihao/go-rmqtool"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/uchatlib"
	"github.com/lvzhihao/zhiya/chatbot"
	"github.com/lvzhihao/zhiya/models"
	"github.com/lvzhihao/zhiya/uchat"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var (
	CODEMAP map[string]string = map[string]string{
		"000000": "",
		"100001": "my_id is empty",
		"100003": "log_serial_no invaild",
		"100007": "robot_serial_no is empty",
	}
	AmrConvertServer string = "http://127.0.0.1:8299"
	ChatBotClient    *chatbot.Client
	MessagePublisher *rmqtool.PublisherTool
)

type ReturnType struct {
	Code  string      `json:"code"`
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
}

func ReturnError(ctx echo.Context, code string, err error) error {
	var ret ReturnType
	ret.Code = code
	if e, ok := CODEMAP[code]; ok {
		ret.Error = e
	} else if err != nil {
		ret.Error = err.Error()
	} else {
		ret.Error = "unknow error"
	}
	ret.Data = nil
	return ctx.JSON(http.StatusOK, ret)
}

func ReturnData(ctx echo.Context, data interface{}) error {
	var ret ReturnType
	ret.Code = "000000"
	ret.Data = data
	return ctx.JSON(http.StatusOK, ret)
}

func SendMessageV2(ctx echo.Context) error {
	var rst map[string]interface{}
	err := json.Unmarshal([]byte(ctx.FormValue("context")), &rst)
	if err != nil {
		return ReturnError(ctx, "100051", err)
	}
	// 补充商户号和发送流水号
	rst["MerchantNo"] = viper.GetString("merchant_no")
	rst["vcRelaSerialNo"] = "api-" + goutils.RandomString(20)
	b, _ := json.Marshal(rst)
	ext := fmt.Sprintf(".%s.%s", rst["MerchantNo"], rst["vcRobotSerialNo"])
	MessagePublisher.PublishExt("uchat.chat.message", ext, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         b,
	})
	return ReturnData(ctx, "success")
}

func OverChatRoomV2(ctx echo.Context) error {
	robotSerialNo := ctx.FormValue("robot_serial_no")
	chatRoomSerialNo := ctx.FormValue("chat_room_serial_no")
	comment := ctx.FormValue("comment")
	if robotSerialNo == "" || chatRoomSerialNo == "" {
		return ReturnError(ctx, "100012", fmt.Errorf("params empty"))
	}
	var robotChat models.RobotChatRoom
	err := DB.Where("robot_serial_no = ?", robotSerialNo).Where("chat_room_serial_no = ?", chatRoomSerialNo).First(&robotChat).Error
	if err != nil {
		return ReturnError(ctx, "100011", err)
	}
	tx := DB.Begin()
	err = robotChat.Close(tx)
	if err != nil {
		tx.Rollback()
		return ReturnError(ctx, "100011", err)
	}
	err = uchatlib.SetChatRoomOver(chatRoomSerialNo, comment, Client)
	if err != nil {
		tx.Rollback()
		return ReturnError(ctx, "100010", err)
	} else {
		tx.Commit()
		return ReturnData(ctx, "success")
	}
}

func GetRobotJoinList(ctx echo.Context) error {
	params := ctx.QueryParams()
	my_id := params.Get("my_id")
	if my_id == "" {
		return ReturnError(ctx, "100001", nil)
	}
	var count int
	page_size := pageParam(params.Get("page_size"), 10)
	if page_size > 100 {
		page_size = 100
	}
	page_num := pageParam(params.Get("page_num"), 1)
	db := DB.Model(&models.RobotJoin{}).Where("my_id = ?", my_id).Where("status = ?", 0)
	db = ParseOrder(db, params.Get("orderby"), []string{"join_date", "chat_room_serial_no", "robot_serial_no"}, []string{"join_date DESC"})
	db = ParseSearch(db, params.Get("search"), []string{"chat_room_nick_name", "robot_nick_name"})
	err := db.Count(&count).Error
	if err != nil {
		return ReturnError(ctx, "100002", err)
	}
	var ret []models.RobotJoin
	//check count todo
	err = db.Offset((page_num - 1) * page_size).Limit(page_size).Find(&ret).Error
	if err != nil {
		return ReturnError(ctx, "100002", err)
	}
	return ReturnData(ctx, map[string]interface{}{
		"current": map[string]interface{}{
			"page_num":  page_num,
			"page_size": page_size,
		},
		"count": count,
		"list":  ret,
	})
}

func DeleteRobotJoin(ctx echo.Context) error {
	my_id := ctx.FormValue("my_id")
	if my_id == "" {
		return ReturnError(ctx, "100001", nil)
	}
	logSerialNoList := strings.Split(ctx.FormValue("log_serial_no"), ";")
	if len(logSerialNoList) == 0 {
		return ReturnError(ctx, "100003", nil)
	}
	var robotJoinList []models.RobotJoin
	// 开群接口小U处判断，这里可以不强制使用事务，减少索表机率
	err := DB.Where("my_id = ?", my_id).Where("status = ?", 0).Where("log_serial_no IN (?)", logSerialNoList).Find(&robotJoinList).Error
	if err != nil {
		return ReturnError(ctx, "100004", err)
	}
	if len(robotJoinList) == 0 {
		return ReturnError(ctx, "100003", nil)
	}
	tx := DB.Begin()
	for _, v := range robotJoinList {
		err := v.SetStatusDelete(tx)
		if err != nil {
			tx.Rollback()
			return ReturnError(ctx, "100005", err)
		}
	}
	tx.Commit()
	return ReturnData(ctx, nil)
}

func OpenChatRoom(ctx echo.Context) error {
	my_id := ctx.FormValue("my_id")
	if my_id == "" {
		return ReturnError(ctx, "100001", nil)
	}
	logSerialNoList := strings.Split(ctx.FormValue("log_serial_no"), ";")
	if len(logSerialNoList) == 0 {
		return ReturnError(ctx, "100003", nil)
	}
	var robotJoinList []models.RobotJoin
	// 开群接口小U处判断，这里可以不强制使用事务，减少索表机率
	err := DB.Where("my_id = ?", my_id).Where("status = ?", 0).Where("log_serial_no IN (?)", logSerialNoList).Find(&robotJoinList).Error
	if err != nil {
		return ReturnError(ctx, "100004", err)
	}
	if len(robotJoinList) == 0 {
		return ReturnError(ctx, "100003", nil)
	}
	var realLogSerialNoList []string
	tx := DB.Begin()
	for _, v := range robotJoinList {
		err := v.SetStatusOpen(tx)
		if err != nil {
			tx.Rollback()
			return ReturnError(ctx, "100005", err)
		}
		realLogSerialNoList = append(realLogSerialNoList, v.LogSerialNo)
	}
	ret, err := Client.PullRobotInChatRoomOpenChatRoom(map[string]string{
		"vcSerialNoList": strings.Join(realLogSerialNoList, ";"),
	}) //pull open chatroom
	if err != nil {
		tx.Rollback()
		return ReturnError(ctx, "100006", err)
	} else {
		log.Println("debug:", ret)
		tx.Commit()
		return ReturnData(ctx, nil)
	}
}

func UpdateRobotInfo(ctx echo.Context) error {
	robotSerialNo := ctx.FormValue("robot_serial_no")
	if robotSerialNo == "" {
		return ReturnError(ctx, "100007", nil)
	}
	var robot models.Robot
	err := DB.Where("serial_no = ?", robotSerialNo).Where("status = ?", 10).First(&robot).Error
	if err != nil {
		return ReturnError(ctx, "100008", err)
	}
	err = Client.RobotInfoModify(map[string]string{
		"vcRobotSerialNo": robot.SerialNo,
		//"vcRobotWxId":     robot.WxId,
		"vcHeadImgUrl": goutils.ToString(ctx.FormValue("head_img_url")),
		"vcNickName":   goutils.ToString(ctx.FormValue("nick_name")),
		"vcRemarkName": goutils.ToString(ctx.FormValue("remark_name")),
		"vcSign":       goutils.ToString(ctx.FormValue("person_sign")),
	})
	if err != nil {
		return ReturnError(ctx, "100009", err)
	} else {
		return ReturnData(ctx, nil)
	}
}

func EnusreWorkTemplateRule(work *models.WorkTemplate) {
	// default tuling
	switch work.CmdType {
	case "shop.intelligent.chatting":
		rst, err := ChatBotClient.ApplyRule(work.WorkTemplateId, "tuling")
		if err != nil {
			log.Println("EnusreChatbotRule Error:", err)
		}
		log.Println(rst)
	}
}

func CreateWorkTemplate(ctx echo.Context) error {
	myId := ctx.FormValue("my_id")
	subId := ctx.FormValue("sub_id")
	name := ctx.FormValue("name")
	cmdType := ctx.FormValue("cmd_type")
	cmdValue := ctx.FormValue("cmd_value")
	cmdParams := ctx.FormValue("cmd_params")
	cmdReply := ctx.FormValue("cmd_reply")
	var status int32
	if ctx.FormValue("status") != "" {
		status = goutils.ToInt32(ctx.FormValue("status"))
	} else {
		status = 0 //不修改
	}
	ret, err := uchat.CreateWorkTemplate(DB, myId, subId, name, cmdType, cmdValue, cmdParams, cmdReply, int8(status))
	if err != nil {
		return ReturnError(ctx, "100013", err)
	} else {
		EnusreWorkTemplateRule(ret)
		return ReturnData(ctx, ret)
	}
}

func UpdateWorkTemplate(ctx echo.Context) error {
	workTemplateId := ctx.FormValue("work_template_id")
	name := ctx.FormValue("name")
	cmdValue := goutils.EchoCtxFormNullString(ctx, "cmd_value")
	cmdParams := goutils.EchoCtxFormNullString(ctx, "cmd_params")
	cmdReply := goutils.EchoCtxFormNullString(ctx, "cmd_reply")
	var status int32
	if ctx.FormValue("status") != "" {
		status = goutils.ToInt32(ctx.FormValue("status"))
	} else {
		status = -1 //不修改
	}
	ret, err := uchat.UpdateWorkTemplate(DB, workTemplateId, name, cmdValue, cmdParams, cmdReply, int8(status))
	if err != nil {
		return ReturnError(ctx, "100014", err)
	} else {
		EnusreWorkTemplateRule(ret)
		return ReturnData(ctx, ret)
	}
}

func WorkTemplateList(ctx echo.Context) error {
	myId := ctx.FormValue("my_id")
	subId := ctx.FormValue("sub_id")
	cmdType := ctx.FormValue("cmd_type")
	ret, err := uchat.ListWorkTemplate(DB, myId, subId, cmdType)
	if err != nil {
		return ReturnError(ctx, "100015", err)
	} else {
		return ReturnData(ctx, ret)
	}
}

func WorkTemplate(ctx echo.Context) error {
	workTemplateId := ctx.FormValue("work_template_id")
	ret, err := uchat.GetWorkTemplate(DB, workTemplateId)
	if err != nil {
		return ReturnError(ctx, "100017", err)
	} else {
		return ReturnData(ctx, ret)
	}
}

func SetWorkTemplateDefault(ctx echo.Context) error {
	myId := ctx.FormValue("my_id")
	subId := ctx.FormValue("sub_id")
	workTemplateId := ctx.FormValue("work_template_id")
	ret, err := uchat.SetDefaultWorkTemplate(DB, myId, subId, workTemplateId)
	if err != nil {
		return ReturnError(ctx, "100016", err)
	} else {
		return ReturnData(ctx, ret)
	}
}

func GetChatRoomTag(ctx echo.Context) error {
	myId := ctx.FormValue("my_id")
	ret, err := models.GetChatRoomTagByMyId(DB, myId)
	if err != nil {
		return ReturnError(ctx, "100060", err)
	} else {
		return ReturnData(ctx, ret)
	}
}

func CmdTypeList(ctx echo.Context) error {
	list, err := uchat.ListCmdType(DB)
	if err != nil {
		return ReturnError(ctx, "100020", err)
	} else {
		return ReturnData(ctx, list)
	}
}

func ApplyWorkTemplate(ctx echo.Context) error {
	myId := ctx.FormValue("my_id")
	subId := ctx.FormValue("sub_id")
	workTemplateId := ctx.FormValue("work_template_id")
	chatRoomList := strings.Split(strings.TrimSpace(ctx.FormValue("chat_rooms_list")), ",")
	ret, err := uchat.ApplyChatRoomTemplate(DB, myId, subId, workTemplateId, chatRoomList)
	if err != nil {
		return ReturnError(ctx, "100018", err)
	} else {
		return ReturnData(ctx, ret)
	}
}

func GetChatRoomTemplates(ctx echo.Context) error {
	chat_room_serial_no := ctx.FormValue("chat_room_serial_no")
	ret, err := uchat.GetChatRoomTemplates(DB, chat_room_serial_no)
	if err != nil {
		return ReturnError(ctx, "100019", err)
	} else {
		return ReturnData(ctx, ret)
	}
}

func GetValidRobot(ctx echo.Context) error {
	myId := ctx.FormValue("my_id")
	//subId := ctx.FormValue("sub_id")
	limit := ctx.FormValue("limit")

	limitNum := 10
	if limit != "" {
		limitNum = int(goutils.ToInt(limit))
	}

	// 查找此用户可用当天加群机器人
	robots, err := models.FindValidCodeRobotByMyId(DB, myId, limitNum)
	if err != nil {
		return ReturnError(ctx, "100024", err)
	} else {
		//随机选取一个机器人
		numn := rand.Intn(len(robots))
		return ReturnData(ctx, robots[numn])
	}
}

func UpdateRobotExpireTime(ctx echo.Context) error {
	myId := ctx.FormValue("my_id")
	expire := ctx.FormValue("expire")

	myRobots, err := models.FindRobotByMyId(DB, myId)
	if err != nil {
		return ReturnError(ctx, "100025", err)
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	date, err := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s 23:59:59", strings.TrimSpace(expire)), loc)
	if err != nil {
		return ReturnError(ctx, "100026", err)
	}
	for _, robot := range myRobots {
		robot.ExpireTime = date
		err := DB.Save(&robot).Error
		if err != nil {
			return ReturnError(ctx, "100027", err)
		}
	}
	return ReturnData(ctx, "success")
}

func AmrConver(ctx echo.Context) error {
	params := ctx.QueryParams()
	source := params.Get("source")
	if source == "" {
		return ReturnError(ctx, "100028", nil)
	}
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	URL, err := url.Parse(AmrConvertServer)
	if err != nil && strings.ToLower(URL.Scheme) == "https" {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	p := url.Values{}
	p.Set("source", source)
	req, err := http.NewRequest(
		"GET",
		strings.TrimRight(AmrConvertServer, "/")+"/api/decoder?"+p.Encode(),
		nil,
	)
	if err != nil {
		return ReturnError(ctx, "100029", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return ReturnError(ctx, "100030", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ReturnError(ctx, "100031", err)
	}
	var rst map[string]interface{}
	err = json.Unmarshal(b, &rst)
	if err != nil {
		return ReturnError(ctx, "100032", err)
	}
	if status, ok := rst["status"]; ok {
		if strings.ToLower(goutils.ToString(status)) == "success" {
			var data []interface{}
			b, _ := json.Marshal(rst["data"])
			err := json.Unmarshal(b, &data)
			if err != nil {
				return ReturnError(ctx, "100035", fmt.Errorf("%v", rst))
			} else {
				return ReturnData(ctx, data[0])
			}
		} else {
			return ReturnError(ctx, "100033", fmt.Errorf("%s", rst["message"]))
		}
	} else {
		return ReturnError(ctx, "100034", fmt.Errorf("%v", rst))
	}
}

func UpdateChatRoomRobotNickName(ctx echo.Context) error {
	chatRoomSerialNo := ctx.FormValue("chat_room_serial_no")
	nickName := ctx.FormValue("nick_name")
	if chatRoomSerialNo == "" {
		return ReturnError(ctx, "100040", fmt.Errorf("chat_room_serial_no is empty"))
	}
	if nickName == "" {
		return ReturnError(ctx, "100041", fmt.Errorf("nick_name is empty"))
	}
	err := Client.ChatRoomRobotNickNameModify(map[string]string{
		"vcChatRoomSerialNo": chatRoomSerialNo,
		"vcNickName":         nickName,
	})
	if err != nil {
		return ReturnError(ctx, "100042", err)
	} else {
		return ReturnData(ctx, "success")
	}
}

func GetChatRoomInfo(ctx echo.Context) error {
	params := ctx.QueryParams()
	chatRoomSerialNoList := params.Get("chat_room_serial_no_list")
	if chatRoomSerialNoList == "" {
		return ReturnError(ctx, "100044", fmt.Errorf("chat_room_serial_no_list is empty"))
	}
	ids := strings.Split(chatRoomSerialNoList, ",")
	var rst []models.ChatRoom
	err := DB.Where("chat_room_serial_no IN (?)", ids).Find(&rst).Error
	if err != nil {
		return ReturnError(ctx, "100045", err)
	} else {
		return ReturnData(ctx, rst)
	}
}

func pageParam(input interface{}, def int) (num int) {
	num = int(goutils.ToInt32(input))
	if num == 0 {
		num = def
	}
	return
}

func ParseOrder(db *gorm.DB, input interface{}, allow []string, def []string) *gorm.DB {
	var ret []string
	list := strings.Split(goutils.ToString(input), ";")
	for _, v := range list {
		data := strings.Split(v, ":")
		if len(data) == 2 && goutils.InStringSlice(allow, data[1]) {
			switch strings.ToLower(data[2]) {
			case "asc":
				ret = append(ret, data[1]+" "+"ASC")
			default:
				ret = append(ret, data[1]+" "+"DESC")
			}
		}
	}
	if len(ret) == 0 {
		ret = def
	}
	for _, o := range ret {
		db = db.Order(o)
	}
	return db
}

func ParseSearch(db *gorm.DB, input interface{}, allow []string) *gorm.DB {
	list := strings.Split(goutils.ToString(input), ";")
	for _, v := range list {
		data := strings.Split(v, ":")
		if len(data) == 2 && goutils.InStringSlice(allow, data[1]) {
			db = db.Where(fmt.Sprintf("%s LIKE ?", data[1]), fmt.Sprintf("%%%s%%", data[2]))
		}
	}
	return db
}
