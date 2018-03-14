package uchat

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/zhiya/models"
)

func CreateWorkTemplate(db *gorm.DB, myId, subId, name, cmdType, cmdValue, cmdParams, cmdReply string) (ret *models.WorkTemplate, err error) {
	_, err = GetCmdType(db, cmdType)
	if err != nil {
		err = fmt.Errorf("cmdType error")
		return
	}
	if myId == "" {
		err = fmt.Errorf("myId is empty")
		return
	}
	if name == "" {
		err = fmt.Errorf("name is empty")
	}
	ret = &models.WorkTemplate{
		MyId:      myId,
		SubId:     subId,
		Name:      name,
		CmdType:   cmdType,
		CmdValue:  cmdValue,
		CmdParams: cmdParams,
		CmdReply:  cmdReply,
		IsDefault: false,
		Status:    0,
	}
	tx := db.Begin()
	err = tx.Save(ret).Error
	if err != nil {
		tx.Rollback()
		return
	}
	ret.WorkTemplateId, err = HashID.Encode([]int{int(ret.ID)})
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Save(ret).Error
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return
}

func UpdateWorkTemplate(db *gorm.DB, workTemplateId, name string, cmdValue, cmdParams, cmdReply *goutils.NullString, status int8) (ret *models.WorkTemplate, err error) {
	ret = &models.WorkTemplate{}
	err = db.Where("work_template_id = ?", workTemplateId).First(ret).Error
	if err != nil {
		return
	}
	if name != "" {
		ret.Name = name
	}
	if status >= 0 && status <= 2 {
		if ret.IsDefault && status == 2 {
			err = fmt.Errorf("default template don't delete, please change default template")
			return
		} // 判断如果是默认模板的话不能停用或删除
		ret.Status = status
	}
	if cmdValue.Valid {
		ret.CmdValue = cmdValue.String
	}
	if cmdParams.Valid {
		ret.CmdParams = cmdParams.String
	}
	if cmdReply.Valid {
		ret.CmdReply = cmdReply.String
	}
	err = db.Save(ret).Error
	return
}

func ListWorkTemplate(db *gorm.DB, myId, subId, cmdType string) (list *[]models.WorkTemplate, err error) {
	list = &[]models.WorkTemplate{}
	if cmdType == "" {
		err = db.Where("my_id = ?", myId).Where("sub_id = ?", subId).Where("status IN (?)", []int8{0, 1}).Find(list).Error
	} else {
		err = db.Where("my_id = ?", myId).Where("sub_id = ?", subId).Where("cmd_type = ?", cmdType).Where("status IN (?)", []int8{0, 1}).Find(list).Error
	}
	return
}

func GetWorkTemplate(db *gorm.DB, workTemplateId string) (ret interface{}, err error) {
	var temp models.WorkTemplate
	err = db.Where("work_template_id = ?", workTemplateId).First(&temp).Error
	if err != nil {
		return
	}
	cw, _ := GetChatRoomTemplate(db, temp.MyId, temp.SubId, temp.WorkTemplateId)
	ret = struct {
		models.WorkTemplate
		AllowChatRooms *[]models.ChatRoomWorkTemplate `json:"allow_chat_rooms"`
	}{
		temp,
		cw,
	}
	return
}

func SetDefaultWorkTemplate(db *gorm.DB, myId, subId, workTemplateId string) (ret *models.WorkTemplate, err error) {
	ret = &models.WorkTemplate{}
	err = db.Where("work_template_id = ?", workTemplateId).Where("my_id = ?", myId).Where("sub_id = ?", subId).Where("status = ?", 0).First(ret).Error //只有启动的模板才能设备成默认模板
	if err != nil {
		return
	}
	tx := db.Begin()
	err = tx.Model(&models.WorkTemplate{}).Where("my_id = ?", myId).Where("sub_id = ?", subId).Where("cmd_type = ?", ret.CmdType).Update("is_default", false).Error //这个类型下的模板取消所有默认
	if err != nil {
		tx.Rollback()
		return
	}
	ret.IsDefault = true
	err = tx.Save(ret).Error
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return
}

