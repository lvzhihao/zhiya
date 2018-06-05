package rmqtool

import (
	"fmt"
	"net/http"
	"net/url"
)

func (c *APIClient) ListQueues(vhost string) ([]map[string]interface{}, error) {
	if vhost == "" {
		return c.readSlice("/queues")
	} else {
		return c.readSlice([]string{"queues", vhost})
	}
}

func APIListQueues(api, user, passwd, vhost string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListQueues(vhost)
}

func (c *APIClient) Queue(vhost, name string) (map[string]interface{}, error) {
	return c.readMap([]string{"queues", vhost, name})
}

func APIQueue(api, user, passwd, vhost, name string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).Queue(vhost, name)
}

func (c *APIClient) CreateQueue(vhost, name string, params map[string]interface{}) error {
	// `{"auto_delete":false, "durable":true, "arguments":[]}`
	return c.create([]string{"queues", vhost, name}, params)
}

func APICreateQueue(api, user, passwd, vhost, name string, params map[string]interface{}) error {
	return NewAPIClient(api, user, passwd).CreateQueue(vhost, name, params)
}

func (c *APIClient) DeleteQueue(vhost, name string) error {
	req, err := c.NewRequest("delete", []string{"queues", vhost, name}, nil)
	if err != nil {
		return err
	}
	query := &url.Values{}
	//query.Add("if-empty", "true")
	query.Add("if-unused", "true")
	req.URL.RawQuery = query.Encode()
	resp, err := c.Do(req)
	return c.ScanDelete(resp, err)
}

func APIDeleteQueue(api, user, passwd, vhost, name string) error {
	return NewAPIClient(api, user, passwd).DeleteQueue(vhost, name)
}

func (c *APIClient) ForceDeleteQueue(vhost, name string) error {
	return c.delete([]string{"queues", vhost, name})
}

func APIForceDeleteQueue(api, user, passwd, vhost, name string) error {
	return NewAPIClient(api, user, passwd).ForceDeleteQueue(vhost, name)
}

func (c *APIClient) PurgeQueue(vhost, name string) error {
	return c.delete([]string{"queues", vhost, name, "contents"})
}

func APIPurgeQueue(api, user, passwd, vhost, name string) error {
	return NewAPIClient(api, user, passwd).PurgeQueue(vhost, name)
}

func (c *APIClient) QueueBindings(vhost, name string) ([]map[string]interface{}, error) {
	return c.readSlice([]string{"queues", vhost, name, "bindings"})
}

func APIQueueBindings(api, user, passwd, vhost, name string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).QueueBindings(vhost, name)
}

func (c *APIClient) QueueExchangeBindings(vhost, exchange, name string) ([]map[string]interface{}, error) {
	return c.readSlice([]string{"bindings", vhost, "e", exchange, "q", name})
}

func APIQueueExchangeBindings(api, user, passwd, vhost, exchange, name string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).QueueExchangeBindings(vhost, exchange, name)
}

func (c *APIClient) QueueMessages(vhost, name string, params map[string]interface{}) ([]map[string]interface{}, error) {
	resp, err := c.POST([]string{"queues", vhost, name, "get"}, params)
	return c.ScanSlice(resp, err)
}

func APIQueueMessages(api, user, passwd, vhost, name string, params map[string]interface{}) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).QueueMessages(vhost, name, params)
}

func (c *APIClient) QueueCreateExchangeBinding(vhost, exchange, name, key string, args map[string]interface{}) (string, error) {
	params := make(map[string]interface{}, 0)
	if key != "" {
		params["routing_key"] = key
	}
	if args != nil {
		params["arguments"] = args
	}
	resp, err := c.POST([]string{"bindings", vhost, "e", exchange, "q", name}, params)
	if err != nil {
		return "", err
	}
	if (resp.StatusCode == http.StatusOK) || (resp.StatusCode == http.StatusCreated) {
		return resp.Header.Get("Location"), nil
	} else {
		return "", fmt.Errorf("API Response Status Error: %d, %v", resp.StatusCode, resp)
	}
}

func APIQueueCreateExchangeBinding(api, user, passwd, vhost, exchange, name, key string, args map[string]interface{}) (string, error) {
	return NewAPIClient(api, user, passwd).QueueCreateExchangeBinding(vhost, exchange, name, key, args)
}

func (c *APIClient) QueueExchangeBinding(vhost, exchange, name, props string) (map[string]interface{}, error) {
	return c.readMap([]string{"bindings", vhost, "e", exchange, "q", name, props})
}

func APIQueueExchangeBinding(api, user, passwd, vhost, exchange, name, props string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).QueueExchangeBinding(vhost, exchange, name, props)
}

func (c *APIClient) QueueDeleteExchangeBinding(vhost, exchange, name, props string) error {
	return c.delete([]string{"bindings", vhost, "e", exchange, "q", name, props})
}

func APIQueueDeleteExchangeBinding(api, user, passwd, vhost, exchange, name, props string) error {
	return NewAPIClient(api, user, passwd).QueueDeleteExchangeBinding(vhost, exchange, name, props)
}

func (c *APIClient) Bindings(vhost string) ([]map[string]interface{}, error) {
	if vhost == "" {
		return c.readSlice("bindings")
	} else {
		return c.readSlice([]string{"bindings", vhost})
	}
}

func (c *APIClient) APIBindings(api, user, passwd, vhost string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).Bindings(vhost)
}
