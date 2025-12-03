package config

type SysConfig struct {
	Service     *ServiceInfo `json:"service" yaml:"service"`
	Database    *MySqlConfig `json:"database" yaml:"database"`
	Consul      *ConsulInfo  `json:"consul" yaml:"consul"`
	Etcd        *Etcd        `json:"etcd" yaml:"etcd"`
	Transaction *Transaction `yaml:"transaction" json:"transaction"`
	Broker      *Broker      `json:"broker" yaml:"broker"`
}

type ServiceInfo struct {
	Name           string `json:"name" yaml:"name"`
	Version        string `json:"version" yaml:"version"`
	Listen         string `json:"listen" yaml:"listen"`
	Port           uint   `json:"port" yaml:"port"`
	Debug          bool   `json:"debug" yaml:"debug"`
	HeathCheckAddr string `json:"heath_check_addr" yaml:"heath_check_addr"`
	Qps            int    `json:"qps" yaml:"qps"`
}

type Etcd struct {
	Hosts            []string `json:"hosts" yaml:"hosts"`
	DialTimeout      int64    `json:"dial_timeout" yaml:"dial_timeout"`
	Username         string   `yaml:"username" json:"username"`
	Password         string   `yaml:"password" json:"password"`
	AutoSyncInterval int64    `yaml:"auto_sync_interval" json:"auto_sync_interval"`
	Prefix           string   `yaml:"prefix" json:"prefix"`
}

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

// MySqlConfig mysql数据库配置
type MySqlConfig struct {
	Host            string `json:"host" yaml:"host"`
	Port            int64  `json:"port" yaml:"port"`
	User            string `json:"user" yaml:"user"`
	Password        string `json:"password" yaml:"password"`
	Database        string `json:"database" yaml:"database"`
	Loc             string `json:"loc" yaml:"loc"`
	Charset         string `json:"charset" yaml:"charset"`
	MaxOpenConns    int    `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifeTime uint   `json:"conn_max_life_time" yaml:"conn_max_life_time"`
}

// Transaction 事务管理
type Transaction struct {
	Driver string `json:"driver" yaml:"driver"`
	Host   string `json:"host" yaml:"host"`
}

type Broker struct {
	Driver     string   `json:"driver" yaml:"driver"`
	Kafka      *Kafka   `json:"kafka" yaml:"kafka"`
	Publisher  []string `json:"publisher" yaml:"publisher"`
	Subscriber []string `json:"subscriber" yaml:"subscriber"`
}

type Kafka struct {
	Hosts        []string       `json:"hosts" yaml:"hosts"`
	DialTimeout  int            `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  int            `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout int            `json:"write_timeout" yaml:"write_timeout"`
	Producer     *KafkaProducer `json:"producer" yaml:"producer"`
	Consumer     *KafkaConsumer `json:"consumer" yaml:"consumer"`
}

type KafkaProducer struct {
	MaxRetry        int `json:"max_retry" yaml:"max_retry"`
	MaxRetryBackOff int `json:"max_retry_back_off" yaml:"max_retry_back_off"`
	FlushBytes      int `json:"flush_bytes" yaml:"flush_bytes"`
	MaxOpenRequests int `json:"max_open_requests" yaml:"max_open_requests"`
}

type KafkaConsumer struct {
	Group            *KafkaConsumerGroup `json:"group" yaml:"group"`
	AutoCommitOffset bool                `json:"auto_commit_offset" yaml:"auto_commit_offset"`
}

type KafkaConsumerGroup struct {
	SessionTimeout int `json:"session_timeout" yaml:"session_timeout"`
}
