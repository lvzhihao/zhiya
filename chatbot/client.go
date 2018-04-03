package chatbot

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/lvzhihao/goutils"
)

const (
	RULE_ROBOT_TULING = "tuling"
)

var (
	DefaultTransportInsecureSkipVerify bool = true
	DefaultTransportDisableCompression bool = true
)

type Result struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Error  string      `json:"error"`
}

type Client struct {
	config     *ClientConfig
	httpClient *http.Client
}

type ClientConfig struct {
	ApiHost        string
	MerchantNo     string
	MerchantSecret string
}

func NewClient(config *ClientConfig) *Client {
	c := &Client{
		config: config,
	}
	c.init()
	return c
}

func (c *Client) init() {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	URL, err := url.Parse(c.config.ApiHost)
	if err != nil && strings.ToLower(URL.Scheme) == "https" {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: DefaultTransportInsecureSkipVerify,
			},
			DisableCompression: DefaultTransportDisableCompression,
		}
	}
	c.httpClient = httpClient
}

type ApplyRuleResult struct {
	MerchantNo string `json:"merchant_no"`
	RuleId     string `json:"rule_id"`
	Type       string `json:"type"`
	RobotId    int    `json:"robot_id"`
	Params     string `json:"params"`
	Status     bool   `json:"status"`
}

func (c *Client) ApplyRule(ruleId, ruleType string) (rst *ApplyRuleResult, err error) {
	params := make(map[string]interface{}, 0)
	params["rule_id"] = ruleId
	params["type"] = ruleType
	res := c.Action("api/ApplyRule", "post", params)
	if res.Status == "success" {
		b, _ := json.Marshal(res.Data)
		err = json.Unmarshal([]byte(b), &rst)
	} else {
		err = fmt.Errorf(res.Error)
	}
	return

}

type SendMessageResult struct {
	Code    int                      `json:"code"`
	UserId  string                   `json:"user_id"`
	Text    string                   `json:"text"`
	Url     string                   `json:"url"`
	List    []map[string]interface{} `json:"list"`
	Proetry map[string]interface{}   `json:"proerty"`
}

func (c *Client) SendMessage(ruleId, ruleType, msg, userId string) (rst *SendMessageResult, err error) {
	params := make(map[string]interface{}, 0)
	params["rule_id"] = ruleId
	params["type"] = ruleType
	params["message"] = msg
	params["user_id"] = userId
	res := c.Action("api/SendMessage", "post", params)
	if res.Status == "success" {
		b, _ := json.Marshal(res.Data)
		err = json.Unmarshal([]byte(b), &rst)
	} else {
		err = fmt.Errorf(res.Error)
	}
	return
}

func (c *Client) Action(path, method string, params map[string]interface{}) *Result {
	req, err := c.request(path, method, params)
	if err != nil {
		return &Result{
			Status: "fail",
			Error:  err.Error(),
		}
	} else {
		rst, err := c.scan(c.do(req))
		if err != nil {
			return &Result{
				Status: "fail",
				Error:  err.Error(),
			}
		} else {
			return rst
		}
	}
}

func (c *Client) scan(resp *http.Response, err error) (*Result, error) {
	var rst Result
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &rst)
	return &rst, err
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *Client) sign(params map[string]interface{}) string {
	params["merchant_no"] = c.config.MerchantNo
	params["timestamp"] = time.Now().Unix()
	var keys []string
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	signOrigin := ""
	for _, k := range keys {
		signOrigin += k + goutils.ToString(params[k])
	}
	sign := fmt.Sprintf("%x", md5.Sum([]byte(signOrigin+c.config.MerchantSecret)))
	return strings.ToUpper(sign)
}

func (c *Client) request(path, method string, params map[string]interface{}) (req *http.Request, err error) {
	params["sign"] = c.sign(params)
	p := url.Values{}
	for k, v := range params {
		p.Set(k, goutils.ToString(v))
	}
	switch strings.ToLower(method) {
	case "get":
		req, err = http.NewRequest(
			strings.ToUpper(method),
			strings.TrimRight(c.config.ApiHost, "/")+"/"+path+"?"+p.Encode(),
			nil,
		)
	default:
		req, err = http.NewRequest(
			strings.ToUpper(method),
			strings.TrimRight(c.config.ApiHost, "/")+"/"+path,
			bytes.NewBufferString(p.Encode()),
		)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", goutils.ToString(len(p.Encode())))
	}
	return
}
