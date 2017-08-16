package goutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	RMQHttpApiTimeout time.Duration = 10 * time.Second
)

type RMQHttpApi struct {
	host   string
	user   string
	passwd string
	client *http.Client
}

type RMQPermission struct {
	User      string `json:"user"`
	Vhost     string `json:"vhost"`
	Configure string `json:"configure"`
	Write     string `json:"write"`
	Read      string `json:"read"`
}

func NewRMQHttpApi(host, user, passwd string) (c *RMQHttpApi, err error) {
	c = &RMQHttpApi{
		host:   host,
		user:   user,
		passwd: passwd,
		client: &http.Client{
			Timeout: RMQHttpApiTimeout,
		},
	}
	return
}

func (c *RMQHttpApi) GET(apiPath string) (*http.Response, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", strings.TrimRight(c.host, "/"), strings.TrimLeft(apiPath, "/")), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.passwd)
	req.Header.Add("Content-Type", "application/json")
	return c.client.Do(req)
}

func (c *RMQHttpApi) Scan(rsp *http.Response, err error) (int, []byte, error) {
	if err != nil {
		return 0, nil, err
	}
	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return 0, nil, err
	}
	return rsp.StatusCode, b, err
}

func (c *RMQHttpApi) WhoAmI() (map[string]string, error) {
	status, result, err := c.Scan(c.GET("api/whoami"))
	var ret map[string]string
	if status == 200 {
		err = json.Unmarshal(result, &ret)
		return ret, err
	} else {
		return nil, fmt.Errorf("status error: %d", status)
	}
}

func (c *RMQHttpApi) Permissions() ([]RMQPermission, error) {
	status, result, err := c.Scan(c.GET("api/permissions"))
	var ret []RMQPermission
	if status == 200 {
		err = json.Unmarshal(result, &ret)
		return ret, err
	} else {
		return nil, fmt.Errorf("status error: %d", status)
	}
}

func RegisterQueue(api, user, passwd, vhost, name, exchange, key string) error {
	client := &http.Client{}
	b := bytes.NewBufferString(`{"auto_delete":false, "durable":true, "arguments":[]}`)
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/queues/%s/%s", api, vhost, name), b)
	if err != nil {
		return err
	}
	// enusre queue
	req.SetBasicAuth(user, passwd)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("CreateQueue StatusError: %d, %v", resp.StatusCode, resp)
	}
	if exchange != "" && key != "" {
		b = bytes.NewBufferString(`{"routing_key":"` + key + `", "arguments":[]}`)
		// ensure binding
		req, err = http.NewRequest(
			"POST",
			fmt.Sprintf("%s/bindings/%s/e/%s/q/%s", api, vhost, exchange, name),
			b)
		req.SetBasicAuth(user, passwd)
		req.Header.Add("Content-Type", "application/json")
		resp, err = client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("BindRoutingKey StatusError: %d, %v", resp.StatusCode, resp)
		}
	}
	return nil
}
