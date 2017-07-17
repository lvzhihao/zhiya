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
)

// 小U机器客户端
type UchatClient struct {
	MarchantNo     string
	MarchantSecret string
	Error          error
	DefaultTimeout time.Duration
}

/*
  初始化一个新的实例
*/
func NewClient(marchantNo, marchantSecret string) *UchatClient {
	return &UchatClient{
		MarchantNo:     marchantNo,
		MarchantSecret: marchantSecret,
		DefaultTimeout: 10 * time.Second,
	}
}

/*
  发送请求
*/
func (c *UchatClient) Post(action string, ctx interface{}) ([]byte, error) {
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

/*
  签名信息
*/
func (c *UchatClient) Sign(strCtx string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(strCtx+c.MarchantSecret)))
}

type RobotListResult struct {
	RobotInfo []map[string]string
}

/*
  获取设备列表
*/
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

/*
  获取验证码
*/
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

/*
  获取设备群列表
*/
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

/*
  发起群会员列表回调
*/
func (c *UchatClient) ChatRoomUserInfo(ctx map[string]string) error {
	ctx["MerchantNo"] = c.MarchantNo
	_, err := c.Post("http://skyagent.shequnguanjia.com/Merchant.asmx/ChatRoomUserInfo", ctx)
	return err
}

/*
  开启时时群信息
*/
func (c *UchatClient) ChatRoomOpenGetMessages(ctx map[string]string) error {
	ctx["MerchantNo"] = c.MarchantNo
	_, err := c.Post("http://skyagent.shequnguanjia.com/Merchant.asmx/ChatRoomOpenGetMessages", ctx)
	return err
}

/*
  关闭时时群信息
*/
func (c *UchatClient) ChatRoomCloseGetMessages(ctx map[string]string) error {
	ctx["MerchantNo"] = c.MarchantNo
	_, err := c.Post("http://skyagent.shequnguanjia.com/Merchant.asmx/ChatRoomCloseGetMessages", ctx)
	return err
}

/*
  关闭群
*/
func (c *UchatClient) ChatRoomOver(ctx map[string]string) error {
	ctx["MerchantNo"] = c.MarchantNo
	_, err := c.Post("http://skyagent.shequnguanjia.com/Merchant.asmx/ChatRoomOver", ctx)
	return err
}

/*
  群推送消息
*/
func (c *UchatClient) SendMessage(ctx map[string]interface{}) error {
	ctx["MerchantNo"] = c.MarchantNo
	_, err := c.Post("http://skyagent.shequnguanjia.com/Merchant.asmx/MerchantSendMessages", ctx)
	return err
}
