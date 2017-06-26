package uchat

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/goutils"
	"github.com/lvzhihao/uchat/models"
)

type UchatClient struct {
	MarchantNo     string
	MarchantSecret string
	Error          error
	DefaultTimeout time.Duration
}

func NewClient(marchantNo, marchantSecret string) *UchatClient {
	return &UchatClient{
		MarchantNo:     marchantNo,
		MarchantSecret: marchantSecret,
		DefaultTimeout: 10 * time.Second,
	}
}

func (c *UchatClient) Post(action string, ctx map[string]string) ([]byte, error) {
	b, err := json.Marshal(ctx)
	if err != nil {
		return nil, err
	}
	p := url.Values{}
	p.Set("strContext", string(b))
	p.Set("strSign", c.Sign(string(b)))
	req, err := http.NewRequest("POST", action, bytes.NewBufferString(p.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(p.Encode())))
	client := &http.Client{
		Timeout: c.DefaultTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var rst map[string]interface{}
	err = json.Unmarshal(b, &rst)
	if err != nil {
		return nil, err
	}
	if nResult, ok := rst["nResult"].(string); !ok || nResult != "1" {
		if vcResult, ok := rst["vcResult"].(string); ok {
			return nil, errors.New(vcResult)
		} else {
			return nil, errors.New("未知错误")
		}
	}
	b, err = json.Marshal(rst["Data"])
	if err != nil {
		return nil, err
	}
	var data []interface{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(data[0])
}

func (c *UchatClient) Sign(strCtx string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(strCtx+c.MarchantSecret)))
}

type RobotListResult struct {
	RobotInfo []map[string]string
}

func (c *UchatClient) RobotList() ([]map[string]string, error) {
	ctx := make(map[string]string, 0)
	ctx["MerchantNo"] = c.MarchantNo
	data, err := c.Post("http://skyagent.shequnguanjia.com/Merchant.asmx/RobotList", ctx)
	if err != nil {
		return nil, err
	} else {
		var rst RobotListResult
		err := json.Unmarshal(data, &rst)
		if err != nil {
			return nil, err
		} else {
			return rst.RobotInfo, nil
		}
	}
}

type ApplyCodeList struct {
	ApplyCodeData []map[string]string
}

func (c *UchatClient) ApplyCodeList(ctx map[string]string) ([]map[string]string, error) {
	ctx["MerchantNo"] = c.MarchantNo
	data, err := c.Post("http://skyagent.shequnguanjia.com/Merchant.asmx/ApplyCodeList", ctx)
	if err != nil {
		return nil, err
	} else {
		var rst ApplyCodeList
		err := json.Unmarshal(data, &rst)
		if err != nil {
			return nil, err
		} else {
			return rst.ApplyCodeData, nil
		}
	}
}

type ChatRoomList struct {
	ChatRoomData []map[string]string
}

func (c *UchatClient) ChatRoomList(ctx map[string]string) ([]map[string]string, error) {
	ctx["MerchantNo"] = c.MarchantNo
	data, err := c.Post("http://skyagent.shequnguanjia.com/Merchant.asmx/ChatRoomList", ctx)
	if err != nil {
		return nil, err
	} else {
		var rst ChatRoomList
		err := json.Unmarshal(data, &rst)
		if err != nil {
			return nil, err
		} else {
			return rst.ChatRoomData, nil
		}
	}
}

func (c *UchatClient) ChatRoomUserInfo(ctx map[string]string) error {
	ctx["MerchantNo"] = c.MarchantNo
	_, err := c.Post("http://skyagent.shequnguanjia.com/Merchant.asmx/ChatRoomUserInfo", ctx)
	return err
}

func SyncRobots(client *UchatClient, db *gorm.DB) error {
	rst, err := client.RobotList()
	if err != nil {
		return err
	}
	for _, v := range rst {
		robot := models.Robot{
			SerialNo:       v["vcSerialNo"],
			ChatRoomCount:  goutils.ToInt(v["nChatRoomCount"]),
			NickName:       v["vcNickName"],
			Base64NickName: v["vcBase64NickName"],
			HeadImages:     v["vcHeadImages"],
			CodeImages:     v["vcCodeImages"],
			Status:         int32(goutils.ToInt(v["nStatus"])),
		}
		err := db.Where(models.Robot{SerialNo: v["vcSerialNo"]}).Assign(robot).FirstOrCreate(&robot).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func SyncRobotChatRooms(RobotSerialNo string, client *UchatClient, db *gorm.DB) error {
	ctx := make(map[string]string, 0)
	ctx["vcRobotSerialNo"] = RobotSerialNo
	rst, err := client.ChatRoomList(ctx)
	if err != nil {
		return err
	}
	for _, v := range rst {
		room := models.ChatRoom{
			ChatRoomSerialNo: v["vcChatRoomSerialNo"],
			WxUserSerialNo:   v["vcWxUserSerialNo"],
			Name:             v["vcName"],
			Base64Name:       v["vcBase64Name"],
		}
		err := db.Where(models.ChatRoom{ChatRoomSerialNo: v["vcChatRoomSerialNo"]}).Assign(room).FirstOrCreate(&room).Error
		if err != nil {
			return err
		}
	}
	return nil
}
