package rmqtool

func (c *APIClient) ListChannels() ([]map[string]interface{}, error) {
	return c.readSlice("channels")
}

func APIListChannels(api, user, passwd string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListChannels()
}

func (c *APIClient) ListVhostChannels(name string) ([]map[string]interface{}, error) {
	return c.readSlice([]string{"vhosts", name, "channels"})
}

func APIListVhostChannels(api, user, passwd, name string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListVhostChannels(name)
}

func (c *APIClient) Channel(name string) (map[string]interface{}, error) {
	return c.readMap([]string{"channels", name})
}

func APIChannel(api, user, passwd, name string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).Channel(name)
}

func (c *APIClient) ListConsumers() ([]map[string]interface{}, error) {
	return c.readSlice("consumers")
}

func APIListConsumers(api, user, passwd string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListConsumers()
}

func (c *APIClient) ListVhostConsumers(name string) ([]map[string]interface{}, error) {
	return c.readSlice([]string{"consumers", name})
}

func APIListVhostConsumers(api, user, passwd, name string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListVhostConsumers(name)
}
