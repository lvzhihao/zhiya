package uchat

import (
	"fmt"

	"github.com/jinzhu/gorm"
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

func UpdateWorkTemplate(db *gorm.DB, workTemplateId, name, cmdValue, cmdParams, cmdReply string, status int8) (ret *models.WorkTemplate, err error) {
	ret = &models.WorkTemplate{}
	err = db.Where("work_template_id = ?", workTemplateId).First(ret).Error
	if err != nil {
		return
	}
	if name != "" {
		ret.Name = name
	}
	if status >= 0 && status <= 2 {
		if ret.IsDefault && status != 0 {
			err = fmt.Errorf("default template must enabled, please change default template")
			return
		} // 判断如果是默认模板的话不能停用或删除
		ret.Status = status
	}
	ret.CmdValue = cmdValue
	ret.CmdParams = cmdParams
	ret.CmdReply = cmdReply
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

func GetWorkTemplate(db *gorm.DB, workTemplateId string) (ret *models.WorkTemplate, err error) {
	ret = &models.WorkTemplate{}
	err = db.Where("work_template_id = ?", workTemplateId).First(ret).Error
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
