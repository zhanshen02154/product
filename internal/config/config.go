package config

type SysConfig struct {
	Service ServiceInfo `json:"service" yaml:"service"`
	Database MySqlConfig `json:"database" yaml:"database"`
	Consul  ConsulInfo  `json:"consul" yaml:"consul"`
}

type ServiceInfo struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
	Listen  string `json:"listen" yaml:"listen"`
	Port    uint   `json:"port" yaml:"port"`
}