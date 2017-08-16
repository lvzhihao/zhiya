package tuling

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/lvzhihao/goutils"
)

var TulingErrorCode = map[int]string{
	40001: "参数 key 错误",
	40002: "请求内容 info 为空",
	40004: "当天请求次数已使用完",
	40007: "数据格式异常",
}

var TulingTypeCode = map[int]string{
	100000: "文本类",
	200000: "链接类",
	302000: "新闻类",
	308000: "菜谱类",
	313000: "儿歌类",
	314000: "诗词类",
}

type TulingApiResult struct {
	Code     int                  `json:"code"`
	Text     string               `json:"text"`
	Url      string               `json:"url"`
	List     []TulingListStruct   `json:"list"`
	Function TulingFunctionStruct `json:"function"`
	listVer  int                  `json:"-"`
}

type TulingListStruct struct {
	Article   string `json:"article"`
	Source    string `json:"source"`
	Name      string `json:"name"`
	Info      string `json:"info"`
	Icon      string `json:"icon"`
	DetailUrl string `json:"detailurl"`
}

type TulingFunctionStruct struct {
	Song   string `json:"song"`
	Singer string `json:"singer"`
	Author string `json:"author"`
	Name   string `json:"name"`
}

type TulingClient struct {
	Config TulingClientConfig
}

type TulingClientConfig struct {
	ApiUrl         string
	ApiKey         string
	ApiSecret      string
	DefaultTimeout time.Duration
}

func (c *TulingApiResult) Error() error {
	if c.IsError() {
		return errors.New(c.Text)
	} else {
		return nil
	}
}

func (c *TulingApiResult) IsError() bool {
	for k, _ := range TulingErrorCode {
		if k == c.Code {
			return true
		}
	}
	return false
}

func (c *TulingApiResult) CodeName() string {
	if c.IsError() {
		return "错误"
	} else {
		if name, ok := TulingTypeCode[c.Code]; ok {
			return name
		} else {
			return "未知类型"
		}
	}
}

func NewTulingClient(c TulingClientConfig) *TulingClient {
	client := &TulingClient{
		Config: c,
	}
	return client
}

// 图灵只有java sdk，golang标准cbc加密接口返回错误，todo fix
func TulinApiSign(key, secret string, b []byte) ([]byte, error) {
	if secret == "" {
		return b, nil
	} else {
		tm := goutils.ToString(time.Now().UnixNano())
		paramsKey := fmt.Sprintf("%s%s%s", secret, tm, key)
		md5Key := md5.Sum([]byte(paramsKey))
		data, err := goutils.AesCBCEncrypt(md5Key[:16], goutils.ToString(b))
		if err != nil {
			return nil, err
		}
		ret := make(map[string]string, 0)
		ret["key"] = key
		ret["timestamp"] = tm
		ret["data"] = data
		return json.Marshal(ret)
	}
}

func (c *TulingClient) Sign(b []byte) ([]byte, error) {
	return TulinApiSign(c.Config.ApiKey, c.Config.ApiSecret, b)
}

func (c *TulingClient) Do(ctx map[string]interface{}) (*TulingApiResult, error) {
	ctx["key"] = c.Config.ApiKey //set api key
	b, err := json.Marshal(ctx)
	if err != nil {
		return nil, err
	}
	signB, err := c.Sign(b)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.Config.ApiUrl, bytes.NewBuffer(signB))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	client := &http.Client{
		Timeout: c.Config.DefaultTimeout,
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
	var ret TulingApiResult
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return nil, err
	} else {
		return &ret, nil
	}
}