func ListCmdType(db *gorm.DB) (list *[]models.CmdType, err error) {
	list = &[]models.CmdType{}
	err = db.Find(&list).Error
	return
}

func GetCmdType(db *gorm.DB, typeFlag string) (ret *models.CmdType, err error) {
	ret = &models.CmdType{}
	err = db.Where("type_flag = ?", typeFlag).First(ret).Error
	return
}

func GetChatRoomTemplate(db *gorm.DB, myId, subId, workTemplateId string) (list *[]models.ChatRoomWorkTemplate, err error) {
	ret := &models.WorkTemplate{}
	err = db.Where("my_id = ?", myId).Where("sub_id = ?", subId).Where("status IN (?)", []int8{0, 1}).Where("work_template_id = ?", workTemplateId).First(ret).Error
	if err != nil {
		return
	}
	list = &[]models.ChatRoomWorkTemplate{}
	err = db.Where("work_template_id = ?", ret.WorkTemplateId).Find(list).Error
	return
}

func CountWorkTemplateChatRoom(db *gorm.DB, myId, subId string) error {
	var workTemplateList []models.WorkTemplate
	err := db.Where("my_id = ?", myId).Where("sub_id = ?", subId).Where("status IN (?)", []int8{0, 1}).Find(&workTemplateList).Error
	if err != nil {
		return err
	}
	for _, temp := range workTemplateList {
		var count int32
		err := db.Model(&models.ChatRoomWorkTemplate{}).Where("work_template_id = ?", temp.WorkTemplateId).Count(&count).Error
		if err != nil {
			continue
		}
		temp.ChatRoomCount = count
		db.Save(&temp)
	} // 些数据不影响主流程，不使用事务，如出错则不处理
	return nil
}

func ApplyChatRoomTemplate(db *gorm.DB, myId, subId, workTemplateId string, chatRoomList []string) (data interface{}, err error) {
	ret := &models.WorkTemplate{}
	err = db.Where("my_id = ?", myId).Where("sub_id = ?", subId).Where("status IN (?)", []int8{0, 1}).Where("work_template_id = ?", workTemplateId).First(ret).Error
	if err != nil {
		return
	}
	var realRooms []models.RobotChatRoom
	//todo fix expires time
	err = db.Where("my_id = ?", myId).Where("sub_id = ?", subId).Where("is_open = ?", true).Where("chat_room_serial_no IN (?)", chatRoomList).Find(&realRooms).Error
	if err != nil {
		return
	}
	if len(realRooms) == 0 {
		err = fmt.Errorf("There is no valid ChatRoomSeraialNo")
		return
	}
	var reals []string
	for _, room := range realRooms {
		reals = append(reals, room.ChatRoomSerialNo)
	}
	//尽量节省索表时，不使用replace into等操作，尽量使事务操作在500ms以内
	tx := db.Begin()
	err = tx.Model(&models.ChatRoomWorkTemplate{}).Where("chat_room_serial_no IN (?)", reals).Where("cmd_type = ?", ret.CmdType).Update("work_template_id", ret.WorkTemplateId).Error
	if err != nil {
		tx.Rollback()
		return
	}
	var existsChatRoom []models.ChatRoomWorkTemplate
	err = tx.Where("work_template_id = ?", ret.WorkTemplateId).Find(&existsChatRoom).Error
	if err != nil {
		tx.Rollback()
		return
	}
	var exists []string
	for _, room := range existsChatRoom {
		exists = append(exists, room.ChatRoomSerialNo)
	}
	//INSERT INTO table (a, b, c) VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?)
	var inserts []string
	var values []interface{}
	for _, serialNo := range reals {
		if goutils.InStringSlice(exists, serialNo) {
			continue
		}
		inserts = append(inserts, "(?, ?, ?, ?)")
		values = append(values, time.Now(), serialNo, ret.CmdType, ret.WorkTemplateId)
	}
	if len(inserts) > 0 {
		tableName := db.NewScope(&models.ChatRoomWorkTemplate{}).TableName()
		err = tx.Exec("INSERT INTO `"+tableName+"` (`created_at`, `chat_room_serial_no`, `cmd_type`, `work_template_id`) VALUES "+strings.Join(inserts, ", "), values...).Error
		if err != nil {
			tx.Rollback()
			return
		}
	}
	err = tx.Commit().Error
	if err != nil {
		return
	}
	/*old
	for _, room := range realRooms {
		var obj models.ChatRoomWorkTemplate
		err = db.Where(models.ChatRoomWorkTemplate{ChatRoomSerialNo: room.ChatRoomSerialNo, CmdType: ret.CmdType}).FirstOrInit(&obj).Error
		if err != nil {
			return
		}
		obj.WorkTemplateId = ret.WorkTemplateId
		err = tx.Save(&obj).Error
		if err != nil {
			tx.Rollback()
			return
		}
	}
	*/
	go CountWorkTemplateChatRoom(db, myId, subId) //此运营模板所绑定的群，可以异高操作
	data, err = GetWorkTemplate(db, ret.WorkTemplateId)
	return
}

