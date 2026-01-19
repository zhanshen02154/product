package config

import (
	"errors"
	"github.com/go-micro/plugins/v4/config/source/consul"
	"github.com/zhanshen02154/product/pkg/env"
	"go-micro.dev/v4/config"
	"go-micro.dev/v4/logger"
	"strings"
)

type SysConfig struct {
	Service     *ServiceInfo `json:"service" yaml:"service"`
	Database    *MySqlConfig `json:"database" yaml:"database"`
	Consul      *ConsulInfo  `json:"consul" yaml:"consul"`
	Transaction *Transaction `yaml:"transaction" json:"transaction"`
	Broker      *Broker      `json:"broker" yaml:"broker"`
	Tracer      *Tracer      `json:"tracer" yaml:"tracer"`
	Redis       *Redis       `json:"redis" yaml:"redis"`
}

type ServiceInfo struct {
	Name                 string `json:"name" yaml:"name"`
	Version              string `json:"version" yaml:"version"`
	Listen               string `json:"listen" yaml:"listen"`
	Port                 uint   `json:"port" yaml:"port"`
	Debug                bool   `json:"debug" yaml:"debug"`
	HeathCheckAddr       string `json:"heath_check_addr" yaml:"heath_check_addr"`
	Qps                  int    `json:"qps" yaml:"qps"`
	RequestSlowThreshold int64  `json:"request_slow_threshold" yaml:"request_slow_threshold"`
	LogLevel             string `json:"log_level" yaml:"log_level"`
}

// Redis Redis配置
type Redis struct {
	Addr           string `json:"addr" yaml:"addr"`
	Password       string `json:"password" yaml:"password"`
	Database       int    `json:"database" yaml:"database"`
	PoolSize       int    `json:"pool_size" yaml:"pool_size"`
	DialTimeout    int    `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout    int    `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout   int    `json:"write_timeout" yaml:"write_timeout"`
	MinIdleConns   int    `json:"min_idle_conns" yaml:"min_idle_conns"`
	Prefix         string `json:"prefix" yaml:"prefix"`
	LockTries      int    `json:"lock_tries" yaml:"lock_tries"`
	LockRetryDelay int    `json:"lock_retry_delay" yaml:"lock_retry_delay"`
	LockDB         int    `json:"lock_db" yaml:"lock_db"`
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
	Dsn             string `json:"dsn" yaml:"dsn"`
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
	Driver                 string   `json:"driver" yaml:"driver"`
	Kafka                  *Kafka   `json:"kafka" yaml:"kafka"`
	Publisher              []string `json:"publisher" yaml:"publisher"`
	PublishTimeThreshold   int64    `json:"publish_time_threshold" yaml:"publish_time_threshold"`
	SubscribeSlowThreshold int64    `json:"subscribe_slow_threshold" yaml:"subscribe_slow_threshold"`
}

type Kafka struct {
	Hosts             []string       `json:"hosts" yaml:"hosts"`
	DialTimeout       int            `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout       int            `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout      int            `json:"write_timeout" yaml:"write_timeout"`
	Producer          *KafkaProducer `json:"producer" yaml:"producer"`
	Consumer          *KafkaConsumer `json:"consumer" yaml:"consumer"`
	ChannelBufferSize int            `json:"channel_buffer_size" yaml:"channel_buffer_size"`
}

type KafkaProducer struct {
	MaxRetry        int `json:"max_retry" yaml:"max_retry"`
	MaxRetryBackOff int `json:"max_retry_back_off" yaml:"max_retry_back_off"`
	FlushBytes      int `json:"flush_bytes" yaml:"flush_bytes"`
	MaxOpenRequests int `json:"max_open_requests" yaml:"max_open_requests"`
	FlushFrequency  int `json:"flush_frequency" yaml:"flush_frequency"`
}

