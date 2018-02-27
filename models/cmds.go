package models

import "github.com/jinzhu/gorm"

// 触发指令
type CmdType struct {
	gorm.Model
	TypeFlag string `gorm:"size:50;unique_index" json:"type_flag"` //定义标识，方便系统回调，命令进入回调队列
	TypeName string `gorm:"size:50" json:"type_name"`              //定义名称
}

/*
  获取触发指令
  如果没有则初始化
*/

func (c *CmdType) Ensure(db *gorm.DB, typeFlag string) error {
	return db.Where(CmdType{TypeFlag: typeFlag}).FirstOrInit(c).Error
}

/*
  群指令
*/
type ChatRoomCmd struct {
	gorm.Model
	ChatRoomSerialNo string `gorm:"size:100" json:"chat_room_serial_no"`        //群编号
	CmdType          string `gorm:"size:50;index:idx_cmd_type" json:"cmd_type"` //指令类型
	CmdValue         string `gorm:"size:100" json:"cmd_value"`                  //指令内容
	CmdParams        string `gorm:"size:500" json:"cmd_params"`                 //指令参数
	CmdReply         string `gorm:"type:text(10000)" json:"cmd_reply"`          //指令回复
	IsOpen           bool   `gorm:"index:idx_is_open" json:"is_open"`           //是否启用
}

/*
  供应商指令
*/
type MyCmd struct {
	gorm.Model
	MyId      string `gorm:"size:100" json:"my_id"`                      //供应商ID
	CmdType   string `gorm:"size:50;index:idx_cmd_type" json:"cmd_type"` //指令类型
	CmdValue  string `gorm:"size:100" json:"cmd_value"`                  //指令内容
	CmdParams string `gorm:"size:500" json:"cmd_params"`                 //指令参数
	CmdReply  string `gorm:"type:text(10000)" json:"cmd_reply"`          //指令回复
	IsOpen    bool   `gorm:"index:idx_is_open" json:"is_open"`           //是否启用
}

/*
  代理商指令
*/
type SubCmd struct {
	gorm.Model
	SubId     string `gorm:"size:100" json:"sub_id"`                     //代理商ID
	CmdType   string `gorm:"size:50;index:idx_cmd_type" json:"cmd_type"` //指令类型
	CmdValue  string `gorm:"size:100" json:"cmd_value"`                  //指令内容
	CmdParams string `gorm:"size:500" json:"cmd_params"`                 //指令参数
	CmdReply  string `gorm:"type:text(10000)" json:"cmd_reply"`          //指令回复
	IsOpen    bool   `gorm:"index:idx_is_open" json:"is_open"`           //是否启用
}

// TAG指令
type TagCmd struct {
	gorm.Model
	TagId     string `gorm:"size:100" json:"tag_id"`                     //分组TAG
	CmdType   string `gorm:"size:50;index:idx_cmd_type" json:"cmd_type"` //指令类型
	CmdValue  string `gorm:"size:100" json:"cmd_value"`                  //指令内容
	CmdParams string `gorm:"size:500" json:"cmd_params"`                 //指令参数
	CmdReply  string `gorm:"type:text(10000)" json:"cmd_reply"`          //指令回复
	IsOpen    bool   `gorm:"index:idx_is_open" json:"is_open"`           //是否启用
}

// 指令模板
type WorkTemplate struct {
	gorm.Model
	WorkTemplateId string `gorm:"size:50;not null;unique_index" json:"work_template_id"` // 模板ID
	MyId           string `gorm:"size:100;not null;index" json:"my_id"`                  // 第三方绑定用户标识
	SubId          string `gorm:"size:100;index" json:"sub_id"`                          // 子商户标识，如果存在
	Name           string `gorm:"size:200" json:"name"`                                  // 模板名称
	IsDefault      bool   `json:"is_default"`                                            // 是否默认模板
	Status         int8   `json:"status"`                                                // 状态： 0 启用 1 关闭 2 不允许修改
}

// 模板指令详细
type WorkTemplateCmd struct {
	gorm.Model
	MyId           string `gorm:"size:100;not null;index" json:"my_id"`                                //第三方绑定用户标识
	SubId          string `gorm:"size:100;index" json:"sub_id"`                                        //子商户标识，如果存在
	WorkTempalteId string `gorm:"size:50;not null;index:idx_work_template_id" json:"work_template_id"` //模板ID
	CmdType        string `gorm:"size:50;not null;index:idx_cmd_type" json:"cmd_type"`                 //指令类型
	CmdValue       string `gorm:"size:100" json:"cmd_value"`                                           //指令内容
	CmdParams      string `gorm:"size:500" json:"cmd_params"`                                          //指令参数
	CmdReply       string `gorm:"type:text(10000)" json:"cmd_reply"`                                   //指令回复
	IsOpen         bool   `gorm:"index:idx_is_open" json:"is_open"`                                    //是否启用
}