func GetChatRoomTemplates(db *gorm.DB, chat_room_serial_no string) (data interface{}, err error) {
	var chatRoom models.ChatRoom
	err = db.Where("chat_room_serial_no = ?", chat_room_serial_no).First(&chatRoom).Error
	if err != nil {
		return
	}
	var temps []models.ChatRoomWorkTemplate
	err = db.Where("chat_room_serial_no = ?", chatRoom.ChatRoomSerialNo).Find(&temps).Error
	if err != nil {
		return
	}
	var templateIds []string
	var workTemplate []models.WorkTemplate
	if len(temps) > 0 {
		for _, temp := range temps {
			templateIds = append(templateIds, temp.WorkTemplateId)
		}
		err = db.Where("work_template_id IN (?)", templateIds).Find(&workTemplate).Error
		if err != nil {
			return
		}
	}
	data = struct {
		models.ChatRoom
		UsedWorkTemplates []models.WorkTemplate `json:"used_work_templates"`
	}{
		chatRoom,
		workTemplate,
	}
	return
}

func GetChatRoomValidTemplate(db *gorm.DB, chat_room_serial_no, cmd_type string) (data *models.WorkTemplate, err error) {
	data = &models.WorkTemplate{}
	var robotChatRoom models.RobotChatRoom
	// 群是否有关联设备并已经开启
	err = db.Where("chat_room_serial_no = ?", chat_room_serial_no).Where("is_open = ?", 1).First(&robotChatRoom).Error
	if err != nil {
		return
	}
	var ct models.ChatRoomWorkTemplate
	// 群目前所关联类型的运营模板
	err = db.Where("chat_room_serial_no = ?", robotChatRoom.ChatRoomSerialNo).Where("cmd_type = ?", cmd_type).First(&ct).Error
	if err == nil && ct.WorkTemplateId != "" {
		// 关联模板是否为群当前所属商家并且运营模板当前为开启
		err = db.Where("work_template_id = ?", ct.WorkTemplateId).Where("my_id = ?", robotChatRoom.MyId).Where("sub_id = ?", robotChatRoom.SubId).Where("status = ?", 0).First(&data).Error
		if err == nil {
			return
		}
	}
	// 读取前当商家关联类型默认模板并已经开启
	err = db.Where("cmd_type = ?", cmd_type).Where("my_id = ?", robotChatRoom.MyId).Where("sub_id = ?", robotChatRoom.SubId).Where("is_default = ?", true).Where("status = ?", 0).First(&data).Error
	return
}
