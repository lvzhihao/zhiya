package rmqtool

func (c *APIClient) ListConnections() ([]map[string]interface{}, error) {
	return c.readSlice("connections")
}

func APIListConnections(api, user, passwd string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListConnections()
}

func (c *APIClient) ListVhostConnections(vhost string) ([]map[string]interface{}, error) {
	return c.readSlice([]string{"vhosts", vhost, "connections"})
}

func APIListVhostConnections(api, user, passwd, vhost string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListVhostConnections(vhost)
}

func (c *APIClient) Connection(name string) (map[string]interface{}, error) {
	return c.readMap([]string{"connections", name})
}

func APIConnection(api, user, passwd, name string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).Connection(name)
}

func (c *APIClient) ForceDeleteConnection(name, reason string) error {
	req, err := c.NewRequest("delete", []string{"connections", name}, nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-Reason", reason)
	resp, err := c.Do(req)
	return c.ScanDelete(resp, err)
}

func APIForceDeleteConnection(api, user, passwd, name, reason string) error {
	return NewAPIClient(api, user, passwd).ForceDeleteConnection(name, reason)
}

func (c *APIClient) ListConnectionChannels(name string) ([]map[string]interface{}, error) {
	return c.readSlice([]string{"connections", name, "channels"})
}

func APIListConnectionChannels(api, user, passwd, name string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListConnectionChannels(name)
}
