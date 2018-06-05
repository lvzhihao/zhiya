package rmqtool

func (c *APIClient) QuickRegisterQueue(vhost, name string, bindings map[string][]string) error {
	err := c.CreateQueue(vhost, name, nil)
	if err != nil {
		return err
	}
	for exchange, keys := range bindings {
		for _, key := range keys {
			if exchange != "" && key != "" {
				_, err := c.QueueCreateExchangeBinding(vhost, exchange, name, key, nil)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func APIQuickRegisterQueue(api, user, passwd, vhost, name string, bindings map[string][]string) error {
	return NewAPIClient(api, user, passwd).QuickRegisterQueue(vhost, name, bindings)
}
