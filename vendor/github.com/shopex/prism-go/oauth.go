package prism

import (
	// "fmt"
	"net/http"
	"net/url"
	"path/filepath"
)

type OAuthInfo struct {
	AccessToken    string                 `json:"access_token"`
	Data           map[string]interface{} `json:"data"`
	Expires        int                    `json:"expires_in"`
	RefreshExpires int                    `json:"refresh_expires"`
	RefreshToken   string                 `json:"refresh_token"`
	SessionId      string                 `json:"session_id"`
	client         *Client
}

func (c *Client) RequrireOAuth(req *http.Request, w http.ResponseWriter) (info *OAuthInfo, err error) {
	code := req.URL.Query().Get("code")

	var redirect_url string
	req.URL.Host = req.Host

	if code == "" {
		p := &url.Values{}
		p.Set("client_id", c.Key)
		p.Set("response_type", "code")
		p.Set("redirect_uri", req.URL.String())
		redirect_url = c.oauth_url("authorize", p)

        w.Header().Set("Location", redirect_url)
        w.WriteHeader(302)
        w.Write([]byte("loading..."))
	} else {
		c2 := *c
		c2.Server = c.oauth_url("", nil)
		var rsp *Response
		rsp, err = c2.Post("token", &map[string]interface{}{
			"code":       code,
			"grant_type": "authorization_code",
		})

		if err == nil {
			info = &OAuthInfo{}
			rsp.Unmarshal(info)
			info.client = c
		}
		q2 := req.URL.Query()
		q2.Del("code")
		req.URL.RawQuery = q2.Encode()
		redirect_url = req.URL.String()
	}
    //todo: fix 
	return
}

func (c *Client) oauth_url(action string, params *url.Values) string {
	u := *c.serverUrl
	u.Path, _ = filepath.Abs(u.Path + "/../oauth/" + action)
	if params != nil {
		u.RawQuery = params.Encode()
	}
	return u.String()
}