type KafkaConsumer struct {
	Group             *KafkaConsumerGroup `json:"group" yaml:"group"`
	FetchMin          int32               `json:"fetch_min" yaml:"fetch_min"`
	FetchMax          int32               `json:"fetch_max" yaml:"fetch_max"`
	MaxProcessingTime int64               `json:"max_processing_time" yaml:"max_processing_time"`
}

type KafkaConsumerGroup struct {
	SessionTimeout    int `json:"session_timeout" yaml:"session_timeout"`
	HeartbeatInterval int `json:"heartbeat_interval" yaml:"heartbeat_interval"`
}

// Tracer 链路追踪
type Tracer struct {
	SampleRate float64 `json:"sample_rate" yaml:"sample_rate"`
	Client     struct {
		Insecure bool   `json:"insecure"`
		Endpoint string `json:"endpoint" yaml:"endpoint"`
		Timeout  int    `json:"timeout" yaml:"timeout"`
		Retry    struct {
			Enabled         bool `json:"enabled" yaml:"enabled"`
			InitialInterval int  `json:"initial_interval" yaml:"initial_interval"`
			MaxInterval     int  `json:"max_interval" yaml:"max_interval"`
			MaxElapsedTime  int  `json:"max_elapsed_time" yaml:"max_elapsed_time"`
		} `json:"retry" yaml:"retry"`
	} `json:"client" yaml:"client"`
}

// CheckConfig 检查配置
func (c *SysConfig) CheckConfig() error {
	if c.Service == nil {
		return errors.New("service info is nil")
	}
	if c.Consul.RegistryAddrs == nil || len(c.Consul.RegistryAddrs) == 0 {
		return errors.New("consul registry addresses cannot be empty")
	}
	if c.Broker.SubscribeSlowThreshold <= 0 || c.Broker.Kafka.Consumer.MaxProcessingTime <= 0 {
		return errors.New("invalid subscribe_slow_threshold or max_processing_time")
	}
	if c.Broker.SubscribeSlowThreshold >= c.Broker.Kafka.Consumer.MaxProcessingTime {
		return errors.New("subscribe_slow_threshold must less than kafka.consumer.max_processing_time")
	}

	// 检查Redis配置
	if c.Redis == nil {
		return errors.New("redis config is nil")
	} else {
		if c.Redis.Addr == "" {
			return errors.New("redis addr is empty")
		}
		if c.Redis.LockRetryDelay == 0 {
			c.Redis.LockRetryDelay = 500
		}
		if c.Redis.LockTries == 0 {
			c.Redis.LockTries = 3
		}
		if c.Redis.PoolSize == 0 {
			c.Redis.PoolSize = 10
		}
		if c.Redis.MinIdleConns == 0 {
			c.Redis.MinIdleConns = 1
		}
	}
	logLevels := [3]string{"info", "warn", "error"}
	if c.Service.LogLevel == "" {
		c.Service.LogLevel = "info"
	} else {
		c.Service.LogLevel = strings.ToLower(c.Service.LogLevel)
		invalidLogLevel := true
		for _, item := range logLevels {
			if item == c.Service.LogLevel {
				invalidLogLevel = false
				break
			}
		}
		if invalidLogLevel {
			return errors.New("invalid log level")
		}
	}

	return nil
}

// GetConfig 从consul获取配置
func GetConfig() (config.Config, error) {
	// 从consul获取配置
	consulHost := env.GetEnv("CONSUL_HOST", "127.0.0.1:8500")
	consulPrefix := env.GetEnv("CONSUL_PREFIX", "/micro/")
	consulSource := consul.NewSource(
		// Set configuration address
		consul.WithAddress(consulHost),
		consul.WithPrefix(consulPrefix),
		consul.StripPrefix(true),
	)
	configInfo, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	// Load config
	if err := configInfo.Load(consulSource); err != nil {
		logger.Error("failed to load source on consul: ", err)
		if err := configInfo.Close(); err != nil {
			logger.Error("failed to close config", err)
		}
		return nil, err
	}
	return configInfo, nil
}
