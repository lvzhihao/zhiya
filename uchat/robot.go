package uchat

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

type UcahtClient struct {
	MarchantNo     string
	MarchantSecret string
	Error          error
}

func (c *UcahtClient) Post(url string, ctx map[string]interface{}) (string, error) {
	b, err := json.Marshal(ctx)
	if err != nil {
		return "", err
	}
	params := url.Values{}
	url.Set("strContext", string(b))
	url.Set("strSign", c.Sgin(string(b))
	return "", nil
}

func (c *UcahtClient) Sign(strCtx string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(strCtx+c.MarchantSecret)))
}

func (c *UchatClient) FetchRobotsList() ([]*map[string]interface{}, error) {
	ctx := make(map[string]interface{}, 0)
	ctx["MerchantNo"] = c.MarchantNo
	data, err := c.Post("http://skyagent.shequnguanjia.com/Merchant.asmx/RobotList", ctx)
	if err != nil {
		return nil, err
	} else {
		var rst []*map[string]interface{}
		err := json.Unmarshal([]byte(data), &rst)
		if err != nil {
			return nil, err
		} else {
			return rst, nil
		}
	}
}
