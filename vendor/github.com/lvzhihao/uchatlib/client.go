package uchatlib

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/lvzhihao/goutils"
)

var (
	UchatApiPrefix                     string         = "http://skyagent.shequnguanjia.com/Merchant.asmx" //接口地址
	UchatTimeZone                      string         = "Asia/Shanghai"                                   //时区设置
	UchatTimeLocation                  *time.Location                                                     //当前时区
	DefaultTransportInsecureSkipVerify bool           = true
	DefaultTransportDisableCompression bool           = true
)

func init() {
	SetTimeZone(UchatTimeZone)
	if os.Getenv("UCHAT_TIMEZONE") != "" {
		loc, err := time.LoadLocation(os.Getenv("UCHAT_TIMEZONE"))
		// loc set success
		if err == nil {
			UchatTimeZone = os.Getenv("UCHAT_TIMEZONE")
			UchatTimeLocation = loc
		}
	}
}

// 设置时区
func SetTimeZone(zone string) error {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return err
	} else {
		UchatTimeZone = zone
		UchatTimeLocation = loc
		return nil
	}
}

// 小U机器客户端
type UchatClient struct {
	ApiPrefix      string        //API
	MarchantNo     string        //商户号
	MarchantSecret string        //商户密钥
	Error          error         //最后一次错误
	DefaultTimeout time.Duration //默认请求超时
}

// 小U机器接口返回格式
type UchatClientResult struct {
	Result string        `json:"nResult"`  //1为成功, -1为失败
	Error  string        `json:"vcResult"` //文字描述
	Data   []interface{} `json:"Data"`     //返回结果
}

//  初始化一个新的实例
//  client := NewClient("marchantNo", "secret")
//  OR:
//  uchatlib.DefaultTimeout = 30 * time.Second
//  client := NewClient("marchantNo", "secret")
func NewClient(marchantNo, marchantSecret string) *UchatClient {
	return &UchatClient{
		ApiPrefix:      UchatApiPrefix,
		MarchantNo:     marchantNo,
		MarchantSecret: marchantSecret,
		DefaultTimeout: 10 * time.Second,
	}
}

// 发送动作
// b, err := client.Action("RobotList", nil)
// OR:
// ctx := make(map[string]string, 0)
// ctx["vcRobotSerialNo"] = "xxxxx"
// b, err := client.Action("ChatRoomList", ctx)
func (c *UchatClient) Action(action string, ctx interface{}) ([]byte, error) {
	req, err := c.request(action, ctx)
	if err != nil {
		return nil, err
	}
	rsp, err := c.Do(req)
	return c.Scan(rsp, err)
}

func (c *UchatClient) request(action string, ctx interface{}) (*http.Request, error) {
	var b []byte
	var err error
	switch ctx.(type) {
	case nil:
		input := make(map[string]interface{}, 0)
		input["MerchantNo"] = c.MarchantNo
		b, err = json.Marshal(input)
	case map[string]string:
		input := ctx.(map[string]string)
		input["MerchantNo"] = c.MarchantNo
		b, err = json.Marshal(input)
	case map[string]interface{}:
		input := ctx.(map[string]interface{})
		input["MerchantNo"] = c.MarchantNo
		b, err = json.Marshal(input)
	default:
		return nil, errors.New("params error")
	}
	if err != nil {
		return nil, err
	}
	p := url.Values{}
	p.Set("strContext", string(b))
	p.Set("strSign", c.Sign(string(b)))
	req, err := http.NewRequest(
		"POST",
		strings.TrimRight(c.ApiPrefix, "/")+"/"+strings.TrimLeft(action, "/"),
		bytes.NewBufferString(p.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", goutils.ToString(len(p.Encode())))
	return req, nil
}

// 审查请求结果
func (c *UchatClient) Scan(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
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

// 发送请求
func (c *UchatClient) Do(req *http.Request) (*http.Response, error) {
	client := &http.Client{
		Timeout: c.DefaultTimeout,
	}
	URL, err := url.Parse(c.ApiPrefix)
	if err != nil && strings.ToLower(URL.Scheme) == "https" {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: DefaultTransportInsecureSkipVerify,
			},
			DisableCompression: DefaultTransportDisableCompression,
		}
	}
	return client.Do(req)
}

// 签名信息
func (c *UchatClient) Sign(strCtx string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(strCtx+c.MarchantSecret)))
}

