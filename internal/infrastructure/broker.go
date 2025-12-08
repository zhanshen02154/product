package infrastructure

import (
	"github.com/Shopify/sarama"
	"github.com/go-micro/plugins/v4/broker/kafka"
	"github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/broker"
	"log"
	"os"
	"time"
)

// 加载Kafka配置
func loadKafkaConfig(conf *config.Kafka) *sarama.Config {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.ClientID = "product-client"
	kafkaConfig.Version = sarama.V3_0_0_0
	kafkaConfig.Net.DialTimeout = time.Duration(conf.DialTimeout) * time.Second
	kafkaConfig.Net.ReadTimeout = time.Duration(conf.ReadTimeout) * time.Second
	kafkaConfig.Net.WriteTimeout = time.Duration(conf.WriteTimeout) * time.Second
	kafkaConfig.Producer.Retry.Backoff = time.Duration(conf.Producer.MaxRetryBackOff) * time.Second
	kafkaConfig.Producer.Retry.Max = conf.Producer.MaxRetry
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Flush.Bytes = conf.Producer.FlushBytes
	kafkaConfig.Producer.Flush.Frequency = 100 * time.Millisecond
	kafkaConfig.Producer.Compression = sarama.CompressionGZIP
	kafkaConfig.Producer.Partitioner = sarama.NewHashPartitioner
	kafkaConfig.Producer.Idempotent = false
	kafkaConfig.Metadata.AllowAutoTopicCreation = false
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConfig.Consumer.Fetch.Max = conf.Consumer.Group.FetchMax
	kafkaConfig.Consumer.Fetch.Min = conf.Consumer.Group.FetchMin
	kafkaConfig.Consumer.Fetch.Default = 1024 * 1024
	kafkaConfig.Consumer.MaxProcessingTime = 300000 * time.Millisecond
	kafkaConfig.Net.MaxOpenRequests = 8
	kafkaConfig.Consumer.Group.Session.Timeout = time.Second * time.Duration(conf.Consumer.Group.SessionTimeout)
	kafkaConfig.Consumer.Group.Heartbeat.Interval = time.Duration(conf.Consumer.Group.HeartbeatInterval) * time.Second
	kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	return kafkaConfig
}

// NewKafkaBroker 创建Broker
func NewKafkaBroker(conf *config.Kafka) broker.Broker {
	sarama.Logger = log.New(os.Stdout, "[Sarama]", log.LstdFlags)
	return kafka.NewBroker(
		broker.Addrs(conf.Hosts...),
		kafka.BrokerConfig(loadKafkaConfig(conf)),
	)
}
