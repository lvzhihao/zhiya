package rmqtool

import (
	"fmt"
	"net/http"
	"net/url"
)

func (c *APIClient) ListExchanges() ([]map[string]interface{}, error) {
	return c.readSlice("exchanges")
}

func APIListExchanges(api, user, passwd string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListExchanges()
}

func (c *APIClient) ListVhostExchanges(name string) ([]map[string]interface{}, error) {
	return c.readSlice([]string{"exchanges", name})
}

func APIListVhostExchanges(api, user, passwd, name string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListVhostExchanges(name)
}

func (c *APIClient) Exchange(vhost, name string) (map[string]interface{}, error) {
	return c.readMap([]string{"exchanges", vhost, name})
}

func APIExchange(api, user, passwd, vhost, name string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).Exchange(vhost, name)
}

func (c *APIClient) CreateExchange(vhost, name string, params map[string]interface{}) error {
	// `{"type":"topic","auto_delete":false,"durable":true,"internal":false,"arguments":[]}`
	return c.create([]string{"exchanges", vhost, name}, params)
}

func APICreateExchange(api, user, passwd, vhost, name string, params map[string]interface{}) error {
	return NewAPIClient(api, user, passwd).CreateExchange(vhost, name, params)
}

func (c *APIClient) DeleteExchange(vhost, name string) error {
	req, err := c.NewRequest("delete", []string{"exchanges", vhost, name}, nil)
	if err != nil {
		return err
	}
	query := &url.Values{}
	query.Add("if-unused", "true")
	req.URL.RawQuery = query.Encode()
	resp, err := c.Do(req)
	return c.ScanDelete(resp, err)
}

func APIDeleteExchange(api, user, passwd, vhost, name string) error {
	return NewAPIClient(api, user, passwd).DeleteExchange(vhost, name)
}

func (c *APIClient) ForceDeleteExchange(vhost, name string) error {
	return c.delete([]string{"exchanges", vhost, name})
}

func APIForceDeleteExchange(api, user, passwd, vhost, name string) error {
	return NewAPIClient(api, user, passwd).ForceDeleteExchange(vhost, name)
}

func (c *APIClient) ListExchangeSourceBindings(vhost, name string) ([]map[string]interface{}, error) {
	return c.readSlice([]string{"exchanges", vhost, name, "bindings", "source"})
}

func APIListExchangeSourceBindings(api, user, passwd, vhost, name string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListExchangeSourceBindings(vhost, name)
}

func (c *APIClient) ListExchangeDestinationBindings(vhost, name string) ([]map[string]interface{}, error) {
	return c.readSlice([]string{"exchanges", vhost, name, "bindings", "destination"})
}

func APIListExchangeDestinationBindings(api, user, passwd, vhost, name string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListExchangeDestinationBindings(vhost, name)
}

func (c *APIClient) ExchangePublish(vhost, name string, params map[string]interface{}) (map[string]interface{}, error) {
	req, err := c.NewRequest("post", []string{"exchanges", vhost, name, "publish"}, params)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	return c.ScanMap(resp, err)
}

func APIExchangePublish(api, user, passwd, vhost, name string, params map[string]interface{}) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ExchangePublish(vhost, name, params)
}

func (c *APIClient) ExchangeBindings(vhost, source, destination string) ([]map[string]interface{}, error) {
	return c.readSlice([]string{"bindings", vhost, "e", source, "e", destination})
}

func APIExchangeBindings(api, user, passwd, vhost, source, destination string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ExchangeBindings(vhost, source, destination)
}

func (c *APIClient) ExchangeCreateBinding(vhost, source, destination, key string, args map[string]interface{}) (string, error) {
	params := make(map[string]interface{}, 0)
	if key != "" {
		params["routing_key"] = key
	}
	if args != nil {
		params["arguments"] = args
	}
	resp, err := c.POST([]string{"bindings", vhost, "e", source, "e", destination}, params)
	if err != nil {
		return "", err
	}
	if (resp.StatusCode == http.StatusOK) || (resp.StatusCode == http.StatusCreated) {
		return resp.Header.Get("Location"), nil
	} else {
		return "", fmt.Errorf("API Response Status Error: %d, %v", resp.StatusCode, resp)
	}
}

func APIExchangeCreateBinding(api, user, passwd, vhost, source, destination, key string, args map[string]interface{}) (string, error) {
	return NewAPIClient(api, user, passwd).ExchangeCreateBinding(vhost, source, destination, key, args)
}

func (c *APIClient) ExchangeBinding(vhost, source, destination, props string) (map[string]interface{}, error) {
	return c.readMap([]string{"bindings", vhost, "e", source, "e", destination, props})
}

func APIExchangeBinding(api, user, passwd, vhost, source, destination, props string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ExchangeBinding(vhost, source, destination, props)
}

func (c *APIClient) ExchangeDeleteBinding(vhost, source, destination, props string) error {
	return c.delete([]string{"bindings", vhost, "e", source, "e", destination, props})
}

func APIExchangeDeleteBinding(api, user, passwd, vhost, source, destination, props string) error {
	return NewAPIClient(api, user, passwd).ExchangeDeleteBinding(vhost, source, destination, props)
}
