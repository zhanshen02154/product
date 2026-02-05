package infrastructure

import (
	"github.com/zhanshen02154/product/internal/infrastructure/event/wrapper"
	"time"

	"github.com/Shopify/sarama"
	"github.com/go-micro/plugins/v4/broker/kafka"
	"github.com/zhanshen02154/product/internal/config"
	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/logger"
)

// 加载Kafka配置
func loadKafkaConfig(conf *config.Kafka) *sarama.Config {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.ClientID = "product-client"
	kafkaConfig.Version = sarama.V3_0_0_0
	kafkaConfig.ChannelBufferSize = conf.ChannelBufferSize
	kafkaConfig.Net.DialTimeout = time.Duration(conf.DialTimeout) * time.Second
	kafkaConfig.Net.ReadTimeout = time.Duration(conf.ReadTimeout) * time.Second
	kafkaConfig.Net.WriteTimeout = time.Duration(conf.WriteTimeout) * time.Second
	kafkaConfig.Producer.Retry.Backoff = time.Duration(conf.Producer.MaxRetryBackOff) * time.Second
	kafkaConfig.Producer.Retry.Max = conf.Producer.MaxRetry
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Flush.Bytes = conf.Producer.FlushBytes
	kafkaConfig.Producer.Flush.Frequency = time.Duration(conf.Producer.FlushFrequency) * time.Millisecond
	kafkaConfig.Producer.Compression = sarama.CompressionGZIP
	kafkaConfig.Producer.Partitioner = sarama.NewHashPartitioner
	kafkaConfig.Producer.Idempotent = true
	kafkaConfig.Metadata.AllowAutoTopicCreation = false
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConfig.Consumer.Fetch.Max = conf.Consumer.FetchMax
	kafkaConfig.Consumer.Fetch.Min = conf.Consumer.FetchMin
	kafkaConfig.Consumer.Fetch.Default = 10240
	kafkaConfig.Consumer.Offsets.AutoCommit.Interval = 5 * time.Second
	kafkaConfig.Consumer.MaxProcessingTime = time.Duration(conf.Consumer.MaxProcessingTime) * time.Millisecond
	kafkaConfig.Net.MaxOpenRequests = 1
	kafkaConfig.Consumer.Group.Session.Timeout = time.Second * time.Duration(conf.Consumer.Group.SessionTimeout)
	kafkaConfig.Consumer.Group.Heartbeat.Interval = time.Duration(conf.Consumer.Group.HeartbeatInterval) * time.Second
	kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	return kafkaConfig
}

// NewKafkaBroker 创建Broker
func NewKafkaBroker(conf *config.Kafka, opts ...broker.Option) broker.Broker {
	// 将额外传入的 broker.Option 直接透传给 kafka.NewBroker，便于注入 AsyncProducer channels
	options := []broker.Option{
		broker.Addrs(conf.Hosts...),
		kafka.BrokerConfig(loadKafkaConfig(conf)),
		broker.Logger(logger.DefaultLogger),
		broker.ErrorHandler(wrapper.ErrorHandler()),
	}
	options = append(options, opts...)
	return kafka.NewBroker(options...)
}
