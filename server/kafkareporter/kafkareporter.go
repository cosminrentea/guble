package kafkareporter

import (
	"github.com/Shopify/sarama"
	"time"
	"io"
)

type Reporter interface {
	io.Closer
	Report([]byte)
}

type Config struct {
	Brokers []string
	Topic   string
}

type reporter struct {
	Config

	producer sarama.AsyncProducer
}

func NewReporter(c Config) (Reporter, error) {
	logger.WithField("config", c).Info("NewReporter")
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V0_10_1_0
	saramaConfig.Producer.Return.Errors = false
	saramaConfig.Producer.Retry.Max = 10
	saramaConfig.Producer.Retry.Backoff = time.Second
	p, err := sarama.NewAsyncProducer(c.Brokers, saramaConfig)
	if err != nil {
		logger.WithError(err).Error("Could not create AsyncProducer")
		return nil, err
	}
	return &reporter{
		Config:   c,
		producer: p,
	}, nil
}

func (r *reporter) Report(b []byte) {
	r.producer.Input() <- &sarama.ProducerMessage{
		Topic: r.Topic,
		Value: sarama.ByteEncoder(b),
	}
}

func (r *reporter) Close() error {
	logger.Info("Close")
	if err := r.producer.Close(); err != nil {
		logger.WithError(err).Error("Could not close Kafka Producer")
		return err
	}
	return nil
}
