package kafkareporter

import (
	"github.com/Shopify/sarama"
)

type Reporter interface {
	Report([]byte)
	Stop() error
}

type KafkaReporterConfig struct {
	BrokerAddr string
	Topic      string
}

type kafkaReporter struct {
	KafkaReporterConfig
	producer sarama.AsyncProducer
}

func NewReporter(c KafkaReporterConfig) (Reporter, error) {
	logger.WithField("config", c).Info("NewReporter")
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V0_10_1_0
	producer, err := sarama.NewAsyncProducer([]string{c.BrokerAddr}, saramaConfig)
	if err != nil {
		//TODO Cosmin
		return nil, err
	}
	return kafkaReporter{c, producer}, nil
}

func (kr *kafkaReporter) Report(b []byte) {
	kr.producer.Input() <- &sarama.ProducerMessage{
		Topic: kr.Topic,
		Value: sarama.ByteEncoder(b),
	}
}

func (kr *kafkaReporter) Stop() error {
	logger.Info("Stop")
	if err := kr.producer.Close(); err != nil {
		//TODO Cosmin
		return err
	}
	return nil
}
