package shorten

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	DefaultTLRUCacheSize int           = 1000
	TAPI                 string        = "https://api.weibo.com/2/short_url"
	TTestURL             string        = "http://www.baidu.com"
	TTestTimeout         time.Duration = 2 * time.Second
	TTestInterval        time.Duration = 3 * time.Second
)

type TConfig struct {
	Source string
	Token  string
}

type T struct {
	config      *TConfig
	lk          sync.Mutex
	cache       *LRU //todo
	shortEnable bool
	testClient  *http.Client
}

func NewT(config *TConfig) *T {
	cache, _ := NewLRU(DefaultTLRUCacheSize, nil)
	t := &T{
		config:      config,
		cache:       cache,
		shortEnable: true,
		testClient: &http.Client{
			Timeout: TTestTimeout,
		},
	}
	t.Init()
	return t
}

func (t *T) Init() {
	t.testStatus()
}

func (t *T) testStatus() {
	short := t.shorts([]string{TTestURL})
	if short[0] == TTestURL {
		t.lk.Lock()
		defer t.lk.Unlock()
		t.shortEnable = false
	} else {
		rsp, err := t.testClient.Get(short[0])
		t.lk.Lock()
		defer t.lk.Unlock()
		if err != nil || rsp.StatusCode != 200 {
			t.shortEnable = false
		} else {
			t.shortEnable = true
		}
	}
	time.AfterFunc(TTestInterval, t.testStatus)
}

func (t *T) Short(link string) string {
	ret := t.Shorts([]string{link})
	return ret[0]
}

func (t *T) Shorts(links []string) []string {
	t.lk.Lock()
	if t.shortEnable != true {
		t.lk.Unlock()
		return links
	}
	t.lk.Unlock()
	return t.shorts(links)
}

func (t *T) shorts(links []string) []string {
	p := url.Values{}
	for _, link := range links {
		p.Add("url_long", link)
	}
	b, err := t.API("/shorten.json", p)
	if err != nil {
		return links
	}
	var rst map[string]interface{}
	err = json.Unmarshal(b, &rst)
	if err != nil {
		return links
	}
	if _, ok := rst["urls"]; !ok {
		return links
	}
	var data []map[string]interface{}
	b, _ = json.Marshal(rst["urls"])
	err = json.Unmarshal(b, &data)
	if len(links) == 0 {
		return links
	}
	var ret []string
	for k, v := range data {
		if short, ok := v["url_short"]; ok {
			ret = append(ret, fmt.Sprintf("%s", short))
		} else {
			ret = append(ret, links[k])
		}
	}
	return ret
}

func (t *T) API(path string, p url.Values) ([]byte, error) {
	if t.config.Token != "" {
		p.Set("token", t.config.Token)
	} else if t.config.Source != "" {
		p.Set("source", t.config.Source)
	} else {
		return nil, fmt.Errorf("Config Error: %t", t.config)
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s?%s", TAPI, path, p.Encode()), nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var rst map[string]interface{}
	err = json.Unmarshal(b, &rst)
	if err != nil {
		return nil, err
	}
	if _, ok := rst["error_code"]; ok {
		return nil, fmt.Errorf("%s", rst["error"])
	}
	return b, nil
}
