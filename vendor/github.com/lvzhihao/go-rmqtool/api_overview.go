package rmqtool

import "net/url"

func (c *APIClient) Overview() (map[string]interface{}, error) {
	return c.readMap("/overview")
}

func APIOverview(api, user, passwd string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).Overview()
}

func (c *APIClient) ListNodes() ([]map[string]interface{}, error) {
	return c.readSlice("/nodes")
}

func APIListNodes(api, user, passwd string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListNodes()
}

func (c *APIClient) Node(name string, params map[string]string) (map[string]interface{}, error) {
	req, err := c.NewRequest("GET", []string{"nodes", name}, nil)
	if err != nil {
		return nil, err
	}
	query := &url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}
	req.URL.RawQuery = query.Encode()
	resp, err := c.Do(req)
	return c.ScanMap(resp, err)
}

func APINode(api, user, passwd, name string, params map[string]string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).Node(name, params)
}

func (c *APIClient) ListExtensions() ([]map[string]interface{}, error) {
	return c.readSlice("/extensions")
}

func APIListExtensions(api, user, passwd string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListExtensions()
}

func (c *APIClient) Definitions() (map[string]interface{}, error) {
	return c.readMap("/definitions")
}

func APIDefinitions(api, user, passwd string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).Definitions()
}

func (c *APIClient) ChangeDefinitions(config map[string]interface{}) error {
	return c.update("/definitions", config)
}

func APIChangeDefinitions(api, user, passwd string, config map[string]interface{}) error {
	return NewAPIClient(api, user, passwd).ChangeDefinitions(config)
}

func (c *APIClient) VhostDefinitions(vhost string) (map[string]interface{}, error) {
	return c.readMap([]string{"definitions", vhost})
}

func APIVhostDefinitions(api, user, passwd, vhost string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).VhostDefinitions(vhost)
}

func (c *APIClient) ChangeVhostDefinitions(vhost string, config map[string]interface{}) error {
	return c.update([]string{"definitions", vhost}, config)
}

func APIChangeVhostDefinitions(api, user, passwd, vhost string, config map[string]interface{}) error {
	return NewAPIClient(api, user, passwd).ChangeVhostDefinitions(vhost, config)
}
