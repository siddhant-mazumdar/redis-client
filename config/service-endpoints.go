package config

type serviceEndpoints struct {
	ServiceName string `mapstructure:"serviceName"`
	Endpoint    string `mapstructure:"endpoint"`
}

func (c config) GetServiceEndpoint(name string) string {
	for _, serviceEndpoint := range c.ServiceEndpoints {
		if serviceEndpoint.ServiceName == name {
			return serviceEndpoint.Endpoint
		}
	}
	return ""
}
