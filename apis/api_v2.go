package apis

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/lvzhihao/goutils"
)

var (
	CODEMAP map[string]string = map[string]string{
		"000000": "",
		"100001": "my_id is empty",
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

func SendMessageV2(ctx echo.Context) error {
	return SendMessage(ctx)
}

func GetRobotJoinListv2(ctx echo.Context) error {
	params := ctx.QueryParams()
	log.Fatal(params)
	my_id := params.Get("my_id")
	if my_id == "" {
		return ReturnError(ctx, "100001", nil)
	}
	page_size := pageParam(params.Get("page_size"), 10)
	page_num := pageParam(params.Get("page_num"), 1)
	orderby := orderParam(params.Get("orderby"), []string{"join_data", "chat_room_serial_no", "robot_serial_no"}, []string{"join_data DESC"})
	search := searchParam(params.Get("search"), []string{"chat_room_nick_name", "robot_nick_name"}, []string{})
	return nil
}

func pageParam(input interface{}, def int32) (num int32) {
	num = goutils.ToInt32(input)
	if num == 0 {
		num = def
	}
	return
}

func orderParam(input interface{}, allow []string, def []string) (ret []string) {
	list := strings.Split(goutils.ToString(input), ";")
	for _, v := range list {
		data := strings.Split(v, ":")
		if goutils.InStringSlice(allow, data[1]) {
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
	return
}

func searchParam(input interface{}, allow []string, def []string) (ret map[string]string) {
	return
}
