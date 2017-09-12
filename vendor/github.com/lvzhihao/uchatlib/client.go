package uchatlib

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/lvzhihao/goutils"
)

var (
	UchatApiPrefix string = "http://skyagent.shequnguanjia.com/Merchant.asmx"
)

// 小U机器客户端
type UchatClient struct {
	MarchantNo     string
	MarchantSecret string
	Error          error
	DefaultTimeout time.Duration
}

// 小U机器接口返回格式
type UchatClientResult struct {
	Result string        `json:"nResult"`
	Error  string        `json:"vcResult"`
	Data   []interface{} `json:"data"`
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
func (c *UchatClient) Do(action string, ctx interface{}) ([]byte, error) {
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
	req.Header.Add("Content-Length", goutils.ToString(len(p.Encode())))
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
	var rst UchatClientResult
	err = json.Unmarshal(b, &rst)
	if err != nil {
		return nil, err
	}
	if rst.Result != "1" {
		if rst.Error != "" {
			return nil, errors.New(rst.Error)
		} else {
			return nil, errors.New("未知错误")
		}
	}
	return json.Marshal(rst.Data[0])
}

/*
  签名信息
*/
func (c *UchatClient) Sign(strCtx string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(strCtx+c.MarchantSecret)))
}

type UchatClientGlobalResultList map[string]interface{}

/*
  全局
*/
func (c *UchatClient) ScanGlobalResultList(key string, data []byte, err error) ([]map[string]string, error) {
	if err != nil {
		return nil, err
	}
	var rst UchatClientGlobalResultList
	err = json.Unmarshal(data, &rst)
	if err != nil {
		return nil, err
	}
	if _, ok := rst[key]; !ok {
		return nil, errors.New(key + " not found")
	}
	var b []byte
	b, err = json.Marshal(rst[key])
	if err != nil {
		return nil, err
	}
	var ret []map[string]string
	err = json.Unmarshal(b, &ret)
	return ret, err
}

/*
  获取设备列表
*/
func (c *UchatClient) RobotList() ([]map[string]string, error) {
	ctx := make(map[string]string, 0)
	ctx["MerchantNo"] = c.MarchantNo
	data, err := c.Do(UchatApiPrefix+"/RobotList", ctx)
	return c.ScanGlobalResultList("RobotInfo", data, err)
}

/*
  获取验证码
*/
func (c *UchatClient) ApplyCodeList(ctx map[string]string) ([]map[string]string, error) {
	ctx["MerchantNo"] = c.MarchantNo
	data, err := c.Do(UchatApiPrefix+"/ApplyCodeList", ctx)
	return c.ScanGlobalResultList("ApplyCodeData", data, err)
}

/*
  获取设备群列表
*/
func (c *UchatClient) ChatRoomList(ctx map[string]string) ([]map[string]string, error) {
	ctx["MerchantNo"] = c.MarchantNo
	b, err := c.Do(UchatApiPrefix+"/ChatRoomList", ctx)
	return c.ScanGlobalResultList("ChatRoomData", b, err)
}

/*
 查询群状态接口
*/
func (c *UchatClient) ChatRoomStatus(ctx map[string]string) (map[string]string, error) {
	ctx["MerchantNo"] = c.MarchantNo
	b, err := c.Do(UchatApiPrefix+"/ChatRoomStatus", ctx)
	if err != nil {
		return nil, err
	} else {
		var rst map[string]string
		err := json.Unmarshal(b, &rst)
		return rst, err
	}
}

/*
  发起群会员列表回调
*/
func (c *UchatClient) ChatRoomUserInfo(ctx map[string]string) error {
	ctx["MerchantNo"] = c.MarchantNo
	_, err := c.Do(UchatApiPrefix+"/ChatRoomUserInfo", ctx)
	return err
}

/*
  开启时时群信息
*/
func (c *UchatClient) ChatRoomOpenGetMessages(ctx map[string]string) error {
	ctx["MerchantNo"] = c.MarchantNo
	_, err := c.Do(UchatApiPrefix+"/ChatRoomOpenGetMessages", ctx)
	return err
}

/*
  关闭时时群信息
*/
func (c *UchatClient) ChatRoomCloseGetMessages(ctx map[string]string) error {
	ctx["MerchantNo"] = c.MarchantNo
	_, err := c.Do(UchatApiPrefix+"/ChatRoomCloseGetMessages", ctx)
	return err
}

/*
  关闭群
*/
func (c *UchatClient) ChatRoomOver(ctx map[string]string) error {
	ctx["MerchantNo"] = c.MarchantNo
	_, err := c.Do(UchatApiPrefix+"/ChatRoomOver", ctx)
	return err
}

/*
  群推送消息
*/
func (c *UchatClient) SendMessage(ctx map[string]interface{}) error {
	ctx["MerchantNo"] = c.MarchantNo
	_, err := c.Do(UchatApiPrefix+"/MerchantSendMessages", ctx)
	return err
}

/*
 机器人添加用户好友
*/
func (c *UchatClient) RobotAddUser(ctx map[string]string) error {
	ctx["MerchantNo"] = c.MarchantNo
	_, err := c.Do(UchatApiPrefix+"/RobotAddUser", ctx)
	return err
}