// 全局返回结果
type UchatClientGlobalResultList map[string]interface{}

// 审查全局返回结果
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

// 获取机器人列表接口
// 可通过该接口获取所购买的机器人列表，包含机器人编号，头像、昵称、状态、已开通群数等。
func (c *UchatClient) RobotList() ([]map[string]string, error) {
	b, err := c.Action("RobotList", nil)
	return c.ScanGlobalResultList("RobotInfo", b, err)
}

// 查询机器人已开通群接口
// 商家可通过该接口查询机器人已开通的群，包含群名称、群编号、群主编号等。
func (c *UchatClient) ChatRoomList(ctx map[string]string) ([]map[string]string, error) {
	b, err := c.Action("ChatRoomList", ctx)
	return c.ScanGlobalResultList("ChatRoomData", b, err)
}

// 机器人添加用户好友
// 商家可以通过机器人所提供的“机器人添加用户好友接口”去让更多的用户去能够添加机器人为好友，方便于群外的私聊信息的发送与朋友圈的推送。
func (c *UchatClient) RobotAddUser(ctx map[string]string) error {
	_, err := c.Action("RobotAddUser", ctx)
	return err
}

// 机器人信息编辑接口
// 商家通过机器人信息编辑接口可以对所拥有的机器人进行信息编辑，比如头像、昵称、个人签名等。
func (c *UchatClient) RobotInfoModify(ctx map[string]string) error {
	_, err := c.Action("RobotInfoModify", ctx)
	return err
}

// 修改群内机器人昵称接口
// 可以通过修改群内机器人昵称对群内机器人的昵称进行修改。
func (c *UchatClient) ChatRoomRobotNickNameModify(ctx map[string]string) error {
	_, err := c.Action("ChatRoomRobotNickNameModify", ctx)
	return err
}

// 获取机器人验证码接口
// 商家可通过该接口获取机器人的信息以及开群验证码，通过扫码添加好友后开群。（注：机器人只有以扫码方式申请添加好友才会自动通过，以其他方式添加好友则不会通过。）
func (c *UchatClient) ApplyCodeList(ctx map[string]string) ([]map[string]string, error) {
	data, err := c.Action("ApplyCodeList", ctx)
	return c.ScanGlobalResultList("ApplyCodeData", data, err)
}

// 批量开群接口
// 商家通过调用批量开群接口直接开群。
func (c *UchatClient) PullRobotInChatRoomOpenChatRoom(ctx map[string]string) ([]map[string]string, error) {
	data, err := c.Action("PullRobotInChatRoomOpenChatRoom", ctx)
	return c.ScanGlobalResultList("CanNotSerialNoData", data, err)
}

// 批量建群开群接口
// 商家通过在建群宝首页上注册一个账户，之后会获得一个UserID，商家需要将建群宝的UserID填写好，去调用“批量建群开群接口”。可以进行建群并开群。
func (c *UchatClient) AddNewChatRooms(ctx map[string]string) error {
	_, err := c.Action("AddNewChatRooms", ctx)
	return err
}

// 群注销接口
// 当需要注销已经开通了的群时，商家可直接调用群注销接口注销，群注销后，机器人会退出群聊，同时相应的机器人可开群的权限增加一个。
func (c *UchatClient) ChatRoomOver(ctx map[string]string) error {
	_, err := c.Action("ChatRoomOver", ctx)
	return err
}

// 机器人设置分组
// 将多个机器人设置为一组，用于多号。
func (c *UchatClient) RobotGroupInsert(ctx map[string]string) (string, error) {
	_, err := c.Action("RobotGroupInsert", ctx)
	//只有这个接口不是标准格式返回
	//todo xxx
	return "xxx", err
}

// 机器人组开群接口
// 商家通过调用机器人组开群接口开群，并将其他两个机器人拉入群中，用于多号。
func (c *UchatClient) RobotGroupOpenChatRoom(ctx map[string]string) error {
	_, err := c.Action("RobotGroupOpenChatRoom", ctx)
	return err
}

