package apis

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/zhiya/models"
)

var (
	CODEMAP map[string]string = map[string]string{
		"000000": "",
		"100001": "my_id is empty",
		"100003": "log_serial_no invaild",
	}
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
	return SendMessage(ctx)
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
