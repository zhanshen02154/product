package monitor

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"go-micro.dev/v4/client"
)

var (
	metricPrefix         = "kafka_"
	MessageProducedCount *prometheus.CounterVec
	ProduceDuration      *prometheus.HistogramVec
	MessagesInFlight     *prometheus.GaugeVec
)

func init() {
	if MessageProducedCount == nil {
		MessageProducedCount = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: metricPrefix + "messages_produced_count",
			Help: "kafka message produced to topic count",
		}, []string{"topic", "status", "service", "version"})
	}

	if ProduceDuration == nil {
		ProduceDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    metricPrefix + "messages_produce_duration",
			Help:    "kafka message produce duration by topic",
			Buckets: prometheus.DefBuckets,
		}, []string{"topic", "service", "version"})
	}

	if MessagesInFlight == nil {
		MessagesInFlight = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: metricPrefix + "messages_in_flight",
			Help: "kafka message in flight by topic",
		}, []string{"topic", "service", "version"})
	}

	prometheus.MustRegister(MessageProducedCount, ProduceDuration, MessagesInFlight)
}

type monitorOptions struct {
	name    string
	version string
}

type Option func(opts *monitorOptions)

type asyncProducerWrapper struct {
	client.Client
	opts *monitorOptions
}

// WithName 名称
func WithName(name string) Option {
	return func(opts *monitorOptions) {
		opts.name = name
	}
}

// WithVersion 版本
func WithVersion(version string) Option {
	return func(opts *monitorOptions) {
		opts.version = version
	}
}

// Publish 发布
func (apw *asyncProducerWrapper) Publish(ctx context.Context, msg client.Message, opts ...client.PublishOption) error {
	err := apw.Client.Publish(ctx, msg, opts...)
	if err == nil {
		MessagesInFlight.WithLabelValues(msg.Topic(), apw.opts.name, apw.opts.version).Inc()
	}
	return err
}

// NewClientWrapper 发布事件包装器
func NewClientWrapper(opts ...Option) client.Wrapper {
	monitorOpts := monitorOptions{}
	for _, opt := range opts {
		opt(&monitorOpts)
	}

	return func(c client.Client) client.Client {
		handler := &asyncProducerWrapper{
			opts:   &monitorOpts,
			Client: c,
		}

		return handler
	}
}