// 修改群资料接口
// 当调用这个接口的时候，我们可以通过机器人来对群设置群名称、群公告和群聊邀请确认开关。注意点：除修改群名称时100人以下的群不需要机器人是群主外，其他情况都需要机器人是群主。
func (c *UchatClient) ChatRoomInfoModify(ctx map[string]string) error {
	_, err := c.Action("ChatRoomInfoModify", ctx)
	return err
}

// 查询群状态接口
// 商家通过机器人提供的“查询群状态的接口”可以实时的知道所需要知道群的状态和机器人状态，也就是目前群是否还存在，机器人是否还工作。
func (c *UchatClient) ChatRoomStatus(ctx map[string]string) (map[string]string, error) {
	b, err := c.Action("ChatRoomStatus", ctx)
	if err != nil {
		return nil, err
	} else {
		var rst map[string]string
		err := json.Unmarshal(b, &rst)
		return rst, err
	}
}

// 入群欢迎语接口
// 当群内有新成员加加入的时候，APP会自动发送入群欢迎语。当商户设置入群欢迎语以后，设置的信息会作为群信息同步到APP。
func (c *UchatClient) ChatRoomWelcomeMsgConfig(ctx map[string]string) error {
	_, err := c.Action("ChatRoomWelcomeMsgConfig", ctx)
	return err
}

// 转让群主接口
// 当机器人为群主时，可以通过调用这个接口，将群主转让给指定的群成员。
func (c *UchatClient) ChatRoomAdminChange(ctx map[string]string) error {
	_, err := c.Action("ChatRoomAdminChange", ctx)
	return err
}

// 群内踢人接口
// 当机器人为群主时，群内有群成员因为特殊原因需要被请出群外时，商家调用这个接口，然后机器人会主动帮忙踢出这个人。
func (c *UchatClient) ChatRoomKicking(ctx map[string]interface{}) error {
	_, err := c.Action("ChatRoomKicking", ctx)
	return err
}

// 群内关键字设置接口
// 商家通过该接口设置关键字，当群内有成员发送的消息触发了关键词时，会通过【群内关键字触发消息回调接口】将发言成员信息与发言内容通知给商家。
func (c *UchatClient) MerchantCmd(ctx map[string]interface{}) error {
	_, err := c.Action("MerchantCmd", ctx)
	return err
}

// 获取群二维码
func (c *UchatClient) GetQrCode(ctx map[string]string) error {
	_, err := c.Action("GetQrCode", ctx)
	return err
}

// 获取群成员信息
// 商家可以通过机器人所提供的“群成员信息接口”可以了解到群成员信息列表。
func (c *UchatClient) ChatRoomUserInfo(ctx map[string]string) error {
	_, err := c.Action("ChatRoomUserInfo", ctx)
	return err
}

// 群成员匹配
// 商家可以通过机器人所提供的“群成员匹配接口”与自身的会员系统打通，对群成员进行管理。
func (c *UchatClient) ChatRoomUserMatch(ctx map[string]string) ([]map[string]string, error) {
	data, err := c.Action("ChatRoomUserMatch", ctx)
	return c.ScanGlobalResultList("MatchData", data, err)
}

// 群内消息发送接口
// 通过机器人在群内发送信息，可支持直接群发，@all群发和@单个人或多人群发。
func (c *UchatClient) MerchantSendMessages(ctx map[string]interface{}) error {
	_, err := c.Action("MerchantSendMessages", ctx)
	return err
}

// 开启群内实时消息
// 商家可通过“开启群内实时消息”的接口选择需要实时抓取聊天信息的群。
func (c *UchatClient) ChatRoomOpenGetMessages(ctx map[string]string) error {
	_, err := c.Action("ChatRoomOpenGetMessages", ctx)
	return err
}

// 关闭群内实时消息
// 商家可通过“关闭群内实时消息接口”关闭对群内实时聊天信息的抓取，关闭后，群内聊天信息不再发送给商家。
func (c *UchatClient) ChatRoomCloseGetMessages(ctx map[string]string) error {
	_, err := c.Action("ChatRoomCloseGetMessages", ctx)
	return err
}

//todo delete
/*
  群推送消息
*/
func (c *UchatClient) SendMessage(ctx map[string]interface{}) error {
	_, err := c.Action("MerchantSendMessages", ctx)
	return err
}
