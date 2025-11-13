package config

// ConsulInfo consul配置信息
type ConsulInfo struct {
	Addr             string   `json:"addr" yaml:"addr"`
	Port             uint     `json:"port" yaml:"port"`
	Prefix           string   `json:"prefix" yaml:"prefix"`
	Timeout          int32    `json:"timeout" yaml:"timeout"`
	RegisterInterval uint     `json:"register_interval" yaml:"register_interval"`
	RegisterTtl      uint     `json:"register_ttl" yaml:"register_ttl"`
	Token            string   `json:"token" yaml:"token"`
	RegistryAddrs    []string `json:"registry_addrs" yaml:"registry_addrs"`
}
